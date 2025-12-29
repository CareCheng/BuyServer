package config

import (
	"os"
	"strconv"
	"strings"
)

// Environment 环境类型
type Environment string

const (
	// EnvDevelopment 开发环境
	EnvDevelopment Environment = "development"
	// EnvProduction 生产环境
	EnvProduction Environment = "production"
	// EnvTesting 测试环境
	EnvTesting Environment = "testing"
)

// EnvironmentConfig 环境相关配置
type EnvironmentConfig struct {
	// 当前环境
	Env Environment

	// 安全设置
	SecureCookie bool   // Cookie是否启用Secure标志
	CookieDomain string // Cookie作用域

	// 日志设置
	LogLevel   string // 日志级别: debug, info, warn, error
	LogFormat  string // 日志格式: json, text
	LogOutput  string // 日志输出: stdout, file, both
	LogFile    string // 日志文件路径

	// 调试设置
	EnableDebug       bool // 是否启用调试模式
	EnablePprof       bool // 是否启用pprof性能分析
	EnableSQLLog      bool // 是否启用SQL日志
	EnableRequestLog  bool // 是否启用请求日志

	// CORS设置
	AllowOrigins     []string // 允许的跨域来源
	AllowCredentials bool     // 是否允许携带凭证

	// 性能设置
	RateLimitEnabled bool // 是否启用限流
	MaxRequestBody   int  // 最大请求体大小（字节）
}

// 默认环境配置
var defaultEnvConfigs = map[Environment]*EnvironmentConfig{
	EnvDevelopment: {
		Env:               EnvDevelopment,
		SecureCookie:      false,
		CookieDomain:      "",
		LogLevel:          "debug",
		LogFormat:         "text",
		LogOutput:         "stdout",
		LogFile:           "logs/app.log",
		EnableDebug:       true,
		EnablePprof:       false,
		EnableSQLLog:      true,
		EnableRequestLog:  true,
		AllowOrigins:      []string{"*"},
		AllowCredentials:  true,
		RateLimitEnabled:  false,
		MaxRequestBody:    10 * 1024 * 1024, // 10MB
	},
	EnvProduction: {
		Env:               EnvProduction,
		SecureCookie:      true,
		CookieDomain:      "",
		LogLevel:          "info",
		LogFormat:         "json",
		LogOutput:         "both",
		LogFile:           "logs/app.log",
		EnableDebug:       false,
		EnablePprof:       false,
		EnableSQLLog:      false,
		EnableRequestLog:  true,
		AllowOrigins:      []string{}, // 生产环境需要明确配置
		AllowCredentials:  true,
		RateLimitEnabled:  true,
		MaxRequestBody:    5 * 1024 * 1024, // 5MB
	},
	EnvTesting: {
		Env:               EnvTesting,
		SecureCookie:      false,
		CookieDomain:      "",
		LogLevel:          "debug",
		LogFormat:         "text",
		LogOutput:         "stdout",
		LogFile:           "logs/test.log",
		EnableDebug:       true,
		EnablePprof:       false,
		EnableSQLLog:      true,
		EnableRequestLog:  false,
		AllowOrigins:      []string{"*"},
		AllowCredentials:  true,
		RateLimitEnabled:  false,
		MaxRequestBody:    10 * 1024 * 1024, // 10MB
	},
}

// GlobalEnvConfig 全局环境配置
var GlobalEnvConfig *EnvironmentConfig

// InitEnvironmentConfig 初始化环境配置
// 从环境变量读取配置，如果不存在则使用默认值
func InitEnvironmentConfig() *EnvironmentConfig {
	// 获取环境变量 APP_ENV，默认为开发环境
	env := Environment(strings.ToLower(getEnvOrDefault("APP_ENV", string(EnvDevelopment))))

	// 获取基础配置
	config := defaultEnvConfigs[env]
	if config == nil {
		config = defaultEnvConfigs[EnvDevelopment]
	}

	// 复制一份以避免修改默认配置
	envConfig := *config

	// 从环境变量覆盖配置
	envConfig.overrideFromEnv()

	GlobalEnvConfig = &envConfig
	return &envConfig
}

// overrideFromEnv 从环境变量覆盖配置
func (c *EnvironmentConfig) overrideFromEnv() {
	// 安全设置
	if v := os.Getenv("SECURE_COOKIE"); v != "" {
		c.SecureCookie = v == "true" || v == "1"
	}
	if v := os.Getenv("COOKIE_DOMAIN"); v != "" {
		c.CookieDomain = v
	}

	// 日志设置
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		c.LogFormat = v
	}
	if v := os.Getenv("LOG_OUTPUT"); v != "" {
		c.LogOutput = v
	}
	if v := os.Getenv("LOG_FILE"); v != "" {
		c.LogFile = v
	}

	// 调试设置
	if v := os.Getenv("ENABLE_DEBUG"); v != "" {
		c.EnableDebug = v == "true" || v == "1"
	}
	if v := os.Getenv("ENABLE_PPROF"); v != "" {
		c.EnablePprof = v == "true" || v == "1"
	}
	if v := os.Getenv("ENABLE_SQL_LOG"); v != "" {
		c.EnableSQLLog = v == "true" || v == "1"
	}
	if v := os.Getenv("ENABLE_REQUEST_LOG"); v != "" {
		c.EnableRequestLog = v == "true" || v == "1"
	}

	// CORS设置
	if v := os.Getenv("ALLOW_ORIGINS"); v != "" {
		c.AllowOrigins = strings.Split(v, ",")
	}
	if v := os.Getenv("ALLOW_CREDENTIALS"); v != "" {
		c.AllowCredentials = v == "true" || v == "1"
	}

	// 性能设置
	if v := os.Getenv("RATE_LIMIT_ENABLED"); v != "" {
		c.RateLimitEnabled = v == "true" || v == "1"
	}
	if v := os.Getenv("MAX_REQUEST_BODY"); v != "" {
		if size, err := strconv.Atoi(v); err == nil {
			c.MaxRequestBody = size
		}
	}
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment 是否为开发环境
func (c *EnvironmentConfig) IsDevelopment() bool {
	return c.Env == EnvDevelopment
}

// IsProduction 是否为生产环境
func (c *EnvironmentConfig) IsProduction() bool {
	return c.Env == EnvProduction
}

// IsTesting 是否为测试环境
func (c *EnvironmentConfig) IsTesting() bool {
	return c.Env == EnvTesting
}

// GetLogLevel 获取日志级别
func (c *EnvironmentConfig) GetLogLevel() string {
	return c.LogLevel
}

// ShouldLog 判断是否应该记录指定级别的日志
func (c *EnvironmentConfig) ShouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}

	configLevel, ok := levels[strings.ToLower(c.LogLevel)]
	if !ok {
		configLevel = 1 // 默认info级别
	}

	targetLevel, ok := levels[strings.ToLower(level)]
	if !ok {
		return true
	}

	return targetLevel >= configLevel
}

// ==================== 辅助函数 ====================

// GetEnv 获取当前环境
func GetEnv() Environment {
	if GlobalEnvConfig != nil {
		return GlobalEnvConfig.Env
	}
	return EnvDevelopment
}

// IsProd 是否为生产环境（快捷方法）
func IsProd() bool {
	return GetEnv() == EnvProduction
}

// IsDev 是否为开发环境（快捷方法）
func IsDev() bool {
	return GetEnv() == EnvDevelopment
}

// IsTest 是否为测试环境（快捷方法）
func IsTest() bool {
	return GetEnv() == EnvTesting
}
