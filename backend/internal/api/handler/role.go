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
type CreateRoleRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	Category       string `json:"category"`
	SystemPrompt   string `json:"systemPrompt" binding:"required"`
	WelcomeMessage string `json:"welcomeMessage"`
}

// List 获取角色列表
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

// Chat 与角色对话
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

	// TODO: 集成AI服务进行对话
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"role":    role.Name,
			"message": "收到消息: " + req.Message,
			"reply":   "这是AI的回复（待集成OpenAI API）",
		},
	})
}
