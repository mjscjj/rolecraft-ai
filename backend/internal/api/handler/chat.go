package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/models"
	"rolecraft-ai/internal/service/ai"
)

// ChatHandler 对话处理器
type ChatHandler struct {
	db     *gorm.DB
	config *config.Config
	openai *ai.OpenAIClient
}

// NewChatHandler 创建对话处理器
func NewChatHandler(db *gorm.DB, cfg *config.Config) *ChatHandler {
	var openaiClient *ai.OpenAIClient
	if cfg.OpenAIKey != "" {
		openaiClient = ai.NewOpenAIClient(ai.OpenAIConfig{
			APIKey: cfg.OpenAIKey,
			Model:  "gpt-4",
		})
	}
	return &ChatHandler{
		db:     db,
		config: cfg,
		openai: openaiClient,
	}
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	RoleID string `json:"roleId" binding:"required"`
	Title  string `json:"title"`
	Mode   string `json:"mode"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Content     string   `json:"content" binding:"required"`
	Attachments []string `json:"attachments"`
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
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).Preload("Role").First(&session); result.Error != nil {
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

// Chat 发送消息（普通响应）
func (h *ChatHandler) Chat(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).Preload("Role").First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	userMsg := models.Message{
		ID:        models.NewUUID(),
		SessionID: session.ID,
		Role:      "user",
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
	h.db.Create(&userMsg)

	var history []models.Message
	h.db.Where("session_id = ?", sessionId).Order("created_at ASC").Limit(20).Find(&history)

	var role models.Role
	if session.Role != nil {
		role = *session.Role
	} else {
		h.db.First(&role, "id = ?", session.RoleID)
	}

	messages := []ai.ChatMessage{
		{Role: "system", Content: role.SystemPrompt},
	}
	for _, msg := range history {
		messages = append(messages, ai.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	var assistantContent string
	if h.openai != nil {
		resp, err := h.openai.ChatCompletion(c.Request.Context(), messages, 0.7)
		if err != nil {
			assistantContent = fmt.Sprintf("AI 服务错误: %v", err)
		} else if len(resp.Choices) > 0 {
			assistantContent = resp.Choices[0].Message.Content
		}
	} else {
		assistantContent = fmt.Sprintf("收到: \"%s\"。（OpenAI API Key 未配置）", req.Content)
	}

	assistantMsg := models.Message{
		ID:        models.NewUUID(),
		SessionID: session.ID,
		Role:      "assistant",
		Content:   assistantContent,
		CreatedAt: time.Now(),
	}
	h.db.Create(&assistantMsg)

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

// ChatStream 发送消息（流式响应 SSE）
func (h *ChatHandler) ChatStream(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).Preload("Role").First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	userMsg := models.Message{
		ID:        models.NewUUID(),
		SessionID: session.ID,
		Role:      "user",
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
	h.db.Create(&userMsg)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	var history []models.Message
	h.db.Where("session_id = ?", sessionId).Order("created_at ASC").Limit(20).Find(&history)

	var role models.Role
	if session.Role != nil {
		role = *session.Role
	} else {
		h.db.First(&role, "id = ?", session.RoleID)
	}

	messages := []ai.ChatMessage{
		{Role: "system", Content: role.SystemPrompt},
	}
	for _, msg := range history {
		messages = append(messages, ai.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	var fullContent string

	if h.openai != nil {
		chunkChan, errChan := h.openai.ChatCompletionStream(c.Request.Context(), messages, 0.7)

		for {
			select {
			case chunk, ok := <-chunkChan:
				if !ok {
					goto done
				}
				if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
					content := chunk.Choices[0].Delta.Content
					fullContent += content
					data := map[string]interface{}{"content": content, "done": false}
					jsonData, _ := json.Marshal(data)
					fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
					flusher.Flush()
				}
			case err := <-errChan:
				if err != nil {
					data := map[string]interface{}{"error": err.Error(), "done": true}
					jsonData, _ := json.Marshal(data)
					fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
					flusher.Flush()
				}
				goto done
			case <-c.Request.Context().Done():
				goto done
			}
		}
	done:
	} else {
		mockResponse := "OpenAI API Key 未配置。请设置 OPENAI_API_KEY 环境变量。"
		for i, char := range mockResponse {
			fullContent += string(char)
			data := map[string]interface{}{"content": string(char), "done": i == len(mockResponse)-1}
			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
			flusher.Flush()
			time.Sleep(30 * time.Millisecond)
		}
	}

	data := map[string]interface{}{"done": true}
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
	flusher.Flush()

	if fullContent != "" {
		assistantMsg := models.Message{
			ID:        models.NewUUID(),
			SessionID: session.ID,
			Role:      "assistant",
			Content:   fullContent,
			CreatedAt: time.Now(),
		}
		h.db.Create(&assistantMsg)
		h.db.Model(&session).Update("updated_at", time.Now())
	}
}
