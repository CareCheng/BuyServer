package api

// ==================== Swagger API 文档注解 ====================
// 本文件包含 Swagger/OpenAPI 文档的基本配置和通用定义
// 使用 swaggo/swag 工具生成文档
// 安装: go install github.com/swaggo/swag/cmd/swag@latest
// 生成: swag init -g cmd/server/main.go -o docs

// @title           KamiServer 用户端 API
// @version         1.0
// @description     卡密购买系统用户端 API 接口文档
// @termsOfService  http://swagger.io/terms/

// @contact.name    技术支持
// @contact.email   support@example.com

// @license.name    私有
// @license.url     http://example.com/license

// @host            localhost:8080
// @BasePath        /api

// @securityDefinitions.apikey  CookieAuth
// @in                          cookie
// @name                        user_session
// @description                 用户会话Cookie认证

// @securityDefinitions.apikey  AdminCookieAuth
// @in                          cookie
// @name                        admin_session
// @description                 管理员会话Cookie认证

// ==================== 通用响应模型 ====================

// SuccessResponse 成功响应
// @Description 操作成功的通用响应结构
type SwaggerSuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message,omitempty" example:"操作成功"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应
// @Description 操作失败的通用响应结构
type SwaggerErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"参数错误"`
	Code    int    `json:"code,omitempty" example:"1001"`
}

// PagedResponse 分页响应
// @Description 分页数据的通用响应结构
type SwaggerPagedResponse struct {
	Success   bool        `json:"success" example:"true"`
	Data      interface{} `json:"data"`
	Total     int64       `json:"total" example:"100"`
	Page      int         `json:"page" example:"1"`
	PageSize  int         `json:"page_size" example:"20"`
	TotalPage int         `json:"pages" example:"5"`
}

// ==================== 用户相关模型 ====================

// UserLoginRequest 用户登录请求
// @Description 用户登录请求参数
type SwaggerUserLoginRequest struct {
	Username    string `json:"username" binding:"required" example:"testuser"`
	Password    string `json:"password" binding:"required" example:"password123"`
	CaptchaID   string `json:"captcha_id,omitempty" example:"abc123"`
	CaptchaCode string `json:"captcha_code,omitempty" example:"1234"`
	Remember    bool   `json:"remember,omitempty" example:"false"`
}

// UserRegisterRequest 用户注册请求
// @Description 用户注册请求参数
type SwaggerUserRegisterRequest struct {
	Username        string `json:"username" binding:"required" example:"newuser"`
	Email           string `json:"email" binding:"required" example:"user@example.com"`
	EmailCode       string `json:"email_code" binding:"required" example:"123456"`
	Password        string `json:"password" binding:"required" example:"password123"`
	ConfirmPassword string `json:"confirm_password" binding:"required" example:"password123"`
	Phone           string `json:"phone,omitempty" example:"13800138000"`
}

// UserInfo 用户信息
// @Description 用户基本信息
type SwaggerUserInfo struct {
	ID            uint   `json:"id" example:"1"`
	Username      string `json:"username" example:"testuser"`
	Email         string `json:"email" example:"user@example.com"`
	EmailVerified bool   `json:"email_verified" example:"true"`
	Phone         string `json:"phone" example:"13800138000"`
	Enable2FA     bool   `json:"enable_2fa" example:"false"`
	CreatedAt     string `json:"created_at" example:"2024-01-01 12:00:00"`
}

// ==================== 商品相关模型 ====================

// Product 商品信息
// @Description 商品详细信息
type SwaggerProduct struct {
	ID           uint    `json:"id" example:"1"`
	Name         string  `json:"name" example:"月卡会员"`
	Description  string  `json:"description" example:"30天会员服务"`
	Price        float64 `json:"price" example:"30.00"`
	Duration     int     `json:"duration" example:"30"`
	DurationUnit string  `json:"duration_unit" example:"天"`
	Stock        int     `json:"stock" example:"100"`
	Status       int     `json:"status" example:"1"`
	ImageURL     string  `json:"image_url" example:"/product/1/image.jpg"`
	CategoryID   uint    `json:"category_id" example:"1"`
}

// ==================== 订单相关模型 ====================

// Order 订单信息
// @Description 订单详细信息
type SwaggerOrder struct {
	ID          uint    `json:"id" example:"1"`
	OrderNo     string  `json:"order_no" example:"ORD_20240101120000001"`
	UserID      uint    `json:"user_id" example:"1"`
	Username    string  `json:"username" example:"testuser"`
	ProductID   uint    `json:"product_id" example:"1"`
	ProductName string  `json:"product_name" example:"月卡会员"`
	Price       float64 `json:"price" example:"30.00"`
	Status      int     `json:"status" example:"1"`
	PayMethod   string  `json:"pay_method" example:"paypal"`
	KamiCode    string  `json:"kami_code,omitempty" example:"XXXX-XXXX-XXXX-XXXX"`
	CreatedAt   string  `json:"created_at" example:"2024-01-01 12:00:00"`
	PaidAt      string  `json:"paid_at,omitempty" example:"2024-01-01 12:05:00"`
}

// CreateOrderRequest 创建订单请求
// @Description 创建订单请求参数
type SwaggerCreateOrderRequest struct {
	ProductID uint `json:"product_id" binding:"required" example:"1"`
}

// ==================== 支付相关模型 ====================

// PaymentMethod 支付方式
// @Description 支付方式信息
type SwaggerPaymentMethod struct {
	ID      string `json:"id" example:"paypal"`
	Name    string `json:"name" example:"PayPal"`
	Enabled bool   `json:"enabled" example:"true"`
	Icon    string `json:"icon,omitempty" example:"/icons/paypal.png"`
}

// PayPalCreateResponse PayPal支付创建响应
// @Description PayPal支付创建后的响应信息
type SwaggerPayPalCreateResponse struct {
	Success       bool   `json:"success" example:"true"`
	PayPalOrderID string `json:"paypal_order_id" example:"5O190127TN364715T"`
	ApproveURL    string `json:"approve_url" example:"https://www.sandbox.paypal.com/checkoutnow?token=5O190127TN364715T"`
}

// ==================== 客服系统模型 ====================

// SupportTicket 工单信息
// @Description 客服工单详细信息
type SwaggerSupportTicket struct {
	ID         uint   `json:"id" example:"1"`
	TicketNo   string `json:"ticket_no" example:"TK20240101000001"`
	UserID     uint   `json:"user_id" example:"1"`
	Username   string `json:"username" example:"testuser"`
	Subject    string `json:"subject" example:"支付问题咨询"`
	Status     string `json:"status" example:"open"`
	Priority   string `json:"priority" example:"normal"`
	Category   string `json:"category" example:"payment"`
	CreatedAt  string `json:"created_at" example:"2024-01-01 12:00:00"`
	UpdatedAt  string `json:"updated_at" example:"2024-01-01 12:30:00"`
	ClosedAt   string `json:"closed_at,omitempty" example:""`
	AssignedTo string `json:"assigned_to,omitempty" example:"staff1"`
}

// CreateTicketRequest 创建工单请求
// @Description 创建工单请求参数
type SwaggerCreateTicketRequest struct {
	Subject  string `json:"subject" binding:"required" example:"支付问题"`
	Content  string `json:"content" binding:"required" example:"我的订单支付失败了"`
	Category string `json:"category,omitempty" example:"payment"`
	Priority string `json:"priority,omitempty" example:"normal"`
	OrderNo  string `json:"order_no,omitempty" example:"ORD_20240101120000001"`
}

// ==================== 优惠券相关模型 ====================

// Coupon 优惠券信息
// @Description 优惠券详细信息
type SwaggerCoupon struct {
	ID           uint    `json:"id" example:"1"`
	Code         string  `json:"code" example:"DISCOUNT10"`
	Name         string  `json:"name" example:"新用户优惠"`
	Type         string  `json:"type" example:"percent"`
	Value        float64 `json:"value" example:"10"`
	MinAmount    float64 `json:"min_amount" example:"50"`
	MaxDiscount  float64 `json:"max_discount" example:"20"`
	UsageLimit   int     `json:"usage_limit" example:"100"`
	UsedCount    int     `json:"used_count" example:"50"`
	StartAt      string  `json:"start_at" example:"2024-01-01 00:00:00"`
	EndAt        string  `json:"end_at" example:"2024-12-31 23:59:59"`
	Status       int     `json:"status" example:"1"`
	ApplicableTo string  `json:"applicable_to" example:"all"`
}

// ValidateCouponRequest 验证优惠券请求
// @Description 验证优惠券请求参数
type SwaggerValidateCouponRequest struct {
	Code      string `json:"code" binding:"required" example:"DISCOUNT10"`
	ProductID uint   `json:"product_id" binding:"required" example:"1"`
	Amount    float64 `json:"amount" binding:"required" example:"100.00"`
}

// ValidateCouponResponse 验证优惠券响应
// @Description 验证优惠券响应信息
type SwaggerValidateCouponResponse struct {
	Success    bool    `json:"success" example:"true"`
	Valid      bool    `json:"valid" example:"true"`
	Discount   float64 `json:"discount" example:"10.00"`
	FinalPrice float64 `json:"final_price" example:"90.00"`
	Message    string  `json:"message,omitempty" example:"优惠券有效"`
}

// ==================== 余额系统模型 ====================

// UserBalance 用户余额
// @Description 用户余额信息
type SwaggerUserBalance struct {
	ID        uint    `json:"id" example:"1"`
	UserID    uint    `json:"user_id" example:"1"`
	Balance   float64 `json:"balance" example:"100.00"`
	Frozen    float64 `json:"frozen" example:"0.00"`
	UpdatedAt string  `json:"updated_at" example:"2024-01-01 12:00:00"`
}

// BalanceLog 余额变动记录
// @Description 余额变动记录信息
type SwaggerBalanceLog struct {
	ID        uint    `json:"id" example:"1"`
	UserID    uint    `json:"user_id" example:"1"`
	Type      string  `json:"type" example:"recharge"`
	Amount    float64 `json:"amount" example:"100.00"`
	Balance   float64 `json:"balance" example:"100.00"`
	OrderNo   string  `json:"order_no,omitempty" example:"RCH20240101000001"`
	Remark    string  `json:"remark" example:"充值"`
	CreatedAt string  `json:"created_at" example:"2024-01-01 12:00:00"`
}

// ==================== FAQ模型 ====================

// FAQ 常见问题
// @Description FAQ详细信息
type SwaggerFAQ struct {
	ID         uint   `json:"id" example:"1"`
	CategoryID uint   `json:"category_id" example:"1"`
	Question   string `json:"question" example:"如何支付？"`
	Answer     string `json:"answer" example:"我们支持多种支付方式..."`
	ViewCount  int    `json:"view_count" example:"100"`
	Helpful    int    `json:"helpful" example:"50"`
	Status     int    `json:"status" example:"1"`
	SortOrder  int    `json:"sort_order" example:"1"`
}

// FAQCategory FAQ分类
// @Description FAQ分类信息
type SwaggerFAQCategory struct {
	ID        uint   `json:"id" example:"1"`
	Name      string `json:"name" example:"支付问题"`
	Icon      string `json:"icon" example:"payment"`
	SortOrder int    `json:"sort_order" example:"1"`
	Status    int    `json:"status" example:"1"`
}
