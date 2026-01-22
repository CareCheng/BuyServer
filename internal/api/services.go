// Package api 提供 HTTP API 处理器
// services.go - 服务变量声明和初始化
package api

import (
	"time"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
	"user-frontend/internal/service"
)

// ==================== 核心服务 ====================
var (
	UserSvc         *service.UserService         // 用户服务
	AdminSvc        *service.AdminService        // 管理员服务
	OrderSvc        *service.OrderService        // 订单服务
	ProductSvc      *service.ProductService      // 商品服务
	EmailSvc        *service.EmailService        // 邮箱服务
	ConfigSvc       *service.ConfigService       // 配置服务（主数据库存储）
	DBConfigSvc     *service.ConfigService       // 数据库配置服务（SQLite配置数据库）
	SecuritySvc     *service.SecurityService     // 安全服务
	LogSvc          *service.LogService          // 日志服务
	AnnouncementSvc *service.AnnouncementService // 公告服务
	CategorySvc     *service.CategoryService     // 分类服务
	CouponSvc       *service.CouponService       // 优惠券服务
	BackupSvc       *service.BackupService       // 备份服务
	SessionSvc      *service.SessionService      // 会话服务（数据库持久化）
	SupportSvc      *service.SupportService      // 客服支持服务
	ManualKamiSvc   *service.ManualKamiService   // 手动卡密服务
)

// ==================== 扩展服务 ====================
var (
	BalanceSvc           *service.BalanceService           // 余额服务
	BalanceAlertSvc      *service.BalanceAlertService      // 余额告警服务
	PayPasswordSvc       *service.PayPasswordService       // 支付密码服务
	PointsSvc            *service.PointsService            // 积分服务
	CartSvc              *service.CartService              // 购物车服务
	FavoriteSvc          *service.FavoriteService          // 收藏服务
	InvoiceSvc           *service.InvoiceService           // 发票服务
	DeviceSvc            *service.DeviceService            // 设备管理服务
	LoginAlertSvc        *service.LoginAlertService        // 登录提醒服务
	RenewalSvc           *service.RenewalService           // 续费服务
	AccountDeletionSvc   *service.AccountDeletionService   // 账户注销服务
	ReviewSvc            *service.ReviewService            // 商品评价服务
	FAQSvc               *service.FAQService               // FAQ服务
	MonitorSvc           *service.MonitorService           // 系统监控服务
	RoleSvc              *service.RoleService              // 角色权限服务
	TaskSvc              *service.TaskService              // 定时任务服务
	KnowledgeSvc         *service.KnowledgeService         // 知识库服务
	UndoSvc              *service.UndoService              // 操作撤销服务
	AutoReplySvc         *service.AutoReplyService         // 智能客服服务
	SensitiveSvc         *service.SensitiveService         // 敏感操作服务
	TicketTemplateSvc    *service.TicketTemplateService    // 工单模板服务
	ExportSvc            *service.ExportService            // 数据导出服务
	ProductImageSvc      *service.ProductImageService      // 商品图片服务
	StatsSvc             *service.StatsService             // 统计服务
	RechargePromoSvc     *service.RechargePromoService     // 充值优惠服务
	HomepageSvc          *service.HomepageService          // 首页配置服务
)

// InitDBConfigService 初始化数据库配置服务（在主数据库初始化之前调用）
func InitDBConfigService(configSvc *service.ConfigService) {
	DBConfigSvc = configSvc
}

// InitServices 初始化所有服务
func InitServices(cfg *config.Config) {
	// 设置黑名单回调
	service.SetBlacklistCallback(func(key string, duration time.Duration) {
		AddToBlacklist(key, duration)
	})

	if model.DBConnected {
		repo := repository.NewRepository(model.DB)
		
		// 初始化核心服务
		initCoreServices(repo, cfg)
		
		// 初始化扩展服务
		initExtendedServices(repo)
		
		// 只有在初始化设置完成后（密码不是默认值）才创建管理员
		// 避免用默认密码 admin123 创建管理员
		if ConfigSvc != nil && !ConfigSvc.NeedsInitialSetup() {
			AdminSvc.InitDefaultAdmin(cfg.ServerConfig.AdminUsername, cfg.ServerConfig.AdminPassword)
		}

		// 启动定时任务
		go startScheduledTasks()
	}
}

// initCoreServices 初始化核心服务
func initCoreServices(repo *repository.Repository, cfg *config.Config) {
	UserSvc = service.NewUserService(repo)
	AdminSvc = service.NewAdminService(repo)
	ProductSvc = service.NewProductService(repo)

	// 复用DBConfigSvc并设置repo，而不是创建新的ConfigService
	if DBConfigSvc != nil {
		DBConfigSvc.SetRepo(repo)
		ConfigSvc = DBConfigSvc
	} else {
		ConfigSvc = service.NewConfigService(repo)
	}

	// 从数据库加载系统配置
	loadSystemConfig(cfg)

	// 从数据库加载邮箱配置
	loadEmailConfig(cfg)
	EmailSvc = service.NewEmailService(repo, &cfg.EmailConfig)

	// 从数据库加载支付配置
	if paymentCfg, err := ConfigSvc.GetPaymentConfig(); err == nil {
		cfg.PaymentConfig = *paymentCfg
	}

	// 订单服务
	OrderSvc = service.NewOrderService(repo, cfg)
	OrderSvc.SetConfigService(ConfigSvc)

	// 初始化安全服务
	SecuritySvc = service.NewSecurityService(repo)

	// 初始化日志服务（文件存储版本，不再使用数据库）
	LogSvc = service.NewLogService()

	// 初始化公告服务
	AnnouncementSvc = service.NewAnnouncementService(repo)

	// 初始化分类服务
	CategorySvc = service.NewCategoryService(repo)

	// 初始化优惠券服务
	CouponSvc = service.NewCouponService(repo)

	// 初始化备份服务
	BackupSvc = service.NewBackupService(repo, cfg.ConfigDir)

	// 初始化会话服务
	SessionSvc = service.NewSessionService(repo)

	// 初始化客服支持服务
	SupportSvc = service.NewSupportService(repo)
	SupportSvc.SetEmailService(EmailSvc)

	// 初始化手动卡密服务
	ManualKamiSvc = service.NewManualKamiService(repo)
	OrderSvc.SetManualKamiService(ManualKamiSvc)
}

// loadSystemConfig 从数据库加载系统配置
func loadSystemConfig(cfg *config.Config) {
	if sysCfg, err := ConfigSvc.GetSystemConfig(); err == nil && sysCfg.SystemTitle != "" {
		cfg.ServerConfig.SystemTitle = sysCfg.SystemTitle
		cfg.ServerConfig.AdminSuffix = sysCfg.AdminSuffix
		cfg.ServerConfig.EnableLogin = sysCfg.EnableLogin
		cfg.ServerConfig.AdminUsername = sysCfg.AdminUsername
		if sysCfg.AdminPassword != "" {
			cfg.ServerConfig.AdminPassword = sysCfg.AdminPassword
		}
		cfg.ServerConfig.Enable2FA = sysCfg.Enable2FA
		cfg.ServerConfig.TOTPSecret = sysCfg.TOTPSecret
	}
}

// loadEmailConfig 从数据库加载邮箱配置
func loadEmailConfig(cfg *config.Config) {
	if emailCfg, err := ConfigSvc.GetEmailConfig(); err == nil {
		cfg.EmailConfig = *emailCfg
	}
}


// initExtendedServices 初始化扩展服务
func initExtendedServices(repo *repository.Repository) {
	// 余额服务
	BalanceSvc = service.NewBalanceService(repo)
	BalanceSvc.SetConfigService(ConfigSvc) // 设置配置服务引用

	// 余额告警服务
	BalanceAlertSvc = service.NewBalanceAlertService(repo)
	BalanceAlertSvc.SetConfigService(ConfigSvc) // 设置配置服务引用

	// 支付密码服务
	PayPasswordSvc = service.NewPayPasswordService(repo)

	// 积分服务
	PointsSvc = service.NewPointsService(repo)

	// 购物车服务
	CartSvc = service.NewCartService(repo)

	// 收藏服务
	FavoriteSvc = service.NewFavoriteService(repo)

	// 发票服务
	InvoiceSvc = service.NewInvoiceService(repo, EmailSvc)

	// 设备管理服务
	DeviceSvc = service.NewDeviceService(repo)

	// 登录提醒服务
	LoginAlertSvc = service.NewLoginAlertService(repo, EmailSvc)

	// 续费服务
	RenewalSvc = service.NewRenewalService(repo, EmailSvc)

	// 账户注销服务
	AccountDeletionSvc = service.NewAccountDeletionService(repo, EmailSvc)

	// 商品评价服务
	ReviewSvc = service.NewReviewService(repo)

	// FAQ服务
	FAQSvc = service.NewFAQService(repo)

	// 系统监控服务
	MonitorSvc = service.NewMonitorService(repo)

	// 角色权限服务
	RoleSvc = service.NewRoleService(repo)

	// 定时任务服务
	TaskSvc = service.NewTaskService(repo)

	// 知识库服务
	KnowledgeSvc = service.NewKnowledgeService(repo)

	// 操作撤销服务
	UndoSvc = service.NewUndoService(repo)

	// 智能客服服务
	AutoReplySvc = service.NewAutoReplyService(repo)

	// 敏感操作服务
	SensitiveSvc = service.NewSensitiveService(repo, EmailSvc)

	// 工单模板服务
	TicketTemplateSvc = service.NewTicketTemplateService(repo)

	// 数据导出服务
	ExportSvc = service.NewExportService(repo)

	// 商品图片服务
	ProductImageSvc = service.NewProductImageService(repo)

	// 统计服务
	StatsSvc = service.NewStatsService(repo)

	// 充值优惠服务
	RechargePromoSvc = service.NewRechargePromoService(repo)

	// 设置余额服务的优惠服务引用
	BalanceSvc.SetPromoService(RechargePromoSvc)

	// 首页配置服务
	HomepageSvc = service.NewHomepageService(model.DB)
}

// startScheduledTasks 启动定时任务
func startScheduledTasks() {
	// 每分钟执行一次
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		// 取消过期订单（30分钟未支付）
		if OrderSvc != nil {
			OrderSvc.CancelExpiredOrders(30)
		}
		// 清理安全服务过期记录
		if SecuritySvc != nil {
			SecuritySvc.CleanupExpiredRecords()
		}
		// 清理过期会话（数据库）
		if SessionSvc != nil {
			SessionSvc.CleanupExpiredSessions()
		}
		// 清理过期令牌
		CleanupExpiredTokens()
		// 清理过期客服会话
		if SupportSvc != nil {
			SupportSvc.CleanupExpiredStaffSessions()
		}
	}
}
