package model

import (
	"time"

	"gorm.io/gorm"
)

// FAQ 常见问题模型
type FAQ struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	CategoryID uint           `gorm:"index" json:"category_id"`                // 分类ID
	Question   string         `gorm:"type:varchar(500)" json:"question"`       // 问题
	Answer     string         `gorm:"type:text" json:"answer"`                 // 答案
	SortOrder  int            `gorm:"default:0" json:"sort_order"`             // 排序
	ViewCount  int            `gorm:"default:0" json:"view_count"`             // 浏览次数
	Helpful    int            `gorm:"default:0" json:"helpful"`                // 有帮助数
	NotHelpful int            `gorm:"default:0" json:"not_helpful"`            // 无帮助数
	Status     int            `gorm:"default:1" json:"status"`                 // 状态：1启用 0禁用
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// FAQCategory FAQ分类模型
type FAQCategory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100)" json:"name"`  // 分类名称
	Icon      string         `gorm:"type:varchar(50)" json:"icon"`   // 图标
	SortOrder int            `gorm:"default:0" json:"sort_order"`    // 排序
	Status    int            `gorm:"default:1" json:"status"`        // 状态：1启用 0禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// FAQFeedback FAQ反馈记录（防止重复反馈）
type FAQFeedback struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FAQID     uint      `gorm:"index" json:"faq_id"`
	UserID    uint      `gorm:"index" json:"user_id"`             // 0表示游客
	SessionID string    `gorm:"type:varchar(64);index" json:"session_id"` // 游客会话ID
	Helpful   bool      `json:"helpful"`                          // true有帮助 false无帮助
	CreatedAt time.Time `json:"created_at"`
}

// TableName 设置表名
func (FAQ) TableName() string {
	return "faqs"
}

func (FAQCategory) TableName() string {
	return "faq_categories"
}

func (FAQFeedback) TableName() string {
	return "faq_feedbacks"
}
