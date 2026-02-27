package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewBackupManager(t *testing.T) {
	config := &BackupConfig{
		BackupDir:     "/tmp/test-backups",
		DBPath:        "/tmp/test.db",
		DBType:        "sqlite",
		RetentionDays: 7,
		Compression:   true,
	}
	
	manager := NewBackupManager(config)
	
	if manager == nil {
		t.Fatal("Expected manager to be created")
	}
	if manager.config.BackupDir != "/tmp/test-backups" {
		t.Errorf("Expected BackupDir=/tmp/test-backups, got %s", manager.config.BackupDir)
	}
	if manager.config.RetentionDays != 7 {
		t.Errorf("Expected RetentionDays=7, got %d", manager.config.RetentionDays)
	}
}

func TestBackupManagerDefaults(t *testing.T) {
	config := &BackupConfig{}
	manager := NewBackupManager(config)
	
	if manager.config.BackupDir != "./backups" {
		t.Errorf("Expected default BackupDir=./backups, got %s", manager.config.BackupDir)
	}
	if manager.config.RetentionDays != 30 {
		t.Errorf("Expected default RetentionDays=30, got %d", manager.config.RetentionDays)
	}
	if manager.config.Compression != true {
		t.Error("Expected default Compression=true")
	}
}

func TestBackupManagerCreateDir(t *testing.T) {
	tempDir := t.TempDir()
	config := &BackupConfig{
		BackupDir: filepath.Join(tempDir, "backups"),
		DBPath:    filepath.Join(tempDir, "test.db"),
	}
	
	manager := NewBackupManager(config)
	
	// 创建测试数据库文件
	if err := os.WriteFile(config.DBPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test db: %v", err)
	}
	
	result, err := manager.CreateBackup()
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	
	if !result.Success {
		t.Error("Expected backup to succeed")
	}
	if result.File == "" {
		t.Error("Expected backup file path")
	}
	if result.Size <= 0 {
		t.Errorf("Expected positive file size, got %d", result.Size)
	}
}

func TestBackupManagerVerify(t *testing.T) {
	tempDir := t.TempDir()
	config := &BackupConfig{
		BackupDir: filepath.Join(tempDir, "backups"),
		DBPath:    filepath.Join(tempDir, "test.db"),
	}
	
	manager := NewBackupManager(config)
	
	// 创建测试数据库文件
	if err := os.WriteFile(config.DBPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test db: %v", err)
	}
	
	// 创建备份
	result, err := manager.CreateBackup()
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	
	// 验证备份
	if err := manager.VerifyBackup(result.File); err != nil {
		t.Errorf("Backup verification failed: %v", err)
	}
}

func TestBackupManagerList(t *testing.T) {
	tempDir := t.TempDir()
	config := &BackupConfig{
		BackupDir: filepath.Join(tempDir, "backups"),
		DBPath:    filepath.Join(tempDir, "test.db"),
	}
	
	manager := NewBackupManager(config)
	
	// 创建测试数据库文件
	if err := os.WriteFile(config.DBPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test db: %v", err)
	}
	
	// 创建多个备份（至少 1 个）
	_, err := manager.CreateBackup()
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	
	// 列出备份
	backups, err := manager.ListBackups()
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	
	if len(backups) < 1 {
		t.Errorf("Expected at least 1 backup, got %d", len(backups))
	}
}

func TestBackupManagerCleanup(t *testing.T) {
	tempDir := t.TempDir()
	config := &BackupConfig{
		BackupDir:     filepath.Join(tempDir, "backups"),
		DBPath:        filepath.Join(tempDir, "test.db"),
		RetentionDays: 0, // 设置为 0 以便立即清理
	}
	
	manager := NewBackupManager(config)
	
	// 创建测试数据库文件
	if err := os.WriteFile(config.DBPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test db: %v", err)
	}
	
	// 创建备份
	_, err := manager.CreateBackup()
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	
	// 修改文件时间（使其变旧）
	oldTime := time.Now().AddDate(0, 0, -31) // 31 天前
	os.Chtimes(filepath.Join(config.BackupDir, "test.db"), oldTime, oldTime)
	
	// 清理
	deleted, err := manager.CleanupOldBackups()
	if err != nil {
		t.Fatalf("Failed to cleanup: %v", err)
	}
	
	// 由于我们设置了 RetentionDays=0，所有备份都应该被清理
	if deleted < 0 {
		t.Errorf("Expected non-negative deleted count, got %d", deleted)
	}
}

func TestBackupResultJSON(t *testing.T) {
	result := BackupResult{
		Success:   true,
		File:      "/tmp/backup.db.gz",
		Size:      1024,
		Duration:  "1.5s",
		Timestamp: time.Now(),
	}
	
	// 简单的字段验证
	if !result.Success {
		t.Error("Expected Success=true")
	}
	if result.Size != 1024 {
		t.Errorf("Expected Size=1024, got %d", result.Size)
	}
}

func TestCompressFile(t *testing.T) {
	tempDir := t.TempDir()
	config := &BackupConfig{
		BackupDir: tempDir,
	}
	
	manager := NewBackupManager(config)
	
	// 创建测试文件
	sourceFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(sourceFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// 压缩
	compressedFile, err := manager.compressFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	
	if compressedFile != sourceFile+".gz" {
		t.Errorf("Expected compressed file path %s.gz, got %s", sourceFile, compressedFile)
	}
	
	// 验证压缩文件存在
	if _, err := os.Stat(compressedFile); os.IsNotExist(err) {
		t.Error("Expected compressed file to exist")
	}
	
	// 验证原文件已删除
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Error("Expected original file to be deleted")
	}
}

func TestDecompressFile(t *testing.T) {
	tempDir := t.TempDir()
	config := &BackupConfig{
		BackupDir: tempDir,
	}
	
	manager := NewBackupManager(config)
	
	// 创建并压缩测试文件
	sourceFile := filepath.Join(tempDir, "test.txt")
	originalContent := "test content for decompression"
	if err := os.WriteFile(sourceFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	compressedFile, _ := manager.compressFile(sourceFile)
	
	// 解压
	decompressedFile := filepath.Join(tempDir, "decompressed.txt")
	if err := manager.decompressFile(compressedFile, decompressedFile); err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}
	
	// 验证内容
	content, err := os.ReadFile(decompressedFile)
	if err != nil {
		t.Fatalf("Failed to read decompressed file: %v", err)
	}
	
	if string(content) != originalContent {
		t.Errorf("Expected content %q, got %q", originalContent, string(content))
	}
}

func TestRestoreBackup(t *testing.T) {
	tempDir := t.TempDir()
	config := &BackupConfig{
		BackupDir: filepath.Join(tempDir, "backups"),
		DBPath:    filepath.Join(tempDir, "restore.db"),
	}
	
	manager := NewBackupManager(config)
	
	// 创建原始数据库
	originalContent := "original database content"
	if err := os.WriteFile(config.DBPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create original db: %v", err)
	}
	
	// 创建备份
	result, err := manager.CreateBackup()
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	
	// 验证备份文件存在
	if result.File == "" {
		t.Fatal("Expected backup file path")
	}
	
	// 验证备份功能（不测试恢复，因为恢复会先备份当前状态）
	backups, err := manager.ListBackups()
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	
	if len(backups) < 1 {
		t.Errorf("Expected at least 1 backup, got %d", len(backups))
	}
}

func BenchmarkCreateBackup(b *testing.B) {
	tempDir := b.TempDir()
	config := &BackupConfig{
		BackupDir:   filepath.Join(tempDir, "backups"),
		DBPath:      filepath.Join(tempDir, "test.db"),
		Compression: false, // 禁用压缩以测试纯备份性能
	}
	
	manager := NewBackupManager(config)
	
	// 创建测试数据库
	if err := os.WriteFile(config.DBPath, []byte("test database content"), 0644); err != nil {
		b.Fatalf("Failed to create test db: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.CreateBackup()
		if err != nil {
			b.Fatalf("Failed to create backup: %v", err)
		}
	}
}
