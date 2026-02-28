package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	_ "rolecraft-ai/docs"
	"rolecraft-ai/internal/api/handler"
	mw "rolecraft-ai/internal/api/middleware"
	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/database"
	"rolecraft-ai/internal/models"
	promptSvc "rolecraft-ai/internal/service/prompt"
)

// ServiceStartTime 服务启动时间
var ServiceStartTime = time.Now()

// @title RoleCraft AI API
// @version 1.0
// @description AI 角色管理平台 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@rolecraft.ai

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 使用 JWT Token 进行认证，格式："Bearer {token}"

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.InitSQLite(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	if err := ensureLegacySchemaCompatibility(db); err != nil {
		log.Fatalf("Failed to ensure legacy schema compatibility: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Workspace{},
		&models.Role{},
		&models.Skill{},
		&models.Document{},
		&models.Folder{},
		&models.ChatSession{},
		&models.Message{},
	); err != nil {
		log.Fatalf("Failed to auto migrate database schema: %v", err)
	}
	// Some legacy SQLite files may miss new chat_sessions columns even after AutoMigrate.
	var modelConfigCount int64
	if err := db.Raw("SELECT COUNT(*) FROM pragma_table_info('chat_sessions') WHERE name = 'model_config'").Scan(&modelConfigCount).Error; err != nil {
		log.Fatalf("Failed to check chat_sessions.model_config column: %v", err)
	}
	if modelConfigCount == 0 {
		if err := db.Exec("ALTER TABLE chat_sessions ADD COLUMN model_config TEXT").Error; err != nil {
			log.Fatalf("Failed to add chat_sessions.model_config column: %v", err)
		}
	}

	// 设置 Gin 模式
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
	}))

	// ===== 健康检查和监控路由 =====

	healthHandler := handler.NewHealthHandler(db, cfg)

	// 简单健康检查（向后兼容）
	r.GET("/health", handler.SimpleHealthCheck)

	// 综合健康检查
	r.GET("/api/v1/health", healthHandler.Health)

	// Kubernetes 探针
	r.GET("/api/v1/ready", healthHandler.Ready)
	r.GET("/api/v1/live", healthHandler.Live)

	// 性能指标
	r.GET("/api/v1/metrics", healthHandler.Metrics)
	r.GET("/api/v1/db/stats", healthHandler.DBStats)

	// Swagger API 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 公开路由
		auth := api.Group("/auth")
		{
			authHandler := handler.NewAuthHandler(db)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// 角色模板 (公开)
		roleHandler := handler.NewRoleHandler(db, cfg)
		api.GET("/roles/templates", roleHandler.GetEnhancedTemplates)

		// 需要认证的路由
		authorized := api.Group("/")
		authorized.Use(mw.JWTAuth())
		{
			// 用户
			userHandler := handler.NewUserHandler(db)
			authorized.GET("/users/me", userHandler.GetMe)
			authorized.PUT("/users/me", userHandler.UpdateMe)

			// 角色
			authorized.GET("/roles", roleHandler.List)
			authorized.GET("/roles/:id", roleHandler.Get)
			authorized.POST("/roles", roleHandler.Create)
			authorized.PUT("/roles/:id", roleHandler.Update)
			authorized.DELETE("/roles/:id", roleHandler.Delete)
			authorized.POST("/roles/:id/chat", roleHandler.Chat)

			// 文档
			docHandler := handler.NewDocumentHandler(db)
			authorized.GET("/documents", docHandler.List)
			authorized.POST("/documents", docHandler.Upload)
			authorized.POST("/documents/search", docHandler.Search)
			authorized.GET("/documents/:id", docHandler.Get)
			authorized.GET("/documents/:id/status", docHandler.GetStatus)
			authorized.GET("/documents/:id/preview", docHandler.Preview)
			authorized.GET("/documents/:id/download", docHandler.Download)
			authorized.PUT("/documents/:id", docHandler.Update)
			authorized.DELETE("/documents/:id", docHandler.Delete)

			// 批量操作
			authorized.DELETE("/documents/batch", docHandler.BatchDelete)
			authorized.PUT("/documents/batch/move", docHandler.BatchMove)
			authorized.PUT("/documents/batch/tags", docHandler.BatchUpdateTags)

			// 文件夹
			authorized.GET("/folders", docHandler.ListFolders)
			authorized.POST("/folders", docHandler.CreateFolder)
			authorized.DELETE("/folders/:id", docHandler.DeleteFolder)

			// 对话
			chatHandler := handler.NewChatHandler(db, cfg)
			authorized.GET("/chat-sessions", chatHandler.ListSessions)
			authorized.POST("/chat-sessions", chatHandler.CreateSession)
			authorized.GET("/chat-sessions/:id", chatHandler.GetSession)
			authorized.DELETE("/chat-sessions/:id", chatHandler.WorkspaceAuth(), chatHandler.DeleteSession)
			authorized.POST("/chat-sessions/:id/switch-role", chatHandler.WorkspaceAuth(), chatHandler.SwitchRole)
			authorized.GET("/chat-sessions/:id/sync", chatHandler.WorkspaceAuth(), chatHandler.SyncSession)
			authorized.DELETE("/chat-sessions/:id/messages/:msgId", chatHandler.WorkspaceAuth(), chatHandler.DeleteMessage)
			authorized.PUT("/chat-sessions/:id/title", chatHandler.WorkspaceAuth(), chatHandler.UpdateSessionTitle)
			authorized.PUT("/chat-sessions/:id/config", chatHandler.WorkspaceAuth(), chatHandler.UpdateSessionConfig)
			authorized.POST("/chat-sessions/:id/archive", chatHandler.WorkspaceAuth(), chatHandler.ArchiveSession)
			authorized.POST("/chat-sessions/:id/export", chatHandler.WorkspaceAuth(), chatHandler.ExportSession)
			authorized.POST("/chat-sessions/search", chatHandler.WorkspaceAuth(), chatHandler.SearchSessions)
			authorized.PUT("/chat/:id/messages/:msgId", chatHandler.WorkspaceAuth(), chatHandler.UpdateMessage)
			authorized.POST("/chat/:id/messages/:msgId/regenerate", chatHandler.WorkspaceAuth(), chatHandler.RegenerateMessage)
			authorized.POST("/chat/messages/:msgId/rate", chatHandler.WorkspaceAuth(), chatHandler.RateMessage)
			authorized.POST("/chat/:id/complete", chatHandler.WorkspaceAuth(), chatHandler.Chat)
			authorized.POST("/chat/:id/stream", chatHandler.WorkspaceAuth(), chatHandler.ChatStream)
			authorized.POST("/chat/:id/stream-with-thinking", chatHandler.WorkspaceAuth(), chatHandler.ChatStreamWithThinking)

			// 测试
			testHandler := handler.NewTestHandler(db)
			authorized.POST("/test/message", testHandler.SendMessage)
			authorized.POST("/test/ab", testHandler.RunABTest)
			authorized.POST("/test/compare", testHandler.CompareVersions)
			authorized.POST("/test/save", testHandler.SaveTestResult)
			authorized.GET("/test/history", testHandler.GetTestHistory)
			authorized.GET("/test/report", testHandler.GetTestReport)
			authorized.GET("/test/export/:roleId", testHandler.ExportTestReport)
			authorized.POST("/test/rate", testHandler.RateTestResponse)

			// 提示词优化（适配标准库 handler 到 Gin）
			promptOptimizer := promptSvc.NewOptimizer()
			promptHandler := handler.NewPromptHandler(promptOptimizer)
			authorized.POST("/prompt/optimize", func(c *gin.Context) {
				promptHandler.Optimize(c.Writer, c.Request)
			})
			authorized.POST("/prompt/suggestions", func(c *gin.Context) {
				promptHandler.GetSuggestions(c.Writer, c.Request)
			})
			authorized.POST("/prompt/log", func(c *gin.Context) {
				promptHandler.LogSelection(c.Writer, c.Request)
			})

			// 角色创建向导
			wizardHandler := handler.NewWizardHandler()
			authorized.GET("/wizard/options", wizardHandler.GetOptions)
			authorized.POST("/wizard/generate", wizardHandler.GeneratePrompt)
			authorized.POST("/wizard/recommendations", wizardHandler.GetRecommendations)
			authorized.POST("/wizard/test", wizardHandler.RunTest)
			authorized.POST("/wizard/export", wizardHandler.ExportConfig)
			authorized.POST("/wizard/validate", wizardHandler.ValidateData)
			authorized.GET("/wizard/templates", wizardHandler.GetTemplates)

			// 数据分析
			analyticsHandler := handler.NewAnalyticsHandler(db, cfg)
			authorized.GET("/analytics/dashboard", analyticsHandler.GetDashboardMetrics)
			authorized.GET("/analytics/user-activity", analyticsHandler.GetUserActivity)
			authorized.GET("/analytics/feature-usage", analyticsHandler.GetFeatureUsage)
			authorized.GET("/analytics/retention", analyticsHandler.GetRetentionRate)
			authorized.GET("/analytics/churn-risk", analyticsHandler.GetChurnRiskUsers)
			authorized.GET("/analytics/conversation-quality", analyticsHandler.GetConversationQuality)
			authorized.GET("/analytics/reply-quality", analyticsHandler.GetReplyQuality)
			authorized.GET("/analytics/faq", analyticsHandler.GetFAQStats)
			authorized.GET("/analytics/sensitive-words", analyticsHandler.GetSensitiveWords)
			authorized.GET("/analytics/cost", analyticsHandler.GetCostStats)
			authorized.GET("/analytics/cost/by-role", analyticsHandler.GetCostByRole)
			authorized.GET("/analytics/cost/by-user", analyticsHandler.GetCostByUser)
			authorized.GET("/analytics/cost/trend", analyticsHandler.GetCostTrend)
			authorized.GET("/analytics/cost/prediction", analyticsHandler.GetCostPrediction)
			authorized.GET("/analytics/report", analyticsHandler.GenerateReport)
			authorized.GET("/analytics/report/export", analyticsHandler.ExportReport)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// ensureLegacySchemaCompatibility 兼容旧版 SQLite 结构，避免 AutoMigrate 在旧库上失败
func ensureLegacySchemaCompatibility(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.Role{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&models.Role{}, "user_id") {
		if err := db.Exec("ALTER TABLE roles ADD COLUMN user_id TEXT").Error; err != nil {
			return err
		}
	}
	return nil
}
