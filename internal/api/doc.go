// Package api HTTP API 处理层
// 本包包含所有 HTTP API 的路由注册和请求处理。
//
// 主要组件：
//   - handler.go: 路由注册和服务初始化
//   - user_handler.go: 用户认证相关 API
//   - order_handler.go: 订单相关 API
//   - product_handler.go: 商品相关 API
//   - payment_handler.go: 支付相关 API
//   - admin_handler.go: 管理员相关 API
//   - support_handler.go: 客服工单相关 API
//   - middleware.go: 中间件（认证、限流、CSRF等）
//   - response_helper.go: 响应辅助函数
//   - error_codes.go: 统一错误码定义
//
// API 设计规范：
//   - 所有 API 路径以 /api 开头
//   - 用户 API: /api/user/*
//   - 订单 API: /api/orders/*
//   - 商品 API: /api/products/*
//   - 支付 API: /api/payment/*
//   - 管理员 API: /api/admin/*
//   - 客服 API: /api/support/*
//
// 响应格式：
//
//	// 成功响应
//	{
//	    "success": true,
//	    "message": "操作成功",
//	    "data": { ... }
//	}
//
//	// 错误响应
//	{
//	    "success": false,
//	    "error": "错误描述",
//	    "code": 1001
//	}
//
//	// 分页响应
//	{
//	    "success": true,
//	    "data": [...],
//	    "total": 100,
//	    "page": 1,
//	    "page_size": 20,
//	    "pages": 5
//	}
//
// 认证方式：
//   - 用户认证: Cookie (user_session)
//   - 管理员认证: Cookie (admin_session)
//   - CSRF 保护: X-CSRF-Token 请求头
//
// 限流策略：
//   - 登录接口: 5次/分钟/IP
//   - 注册接口: 3次/分钟/IP
//   - 发送验证码: 1次/分钟/邮箱
//   - 普通接口: 60次/分钟/IP
package api

// ==================== API 设计说明 ====================
//
// 1. RESTful 设计
//    - GET: 获取资源
//    - POST: 创建资源
//    - PUT: 更新资源
//    - DELETE: 删除资源
//
// 2. HTTP 状态码使用
//    - 200: 成功
//    - 400: 请求参数错误
//    - 401: 未认证
//    - 403: 无权限
//    - 404: 资源不存在
//    - 429: 请求过于频繁
//    - 500: 服务器内部错误
//
// 3. 错误处理
//    - 业务错误通过 JSON 响应返回
//    - 错误信息使用中文便于前端展示
//    - 错误码用于前端判断错误类型
//
// 4. 安全措施
//    - CSRF Token 验证
//    - 请求频率限制
//    - 敏感操作二次验证
//    - SQL 注入防护（使用参数化查询）
//    - XSS 防护（响应头设置）
//
// 5. 日志记录
//    - 所有请求记录访问日志
//    - 敏感操作记录操作日志
//    - 错误响应记录错误日志
//
// ==================== 全局服务变量 ====================
// 以下变量在 handler.go 中初始化

// 服务层实例说明:
// - UserSvc: 用户服务
// - SessionSvc: 会话服务
// - OrderSvc: 订单服务
// - ProductSvc: 商品服务
// - PaymentSvc: 支付服务
// - EmailSvc: 邮件服务
// - BalanceSvc: 余额服务
// - PointsSvc: 积分服务
// - CouponSvc: 优惠券服务
// - TicketSvc: 工单服务
// - NotificationSvc: 通知服务
// - LogSvc: 日志服务
// - SecuritySvc: 安全服务
// - DeviceSvc: 设备管理服务

// ==================== 中间件说明 ====================
//
// 1. AuthMiddleware - 用户认证中间件
//    检查 user_session Cookie，验证用户登录状态
//
// 2. AdminAuthMiddleware - 管理员认证中间件
//    检查 admin_session Cookie，验证管理员登录状态
//
// 3. CSRFMiddleware - CSRF 防护中间件
//    验证 X-CSRF-Token 请求头
//
// 4. RateLimitMiddleware - 限流中间件
//    基于 IP 地址的请求频率限制
//
// 5. CORSMiddleware - 跨域中间件
//    处理跨域请求的预检和响应头
//
// 6. LoggingMiddleware - 日志中间件
//    记录请求和响应信息
