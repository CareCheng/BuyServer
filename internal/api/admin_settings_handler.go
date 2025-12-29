package api

import (
	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// ==================== 系统设置相关 API ====================

// AdminGetSettings 获取系统设置（从数据库）
func AdminGetSettings(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	sysCfg, err := ConfigSvc.GetSystemConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取配置失败"})
		return
	}

	// 获取服务器端口配置
	serverPort, _ := ConfigSvc.GetServerPort()
	if serverPort <= 0 {
		serverPort = 8080
	}

	c.JSON(200, gin.H{
		"success": true,
		"settings": gin.H{
			"system_title":            sysCfg.SystemTitle,
			"admin_suffix":            sysCfg.AdminSuffix,
			"enable_login":            sysCfg.EnableLogin,
			"admin_username":          sysCfg.AdminUsername,
			"enable_2fa":              sysCfg.Enable2FA,
			"totp_secret":             sysCfg.TOTPSecret,
			"server_port":             serverPort,
		},
	})
}

// AdminSaveSettings 保存系统设置（到数据库）
func AdminSaveSettings(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	var req struct {
		SystemTitle string `json:"system_title"`
		AdminSuffix string `json:"admin_suffix"`
		ServerPort  int    `json:"server_port"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取当前配置，如果获取失败则使用默认配置
	sysCfg, err := ConfigSvc.GetSystemConfig()
	if err != nil || sysCfg == nil {
		sysCfg = &service.SystemConfig{
			SystemTitle:     config.GlobalConfig.ServerConfig.SystemTitle,
			AdminSuffix:     config.GlobalConfig.ServerConfig.AdminSuffix,
			EnableLogin:     config.GlobalConfig.ServerConfig.EnableLogin,
			AdminUsername:   config.GlobalConfig.ServerConfig.AdminUsername,
			AdminPassword:   config.GlobalConfig.ServerConfig.AdminPassword,
			Enable2FA:       config.GlobalConfig.ServerConfig.Enable2FA,
			TOTPSecret:      config.GlobalConfig.ServerConfig.TOTPSecret,
			EnableWhitelist: false,
			IPWhitelist:     []string{},
		}
	}

	// 更新基本设置字段
	if req.SystemTitle != "" {
		sysCfg.SystemTitle = req.SystemTitle
	}
	if req.AdminSuffix != "" {
		sysCfg.AdminSuffix = req.AdminSuffix
	}
	if err := ConfigSvc.SaveSystemConfig(sysCfg); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存配置失败: " + err.Error()})
		return
	}

	// 保存服务器端口配置
	needRestart := false
	if req.ServerPort > 0 && req.ServerPort <= 65535 {
		currentPort, _ := ConfigSvc.GetServerPort()
		if currentPort != req.ServerPort {
			if err := ConfigSvc.SaveServerPort(req.ServerPort); err != nil {
				c.JSON(500, gin.H{"success": false, "error": "保存端口配置失败: " + err.Error()})
				return
			}
			needRestart = true
		}
	}

	// 同步更新全局配置
	config.GlobalConfig.ServerConfig.SystemTitle = sysCfg.SystemTitle
	config.GlobalConfig.ServerConfig.AdminSuffix = sysCfg.AdminSuffix

	message := "设置已保存"
	if needRestart {
		message = "设置已保存，端口更改需要重启程序后生效"
	}
	c.JSON(200, gin.H{"success": true, "message": message, "need_restart": needRestart})
}

// AdminSaveSecuritySettings 保存安全设置（到数据库）
func AdminSaveSecuritySettings(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	var req struct {
		EnableLogin   bool   `json:"enable_login"`
		AdminUsername string `json:"admin_username"`
		AdminPassword string `json:"admin_password"`
		Enable2FA     bool   `json:"enable_2fa"`
		TOTPSecret    string `json:"totp_secret"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取当前配置，如果获取失败则使用默认配置
	sysCfg, err := ConfigSvc.GetSystemConfig()
	if err != nil || sysCfg == nil {
		sysCfg = &service.SystemConfig{
			SystemTitle:     config.GlobalConfig.ServerConfig.SystemTitle,
			AdminSuffix:     config.GlobalConfig.ServerConfig.AdminSuffix,
			EnableLogin:     true,
			AdminUsername:   "admin",
			AdminPassword:   "admin123",
			Enable2FA:       false,
			EnableWhitelist: false,
			IPWhitelist:     []string{},
		}
	}

	// 检查新用户名是否与 admins 表中的用户名冲突
	if req.AdminUsername != "" && req.AdminUsername != sysCfg.AdminUsername {
		if model.DBConnected && RoleSvc != nil {
			if _, err := RoleSvc.GetAdminByUsername(req.AdminUsername); err == nil {
				c.JSON(400, gin.H{"success": false, "error": "用户名与多管理员系统中的账户冲突，请使用其他用户名"})
				return
			}
		}
	}

	// 更新安全相关字段
	sysCfg.EnableLogin = req.EnableLogin
	if req.AdminUsername != "" {
		sysCfg.AdminUsername = req.AdminUsername
	}
	if req.AdminPassword != "" {
		sysCfg.AdminPassword = req.AdminPassword
	}
	sysCfg.Enable2FA = req.Enable2FA
	sysCfg.TOTPSecret = req.TOTPSecret

	if err := ConfigSvc.SaveSystemConfig(sysCfg); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存配置失败: " + err.Error()})
		return
	}

	// 同步更新全局配置
	config.GlobalConfig.ServerConfig.EnableLogin = sysCfg.EnableLogin
	config.GlobalConfig.ServerConfig.AdminUsername = sysCfg.AdminUsername
	if req.AdminPassword != "" {
		config.GlobalConfig.ServerConfig.AdminPassword = sysCfg.AdminPassword
	}
	config.GlobalConfig.ServerConfig.Enable2FA = sysCfg.Enable2FA
	config.GlobalConfig.ServerConfig.TOTPSecret = sysCfg.TOTPSecret

	c.JSON(200, gin.H{"success": true, "message": "安全设置已保存"})
}

// AdminGenerate2FASecret 生成2FA密钥
func AdminGenerate2FASecret(c *gin.Context) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.GlobalConfig.ServerConfig.SystemTitle,
		AccountName: config.GlobalConfig.ServerConfig.AdminUsername,
	})
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "生成密钥失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"secret":  key.Secret(),
		"url":     key.URL(),
	})
}

// AdminVerify2FACode 验证2FA验证码
func AdminVerify2FACode(c *gin.Context) {
	var req struct {
		Code   string `json:"code" binding:"required"`
		Secret string `json:"secret" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if totp.Validate(req.Code, req.Secret) {
		c.JSON(200, gin.H{"success": true, "message": "验证通过"})
	} else {
		c.JSON(400, gin.H{"success": false, "error": "验证码错误"})
	}
}
