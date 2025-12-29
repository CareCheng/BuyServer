package model

import (
	"time"

	"gorm.io/gorm"
)

// ManualKami 手动卡密模型
// 用于存储管理员手动导入的卡密，不通过服务端生成
type ManualKami struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProductID uint           `gorm:"index" json:"product_id"`       // 关联商品ID
	KamiCode  string         `gorm:"type:varchar(255)" json:"kami_code"` // 卡密内容
	Status    int            `gorm:"default:0" json:"status"`       // 状态：0可用 1已售出 2已禁用
	OrderID   uint           `gorm:"default:0" json:"order_id"`     // 关联订单ID（售出后填充）
	OrderNo   string         `gorm:"type:varchar(64)" json:"order_no"` // 关联订单号
	SoldAt    *time.Time     `json:"sold_at"`                       // 售出时间
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ManualKamiStatus 卡密状态常量
const (
	ManualKamiStatusAvailable = 0 // 可用
	ManualKamiStatusSold      = 1 // 已售出
	ManualKamiStatusDisabled  = 2 // 已禁用
)

// ProductType 商品类型常量
const (
	ProductTypeManual = 1 // 手动卡密（管理员手动导入，默认模式）
)

// TableName 设置表名
func (ManualKami) TableName() string {
	return "manual_kamis"
}
