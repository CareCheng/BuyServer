package model

import (
	"time"
)

// CartItem 购物车项
type CartItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`                          // 用户ID
	ProductID uint      `gorm:"index" json:"product_id"`                       // 商品ID
	Quantity  int       `gorm:"default:1" json:"quantity"`                     // 数量
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// CartSummary 购物车汇总
type CartSummary struct {
	Items      []CartItem `json:"items"`       // 购物车项列表
	TotalCount int        `json:"total_count"` // 商品总数
	TotalPrice float64    `json:"total_price"` // 总价
}
