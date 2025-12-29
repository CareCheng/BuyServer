// Package model 数据模型
// recharge_promo.go - 充值优惠活动模型
package model

import (
	"time"

	"gorm.io/gorm"
)

// RechargePromo 充值优惠活动
type RechargePromo struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100" json:"name"`                      // 活动名称
	Description string         `gorm:"size:500" json:"description"`               // 活动描述
	PromoType   string         `gorm:"size:20;index" json:"promo_type"`           // 优惠类型：discount折扣/bonus赠金/percent百分比赠送
	MinAmount   float64        `gorm:"type:decimal(10,2)" json:"min_amount"`      // 最低充值金额（门槛）
	MaxAmount   float64        `gorm:"type:decimal(10,2)" json:"max_amount"`      // 最高充值金额（0表示不限）
	Value       float64        `gorm:"type:decimal(10,2)" json:"value"`           // 优惠值（折扣率/赠金金额/赠送百分比）
	MaxBonus    float64        `gorm:"type:decimal(10,2)" json:"max_bonus"`       // 最大赠送金额（0表示不限，用于百分比赠送）
	Priority    int            `gorm:"default:0" json:"priority"`                 // 优先级（数字越大优先级越高）
	StackMode   string         `gorm:"size:20;default:'best'" json:"stack_mode"`  // 叠加模式：best最优/first首个/all全部
	PerUserLimit int           `gorm:"default:0" json:"per_user_limit"`           // 每用户限用次数（0表示不限）
	TotalLimit  int            `gorm:"default:0" json:"total_limit"`              // 总使用次数限制（0表示不限）
	UsedCount   int            `gorm:"default:0" json:"used_count"`               // 已使用次数
	StartAt     *time.Time     `json:"start_at"`                                  // 开始时间
	EndAt       *time.Time     `json:"end_at"`                                    // 结束时间
	Status      int            `gorm:"default:1;index" json:"status"`             // 状态：1启用 0禁用
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// RechargePromoUsage 充值优惠使用记录
type RechargePromoUsage struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PromoID     uint      `gorm:"index" json:"promo_id"`                     // 优惠活动ID
	UserID      uint      `gorm:"index" json:"user_id"`                      // 用户ID
	RechargeNo  string    `gorm:"size:64;index" json:"recharge_no"`          // 充值单号
	Amount      float64   `gorm:"type:decimal(10,2)" json:"amount"`          // 充值金额
	BonusAmount float64   `gorm:"type:decimal(10,2)" json:"bonus_amount"`    // 赠送金额
	DiscountAmount float64 `gorm:"type:decimal(10,2)" json:"discount_amount"` // 折扣金额（实际少付的金额）
	CreatedAt   time.Time `json:"created_at"`
}

// 优惠类型常量
const (
	PromoTypeDiscount = "discount" // 折扣（如9折，value=0.9）
	PromoTypeBonus    = "bonus"    // 固定赠金（如充100送10，value=10）
	PromoTypePercent  = "percent"  // 百分比赠送（如充值送10%，value=10）
)

// 叠加模式常量
const (
	StackModeBest  = "best"  // 选择最优惠的一个
	StackModeFirst = "first" // 选择第一个匹配的
	StackModeAll   = "all"   // 所有优惠叠加
)

// IsActive 检查活动是否有效
func (p *RechargePromo) IsActive() bool {
	if p.Status != 1 {
		return false
	}
	now := time.Now()
	if p.StartAt != nil && now.Before(*p.StartAt) {
		return false
	}
	if p.EndAt != nil && now.After(*p.EndAt) {
		return false
	}
	if p.TotalLimit > 0 && p.UsedCount >= p.TotalLimit {
		return false
	}
	return true
}

// CalculateBonus 计算赠送金额
func (p *RechargePromo) CalculateBonus(rechargeAmount float64) float64 {
	if rechargeAmount < p.MinAmount {
		return 0
	}
	if p.MaxAmount > 0 && rechargeAmount > p.MaxAmount {
		return 0
	}

	var bonus float64
	switch p.PromoType {
	case PromoTypeBonus:
		// 固定赠金
		bonus = p.Value
	case PromoTypePercent:
		// 百分比赠送
		bonus = rechargeAmount * p.Value / 100
		if p.MaxBonus > 0 && bonus > p.MaxBonus {
			bonus = p.MaxBonus
		}
	case PromoTypeDiscount:
		// 折扣不产生赠金，而是减少实付金额
		bonus = 0
	}
	return bonus
}

// CalculateDiscount 计算折扣金额（实际少付的金额）
func (p *RechargePromo) CalculateDiscount(rechargeAmount float64) float64 {
	if rechargeAmount < p.MinAmount {
		return 0
	}
	if p.MaxAmount > 0 && rechargeAmount > p.MaxAmount {
		return 0
	}

	if p.PromoType == PromoTypeDiscount {
		// 折扣：value 表示折扣率，如 0.9 表示9折
		return rechargeAmount * (1 - p.Value)
	}
	return 0
}

// GetPromoTypeText 获取优惠类型文本
func (p *RechargePromo) GetPromoTypeText() string {
	switch p.PromoType {
	case PromoTypeDiscount:
		return "充值折扣"
	case PromoTypeBonus:
		return "固定赠金"
	case PromoTypePercent:
		return "百分比赠送"
	default:
		return "未知"
	}
}

// TableName 设置表名
func (RechargePromo) TableName() string {
	return "recharge_promos"
}

// TableName 设置表名
func (RechargePromoUsage) TableName() string {
	return "recharge_promo_usages"
}
