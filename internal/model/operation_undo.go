package model

import (
	"time"

	"gorm.io/gorm"
)

// UndoOperationType 可撤销操作类型
const (
	UndoTypeProductDelete   = "product_delete"   // 删除商品
	UndoTypeProductDisable  = "product_disable"  // 禁用商品
	UndoTypeUserDelete      = "user_delete"      // 删除用户
	UndoTypeUserDisable     = "user_disable"     // 禁用用户
	UndoTypeCouponDelete    = "coupon_delete"    // 删除优惠券
	UndoTypeCouponDisable   = "coupon_disable"   // 禁用优惠券
	UndoTypeCategoryDelete  = "category_delete"  // 删除分类
	UndoTypeAnnouncementDelete = "announcement_delete" // 删除公告
)

// UndoOperation 可撤销操作记录
type UndoOperation struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	OperationType string        `gorm:"size:50;index" json:"operation_type"`  // 操作类型
	TargetType   string         `gorm:"size:50" json:"target_type"`           // 目标类型：product/user/coupon等
	TargetID     uint           `gorm:"index" json:"target_id"`               // 目标ID
	TargetName   string         `gorm:"size:200" json:"target_name"`          // 目标名称（用于显示）
	OriginalData string         `gorm:"type:text" json:"original_data"`       // 原始数据（JSON）
	AdminID      uint           `gorm:"index" json:"admin_id"`                // 操作管理员ID
	AdminName    string         `gorm:"size:50" json:"admin_name"`            // 操作管理员名称
	Status       int            `gorm:"default:0" json:"status"`              // 状态：0可撤销 1已撤销 2已过期
	UndoneAt     *time.Time     `json:"undone_at"`                            // 撤销时间
	UndoneBy     string         `gorm:"size:50" json:"undone_by"`             // 撤销人
	ExpireAt     time.Time      `gorm:"index" json:"expire_at"`               // 过期时间
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// UndoConfig 撤销配置
type UndoConfig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Enabled         bool      `gorm:"default:true" json:"enabled"`           // 是否启用撤销功能
	RetentionHours  int       `gorm:"default:24" json:"retention_hours"`     // 保留时长（小时）
	AllowedTypes    string    `gorm:"type:text" json:"allowed_types"`        // 允许撤销的操作类型（JSON数组）
	UpdatedAt       time.Time `json:"updated_at"`
}
