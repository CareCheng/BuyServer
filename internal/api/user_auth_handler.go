// Package api 提供 HTTP API 处理器
// user_auth_handler.go - 用户认证相关处理器（登录、注册、登出）
package api

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"user-frontend/internal/config"
	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// IndexPage 首页
func IndexPage(c *gin.Context) {
	c.HTML(200, "index.html", gin.H{
		"title": config.GlobalConfig.ServerConfig.SystemTitle,
	})
}

// UserLoginPage 用户登录页面
func UserLoginPage(c *gin.Context) {
	c.HTML(200, "login.html", gin.H{
		"title": "用户登录",
	})
}

// UserRegisterPage 用户注册页面
func UserRegisterPage(c *gin.Context) {
	c.HTML(200, "register.html", gin.H{
		"title": "用户注册",
	})
}

// ProductListPage 商品列表页面
func ProductListPage(c *gin.Context) {
	c.HTML(200, "products.html", gin.H{
		"title": "商品列表",
	})
}

// UserCenterPage 用户中心页面
func UserCenterPage(c *gin.Context) {
	c.HTML(200, "user_center.html", gin.H{
		"title": "个人中心",
	})
}

// UserRegister 用户注册
func UserRegister(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Username        string `json:"username" binding:"required"`
		Email           string `json:"email" binding:"required"`
		EmailCode       string `json:"email_code" binding:"required"`
		Password        string `json:"password" binding:"required"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
		Phone           string `json:"phone"`
		CaptchaID       string `json:"captcha_id"`
		CaptchaCode     string `json:"captcha_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请填写完整信息（包括邮箱和验证码）"})
		return
	}

	// 验证图形验证码
	if req.CaptchaID != "" && req.CaptchaCode != "" {
		if !VerifyCaptchaCode(req.CaptchaID, req.CaptchaCode) {
			c.JSON(400, gin.H{"success": false, "error": "图形验证码错误"})
			return
		}
	}

	// 验证邮箱验证码
	if EmailSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "邮箱服务未初始化"})
		return
	}
	if !EmailSvc.VerifyCode(req.Email, req.EmailCode, "register") {
		c.JSON(400, gin.H{"success": false, "error": "邮箱验证码错误或已过期"})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(400, gin.H{"success": false, "error": "两次密码不一致"})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(400, gin.H{"success": false, "error": "密码长度至少6位"})
		return
	}

	user, err := UserSvc.RegisterWithVerifiedEmail(req.Username, req.Email, req.Password, req.Phone)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "注册成功",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

// UserLogin 用户登录
func UserLogin(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	clientIP := c.ClientIP()

	// 检查IP是否被锁定
	if SecuritySvc != nil {
		if locked, remaining := SecuritySvc.IsLoginLocked(clientIP); locked {
			c.JSON(429, gin.H{
				"success":   false,
				"error":     "登录尝试次数过多，请稍后再试",
				"locked":    true,
				"remaining": int(remaining.Minutes()),
			})
			return
		}
	}

	var req struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		CaptchaID   string `json:"captcha_id"`
		CaptchaCode string `json:"captcha_code"`
		TOTPCode    string `json:"totp_code"`
		EmailCode   string `json:"email_code"`
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

	user, err := UserSvc.Login(req.Username, req.Password, c.ClientIP())
	if err != nil {
		// 记录登录失败
		if SecuritySvc != nil {
			SecuritySvc.RecordLoginAttempt(req.Username, clientIP, false)
			failCount := SecuritySvc.GetLoginFailureCount(clientIP)
			if failCount >= 3 {
				c.JSON(400, gin.H{
					"success":      false,
					"error":        err.Error(),
					"fail_count":   failCount,
					"max_attempts": 5,
					"warning":      "登录失败次数过多将被临时锁定",
				})
				return
			}
		}
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 登录成功，清除失败记录
	if SecuritySvc != nil {
		SecuritySvc.RecordLoginAttempt(req.Username, clientIP, true)
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogUserActionSimple(user.ID, user.Username, "login", "user", "", nil, clientIP, c.GetHeader("User-Agent"))
	}

	// 检查是否启用了两步验证
	if user.Enable2FA {
		// 生成登录验证令牌，跳转到独立验证页面
		tokenBytes := make([]byte, 32)
		rand.Read(tokenBytes)
		verifyToken := hex.EncodeToString(tokenBytes)

		loginTokensMu.Lock()
		loginTokens[verifyToken] = &LoginToken{
			UserID:          user.ID,
			Username:        user.Username,
			Email:           user.Email,
			PreferEmailAuth: user.PreferEmailAuth,
			ExpiresAt:       time.Now().Add(10 * time.Minute),
		}
		loginTokensMu.Unlock()

		c.JSON(200, gin.H{
			"success":      false,
			"require_2fa":  true,
			"verify_token": verifyToken,
			"message":      "请完成二次验证",
		})
		return
	}

	// 创建会话（数据库持久化）
	if SessionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "会话服务未初始化"})
		return
	}
	sessionID, err := SessionSvc.CreateUserSession(user.ID, user.Username, c.ClientIP(), c.GetHeader("User-Agent"), req.Remember)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建会话失败"})
		return
	}

	// 记录登录设备
	if DeviceSvc != nil {
		DeviceSvc.RecordLoginDevice(user.ID, sessionID, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	// 设置Cookie
	maxAge := 7200 // 2小时
	if req.Remember {
		maxAge = 604800 // 7天
	}
	SetSecureCookie(c, "user_session", sessionID, maxAge, true)

	// 设置CSRF令牌
	csrfToken := SetCSRFCookie(c, sessionID)

	c.JSON(200, gin.H{
		"success":    true,
		"message":    "登录成功",
		"csrf_token": csrfToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// UserLogout 用户登出
func UserLogout(c *gin.Context) {
	sessionID, _ := c.Cookie("user_session")
	if sessionID != "" && SessionSvc != nil {
		SessionSvc.DeleteUserSession(sessionID)
	}
	c.SetCookie("user_session", "", -1, "/", "", false, true)
	c.JSON(200, gin.H{"success": true, "message": "已退出登录"})
}
