package api

import (
	"user-frontend/internal/config"
	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	status := "healthy"
	dbStatus := "connected"

	if !model.DBConnected {
		dbStatus = "disconnected"
		status = "degraded"
	}

	c.JSON(200, gin.H{
		"status":    status,
		"database":  dbStatus,
		"timestamp": c.GetInt64("request_time"),
	})
}

// GetAnnouncements 获取公告列表（公开）
func GetAnnouncements(c *gin.Context) {
	if AnnouncementSvc == nil {
		c.JSON(200, gin.H{"success": true, "announcements": []interface{}{}})
		return
	}

	announcements, err := AnnouncementSvc.GetActiveAnnouncements()
	if err != nil {
		c.JSON(200, gin.H{"success": true, "announcements": []interface{}{}})
		return
	}

	c.JSON(200, gin.H{"success": true, "announcements": announcements})
}

// GetCategories 获取分类列表（公开）
func GetCategories(c *gin.Context) {
	if CategorySvc == nil {
		c.JSON(200, gin.H{"success": true, "categories": []interface{}{}})
		return
	}

	categories, err := CategorySvc.GetAllCategories(true)
	if err != nil {
		c.JSON(200, gin.H{"success": true, "categories": []interface{}{}})
		return
	}

	c.JSON(200, gin.H{"success": true, "categories": categories})
}

// QueryOrderPublic 公开订单查询（通过订单号+邮箱）
func QueryOrderPublic(c *gin.Context) {
	var req struct {
		OrderNo string `json:"order_no" binding:"required"`
		Email   string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请输入订单号和邮箱"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	order, err := OrderSvc.GetOrderByOrderNoAndEmail(req.OrderNo, req.Email)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在或邮箱不匹配"})
		return
	}

	// 返回订单信息（隐藏部分敏感信息）
	c.JSON(200, gin.H{
		"success": true,
		"order": gin.H{
			"order_no":       order.OrderNo,
			"product_name":   order.ProductName,
			"price":          order.Price,
			"duration":       order.Duration,
			"duration_unit":  order.DurationUnit,
			"status":         order.Status,
			"payment_method": order.PaymentMethod,
			"payment_time":   order.PaymentTime,
			"kami_code":      order.KamiCode,
			"created_at":     order.CreatedAt,
		},
	})
}

// GetPaymentMethods 获取可用的支付方式（公开）
// 返回当前启用的支付方式列表，不包含敏感配置信息
func GetPaymentMethods(c *gin.Context) {
	paymentCfg := config.GlobalConfig.PaymentConfig

	c.JSON(200, gin.H{
		"success": true,
		"methods": gin.H{
			"paypal": gin.H{
				"enabled": paymentCfg.PayPal.Enabled,
				"sandbox": paymentCfg.PayPal.Sandbox,
			},
			"stripe": gin.H{
				"enabled": paymentCfg.StripeEnabled,
			},
			"alipay_f2f": gin.H{
				"enabled": paymentCfg.AlipayF2F.Enabled,
			},
			"wechat_pay": gin.H{
				"enabled": paymentCfg.WechatPay.Enabled,
			},
			"yi_pay": gin.H{
				"enabled": paymentCfg.YiPay.Enabled,
			},
			"usdt": gin.H{
				"enabled": paymentCfg.USDTEnabled,
				"network": paymentCfg.USDTNetwork,
			},
			"balance": gin.H{
				"enabled": true, // 余额支付始终可用
			},
		},
	})
}
