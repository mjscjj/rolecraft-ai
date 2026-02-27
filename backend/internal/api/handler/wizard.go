package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rolecraft-ai/internal/service/prompt"
)

// WizardHandler 向导处理器
type WizardHandler struct {
	generator *prompt.PromptGenerator
}

// NewWizardHandler 创建向导处理器
func NewWizardHandler() *WizardHandler {
	return &WizardHandler{
		generator: prompt.NewPromptGenerator(),
	}
}

// WizardDataRequest 向导数据请求
type WizardDataRequest struct {
	Name              string   `json:"name" binding:"required"`
	Purpose           string   `json:"purpose" binding:"required"`
	Style             string   `json:"style" binding:"required"`
	Expertise         []string `json:"expertise"`
	Avoidances        []string `json:"avoidances"`
	SpecialRequirements string `json:"specialRequirements"`
	TestMessage       string   `json:"testMessage"`
	Satisfaction      *int     `json:"satisfaction"`
}

// GeneratePrompt 生成提示词
// POST /api/wizard/generate
func (h *WizardHandler) GeneratePrompt(c *gin.Context) {
	var req WizardDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误：" + err.Error(),
		})
		return
	}

	// 转换为内部数据结构
	data := prompt.WizardData{
		Name:              req.Name,
		Purpose:           req.Purpose,
		Style:             req.Style,
		Expertise:         req.Expertise,
		Avoidances:        req.Avoidances,
		SpecialRequirements: req.SpecialRequirements,
		TestMessage:       req.TestMessage,
		Satisfaction:      req.Satisfaction,
	}

	// 生成提示词
	generated := h.generator.GeneratePrompt(data)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    generated,
	})
}

// GetRecommendations 获取智能推荐
// POST /api/wizard/recommendations
func (h *WizardHandler) GetRecommendations(c *gin.Context) {
	var req WizardDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误：" + err.Error(),
		})
		return
	}

	data := prompt.WizardData{
		Name:              req.Name,
		Purpose:           req.Purpose,
		Style:             req.Style,
		Expertise:         req.Expertise,
		Avoidances:        req.Avoidances,
		SpecialRequirements: req.SpecialRequirements,
	}

	recommendations := h.generator.GetRecommendations(data)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    recommendations,
		"total":   len(recommendations),
	})
}

// RunTest 运行测试对话
// POST /api/wizard/test
func (h *WizardHandler) RunTest(c *gin.Context) {
	var req struct {
		WizardDataRequest
		TestMessage string `json:"testMessage" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误：" + err.Error(),
		})
		return
	}

	data := prompt.WizardData{
		Name:              req.Name,
		Purpose:           req.Purpose,
		Style:             req.Style,
		Expertise:         req.Expertise,
		Avoidances:        req.Avoidances,
		SpecialRequirements: req.SpecialRequirements,
	}

	result := h.generator.RunTest(data, req.TestMessage)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// GetOptions 获取所有配置选项
// GET /api/wizard/options
func (h *WizardHandler) GetOptions(c *gin.Context) {
	options := gin.H{
		"purposes":   prompt.Purposes,
		"styles":     prompt.Styles,
		"expertise":  prompt.ExpertiseAreas,
		"avoidances": prompt.Avoidances,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    options,
	})
}

// ExportConfig 导出角色配置
// POST /api/wizard/export
func (h *WizardHandler) ExportConfig(c *gin.Context) {
	var req WizardDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误：" + err.Error(),
		})
		return
	}

	data := prompt.WizardData{
		Name:              req.Name,
		Purpose:           req.Purpose,
		Style:             req.Style,
		Expertise:         req.Expertise,
		Avoidances:        req.Avoidances,
		SpecialRequirements: req.SpecialRequirements,
	}

	config := h.generator.ExportRoleConfig(data)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    config,
	})
}

// ValidateData 验证向导数据
// POST /api/wizard/validate
func (h *WizardHandler) ValidateData(c *gin.Context) {
	var req WizardDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误：" + err.Error(),
		})
		return
	}

	errors := []string{}
	warnings := []string{}

	// 验证必填字段
	if len(req.Name) == 0 {
		errors = append(errors, "角色名称不能为空")
	} else if len(req.Name) > 20 {
		warnings = append(warnings, "角色名称建议不超过 20 个字")
	}

	if req.Purpose == "" {
		errors = append(errors, "请选择主要用途")
	}

	if req.Style == "" {
		errors = append(errors, "请选择说话风格")
	}

	// 验证专业领域
	if len(req.Expertise) == 0 {
		errors = append(errors, "请至少选择一个专业领域")
	} else if len(req.Expertise) > 5 {
		warnings = append(warnings, "专业领域建议不超过 5 个，过多可能导致焦点分散")
	}

	// 验证特殊要求
	if len(req.SpecialRequirements) > 500 {
		warnings = append(warnings, "特殊要求内容较长，建议精简")
	}

	// 返回验证结果
	valid := len(errors) == 0
	
	response := gin.H{
		"code":     200,
		"message":  "success",
		"data": gin.H{
			"valid":    valid,
			"errors":   errors,
			"warnings": warnings,
		},
	}

	if !valid {
		response["code"] = 400
		response["message"] = "验证未通过"
	}

	c.JSON(http.StatusOK, response)
}

// GetTemplates 获取推荐模板
// GET /api/wizard/templates
func (h *WizardHandler) GetTemplates(c *gin.Context) {
	category := c.Query("category")
	
	templates := []gin.H{
		{
			"id":          "template_001",
			"name":        "智能助理",
			"description": "全能型办公助手，处理日常事务",
			"purpose":     "assistant",
			"style":       "professional",
			"expertise":   []string{"business"},
			"preview":     "你好！我是你的智能助理，有什么可以帮你的吗？",
		},
		{
			"id":          "template_002",
			"name":        "营销专家",
			"description": "专业的营销策划助手",
			"purpose":     "creator",
			"style":       "humorous",
			"expertise":   []string{"marketing"},
			"preview":     "你好！让我们一起制定出色的营销策略吧！",
		},
		{
			"id":          "template_003",
			"name":        "编程导师",
			"description": "经验丰富的软件工程师",
			"purpose":     "teacher",
			"style":       "friendly",
			"expertise":   []string{"tech"},
			"preview":     "你好！我是你的编程导师，有什么问题尽管问我！",
		},
		{
			"id":          "template_004",
			"name":        "心理咨询师",
			"description": "专业的心理健康支持者",
			"purpose":     "companion",
			"style":       "friendly",
			"expertise":   []string{"health"},
			"preview":     "你好，我在这里倾听你的心声。今天想聊些什么呢？",
		},
	}

	// 分类筛选
	if category != "" && category != "all" {
		filtered := []gin.H{}
		for _, t := range templates {
			if expertise, ok := t["expertise"].([]string); ok {
				for _, e := range expertise {
					if e == category {
						filtered = append(filtered, t)
						break
					}
				}
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
