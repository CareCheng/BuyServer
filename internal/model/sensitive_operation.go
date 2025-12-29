package model

import (
	"time"
)

// SensitiveOperationToken 敏感操作验证令牌
// 用于修改密码、绑定邮箱、禁用2FA等敏感操作的二次验证
type SensitiveOperationToken struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index" json:"user_id"`
	Token         string    `gorm:"type:varchar(64);uniqueIndex" json:"token"`
	OperationType string    `gorm:"type:varchar(50)" json:"operation_type"` // change_password, bind_email, disable_2fa, delete_account
	Verified      bool      `gorm:"default:false" json:"verified"`          // 是否已验证
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// TableName 设置表名
func (SensitiveOperationToken) TableName() string {
	return "sensitive_operation_tokens"
}

// 敏感操作类型常量
const (
	OpTypeChangePassword = "change_password" // 修改密码
	OpTypeBindEmail      = "bind_email"      // 绑定/更换邮箱
	OpTypeDisable2FA     = "disable_2fa"     // 禁用两步验证
	OpTypeDeleteAccount  = "delete_account"  // 注销账户
	OpTypeChangePhone    = "change_phone"    // 更换手机号
)
