package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	db *gorm.DB
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

// CreateRoleRequest 创建角色请求
// @Description 创建或更新角色的请求体
type CreateRoleRequest struct {
	// 角色名称
	Name string `json:"name" binding:"required" example:"智能助理"`
	// 角色描述
	Description string `json:"description" example:"全能型办公助手"`
	// 角色分类
	Category string `json:"category" example:"通用"`
	// 系统提示词
	SystemPrompt string `json:"systemPrompt" binding:"required" example:"你是一位智能助理..."`
	// 欢迎消息
	WelcomeMessage string `json:"welcomeMessage" example:"你好！有什么可以帮你的吗？"`
}

// List 获取角色列表
// @Summary 获取角色列表
// @Description 获取所有角色，支持分类筛选和模板筛选
// @Tags 角色
// @Produce json
// @Security BearerAuth
// @Param category query string false "分类筛选"
// @Param template query string false "是否仅返回模板 (true/false)"
// @Success 200 {object} map[string]interface{} "角色列表"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	var roles []models.Role

	query := h.db.Preload("Skills").Preload("Documents")

	// 分类筛选
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// 只显示模板
	if c.Query("template") == "true" {
		query = query.Where("is_template = ?", true)
	}

	if result := query.Find(&roles); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    roles,
	})
}

// Get 获取单个角色
// @Summary 获取角色详情
// @Description 根据 ID 获取单个角色的详细信息
// @Tags 角色
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色 ID"
// @Success 200 {object} map[string]interface{} "角色详情"
// @Failure 404 {object} map[string]string "角色不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/roles/{id} [get]
func (h *RoleHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var role models.Role
	if result := h.db.Preload("Skills").Preload("Documents").First(&role, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    role,
	})
}

// Create 创建角色
// @Summary 创建角色
// @Description 创建一个新的 AI 角色
// @Tags 角色
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRoleRequest true "角色信息"
// @Success 201 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := models.Role{
		ID:             models.NewUUID(),
		Name:           req.Name,
		Description:    req.Description,
		Category:       req.Category,
		SystemPrompt:   req.SystemPrompt,
		WelcomeMessage: req.WelcomeMessage,
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

// Update 更新角色
// @Summary 更新角色
// @Description 更新指定角色的信息
// @Tags 角色
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色 ID"
// @Param request body CreateRoleRequest true "角色信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 404 {object} map[string]string "角色不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
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

	// 更新字段
	role.Name = req.Name
	role.Description = req.Description
	role.Category = req.Category
	role.SystemPrompt = req.SystemPrompt
	role.WelcomeMessage = req.WelcomeMessage

	if result := h.db.Save(&role); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    role,
	})
}

// Delete 删除角色
// @Summary 删除角色
// @Description 删除指定的角色
// @Tags 角色
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色 ID"
// @Success 200 {object} map[string]string "删除成功"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if result := h.db.Delete(&models.Role{}, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// GetTemplates 获取内置模板
// @Summary 获取角色模板
// @Description 获取系统内置的角色模板 (无需认证)
// @Tags 角色
// @Produce json
// @Success 200 {object} map[string]interface{} "角色模板列表"
// @Router /api/v1/roles/templates [get]
func (h *RoleHandler) GetTemplates(c *gin.Context) {
	templates := []models.Role{
		{
			ID:             "11111111-1111-1111-1111-111111111111",
			Name:           "智能助理",
			Description:    "全能型办公助手，帮助处理日常事务、撰写邮件、安排日程",
			Category:       "通用",
			SystemPrompt:   "你是一位智能助理，擅长帮助用户处理各种办公任务。请用友好、专业的态度回答用户的问题。",
			WelcomeMessage: "你好！我是你的智能助理，有什么可以帮你的吗？",
			IsTemplate:     true,
		},
		{
			ID:             "22222222-2222-2222-2222-222222222222",
			Name:           "营销专家",
			Description:    "专业的营销策划助手，帮助制定营销策略、撰写文案",
			Category:       "营销",
			SystemPrompt:   "你是一位资深的营销专家，精通各种营销策略和内容创作。请提供有创意、可执行的营销建议。",
			WelcomeMessage: "你好！我是你的营销顾问，让我们一起制定出色的营销策略吧！",
			IsTemplate:     true,
		},
		{
			ID:             "33333333-3333-3333-3333-333333333333",
			Name:           "法务顾问",
			Description:    "合同审查与法律咨询专家，协助审查合同条款、解答法律问题",
			Category:       "法律",
			SystemPrompt:   "你是一位专业的法务顾问，擅长合同审查和法律咨询。请提供准确、实用的法律建议。",
			WelcomeMessage: "你好！我是你的法务顾问，有什么法律问题需要咨询吗？",
			IsTemplate:     true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    templates,
	})
}

// ChatRequest 对话请求
// @Description 与角色对话的请求体
type ChatRequest struct {
	// 用户消息内容
	Message string `json:"message" binding:"required" example:"你好，请帮我写一封邮件"`
}

// Chat 与角色对话
// @Summary 与角色对话
// @Description 发送消息给指定角色并获取 AI 回复
// @Tags 角色
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色 ID"
// @Param request body ChatRequest true "对话内容"
// @Success 200 {object} map[string]interface{} "对话响应"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 404 {object} map[string]string "角色不存在"
// @Router /api/v1/roles/{id}/chat [post]
func (h *RoleHandler) Chat(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if result := h.db.First(&role, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// TODO: 集成 AI 服务进行对话
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"role":    role.Name,
			"message": "收到消息：" + req.Message,
			"reply":   "这是 AI 的回复（待集成 OpenAI API）",
		},
	})
}
