package api

import (
	"io"
	"net/http"

	"user-frontend/internal/config"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

var StripeSvc *service.StripeService

// InitStripeService 初始化Stripe服务
func InitStripeService(cfg *config.Config) {
	StripeSvc = service.NewStripeService(cfg)
}

// StripeGetConfig 获取Stripe配置（前端使用）
// GET /api/stripe/config
func StripeGetConfig(c *gin.Context) {
	if StripeSvc == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"enabled":         false,
				"publishable_key": "",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled":         StripeSvc.IsEnabled(),
			"publishable_key": StripeSvc.GetPublishableKey(),
		},
	})
}

// StripeCreateCheckoutSession 创建Stripe Checkout会话
// POST /api/stripe/create
// 安全特性：验证用户订单归属
func StripeCreateCheckoutSession(c *gin.Context) {
	if StripeSvc == nil || !StripeSvc.IsEnabled() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Stripe支付未启用",
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

	// 验证订单状态（0=待支付，1=已支付，2=已完成，3=已取消）
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

	// 获取基础URL
	baseURL := c.Request.Header.Get("Origin")
	if baseURL == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		baseURL = scheme + "://" + c.Request.Host
	}

	// 创建Checkout会话
	session, err := StripeSvc.CreateCheckoutSessionForOrder(
		order.OrderNo,
		order.Price,
		productName,
		baseURL,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建支付会话失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"session_id":   session.ID,
			"checkout_url": session.URL,
		},
	})
}

// StripeVerifyPayment 验证Stripe支付状态
// GET /api/stripe/verify/:session_id
func StripeVerifyPayment(c *gin.Context) {
	if StripeSvc == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Stripe服务未初始化",
		})
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少session_id参数",
		})
		return
	}

	result, err := StripeSvc.VerifyPayment(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "验证支付状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// StripeWebhook 处理Stripe Webhook回调
// POST /stripe/webhook
// 安全特性：
//   - 强制验证签名
//   - 验证支付金额
func StripeWebhook(c *gin.Context) {
	if StripeSvc == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stripe服务未初始化"})
		return
	}

	// 读取请求体
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}

	// 获取签名头
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		// 记录安全日志
		if LogSvc != nil {
			LogSvc.LogSecurityEvent(
				"stripe_webhook_missing_signature",
				c.ClientIP(),
				c.GetHeader("User-Agent"),
				nil,
			)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少签名头"})
		return
	}

	// 验证签名并解析事件
	event, err := StripeSvc.VerifyWebhookSignature(payload, signature)
	if err != nil {
		// 记录安全日志
		if LogSvc != nil {
			LogSvc.LogSecurityEvent(
				"stripe_webhook_signature_failed",
				c.ClientIP(),
				c.GetHeader("User-Agent"),
				map[string]interface{}{"error": err.Error()},
			)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "签名验证失败: " + err.Error()})
		return
	}

	// 处理不同类型的事件
	switch event.Type {
	case "checkout.session.completed":
		// Checkout会话完成
		orderNo, paidAmount, err := StripeSvc.ParseCheckoutSessionCompletedWithAmount(event.Data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "解析事件失败: " + err.Error()})
			return
		}
		// 完成订单（带金额验证）
		if OrderSvc != nil && orderNo != "" {
			_, err := OrderSvc.ProcessPaymentWithAmount(orderNo, "Stripe", event.ID, paidAmount)
			if err != nil {
				// 记录支付处理失败
				if LogSvc != nil {
					LogSvc.LogSecurityEvent(
						"stripe_payment_process_failed",
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

	case "payment_intent.succeeded":
		// 支付意图成功
		orderNo, paidAmount, err := StripeSvc.ParsePaymentIntentSucceededWithAmount(event.Data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "解析事件失败: " + err.Error()})
			return
		}
		// 完成订单（带金额验证）
		if OrderSvc != nil && orderNo != "" {
			_, err := OrderSvc.ProcessPaymentWithAmount(orderNo, "Stripe", event.ID, paidAmount)
			if err != nil {
				// 记录安全日志
				if LogSvc != nil {
					LogSvc.LogSecurityEvent(
						"stripe_payment_process_failed",
						c.ClientIP(),
						c.GetHeader("User-Agent"),
						map[string]interface{}{
							"order_no":    orderNo,
							"paid_amount": paidAmount,
							"error":       err.Error(),
						},
					)
				}
			}
		}

	case "payment_intent.payment_failed":
		// 支付失败 - 记录日志
		if LogSvc != nil {
			LogSvc.LogAdminActionSimple("system", "stripe_payment_failed", "webhook", event.ID, string(event.Data), "", "")
		}
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"received": true})
}

// StripeTestConnection 测试Stripe连接（管理员）
// POST /api/admin/stripe/test
func StripeTestConnection(c *gin.Context) {
	if StripeSvc == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Stripe服务未初始化",
		})
		return
	}

	err := StripeSvc.TestConnection()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Stripe连接测试成功",
	})
}
