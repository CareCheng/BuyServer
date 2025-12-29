// Package model 数据模型层
// 本包定义所有数据库表对应的数据模型结构体。
//
// 主要模型：
//   - User: 用户模型
//   - Order: 订单模型
//   - Product: 商品模型
//   - ProductCategory: 商品分类模型
//   - Coupon: 优惠券模型
//   - UserCoupon: 用户优惠券关联模型
//   - Balance: 用户余额模型
//   - BalanceLog: 余额变动记录模型
//   - Points: 用户积分模型
//   - PointsLog: 积分变动记录模型
//   - CartItem: 购物车项模型
//   - Favorite: 收藏模型
//   - SupportTicket: 客服工单模型
//   - TicketMessage: 工单消息模型
//   - Announcement: 公告模型
//   - FAQ: 常见问题模型
//   - FAQCategory: FAQ分类模型
//   - UserSession: 用户会话模型
//   - AdminSession: 管理员会话模型
//   - LoginDevice: 登录设备模型
//
// 注意：OperationLog 已移除，操作日志现在使用加密的CSV文件存储
// 参见 service/log_service.go 中的 LogEntry 结构
//
// 数据库支持：
//   - MySQL 5.7+
//   - PostgreSQL 9.6+
//   - SQLite 3.x
//
// 表命名规范：
//   - 使用蛇形命名法（snake_case）
//   - 复数形式（如 users, orders）
//   - 关联表使用下划线连接（如 user_coupons）
//
// 字段命名规范：
//   - 使用蛇形命名法
//   - 主键统一使用 id
//   - 外键使用 表名单数_id（如 user_id）
//   - 时间字段使用 _at 后缀（如 created_at）
//   - 布尔字段使用 is_ 或 enable_ 前缀
//
// 常量定义说明：
//
// 用户状态 (UserStatus*):
//   - UserStatusInactive = 0  // 未激活
//   - UserStatusActive   = 1  // 正常
//   - UserStatusBanned   = 2  // 已禁用
//
// 订单状态 (OrderStatus*):
//   - OrderStatusPending   = 0 // 待支付
//   - OrderStatusPaid      = 1 // 已支付
//   - OrderStatusCompleted = 2 // 已完成
//   - OrderStatusCancelled = 3 // 已取消
//   - OrderStatusRefunded  = 4 // 已退款
//
// 商品状态 (ProductStatus*):
//   - ProductStatusOff = 0 // 下架
//   - ProductStatusOn  = 1 // 上架
//
// 商品类型 (ProductType*):
//   - ProductTypeManual = 1 // 手动卡密（默认模式）
//
// 优惠券类型 (CouponType*):
//   - CouponTypePercent = "percent" // 百分比折扣
//   - CouponTypeFixed   = "fixed"   // 固定金额
//
// 工单状态 (TicketStatus*):
//   - TicketStatusOpen       = "open"        // 待处理
//   - TicketStatusInProgress = "in_progress" // 处理中
//   - TicketStatusClosed     = "closed"      // 已关闭
//
// 工单优先级 (TicketPriority*):
//   - TicketPriorityLow    = "low"    // 低
//   - TicketPriorityNormal = "normal" // 普通
//   - TicketPriorityHigh   = "high"   // 高
//   - TicketPriorityUrgent = "urgent" // 紧急
//
// 模型关系说明：
//
// User 1:N Order        - 一个用户可以有多个订单
// User 1:N CartItem     - 一个用户可以有多个购物车项
// User 1:N Favorite     - 一个用户可以有多个收藏
// User 1:N UserCoupon   - 一个用户可以有多个优惠券
// User 1:1 Balance      - 一个用户对应一个余额
// User 1:1 Points       - 一个用户对应一个积分
// User 1:N SupportTicket - 一个用户可以有多个工单
// User 1:N LoginDevice  - 一个用户可以有多个登录设备
//
// Product N:1 ProductCategory - 多个商品属于一个分类
// Product 1:N Order           - 一个商品可以有多个订单
// Product 1:N ManualKami      - 一个商品可以有多个手动卡密
//
// Order 1:N OrderPayment - 一个订单可以有多次支付尝试
//
// SupportTicket 1:N TicketMessage - 一个工单可以有多条消息
//
// FAQCategory 1:N FAQ - 一个分类下可以有多个FAQ
//
// 数据库索引说明：
//
// users 表索引:
//   - idx_users_username (username) - 唯一索引
//   - idx_users_email (email) - 唯一索引
//   - idx_users_status (status) - 普通索引
//
// orders 表索引:
//   - idx_orders_order_no (order_no) - 唯一索引
//   - idx_orders_user_id (user_id) - 普通索引
//   - idx_orders_status (status) - 普通索引
//   - idx_orders_created_at (created_at) - 普通索引
//
// products 表索引:
//   - idx_products_status (status) - 普通索引
//   - idx_products_category_id (category_id) - 普通索引
//
// 更多索引定义请参考各模型文件中的 gorm tag
package model
