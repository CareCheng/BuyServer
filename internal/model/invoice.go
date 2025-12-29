package model

import (
	"time"

	"gorm.io/gorm"
)

// InvoiceType 发票类型
const (
	InvoiceTypePersonal   = "personal"   // 个人发票
	InvoiceTypeEnterprise = "enterprise" // 企业发票
)

// InvoiceStatus 发票状态
const (
	InvoiceStatusPending  = 0 // 待开具
	InvoiceStatusIssued   = 1 // 已开具
	InvoiceStatusRejected = 2 // 已拒绝
	InvoiceStatusCanceled = 3 // 已取消
)

// Invoice 发票记录
type Invoice struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	InvoiceNo     string         `gorm:"size:64;uniqueIndex" json:"invoice_no"`      // 发票编号
	UserID        uint           `gorm:"index" json:"user_id"`                       // 用户ID
	OrderNo       string         `gorm:"size:64;index" json:"order_no"`              // 关联订单号
	Type          string         `gorm:"size:20;default:personal" json:"type"`       // 发票类型
	TitleType     string         `gorm:"size:20" json:"title_type"`                  // 抬头类型：personal/enterprise
	Title         string         `gorm:"size:200" json:"title"`                      // 发票抬头
	TaxNo         string         `gorm:"size:50" json:"tax_no"`                      // 税号（企业发票）
	Amount        float64        `json:"amount"`                                     // 发票金额
	Email         string         `gorm:"size:255" json:"email"`                      // 接收邮箱
	Phone         string         `gorm:"size:20" json:"phone"`                       // 联系电话
	Address       string         `gorm:"size:500" json:"address"`                    // 企业地址（企业发票）
	BankName      string         `gorm:"size:100" json:"bank_name"`                  // 开户银行（企业发票）
	BankAccount   string         `gorm:"size:50" json:"bank_account"`                // 银行账号（企业发票）
	Content       string         `gorm:"size:200;default:信息服务费" json:"content"` // 发票内容
	Remark        string         `gorm:"size:500" json:"remark"`                     // 备注
	Status        int            `gorm:"default:0" json:"status"`                    // 状态
	RejectReason  string         `gorm:"size:500" json:"reject_reason"`              // 拒绝原因
	InvoiceURL    string         `gorm:"size:500" json:"invoice_url"`                // 电子发票URL
	IssuedAt      *time.Time     `json:"issued_at"`                                  // 开具时间
	IssuedBy      string         `gorm:"size:50" json:"issued_by"`                   // 开具人
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// InvoiceTitle 发票抬头（用户保存的常用抬头）
type InvoiceTitle struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"index" json:"user_id"`                   // 用户ID
	Type        string         `gorm:"size:20;default:personal" json:"type"`   // 类型：personal/enterprise
	Title       string         `gorm:"size:200" json:"title"`                  // 抬头名称
	TaxNo       string         `gorm:"size:50" json:"tax_no"`                  // 税号
	Address     string         `gorm:"size:500" json:"address"`                // 企业地址
	Phone       string         `gorm:"size:20" json:"phone"`                   // 联系电话
	BankName    string         `gorm:"size:100" json:"bank_name"`              // 开户银行
	BankAccount string         `gorm:"size:50" json:"bank_account"`            // 银行账号
	IsDefault   bool           `gorm:"default:false" json:"is_default"`        // 是否默认
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// InvoiceConfig 发票配置
type InvoiceConfig struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Enabled        bool      `gorm:"default:false" json:"enabled"`                    // 是否启用发票功能
	MinAmount      float64   `gorm:"default:0" json:"min_amount"`                     // 最低开票金额
	AutoIssue      bool      `gorm:"default:false" json:"auto_issue"`                 // 是否自动开具
	AllowPersonal  bool      `gorm:"default:true" json:"allow_personal"`              // 允许个人发票
	AllowEnterprise bool     `gorm:"default:true" json:"allow_enterprise"`            // 允许企业发票
	DefaultContent string    `gorm:"size:200;default:信息服务费" json:"default_content"` // 默认发票内容
	CompanyName    string    `gorm:"size:200" json:"company_name"`                    // 开票公司名称
	CompanyTaxNo   string    `gorm:"size:50" json:"company_tax_no"`                   // 开票公司税号
	CompanyAddress string    `gorm:"size:500" json:"company_address"`                 // 开票公司地址
	CompanyPhone   string    `gorm:"size:20" json:"company_phone"`                    // 开票公司电话
	CompanyBank    string    `gorm:"size:100" json:"company_bank"`                    // 开票公司开户银行
	CompanyAccount string    `gorm:"size:50" json:"company_account"`                  // 开票公司银行账号
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 设置表名
func (Invoice) TableName() string {
	return "invoices"
}

// TableName 设置表名
func (InvoiceTitle) TableName() string {
	return "invoice_titles"
}

// TableName 设置表名
func (InvoiceConfig) TableName() string {
	return "invoice_configs"
}
