package handler_test

import (
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

func setupAnalyticsTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.ChatSession{},
		&models.Message{},
		&models.Document{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func TestExportReport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAnalyticsTestDB(t)
	analyticsHandler := handler.NewAnalyticsHandler(db, &config.Config{})

	t.Run("export markdown", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/analytics/report/export?type=weekly&format=markdown", nil)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		analyticsHandler.ExportReport(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/markdown")
		assert.Contains(t, w.Header().Get("Content-Disposition"), ".md")
		assert.Contains(t, w.Body.String(), "# RoleCraft 数据分析周报")
	})

	t.Run("export json", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/analytics/report/export?type=monthly&format=json", nil)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		analyticsHandler.ExportReport(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
		assert.Contains(t, w.Header().Get("Content-Disposition"), ".json")

		var payload map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &payload)
		assert.NoError(t, err)
		assert.Equal(t, "monthly", payload["reportType"])
	})

	t.Run("unsupported format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/analytics/report/export?type=weekly&format=pdf", nil)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		analyticsHandler.ExportReport(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetCostByUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAnalyticsTestDB(t)
	analyticsHandler := handler.NewAnalyticsHandler(db, &config.Config{})

	user := models.User{
		ID:           "user-1",
		Email:        "cost@example.com",
		PasswordHash: "hashed",
		Name:         "Cost User",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	session := models.ChatSession{
		ID:     "session-1",
		UserID: user.ID,
		Title:  "cost session",
	}
	if err := db.Create(&session).Error; err != nil {
		t.Fatalf("failed to seed session: %v", err)
	}

	msg := models.Message{
		ID:         "msg-1",
		SessionID:  session.ID,
		Role:       "assistant",
		Content:    "hello",
		TokensUsed: 123,
	}
	if err := db.Create(&msg).Error; err != nil {
		t.Fatalf("failed to seed message: %v", err)
	}

	req, _ := http.NewRequest("GET", "/api/v1/analytics/cost/by-user", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	analyticsHandler.GetCostByUser(ctx)

	assert.Equal(t, http.StatusOK, w.Code)

	var payload map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), payload["code"])

	items, ok := payload["data"].([]any)
	assert.True(t, ok)
	assert.NotEmpty(t, items)
}
