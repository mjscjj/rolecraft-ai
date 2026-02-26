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
)

// ChatHandler 对话处理器
type ChatHandler struct {
	db     *gorm.DB
	config *config.Config
}

// NewChatHandler 创建对话处理器
func NewChatHandler(db *gorm.DB, cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		db:     db,
		config: cfg,
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

// callMockAI 调用 Mock AI（当 AnythingLLM 未配置时使用）
func (h *ChatHandler) callMockAI(message string) string {
	// 简单的 Mock AI 回复
	return "你好！我是 Mock AI 助手。我现在可以回答你的问题：" + message
}
