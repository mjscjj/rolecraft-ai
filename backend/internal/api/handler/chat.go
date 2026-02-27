package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/models"
	"rolecraft-ai/internal/service/anythingllm"
	"rolecraft-ai/internal/service/thinking"
)

// ChatHandler 对话处理器
type ChatHandler struct {
	db            *gorm.DB
	config        *config.Config
	thinkingSvc   *thinking.Service
}

// NewChatHandler 创建对话处理器
func NewChatHandler(db *gorm.DB, cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		db:          db,
		config:      cfg,
		thinkingSvc: thinking.NewService(),
	}
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	RoleID          string `json:"roleId" binding:"required"`
	Title           string `json:"title"`
	Mode            string `json:"mode"`
	AnythingLLMSlug string `json:"anythingLLMSlug"` // AnythingLLM Workspace Slug
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Content     string   `json:"content" binding:"required"`
	Attachments []string `json:"attachments"`
}

// AnythingLLMChatRequest AnythingLLM 聊天请求
type AnythingLLMChatRequest struct {
	Message string `json:"message"`
	Mode    string `json:"mode,omitempty"`
}

// AnythingLLMChatResponse AnythingLLM 聊天响应
type AnythingLLMChatResponse struct {
	Response string `json:"response"`
}

// ListSessions 获取对话会话列表
func (h *ChatHandler) ListSessions(c *gin.Context) {
	userId, _ := c.Get("userId")

	var sessions []models.ChatSession
	if result := h.db.Where("user_id = ?", userId).Order("updated_at DESC").Limit(50).Find(&sessions); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    sessions,
	})
}

// CreateSession 创建新会话
func (h *ChatHandler) CreateSession(c *gin.Context) {
	userId, _ := c.Get("userId")

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if result := h.db.First(&role, "id = ?", req.RoleID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	mode := req.Mode
	if mode == "" {
		mode = "quick"
	}
	title := req.Title
	if title == "" {
		title = fmt.Sprintf("与 %s 的对话", role.Name)
	}

	session := models.ChatSession{
		ID:        models.NewUUID(),
		UserID:    userId.(string),
		RoleID:    req.RoleID,
		Title:     title,
		Mode:      mode,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 存储 AnythingLLM Slug
	if req.AnythingLLMSlug != "" {
		session.AnythingLLMSlug = req.AnythingLLMSlug
	}

	if result := h.db.Create(&session); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "success",
		"data":    session,
	})
}

// GetSession 获取会话详情
func (h *ChatHandler) GetSession(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var messages []models.Message
	h.db.Where("session_id = ?", sessionId).Order("created_at ASC").Find(&messages)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"session":  session,
			"messages": messages,
		},
	})
}

// getAnythingLLMSlug 获取会话的 AnythingLLM Slug
func (h *ChatHandler) getAnythingLLMSlug(session models.ChatSession) (string, error) {
	// 优先从会话字段获取
	if session.AnythingLLMSlug != "" {
		return session.AnythingLLMSlug, nil
	}

	// 从角色元数据获取
	var role models.Role
	if result := h.db.First(&role, "id = ?", session.RoleID); result.Error != nil {
		return "", fmt.Errorf("role not found")
	}

	if role.ModelConfig != "" {
		var modelConfig map[string]interface{}
		if err := json.Unmarshal([]byte(role.ModelConfig), &modelConfig); err == nil {
			if slug, ok := modelConfig["anythingllm_slug"].(string); ok && slug != "" {
				return slug, nil
			}
		}
	}

	return "", fmt.Errorf("anythingllm slug not found")
}

// callAnythingLLM 调用 AnythingLLM Chat API
func (h *ChatHandler) callAnythingLLM(slug, message string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/workspace/%s/chat", h.config.AnythingLLMURL, slug)

	reqBody := AnythingLLMChatRequest{
		Message: message,
		Mode:    "chat",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.config.AnythingLLMKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anythingllm API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var chatResp AnythingLLMChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return chatResp.Response, nil
}

// Chat 发送消息（普通响应）- 集成 AnythingLLM
func (h *ChatHandler) Chat(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 保存用户消息
	userMsg := models.Message{
		ID:        models.NewUUID(),
		SessionID: session.ID,
		Role:      "user",
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
	h.db.Create(&userMsg)

	// 获取 AnythingLLM Slug
	var assistantContent string
	slug, err := h.getAnythingLLMSlug(session)
	if err != nil {
		// 如果没有配置 AnythingLLM，使用 Mock AI
		assistantContent = h.callMockAI(req.Content)
	} else {
		// 调用 AnythingLLM API
		assistantContent, err = h.callAnythingLLM(slug, req.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "anythingllm API error: " + err.Error(),
			})
			return
		}
	}

	// 保存助手消息
	assistantMsg := models.Message{
		ID:        models.NewUUID(),
		SessionID: session.ID,
		Role:      "assistant",
		Content:   assistantContent,
		CreatedAt: time.Now(),
	}
	h.db.Create(&assistantMsg)

	// 更新会话时间
	h.db.Model(&session).Update("updated_at", time.Now())

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"userMessage":      userMsg,
			"assistantMessage": assistantMsg,
		},
	})
}

// ChatStream 发送消息（流式响应 SSE）- 集成 AnythingLLM
func (h *ChatHandler) ChatStream(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 保存用户消息
	userMsg := models.Message{
		ID:        models.NewUUID(),
		SessionID: session.ID,
		Role:      "user",
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
	h.db.Create(&userMsg)

	// 获取 AnythingLLM Slug
	slug, err := h.getAnythingLLMSlug(session)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "anythingllm workspace not configured: " + err.Error(),
		})
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// 调用 AnythingLLM Stream API
	url := fmt.Sprintf("%s/api/v1/workspace/%s/stream-chat", h.config.AnythingLLMURL, slug)

	reqBody := AnythingLLMChatRequest{
		Message: req.Content,
		Mode:    "chat",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal request"})
		return
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+h.config.AnythingLLMKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to anythingllm"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("anythingllm error: %s", string(body))})
		return
	}

	// 流式读取响应
	var fullContent strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			// 发送错误到客户端
			data := map[string]interface{}{"error": err.Error(), "done": true}
			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
			flusher.Flush()
			break
		}

		// 提取内容
		if content, ok := chunk["response"].(string); ok && content != "" {
			fullContent.WriteString(content)
			data := map[string]interface{}{"content": content, "done": false}
			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
			flusher.Flush()
		}

		// 检查是否结束
		if done, ok := chunk["done"].(bool); ok && done {
			data := map[string]interface{}{"done": true}
			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
			flusher.Flush()
			break
		}
	}

	// 保存完整的助手消息
	if fullContent.Len() > 0 {
		assistantMsg := models.Message{
			ID:        models.NewUUID(),
			SessionID: session.ID,
			Role:      "assistant",
			Content:   fullContent.String(),
			CreatedAt: time.Now(),
		}
		h.db.Create(&assistantMsg)
		h.db.Model(&session).Update("updated_at", time.Now())
	}
}

// SyncSession 从 AnythingLLM 同步对话历史
func (h *ChatHandler) SyncSession(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	slug, err := h.getAnythingLLMSlug(session)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "anythingllm workspace not configured: " + err.Error(),
		})
		return
	}

	// 调用 AnythingLLM 获取历史对话 API
	// 注意：AnythingLLM 可能需要特定的 API 端点来获取历史
	url := fmt.Sprintf("%s/api/v1/workspace/%s/chats", h.config.AnythingLLMURL, slug)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+h.config.AnythingLLMKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch from anythingllm"})
		return
	}
	defer resp.Body.Close()

	var history []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}

	// 同步到本地数据库（可选）
	// 这里可以根据需要实现具体的同步逻辑

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"history": history,
		},
	})
}

// DeleteMessage 删除指定消息
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")
	messageId := c.Param("msgId")

	// 验证会话属于用户
	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 删除本地消息
	var msg models.Message
	if result := h.db.Where("id = ? AND session_id = ?", messageId, sessionId).First(&msg); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	if result := h.db.Delete(&msg); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete message"})
		return
	}

	// 同步删除到 AnythingLLM
	// 注意：AnythingLLM 不支持删除单个消息，只能清空整个聊天历史
	// 如果删除的是最后一条消息，可以选择清空 AnythingLLM 的聊天历史
	if session.AnythingLLMSlug != "" {
		// 检查是否还有其他消息
		var remainingCount int64
		h.db.Model(&models.Message{}).Where("session_id = ?", sessionId).Count(&remainingCount)
		
		// 如果没有其他消息了，清空 AnythingLLM 的聊天历史
		if remainingCount == 0 {
			// 使用 AnythingLLM client 删除聊天历史
			anythingLLMClient := anythingllm.NewAnythingLLMClient(h.config.AnythingLLMURL, h.config.AnythingLLMKey)
			if err := anythingLLMClient.DeleteChatHistory(userId.(string)); err != nil {
				// 记录错误但不影响本地删除结果
				// 可以在日志中记录：log.Printf("Failed to delete AnythingLLM chat history: %v", err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"deleted": true,
		},
	})
}

// SwitchRoleRequest 切换角色请求
type SwitchRoleRequest struct {
	RoleID string `json:"roleId" binding:"required"`
}

// SwitchRole 切换会话角色
// @Summary 切换会话角色
// @Description 在对话中切换到另一个角色，保持对话历史
// @Tags 对话
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "会话 ID"
// @Param request body SwitchRoleRequest true "角色 ID"
// @Success 200 {object} map[string]interface{} "切换成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 404 {object} map[string]string "会话或角色不存在"
// @Router /api/v1/chat-sessions/{id}/switch-role [post]
func (h *ChatHandler) SwitchRole(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var req SwitchRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证会话属于用户
	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 验证新角色存在
	var role models.Role
	if result := h.db.First(&role, "id = ?", req.RoleID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// 更新会话的角色 ID
	oldRoleID := session.RoleID
	session.RoleID = req.RoleID
	session.UpdatedAt = time.Now()

	if result := h.db.Save(&session); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to switch role"})
		return
	}

	// 添加系统消息记录角色切换
	systemMsg := models.Message{
		ID:        models.NewUUID(),
		SessionID: session.ID,
		Role:      "system",
		Content:   "角色已切换",
		CreatedAt: time.Now(),
	}
	h.db.Create(&systemMsg)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"sessionId":   session.ID,
			"oldRoleId":   oldRoleID,
			"newRoleId":   req.RoleID,
			"newRoleName": role.Name,
		},
	})
}

// WorkspaceAuth Workspace 认证中间件
func (h *ChatHandler) WorkspaceAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		sessionId := c.Param("id")
		if sessionId == "" {
			c.Next()
			return
		}

		// 验证会话属于当前用户
		var session models.ChatSession
		if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this workspace"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UpdateSessionTitleRequest 更新会话标题请求
type UpdateSessionTitleRequest struct {
	Title string `json:"title" binding:"required"`
}

// UpdateSessionTitle 更新会话标题
func (h *ChatHandler) UpdateSessionTitle(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var req UpdateSessionTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	session.Title = req.Title
	session.UpdatedAt = time.Now()

	if result := h.db.Save(&session); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update title"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"sessionId": session.ID,
			"title":     session.Title,
		},
	})
}

// ArchiveSessionRequest 归档会话请求
type ArchiveSessionRequest struct {
	IsArchived bool `json:"isArchived"`
}

// ArchiveSession 归档/取消归档会话
func (h *ChatHandler) ArchiveSession(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var req ArchiveSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 使用 ModelConfig 字段存储归档状态（临时方案）
	var config map[string]interface{}
	if session.ModelConfig != "" {
		json.Unmarshal([]byte(session.ModelConfig), &config)
	}
	if config == nil {
		config = make(map[string]interface{})
	}
	config["isArchived"] = req.IsArchived

	configJSON, _ := json.Marshal(config)
	session.ModelConfig = models.JSON(configJSON)
	session.UpdatedAt = time.Now()

	if result := h.db.Save(&session); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to archive session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"sessionId":  session.ID,
			"isArchived": req.IsArchived,
		},
	})
}

// ExportSessionRequest 导出会话请求
type ExportSessionRequest struct {
	Format string `json:"format" binding:"required"` // markdown, pdf, json
}

// ExportSession 导出会话
func (h *ChatHandler) ExportSession(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var req ExportSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var messages []models.Message
	if result := h.db.Where("session_id = ?", sessionId).Order("created_at ASC").Find(&messages); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load messages"})
		return
	}

	var exportContent string
	var contentType string
	var filename string

	switch req.Format {
	case "markdown":
		exportContent = h.exportToMarkdown(session, messages)
		contentType = "text/markdown"
		filename = fmt.Sprintf("%s.md", session.Title)
	case "json":
		exportContent = h.exportToJSON(session, messages)
		contentType = "application/json"
		filename = fmt.Sprintf("%s.json", session.Title)
	case "pdf":
		// PDF 导出需要额外库，暂时返回 markdown 格式
		exportContent = h.exportToMarkdown(session, messages)
		contentType = "text/markdown"
		filename = fmt.Sprintf("%s.md", session.Title)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported format"})
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, contentType, []byte(exportContent))
}

// exportToMarkdown 导出为 Markdown 格式
func (h *ChatHandler) exportToMarkdown(session models.ChatSession, messages []models.Message) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", session.Title))
	sb.WriteString(fmt.Sprintf("**角色**: %s\n", session.RoleID))
	sb.WriteString(fmt.Sprintf("**创建时间**: %s\n\n", session.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString("---\n\n")

	for _, msg := range messages {
		if msg.Role == "system" {
			continue
		}

		roleName := "用户"
		if msg.Role == "assistant" {
			roleName = "AI"
		}

		sb.WriteString(fmt.Sprintf("### %s - %s\n\n", roleName, msg.CreatedAt.Format("15:04:05")))
		sb.WriteString(msg.Content)
		sb.WriteString("\n\n---\n\n")
	}

	return sb.String()
}

// exportToJSON 导出为 JSON 格式
func (h *ChatHandler) exportToJSON(session models.ChatSession, messages []models.Message) string {
	data := map[string]interface{}{
		"session": session,
		"messages": messages,
	}
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonData)
}

// UpdateMessageRequest 更新消息请求
type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// UpdateMessage 更新消息（编辑用户消息）
func (h *ChatHandler) UpdateMessage(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")
	messageId := c.Param("msgId")

	var req UpdateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证会话属于用户
	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 更新消息
	var msg models.Message
	if result := h.db.Where("id = ? AND session_id = ?", messageId, sessionId).First(&msg); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	// 只允许编辑用户自己的消息
	if msg.Role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can only edit user messages"})
		return
	}

	msg.Content = req.Content
	if result := h.db.Save(&msg); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"messageId": msg.ID,
			"content":   msg.Content,
		},
	})
}

// RegenerateMessageRequest 重新生成请求
type RegenerateMessageRequest struct {
	Content string `json:"content"`
}

// RegenerateMessage 重新生成 AI 回复
func (h *ChatHandler) RegenerateMessage(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")
	messageId := c.Param("msgId")

	// 验证会话属于用户
	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 找到要重新生成的消息
	var msg models.Message
	if result := h.db.Where("id = ? AND session_id = ?", messageId, sessionId).First(&msg); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	// 找到最后一条用户消息
	var lastUserMsg models.Message
	if result := h.db.Where("session_id = ? AND role = 'user' AND created_at <= ?", sessionId, msg.CreatedAt).
		Order("created_at DESC").First(&lastUserMsg); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no user message found"})
		return
	}

	content := lastUserMsg.Content

	// 获取 AnythingLLM Slug
	var assistantContent string
	slug, err := h.getAnythingLLMSlug(session)
	if err != nil {
		assistantContent = h.callMockAI(content)
	} else {
		assistantContent, err = h.callAnythingLLM(slug, content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "anythingllm API error: " + err.Error(),
			})
			return
		}
	}

	// 更新或创建新的助手消息
	msg.Content = assistantContent
	msg.CreatedAt = time.Now()
	h.db.Save(&msg)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"assistantMessage": msg,
		},
	})
}

// RateMessageRequest 评分请求
type RateMessageRequest struct {
	Rating string `json:"rating" binding:"required"` // up, down
}

// RateMessage 对消息评分（点赞/点踩）
func (h *ChatHandler) RateMessage(c *gin.Context) {
	userId, _ := c.Get("userId")
	messageId := c.Param("msgId")

	var req RateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Rating != "up" && req.Rating != "down" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be 'up' or 'down'"})
		return
	}

	// 验证消息存在
	var msg models.Message
	if result := h.db.Where("id = ?", messageId).First(&msg); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	// 验证会话属于用户
	var session models.ChatSession
	if result := h.db.Where("id = ?", msg.SessionID).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	if session.UserID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// 存储评分到元数据（临时方案，使用 Sources 字段）
	var metadata map[string]interface{}
	if msg.Sources != "" {
		json.Unmarshal([]byte(msg.Sources), &metadata)
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["rating"] = req.Rating
	metadata["ratedBy"] = userId
	metadata["ratedAt"] = time.Now().Unix()

	metadataJSON, _ := json.Marshal(metadata)
	msg.Sources = models.JSON(metadataJSON)
	h.db.Save(&msg)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"messageId": messageId,
			"rating":    req.Rating,
		},
	})
}

// SearchSessionsRequest 搜索会话请求
type SearchSessionsRequest struct {
	Query string `json:"query" binding:"required"`
}

// SearchSessions 搜索会话（基于消息内容）
func (h *ChatHandler) SearchSessions(c *gin.Context) {
	userId, _ := c.Get("userId")

	var req SearchSessionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 搜索包含关键词的会话
	var sessions []models.ChatSession
	query := "%" + req.Query + "%"
	
	// 通过消息内容搜索会话
	var messageSessionIDs []string
	h.db.Table("messages").
		Where("session_id IN (SELECT id FROM chat_sessions WHERE user_id = ?)", userId).
		Where("content LIKE ?", query).
		Distinct("session_id").
		Pluck("session_id", &messageSessionIDs)

	// 通过标题搜索会话
	h.db.Where("user_id = ? AND title LIKE ?", userId, query).
		Or("id IN ?", messageSessionIDs).
		Order("updated_at DESC").
		Find(&sessions)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    sessions,
	})
}

// ChatStreamWithThinking 发送消息（流式响应 SSE + 深度思考过程）
func (h *ChatHandler) ChatStreamWithThinking(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 保存用户消息
	userMsg := models.Message{
		ID:        models.NewUUID(),
		SessionID: session.ID,
		Role:      "user",
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
	h.db.Create(&userMsg)

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// 创建思考过程发送器
	sender := h.thinkingSvc.NewStreamThinkingSender(func(chunk thinking.StreamChunk) {
		jsonData, _ := json.Marshal(chunk)
		fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
		flusher.Flush()
	})

	// 流式推送思考步骤
	// 步骤 1: 理解问题
	sender.AddThinkingStep(thinking.ThinkingUnderstand, "理解问题："+truncateString(req.Content, 50))
	
	// 步骤 2: 分析要素
	sender.AddThinkingStep(thinking.ThinkingAnalyze, "分析关键要素和约束条件")
	
	// 步骤 3: 检索知识
	sender.AddThinkingStep(thinking.ThinkingSearch, "从知识库检索相关信息")
	
	// 步骤 4: 组织答案
	sender.AddThinkingStep(thinking.ThinkingOrganize, "组织回答结构和逻辑")
	
	// 获取实际响应（从 AnythingLLM 或 Mock AI）
	var assistantContent string
	slug, err := h.getAnythingLLMSlug(session)
	if err != nil {
		// 使用 Mock AI
		assistantContent = h.callMockAI(req.Content)
	} else {
		// 调用 AnythingLLM API
		assistantContent, err = h.callAnythingLLM(slug, req.Content)
		if err != nil {
			// 发送错误
			jsonData, _ := json.Marshal(map[string]interface{}{
				"type": "error",
				"data": map[string]string{"message": err.Error()},
			})
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
			flusher.Flush()
			return
		}
	}
	
	// 步骤 5: 得出结论
	sender.AddThinkingStep(thinking.ThinkingConclude, "综合以上分析得出结论")
	
	// 完成思考过程
	sender.Complete()
	
	// 发送最终答案
	sender.SendAnswer(assistantContent)
	
	// 发送完成标记
	jsonData, _ := json.Marshal(thinking.StreamChunk{
		Type: "done",
		Done: true,
	})
	fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
	flusher.Flush()

	// 保存完整的助手消息
	if assistantContent != "" {
		assistantMsg := models.Message{
			ID:        models.NewUUID(),
			SessionID: session.ID,
			Role:      "assistant",
			Content:   assistantContent,
			CreatedAt: time.Now(),
		}
		h.db.Create(&assistantMsg)
		h.db.Model(&session).Update("updated_at", time.Now())
	}
}

// callMockAI 调用 Mock AI（当 AnythingLLM 未配置时使用）
func (h *ChatHandler) callMockAI(message string) string {
	// 简单的 Mock AI 回复
	return "你好！我是 Mock AI 助手。我现在可以回答你的问题：" + message
}

// truncateString 截断字符串（辅助函数）
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
