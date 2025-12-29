package service

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

type BackupService struct {
	repo      *repository.Repository
	backupDir string
}

func NewBackupService(repo *repository.Repository, baseDir string) *BackupService {
	backupDir := filepath.Join(baseDir, "backups")
	os.MkdirAll(backupDir, 0755)
	return &BackupService{
		repo:      repo,
		backupDir: backupDir,
	}
}

// CreateBackup 创建数据库备份
func (s *BackupService) CreateBackup(dbConfig *config.DBConfig, createdBy, remark string) (*model.DatabaseBackup, error) {
	timestamp := time.Now().Format("20060102_150405")
	var filename string
	var filePath string
	var fileSize int64

	switch dbConfig.Type {
	case "sqlite":
		filename = fmt.Sprintf("backup_sqlite_%s.zip", timestamp)
		filePath = filepath.Join(s.backupDir, filename)
		size, err := s.backupSQLite(dbConfig.Database, filePath)
		if err != nil {
			return nil, fmt.Errorf("SQLite备份失败: %v", err)
		}
		fileSize = size

	case "mysql":
		filename = fmt.Sprintf("backup_mysql_%s.sql", timestamp)
		filePath = filepath.Join(s.backupDir, filename)
		size, err := s.backupMySQL(dbConfig, filePath)
		if err != nil {
			return nil, fmt.Errorf("MySQL备份失败: %v", err)
		}
		fileSize = size

	case "postgres":
		filename = fmt.Sprintf("backup_postgres_%s.sql", timestamp)
		filePath = filepath.Join(s.backupDir, filename)
		size, err := s.backupPostgres(dbConfig, filePath)
		if err != nil {
			return nil, fmt.Errorf("PostgreSQL备份失败: %v", err)
		}
		fileSize = size

	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbConfig.Type)
	}

	// 保存备份记录
	backup := &model.DatabaseBackup{
		Filename:  filename,
		FilePath:  filePath,
		FileSize:  fileSize,
		DBType:    dbConfig.Type,
		Remark:    remark,
		CreatedBy: createdBy,
	}

	if err := s.repo.CreateBackupRecord(backup); err != nil {
		// 删除已创建的备份文件
		os.Remove(filePath)
		return nil, err
	}

	return backup, nil
}

// backupSQLite 备份SQLite数据库
func (s *BackupService) backupSQLite(dbPath, backupPath string) (int64, error) {
	// 创建zip文件
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return 0, err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 打开源数据库文件
	srcFile, err := os.Open(dbPath)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	// 获取文件信息
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return 0, err
	}

	// 创建zip条目
	header, err := zip.FileInfoHeader(srcInfo)
	if err != nil {
		return 0, err
	}
	header.Name = filepath.Base(dbPath)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return 0, err
	}

	// 复制数据
	_, err = io.Copy(writer, srcFile)
	if err != nil {
		return 0, err
	}

	// 关闭zip写入器以获取正确的文件大小
	zipWriter.Close()
	zipFile.Close()

	// 获取备份文件大小
	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		return 0, err
	}

	return backupInfo.Size(), nil
}

// backupMySQL 备份MySQL数据库（使用SQL导出）
func (s *BackupService) backupMySQL(cfg *config.DBConfig, backupPath string) (int64, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	return s.exportToSQL(db, cfg.Database, backupPath, "mysql")
}

// backupPostgres 备份PostgreSQL数据库（使用SQL导出）
func (s *BackupService) backupPostgres(cfg *config.DBConfig, backupPath string) (int64, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	return s.exportToSQL(db, cfg.Database, backupPath, "postgres")
}

// exportToSQL 导出数据库为SQL文件
func (s *BackupService) exportToSQL(db *sql.DB, dbName, backupPath, dbType string) (int64, error) {
	file, err := os.Create(backupPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// 写入头部信息
	header := fmt.Sprintf("-- Database Backup\n-- Database: %s\n-- Type: %s\n-- Created: %s\n\n",
		dbName, dbType, time.Now().Format("2006-01-02 15:04:05"))
	file.WriteString(header)

	// 获取所有表
	var tables []string
	var query string
	if dbType == "mysql" {
		query = "SHOW TABLES"
	} else {
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public'"
	}

	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
	}

	// 导出每个表的数据
	for _, table := range tables {
		file.WriteString(fmt.Sprintf("\n-- Table: %s\n", table))

		// 获取表数据
		dataRows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			continue
		}

		columns, _ := dataRows.Columns()
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		for dataRows.Next() {
			dataRows.Scan(valuePtrs...)

			var valueStrings []string
			for _, v := range values {
				if v == nil {
					valueStrings = append(valueStrings, "NULL")
				} else {
					switch val := v.(type) {
					case []byte:
						valueStrings = append(valueStrings, fmt.Sprintf("'%s'", strings.ReplaceAll(string(val), "'", "''")))
					case string:
						valueStrings = append(valueStrings, fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''")))
					case time.Time:
						valueStrings = append(valueStrings, fmt.Sprintf("'%s'", val.Format("2006-01-02 15:04:05")))
					default:
						valueStrings = append(valueStrings, fmt.Sprintf("%v", val))
					}
				}
			}

			insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n",
				table, strings.Join(columns, ", "), strings.Join(valueStrings, ", "))
			file.WriteString(insertSQL)
		}
		dataRows.Close()
	}

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

// GetAllBackups 获取所有备份记录
func (s *BackupService) GetAllBackups() ([]model.DatabaseBackup, error) {
	return s.repo.GetAllBackups()
}

// GetBackupByID 根据ID获取备份记录
func (s *BackupService) GetBackupByID(id uint) (*model.DatabaseBackup, error) {
	return s.repo.GetBackupByID(id)
}

// DeleteBackup 删除备份
func (s *BackupService) DeleteBackup(id uint) error {
	backup, err := s.repo.GetBackupByID(id)
	if err != nil {
		return err
	}

	// 删除文件
	if backup.FilePath != "" {
		os.Remove(backup.FilePath)
	}

	// 删除记录
	return s.repo.DeleteBackupRecord(id)
}

// GetBackupFilePath 获取备份文件路径
func (s *BackupService) GetBackupFilePath(id uint) (string, string, error) {
	backup, err := s.repo.GetBackupByID(id)
	if err != nil {
		return "", "", err
	}

	// 检查文件是否存在
	if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("备份文件不存在")
	}

	return backup.FilePath, backup.Filename, nil
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// ExportDataAsJSON 导出数据为JSON格式（用于简单备份）
func (s *BackupService) ExportDataAsJSON(createdBy string) (*model.DatabaseBackup, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_json_%s.json", timestamp)
	filePath := filepath.Join(s.backupDir, filename)

	// 收集所有数据
	data := make(map[string]interface{})

	// 这里可以根据需要导出各种数据
	// 示例：导出基本统计信息
	data["export_time"] = time.Now().Format("2006-01-02 15:04:05")
	data["export_by"] = createdBy

	// 写入JSON文件
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return nil, err
	}

	fileInfo, _ := os.Stat(filePath)

	backup := &model.DatabaseBackup{
		Filename:  filename,
		FilePath:  filePath,
		FileSize:  fileInfo.Size(),
		DBType:    "json",
		Remark:    "JSON数据导出",
		CreatedBy: createdBy,
	}

	if err := s.repo.CreateBackupRecord(backup); err != nil {
		os.Remove(filePath)
		return nil, err
	}

	return backup, nil
}
