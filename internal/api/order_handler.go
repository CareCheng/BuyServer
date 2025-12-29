package api

import (
	"strconv"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// GetProducts 获取商品列表
func GetProducts(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	products, err := ProductSvc.GetAllProducts(true)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":  true,
		"products": products,
	})
}

// GetProduct 获取单个商品
func GetProduct(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	product, err := ProductSvc.GetProductByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "商品不存在"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"product": product,
	})
}

// CreateOrder 创建订单
func CreateOrder(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	username := c.GetString("username")

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity"` // 购买数量，默认为1
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 数量默认为1，最小为1
	quantity := req.Quantity
	if quantity < 1 {
		quantity = 1
	}

	order, err := OrderSvc.CreateOrderWithQuantity(userID, username, req.ProductID, quantity, c.ClientIP())
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":  true,
		"message":  "订单创建成功",
		"order_no": order.OrderNo,
		"order":    order,
	})
}

// OrderDetail 订单详情
func OrderDetail(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	orderNo := c.Param("order_no")

	order, err := OrderSvc.GetOrderByOrderNo(orderNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	if order.UserID != userID {
		c.JSON(403, gin.H{"success": false, "error": "无权查看此订单"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"order":   order,
	})
}

// CancelOrder 取消订单
func CancelOrder(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		OrderNo string `json:"order_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := OrderSvc.CancelOrder(req.OrderNo, userID); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "订单已取消"})
}


// PayPalCreatePayment 创建PayPal支付
func PayPalCreatePayment(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		OrderNo string `json:"order_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取订单
	order, err := OrderSvc.GetOrderByOrderNo(req.OrderNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	if order.UserID != userID {
		c.JSON(403, gin.H{"success": false, "error": "无权操作此订单"})
		return
	}

	if order.Status != 0 {
		c.JSON(400, gin.H{"success": false, "error": "订单状态异常"})
		return
	}

	// 检查PayPal配置
	paypalConfig := &config.GlobalConfig.PaymentConfig.PayPal
	if !paypalConfig.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "PayPal支付未启用"})
		return
	}

	// 创建PayPal服务
	paypalSvc := service.NewPayPalService(paypalConfig)

	// 创建PayPal订单
	description := order.ProductName + " - " + order.OrderNo
	paypalOrder, err := paypalSvc.CreateOrder(order.OrderNo, order.Price, description)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建PayPal订单失败: " + err.Error()})
		return
	}

	// 获取支付链接
	var approveURL string
	for _, link := range paypalOrder.Links {
		if link.Rel == "approve" {
			approveURL = link.Href
			break
		}
	}

	if approveURL == "" {
		c.JSON(500, gin.H{"success": false, "error": "获取支付链接失败"})
		return
	}

	c.JSON(200, gin.H{
		"success":         true,
		"paypal_order_id": paypalOrder.ID,
		"approve_url":     approveURL,
	})
}

// PayPalCapturePayment 捕获PayPal支付（用户授权后调用）
func PayPalCapturePayment(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		OrderNo       string `json:"order_no" binding:"required"`
		PayPalOrderID string `json:"paypal_order_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取订单
	order, err := OrderSvc.GetOrderByOrderNo(req.OrderNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	if order.UserID != userID {
		c.JSON(403, gin.H{"success": false, "error": "无权操作此订单"})
		return
	}

	if order.Status != 0 {
		c.JSON(400, gin.H{"success": false, "error": "订单状态异常"})
		return
	}

	// 检查PayPal配置
	paypalConfig := &config.GlobalConfig.PaymentConfig.PayPal
	if !paypalConfig.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "PayPal支付未启用"})
		return
	}

	// 创建PayPal服务
	paypalSvc := service.NewPayPalService(paypalConfig)

	// 捕获支付
	captureResp, err := paypalSvc.CaptureOrder(req.PayPalOrderID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "捕获支付失败: " + err.Error()})
		return
	}

	// 检查支付状态
	if captureResp.Status != "COMPLETED" {
		c.JSON(400, gin.H{"success": false, "error": "支付未完成，状态: " + captureResp.Status})
		return
	}

	// 获取 PayPal 支付订单号
	paymentNo := req.PayPalOrderID
	if len(captureResp.PurchaseUnits) > 0 && len(captureResp.PurchaseUnits[0].Payments.Captures) > 0 {
		paymentNo = captureResp.PurchaseUnits[0].Payments.Captures[0].ID
	}

	// 处理订单支付（传递支付订单号）
	order, err = OrderSvc.ProcessPayment(req.OrderNo, "PayPal", paymentNo)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "处理订单失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":   true,
		"message":   "支付成功",
		"order_no":  order.OrderNo,
		"kami_code": order.KamiCode,
		"order":     order,
	})
}

// PayPalReturn PayPal支付返回页面处理
func PayPalReturn(c *gin.Context) {
	// 从URL参数获取token（PayPal订单ID）
	token := c.Query("token")
	
	// 重定向到支付结果页面处理
	c.Redirect(302, "/payment/result?paypal_order_id="+token)
}

// PayPalCancel PayPal支付取消处理
func PayPalCancel(c *gin.Context) {
	c.Redirect(302, "/payment/cancel")
}
