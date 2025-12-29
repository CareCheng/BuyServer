package api

import (
	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// ==========================================
//         支付宝当面付 API
// ==========================================

// AlipayCreatePayment 创建支付宝当面付订单
// 生成二维码供用户扫码支付
func AlipayCreatePayment(c *gin.Context) {
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

	// 检查支付宝配置
	alipayCfg := &config.GlobalConfig.PaymentConfig.AlipayF2F
	if !alipayCfg.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "支付宝支付未启用"})
		return
	}

	// 创建支付宝服务
	alipaySvc := service.NewAlipayService(alipayCfg)

	// 创建支付宝订单，获取二维码
	description := order.ProductName + " - " + order.OrderNo
	qrCode, err := alipaySvc.CreatePreOrder(order.OrderNo, order.Price, description)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建支付宝订单失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"qr_code": qrCode,
	})
}

// AlipayNotify 支付宝异步通知回调
func AlipayNotify(c *gin.Context) {
	// 检查支付宝配置
	alipayCfg := &config.GlobalConfig.PaymentConfig.AlipayF2F
	if !alipayCfg.Enabled {
		c.String(400, "fail")
		return
	}

	if OrderSvc == nil {
		c.String(500, "fail")
		return
	}

	alipaySvc := service.NewAlipayService(alipayCfg)

	// 解析通知参数
	if err := c.Request.ParseForm(); err != nil {
		c.String(400, "fail")
		return
	}

	// 验证签名并获取订单信息
	orderNo, tradeNo, err := alipaySvc.VerifyNotify(c.Request.PostForm)
	if err != nil {
		c.String(400, "fail")
		return
	}

	// 处理订单支付
	_, err = OrderSvc.ProcessPayment(orderNo, "支付宝", tradeNo)
	if err != nil {
		// 记录错误但仍返回成功，避免重复通知
		c.String(200, "success")
		return
	}

	c.String(200, "success")
}

// AlipayQueryStatus 查询支付宝订单状态
func AlipayQueryStatus(c *gin.Context) {
	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	orderNo := c.Param("order_no")

	// 获取订单
	order, err := OrderSvc.GetOrderByOrderNo(orderNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	if order.UserID != userID {
		c.JSON(403, gin.H{"success": false, "error": "无权查询此订单"})
		return
	}

	// 如果订单已完成，直接返回
	if order.Status == 2 {
		c.JSON(200, gin.H{
			"success": true,
			"paid":    true,
			"order":   order,
		})
		return
	}

	// 检查支付宝配置
	alipayCfg := &config.GlobalConfig.PaymentConfig.AlipayF2F
	if !alipayCfg.Enabled {
		c.JSON(200, gin.H{"success": true, "paid": false})
		return
	}

	// 查询支付宝订单状态
	alipaySvc := service.NewAlipayService(alipayCfg)
	paid, tradeNo, err := alipaySvc.QueryOrder(orderNo)
	if err != nil {
		c.JSON(200, gin.H{"success": true, "paid": false})
		return
	}

	// 如果已支付，处理订单
	if paid {
		order, err = OrderSvc.ProcessPayment(orderNo, "支付宝", tradeNo)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "处理订单失败"})
			return
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"paid":    paid,
		"order":   order,
	})
}

// ==========================================
//         微信支付 API
// ==========================================

// WechatCreatePayment 创建微信支付订单
func WechatCreatePayment(c *gin.Context) {
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

	// 检查微信支付配置
	wechatCfg := &config.GlobalConfig.PaymentConfig.WechatPay
	if !wechatCfg.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "微信支付未启用"})
		return
	}

	// 创建微信支付服务
	wechatSvc := service.NewWechatPayService(wechatCfg)

	// 创建微信支付订单，获取二维码
	description := order.ProductName
	qrCode, err := wechatSvc.CreateNativeOrder(order.OrderNo, order.Price, description)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建微信支付订单失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"qr_code": qrCode,
	})
}

// WechatNotify 微信支付异步通知回调
func WechatNotify(c *gin.Context) {
	// 检查微信支付配置
	wechatCfg := &config.GlobalConfig.PaymentConfig.WechatPay
	if !wechatCfg.Enabled {
		c.XML(400, gin.H{"return_code": "FAIL", "return_msg": "支付未启用"})
		return
	}

	if OrderSvc == nil {
		c.XML(500, gin.H{"return_code": "FAIL", "return_msg": "服务未初始化"})
		return
	}

	wechatSvc := service.NewWechatPayService(wechatCfg)

	// 验证签名并获取订单信息
	orderNo, tradeNo, err := wechatSvc.VerifyNotify(c.Request)
	if err != nil {
		c.XML(400, gin.H{"return_code": "FAIL", "return_msg": err.Error()})
		return
	}

	// 处理订单支付
	_, err = OrderSvc.ProcessPayment(orderNo, "微信支付", tradeNo)
	if err != nil {
		// 记录错误但仍返回成功
	}

	c.XML(200, gin.H{"return_code": "SUCCESS", "return_msg": "OK"})
}

// WechatQueryStatus 查询微信支付订单状态
func WechatQueryStatus(c *gin.Context) {
	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	orderNo := c.Param("order_no")

	// 获取订单
	order, err := OrderSvc.GetOrderByOrderNo(orderNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	if order.UserID != userID {
		c.JSON(403, gin.H{"success": false, "error": "无权查询此订单"})
		return
	}

	// 如果订单已完成，直接返回
	if order.Status == 2 {
		c.JSON(200, gin.H{
			"success": true,
			"paid":    true,
			"order":   order,
		})
		return
	}

	// 检查微信支付配置
	wechatCfg := &config.GlobalConfig.PaymentConfig.WechatPay
	if !wechatCfg.Enabled {
		c.JSON(200, gin.H{"success": true, "paid": false})
		return
	}

	// 查询微信支付订单状态
	wechatSvc := service.NewWechatPayService(wechatCfg)
	paid, tradeNo, err := wechatSvc.QueryOrder(orderNo)
	if err != nil {
		c.JSON(200, gin.H{"success": true, "paid": false})
		return
	}

	// 如果已支付，处理订单
	if paid {
		order, err = OrderSvc.ProcessPayment(orderNo, "微信支付", tradeNo)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "处理订单失败"})
			return
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"paid":    paid,
		"order":   order,
	})
}

// ==========================================
//         易支付 API
// ==========================================

// YiPayCreatePayment 创建易支付订单
func YiPayCreatePayment(c *gin.Context) {
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

	// 检查易支付配置
	yipayCfg := &config.GlobalConfig.PaymentConfig.YiPay
	if !yipayCfg.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "易支付未启用"})
		return
	}

	// 创建易支付服务
	yipaySvc := service.NewYiPayService(yipayCfg)

	// 创建易支付订单，获取支付URL
	productName := order.ProductName
	payURL, err := yipaySvc.CreateOrder(order.OrderNo, order.Price, productName)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建易支付订单失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"pay_url": payURL,
	})
}

// YiPayNotify 易支付异步通知回调
func YiPayNotify(c *gin.Context) {
	// 检查易支付配置
	yipayCfg := &config.GlobalConfig.PaymentConfig.YiPay
	if !yipayCfg.Enabled {
		c.String(400, "fail")
		return
	}

	if OrderSvc == nil {
		c.String(500, "fail")
		return
	}

	yipaySvc := service.NewYiPayService(yipayCfg)

	// 验证签名并获取订单信息
	orderNo, tradeNo, err := yipaySvc.VerifyNotify(c.Request)
	if err != nil {
		c.String(400, "fail")
		return
	}

	// 处理订单支付
	_, err = OrderSvc.ProcessPayment(orderNo, "易支付", tradeNo)
	if err != nil {
		// 记录错误但仍返回成功
	}

	c.String(200, "success")
}

// YiPayReturn 易支付同步返回处理
func YiPayReturn(c *gin.Context) {
	orderNo := c.Query("out_trade_no")
	tradeNo := c.Query("trade_no")

	// 重定向到支付结果页面
	c.Redirect(302, "/payment/result?out_trade_no="+orderNo+"&trade_no="+tradeNo)
}

// YiPayCallback 易支付回调验证（前端调用）
func YiPayCallback(c *gin.Context) {
	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		OutTradeNo string `json:"out_trade_no" binding:"required"`
		TradeNo    string `json:"trade_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 检查易支付配置
	yipayCfg := &config.GlobalConfig.PaymentConfig.YiPay
	if !yipayCfg.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "易支付未启用"})
		return
	}

	// 查询订单
	order, err := OrderSvc.GetOrderByOrderNo(req.OutTradeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	// 如果订单已完成，直接返回
	if order.Status == 2 {
		c.JSON(200, gin.H{
			"success":   true,
			"order_no":  order.OrderNo,
			"kami_code": order.KamiCode,
		})
		return
	}

	// 处理订单支付
	order, err = OrderSvc.ProcessPayment(req.OutTradeNo, "易支付", req.TradeNo)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "处理订单失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":   true,
		"order_no":  order.OrderNo,
		"kami_code": order.KamiCode,
	})
}

// ==========================================
//         充值订单支付 API
// ==========================================

// YiPayCreateRechargePayment 创建充值订单的易支付支付
func YiPayCreateRechargePayment(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "余额服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		RechargeNo string `json:"recharge_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取充值订单
	order, err := BalanceSvc.GetRechargeOrder(req.RechargeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "充值订单不存在"})
		return
	}

	// 验证订单归属
	if order.UserID != userID {
		c.JSON(403, gin.H{"success": false, "error": "无权操作此订单"})
		return
	}

	// 验证订单状态
	if order.Status != 0 { // 0 = 待支付
		c.JSON(400, gin.H{"success": false, "error": "订单状态异常，无法支付"})
		return
	}

	// 检查易支付配置
	yipayCfg := &config.GlobalConfig.PaymentConfig.YiPay
	if !yipayCfg.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "易支付未启用"})
		return
	}

	// 创建易支付服务
	yipaySvc := service.NewYiPayService(yipayCfg)

	// 创建易支付订单，获取支付URL
	// 使用实际支付金额（可能有折扣）
	payAmount := order.PayAmount
	if payAmount <= 0 {
		payAmount = order.Amount
	}
	productName := "余额充值"
	payURL, err := yipaySvc.CreateOrder(order.RechargeNo, payAmount, productName)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建易支付订单失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"pay_url": payURL,
	})
}

// YiPayRechargeNotify 充值订单易支付异步通知回调
func YiPayRechargeNotify(c *gin.Context) {
	// 检查易支付配置
	yipayCfg := &config.GlobalConfig.PaymentConfig.YiPay
	if !yipayCfg.Enabled {
		c.String(400, "fail")
		return
	}

	if BalanceSvc == nil {
		c.String(500, "fail")
		return
	}

	yipaySvc := service.NewYiPayService(yipayCfg)

	// 验证签名并获取订单信息
	rechargeNo, tradeNo, err := yipaySvc.VerifyNotify(c.Request)
	if err != nil {
		c.String(400, "fail")
		return
	}

	// 完成充值订单
	err = BalanceSvc.CompleteRechargeOrder(rechargeNo, tradeNo)
	if err != nil {
		// 记录错误但仍返回成功，避免重复通知
	}

	c.String(200, "success")
}

// YiPayRechargeCallback 充值订单易支付回调验证（前端调用）
func YiPayRechargeCallback(c *gin.Context) {
	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "余额服务未初始化"})
		return
	}

	var req struct {
		OutTradeNo string `json:"out_trade_no" binding:"required"`
		TradeNo    string `json:"trade_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 检查易支付配置
	yipayCfg := &config.GlobalConfig.PaymentConfig.YiPay
	if !yipayCfg.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "易支付未启用"})
		return
	}

	// 查询充值订单
	order, err := BalanceSvc.GetRechargeOrder(req.OutTradeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "充值订单不存在"})
		return
	}

	// 如果订单已完成，直接返回
	if order.Status == 1 { // 1 = 已支付
		c.JSON(200, gin.H{
			"success":      true,
			"recharge_no":  order.RechargeNo,
			"amount":       order.Amount,
			"bonus_amount": order.BonusAmount,
			"total_credit": order.TotalCredit,
		})
		return
	}

	// 完成充值订单
	err = BalanceSvc.CompleteRechargeOrder(req.OutTradeNo, req.TradeNo)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "处理充值订单失败: " + err.Error()})
		return
	}

	// 重新获取订单信息
	order, _ = BalanceSvc.GetRechargeOrder(req.OutTradeNo)

	c.JSON(200, gin.H{
		"success":      true,
		"recharge_no":  order.RechargeNo,
		"amount":       order.Amount,
		"bonus_amount": order.BonusAmount,
		"total_credit": order.TotalCredit,
	})
}

// ==========================================
//         充值订单 - 支付宝当面付 API
// ==========================================

// AlipayCreateRechargePayment 创建充值订单的支付宝当面付订单
func AlipayCreateRechargePayment(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "余额服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		RechargeNo string `json:"recharge_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取充值订单
	order, err := BalanceSvc.GetRechargeOrder(req.RechargeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "充值订单不存在"})
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

	// 检查支付宝配置
	alipayCfg := &config.GlobalConfig.PaymentConfig.AlipayF2F
	if !alipayCfg.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "支付宝支付未启用"})
		return
	}

	// 创建支付宝服务
	alipaySvc := service.NewAlipayService(alipayCfg)

	// 使用实际支付金额
	payAmount := order.PayAmount
	if payAmount <= 0 {
		payAmount = order.Amount
	}

	// 创建支付宝订单，获取二维码
	description := "余额充值 - " + order.RechargeNo
	qrCode, err := alipaySvc.CreatePreOrder(order.RechargeNo, payAmount, description)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建支付宝订单失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"qr_code": qrCode,
	})
}

// AlipayRechargeQueryStatus 查询充值订单支付宝支付状态
func AlipayRechargeQueryStatus(c *gin.Context) {
	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	rechargeNo := c.Param("recharge_no")

	// 获取充值订单
	order, err := BalanceSvc.GetRechargeOrder(rechargeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	if order.UserID != userID {
		c.JSON(403, gin.H{"success": false, "error": "无权查询此订单"})
		return
	}

	// 如果订单已完成，直接返回
	if order.Status == 1 {
		c.JSON(200, gin.H{
			"success":      true,
			"paid":         true,
			"recharge_no":  order.RechargeNo,
			"amount":       order.Amount,
			"bonus_amount": order.BonusAmount,
			"total_credit": order.TotalCredit,
		})
		return
	}

	// 检查支付宝配置
	alipayCfg := &config.GlobalConfig.PaymentConfig.AlipayF2F
	if !alipayCfg.Enabled {
		c.JSON(200, gin.H{"success": true, "paid": false})
		return
	}

	// 查询支付宝订单状态
	alipaySvc := service.NewAlipayService(alipayCfg)
	paid, tradeNo, err := alipaySvc.QueryOrder(rechargeNo)
	if err != nil {
		c.JSON(200, gin.H{"success": true, "paid": false})
		return
	}

	// 如果已支付，完成充值订单
	if paid {
		err = BalanceSvc.CompleteRechargeOrder(rechargeNo, tradeNo)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "处理订单失败"})
			return
		}
		order, _ = BalanceSvc.GetRechargeOrder(rechargeNo)
	}

	c.JSON(200, gin.H{
		"success":      true,
		"paid":         paid,
		"recharge_no":  order.RechargeNo,
		"amount":       order.Amount,
		"bonus_amount": order.BonusAmount,
		"total_credit": order.TotalCredit,
	})
}

// ==========================================
//         充值订单 - 微信支付 API
// ==========================================

// WechatCreateRechargePayment 创建充值订单的微信支付订单
func WechatCreateRechargePayment(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "余额服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		RechargeNo string `json:"recharge_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取充值订单
	order, err := BalanceSvc.GetRechargeOrder(req.RechargeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "充值订单不存在"})
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

	// 检查微信支付配置
	wechatCfg := &config.GlobalConfig.PaymentConfig.WechatPay
	if !wechatCfg.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "微信支付未启用"})
		return
	}

	// 创建微信支付服务
	wechatSvc := service.NewWechatPayService(wechatCfg)

	// 使用实际支付金额
	payAmount := order.PayAmount
	if payAmount <= 0 {
		payAmount = order.Amount
	}

	// 创建微信支付订单，获取二维码
	description := "余额充值"
	qrCode, err := wechatSvc.CreateNativeOrder(order.RechargeNo, payAmount, description)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建微信支付订单失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"qr_code": qrCode,
	})
}

// WechatRechargeQueryStatus 查询充值订单微信支付状态
func WechatRechargeQueryStatus(c *gin.Context) {
	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	rechargeNo := c.Param("recharge_no")

	// 获取充值订单
	order, err := BalanceSvc.GetRechargeOrder(rechargeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	if order.UserID != userID {
		c.JSON(403, gin.H{"success": false, "error": "无权查询此订单"})
		return
	}

	// 如果订单已完成，直接返回
	if order.Status == 1 {
		c.JSON(200, gin.H{
			"success":      true,
			"paid":         true,
			"recharge_no":  order.RechargeNo,
			"amount":       order.Amount,
			"bonus_amount": order.BonusAmount,
			"total_credit": order.TotalCredit,
		})
		return
	}

	// 检查微信支付配置
	wechatCfg := &config.GlobalConfig.PaymentConfig.WechatPay
	if !wechatCfg.Enabled {
		c.JSON(200, gin.H{"success": true, "paid": false})
		return
	}

	// 查询微信支付订单状态
	wechatSvc := service.NewWechatPayService(wechatCfg)
	paid, tradeNo, err := wechatSvc.QueryOrder(rechargeNo)
	if err != nil {
		c.JSON(200, gin.H{"success": true, "paid": false})
		return
	}

	// 如果已支付，完成充值订单
	if paid {
		err = BalanceSvc.CompleteRechargeOrder(rechargeNo, tradeNo)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "处理订单失败"})
			return
		}
		order, _ = BalanceSvc.GetRechargeOrder(rechargeNo)
	}

	c.JSON(200, gin.H{
		"success":      true,
		"paid":         paid,
		"recharge_no":  order.RechargeNo,
		"amount":       order.Amount,
		"bonus_amount": order.BonusAmount,
		"total_credit": order.TotalCredit,
	})
}

// ==========================================
//         充值订单 - PayPal API
// ==========================================

// PayPalCreateRechargePayment 创建充值订单的 PayPal 支付
func PayPalCreateRechargePayment(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "余额服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		RechargeNo string `json:"recharge_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取充值订单
	order, err := BalanceSvc.GetRechargeOrder(req.RechargeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "充值订单不存在"})
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

	// 检查 PayPal 配置
	paypalConfig := &config.GlobalConfig.PaymentConfig.PayPal
	if !paypalConfig.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "PayPal 支付未启用"})
		return
	}

	// 创建 PayPal 服务
	paypalSvc := service.NewPayPalService(paypalConfig)

	// 使用实际支付金额
	payAmount := order.PayAmount
	if payAmount <= 0 {
		payAmount = order.Amount
	}

	// 创建 PayPal 订单
	description := "余额充值 - " + order.RechargeNo
	paypalOrder, err := paypalSvc.CreateOrder(order.RechargeNo, payAmount, description)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建 PayPal 订单失败: " + err.Error()})
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

// PayPalCaptureRechargePayment 捕获充值订单的 PayPal 支付
func PayPalCaptureRechargePayment(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "余额服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		RechargeNo    string `json:"recharge_no" binding:"required"`
		PayPalOrderID string `json:"paypal_order_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取充值订单
	order, err := BalanceSvc.GetRechargeOrder(req.RechargeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "充值订单不存在"})
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

	// 检查 PayPal 配置
	paypalConfig := &config.GlobalConfig.PaymentConfig.PayPal
	if !paypalConfig.Enabled {
		c.JSON(400, gin.H{"success": false, "error": "PayPal 支付未启用"})
		return
	}

	// 创建 PayPal 服务
	paypalSvc := service.NewPayPalService(paypalConfig)

	// 捕获支付
	captureResult, err := paypalSvc.CaptureOrder(req.PayPalOrderID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "捕获支付失败: " + err.Error()})
		return
	}

	// 检查支付状态
	if captureResult.Status != "COMPLETED" {
		c.JSON(400, gin.H{"success": false, "error": "支付未完成"})
		return
	}

	// 完成充值订单
	err = BalanceSvc.CompleteRechargeOrder(req.RechargeNo, req.PayPalOrderID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "处理充值订单失败: " + err.Error()})
		return
	}

	// 重新获取订单信息
	order, _ = BalanceSvc.GetRechargeOrder(req.RechargeNo)

	c.JSON(200, gin.H{
		"success":      true,
		"recharge_no":  order.RechargeNo,
		"amount":       order.Amount,
		"bonus_amount": order.BonusAmount,
		"total_credit": order.TotalCredit,
	})
}

// ==========================================
//         充值订单 - Stripe API
// ==========================================

// StripeCreateRechargeCheckoutSession 创建充值订单的 Stripe Checkout Session
func StripeCreateRechargeCheckoutSession(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "余额服务未初始化"})
		return
	}

	if StripeSvc == nil || !StripeSvc.IsEnabled() {
		c.JSON(400, gin.H{"success": false, "error": "Stripe 支付未启用"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		RechargeNo string `json:"recharge_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取充值订单
	order, err := BalanceSvc.GetRechargeOrder(req.RechargeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "充值订单不存在"})
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

	// 使用实际支付金额
	payAmount := order.PayAmount
	if payAmount <= 0 {
		payAmount = order.Amount
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

	// 创建 Stripe Checkout Session
	productName := "余额充值"
	successURL := baseURL + "/payment/result?type=recharge&recharge_no=" + order.RechargeNo + "&status=success"
	cancelURL := baseURL + "/payment?type=recharge&recharge_no=" + order.RechargeNo

	session, err := StripeSvc.CreateCheckoutSession(
		order.RechargeNo,
		int64(payAmount*100), // 转换为分
		productName,
		successURL,
		cancelURL,
	)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建 Stripe 订单失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"session_id": session.ID,
		"url":        session.URL,
	})
}

// ==========================================
//         充值订单 - USDT API
// ==========================================

// USDTCreateRechargePayment 创建充值订单的 USDT 支付
func USDTCreateRechargePayment(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "余额服务未初始化"})
		return
	}

	if USDTSvc == nil || !USDTSvc.IsEnabled() {
		c.JSON(400, gin.H{"success": false, "error": "USDT 支付未启用"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		RechargeNo string `json:"recharge_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取充值订单
	order, err := BalanceSvc.GetRechargeOrder(req.RechargeNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "充值订单不存在"})
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

	// 使用实际支付金额
	payAmount := order.PayAmount
	if payAmount <= 0 {
		payAmount = order.Amount
	}

	// 创建 USDT 支付
	paymentReq := &service.USDTPaymentRequest{
		OrderNo:     order.RechargeNo,
		Amount:      payAmount,
		Currency:    "CNY",
		Description: "余额充值",
	}

	payment, err := USDTSvc.CreatePayment(paymentReq)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建 USDT 订单失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":        true,
		"payment_id":     payment.PaymentID,
		"wallet_address": payment.WalletAddress,
		"amount_usdt":    payment.Amount,
		"network":        payment.Network,
	})
}

// USDTGetRechargePaymentStatus 获取充值订单的 USDT 支付状态
func USDTGetRechargePaymentStatus(c *gin.Context) {
	if USDTSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "USDT 服务未初始化"})
		return
	}

	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "余额服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	paymentID := c.Param("payment_id")

	// paymentID 就是 rechargeNo
	order, err := BalanceSvc.GetRechargeOrder(paymentID)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "充值订单不存在"})
		return
	}

	if order.UserID != userID {
		c.JSON(403, gin.H{"success": false, "error": "无权查询此订单"})
		return
	}

	// 如果订单已完成，直接返回
	if order.Status == 1 {
		c.JSON(200, gin.H{
			"success":      true,
			"status":       "completed",
			"recharge_no":  order.RechargeNo,
			"amount":       order.Amount,
			"bonus_amount": order.BonusAmount,
			"total_credit": order.TotalCredit,
		})
		return
	}

	// 查询 USDT 支付状态
	status, err := USDTSvc.GetPaymentStatus(paymentID)
	if err != nil {
		c.JSON(200, gin.H{
			"success":     true,
			"status":      "waiting",
			"recharge_no": order.RechargeNo,
		})
		return
	}

	c.JSON(200, gin.H{
		"success":     true,
		"status":      status.Status,
		"recharge_no": order.RechargeNo,
		"tx_hash":     status.TxHash,
	})
}
