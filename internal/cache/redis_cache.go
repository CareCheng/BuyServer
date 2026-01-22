// Package cache 提供统一的缓存抽象层
// redis_cache.go - Redis 缓存实现
package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis 缓存实现
type RedisCache struct {
	client redis.UniversalClient
	ctx    context.Context
}

// RedisConfig Redis 配置
type RedisConfig struct {
	// 基本配置
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"` // standalone/sentinel/cluster

	// 单机模式配置
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`

	// 哨兵模式配置
	SentinelAddrs    []string `json:"sentinel_addrs"`
	SentinelMaster   string   `json:"sentinel_master"`
	SentinelPassword string   `json:"sentinel_password"`

	// 集群模式配置
	ClusterAddrs []string `json:"cluster_addrs"`

	// 连接池配置
	PoolSize     int `json:"pool_size"`
	MinIdleConns int `json:"min_idle_conns"`
	MaxRetries   int `json:"max_retries"`
	DialTimeout  int `json:"dial_timeout"`
	ReadTimeout  int `json:"read_timeout"`
	WriteTimeout int `json:"write_timeout"`

	// 高级配置
	KeyPrefix     string `json:"key_prefix"`
	DefaultTTL    int    `json:"default_ttl"`
	EnableMetrics bool   `json:"enable_metrics"`

	// TLS 配置
	TLSEnabled bool   `json:"tls_enabled"`
	TLSCert    string `json:"tls_cert"`
	TLSKey     string `json:"tls_key"`
	TLSCACert  string `json:"tls_ca_cert"`
}

// NewRedisCache 创建 Redis 缓存实例
func NewRedisCache(cfg *RedisConfig) (*RedisCache, error) {
	var client redis.UniversalClient

	// TLS 配置
	var tlsConfig *tls.Config
	if cfg.TLSEnabled {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true, // 根据需要配置
		}
	}

	switch cfg.Mode {
	case "sentinel":
		// 哨兵模式
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       cfg.SentinelMaster,
			SentinelAddrs:    cfg.SentinelAddrs,
			SentinelPassword: cfg.SentinelPassword,
			Password:         cfg.Password,
			DB:               cfg.Database,
			PoolSize:         cfg.PoolSize,
			MinIdleConns:     cfg.MinIdleConns,
			MaxRetries:       cfg.MaxRetries,
			DialTimeout:      time.Duration(cfg.DialTimeout) * time.Second,
			ReadTimeout:      time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:     time.Duration(cfg.WriteTimeout) * time.Second,
			TLSConfig:        tlsConfig,
		})

	case "cluster":
		// 集群模式
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.ClusterAddrs,
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
			TLSConfig:    tlsConfig,
		})

	default:
		// 单机模式
		client = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password:     cfg.Password,
			DB:           cfg.Database,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
			TLSConfig:    tlsConfig,
		})
	}

	ctx := context.Background()

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis 连接失败: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// Get 获取缓存值
func (c *RedisCache) Get(key string) (interface{}, bool) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		return nil, false
	}

	// 尝试反序列化 JSON
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// 如果不是 JSON，返回原始字符串
		return val, true
	}
	return result, true
}

// GetString 获取字符串类型缓存值
func (c *RedisCache) GetString(key string) (string, bool) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return val, true
}

// Set 设置缓存值
func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	// 序列化为 JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(c.ctx, key, data, ttl).Err()
}

// SetString 设置字符串类型缓存值
func (c *RedisCache) SetString(key string, value string, ttl time.Duration) error {
	return c.client.Set(c.ctx, key, value, ttl).Err()
}

// Delete 删除缓存
func (c *RedisCache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// Exists 检查键是否存在
func (c *RedisCache) Exists(key string) bool {
	result, err := c.client.Exists(c.ctx, key).Result()
	return err == nil && result > 0
}

// Expire 设置过期时间
func (c *RedisCache) Expire(key string, ttl time.Duration) error {
	return c.client.Expire(c.ctx, key, ttl).Err()
}

// TTL 获取剩余过期时间
func (c *RedisCache) TTL(key string) (time.Duration, error) {
	return c.client.TTL(c.ctx, key).Result()
}

// Incr 原子自增
func (c *RedisCache) Incr(key string) (int64, error) {
	return c.client.Incr(c.ctx, key).Result()
}

// IncrBy 原子自增指定值
func (c *RedisCache) IncrBy(key string, delta int64) (int64, error) {
	return c.client.IncrBy(c.ctx, key, delta).Result()
}

// Decr 原子自减
func (c *RedisCache) Decr(key string) (int64, error) {
	return c.client.Decr(c.ctx, key).Result()
}

// Keys 获取匹配模式的所有键
func (c *RedisCache) Keys(pattern string) ([]string, error) {
	return c.client.Keys(c.ctx, pattern).Result()
}

// Ping 健康检查
func (c *RedisCache) Ping() error {
	return c.client.Ping(c.ctx).Err()
}

// Close 关闭连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// FlushDB 清空当前数据库（危险操作）
func (c *RedisCache) FlushDB() error {
	return c.client.FlushDB(c.ctx).Err()
}

// Info 获取 Redis 服务器信息
func (c *RedisCache) Info() (string, error) {
	return c.client.Info(c.ctx).Result()
}

// DBSize 获取当前数据库的键数量
func (c *RedisCache) DBSize() (int64, error) {
	return c.client.DBSize(c.ctx).Result()
}
