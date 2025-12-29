package repository

import (
	"time"

	"user-frontend/internal/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetDB 获取数据库连接（供需要直接操作数据库的服务使用）
func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

// User 相关操作
func (r *Repository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *Repository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *Repository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *Repository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *Repository) GetAllUsers(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	
	r.db.Model(&model.User{}).Count(&total)
	err := r.db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&users).Error
	return users, total, err
}

// AdminUser 相关操作
func (r *Repository) CreateAdminUser(admin *model.AdminUser) error {
	return r.db.Create(admin).Error
}

func (r *Repository) GetAdminByUsername(username string) (*model.AdminUser, error) {
	var admin model.AdminUser
	err := r.db.Where("username = ?", username).First(&admin).Error
	return &admin, err
}

func (r *Repository) UpdateAdminUser(admin *model.AdminUser) error {
	return r.db.Save(admin).Error
}

func (r *Repository) GetAllAdmins() ([]model.AdminUser, error) {
	var admins []model.AdminUser
	err := r.db.Find(&admins).Error
	return admins, err
}

// Product 相关操作
func (r *Repository) CreateProduct(product *model.Product) error {
	return r.db.Create(product).Error
}

func (r *Repository) GetProductByID(id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.First(&product, id).Error
	return &product, err
}

func (r *Repository) UpdateProduct(product *model.Product) error {
	return r.db.Save(product).Error
}

// DecrementProductStock 原子减少商品库存
// 使用数据库级别的原子操作，防止并发超卖
// 返回：affected 影响行数，error 错误
func (r *Repository) DecrementProductStock(productID uint, quantity int) (int64, error) {
	result := r.db.Model(&model.Product{}).
		Where("id = ? AND stock >= ? AND stock != -1", productID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))
	return result.RowsAffected, result.Error
}

// IncrementProductStock 原子增加商品库存（用于订单取消/退款）
func (r *Repository) IncrementProductStock(productID uint, quantity int) error {
	return r.db.Model(&model.Product{}).
		Where("id = ? AND stock != -1", productID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}

func (r *Repository) DeleteProduct(id uint) error {
	return r.db.Delete(&model.Product{}, id).Error
}

func (r *Repository) GetAllProducts(onlyActive bool) ([]model.Product, error) {
	var products []model.Product
	query := r.db.Order("sort_order ASC, id DESC")
	if onlyActive {
		query = query.Where("status = ?", 1)
	}
	err := query.Find(&products).Error
	return products, err
}

func (r *Repository) GetProductsWithPagination(page, pageSize int, onlyActive bool) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64
	
	query := r.db.Model(&model.Product{})
	if onlyActive {
		query = query.Where("status = ?", 1)
	}
	query.Count(&total)
	
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("sort_order ASC, id DESC").Find(&products).Error
	return products, total, err
}

// Order 相关操作
func (r *Repository) CreateOrder(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *Repository) GetOrderByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.First(&order, id).Error
	return &order, err
}

func (r *Repository) GetOrderByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

func (r *Repository) UpdateOrder(order *model.Order) error {
	return r.db.Save(order).Error
}

func (r *Repository) GetOrdersByUserID(userID uint, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	
	r.db.Model(&model.Order{}).Where("user_id = ?", userID).Count(&total)
	err := r.db.Where("user_id = ?", userID).Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&orders).Error
	return orders, total, err
}

func (r *Repository) GetAllOrders(page, pageSize int, status *int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	
	query := r.db.Model(&model.Order{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Count(&total)
	
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&orders).Error
	return orders, total, err
}

func (r *Repository) GetOrderStats() (map[string]interface{}, error) {
	var totalOrders int64
	var paidOrders int64
	var totalRevenue float64
	var todayOrders int64

	r.db.Model(&model.Order{}).Count(&totalOrders)
	r.db.Model(&model.Order{}).Where("status >= ?", 1).Count(&paidOrders)
	r.db.Model(&model.Order{}).Where("status >= ?", 1).Select("COALESCE(SUM(price), 0)").Scan(&totalRevenue)

	// 使用Go计算今天的时间范围，兼容所有数据库
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	r.db.Model(&model.Order{}).Where("created_at >= ? AND created_at < ?", todayStart, todayEnd).Count(&todayOrders)

	return map[string]interface{}{
		"total_orders":  totalOrders,
		"paid_orders":   paidOrders,
		"total_revenue": totalRevenue,
		"today_orders":  todayOrders,
	}, nil
}

// SystemSetting 相关操作
func (r *Repository) GetSetting(key string) (string, error) {
	var setting model.SystemSetting
	err := r.db.Where("`key` = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (r *Repository) SetSetting(key, value, remark string) error {
	var setting model.SystemSetting
	err := r.db.Where("`key` = ?", key).First(&setting).Error
	if err != nil {
		// 不存在则创建
		setting = model.SystemSetting{
			Key:    key,
			Value:  value,
			Remark: remark,
		}
		return r.db.Create(&setting).Error
	}
	// 存在则更新
	setting.Value = value
	if remark != "" {
		setting.Remark = remark
	}
	return r.db.Save(&setting).Error
}

func (r *Repository) GetAllSettings() ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := r.db.Find(&settings).Error
	return settings, err
}

// EmailVerifyCode 相关操作
func (r *Repository) CreateEmailVerifyCode(code *model.EmailVerifyCode) error {
	return r.db.Create(code).Error
}

func (r *Repository) GetLatestEmailVerifyCode(email, codeType string) (*model.EmailVerifyCode, error) {
	var code model.EmailVerifyCode
	err := r.db.Where("email = ? AND type = ? AND used = ?", email, codeType, false).
		Order("created_at DESC").First(&code).Error
	return &code, err
}

func (r *Repository) MarkEmailVerifyCodeUsed(id uint) error {
	return r.db.Model(&model.EmailVerifyCode{}).Where("id = ?", id).Update("used", true).Error
}

// CleanExpiredEmailCodes 清理过期的验证码
func (r *Repository) CleanExpiredEmailCodes() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&model.EmailVerifyCode{}).Error
}

// ==================== 邮箱配置相关操作 ====================

// GetEmailConfig 获取邮箱配置
func (r *Repository) GetEmailConfig() (*model.EmailConfigDB, error) {
	var config model.EmailConfigDB
	err := r.db.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveEmailConfig 保存邮箱配置
func (r *Repository) SaveEmailConfig(config *model.EmailConfigDB) error {
	var existing model.EmailConfigDB
	err := r.db.First(&existing).Error
	if err != nil {
		// 不存在则创建
		return r.db.Create(config).Error
	}
	// 存在则更新
	config.ID = existing.ID
	return r.db.Save(config).Error
}

// ==================== 支付配置相关操作 ====================

// GetPaymentConfig 获取指定类型的支付配置
func (r *Repository) GetPaymentConfig(paymentType string) (*model.PaymentConfigDB, error) {
	var config model.PaymentConfigDB
	err := r.db.Where("payment_type = ?", paymentType).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetAllPaymentConfigs 获取所有支付配置
func (r *Repository) GetAllPaymentConfigs() ([]model.PaymentConfigDB, error) {
	var configs []model.PaymentConfigDB
	err := r.db.Find(&configs).Error
	return configs, err
}

// SavePaymentConfig 保存支付配置
func (r *Repository) SavePaymentConfig(config *model.PaymentConfigDB) error {
	var existing model.PaymentConfigDB
	err := r.db.Where("payment_type = ?", config.PaymentType).First(&existing).Error
	if err != nil {
		// 不存在则创建
		return r.db.Create(config).Error
	}
	// 存在则更新
	config.ID = existing.ID
	return r.db.Save(config).Error
}

// ==================== 系统配置相关操作 ====================

// GetSystemConfig 获取系统配置
func (r *Repository) GetSystemConfig() (*model.SystemConfigDB, error) {
	var config model.SystemConfigDB
	err := r.db.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveSystemConfig 保存系统配置
func (r *Repository) SaveSystemConfig(config *model.SystemConfigDB) error {
	var existing model.SystemConfigDB
	err := r.db.First(&existing).Error
	if err != nil {
		// 不存在则创建
		return r.db.Create(config).Error
	}
	// 存在则更新
	config.ID = existing.ID
	return r.db.Save(config).Error
}

// ==================== 登录尝试相关操作 ====================

func (r *Repository) CreateLoginAttempt(attempt *model.LoginAttempt) error {
	return r.db.Create(attempt).Error
}

func (r *Repository) GetRecentLoginAttempts(username string, since time.Time) ([]model.LoginAttempt, error) {
	var attempts []model.LoginAttempt
	err := r.db.Where("username = ? AND created_at > ?", username, since).Order("created_at DESC").Find(&attempts).Error
	return attempts, err
}

// ==================== 公告相关操作 ====================

func (r *Repository) CreateAnnouncement(announcement *model.Announcement) error {
	return r.db.Create(announcement).Error
}

func (r *Repository) UpdateAnnouncement(announcement *model.Announcement) error {
	return r.db.Save(announcement).Error
}

func (r *Repository) DeleteAnnouncement(id uint) error {
	return r.db.Delete(&model.Announcement{}, id).Error
}

func (r *Repository) GetAnnouncementByID(id uint) (*model.Announcement, error) {
	var announcement model.Announcement
	err := r.db.First(&announcement, id).Error
	return &announcement, err
}

func (r *Repository) GetAllAnnouncements() ([]model.Announcement, error) {
	var announcements []model.Announcement
	err := r.db.Order("sort_order ASC, id DESC").Find(&announcements).Error
	return announcements, err
}

func (r *Repository) GetActiveAnnouncements() ([]model.Announcement, error) {
	var announcements []model.Announcement
	now := time.Now()
	err := r.db.Where("status = ? AND (start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", 1, now, now).
		Order("sort_order ASC, id DESC").Find(&announcements).Error
	return announcements, err
}

// ==================== 商品分类相关操作 ====================

func (r *Repository) CreateCategory(category *model.ProductCategory) error {
	return r.db.Create(category).Error
}

func (r *Repository) UpdateCategory(category *model.ProductCategory) error {
	return r.db.Save(category).Error
}

func (r *Repository) DeleteCategory(id uint) error {
	return r.db.Delete(&model.ProductCategory{}, id).Error
}

func (r *Repository) GetCategoryByID(id uint) (*model.ProductCategory, error) {
	var category model.ProductCategory
	err := r.db.First(&category, id).Error
	return &category, err
}

func (r *Repository) GetAllCategories(onlyActive bool) ([]model.ProductCategory, error) {
	var categories []model.ProductCategory
	query := r.db.Order("sort_order ASC, id ASC")
	if onlyActive {
		query = query.Where("status = ?", 1)
	}
	err := query.Find(&categories).Error
	return categories, err
}

// ==================== 订单高级查询 ====================

// OrderSearchParams 订单搜索参数
type OrderSearchParams struct {
	OrderNo   string
	Username  string
	Status    *int
	StartDate *time.Time
	EndDate   *time.Time
}

func (r *Repository) SearchOrders(params *OrderSearchParams, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{})

	if params.OrderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+params.OrderNo+"%")
	}
	if params.Username != "" {
		query = query.Where("username LIKE ?", "%"+params.Username+"%")
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.StartDate != nil {
		query = query.Where("created_at >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		query = query.Where("created_at <= ?", *params.EndDate)
	}

	query.Count(&total)
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&orders).Error
	return orders, total, err
}

// GetOrderByOrderNoAndEmail 通过订单号和邮箱查询订单（未登录查询）
func (r *Repository) GetOrderByOrderNoAndEmail(orderNo, email string) (*model.Order, error) {
	var order model.Order
	err := r.db.Joins("JOIN users ON users.id = orders.user_id").
		Where("orders.order_no = ? AND users.email = ?", orderNo, email).
		First(&order).Error
	return &order, err
}

// ==================== 统计相关 ====================

// GetOrderStatsByDateRange 获取日期范围内的订单统计
func (r *Repository) GetOrderStatsByDateRange(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// 按天统计
	rows, err := r.db.Model(&model.Order{}).
		Select("DATE(created_at) as date, COUNT(*) as count, SUM(CASE WHEN status >= 1 THEN price ELSE 0 END) as revenue").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var date string
		var count int64
		var revenue float64
		rows.Scan(&date, &count, &revenue)
		results = append(results, map[string]interface{}{
			"date":    date,
			"count":   count,
			"revenue": revenue,
		})
	}

	return results, nil
}

// CancelExpiredOrders 取消过期订单
func (r *Repository) CancelExpiredOrders(expireMinutes int) (int64, error) {
	expireTime := time.Now().Add(-time.Duration(expireMinutes) * time.Minute)
	result := r.db.Model(&model.Order{}).
		Where("status = ? AND created_at < ?", 0, expireTime).
		Update("status", 3) // 3: 已取消
	return result.RowsAffected, result.Error
}

// GetProductsByCategory 按分类获取商品
func (r *Repository) GetProductsByCategory(categoryID uint, onlyActive bool) ([]model.Product, error) {
	var products []model.Product
	query := r.db.Where("category_id = ?", categoryID).Order("sort_order ASC, id DESC")
	if onlyActive {
		query = query.Where("status = ?", 1)
	}
	err := query.Find(&products).Error
	return products, err
}

// ==================== 优惠券相关操作 ====================

func (r *Repository) CreateCoupon(coupon *model.Coupon) error {
	return r.db.Create(coupon).Error
}

func (r *Repository) UpdateCoupon(coupon *model.Coupon) error {
	return r.db.Save(coupon).Error
}

func (r *Repository) DeleteCoupon(id uint) error {
	return r.db.Delete(&model.Coupon{}, id).Error
}

func (r *Repository) GetCouponByID(id uint) (*model.Coupon, error) {
	var coupon model.Coupon
	err := r.db.First(&coupon, id).Error
	return &coupon, err
}

func (r *Repository) GetCouponByCode(code string) (*model.Coupon, error) {
	var coupon model.Coupon
	err := r.db.Where("code = ?", code).First(&coupon).Error
	return &coupon, err
}

func (r *Repository) GetAllCoupons() ([]model.Coupon, error) {
	var coupons []model.Coupon
	err := r.db.Order("id DESC").Find(&coupons).Error
	return coupons, err
}

func (r *Repository) GetActiveCoupons() ([]model.Coupon, error) {
	var coupons []model.Coupon
	now := time.Now()
	err := r.db.Where("status = ? AND (start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?) AND (total_count = -1 OR used_count < total_count)", 1, now, now).
		Order("id DESC").Find(&coupons).Error
	return coupons, err
}

func (r *Repository) IncrementCouponUsedCount(couponID uint) error {
	return r.db.Model(&model.Coupon{}).Where("id = ?", couponID).
		UpdateColumn("used_count", gorm.Expr("used_count + ?", 1)).Error
}

// CouponUsage 相关
func (r *Repository) CreateCouponUsage(usage *model.CouponUsage) error {
	return r.db.Create(usage).Error
}

func (r *Repository) GetUserCouponUsageCount(userID, couponID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.CouponUsage{}).Where("user_id = ? AND coupon_id = ?", userID, couponID).Count(&count).Error
	return count, err
}

func (r *Repository) GetCouponUsages(couponID uint, page, pageSize int) ([]model.CouponUsage, int64, error) {
	var usages []model.CouponUsage
	var total int64

	query := r.db.Model(&model.CouponUsage{}).Where("coupon_id = ?", couponID)
	query.Count(&total)

	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&usages).Error
	return usages, total, err
}

// ==================== 数据库备份相关操作 ====================

func (r *Repository) CreateBackupRecord(backup *model.DatabaseBackup) error {
	return r.db.Create(backup).Error
}

func (r *Repository) GetAllBackups() ([]model.DatabaseBackup, error) {
	var backups []model.DatabaseBackup
	err := r.db.Order("id DESC").Find(&backups).Error
	return backups, err
}

func (r *Repository) GetBackupByID(id uint) (*model.DatabaseBackup, error) {
	var backup model.DatabaseBackup
	err := r.db.First(&backup, id).Error
	return &backup, err
}

func (r *Repository) DeleteBackupRecord(id uint) error {
	return r.db.Delete(&model.DatabaseBackup{}, id).Error
}

// ==================== 用户会话相关操作 ====================

func (r *Repository) CreateUserSession(session *model.UserSession) error {
	return r.db.Create(session).Error
}

func (r *Repository) GetUserSession(sessionID string) (*model.UserSession, error) {
	var session model.UserSession
	err := r.db.Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session).Error
	return &session, err
}

func (r *Repository) UpdateUserSession(session *model.UserSession) error {
	return r.db.Save(session).Error
}

func (r *Repository) DeleteUserSession(sessionID string) error {
	return r.db.Where("session_id = ?", sessionID).Delete(&model.UserSession{}).Error
}

func (r *Repository) DeleteExpiredUserSessions() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&model.UserSession{}).Error
}

func (r *Repository) DeleteUserSessionsByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserSession{}).Error
}

// ==================== 管理员会话相关操作 ====================

func (r *Repository) CreateAdminSession(session *model.AdminSession) error {
	return r.db.Create(session).Error
}

func (r *Repository) GetAdminSession(sessionID string) (*model.AdminSession, error) {
	var session model.AdminSession
	err := r.db.Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session).Error
	return &session, err
}

func (r *Repository) UpdateAdminSession(session *model.AdminSession) error {
	return r.db.Save(session).Error
}

func (r *Repository) DeleteAdminSession(sessionID string) error {
	return r.db.Where("session_id = ?", sessionID).Delete(&model.AdminSession{}).Error
}

func (r *Repository) DeleteExpiredAdminSessions() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&model.AdminSession{}).Error
}

// ==================== 登录失败记录相关操作 ====================

func (r *Repository) GetLoginFailureRecord(key string) (*model.LoginFailureRecord, error) {
	var record model.LoginFailureRecord
	err := r.db.Where("`key` = ?", key).First(&record).Error
	return &record, err
}

func (r *Repository) SaveLoginFailureRecord(record *model.LoginFailureRecord) error {
	var existing model.LoginFailureRecord
	err := r.db.Where("`key` = ?", record.Key).First(&existing).Error
	if err != nil {
		return r.db.Create(record).Error
	}
	record.ID = existing.ID
	return r.db.Save(record).Error
}

func (r *Repository) DeleteLoginFailureRecord(key string) error {
	return r.db.Where("`key` = ?", key).Delete(&model.LoginFailureRecord{}).Error
}

func (r *Repository) DeleteExpiredLoginFailureRecords(window time.Duration) error {
	expireTime := time.Now().Add(-window * 2)
	return r.db.Where("updated_at < ?", expireTime).Delete(&model.LoginFailureRecord{}).Error
}


// ==================== 手动卡密相关操作 ====================

// CreateManualKami 创建手动卡密
func (r *Repository) CreateManualKami(kami *model.ManualKami) error {
	return r.db.Create(kami).Error
}

// GetManualKamiByID 根据ID获取卡密
func (r *Repository) GetManualKamiByID(id uint) (*model.ManualKami, error) {
	var kami model.ManualKami
	err := r.db.First(&kami, id).Error
	return &kami, err
}

// UpdateManualKami 更新卡密
func (r *Repository) UpdateManualKami(kami *model.ManualKami) error {
	return r.db.Save(kami).Error
}

// DeleteManualKami 删除卡密
func (r *Repository) DeleteManualKami(id uint) error {
	return r.db.Delete(&model.ManualKami{}, id).Error
}

// GetAvailableManualKami 获取一个可用的卡密
func (r *Repository) GetAvailableManualKami(productID uint) (*model.ManualKami, error) {
	var kami model.ManualKami
	err := r.db.Where("product_id = ? AND status = ?", productID, 0).
		Order("id ASC").First(&kami).Error
	return &kami, err
}

// GetManualKamiCodesByProductID 获取商品的所有卡密码（用于去重）
func (r *Repository) GetManualKamiCodesByProductID(productID uint) ([]string, error) {
	var codes []string
	err := r.db.Model(&model.ManualKami{}).
		Where("product_id = ?", productID).
		Pluck("kami_code", &codes).Error
	return codes, err
}

// GetManualKamisByProductID 分页获取商品的卡密列表
func (r *Repository) GetManualKamisByProductID(productID uint, page, pageSize int, status *int) ([]model.ManualKami, int64, error) {
	var kamis []model.ManualKami
	var total int64

	query := r.db.Model(&model.ManualKami{}).Where("product_id = ?", productID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Count(&total)

	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&kamis).Error
	return kamis, total, err
}

// GetManualKamiStats 获取商品的卡密统计
func (r *Repository) GetManualKamiStats(productID uint) (map[string]int64, error) {
	stats := map[string]int64{
		"total":     0,
		"available": 0,
		"sold":      0,
		"disabled":  0,
	}

	var total, available, sold, disabled int64

	// 总数
	r.db.Model(&model.ManualKami{}).Where("product_id = ?", productID).Count(&total)
	stats["total"] = total

	// 可用
	r.db.Model(&model.ManualKami{}).Where("product_id = ? AND status = ?", productID, 0).Count(&available)
	stats["available"] = available

	// 已售出
	r.db.Model(&model.ManualKami{}).Where("product_id = ? AND status = ?", productID, 1).Count(&sold)
	stats["sold"] = sold

	// 已禁用
	r.db.Model(&model.ManualKami{}).Where("product_id = ? AND status = ?", productID, 2).Count(&disabled)
	stats["disabled"] = disabled

	return stats, nil
}
