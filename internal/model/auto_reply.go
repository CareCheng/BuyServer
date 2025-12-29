package model

import (
	"time"

	"gorm.io/gorm"
)

// AutoReplyRule 自动回复规则
type AutoReplyRule struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100" json:"name"`                  // 规则名称
	Keywords    string         `gorm:"type:text" json:"keywords"`             // 关键词（JSON数组）
	MatchType   string         `gorm:"size:20;default:contains" json:"match_type"` // 匹配类型：contains/exact/regex
	Reply       string         `gorm:"type:text" json:"reply"`                // 回复内容
	Priority    int            `gorm:"default:0" json:"priority"`             // 优先级（数字越大优先级越高）
	Status      int            `gorm:"default:1" json:"status"`               // 状态：1启用 0禁用
	HitCount    int            `gorm:"default:0" json:"hit_count"`            // 命中次数
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// AutoReplyConfig 智能客服配置
type AutoReplyConfig struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Enabled           bool      `gorm:"default:false" json:"enabled"`              // 是否启用智能客服
	WelcomeMessage    string    `gorm:"type:text" json:"welcome_message"`          // 欢迎语
	NoMatchReply      string    `gorm:"type:text" json:"no_match_reply"`           // 无匹配时的回复
	TransferKeywords  string    `gorm:"type:text" json:"transfer_keywords"`        // 转人工关键词（JSON数组）
	TransferMessage   string    `gorm:"type:text" json:"transfer_message"`         // 转人工提示语
	WorkingHoursOnly  bool      `gorm:"default:false" json:"working_hours_only"`   // 仅工作时间启用
	WorkingHoursStart string    `gorm:"size:10" json:"working_hours_start"`        // 工作时间开始（如 09:00）
	WorkingHoursEnd   string    `gorm:"size:10" json:"working_hours_end"`          // 工作时间结束（如 18:00）
	UpdatedAt         time.Time `json:"updated_at"`
}

// AutoReplyLog 自动回复日志
type AutoReplyLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SessionID   string    `gorm:"size:64;index" json:"session_id"`   // 会话ID
	UserID      uint      `gorm:"index" json:"user_id"`              // 用户ID（0表示游客）
	UserMessage string    `gorm:"type:text" json:"user_message"`     // 用户消息
	RuleID      uint      `gorm:"index" json:"rule_id"`              // 匹配的规则ID（0表示无匹配）
	RuleName    string    `gorm:"size:100" json:"rule_name"`         // 规则名称
	BotReply    string    `gorm:"type:text" json:"bot_reply"`        // 机器人回复
	Transferred bool      `gorm:"default:false" json:"transferred"`  // 是否转人工
	CreatedAt   time.Time `json:"created_at"`
}
