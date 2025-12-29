package model

import (
	"time"
)

// LoginDevice 登录设备记录
// 用于记录用户的登录设备信息，支持设备管理和踢出功能
type LoginDevice struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"index" json:"user_id"`                          // 用户ID
	SessionID  string     `gorm:"type:varchar(64);index" json:"session_id"`      // 关联的会话ID
	DeviceName string     `gorm:"type:varchar(100)" json:"device_name"`          // 设备名称（从User-Agent解析）
	DeviceType string     `gorm:"type:varchar(50)" json:"device_type"`           // 设备类型：PC/Mobile/Tablet
	Browser    string     `gorm:"type:varchar(50)" json:"browser"`               // 浏览器
	OS         string     `gorm:"type:varchar(50)" json:"os"`                    // 操作系统
	IP         string     `gorm:"type:varchar(50)" json:"ip"`                    // 登录IP
	Location   string     `gorm:"type:varchar(100)" json:"location"`             // IP归属地
	IsCurrent  bool       `gorm:"default:false" json:"is_current"`               // 是否当前设备（运行时计算）
	LastActive time.Time  `json:"last_active"`                                   // 最后活跃时间
	CreatedAt  time.Time  `json:"created_at"`                                    // 首次登录时间
}

// LoginHistory 登录历史记录
// 记录用户的登录历史，用于安全审计
type LoginHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`                     // 用户ID
	Username  string    `gorm:"type:varchar(100)" json:"username"`        // 用户名
	IP        string    `gorm:"type:varchar(50)" json:"ip"`               // 登录IP
	Location  string    `gorm:"type:varchar(100)" json:"location"`        // IP归属地
	Device    string    `gorm:"type:varchar(100)" json:"device"`          // 设备信息
	Browser   string    `gorm:"type:varchar(50)" json:"browser"`          // 浏览器
	OS        string    `gorm:"type:varchar(50)" json:"os"`               // 操作系统
	Status    int       `gorm:"default:1" json:"status"`                  // 状态：1成功 0失败
	FailReason string   `gorm:"type:varchar(200)" json:"fail_reason"`     // 失败原因
	CreatedAt time.Time `json:"created_at"`                               // 登录时间
}

// TableName 设置表名
func (LoginDevice) TableName() string {
	return "login_devices"
}

func (LoginHistory) TableName() string {
	return "login_histories"
}
