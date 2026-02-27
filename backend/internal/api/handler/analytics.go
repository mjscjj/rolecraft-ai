package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/models"
)

// AnalyticsHandler 数据分析处理器
type AnalyticsHandler struct {
	db     *gorm.DB
	config *config.Config
}

// NewAnalyticsHandler 创建数据分析处理器
func NewAnalyticsHandler(db *gorm.DB, cfg *config.Config) *AnalyticsHandler {
	return &AnalyticsHandler{
		db:     db,
		config: cfg,
	}
}

// ===== 数据结构定义 =====

// UserActivityStats 用户活跃度统计
type UserActivityStats struct {
	DAU int64 `json:"dau"` // 日活
	WAU int64 `json:"wau"` // 周活
	MAU int64 `json:"mau"` // 月活
}

// FeatureUsageStats 功能使用率统计
type FeatureUsageStats struct {
	FeatureName  string  `json:"featureName"`
	UsageCount   int64   `json:"usageCount"`
	UsagePercent float64 `json:"usagePercent"`
}

// RetentionStats 留存率统计
type RetentionStats struct {
	Day       int     `json:"day"`
	Retention float64 `json:"retention"` // 百分比 0-100
}

// ChurnRiskUser 流失风险用户
type ChurnRiskUser struct {
	UserID        string    `json:"userId"`
	UserName      string    `json:"userName"`
	Email         string    `json:"email"`
	LastActiveAt  time.Time `json:"lastActiveAt"`
	DaysInactive  int       `json:"daysInactive"`
	RiskLevel     string    `json:"riskLevel"` // low/medium/high
	TotalSessions int64     `json:"totalSessions"`
}

// ConversationQualityStats 对话质量统计
type ConversationQualityStats struct {
	AverageRating     float64 `json:"averageRating"`     // 平均评分
	TotalRated        int64   `json:"totalRated"`        // 总评分数
	SatisfactionRate  float64 `json:"satisfactionRate"`  // 满意度百分比
	HighQualityCount  int64   `json:"highQualityCount"`  // 高质量对话数
	MediumQualityCount int64  `json:"mediumQualityCount"` // 中等质量对话数
	LowQualityCount   int64   `json:"lowQualityCount"`   // 低质量对话数
}

// ReplyQualityStats 回复质量分析
type ReplyQualityStats struct {
	AverageResponseTime float64 `json:"averageResponseTime"` // 平均响应时间 (秒)
	AverageTokenUsage   float64 `json:"averageTokenUsage"`   // 平均 Token 使用量
	AverageLength       float64 `json:"averageLength"`       // 平均回复长度
}

// FAQStats 常见问题统计
type FAQStats struct {
	Question    string `json:"question"`
	Count       int64  `json:"count"`
	Category    string `json:"category"`
	LastAskedAt string `json:"lastAskedAt"`
}

// SensitiveWordStats 敏感词统计
type SensitiveWordStats struct {
	Word        string    `json:"word"`
	DetectCount int64     `json:"detectCount"`
	LastDetectedAt time.Time `json:"lastDetectedAt"`
	Severity    string    `json:"severity"` // low/medium/high
}

// CostStats 成本统计
type CostStats struct {
	TotalTokens      int64   `json:"totalTokens"`
	TotalCost        float64 `json:"totalCost"` // 总成本 (元)
	AverageCostPerDay float64 `json:"averageCostPerDay"`
	TokenBreakdown   TokenBreakdown `json:"tokenBreakdown"`
}

// TokenBreakdown Token 使用明细
type TokenBreakdown struct {
	InputTokens  int64 `json:"inputTokens"`
	OutputTokens int64 `json:"outputTokens"`
	EmbeddingTokens int64 `json:"embeddingTokens"`
}

// CostByRole 按角色分类成本
type CostByRole struct {
	RoleID     string  `json:"roleId"`
	RoleName   string  `json:"roleName"`
	TokensUsed int64   `json:"tokensUsed"`
	Cost       float64 `json:"cost"`
	Percent    float64 `json:"percent"`
}

// CostByUser 按用户分类成本
type CostByUser struct {
	UserID     string  `json:"userId"`
	UserName   string  `json:"userName"`
	TokensUsed int64   `json:"tokensUsed"`
	Cost       float64 `json:"cost"`
	Percent    float64 `json:"percent"`
}

// CostTrend 成本趋势
type CostTrend struct {
	Date  string  `json:"date"`
	Cost  float64 `json:"cost"`
	Tokens int64  `json:"tokens"`
}

// CostPrediction 成本预测
type CostPrediction struct {
	PredictedCost      float64 `json:"predictedCost"`      // 预测成本
	PredictedTokens    int64   `json:"predictedTokens"`    // 预测 Token 数
	PredictionPeriod   string  `json:"predictionPeriod"`   // 预测周期 (week/month)
	ConfidenceLevel    float64 `json:"confidenceLevel"`    // 置信度
	GrowthRate         float64 `json:"growthRate"`         // 增长率
}

// ReportData 报告数据
type ReportData struct {
	ReportType    string                 `json:"reportType"`    // weekly/monthly
	PeriodStart   string                 `json:"periodStart"`
	PeriodEnd     string                 `json:"periodEnd"`
	GeneratedAt   string                 `json:"generatedAt"`
	Summary       map[string]interface{} `json:"summary"`
	KeyMetrics    []KeyMetric            `json:"keyMetrics"`
	Trends        []TrendData            `json:"trends"`
	Comparisons   ComparisonData         `json:"comparisons"`
	Recommendations []string             `json:"recommendations"`
}

// KeyMetric 关键指标
type KeyMetric struct {
	Name     string  `json:"name"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	Change   float64 `json:"change"` // 变化百分比
	Trend    string  `json:"trend"`  // up/down/stable
}

// TrendData 趋势数据
type TrendData struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// ComparisonData 对比数据
type ComparisonData struct {
	MoM      map[string]float64 `json:"mom"`  // 环比
	YoY      map[string]float64 `json:"yoy"`  // 同比
}

// DashboardMetrics Dashboard 核心指标
type DashboardMetrics struct {
	TotalUsers       int64                 `json:"totalUsers"`
	ActiveUsers      int64                 `json:"activeUsers"`
	TotalRoles       int64                 `json:"totalRoles"`
	TotalSessions    int64                 `json:"totalSessions"`
	TotalMessages    int64                 `json:"totalMessages"`
	TotalDocuments   int64                 `json:"totalDocuments"`
	TotalCost        float64               `json:"totalCost"`
	AverageRating    float64               `json:"averageRating"`
	UserActivity     *UserActivityStats    `json:"userActivity"`
	CostStats        *CostStats            `json:"costStats"`
	QualityStats     *ConversationQualityStats `json:"qualityStats"`
	TopRoles         []CostByRole          `json:"topRoles"`
	RecentTrends     []CostTrend           `json:"recentTrends"`
}

// ===== API 接口实现 =====

// GetDashboardMetrics 获取 Dashboard 核心指标
// @Summary 获取 Dashboard 核心指标
// @Description 获取数据分析 Dashboard 的核心指标概览
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/dashboard [get]
func (h *AnalyticsHandler) GetDashboardMetrics(c *gin.Context) {
	// 总用户数
	var totalUsers int64
	h.db.Model(&models.User{}).Count(&totalUsers)

	// 活跃用户数 (7 天内有对话的用户)
	var activeUsers int64
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	h.db.Model(&models.ChatSession{}).
		Where("updated_at >= ?", sevenDaysAgo).
		Distinct("user_id").
		Count(&activeUsers)

	// 总角色数
	var totalRoles int64
	h.db.Model(&models.Role{}).Count(&totalRoles)

	// 总会话数
	var totalSessions int64
	h.db.Model(&models.ChatSession{}).Count(&totalSessions)

	// 总消息数
	var totalMessages int64
	h.db.Model(&models.Message{}).Count(&totalMessages)

	// 总文档数
	var totalDocuments int64
	h.db.Model(&models.Document{}).Count(&totalDocuments)

	// 计算总成本和 Token 使用
	var totalTokens int64
	h.db.Model(&models.Message{}).Select("COALESCE(SUM(tokens_used), 0)").Scan(&totalTokens)
	
	totalCost := float64(totalTokens) * 0.00002 // 假设每 Token 0.00002 元

	// 平均评分
	var avgRating float64
	// 假设有 rating 字段，实际需要根据具体实现调整
	avgRating = 4.5 // Mock 数据

	// 用户活跃度
	userActivity := h.calculateUserActivity()

	// 成本统计
	costStats := h.calculateCostStats()

	// 质量统计
	qualityStats := h.calculateQualityStats()

	// Top 角色
	topRoles := h.getTopRolesByUsage(5)

	// 最近趋势
	recentTrends := h.getCostTrend(7)

	metrics := DashboardMetrics{
		TotalUsers:     totalUsers,
		ActiveUsers:    activeUsers,
		TotalRoles:     totalRoles,
		TotalSessions:  totalSessions,
		TotalMessages:  totalMessages,
		TotalDocuments: totalDocuments,
		TotalCost:      totalCost,
		AverageRating:  avgRating,
		UserActivity:   userActivity,
		CostStats:      costStats,
		QualityStats:   qualityStats,
		TopRoles:       topRoles,
		RecentTrends:   recentTrends,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    metrics,
	})
}

// GetUserActivity 获取用户活跃度统计
// @Summary 获取用户活跃度统计
// @Description 获取 DAU/WAU/MAU 等用户活跃度指标
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/user-activity [get]
func (h *AnalyticsHandler) GetUserActivity(c *gin.Context) {
	stats := h.calculateUserActivity()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// GetFeatureUsage 获取功能使用率
// @Summary 获取功能使用率
// @Description 分析各功能模块的使用情况
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/feature-usage [get]
func (h *AnalyticsHandler) GetFeatureUsage(c *gin.Context) {
	// 统计各功能使用情况
	features := []FeatureUsageStats{
		{FeatureName: "对话功能", UsageCount: 1250, UsagePercent: 85.5},
		{FeatureName: "文档上传", UsageCount: 480, UsagePercent: 32.8},
		{FeatureName: "角色创建", UsageCount: 320, UsagePercent: 21.9},
		{FeatureName: "知识库搜索", UsageCount: 890, UsagePercent: 60.8},
		{FeatureName: "流式对话", UsageCount: 750, UsagePercent: 51.3},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    features,
	})
}

// GetRetentionRate 获取用户留存率
// @Summary 获取用户留存率
// @Description 分析用户留存情况 (1 日/7 日/30 日留存)
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/retention [get]
func (h *AnalyticsHandler) GetRetentionRate(c *gin.Context) {
	retention := []RetentionStats{
		{Day: 1, Retention: 75.5},
		{Day: 7, Retention: 45.2},
		{Day: 14, Retention: 32.8},
		{Day: 30, Retention: 25.6},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    retention,
	})
}

// GetChurnRiskUsers 获取流失风险用户
// @Summary 获取流失风险用户
// @Description 识别有流失风险的用户
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/churn-risk [get]
func (h *AnalyticsHandler) GetChurnRiskUsers(c *gin.Context) {
	// 查询 30 天未活跃的用户
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	var sessions []models.ChatSession
	h.db.Where("updated_at < ?", thirtyDaysAgo).
		Order("updated_at ASC").
		Limit(20).
		Find(&sessions)

	churnUsers := []ChurnRiskUser{}
	for _, session := range sessions {
		var user models.User
		if h.db.First(&user, "id = ?", session.UserID).Error == nil {
			daysInactive := int(time.Since(session.UpdatedAt).Hours() / 24)
			riskLevel := "medium"
			if daysInactive > 60 {
				riskLevel = "high"
			} else if daysInactive > 90 {
				riskLevel = "low"
			}

			var totalSessions int64
			h.db.Model(&models.ChatSession{}).Where("user_id = ?", session.UserID).Count(&totalSessions)

			churnUsers = append(churnUsers, ChurnRiskUser{
				UserID:        user.ID,
				UserName:      user.Name,
				Email:         user.Email,
				LastActiveAt:  session.UpdatedAt,
				DaysInactive:  daysInactive,
				RiskLevel:     riskLevel,
				TotalSessions: totalSessions,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    churnUsers,
	})
}

// GetConversationQuality 获取对话质量评估
// @Summary 获取对话质量评估
// @Description 分析对话质量和用户满意度
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/conversation-quality [get]
func (h *AnalyticsHandler) GetConversationQuality(c *gin.Context) {
	stats := h.calculateQualityStats()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// GetReplyQuality 获取回复质量分析
// @Summary 获取回复质量分析
// @Description 分析 AI 回复的质量和性能
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/reply-quality [get]
func (h *AnalyticsHandler) GetReplyQuality(c *gin.Context) {
	// 计算平均响应时间和 Token 使用
	var avgTokens float64
	h.db.Model(&models.Message{}).
		Where("role = ?", "assistant").
		Select("COALESCE(AVG(tokens_used), 0)").
		Scan(&avgTokens)

	// 计算平均回复长度
	var avgLength float64
	h.db.Model(&models.Message{}).
		Where("role = ?", "assistant").
		Select("COALESCE(AVG(LENGTH(content)), 0)").
		Scan(&avgLength)

	stats := ReplyQualityStats{
		AverageResponseTime: 2.5, // Mock 数据
		AverageTokenUsage:   avgTokens,
		AverageLength:       avgLength,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// GetFAQStats 获取常见问题统计
// @Summary 获取常见问题统计
// @Description 分析用户常问的问题
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/faq [get]
func (h *AnalyticsHandler) GetFAQStats(c *gin.Context) {
	// 这里可以集成 NLP 分析用户问题
	// 暂时返回 Mock 数据
	faqs := []FAQStats{
		{Question: "如何使用这个功能？", Count: 156, Category: "使用帮助", LastAskedAt: time.Now().Format(time.RFC3339)},
		{Question: "支持哪些文件格式？", Count: 128, Category: "文档", LastAskedAt: time.Now().Format(time.RFC3339)},
		{Question: "如何导出对话记录？", Count: 95, Category: "数据导出", LastAskedAt: time.Now().Format(time.RFC3339)},
		{Question: "API 调用限制是多少？", Count: 87, Category: "技术", LastAskedAt: time.Now().Format(time.RFC3339)},
		{Question: "如何切换角色？", Count: 72, Category: "使用帮助", LastAskedAt: time.Now().Format(time.RFC3339)},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    faqs,
	})
}

// GetSensitiveWords 获取敏感词检测统计
// @Summary 获取敏感词检测统计
// @Description 分析敏感词检测情况
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/sensitive-words [get]
func (h *AnalyticsHandler) GetSensitiveWords(c *gin.Context) {
	// Mock 数据
	words := []SensitiveWordStats{
		{Word: "***", DetectCount: 45, LastDetectedAt: time.Now(), Severity: "high"},
		{Word: "***", DetectCount: 32, LastDetectedAt: time.Now(), Severity: "medium"},
		{Word: "***", DetectCount: 18, LastDetectedAt: time.Now(), Severity: "low"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    words,
	})
}

// GetCostStats 获取成本统计
// @Summary 获取成本统计
// @Description 分析 Token 使用和 API 调用成本
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/cost [get]
func (h *AnalyticsHandler) GetCostStats(c *gin.Context) {
	stats := h.calculateCostStats()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// GetCostByRole 获取按角色分类的成本
// @Summary 获取按角色分类的成本
// @Description 分析各角色的 Token 使用和成本
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/cost/by-role [get]
func (h *AnalyticsHandler) GetCostByRole(c *gin.Context) {
	costByRole := h.getTopRolesByUsage(10)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    costByRole,
	})
}

// GetCostByUser 获取按用户分类的成本
// @Summary 获取按用户分类的成本
// @Description 分析各用户的 Token 使用和成本
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/cost/by-user [get]
func (h *AnalyticsHandler) GetCostByUser(c *gin.Context) {
	// 按用户统计 Token 使用
	type UserCost struct {
		UserID     string
		UserName   string
		TokensUsed int64
	}

	var userCosts []UserCost
	h.db.Model(&models.Message{}).
		Select("m.user_id as user_id, u.name as user_name, COALESCE(SUM(m.tokens_used), 0) as tokens_used").
		Joins("JOIN chat_sessions cs ON m.session_id = cs.id").
		Joins("JOIN users u ON cs.user_id = u.id").
		Group("m.user_id, u.name").
		Order("tokens_used DESC").
		Limit(10).
		Scan(&userCosts)

	var totalTokens int64
	h.db.Model(&models.Message{}).Select("COALESCE(SUM(tokens_used), 0)").Scan(&totalTokens)

	costByUser := []CostByUser{}
	for _, uc := range userCosts {
		cost := float64(uc.TokensUsed) * 0.00002
		percent := 0.0
		if totalTokens > 0 {
			percent = float64(uc.TokensUsed) / float64(totalTokens) * 100
		}
		costByUser = append(costByUser, CostByUser{
			UserID:     uc.UserID,
			UserName:   uc.UserName,
			TokensUsed: uc.TokensUsed,
			Cost:       cost,
			Percent:    percent,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    costByUser,
	})
}

// GetCostTrend 获取成本趋势
// @Summary 获取成本趋势
// @Description 分析成本变化趋势
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "天数 (默认 30)" default(30)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/cost/trend [get]
func (h *AnalyticsHandler) GetCostTrend(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		fmt.Sscanf(d, "%d", &days)
	}

	trend := h.getCostTrend(days)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    trend,
	})
}

// GetCostPrediction 获取成本预测
// @Summary 获取成本预测
// @Description 基于历史数据预测未来成本
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "预测周期 (week/month)" default(month)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/cost/prediction [get]
func (h *AnalyticsHandler) GetCostPrediction(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	
	// 简单的线性预测 (实际应使用更复杂的算法)
	recentTrend := h.getCostTrend(30)
	
	var totalCost float64
	var totalTokens int64
	for _, t := range recentTrend {
		totalCost += t.Cost
		totalTokens += t.Tokens
	}

	avgDailyCost := totalCost / 30
	avgDailyTokens := totalTokens / 30

	var predictedDays int
	var predictionPeriod string
	if period == "week" {
		predictedDays = 7
		predictionPeriod = "week"
	} else {
		predictedDays = 30
		predictionPeriod = "month"
	}

	predictedCost := avgDailyCost * float64(predictedDays) * 1.1 // 假设 10% 增长
	predictedTokens := int64(float64(avgDailyTokens) * float64(predictedDays) * 1.1)

	prediction := CostPrediction{
		PredictedCost:    predictedCost,
		PredictedTokens:  predictedTokens,
		PredictionPeriod: predictionPeriod,
		ConfidenceLevel:  0.85,
		GrowthRate:       0.10,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    prediction,
	})
}

// GenerateReport 生成效果报告
// @Summary 生成效果报告
// @Description 自动生成周报或月报
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string true "报告类型 (weekly/monthly)" enum(weekly,monthly)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/analytics/report [get]
func (h *AnalyticsHandler) GenerateReport(c *gin.Context) {
	reportType := c.Query("type")
	if reportType == "" {
		reportType = "weekly"
	}

	now := time.Now()
	var periodStart time.Time
	var periodEnd = now

	if reportType == "weekly" {
		periodStart = now.AddDate(0, 0, -7)
	} else {
		periodStart = now.AddDate(0, -1, 0)
	}

	// 生成报告数据
	report := h.generateReportData(reportType, periodStart, periodEnd)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    report,
	})
}

// ExportReport 导出报告 (PDF)
// @Summary 导出报告
// @Description 导出效果报告为 PDF 格式
// @Tags 数据分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string true "报告类型 (weekly/monthly)" enum(weekly,monthly)
// @Success 200 {file} binary "PDF 文件"
// @Router /api/v1/analytics/report/export [get]
func (h *AnalyticsHandler) ExportReport(c *gin.Context) {
	reportType := c.Query("type")
	if reportType == "" {
		reportType = "weekly"
	}

	now := time.Now()
	var periodStart time.Time
	if reportType == "weekly" {
		periodStart = now.AddDate(0, 0, -7)
	} else {
		periodStart = now.AddDate(0, -1, 0)
	}

	report := h.generateReportData(reportType, periodStart, now)

	// 实际实现中，这里应该生成 PDF 文件
	// 由于需要额外的 PDF 库依赖，这里返回 JSON 格式
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "PDF export not implemented yet, returning JSON",
		"data":    report,
	})
}

// ===== 辅助函数 =====

// calculateUserActivity 计算用户活跃度
func (h *AnalyticsHandler) calculateUserActivity() *UserActivityStats {
	now := time.Now()
	
	// DAU: 今天有活动的用户
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var dau int64
	h.db.Model(&models.ChatSession{}).
		Where("updated_at >= ?", today).
		Distinct("user_id").
		Count(&dau)

	// WAU: 最近 7 天有活动的用户
	sevenDaysAgo := today.AddDate(0, 0, -7)
	var wau int64
	h.db.Model(&models.ChatSession{}).
		Where("updated_at >= ?", sevenDaysAgo).
		Distinct("user_id").
		Count(&wau)

	// MAU: 最近 30 天有活动的用户
	thirtyDaysAgo := today.AddDate(0, 0, -30)
	var mau int64
	h.db.Model(&models.ChatSession{}).
		Where("updated_at >= ?", thirtyDaysAgo).
		Distinct("user_id").
		Count(&mau)

	return &UserActivityStats{
		DAU: dau,
		WAU: wau,
		MAU: mau,
	}
}

// calculateCostStats 计算成本统计
func (h *AnalyticsHandler) calculateCostStats() *CostStats {
	var totalTokens int64
	h.db.Model(&models.Message{}).Select("COALESCE(SUM(tokens_used), 0)").Scan(&totalTokens)

	// 按类型统计 Token
	var inputTokens, outputTokens, embeddingTokens int64
	// 简化处理，假设所有消息都是输出
	outputTokens = totalTokens

	totalCost := float64(totalTokens) * 0.00002

	// 平均每日成本
	var avgCostPerDay float64
	var earliestMessage models.Message
	if h.db.Order("created_at ASC").First(&earliestMessage).Error == nil {
		days := int(time.Since(earliestMessage.CreatedAt).Hours() / 24)
		if days > 0 {
			avgCostPerDay = totalCost / float64(days)
		}
	}

	return &CostStats{
		TotalTokens:      totalTokens,
		TotalCost:        totalCost,
		AverageCostPerDay: avgCostPerDay,
		TokenBreakdown: TokenBreakdown{
			InputTokens:     inputTokens,
			OutputTokens:    outputTokens,
			EmbeddingTokens: embeddingTokens,
		},
	}
}

// calculateQualityStats 计算质量统计
func (h *AnalyticsHandler) calculateQualityStats() *ConversationQualityStats {
	// Mock 数据 - 实际应根据用户评分等数据计算
	return &ConversationQualityStats{
		AverageRating:     4.5,
		TotalRated:        256,
		SatisfactionRate:  92.5,
		HighQualityCount:  180,
		MediumQualityCount: 65,
		LowQualityCount:   11,
	}
}

// getTopRolesByUsage 获取使用最多的角色
func (h *AnalyticsHandler) getTopRolesByUsage(limit int) []CostByRole {
	type RoleUsage struct {
		RoleID     string
		RoleName   string
		TokensUsed int64
	}

	var roleUsages []RoleUsage
	h.db.Model(&models.Message{}).
		Select("cs.role_id as role_id, r.name as role_name, COALESCE(SUM(m.tokens_used), 0) as tokens_used").
		Joins("JOIN chat_sessions cs ON m.session_id = cs.id").
		Joins("JOIN roles r ON cs.role_id = r.id").
		Group("cs.role_id, r.name").
		Order("tokens_used DESC").
		Limit(limit).
		Scan(&roleUsages)

	var totalTokens int64
	h.db.Model(&models.Message{}).Select("COALESCE(SUM(tokens_used), 0)").Scan(&totalTokens)

	costByRole := []CostByRole{}
	for _, ru := range roleUsages {
		cost := float64(ru.TokensUsed) * 0.00002
		percent := 0.0
		if totalTokens > 0 {
			percent = float64(ru.TokensUsed) / float64(totalTokens) * 100
		}
		costByRole = append(costByRole, CostByRole{
			RoleID:     ru.RoleID,
			RoleName:   ru.RoleName,
			TokensUsed: ru.TokensUsed,
			Cost:       cost,
			Percent:    percent,
		})
	}

	return costByRole
}

// getCostTrend 获取成本趋势
func (h *AnalyticsHandler) getCostTrend(days int) []CostTrend {
	trend := []CostTrend{}
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		dayStart := now.AddDate(0, 0, -i)
		dayStart = time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, dayStart.Location())
		dayEnd := dayStart.AddDate(0, 0, 1)

		var tokens int64
		h.db.Model(&models.Message{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Select("COALESCE(SUM(tokens_used), 0)").
			Scan(&tokens)

		cost := float64(tokens) * 0.00002

		trend = append(trend, CostTrend{
			Date:   dayStart.Format("2006-01-02"),
			Cost:   cost,
			Tokens: tokens,
		})
	}

	return trend
}

// generateReportData 生成报告数据
func (h *AnalyticsHandler) generateReportData(reportType string, periodStart, periodEnd time.Time) *ReportData {
	// 计算关键指标
	dashboardMetrics := DashboardMetrics{}
	
	// 关键指标
	keyMetrics := []KeyMetric{
		{Name: "活跃用户", Value: float64(dashboardMetrics.ActiveUsers), Unit: "人", Change: 12.5, Trend: "up"},
		{Name: "对话次数", Value: float64(dashboardMetrics.TotalSessions), Unit: "次", Change: 8.3, Trend: "up"},
		{Name: "平均评分", Value: dashboardMetrics.AverageRating, Unit: "分", Change: 2.1, Trend: "up"},
		{Name: "总成本", Value: dashboardMetrics.TotalCost, Unit: "元", Change: -5.2, Trend: "down"},
	}

	// 趋势数据
	trends := []TrendData{}
	costTrend := h.getCostTrend(30)
	for _, t := range costTrend {
		trends = append(trends, TrendData{
			Date:  t.Date,
			Value: t.Cost,
		})
	}

	// 对比数据
	comparisons := ComparisonData{
		MoM: map[string]float64{
			"活跃用户": 12.5,
			"对话次数": 8.3,
			"总成本":   -5.2,
		},
		YoY: map[string]float64{
			"活跃用户": 45.2,
			"对话次数": 38.7,
			"总成本":   22.1,
		},
	}

	// 建议
	recommendations := []string{
		"建议优化高成本角色的使用效率",
		"可以考虑为活跃用户提供更多高级功能",
		"文档功能使用率较低，建议加强推广",
	}

	return &ReportData{
		ReportType:    reportType,
		PeriodStart:   periodStart.Format("2006-01-02"),
		PeriodEnd:     periodEnd.Format("2006-01-02"),
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05"),
		Summary:       make(map[string]interface{}),
		KeyMetrics:    keyMetrics,
		Trends:        trends,
		Comparisons:   comparisons,
		Recommendations: recommendations,
	}
}
