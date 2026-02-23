package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Init 初始化数据库 (SQLite)
func Init(databaseURL string) (*gorm.DB, error) {
	return InitSQLite(databaseURL)
}

// InitSQLite 初始化 SQLite 数据库
func InitSQLite(dbPath string) (*gorm.DB, error) {
	// 强制使用 SQLite，忽略 databaseURL 参数
	// 检查环境变量 DB_PATH
	dbPath = os.Getenv("DB_PATH")
	if dbPath == "" {
		// 获取当前工作目录
		wd, _ := os.Getwd()
		dbPath = filepath.Join(wd, "rolecraft.db")
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(dbPath)
	if err == nil {
		dbPath = absPath
	}

	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		os.MkdirAll(dir, 0755)
	}

	fmt.Printf("📦 Using SQLite database: %s\n", dbPath)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	return db, nil
}
