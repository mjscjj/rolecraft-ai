package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

// TestHandler 测试处理器
type TestHandler struct {
	db *gorm.DB
}

// NewTestHandler 创建测试处理器
func NewTestHandler(db *gorm.DB) *TestHandler {
	return &TestHandler{
		db: db,
	}
}

// TestMessageRequest 测试消息请求
type TestMessageRequest struct {
	Content        string                 `json:"content" binding:"required"`
	SystemPrompt   string                 `json:"systemPrompt"`
	ModelConfig    map[string]interface{} `json:"modelConfig"`
	RoleName       string                 `json:"roleName"`
}

// TestMessageResponse 测试消息响应
type TestMessageResponse struct {
	Content      string                 `json:"content"`
	ResponseTime float64                `json:"responseTime"` // seconds
	Tokens       int                    `json:"tokens"`
	Model        string                 `json:"model"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ABTestVersion A/B 测试版本
type ABTestVersion struct {
	VersionID    string                 `json:"versionId"`
	VersionName  string                 `json:"versionName"`
	SystemPrompt string                 `json:"systemPrompt"`
	ModelConfig  map[string]interface{} `json:"modelConfig"`
}

// ABTestRequest A/B 测试请求
type ABTestRequest struct {
	Versions []ABTestVersion `json:"versions" binding:"required,min=2"`
	Question string          `json:"question" binding:"required"`
}

// ABTestResult A/B 测试结果
type ABTestResult struct {
	TestID      string             `json:"testId"`
	Question    string             `json:"question"`
	Results     []ABTestResultItem `json:"results"`
	WinnerID    string             `json:"winnerId,omitempty"`
	CreatedAt   time.Time          `json:"createdAt"`
}

// ABTestResultItem A/B 测试结果项
type ABTestResultItem struct {
	VersionID    string  `json:"versionId"`
	VersionName  string  `json:"versionName"`
	Response     string  `json:"response"`
	ResponseTime float64 `json:"responseTime"`
	Score        float64 `json:"score"`
	Rating       int     `json:"rating"` // 1-5
	Feedback     string  `json:"feedback"`
}

// TestHistory 测试历史
type TestHistory struct {
	TestID       string    `json:"testId"`
	RoleID       string    `json:"roleId"`
	RoleName     string    `json:"roleName"`
	TestType     string    `json:"testType"` // single/ab
	Question     string    `json:"question"`
	Response     string    `json:"response"`
	Rating       int       `json:"rating"`
	Feedback     string    `json:"feedback"`
	CreatedAt    time.Time `json:"createdAt"`
}

// TestReportRequest 测试报告请求
type TestReportRequest struct {
	RoleID    string `json:"roleId"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// TestReportResponse 测试报告响应
type TestReportResponse struct {
	RoleID         string              `json:"roleId"`
	RoleName       string              `json:"roleName"`
	TotalTests     int                 `json:"totalTests"`
	AverageRating  float64             `json:"averageRating"`
	PassRate       float64             `json:"passRate"`
	TestsByRating  map[int]int         `json:"testsByRating"`
	ImprovementTrend []RatingTrendItem `json:"improvementTrend"`
	Suggestions    []string            `json:"suggestions"`
	ExportURL      string              `json:"exportUrl"`
}

// RatingTrendItem 评分趋势项
type RatingTrendItem struct {
	Date       string  `json:"date"`
	AvgRating  float64 `json:"avgRating"`
	TestCount  int     `json:"testCount"`
}

// SendMessage 发送测试消息
func (h *TestHandler) SendMessage(c *gin.Context) {
	var req TestMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 模拟 AI 回复
	response := h.generateMockResponse(req)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": response,
	})
}

// generateMockResponse 生成模拟回复
func (h *TestHandler) generateMockResponse(req TestMessageRequest) TestMessageResponse {
	startTime := time.Now()

	// 基于系统提示词生成回复
	response := h.createResponse(req.SystemPrompt, req.Content, req.RoleName)

	responseTime := time.Since(startTime).Seconds()

	return TestMessageResponse{
		Content:      response,
		ResponseTime: responseTime,
		Tokens:       len(response) / 4, // 估算
		Model:        "mock-model",
		Metadata: map[string]interface{}{
			"promptLength": len(req.SystemPrompt),
			"messageLength": len(req.Content),
		},
	}
}

// createResponse 创建回复
func (h *TestHandler) createResponse(systemPrompt, content, roleName string) string {
	prompt := strings.ToLower(systemPrompt)
	
	// 根据角色类型生成不同的回复
	if strings.Contains(prompt, "营销") || strings.Contains(prompt, "创意") {
		return fmt.Sprintf("【%s】这是一个很有创意的问题！基于我的专业经验，我建议我们可以从以下几个角度来思考...", roleName)
	} else if strings.Contains(prompt, "法律") || strings.Contains(prompt, "合同") {
		return fmt.Sprintf("【%s】从法律角度来看，这个问题需要注意以下几点：首先，我们需要明确相关条款的法律效力...", roleName)
	} else if strings.Contains(prompt, "心理") || strings.Contains(prompt, "咨询") {
		return fmt.Sprintf("【%s】我理解你的感受。能详细说说是什么让你有这样的想法吗？我会一直在这里倾听...", roleName)
	} else if strings.Contains(prompt, "技术") || strings.Contains(prompt, "编程") {
		return fmt.Sprintf("【%s】从技术层面分析，这个问题的核心在于...让我给你举个代码示例...", roleName)
	} else {
		return fmt.Sprintf("【%s】你好！收到你的问题：%s。让我来帮你解答...", roleName, content)
	}
}

// RunABTest 运行 A/B 测试
func (h *TestHandler) RunABTest(c *gin.Context) {
	var req ABTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Versions) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "至少需要 2 个版本进行对比"})
		return
	}

	results := []ABTestResultItem{}
	
	// 对每个版本进行测试
	for _, version := range req.Versions {
		result := h.testVersion(version, req.Question)
		results = append(results, result)
	}

	// 自动选择优胜者
	winnerID := h.selectWinner(results)

	testResult := ABTestResult{
		TestID:    models.NewUUID(),
		Question:  req.Question,
		Results:   results,
		WinnerID:  winnerID,
		CreatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    testResult,
	})
}

// testVersion 测试单个版本
func (h *TestHandler) testVersion(version ABTestVersion, question string) ABTestResultItem {
	startTime := time.Now()
	
	// 生成回复
	response := h.createResponse(version.SystemPrompt, question, version.VersionName)
	
	responseTime := time.Since(startTime).Seconds()
	
	// 模拟评分（基于回复质量）
	score := h.evaluateResponse(response, question)
	rating := int(score / 20) // 转换为 1-5 星
	if rating < 1 {
		rating = 1
	}
	if rating > 5 {
		rating = 5
	}

	return ABTestResultItem{
		VersionID:    version.VersionID,
		VersionName:  version.VersionName,
		Response:     response,
		ResponseTime: responseTime,
		Score:        score,
		Rating:       rating,
		Feedback:     h.generateFeedback(response, score),
	}
}

// evaluateResponse 评估回复质量
func (h *TestHandler) evaluateResponse(response, question string) float64 {
	// 简化评估逻辑
	baseScore := 70.0
	
	// 回复长度适中加分
	if len(response) > 50 && len(response) < 500 {
		baseScore += 10
	}
	
	// 包含关键词加分
	if strings.Contains(response, "建议") || strings.Contains(response, "分析") {
		baseScore += 10
	}
	
	// 回复相关性（简化）
	if len(question) > 0 && len(response) > len(question) {
		baseScore += 10
	}
	
	return baseScore
}

// generateFeedback 生成反馈
func (h *TestHandler) generateFeedback(response string, score float64) string {
	if score >= 90 {
		return "回复质量优秀，专业且全面"
	} else if score >= 75 {
		return "回复质量良好，基本满足需求"
	} else if score >= 60 {
		return "回复质量一般，有待改进"
	} else {
		return "回复质量较差，建议优化提示词"
	}
}

// selectWinner 选择优胜者
func (h *TestHandler) selectWinner(results []ABTestResultItem) string {
	if len(results) == 0 {
		return ""
	}
	
	// 按评分排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	return results[0].VersionID
}

// SaveTestResult 保存测试结果
func (h *TestHandler) SaveTestResult(c *gin.Context) {
	var req struct {
		RoleID     string `json:"roleId"`
		RoleName   string `json:"roleName"`
		TestType   string `json:"testType"`
		Question   string `json:"question"`
		Response   string `json:"response"`
		Rating     int    `json:"rating"`
		Feedback   string `json:"feedback"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存到数据库（简化实现）
	// 实际应创建 TestHistory 模型并保存
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"saved": true,
			"testId": models.NewUUID(),
		},
	})
}

// GetTestHistory 获取测试历史
func (h *TestHandler) GetTestHistory(c *gin.Context) {
	roleID := c.Query("roleId")
	
	// 查询测试历史（简化实现）
	history := []TestHistory{}
	
	// 模拟数据
	for i := 0; i < 5; i++ {
		history = append(history, TestHistory{
			TestID:    models.NewUUID(),
			RoleID:    roleID,
			RoleName:  "测试角色",
			TestType:  "single",
			Question:  fmt.Sprintf("测试问题 %d", i+1),
			Response:  fmt.Sprintf("这是测试回复 %d", i+1),
			Rating:    4 + (i % 2),
			Feedback:  "回复质量良好",
			CreatedAt: time.Now().Add(-time.Duration(i) * 24 * time.Hour),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    history,
		"total":   len(history),
	})
}

// GetTestReport 获取测试报告
func (h *TestHandler) GetTestReport(c *gin.Context) {
	roleID := c.Query("roleId")
	
	// 获取角色信息
	var role models.Role
	if result := h.db.First(&role, "id = ?", roleID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	
	// 生成测试报告
	report := h.generateTestReport(role)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    report,
	})
}

// generateTestReport 生成测试报告
func (h *TestHandler) generateTestReport(role models.Role) TestReportResponse {
	// 模拟测试数据
	testsByRating := map[int]int{
		5: 15,
		4: 25,
		3: 8,
		2: 2,
		1: 0,
	}
	
	totalTests := 50
	avgRating := 4.1
	passRate := 86.0
	
	// 生成趋势数据
	trend := []RatingTrendItem{}
	for i := 6; i >= 0; i-- {
		date := time.Now().Add(-time.Duration(i) * 24 * time.Hour).Format("2006-01-02")
		trend = append(trend, RatingTrendItem{
			Date:      date,
			AvgRating: 3.8 + float64(7-i)*0.05,
			TestCount: 5 + i*2,
		})
	}
	
	// 生成改进建议
	suggestions := []string{
		"系统提示词可以更具体一些，增加角色背景描述",
		"建议增加示例对话，帮助 AI 更好理解角色定位",
		"可以考虑调整模型参数，提升回复创造性",
	}

	return TestReportResponse{
		RoleID:         role.ID,
		RoleName:       role.Name,
		TotalTests:     totalTests,
		AverageRating:  avgRating,
		PassRate:       passRate,
		TestsByRating:  testsByRating,
		ImprovementTrend: trend,
		Suggestions:    suggestions,
		ExportURL:      fmt.Sprintf("/api/v1/test/export/%s", role.ID),
	}
}

// ExportTestReport 导出测试报告
func (h *TestHandler) ExportTestReport(c *gin.Context) {
	roleID := c.Param("roleId")
	format := c.Query("format")
	
	if format == "" {
		format = "pdf"
	}
	
	// 生成导出内容（简化）
	content := fmt.Sprintf("测试报告 - 角色 ID: %s\n导出时间：%s\n格式：%s", 
		roleID, time.Now().Format("2006-01-02 15:04:05"), format)
	
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"test_report_%s.%s\"", roleID, format))
	c.Data(http.StatusOK, "application/octet-stream", []byte(content))
}

// RateTestResponse 评分测试回复
func (h *TestHandler) RateTestResponse(c *gin.Context) {
	var req struct {
		TestID   string `json:"testId" binding:"required"`
		Rating   int    `json:"rating" binding:"required,min=1,max=5"`
		Feedback string `json:"feedback"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存评分（简化实现）
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"rated": true,
			"testId": req.TestID,
			"rating": req.Rating,
		},
	})
}

// CompareVersions 对比多个版本
func (h *TestHandler) CompareVersions(c *gin.Context) {
	var req struct {
		VersionIDs []string `json:"versionIds" binding:"required,min=2"`
		Question   string   `json:"question" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 并排对比测试
	results := []ABTestResultItem{}
	for _, versionID := range req.VersionIDs {
		version := ABTestVersion{
			VersionID:   versionID,
			VersionName: fmt.Sprintf("版本 %s", versionID),
			SystemPrompt: "默认系统提示词",
		}
		result := h.testVersion(version, req.Question)
		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"question": req.Question,
			"results":  results,
		},
	})
}
