// Package api 提供 HTTP API 处理器
// balance_config_handler.go - 余额配置管理 API
package api

import (
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminGetBalanceConfig 管理员获取余额配置
func AdminGetBalanceConfig(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "配置服务未初始化"})
		return
	}

	cfg, err := ConfigSvc.GetBalanceConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取配置失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": cfg})
}

// AdminSaveBalanceConfig 管理员保存余额配置
func AdminSaveBalanceConfig(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "配置服务未初始化"})
		return
	}

	var req service.BalanceConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误: " + err.Error()})
		return
	}

	if err := ConfigSvc.SaveBalanceConfig(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "修改余额配置", "balance_config", "", req, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "保存成功"})
}

// GetBalanceConfigPublic 获取余额配置（公开部分，供前端显示限制信息）
func GetBalanceConfigPublic(c *gin.Context) {
	if ConfigSvc == nil {
		// 返回默认配置
		c.JSON(200, gin.H{
			"success": true,
			"data": gin.H{
				"min_recharge_amount": service.DefaultBalanceConfig.MinRechargeAmount,
				"max_recharge_amount": service.DefaultBalanceConfig.MaxRechargeAmount,
				"max_balance_limit":   service.DefaultBalanceConfig.MaxBalanceLimit,
			},
		})
		return
	}

	cfg, _ := ConfigSvc.GetBalanceConfig()
	// 只返回用户需要知道的限制信息
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"min_recharge_amount": cfg.MinRechargeAmount,
			"max_recharge_amount": cfg.MaxRechargeAmount,
			"max_balance_limit":   cfg.MaxBalanceLimit,
		},
	})
}
