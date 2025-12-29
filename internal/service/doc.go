// Package service 业务逻辑层
// 本包包含所有业务逻辑处理，是连接 API 层和数据访问层的桥梁。
//
// 主要服务包括：
//   - UserService: 用户管理服务，处理注册、登录、密码修改等
//   - OrderService: 订单管理服务，处理订单创建、支付、查询等
//   - ProductService: 商品管理服务，处理商品增删改查
//   - PaymentService: 支付服务，处理各种支付方式的调用和回调
//   - SessionService: 会话管理服务，处理用户登录会话
//   - EmailService: 邮件服务，处理验证码发送和邮件通知
//   - BalanceService: 余额服务，处理用户余额充值和消费
//   - PointsService: 积分服务，处理积分的获取和使用
//   - CouponService: 优惠券服务，处理优惠券的创建和使用
//   - TicketService: 工单服务，处理客服工单
//
// 设计原则：
//   - 每个服务只处理自己领域的业务逻辑
//   - 服务之间通过依赖注入的方式协作
//   - 所有数据库操作通过 repository 层完成
//   - 错误信息使用中文，便于前端直接展示
//
// 使用示例：
//
//	// 创建用户服务
//	userSvc := service.NewUserService(repo)
//
//	// 用户注册
//	user, err := userSvc.Register("username", "email@example.com", "password", "phone")
//	if err != nil {
//	    log.Printf("注册失败: %v", err)
//	}
package service

// ==================== 服务层设计说明 ====================
//
// 1. 分层架构
//    API Layer (handler) -> Service Layer -> Repository Layer -> Database
//
// 2. 职责划分
//    - Handler: 处理 HTTP 请求/响应，参数校验
//    - Service: 业务逻辑处理，事务管理
//    - Repository: 数据访问，SQL 查询
//
// 3. 错误处理
//    - 服务层返回业务错误，使用中文描述
//    - 技术错误（如数据库错误）包装为业务错误后返回
//    - 不暴露内部实现细节
//
// 4. 事务处理
//    - 跨表操作使用事务保证数据一致性
//    - 事务在服务层管理，repository 不处理事务
//
// 5. 缓存策略
//    - 热点数据使用内存缓存
//    - 配置信息使用数据库缓存
//    - 会话数据使用数据库持久化
//
// ==================== 服务初始化顺序 ====================
//
// 服务初始化有依赖关系，需按以下顺序：
// 1. ConfigService - 配置服务（最先初始化）
// 2. EmailService - 邮件服务
// 3. UserService - 用户服务
// 4. SessionService - 会话服务
// 5. ProductService - 商品服务
// 6. OrderService - 订单服务（依赖商品服务）
// 7. PaymentService - 支付服务（依赖订单服务）
// 8. BalanceService - 余额服务
// 9. PointsService - 积分服务
// 10. CouponService - 优惠券服务
// 11. TicketService - 工单服务
// 12. NotificationService - 通知服务
// 13. LogService - 日志服务
//
// ==================== 常量定义 ====================

// OrderStatus 订单状态常量
const (
	OrderStatusPending   = 0 // 待支付
	OrderStatusPaid      = 1 // 已支付（等待发货）
	OrderStatusCompleted = 2 // 已完成
	OrderStatusCancelled = 3 // 已取消
	OrderStatusRefunded  = 4 // 已退款
)

// UserStatus 用户状态常量
const (
	UserStatusActive   = 1 // 正常
	UserStatusInactive = 0 // 未激活
	UserStatusBanned   = 2 // 已禁用
)

// PaymentMethod 支付方式常量
const (
	PaymentMethodPayPal   = "paypal"   // PayPal
	PaymentMethodAlipay   = "alipay"   // 支付宝
	PaymentMethodWechat   = "wechat"   // 微信支付
	PaymentMethodStripe   = "stripe"   // Stripe
	PaymentMethodYiPay    = "yipay"    // 易支付
	PaymentMethodUSDT     = "usdt"     // USDT
	PaymentMethodBalance  = "balance"  // 余额支付
)

// TicketStatus 工单状态常量
const (
	TicketStatusOpen       = "open"        // 待处理
	TicketStatusInProgress = "in_progress" // 处理中
	TicketStatusClosed     = "closed"      // 已关闭
)

// PointsType 积分类型常量
const (
	PointsTypeRegister = "register" // 注册奖励
	PointsTypePurchase = "purchase" // 购买获得
	PointsTypeConsume  = "consume"  // 积分消费
	PointsTypeAdmin    = "admin"    // 管理员调整
)

// BalanceType 余额变动类型常量
const (
	BalanceTypeRecharge = "recharge" // 充值
	BalanceTypePayment  = "payment"  // 支付
	BalanceTypeRefund   = "refund"   // 退款
	BalanceTypeAdmin    = "admin"    // 管理员调整
)
