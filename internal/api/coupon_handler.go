package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ==================== 优惠券管理 ====================

// AdminGetCoupons 获取优惠券列表
func AdminGetCoupons(c *gin.Context) {
	if CouponSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	coupons, err := CouponSvc.GetAllCoupons()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "coupons": coupons})
}

// AdminCreateCoupon 创建优惠券
func AdminCreateCoupon(c *gin.Context) {
	if CouponSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Name         string  `json:"name" binding:"required"`
		Code         string  `json:"code"`
		Type         string  `json:"type" binding:"required"`
		Value        float64 `json:"value" binding:"required"`
		MinAmount    float64 `json:"min_amount"`
		MaxDiscount  float64 `json:"max_discount"`
		TotalCount   int     `json:"total_count"`
		PerUserLimit int     `json:"per_user_limit"`
		ProductIDs   string  `json:"product_ids"`
		CategoryIDs  string  `json:"category_ids"`
		StartAt      string  `json:"start_at"`
		EndAt        string  `json:"end_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	var startAt, endAt *time.Time
	if req.StartAt != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", req.StartAt)
		startAt = &t
	}
	if req.EndAt != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", req.EndAt)
		endAt = &t
	}

	if req.TotalCount == 0 {
		req.TotalCount = -1 // 默认无限
	}
	if req.PerUserLimit == 0 {
		req.PerUserLimit = 1 // 默认每人1次
	}

	coupon, err := CouponSvc.CreateCoupon(
		req.Name, req.Code, req.Type, req.Value,
		req.MinAmount, req.MaxDiscount,
		req.TotalCount, req.PerUserLimit,
		req.ProductIDs, req.CategoryIDs,
		startAt, endAt,
	)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(c.GetString("admin_username"), "create", "coupon", strconv.Itoa(int(coupon.ID)), req.Name, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "coupon": coupon})
}

// AdminUpdateCoupon 更新优惠券
func AdminUpdateCoupon(c *gin.Context) {
	if CouponSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var req struct {
		Name         string  `json:"name"`
		Type         string  `json:"type"`
		Value        float64 `json:"value"`
		MinAmount    float64 `json:"min_amount"`
		MaxDiscount  float64 `json:"max_discount"`
		TotalCount   int     `json:"total_count"`
		PerUserLimit int     `json:"per_user_limit"`
		ProductIDs   string  `json:"product_ids"`
		CategoryIDs  string  `json:"category_ids"`
		StartAt      string  `json:"start_at"`
		EndAt        string  `json:"end_at"`
		Status       int     `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	var startAt, endAt *time.Time
	if req.StartAt != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", req.StartAt)
		startAt = &t
	}
	if req.EndAt != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", req.EndAt)
		endAt = &t
	}

	coupon, err := CouponSvc.UpdateCoupon(
		uint(id), req.Name, req.Type, req.Value,
		req.MinAmount, req.MaxDiscount,
		req.TotalCount, req.PerUserLimit,
		req.ProductIDs, req.CategoryIDs,
		startAt, endAt, req.Status,
	)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(c.GetString("admin_username"), "update", "coupon", idStr, req.Name, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "coupon": coupon})
}

// AdminDeleteCoupon 删除优惠券
func AdminDeleteCoupon(c *gin.Context) {
	if CouponSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := CouponSvc.DeleteCoupon(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(c.GetString("admin_username"), "delete", "coupon", idStr, "", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "优惠券已删除"})
}

// AdminGetCouponUsages 获取优惠券使用记录
func AdminGetCouponUsages(c *gin.Context) {
	if CouponSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	usages, total, err := CouponSvc.GetCouponUsages(uint(id), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"usages":  usages,
		"total":   total,
		"page":    page,
	})
}

// ==================== 用户端优惠券验证 ====================

// ValidateCoupon 验证优惠券
func ValidateCoupon(c *gin.Context) {
	if CouponSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Code       string  `json:"code" binding:"required"`
		ProductID  uint    `json:"product_id" binding:"required"`
		CategoryID uint    `json:"category_id"`
		Amount     float64 `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	coupon, discount, err := CouponSvc.ValidateCoupon(req.Code, userID, req.ProductID, req.CategoryID, req.Amount)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":      true,
		"coupon_id":    coupon.ID,
		"coupon_name":  coupon.Name,
		"coupon_type":  coupon.Type,
		"discount":     discount,
		"final_amount": req.Amount - discount,
	})
}
