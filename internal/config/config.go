package config

import (
	"os"
	"sync"
)

// Config 全局配置
type Config struct {
	DBConfig      DBConfig      `json:"db_config"`       // 数据库配置（从SQLite配置数据库加载）
	ServerConfig  ServerConfig  `json:"server_config"`   // 服务器配置（运行时从主数据库加载）
	PaymentConfig PaymentConfig `json:"payment_config"`  // 支付配置（运行时从主数据库加载）
	EmailConfig   EmailConfig   `json:"email_config"`    // 邮箱配置（运行时从主数据库加载）
	ConfigDir     string        `json:"-"`
	mu            sync.RWMutex
}

// EmailConfig 邮箱配置
type EmailConfig struct {
	Enabled      bool   `json:"enabled"`
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	FromName     string `json:"from_name"`
	FromEmail    string `json:"from_email"`
	Encryption   string `json:"encryption"` // 加密方式：none/ssl/starttls
	CodeLength   int    `json:"code_length"`
}

// PaymentConfig 支付配置
type PaymentConfig struct {
	AlipayF2F AlipayF2FConfig `json:"alipay_f2f"`
	WechatPay WechatPayConfig `json:"wechat_pay"`
	YiPay     YiPayConfig     `json:"yi_pay"`
	PayPal    PayPalConfig    `json:"paypal"`
	// Stripe支付配置
	StripeEnabled        bool   `json:"stripe_enabled"`         // 是否启用Stripe
	StripePublishableKey string `json:"stripe_publishable_key"` // Stripe公钥（前端使用）
	StripeSecretKey      string `json:"stripe_secret_key"`      // Stripe私钥（后端使用）
	StripeWebhookSecret  string `json:"stripe_webhook_secret"`  // Webhook签名密钥
	StripeCurrency       string `json:"stripe_currency"`        // 货币代码，默认usd
	// USDT支付配置
	USDTEnabled       bool    `json:"usdt_enabled"`        // 是否启用USDT
	USDTNetwork       string  `json:"usdt_network"`        // 网络类型：TRC20, ERC20, BEP20
	USDTWalletAddress string  `json:"usdt_wallet_address"` // 收款钱包地址
	USDTAPIProvider   string  `json:"usdt_api_provider"`   // API提供商：nowpayments, coingate, manual
	USDTAPIKey        string  `json:"usdt_api_key"`        // API密钥
	USDTAPISecret     string  `json:"usdt_api_secret"`     // API密钥（部分提供商需要）
	USDTWebhookSecret string  `json:"usdt_webhook_secret"` // Webhook签名密钥
	USDTExchangeRate  float64 `json:"usdt_exchange_rate"`  // 汇率（手动模式使用）
	USDTMinAmount     float64 `json:"usdt_min_amount"`     // 最小支付金额（USDT）
	USDTConfirmations int     `json:"usdt_confirmations"`  // 需要的确认数
}

// PayPalConfig PayPal支付配置
type PayPalConfig struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Sandbox      bool   `json:"sandbox"`
	Currency     string `json:"currency"`
	ReturnURL    string `json:"return_url"`
	CancelURL    string `json:"cancel_url"`
}

// DBConfig 数据库配置（从SQLite配置数据库加载）
type DBConfig struct {
	Type     string `json:"type"` // mysql, postgres, sqlite
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// ServerConfig 服务器配置（运行时从主数据库加载）
type ServerConfig struct {
	Port          int    `json:"port"`
	UseHTTPS      bool   `json:"use_https"`
	CertFile      string `json:"cert_file"`
	KeyFile       string `json:"key_file"`
	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`
	AdminSuffix   string `json:"admin_suffix"`
	EnableLogin   bool   `json:"enable_login"`
	Enable2FA     bool   `json:"enable_2fa"`
	TOTPSecret    string `json:"totp_secret"`
	SystemTitle   string `json:"system_title"`
}

// AlipayF2FConfig 支付宝当面付配置
type AlipayF2FConfig struct {
	Enabled    bool   `json:"enabled"`
	AppID      string `json:"app_id"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	NotifyURL  string `json:"notify_url"`
}

// WechatPayConfig 微信支付配置
type WechatPayConfig struct {
	Enabled   bool   `json:"enabled"`
	AppID     string `json:"app_id"`
	MchID     string `json:"mch_id"`
	APIKey    string `json:"api_key"`
	NotifyURL string `json:"notify_url"`
}

// YiPayConfig 易支付配置
type YiPayConfig struct {
	Enabled   bool   `json:"enabled"`
	APIURL    string `json:"api_url"`
	PID       string `json:"pid"`
	Key       string `json:"key"`
	NotifyURL string `json:"notify_url"`
	ReturnURL string `json:"return_url"`
}

var (
	GlobalConfig *Config
	once         sync.Once
)

func InitConfig(configDir string) (*Config, error) {
	var err error
	once.Do(func() {
		GlobalConfig = &Config{
			ConfigDir: configDir,
		}
		os.MkdirAll(configDir, 0755)
		err = GlobalConfig.LoadAll()
	})
	return GlobalConfig, err
}

// LoadAll 加载所有配置（数据库配置从SQLite配置数据库加载，其他配置从主数据库加载）
func (c *Config) LoadAll() error {
	// 数据库配置将通过 ConfigService 从 SQLite 配置数据库加载
	// 这里只设置默认值
	c.DBConfig = DBConfig{Type: "sqlite", Database: "user_config/user_data.db", Port: 3306}

	// 设置默认的服务器配置（实际值从数据库加载）
	c.ServerConfig = ServerConfig{
		Port:          8080,
		AdminUsername: "admin",
		AdminPassword: "admin123",
		AdminSuffix:   "manage",
		EnableLogin:   true,
		Enable2FA:     false,
		SystemTitle:   "卡密购买系统",
	}

	// 设置默认的邮箱配置（实际值从数据库加载）
	c.EmailConfig = EmailConfig{
		SMTPPort:   465,
		Encryption: "ssl",
		CodeLength: 6,
	}

	// 支付配置默认为空（实际值从数据库加载）
	c.PaymentConfig = PaymentConfig{}

	return nil
}

// SetDBConfig 设置数据库配置（由 ConfigService 调用）
func (c *Config) SetDBConfig(cfg DBConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.DBConfig = cfg
}
