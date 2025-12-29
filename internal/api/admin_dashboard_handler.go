package api

import (
	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 仪表盘相关 API ====================

// AdminDashboard 管理仪表盘
func AdminDashboard(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(200, gin.H{
			"success":      true,
			"db_connected": false,
		})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats, _ := OrderSvc.GetOrderStats()
	c.JSON(200, gin.H{
		"success":      true,
		"db_connected": true,
		"stats":        stats,
	})
}
