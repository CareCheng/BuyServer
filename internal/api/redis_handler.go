package api

import (
	"strings"

	"user-frontend/internal/cache"
	"user-frontend/internal/config"

	"github.com/gin-gonic/gin"
)

// ==================== Redis 配置管理 API ====================

// SaveRedisConfigRequest 保存Redis配置请求
type SaveRedisConfigRequest struct {
	// 基础配置
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode" binding:"omitempty,oneof=standalone sentinel cluster"` // standalone/sentinel/cluster
	Prefix  string `json:"prefix"`

	// 单机模式配置
	Host string `json:"host"`
	Port int    `json:"port"`

	// Sentinel模式配置
	SentinelAddrs  string `json:"sentinel_addrs"`
	SentinelMaster string `json:"sentinel_master"`

	// Cluster模式配置
	ClusterAddrs string `json:"cluster_addrs"`

	// 通用配置
	Password string `json:"password"`
	DB       int    `json:"db"`

	// 连接池配置
	PoolSize     int `json:"pool_size"`
	MinIdleConns int `json:"min_idle_conns"`
	MaxRetries   int `json:"max_retries"`

	// 超时配置(秒)
	DialTimeout  int `json:"dial_timeout"`
	ReadTimeout  int `json:"read_timeout"`
	WriteTimeout int `json:"write_timeout"`

	// TLS配置
	TLSEnabled  bool   `json:"tls_enabled"`
	TLSCertPath string `json:"tls_cert_path"`
	TLSKeyPath  string `json:"tls_key_path"`
	TLSCAPath   string `json:"tls_ca_path"`
}

// AdminGetRedisConfig 获取Redis配置
// @Summary 获取Redis配置
// @Description 获取当前的Redis配置信息
// @Tags 管理员-Redis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/admin/redis/config [get]
func AdminGetRedisConfig(c *gin.Context) {
	if DBConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "配置服务未初始化"})
		return
	}

	config, err := DBConfigSvc.GetRedisConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取Redis配置失败: " + err.Error()})
		return
	}

	// 隐藏敏感信息
	if config.Password != "" {
		config.Password = "******"
	}
	if config.SentinelPassword != "" {
		config.SentinelPassword = "******"
	}
	if config.TLSKey != "" {
		config.TLSKey = "[已配置]"
	}

	c.JSON(200, gin.H{"success": true, "config": config})
}

// AdminSaveRedisConfig 保存Redis配置
// @Summary 保存Redis配置
// @Description 保存Redis配置信息，保存后需要手动测试连接
// @Tags 管理员-Redis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SaveRedisConfigRequest true "Redis配置"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/admin/redis/config [post]
func AdminSaveRedisConfig(c *gin.Context) {
	if DBConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "配置服务未初始化"})
		return
	}

	var req SaveRedisConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误: " + err.Error()})
		return
	}

	// 验证模式相关配置
	switch req.Mode {
	case "standalone", "":
		if req.Host == "" {
			req.Host = "127.0.0.1"
		}
		if req.Port == 0 {
			req.Port = 6379
		}
		req.Mode = "standalone"
	case "sentinel":
		if req.SentinelAddrs == "" {
			c.JSON(400, gin.H{"success": false, "error": "Sentinel模式需要配置sentinel_addrs"})
			return
		}
		if req.SentinelMaster == "" {
			c.JSON(400, gin.H{"success": false, "error": "Sentinel模式需要配置sentinel_master"})
			return
		}
	case "cluster":
		if req.ClusterAddrs == "" {
			c.JSON(400, gin.H{"success": false, "error": "Cluster模式需要配置cluster_addrs"})
			return
		}
	}

	// 设置默认值
	if req.Prefix == "" {
		req.Prefix = "user:"
	}
	if req.PoolSize == 0 {
		req.PoolSize = 10
	}
	if req.MinIdleConns == 0 {
		req.MinIdleConns = 5
	}
	if req.MaxRetries == 0 {
		req.MaxRetries = 3
	}
	if req.DialTimeout == 0 {
		req.DialTimeout = 5
	}
	if req.ReadTimeout == 0 {
		req.ReadTimeout = 3
	}
	if req.WriteTimeout == 0 {
		req.WriteTimeout = 3
	}

	// 构建配置对象（使用运行时配置类型）
	redisConfig := &config.RedisConfig{
		Enabled:       req.Enabled,
		Mode:          req.Mode,
		KeyPrefix:     req.Prefix,
		Host:          req.Host,
		Port:          req.Port,
		SentinelAddrs: parseSentinelAddrs(req.SentinelAddrs),
		SentinelMaster: req.SentinelMaster,
		ClusterAddrs:  parseClusterAddrs(req.ClusterAddrs),
		Password:      req.Password,
		Database:      req.DB,
		PoolSize:      req.PoolSize,
		MinIdleConns:  req.MinIdleConns,
		MaxRetries:    req.MaxRetries,
		DialTimeout:   req.DialTimeout,
		ReadTimeout:   req.ReadTimeout,
		WriteTimeout:  req.WriteTimeout,
		TLSEnabled:    req.TLSEnabled,
		TLSCert:       req.TLSCertPath,
		TLSKey:        req.TLSKeyPath,
		TLSCACert:     req.TLSCAPath,
	}

	if err := DBConfigSvc.SaveRedisConfig(redisConfig); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存Redis配置失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Redis配置保存成功，请测试连接"})
}

// parseSentinelAddrs 解析哨兵地址字符串
func parseSentinelAddrs(addrs string) []string {
	if addrs == "" {
		return nil
	}
	parts := strings.Split(addrs, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// parseClusterAddrs 解析集群地址字符串
func parseClusterAddrs(addrs string) []string {
	return parseSentinelAddrs(addrs)
}

// convertToCacheConfig 将配置转换为缓存配置
func convertToCacheConfig(cfg *config.RedisConfig) *cache.RedisConfig {
	if cfg == nil {
		return nil
	}
	return &cache.RedisConfig{
		Enabled:          cfg.Enabled,
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
		KeyPrefix:        cfg.KeyPrefix,
		DefaultTTL:       cfg.DefaultTTL,
		EnableMetrics:    cfg.EnableMetrics,
		TLSEnabled:       cfg.TLSEnabled,
		TLSCert:          cfg.TLSCert,
		TLSKey:           cfg.TLSKey,
		TLSCACert:        cfg.TLSCACert,
	}
}

// AdminTestRedisConnection 测试Redis连接
// @Summary 测试Redis连接
// @Description 测试当前Redis配置的连接是否正常
// @Tags 管理员-Redis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/admin/redis/test [post]
func AdminTestRedisConnection(c *gin.Context) {
	if DBConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "配置服务未初始化"})
		return
	}

	// 获取当前配置
	cfg, err := DBConfigSvc.GetRedisConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取Redis配置失败: " + err.Error()})
		return
	}

	// 测试连接
	success, message, latency := DBConfigSvc.TestRedisConnection(cfg)
	
	c.JSON(200, gin.H{
		"success": success,
		"result": gin.H{
			"connected": success,
			"message":   message,
			"latency":   latency,
		},
	})
}

// AdminGetCacheStats 获取缓存统计信息
// @Summary 获取缓存统计
// @Description 获取缓存的使用统计信息
// @Tags 管理员-Redis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/admin/redis/stats [get]
func AdminGetCacheStats(c *gin.Context) {
	manager := cache.GetCacheManager()
	if manager == nil {
		c.JSON(500, gin.H{"success": false, "error": "缓存管理器未初始化"})
		return
	}

	stats := manager.GetStats()
	c.JSON(200, gin.H{"success": true, "stats": stats})
}

// CacheDashboardStats 仪表盘统计数据
type CacheDashboardStats struct {
	// 基本信息
	Mode           string `json:"mode"`             // 缓存模式: local/redis-standalone/redis-sentinel/redis-cluster
	Status         string `json:"status"`           // 状态: connected/disconnected/degraded
	Version        string `json:"version"`          // Redis版本（如果是Redis模式）
	Uptime         string `json:"uptime"`           // 运行时间
	UptimeSeconds  int64  `json:"uptime_seconds"`   // 运行时间（秒）
	
	// 性能指标
	HitRate        float64 `json:"hit_rate"`        // 命中率（百分比）
	Hits           int64   `json:"hits"`            // 命中次数
	Misses         int64   `json:"misses"`          // 未命中次数
	TotalRequests  int64   `json:"total_requests"`  // 总请求次数
	OpsPerSecond   float64 `json:"ops_per_second"`  // 每秒操作数
	
	// 内存信息
	MemoryUsed     string  `json:"memory_used"`       // 已用内存
	MemoryUsedBytes int64  `json:"memory_used_bytes"` // 已用内存（字节）
	MemoryPeak     string  `json:"memory_peak"`       // 峰值内存
	MemoryPeakBytes int64  `json:"memory_peak_bytes"` // 峰值内存（字节）
	MemoryLimit    string  `json:"memory_limit"`      // 内存限制
	MemoryPolicy   string  `json:"memory_policy"`     // 内存淘汰策略
	
	// 键空间信息
	KeysCount      int64  `json:"keys_count"`       // 总键数
	ExpiringKeys   int64  `json:"expiring_keys"`    // 设置过期时间的键数
	ExpiredKeys    int64  `json:"expired_keys"`     // 已过期删除的键数
	EvictedKeys    int64  `json:"evicted_keys"`     // 被淘汰的键数
	
	// 连接信息
	ConnectedClients  int `json:"connected_clients"`   // 当前连接的客户端数
	MaxClients        int `json:"max_clients"`         // 最大客户端数
	BlockedClients    int `json:"blocked_clients"`     // 阻塞的客户端数
	
	// 复制信息（仅主从模式）
	Role              string `json:"role"`               // 角色: master/slave
	ConnectedSlaves   int    `json:"connected_slaves"`   // 连接的从节点数
	
	// 持久化信息
	RDBEnabled       bool   `json:"rdb_enabled"`         // RDB是否启用
	AOFEnabled       bool   `json:"aof_enabled"`         // AOF是否启用
	LastSaveTime     string `json:"last_save_time"`      // 最后保存时间
	LastSaveStatus   string `json:"last_save_status"`    // 最后保存状态
	
	// 故障转移信息
	Failovers        int64  `json:"failovers"`           // 故障转移次数
	LastError        string `json:"last_error"`          // 最后一次错误
	LastErrorTime    string `json:"last_error_time"`     // 最后错误时间
	
	// 本地缓存信息（降级模式时使用）
	LocalCacheSize   int    `json:"local_cache_size"`    // 本地缓存条目数
	LocalCacheMemory string `json:"local_cache_memory"`  // 本地缓存估算内存
}

// AdminGetCacheDashboard 获取缓存仪表盘数据
// @Summary 获取缓存仪表盘
// @Description 获取详细的缓存仪表盘统计数据
// @Tags 管理员-Redis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/admin/redis/dashboard [get]
func AdminGetCacheDashboard(c *gin.Context) {
	manager := cache.GetCacheManager()
	if manager == nil {
		c.JSON(500, gin.H{"success": false, "error": "缓存管理器未初始化"})
		return
	}

	dashboard := manager.GetDashboard()
	c.JSON(200, gin.H{"success": true, "dashboard": dashboard})
}

// AdminGetCacheKeys 获取缓存键列表
// @Summary 获取缓存键列表
// @Description 获取匹配模式的缓存键列表
// @Tags 管理员-Redis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param pattern query string false "匹配模式，默认*"
// @Param limit query int false "返回数量限制，默认100"
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/admin/redis/keys [get]
func AdminGetCacheKeys(c *gin.Context) {
	manager := cache.GetCacheManager()
	if manager == nil {
		c.JSON(500, gin.H{"success": false, "error": "缓存管理器未初始化"})
		return
	}

	pattern := c.DefaultQuery("pattern", "*")
	// 安全限制：最多返回100个键
	keys, err := manager.Keys(pattern)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取键列表失败: " + err.Error()})
		return
	}

	// 限制返回数量
	limit := 100
	if len(keys) > limit {
		keys = keys[:limit]
	}

	c.JSON(200, gin.H{
		"success": true,
		"keys":    keys,
		"count":   len(keys),
		"pattern": pattern,
	})
}

// AdminDeleteCacheKey 删除指定缓存键
// @Summary 删除缓存键
// @Description 删除指定的缓存键
// @Tags 管理员-Redis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key query string true "要删除的键名"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/admin/redis/key [delete]
func AdminDeleteCacheKey(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(400, gin.H{"success": false, "error": "请指定要删除的键名"})
		return
	}

	manager := cache.GetCacheManager()
	if manager == nil {
		c.JSON(500, gin.H{"success": false, "error": "缓存管理器未初始化"})
		return
	}

	if err := manager.Delete(key); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除键失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "键已删除"})
}

// AdminGetCacheKeyInfo 获取键信息
// @Summary 获取缓存键信息
// @Description 获取指定缓存键的详细信息
// @Tags 管理员-Redis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key query string true "键名"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/admin/redis/key/info [get]
func AdminGetCacheKeyInfo(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(400, gin.H{"success": false, "error": "请指定键名"})
		return
	}

	manager := cache.GetCacheManager()
	if manager == nil {
		c.JSON(500, gin.H{"success": false, "error": "缓存管理器未初始化"})
		return
	}

	// 检查键是否存在
	if !manager.Exists(key) {
		c.JSON(404, gin.H{"success": false, "error": "键不存在"})
		return
	}

	// 获取TTL
	ttl, _ := manager.TTL(key)

	// 获取值
	value, _ := manager.Get(key)

	c.JSON(200, gin.H{
		"success": true,
		"key":     key,
		"exists":  true,
		"ttl":     ttl.String(),
		"value":   value,
	})
}

// AdminFlushCache 清空缓存
// @Summary 清空缓存
// @Description 清空所有缓存数据（危险操作）
// @Tags 管理员-Redis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param confirm query bool true "确认清空"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/admin/redis/flush [post]
func AdminFlushCache(c *gin.Context) {
	confirm := c.Query("confirm")
	if confirm != "true" {
		c.JSON(400, gin.H{"success": false, "error": "请确认清空操作，传入confirm=true"})
		return
	}

	manager := cache.GetCacheManager()
	if manager == nil {
		c.JSON(500, gin.H{"success": false, "error": "缓存管理器未初始化"})
		return
	}

	if err := manager.FlushAll(); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "清空缓存失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "缓存已清空"})
}

// AdminRefreshCache 刷新缓存连接
// @Summary 刷新缓存连接
// @Description 使用最新配置重新初始化缓存连接
// @Tags 管理员-Redis
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/admin/redis/refresh [post]
func AdminRefreshCache(c *gin.Context) {
	if DBConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "配置服务未初始化"})
		return
	}

	// 获取最新配置
	redisConfig, err := DBConfigSvc.GetRedisConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取Redis配置失败: " + err.Error()})
		return
	}

	if !redisConfig.Enabled {
		// 如果禁用Redis，切换到本地缓存
		cache.InitCacheManager(nil)
		c.JSON(200, gin.H{"success": true, "message": "已切换到本地缓存模式"})
		return
	}

	// 转换为缓存配置
	cacheConfig := convertToCacheConfig(redisConfig)

	// 重新初始化缓存管理器
	cache.InitCacheManager(cacheConfig)

	// 测试新连接
	manager := cache.GetCacheManager()
	if manager == nil {
		c.JSON(500, gin.H{"success": false, "error": "缓存管理器初始化失败"})
		return
	}

	if err := manager.Ping(); err != nil {
		c.JSON(200, gin.H{
			"success": true,
			"message": "缓存已刷新，但Redis连接失败，已降级到本地缓存",
			"warning": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "缓存连接刷新成功"})
}
