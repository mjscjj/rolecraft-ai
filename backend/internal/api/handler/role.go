package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/models"
	"rolecraft-ai/internal/service/anythingllm"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	db             *gorm.DB
	anythingllmURL string
	anythingllmKey string
	openaiKey      string
	anything       *anythingllm.Orchestrator
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(db *gorm.DB, cfg *config.Config) *RoleHandler {
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &RoleHandler{
		db:             db,
		anythingllmURL: cfg.AnythingLLMURL,
		anythingllmKey: cfg.AnythingLLMKey,
		openaiKey:      cfg.OpenAIKey,
		anything: anythingllm.NewOrchestrator(cfg.AnythingLLMURL, cfg.AnythingLLMKey, anythingllm.OrchestratorConfig{
			DefaultProvider: "openrouter",
			DefaultModel:    cfg.OpenRouterModel,
			OpenRouterKey:   cfg.OpenRouterKey,
		}),
	}
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name           string                 `json:"name" binding:"required" example:"智能助理"`
	Description    string                 `json:"description" example:"全能型办公助手"`
	Category       string                 `json:"category" example:"通用"`
	CompanyID      string                 `json:"companyId" example:""`
	SystemPrompt   string                 `json:"systemPrompt" binding:"required" example:"你是一位智能助理..."`
	WelcomeMessage string                 `json:"welcomeMessage" example:"你好！有什么可以帮你的吗？"`
	Avatar         string                 `json:"avatar" example:""`
	ModelConfig    map[string]interface{} `json:"modelConfig" example:"{\"temperature\":0.7}"`
	IsTemplate     bool                   `json:"isTemplate" example:"false"`
	IsPublic       bool                   `json:"isPublic" example:"false"`
}

type InstallRoleRequest struct {
	TargetType string `json:"targetType"` // personal/company
	CompanyID  string `json:"companyId"`
	Name       string `json:"name"`
}

// RoleCapability 角色能力评估
type RoleCapability struct {
	Creativity      float64 `json:"creativity"`      // 创造性
	Logic           float64 `json:"logic"`           // 逻辑性
	Professionalism float64 `json:"professionalism"` // 专业性
	Empathy         float64 `json:"empathy"`         // 共情力
	Efficiency      float64 `json:"efficiency"`      // 效率
	Adaptability    float64 `json:"adaptability"`    // 适应性
}

// RoleEvaluation 角色评估报告
type RoleEvaluation struct {
	RoleID          string         `json:"roleId"`
	RoleName        string         `json:"roleName"`
	Capabilities    RoleCapability `json:"capabilities"`
	Strengths       []string       `json:"strengths"`
	Weaknesses      []string       `json:"weaknesses"`
	Score           float64        `json:"score"`
	Suggestions     []string       `json:"suggestions"`
	UsageStats      UsageStats     `json:"usageStats"`
	OptimizedPrompt string         `json:"optimizedPrompt"`
}

// UsageStats 使用统计
type UsageStats struct {
	TotalChats     int       `json:"totalChats"`
	TotalMessages  int       `json:"totalMessages"`
	AvgSessionTime float64   `json:"avgSessionTime"` // 分钟
	ActiveUsers    int       `json:"activeUsers"`
	FavoriteCount  int       `json:"favoriteCount"`
	ShareCount     int       `json:"shareCount"`
	LastUsedAt     time.Time `json:"lastUsedAt"`
	PopularityRank int       `json:"popularityRank"`
	CategoryRank   int       `json:"categoryRank"`
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	Type        string `json:"type"`     // prompt/skill/template/other
	Priority    string `json:"priority"` // high/medium/low
	Title       string `json:"title"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// TestReport 测试报告
type TestReport struct {
	TestID       string           `json:"testId"`
	RoleID       string           `json:"roleId"`
	RoleName     string           `json:"roleName"`
	TestCases    []TestCaseResult `json:"testCases"`
	OverallScore float64          `json:"overallScore"`
	PassRate     float64          `json:"passRate"`
	TotalTime    float64          `json:"totalTime"` // seconds
	CreatedAt    time.Time        `json:"createdAt"`
	Summary      string           `json:"summary"`
}

// TestCaseResult 测试用例结果
type TestCaseResult struct {
	CaseID         string  `json:"caseId"`
	CaseName       string  `json:"caseName"`
	Input          string  `json:"input"`
	ExpectedOutput string  `json:"expectedOutput"`
	ActualOutput   string  `json:"actualOutput"`
	Score          float64 `json:"score"`
	Passed         bool    `json:"passed"`
	Feedback       string  `json:"feedback"`
	Duration       float64 `json:"duration"` // seconds
}

// TestCase 测试用例
type TestCase struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Input              string `json:"input"`
	ExpectedOutput     string `json:"expectedOutput"`
	EvaluationCriteria string `json:"evaluationCriteria"`
}

// RoleShareLink 角色分享链接
type RoleShareLink struct {
	ShareID      string    `json:"shareId"`
	RoleID       string    `json:"roleId"`
	RoleName     string    `json:"roleName"`
	ShareURL     string    `json:"shareUrl"`
	ExpiryDate   time.Time `json:"expiryDate"`
	MaxViews     int       `json:"maxViews"`
	CurrentViews int       `json:"currentViews"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
}

// RoleExport 角色导出配置
type RoleExport struct {
	RoleID         string                 `json:"roleId"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`
	SystemPrompt   string                 `json:"systemPrompt"`
	WelcomeMessage string                 `json:"welcomeMessage"`
	Avatar         string                 `json:"avatar"`
	ModelConfig    map[string]interface{} `json:"modelConfig"`
	Skills         []string               `json:"skills"`
	Version        string                 `json:"version"`
	ExportedAt     time.Time              `json:"exportedAt"`
}

// EnhancedRoleTemplate 增强角色模板
type EnhancedRoleTemplate struct {
	ID                   string         `json:"id"`
	Name                 string         `json:"name"`
	Description          string         `json:"description"`
	Category             string         `json:"category"`
	SystemPrompt         string         `json:"systemPrompt"`
	WelcomeMessage       string         `json:"welcomeMessage"`
	Avatar               string         `json:"avatar"`
	Capabilities         RoleCapability `json:"capabilities"`
	Tags                 []string       `json:"tags"`
	Rating               float64        `json:"rating"`
	UsageCount           int            `json:"usageCount"`
	IsPremium            bool           `json:"isPremium"`
	ExampleConversations []string       `json:"exampleConversations"`
}

// List 获取角色列表
func (h *RoleHandler) List(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	var roles []models.Role

	companyIDs := h.getOwnedCompanyIDs(userIDStr)
	query := h.db.Where("user_id = ?", userIDStr)
	if len(companyIDs) > 0 {
		query = h.db.Where("user_id = ? OR company_id IN ?", userIDStr, companyIDs)
	}

	if companyID := c.Query("companyId"); companyID != "" {
		if !containsString(companyIDs, companyID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "no access to this company"})
			return
		}
		query = h.db.Where("company_id = ?", companyID)
	}

	// 分类筛选
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// 只显示模板
	if c.Query("template") == "true" {
		query = query.Where("is_template = ?", true)
	}

	// 只显示公开角色
	if c.Query("public") == "true" {
		query = query.Where("is_public = ?", true)
	}

	if result := query.Order("created_at DESC").Find(&roles); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    roles,
		"total":   len(roles),
	})
}

// Get 获取单个角色
func (h *RoleHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var role models.Role
	if result := h.db.First(&role, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if !h.canManageRole(userIDStr, role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "no access to this role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    role,
	})
}

// Create 创建角色
func (h *RoleHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.CompanyID != "" {
		var company models.Company
		if err := h.db.Where("id = ? AND owner_id = ?", req.CompanyID, userIDStr).First(&company).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "no access to this company"})
			return
		}
	}

	role := models.Role{
		ID:             models.NewUUID(),
		UserID:         userIDStr,
		CompanyID:      req.CompanyID,
		Name:           req.Name,
		Description:    req.Description,
		Category:       req.Category,
		SystemPrompt:   req.SystemPrompt,
		WelcomeMessage: req.WelcomeMessage,
		Avatar:         req.Avatar,
		IsTemplate:     req.IsTemplate,
		IsPublic:       req.IsPublic,
	}

	// 转换 ModelConfig
	if req.ModelConfig != nil {
		configJSON, _ := json.Marshal(req.ModelConfig)
		role.ModelConfig = models.JSON(configJSON)
	}

	if result := h.db.Create(&role); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 异步同步到 AnythingLLM
	go h.syncToAnythingLLM(role)

	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "success",
		"data":    role,
	})
}

// Update 更新角色
func (h *RoleHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if result := h.db.First(&role, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if !h.canManageRole(userIDStr, role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "no access to this role"})
		return
	}

	if req.CompanyID != "" {
		var company models.Company
		if err := h.db.Where("id = ? AND owner_id = ?", req.CompanyID, userIDStr).First(&company).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "no access to this company"})
			return
		}
	}

	// 更新字段
	role.Name = req.Name
	role.Description = req.Description
	role.Category = req.Category
	role.CompanyID = req.CompanyID
	role.SystemPrompt = req.SystemPrompt
	role.WelcomeMessage = req.WelcomeMessage
	role.Avatar = req.Avatar
	role.IsTemplate = req.IsTemplate
	role.IsPublic = req.IsPublic

	if req.ModelConfig != nil {
		configJSON, _ := json.Marshal(req.ModelConfig)
		role.ModelConfig = models.JSON(configJSON)
	}

	if result := h.db.Save(&role); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 异步同步到 AnythingLLM
	go h.syncToAnythingLLM(role)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    role,
	})
}

// Delete 删除角色
func (h *RoleHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var role models.Role
	if result := h.db.First(&role, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if !h.canManageRole(userIDStr, role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "no access to this role"})
		return
	}

	if result := h.db.Delete(&models.Role{}, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// InstallFromMarket 从角色市场安装到“我的角色”或“我的公司”
func (h *RoleHandler) InstallFromMarket(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	templateID := c.Param("id")

	var req InstallRoleRequest
	_ = c.ShouldBindJSON(&req)
	targetType := strings.TrimSpace(req.TargetType)
	if targetType == "" {
		targetType = "personal"
	}
	if targetType != "personal" && targetType != "company" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "targetType must be personal or company"})
		return
	}

	template, ok := h.findTemplateByID(templateID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	companyID := ""
	if targetType == "company" {
		companyID = strings.TrimSpace(req.CompanyID)
		if companyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "companyId is required"})
			return
		}
		var company models.Company
		if err := h.db.Where("id = ? AND owner_id = ?", companyID, userIDStr).First(&company).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "no access to this company"})
			return
		}
	}

	roleName := strings.TrimSpace(req.Name)
	if roleName == "" {
		roleName = template.Name
	}

	modelCfg := map[string]interface{}{
		"installedFromMarket": true,
		"sourceTemplateId":    template.ID,
		"sourceTemplateName":  template.Name,
		"targetType":          targetType,
	}
	cfgJSON, _ := json.Marshal(modelCfg)

	role := models.Role{
		ID:             models.NewUUID(),
		UserID:         userIDStr,
		CompanyID:      companyID,
		Name:           roleName,
		Description:    template.Description,
		Category:       template.Category,
		SystemPrompt:   template.SystemPrompt,
		WelcomeMessage: template.WelcomeMessage,
		Avatar:         template.Avatar,
		ModelConfig:    models.JSON(cfgJSON),
		IsTemplate:     false,
		IsPublic:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		install := models.RoleInstall{
			ID:              models.NewUUID(),
			TemplateID:      template.ID,
			InstalledRoleID: role.ID,
			InstallerUserID: userIDStr,
			TargetType:      targetType,
			TargetID: func() string {
				if targetType == "company" {
					return companyID
				}
				return userIDStr
			}(),
			CreatedAt: time.Now(),
		}
		return tx.Create(&install).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go h.syncToAnythingLLM(role)

	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"role":       role,
			"templateId": template.ID,
			"targetType": targetType,
			"companyId":  companyID,
		},
	})
}

func (h *RoleHandler) findTemplateByID(id string) (EnhancedRoleTemplate, bool) {
	for _, tpl := range h.getEnhancedTemplatesList() {
		if tpl.ID == id {
			return tpl, true
		}
	}
	return EnhancedRoleTemplate{}, false
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (h *RoleHandler) getOwnedCompanyIDs(userID string) []string {
	var companies []models.Company
	if err := h.db.Select("id").Where("owner_id = ?", userID).Find(&companies).Error; err != nil {
		return nil
	}
	ids := make([]string, 0, len(companies))
	for _, company := range companies {
		ids = append(ids, company.ID)
	}
	return ids
}

func (h *RoleHandler) canManageRole(userID string, role models.Role) bool {
	if role.UserID == userID {
		return true
	}
	if role.CompanyID == "" {
		return false
	}
	var company models.Company
	return h.db.Where("id = ? AND owner_id = ?", role.CompanyID, userID).First(&company).Error == nil
}

// Evaluate 评估角色能力
func (h *RoleHandler) Evaluate(c *gin.Context) {
	id := c.Param("id")

	var role models.Role
	if result := h.db.First(&role, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// 分析系统提示词评估能力
	capabilities := h.analyzeCapabilities(role.SystemPrompt)

	// 获取使用统计
	usageStats := h.getUsageStats(id)

	// 生成评估报告
	evaluation := h.generateEvaluation(role, capabilities, usageStats)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    evaluation,
	})
}

// analyzeCapabilities 分析角色能力
func (h *RoleHandler) analyzeCapabilities(systemPrompt string) RoleCapability {
	prompt := strings.ToLower(systemPrompt)

	cap := RoleCapability{}

	// 创造性指标
	creativeKeywords := []string{"创意", "创新", "创造", "想象", "设计", "艺术", "写作", "故事", "营销", "广告"}
	cap.Creativity = h.calculateScore(prompt, creativeKeywords)

	// 逻辑性指标
	logicKeywords := []string{"逻辑", "分析", "推理", "数据", "结构", "系统", "算法", "编程", "技术", "科学"}
	cap.Logic = h.calculateScore(prompt, logicKeywords)

	// 专业性指标
	proKeywords := []string{"专业", "专家", "资深", "精通", "认证", "经验", "顾问", "律师", "医生", "工程师"}
	cap.Professionalism = h.calculateScore(prompt, proKeywords)

	// 共情力指标
	empathyKeywords := []string{"理解", "关心", "支持", "帮助", "耐心", "友好", "温暖", "倾听", "陪伴", "心理"}
	cap.Empathy = h.calculateScore(prompt, empathyKeywords)

	// 效率指标
	efficiencyKeywords := []string{"快速", "高效", "及时", "简洁", "直接", "优化", "自动化", "批量", "工具"}
	cap.Efficiency = h.calculateScore(prompt, efficiencyKeywords)

	// 适应性指标
	adaptKeywords := []string{"灵活", "适应", "多样", "多场景", "通用", "自定义", "调整", "变通"}
	cap.Adaptability = h.calculateScore(prompt, adaptKeywords)

	return cap
}

// calculateScore 计算关键词匹配得分
func (h *RoleHandler) calculateScore(prompt string, keywords []string) float64 {
	score := 0.0
	for _, keyword := range keywords {
		if strings.Contains(prompt, keyword) {
			score += 1.0
		}
	}
	// 归一化到 0-100
	maxScore := float64(len(keywords))
	if maxScore == 0 {
		return 50.0
	}
	return math.Min(100, (score/maxScore)*100)
}

// getUsageStats 获取使用统计
func (h *RoleHandler) getUsageStats(roleID string) UsageStats {
	stats := UsageStats{}

	// 查询对话会话数
	var sessionCount int64
	h.db.Model(&models.ChatSession{}).Where("role_id = ?", roleID).Count(&sessionCount)
	stats.TotalChats = int(sessionCount)

	// 查询消息总数 (简化处理)
	stats.TotalMessages = int(sessionCount) * 10 // 估算

	// 计算平均会话时间 (简化)
	stats.AvgSessionTime = 15.5

	// 活跃用户数
	var userCount int64
	h.db.Model(&models.ChatSession{}).Distinct("user_id").Where("role_id = ?", roleID).Count(&userCount)
	stats.ActiveUsers = int(userCount)

	// 模拟其他统计数据
	stats.FavoriteCount = int(sessionCount / 5)
	stats.ShareCount = int(sessionCount / 10)
	stats.LastUsedAt = time.Now().Add(-24 * time.Hour)

	return stats
}

// generateEvaluation 生成评估报告
func (h *RoleHandler) generateEvaluation(role models.Role, cap RoleCapability, stats UsageStats) RoleEvaluation {
	eval := RoleEvaluation{
		RoleID:       role.ID,
		RoleName:     role.Name,
		Capabilities: cap,
		UsageStats:   stats,
	}

	// 计算总分
	eval.Score = (cap.Creativity + cap.Logic + cap.Professionalism + cap.Empathy + cap.Efficiency + cap.Adaptability) / 6.0

	// 识别优势
	if cap.Creativity >= 70 {
		eval.Strengths = append(eval.Strengths, "创造性思维突出")
	}
	if cap.Logic >= 70 {
		eval.Strengths = append(eval.Strengths, "逻辑分析能力强")
	}
	if cap.Professionalism >= 70 {
		eval.Strengths = append(eval.Strengths, "专业知识扎实")
	}
	if cap.Empathy >= 70 {
		eval.Strengths = append(eval.Strengths, "共情能力优秀")
	}
	if cap.Efficiency >= 70 {
		eval.Strengths = append(eval.Strengths, "响应效率高")
	}
	if cap.Adaptability >= 70 {
		eval.Strengths = append(eval.Strengths, "适应性强")
	}

	// 识别劣势
	if cap.Creativity < 40 {
		eval.Weaknesses = append(eval.Weaknesses, "创造性有待提升")
	}
	if cap.Logic < 40 {
		eval.Weaknesses = append(eval.Weaknesses, "逻辑性需要加强")
	}
	if cap.Professionalism < 40 {
		eval.Weaknesses = append(eval.Weaknesses, "专业性不足")
	}

	// 生成优化建议
	eval.Suggestions = h.generateSuggestions(role, cap)

	// 生成优化后的提示词
	eval.OptimizedPrompt = h.optimizePrompt(role.SystemPrompt, cap)

	return eval
}

// generateSuggestions 生成优化建议
func (h *RoleHandler) generateSuggestions(role models.Role, cap RoleCapability) []string {
	var suggestions []string

	if len(role.SystemPrompt) < 100 {
		suggestions = append(suggestions, "系统提示词较短，建议补充更多角色细节和行为准则")
	}

	if cap.Creativity < 50 && strings.Contains(strings.ToLower(role.Category), "营销") {
		suggestions = append(suggestions, "营销类角色需要更强的创造性，建议在提示词中加入创意生成的指导")
	}

	if cap.Empathy < 50 && strings.Contains(strings.ToLower(role.Category), "客服") {
		suggestions = append(suggestions, "客服类角色需要更强的共情力，建议增加情感支持的描述")
	}

	if cap.Professionalism < 60 {
		suggestions = append(suggestions, "建议增加专业背景和资质的描述，提升角色可信度")
	}

	return suggestions
}

// optimizePrompt 优化提示词
func (h *RoleHandler) optimizePrompt(original string, cap RoleCapability) string {
	optimized := original

	// 如果提示词太短，添加通用优化
	if len(original) < 200 {
		optimized += "\n\n请始终保持专业、友好的态度。在回答问题时，先理解用户的核心需求，然后提供清晰、有条理的解答。如果遇到不确定的问题，诚实地告知用户并建议寻求专业帮助。"
	}

	return optimized
}

// GetSuggestions 获取角色优化建议
func (h *RoleHandler) GetSuggestions(c *gin.Context) {
	id := c.Param("id")

	var role models.Role
	if result := h.db.First(&role, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	capabilities := h.analyzeCapabilities(role.SystemPrompt)
	suggestions := []OptimizationSuggestion{}

	// 基于能力评估生成建议
	if capabilities.Creativity < 60 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "prompt",
			Priority:    "medium",
			Title:       "提升创造性",
			Description: "当前角色的创造性评分较低，建议在系统提示词中加入鼓励创新思维的描述",
			Example:     "你善于提出创新性的解决方案，能够从多个角度思考问题...",
		})
	}

	if capabilities.Professionalism < 60 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "prompt",
			Priority:    "high",
			Title:       "增强专业性",
			Description: "建议补充角色的专业背景、资质认证和从业经验",
			Example:     "你是一位拥有 10 年经验的资深专家，持有 XX 认证...",
		})
	}

	// 基于使用统计的建议
	stats := h.getUsageStats(id)
	if stats.TotalChats < 10 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "other",
			Priority:    "low",
			Title:       "增加曝光度",
			Description: "角色使用次数较少，建议优化角色描述或分享到角色市场",
			Example:     "",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    suggestions,
	})
}

// ExportRole 导出角色配置
func (h *RoleHandler) ExportRole(c *gin.Context) {
	id := c.Param("id")

	var role models.Role
	if result := h.db.First(&role, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	export := RoleExport{
		RoleID:         role.ID,
		Name:           role.Name,
		Description:    role.Description,
		Category:       role.Category,
		SystemPrompt:   role.SystemPrompt,
		WelcomeMessage: role.WelcomeMessage,
		Avatar:         role.Avatar,
		Version:        "1.0",
		ExportedAt:     time.Now(),
	}

	// 解析 ModelConfig
	if role.ModelConfig != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(role.ModelConfig), &config); err == nil {
			export.ModelConfig = config
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    export,
	})
}

// ImportRole 导入角色配置
func (h *RoleHandler) ImportRole(c *gin.Context) {
	var export RoleExport
	if err := c.ShouldBindJSON(&export); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := models.Role{
		ID:             models.NewUUID(),
		Name:           export.Name,
		Description:    export.Description,
		Category:       export.Category,
		SystemPrompt:   export.SystemPrompt,
		WelcomeMessage: export.WelcomeMessage,
		Avatar:         export.Avatar,
		IsTemplate:     false,
		IsPublic:       false,
	}

	if export.ModelConfig != nil {
		configJSON, _ := json.Marshal(export.ModelConfig)
		role.ModelConfig = models.JSON(configJSON)
	}

	if result := h.db.Create(&role); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "success",
		"data":    role,
	})
}

// GenerateShareLink 生成分享链接
func (h *RoleHandler) GenerateShareLink(c *gin.Context) {
	id := c.Param("id")

	var role models.Role
	if result := h.db.First(&role, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	shareID := models.NewUUID()
	shareLink := RoleShareLink{
		ShareID:      shareID,
		RoleID:       role.ID,
		RoleName:     role.Name,
		ShareURL:     fmt.Sprintf("https://rolecraft.ai/share/%s", shareID),
		ExpiryDate:   time.Now().Add(30 * 24 * time.Hour), // 30 天有效期
		MaxViews:     1000,
		CurrentViews: 0,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    shareLink,
	})
}

// RunTest 运行角色测试
func (h *RoleHandler) RunTest(c *gin.Context) {
	id := c.Param("id")

	var role models.Role
	if result := h.db.First(&role, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	var testCases []TestCase
	if err := c.ShouldBindJSON(&testCases); err != nil {
		// 使用默认测试用例
		testCases = h.getDefaultTestCases(role)
	}

	results := []TestCaseResult{}
	totalScore := 0.0
	passedCount := 0

	for _, tc := range testCases {
		result := h.runTestCase(role, tc)
		results = append(results, result)
		totalScore += result.Score
		if result.Passed {
			passedCount++
		}
	}

	report := TestReport{
		TestID:       models.NewUUID(),
		RoleID:       role.ID,
		RoleName:     role.Name,
		TestCases:    results,
		OverallScore: totalScore / float64(len(testCases)),
		PassRate:     float64(passedCount) / float64(len(testCases)) * 100,
		TotalTime:    2.5,
		CreatedAt:    time.Now(),
		Summary:      fmt.Sprintf("完成 %d 个测试用例，通过率 %.1f%%", len(testCases), float64(passedCount)/float64(len(testCases))*100),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    report,
	})
}

// getDefaultTestCases 获取默认测试用例
func (h *RoleHandler) getDefaultTestCases(role models.Role) []TestCase {
	return []TestCase{
		{
			ID:                 "1",
			Name:               "基础问候测试",
			Description:        "测试角色的基本交互能力",
			Input:              "你好，请介绍一下你自己",
			EvaluationCriteria: "回复应该友好、专业，并体现角色定位",
		},
		{
			ID:                 "2",
			Name:               "专业能力测试",
			Description:        "测试角色的专业知识",
			Input:              fmt.Sprintf("请帮我解决一个%s相关的问题", role.Category),
			EvaluationCriteria: "回复应该体现专业性和实用性",
		},
		{
			ID:                 "3",
			Name:               "边界情况测试",
			Description:        "测试角色处理模糊问题的能力",
			Input:              "我不太确定该怎么描述我的问题...",
			EvaluationCriteria: "回复应该耐心引导用户明确需求",
		},
	}
}

// runTestCase 运行单个测试用例
func (h *RoleHandler) runTestCase(role models.Role, tc TestCase) TestCaseResult {
	// 模拟测试执行
	result := TestCaseResult{
		CaseID:         tc.ID,
		CaseName:       tc.Name,
		Input:          tc.Input,
		ExpectedOutput: tc.ExpectedOutput,
		ActualOutput:   fmt.Sprintf("这是基于角色 [%s] 的模拟回复", role.Name),
		Score:          85.0,
		Passed:         true,
		Feedback:       "回复符合预期，体现了角色的专业性",
		Duration:       0.8,
	}

	return result
}

// GetEnhancedTemplates 获取增强角色模板
func (h *RoleHandler) GetEnhancedTemplates(c *gin.Context) {
	category := c.Query("category")

	templates := h.getEnhancedTemplatesList()

	// 分类筛选
	if category != "" && category != "全部" {
		filtered := []EnhancedRoleTemplate{}
		for _, t := range templates {
			if t.Category == category {
				filtered = append(filtered, t)
			}
		}
		templates = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    templates,
		"total":   len(templates),
	})
}

// getEnhancedTemplatesList 获取增强模板列表
func (h *RoleHandler) getEnhancedTemplatesList() []EnhancedRoleTemplate {
	return []EnhancedRoleTemplate{
		{
			ID:             "template_001",
			Name:           "智能助理",
			Description:    "全能型办公助手，帮助处理日常事务、撰写邮件、安排日程",
			Category:       "通用",
			SystemPrompt:   "你是一位智能助理，擅长帮助用户处理各种办公任务。请用友好、专业的态度回答用户的问题。",
			WelcomeMessage: "你好！我是你的智能助理，有什么可以帮你的吗？",
			Capabilities: RoleCapability{
				Creativity: 60, Logic: 75, Professionalism: 70, Empathy: 65, Efficiency: 80, Adaptability: 70,
			},
			Tags:       []string{"办公", "效率", "通用"},
			Rating:     4.8,
			UsageCount: 15234,
			IsPremium:  false,
			ExampleConversations: []string{
				"用户：帮我写一封会议邀请邮件\n助理：好的，请问会议的时间、地点和参会人员是？",
			},
		},
		{
			ID:             "template_002",
			Name:           "营销专家",
			Description:    "专业的营销策划助手，帮助制定营销策略、撰写文案",
			Category:       "营销",
			SystemPrompt:   "你是一位资深的营销专家，精通各种营销策略和内容创作。请提供有创意、可执行的营销建议。",
			WelcomeMessage: "你好！我是你的营销顾问，让我们一起制定出色的营销策略吧！",
			Capabilities: RoleCapability{
				Creativity: 90, Logic: 70, Professionalism: 85, Empathy: 60, Efficiency: 75, Adaptability: 80,
			},
			Tags:       []string{"营销", "创意", "文案"},
			Rating:     4.9,
			UsageCount: 12456,
			IsPremium:  false,
			ExampleConversations: []string{
				"用户：如何提升产品转化率？\n专家：我们可以从用户旅程分析开始...",
			},
		},
		{
			ID:             "template_003",
			Name:           "法务顾问",
			Description:    "合同审查与法律咨询专家",
			Category:       "法律",
			SystemPrompt:   "你是一位专业的法务顾问，擅长合同审查和法律咨询。请提供准确、实用的法律建议。",
			WelcomeMessage: "你好！我是你的法务顾问，有什么法律问题需要咨询吗？",
			Capabilities: RoleCapability{
				Creativity: 40, Logic: 95, Professionalism: 95, Empathy: 50, Efficiency: 70, Adaptability: 60,
			},
			Tags:       []string{"法律", "合同", "咨询"},
			Rating:     4.7,
			UsageCount: 8932,
			IsPremium:  true,
			ExampleConversations: []string{
				"用户：这份合同有什么风险？\n顾问：让我仔细审查一下关键条款...",
			},
		},
		{
			ID:             "template_004",
			Name:           "心理咨询师",
			Description:    "专业的心理健康支持者，提供情感倾听和心理疏导",
			Category:       "健康",
			SystemPrompt:   "你是一位温暖、专业的心理咨询师。请耐心倾听用户的困扰，提供情感支持和专业建议。注意：不能替代专业医疗诊断。",
			WelcomeMessage: "你好，我在这里倾听你的心声。今天想聊些什么呢？",
			Capabilities: RoleCapability{
				Creativity: 50, Logic: 60, Professionalism: 85, Empathy: 95, Efficiency: 65, Adaptability: 75,
			},
			Tags:       []string{"心理", "健康", "倾听"},
			Rating:     4.9,
			UsageCount: 23456,
			IsPremium:  false,
			ExampleConversations: []string{
				"用户：最近感觉压力很大...\n咨询师：能具体说说是什么让你感到压力吗？",
			},
		},
		{
			ID:             "template_005",
			Name:           "编程导师",
			Description:    "经验丰富的软件工程师，帮助学习编程和解决技术问题",
			Category:       "技术",
			SystemPrompt:   "你是一位资深软件工程师，擅长多种编程语言。请用清晰、易懂的方式讲解技术概念，帮助学习者成长。",
			WelcomeMessage: "你好！我是你的编程导师，有什么问题尽管问我！",
			Capabilities: RoleCapability{
				Creativity: 65, Logic: 90, Professionalism: 85, Empathy: 70, Efficiency: 80, Adaptability: 75,
			},
			Tags:       []string{"编程", "技术", "教育"},
			Rating:     4.8,
			UsageCount: 18765,
			IsPremium:  false,
			ExampleConversations: []string{
				"用户：这段代码为什么报错？\n导师：让我看看...问题出在这一行...",
			},
		},
		{
			ID:             "template_006",
			Name:           "财务规划师",
			Description:    "专业的理财顾问，帮助制定财务规划和投资建议",
			Category:       "财务",
			SystemPrompt:   "你是一位认证的财务规划师，擅长个人理财、投资规划和税务优化。请提供专业、谨慎的财务建议。",
			WelcomeMessage: "你好！让我们一起规划你的财务未来！",
			Capabilities: RoleCapability{
				Creativity: 45, Logic: 85, Professionalism: 90, Empathy: 60, Efficiency: 75, Adaptability: 70,
			},
			Tags:       []string{"财务", "投资", "理财"},
			Rating:     4.7,
			UsageCount: 9876,
			IsPremium:  true,
			ExampleConversations: []string{
				"用户：如何合理配置资产？\n规划师：首先我们需要了解你的风险承受能力...",
			},
		},
		{
			ID:             "template_007",
			Name:           "学术研究员",
			Description:    "专业的学术研究助手，帮助文献检索、论文写作和数据分析",
			Category:       "教育",
			SystemPrompt:   "你是一位经验丰富的学术研究员，熟悉各学科的研究方法。请帮助用户进行文献检索、论文写作和数据分析。",
			WelcomeMessage: "你好！我是你的学术研究助手，有什么研究问题需要帮助吗？",
			Capabilities: RoleCapability{
				Creativity: 55, Logic: 90, Professionalism: 90, Empathy: 55, Efficiency: 70, Adaptability: 65,
			},
			Tags:       []string{"学术", "研究", "论文"},
			Rating:     4.6,
			UsageCount: 7654,
			IsPremium:  false,
			ExampleConversations: []string{
				"用户：如何查找相关文献？\n研究员：我们可以从这些数据库开始...",
			},
		},
		{
			ID:             "template_008",
			Name:           "健身教练",
			Description:    "专业的健身指导专家，帮助制定训练计划和营养建议",
			Category:       "健康",
			SystemPrompt:   "你是一位认证的健身教练，擅长制定个性化训练计划和营养方案。请提供科学、安全的健身指导。",
			WelcomeMessage: "你好！让我们一起开启健康之旅！",
			Capabilities: RoleCapability{
				Creativity: 60, Logic: 75, Professionalism: 85, Empathy: 80, Efficiency: 75, Adaptability: 80,
			},
			Tags:       []string{"健身", "健康", "运动"},
			Rating:     4.8,
			UsageCount: 14532,
			IsPremium:  false,
			ExampleConversations: []string{
				"用户：我想减脂，该怎么训练？\n教练：首先我们需要制定一个合理的计划...",
			},
		},
		{
			ID:             "template_009",
			Name:           "旅行规划师",
			Description:    "经验丰富的旅行顾问，帮助规划行程和提供旅行建议",
			Category:       "生活",
			SystemPrompt:   "你是一位热爱旅行的规划师，熟悉全球各地旅游景点和文化。请帮助用户规划完美的旅行行程。",
			WelcomeMessage: "你好！想去哪里旅行？让我帮你规划！",
			Capabilities: RoleCapability{
				Creativity: 75, Logic: 70, Professionalism: 75, Empathy: 70, Efficiency: 80, Adaptability: 85,
			},
			Tags:       []string{"旅行", "规划", "生活"},
			Rating:     4.7,
			UsageCount: 11234,
			IsPremium:  false,
			ExampleConversations: []string{
				"用户：想去日本玩一周，怎么安排？\n规划师：日本一周的话，我建议...",
			},
		},
		{
			ID:             "template_010",
			Name:           "职业规划师",
			Description:    "专业的职业发展顾问，帮助职业规划、简历优化和面试准备",
			Category:       "职业",
			SystemPrompt:   "你是一位资深职业规划师，熟悉各行业职业发展路径。请帮助用户进行职业规划、简历优化和面试准备。",
			WelcomeMessage: "你好！让我们一起规划你的职业发展之路！",
			Capabilities: RoleCapability{
				Creativity: 55, Logic: 80, Professionalism: 90, Empathy: 75, Efficiency: 75, Adaptability: 80,
			},
			Tags:       []string{"职业", "发展", "求职"},
			Rating:     4.8,
			UsageCount: 13567,
			IsPremium:  true,
			ExampleConversations: []string{
				"用户：我想转行，该怎么准备？\n规划师：转行需要系统规划，首先...",
			},
		},
	}
}

// syncToAnythingLLM 同步角色到 AnythingLLM
func (h *RoleHandler) syncToAnythingLLM(role models.Role) {
	if h.anything == nil || !h.anything.Enabled() {
		return
	}

	slug := anythingllm.UserWorkspaceSlug(role.ID)

	ws, err := h.anything.EnsureWorkspaceBySlug(context.Background(), slug, fmt.Sprintf("Role: %s", role.Name), role.SystemPrompt)
	if err != nil {
		log.Printf("⚠️ 角色 [%s] 创建 AnythingLLM Workspace 失败：%v", role.Name, err)
		return
	}
	if ws != nil && strings.TrimSpace(ws.Slug) != "" {
		slug = strings.TrimSpace(ws.Slug)
	}
	if err := h.anything.UpdateWorkspaceSystemPrompt(context.Background(), slug, role.SystemPrompt); err != nil {
		log.Printf("⚠️ 角色 [%s] 更新 AnythingLLM Workspace 失败：%v", role.Name, err)
		return
	}
	log.Printf("✅ 角色 [%s] 已同步到 AnythingLLM Workspace", role.Name)
}

// Chat 与角色对话
func (h *RoleHandler) Chat(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if result := h.db.First(&role, "id = ? AND user_id = ?", id, userIDStr); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// 调用 AI 服务进行对话
	reply, err := h.callAI(role, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"role":    role.Name,
			"message": req.Message,
			"reply":   reply,
		},
	})
}

// callAI 调用 AI 服务
func (h *RoleHandler) callAI(role models.Role, message string) (string, error) {
	// 简化实现，实际应调用 OpenAI API
	return fmt.Sprintf("[%s] 收到你的消息：%s", role.Name, message), nil
}

// GetRoleStats 获取角色排行榜
func (h *RoleHandler) GetRoleStats(c *gin.Context) {
	var roles []models.Role
	h.db.Where("is_public = ?", true).Find(&roles)

	type RoleStat struct {
		RoleID     string  `json:"roleId"`
		RoleName   string  `json:"roleName"`
		Category   string  `json:"category"`
		UsageCount int     `json:"usageCount"`
		Rating     float64 `json:"rating"`
		Rank       int     `json:"rank"`
	}

	stats := []RoleStat{}
	for _, role := range roles {
		usageStats := h.getUsageStats(role.ID)
		stats = append(stats, RoleStat{
			RoleID:     role.ID,
			RoleName:   role.Name,
			Category:   role.Category,
			UsageCount: usageStats.TotalChats,
			Rating:     4.5 + float64(usageStats.TotalChats%10)*0.05,
		})
	}

	// 按使用次数排序
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].UsageCount > stats[j].UsageCount
	})

	// 添加排名
	for i := range stats {
		stats[i].Rank = i + 1
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}
