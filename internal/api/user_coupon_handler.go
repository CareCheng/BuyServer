package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ==================== 用户优惠券 API ====================

// GetMyCoupons 获取我的优惠券列表
// 返回用户持有的所有优惠券
func GetMyCoupons(c *gin.Context) {
	if CouponSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	coupons, total, err := CouponSvc.GetUserCoupons(userID.(uint), status, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取优惠券列表失败"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(200, gin.H{
		"success":     true,
		"data":        coupons,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetMyAvailableCoupons 获取我的可用优惠券列表
// 返回未使用且未过期的优惠券，可根据订单金额筛选
func GetMyAvailableCoupons(c *gin.Context) {
	if CouponSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	orderAmount, _ := strconv.ParseFloat(c.Query("order_amount"), 64)

	coupons, err := CouponSvc.GetUserAvailableCoupons(userID.(uint), orderAmount)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取可用优惠券失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    coupons,
		"count":   len(coupons),
	})
}

// GetMyCouponDetail 获取我的优惠券详情
func GetMyCouponDetail(c *gin.Context) {
	if CouponSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	couponID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的优惠券ID"})
		return
	}

	coupon, err := CouponSvc.GetUserCouponByID(userID.(uint), uint(couponID))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "优惠券不存在"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    coupon,
	})
}

// GetMyCouponCount 获取我的优惠券数量统计
func GetMyCouponCount(c *gin.Context) {
	if CouponSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	// 获取各状态的数量
	unusedCount, _, _ := CouponSvc.GetUserCoupons(userID.(uint), 0, 1, 1) // 未使用
	usedCount, _, _ := CouponSvc.GetUserCoupons(userID.(uint), 1, 1, 1)   // 已使用
	expiredCount, _, _ := CouponSvc.GetUserCoupons(userID.(uint), 2, 1, 1) // 已过期

	// 获取可用数量（未使用且未过期）
	availableCoupons, _ := CouponSvc.GetUserAvailableCoupons(userID.(uint), 0)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"unused":    len(unusedCount),
			"used":      len(usedCount),
			"expired":   len(expiredCount),
			"available": len(availableCoupons),
		},
	})
}
