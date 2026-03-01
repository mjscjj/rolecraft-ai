package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

type WorkHandler struct {
	db *gorm.DB
}

func NewWorkHandler(db *gorm.DB) *WorkHandler {
	return &WorkHandler{db: db}
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

func normalizeTimezone(tz string) string {
	value := strings.TrimSpace(tz)
	if value == "" {
		return "Asia/Shanghai"
	}
	return value
}

func parseTimeInLocation(raw string, location *time.Location) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
	var lastErr error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, raw, location)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

func computeNextRunAt(triggerType, triggerValue, timezone string, now time.Time) (*time.Time, error) {
	mode := strings.TrimSpace(triggerType)
	value := strings.TrimSpace(triggerValue)
	location, err := time.LoadLocation(normalizeTimezone(timezone))
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}

	switch mode {
	case "", "manual":
		return nil, nil
	case "once":
		if value == "" {
			return nil, fmt.Errorf("triggerValue required when triggerType=once")
		}
		parsed, err := parseTimeInLocation(value, location)
		if err != nil {
			return nil, fmt.Errorf("invalid once triggerValue: %w", err)
		}
		return &parsed, nil
	case "daily":
		if value == "" {
			return nil, fmt.Errorf("triggerValue required when triggerType=daily (HH:MM)")
		}
		parts := strings.Split(value, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid daily triggerValue, expected HH:MM")
		}
		hour, err := strconv.Atoi(parts[0])
		if err != nil || hour < 0 || hour > 23 {
			return nil, fmt.Errorf("invalid daily hour")
		}
		minute, err := strconv.Atoi(parts[1])
		if err != nil || minute < 0 || minute > 59 {
			return nil, fmt.Errorf("invalid daily minute")
		}
		base := now.In(location)
		next := time.Date(base.Year(), base.Month(), base.Day(), hour, minute, 0, 0, location)
		if !next.After(base) {
			next = next.Add(24 * time.Hour)
		}
		return &next, nil
	case "interval_hours":
		if value == "" {
			return nil, fmt.Errorf("triggerValue required when triggerType=interval_hours")
		}
		hours, err := strconv.Atoi(value)
		if err != nil || hours <= 0 || hours > 720 {
			return nil, fmt.Errorf("invalid interval hours, expected 1~720")
		}
		next := now.In(location).Add(time.Duration(hours) * time.Hour)
		return &next, nil
	default:
		return nil, fmt.Errorf("unsupported triggerType: %s", mode)
	}
}

func defaultAsyncStatus(triggerType string) string {
	switch strings.TrimSpace(triggerType) {
	case "", "manual":
		return "idle"
	default:
		return "scheduled"
	}
}

func buildExecutionSummary(work models.Work, now time.Time) string {
	switch work.Type {
	case "report":
		return fmt.Sprintf("[%s] 自动汇报已生成：%s（规则：%s）", now.Format("2006-01-02 15:04"), work.Name, strings.TrimSpace(work.ReportRule))
	case "analyze":
		return fmt.Sprintf("[%s] 文件分析任务已完成：%s（输入源：%s）", now.Format("2006-01-02 15:04"), work.Name, strings.TrimSpace(work.InputSource))
	default:
		return fmt.Sprintf("[%s] 异步任务已执行：%s", now.Format("2006-01-02 15:04"), work.Name)
	}
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
	timezone := normalizeTimezone(req.Timezone)
	nextRunAt, err := computeNextRunAt(triggerType, req.TriggerValue, timezone, time.Now())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asyncStatus := req.AsyncStatus
	if strings.TrimSpace(asyncStatus) == "" {
		asyncStatus = defaultAsyncStatus(triggerType)
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
			work.Timezone = normalizeTimezone(req.Timezone)
		}
		nextRunAt, err := computeNextRunAt(work.TriggerType, work.TriggerValue, work.Timezone, time.Now())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		work.NextRunAt = nextRunAt
		if req.AsyncStatus == "" {
			work.AsyncStatus = defaultAsyncStatus(work.TriggerType)
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

// Run 立即执行工作区任务（MVP：异步执行模拟）
func (h *WorkHandler) Run(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var work models.Work
	if err := h.db.Where("id = ? AND user_id = ?", id, userIDStr).First(&work).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	now := time.Now()
	work.AsyncStatus = "running"
	work.LastRunAt = &now
	work.Status = "in_progress"
	work.ResultSummary = buildExecutionSummary(work, now)

	nextRunAt, err := computeNextRunAt(work.TriggerType, work.TriggerValue, work.Timezone, now)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if work.TriggerType == "once" {
		nextRunAt = nil
	}
	work.NextRunAt = nextRunAt
	if work.NextRunAt != nil && work.TriggerType != "once" {
		work.AsyncStatus = "scheduled"
	} else {
		work.AsyncStatus = "completed"
	}
	work.Status = "done"
	work.UpdatedAt = now

	if err := h.db.Save(&work).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    work,
	})
}
