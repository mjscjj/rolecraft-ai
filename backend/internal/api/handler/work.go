package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
	workspaceSvc "rolecraft-ai/internal/service/workspace"
)

type WorkHandler struct {
	db     *gorm.DB
	runner *workspaceSvc.Runner
}

type AgentRunResponse struct {
	ID            string                 `json:"id"`
	WorkID        string                 `json:"workId"`
	UserID        string                 `json:"userId"`
	CompanyID     string                 `json:"companyId,omitempty"`
	TriggerSource string                 `json:"triggerSource"`
	Status        string                 `json:"status"`
	Summary       string                 `json:"summary"`
	FinalAnswer   string                 `json:"finalAnswer"`
	Confidence    float64                `json:"confidence"`
	Trace         map[string]interface{} `json:"trace,omitempty"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
	StartedAt     *time.Time             `json:"startedAt,omitempty"`
	FinishedAt    *time.Time             `json:"finishedAt,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

func NewWorkHandler(db *gorm.DB, runner *workspaceSvc.Runner) *WorkHandler {
	return &WorkHandler{db: db, runner: runner}
}

type WorkRequest struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	CompanyID     string                 `json:"companyId"`
	Status        string                 `json:"status"`
	Priority      string                 `json:"priority"`
	RoleID        string                 `json:"roleId"`
	Type          string                 `json:"type"`         // general/report/analyze
	TriggerType   string                 `json:"triggerType"`  // manual/once/daily/interval_hours
	TriggerValue  string                 `json:"triggerValue"` // 09:00 / 4 / RFC3339
	Timezone      string                 `json:"timezone"`
	AsyncStatus   string                 `json:"asyncStatus"`
	InputSource   string                 `json:"inputSource"`
	ReportRule    string                 `json:"reportRule"`
	ResultSummary string                 `json:"resultSummary"`
	Config        map[string]interface{} `json:"config"`
}

type BatchRunRequest struct {
	IDs         []string `json:"ids"`
	MaxParallel int      `json:"maxParallel"`
}

type BatchRunItemResponse struct {
	WorkID string            `json:"workId"`
	Status string            `json:"status"`
	Error  string            `json:"error,omitempty"`
	Run    *AgentRunResponse `json:"run,omitempty"`
}

func parseJSONMap(raw models.JSON) map[string]interface{} {
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return nil
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return map[string]interface{}{
			"raw": text,
		}
	}
	if len(payload) == 0 {
		return nil
	}
	return payload
}

func toAgentRunResponse(run models.AgentRun) AgentRunResponse {
	return AgentRunResponse{
		ID:            run.ID,
		WorkID:        run.WorkID,
		UserID:        run.UserID,
		CompanyID:     run.CompanyID,
		TriggerSource: run.TriggerSource,
		Status:        run.Status,
		Summary:       run.Summary,
		FinalAnswer:   run.FinalAnswer,
		Confidence:    run.Confidence,
		Trace:         parseJSONMap(run.Trace),
		ErrorMessage:  run.ErrorMessage,
		StartedAt:     run.StartedAt,
		FinishedAt:    run.FinishedAt,
		CreatedAt:     run.CreatedAt,
		UpdatedAt:     run.UpdatedAt,
	}
}

func toAgentRunResponses(runs []models.AgentRun) []AgentRunResponse {
	if len(runs) == 0 {
		return []AgentRunResponse{}
	}
	resp := make([]AgentRunResponse, 0, len(runs))
	for _, run := range runs {
		resp = append(resp, toAgentRunResponse(run))
	}
	return resp
}

func (h *WorkHandler) List(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	var works []models.Work
	query := h.db.Where("user_id = ?", userIDStr)

	if companyID := c.Query("companyId"); companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if triggerType := c.Query("triggerType"); triggerType != "" {
		query = query.Where("trigger_type = ?", triggerType)
	}
	if asyncStatus := c.Query("asyncStatus"); asyncStatus != "" {
		query = query.Where("async_status = ?", asyncStatus)
	}

	if err := query.Order("created_at DESC").Find(&works).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": works})
}

func (h *WorkHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	var req WorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	if req.CompanyID != "" {
		var company models.Company
		if err := h.db.Where("id = ? AND owner_id = ?", req.CompanyID, userIDStr).First(&company).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "no access to this company"})
			return
		}
	}

	status := req.Status
	if status == "" {
		status = "todo"
	}
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	workType := req.Type
	if strings.TrimSpace(workType) == "" {
		workType = "general"
	}
	triggerType := req.TriggerType
	if strings.TrimSpace(triggerType) == "" {
		triggerType = "manual"
	}
	timezone := workspaceSvc.NormalizeTimezone(req.Timezone)
	nextRunAt, err := workspaceSvc.ComputeNextRunAt(triggerType, req.TriggerValue, timezone, time.Now())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asyncStatus := req.AsyncStatus
	if strings.TrimSpace(asyncStatus) == "" {
		asyncStatus = workspaceSvc.DefaultAsyncStatus(triggerType)
	}
	configJSON := models.JSON("")
	if req.Config != nil {
		configJSON = models.ToJSON(req.Config)
	}

	work := models.Work{
		ID:            models.NewUUID(),
		UserID:        userIDStr,
		CompanyID:     req.CompanyID,
		Name:          strings.TrimSpace(req.Name),
		Description:   strings.TrimSpace(req.Description),
		Status:        status,
		Priority:      priority,
		RoleID:        req.RoleID,
		Type:          workType,
		TriggerType:   triggerType,
		TriggerValue:  strings.TrimSpace(req.TriggerValue),
		Timezone:      timezone,
		NextRunAt:     nextRunAt,
		AsyncStatus:   asyncStatus,
		InputSource:   strings.TrimSpace(req.InputSource),
		ReportRule:    strings.TrimSpace(req.ReportRule),
		ResultSummary: strings.TrimSpace(req.ResultSummary),
		Config:        configJSON,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.db.Create(&work).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 200, "message": "success", "data": work})
}

func (h *WorkHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var req WorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var work models.Work
	if err := h.db.Where("id = ? AND user_id = ?", id, userIDStr).First(&work).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work not found"})
		return
	}

	if req.CompanyID != "" {
		var company models.Company
		if err := h.db.Where("id = ? AND owner_id = ?", req.CompanyID, userIDStr).First(&company).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "no access to this company"})
			return
		}
	}

	if strings.TrimSpace(req.Name) != "" {
		work.Name = strings.TrimSpace(req.Name)
	}
	work.Description = strings.TrimSpace(req.Description)
	work.CompanyID = req.CompanyID
	if req.Status != "" {
		work.Status = req.Status
	}
	if req.Priority != "" {
		work.Priority = req.Priority
	}
	if req.Type != "" {
		work.Type = req.Type
	}
	if req.TriggerType != "" {
		work.TriggerType = req.TriggerType
	}
	if req.TriggerValue != "" || req.TriggerType != "" || req.Timezone != "" {
		if req.TriggerValue != "" {
			work.TriggerValue = strings.TrimSpace(req.TriggerValue)
		}
		if req.Timezone != "" {
			work.Timezone = workspaceSvc.NormalizeTimezone(req.Timezone)
		}
		nextRunAt, err := workspaceSvc.ComputeNextRunAt(work.TriggerType, work.TriggerValue, work.Timezone, time.Now())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		work.NextRunAt = nextRunAt
		if req.AsyncStatus == "" {
			work.AsyncStatus = workspaceSvc.DefaultAsyncStatus(work.TriggerType)
		}
	}
	if req.AsyncStatus != "" {
		work.AsyncStatus = req.AsyncStatus
	}
	work.InputSource = strings.TrimSpace(req.InputSource)
	work.ReportRule = strings.TrimSpace(req.ReportRule)
	if req.ResultSummary != "" {
		work.ResultSummary = strings.TrimSpace(req.ResultSummary)
	}
	if req.Config != nil {
		work.Config = models.ToJSON(req.Config)
	}
	work.RoleID = req.RoleID
	work.UpdatedAt = time.Now()

	if err := h.db.Save(&work).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": work})
}

func (h *WorkHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	if err := h.db.Delete(&models.Work{}, "id = ? AND user_id = ?", id, userIDStr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// Run 立即执行工作区任务（多 Agent 协商）
func (h *WorkHandler) Run(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	work, claimed, err := h.runner.ClaimWork(id, userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	if !claimed {
		c.JSON(http.StatusConflict, gin.H{"error": "workspace task is running"})
		return
	}

	run, runErr := h.runner.ExecuteClaimed(c.Request.Context(), &work, "manual")
	if runErr != nil {
		// 返回最新状态给前端，便于提示
		var latest models.Work
		_ = h.db.Where("id = ? AND user_id = ?", id, userIDStr).First(&latest).Error
		var runPayload interface{}
		if run != nil {
			runPayload = toAgentRunResponse(*run)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": runErr.Error(),
			"data": gin.H{
				"work": latest,
				"run":  runPayload,
			},
		})
		return
	}

	var latest models.Work
	_ = h.db.Where("id = ? AND user_id = ?", id, userIDStr).First(&latest).Error

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"work": latest,
			"run":  toAgentRunResponse(*run),
		},
	})
}

func dedupeIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

// BatchRun 批量执行工作区任务（并发）
func (h *WorkHandler) BatchRun(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	var req BatchRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ids := dedupeIDs(req.IDs)
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids is required"})
		return
	}
	if len(ids) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch size exceeded, max 50"})
		return
	}

	maxParallel := req.MaxParallel
	if maxParallel <= 0 {
		maxParallel = 3
	}
	if maxParallel > 10 {
		maxParallel = 10
	}

	type indexedItem struct {
		index int
		item  BatchRunItemResponse
	}

	sem := make(chan struct{}, maxParallel)
	resultCh := make(chan indexedItem, len(ids))
	var wg sync.WaitGroup

	for idx, id := range ids {
		wg.Add(1)
		go func(index int, workID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := BatchRunItemResponse{
				WorkID: workID,
				Status: "failed",
			}

			work, claimed, err := h.runner.ClaimWork(workID, userIDStr)
			if err != nil {
				result.Status = "not_found"
				result.Error = "workspace not found"
				resultCh <- indexedItem{index: index, item: result}
				return
			}
			if !claimed {
				result.Status = "busy"
				result.Error = "workspace task is running"
				resultCh <- indexedItem{index: index, item: result}
				return
			}

			run, runErr := h.runner.ExecuteClaimed(c.Request.Context(), &work, "batch")
			if run != nil {
				runResp := toAgentRunResponse(*run)
				result.Run = &runResp
			}
			if runErr != nil {
				result.Status = "failed"
				if result.Run != nil && strings.TrimSpace(result.Run.ErrorMessage) != "" {
					result.Error = result.Run.ErrorMessage
				} else {
					result.Error = runErr.Error()
				}
				resultCh <- indexedItem{index: index, item: result}
				return
			}

			result.Status = "completed"
			resultCh <- indexedItem{index: index, item: result}
		}(idx, id)
	}

	wg.Wait()
	close(resultCh)

	items := make([]BatchRunItemResponse, len(ids))
	successCount := 0
	failedCount := 0
	busyCount := 0
	notFoundCount := 0
	for payload := range resultCh {
		items[payload.index] = payload.item
		switch payload.item.Status {
		case "completed":
			successCount++
		case "busy":
			busyCount++
			failedCount++
		case "not_found":
			notFoundCount++
			failedCount++
		default:
			failedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"items":         items,
			"successCount":  successCount,
			"failedCount":   failedCount,
			"busyCount":     busyCount,
			"notFoundCount": notFoundCount,
		},
	})
}

// ListRuns 获取工作区任务执行记录
func (h *WorkHandler) ListRuns(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var work models.Work
	if err := h.db.Where("id = ? AND user_id = ?", id, userIDStr).First(&work).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	limit := 20
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err == nil && value > 0 && value <= 100 {
			limit = value
		}
	}

	var runs []models.AgentRun
	if err := h.db.
		Where("work_id = ? AND user_id = ?", id, userIDStr).
		Order("created_at DESC").
		Limit(limit).
		Find(&runs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    toAgentRunResponses(runs),
	})
}

// GetRun 获取单次执行记录详情
func (h *WorkHandler) GetRun(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	workID := c.Param("id")
	runID := c.Param("runId")

	var work models.Work
	if err := h.db.Where("id = ? AND user_id = ?", workID, userIDStr).First(&work).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	var run models.AgentRun
	if err := h.db.
		Where("id = ? AND work_id = ? AND user_id = ?", runID, workID, userIDStr).
		First(&run).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "run not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    toAgentRunResponse(run),
	})
}
