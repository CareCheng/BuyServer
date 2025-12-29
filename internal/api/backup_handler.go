package api

import (
	"net/http"
	"strconv"

	"user-frontend/internal/config"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// ==================== 数据库备份管理 ====================

// AdminGetBackups 获取备份列表
func AdminGetBackups(c *gin.Context) {
	if BackupSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	backups, err := BackupSvc.GetAllBackups()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 格式化文件大小
	type BackupInfo struct {
		ID           uint   `json:"id"`
		Filename     string `json:"filename"`
		FileSizeText string `json:"file_size_text"`
		FileSize     int64  `json:"file_size"`
		DBType       string `json:"db_type"`
		Remark       string `json:"remark"`
		CreatedBy    string `json:"created_by"`
		CreatedAt    string `json:"created_at"`
	}

	var backupInfos []BackupInfo
	for _, b := range backups {
		backupInfos = append(backupInfos, BackupInfo{
			ID:           b.ID,
			Filename:     b.Filename,
			FileSizeText: service.FormatFileSize(b.FileSize),
			FileSize:     b.FileSize,
			DBType:       b.DBType,
			Remark:       b.Remark,
			CreatedBy:    b.CreatedBy,
			CreatedAt:    b.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(200, gin.H{"success": true, "backups": backupInfos})
}

// AdminCreateBackup 创建备份
func AdminCreateBackup(c *gin.Context) {
	if BackupSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Remark string `json:"remark"`
	}
	c.ShouldBindJSON(&req)

	// 获取数据库配置
	dbConfig := &config.GlobalConfig.DBConfig
	if dbConfig.Type == "" {
		c.JSON(500, gin.H{"success": false, "error": "数据库配置不存在"})
		return
	}

	adminUsername := c.GetString("admin_username")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	backup, err := BackupSvc.CreateBackup(dbConfig, adminUsername, req.Remark)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "create", "backup", strconv.Itoa(int(backup.ID)), backup.Filename, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"backup": gin.H{
			"id":             backup.ID,
			"filename":       backup.Filename,
			"file_size_text": service.FormatFileSize(backup.FileSize),
			"file_size":      backup.FileSize,
			"db_type":        backup.DBType,
			"remark":         backup.Remark,
			"created_by":     backup.CreatedBy,
			"created_at":     backup.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		"message": "备份创建成功",
	})
}

// AdminDownloadBackup 下载备份
func AdminDownloadBackup(c *gin.Context) {
	if BackupSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	filePath, filename, err := BackupSvc.GetBackupFilePath(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(c.GetString("admin_username"), "download", "backup", idStr, filename, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.File(filePath)
}

// AdminDeleteBackup 删除备份
func AdminDeleteBackup(c *gin.Context) {
	if BackupSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := BackupSvc.DeleteBackup(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(c.GetString("admin_username"), "delete", "backup", idStr, "", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "备份已删除"})
}

// AdminGetBackupInfo 获取备份信息（数据库类型等）
func AdminGetBackupInfo(c *gin.Context) {
	dbConfig := &config.GlobalConfig.DBConfig
	if dbConfig.Type == "" {
		c.JSON(200, gin.H{
			"success":      true,
			"db_type":      "unknown",
			"db_connected": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"db_type":      dbConfig.Type,
		"db_connected": true,
	})
}
