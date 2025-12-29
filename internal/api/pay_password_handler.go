package api

import (
	"github.com/gin-gonic/gin"
)

// ==================== 支付密码 API ====================

// GetPayPasswordStatus 获取支付密码状态
func GetPayPasswordStatus(c *gin.Context) {
	if PayPasswordSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")

	isSet, isLocked, lockRemaining, err := PayPasswordSvc.GetPayPasswordStatus(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"is_set":                 isSet,
			"is_locked":              isLocked,
			"lock_remaining_seconds": lockRemaining,
		},
	})
}

// SetPayPassword 设置支付密码
func SetPayPassword(c *gin.Context) {
	if PayPasswordSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")

	var req struct {
		Password      string `json:"password" binding:"required"`
		LoginPassword string `json:"login_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := PayPasswordSvc.SetPayPassword(userID.(uint), req.Password, req.LoginPassword); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "支付密码设置成功"})
}

// UpdatePayPassword 修改支付密码
func UpdatePayPassword(c *gin.Context) {
	if PayPasswordSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := PayPasswordSvc.UpdatePayPassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "支付密码修改成功"})
}

// ResetPayPassword 重置支付密码（需要先通过邮箱验证）
func ResetPayPassword(c *gin.Context) {
	if PayPasswordSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")

	var req struct {
		NewPassword string `json:"new_password" binding:"required"`
		EmailCode   string `json:"email_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取用户邮箱
	user, err := UserSvc.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取用户信息失败"})
		return
	}

	// 验证邮箱验证码
	if EmailSvc != nil {
		if !EmailSvc.VerifyCode(user.Email, req.EmailCode, "reset_pay_password") {
			c.JSON(400, gin.H{"success": false, "error": "验证码错误或已过期"})
			return
		}
	}

	if err := PayPasswordSvc.ResetPayPassword(userID.(uint), req.NewPassword); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "支付密码重置成功"})
}

// VerifyPayPassword 验证支付密码（用于前端预验证）
func VerifyPayPassword(c *gin.Context) {
	if PayPasswordSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")

	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := PayPasswordSvc.VerifyPayPassword(userID.(uint), req.Password); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "验证成功"})
}

// SendResetPayPasswordCode 发送重置支付密码验证码
func SendResetPayPasswordCode(c *gin.Context) {
	if EmailSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "邮件服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")

	// 获取用户邮箱
	user, err := UserSvc.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取用户信息失败"})
		return
	}

	if user.Email == "" {
		c.JSON(400, gin.H{"success": false, "error": "请先绑定邮箱"})
		return
	}

	// 发送验证码
	if err := EmailSvc.SendVerifyCode(user.Email, "reset_pay_password"); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "发送验证码失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "验证码已发送"})
}
