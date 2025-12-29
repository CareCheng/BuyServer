// Package api 提供 HTTP API 处理器
// user_password_handler.go - 用户密码找回处理器
package api

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"user-frontend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// 重置密码令牌存储
var (
	resetTokens   = make(map[string]*ResetToken)
	resetTokensMu sync.RWMutex
)

// ResetToken 重置密码令牌
type ResetToken struct {
	Username  string
	ExpiresAt time.Time
}

// CleanupExpiredTokens 清理过期的令牌（由定时任务调用）
func CleanupExpiredTokens() {
	now := time.Now()

	// 清理过期的重置密码令牌
	resetTokensMu.Lock()
	for token, data := range resetTokens {
		if now.After(data.ExpiresAt) {
			delete(resetTokens, token)
		}
	}
	resetTokensMu.Unlock()

	// 清理过期的登录验证令牌
	loginTokensMu.Lock()
	for token, data := range loginTokens {
		if now.After(data.ExpiresAt) {
			delete(loginTokens, token)
		}
	}
	loginTokensMu.Unlock()
}

// ForgotPasswordPage 忘记密码页面
func ForgotPasswordPage(c *gin.Context) {
	c.HTML(200, "forgot_password.html", gin.H{
		"title": "找回密码",
	})
}

// ForgotPasswordCheck 检查用户是否存在并返回邮箱信息
func ForgotPasswordCheck(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请输入用户名"})
		return
	}

	user, err := UserSvc.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "用户不存在"})
		return
	}

	if user.Email == "" {
		c.JSON(400, gin.H{"success": false, "error": "该用户未绑定邮箱，无法找回密码"})
		return
	}

	c.JSON(200, gin.H{
		"success":      true,
		"email":        user.Email,
		"masked_email": maskEmail(user.Email),
		"has_2fa":      user.Enable2FA,
	})
}

// ForgotPasswordVerify 验证身份（邮箱验证码或TOTP）
func ForgotPasswordVerify(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Username  string `json:"username" binding:"required"`
		EmailCode string `json:"email_code"`
		TOTPCode  string `json:"totp_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	user, err := UserSvc.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "用户不存在"})
		return
	}

	// 验证身份
	verified := false
	if user.Enable2FA && req.TOTPCode != "" {
		// TOTP验证
		if totp.Validate(req.TOTPCode, user.TOTPSecret) {
			verified = true
		} else {
			c.JSON(400, gin.H{"success": false, "error": "动态口令错误"})
			return
		}
	} else if req.EmailCode != "" {
		// 邮箱验证码验证
		if EmailSvc != nil && EmailSvc.VerifyCode(user.Email, req.EmailCode, "reset_password") {
			verified = true
		} else {
			c.JSON(400, gin.H{"success": false, "error": "邮箱验证码错误或已过期"})
			return
		}
	} else {
		c.JSON(400, gin.H{"success": false, "error": "请提供验证码"})
		return
	}

	if verified {
		// 生成重置令牌
		tokenBytes := make([]byte, 32)
		rand.Read(tokenBytes)
		token := hex.EncodeToString(tokenBytes)

		resetTokensMu.Lock()
		resetTokens[token] = &ResetToken{
			Username:  req.Username,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}
		resetTokensMu.Unlock()

		c.JSON(200, gin.H{
			"success":     true,
			"reset_token": token,
		})
	}
}

// ForgotPasswordReset 重置密码
func ForgotPasswordReset(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Username    string `json:"username" binding:"required"`
		ResetToken  string `json:"reset_token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(400, gin.H{"success": false, "error": "密码长度至少6位"})
		return
	}

	// 验证令牌
	resetTokensMu.RLock()
	tokenData, exists := resetTokens[req.ResetToken]
	resetTokensMu.RUnlock()

	if !exists {
		c.JSON(400, gin.H{"success": false, "error": "无效的重置令牌"})
		return
	}

	if time.Now().After(tokenData.ExpiresAt) {
		resetTokensMu.Lock()
		delete(resetTokens, req.ResetToken)
		resetTokensMu.Unlock()
		c.JSON(400, gin.H{"success": false, "error": "重置令牌已过期，请重新验证"})
		return
	}

	if tokenData.Username != req.Username {
		c.JSON(400, gin.H{"success": false, "error": "无效的重置令牌"})
		return
	}

	// 重置密码
	user, err := UserSvc.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "用户不存在"})
		return
	}

	passwordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "密码加密失败"})
		return
	}

	user.PasswordHash = passwordHash
	// 重置密码后关闭两步验证
	user.Enable2FA = false
	user.TOTPSecret = ""
	user.PreferEmailAuth = false

	if err := UserSvc.UpdateUser(user); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "重置密码失败"})
		return
	}

	// 删除令牌
	resetTokensMu.Lock()
	delete(resetTokens, req.ResetToken)
	resetTokensMu.Unlock()

	c.JSON(200, gin.H{"success": true, "message": "密码重置成功，两步验证已关闭，请重新设置"})
}
