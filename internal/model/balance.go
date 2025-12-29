package model

import (
	"time"
)

// UserBalance 用户余额
type UserBalance struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex" json:"user_id"`           // 用户ID
	Balance   float64   `gorm:"default:0;type:decimal(10,2)" json:"balance"`    // 可用余额
	Frozen    float64   `gorm:"default:0;type:decimal(10,2)" json:"frozen"`     // 冻结金额
	TotalIn   float64   `gorm:"default:0;type:decimal(10,2)" json:"total_in"`   // 累计充值
	TotalOut  float64   `gorm:"default:0;type:decimal(10,2)" json:"total_out"`  // 累计消费
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BalanceLog 余额变动记录
type BalanceLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index" json:"user_id"`                          // 用户ID
	Type          string    `gorm:"size:20;index" json:"type"`                     // 类型：recharge/consume/refund/withdraw/freeze/unfreeze
	Amount        float64   `gorm:"type:decimal(10,2)" json:"amount"`              // 变动金额（正数增加，负数减少）
	BeforeBalance float64   `gorm:"type:decimal(10,2)" json:"before_balance"`      // 变动前余额
	AfterBalance  float64   `gorm:"type:decimal(10,2)" json:"after_balance"`       // 变动后余额
	OrderNo       string    `gorm:"size:64;index" json:"order_no"`                 // 关联订单号
	RechargeNo    string    `gorm:"size:64;index" json:"recharge_no"`              // 充值单号
	Remark        string    `gorm:"size:500" json:"remark"`                        // 备注
	OperatorID    uint      `gorm:"default:0" json:"operator_id"`                  // 操作者ID（管理员调整时）
	OperatorType  string    `gorm:"size:20;default:'user'" json:"operator_type"`   // 操作者类型：user/admin/system
	ClientIP      string    `gorm:"size:50" json:"client_ip"`                      // 客户端IP
	CreatedAt     time.Time `json:"created_at"`
}

// RechargeOrder 充值订单
type RechargeOrder struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	RechargeNo    string     `gorm:"size:64;uniqueIndex" json:"recharge_no"`     // 充值单号
	UserID        uint       `gorm:"index" json:"user_id"`                       // 用户ID
	Amount        float64    `gorm:"type:decimal(10,2)" json:"amount"`           // 充值金额
	PayAmount     float64    `gorm:"type:decimal(10,2)" json:"pay_amount"`       // 实际支付金额（折扣后）
	BonusAmount   float64    `gorm:"type:decimal(10,2)" json:"bonus_amount"`     // 赠送金额
	TotalCredit   float64    `gorm:"type:decimal(10,2)" json:"total_credit"`     // 总到账金额
	PromoID       uint       `gorm:"default:0" json:"promo_id"`                  // 使用的优惠活动ID
	PromoName     string     `gorm:"size:100" json:"promo_name"`                 // 优惠活动名称
	PaymentMethod string     `gorm:"size:50" json:"payment_method"`              // 支付方式
	PaymentNo     string     `gorm:"size:100" json:"payment_no"`                 // 第三方支付单号
	Status        int        `gorm:"default:0;index" json:"status"`              // 状态：0待支付 1已支付 2已取消 3已退款
	PaidAt        *time.Time `json:"paid_at"`                                    // 支付时间
	ExpireAt      time.Time  `json:"expire_at"`                                  // 过期时间
	Remark        string     `gorm:"size:500" json:"remark"`                     // 备注
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// 余额变动类型常量
const (
	BalanceTypeRecharge = "recharge" // 充值
	BalanceTypeConsume  = "consume"  // 消费
	BalanceTypeRefund   = "refund"   // 退款
	BalanceTypeWithdraw = "withdraw" // 提现
	BalanceTypeFreeze   = "freeze"   // 冻结
	BalanceTypeUnfreeze = "unfreeze" // 解冻
	BalanceTypeGift     = "gift"     // 赠送
	BalanceTypeAdjust   = "adjust"   // 调整
)

// 充值订单状态常量
const (
	RechargeStatusPending   = 0 // 待支付
	RechargeStatusPaid      = 1 // 已支付
	RechargeStatusCancelled = 2 // 已取消
	RechargeStatusRefunded  = 3 // 已退款
)

// GetStatusText 获取充值订单状态文本
func (r *RechargeOrder) GetStatusText() string {
	switch r.Status {
	case RechargeStatusPending:
		return "待支付"
	case RechargeStatusPaid:
		return "已支付"
	case RechargeStatusCancelled:
		return "已取消"
	case RechargeStatusRefunded:
		return "已退款"
	default:
		return "未知"
	}
}

// GetTypeText 获取余额变动类型文本
func (b *BalanceLog) GetTypeText() string {
	switch b.Type {
	case BalanceTypeRecharge:
		return "充值"
	case BalanceTypeConsume:
		return "消费"
	case BalanceTypeRefund:
		return "退款"
	case BalanceTypeWithdraw:
		return "提现"
	case BalanceTypeFreeze:
		return "冻结"
	case BalanceTypeUnfreeze:
		return "解冻"
	case BalanceTypeGift:
		return "赠送"
	case BalanceTypeAdjust:
		return "调整"
	default:
		return "未知"
	}
}

// TableName 设置表名
func (UserBalance) TableName() string {
	return "user_balances"
}

// TableName 设置表名
func (BalanceLog) TableName() string {
	return "balance_logs"
}

// TableName 设置表名
func (RechargeOrder) TableName() string {
	return "recharge_orders"
}
