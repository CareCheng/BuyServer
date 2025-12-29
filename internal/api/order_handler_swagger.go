package api

// ==================== 商品 API 文档注解 ====================

// GetProductList 获取商品列表
// @Summary      获取商品列表
// @Description  获取所有上架商品的列表，支持分页和筛选
// @Tags         商品
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        category_id query int false "分类ID"
// @Param        keyword query string false "搜索关键词"
// @Success      200 {object} SwaggerPagedResponse{data=[]SwaggerProduct} "商品列表"
// @Failure      500 {object} SwaggerErrorResponse "服务器错误"
// @Router       /api/products [get]

// GetProductDetail 获取商品详情
// @Summary      获取商品详情
// @Description  获取指定商品的详细信息
// @Tags         商品
// @Produce      json
// @Param        id path int true "商品ID"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerProduct} "商品详情"
// @Failure      400 {object} SwaggerErrorResponse "商品不存在"
// @Router       /api/products/{id} [get]

// GetProductCategories 获取商品分类
// @Summary      获取商品分类列表
// @Description  获取所有商品分类
// @Tags         商品
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=[]object{id=int,name=string,sort_order=int}} "分类列表"
// @Router       /api/products/categories [get]

// ==================== 订单 API 文档注解 ====================

// CreateOrder 创建订单
// @Summary      创建订单
// @Description  创建新的商品订单
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        request body SwaggerCreateOrderRequest true "订单信息"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerOrder} "订单创建成功"
// @Failure      400 {object} SwaggerErrorResponse "参数错误或库存不足"
// @Failure      500 {object} SwaggerErrorResponse "创建失败"
// @Security     CookieAuth
// @Router       /api/orders [post]

// GetUserOrders 获取用户订单列表
// @Summary      获取我的订单
// @Description  获取当前用户的订单列表
// @Tags         订单
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        status query int false "订单状态"
// @Success      200 {object} SwaggerPagedResponse{data=[]SwaggerOrder} "订单列表"
// @Security     CookieAuth
// @Router       /api/orders [get]

// GetOrderDetail 获取订单详情
// @Summary      获取订单详情
// @Description  获取指定订单的详细信息
// @Tags         订单
// @Produce      json
// @Param        order_no path string true "订单号"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerOrder} "订单详情"
// @Failure      400 {object} SwaggerErrorResponse "订单不存在"
// @Security     CookieAuth
// @Router       /api/orders/{order_no} [get]

// CancelOrder 取消订单
// @Summary      取消订单
// @Description  取消未支付的订单
// @Tags         订单
// @Produce      json
// @Param        order_no path string true "订单号"
// @Success      200 {object} SwaggerSuccessResponse "取消成功"
// @Failure      400 {object} SwaggerErrorResponse "订单不存在或无法取消"
// @Security     CookieAuth
// @Router       /api/orders/{order_no}/cancel [post]

// ==================== 支付 API 文档注解 ====================

// GetPaymentMethods 获取支付方式
// @Summary      获取可用支付方式
// @Description  获取系统支持的所有支付方式
// @Tags         支付
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=[]SwaggerPaymentMethod} "支付方式列表"
// @Router       /api/payment/methods [get]

// CreatePayPalPayment 创建PayPal支付
// @Summary      创建PayPal支付
// @Description  为订单创建PayPal支付链接
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        request body object{order_no=string} true "订单号"
// @Success      200 {object} SwaggerPayPalCreateResponse "PayPal支付信息"
// @Failure      400 {object} SwaggerErrorResponse "订单不存在或已支付"
// @Security     CookieAuth
// @Router       /api/payment/paypal/create [post]

// CapturePayPalPayment 捕获PayPal支付
// @Summary      完成PayPal支付
// @Description  用户完成PayPal支付后的回调处理
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        request body object{order_no=string,paypal_order_id=string} true "支付信息"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerOrder} "支付成功"
// @Failure      400 {object} SwaggerErrorResponse "支付失败"
// @Security     CookieAuth
// @Router       /api/payment/paypal/capture [post]

// CreateAlipayPayment 创建支付宝支付
// @Summary      创建支付宝支付
// @Description  为订单创建支付宝面对面支付
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        request body object{order_no=string} true "订单号"
// @Success      200 {object} SwaggerSuccessResponse{data=object{qr_code=string}} "支付宝二维码"
// @Failure      400 {object} SwaggerErrorResponse "订单不存在或已支付"
// @Security     CookieAuth
// @Router       /api/payment/alipay/create [post]

// CreateWechatPayment 创建微信支付
// @Summary      创建微信支付
// @Description  为订单创建微信扫码支付
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        request body object{order_no=string} true "订单号"
// @Success      200 {object} SwaggerSuccessResponse{data=object{qr_code=string}} "微信支付二维码"
// @Failure      400 {object} SwaggerErrorResponse "订单不存在或已支付"
// @Security     CookieAuth
// @Router       /api/payment/wechat/create [post]

// CreateStripePayment 创建Stripe支付
// @Summary      创建Stripe支付
// @Description  为订单创建Stripe支付
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        request body object{order_no=string} true "订单号"
// @Success      200 {object} SwaggerSuccessResponse{data=object{client_secret=string,payment_intent_id=string}} "Stripe支付信息"
// @Failure      400 {object} SwaggerErrorResponse "订单不存在或已支付"
// @Security     CookieAuth
// @Router       /api/payment/stripe/create [post]

// CreateUSDTPayment 创建USDT支付
// @Summary      创建USDT支付
// @Description  为订单创建USDT加密货币支付
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        request body object{order_no=string,network=string} true "订单号和网络类型(TRC20/ERC20)"
// @Success      200 {object} SwaggerSuccessResponse{data=object{address=string,amount=string,network=string}} "USDT支付信息"
// @Failure      400 {object} SwaggerErrorResponse "订单不存在或已支付"
// @Security     CookieAuth
// @Router       /api/payment/usdt/create [post]

// BalancePay 余额支付
// @Summary      余额支付
// @Description  使用账户余额支付订单
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        request body object{order_no=string} true "订单号"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerOrder} "支付成功"
// @Failure      400 {object} SwaggerErrorResponse "余额不足或订单无效"
// @Security     CookieAuth
// @Router       /api/payment/balance/pay [post]

// CheckPaymentStatus 查询支付状态
// @Summary      查询支付状态
// @Description  查询订单的支付状态
// @Tags         支付
// @Produce      json
// @Param        order_no query string true "订单号"
// @Success      200 {object} SwaggerSuccessResponse{data=object{status=int,status_text=string,paid=bool}} "支付状态"
// @Security     CookieAuth
// @Router       /api/payment/status [get]

// ==================== 优惠券 API 文档注解 ====================

// ValidateCoupon 验证优惠券
// @Summary      验证优惠券
// @Description  验证优惠券是否可用并计算折扣
// @Tags         优惠券
// @Accept       json
// @Produce      json
// @Param        request body SwaggerValidateCouponRequest true "验证信息"
// @Success      200 {object} SwaggerValidateCouponResponse "验证结果"
// @Failure      400 {object} SwaggerErrorResponse "优惠券无效"
// @Security     CookieAuth
// @Router       /api/coupons/validate [post]

// GetUserCoupons 获取用户优惠券
// @Summary      获取我的优惠券
// @Description  获取当前用户持有的优惠券列表
// @Tags         优惠券
// @Produce      json
// @Param        status query string false "状态筛选(valid/used/expired)"
// @Success      200 {object} SwaggerSuccessResponse{data=[]SwaggerCoupon} "优惠券列表"
// @Security     CookieAuth
// @Router       /api/coupons/my [get]

// ClaimCoupon 领取优惠券
// @Summary      领取优惠券
// @Description  通过优惠码领取优惠券
// @Tags         优惠券
// @Accept       json
// @Produce      json
// @Param        request body object{code=string} true "优惠码"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerCoupon} "领取成功"
// @Failure      400 {object} SwaggerErrorResponse "优惠券无效或已领取"
// @Security     CookieAuth
// @Router       /api/coupons/claim [post]

// ==================== 余额系统 API 文档注解 ====================

// GetUserBalance 获取用户余额
// @Summary      获取账户余额
// @Description  获取当前用户的账户余额信息
// @Tags         余额
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerUserBalance} "余额信息"
// @Security     CookieAuth
// @Router       /api/balance [get]

// GetBalanceLogs 获取余额变动记录
// @Summary      获取余额明细
// @Description  获取账户余额的变动记录
// @Tags         余额
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        type query string false "类型筛选"
// @Success      200 {object} SwaggerPagedResponse{data=[]SwaggerBalanceLog} "余额记录"
// @Security     CookieAuth
// @Router       /api/balance/logs [get]

// CreateRecharge 创建充值订单
// @Summary      创建充值订单
// @Description  创建账户充值订单
// @Tags         余额
// @Accept       json
// @Produce      json
// @Param        request body object{amount=number,pay_method=string} true "充值信息"
// @Success      200 {object} SwaggerSuccessResponse{data=object{order_no=string,amount=number}} "充值订单"
// @Failure      400 {object} SwaggerErrorResponse "金额无效"
// @Security     CookieAuth
// @Router       /api/balance/recharge [post]

// ==================== 积分系统 API 文档注解 ====================

// GetUserPoints 获取用户积分
// @Summary      获取积分余额
// @Description  获取当前用户的积分信息
// @Tags         积分
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{points=int,level=int,level_name=string}} "积分信息"
// @Security     CookieAuth
// @Router       /api/points [get]

// GetPointsLogs 获取积分记录
// @Summary      获取积分明细
// @Description  获取积分的获取和使用记录
// @Tags         积分
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200 {object} SwaggerPagedResponse{data=[]object{id=int,type=string,points=int,remark=string,created_at=string}} "积分记录"
// @Security     CookieAuth
// @Router       /api/points/logs [get]

// ==================== 购物车 API 文档注解 ====================

// GetCart 获取购物车
// @Summary      获取购物车
// @Description  获取当前用户的购物车内容
// @Tags         购物车
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{items=[]object{id=int,product_id=int,product_name=string,price=number,quantity=int},total=number}} "购物车内容"
// @Security     CookieAuth
// @Router       /api/cart [get]

// AddToCart 添加到购物车
// @Summary      添加商品到购物车
// @Description  将商品添加到购物车
// @Tags         购物车
// @Accept       json
// @Produce      json
// @Param        request body object{product_id=int,quantity=int} true "商品信息"
// @Success      200 {object} SwaggerSuccessResponse "添加成功"
// @Failure      400 {object} SwaggerErrorResponse "商品不存在或库存不足"
// @Security     CookieAuth
// @Router       /api/cart/add [post]

// UpdateCartItem 更新购物车商品
// @Summary      更新购物车商品数量
// @Description  修改购物车中商品的数量
// @Tags         购物车
// @Accept       json
// @Produce      json
// @Param        id path int true "购物车项ID"
// @Param        request body object{quantity=int} true "数量"
// @Success      200 {object} SwaggerSuccessResponse "更新成功"
// @Failure      400 {object} SwaggerErrorResponse "商品不存在"
// @Security     CookieAuth
// @Router       /api/cart/{id} [put]

// RemoveFromCart 从购物车移除
// @Summary      从购物车移除商品
// @Description  从购物车中移除指定商品
// @Tags         购物车
// @Produce      json
// @Param        id path int true "购物车项ID"
// @Success      200 {object} SwaggerSuccessResponse "移除成功"
// @Security     CookieAuth
// @Router       /api/cart/{id} [delete]

// ClearCart 清空购物车
// @Summary      清空购物车
// @Description  清空购物车中的所有商品
// @Tags         购物车
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse "清空成功"
// @Security     CookieAuth
// @Router       /api/cart/clear [post]

// ==================== 收藏夹 API 文档注解 ====================

// GetFavorites 获取收藏列表
// @Summary      获取收藏列表
// @Description  获取当前用户收藏的商品列表
// @Tags         收藏夹
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200 {object} SwaggerPagedResponse{data=[]SwaggerProduct} "收藏列表"
// @Security     CookieAuth
// @Router       /api/favorites [get]

// AddFavorite 添加收藏
// @Summary      收藏商品
// @Description  将商品添加到收藏夹
// @Tags         收藏夹
// @Accept       json
// @Produce      json
// @Param        request body object{product_id=int} true "商品ID"
// @Success      200 {object} SwaggerSuccessResponse "收藏成功"
// @Failure      400 {object} SwaggerErrorResponse "商品不存在或已收藏"
// @Security     CookieAuth
// @Router       /api/favorites [post]

// RemoveFavorite 取消收藏
// @Summary      取消收藏
// @Description  从收藏夹中移除商品
// @Tags         收藏夹
// @Produce      json
// @Param        product_id path int true "商品ID"
// @Success      200 {object} SwaggerSuccessResponse "取消成功"
// @Security     CookieAuth
// @Router       /api/favorites/{product_id} [delete]
