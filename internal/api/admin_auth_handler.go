package api

import (
	"log"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// ==================== 管理员认证相关 API ====================

// CheckInitialSetup 检查是否需要初始化设置
func CheckInitialSetup(c *gin.Context) {
	needsSetup := false
	
	if ConfigSvc != nil {
		needsSetup = ConfigSvc.NeedsInitialSetup()
		log.Printf("[CheckInitialSetup] ConfigSvc存在, needsSetup=%v", needsSetup)
	} else {
		// ConfigSvc 未初始化，检查全局配置
		cfg := config.GlobalConfig.ServerConfig
		needsSetup = cfg.AdminPassword == "admin123" || cfg.AdminPassword == ""
		log.Printf("[CheckInitialSetup] ConfigSvc为nil, 检查GlobalConfig, password=%s, needsSetup=%v", cfg.AdminPassword, needsSetup)
	}
	
	c.JSON(200, gin.H{
		"success":      true,
		"needs_setup":  needsSetup,
	})
}

// SetInitialPassword 设置初始管理员密码
func SetInitialPassword(c *gin.Context) {
	var req struct {
		Password        string `json:"password" binding:"required,min=6"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "密码长度至少6位"})
		return
	}
	
	if req.Password != req.ConfirmPassword {
		c.JSON(400, gin.H{"success": false, "error": "两次输入的密码不一致"})
		return
	}
	
	// 检查是否需要初始化
	needsSetup := false
	if ConfigSvc != nil {
		needsSetup = ConfigSvc.NeedsInitialSetup()
	} else {
		cfg := config.GlobalConfig.ServerConfig
		needsSetup = cfg.AdminPassword == "admin123" || cfg.AdminPassword == ""
	}
	
	if !needsSetup {
		c.JSON(400, gin.H{"success": false, "error": "初始密码已设置，无法重复设置"})
		return
	}
	
	// 设置密码到配置
	if ConfigSvc != nil {
		if err := ConfigSvc.SetInitialPassword(req.Password); err != nil {
			c.JSON(500, gin.H{"success": false, "error": err.Error()})
			return
		}
	} else {
		// 更新全局配置（内存中）
		config.GlobalConfig.ServerConfig.AdminPassword = req.Password
	}

	// 同时创建或更新数据库中的管理员记录
	if model.DBConnected {
		// 优先使用 RoleSvc 创建新的 Admin 记录（角色权限系统）
		if RoleSvc != nil {
			// 检查是否已存在 admin 用户
			existingAdmin, _ := RoleSvc.GetAdminByUsername("admin")
			if existingAdmin == nil {
				// 创建默认超级管理员
				if err := RoleSvc.CreateSuperAdmin("admin", req.Password); err != nil {
					// 记录错误但不中断，配置已保存成功
					c.JSON(200, gin.H{
						"success": true,
						"message": "管理员密码设置成功（注意：数据库管理员创建失败，请使用配置文件登录）",
						"warning": err.Error(),
					})
					return
				}
			} else {
				// 更新现有管理员密码
				if err := RoleSvc.UpdateAdminPassword(existingAdmin.ID, req.Password); err != nil {
					c.JSON(200, gin.H{
						"success": true,
						"message": "管理员密码设置成功（注意：数据库密码更新失败）",
						"warning": err.Error(),
					})
					return
				}
			}
		} else if AdminSvc != nil {
			// 回退到旧的 AdminService
			if err := AdminSvc.InitDefaultAdmin("admin", req.Password); err != nil {
				// 记录错误但不中断
				c.JSON(200, gin.H{
					"success": true,
					"message": "管理员密码设置成功",
					"warning": err.Error(),
				})
				return
			}
		}
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"message": "管理员密码设置成功",
	})
}

// AdminLoginPage 管理员登录页面
func AdminLoginPage(c *gin.Context) {
	c.HTML(200, "admin_login.html", gin.H{
		"title": "管理员登录",
	})
}

// AdminTOTPPage 管理员TOTP验证页面
func AdminTOTPPage(c *gin.Context) {
	c.HTML(200, "admin_totp.html", gin.H{
		"title": "两步验证",
	})
}

// AdminLogin 管理员登录
func AdminLogin(c *gin.Context) {
	// 检查是否需要初始化设置
	if ConfigSvc != nil && ConfigSvc.NeedsInitialSetup() {
		c.JSON(400, gin.H{"success": false, "error": "请先完成初始化设置", "needs_setup": true})
		return
	}

	if !model.DBConnected {
		// 数据库未连接时使用配置文件中的管理员账号
		var req struct {
			Username    string `json:"username" binding:"required"`
			Password    string `json:"password" binding:"required"`
			CaptchaID   string `json:"captcha_id"`
			CaptchaCode string `json:"captcha_code"`
			Remember    bool   `json:"remember"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": "参数错误"})
			return
		}

		// 验证验证码
		if req.CaptchaID != "" && req.CaptchaCode != "" {
			if !VerifyCaptchaCode(req.CaptchaID, req.CaptchaCode) {
				c.JSON(400, gin.H{"success": false, "error": "验证码错误"})
				return
			}
		}

		cfg := config.GlobalConfig
		if req.Username != cfg.ServerConfig.AdminUsername || req.Password != cfg.ServerConfig.AdminPassword {
			c.JSON(400, gin.H{"success": false, "error": "用户名或密码错误"})
			return
		}

		// 创建会话（数据库持久化）
		if SessionSvc == nil {
			c.JSON(500, gin.H{"success": false, "error": "会话服务未初始化"})
			return
		}
		sessionID, err := SessionSvc.CreateAdminSession(req.Username, "super_admin", c.ClientIP(), c.GetHeader("User-Agent"), req.Remember)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "创建会话失败"})
			return
		}

		// 如果未启用2FA则直接验证通过
		if !cfg.ServerConfig.Enable2FA {
			SessionSvc.SetAdminSessionVerified(sessionID)
		}

		maxAge := 3600
		if req.Remember {
			maxAge = 86400
		}
		SetSecureCookie(c, "admin_session", sessionID, maxAge, true)
		SetCSRFCookie(c, sessionID)

		if cfg.ServerConfig.Enable2FA {
			c.JSON(200, gin.H{
				"success":      true,
				"require_totp": true,
				"message":      "请完成两步验证",
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": "登录成功"})
		return
	}

	// 数据库已连接，使用数据库中的管理员账号
	if AdminSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		CaptchaID   string `json:"captcha_id"`
		CaptchaCode string `json:"captcha_code"`
		Remember    bool   `json:"remember"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 验证验证码
	if req.CaptchaID != "" && req.CaptchaCode != "" {
		if !VerifyCaptchaCode(req.CaptchaID, req.CaptchaCode) {
			c.JSON(400, gin.H{"success": false, "error": "验证码错误"})
			return
		}
	}

	// 优先从新的 admins 表验证（角色权限系统）
	var adminUsername string
	var adminRole string
	var enable2FA bool
	var verified bool

	if RoleSvc != nil {
		newAdmin, err := RoleSvc.VerifyAdminPassword(req.Username, req.Password)
		if err == nil {
			// 新表验证成功
			adminUsername = newAdmin.Username
			if newAdmin.Role != nil {
				adminRole = newAdmin.Role.Name
			} else {
				adminRole = "admin"
			}
			enable2FA = newAdmin.Enable2FA
			verified = true
			// 更新登录信息
			RoleSvc.UpdateAdminLoginInfo(newAdmin.ID, c.ClientIP())
		}
	}

	// 新表验证失败，尝试旧的 admin_users 表
	if !verified && AdminSvc != nil {
		oldAdmin, err := AdminSvc.Login(req.Username, req.Password, c.ClientIP())
		if err == nil {
			adminUsername = oldAdmin.Username
			adminRole = oldAdmin.Role
			enable2FA = oldAdmin.Enable2FA
			verified = true
		}
	}

	// 数据库表都没有记录，回退到配置文件验证
	if !verified {
		cfg := config.GlobalConfig
		if req.Username == cfg.ServerConfig.AdminUsername && req.Password == cfg.ServerConfig.AdminPassword {
			adminUsername = cfg.ServerConfig.AdminUsername
			adminRole = "super_admin"
			enable2FA = cfg.ServerConfig.Enable2FA
			verified = true
		}
	}

	// 所有验证方式都失败
	if !verified {
		c.JSON(400, gin.H{"success": false, "error": "用户名或密码错误"})
		return
	}

	// 创建会话（数据库持久化）
	if SessionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "会话服务未初始化"})
		return
	}
	sessionID, err := SessionSvc.CreateAdminSession(adminUsername, adminRole, c.ClientIP(), c.GetHeader("User-Agent"), req.Remember)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建会话失败"})
		return
	}

	// 如果未启用2FA则直接验证通过
	if !enable2FA {
		SessionSvc.SetAdminSessionVerified(sessionID)
	}

	maxAge := 3600
	if req.Remember {
		maxAge = 86400
	}
	SetSecureCookie(c, "admin_session", sessionID, maxAge, true)
	SetCSRFCookie(c, sessionID)

	if enable2FA {
		c.JSON(200, gin.H{
			"success":      true,
			"require_totp": true,
			"message":      "请完成两步验证",
		})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "登录成功"})
}

// AdminVerifyTOTP 管理员TOTP验证
func AdminVerifyTOTP(c *gin.Context) {
	sessionID, _ := c.Cookie("admin_session")
	if sessionID == "" {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	if SessionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "会话服务未初始化"})
		return
	}

	session, err := SessionSvc.GetAdminSession(sessionID)
	if err != nil {
		c.JSON(401, gin.H{"success": false, "error": "会话已过期"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取TOTP密钥
	var totpSecret string
	if model.DBConnected {
		_, secret, _ := AdminSvc.GetAdmin2FAStatus(session.Username)
		totpSecret = secret
	} else {
		totpSecret = config.GlobalConfig.ServerConfig.TOTPSecret
	}

	if !totp.Validate(req.Code, totpSecret) {
		c.JSON(400, gin.H{"success": false, "error": "验证码错误"})
		return
	}

	// 更新会话状态
	SessionSvc.SetAdminSessionVerified(sessionID)

	c.JSON(200, gin.H{"success": true, "message": "验证成功"})
}

// AdminLogout 管理员登出
func AdminLogout(c *gin.Context) {
	sessionID, _ := c.Cookie("admin_session")
	if sessionID != "" && SessionSvc != nil {
		SessionSvc.DeleteAdminSession(sessionID)
	}

	// 清除Cookie
	c.SetCookie("admin_session", "", -1, "/", "", false, true)
	c.SetCookie("csrf_token", "", -1, "/", "", false, false)

	c.JSON(200, gin.H{"success": true, "message": "已退出登录"})
}

// AdminInfo 获取当前管理员信息
func AdminInfo(c *gin.Context) {
	username := c.GetString("admin_username")
	role := c.GetString("admin_role")

	if username == "" {
		c.JSON(401, gin.H{"success": false, "error": "未登录"})
		return
	}

	// 获取权限列表
	var permissions []string
	isSuperAdmin := false

	// 超级管理员（系统配置账户或 super_admin 角色）拥有所有权限
	if role == "super_admin" {
		isSuperAdmin = true
		for _, p := range model.AllPermissions {
			permissions = append(permissions, p.Code)
		}
	} else if model.DBConnected && RoleSvc != nil {
		// 从数据库获取管理员权限
		admin, err := RoleSvc.GetAdminByUsername(username)
		if err == nil && admin != nil {
			perms, err := RoleSvc.GetRolePermissions(admin.RoleID)
			if err == nil {
				permissions = perms
			}
			// 检查是否是超级管理员角色
			if admin.Role != nil && admin.Role.Name == "super_admin" {
				isSuperAdmin = true
				permissions = make([]string, 0, len(model.AllPermissions))
				for _, p := range model.AllPermissions {
					permissions = append(permissions, p.Code)
				}
			}
		}
	}

	// 如果没有获取到权限，给予基本的仪表盘查看权限
	if len(permissions) == 0 && !isSuperAdmin {
		permissions = []string{"dashboard:view"}
	}

	c.JSON(200, gin.H{
		"success": true,
		"admin": gin.H{
			"username":       username,
			"role":           role,
			"permissions":    permissions,
			"is_super_admin": isSuperAdmin,
		},
	})
}

// AdminEnable2FA 启用管理员2FA
func AdminEnable2FA(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	username := c.GetString("admin_username")

	var req struct {
		Secret string `json:"secret" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if !totp.Validate(req.Code, req.Secret) {
		c.JSON(400, gin.H{"success": false, "error": "验证码错误"})
		return
	}

	if err := AdminSvc.Enable2FA(username, req.Secret); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "两步验证已启用"})
}

// AdminDisable2FA 禁用管理员2FA
func AdminDisable2FA(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	username := c.GetString("admin_username")

	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 验证密码
	admin, _ := AdminSvc.GetAdminByUsername(username)
	if admin == nil || !utils.CheckPassword(req.Password, admin.PasswordHash) {
		c.JSON(400, gin.H{"success": false, "error": "密码错误"})
		return
	}
	if err := AdminSvc.Disable2FA(username); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "两步验证已禁用"})
}

// AdminGet2FAStatus 获取管理员2FA状态
func AdminGet2FAStatus(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	username := c.GetString("admin_username")
	enabled, _, _ := AdminSvc.GetAdmin2FAStatus(username)

	c.JSON(200, gin.H{
		"success": true,
		"enabled": enabled,
	})
}
