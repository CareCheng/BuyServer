package service

import (
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// StatsService 统计报表服务
type StatsService struct {
	repo *repository.Repository
}

// NewStatsService 创建统计报表服务实例
func NewStatsService(repo *repository.Repository) *StatsService {
	return &StatsService{repo: repo}
}

// SalesStats 销售统计数据
type SalesStats struct {
	TotalRevenue     float64 `json:"total_revenue"`      // 总收入
	TotalOrders      int64   `json:"total_orders"`       // 总订单数
	PaidOrders       int64   `json:"paid_orders"`        // 已支付订单数
	CompletedOrders  int64   `json:"completed_orders"`   // 已完成订单数
	CancelledOrders  int64   `json:"cancelled_orders"`   // 已取消订单数
	RefundedOrders   int64   `json:"refunded_orders"`    // 已退款订单数
	AvgOrderValue    float64 `json:"avg_order_value"`    // 平均订单金额
	ConversionRate   float64 `json:"conversion_rate"`    // 转化率（已支付/总订单）
}

// DailySalesData 每日销售数据
type DailySalesData struct {
	Date     string  `json:"date"`
	Revenue  float64 `json:"revenue"`
	Orders   int64   `json:"orders"`
	NewUsers int64   `json:"new_users"`
}

// ProductSalesData 商品销售数据
type ProductSalesData struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	SalesCount  int64   `json:"sales_count"`
	Revenue     float64 `json:"revenue"`
}

// PaymentMethodStats 支付方式统计
type PaymentMethodStats struct {
	Method  string  `json:"method"`
	Count   int64   `json:"count"`
	Revenue float64 `json:"revenue"`
	Percent float64 `json:"percent"`
}

// UserStats 用户统计数据
type UserStats struct {
	TotalUsers      int64   `json:"total_users"`       // 总用户数
	ActiveUsers     int64   `json:"active_users"`      // 活跃用户数（30天内登录）
	NewUsersToday   int64   `json:"new_users_today"`   // 今日新增
	NewUsersWeek    int64   `json:"new_users_week"`    // 本周新增
	NewUsersMonth   int64   `json:"new_users_month"`   // 本月新增
	VerifiedUsers   int64   `json:"verified_users"`    // 已验证邮箱用户
	Enable2FAUsers  int64   `json:"enable_2fa_users"`  // 启用2FA用户
	RetentionRate   float64 `json:"retention_rate"`    // 留存率
}

// GetSalesStats 获取销售统计
// 参数：
//   - startDate: 开始日期
//   - endDate: 结束日期
// 返回：
//   - 销售统计数据
//   - 错误信息
func (s *StatsService) GetSalesStats(startDate, endDate time.Time) (*SalesStats, error) {
	stats := &SalesStats{}

	// 基础查询条件
	query := s.repo.GetDB().Model(&model.Order{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate)

	// 总订单数
	query.Count(&stats.TotalOrders)

	// 各状态订单数
	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at BETWEEN ? AND ? AND status IN ?", startDate, endDate, []int{1, 2}).
		Count(&stats.PaidOrders)

	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, 2).
		Count(&stats.CompletedOrders)

	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, 3).
		Count(&stats.CancelledOrders)

	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, 4).
		Count(&stats.RefundedOrders)

	// 总收入（已支付和已完成的订单）
	var revenue struct {
		Total float64
	}
	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at BETWEEN ? AND ? AND status IN ?", startDate, endDate, []int{1, 2}).
		Select("COALESCE(SUM(price), 0) as total").
		Scan(&revenue)
	stats.TotalRevenue = revenue.Total

	// 平均订单金额
	if stats.PaidOrders > 0 {
		stats.AvgOrderValue = stats.TotalRevenue / float64(stats.PaidOrders)
	}

	// 转化率
	if stats.TotalOrders > 0 {
		stats.ConversionRate = float64(stats.PaidOrders) / float64(stats.TotalOrders) * 100
	}

	return stats, nil
}

// GetDailySalesData 获取每日销售数据
// 参数：
//   - days: 天数
// 返回：
//   - 每日销售数据列表
//   - 错误信息
func (s *StatsService) GetDailySalesData(days int) ([]DailySalesData, error) {
	var result []DailySalesData

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	// 生成日期列表
	dateMap := make(map[string]*DailySalesData)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dateMap[dateStr] = &DailySalesData{Date: dateStr}
	}

	// 查询订单数据
	var orderData []struct {
		Date    string
		Revenue float64
		Orders  int64
	}
	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at >= ? AND status IN ?", startDate, []int{1, 2}).
		Select("DATE(created_at) as date, COALESCE(SUM(price), 0) as revenue, COUNT(*) as orders").
		Group("DATE(created_at)").
		Scan(&orderData)

	for _, od := range orderData {
		if data, ok := dateMap[od.Date]; ok {
			data.Revenue = od.Revenue
			data.Orders = od.Orders
		}
	}

	// 查询新用户数据
	var userData []struct {
		Date     string
		NewUsers int64
	}
	s.repo.GetDB().Model(&model.User{}).
		Where("created_at >= ?", startDate).
		Select("DATE(created_at) as date, COUNT(*) as new_users").
		Group("DATE(created_at)").
		Scan(&userData)

	for _, ud := range userData {
		if data, ok := dateMap[ud.Date]; ok {
			data.NewUsers = ud.NewUsers
		}
	}

	// 转换为有序列表
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if data, ok := dateMap[dateStr]; ok {
			result = append(result, *data)
		}
	}

	return result, nil
}

// GetProductSalesRanking 获取商品销售排行
// 参数：
//   - startDate: 开始日期
//   - endDate: 结束日期
//   - limit: 数量限制
// 返回：
//   - 商品销售数据列表
//   - 错误信息
func (s *StatsService) GetProductSalesRanking(startDate, endDate time.Time, limit int) ([]ProductSalesData, error) {
	var result []ProductSalesData

	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at BETWEEN ? AND ? AND status IN ?", startDate, endDate, []int{1, 2}).
		Select("product_id, product_name, COUNT(*) as sales_count, COALESCE(SUM(price), 0) as revenue").
		Group("product_id, product_name").
		Order("revenue DESC").
		Limit(limit).
		Scan(&result)

	return result, nil
}

// GetPaymentMethodStats 获取支付方式统计
// 参数：
//   - startDate: 开始日期
//   - endDate: 结束日期
// 返回：
//   - 支付方式统计列表
//   - 错误信息
func (s *StatsService) GetPaymentMethodStats(startDate, endDate time.Time) ([]PaymentMethodStats, error) {
	var result []PaymentMethodStats
	var totalRevenue float64

	// 查询各支付方式数据
	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at BETWEEN ? AND ? AND status IN ? AND payment_method != ''", startDate, endDate, []int{1, 2}).
		Select("payment_method as method, COUNT(*) as count, COALESCE(SUM(price), 0) as revenue").
		Group("payment_method").
		Order("revenue DESC").
		Scan(&result)

	// 计算总收入
	for _, r := range result {
		totalRevenue += r.Revenue
	}

	// 计算百分比
	for i := range result {
		if totalRevenue > 0 {
			result[i].Percent = result[i].Revenue / totalRevenue * 100
		}
	}

	return result, nil
}

// GetUserStats 获取用户统计
// 返回：
//   - 用户统计数据
//   - 错误信息
func (s *StatsService) GetUserStats() (*UserStats, error) {
	stats := &UserStats{}
	now := time.Now()

	// 总用户数
	s.repo.GetDB().Model(&model.User{}).Count(&stats.TotalUsers)

	// 活跃用户（30天内登录）
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	s.repo.GetDB().Model(&model.User{}).
		Where("last_login_at >= ?", thirtyDaysAgo).
		Count(&stats.ActiveUsers)

	// 今日新增
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	s.repo.GetDB().Model(&model.User{}).
		Where("created_at >= ?", todayStart).
		Count(&stats.NewUsersToday)

	// 本周新增
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	s.repo.GetDB().Model(&model.User{}).
		Where("created_at >= ?", weekStart).
		Count(&stats.NewUsersWeek)

	// 本月新增
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	s.repo.GetDB().Model(&model.User{}).
		Where("created_at >= ?", monthStart).
		Count(&stats.NewUsersMonth)

	// 已验证邮箱用户（使用安全查询，忽略可能不存在的列错误）
	if err := s.repo.GetDB().Model(&model.User{}).
		Where("email_verified = ?", true).
		Count(&stats.VerifiedUsers).Error; err != nil {
		// 如果列不存在，忽略错误，保持默认值0
		stats.VerifiedUsers = 0
	}

	// 启用2FA用户（使用安全查询，忽略可能不存在的列错误）
	if err := s.repo.GetDB().Model(&model.User{}).
		Where("enable_2fa = ?", true).
		Count(&stats.Enable2FAUsers).Error; err != nil {
		// 如果列不存在，忽略错误，保持默认值0
		stats.Enable2FAUsers = 0
	}

	// 留存率（30天内注册且有登录记录的用户比例）
	var newUsers int64
	var retainedUsers int64
	s.repo.GetDB().Model(&model.User{}).
		Where("created_at >= ?", thirtyDaysAgo).
		Count(&newUsers)
	s.repo.GetDB().Model(&model.User{}).
		Where("created_at >= ? AND last_login_at > created_at", thirtyDaysAgo).
		Count(&retainedUsers)
	if newUsers > 0 {
		stats.RetentionRate = float64(retainedUsers) / float64(newUsers) * 100
	}

	return stats, nil
}

// GetHourlySalesData 获取24小时销售数据
// 返回：
//   - 每小时销售数据
//   - 错误信息
func (s *StatsService) GetHourlySalesData() ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 初始化24小时数据
	hourlyData := make(map[int]map[string]interface{})
	for i := 0; i < 24; i++ {
		hourlyData[i] = map[string]interface{}{
			"hour":    i,
			"orders":  int64(0),
			"revenue": float64(0),
		}
	}

	// 查询今日订单数据
	var orderData []struct {
		Hour    int
		Orders  int64
		Revenue float64
	}
	s.repo.GetDB().Model(&model.Order{}).
		Where("created_at >= ? AND status IN ?", todayStart, []int{1, 2}).
		Select("HOUR(created_at) as hour, COUNT(*) as orders, COALESCE(SUM(price), 0) as revenue").
		Group("HOUR(created_at)").
		Scan(&orderData)

	for _, od := range orderData {
		if data, ok := hourlyData[od.Hour]; ok {
			data["orders"] = od.Orders
			data["revenue"] = od.Revenue
		}
	}

	// 转换为有序列表
	for i := 0; i < 24; i++ {
		result = append(result, hourlyData[i])
	}

	return result, nil
}

// GetCategoryStats 获取分类销售统计
// 参数：
//   - startDate: 开始日期
//   - endDate: 结束日期
// 返回：
//   - 分类销售数据
//   - 错误信息
func (s *StatsService) GetCategoryStats(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	// 获取所有分类
	var categories []model.ProductCategory
	s.repo.GetDB().Find(&categories)

	categoryMap := make(map[uint]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	// 查询各分类销售数据
	var salesData []struct {
		CategoryID uint
		Count      int64
		Revenue    float64
	}

	s.repo.GetDB().Table("orders").
		Joins("JOIN products ON orders.product_id = products.id").
		Where("orders.created_at BETWEEN ? AND ? AND orders.status IN ?", startDate, endDate, []int{1, 2}).
		Select("products.category_id, COUNT(*) as count, COALESCE(SUM(orders.price), 0) as revenue").
		Group("products.category_id").
		Scan(&salesData)

	for _, sd := range salesData {
		categoryName := "未分类"
		if name, ok := categoryMap[sd.CategoryID]; ok {
			categoryName = name
		}
		result = append(result, map[string]interface{}{
			"category_id":   sd.CategoryID,
			"category_name": categoryName,
			"count":         sd.Count,
			"revenue":       sd.Revenue,
		})
	}

	return result, nil
}
