// Package cache 提供统一的缓存抽象层
// cache_metrics.go - 缓存监控指标
package cache

import (
	"sync/atomic"
	"time"
)

// CacheMetrics 缓存监控指标
type CacheMetrics struct {
	// 命中统计
	hits   atomic.Int64
	misses atomic.Int64

	// 故障转移统计
	failovers atomic.Int64

	// 启动时间
	startTime time.Time

	// 最后一次错误
	lastError     atomic.Value // string
	lastErrorTime atomic.Value // time.Time
}

// NewCacheMetrics 创建缓存监控指标
func NewCacheMetrics() *CacheMetrics {
	m := &CacheMetrics{
		startTime: time.Now(),
	}
	m.lastError.Store("")
	m.lastErrorTime.Store(time.Time{})
	return m
}

// RecordHit 记录缓存命中
func (m *CacheMetrics) RecordHit() {
	m.hits.Add(1)
}

// RecordMiss 记录缓存未命中
func (m *CacheMetrics) RecordMiss() {
	m.misses.Add(1)
}

// RecordFailover 记录故障转移
func (m *CacheMetrics) RecordFailover() {
	m.failovers.Add(1)
}

// RecordError 记录错误
func (m *CacheMetrics) RecordError(err string) {
	m.lastError.Store(err)
	m.lastErrorTime.Store(time.Now())
}

// GetHits 获取命中次数
func (m *CacheMetrics) GetHits() int64 {
	return m.hits.Load()
}

// GetMisses 获取未命中次数
func (m *CacheMetrics) GetMisses() int64 {
	return m.misses.Load()
}

// GetHitRate 获取命中率
func (m *CacheMetrics) GetHitRate() float64 {
	hits := m.hits.Load()
	misses := m.misses.Load()
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// GetFailovers 获取故障转移次数
func (m *CacheMetrics) GetFailovers() int64 {
	return m.failovers.Load()
}

// GetUptime 获取运行时间
func (m *CacheMetrics) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

// GetLastError 获取最后一次错误
func (m *CacheMetrics) GetLastError() string {
	if v := m.lastError.Load(); v != nil {
		return v.(string)
	}
	return ""
}

// GetLastErrorTime 获取最后一次错误时间
func (m *CacheMetrics) GetLastErrorTime() time.Time {
	if v := m.lastErrorTime.Load(); v != nil {
		return v.(time.Time)
	}
	return time.Time{}
}

// CacheStats 缓存统计信息（用于 API 返回）
type CacheStats struct {
	RedisEnabled   bool    `json:"redis_enabled"`
	RedisHealthy   bool    `json:"redis_healthy"`
	LocalCacheSize int     `json:"local_cache_size"`
	Hits           int64   `json:"hits"`
	Misses         int64   `json:"misses"`
	Failovers      int64   `json:"failovers"`
	HitRate        string  `json:"hit_rate"`
	Uptime         string  `json:"uptime"`
	LastError      string  `json:"last_error,omitempty"`
}

// CacheDashboard 缓存仪表盘数据（详细统计）
type CacheDashboard struct {
	// 基本信息
	Mode           string `json:"mode"`             // 缓存模式: local/redis-standalone/redis-sentinel/redis-cluster
	Status         string `json:"status"`           // 状态: connected/disconnected/degraded
	Version        string `json:"version"`          // Redis版本（如果是Redis模式）
	Uptime         string `json:"uptime"`           // 运行时间
	UptimeSeconds  int64  `json:"uptime_seconds"`   // 运行时间（秒）
	
	// 性能指标
	HitRate        float64 `json:"hit_rate"`        // 命中率（百分比）
	HitRateStr     string  `json:"hit_rate_str"`    // 命中率（字符串）
	Hits           int64   `json:"hits"`            // 命中次数
	Misses         int64   `json:"misses"`          // 未命中次数
	TotalRequests  int64   `json:"total_requests"`  // 总请求次数
	OpsPerSecond   float64 `json:"ops_per_second"`  // 每秒操作数
	
	// 内存信息
	MemoryUsed      string  `json:"memory_used"`       // 已用内存
	MemoryUsedBytes int64   `json:"memory_used_bytes"` // 已用内存（字节）
	MemoryPeak      string  `json:"memory_peak"`       // 峰值内存
	MemoryPeakBytes int64   `json:"memory_peak_bytes"` // 峰值内存（字节）
	MemoryLimit     string  `json:"memory_limit"`      // 内存限制
	MemoryPolicy    string  `json:"memory_policy"`     // 内存淘汰策略
	
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

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats() *CacheStats {
	hitRate := cm.metrics.GetHitRate()
	return &CacheStats{
		RedisEnabled:   cm.redisEnabled,
		RedisHealthy:   cm.redisHealthy.Load(),
		LocalCacheSize: cm.local.Size(),
		Hits:           cm.metrics.GetHits(),
		Misses:         cm.metrics.GetMisses(),
		Failovers:      cm.metrics.GetFailovers(),
		HitRate:        formatPercent(hitRate),
		Uptime:         formatDuration(cm.metrics.GetUptime()),
		LastError:      cm.metrics.GetLastError(),
	}
}

// GetDashboard 获取缓存仪表盘数据
func (cm *CacheManager) GetDashboard() *CacheDashboard {
	dashboard := &CacheDashboard{
		Mode:            "local",
		Status:          "connected",
		Uptime:          formatDuration(cm.metrics.GetUptime()),
		UptimeSeconds:   int64(cm.metrics.GetUptime().Seconds()),
		HitRate:         cm.metrics.GetHitRate(),
		HitRateStr:      formatPercent(cm.metrics.GetHitRate()),
		Hits:            cm.metrics.GetHits(),
		Misses:          cm.metrics.GetMisses(),
		TotalRequests:   cm.metrics.GetHits() + cm.metrics.GetMisses(),
		Failovers:       cm.metrics.GetFailovers(),
		LastError:       cm.metrics.GetLastError(),
		LocalCacheSize:  cm.local.Size(),
		LocalCacheMemory: formatBytes(int64(cm.local.Size() * 256)), // 估算每个条目256字节
	}
	
	// 计算每秒操作数
	uptimeSeconds := cm.metrics.GetUptime().Seconds()
	if uptimeSeconds > 0 {
		dashboard.OpsPerSecond = float64(dashboard.TotalRequests) / uptimeSeconds
	}
	
	// 最后错误时间
	if lastErrTime := cm.metrics.GetLastErrorTime(); !lastErrTime.IsZero() {
		dashboard.LastErrorTime = lastErrTime.Format("2006-01-02 15:04:05")
	}
	
	// Redis 模式
	if cm.redisEnabled {
		if cm.config != nil {
			switch cm.config.Mode {
			case "sentinel":
				dashboard.Mode = "redis-sentinel"
			case "cluster":
				dashboard.Mode = "redis-cluster"
			default:
				dashboard.Mode = "redis-standalone"
			}
		}
		
		if cm.redisHealthy.Load() {
			dashboard.Status = "connected"
			
			// 获取 Redis 详细信息
			if cm.redis != nil {
				info, err := cm.redis.Info()
				if err == nil {
					parseRedisInfo(info, dashboard)
				}
				
				// 获取键数量
				if dbSize, err := cm.redis.DBSize(); err == nil {
					dashboard.KeysCount = dbSize
				}
			}
		} else {
			dashboard.Status = "degraded"
		}
	}
	
	return dashboard
}

// parseRedisInfo 解析 Redis INFO 命令的输出
func parseRedisInfo(info string, dashboard *CacheDashboard) {
	lines := splitLines(info)
	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		parts := splitKV(line)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		
		switch key {
		case "redis_version":
			dashboard.Version = value
		case "used_memory":
			if n := parseInt64(value); n > 0 {
				dashboard.MemoryUsedBytes = n
				dashboard.MemoryUsed = formatBytes(n)
			}
		case "used_memory_peak":
			if n := parseInt64(value); n > 0 {
				dashboard.MemoryPeakBytes = n
				dashboard.MemoryPeak = formatBytes(n)
			}
		case "maxmemory":
			if n := parseInt64(value); n > 0 {
				dashboard.MemoryLimit = formatBytes(n)
			}
		case "maxmemory_policy":
			dashboard.MemoryPolicy = value
		case "connected_clients":
			dashboard.ConnectedClients = int(parseInt64(value))
		case "maxclients":
			dashboard.MaxClients = int(parseInt64(value))
		case "blocked_clients":
			dashboard.BlockedClients = int(parseInt64(value))
		case "role":
			dashboard.Role = value
		case "connected_slaves":
			dashboard.ConnectedSlaves = int(parseInt64(value))
		case "rdb_last_save_time":
			if n := parseInt64(value); n > 0 {
				dashboard.RDBEnabled = true
				dashboard.LastSaveTime = formatUnixTime(n)
			}
		case "rdb_last_bgsave_status":
			dashboard.LastSaveStatus = value
		case "aof_enabled":
			dashboard.AOFEnabled = value == "1"
		case "expired_keys":
			dashboard.ExpiredKeys = parseInt64(value)
		case "evicted_keys":
			dashboard.EvictedKeys = parseInt64(value)
		case "keyspace_hits":
			// Redis 自身的命中统计
		case "keyspace_misses":
			// Redis 自身的未命中统计
		}
	}
}

// splitLines 按行分割字符串
func splitLines(s string) []string {
	var lines []string
	var current []byte
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			if len(current) > 0 && current[len(current)-1] == '\r' {
				current = current[:len(current)-1]
			}
			lines = append(lines, string(current))
			current = nil
		} else {
			current = append(current, s[i])
		}
	}
	if len(current) > 0 {
		lines = append(lines, string(current))
	}
	return lines
}

// splitKV 按 : 分割键值对
func splitKV(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}

// parseInt64 解析整数
func parseInt64(s string) int64 {
	var n int64
	negative := false
	for i := 0; i < len(s); i++ {
		if s[i] == '-' && i == 0 {
			negative = true
			continue
		}
		if s[i] >= '0' && s[i] <= '9' {
			n = n*10 + int64(s[i]-'0')
		}
	}
	if negative {
		return -n
	}
	return n
}

// formatUnixTime 格式化 Unix 时间戳
func formatUnixTime(ts int64) string {
	t := time.Unix(ts, 0)
	return t.Format("2006-01-02 15:04:05")
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	
	if bytes < KB {
		return itoa(bytes) + " B"
	}
	if bytes < MB {
		return formatFloat(float64(bytes)/float64(KB), 2) + " KB"
	}
	if bytes < GB {
		return formatFloat(float64(bytes)/float64(MB), 2) + " MB"
	}
	return formatFloat(float64(bytes)/float64(GB), 2) + " GB"
}

// formatPercent 格式化百分比
func formatPercent(value float64) string {
	return formatFloat(value, 2) + "%"
}

// formatFloat 格式化浮点数
func formatFloat(value float64, precision int) string {
	format := "%." + string(rune('0'+precision)) + "f"
	return sprintf(format, value)
}

// sprintf 简单的格式化函数
func sprintf(format string, value float64) string {
	// 使用简单的整数和小数分离方式
	intPart := int64(value)
	decPart := int64((value - float64(intPart)) * 100)
	if decPart < 0 {
		decPart = -decPart
	}
	if decPart < 10 {
		return itoa(intPart) + ".0" + itoa(decPart)
	}
	return itoa(intPart) + "." + itoa(decPart)
}

// itoa 整数转字符串
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return itoa(int64(d.Seconds())) + "s"
	}
	if d < time.Hour {
		return itoa(int64(d.Minutes())) + "m " + itoa(int64(d.Seconds())%60) + "s"
	}
	if d < 24*time.Hour {
		return itoa(int64(d.Hours())) + "h " + itoa(int64(d.Minutes())%60) + "m"
	}
	days := int64(d.Hours()) / 24
	hours := int64(d.Hours()) % 24
	return itoa(days) + "d " + itoa(hours) + "h"
}
