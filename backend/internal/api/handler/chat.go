package handler

import (
	"context"
	"encoding/json"
	"fmt"
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
	db          *gorm.DB
	config      *config.Config
	thinkingSvc *thinking.Service
	anything    *anythingllm.Orchestrator
}

// NewChatHandler 创建对话处理器
func NewChatHandler(db *gorm.DB, cfg *config.Config) *ChatHandler {
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &ChatHandler{
		db:          db,
		config:      cfg,
		thinkingSvc: thinking.NewService(),
		anything: anythingllm.NewOrchestrator(cfg.AnythingLLMURL, cfg.AnythingLLMKey, anythingllm.OrchestratorConfig{
			DefaultProvider: "openrouter",
			DefaultModel:    cfg.OpenRouterModel,
			OpenRouterKey:   cfg.OpenRouterKey,
		}),
	}
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	RoleID          string                 `json:"roleId" binding:"required"`
	Title           string                 `json:"title"`
	Mode            string                 `json:"mode"`
	AnythingLLMSlug string                 `json:"anythingLLMSlug"` // AnythingLLM Workspace Slug
	ModelConfig     map[string]interface{} `json:"modelConfig"`
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
	Response     string                   `json:"response"`
	TextResponse string                   `json:"textResponse"`
	Error        string                   `json:"error"`
	Thoughts     []string                 `json:"thoughts"`
	Sources      []map[string]interface{} `json:"sources"`
	Type         string                   `json:"type"`
}

type ChatMode string

const (
	ChatModeChat  ChatMode = "chat"
	ChatModeAgent ChatMode = "agent"
)

type AnythingLLMResult struct {
	Content  string
	Sources  []map[string]interface{}
	Thoughts []string
	Type     string
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
	userIDStr, _ := userId.(string)

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if result := h.db.First(&role, "id = ? AND user_id = ?", req.RoleID, userIDStr); result.Error != nil {
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
		UserID:    userIDStr,
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
	if req.ModelConfig != nil {
		configJSON, _ := json.Marshal(req.ModelConfig)
		session.ModelConfig = models.JSON(configJSON)
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

// DeleteSession 删除会话及其消息
func (h *ChatHandler) DeleteSession(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", sessionId).Delete(&models.Message{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.ChatSession{}, "id = ? AND user_id = ?", sessionId, userId).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"deleted": true,
		},
	})
}

func normalizeUserWorkspaceSlug(userID string) string {
	return anythingllm.UserWorkspaceSlug(userID)
}

// getAnythingLLMSlug 获取会话的 AnythingLLM Slug
func (h *ChatHandler) getAnythingLLMSlug(userID string, session models.ChatSession) (string, error) {
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

	// 最后回退到按用户自动生成的 workspace slug（与 youmind 对齐）
	return normalizeUserWorkspaceSlug(userID), nil
}

func (h *ChatHandler) ensureAnythingLLMWorkspace(userID string, session *models.ChatSession) (string, error) {
	slug, err := h.getAnythingLLMSlug(userID, *session)
	if err != nil {
		return "", err
	}

	if h.anything == nil || !h.anything.Enabled() {
		return "", fmt.Errorf("anythingllm is not configured")
	}
	ws, err := h.anything.EnsureWorkspaceBySlug(context.Background(), slug, slug, "")
	if err != nil {
		return "", fmt.Errorf("failed to ensure workspace: %w", err)
	}
	if ws != nil && strings.TrimSpace(ws.Slug) != "" {
		slug = strings.TrimSpace(ws.Slug)
	}

	// 回写 session，避免重复推断
	if session.AnythingLLMSlug != slug {
		session.AnythingLLMSlug = slug
		_ = h.db.Save(session).Error
	}

	return slug, nil
}

func (h *ChatHandler) resolveRuntimeProvider(session models.ChatSession) string {
	cfg := h.parseSessionModelConfig(session)
	for _, key := range []string{"provider", "chatProvider"} {
		if value, ok := cfg[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return "openrouter"
}

func (h *ChatHandler) resolveRuntimeModel(session models.ChatSession) string {
	cfg := h.parseSessionModelConfig(session)
	for _, key := range []string{"model", "modelId", "chatModel"} {
		if value, ok := cfg[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	if strings.TrimSpace(h.config.OpenRouterModel) != "" {
		return strings.TrimSpace(h.config.OpenRouterModel)
	}
	return ""
}

// callAnythingLLM 调用 AnythingLLM Chat API
func (h *ChatHandler) callAnythingLLMWithMode(session models.ChatSession, slug, message string, mode ChatMode, sessionID string) (*AnythingLLMResult, error) {
	if h.anything == nil || !h.anything.Enabled() {
		return nil, fmt.Errorf("anythingllm is not configured")
	}
	modeValue := "chat"
	if mode == ChatModeAgent {
		modeValue = "agent"
	}
	result, err := h.anything.Chat(context.Background(), anythingllm.ChatPayload{
		WorkspaceSlug: slug,
		Message:       message,
		Mode:          modeValue,
		Model:         h.resolveRuntimeModel(session),
		Provider:      h.resolveRuntimeProvider(session),
		SessionID:     sessionID,
	})
	if err != nil {
		return nil, err
	}
	return &AnythingLLMResult{
		Content:  result.Content,
		Sources:  result.Sources,
		Thoughts: result.Thoughts,
		Type:     result.Type,
	}, nil
}

func (h *ChatHandler) parseSessionModelConfig(session models.ChatSession) map[string]interface{} {
	if session.ModelConfig == "" {
		return map[string]interface{}{}
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(session.ModelConfig), &cfg); err != nil || cfg == nil {
		return map[string]interface{}{}
	}
	return cfg
}

func (h *ChatHandler) resolveAnythingLLMKey(session models.ChatSession) string {
	// 优先使用服务端系统 Key，避免前端历史 custom key 失效导致对话不可用。
	if strings.TrimSpace(h.config.AnythingLLMKey) != "" {
		return h.config.AnythingLLMKey
	}
	cfg := h.parseSessionModelConfig(session)
	if k, ok := cfg["customAPIKey"].(string); ok && strings.TrimSpace(k) != "" {
		return strings.TrimSpace(k)
	}
	return ""
}

func (h *ChatHandler) resolveChatMode(session models.ChatSession) ChatMode {
	cfg := h.parseSessionModelConfig(session)
	chatMode, _ := cfg["chatMode"].(string)
	if chatMode == "deep" || strings.EqualFold(chatMode, "agent") {
		return ChatModeAgent
	}
	return ChatModeChat
}

func buildAssistantSources(mode ChatMode, result *AnythingLLMResult) models.JSON {
	sources := []map[string]interface{}{}
	if result != nil && len(result.Sources) > 0 {
		sources = result.Sources
	}
	thoughts := []string{}
	if result != nil && len(result.Thoughts) > 0 {
		thoughts = result.Thoughts
	}
	payload := map[string]interface{}{
		"mode":     string(mode),
		"sources":  sources,
		"thoughts": thoughts,
	}
	return models.ToJSON(payload)
}

func (h *ChatHandler) buildKnowledgeContext(userID string, session models.ChatSession) string {
	cfg := h.parseSessionModelConfig(session)
	scope, _ := cfg["knowledgeScope"].(string)
	if scope == "" || scope == "none" {
		return ""
	}

	query := h.db.Model(&models.Document{}).Where("user_id = ? AND status = ?", userID, "completed")
	if strings.HasPrefix(scope, "folder:") {
		folderID := strings.TrimPrefix(scope, "folder:")
		if folderID != "" && folderID != "default" {
			query = query.Where("folder_id = ?", folderID)
		}
	}

	var docs []models.Document
	if err := query.Order("updated_at DESC").Limit(8).Find(&docs).Error; err != nil || len(docs) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("可参考知识库文档：\n")
	for i, d := range docs {
		b.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, d.Name, d.FileType))
	}
	return b.String()
}

func clipText(value string, limit int) string {
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
}

func (h *ChatHandler) buildAttachmentContext(userID string, attachments []string) string {
	if len(attachments) == 0 {
		return ""
	}

	var docs []models.Document
	if err := h.db.
		Where("id IN ? AND user_id = ? AND status = ?", attachments, userID, "completed").
		Find(&docs).Error; err != nil || len(docs) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("本轮用户上传并指定参考的文档内容：\n")

	for i, doc := range docs {
		content, err := (&DocumentHandler{}).extractPreviewContent(doc)
		if err != nil {
			b.WriteString(fmt.Sprintf("%d. %s (%s)\n内容提取失败，优先按文档标题和已有知识检索回答。\n\n", i+1, doc.Name, doc.FileType))
			continue
		}

		trimmed := strings.TrimSpace(content)
		if trimmed == "" {
			b.WriteString(fmt.Sprintf("%d. %s (%s)\n文档内容为空。\n\n", i+1, doc.Name, doc.FileType))
			continue
		}

		b.WriteString(fmt.Sprintf("%d. %s (%s)\n%s\n\n", i+1, doc.Name, doc.FileType, clipText(trimmed, 2000)))
	}

	return b.String()
}

func (h *ChatHandler) buildComposedMessage(userID string, session models.ChatSession, userMessage string, attachments []string) string {
	var role models.Role
	rolePrompt := ""
	if err := h.db.Where("id = ? AND user_id = ?", session.RoleID, userID).First(&role).Error; err == nil {
		rolePrompt = role.SystemPrompt
	}
	cfg := h.parseSessionModelConfig(session)
	chatMode, _ := cfg["chatMode"].(string)
	kbContext := h.buildKnowledgeContext(userID, session)
	attachmentContext := h.buildAttachmentContext(userID, attachments)

	// Deep mode: use AnythingLLM agent invocation so web-browsing skill can be called.
	if chatMode == "deep" {
		var b strings.Builder
		b.WriteString("@agent ")
		if rolePrompt != "" {
			b.WriteString("角色设定：")
			b.WriteString(rolePrompt)
			b.WriteString("\n")
		}
		if kbContext != "" {
			b.WriteString(kbContext)
			b.WriteString("\n")
		}
		if attachmentContext != "" {
			b.WriteString(attachmentContext)
			b.WriteString("\n")
		}
		b.WriteString("请优先联网搜索最新信息，并给出可点击来源链接。\n")
		b.WriteString("用户问题：")
		b.WriteString(userMessage)
		return b.String()
	}

	if rolePrompt == "" && kbContext == "" && attachmentContext == "" {
		return userMessage
	}

	var b strings.Builder
	if rolePrompt != "" {
		b.WriteString("角色设定：\n")
		b.WriteString(rolePrompt)
		b.WriteString("\n\n")
	}
	if kbContext != "" {
		b.WriteString(kbContext)
		b.WriteString("\n")
	}
	if attachmentContext != "" {
		b.WriteString(attachmentContext)
		b.WriteString("\n")
	}
	b.WriteString("请遵循角色设定，并优先利用知识库信息回答。\n\n")
	b.WriteString("用户问题：\n")
	b.WriteString(userMessage)
	return b.String()
}

// Chat 发送消息（普通响应）- 集成 AnythingLLM
func (h *ChatHandler) Chat(c *gin.Context) {
	userId, _ := c.Get("userId")
	userIDStr, _ := userId.(string)
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
	composedMessage := h.buildComposedMessage(userIDStr, session, req.Content, req.Attachments)
	slug, err := h.ensureAnythingLLMWorkspace(userIDStr, &session)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "anythingllm workspace not configured: " + err.Error(),
		})
		return
	}

	// 调用 AnythingLLM API
	mode := h.resolveChatMode(session)
	aiResult, err := h.callAnythingLLMWithMode(session, slug, composedMessage, mode, session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "anythingllm API error: " + err.Error(),
		})
		return
	}
	assistantContent = aiResult.Content

	// 保存助手消息
	assistantMsg := models.Message{
		ID:        models.NewUUID(),
		SessionID: session.ID,
		Role:      "assistant",
		Content:   assistantContent,
		Sources:   buildAssistantSources(mode, aiResult),
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
	userIDStr, _ := userId.(string)
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
	slug, err := h.ensureAnythingLLMWorkspace(userIDStr, &session)
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

	// 为保证稳定，服务端统一调用 chat API，再以 SSE 输出给前端。
	mode := h.resolveChatMode(session)
	aiResult, err := h.callAnythingLLMWithMode(
		session,
		slug,
		h.buildComposedMessage(userIDStr, session, req.Content, req.Attachments),
		mode,
		session.ID,
	)
	if err != nil {
		data := map[string]interface{}{"error": err.Error(), "done": true}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
		flusher.Flush()
		return
	}
	assistantContent := aiResult.Content

	var fullContent strings.Builder
	fullContent.WriteString(assistantContent)
	data := map[string]interface{}{"content": assistantContent, "done": false}
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
	flusher.Flush()

	// 保存完整的助手消息
	assistantMessageID := ""
	if fullContent.Len() > 0 {
		assistantMsg := models.Message{
			ID:        models.NewUUID(),
			SessionID: session.ID,
			Role:      "assistant",
			Content:   fullContent.String(),
			Sources:   buildAssistantSources(mode, aiResult),
			CreatedAt: time.Now(),
		}
		h.db.Create(&assistantMsg)
		h.db.Model(&session).Update("updated_at", time.Now())
		assistantMessageID = assistantMsg.ID
	}

	// 统一由服务端发送最终 done 事件，并附带真实 message ID
	doneData := map[string]interface{}{
		"done": true,
	}
	if assistantMessageID != "" {
		doneData["assistantMessageId"] = assistantMessageID
	}
	doneData["meta"] = map[string]interface{}{
		"mode":     string(mode),
		"sources":  aiResult.Sources,
		"thoughts": aiResult.Thoughts,
	}
	doneJSON, _ := json.Marshal(doneData)
	fmt.Fprintf(c.Writer, "data: %s\n\n", doneJSON)
	flusher.Flush()
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

	userIDStr, _ := userId.(string)
	slug, err := h.ensureAnythingLLMWorkspace(userIDStr, &session)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "anythingllm workspace not configured: " + err.Error(),
		})
		return
	}

	history, err := h.anything.GetChatHistory(context.Background(), slug, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch from anythingllm: " + err.Error()})
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
	if h.anything != nil && h.anything.Enabled() {
		// 检查是否还有其他消息
		var remainingCount int64
		h.db.Model(&models.Message{}).Where("session_id = ?", sessionId).Count(&remainingCount)

		// 如果没有其他消息了，清空 AnythingLLM 的聊天历史
		if remainingCount == 0 {
			slug := session.AnythingLLMSlug
			if strings.TrimSpace(slug) == "" {
				if uid, ok := userId.(string); ok {
					slug = normalizeUserWorkspaceSlug(uid)
				}
			}
			_ = h.anything.DeleteChatHistory(context.Background(), slug)
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
	userIDStr, _ := userId.(string)
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
	if result := h.db.First(&role, "id = ? AND user_id = ?", req.RoleID, userIDStr); result.Error != nil {
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

type UpdateSessionConfigRequest struct {
	ModelConfig map[string]interface{} `json:"modelConfig" binding:"required"`
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

// UpdateSessionConfig 更新会话配置（模型/温度/知识范围等）
func (h *ChatHandler) UpdateSessionConfig(c *gin.Context) {
	userId, _ := c.Get("userId")
	sessionId := c.Param("id")

	var req UpdateSessionConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userId).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	configJSON, _ := json.Marshal(req.ModelConfig)
	session.ModelConfig = models.JSON(configJSON)
	session.UpdatedAt = time.Now()

	if result := h.db.Save(&session); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"sessionId":   session.ID,
			"modelConfig": req.ModelConfig,
		},
	})
}

// ArchiveSessionRequest 归档会话请求
type ArchiveSessionRequest struct {
	IsArchived bool `json:"isArchived"`
}

// ArchiveSession 归档/取消归档会话
func (h *ChatHandler) ArchiveSession(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		if v, ok := c.Get("userId"); ok {
			userID = fmt.Sprint(v)
		}
	}
	sessionId := c.Param("id")

	var req ArchiveSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userID).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 使用 ModelConfig 字段存储归档状态
	var config map[string]interface{}
	if session.ModelConfig != "" {
		_ = json.Unmarshal([]byte(session.ModelConfig), &config)
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

// FeedbackRequest 反馈请求
type FeedbackRequest struct {
	Type string `json:"type" binding:"required,oneof=like dislike"`
}

// AddFeedback 添加反馈
func (h *ChatHandler) AddFeedback(c *gin.Context) {
	messageId := c.Param("id")
	if messageId == "" {
		messageId = c.Param("msgId")
	}

	userID := c.GetString("userId")
	if userID == "" {
		if v, ok := c.Get("userId"); ok {
			userID = fmt.Sprint(v)
		}
	}

	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}

	// 验证消息所有权
	var msg models.Message
	if err := h.db.Where("id = ?", messageId).First(&msg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "消息不存在"})
		return
	}

	var session models.ChatSession
	if err := h.db.Where("id = ? AND user_id = ?", msg.SessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权操作此消息"})
		return
	}

	// 更新反馈计数
	updates := make(map[string]interface{})
	if req.Type == "like" {
		updates["likes"] = gorm.Expr("COALESCE(likes, 0) + 1")
	} else {
		updates["dislikes"] = gorm.Expr("COALESCE(dislikes, 0) + 1")
	}

	if err := h.db.Model(&msg).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "反馈失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "反馈成功"})
}

// ExportSessionRequest 导出会话请求
type ExportSessionRequest struct {
	Format string `json:"format" binding:"required"` // markdown, pdf, json
}

// ExportSession 导出会话
func (h *ChatHandler) ExportSession(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		if v, ok := c.Get("userId"); ok {
			userID = fmt.Sprint(v)
		}
	}
	sessionId := c.Param("id")

	format := "markdown"
	if c.Request.Method == http.MethodGet {
		format = c.DefaultQuery("format", "md")
	} else {
		var req ExportSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		format = req.Format
	}

	switch strings.ToLower(format) {
	case "md":
		format = "markdown"
	case "json":
		format = "json"
	case "pdf":
		format = "pdf"
	default:
		format = "markdown"
	}

	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userID).First(&session); result.Error != nil {
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

	switch format {
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

	// 兼容旧版 GET 导出响应格式（用于历史客户端和旧测试）
	if c.Request.Method == http.MethodGet {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "导出成功",
			"data":    exportContent,
		})
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
		"session":  session,
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
	msg.IsEdited = true
	msg.UpdatedAt = time.Now()
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
	userID := c.GetString("userId")
	if userID == "" {
		if v, ok := c.Get("userId"); ok {
			userID = fmt.Sprint(v)
		}
	}
	sessionId := c.Param("id")
	messageId := c.Param("msgId")

	// 验证会话属于用户
	var session models.ChatSession
	if result := h.db.Where("id = ? AND user_id = ?", sessionId, userID).First(&session); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 找到要重新生成的消息
	var msg models.Message
	if result := h.db.Where("id = ? AND session_id = ?", messageId, sessionId).First(&msg); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}
	if msg.Role != "assistant" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can only regenerate assistant messages"})
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
	var aiResult *AnythingLLMResult
	slug, err := h.ensureAnythingLLMWorkspace(userID, &session)
	if err != nil {
		assistantContent = h.callMockAI(content)
	} else {
		composed := h.buildComposedMessage(userID, session, content, nil)
		mode := h.resolveChatMode(session)
		aiResult, err = h.callAnythingLLMWithMode(session, slug, composed, mode, session.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "anythingllm API error: " + err.Error(),
			})
			return
		}
		assistantContent = aiResult.Content
	}

	// 更新或创建新的助手消息
	msg.Content = assistantContent
	msg.Sources = buildAssistantSources(h.resolveChatMode(session), aiResult)
	msg.CreatedAt = time.Now()
	if err := h.db.Save(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save regenerated message"})
		return
	}

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
	userID := c.GetString("userId")
	if userID == "" {
		if v, ok := c.Get("userId"); ok {
			userID = fmt.Sprint(v)
		}
	}
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

	if session.UserID != userID {
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
	metadata["ratedBy"] = userID
	metadata["ratedAt"] = time.Now().Unix()

	if req.Rating == "up" {
		msg.Likes++
	} else {
		msg.Dislikes++
	}

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

	// 获取实际响应（stream-with-thinking endpoint 固定按 agent 模式执行，避免会话配置落库延迟导致模式漂移）
	var assistantContent string
	var aiResult *AnythingLLMResult
	mode := ChatModeAgent
	userIDStr, _ := userId.(string)
	slug, err := h.ensureAnythingLLMWorkspace(userIDStr, &session)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{
			"type": "error",
			"data": map[string]string{"message": "anythingllm workspace not configured: " + err.Error()},
		})
		fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
		flusher.Flush()
		return
	} else {
		// 调用 AnythingLLM API
		composed := h.buildComposedMessage(userIDStr, session, req.Content, req.Attachments)
		composed = strings.TrimSpace(composed)
		if !strings.HasPrefix(composed, "@agent") {
			composed = "@agent " + composed
		}
		aiResult, err = h.callAnythingLLMWithMode(session, slug, composed, mode, session.ID)
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
		assistantContent = aiResult.Content
	}

	// 步骤 5: 得出结论
	sender.AddThinkingStep(thinking.ThinkingConclude, "综合以上分析得出结论")

	// 完成思考过程
	sender.Complete()

	// 发送最终答案
	sender.SendAnswer(assistantContent)

	// 保存完整的助手消息
	assistantMessageID := ""
	if assistantContent != "" {
		assistantMsg := models.Message{
			ID:        models.NewUUID(),
			SessionID: session.ID,
			Role:      "assistant",
			Content:   assistantContent,
			Sources:   buildAssistantSources(mode, aiResult),
			CreatedAt: time.Now(),
		}
		h.db.Create(&assistantMsg)
		h.db.Model(&session).Update("updated_at", time.Now())
		assistantMessageID = assistantMsg.ID
	}

	// 发送完成标记（附带真实消息 ID，便于前端替换临时 ID）
	doneChunk := map[string]interface{}{
		"type": "done",
		"done": true,
	}
	if assistantMessageID != "" {
		doneChunk["assistantMessageId"] = assistantMessageID
	}
	doneChunk["meta"] = map[string]interface{}{
		"mode":     string(mode),
		"sources":  aiResult.Sources,
		"thoughts": aiResult.Thoughts,
	}
	jsonData, _ := json.Marshal(doneChunk)
	fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
	flusher.Flush()
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
