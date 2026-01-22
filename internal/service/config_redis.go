// Package service 提供业务逻辑服务
// config_redis.go - Redis 配置管理
//
// 本模块负责 Redis 配置的读取和保存。
// 配置存储在 SQLite 配置数据库中。
package service

import (
	"encoding/json"
	"time"

	"user-frontend/internal/cache"
	"user-frontend/internal/config"
	"user-frontend/internal/model"
)

// ==================== Redis 配置管理 ====================

// GetRedisConfig 获取 Redis 配置
//
// 从 SQLite 配置数据库读取 Redis 配置。
// 如果配置不存在，返回默认配置。
func (s *ConfigService) GetRedisConfig() (*config.RedisConfig, error) {
	if s.configDB == nil {
		return s.getDefaultRedisConfig(), nil
	}

	var dbConfig model.RedisConfigDB
	if err := s.configDB.First(&dbConfig).Error; err != nil {
		// 配置不存在，返回默认值
		return s.getDefaultRedisConfig(), nil
	}

	return s.convertRedisConfig(&dbConfig)
}

// SaveRedisConfig 保存 Redis 配置
//
// 将 Redis 配置保存到 SQLite 配置数据库。
// 敏感信息（密码）会进行加密存储。
func (s *ConfigService) SaveRedisConfig(cfg *config.RedisConfig) error {
	if s.configDB == nil {
		return nil
	}

	// 查找现有配置
	var dbConfig model.RedisConfigDB
	s.configDB.First(&dbConfig)

	// 更新配置
	dbConfig.Enabled = cfg.Enabled
	dbConfig.Mode = cfg.Mode
	dbConfig.Host = cfg.Host
	dbConfig.Port = cfg.Port
	dbConfig.Database = cfg.Database
	dbConfig.PoolSize = cfg.PoolSize
	dbConfig.MinIdleConns = cfg.MinIdleConns
	dbConfig.MaxRetries = cfg.MaxRetries
	dbConfig.DialTimeout = cfg.DialTimeout
	dbConfig.ReadTimeout = cfg.ReadTimeout
	dbConfig.WriteTimeout = cfg.WriteTimeout
	dbConfig.KeyPrefix = cfg.KeyPrefix
	dbConfig.DefaultTTL = cfg.DefaultTTL
	dbConfig.EnableMetrics = cfg.EnableMetrics
	dbConfig.TLSEnabled = cfg.TLSEnabled
	dbConfig.TLSCert = cfg.TLSCert
	dbConfig.TLSCACert = cfg.TLSCACert
	dbConfig.SentinelMaster = cfg.SentinelMaster

	// 加密密码
	if cfg.Password != "" {
		encrypted, err := encryptPassword(cfg.Password)
		if err != nil {
			return err
		}
		dbConfig.Password = encrypted
	}

	// 加密哨兵密码
	if cfg.SentinelPassword != "" {
		encrypted, err := encryptPassword(cfg.SentinelPassword)
		if err != nil {
			return err
		}
		dbConfig.SentinelPassword = encrypted
	}

	// 加密 TLS 私钥
	if cfg.TLSKey != "" {
		encrypted, err := encryptPassword(cfg.TLSKey)
		if err != nil {
			return err
		}
		dbConfig.TLSKey = encrypted
	}

	// 序列化地址列表
	if len(cfg.SentinelAddrs) > 0 {
		addrsJSON, _ := json.Marshal(cfg.SentinelAddrs)
		dbConfig.SentinelAddrs = string(addrsJSON)
	}

	if len(cfg.ClusterAddrs) > 0 {
		addrsJSON, _ := json.Marshal(cfg.ClusterAddrs)
		dbConfig.ClusterAddrs = string(addrsJSON)
	}

	// 保存
	if dbConfig.ID == 0 {
		return s.configDB.Create(&dbConfig).Error
	}
	return s.configDB.Save(&dbConfig).Error
}

// getDefaultRedisConfig 获取默认 Redis 配置
func (s *ConfigService) getDefaultRedisConfig() *config.RedisConfig {
	return &config.RedisConfig{
		Enabled:       false,
		Mode:          "standalone",
		Host:          "127.0.0.1",
		Port:          6379,
		Database:      0,
		PoolSize:      10,
		MinIdleConns:  5,
		MaxRetries:    3,
		DialTimeout:   5,
		ReadTimeout:   3,
		WriteTimeout:  3,
		KeyPrefix:     "user:",
		DefaultTTL:    300,
		EnableMetrics: true,
	}
}

// convertRedisConfig 转换数据库模型为配置结构
func (s *ConfigService) convertRedisConfig(db *model.RedisConfigDB) (*config.RedisConfig, error) {
	cfg := &config.RedisConfig{
		Enabled:       db.Enabled,
		Mode:          db.Mode,
		Host:          db.Host,
		Port:          db.Port,
		Database:      db.Database,
		PoolSize:      db.PoolSize,
		MinIdleConns:  db.MinIdleConns,
		MaxRetries:    db.MaxRetries,
		DialTimeout:   db.DialTimeout,
		ReadTimeout:   db.ReadTimeout,
		WriteTimeout:  db.WriteTimeout,
		KeyPrefix:     db.KeyPrefix,
		DefaultTTL:    db.DefaultTTL,
		EnableMetrics: db.EnableMetrics,
		TLSEnabled:    db.TLSEnabled,
		TLSCert:       db.TLSCert,
		TLSCACert:     db.TLSCACert,
		SentinelMaster: db.SentinelMaster,
	}

	// 解密密码
	if db.Password != "" {
		if decrypted, err := decryptPassword(db.Password); err == nil {
			cfg.Password = decrypted
		}
	}

	// 解密哨兵密码
	if db.SentinelPassword != "" {
		if decrypted, err := decryptPassword(db.SentinelPassword); err == nil {
			cfg.SentinelPassword = decrypted
		}
	}

	// 解密 TLS 私钥
	if db.TLSKey != "" {
		if decrypted, err := decryptPassword(db.TLSKey); err == nil {
			cfg.TLSKey = decrypted
		}
	}

	// 解析地址列表
	if db.SentinelAddrs != "" {
		json.Unmarshal([]byte(db.SentinelAddrs), &cfg.SentinelAddrs)
	}
	if db.ClusterAddrs != "" {
		json.Unmarshal([]byte(db.ClusterAddrs), &cfg.ClusterAddrs)
	}

	return cfg, nil
}

// TestRedisConnection 测试 Redis 连接
//
// 使用提供的配置测试 Redis 连接是否正常。
func (s *ConfigService) TestRedisConnection(cfg *config.RedisConfig) (bool, string, int64) {
	if cfg == nil {
		return false, "配置为空", 0
	}

	// 转换为 cache.RedisConfig
	cacheConfig := &cache.RedisConfig{
		Enabled:          true, // 测试时强制启用
		Mode:             cfg.Mode,
		Host:             cfg.Host,
		Port:             cfg.Port,
		Password:         cfg.Password,
		Database:         cfg.Database,
		SentinelAddrs:    cfg.SentinelAddrs,
		SentinelMaster:   cfg.SentinelMaster,
		SentinelPassword: cfg.SentinelPassword,
		ClusterAddrs:     cfg.ClusterAddrs,
		PoolSize:         cfg.PoolSize,
		MinIdleConns:     cfg.MinIdleConns,
		MaxRetries:       cfg.MaxRetries,
		DialTimeout:      cfg.DialTimeout,
		ReadTimeout:      cfg.ReadTimeout,
		WriteTimeout:     cfg.WriteTimeout,
		TLSEnabled:       cfg.TLSEnabled,
		TLSCert:          cfg.TLSCert,
		TLSKey:           cfg.TLSKey,
		TLSCACert:        cfg.TLSCACert,
	}

	// 记录开始时间
	startTime := time.Now()

	// 尝试创建连接
	redisCache, err := cache.NewRedisCache(cacheConfig)
	if err != nil {
		return false, err.Error(), 0
	}
	defer redisCache.Close()

	// 计算耗时（毫秒）
	duration := time.Since(startTime).Milliseconds()

	return true, "连接成功", duration
}
