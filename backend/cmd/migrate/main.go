package main

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"rolecraft-ai/internal/database"
	"rolecraft-ai/internal/models"
)

func main() {
	db, err := database.Init("")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🚀 Running migrations...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Workspace{},
		&models.Role{},
		&models.Skill{},
		&models.Document{},
		&models.ChatSession{},
		&models.Message{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✅ Migrations completed!")
	seedData(db)
}

func seedData(db *gorm.DB) {
	fmt.Println("🌱 Seeding data...")

	templates := []models.Role{
		{
			ID:             "role-001",
			Name:           "智能助理",
			Description:    "全能型办公助手",
			Category:       "通用",
			SystemPrompt:   "你是一位智能助理，擅长帮助用户处理各种办公任务。",
			WelcomeMessage: "你好！有什么可以帮你的吗？",
			IsTemplate:     true,
		},
		{
			ID:             "role-002",
			Name:           "营销专家",
			Description:    "营销策划助手",
			Category:       "营销",
			SystemPrompt:   "你是一位资深的营销专家。",
			WelcomeMessage: "你好！我是你的营销顾问。",
			IsTemplate:     true,
		},
		{
			ID:             "role-003",
			Name:           "法务顾问",
			Description:    "法律咨询专家",
			Category:       "法律",
			SystemPrompt:   "你是一位专业的法务顾问。",
			WelcomeMessage: "你好！有什么法律问题？",
			IsTemplate:     true,
		},
	}

	for _, t := range templates {
		if err := db.Create(&t).Error; err != nil {
			fmt.Printf("  ⚠️  %s 已存在\n", t.Name)
		} else {
			fmt.Printf("  ✅ %s\n", t.Name)
		}
	}

	fmt.Println("✅ Seeding completed!")
}
