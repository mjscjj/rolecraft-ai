package main

import (
	"archive/zip"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupConfig 备份配置
type BackupConfig struct {
	BackupDir      string        `json:"backup_dir"`
	DBPath         string        `json:"db_path"`
	DBType         string        `json:"db_type"` // sqlite, postgres
	PGConnection   string        `json:"pg_connection"`
	RetentionDays  int           `json:"retention_days"`
	Compression    bool          `json:"compression"`
	VerifyAfter    bool          `json:"verify_after"`
}

// BackupResult 备份结果
type BackupResult struct {
	Success   bool      `json:"success"`
	File      string    `json:"file,omitempty"`
	Size      int64     `json:"size"`
	Duration  string    `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// BackupManager 备份管理器
type BackupManager struct {
	config *BackupConfig
}

// NewBackupManager 创建备份管理器
func NewBackupManager(config *BackupConfig) *BackupManager {
	if config.BackupDir == "" {
		config.BackupDir = "./backups"
	}
	if config.DBType == "" {
		config.DBType = "sqlite"
	}
	if config.RetentionDays == 0 {
		config.RetentionDays = 30
	}
	if config.Compression == false {
		config.Compression = true
	}
	if config.VerifyAfter == false {
		config.VerifyAfter = true
	}

	return &BackupManager{
		config: config,
	}
}

// CreateBackup 创建备份
func (m *BackupManager) CreateBackup() (*BackupResult, error) {
	startTime := time.Now()

	// 创建备份目录
	if err := os.MkdirAll(m.config.BackupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// 生成备份文件名
	timestamp := startTime.Format("20060102_150405")
	backupFile := filepath.Join(m.config.BackupDir, fmt.Sprintf("rolecraft_%s.db", timestamp))

	var err error
	if m.config.DBType == "postgres" {
		err = m.backupPostgres(backupFile)
	} else {
		err = m.backupSQLite(backupFile)
	}

	if err != nil {
		return &BackupResult{
			Success:   false,
			Error:     err.Error(),
			Timestamp: startTime,
		}, err
	}

	// 压缩
	if m.config.Compression {
		backupFile, err = m.compressFile(backupFile)
		if err != nil {
			return nil, fmt.Errorf("failed to compress backup: %w", err)
		}
	}

	// 验证
	if m.config.VerifyAfter {
		if err := m.VerifyBackup(backupFile); err != nil {
			return nil, fmt.Errorf("backup verification failed: %w", err)
		}
	}

	// 获取文件大小
	fileInfo, err := os.Stat(backupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// 清理旧备份
	m.CleanupOldBackups()

	return &BackupResult{
		Success:   true,
		File:      backupFile,
		Size:      fileInfo.Size(),
		Duration:  time.Since(startTime).String(),
		Timestamp: startTime,
	}, nil
}

// backupSQLite 备份 SQLite 数据库
func (m *BackupManager) backupSQLite(backupFile string) error {
	if _, err := os.Stat(m.config.DBPath); os.IsNotExist(err) {
		return fmt.Errorf("database file does not exist: %s", m.config.DBPath)
	}

	// 复制文件
	sourceFile, err := os.Open(m.config.DBPath)
	if err != nil {
		return fmt.Errorf("failed to open source database: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupFile)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy database: %w", err)
	}

	return nil
}

// backupPostgres 备份 PostgreSQL 数据库
func (m *BackupManager) backupPostgres(backupFile string) error {
	if m.config.PGConnection == "" {
		return fmt.Errorf("PostgreSQL connection string not provided")
	}

	// 连接到数据库
	db, err := sql.Open("postgres", m.config.PGConnection)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// 这里应该使用 pg_dump，但为了简化，我们导出为 SQL
	// 生产环境应该调用 pg_dump 命令
	return fmt.Errorf("PostgreSQL backup requires pg_dump command")
}

// compressFile 压缩文件
func (m *BackupManager) compressFile(sourceFile string) (string, error) {
	destFile := sourceFile + ".gz"

	source, err := os.Open(sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(destFile)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dest.Close()

	gzWriter := gzip.NewWriter(dest)
	defer gzWriter.Close()

	_, err = io.Copy(gzWriter, source)
	if err != nil {
		return "", fmt.Errorf("failed to compress file: %w", err)
	}

	// 删除原文件
	os.Remove(sourceFile)

	return destFile, nil
}

// VerifyBackup 验证备份文件
func (m *BackupManager) VerifyBackup(backupFile string) error {
	// 检查文件存在
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupFile)
	}

	// 如果是压缩文件，验证压缩完整性
	if strings.HasSuffix(backupFile, ".gz") {
		file, err := os.Open(backupFile)
		if err != nil {
			return err
		}
		defer file.Close()

		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("invalid gzip file: %w", err)
		}
		defer gzReader.Close()

		// 尝试读取部分内容
		buf := make([]byte, 1024)
		_, err = gzReader.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("gzip file corrupted: %w", err)
		}
	}

	return nil
}

// CleanupOldBackups 清理旧备份
func (m *BackupManager) CleanupOldBackups() (int, error) {
	cutoffTime := time.Now().AddDate(0, 0, -m.config.RetentionDays)
	deletedCount := 0

	entries, err := os.ReadDir(m.config.BackupDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 检查是否是备份文件
		if !strings.HasPrefix(entry.Name(), "rolecraft_") {
			continue
		}

		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		if fileInfo.ModTime().Before(cutoffTime) {
			filePath := filepath.Join(m.config.BackupDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				fmt.Printf("Failed to delete old backup %s: %v\n", filePath, err)
				continue
			}
			deletedCount++
			fmt.Printf("Deleted old backup: %s\n", filePath)
		}
	}

	return deletedCount, nil
}

// ListBackups 列出所有备份
func (m *BackupManager) ListBackups() ([]BackupResult, error) {
	var backups []BackupResult

	entries, err := os.ReadDir(m.config.BackupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasPrefix(entry.Name(), "rolecraft_") {
			continue
		}

		filePath := filepath.Join(m.config.BackupDir, entry.Name())
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupResult{
			Success:   true,
			File:      filePath,
			Size:      fileInfo.Size(),
			Timestamp: fileInfo.ModTime(),
		})
	}

	// 按时间排序（最新的在前）
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// RestoreBackup 恢复备份
func (m *BackupManager) RestoreBackup(backupFile string) error {
	// 验证备份文件
	if err := m.VerifyBackup(backupFile); err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}

	// 创建当前数据库的备份
	if _, err := os.Stat(m.config.DBPath); err == nil {
		fmt.Println("Creating backup of current database...")
		if _, err := m.CreateBackup(); err != nil {
			return fmt.Errorf("failed to backup current database: %w", err)
		}
	}

	// 解压（如果需要）
	sourceFile := backupFile
	tempFile := ""
	
	if strings.HasSuffix(backupFile, ".gz") {
		tempFile = strings.TrimSuffix(backupFile, ".gz")
		if err := m.decompressFile(backupFile, tempFile); err != nil {
			return fmt.Errorf("failed to decompress backup: %w", err)
		}
		sourceFile = tempFile
		defer os.Remove(tempFile)
	}

	// 恢复数据库
	source, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(m.config.DBPath)
	if err != nil {
		return fmt.Errorf("failed to create database file: %w", err)
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}

	fmt.Println("Database restored successfully")
	return nil
}

// decompressFile 解压文件
func (m *BackupManager) decompressFile(sourceFile, destFile string) error {
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	gzReader, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	dest, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, gzReader)
	return err
}

// CreateZipBackup 创建包含数据库和配置的 ZIP 备份
func (m *BackupManager) CreateZipBackup() (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	zipFile := filepath.Join(m.config.BackupDir, fmt.Sprintf("rolecraft_backup_%s.zip", timestamp))

	// 创建 ZIP 文件
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		return "", err
	}
	defer zipWriter.Close()

	zipArchive := zip.NewWriter(zipWriter)
	defer zipArchive.Close()

	// 添加数据库文件
	if err := m.addToZip(zipArchive, m.config.DBPath, "database/rolecraft.db"); err != nil {
		return "", err
	}

	// 添加配置文件（如果存在）
	configFiles := []string{
		".env",
		"config.json",
	}
	for _, configFile := range configFiles {
		configPath := filepath.Join(filepath.Dir(m.config.DBPath), configFile)
		if _, err := os.Stat(configPath); err == nil {
			m.addToZip(zipArchive, configPath, "config/"+filepath.Base(configPath))
		}
	}

	return zipFile, nil
}

// addToZip 添加文件到 ZIP
func (m *BackupManager) addToZip(zipWriter *zip.Writer, sourcePath, destPath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(destPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// ExportToJSON 导出备份信息为 JSON
func (m *BackupManager) ExportToJSON(outputFile string) error {
	backups, err := m.ListBackups()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(backups, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, data, 0644)
}

// ScheduleBackup 定时备份（需要配合 cron 使用）
func ScheduleBackup(ctx context.Context, interval time.Duration, callback func(*BackupResult)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			manager := NewBackupManager(&BackupConfig{})
			result, err := manager.CreateBackup()
			if callback != nil {
				callback(result)
			}
			if err != nil {
				fmt.Printf("Scheduled backup failed: %v\n", err)
			}
		}
	}
}

func main() {
	// 示例：创建备份
	config := &BackupConfig{
		BackupDir:     "./backups",
		DBPath:        "./rolecraft.db",
		DBType:        "sqlite",
		RetentionDays: 30,
		Compression:   true,
	}

	manager := NewBackupManager(config)

	// 创建备份
	result, err := manager.CreateBackup()
	if err != nil {
		fmt.Printf("Backup failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Backup successful: %s (%d bytes, %s)\n", result.File, result.Size, result.Duration)

	// 列出备份
	backups, _ := manager.ListBackups()
	fmt.Printf("\nTotal backups: %d\n", len(backups))
	for _, backup := range backups {
		fmt.Printf("  - %s (%d bytes)\n", backup.File, backup.Size)
	}
}
