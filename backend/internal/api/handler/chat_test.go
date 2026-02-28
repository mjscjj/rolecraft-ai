package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"rolecraft-ai/internal/api/handler"
	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/models"
)

// TestUpdateMessage 测试编辑消息
func TestUpdateMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db := setupTestDB(t)
	cfg := &config.Config{}
	chatHandler := handler.NewChatHandler(db, cfg)

	// 创建测试用户
	user := models.User{
		ID:           "test-user",
		Email:        "test@example.com",
		PasswordHash: "hashed",
	}
	db.Create(&user)

	// 创建测试会话
	session := models.ChatSession{
		ID:     "test-session",
		UserID: user.ID,
		Title:  "Test Session",
	}
	db.Create(&session)

	// 创建测试消息
	message := models.Message{
		ID:        "test-message",
		SessionID: session.ID,
		Role:      "user",
		Content:   "Original content",
	}
	db.Create(&message)

	// 测试编辑消息
	t.Run("Update user message", func(t *testing.T) {
		reqBody := map[string]string{"content": "Updated content"}
		jsonBody, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("PUT", "/api/v1/messages/"+message.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("userId", user.ID)
		ctx.Params = []gin.Param{{Key: "id", Value: message.ID}}
		
		chatHandler.UpdateMessage(ctx)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		// 验证数据库更新
		var updatedMsg models.Message
		db.First(&updatedMsg, message.ID)
		assert.Equal(t, "Updated content", updatedMsg.Content)
		assert.True(t, updatedMsg.IsEdited)
	})

	// 测试不能编辑 AI 消息
	t.Run("Cannot edit assistant message", func(t *testing.T) {
		aiMsg := models.Message{
			ID:        "test-ai-message",
			SessionID: session.ID,
			Role:      "assistant",
			Content:   "AI response",
		}
		db.Create(&aiMsg)

		reqBody := map[string]string{"content": "Hacked content"}
		jsonBody, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("PUT", "/api/v1/messages/"+aiMsg.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("userId", user.ID)
		ctx.Params = []gin.Param{{Key: "id", Value: aiMsg.ID}}
		
		chatHandler.UpdateMessage(ctx)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestAddFeedback 测试添加反馈
func TestAddFeedback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db := setupTestDB(t)
	cfg := &config.Config{}
	chatHandler := handler.NewChatHandler(db, cfg)

	// 创建测试数据
	user := models.User{ID: "test-user", Email: "test@example.com", PasswordHash: "hashed"}
	db.Create(&user)

	session := models.ChatSession{ID: "test-session", UserID: user.ID, Title: "Test"}
	db.Create(&session)

	message := models.Message{ID: "test-message", SessionID: session.ID, Role: "assistant", Content: "AI response"}
	db.Create(&message)

	t.Run("Add like feedback", func(t *testing.T) {
		reqBody := map[string]string{"type": "like"}
		jsonBody, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/v1/messages/"+message.ID+"/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("userId", user.ID)
		ctx.Params = []gin.Param{{Key: "id", Value: message.ID}}
		
		chatHandler.AddFeedback(ctx)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var updatedMsg models.Message
		db.First(&updatedMsg, message.ID)
		assert.Equal(t, 1, updatedMsg.Likes)
		assert.Equal(t, 0, updatedMsg.Dislikes)
	})

	t.Run("Add dislike feedback", func(t *testing.T) {
		reqBody := map[string]string{"type": "dislike"}
		jsonBody, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/v1/messages/"+message.ID+"/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("userId", user.ID)
		ctx.Params = []gin.Param{{Key: "id", Value: message.ID}}
		
		chatHandler.AddFeedback(ctx)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var updatedMsg models.Message
		db.First(&updatedMsg, message.ID)
		assert.Equal(t, 1, updatedMsg.Likes)
		assert.Equal(t, 1, updatedMsg.Dislikes)
	})
}

// TestExportSession 测试导出会话
func TestExportSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db := setupTestDB(t)
	cfg := &config.Config{}
	chatHandler := handler.NewChatHandler(db, cfg)

	// 创建测试数据
	user := models.User{ID: "test-user", Email: "test@example.com", PasswordHash: "hashed"}
	db.Create(&user)

	session := models.ChatSession{ID: "test-session", UserID: user.ID, Title: "Test Session"}
	db.Create(&session)

	message1 := models.Message{ID: "msg1", SessionID: session.ID, Role: "user", Content: "Hello"}
	message2 := models.Message{ID: "msg2", SessionID: session.ID, Role: "assistant", Content: "Hi there"}
	db.Create(&message1)
	db.Create(&message2)

	t.Run("Export as Markdown", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/chat-sessions/"+session.ID+"/export?format=md", nil)
		
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("userId", user.ID)
		ctx.Params = []gin.Param{{Key: "id", Value: session.ID}}
		
		chatHandler.ExportSession(ctx)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "# Test Session")
		assert.Contains(t, w.Body.String(), "Hello")
		assert.Contains(t, w.Body.String(), "Hi there")
	})

	t.Run("Export as JSON", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/chat-sessions/"+session.ID+"/export?format=json", nil)
		
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("userId", user.ID)
		ctx.Params = []gin.Param{{Key: "id", Value: session.ID}}
		
		chatHandler.ExportSession(ctx)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(0), response["code"])
	})
}

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	
	// 自动迁移
	db.AutoMigrate(&models.User{}, &models.ChatSession{}, &models.Message{})
	
	return db
}
