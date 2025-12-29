package model

import (
	"time"
)

// LoginAlert 异地登录提醒记录
type LoginAlert struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index" json:"user_id"`                      // 用户ID
	Username     string    `gorm:"type:varchar(100)" json:"username"`         // 用户名
	IP           string    `gorm:"type:varchar(50)" json:"ip"`                // 登录IP
	Location     string    `gorm:"type:varchar(200)" json:"location"`         // IP归属地
	PreviousIP   string    `gorm:"type:varchar(50)" json:"previous_ip"`       // 上次登录IP
	PrevLocation string    `gorm:"type:varchar(200)" json:"prev_location"`    // 上次登录归属地
	DeviceInfo   string    `gorm:"type:varchar(500)" json:"device_info"`      // 设备信息
	AlertType    string    `gorm:"type:varchar(50)" json:"alert_type"`        // 提醒类型：new_location, new_device, suspicious
	EmailSent    bool      `gorm:"default:false" json:"email_sent"`           // 是否已发送邮件
	EmailSentAt  *time.Time `json:"email_sent_at"`                            // 邮件发送时间
	Acknowledged bool      `gorm:"default:false" json:"acknowledged"`         // 用户是否已确认
	CreatedAt    time.Time `json:"created_at"`
}

// UserLoginLocation 用户常用登录地点
type UserLoginLocation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`                      // 用户ID
	IP        string    `gorm:"type:varchar(50)" json:"ip"`                // IP地址
	Location  string    `gorm:"type:varchar(200)" json:"location"`         // 归属地
	LoginCount int      `gorm:"default:1" json:"login_count"`              // 登录次数
	LastLoginAt time.Time `json:"last_login_at"`                           // 最后登录时间
	IsTrusted  bool      `gorm:"default:false" json:"is_trusted"`          // 是否为可信地点
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// LoginAlertType 登录提醒类型常量
const (
	AlertTypeNewLocation = "new_location" // 新地点登录
	AlertTypeNewDevice   = "new_device"   // 新设备登录
	AlertTypeSuspicious  = "suspicious"   // 可疑登录
)

// TableName 设置表名
func (LoginAlert) TableName() string {
	return "login_alerts"
}

func (UserLoginLocation) TableName() string {
	return "user_login_locations"
}
