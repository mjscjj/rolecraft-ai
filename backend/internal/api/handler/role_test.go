package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"rolecraft-ai/internal/api/handler"
	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/models"
)

func setupRoleTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Role{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

func TestRoleChatOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRoleTestDB(t)
	roleHandler := handler.NewRoleHandler(db, &config.Config{})

	userA := models.User{ID: "user-a", Email: "a@example.com", PasswordHash: "hashed"}
	userB := models.User{ID: "user-b", Email: "b@example.com", PasswordHash: "hashed"}
	assert.NoError(t, db.Create(&userA).Error)
	assert.NoError(t, db.Create(&userB).Error)

	role := models.Role{
		ID:           "role-b",
		UserID:       userB.ID,
		Name:         "B Role",
		SystemPrompt: "You are B",
	}
	assert.NoError(t, db.Create(&role).Error)

	body, _ := json.Marshal(map[string]string{"message": "hello"})

	t.Run("cannot chat with other user's role", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/roles/"+role.ID+"/chat", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("userId", userA.ID)
		ctx.Params = []gin.Param{{Key: "id", Value: role.ID}}

		roleHandler.Chat(ctx)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("can chat with own role", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/roles/"+role.ID+"/chat", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("userId", userB.ID)
		ctx.Params = []gin.Param{{Key: "id", Value: role.ID}}

		roleHandler.Chat(ctx)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, float64(200), resp["code"])
	})
}
