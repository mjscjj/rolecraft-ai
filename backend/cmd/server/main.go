package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"rolecraft-ai/internal/api/handler"
	"rolecraft-ai/internal/api/middleware"
	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/database"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.InitSQLite(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 设置Gin模式
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API路由组
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
		roleHandler := handler.NewRoleHandler(db)
		api.GET("/roles/templates", roleHandler.GetTemplates)

		// 需要认证的路由
		authorized := api.Group("/")
		authorized.Use(middleware.JWTAuth())
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
			authorized.GET("/documents/:id", docHandler.Get)
			authorized.DELETE("/documents/:id", docHandler.Delete)

			// 对话
			chatHandler := handler.NewChatHandler(db, cfg)
			authorized.GET("/chat-sessions", chatHandler.ListSessions)
			authorized.POST("/chat-sessions", chatHandler.CreateSession)
			authorized.GET("/chat-sessions/:id", chatHandler.GetSession)
			authorized.POST("/chat/:id/complete", chatHandler.Chat)
			authorized.POST("/chat/:id/stream", chatHandler.ChatStream)
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