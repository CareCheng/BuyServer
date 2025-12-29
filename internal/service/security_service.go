package service

import (
	"sync"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// SecurityService 安全服务
type SecurityService struct {
	repo *repository.Repository
}

// API限流记录（内存缓存，不需要持久化）
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

	// 记录失败（持久化到数据库）
	s.recordFailure(username)
	s.recordFailure(ip)
}

func (s *SecurityService) recordFailure(key string) {
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
	if s.repo != nil {
		s.repo.DeleteLoginFailureRecord(key)
	}
}

// GetLoginFailureCount 获取登录失败次数
func (s *SecurityService) GetLoginFailureCount(key string) int {
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

// CheckRateLimit 检查API限流（内存缓存，不需要持久化）
func (s *SecurityService) CheckRateLimit(key string) (bool, int) {
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

	// 清理内存中的限流记录
	rateLimitsMu.Lock()
	for key, info := range rateLimits {
		if time.Since(info.WindowStart) > RateLimitWindow*2 {
			delete(rateLimits, key)
		}
	}
	rateLimitsMu.Unlock()
}
