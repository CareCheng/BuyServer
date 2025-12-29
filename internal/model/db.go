package model

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"user-frontend/internal/config"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var DBConnected bool

func InitDB(cfg *config.DBConfig) error {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dir := filepath.Dir(cfg.Database)
		if dir != "." && dir != "" {
			os.MkdirAll(dir, 0755)
		}
		dialector = sqlite.Open(cfg.Database)
	case "mysql":
		fallthrough
	default:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		dialector = mysql.Open(dsn)
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		DBConnected = false
		return err
	}

	// 自动迁移（注意：OperationLog 已改为文件存储，不再使用数据库）
	err = DB.AutoMigrate(&User{}, &Order{}, &Product{}, &AdminUser{}, &SystemSetting{}, &EmailVerifyCode{}, &EmailConfigDB{}, &PaymentConfigDB{}, &SystemConfigDB{}, &LoginAttempt{}, &Announcement{}, &ProductCategory{}, &Coupon{}, &CouponUsage{}, &DatabaseBackup{}, &UserSession{}, &AdminSession{}, &LoginFailureRecord{},
		// 客服支持系统
		&SupportTicket{}, &SupportMessage{}, &SupportStaff{}, &SupportStaffSession{}, &SupportConfigDB{}, &LiveChat{}, &LiveChatMessage{},
		// 手动卡密
		&ManualKami{},
		// FAQ系统
		&FAQ{}, &FAQCategory{}, &FAQFeedback{},
		// 登录设备管理
		&LoginDevice{}, &LoginHistory{},
		// 敏感操作验证
		&SensitiveOperationToken{},
		// 续费提醒
		&RenewalReminder{},
		// 商品评价
		&ProductReview{},
		// 账户注销
		&AccountDeletionRequest{},
		// 异地登录提醒
		&LoginAlert{}, &UserLoginLocation{},
		// 工单模板
		&TicketTemplate{},
		// 知识库
		&KnowledgeCategory{}, &KnowledgeArticle{},
		// 角色权限管理
		&AdminRole{}, &Admin{},
		// 余额系统
		&UserBalance{}, &BalanceLog{}, &RechargeOrder{}, &BalanceAlert{},
		// 商品多图片
		&ProductImage{},
		// 购物车
		&CartItem{},
		// 商品收藏
		&ProductFavorite{},
		// 积分系统
		&UserPoints{}, &PointsLog{}, &PointsRule{}, &PointsExchange{},
		// 定时任务
		&ScheduledTask{}, &TaskLog{},
		// 发票系统
		&Invoice{}, &InvoiceTitle{}, &InvoiceConfig{},
		// 操作撤销
		&UndoOperation{}, &UndoConfig{},
		// 智能客服
		&AutoReplyRule{}, &AutoReplyConfig{}, &AutoReplyLog{},
		// 用户优惠券
		&UserCoupon{},
		// 工单附件
		&SupportAttachment{},
		// 充值优惠活动
		&RechargePromo{}, &RechargePromoUsage{},
		// 首页配置
		&HomepageConfig{})
	if err != nil {
		DBConnected = false
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		DBConnected = false
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DBConnected = true
	return nil
}

// TestConnection 测试数据库连接
func TestConnection(cfg *config.DBConfig) error {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.Database)
	case "mysql":
		fallthrough
	default:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		dialector = mysql.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return sqlDB.Ping()
}
