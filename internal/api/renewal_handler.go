package api

import (
	"github.com/gin-gonic/gin"
)

// GetUserKamis 获取用户的卡密列表（用于续费页面）
func GetUserKamis(c *gin.Context) {
	userID := c.GetUint("user_id")

	if RenewalSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	kamis, err := RenewalSvc.GetUserKamis(userID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"kamis":   kamis,
	})
}

// GetExpiringKamis 获取即将过期的卡密
func GetExpiringKamis(c *gin.Context) {
	userID := c.GetUint("user_id")

	if RenewalSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 获取用户的所有卡密
	kamis, err := RenewalSvc.GetUserKamis(userID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 筛选即将过期的（7天内）
	var expiringKamis []interface{}
	for _, kami := range kamis {
		if !kami.IsExpired && kami.DaysLeft <= 7 {
			expiringKamis = append(expiringKamis, kami)
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"kamis":   expiringKamis,
		"count":   len(expiringKamis),
	})
}

// GetExpiredKamis 获取已过期的卡密
func GetExpiredKamis(c *gin.Context) {
	userID := c.GetUint("user_id")

	if RenewalSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 获取用户的所有卡密
	kamis, err := RenewalSvc.GetUserKamis(userID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 筛选已过期的
	var expiredKamis []interface{}
	for _, kami := range kamis {
		if kami.IsExpired {
			expiredKamis = append(expiredKamis, kami)
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"kamis":   expiredKamis,
		"count":   len(expiredKamis),
	})
}

// RequestRenewalReminder 请求发送续费提醒（手动触发）
func RequestRenewalReminder(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		OrderNo    string `json:"order_no" binding:"required"`
		RemindType string `json:"remind_type"` // 可选，默认根据剩余天数自动判断
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if RenewalSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 如果未指定提醒类型，使用默认值
	remindType := req.RemindType
	if remindType == "" {
		remindType = "manual"
	}

	if err := RenewalSvc.SendRenewalReminder(userID, req.OrderNo, remindType); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "续费提醒已发送",
	})
}

// GetRenewalHistory 获取续费提醒历史
func GetRenewalHistory(c *gin.Context) {
	userID := c.GetUint("user_id")

	if RenewalSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	history, err := RenewalSvc.GetRenewalHistory(userID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"history": history,
	})
}

// GetRenewalStats 获取续费统计（用于用户中心显示）
func GetRenewalStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	if RenewalSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	kamis, err := RenewalSvc.GetUserKamis(userID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 统计
	var totalKamis, activeKamis, expiringKamis, expiredKamis int
	for _, kami := range kamis {
		totalKamis++
		if kami.IsExpired {
			expiredKamis++
		} else {
			activeKamis++
			if kami.DaysLeft <= 7 {
				expiringKamis++
			}
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"stats": gin.H{
			"total":    totalKamis,
			"active":   activeKamis,
			"expiring": expiringKamis, // 7天内即将过期
			"expired":  expiredKamis,
		},
	})
}
