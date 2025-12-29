package model

import (
	"time"
)

// RenewalReminder 续费提醒记录
// 用于记录已发送的续费提醒，避免重复发送
type RenewalReminder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	OrderID   uint      `gorm:"index" json:"order_id"`
	OrderNo   string    `gorm:"type:varchar(64)" json:"order_no"`
	KamiCode  string    `gorm:"type:varchar(255)" json:"kami_code"`
	ExpireAt  time.Time `json:"expire_at"`           // 卡密过期时间
	RemindAt  time.Time `json:"remind_at"`           // 提醒发送时间
	RemindType string   `gorm:"type:varchar(20)" json:"remind_type"` // 提醒类型：7day, 3day, 1day, expired
	CreatedAt time.Time `json:"created_at"`
}

// TableName 设置表名
func (RenewalReminder) TableName() string {
	return "renewal_reminders"
}

// 续费提醒类型常量
const (
	RemindType7Day   = "7day"    // 7天前提醒
	RemindType3Day   = "3day"    // 3天前提醒
	RemindType1Day   = "1day"    // 1天前提醒
	RemindTypeExpired = "expired" // 已过期提醒
)
