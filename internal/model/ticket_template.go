package model

import (
	"time"

	"gorm.io/gorm"
)

// TicketTemplate 工单模板模型
type TicketTemplate struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100)" json:"name"`           // 模板名称
	Description string         `gorm:"type:varchar(500)" json:"description"`    // 模板描述
	Category    string         `gorm:"type:varchar(50)" json:"category"`        // 模板分类：order, payment, product, account, other
	Subject     string         `gorm:"type:varchar(200)" json:"subject"`        // 预设主题
	Content     string         `gorm:"type:text" json:"content"`                // 预设内容模板
	Fields      string         `gorm:"type:text" json:"fields"`                 // 自定义字段（JSON格式）
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`            // 图标
	SortOrder   int            `gorm:"default:0" json:"sort_order"`             // 排序
	UseCount    int            `gorm:"default:0" json:"use_count"`              // 使用次数
	Status      int            `gorm:"default:1" json:"status"`                 // 状态：1启用 0禁用
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TicketTemplateField 工单模板自定义字段
type TicketTemplateField struct {
	Name        string   `json:"name"`        // 字段名称
	Label       string   `json:"label"`       // 显示标签
	Type        string   `json:"type"`        // 字段类型：text, textarea, select, number
	Required    bool     `json:"required"`    // 是否必填
	Placeholder string   `json:"placeholder"` // 占位符
	Options     []string `json:"options"`     // 选项（select类型使用）
}

// TicketTemplateCategory 工单模板分类常量
const (
	TemplateOrderCategory   = "order"   // 订单相关
	TemplatePaymentCategory = "payment" // 支付相关
	TemplateProductCategory = "product" // 商品相关
	TemplateAccountCategory = "account" // 账户相关
	TemplateOtherCategory   = "other"   // 其他
)

// TableName 设置表名
func (TicketTemplate) TableName() string {
	return "ticket_templates"
}
