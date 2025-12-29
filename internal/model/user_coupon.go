package model

import (
	"time"

	"gorm.io/gorm"
)

// UserCoupon 用户优惠券（记录用户持有的优惠券）
type UserCoupon struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"index;not null" json:"user_id"`                    // 用户ID
	CouponID   uint           `gorm:"index;not null" json:"coupon_id"`                  // 优惠券ID
	CouponCode string         `gorm:"type:varchar(50)" json:"coupon_code"`              // 优惠券码
	CouponName string         `gorm:"type:varchar(100)" json:"coupon_name"`             // 优惠券名称
	Source     string         `gorm:"type:varchar(50);default:'exchange'" json:"source"` // 来源：exchange(积分兑换)、gift(赠送)、activity(活动)
	SourceID   uint           `gorm:"default:0" json:"source_id"`                       // 来源ID（如兑换记录ID）
	Status     int            `gorm:"default:0" json:"status"`                          // 状态：0未使用、1已使用、2已过期、3已作废
	UsedAt     *time.Time     `json:"used_at"`                                          // 使用时间
	UsedOrder  string         `gorm:"type:varchar(64)" json:"used_order"`               // 使用的订单号
	ExpireAt   *time.Time     `json:"expire_at"`                                        // 过期时间
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	
	// 关联字段
	Coupon *Coupon `gorm:"foreignKey:CouponID" json:"coupon,omitempty"`
}

// TableName 设置表名
func (UserCoupon) TableName() string {
	return "user_coupons"
}

// 用户优惠券状态常量
const (
	UserCouponStatusUnused   = 0 // 未使用
	UserCouponStatusUsed     = 1 // 已使用
	UserCouponStatusExpired  = 2 // 已过期
	UserCouponStatusInvalid  = 3 // 已作废
)

// 用户优惠券来源常量
const (
	UserCouponSourceExchange = "exchange" // 积分兑换
	UserCouponSourceGift     = "gift"     // 赠送
	UserCouponSourceActivity = "activity" // 活动
)
