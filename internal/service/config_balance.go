// Package service 提供业务逻辑服务
// config_balance.go - 余额系统配置管理
package service

import (
	"encoding/json"
	"fmt"
)

// BalanceConfig 余额系统配置结构
type BalanceConfig struct {
	// 充值限制
	MinRechargeAmount float64 `json:"min_recharge_amount"` // 单笔充值最小金额（元）
	MaxRechargeAmount float64 `json:"max_recharge_amount"` // 单笔充值最大金额（元）
	MaxDailyRecharge  float64 `json:"max_daily_recharge"`  // 每日充值上限（元）
	MaxBalanceLimit   float64 `json:"max_balance_limit"`   // 用户余额上限（元）

	// 告警阈值
	LargeRechargeThreshold    float64 `json:"large_recharge_threshold"`     // 大额充值告警阈值
	LargeConsumeThreshold     float64 `json:"large_consume_threshold"`      // 大额消费告警阈值
	FrequentRechargeCount     int     `json:"frequent_recharge_count"`      // 频繁充值告警次数（每小时）
	FrequentConsumeCount      int     `json:"frequent_consume_count"`       // 频繁消费告警次数（每小时）
	LargeAdminAdjustThreshold float64 `json:"large_admin_adjust_threshold"` // 管理员大额调整告警阈值
}

// DefaultBalanceConfig 默认余额配置
var DefaultBalanceConfig = BalanceConfig{
	MinRechargeAmount:         1.0,
	MaxRechargeAmount:         50000.0,
	MaxDailyRecharge:          100000.0,
	MaxBalanceLimit:           100000.0,
	LargeRechargeThreshold:    1000.0,
	LargeConsumeThreshold:     500.0,
	FrequentRechargeCount:     5,
	FrequentConsumeCount:      10,
	LargeAdminAdjustThreshold: 1000.0,
}

// 全局余额配置缓存
var cachedBalanceConfig *BalanceConfig

// GetBalanceConfig 获取余额配置
func (s *ConfigService) GetBalanceConfig() (*BalanceConfig, error) {
	// 如果有缓存，直接返回
	if cachedBalanceConfig != nil {
		return cachedBalanceConfig, nil
	}

	// 检查 repo 是否已初始化
	if s.repo == nil {
		return &DefaultBalanceConfig, nil
	}

	// 从数据库读取配置
	value, err := s.repo.GetSetting("balance_config")
	if err != nil || value == "" {
		// 数据库中没有配置，返回默认值
		return &DefaultBalanceConfig, nil
	}

	// 解析 JSON 配置
	var cfg BalanceConfig
	if err := json.Unmarshal([]byte(value), &cfg); err != nil {
		return &DefaultBalanceConfig, nil
	}

	// 验证配置有效性，无效则使用默认值
	if cfg.MinRechargeAmount <= 0 {
		cfg.MinRechargeAmount = DefaultBalanceConfig.MinRechargeAmount
	}
	if cfg.MaxRechargeAmount <= 0 {
		cfg.MaxRechargeAmount = DefaultBalanceConfig.MaxRechargeAmount
	}
	if cfg.MaxDailyRecharge <= 0 {
		cfg.MaxDailyRecharge = DefaultBalanceConfig.MaxDailyRecharge
	}
	if cfg.MaxBalanceLimit <= 0 {
		cfg.MaxBalanceLimit = DefaultBalanceConfig.MaxBalanceLimit
	}
	if cfg.LargeRechargeThreshold <= 0 {
		cfg.LargeRechargeThreshold = DefaultBalanceConfig.LargeRechargeThreshold
	}
	if cfg.LargeConsumeThreshold <= 0 {
		cfg.LargeConsumeThreshold = DefaultBalanceConfig.LargeConsumeThreshold
	}
	if cfg.FrequentRechargeCount <= 0 {
		cfg.FrequentRechargeCount = DefaultBalanceConfig.FrequentRechargeCount
	}
	if cfg.FrequentConsumeCount <= 0 {
		cfg.FrequentConsumeCount = DefaultBalanceConfig.FrequentConsumeCount
	}
	if cfg.LargeAdminAdjustThreshold <= 0 {
		cfg.LargeAdminAdjustThreshold = DefaultBalanceConfig.LargeAdminAdjustThreshold
	}

	// 缓存配置
	cachedBalanceConfig = &cfg
	return &cfg, nil
}

// SaveBalanceConfig 保存余额配置
func (s *ConfigService) SaveBalanceConfig(cfg *BalanceConfig) error {
	if s.repo == nil {
		return fmt.Errorf("数据库未连接")
	}

	// 验证配置
	if cfg.MinRechargeAmount <= 0 {
		return fmt.Errorf("单笔充值最小金额必须大于0")
	}
	if cfg.MaxRechargeAmount <= 0 {
		return fmt.Errorf("单笔充值最大金额必须大于0")
	}
	if cfg.MinRechargeAmount > cfg.MaxRechargeAmount {
		return fmt.Errorf("单笔充值最小金额不能大于最大金额")
	}
	if cfg.MaxDailyRecharge <= 0 {
		return fmt.Errorf("每日充值上限必须大于0")
	}
	if cfg.MaxBalanceLimit <= 0 {
		return fmt.Errorf("用户余额上限必须大于0")
	}

	// 序列化为 JSON
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("配置序列化失败: %v", err)
	}

	// 保存到数据库
	err = s.repo.SetSetting("balance_config", string(data), "余额系统配置")
	if err != nil {
		return err
	}

	// 清除缓存
	cachedBalanceConfig = nil
	return nil
}

// ClearBalanceConfigCache 清除余额配置缓存
func ClearBalanceConfigCache() {
	cachedBalanceConfig = nil
}

// GetBalanceLimits 获取余额限制配置（供 BalanceService 使用）
func (s *ConfigService) GetBalanceLimits() (minRecharge, maxRecharge, maxDaily, maxBalance float64) {
	cfg, _ := s.GetBalanceConfig()
	return cfg.MinRechargeAmount, cfg.MaxRechargeAmount, cfg.MaxDailyRecharge, cfg.MaxBalanceLimit
}

// GetBalanceAlertThresholds 获取余额告警阈值配置（供 BalanceAlertService 使用）
func (s *ConfigService) GetBalanceAlertThresholds() (largeRecharge, largeConsume, largeAdminAdjust float64, freqRecharge, freqConsume int) {
	cfg, _ := s.GetBalanceConfig()
	return cfg.LargeRechargeThreshold, cfg.LargeConsumeThreshold, cfg.LargeAdminAdjustThreshold, cfg.FrequentRechargeCount, cfg.FrequentConsumeCount
}
