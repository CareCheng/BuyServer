package api

import (
	"github.com/gin-gonic/gin"
)

// AdminGetSystemInfo 获取系统信息
// GET /api/admin/monitor/system
func AdminGetSystemInfo(c *gin.Context) {
	if MonitorSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	info := MonitorSvc.GetSystemInfo()
	c.JSON(200, gin.H{"success": true, "data": info})
}

// AdminGetMemoryStats 获取内存统计
// GET /api/admin/monitor/memory
func AdminGetMemoryStats(c *gin.Context) {
	if MonitorSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats := MonitorSvc.GetMemoryStats()
	c.JSON(200, gin.H{"success": true, "data": stats})
}

// AdminGetDatabaseStats 获取数据库统计
// GET /api/admin/monitor/database
func AdminGetDatabaseStats(c *gin.Context) {
	if MonitorSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats, err := MonitorSvc.GetDatabaseStats()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取统计失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": stats})
}

// AdminGetHealthStatus 获取健康状态
// GET /api/admin/monitor/health
func AdminGetHealthStatus(c *gin.Context) {
	if MonitorSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	status := MonitorSvc.GetHealthStatus()
	c.JSON(200, gin.H{"success": true, "data": status})
}

// AdminGetRealtimeStats 获取实时统计
// GET /api/admin/monitor/realtime
func AdminGetRealtimeStats(c *gin.Context) {
	if MonitorSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats := MonitorSvc.GetRealtimeStats()
	c.JSON(200, gin.H{"success": true, "data": stats})
}

// AdminGetMonitorOverview 获取监控概览
// GET /api/admin/monitor/overview
func AdminGetMonitorOverview(c *gin.Context) {
	if MonitorSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	overview := gin.H{
		"system":   MonitorSvc.GetSystemInfo(),
		"memory":   MonitorSvc.GetMemoryStats(),
		"health":   MonitorSvc.GetHealthStatus(),
		"realtime": MonitorSvc.GetRealtimeStats(),
	}

	// 获取数据库统计
	if dbStats, err := MonitorSvc.GetDatabaseStats(); err == nil {
		overview["database"] = dbStats
	}

	c.JSON(200, gin.H{"success": true, "data": overview})
}

// PublicHealthCheck 公开健康检查接口
// GET /api/health
func PublicHealthCheck(c *gin.Context) {
	if MonitorSvc == nil {
		c.JSON(503, gin.H{
			"status":  "unhealthy",
			"message": "服务未初始化",
		})
		return
	}

	status := MonitorSvc.GetHealthStatus()
	if status["status"] == "healthy" {
		c.JSON(200, status)
	} else {
		c.JSON(503, status)
	}
}
