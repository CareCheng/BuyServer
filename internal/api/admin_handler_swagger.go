package api

// ==================== 管理员认证 API 文档注解 ====================

// AdminLogin 管理员登录
// @Summary      管理员登录
// @Description  管理员账号登录
// @Tags         管理员-认证
// @Accept       json
// @Produce      json
// @Param        request body object{username=string,password=string,captcha_id=string,captcha_code=string} true "登录信息"
// @Success      200 {object} SwaggerSuccessResponse{data=object{id=int,username=string,role=string}} "登录成功"
// @Failure      400 {object} SwaggerErrorResponse "用户名或密码错误"
// @Router       /api/admin/login [post]

// AdminLogout 管理员登出
// @Summary      管理员登出
// @Description  管理员退出登录
// @Tags         管理员-认证
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse "登出成功"
// @Security     AdminCookieAuth
// @Router       /api/admin/logout [post]

// AdminInfo 获取管理员信息
// @Summary      获取管理员信息
// @Description  获取当前登录管理员的信息
// @Tags         管理员-认证
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{id=int,username=string,role=string,permissions=[]string}} "管理员信息"
// @Security     AdminCookieAuth
// @Router       /api/admin/info [get]

// ==================== 商品管理 API 文档注解 ====================

// AdminGetProducts 获取商品列表
// @Summary      获取商品列表（管理）
// @Description  获取所有商品列表，包含未上架商品
// @Tags         管理员-商品管理
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        status query int false "状态筛选"
// @Param        category_id query int false "分类筛选"
// @Success      200 {object} SwaggerPagedResponse{data=[]SwaggerProduct} "商品列表"
// @Security     AdminCookieAuth
// @Router       /api/admin/products [get]

// AdminCreateProduct 创建商品
// @Summary      创建商品
// @Description  创建新商品
// @Tags         管理员-商品管理
// @Accept       json
// @Produce      json
// @Param        request body SwaggerProduct true "商品信息"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerProduct} "创建成功"
// @Failure      400 {object} SwaggerErrorResponse "参数错误"
// @Security     AdminCookieAuth
// @Router       /api/admin/products [post]

// AdminUpdateProduct 更新商品
// @Summary      更新商品
// @Description  更新商品信息
// @Tags         管理员-商品管理
// @Accept       json
// @Produce      json
// @Param        id path int true "商品ID"
// @Param        request body SwaggerProduct true "商品信息"
// @Success      200 {object} SwaggerSuccessResponse "更新成功"
// @Failure      400 {object} SwaggerErrorResponse "商品不存在"
// @Security     AdminCookieAuth
// @Router       /api/admin/products/{id} [put]

// AdminDeleteProduct 删除商品
// @Summary      删除商品
// @Description  删除指定商品
// @Tags         管理员-商品管理
// @Produce      json
// @Param        id path int true "商品ID"
// @Success      200 {object} SwaggerSuccessResponse "删除成功"
// @Failure      400 {object} SwaggerErrorResponse "商品不存在"
// @Security     AdminCookieAuth
// @Router       /api/admin/products/{id} [delete]

// AdminUploadProductImage 上传商品图片
// @Summary      上传商品图片
// @Description  上传商品展示图片
// @Tags         管理员-商品管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path int true "商品ID"
// @Param        image formData file true "图片文件"
// @Success      200 {object} SwaggerSuccessResponse{data=object{url=string}} "上传成功"
// @Failure      400 {object} SwaggerErrorResponse "上传失败"
// @Security     AdminCookieAuth
// @Router       /api/admin/products/{id}/image [post]

// ==================== 订单管理 API 文档注解 ====================

// AdminGetOrders 获取订单列表
// @Summary      获取订单列表（管理）
// @Description  获取所有用户订单列表
// @Tags         管理员-订单管理
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        status query int false "订单状态"
// @Param        order_no query string false "订单号搜索"
// @Param        username query string false "用户名搜索"
// @Param        start_date query string false "开始日期"
// @Param        end_date query string false "结束日期"
// @Success      200 {object} SwaggerPagedResponse{data=[]SwaggerOrder} "订单列表"
// @Security     AdminCookieAuth
// @Router       /api/admin/orders [get]

// AdminGetOrderDetail 获取订单详情
// @Summary      获取订单详情（管理）
// @Description  获取指定订单的详细信息
// @Tags         管理员-订单管理
// @Produce      json
// @Param        order_no path string true "订单号"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerOrder} "订单详情"
// @Failure      400 {object} SwaggerErrorResponse "订单不存在"
// @Security     AdminCookieAuth
// @Router       /api/admin/orders/{order_no} [get]

// AdminConfirmPayment 确认支付
// @Summary      手动确认支付
// @Description  管理员手动确认订单已支付
// @Tags         管理员-订单管理
// @Accept       json
// @Produce      json
// @Param        order_no path string true "订单号"
// @Param        request body object{remark=string} false "备注"
// @Success      200 {object} SwaggerSuccessResponse "确认成功"
// @Failure      400 {object} SwaggerErrorResponse "订单不存在或状态不正确"
// @Security     AdminCookieAuth
// @Router       /api/admin/orders/{order_no}/confirm [post]

// AdminRefundOrder 订单退款
// @Summary      订单退款
// @Description  对已支付订单进行退款
// @Tags         管理员-订单管理
// @Accept       json
// @Produce      json
// @Param        order_no path string true "订单号"
// @Param        request body object{reason=string,refund_amount=number} true "退款信息"
// @Success      200 {object} SwaggerSuccessResponse "退款成功"
// @Failure      400 {object} SwaggerErrorResponse "订单不存在或无法退款"
// @Security     AdminCookieAuth
// @Router       /api/admin/orders/{order_no}/refund [post]

// AdminExportOrders 导出订单
// @Summary      导出订单
// @Description  导出订单数据为Excel文件
// @Tags         管理员-订单管理
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        status query int false "订单状态"
// @Param        start_date query string false "开始日期"
// @Param        end_date query string false "结束日期"
// @Success      200 {file} file "Excel文件"
// @Security     AdminCookieAuth
// @Router       /api/admin/orders/export [get]

// ==================== 用户管理 API 文档注解 ====================

// AdminGetUsers 获取用户列表
// @Summary      获取用户列表
// @Description  获取所有注册用户列表
// @Tags         管理员-用户管理
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        keyword query string false "搜索关键词"
// @Param        status query int false "状态筛选"
// @Success      200 {object} SwaggerPagedResponse{data=[]SwaggerUserInfo} "用户列表"
// @Security     AdminCookieAuth
// @Router       /api/admin/users [get]

// AdminGetUserDetail 获取用户详情
// @Summary      获取用户详情
// @Description  获取指定用户的详细信息
// @Tags         管理员-用户管理
// @Produce      json
// @Param        id path int true "用户ID"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerUserInfo} "用户详情"
// @Failure      400 {object} SwaggerErrorResponse "用户不存在"
// @Security     AdminCookieAuth
// @Router       /api/admin/users/{id} [get]

// AdminUpdateUser 更新用户信息
// @Summary      更新用户信息
// @Description  更新指定用户的信息
// @Tags         管理员-用户管理
// @Accept       json
// @Produce      json
// @Param        id path int true "用户ID"
// @Param        request body object{email=string,phone=string,status=int} true "用户信息"
// @Success      200 {object} SwaggerSuccessResponse "更新成功"
// @Failure      400 {object} SwaggerErrorResponse "用户不存在"
// @Security     AdminCookieAuth
// @Router       /api/admin/users/{id} [put]

// AdminBanUser 禁用用户
// @Summary      禁用/启用用户
// @Description  禁用或启用指定用户
// @Tags         管理员-用户管理
// @Accept       json
// @Produce      json
// @Param        id path int true "用户ID"
// @Param        request body object{banned=bool,reason=string} true "禁用信息"
// @Success      200 {object} SwaggerSuccessResponse "操作成功"
// @Failure      400 {object} SwaggerErrorResponse "用户不存在"
// @Security     AdminCookieAuth
// @Router       /api/admin/users/{id}/ban [post]

// AdminResetUserPassword 重置用户密码
// @Summary      重置用户密码
// @Description  重置指定用户的登录密码
// @Tags         管理员-用户管理
// @Accept       json
// @Produce      json
// @Param        id path int true "用户ID"
// @Param        request body object{new_password=string} true "新密码"
// @Success      200 {object} SwaggerSuccessResponse "重置成功"
// @Failure      400 {object} SwaggerErrorResponse "用户不存在"
// @Security     AdminCookieAuth
// @Router       /api/admin/users/{id}/reset-password [post]

// ==================== 优惠券管理 API 文档注解 ====================

// AdminGetCoupons 获取优惠券列表
// @Summary      获取优惠券列表
// @Description  获取所有优惠券列表
// @Tags         管理员-优惠券管理
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        status query int false "状态筛选"
// @Success      200 {object} SwaggerPagedResponse{data=[]SwaggerCoupon} "优惠券列表"
// @Security     AdminCookieAuth
// @Router       /api/admin/coupons [get]

// AdminCreateCoupon 创建优惠券
// @Summary      创建优惠券
// @Description  创建新的优惠券
// @Tags         管理员-优惠券管理
// @Accept       json
// @Produce      json
// @Param        request body SwaggerCoupon true "优惠券信息"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerCoupon} "创建成功"
// @Failure      400 {object} SwaggerErrorResponse "参数错误"
// @Security     AdminCookieAuth
// @Router       /api/admin/coupons [post]

// AdminUpdateCoupon 更新优惠券
// @Summary      更新优惠券
// @Description  更新优惠券信息
// @Tags         管理员-优惠券管理
// @Accept       json
// @Produce      json
// @Param        id path int true "优惠券ID"
// @Param        request body SwaggerCoupon true "优惠券信息"
// @Success      200 {object} SwaggerSuccessResponse "更新成功"
// @Failure      400 {object} SwaggerErrorResponse "优惠券不存在"
// @Security     AdminCookieAuth
// @Router       /api/admin/coupons/{id} [put]

// AdminDeleteCoupon 删除优惠券
// @Summary      删除优惠券
// @Description  删除指定优惠券
// @Tags         管理员-优惠券管理
// @Produce      json
// @Param        id path int true "优惠券ID"
// @Success      200 {object} SwaggerSuccessResponse "删除成功"
// @Failure      400 {object} SwaggerErrorResponse "优惠券不存在"
// @Security     AdminCookieAuth
// @Router       /api/admin/coupons/{id} [delete]

// ==================== 支付配置 API 文档注解 ====================

// AdminGetPaymentConfig 获取支付配置
// @Summary      获取支付配置
// @Description  获取各支付方式的配置信息
// @Tags         管理员-支付配置
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{paypal=object,alipay=object,wechat=object,stripe=object,usdt=object}} "支付配置"
// @Security     AdminCookieAuth
// @Router       /api/admin/payment/config [get]

// AdminUpdatePaymentConfig 更新支付配置
// @Summary      更新支付配置
// @Description  更新指定支付方式的配置
// @Tags         管理员-支付配置
// @Accept       json
// @Produce      json
// @Param        method path string true "支付方式(paypal/alipay/wechat/stripe/usdt)"
// @Param        request body object{enabled=bool,config=object} true "配置信息"
// @Success      200 {object} SwaggerSuccessResponse "更新成功"
// @Failure      400 {object} SwaggerErrorResponse "参数错误"
// @Security     AdminCookieAuth
// @Router       /api/admin/payment/config/{method} [put]

// ==================== 系统设置 API 文档注解 ====================

// AdminGetSettings 获取系统设置
// @Summary      获取系统设置
// @Description  获取系统各项设置
// @Tags         管理员-系统设置
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{site_name=string,site_logo=string,contact_email=string,register_enabled=bool}} "系统设置"
// @Security     AdminCookieAuth
// @Router       /api/admin/settings [get]

// AdminUpdateSettings 更新系统设置
// @Summary      更新系统设置
// @Description  更新系统设置
// @Tags         管理员-系统设置
// @Accept       json
// @Produce      json
// @Param        request body object{site_name=string,site_logo=string,contact_email=string,register_enabled=bool} true "设置信息"
// @Success      200 {object} SwaggerSuccessResponse "更新成功"
// @Security     AdminCookieAuth
// @Router       /api/admin/settings [put]

// ==================== 数据统计 API 文档注解 ====================

// AdminGetDashboard 获取仪表盘数据
// @Summary      获取仪表盘数据
// @Description  获取管理后台首页统计数据
// @Tags         管理员-数据统计
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{today_orders=int,today_revenue=number,total_users=int,pending_tickets=int}} "统计数据"
// @Security     AdminCookieAuth
// @Router       /api/admin/dashboard [get]

// AdminGetStatistics 获取统计报表
// @Summary      获取统计报表
// @Description  获取指定时间范围的统计数据
// @Tags         管理员-数据统计
// @Produce      json
// @Param        type query string true "统计类型(orders/revenue/users)"
// @Param        start_date query string true "开始日期"
// @Param        end_date query string true "结束日期"
// @Param        group_by query string false "分组方式(day/week/month)" default(day)
// @Success      200 {object} SwaggerSuccessResponse{data=[]object{date=string,value=number}} "统计数据"
// @Security     AdminCookieAuth
// @Router       /api/admin/statistics [get]

// ==================== 操作日志 API 文档注解 ====================

// AdminGetOperationLogs 获取操作日志
// @Summary      获取操作日志
// @Description  获取管理员操作日志
// @Tags         管理员-操作日志
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        admin_id query int false "管理员ID"
// @Param        action query string false "操作类型"
// @Param        start_date query string false "开始日期"
// @Param        end_date query string false "结束日期"
// @Success      200 {object} SwaggerPagedResponse{data=[]object{id=int,admin_id=int,admin_name=string,action=string,target=string,ip=string,created_at=string}} "操作日志"
// @Security     AdminCookieAuth
// @Router       /api/admin/logs [get]
