package api

import (
	"github.com/gin-gonic/gin"
)

// RequestSensitiveVerification 请求敏感操作验证
// 用户在执行敏感操作前调用此接口获取验证令牌
func RequestSensitiveVerification(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		OperationType string `json:"operation_type" binding:"required"` // change_password, bind_email, disable_2fa, delete_account
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 验证操作类型
	validTypes := map[string]bool{
		"change_password": true,
		"bind_email":      true,
		"disable_2fa":     true,
		"delete_account":  true,
		"change_phone":    true,
	}
	if !validTypes[req.OperationType] {
		c.JSON(400, gin.H{"success": false, "error": "无效的操作类型"})
		return
	}

	if SensitiveSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 获取用户可用的验证方式
	methods, err := SensitiveSvc.GetUserVerificationMethods(userID)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 生成验证令牌
	token, err := SensitiveSvc.RequestVerification(userID, req.OperationType)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"token":   token,
		"methods": methods,
		"message": "请选择验证方式完成身份验证",
	})
}

// VerifySensitiveOperation 验证敏感操作
func VerifySensitiveOperation(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Token      string `json:"token" binding:"required"`
		Method     string `json:"method" binding:"required"` // password, email, totp
		Password   string `json:"password"`
		EmailCode  string `json:"email_code"`
		TOTPCode   string `json:"totp_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if SensitiveSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var err error
	switch req.Method {
	case "password":
		if req.Password == "" {
			c.JSON(400, gin.H{"success": false, "error": "请输入密码"})
			return
		}
		err = SensitiveSvc.VerifyWithPassword(userID, req.Token, req.Password)
	case "email":
		if req.EmailCode == "" {
			c.JSON(400, gin.H{"success": false, "error": "请输入邮箱验证码"})
			return
		}
		err = SensitiveSvc.VerifyWithEmail(userID, req.Token, req.EmailCode)
	case "totp":
		if req.TOTPCode == "" {
			c.JSON(400, gin.H{"success": false, "error": "请输入动态口令"})
			return
		}
		err = SensitiveSvc.VerifyWithTOTP(userID, req.Token, req.TOTPCode)
	default:
		c.JSON(400, gin.H{"success": false, "error": "无效的验证方式"})
		return
	}

	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "验证成功，请在10分钟内完成操作",
	})
}

// SendSensitiveVerificationEmail 发送敏感操作验证邮件
func SendSensitiveVerificationEmail(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		OperationType string `json:"operation_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if SensitiveSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := SensitiveSvc.SendVerificationEmail(userID, req.OperationType); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "验证码已发送",
	})
}

// UpdatePasswordWithVerification 带验证的密码修改
func UpdatePasswordWithVerification(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	var req struct {
		Token       string `json:"token" binding:"required"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(400, gin.H{"success": false, "error": "新密码长度至少6位"})
		return
	}

	if SensitiveSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 检查验证状态
	verified, err := SensitiveSvc.CheckVerified(userID, req.Token, "change_password")
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !verified {
		c.JSON(400, gin.H{"success": false, "error": "请先完成身份验证"})
		return
	}

	// 执行密码修改
	if err := UserSvc.UpdatePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 消费令牌
	SensitiveSvc.ConsumeToken(userID, req.Token)

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogUserActionSimple(userID, username, "change_password", "user", "", nil, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "密码修改成功"})
}

// BindEmailWithVerification 带验证的邮箱绑定
func BindEmailWithVerification(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	var req struct {
		Token         string `json:"token" binding:"required"`
		Email         string `json:"email" binding:"required"`
		EmailCode     string `json:"email_code" binding:"required"` // 新邮箱的验证码
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if SensitiveSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 检查验证状态
	verified, err := SensitiveSvc.CheckVerified(userID, req.Token, "bind_email")
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !verified {
		c.JSON(400, gin.H{"success": false, "error": "请先完成身份验证"})
		return
	}

	// 验证新邮箱的验证码
	if EmailSvc == nil || !EmailSvc.VerifyCode(req.Email, req.EmailCode, "register") {
		c.JSON(400, gin.H{"success": false, "error": "邮箱验证码错误或已过期"})
		return
	}

	// 执行邮箱绑定
	if err := UserSvc.BindEmail(userID, req.Email); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 消费令牌
	SensitiveSvc.ConsumeToken(userID, req.Token)

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogUserActionSimple(userID, username, "bind_email", "user", "", map[string]interface{}{"email": req.Email}, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "邮箱绑定成功"})
}

// Disable2FAWithVerification 带验证的禁用两步验证
func Disable2FAWithVerification(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if SensitiveSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 检查验证状态
	verified, err := SensitiveSvc.CheckVerified(userID, req.Token, "disable_2fa")
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !verified {
		c.JSON(400, gin.H{"success": false, "error": "请先完成身份验证"})
		return
	}

	// 执行禁用两步验证
	if err := UserSvc.Disable2FA(userID); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 消费令牌
	SensitiveSvc.ConsumeToken(userID, req.Token)

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogUserActionSimple(userID, username, "disable_2fa", "user", "", nil, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "两步验证已禁用"})
}
