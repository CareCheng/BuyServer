package model

import (
	"time"

	"gorm.io/datatypes"
)

// HomepageConfig 首页配置模型
type HomepageConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Template  string         `gorm:"size:50;default:modern" json:"template"`           // 模板名称: modern, minimal, gradient, card, hero, business
	Config    datatypes.JSON `gorm:"type:json" json:"config"`                          // JSON 配置数据
	IsActive  bool           `gorm:"default:true" json:"is_active"`                    // 是否启用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// HomepageSection 首页区块配置
type HomepageSection struct {
	Type    string                 `json:"type"`    // hero, features, announcement, products, stats, testimonials, cta
	Enabled bool                   `json:"enabled"` // 是否启用
	Order   int                    `json:"order"`   // 排序
	Config  map[string]interface{} `json:"config"`  // 区块配置
}

// HomepageFullConfig 完整首页配置
type HomepageFullConfig struct {
	// 基础设置
	Template       string `json:"template"`        // 模板名称
	PrimaryColor   string `json:"primary_color"`   // 主色调
	SecondaryColor string `json:"secondary_color"` // 次色调

	// 高级模式（自定义 HTML）
	AdvancedMode bool   `json:"advanced_mode"` // 是否启用高级模式
	CustomHTML   string `json:"custom_html"`   // 自定义 HTML 代码
	CustomCSS    string `json:"custom_css"`    // 自定义 CSS 样式
	CustomJS     string `json:"custom_js"`     // 自定义 JavaScript 代码
	
	// Logo 设置
	LogoType string `json:"logo_type"` // text, image, emoji
	LogoText string `json:"logo_text"` // Logo 文字
	LogoImage string `json:"logo_image"` // Logo 图片 URL
	LogoEmoji string `json:"logo_emoji"` // Logo Emoji
	
	// Hero 区块
	HeroEnabled     bool   `json:"hero_enabled"`
	HeroTitle       string `json:"hero_title"`
	HeroSubtitle    string `json:"hero_subtitle"`
	HeroButtonText  string `json:"hero_button_text"`
	HeroButtonLink  string `json:"hero_button_link"`
	HeroBackground  string `json:"hero_background"`  // gradient, image, solid
	HeroBgImage     string `json:"hero_bg_image"`    // 背景图片 URL
	HeroBgColor     string `json:"hero_bg_color"`    // 背景颜色
	
	// 特性区块
	FeaturesEnabled bool              `json:"features_enabled"`
	FeaturesTitle   string            `json:"features_title"`
	Features        []FeatureItem     `json:"features"`
	
	// 公告区块
	AnnouncementEnabled bool   `json:"announcement_enabled"`
	AnnouncementTitle   string `json:"announcement_title"`
	AnnouncementContent string `json:"announcement_content"`
	AnnouncementType    string `json:"announcement_type"` // info, warning, success
	
	// 商品展示区块
	ProductsEnabled bool   `json:"products_enabled"`
	ProductsTitle   string `json:"products_title"`
	ProductsCount   int    `json:"products_count"` // 展示数量
	
	// 统计区块
	StatsEnabled bool       `json:"stats_enabled"`
	Stats        []StatItem `json:"stats"`
	
	// CTA 区块
	CTAEnabled    bool   `json:"cta_enabled"`
	CTATitle      string `json:"cta_title"`
	CTASubtitle   string `json:"cta_subtitle"`
	CTAButtonText string `json:"cta_button_text"`
	CTAButtonLink string `json:"cta_button_link"`
	
	// 页脚设置
	FooterText    string `json:"footer_text"`
	FooterLinks   []FooterLink `json:"footer_links"`
	
	// 浮动按钮
	FloatingButtonEnabled bool   `json:"floating_button_enabled"`
	FloatingButtonIcon    string `json:"floating_button_icon"`
	FloatingButtonLink    string `json:"floating_button_link"`
}

// FeatureItem 特性项
type FeatureItem struct {
	Icon        string `json:"icon"`        // emoji 或 fa 图标
	Title       string `json:"title"`
	Description string `json:"description"`
}

// StatItem 统计项
type StatItem struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
}

// FooterLink 页脚链接
type FooterLink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// TableName 表名
func (HomepageConfig) TableName() string {
	return "homepage_configs"
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig(template string) HomepageFullConfig {
	config := HomepageFullConfig{
		Template:       template,
		PrimaryColor:   "#6366f1",
		SecondaryColor: "#8b5cf6",
		
		LogoType:  "emoji",
		LogoText:  "卡密购买系统",
		LogoEmoji: "🔐",
		
		HeroEnabled:    true,
		HeroTitle:      "欢迎使用卡密购买系统",
		HeroSubtitle:   "安全、便捷的卡密购买平台",
		HeroButtonText: "浏览商品",
		HeroButtonLink: "/products/",
		HeroBackground: "gradient",
		
		FeaturesEnabled: true,
		FeaturesTitle:   "为什么选择我们",
		Features: []FeatureItem{
			{Icon: "🔒", Title: "安全可靠", Description: "采用ECC加密通信，保障交易安全"},
			{Icon: "⚡", Title: "即时发货", Description: "支付成功后立即获取卡密"},
			{Icon: "💬", Title: "售后保障", Description: "专业客服团队，随时为您服务"},
		},
		
		AnnouncementEnabled: false,
		AnnouncementTitle:   "系统公告",
		AnnouncementContent: "",
		AnnouncementType:    "info",
		
		ProductsEnabled: true,
		ProductsTitle:   "热门商品",
		ProductsCount:   6,
		
		StatsEnabled: true,
		Stats: []StatItem{
			{Value: "10000+", Label: "用户数量", Icon: "👥"},
			{Value: "50000+", Label: "成交订单", Icon: "📦"},
			{Value: "99.9%", Label: "好评率", Icon: "⭐"},
			{Value: "24/7", Label: "在线客服", Icon: "💬"},
		},
		
		CTAEnabled:    true,
		CTATitle:      "准备好开始了吗？",
		CTASubtitle:   "立即注册，享受便捷的购买体验",
		CTAButtonText: "立即注册",
		CTAButtonLink: "/register/",
		
		FooterText: "卡密购买系统",
		FooterLinks: []FooterLink{
			{Text: "常见问题", URL: "/faq/"},
			{Text: "联系客服", URL: "/message/"},
		},
		
		FloatingButtonEnabled: true,
		FloatingButtonIcon:    "fa-headset",
		FloatingButtonLink:    "/message/",
	}
	
	return config
}
