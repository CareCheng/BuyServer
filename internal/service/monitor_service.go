package service

import (
	"runtime"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// MonitorService 系统监控服务
type MonitorService struct {
	repo      *repository.Repository
	startTime time.Time
}

// NewMonitorService 创建系统监控服务实例
func NewMonitorService(repo *repository.Repository) *MonitorService {
	return &MonitorService{
		repo:      repo,
		startTime: time.Now(),
	}
}

// SystemInfo 系统信息
type SystemInfo struct {
	GoVersion    string `json:"go_version"`     // Go版本
	GOOS         string `json:"goos"`           // 操作系统
	GOARCH       string `json:"goarch"`         // 架构
	NumCPU       int    `json:"num_cpu"`        // CPU核心数
	NumGoroutine int    `json:"num_goroutine"`  // Goroutine数量
	Uptime       int64  `json:"uptime"`         // 运行时间（秒）
	UptimeStr    string `json:"uptime_str"`     // 运行时间（格式化）
}

// MemoryStats 内存统计
type MemoryStats struct {
	Alloc        uint64  `json:"alloc"`          // 已分配内存（字节）
	TotalAlloc   uint64  `json:"total_alloc"`    // 累计分配内存（字节）
	Sys          uint64  `json:"sys"`            // 系统内存（字节）
	NumGC        uint32  `json:"num_gc"`         // GC次数
	HeapAlloc    uint64  `json:"heap_alloc"`     // 堆内存分配
	HeapSys      uint64  `json:"heap_sys"`       // 堆系统内存
	HeapIdle     uint64  `json:"heap_idle"`      // 堆空闲内存
	HeapInuse    uint64  `json:"heap_inuse"`     // 堆使用中内存
	StackInuse   uint64  `json:"stack_inuse"`    // 栈使用中内存
	AllocMB      float64 `json:"alloc_mb"`       // 已分配内存（MB）
	SysMB        float64 `json:"sys_mb"`         // 系统内存（MB）
	HeapAllocMB  float64 `json:"heap_alloc_mb"`  // 堆内存分配（MB）
}

// DatabaseStats 数据库统计
type DatabaseStats struct {
	TotalUsers        int64 `json:"total_users"`         // 总用户数
	TotalOrders       int64 `json:"total_orders"`        // 总订单数
	TotalProducts     int64 `json:"total_products"`      // 总商品数
	TotalTickets      int64 `json:"total_tickets"`       // 总工单数
	PendingOrders     int64 `json:"pending_orders"`      // 待支付订单
	ActiveSessions    int64 `json:"active_sessions"`     // 活跃会话数
	TodayOrders       int64 `json:"today_orders"`        // 今日订单
	TodayUsers        int64 `json:"today_users"`         // 今日新用户
	TodayRevenue      float64 `json:"today_revenue"`     // 今日收入
}

// APIStats API统计
type APIStats struct {
	TotalRequests   int64   `json:"total_requests"`    // 总请求数
	AvgResponseTime float64 `json:"avg_response_time"` // 平均响应时间（ms）
	ErrorRate       float64 `json:"error_rate"`        // 错误率
}

// GetSystemInfo 获取系统信息
// 返回：
//   - 系统信息
func (s *MonitorService) GetSystemInfo() *SystemInfo {
	uptime := time.Since(s.startTime)
	
	return &SystemInfo{
		GoVersion:    runtime.Version(),
		GOOS:         runtime.GOOS,
		GOARCH:       runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		Uptime:       int64(uptime.Seconds()),
		UptimeStr:    formatDuration(uptime),
	}
}

// GetMemoryStats 获取内存统计
// 返回：
//   - 内存统计
func (s *MonitorService) GetMemoryStats() *MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryStats{
		Alloc:       m.Alloc,
		TotalAlloc:  m.TotalAlloc,
		Sys:         m.Sys,
		NumGC:       m.NumGC,
		HeapAlloc:   m.HeapAlloc,
		HeapSys:     m.HeapSys,
		HeapIdle:    m.HeapIdle,
		HeapInuse:   m.HeapInuse,
		StackInuse:  m.StackInuse,
		AllocMB:     float64(m.Alloc) / 1024 / 1024,
		SysMB:       float64(m.Sys) / 1024 / 1024,
		HeapAllocMB: float64(m.HeapAlloc) / 1024 / 1024,
	}
}

// GetDatabaseStats 获取数据库统计
// 返回：
//   - 数据库统计
//   - 错误信息
func (s *MonitorService) GetDatabaseStats() (*DatabaseStats, error) {
	stats := &DatabaseStats{}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 总用户数
	s.repo.GetDB().Model(&model.User{}).Count(&stats.TotalUsers)

	// 总订单数
	s.repo.GetDB().Model(&model.Order{}).Count(&stats.TotalOrders)

	// 总商品数
	s.repo.GetDB().Model(&model.Product{}).Count(&stats.TotalProducts)

	// 总工单数
	s.repo.GetDB().Model(&model.SupportTicket{}).Count(&stats.TotalTickets)

	// 待支付订单
	s.repo.GetDB().Model(&model.Order{}).Where("status = 0").Count(&stats.PendingOrders)

	// 活跃会话数
	s.repo.GetDB().Model(&model.UserSession{}).Where("expires_at > ?", now).Count(&stats.ActiveSessions)

	// 今日订单
	s.repo.GetDB().Model(&model.Order{}).Where("created_at >= ?", todayStart).Count(&stats.TodayOrders)

	// 今日新用户
	s.repo.GetDB().Model(&model.User{}).Where("created_at >= ?", todayStart).Count(&stats.TodayUsers)

	// 今日收入
	var revenue struct {
		Total float64
	}
	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at >= ? AND status IN ?", todayStart, []int{1, 2}).
		Select("COALESCE(SUM(price), 0) as total").
		Scan(&revenue)
	stats.TodayRevenue = revenue.Total

	return stats, nil
}

// GetHealthStatus 获取健康状态
// 返回：
//   - 健康状态
func (s *MonitorService) GetHealthStatus() map[string]interface{} {
	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// 检查数据库连接
	if s.repo != nil {
		sqlDB, err := s.repo.GetDB().DB()
		if err != nil {
			status["database"] = "error"
			status["status"] = "unhealthy"
		} else if err := sqlDB.Ping(); err != nil {
			status["database"] = "disconnected"
			status["status"] = "unhealthy"
		} else {
			status["database"] = "connected"
		}
	} else {
		status["database"] = "not_initialized"
		status["status"] = "unhealthy"
	}

	// 内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memUsageMB := float64(m.Alloc) / 1024 / 1024
	status["memory_mb"] = memUsageMB

	// 如果内存使用超过1GB，标记为警告
	if memUsageMB > 1024 {
		status["memory_warning"] = true
	}

	// Goroutine数量
	goroutines := runtime.NumGoroutine()
	status["goroutines"] = goroutines

	// 如果Goroutine数量超过10000，标记为警告
	if goroutines > 10000 {
		status["goroutine_warning"] = true
	}

	return status
}

// GetRealtimeStats 获取实时统计（用于仪表盘刷新）
// 返回：
//   - 实时统计数据
func (s *MonitorService) GetRealtimeStats() map[string]interface{} {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	hourAgo := now.Add(-time.Hour)

	stats := make(map[string]interface{})

	// 最近1小时订单数
	var hourOrders int64
	s.repo.GetDB().Model(&model.Order{}).Where("created_at >= ?", hourAgo).Count(&hourOrders)
	stats["hour_orders"] = hourOrders

	// 最近1小时收入
	var hourRevenue struct {
		Total float64
	}
	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at >= ? AND status IN ?", hourAgo, []int{1, 2}).
		Select("COALESCE(SUM(price), 0) as total").
		Scan(&hourRevenue)
	stats["hour_revenue"] = hourRevenue.Total

	// 在线用户数（5分钟内有活动）
	fiveMinAgo := now.Add(-5 * time.Minute)
	var onlineUsers int64
	s.repo.GetDB().Model(&model.UserSession{}).
		Where("updated_at >= ? AND expires_at > ?", fiveMinAgo, now).
		Count(&onlineUsers)
	stats["online_users"] = onlineUsers

	// 待处理工单数
	var pendingTickets int64
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("status = 0").Count(&pendingTickets)
	stats["pending_tickets"] = pendingTickets

	// 今日访问量（通过登录历史估算）
	var todayLogins int64
	s.repo.GetDB().Model(&model.LoginHistory{}).Where("created_at >= ?", todayStart).Count(&todayLogins)
	stats["today_visits"] = todayLogins

	// 系统信息
	stats["goroutines"] = runtime.NumGoroutine()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats["memory_mb"] = float64(m.Alloc) / 1024 / 1024

	return stats
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return formatDurationStr(days, "天") + formatDurationStr(hours, "小时") + formatDurationStr(minutes, "分钟")
	}
	if hours > 0 {
		return formatDurationStr(hours, "小时") + formatDurationStr(minutes, "分钟") + formatDurationStr(seconds, "秒")
	}
	if minutes > 0 {
		return formatDurationStr(minutes, "分钟") + formatDurationStr(seconds, "秒")
	}
	return formatDurationStr(seconds, "秒")
}

func formatDurationStr(value int, unit string) string {
	if value > 0 {
		return string(rune('0'+value/10)) + string(rune('0'+value%10)) + unit
	}
	return ""
}
