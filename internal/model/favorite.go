package model

import "time"

// ProductFavorite 商品收藏模型
// 用户可以收藏感兴趣的商品，方便后续查看和购买
type ProductFavorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`                          // 用户ID
	ProductID uint      `gorm:"index;not null" json:"product_id"`                       // 商品ID
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                       // 收藏时间
}

// TableName 指定表名
func (ProductFavorite) TableName() string {
	return "product_favorites"
}
