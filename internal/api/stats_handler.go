package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminGetSalesStats 获取销售统计
// GET /api/admin/stats/sales
func AdminGetSalesStats(c *gin.Context) {
	// 解析日期参数
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		startDate = time.Now().AddDate(0, -1, 0)
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		endDate = time.Now()
	}
	// 结束日期设为当天结束
	endDate = endDate.Add(24*time.Hour - time.Second)

	if StatsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats, err := StatsSvc.GetSalesStats(startDate, endDate)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取统计失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": stats})
}

// AdminGetDailySalesData 获取每日销售数据
// GET /api/admin/stats/daily
func AdminGetDailySalesData(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	if StatsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	data, err := StatsSvc.GetDailySalesData(days)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取数据失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}

// AdminGetProductRanking 获取商品销售排行
// GET /api/admin/stats/product-ranking
func AdminGetProductRanking(c *gin.Context) {
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", endDateStr)
	endDate = endDate.Add(24*time.Hour - time.Second)

	if limit < 1 || limit > 100 {
		limit = 10
	}

	if StatsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	data, err := StatsSvc.GetProductSalesRanking(startDate, endDate, limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取数据失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}

// AdminGetPaymentStats 获取支付方式统计
// GET /api/admin/stats/payment
func AdminGetPaymentStats(c *gin.Context) {
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", endDateStr)
	endDate = endDate.Add(24*time.Hour - time.Second)

	if StatsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	data, err := StatsSvc.GetPaymentMethodStats(startDate, endDate)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取数据失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}

// AdminGetUserStats 获取用户统计
// GET /api/admin/stats/users
func AdminGetUserStats(c *gin.Context) {
	if StatsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats, err := StatsSvc.GetUserStats()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取统计失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": stats})
}

// AdminGetHourlySales 获取24小时销售数据
// GET /api/admin/stats/hourly
func AdminGetHourlySales(c *gin.Context) {
	if StatsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	data, err := StatsSvc.GetHourlySalesData()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取数据失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}

// AdminGetCategoryStats 获取分类销售统计
// GET /api/admin/stats/category
func AdminGetCategoryStats(c *gin.Context) {
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", endDateStr)
	endDate = endDate.Add(24*time.Hour - time.Second)

	if StatsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	data, err := StatsSvc.GetCategoryStats(startDate, endDate)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取数据失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}

// AdminGetOverviewStats 获取概览统计（综合数据）
// GET /api/admin/stats/overview
func AdminGetOverviewStats(c *gin.Context) {
	if StatsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 今日统计
	todayStats, _ := StatsSvc.GetSalesStats(todayStart, now)

	// 本月统计
	monthStats, _ := StatsSvc.GetSalesStats(monthStart, now)

	// 用户统计
	userStats, _ := StatsSvc.GetUserStats()

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"today": todayStats,
			"month": monthStats,
			"users": userStats,
		},
	})
}
