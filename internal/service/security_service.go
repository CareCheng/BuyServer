package service

import (
	"fmt"
	"sync"
	"time"

	"user-frontend/internal/cache"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// SecurityService 安全服务
type SecurityService struct {
	repo *repository.Repository
}

// API限流记录（内存缓存后备，仅当CacheManager不可用时使用）
var (
	rateLimits   = make(map[string]*RateLimitInfo)
	rateLimitsMu sync.RWMutex
)

// RateLimitInfo 限流信息
type RateLimitInfo struct {
	Count       int
	WindowStart time.Time
}

// 配置常量
const (
	MaxLoginAttempts     = 5                // 最大登录尝试次数
	LoginLockDuration    = 15 * time.Minute // 锁定时长
	LoginFailureWindow   = 10 * time.Minute // 失败计数窗口
	RateLimitWindow      = time.Minute      // 限流窗口
	RateLimitMaxRequests = 60               // 每分钟最大请求数
)

func NewSecurityService(repo *repository.Repository) *SecurityService {
	return &SecurityService{repo: repo}
}

// RecordLoginAttempt 记录登录尝试
func (s *SecurityService) RecordLoginAttempt(username, ip string, success bool) {
	if s.repo != nil {
		attempt := &model.LoginAttempt{
			Username:  username,
			IP:        ip,
			Success:   success,
			CreatedAt: time.Now(),
		}
		s.repo.CreateLoginAttempt(attempt)
	}

	if success {
		// 登录成功，清除失败记录
		s.ClearLoginFailures(username)
		s.ClearLoginFailures(ip)
		return
	}

	// 记录失败（使用缓存 + 数据库双写）
	s.recordFailure(username)
	s.recordFailure(ip)
}

func (s *SecurityService) recordFailure(key string) {
	cm := cache.GetCacheManager()
	
	// 使用缓存记录失败次数
	if cm != nil {
		failKey := cache.LoginFailureKey(key)
		count, err := cm.IncrBy(failKey, 1)
		if err == nil {
			// 设置过期时间（首次设置）
			if count == 1 {
				cm.Expire(failKey, LoginFailureWindow)
			}
			
			// 检查是否需要锁定
			if count >= int64(MaxLoginAttempts) {
				lockKey := cache.LoginLockKey(key)
				cm.Set(lockKey, "locked", LoginLockDuration)
			}
			
			// 如果失败次数过多，加入临时黑名单
			if count >= int64(MaxLoginAttempts*2) {
				s.AddToTempBlacklist(key, 30*time.Minute)
			}
			return
		}
	}

	// 降级到数据库
	s.recordFailureToDatabase(key)
}

// recordFailureToDatabase 将失败记录写入数据库（降级方案）
func (s *SecurityService) recordFailureToDatabase(key string) {
	if s.repo == nil {
		return
	}

	now := time.Now()
	record, err := s.repo.GetLoginFailureRecord(key)
	if err != nil {
		// 不存在，创建新记录
		record = &model.LoginFailureRecord{
			Key:          key,
			FailureCount: 1,
			FirstFailAt:  now,
		}
		s.repo.SaveLoginFailureRecord(record)
		return
	}

	// 检查是否在窗口期内
	if now.Sub(record.FirstFailAt) > LoginFailureWindow {
		// 超出窗口期，重置计数
		record.FailureCount = 1
		record.FirstFailAt = now
		record.LockedAt = nil
	} else {
		record.FailureCount++
		if record.FailureCount >= MaxLoginAttempts {
			record.LockedAt = &now
		}
		// 如果失败次数过多（超过10次），加入临时黑名单
		if record.FailureCount >= MaxLoginAttempts*2 {
			s.AddToTempBlacklist(key, 30*time.Minute)
		}
	}
	s.repo.SaveLoginFailureRecord(record)
}

// AddToTempBlacklist 添加到临时黑名单（通过API中间件实现）
func (s *SecurityService) AddToTempBlacklist(key string, duration time.Duration) {
	// 这里可以调用API层的黑名单功能
	// 由于循环依赖，使用回调方式
	if blacklistCallback != nil {
		blacklistCallback(key, duration)
	}
}

// 黑名单回调函数
var blacklistCallback func(key string, duration time.Duration)

// SetBlacklistCallback 设置黑名单回调
func SetBlacklistCallback(callback func(key string, duration time.Duration)) {
	blacklistCallback = callback
}

// IsLoginLocked 检查是否被锁定
func (s *SecurityService) IsLoginLocked(key string) (bool, time.Duration) {
	cm := cache.GetCacheManager()
	
	// 优先从缓存检查
	if cm != nil {
		lockKey := cache.LoginLockKey(key)
		if cm.Exists(lockKey) {
			ttl, err := cm.TTL(lockKey)
			if err == nil && ttl > 0 {
				return true, ttl
			}
		}
		// 缓存中没有锁定记录
		return false, 0
	}

	// 降级到数据库
	return s.isLoginLockedFromDatabase(key)
}

// isLoginLockedFromDatabase 从数据库检查锁定状态
func (s *SecurityService) isLoginLockedFromDatabase(key string) (bool, time.Duration) {
	if s.repo == nil {
		return false, 0
	}

	record, err := s.repo.GetLoginFailureRecord(key)
	if err != nil {
		return false, 0
	}

	if record.LockedAt == nil {
		return false, 0
	}

	elapsed := time.Since(*record.LockedAt)
	if elapsed >= LoginLockDuration {
		// 锁定已过期，清除记录
		s.repo.DeleteLoginFailureRecord(key)
		return false, 0
	}

	return true, LoginLockDuration - elapsed
}

// ClearLoginFailures 清除登录失败记录
func (s *SecurityService) ClearLoginFailures(key string) {
	cm := cache.GetCacheManager()
	
	// 清除缓存中的记录
	if cm != nil {
		failKey := cache.LoginFailureKey(key)
		lockKey := cache.LoginLockKey(key)
		cm.Delete(failKey)
		cm.Delete(lockKey)
	}

	// 同时清除数据库中的记录
	if s.repo != nil {
		s.repo.DeleteLoginFailureRecord(key)
	}
}

// GetLoginFailureCount 获取登录失败次数
func (s *SecurityService) GetLoginFailureCount(key string) int {
	cm := cache.GetCacheManager()
	
	// 优先从缓存获取
	if cm != nil {
		failKey := cache.LoginFailureKey(key)
		if val, ok := cm.Get(failKey); ok {
			if count, ok := val.(int64); ok {
				return int(count)
			}
			if count, ok := val.(int); ok {
				return count
			}
		}
	}

	// 降级到数据库
	return s.getLoginFailureCountFromDatabase(key)
}

// getLoginFailureCountFromDatabase 从数据库获取失败次数
func (s *SecurityService) getLoginFailureCountFromDatabase(key string) int {
	if s.repo == nil {
		return 0
	}

	record, err := s.repo.GetLoginFailureRecord(key)
	if err != nil {
		return 0
	}

	if time.Since(record.FirstFailAt) > LoginFailureWindow {
		return 0
	}

	return record.FailureCount
}

// CheckRateLimit 检查API限流
func (s *SecurityService) CheckRateLimit(key string) (bool, int) {
	cm := cache.GetCacheManager()
	
	// 优先使用缓存进行限流
	if cm != nil {
		return s.checkRateLimitWithCache(key)
	}

	// 降级到内存限流
	return s.checkRateLimitInMemory(key)
}

// checkRateLimitWithCache 使用缓存进行限流
func (s *SecurityService) checkRateLimitWithCache(key string) (bool, int) {
	cm := cache.GetCacheManager()
	if cm == nil {
		return s.checkRateLimitInMemory(key)
	}

	// 计算当前窗口ID（基于分钟）
	windowID := time.Now().Unix() / 60
	rateLimitKey := cache.RateLimitKey("api", key, windowID)
	
	// 使用原子递增
	count, err := cm.IncrBy(rateLimitKey, 1)
	if err != nil {
		// 缓存失败，降级到内存
		return s.checkRateLimitInMemory(key)
	}

	// 首次设置过期时间
	if count == 1 {
		cm.Expire(rateLimitKey, RateLimitWindow)
	}

	if count > int64(RateLimitMaxRequests) {
		return false, 0
	}

	return true, RateLimitMaxRequests - int(count)
}

// checkRateLimitInMemory 内存限流（降级方案）
func (s *SecurityService) checkRateLimitInMemory(key string) (bool, int) {
	rateLimitsMu.Lock()
	defer rateLimitsMu.Unlock()

	info, exists := rateLimits[key]
	now := time.Now()

	if !exists || now.Sub(info.WindowStart) > RateLimitWindow {
		rateLimits[key] = &RateLimitInfo{
			Count:       1,
			WindowStart: now,
		}
		return true, RateLimitMaxRequests - 1
	}

	if info.Count >= RateLimitMaxRequests {
		return false, 0
	}

	info.Count++
	return true, RateLimitMaxRequests - info.Count
}

// CleanupExpiredRecords 清理过期记录（定期调用）
func (s *SecurityService) CleanupExpiredRecords() {
	// 清理数据库中过期的登录失败记录
	if s.repo != nil {
		s.repo.DeleteExpiredLoginFailureRecords(LoginFailureWindow)
	}

	// 清理内存中的限流记录（当使用缓存时，缓存会自动过期）
	rateLimitsMu.Lock()
	for key, info := range rateLimits {
		if time.Since(info.WindowStart) > RateLimitWindow*2 {
			delete(rateLimits, key)
		}
	}
	rateLimitsMu.Unlock()
}

// GetSecurityStats 获取安全统计信息（用于监控）
func (s *SecurityService) GetSecurityStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	// 内存限流记录数
	rateLimitsMu.RLock()
	stats["memory_rate_limit_entries"] = len(rateLimits)
	rateLimitsMu.RUnlock()

	// 检查缓存是否可用
	cm := cache.GetCacheManager()
	if cm != nil {
		stats["cache_enabled"] = true
		// 尝试获取一些缓存统计
		cacheStats := cm.GetStats()
		stats["cache_stats"] = cacheStats
	} else {
		stats["cache_enabled"] = false
	}

	// 数据库登录失败记录数（如果可用）
	if s.repo != nil {
		count, err := s.repo.CountActiveLoginFailures()
		if err == nil {
			stats["db_active_failures"] = count
		}
	}

	return stats
}

// SetRateLimitCustom 设置自定义限流（用于特定场景）
func (s *SecurityService) SetRateLimitCustom(key string, maxRequests int, window time.Duration) (bool, int) {
	cm := cache.GetCacheManager()
	if cm == nil {
		return true, maxRequests // 降级时不限流
	}

	rateLimitKey := fmt.Sprintf("rl:custom:%s", key)
	
	count, err := cm.IncrBy(rateLimitKey, 1)
	if err != nil {
		return true, maxRequests
	}

	if count == 1 {
		cm.Expire(rateLimitKey, window)
	}

	if count > int64(maxRequests) {
		return false, 0
	}

	return true, maxRequests - int(count)
}
