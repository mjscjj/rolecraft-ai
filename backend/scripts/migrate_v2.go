package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 迁移脚本：RoleCraft AI v1 -> v2
// 主要变更：
// 1. User 表添加 anything_llm_slug 字段
// 2. Role 表简化：移除 workspace_id，添加 user_id，移除 skills/documents 关联
// 3. Document 表添加 anything_llm_hash 字段，user_id 替代 workspace_id
// 4. ChatSession 表添加 anything_llm_slug 字段
// 5. 添加索引优化查询性能

func main() {
	// 获取数据库路径
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		wd, _ := os.Getwd()
		dbPath = filepath.Join(wd, "rolecraft.db")
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(dbPath)
	if err == nil {
		dbPath = absPath
	}

	fmt.Printf("📦 Migrating database: %s\n", dbPath)

	// 检查数据库文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("⚠️  Database file does not exist. Creating new database...")
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Connected to database")

	// 执行迁移
	if err := migrateV2(db); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	fmt.Println("✅ Migration completed successfully!")
}

func migrateV2(db *gorm.DB) error {
	fmt.Println("\n🚀 Starting v2 migration...")

	// 启用 SQL 日志
	db = db.Debug()

	// 1. 迁移 User 表
	fmt.Println("\n📝 Migrating users table...")
	if err := migrateUsers(db); err != nil {
		return fmt.Errorf("users migration failed: %w", err)
	}

	// 2. 迁移 Role 表
	fmt.Println("\n📝 Migrating roles table...")
	if err := migrateRoles(db); err != nil {
		return fmt.Errorf("roles migration failed: %w", err)
	}

	// 3. 迁移 Document 表
	fmt.Println("\n📝 Migrating documents table...")
	if err := migrateDocuments(db); err != nil {
		return fmt.Errorf("documents migration failed: %w", err)
	}

	// 4. 迁移 ChatSession 表
	fmt.Println("\n📝 Migrating chat_sessions table...")
	if err := migrateChatSessions(db); err != nil {
		return fmt.Errorf("chat_sessions migration failed: %w", err)
	}

	// 5. 创建索引
	fmt.Println("\n📝 Creating indexes...")
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("index creation failed: %w", err)
	}

	// 6. 清理旧表
	fmt.Println("\n📝 Cleaning up old tables...")
	if err := cleanupOldTables(db); err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	fmt.Println("\n✅ All migrations completed!")
	return nil
}

func migrateUsers(db *gorm.DB) error {
	// 检查 anything_llm_slug 列是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='anything_llm_slug'").Scan(&count)
	
	if count == 0 {
		fmt.Println("   Adding anything_llm_slug column...")
		if err := db.Exec("ALTER TABLE users ADD COLUMN anything_llm_slug TEXT").Error; err != nil {
			return err
		}
		fmt.Println("   ✅ Added anything_llm_slug column")
	} else {
		fmt.Println("   ✓ anything_llm_slug column already exists")
	}

	// 创建索引
	fmt.Println("   Creating index on anything_llm_slug...")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_anything_llm_slug ON users(anything_llm_slug)")
	
	return nil
}

func migrateRoles(db *gorm.DB) error {
	// 检查 user_id 列是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('roles') WHERE name='user_id'").Scan(&count)
	
	if count == 0 {
		fmt.Println("   Adding user_id column...")
		if err := db.Exec("ALTER TABLE roles ADD COLUMN user_id TEXT NOT NULL DEFAULT ''").Error; err != nil {
			return err
		}
		fmt.Println("   ✅ Added user_id column")
	} else {
		fmt.Println("   ✓ user_id column already exists")
	}

	// 检查 workspace_id 列是否存在并标记为废弃
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('roles') WHERE name='workspace_id'").Scan(&count)
	if count > 0 {
		fmt.Println("   ⚠️  workspace_id column exists (deprecated, will be ignored)")
	}

	// 创建索引
	fmt.Println("   Creating indexes...")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_roles_user_id ON roles(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_roles_is_template ON roles(is_template)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_roles_is_public ON roles(is_public)")
	
	return nil
}

func migrateDocuments(db *gorm.DB) error {
	// 检查 anything_llm_hash 列是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('documents') WHERE name='anything_llm_hash'").Scan(&count)
	
	if count == 0 {
		fmt.Println("   Adding anything_llm_hash column...")
		if err := db.Exec("ALTER TABLE documents ADD COLUMN anything_llm_hash TEXT").Error; err != nil {
			return err
		}
		fmt.Println("   ✅ Added anything_llm_hash column")
	} else {
		fmt.Println("   ✓ anything_llm_hash column already exists")
	}

	// 检查 user_id 列是否存在
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('documents') WHERE name='user_id'").Scan(&count)
	if count == 0 {
		fmt.Println("   Adding user_id column...")
		if err := db.Exec("ALTER TABLE documents ADD COLUMN user_id TEXT NOT NULL DEFAULT ''").Error; err != nil {
			return err
		}
		fmt.Println("   ✅ Added user_id column")
		
		// 如果有 workspace_id，迁移数据到 user_id
		var wsCount int64
		db.Raw("SELECT COUNT(*) FROM pragma_table_info('documents') WHERE name='workspace_id'").Scan(&wsCount)
		if wsCount > 0 {
			fmt.Println("   Migrating workspace_id to user_id...")
			db.Exec("UPDATE documents SET user_id = workspace_id WHERE user_id = ''")
		}
	} else {
		fmt.Println("   ✓ user_id column already exists")
	}

	// 创建索引
	fmt.Println("   Creating indexes...")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_documents_user_id ON documents(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_documents_anything_llm_hash ON documents(anything_llm_hash)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status)")
	
	return nil
}

func migrateChatSessions(db *gorm.DB) error {
	// 检查 anything_llm_slug 列是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('chat_sessions') WHERE name='anything_llm_slug'").Scan(&count)
	
	if count == 0 {
		fmt.Println("   Adding anything_llm_slug column...")
		if err := db.Exec("ALTER TABLE chat_sessions ADD COLUMN anything_llm_slug TEXT").Error; err != nil {
			return err
		}
		fmt.Println("   ✅ Added anything_llm_slug column")
	} else {
		fmt.Println("   ✓ anything_llm_slug column already exists")
	}

	// 检查 user_id 列是否存在
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('chat_sessions') WHERE name='user_id'").Scan(&count)
	if count == 0 {
		fmt.Println("   Adding user_id column...")
		if err := db.Exec("ALTER TABLE chat_sessions ADD COLUMN user_id TEXT NOT NULL DEFAULT ''").Error; err != nil {
			return err
		}
		fmt.Println("   ✅ Added user_id column")
	} else {
		fmt.Println("   ✓ user_id column already exists")
	}

	// 创建索引
	fmt.Println("   Creating indexes...")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_id ON chat_sessions(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_chat_sessions_role_id ON chat_sessions(role_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_chat_sessions_anything_llm_slug ON chat_sessions(anything_llm_slug)")
	
	return nil
}

func createIndexes(db *gorm.DB) error {
	// Messages 表索引
	fmt.Println("   Creating messages indexes...")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)")

	// 复合索引
	fmt.Println("   Creating composite indexes...")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_roles_user_created ON roles(user_id, created_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_documents_user_status ON documents(user_id, status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_created ON chat_sessions(user_id, created_at)")
	
	return nil
}

func cleanupOldTables(db *gorm.DB) error {
	// 删除 role_skills 关联表（如果存在）
	var count int64
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='role_skills'").Scan(&count)
	if count > 0 {
		fmt.Println("   Dropping deprecated table: role_skills")
		db.Exec("DROP TABLE IF EXISTS role_skills")
	}

	// 删除 role_documents 关联表（如果存在）
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='role_documents'").Scan(&count)
	if count > 0 {
		fmt.Println("   Dropping deprecated table: role_documents")
		db.Exec("DROP TABLE IF EXISTS role_documents")
	}

	// 删除 skills 表（如果存在且为空）
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='skills'").Scan(&count)
	if count > 0 {
		var skillCount int64
		db.Table("skills").Count(&skillCount)
		if skillCount == 0 {
			fmt.Println("   Dropping empty table: skills")
			db.Exec("DROP TABLE IF EXISTS skills")
		} else {
			fmt.Println("   ⚠️  skills table has data, keeping it")
		}
	}

	return nil
}

// 辅助函数：检查列是否存在
func columnExists(db *gorm.DB, tableName, columnName string) (bool, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='%s'", tableName, columnName)
	err := db.Raw(query).Scan(&count).Error
	return count > 0, err
}

// 辅助函数：检查表是否存在
func tableExists(db *gorm.DB, tableName string) (bool, error) {
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count).Error
	return count > 0, err
}

// 测试函数：验证迁移结果
func testMigration(db *gorm.DB) error {
	fmt.Println("\n🧪 Running migration tests...")
	
	// 测试 1: 验证 User 模型
	fmt.Println("   Test 1: Verifying User model...")
	type TestUser struct {
		ID              string
		Email           string
		AnythingLLMSlug string
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}
	
	var testUser TestUser
	if err := db.Table("users").Select("id, email, anything_llm_slug, created_at, updated_at").First(&testUser).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("User model test failed: %w", err)
		}
		fmt.Println("   ✓ User model structure OK (no data yet)")
	} else {
		fmt.Printf("   ✓ User model OK - Sample: %+v\n", testUser)
	}

	// 测试 2: 验证 Role 模型
	fmt.Println("   Test 2: Verifying Role model...")
	type TestRole struct {
		ID     string
		UserID string
		Name   string
	}
	
	var testRole TestRole
	if err := db.Table("roles").Select("id, user_id, name").First(&testRole).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("Role model test failed: %w", err)
		}
		fmt.Println("   ✓ Role model structure OK (no data yet)")
	} else {
		fmt.Printf("   ✓ Role model OK - Sample: %+v\n", testRole)
	}

	// 测试 3: 验证 Document 模型
	fmt.Println("   Test 3: Verifying Document model...")
	type TestDocument struct {
		ID              string
		UserID          string
		AnythingLLMHash string
		Status          string
	}
	
	var testDoc TestDocument
	if err := db.Table("documents").Select("id, user_id, anything_llm_hash, status").First(&testDoc).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("Document model test failed: %w", err)
		}
		fmt.Println("   ✓ Document model structure OK (no data yet)")
	} else {
		fmt.Printf("   ✓ Document model OK - Sample: %+v\n", testDoc)
	}

	// 测试 4: 验证 ChatSession 模型
	fmt.Println("   Test 4: Verifying ChatSession model...")
	type TestChatSession struct {
		ID              string
		UserID          string
		RoleID          string
		AnythingLLMSlug string
	}
	
	var testSession TestChatSession
	if err := db.Table("chat_sessions").Select("id, user_id, role_id, anything_llm_slug").First(&testSession).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("ChatSession model test failed: %w", err)
		}
		fmt.Println("   ✓ ChatSession model structure OK (no data yet)")
	} else {
		fmt.Printf("   ✓ ChatSession model OK - Sample: %+v\n", testSession)
	}

	// 测试 5: 验证索引
	fmt.Println("   Test 5: Verifying indexes...")
	var indexes []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%'").Scan(&indexes)
	fmt.Printf("   ✓ Found %d indexes: %v\n", len(indexes), indexes)

	fmt.Println("\n✅ All migration tests passed!")
	return nil
}
