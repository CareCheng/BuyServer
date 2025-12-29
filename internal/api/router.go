// Package api 提供 HTTP API 处理器
// router.go - 路由注册
package api

import (
	"user-frontend/internal/config"
	"user-frontend/internal/static"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	// 全局安全中间件
	r.Use(SecurityHeadersMiddleware())
	r.Use(IPBlacklistMiddleware())
	r.Use(RateLimitMiddleware())

	// 启动安全清理任务
	StartSecurityCleanupTask()

	// 静态文件服务（自动选择嵌入式或外部模式）
	// 注意：SetupStaticRoutes 内部已处理 /product-files 和 /uploads 路由
	static.SetupStaticRoutes(r)

	// CSRF令牌API
	r.GET("/api/csrf-token", GetCSRFToken)

	// 首页配置API（公开）
	r.GET("/api/homepage/config", GetPublicHomepageConfig)

	// 注册各模块路由
	registerUserRoutes(r)
	registerOrderRoutes(r)
	registerPaymentRoutes(r)
	registerProductRoutes(r)
	registerSupportRoutes(r)
	registerAdminRoutes(r, cfg)
	registerSPARoutes(r, cfg)
}

// registerStaticRoutes 注册静态文件路由（已废弃，使用 static.SetupStaticRoutes）
func registerStaticRoutes(r *gin.Engine) {
	// 此函数保留用于兼容，实际逻辑已移至 static 包
	static.SetupStaticRoutes(r)
}

// registerUserRoutes 注册用户相关路由
func registerUserRoutes(r *gin.Engine) {
	userAPI := r.Group("/api/user")
	{
		// 认证相关
		userAPI.POST("/register", UserRegister)
		userAPI.POST("/login", UserLogin)
		userAPI.POST("/logout", UserLogout)
		userAPI.GET("/info", AuthRequired(), UserInfo)
		userAPI.PUT("/info", AuthRequired(), UpdateUserInfo)
		userAPI.POST("/password", AuthRequired(), UpdatePassword)
		userAPI.GET("/orders", AuthRequired(), UserOrders)

		// 2FA相关
		userAPI.POST("/2fa/enable", AuthRequired(), Enable2FA)
		userAPI.POST("/2fa/disable", AuthRequired(), Disable2FA)
		userAPI.GET("/2fa/status", AuthRequired(), Get2FAStatus)
		userAPI.POST("/2fa/generate", AuthRequired(), Generate2FASecret)
		userAPI.POST("/2fa/preference", AuthRequired(), Set2FAPreference)
		userAPI.GET("/2fa/info", Get2FAInfo)
		userAPI.POST("/2fa/verify_login", Verify2FALogin)
		userAPI.POST("/2fa/enable_email", AuthRequired(), Enable2FAEmail)
		userAPI.POST("/2fa/verify_totp", AuthRequired(), VerifyTOTP)

		// 邮箱验证
		userAPI.POST("/email/send_code", SendEmailCode)
		userAPI.POST("/email/verify", VerifyEmailCode)
		userAPI.POST("/email/verify_only", VerifyEmailCodeOnly) // 仅验证不消耗
		userAPI.GET("/email/code_length", GetEmailCodeLength)   // 获取验证码长度
		userAPI.POST("/email/bind", AuthRequired(), BindEmail)

		// 忘记密码
		userAPI.POST("/forgot/check", ForgotPasswordCheck)
		userAPI.POST("/forgot/verify", ForgotPasswordVerify)
		userAPI.POST("/forgot/reset", ForgotPasswordReset)

		// 余额系统
		userAPI.GET("/balance", AuthRequired(), GetMyBalance)
		userAPI.GET("/balance/config", GetBalanceConfigPublic) // 公开的余额配置（限制信息）
		userAPI.GET("/balance/logs", AuthRequired(), GetMyBalanceLogs)
		userAPI.POST("/balance/recharge", AuthRequired(), CreateRechargeOrder)
		userAPI.GET("/balance/recharge/orders", AuthRequired(), GetMyRechargeOrders)
		userAPI.GET("/balance/recharge/:recharge_no", AuthRequired(), GetRechargeOrderDetail)
		userAPI.POST("/balance/recharge/:recharge_no/cancel", AuthRequired(), CancelRechargeOrder)
		userAPI.GET("/balance/promos", GetActiveRechargePromos)                       // 获取有效的充值优惠活动
		userAPI.POST("/balance/promo/calculate", AuthRequired(), CalculateRechargePromo) // 计算充值优惠
		userAPI.GET("/balance/promo/all", AuthRequired(), GetAllApplicablePromos)     // 获取所有适用的优惠方案

		// 支付密码
		userAPI.GET("/pay-password/status", AuthRequired(), GetPayPasswordStatus)
		userAPI.POST("/pay-password/set", AuthRequired(), SetPayPassword)
		userAPI.POST("/pay-password/update", AuthRequired(), UpdatePayPassword)
		userAPI.POST("/pay-password/reset", AuthRequired(), ResetPayPassword)
		userAPI.POST("/pay-password/verify", AuthRequired(), VerifyPayPassword)
		userAPI.POST("/pay-password/send-reset-code", AuthRequired(), SendResetPayPasswordCode)

		// 积分系统
		userAPI.GET("/points", AuthRequired(), GetMyPoints)
		userAPI.GET("/points/logs", AuthRequired(), GetPointsLogs)
		userAPI.GET("/points/exchange/list", AuthRequired(), GetExchangeList)
		userAPI.POST("/points/exchange/coupon", AuthRequired(), ExchangeCoupon)
		userAPI.GET("/points/exchanges", AuthRequired(), GetMyExchanges)

		// 购物车
		userAPI.GET("/cart", AuthRequired(), GetCart)
		userAPI.POST("/cart", AuthRequired(), AddToCart)
		userAPI.POST("/cart/add", AuthRequired(), AddToCart)
		userAPI.PUT("/cart/:id", AuthRequired(), UpdateCartItem)
		userAPI.DELETE("/cart/:id", AuthRequired(), RemoveFromCart)
		userAPI.DELETE("/cart", AuthRequired(), ClearCart)
		userAPI.GET("/cart/count", AuthRequired(), GetCartCount)
		userAPI.POST("/cart/validate", AuthRequired(), ValidateCart)

		// 收藏
		userAPI.POST("/favorite", AuthRequired(), AddFavorite)
		userAPI.DELETE("/favorite/:product_id", AuthRequired(), RemoveFavorite)
		userAPI.GET("/favorites", AuthRequired(), GetFavorites)
		userAPI.GET("/favorite/:product_id/check", AuthRequired(), CheckFavorite)
		userAPI.GET("/favorites/count", AuthRequired(), GetFavoriteCount)

		// 发票
		userAPI.GET("/invoices", AuthRequired(), GetMyInvoices)
		userAPI.GET("/invoice/:invoice_no", AuthRequired(), GetInvoiceDetail)
		userAPI.POST("/invoice", AuthRequired(), ApplyInvoice)
		userAPI.POST("/invoice/:invoice_no/cancel", AuthRequired(), CancelInvoice)
		userAPI.GET("/invoice/titles", AuthRequired(), GetMyInvoiceTitles)
		userAPI.POST("/invoice/title", AuthRequired(), SaveInvoiceTitle)
		userAPI.DELETE("/invoice/title/:id", AuthRequired(), DeleteInvoiceTitle)

		// 设备管理
		userAPI.GET("/devices", AuthRequired(), GetLoginDevices)
		userAPI.DELETE("/device/:id", AuthRequired(), RemoveLoginDevice)
		userAPI.POST("/devices/remove-all", AuthRequired(), RemoveAllOtherDevices)
		userAPI.GET("/login-history", AuthRequired(), GetLoginHistory)

		// 登录提醒
		userAPI.GET("/login-alerts", AuthRequired(), GetLoginAlerts)
		userAPI.GET("/login-alerts/unread-count", AuthRequired(), GetUnacknowledgedAlertCount)
		userAPI.POST("/login-alert/:id/acknowledge", AuthRequired(), AcknowledgeLoginAlert)
		userAPI.GET("/login-locations", AuthRequired(), GetLoginLocations)
		userAPI.POST("/login-location/:id/trust", AuthRequired(), TrustLoginLocation)
		userAPI.DELETE("/login-location/:id", AuthRequired(), RemoveLoginLocation)

		// 卡密/续费
		userAPI.GET("/kamis", AuthRequired(), GetUserKamis)
		userAPI.GET("/kamis/expiring", AuthRequired(), GetExpiringKamis)
		userAPI.GET("/kamis/expired", AuthRequired(), GetExpiredKamis)
		userAPI.POST("/renewal/remind", AuthRequired(), RequestRenewalReminder)
		userAPI.GET("/renewal/history", AuthRequired(), GetRenewalHistory)
		userAPI.GET("/renewal/stats", AuthRequired(), GetRenewalStats)

		// 账户注销
		userAPI.POST("/account/delete", AuthRequired(), RequestAccountDeletion)
		userAPI.POST("/account/delete/cancel", AuthRequired(), CancelAccountDeletion)
		userAPI.GET("/account/delete/status", AuthRequired(), GetAccountDeletionStatus)

		// 商品评价
		userAPI.GET("/reviews", AuthRequired(), GetUserReviews)

		// 敏感操作验证
		userAPI.POST("/sensitive/request", AuthRequired(), RequestSensitiveVerification)
		userAPI.POST("/sensitive/verify", AuthRequired(), VerifySensitiveOperation)
		userAPI.POST("/sensitive/email", AuthRequired(), SendSensitiveVerificationEmail)
		userAPI.POST("/sensitive/password", AuthRequired(), UpdatePasswordWithVerification)
		userAPI.POST("/sensitive/bind-email", AuthRequired(), BindEmailWithVerification)
		userAPI.POST("/sensitive/disable-2fa", AuthRequired(), Disable2FAWithVerification)

		// 数据导出
		userAPI.GET("/export/orders", AuthRequired(), UserExportOrders)

		// 我的优惠券
		userAPI.GET("/coupons", AuthRequired(), GetMyCoupons)
		userAPI.GET("/coupons/available", AuthRequired(), GetMyAvailableCoupons)
		userAPI.GET("/coupon/:id", AuthRequired(), GetMyCouponDetail)
		userAPI.GET("/coupons/count", AuthRequired(), GetMyCouponCount)
	}
}


// registerOrderRoutes 注册订单相关路由
func registerOrderRoutes(r *gin.Engine) {
	orderAPI := r.Group("/api/order")
	{
		orderAPI.POST("/create", AuthRequired(), CreateOrder)
		orderAPI.GET("/detail/:order_no", AuthRequired(), OrderDetail)
		orderAPI.POST("/cancel", AuthRequired(), CancelOrder)
		orderAPI.POST("/pay/balance", AuthRequired(), PayOrderWithBalance)
	}

	// 优惠券验证
	r.POST("/api/coupon/validate", AuthRequired(), ValidateCoupon)

	// 订单查询（未登录，通过订单号+邮箱）
	r.POST("/api/order/query", QueryOrderPublic)
}

// registerPaymentRoutes 注册支付相关路由
func registerPaymentRoutes(r *gin.Engine) {
	// PayPal支付
	paypalAPI := r.Group("/api/paypal")
	{
		paypalAPI.POST("/create", AuthRequired(), PayPalCreatePayment)
		paypalAPI.POST("/capture", AuthRequired(), PayPalCapturePayment)
		// 充值订单支付
		paypalAPI.POST("/recharge/create", AuthRequired(), PayPalCreateRechargePayment)
		paypalAPI.POST("/recharge/capture", AuthRequired(), PayPalCaptureRechargePayment)
	}
	r.GET("/paypal/return", PayPalReturn)
	r.GET("/paypal/cancel", PayPalCancel)

	// 支付宝当面付
	alipayAPI := r.Group("/api/alipay")
	{
		alipayAPI.POST("/create", AuthRequired(), AlipayCreatePayment)
		alipayAPI.GET("/status/:order_no", AuthRequired(), AlipayQueryStatus)
		// 充值订单支付
		alipayAPI.POST("/recharge/create", AuthRequired(), AlipayCreateRechargePayment)
		alipayAPI.GET("/recharge/status/:recharge_no", AuthRequired(), AlipayRechargeQueryStatus)
	}
	r.POST("/alipay/notify", AlipayNotify)

	// 微信支付
	wechatAPI := r.Group("/api/wechat")
	{
		wechatAPI.POST("/create", AuthRequired(), WechatCreatePayment)
		wechatAPI.GET("/status/:order_no", AuthRequired(), WechatQueryStatus)
		// 充值订单支付
		wechatAPI.POST("/recharge/create", AuthRequired(), WechatCreateRechargePayment)
		wechatAPI.GET("/recharge/status/:recharge_no", AuthRequired(), WechatRechargeQueryStatus)
	}
	r.POST("/wechat/notify", WechatNotify)

	// 易支付
	yipayAPI := r.Group("/api/yipay")
	{
		yipayAPI.POST("/create", AuthRequired(), YiPayCreatePayment)
		yipayAPI.POST("/callback", YiPayCallback)
		// 充值订单支付
		yipayAPI.POST("/recharge/create", AuthRequired(), YiPayCreateRechargePayment)
		yipayAPI.POST("/recharge/callback", YiPayRechargeCallback)
	}
	r.POST("/yipay/notify", YiPayNotify)
	r.POST("/yipay/recharge/notify", YiPayRechargeNotify)
	r.GET("/yipay/return", YiPayReturn)

	// Stripe支付
	stripeAPI := r.Group("/api/stripe")
	{
		stripeAPI.GET("/config", StripeGetConfig)
		stripeAPI.POST("/create", AuthRequired(), StripeCreateCheckoutSession)
		stripeAPI.GET("/verify/:session_id", AuthRequired(), StripeVerifyPayment)
		// 充值订单支付
		stripeAPI.POST("/recharge/create", AuthRequired(), StripeCreateRechargeCheckoutSession)
	}
	r.POST("/stripe/webhook", StripeWebhook)

	// USDT支付
	usdtAPI := r.Group("/api/usdt")
	{
		usdtAPI.GET("/config", USDTGetConfig)
		usdtAPI.POST("/create", AuthRequired(), USDTCreatePayment)
		usdtAPI.GET("/status/:payment_id", AuthRequired(), USDTGetPaymentStatus)
		// 充值订单支付
		usdtAPI.POST("/recharge/create", AuthRequired(), USDTCreateRechargePayment)
		usdtAPI.GET("/recharge/status/:payment_id", AuthRequired(), USDTGetRechargePaymentStatus)
	}
	r.POST("/usdt/webhook", USDTWebhook)

	// 支付方式查询
	r.GET("/api/payment/methods", GetPaymentMethods)
}

// registerProductRoutes 注册商品相关路由
func registerProductRoutes(r *gin.Engine) {
	// 商品API（公开）
	r.GET("/api/products", GetProducts)
	r.GET("/api/product/:id", GetProduct)
	r.GET("/api/product/:id/images", GetProductImages)
	r.GET("/api/product/:id/detail-file", GetProductDetailFile)
	r.GET("/api/categories", GetCategories)

	// 公告API（公开）
	r.GET("/api/announcements", GetAnnouncements)

	// FAQ API（公开）
	r.GET("/api/faq/categories", GetFAQCategories)
	r.GET("/api/faq/list", GetFAQList)
	r.GET("/api/faq/detail/:id", GetFAQDetail)
	r.GET("/api/faq/search", SearchFAQs)
	r.GET("/api/faq/hot", GetHotFAQs)
	r.POST("/api/faq/feedback/:id", SubmitFAQFeedback)

	// 商品评价（公开）
	r.GET("/api/product/:id/reviews", GetProductReviews)
	r.GET("/api/product/:id/review-stats", GetProductReviewStats)
	r.POST("/api/review", AuthRequired(), CreateProductReview)
	r.GET("/api/order/:order_no/can-review", AuthRequired(), CheckCanReview)

	// 发票配置（公开）
	r.GET("/api/invoice/config", GetInvoiceConfig)

	// 健康检查
	r.GET("/health", HealthCheck)
	r.GET("/api/health", HealthCheck)

	// 验证码
	r.GET("/api/captcha", CaptchaHandler)
	r.POST("/api/captcha/verify", VerifyCaptcha)
}


// registerSupportRoutes 注册客服支持相关路由
func registerSupportRoutes(r *gin.Engine) {
	// 客服配置（公开）
	r.GET("/api/support/config", GetSupportConfig)

	// 用户端工单 API
	supportAPI := r.Group("/api/support")
	{
		supportAPI.POST("/ticket", OptionalAuth(), CreateTicket)
		supportAPI.GET("/tickets/guest", GetGuestTickets)
		supportAPI.GET("/ticket/:ticket_no", OptionalAuth(), GetTicketDetail)
		supportAPI.POST("/ticket/:ticket_no/reply", OptionalAuth(), ReplyTicket)
		supportAPI.POST("/ticket/:ticket_no/close", OptionalAuth(), CloseTicket)
		supportAPI.POST("/ticket/:ticket_no/upload", OptionalAuth(), UploadTicketAttachment)
		supportAPI.GET("/tickets", AuthRequired(), GetUserTickets)
		supportAPI.POST("/ticket/:ticket_no/rate", OptionalAuth(), RateTicket)
		supportAPI.GET("/ticket/:ticket_no/rating", OptionalAuth(), GetTicketRating)
	}

	// 实时聊天 API
	chatAPI := r.Group("/api/chat")
	{
		chatAPI.POST("/start", OptionalAuth(), StartLiveChat)
		chatAPI.POST("/:session_id/send", OptionalAuth(), SendChatMessage)
		chatAPI.GET("/:session_id/messages", OptionalAuth(), GetChatMessages)
		chatAPI.POST("/:session_id/end", OptionalAuth(), EndLiveChat)
	}

	// WebSocket 实时通信
	r.GET("/ws/user", OptionalAuth(), WSUserConnect)
	r.GET("/ws/staff", WSStaffConnect)

	// 客服后台 API
	staffAPI := r.Group("/api/staff")
	{
		staffAPI.POST("/login", StaffLogin)
		staffAPI.POST("/logout", StaffLogout)
		staffAPI.POST("/2fa/verify", StaffVerify2FA)

		// 需要客服认证的 API
		staffAPI.Use(StaffAuthRequired())
		{
			staffAPI.GET("/info", StaffInfo)
			// 二步验证管理
			staffAPI.GET("/2fa/status", StaffGet2FAStatus)
			staffAPI.POST("/2fa/generate", StaffGenerate2FASecret)
			staffAPI.POST("/2fa/enable", StaffEnable2FA)
			staffAPI.POST("/2fa/disable", StaffDisable2FA)
			// 工单管理
			staffAPI.GET("/tickets", StaffGetTickets)
			staffAPI.GET("/ticket/:ticket_no", StaffGetTicketDetail)
			staffAPI.POST("/ticket/:ticket_no/reply", StaffReplyTicket)
			staffAPI.PUT("/ticket/:ticket_no/status", StaffUpdateTicketStatus)
			staffAPI.POST("/ticket/:ticket_no/assign", StaffAssignTicket)
			staffAPI.GET("/tickets/stats", StaffGetTicketStats)
			// 工单转接与合并
			staffAPI.POST("/ticket/:ticket_no/transfer", StaffTransferTicket)
			staffAPI.POST("/tickets/merge", StaffMergeTickets)
			staffAPI.GET("/staff/online", StaffGetOnlineStaff)
			staffAPI.GET("/ticket/:ticket_no/attachments", StaffGetTicketAttachments)
			staffAPI.POST("/ticket/:ticket_no/upload", StaffUploadTicketAttachment)
			// 实时聊天管理
			staffAPI.GET("/chats/waiting", StaffGetWaitingChats)
			staffAPI.POST("/chat/:chat_id/accept", StaffAcceptChat)
			staffAPI.POST("/chat/:chat_id/send", StaffSendChatMessage)
			staffAPI.GET("/chat/:chat_id/messages", StaffGetChatMessages)
			staffAPI.POST("/chat/:chat_id/end", StaffEndChat)
			// 知识库
			staffAPI.GET("/knowledge/categories", StaffGetKnowledgeCategories)
			staffAPI.GET("/knowledge/articles", StaffGetKnowledgeArticles)
			staffAPI.GET("/knowledge/article/:id", StaffGetKnowledgeArticle)
			staffAPI.GET("/knowledge/search", StaffSearchKnowledge)
			staffAPI.GET("/knowledge/hot", StaffGetHotKnowledge)
			staffAPI.POST("/knowledge/:id/use", StaffUseKnowledge)
		}
	}
}


// registerAdminRoutes 注册管理后台路由
func registerAdminRoutes(r *gin.Engine, cfg *config.Config) {
	adminSuffix := cfg.ServerConfig.AdminSuffix

	// 初始化设置检查（无需认证）
	r.GET("/"+adminSuffix+"/check-setup", CheckInitialSetup)
	r.POST("/"+adminSuffix+"/setup", SetInitialPassword)

	// 管理员登录相关
	r.POST("/"+adminSuffix+"/login", AdminLogin)
	r.POST("/"+adminSuffix+"/totp", AdminVerifyTOTP)
	r.POST("/"+adminSuffix+"/logout", AdminLogout)

	// 管理员信息API
	r.GET("/api/admin/info", AdminAuthRequired(), AdminInfo)

	// 获取管理后台入口配置
	r.GET("/api/admin/suffix", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true, "suffix": adminSuffix})
	})

	// 管理API
	adminAPI := r.Group("/api/admin")
	adminAPI.Use(IPWhitelistMiddleware())
	adminAPI.Use(AdminAuthRequired())
	{
		// 仪表盘
		adminAPI.GET("/dashboard", AdminDashboard)

		// 商品管理
		registerAdminProductRoutes(adminAPI)

		// 订单管理
		registerAdminOrderRoutes(adminAPI)

		// 用户管理
		registerAdminUserRoutes(adminAPI)

		// 系统设置
		registerAdminSettingsRoutes(adminAPI)

		// 内容管理
		registerAdminContentRoutes(adminAPI)

		// 客服系统管理
		registerAdminSupportRoutes(adminAPI)

		// 系统管理
		registerAdminSystemRoutes(adminAPI)

		// 批量操作
		registerAdminBatchRoutes(adminAPI)
	}
}

// registerAdminProductRoutes 注册管理后台商品相关路由
func registerAdminProductRoutes(adminAPI *gin.RouterGroup) {
	adminAPI.GET("/products", AdminGetProducts)
	adminAPI.POST("/product", AdminCreateProduct)
	adminAPI.PUT("/product/:id", AdminUpdateProduct)
	adminAPI.DELETE("/product/:id", AdminDeleteProduct)
	adminAPI.POST("/product/:id/image", AdminUploadProductImage)
	adminAPI.DELETE("/product/:id/image", AdminDeleteProductImage)
	adminAPI.POST("/product/:id/detail-file", SaveProductDetailFile)
	adminAPI.POST("/product/:id/detail-image", UploadProductDetailImage)

	// 手动卡密管理
	adminAPI.POST("/product/:id/kami/import", AdminImportKami)
	adminAPI.GET("/product/:id/kami", AdminGetProductKamis)
	adminAPI.GET("/product/:id/kami/stats", AdminGetKamiStats)
	adminAPI.DELETE("/kami/:id", AdminDeleteKami)
	adminAPI.POST("/kami/:id/disable", AdminDisableKami)
	adminAPI.POST("/kami/:id/enable", AdminEnableKami)
	adminAPI.POST("/kami/batch-delete", AdminBatchDeleteKamis)
}

// registerAdminOrderRoutes 注册管理后台订单相关路由
func registerAdminOrderRoutes(adminAPI *gin.RouterGroup) {
	adminAPI.GET("/orders", AdminGetOrders)
	adminAPI.GET("/orders/search", AdminSearchOrders)
	adminAPI.GET("/order/:id", AdminGetOrder)
}

// registerAdminUserRoutes 注册管理后台用户相关路由
func registerAdminUserRoutes(adminAPI *gin.RouterGroup) {
	adminAPI.GET("/users", AdminGetUsers)
	adminAPI.PUT("/user/:id/status", AdminUpdateUserStatus)
}

// registerAdminSettingsRoutes 注册管理后台设置相关路由
func registerAdminSettingsRoutes(adminAPI *gin.RouterGroup) {
	// 系统设置
	adminAPI.GET("/settings", AdminGetSettings)
	adminAPI.POST("/settings", AdminSaveSettings)
	adminAPI.POST("/settings/security", AdminSaveSecuritySettings)

	// 数据库配置
	adminAPI.GET("/db/config", AdminGetDBConfig)
	adminAPI.POST("/db/config", AdminSaveDBConfig)
	adminAPI.POST("/db/test", AdminTestDBConnection)
	adminAPI.POST("/db/reset-key", AdminResetEncryptionKey)

	// 2FA设置
	adminAPI.POST("/2fa/enable", AdminEnable2FA)
	adminAPI.POST("/2fa/disable", AdminDisable2FA)
	adminAPI.GET("/2fa/status", AdminGet2FAStatus)
	adminAPI.POST("/2fa/generate", AdminGenerate2FASecret)
	adminAPI.POST("/2fa/verify", AdminVerify2FACode)

	// 支付配置
	adminAPI.GET("/payment/config", AdminGetPaymentConfig)
	adminAPI.POST("/payment/config", AdminSavePaymentConfig)

	// 邮箱配置
	adminAPI.GET("/email/config", AdminGetEmailConfig)
	adminAPI.POST("/email/config", AdminSaveEmailConfig)
	adminAPI.POST("/email/test", AdminTestEmail)

	// Stripe/USDT 测试
	adminAPI.POST("/stripe/test", StripeTestConnection)
	adminAPI.POST("/usdt/test", USDTTestConnection)
	adminAPI.POST("/usdt/confirm", AdminConfirmUSDTPayment)
}


// registerAdminContentRoutes 注册管理后台内容管理路由
func registerAdminContentRoutes(adminAPI *gin.RouterGroup) {
	// 首页配置管理
	adminAPI.GET("/homepage/config", AdminGetHomepageConfig)
	adminAPI.POST("/homepage/config", AdminUpdateHomepageConfig)
	adminAPI.GET("/homepage/templates", AdminGetTemplateList)
	adminAPI.GET("/homepage/template/default", AdminGetTemplateDefault)
	adminAPI.POST("/homepage/reset", AdminResetHomepage)

	// 公告管理
	adminAPI.GET("/announcements", AdminGetAnnouncements)
	adminAPI.POST("/announcement", AdminCreateAnnouncement)
	adminAPI.PUT("/announcement/:id", AdminUpdateAnnouncement)
	adminAPI.DELETE("/announcement/:id", AdminDeleteAnnouncement)

	// 分类管理
	adminAPI.GET("/categories", AdminGetCategories)
	adminAPI.POST("/category", AdminCreateCategory)
	adminAPI.PUT("/category/:id", AdminUpdateCategory)
	adminAPI.DELETE("/category/:id", AdminDeleteCategory)

	// 优惠券管理
	adminAPI.GET("/coupons", AdminGetCoupons)
	adminAPI.POST("/coupon", AdminCreateCoupon)
	adminAPI.PUT("/coupon/:id", AdminUpdateCoupon)
	adminAPI.DELETE("/coupon/:id", AdminDeleteCoupon)
	adminAPI.GET("/coupon/:id/usages", AdminGetCouponUsages)

	// FAQ管理
	adminAPI.GET("/faq/categories", AdminGetFAQCategories)
	adminAPI.POST("/faq/category", AdminCreateFAQCategory)
	adminAPI.PUT("/faq/category/:id", AdminUpdateFAQCategory)
	adminAPI.DELETE("/faq/category/:id", AdminDeleteFAQCategory)
	adminAPI.GET("/faqs", AdminGetFAQs)
	adminAPI.POST("/faq", AdminCreateFAQ)
	adminAPI.PUT("/faq/:id", AdminUpdateFAQ)
	adminAPI.DELETE("/faq/:id", AdminDeleteFAQ)

	// 商品评价管理
	adminAPI.GET("/reviews", AdminGetReviews)
	adminAPI.POST("/review/:id/reply", AdminReplyReview)
	adminAPI.PUT("/review/:id/status", AdminUpdateReviewStatus)
	adminAPI.DELETE("/review/:id", AdminDeleteReview)

	// 发票管理
	adminAPI.GET("/invoices", AdminGetInvoices)
	adminAPI.POST("/invoice/:invoice_no/issue", AdminIssueInvoice)
	adminAPI.POST("/invoice/:invoice_no/reject", AdminRejectInvoice)
	adminAPI.GET("/invoice/config", AdminGetInvoiceConfig)
	adminAPI.POST("/invoice/config", AdminSaveInvoiceConfig)
	adminAPI.GET("/invoice/stats", AdminGetInvoiceStats)

	// 余额管理
	adminAPI.GET("/balances", AdminGetBalances)
	adminAPI.GET("/balance/logs", AdminGetBalanceLogs)
	adminAPI.GET("/balance/recharge/orders", AdminGetRechargeOrders)
	adminAPI.POST("/balance/adjust", AdminAdjustBalance)
	adminAPI.POST("/balance/gift", AdminGiftBalance)
	adminAPI.GET("/balance/stats", AdminGetBalanceStats)
	adminAPI.GET("/balance/config", AdminGetBalanceConfig)
	adminAPI.POST("/balance/config", AdminSaveBalanceConfig)

	// 余额告警管理
	adminAPI.GET("/balance/alerts", AdminGetBalanceAlerts)
	adminAPI.GET("/balance/alert/:id", AdminGetBalanceAlertDetail)
	adminAPI.POST("/balance/alert/:id/handle", AdminHandleBalanceAlert)
	adminAPI.GET("/balance/alert/stats", AdminGetBalanceAlertStats)
	adminAPI.POST("/balance/alert/check-mismatch", AdminBatchCheckBalanceMismatch)
	adminAPI.POST("/balance/alert/clean", AdminCleanOldAlerts)

	// 充值优惠活动管理
	adminAPI.GET("/recharge-promos", AdminGetRechargePromos)
	adminAPI.POST("/recharge-promo", AdminCreateRechargePromo)
	adminAPI.PUT("/recharge-promo/:id", AdminUpdateRechargePromo)
	adminAPI.DELETE("/recharge-promo/:id", AdminDeleteRechargePromo)
	adminAPI.POST("/recharge-promo/:id/toggle", AdminToggleRechargePromoStatus)
	adminAPI.GET("/recharge-promo/usages", AdminGetRechargePromoUsages)
	adminAPI.GET("/recharge-promo/stats", AdminGetRechargePromoStats)

	// 积分管理
	adminAPI.GET("/points/users", AdminGetPointsUsers)
	adminAPI.GET("/points/logs", AdminGetPointsLogs)
	adminAPI.GET("/points/rules", AdminGetPointsRules)
	adminAPI.POST("/points/rule", AdminCreatePointsRule)
	adminAPI.PUT("/points/rule/:id", AdminUpdatePointsRule)
	adminAPI.DELETE("/points/rule/:id", AdminDeletePointsRule)
	adminAPI.POST("/points/adjust", AdminAdjustPoints)

	// 账户注销管理
	adminAPI.GET("/account/deletions", AdminGetDeletionRequests)
	adminAPI.POST("/account/deletion/:id/approve", AdminApproveDeletion)
	adminAPI.POST("/account/deletion/:id/reject", AdminRejectDeletion)
}

// registerAdminSupportRoutes 注册管理后台客服系统路由
func registerAdminSupportRoutes(adminAPI *gin.RouterGroup) {
	// 客服系统管理
	adminAPI.GET("/support/config", AdminGetSupportConfig)
	adminAPI.POST("/support/config", AdminSaveSupportConfig)
	adminAPI.GET("/support/staff", AdminGetStaffList)
	adminAPI.POST("/support/staff", AdminCreateStaff)
	adminAPI.PUT("/support/staff/:id", AdminUpdateStaff)
	adminAPI.DELETE("/support/staff/:id", AdminDeleteStaff)
	adminAPI.GET("/support/stats", AdminGetSupportStats)

	// 知识库管理
	adminAPI.GET("/knowledge/categories", AdminGetKnowledgeCategories)
	adminAPI.POST("/knowledge/category", AdminCreateKnowledgeCategory)
	adminAPI.PUT("/knowledge/category/:id", AdminUpdateKnowledgeCategory)
	adminAPI.DELETE("/knowledge/category/:id", AdminDeleteKnowledgeCategory)
	adminAPI.GET("/knowledge/articles", AdminGetKnowledgeArticles)
	adminAPI.POST("/knowledge/article", AdminCreateKnowledgeArticle)
	adminAPI.PUT("/knowledge/article/:id", AdminUpdateKnowledgeArticle)
	adminAPI.DELETE("/knowledge/article/:id", AdminDeleteKnowledgeArticle)

	// 工单模板管理
	adminAPI.GET("/ticket-templates", AdminGetTicketTemplates)
	adminAPI.POST("/ticket-template", AdminCreateTicketTemplate)
	adminAPI.PUT("/ticket-template/:id", AdminUpdateTicketTemplate)
	adminAPI.DELETE("/ticket-template/:id", AdminDeleteTicketTemplate)

	// 智能客服管理
	adminAPI.GET("/auto-reply/config", AdminGetAutoReplyConfig)
	adminAPI.POST("/auto-reply/config", AdminSaveAutoReplyConfig)
	adminAPI.GET("/auto-reply/rules", AdminGetAutoReplyRules)
	adminAPI.POST("/auto-reply/rule", AdminCreateAutoReplyRule)
	adminAPI.PUT("/auto-reply/rule/:id", AdminUpdateAutoReplyRule)
	adminAPI.DELETE("/auto-reply/rule/:id", AdminDeleteAutoReplyRule)
	adminAPI.GET("/auto-reply/logs", AdminGetAutoReplyLogs)
	adminAPI.GET("/auto-reply/stats", AdminGetAutoReplyStats)
}


// registerAdminSystemRoutes 注册管理后台系统管理路由
func registerAdminSystemRoutes(adminAPI *gin.RouterGroup) {
	// 操作日志
	adminAPI.GET("/logs", AdminGetOperationLogs)
	adminAPI.GET("/logs/dates", AdminGetLogDates)
	adminAPI.GET("/logs/config", AdminGetLogConfig)
	adminAPI.POST("/logs/config", AdminUpdateLogConfig)

	// 统计数据
	adminAPI.GET("/stats/chart", AdminGetStatsChart)

	// 数据库备份
	adminAPI.GET("/backups", AdminGetBackups)
	adminAPI.GET("/backup/info", AdminGetBackupInfo)
	adminAPI.POST("/backup", AdminCreateBackup)
	adminAPI.GET("/backup/:id/download", AdminDownloadBackup)
	adminAPI.DELETE("/backup/:id", AdminDeleteBackup)

	// IP黑名单管理
	adminAPI.GET("/blacklist", AdminGetBlacklist)
	adminAPI.DELETE("/blacklist/:ip", AdminRemoveFromBlacklist)
	adminAPI.DELETE("/blacklist", AdminClearBlacklist)

	// IP白名单管理
	adminAPI.GET("/whitelist", AdminGetWhitelist)
	adminAPI.POST("/whitelist", AdminSaveWhitelist)

	// 系统监控
	adminAPI.GET("/monitor/system", AdminGetSystemInfo)
	adminAPI.GET("/monitor/memory", AdminGetMemoryStats)
	adminAPI.GET("/monitor/database", AdminGetDatabaseStats)
	adminAPI.GET("/monitor/health", AdminGetHealthStatus)
	adminAPI.GET("/monitor/realtime", AdminGetRealtimeStats)
	adminAPI.GET("/monitor/overview", AdminGetMonitorOverview)

	// 角色权限管理
	adminAPI.GET("/roles", AdminGetRoles)
	adminAPI.GET("/role/:id", AdminGetRole)
	adminAPI.POST("/role", AdminCreateRole)
	adminAPI.PUT("/role/:id", AdminUpdateRole)
	adminAPI.DELETE("/role/:id", AdminDeleteRole)
	adminAPI.GET("/permissions", AdminGetPermissions)
	adminAPI.GET("/admins", AdminGetAdmins)
	adminAPI.GET("/admin/:id", AdminGetAdmin)
	adminAPI.POST("/admin", AdminCreateAdmin)
	adminAPI.PUT("/admin/:id", AdminUpdateAdmin)
	adminAPI.PUT("/admin/:id/password", AdminUpdateAdminPassword)
	adminAPI.DELETE("/admin/:id", AdminDeleteAdmin)
	adminAPI.GET("/my-permissions", AdminGetMyPermissions)

	// 定时任务管理
	adminAPI.GET("/tasks", AdminGetTasks)
	adminAPI.GET("/tasks/types", AdminGetTaskTypes)
	adminAPI.GET("/tasks/stats", AdminGetTaskStats)
	adminAPI.GET("/tasks/logs", AdminGetTaskLogs)
	adminAPI.POST("/task", AdminCreateTask)
	adminAPI.PUT("/task/:id", AdminUpdateTask)
	adminAPI.DELETE("/task/:id", AdminDeleteTask)
	adminAPI.POST("/task/:id/run", AdminRunTaskNow)
	adminAPI.POST("/task/:id/toggle", AdminToggleTaskStatus)

	// 操作撤销管理
	adminAPI.GET("/undo/operations", AdminGetUndoableOperations)
	adminAPI.GET("/undo/all", AdminGetAllUndoOperations)
	adminAPI.POST("/undo/:id", AdminUndoOperation)
	adminAPI.GET("/undo/config", AdminGetUndoConfig)
	adminAPI.POST("/undo/config", AdminSaveUndoConfig)
	adminAPI.GET("/undo/stats", AdminGetUndoStats)

	// 数据导出
	adminAPI.GET("/export/orders", AdminExportOrders)
	adminAPI.GET("/export/users", AdminExportUsers)
	adminAPI.GET("/export/logs", AdminExportLogs)
	adminAPI.GET("/export/login-history", AdminExportLoginHistory)
}

// registerAdminBatchRoutes 注册管理后台批量操作路由
func registerAdminBatchRoutes(adminAPI *gin.RouterGroup) {
	adminAPI.POST("/products/batch-delete", AdminBatchDeleteProducts)
	adminAPI.POST("/products/batch-status", AdminBatchUpdateProductStatus)
	adminAPI.POST("/users/batch-delete", AdminBatchDeleteUsers)
	adminAPI.POST("/users/batch-status", AdminBatchUpdateUserStatus)
	adminAPI.POST("/orders/batch-delete", AdminBatchDeleteOrders)
	adminAPI.POST("/coupons/batch-delete", AdminBatchDeleteCoupons)
	adminAPI.POST("/coupons/batch-status", AdminBatchUpdateCouponStatus)
	adminAPI.POST("/announcements/batch-delete", AdminBatchDeleteAnnouncements)
	adminAPI.POST("/categories/batch-delete", AdminBatchDeleteCategories)
}

// registerSPARoutes 注册 SPA 前端页面路由
func registerSPARoutes(r *gin.Engine, cfg *config.Config) {
	adminSuffix := cfg.ServerConfig.AdminSuffix

	// 用户前台页面
	r.GET("/", ServeReactPage("index.html"))
	r.GET("/login", ServeReactPage("login/index.html"))
	r.GET("/login/", ServeReactPage("login/index.html"))
	r.GET("/register", ServeReactPage("register/index.html"))
	r.GET("/register/", ServeReactPage("register/index.html"))
	r.GET("/forgot", ServeReactPage("forgot/index.html"))
	r.GET("/forgot/", ServeReactPage("forgot/index.html"))
	r.GET("/verify", ServeReactPage("verify/index.html"))
	r.GET("/verify/", ServeReactPage("verify/index.html"))
	r.GET("/products", ServeReactPage("products/index.html"))
	r.GET("/products/", ServeReactPage("products/index.html"))
	r.GET("/product", ServeReactPage("product/index.html"))
	r.GET("/product/", ServeReactPage("product/index.html"))
	r.GET("/faq", ServeReactPage("faq/index.html"))
	r.GET("/faq/", ServeReactPage("faq/index.html"))
	r.GET("/order/detail", ServeReactPage("order/detail/index.html"))
	r.GET("/order/detail/", ServeReactPage("order/detail/index.html"))
	r.GET("/user", ServeReactPage("user/index.html"))
	r.GET("/user/", ServeReactPage("user/index.html"))

	// 支付相关页面
	r.GET("/payment", ServeReactPage("payment/index.html"))
	r.GET("/payment/", ServeReactPage("payment/index.html"))
	r.GET("/payment/result", ServeReactPage("payment/result/index.html"))
	r.GET("/payment/result/", ServeReactPage("payment/result/index.html"))
	r.GET("/payment/cancel", ServeReactPage("payment/cancel/index.html"))
	r.GET("/payment/cancel/", ServeReactPage("payment/cancel/index.html"))
	r.GET("/payment/qrcode", ServeReactPage("payment/qrcode/index.html"))
	r.GET("/payment/qrcode/", ServeReactPage("payment/qrcode/index.html"))

	// 客服支持页面
	r.GET("/message", ServeReactPage("message/index.html"))
	r.GET("/message/", ServeReactPage("message/index.html"))
	r.GET("/message/ticket/detail", ServeReactPage("message/ticket/detail/index.html"))
	r.GET("/message/ticket/detail/", ServeReactPage("message/ticket/detail/index.html"))
	r.GET("/message/ticket/:ticket_no", ServeReactPage("message/ticket/index.html"))

	// 客服后台页面
	r.GET("/staff", ServeReactPage("staff/index.html"))
	r.GET("/staff/", ServeReactPage("staff/index.html"))
	r.GET("/staff/login", ServeReactPage("staff/login/index.html"))
	r.GET("/staff/login/", ServeReactPage("staff/login/index.html"))

	// 管理后台页面（使用动态后缀）
	r.GET("/"+adminSuffix, ServeReactPage("admin/index.html"))
	r.GET("/"+adminSuffix+"/", ServeReactPage("admin/index.html"))
	r.GET("/"+adminSuffix+"/login", ServeReactPage("admin/login/index.html"))
	r.GET("/"+adminSuffix+"/login/", ServeReactPage("admin/login/index.html"))
	r.GET("/"+adminSuffix+"/totp", ServeReactPage("admin/totp/index.html"))
	r.GET("/"+adminSuffix+"/totp/", ServeReactPage("admin/totp/index.html"))
	r.GET("/"+adminSuffix+"/setup", ServeReactPage("admin/setup/index.html"))
	r.GET("/"+adminSuffix+"/setup/", ServeReactPage("admin/setup/index.html"))
}

// ServeReactPage 返回React静态页面的处理函数
// 自动支持嵌入式和外部资源模式
func ServeReactPage(pagePath string) gin.HandlerFunc {
	return static.ServeEmbeddedPage(pagePath)
}
