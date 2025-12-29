package model

import (
	"time"

	"gorm.io/gorm"
)

// KnowledgeCategory 知识库分类
type KnowledgeCategory struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100)" json:"name"`           // 分类名称
	Description string         `gorm:"type:varchar(500)" json:"description"`    // 分类描述
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`            // 图标
	ParentID    uint           `gorm:"default:0" json:"parent_id"`              // 父分类ID（0为顶级）
	SortOrder   int            `gorm:"default:0" json:"sort_order"`             // 排序
	Status      int            `gorm:"default:1" json:"status"`                 // 状态：1启用 0禁用
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// KnowledgeArticle 知识库文章
type KnowledgeArticle struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	CategoryID uint           `gorm:"index" json:"category_id"`                // 分类ID
	Title      string         `gorm:"type:varchar(200)" json:"title"`          // 标题
	Content    string         `gorm:"type:text" json:"content"`                // 内容（Markdown格式）
	Summary    string         `gorm:"type:varchar(500)" json:"summary"`        // 摘要
	Tags       string         `gorm:"type:varchar(500)" json:"tags"`           // 标签（逗号分隔）
	ViewCount  int            `gorm:"default:0" json:"view_count"`             // 浏览次数
	UseCount   int            `gorm:"default:0" json:"use_count"`              // 使用次数（客服引用）
	Helpful    int            `gorm:"default:0" json:"helpful"`                // 有帮助数
	NotHelpful int            `gorm:"default:0" json:"not_helpful"`            // 无帮助数
	SortOrder  int            `gorm:"default:0" json:"sort_order"`             // 排序
	Status     int            `gorm:"default:1" json:"status"`                 // 状态：1发布 0草稿
	CreatedBy  string         `gorm:"type:varchar(100)" json:"created_by"`     // 创建人
	UpdatedBy  string         `gorm:"type:varchar(100)" json:"updated_by"`     // 更新人
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 设置表名
func (KnowledgeCategory) TableName() string {
	return "knowledge_categories"
}

func (KnowledgeArticle) TableName() string {
	return "knowledge_articles"
}
