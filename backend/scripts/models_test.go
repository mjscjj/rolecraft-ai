package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

var testDB *gorm.DB

// TestMain 设置测试环境
func TestMain(m *testing.M) {
	// 创建临时数据库
	tmpDir := os.TempDir()
	dbPath := filepath.Join(tmpDir, "rolecraft_test.db")
	
	// 删除旧测试数据库
	os.Remove(dbPath)

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	testDB = db

	// 自动迁移模型
	if err := db.AutoMigrate(
		&models.User{},
		&models.Workspace{},
		&models.Role{},
		&models.Skill{},
		&models.Document{},
		&models.ChatSession{},
		&models.Message{},
	); err != nil {
		panic(fmt.Sprintf("Failed to auto migrate: %v", err))
	}

	// 运行测试
	code := m.Run()

	// 清理
	os.Remove(dbPath)

	os.Exit(code)
}

// TestUserCRUD 测试 User 模型的 CRUD 操作
func TestUserCRUD(t *testing.T) {
	fmt.Println("\n=== Testing User CRUD ===")

	// Create
	user := models.User{
		ID:              models.NewUUID(),
		Email:           "test@example.com",
		PasswordHash:    "hashed_password",
		Name:            "Test User",
		Avatar:          "https://example.com/avatar.png",
		AnythingLLMSlug: "test-workspace",
		EmailVerified:   true,
	}

	if err := testDB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("✓ Created user: %s\n", user.ID)

	// Read
	var fetchedUser models.User
	if err := testDB.Where("id = ?", user.ID).First(&fetchedUser).Error; err != nil {
		t.Fatalf("Failed to fetch user: %v", err)
	}
	fmt.Printf("✓ Fetched user: %s (AnythingLLMSlug: %s)\n", fetchedUser.Email, fetchedUser.AnythingLLMSlug)

	// Update
	fetchedUser.Name = "Updated Name"
	fetchedUser.AnythingLLMSlug = "updated-workspace"
	if err := testDB.Save(&fetchedUser).Error; err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}
	fmt.Printf("✓ Updated user: %s\n", fetchedUser.Name)

	// Delete
	if err := testDB.Where("id = ?", user.ID).Delete(&models.User{}).Error; err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	fmt.Printf("✓ Deleted user: %s\n", user.ID)

	// Verify deletion
	var count int64
	testDB.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	if count != 0 {
		t.Fatal("User should be deleted")
	}
	fmt.Println("✓ Verified user deletion")
}

// TestRoleCRUD 测试 Role 模型的 CRUD 操作
func TestRoleCRUD(t *testing.T) {
	fmt.Println("\n=== Testing Role CRUD ===")

	// Create
	role := models.Role{
		ID:             models.NewUUID(),
		UserID:         models.NewUUID(),
		Name:           "AI Assistant",
		Avatar:         "https://example.com/role-avatar.png",
		Description:    "A helpful AI assistant",
		Category:       "assistant",
		SystemPrompt:   "You are a helpful assistant",
		WelcomeMessage: "Hello! How can I help you?",
		ModelConfig:    models.ToJSON(map[string]string{"model": "gpt-4"}),
		IsTemplate:     false,
		IsPublic:       true,
	}

	if err := testDB.Create(&role).Error; err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}
	fmt.Printf("✓ Created role: %s\n", role.ID)

	// Read
	var fetchedRole models.Role
	if err := testDB.Where("id = ?", role.ID).First(&fetchedRole).Error; err != nil {
		t.Fatalf("Failed to fetch role: %v", err)
	}
	fmt.Printf("✓ Fetched role: %s (UserID: %s)\n", fetchedRole.Name, fetchedRole.UserID)

	// Update
	fetchedRole.Description = "Updated description"
	fetchedRole.SystemPrompt = "You are an updated assistant"
	if err := testDB.Save(&fetchedRole).Error; err != nil {
		t.Fatalf("Failed to update role: %v", err)
	}
	fmt.Printf("✓ Updated role: %s\n", fetchedRole.Description)

	// Query by UserID
	var userRoles []models.Role
	if err := testDB.Where("user_id = ?", role.UserID).Find(&userRoles).Error; err != nil {
		t.Fatalf("Failed to query roles by user: %v", err)
	}
	fmt.Printf("✓ Found %d roles for user\n", len(userRoles))

	// Delete
	if err := testDB.Where("id = ?", role.ID).Delete(&models.Role{}).Error; err != nil {
		t.Fatalf("Failed to delete role: %v", err)
	}
	fmt.Printf("✓ Deleted role: %s\n", role.ID)
}

// TestDocumentCRUD 测试 Document 模型的 CRUD 操作
func TestDocumentCRUD(t *testing.T) {
	fmt.Println("\n=== Testing Document CRUD ===")

	// Create
	doc := models.Document{
		ID:              models.NewUUID(),
		UserID:          models.NewUUID(),
		Name:            "Test Document.pdf",
		FileType:        "pdf",
		FileSize:        1024 * 1024, // 1MB
		FilePath:        "/tmp/test.pdf",
		AnythingLLMHash: "abc123def456",
		Status:          "pending",
		ChunkCount:      0,
		Metadata:        models.ToJSON(map[string]string{"author": "Test Author"}),
	}

	if err := testDB.Create(&doc).Error; err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	fmt.Printf("✓ Created document: %s\n", doc.ID)

	// Read
	var fetchedDoc models.Document
	if err := testDB.Where("id = ?", doc.ID).First(&fetchedDoc).Error; err != nil {
		t.Fatalf("Failed to fetch document: %v", err)
	}
	fmt.Printf("✓ Fetched document: %s (AnythingLLMHash: %s, Status: %s)\n", 
		fetchedDoc.Name, fetchedDoc.AnythingLLMHash, fetchedDoc.Status)

	// Update
	fetchedDoc.Status = "processing"
	fetchedDoc.ChunkCount = 10
	if err := testDB.Save(&fetchedDoc).Error; err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}
	fmt.Printf("✓ Updated document status: %s (Chunks: %d)\n", fetchedDoc.Status, fetchedDoc.ChunkCount)

	// Update to completed
	fetchedDoc.Status = "completed"
	fetchedDoc.ChunkCount = 25
	if err := testDB.Save(&fetchedDoc).Error; err != nil {
		t.Fatalf("Failed to complete document: %v", err)
	}
	fmt.Printf("✓ Document completed: %d chunks\n", fetchedDoc.ChunkCount)

	// Query by UserID
	var userDocs []models.Document
	if err := testDB.Where("user_id = ?", doc.UserID).Find(&userDocs).Error; err != nil {
		t.Fatalf("Failed to query documents by user: %v", err)
	}
	fmt.Printf("✓ Found %d documents for user\n", len(userDocs))

	// Query by Status
	var pendingDocs []models.Document
	if err := testDB.Where("status = ?", "pending").Find(&pendingDocs).Error; err != nil {
		t.Fatalf("Failed to query pending documents: %v", err)
	}
	fmt.Printf("✓ Found %d pending documents\n", len(pendingDocs))

	// Delete
	if err := testDB.Where("id = ?", doc.ID).Delete(&models.Document{}).Error; err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}
	fmt.Printf("✓ Deleted document: %s\n", doc.ID)
}

// TestChatSessionCRUD 测试 ChatSession 模型的 CRUD 操作
func TestChatSessionCRUD(t *testing.T) {
	fmt.Println("\n=== Testing ChatSession CRUD ===")

	// Create
	session := models.ChatSession{
		ID:              models.NewUUID(),
		UserID:          models.NewUUID(),
		RoleID:          models.NewUUID(),
		Title:           "Test Conversation",
		Mode:            "quick",
		AnythingLLMSlug: "test-workspace",
	}

	if err := testDB.Create(&session).Error; err != nil {
		t.Fatalf("Failed to create chat session: %v", err)
	}
	fmt.Printf("✓ Created chat session: %s\n", session.ID)

	// Read
	var fetchedSession models.ChatSession
	if err := testDB.Where("id = ?", session.ID).First(&fetchedSession).Error; err != nil {
		t.Fatalf("Failed to fetch chat session: %v", err)
	}
	fmt.Printf("✓ Fetched chat session: %s (Mode: %s, AnythingLLMSlug: %s)\n", 
		fetchedSession.Title, fetchedSession.Mode, fetchedSession.AnythingLLMSlug)

	// Update
	fetchedSession.Title = "Updated Conversation"
	fetchedSession.Mode = "task"
	if err := testDB.Save(&fetchedSession).Error; err != nil {
		t.Fatalf("Failed to update chat session: %v", err)
	}
	fmt.Printf("✓ Updated chat session: %s (Mode: %s)\n", fetchedSession.Title, fetchedSession.Mode)

	// Query by UserID
	var userSessions []models.ChatSession
	if err := testDB.Where("user_id = ?", session.UserID).Find(&userSessions).Error; err != nil {
		t.Fatalf("Failed to query sessions by user: %v", err)
	}
	fmt.Printf("✓ Found %d chat sessions for user\n", len(userSessions))

	// Query by RoleID
	var roleSessions []models.ChatSession
	if err := testDB.Where("role_id = ?", session.RoleID).Find(&roleSessions).Error; err != nil {
		t.Fatalf("Failed to query sessions by role: %v", err)
	}
	fmt.Printf("✓ Found %d chat sessions for role\n", len(roleSessions))

	// Create a message
	message := models.Message{
		ID:         models.NewUUID(),
		SessionID:  session.ID,
		Role:       "user",
		Content:    "Hello, AI!",
		Sources:    models.ToJSON([]string{}),
		TokensUsed: 10,
	}

	if err := testDB.Create(&message).Error; err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	fmt.Printf("✓ Created message in session\n")

	// Query messages by session
	var messages []models.Message
	if err := testDB.Where("session_id = ?", session.ID).Find(&messages).Error; err != nil {
		t.Fatalf("Failed to query messages: %v", err)
	}
	fmt.Printf("✓ Found %d messages in session\n", len(messages))

	// Delete
	if err := testDB.Where("id = ?", session.ID).Delete(&models.ChatSession{}).Error; err != nil {
		t.Fatalf("Failed to delete chat session: %v", err)
	}
	fmt.Printf("✓ Deleted chat session: %s\n", session.ID)
}

// TestIndexes 测试索引效果
func TestIndexes(t *testing.T) {
	fmt.Println("\n=== Testing Indexes ===")

	// 创建测试数据
	userID := models.NewUUID()
	
	// 创建多个 roles
	for i := 0; i < 10; i++ {
		role := models.Role{
			ID:             models.NewUUID(),
			UserID:         userID,
			Name:           fmt.Sprintf("Role %d", i),
			SystemPrompt:   "Test prompt",
			IsPublic:       i%2 == 0,
			IsTemplate:     i%3 == 0,
		}
		testDB.Create(&role)
	}

	// 创建多个 documents
	for i := 0; i < 20; i++ {
		doc := models.Document{
			ID:              models.NewUUID(),
			UserID:          userID,
			Name:            fmt.Sprintf("Doc %d.pdf", i),
			FileType:        "pdf",
			FileSize:        int64(1024 * i),
			AnythingLLMHash: fmt.Sprintf("hash_%d", i),
			Status:          []string{"pending", "processing", "completed"}[i%3],
		}
		testDB.Create(&doc)
	}

	// 创建多个 chat sessions
	for i := 0; i < 15; i++ {
		session := models.ChatSession{
			ID:              models.NewUUID(),
			UserID:          userID,
			RoleID:          models.NewUUID(),
			Title:           fmt.Sprintf("Session %d", i),
			Mode:            []string{"quick", "task"}[i%2],
			AnythingLLMSlug: fmt.Sprintf("workspace_%d", i%3),
		}
		testDB.Create(&session)
	}

	fmt.Println("✓ Created test data")

	// 测试索引查询性能
	start := time.Now()
	var roles []models.Role
	testDB.Where("user_id = ?", userID).Find(&roles)
	duration := time.Since(start)
	fmt.Printf("✓ Indexed query (roles by user): %d results in %v\n", len(roles), duration)

	start = time.Now()
	var docs []models.Document
	testDB.Where("user_id = ? AND status = ?", userID, "completed").Find(&docs)
	duration = time.Since(start)
	fmt.Printf("✓ Indexed query (documents by user+status): %d results in %v\n", len(docs), duration)

	start = time.Now()
	var sessions []models.ChatSession
	testDB.Where("anything_llm_slug = ?", "workspace_0").Find(&sessions)
	duration = time.Since(start)
	fmt.Printf("✓ Indexed query (sessions by anything_llm_slug): %d results in %v\n", len(sessions), duration)

	// 验证索引存在
	type IndexInfo struct {
		Name string
	}
	var indexes []IndexInfo
	testDB.Raw("SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%'").Scan(&indexes)
	fmt.Printf("✓ Verified %d custom indexes exist\n", len(indexes))
}

// TestAnythingLLMIntegration 测试 AnythingLLM 关联
func TestAnythingLLMIntegration(t *testing.T) {
	fmt.Println("\n=== Testing AnythingLLM Integration ===")

	userID := models.NewUUID()
	workspaceSlug := "my-anythingllm-workspace"

	// 创建用户
	user := models.User{
		ID:              userID,
		Email:           "anythingllm@test.com",
		PasswordHash:    "hash",
		AnythingLLMSlug: workspaceSlug,
	}
	testDB.Create(&user)
	fmt.Printf("✓ Created user with AnythingLLM slug: %s\n", user.AnythingLLMSlug)

	// 创建关联的文档
	doc := models.Document{
		ID:              models.NewUUID(),
		UserID:          userID,
		Name:            "Linked Document",
		FileType:        "txt",
		AnythingLLMHash: "anythingllm-doc-hash-123",
		Status:          "completed",
	}
	testDB.Create(&doc)
	fmt.Printf("✓ Created document with AnythingLLM hash: %s\n", doc.AnythingLLMHash)

	// 创建关联的会话
	session := models.ChatSession{
		ID:              models.NewUUID(),
		UserID:          userID,
		RoleID:          models.NewUUID(),
		Title:           "Workspace Chat",
		AnythingLLMSlug: workspaceSlug,
	}
	testDB.Create(&session)
	fmt.Printf("✓ Created chat session linked to AnythingLLM workspace: %s\n", session.AnythingLLMSlug)

	// 查询用户的所有 AnythingLLM 关联
	var userDocs []models.Document
	testDB.Where("user_id = ? AND anything_llm_hash IS NOT NULL", userID).Find(&userDocs)
	fmt.Printf("✓ Found %d documents with AnythingLLM hash\n", len(userDocs))

	var userSessions []models.ChatSession
	testDB.Where("user_id = ? AND anything_llm_slug = ?", userID, workspaceSlug).Find(&userSessions)
	fmt.Printf("✓ Found %d chat sessions in workspace: %s\n", len(userSessions), workspaceSlug)

	// 清理
	testDB.Where("id = ?", doc.ID).Delete(&models.Document{})
	testDB.Where("id = ?", session.ID).Delete(&models.ChatSession{})
	testDB.Where("id = ?", user.ID).Delete(&models.User{})
	fmt.Println("✓ Cleaned up test data")
}

// TestModelValidation 测试模型验证
func TestModelValidation(t *testing.T) {
	fmt.Println("\n=== Testing Model Validation ===")

	// 测试 User 唯一性
	user1 := models.User{
		ID:    models.NewUUID(),
		Email: "unique@test.com",
		PasswordHash: "hash1",
	}
	if err := testDB.Create(&user1).Error; err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2 := models.User{
		ID:    models.NewUUID(),
		Email: "unique@test.com", // 重复 email
		PasswordHash: "hash2",
	}
	if err := testDB.Create(&user2).Error; err == nil {
		t.Fatal("Should fail with duplicate email")
	}
	fmt.Println("✓ User validation: Duplicate email rejected")

	// 清理
	testDB.Where("id = ?", user1.ID).Delete(&models.User{})

	// 测试 Role 关联用户
	roleUserID := models.NewUUID()
	role := models.Role{
		ID:           models.NewUUID(),
		UserID:       roleUserID,
		Name:         "Test Role",
		SystemPrompt: "Test",
	}
	if err := testDB.Create(&role).Error; err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}
	
	// 查询验证
	var roleCount int64
	testDB.Model(&models.Role{}).Where("user_id = ?", roleUserID).Count(&roleCount)
	if roleCount != 1 {
		t.Fatal("Should have 1 role")
	}
	fmt.Println("✓ Role validation: User association works")

	// 清理
	testDB.Where("id = ?", role.ID).Delete(&models.Role{})

	// 测试 Document 关联用户
	docUserID := models.NewUUID()
	doc := models.Document{
		ID:       models.NewUUID(),
		UserID:   docUserID,
		Name:     "Test Doc",
		FileType: "pdf",
	}
	if err := testDB.Create(&doc).Error; err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	
	// 查询验证
	var docCount int64
	testDB.Model(&models.Document{}).Where("user_id = ?", docUserID).Count(&docCount)
	if docCount != 1 {
		t.Fatal("Should have 1 document")
	}
	fmt.Println("✓ Document validation: User association works")

	// 清理
	testDB.Where("id = ?", doc.ID).Delete(&models.Document{})

	// 测试 ChatSession 关联用户
	sessionUserID := models.NewUUID()
	session := models.ChatSession{
		ID:     models.NewUUID(),
		UserID: sessionUserID,
		RoleID: models.NewUUID(),
		Title:  "Test Session",
	}
	if err := testDB.Create(&session).Error; err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	// 查询验证
	var sessionCount int64
	testDB.Model(&models.ChatSession{}).Where("user_id = ?", sessionUserID).Count(&sessionCount)
	if sessionCount != 1 {
		t.Fatal("Should have 1 session")
	}
	fmt.Println("✓ ChatSession validation: User association works")

	// 清理
	testDB.Where("id = ?", session.ID).Delete(&models.ChatSession{})
}
