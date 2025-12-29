// Package service 提供业务逻辑服务
// config_system.go - 系统配置管理方法
package service

import (
	"encoding/json"
	"fmt"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
)

// SystemConfig 系统配置结构
type SystemConfig struct {
	SystemTitle         string   `json:"system_title"`
	AdminSuffix         string   `json:"admin_suffix"`
	EnableLogin         bool     `json:"enable_login"`
	AdminUsername       string   `json:"admin_username"`
	AdminPassword       string   `json:"admin_password"`
	Enable2FA           bool     `json:"enable_2fa"`
	TOTPSecret          string   `json:"totp_secret"`
	EnableWhitelist     bool     `json:"enable_whitelist"`
	IPWhitelist         []string `json:"ip_whitelist"`
}

// GetSystemConfig 获取系统配置
func (s *ConfigService) GetSystemConfig() (*SystemConfig, error) {
	// 检查 repo 是否已初始化
	if s.repo == nil {
		globalCfg := config.GlobalConfig.ServerConfig
		return &SystemConfig{
			SystemTitle:         getStringOrDefault(globalCfg.SystemTitle, "卡密购买系统"),
			AdminSuffix:         getStringOrDefault(globalCfg.AdminSuffix, "manage"),
			EnableLogin:         globalCfg.EnableLogin,
			AdminUsername:       getStringOrDefault(globalCfg.AdminUsername, "admin"),
			AdminPassword:       getStringOrDefault(globalCfg.AdminPassword, "admin123"),
			Enable2FA:           globalCfg.Enable2FA,
			TOTPSecret:          globalCfg.TOTPSecret,
			EnableWhitelist:     false,
			IPWhitelist:         []string{},
		}, nil
	}

	dbConfig, err := s.repo.GetSystemConfig()
	if err != nil {
		globalCfg := config.GlobalConfig.ServerConfig
		return &SystemConfig{
			SystemTitle:         getStringOrDefault(globalCfg.SystemTitle, "卡密购买系统"),
			AdminSuffix:         getStringOrDefault(globalCfg.AdminSuffix, "manage"),
			EnableLogin:         globalCfg.EnableLogin,
			AdminUsername:       getStringOrDefault(globalCfg.AdminUsername, "admin"),
			AdminPassword:       getStringOrDefault(globalCfg.AdminPassword, "admin123"),
			Enable2FA:           globalCfg.Enable2FA,
			TOTPSecret:          globalCfg.TOTPSecret,
			EnableWhitelist:     false,
			IPWhitelist:         []string{},
		}, nil
	}

	// 解析IP白名单JSON
	var ipWhitelist []string
	if dbConfig.IPWhitelist != "" {
		json.Unmarshal([]byte(dbConfig.IPWhitelist), &ipWhitelist)
	}

	return &SystemConfig{
		SystemTitle:         dbConfig.SystemTitle,
		AdminSuffix:         dbConfig.AdminSuffix,
		EnableLogin:         dbConfig.EnableLogin,
		AdminUsername:       dbConfig.AdminUsername,
		AdminPassword:       dbConfig.AdminPassword,
		Enable2FA:           dbConfig.Enable2FA,
		TOTPSecret:          dbConfig.TOTPSecret,
		EnableWhitelist:     dbConfig.EnableWhitelist,
		IPWhitelist:         ipWhitelist,
	}, nil
}

// SaveSystemConfig 保存系统配置
func (s *ConfigService) SaveSystemConfig(cfg *SystemConfig) error {
	if s.repo == nil {
		return fmt.Errorf("数据库未连接")
	}

	// 序列化IP白名单为JSON
	ipWhitelistJSON := "[]"
	if len(cfg.IPWhitelist) > 0 {
		if data, err := json.Marshal(cfg.IPWhitelist); err == nil {
			ipWhitelistJSON = string(data)
		}
	}

	dbConfig := &model.SystemConfigDB{
		SystemTitle:         cfg.SystemTitle,
		AdminSuffix:         cfg.AdminSuffix,
		EnableLogin:         cfg.EnableLogin,
		AdminUsername:       cfg.AdminUsername,
		AdminPassword:       cfg.AdminPassword,
		Enable2FA:           cfg.Enable2FA,
		TOTPSecret:          cfg.TOTPSecret,
		EnableWhitelist:     cfg.EnableWhitelist,
		IPWhitelist:         ipWhitelistJSON,
	}
	return s.repo.SaveSystemConfig(dbConfig)
}

// UpdateSystemTitle 更新系统标题
func (s *ConfigService) UpdateSystemTitle(title string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.SystemTitle = title
	return s.SaveSystemConfig(cfg)
}

// UpdateAdminSuffix 更新管理后台路径后缀
func (s *ConfigService) UpdateAdminSuffix(suffix string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.AdminSuffix = suffix
	return s.SaveSystemConfig(cfg)
}

// UpdateSecuritySettings 更新安全设置
func (s *ConfigService) UpdateSecuritySettings(enableLogin bool, adminUsername, adminPassword string, enable2FA bool, totpSecret string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.EnableLogin = enableLogin
	if adminUsername != "" {
		cfg.AdminUsername = adminUsername
	}
	if adminPassword != "" {
		cfg.AdminPassword = adminPassword
	}
	cfg.Enable2FA = enable2FA
	cfg.TOTPSecret = totpSecret
	return s.SaveSystemConfig(cfg)
}

// GetWhitelistConfig 获取白名单配置
func (s *ConfigService) GetWhitelistConfig() (bool, []string, error) {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return false, nil, err
	}
	return cfg.EnableWhitelist, cfg.IPWhitelist, nil
}

// UpdateWhitelistConfig 更新白名单配置
func (s *ConfigService) UpdateWhitelistConfig(enabled bool, whitelist []string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.EnableWhitelist = enabled
	cfg.IPWhitelist = whitelist
	return s.SaveSystemConfig(cfg)
}

// IsIPInWhitelist 检查IP是否在白名单中
func (s *ConfigService) IsIPInWhitelist(ip string) bool {
	cfg, err := s.GetSystemConfig()
	if err != nil || !cfg.EnableWhitelist {
		return true // 白名单未启用时，所有IP都允许
	}

	for _, whiteIP := range cfg.IPWhitelist {
		if whiteIP == ip {
			return true
		}
	}
	return false
}

// NeedsInitialSetup 检查是否需要初始化设置（首次启动）
// 返回 true 表示需要设置初始密码
func (s *ConfigService) NeedsInitialSetup() bool {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		// 无法获取配置，可能是首次启动
		return true
	}
	
	// 检查密码是否为默认密码
	return cfg.AdminPassword == "admin123" || cfg.AdminPassword == ""
}

// SetInitialPassword 设置初始管理员密码
// 只有在密码为默认值时才允许设置
func (s *ConfigService) SetInitialPassword(newPassword string) error {
	if !s.NeedsInitialSetup() {
		return fmt.Errorf("初始密码已设置，无法重复设置")
	}
	
	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度至少6位")
	}
	
	cfg, err := s.GetSystemConfig()
	if err != nil {
		// 创建新配置
		cfg = &SystemConfig{
			SystemTitle:     "卡密购买系统",
			AdminSuffix:     "manage",
			EnableLogin:     true,
			AdminUsername:   "admin",
			AdminPassword:   newPassword,
			Enable2FA:       false,
			EnableWhitelist: false,
			IPWhitelist:     []string{},
		}
	} else {
		cfg.AdminPassword = newPassword
	}
	
	return s.SaveSystemConfig(cfg)
}
