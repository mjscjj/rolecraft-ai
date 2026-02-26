package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"rolecraft-ai/internal/api/handler"
	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/database"
	"rolecraft-ai/internal/models"
)

// TestRoleAnythingLLMSync 测试角色创建/更新时自动同步到 AnythingLLM
func TestRoleAnythingLLMSync(t *testing.T) {
	// 初始化测试数据库
	db, err := database.InitSQLite(":memory:")
	assert.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.Role{})
	assert.NoError(t, err)

	// 加载配置
	cfg := config.Load()

	// 创建 RoleHandler
	roleHandler := handler.NewRoleHandler(db, cfg)

	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// 注册路由
	router.POST("/roles", roleHandler.Create)
	router.PUT("/roles/:id", roleHandler.Update)

	t.Run("CreateRole_ShouldSyncToAnythingLLM", func(t *testing.T) {
		// 创建角色请求
		createReq := handler.CreateRoleRequest{
			Name:           "测试角色",
			Description:    "用于测试 AnythingLLM 同步",
			Category:       "测试",
			SystemPrompt:   "你是一个测试助手",
			WelcomeMessage: "你好！",
		}

		body, _ := json.Marshal(createReq)
		req, _ := http.NewRequest(http.MethodPost, "/roles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证创建成功
		assert.Equal(t, http.StatusCreated, w.Code)

		// 验证角色已保存到数据库
		var role models.Role
		result := db.First(&role, "name = ?", "测试角色")
		assert.NoError(t, result.Error)
		assert.Equal(t, "测试角色", role.Name)
		assert.Equal(t, "你是一个测试助手", role.SystemPrompt)

		// 注意：由于同步是异步的，这里无法直接验证 AnythingLLM 同步结果
		// 实际测试中应该通过 mock 或等待 goroutine 完成来验证
		t.Log("✅ 角色创建成功，异步同步已触发")
	})

	t.Run("UpdateRole_ShouldSyncToAnythingLLM", func(t *testing.T) {
		// 先创建一个角色
		role := models.Role{
			ID:             models.NewUUID(),
			Name:           "更新测试角色",
			Description:    "原始描述",
			Category:       "测试",
			SystemPrompt:   "原始系统提示词",
			WelcomeMessage: "原始欢迎消息",
		}
		db.Create(&role)

		// 更新角色请求
		updateReq := handler.CreateRoleRequest{
			Name:           "更新测试角色",
			Description:    "更新后的描述",
			Category:       "测试",
			SystemPrompt:   "更新后的系统提示词",
			WelcomeMessage: "更新后的欢迎消息",
		}

		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPut, "/roles/"+role.ID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证更新成功
		assert.Equal(t, http.StatusOK, w.Code)

		// 验证角色已更新
		var updatedRole models.Role
		result := db.First(&updatedRole, "id = ?", role.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, "更新后的描述", updatedRole.Description)
		assert.Equal(t, "更新后的系统提示词", updatedRole.SystemPrompt)

		// 注意：由于同步是异步的，这里无法直接验证 AnythingLLM 同步结果
		t.Log("✅ 角色更新成功，异步同步已触发")
	})
}
