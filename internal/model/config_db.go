package model

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"user-frontend/internal/config"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConfigDB 配置数据库连接（SQLite）
var ConfigDB *gorm.DB

// DBConfigDB 数据库配置（存储在SQLite配置数据库中）
type DBConfigDB struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Type          string    `gorm:"type:varchar(50);default:sqlite" json:"type"`
	Host          string    `gorm:"type:varchar(200)" json:"host"`
	Port          int       `gorm:"default:3306" json:"port"`
	User          string    `gorm:"type:varchar(100)" json:"user"`
	Password      string    `gorm:"type:varchar(255)" json:"password"`
	Database      string    `gorm:"type:varchar(200)" json:"database"`
	ServerPort    int       `gorm:"default:8080" json:"server_port"`         // 服务器监听端口
	EncryptionKey string    `gorm:"type:varchar(100)" json:"encryption_key"` // AES加密密钥（Base64编码）
	KeyLength     int       `gorm:"default:256" json:"key_length"`           // 密钥长度：128/192/256位
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (DBConfigDB) TableName() string {
	return "db_configs"
}

// RedisConfigDB Redis 配置模型
//
// 存储 Redis 连接信息，保存在 SQLite 配置数据库中（与 DBConfigDB 同一位置）。
// 支持单机、哨兵、集群三种模式。
type RedisConfigDB struct {
	// ID 主键，自增
	ID uint `gorm:"primaryKey" json:"id"`

	// ==================== 基本配置 ====================

	// Enabled 是否启用 Redis 缓存
	Enabled bool `gorm:"default:false" json:"enabled"`

	// Mode 部署模式: standalone/sentinel/cluster
	Mode string `gorm:"type:varchar(20);default:standalone" json:"mode"`

	// ==================== 单机模式配置 ====================

	// Host Redis 服务器地址
	Host string `gorm:"type:varchar(200);default:127.0.0.1" json:"host"`

	// Port Redis 服务器端口
	Port int `gorm:"default:6379" json:"port"`

	// Password Redis 认证密码（加密存储）
	Password string `gorm:"type:varchar(255)" json:"password"`

	// Database 数据库编号 (0-15)
	Database int `gorm:"default:0" json:"database"`

	// ==================== 哨兵模式配置 ====================

	// SentinelAddrs 哨兵地址列表（JSON 数组）
	SentinelAddrs string `gorm:"type:text" json:"sentinel_addrs"`

	// SentinelMaster 主节点名称
	SentinelMaster string `gorm:"type:varchar(100)" json:"sentinel_master"`

	// SentinelPassword 哨兵认证密码（加密存储）
	SentinelPassword string `gorm:"type:varchar(255)" json:"sentinel_password"`

	// ==================== 集群模式配置 ====================

	// ClusterAddrs 集群节点地址列表（JSON 数组）
	ClusterAddrs string `gorm:"type:text" json:"cluster_addrs"`

	// ==================== 连接池配置 ====================

	// PoolSize 连接池大小
	PoolSize int `gorm:"default:10" json:"pool_size"`

	// MinIdleConns 最小空闲连接数
	MinIdleConns int `gorm:"default:5" json:"min_idle_conns"`

	// MaxRetries 最大重试次数
	MaxRetries int `gorm:"default:3" json:"max_retries"`

	// DialTimeout 连接超时（秒）
	DialTimeout int `gorm:"default:5" json:"dial_timeout"`

	// ReadTimeout 读取超时（秒）
	ReadTimeout int `gorm:"default:3" json:"read_timeout"`

	// WriteTimeout 写入超时（秒）
	WriteTimeout int `gorm:"default:3" json:"write_timeout"`

	// ==================== 高级配置 ====================

	// KeyPrefix 缓存键前缀
	KeyPrefix string `gorm:"type:varchar(50);default:user:" json:"key_prefix"`

	// DefaultTTL 默认缓存过期时间（秒）
	DefaultTTL int `gorm:"default:300" json:"default_ttl"`

	// EnableMetrics 是否启用监控指标
	EnableMetrics bool `gorm:"default:true" json:"enable_metrics"`

	// ==================== TLS 配置 ====================

	// TLSEnabled 是否启用 TLS 加密
	TLSEnabled bool `gorm:"default:false" json:"tls_enabled"`

	// TLSCert TLS 证书内容（PEM 格式）
	TLSCert string `gorm:"type:text" json:"tls_cert"`

	// TLSKey TLS 私钥内容（PEM 格式，加密存储）
	TLSKey string `gorm:"type:text" json:"tls_key"`

	// TLSCACert CA 证书内容（PEM 格式）
	TLSCACert string `gorm:"type:text" json:"tls_ca_cert"`

	// ==================== 状态信息 ====================

	// LastTestTime 上次连接测试时间
	LastTestTime *time.Time `json:"last_test_time"`

	// LastTestResult 上次测试结果
	LastTestResult bool `json:"last_test_result"`

	// LastTestError 上次测试错误信息
	LastTestError string `gorm:"type:text" json:"last_test_error"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定数据库表名
func (RedisConfigDB) TableName() string {
	return "redis_configs"
}

// ToRedisConfig 转换为运行时配置结构
func (r *RedisConfigDB) ToRedisConfig() *config.RedisConfig {
	cfg := &config.RedisConfig{
		Enabled:       r.Enabled,
		Mode:          r.Mode,
		Host:          r.Host,
		Port:          r.Port,
		Password:      r.Password,
		Database:      r.Database,
		PoolSize:      r.PoolSize,
		MinIdleConns:  r.MinIdleConns,
		MaxRetries:    r.MaxRetries,
		DialTimeout:   r.DialTimeout,
		ReadTimeout:   r.ReadTimeout,
		WriteTimeout:  r.WriteTimeout,
		KeyPrefix:     r.KeyPrefix,
		DefaultTTL:    r.DefaultTTL,
		EnableMetrics: r.EnableMetrics,
		TLSEnabled:    r.TLSEnabled,
		TLSCert:       r.TLSCert,
		TLSKey:        r.TLSKey,
		TLSCACert:     r.TLSCACert,
	}

	// 解析哨兵地址列表
	if r.SentinelAddrs != "" {
		cfg.SentinelAddrs = parseCommaSeparated(r.SentinelAddrs)
	}
	cfg.SentinelMaster = r.SentinelMaster
	cfg.SentinelPassword = r.SentinelPassword

	// 解析集群地址列表
	if r.ClusterAddrs != "" {
		cfg.ClusterAddrs = parseCommaSeparated(r.ClusterAddrs)
	}

	return cfg
}

// parseCommaSeparated 解析逗号分隔的字符串
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// InitConfigDB 初始化配置数据库（SQLite）
func InitConfigDB(configDir string) error {
	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 配置数据库路径
	dbPath := filepath.Join(configDir, "db-config.db")

	var err error
	ConfigDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	// 自动迁移配置表
	if err := ConfigDB.AutoMigrate(&DBConfigDB{}, &RedisConfigDB{}); err != nil {
		return err
	}

	return nil
}
