package model

import (
	"time"
)

// BalanceAlert 余额异常告警记录
type BalanceAlert struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"`                       // 用户ID
	AlertType   string    `gorm:"size:50;index" json:"alert_type"`            // 告警类型
	Level       string    `gorm:"size:20;index" json:"level"`                 // 告警级别：info/warning/critical
	Title       string    `gorm:"size:200" json:"title"`                      // 告警标题
	Content     string    `gorm:"size:2000" json:"content"`                   // 告警内容
	Amount      float64   `gorm:"type:decimal(10,2)" json:"amount"`           // 涉及金额
	RelatedID   string    `gorm:"size:100" json:"related_id"`                 // 关联ID（订单号/充值单号等）
	Status      int       `gorm:"default:0;index" json:"status"`              // 状态：0未处理 1已处理 2已忽略
	HandledBy   uint      `gorm:"default:0" json:"handled_by"`                // 处理人ID
	HandledAt   *time.Time `json:"handled_at"`                                // 处理时间
	HandleRemark string   `gorm:"size:500" json:"handle_remark"`              // 处理备注
	ClientIP    string    `gorm:"size:50" json:"client_ip"`                   // 客户端IP
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 告警类型常量
const (
	AlertTypeLargeRecharge     = "large_recharge"      // 大额充值
	AlertTypeLargeConsume      = "large_consume"       // 大额消费
	AlertTypeFrequentRecharge  = "frequent_recharge"   // 频繁充值
	AlertTypeFrequentConsume   = "frequent_consume"    // 频繁消费
	AlertTypeNegativeBalance   = "negative_balance"    // 余额异常（负数）
	AlertTypeBalanceMismatch   = "balance_mismatch"    // 余额不一致
	AlertTypeSuspiciousIP      = "suspicious_ip"       // 可疑IP操作
	AlertTypeUnfreezeFailure   = "unfreeze_failure"    // 解冻失败
	AlertTypeRefundAnomaly     = "refund_anomaly"      // 退款异常
	AlertTypeAdminAdjust       = "admin_adjust"        // 管理员调整（大额）
)

// 告警级别常量
const (
	AlertLevelInfo     = "info"     // 信息
	AlertLevelWarning  = "warning"  // 警告
	AlertLevelCritical = "critical" // 严重
)

// 告警状态常量
const (
	AlertStatusPending  = 0 // 未处理
	AlertStatusHandled  = 1 // 已处理
	AlertStatusIgnored  = 2 // 已忽略
)

// GetAlertTypeText 获取告警类型文本
func (a *BalanceAlert) GetAlertTypeText() string {
	switch a.AlertType {
	case AlertTypeLargeRecharge:
		return "大额充值"
	case AlertTypeLargeConsume:
		return "大额消费"
	case AlertTypeFrequentRecharge:
		return "频繁充值"
	case AlertTypeFrequentConsume:
		return "频繁消费"
	case AlertTypeNegativeBalance:
		return "余额异常"
	case AlertTypeBalanceMismatch:
		return "余额不一致"
	case AlertTypeSuspiciousIP:
		return "可疑IP操作"
	case AlertTypeUnfreezeFailure:
		return "解冻失败"
	case AlertTypeRefundAnomaly:
		return "退款异常"
	case AlertTypeAdminAdjust:
		return "管理员调整"
	default:
		return "未知"
	}
}

// GetLevelText 获取告警级别文本
func (a *BalanceAlert) GetLevelText() string {
	switch a.Level {
	case AlertLevelInfo:
		return "信息"
	case AlertLevelWarning:
		return "警告"
	case AlertLevelCritical:
		return "严重"
	default:
		return "未知"
	}
}

// GetStatusText 获取告警状态文本
func (a *BalanceAlert) GetStatusText() string {
	switch a.Status {
	case AlertStatusPending:
		return "未处理"
	case AlertStatusHandled:
		return "已处理"
	case AlertStatusIgnored:
		return "已忽略"
	default:
		return "未知"
	}
}

// TableName 设置表名
func (BalanceAlert) TableName() string {
	return "balance_alerts"
}
