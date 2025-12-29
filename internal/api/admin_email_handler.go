package api

import (
	"user-frontend/internal/config"

	"github.com/gin-gonic/gin"
)

// ==================== 邮箱配置相关 API ====================

// AdminGetEmailConfig 获取邮箱配置（从数据库）
func AdminGetEmailConfig(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	emailCfg, err := ConfigSvc.GetEmailConfig()
	if err != nil {
		emailCfg = &config.EmailConfig{SMTPPort: 465, Encryption: "ssl", CodeLength: 6}
	}

	c.JSON(200, gin.H{
		"success": true,
		"config": gin.H{
			"enabled":      emailCfg.Enabled,
			"smtp_host":    emailCfg.SMTPHost,
			"smtp_port":    emailCfg.SMTPPort,
			"smtp_user":    emailCfg.SMTPUser,
			"has_password": emailCfg.SMTPPassword != "",
			"from_name":    emailCfg.FromName,
			"from_email":   emailCfg.FromEmail,
			"encryption":   emailCfg.Encryption,
			"code_length":  emailCfg.CodeLength,
		},
	})
}

// AdminSaveEmailConfig 保存邮箱配置（到数据库）
func AdminSaveEmailConfig(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	var req struct {
		Enabled      bool   `json:"enabled"`
		SMTPHost     string `json:"smtp_host"`
		SMTPPort     int    `json:"smtp_port"`
		SMTPUser     string `json:"smtp_user"`
		SMTPPassword string `json:"smtp_password"`
		FromName     string `json:"from_name"`
		FromEmail    string `json:"from_email"`
		Encryption   string `json:"encryption"` // none/ssl/starttls
		CodeLength   int    `json:"code_length"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取现有配置以保留密码
	existingCfg, _ := ConfigSvc.GetEmailConfig()

	// 验证加密方式
	encryption := req.Encryption
	if encryption != "none" && encryption != "ssl" && encryption != "starttls" {
		encryption = "ssl" // 默认使用 SSL
	}

	emailCfg := &config.EmailConfig{
		Enabled:    req.Enabled,
		SMTPHost:   req.SMTPHost,
		SMTPPort:   req.SMTPPort,
		SMTPUser:   req.SMTPUser,
		FromName:   req.FromName,
		FromEmail:  req.FromEmail,
		Encryption: encryption,
	}

	// 密码处理：如果提供了新密码则使用，否则保留旧密码
	if req.SMTPPassword != "" {
		emailCfg.SMTPPassword = req.SMTPPassword
	} else if existingCfg != nil {
		emailCfg.SMTPPassword = existingCfg.SMTPPassword
	}

	// 验证码长度，默认6位，范围4-8
	if req.CodeLength >= 4 && req.CodeLength <= 8 {
		emailCfg.CodeLength = req.CodeLength
	} else {
		emailCfg.CodeLength = 6
	}

	if err := ConfigSvc.SaveEmailConfig(emailCfg); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存配置失败: " + err.Error()})
		return
	}

	// 更新全局配置和EmailService
	config.GlobalConfig.EmailConfig = *emailCfg
	if EmailSvc != nil {
		// 使用 UpdateConfig 更新配置，避免重建服务导致 repo 丢失
		EmailSvc.UpdateConfig(emailCfg)
	}

	c.JSON(200, gin.H{"success": true, "message": "邮箱配置已保存"})
}

// AdminTestEmail 测试邮箱发送
func AdminTestEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if EmailSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "邮箱服务未初始化"})
		return
	}

	if err := EmailSvc.SendVerifyCode(req.Email, "test"); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "发送失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "测试邮件已发送"})
}
