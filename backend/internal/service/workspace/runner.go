package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/models"
	"rolecraft-ai/internal/service/collab"
)

type Runner struct {
	db           *gorm.DB
	orchestrator *collab.Orchestrator
}

type executionPolicy struct {
	ExecutionMode    string `json:"executionMode"`
	TimeoutSeconds   int    `json:"timeoutSeconds"`
	MaxRetries       int    `json:"maxRetries"`
	RetryDelaySecond int    `json:"retryDelaySeconds"`
	ArchiveToCompany bool   `json:"archiveToCompany"`
	QueueRetryOnFail bool   `json:"queueRetryOnFailure"`
	RetryWindowMins  int    `json:"retryWindowMinutes"`
	MaxFailureCycles int    `json:"maxFailureCycles"`
}

func NewRunner(db *gorm.DB, cfg *config.Config) *Runner {
	return &Runner{
		db:           db,
		orchestrator: collab.NewOrchestrator(cfg),
	}
}

func (r *Runner) ExecuteClaimed(ctx context.Context, work *models.Work, triggerSource string) (*models.AgentRun, error) {
	now := time.Now()
	policy := parseExecutionPolicy(work.Config, work.CompanyID)
	run := models.AgentRun{
		ID:            models.NewUUID(),
		WorkID:        work.ID,
		UserID:        work.UserID,
		CompanyID:     work.CompanyID,
		TriggerSource: triggerSource,
		Status:        "running",
		StartedAt:     &now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.db.Create(&run).Error; err != nil {
		return nil, err
	}

	attempts := make([]map[string]interface{}, 0, policy.MaxRetries+1)
	var result *collab.RunResult
	var runErr error
	totalAttempts := policy.MaxRetries + 1
	for attempt := 1; attempt <= totalAttempts; attempt++ {
		attemptStart := time.Now()
		attemptCtx, cancel := context.WithTimeout(ctx, time.Duration(policy.TimeoutSeconds)*time.Second)
		currentResult, err := r.orchestrator.Run(attemptCtx, collab.RunRequest{
			TaskName:        work.Name,
			TaskDescription: work.Description,
			TaskType:        work.Type,
			InputSource:     work.InputSource,
			ReportRule:      work.ReportRule,
			ExecutionMode:   policy.ExecutionMode,
		})
		cancel()

		attemptLog := map[string]interface{}{
			"attempt":    attempt,
			"durationMs": time.Since(attemptStart).Milliseconds(),
		}
		if err != nil {
			errMsg := sanitizeText(err.Error())
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
				errMsg = fmt.Sprintf("execution timeout after %ds", policy.TimeoutSeconds)
			}
			attemptLog["status"] = "failed"
			attemptLog["error"] = errMsg
			attempts = append(attempts, attemptLog)
			runErr = fmt.Errorf(errMsg)

			if attempt < totalAttempts && ctx.Err() == nil {
				if !waitRetry(ctx, policy.RetryDelaySecond) {
					break
				}
				continue
			}
			break
		}

		result = currentResult
		attemptLog["status"] = "completed"
		attemptLog["summary"] = clip(currentResult.Summary, 120)
		attempts = append(attempts, attemptLog)
		runErr = nil
		break
	}

	finishedAt := time.Now()
	run.FinishedAt = &finishedAt
	run.UpdatedAt = finishedAt
	work.UpdatedAt = finishedAt
	work.LastRunAt = &finishedAt

	tracePayload := map[string]interface{}{
		"attempts": attempts,
		"policy": map[string]interface{}{
			"executionMode":       policy.ExecutionMode,
			"timeoutSeconds":      policy.TimeoutSeconds,
			"maxRetries":          policy.MaxRetries,
			"retryDelaySeconds":   policy.RetryDelaySecond,
			"archiveToCompany":    policy.ArchiveToCompany,
			"queueRetryOnFailure": policy.QueueRetryOnFail,
			"retryWindowMinutes":  policy.RetryWindowMins,
			"maxFailureCycles":    policy.MaxFailureCycles,
		},
	}

	if runErr != nil {
		run.Status = "failed"
		run.ErrorMessage = sanitizeText(runErr.Error())
		run.Summary = "执行失败：" + clip(run.ErrorMessage, 120)
		retryQueued, retryMeta := r.tryQueueFailureRetry(work, policy, finishedAt)
		if retryQueued {
			tracePayload["retryQueue"] = retryMeta
			run.Summary = clip(fmt.Sprintf("%s（已加入重试队列）", run.Summary), 240)
		}
		run.Trace = models.ToJSON(tracePayload)
		if !retryQueued {
			work.AsyncStatus = "failed"
			work.NextRunAt = nil
		}
		work.Status = "todo"
		work.ResultSummary = run.Summary
	} else {
		tracePayload["steps"] = sanitizeSteps(result.Steps)
		tracePayload["nextActions"] = sanitizeList(result.NextActions)
		tracePayload["evidence"] = sanitizeList(result.Evidence)
		run.Status = "completed"
		run.Summary = clip(sanitizeText(result.Summary), 240)
		if len(attempts) > 1 {
			run.Summary = clip(fmt.Sprintf("重试 %d 次后成功。%s", len(attempts)-1, run.Summary), 240)
		}
		run.FinalAnswer = sanitizeText(result.FinalAnswer)
		run.Confidence = result.Confidence
		run.Trace = models.ToJSON(tracePayload)

		work.ResultSummary = run.Summary
		work.Status = "done"
		nextRunAt, err := ComputeNextRunAt(work.TriggerType, work.TriggerValue, work.Timezone, finishedAt)
		if err != nil {
			work.AsyncStatus = "failed"
			run.Status = "failed"
			run.ErrorMessage = sanitizeText(err.Error())
		} else {
			if work.TriggerType == "once" {
				nextRunAt = nil
			}
			work.NextRunAt = nextRunAt
			if nextRunAt == nil {
				work.AsyncStatus = "completed"
			} else {
				work.AsyncStatus = "scheduled"
			}
		}
	}

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&run).Error; err != nil {
			return err
		}
		if err := tx.Save(work).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if run.Status == "failed" {
		return &run, fmt.Errorf(run.ErrorMessage)
	}
	return &run, nil
}

func waitRetry(ctx context.Context, delaySeconds int) bool {
	if delaySeconds <= 0 {
		return true
	}
	timer := time.NewTimer(time.Duration(delaySeconds) * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (r *Runner) tryQueueFailureRetry(work *models.Work, policy executionPolicy, now time.Time) (bool, map[string]interface{}) {
	if !policy.QueueRetryOnFail || strings.TrimSpace(work.TriggerType) == "manual" {
		return false, map[string]interface{}{
			"queued": false,
			"reason": "queue disabled or manual trigger",
		}
	}

	windowStart := now.Add(-time.Duration(policy.RetryWindowMins) * time.Minute)
	var previousFailures int64
	if err := r.db.Model(&models.AgentRun{}).
		Where("work_id = ? AND status = ? AND created_at >= ?", work.ID, "failed", windowStart).
		Count(&previousFailures).Error; err != nil {
		return false, map[string]interface{}{
			"queued": false,
			"reason": "count failures failed",
			"error":  sanitizeText(err.Error()),
		}
	}

	currentCycle := int(previousFailures) + 1
	if currentCycle > policy.MaxFailureCycles {
		return false, map[string]interface{}{
			"queued":             false,
			"reason":             "failure cycles exceeded",
			"currentCycle":       currentCycle,
			"maxFailureCycles":   policy.MaxFailureCycles,
			"retryWindowMinutes": policy.RetryWindowMins,
		}
	}

	delay := policy.RetryDelaySecond
	if delay <= 0 {
		delay = 1
	}
	retryAt := now.Add(time.Duration(delay) * time.Second)
	work.AsyncStatus = "scheduled"
	work.NextRunAt = &retryAt

	return true, map[string]interface{}{
		"queued":             true,
		"retryAt":            retryAt,
		"currentCycle":       currentCycle,
		"maxFailureCycles":   policy.MaxFailureCycles,
		"retryWindowMinutes": policy.RetryWindowMins,
		"retryDelaySeconds":  delay,
	}
}

func (r *Runner) ClaimWork(workID, userID string) (models.Work, bool, error) {
	var work models.Work
	if err := r.db.Where("id = ? AND user_id = ?", workID, userID).First(&work).Error; err != nil {
		return work, false, err
	}

	result := r.db.Model(&models.Work{}).
		Where("id = ? AND user_id = ? AND async_status <> ?", workID, userID, "running").
		Updates(map[string]interface{}{
			"async_status": "running",
			"updated_at":   time.Now(),
		})
	if result.Error != nil {
		return work, false, result.Error
	}
	if result.RowsAffected == 0 {
		return work, false, nil
	}

	if err := r.db.Where("id = ? AND user_id = ?", workID, userID).First(&work).Error; err != nil {
		return work, false, err
	}
	return work, true, nil
}

func clip(s string, limit int) string {
	text := sanitizeText(s)
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}

func sanitizeText(input string) string {
	text := strings.TrimSpace(input)
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		if r == '\n' || r == '\t' {
			b.WriteRune(r)
			continue
		}
		if r < 0x20 || r == 0x7F {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func sanitizeList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		text := sanitizeText(item)
		if text != "" {
			out = append(out, text)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sanitizeSteps(steps []collab.AgentStep) []collab.AgentStep {
	if len(steps) == 0 {
		return nil
	}
	out := make([]collab.AgentStep, 0, len(steps))
	for _, step := range steps {
		agent := sanitizeText(step.Agent)
		purpose := sanitizeText(step.Purpose)
		output := sanitizeText(step.Output)
		if agent == "" && purpose == "" && output == "" {
			continue
		}
		out = append(out, collab.AgentStep{
			Agent:      agent,
			Purpose:    purpose,
			Output:     output,
			DurationMs: step.DurationMs,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseExecutionPolicy(config models.JSON, companyID string) executionPolicy {
	policy := executionPolicy{
		ExecutionMode:    "serial",
		TimeoutSeconds:   180,
		MaxRetries:       1,
		RetryDelaySecond: 3,
		ArchiveToCompany: strings.TrimSpace(companyID) != "",
		QueueRetryOnFail: true,
		RetryWindowMins:  60,
		MaxFailureCycles: 3,
	}

	text := strings.TrimSpace(string(config))
	if text == "" {
		return policy
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return policy
	}

	if value := strings.ToLower(strings.TrimSpace(toString(payload["executionMode"]))); value == "parallel" || value == "serial" {
		policy.ExecutionMode = value
	}
	if value := toInt(payload["timeoutSeconds"]); value >= 30 && value <= 1800 {
		policy.TimeoutSeconds = value
	}
	if value := toInt(payload["maxRetries"]); value >= 0 && value <= 5 {
		policy.MaxRetries = value
	}
	if value := toInt(payload["retryDelaySeconds"]); value >= 0 && value <= 120 {
		policy.RetryDelaySecond = value
	}
	if value, ok := toBool(payload["archiveToCompany"]); ok {
		policy.ArchiveToCompany = value
	}
	if value, ok := toBool(payload["queueRetryOnFailure"]); ok {
		policy.QueueRetryOnFail = value
	}
	if value := toInt(payload["retryWindowMinutes"]); value >= 5 && value <= 1440 {
		policy.RetryWindowMins = value
	}
	if value := toInt(payload["maxFailureCycles"]); value >= 1 && value <= 20 {
		policy.MaxFailureCycles = value
	}
	return policy
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func toInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil {
			return parsed
		}
	}
	return -1
}

func toBool(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes":
			return true, true
		case "0", "false", "no":
			return false, true
		}
	}
	return false, false
}
