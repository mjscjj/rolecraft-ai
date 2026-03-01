package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

type CompanyHandler struct {
	db *gorm.DB
}

func NewCompanyHandler(db *gorm.DB) *CompanyHandler {
	return &CompanyHandler{db: db}
}

type CompanyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CompanyExportRequest struct {
	Format        string   `json:"format"`
	Keyword       string   `json:"keyword"`
	MinConfidence *float64 `json:"minConfidence"`
	From          string   `json:"from"`
	To            string   `json:"to"`
}

type deliveryEntry struct {
	ID          string
	WorkID      string
	WorkName    string
	Summary     string
	FinalAnswer string
	Confidence  float64
	StepCount   int
	NextActions []string
	Evidence    []string
	UpdatedAt   time.Time
}

func parseCompanyJSONMap(raw models.JSON) map[string]interface{} {
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return nil
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return nil
	}
	if len(payload) == 0 {
		return nil
	}
	return payload
}

func parseTracePayload(raw models.JSON) map[string]interface{} {
	return parseCompanyJSONMap(raw)
}

func anyToStringSlice(value interface{}) []string {
	items, ok := value.([]interface{})
	if !ok || len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if ok {
			trimmed := strings.TrimSpace(text)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func (h *CompanyHandler) ensureCompanyOwned(companyID, userID string) (models.Company, error) {
	var company models.Company
	if err := h.db.Where("id = ? AND owner_id = ?", companyID, userID).First(&company).Error; err != nil {
		return company, err
	}
	return company, nil
}

func (h *CompanyHandler) buildCompanyInsights(company models.Company) (gin.H, []gin.H, []deliveryEntry, error) {
	var roleCount int64
	var workCount int64
	var outcomeCount int64
	var docCount int64
	h.db.Model(&models.Role{}).Where("company_id = ?", company.ID).Count(&roleCount)
	h.db.Model(&models.Work{}).Where("company_id = ?", company.ID).Count(&workCount)
	h.db.Model(&models.AgentRun{}).Where("company_id = ? AND status = ?", company.ID, "completed").Count(&outcomeCount)
	h.db.Model(&models.Document{}).Where("company_id = ?", company.ID).Count(&docCount)

	stats := gin.H{
		"roleCount":      roleCount,
		"workCount":      workCount,
		"workspaceCount": workCount,
		"outcomeCount":   outcomeCount,
		"docCount":       docCount,
	}

	var recentRuns []models.AgentRun
	if err := h.db.
		Where("company_id = ? AND status = ?", company.ID, "completed").
		Order("updated_at DESC").
		Limit(30).
		Find(&recentRuns).Error; err != nil {
		return nil, nil, nil, err
	}

	workNameByID := map[string]string{}
	workIDs := make([]string, 0, len(recentRuns))
	for _, item := range recentRuns {
		if strings.TrimSpace(item.WorkID) != "" {
			workIDs = append(workIDs, item.WorkID)
		}
	}
	if len(workIDs) > 0 {
		var works []models.Work
		h.db.Select("id, name").Where("id IN ?", workIDs).Find(&works)
		for _, item := range works {
			workNameByID[item.ID] = item.Name
		}
	}

	outcomes := make([]gin.H, 0, len(recentRuns))
	deliveries := make([]deliveryEntry, 0, len(recentRuns))
	for _, item := range recentRuns {
		trace := parseTracePayload(item.Trace)
		stepCount := 0
		if trace != nil {
			if steps, ok := trace["steps"].([]interface{}); ok {
				stepCount = len(steps)
			}
		}
		outcomes = append(outcomes, gin.H{
			"id":            item.ID,
			"workId":        item.WorkID,
			"workName":      workNameByID[item.WorkID],
			"status":        item.Status,
			"confidence":    item.Confidence,
			"resultSummary": item.Summary,
			"updatedAt":     item.UpdatedAt,
		})
		deliveries = append(deliveries, deliveryEntry{
			ID:          item.ID,
			WorkID:      item.WorkID,
			WorkName:    workNameByID[item.WorkID],
			Summary:     item.Summary,
			FinalAnswer: item.FinalAnswer,
			Confidence:  item.Confidence,
			StepCount:   stepCount,
			NextActions: anyToStringSlice(trace["nextActions"]),
			Evidence:    anyToStringSlice(trace["evidence"]),
			UpdatedAt:   item.UpdatedAt,
		})
	}

	if len(outcomes) == 0 {
		var legacyWorks []models.Work
		h.db.
			Where("company_id = ? AND result_summary <> ''", company.ID).
			Order("updated_at DESC").
			Limit(30).
			Find(&legacyWorks)
		for _, item := range legacyWorks {
			outcomes = append(outcomes, gin.H{
				"id":            item.ID,
				"workId":        item.ID,
				"workName":      item.Name,
				"status":        item.AsyncStatus,
				"confidence":    0.0,
				"resultSummary": item.ResultSummary,
				"updatedAt":     item.UpdatedAt,
			})
			deliveries = append(deliveries, deliveryEntry{
				ID:          item.ID,
				WorkID:      item.ID,
				WorkName:    item.Name,
				Summary:     item.ResultSummary,
				FinalAnswer: "",
				Confidence:  0,
				StepCount:   0,
				NextActions: []string{},
				Evidence:    []string{},
				UpdatedAt:   item.UpdatedAt,
			})
		}
	}

	return stats, outcomes, deliveries, nil
}

func toDeliveryBoardPayload(entries []deliveryEntry) []gin.H {
	if len(entries) == 0 {
		return []gin.H{}
	}
	result := make([]gin.H, 0, len(entries))
	for _, item := range entries {
		result = append(result, gin.H{
			"id":          item.ID,
			"workId":      item.WorkID,
			"workName":    item.WorkName,
			"summary":     item.Summary,
			"finalAnswer": item.FinalAnswer,
			"confidence":  item.Confidence,
			"stepCount":   item.StepCount,
			"nextActions": item.NextActions,
			"evidence":    item.Evidence,
			"updatedAt":   item.UpdatedAt,
		})
	}
	return result
}

func normalizeExportFormat(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "json", "markdown", "md":
		if value == "md" {
			return "markdown"
		}
		return value
	default:
		return "markdown"
	}
}

func parseDateFilter(raw string, endOfDay bool) (*time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return &parsed, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", trimmed, time.Local)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		v := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, parsed.Location())
		return &v, nil
	}
	v := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())
	return &v, nil
}

func applyDeliveryFilter(entries []deliveryEntry, req CompanyExportRequest) ([]deliveryEntry, error) {
	keyword := strings.ToLower(strings.TrimSpace(req.Keyword))
	minConfidence := -1.0
	if req.MinConfidence != nil {
		minConfidence = *req.MinConfidence
	}
	from, err := parseDateFilter(req.From, false)
	if err != nil {
		return nil, fmt.Errorf("invalid from date")
	}
	to, err := parseDateFilter(req.To, true)
	if err != nil {
		return nil, fmt.Errorf("invalid to date")
	}

	result := make([]deliveryEntry, 0, len(entries))
	for _, item := range entries {
		if keyword != "" {
			payload := strings.ToLower(strings.Join([]string{item.WorkName, item.Summary, item.FinalAnswer}, "\n"))
			if !strings.Contains(payload, keyword) {
				continue
			}
		}
		if minConfidence >= 0 && item.Confidence < minConfidence {
			continue
		}
		if from != nil && item.UpdatedAt.Before(*from) {
			continue
		}
		if to != nil && item.UpdatedAt.After(*to) {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func sanitizeFileNamePart(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "company"
	}
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		" ", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	result := replacer.Replace(trimmed)
	result = strings.Trim(result, "-.")
	if result == "" {
		return "company"
	}
	return result
}

func buildMarkdownExport(company models.Company, entries []deliveryEntry, generatedAt time.Time) string {
	lines := make([]string, 0, len(entries)*16)
	lines = append(lines, fmt.Sprintf("# %s 交付看板", company.Name))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("导出时间：%s", generatedAt.Format("2006-01-02 15:04:05")))
	lines = append(lines, "")

	for idx, item := range entries {
		workName := item.WorkName
		if strings.TrimSpace(workName) == "" {
			workName = "任务 " + item.WorkID
		}
		lines = append(lines, fmt.Sprintf("## %d. %s", idx+1, workName))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("- 交付ID：%s", item.ID))
		lines = append(lines, fmt.Sprintf("- 置信度：%.2f", item.Confidence))
		lines = append(lines, fmt.Sprintf("- 步骤数：%d", item.StepCount))
		lines = append(lines, fmt.Sprintf("- 更新时间：%s", item.UpdatedAt.Format("2006-01-02 15:04:05")))
		lines = append(lines, "")
		lines = append(lines, "摘要：")
		if strings.TrimSpace(item.Summary) == "" {
			lines = append(lines, "无摘要")
		} else {
			lines = append(lines, item.Summary)
		}
		if strings.TrimSpace(item.FinalAnswer) != "" {
			lines = append(lines, "")
			lines = append(lines, "最终答案：")
			lines = append(lines, item.FinalAnswer)
		}
		if len(item.NextActions) > 0 {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("下一步：%s", strings.Join(item.NextActions, "；")))
		}
		if len(item.Evidence) > 0 {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("证据：%s", strings.Join(item.Evidence, "；")))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func buildExportContent(company models.Company, entries []deliveryEntry, format string, generatedAt time.Time) (string, string) {
	stamp := generatedAt.Format("20060102-150405")
	fileBase := sanitizeFileNamePart(company.Name)
	if format == "json" {
		payload := map[string]interface{}{
			"company": map[string]interface{}{
				"id":   company.ID,
				"name": company.Name,
			},
			"generatedAt": generatedAt,
			"deliveries":  toDeliveryBoardPayload(entries),
		}
		bytes, _ := json.MarshalIndent(payload, "", "  ")
		return string(bytes), fmt.Sprintf("delivery-board-%s-%s.json", fileBase, stamp)
	}
	return buildMarkdownExport(company, entries, generatedAt), fmt.Sprintf("delivery-board-%s-%s.md", fileBase, stamp)
}

func (h *CompanyHandler) List(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	var companies []models.Company
	if err := h.db.Where("owner_id = ?", userIDStr).Order("created_at DESC").Find(&companies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": companies})
}

func (h *CompanyHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	var req CompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company := models.Company{
		ID:          models.NewUUID(),
		OwnerID:     userIDStr,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.db.Create(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 200, "message": "success", "data": company})
}

func (h *CompanyHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	company, err := h.ensureCompanyOwned(id, userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	stats, outcomes, deliveries, err := h.buildCompanyInsights(company)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"company":        company,
			"stats":          stats,
			"recentOutcomes": outcomes,
			"deliveryBoard":  toDeliveryBoardPayload(deliveries),
		},
	})
}

func (h *CompanyHandler) ListExports(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	companyID := c.Param("id")

	company, err := h.ensureCompanyOwned(companyID, userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	limit := 20
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if value, parseErr := strconv.Atoi(raw); parseErr == nil && value > 0 && value <= 100 {
			limit = value
		}
	}

	var rows []models.CompanyExport
	if err := h.db.
		Where("company_id = ? AND user_id = ?", companyID, userIDStr).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload := make([]gin.H, 0, len(rows))
	for _, item := range rows {
		payload = append(payload, gin.H{
			"id":            item.ID,
			"companyId":     item.CompanyID,
			"companyName":   company.Name,
			"format":        item.Format,
			"fileName":      item.FileName,
			"deliveryCount": item.DeliveryCount,
			"filters":       parseCompanyJSONMap(item.Filters),
			"createdAt":     item.CreatedAt,
			"updatedAt":     item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": payload})
}

func (h *CompanyHandler) GetExport(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	companyID := c.Param("id")
	exportID := c.Param("exportId")

	company, err := h.ensureCompanyOwned(companyID, userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	var item models.CompanyExport
	if err := h.db.
		Where("id = ? AND company_id = ? AND user_id = ?", exportID, companyID, userIDStr).
		First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":            item.ID,
			"companyId":     item.CompanyID,
			"companyName":   company.Name,
			"format":        item.Format,
			"fileName":      item.FileName,
			"deliveryCount": item.DeliveryCount,
			"filters":       parseCompanyJSONMap(item.Filters),
			"content":       item.Content,
			"createdAt":     item.CreatedAt,
			"updatedAt":     item.UpdatedAt,
		},
	})
}

func (h *CompanyHandler) CreateExport(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	companyID := c.Param("id")

	company, err := h.ensureCompanyOwned(companyID, userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	var req CompanyExportRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, _, deliveries, err := h.buildCompanyInsights(company)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filtered, err := applyDeliveryFilter(deliveries, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(filtered) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no delivery records for current filter"})
		return
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].UpdatedAt.After(filtered[j].UpdatedAt)
	})

	format := normalizeExportFormat(req.Format)
	now := time.Now()
	content, fileName := buildExportContent(company, filtered, format, now)

	filters := map[string]interface{}{}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		filters["keyword"] = keyword
	}
	if req.MinConfidence != nil {
		filters["minConfidence"] = *req.MinConfidence
	}
	if from := strings.TrimSpace(req.From); from != "" {
		filters["from"] = from
	}
	if to := strings.TrimSpace(req.To); to != "" {
		filters["to"] = to
	}

	record := models.CompanyExport{
		ID:            models.NewUUID(),
		CompanyID:     company.ID,
		UserID:        userIDStr,
		Format:        format,
		FileName:      fileName,
		DeliveryCount: len(filtered),
		Filters:       models.ToJSON(filters),
		Content:       content,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := h.db.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":            record.ID,
			"companyId":     record.CompanyID,
			"companyName":   company.Name,
			"format":        record.Format,
			"fileName":      record.FileName,
			"deliveryCount": record.DeliveryCount,
			"filters":       parseCompanyJSONMap(record.Filters),
			"content":       record.Content,
			"createdAt":     record.CreatedAt,
			"updatedAt":     record.UpdatedAt,
		},
	})
}

func (h *CompanyHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var req CompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	if err := h.db.Where("id = ? AND owner_id = ?", id, userIDStr).First(&company).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	company.Name = req.Name
	company.Description = req.Description
	company.UpdatedAt = time.Now()

	if err := h.db.Save(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": company})
}

func (h *CompanyHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var company models.Company
	if err := h.db.Where("id = ? AND owner_id = ?", id, userIDStr).First(&company).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Role{}).Where("company_id = ?", id).Update("company_id", "").Error; err != nil {
			return err
		}
		if err := tx.Where("company_id = ?", id).Delete(&models.Work{}).Error; err != nil {
			return err
		}
		if err := tx.Where("company_id = ?", id).Delete(&models.AgentRun{}).Error; err != nil {
			return err
		}
		if err := tx.Where("company_id = ?", id).Delete(&models.CompanyExport{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Document{}).Where("company_id = ?", id).Update("company_id", "").Error; err != nil {
			return err
		}
		return tx.Delete(&models.Company{}, "id = ? AND owner_id = ?", id, userIDStr).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}
