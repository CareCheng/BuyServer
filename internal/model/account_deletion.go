package model

import (
	"time"

	"gorm.io/gorm"
)

// AccountDeletionRequest 账户注销申请模型
type AccountDeletionRequest struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"uniqueIndex" json:"user_id"`                // 用户ID（一个用户只能有一个待处理的申请）
	Username     string         `gorm:"type:varchar(100)" json:"username"`         // 用户名
	Email        string         `gorm:"type:varchar(255)" json:"email"`            // 邮箱
	Reason       string         `gorm:"type:text" json:"reason"`                   // 注销原因
	Status       int            `gorm:"default:0" json:"status"`                   // 状态：0待处理 1已批准 2已拒绝 3已取消
	RejectReason string         `gorm:"type:text" json:"reject_reason"`            // 拒绝原因
	ProcessedBy  string         `gorm:"type:varchar(100)" json:"processed_by"`     // 处理人
	ProcessedAt  *time.Time     `json:"processed_at"`                              // 处理时间
	ScheduledAt  *time.Time     `json:"scheduled_at"`                              // 计划删除时间（批准后7天）
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// AccountDeletionStatus 账户注销状态常量
const (
	DeletionStatusPending   = 0 // 待处理
	DeletionStatusApproved  = 1 // 已批准
	DeletionStatusRejected  = 2 // 已拒绝
	DeletionStatusCancelled = 3 // 已取消
	DeletionStatusCompleted = 4 // 已完成
)

// TableName 设置表名
func (AccountDeletionRequest) TableName() string {
	return "account_deletion_requests"
}
