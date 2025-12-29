package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetLoginAlerts 获取用户的登录提醒记录
// GET /api/user/login-alerts
func GetLoginAlerts(c *gin.Context) {
	userID := c.GetUint("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	if LoginAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	alerts, total, err := LoginAlertSvc.GetUserAlerts(userID, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取记录失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"alerts":    alerts,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetUnacknowledgedAlertCount 获取未确认的提醒数量
// GET /api/user/login-alerts/unread-count
func GetUnacknowledgedAlertCount(c *gin.Context) {
	userID := c.GetUint("user_id")

	if LoginAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	count := LoginAlertSvc.GetUnacknowledgedCount(userID)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"count": count,
		},
	})
}

// AcknowledgeLoginAlert 确认登录提醒
// POST /api/user/login-alert/:id/acknowledge
func AcknowledgeLoginAlert(c *gin.Context) {
	userID := c.GetUint("user_id")

	alertID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "提醒ID无效"})
		return
	}

	if LoginAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := LoginAlertSvc.AcknowledgeAlert(userID, uint(alertID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "确认失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "已确认"})
}

// GetLoginLocations 获取用户的登录地点列表
// GET /api/user/login-locations
func GetLoginLocations(c *gin.Context) {
	userID := c.GetUint("user_id")

	if LoginAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	locations, err := LoginAlertSvc.GetUserLocations(userID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取地点失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    locations,
	})
}

// TrustLoginLocation 将登录地点标记为可信
// POST /api/user/login-location/:id/trust
func TrustLoginLocation(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	locationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "地点ID无效"})
		return
	}

	if LoginAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := LoginAlertSvc.TrustLocation(userID, uint(locationID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "操作失败"})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogUserActionSimple(userID, username, "trust_location", "login_location", strconv.Itoa(int(locationID)), "标记登录地点为可信", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "已标记为可信地点"})
}

// RemoveLoginLocation 移除登录地点
// DELETE /api/user/login-location/:id
func RemoveLoginLocation(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	locationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "地点ID无效"})
		return
	}

	if LoginAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := LoginAlertSvc.RemoveLocation(userID, uint(locationID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogUserActionSimple(userID, username, "remove_location", "login_location", strconv.Itoa(int(locationID)), "移除登录地点", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "已移除"})
}
