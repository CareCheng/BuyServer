// Package api 提供 HTTP API 处理器
// user_email_handler.go - 用户邮箱相关处理器
package api

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// maskEmail 隐藏邮箱中间部分
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	name := parts[0]
	if len(name) <= 2 {
		return name + "***@" + parts[1]
	}
	return name[:2] + "***@" + parts[1]
}

// GetEmailCodeLength 获取验证码长度（公开接口）
func GetEmailCodeLength(c *gin.Context) {
	codeLength := 6 // 默认值
	
	if ConfigSvc != nil {
		emailCfg, err := ConfigSvc.GetEmailConfig()
		if err == nil && emailCfg != nil && emailCfg.CodeLength > 0 {
			codeLength = emailCfg.CodeLength
		}
	}
	
	c.JSON(200, gin.H{
		"success":     true,
		"code_length": codeLength,
	})
}

// VerifyEmailCodeOnly 仅验证邮箱验证码（不消耗，用于实时验证）
func VerifyEmailCodeOnly(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Code     string `json:"code" binding:"required"`
		CodeType string `json:"code_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if EmailSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "邮箱服务未初始化"})
		return
	}

	// 验证但不消耗验证码
	valid := EmailSvc.CheckCodeValid(req.Email, req.Code, req.CodeType)
	c.JSON(200, gin.H{
		"success": true,
		"valid":   valid,
	})
}

// SendEmailCode 发送邮箱验证码
func SendEmailCode(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		CodeType string `json:"code_type" binding:"required"` // register, login, reset_password, enable_2fa
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if EmailSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "邮箱服务未初始化"})
		return
	}

	if err := EmailSvc.SendVerifyCode(req.Email, req.CodeType); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "验证码已发送"})
}

// VerifyEmailCode 验证邮箱验证码
func VerifyEmailCode(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Code     string `json:"code" binding:"required"`
		CodeType string `json:"code_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if EmailSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "邮箱服务未初始化"})
		return
	}

	if EmailSvc.VerifyCode(req.Email, req.Code, req.CodeType) {
		c.JSON(200, gin.H{"success": true, "message": "验证成功"})
	} else {
		c.JSON(400, gin.H{"success": false, "error": "验证码错误或已过期"})
	}
}

// BindEmail 绑定邮箱
func BindEmail(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		Email string `json:"email" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 验证验证码
	if EmailSvc == nil || !EmailSvc.VerifyCode(req.Email, req.Code, "register") {
		c.JSON(400, gin.H{"success": false, "error": "验证码错误或已过期"})
		return
	}

	// 绑定邮箱
	if err := UserSvc.BindEmail(userID, req.Email); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "邮箱绑定成功"})
}
