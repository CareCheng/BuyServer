package model

import (
	"time"

	"gorm.io/gorm"
)

// ProductReview 商品评价模型
type ProductReview struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProductID uint           `gorm:"index" json:"product_id"`                   // 商品ID
	UserID    uint           `gorm:"index" json:"user_id"`                      // 用户ID
	Username  string         `gorm:"type:varchar(100)" json:"username"`         // 用户名
	OrderNo   string         `gorm:"type:varchar(64);index" json:"order_no"`    // 订单号
	Rating    int            `gorm:"default:5" json:"rating"`                   // 评分 1-5星
	Content   string         `gorm:"type:text" json:"content"`                  // 评价内容
	Images    string         `gorm:"type:text" json:"images"`                   // 评价图片（JSON数组）
	Reply     string         `gorm:"type:text" json:"reply"`                    // 商家回复
	ReplyAt   *time.Time     `json:"reply_at"`                                  // 回复时间
	Status    int            `gorm:"default:1" json:"status"`                   // 状态：1显示 0隐藏
	IsAnon    bool           `gorm:"default:false" json:"is_anon"`              // 是否匿名评价
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProductReviewStats 商品评价统计
type ProductReviewStats struct {
	ProductID    uint    `json:"product_id"`
	TotalCount   int64   `json:"total_count"`    // 总评价数
	AvgRating    float64 `json:"avg_rating"`     // 平均评分
	Rating5Count int64   `json:"rating_5_count"` // 5星数量
	Rating4Count int64   `json:"rating_4_count"` // 4星数量
	Rating3Count int64   `json:"rating_3_count"` // 3星数量
	Rating2Count int64   `json:"rating_2_count"` // 2星数量
	Rating1Count int64   `json:"rating_1_count"` // 1星数量
}

// TableName 设置表名
func (ProductReview) TableName() string {
	return "product_reviews"
}
