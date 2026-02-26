package main

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

// 演示数据种子脚本
func main() {
	// 初始化数据库
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 创建演示用户
	demoUser := createDemoUser(db)
	log.Printf("✅ Created demo user: %s", demoUser.Email)

	// 创建预定义角色
	roles := createDemoRoles(db, demoUser.ID)
	log.Printf("✅ Created %d demo roles", len(roles))

	// 创建演示对话
	sessions := createDemoSessions(db, demoUser.ID, roles)
	log.Printf("✅ Created %d demo sessions", len(sessions))

	// 创建演示文档
	documents := createDemoDocuments(db, demoUser.ID)
	log.Printf("✅ Created %d demo documents", len(documents))

	log.Println("🎉 Demo data seeding completed successfully!")
}

func initDB() (*gorm.DB, error) {
	// 使用 SQLite 数据库
	dbPath := "rolecraft.db"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func createDemoUser(db *gorm.DB) *models.User {
	// 检查是否已存在
	var existing models.User
	if db.Where("email = ?", "demo@rolecraft.ai").First(&existing).Error == nil {
		return &existing
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("demo123456"), bcrypt.DefaultCost)

	user := &models.User{
		ID:           models.NewUUID(),
		Email:        "demo@rolecraft.ai",
		PasswordHash: string(hashedPassword),
		Name:         "演示用户",
		Avatar:       "",
	}

	db.Create(user)
	return user
}

func createDemoRoles(db *gorm.DB, userID string) []*models.Role {
	roleTemplates := []struct {
		Name           string
		Description    string
		Category       string
		SystemPrompt   string
		WelcomeMessage string
	}{
		{
			Name:           "客服助手",
			Description:    "专业友好的客户服务代表",
			Category:       "商业",
			SystemPrompt:   "你是一名专业、耐心的客服代表。你的职责是：1. 快速响应客户问题 2. 提供准确的解决方案 3. 保持友好和专业的语气",
			WelcomeMessage: "您好！我是客服助手，有什么可以帮您的吗？",
		},
		{
			Name:           "写作助手",
			Description:    "专业文案和内容创作专家",
			Category:       "创作",
			SystemPrompt:   "你是一名专业的写作助手。你擅长文章撰写、文案创作、内容优化。",
			WelcomeMessage: "你好！我是写作助手，让我帮你创作出色的内容吧！",
		},
		{
			Name:           "代码助手",
			Description:    "全栈开发专家",
			Category:       "技术",
			SystemPrompt:   "你是一名经验丰富的全栈开发者。你擅长代码编写、架构设计、调试优化。",
			WelcomeMessage: "嗨！我是代码助手，有什么技术问题需要帮助吗？",
		},
		{
			Name:           "学习导师",
			Description:    "个性化学习指导专家",
			Category:       "教育",
			SystemPrompt:   "你是一名经验丰富的学习导师。你擅长制定学习计划、解释复杂概念。",
			WelcomeMessage: "欢迎来到学习之旅！我是你的导师，今天想学习什么？",
		},
		{
			Name:           "营销专家",
			Description:    "数字营销和增长专家",
			Category:       "商业",
			SystemPrompt:   "你是一名资深营销专家。你擅长营销策略、社交媒体营销、内容营销。",
			WelcomeMessage: "你好！我是营销专家，让我们一起提升你的业务增长！",
		},
	}

	var roles []*models.Role
	for _, t := range roleTemplates {
		role := &models.Role{
			ID:             models.NewUUID(),
			UserID:         userID,
			Name:           t.Name,
			Description:    t.Description,
			Category:       t.Category,
			SystemPrompt:   t.SystemPrompt,
			WelcomeMessage: t.WelcomeMessage,
			IsTemplate:     true,
			IsPublic:       true,
		}
		db.Create(role)
		roles = append(roles, role)
	}

	return roles
}

func createDemoSessions(db *gorm.DB, userID string, roles []*models.Role) []*models.ChatSession {
	if len(roles) == 0 {
		return nil
	}

	sessions := []*models.ChatSession{
		{
			ID:        models.NewUUID(),
			UserID:    userID,
			RoleID:    roles[0].ID,
			Title:     "客服咨询演示",
			Mode:      "quick",
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        models.NewUUID(),
			UserID:    userID,
			RoleID:    roles[1].ID,
			Title:     "写作协助演示",
			Mode:      "task",
			CreatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			ID:        models.NewUUID(),
			UserID:    userID,
			RoleID:    roles[2].ID,
			Title:     "代码咨询演示",
			Mode:      "quick",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	for _, session := range sessions {
		db.Create(session)
	}

	return sessions
}

func createDemoDocuments(db *gorm.DB, userID string) []*models.Document {
	docTemplates := []struct {
		Name     string
		FileType string
		Status   string
	}{
		{"产品使用手册", "pdf", "completed"},
		{"常见问题 FAQ", "md", "completed"},
		{"最佳实践指南", "md", "completed"},
	}

	var documents []*models.Document
	for _, t := range docTemplates {
		doc := &models.Document{
			ID:        models.NewUUID(),
			UserID:    userID,
			Name:      t.Name,
			FileType:  t.FileType,
			Status:    t.Status,
			CreatedAt: time.Now(),
		}
		db.Create(doc)
		documents = append(documents, doc)
	}

	return documents
}
