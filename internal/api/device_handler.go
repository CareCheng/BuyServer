package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ==================== 登录设备管理 API ====================

// GetLoginDevices 获取登录设备列表
// 返回用户当前所有登录的设备信息
func GetLoginDevices(c *gin.Context) {
	if DeviceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	// 获取当前会话ID
	sessionID, _ := c.Cookie("user_session")

	devices, err := DeviceSvc.GetUserDevices(userID.(uint), sessionID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取设备列表失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": devices})
}

// RemoveLoginDevice 移除登录设备（踢出）
// 将指定设备从登录状态中移除
func RemoveLoginDevice(c *gin.Context) {
	if DeviceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	deviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的设备ID"})
		return
	}

	// 获取当前会话ID
	sessionID, _ := c.Cookie("user_session")

	if err := DeviceSvc.RemoveDevice(userID.(uint), uint(deviceID), sessionID); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "移除设备失败"})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		username, _ := c.Get("username")
		LogSvc.LogUserActionSimple(userID.(uint), username.(string), "remove_device", "login_device",
			strconv.FormatUint(deviceID, 10), "移除登录设备", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "设备已移除"})
}

// RemoveAllOtherDevices 移除所有其他设备
// 将除当前设备外的所有设备从登录状态中移除
func RemoveAllOtherDevices(c *gin.Context) {
	if DeviceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	// 获取当前会话ID
	sessionID, _ := c.Cookie("user_session")

	count, err := DeviceSvc.RemoveAllOtherDevices(userID.(uint), sessionID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "移除设备失败"})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		username, _ := c.Get("username")
		LogSvc.LogUserActionSimple(userID.(uint), username.(string), "remove_all_devices", "login_device",
			"", "移除所有其他设备", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "已移除其他设备", "count": count})
}

// GetLoginHistory 获取登录历史
// 返回用户的登录历史记录
func GetLoginHistory(c *gin.Context) {
	if DeviceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	histories, total, err := DeviceSvc.GetLoginHistory(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取登录历史失败"})
		return
	}

	c.JSON(200, gin.H{
		"success":   true,
		"data":      histories,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
