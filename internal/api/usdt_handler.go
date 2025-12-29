package api

import (
	"io"
	"net/http"

	"user-frontend/internal/config"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

var USDTSvc *service.USDTService

// InitUSDTService 初始化USDT服务
func InitUSDTService(cfg *config.Config) {
	USDTSvc = service.NewUSDTService(cfg)
}

// USDTGetConfig 获取USDT配置（前端使用）
// GET /api/usdt/config
func USDTGetConfig(c *gin.Context) {
	if USDTSvc == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"enabled": false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    USDTSvc.GetConfig(),
	})
}

// USDTCreatePayment 创建USDT支付
// POST /api/usdt/create
// 安全特性：验证用户订单归属
func USDTCreatePayment(c *gin.Context) {
	if USDTSvc == nil || !USDTSvc.IsEnabled() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "USDT支付未启用",
		})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		OrderNo string `json:"order_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "参数错误",
		})
		return
	}

	// 获取订单信息
	if OrderSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "订单服务未初始化",
		})
		return
	}

	// 【安全检查】验证订单归属
	order, err := OrderSvc.ValidateOrderOwnership(req.OrderNo, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 验证订单状态（0=待支付）
	if order.Status != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "订单状态不正确",
		})
		return
	}

	// 获取商品名称
	productName := "商品购买"
	if ProductSvc != nil {
		if product, err := ProductSvc.GetProductByID(order.ProductID); err == nil {
			productName = product.Name
		}
	}

	// 创建USDT支付
	paymentReq := &service.USDTPaymentRequest{
		OrderNo:     order.OrderNo,
		Amount:      order.Price,
		Currency:    "CNY", // 默认人民币
		Description: productName,
	}

	payment, err := USDTSvc.CreatePayment(paymentReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建支付失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    payment,
	})
}

// USDTGetPaymentStatus 获取USDT支付状态
// GET /api/usdt/status/:payment_id
func USDTGetPaymentStatus(c *gin.Context) {
	if USDTSvc == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "USDT服务未初始化",
		})
		return
	}

	paymentID := c.Param("payment_id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少payment_id参数",
		})
		return
	}

	status, err := USDTSvc.GetPaymentStatus(paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取支付状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// USDTWebhook 处理USDT Webhook回调
// POST /usdt/webhook
// 安全特性：
//   - 强制验证签名（必须配置webhook密钥）
//   - 验证支付金额
func USDTWebhook(c *gin.Context) {
	if USDTSvc == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "USDT服务未初始化"})
		return
	}

	// 读取请求体
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}

	// 获取签名头（不同提供商可能使用不同的头）
	signature := c.GetHeader("X-Nowpayments-Sig")
	if signature == "" {
		signature = c.GetHeader("X-Coingate-Signature")
	}

	// 【安全检查】强制验证签名
	// 如果配置了webhook密钥，必须验证签名
	if err := USDTSvc.VerifyWebhook(payload, signature); err != nil {
		// 记录安全日志
		if LogSvc != nil {
			LogSvc.LogSecurityEvent(
				"usdt_webhook_signature_failed",
				c.ClientIP(),
				c.GetHeader("User-Agent"),
				map[string]interface{}{
					"error":     err.Error(),
					"signature": signature,
				},
			)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "签名验证失败"})
		return
	}

	// 解析事件
	orderNo, status, paidAmount, err := USDTSvc.ParseWebhookEventWithAmount(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析事件失败: " + err.Error()})
		return
	}

	// 处理支付完成
	if status == "confirmed" || status == "finished" || status == "paid" {
		if OrderSvc != nil && orderNo != "" {
			// 使用带金额验证的支付处理方法
			_, err := OrderSvc.ProcessPaymentWithAmount(orderNo, "USDT", "", paidAmount)
			if err != nil {
				// 记录支付处理失败
				if LogSvc != nil {
					LogSvc.LogSecurityEvent(
						"usdt_payment_process_failed",
						c.ClientIP(),
						c.GetHeader("User-Agent"),
						map[string]interface{}{
							"order_no":    orderNo,
							"paid_amount": paidAmount,
							"error":       err.Error(),
						},
					)
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// USDTTestConnection 测试USDT连接（管理员）
// POST /api/admin/usdt/test
func USDTTestConnection(c *gin.Context) {
	if USDTSvc == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "USDT服务未初始化",
		})
		return
	}

	err := USDTSvc.TestConnection()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "USDT连接测试成功",
	})
}

// AdminConfirmUSDTPayment 管理员确认USDT支付（手动模式）
// POST /api/admin/usdt/confirm
func AdminConfirmUSDTPayment(c *gin.Context) {
	var req struct {
		OrderNo string `json:"order_no" binding:"required"`
		TxHash  string `json:"tx_hash"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "参数错误",
		})
		return
	}

	if OrderSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "订单服务未初始化",
		})
		return
	}

	// 完成订单
	_, err := OrderSvc.ProcessPayment(req.OrderNo, "usdt", req.TxHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "确认支付失败: " + err.Error(),
		})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(
			adminUsername.(string),
			"confirm_usdt_payment",
			"order",
			req.OrderNo,
			gin.H{"tx_hash": req.TxHash},
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "支付确认成功",
	})
}
