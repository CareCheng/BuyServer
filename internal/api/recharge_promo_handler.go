// Package api 提供 HTTP API 处理器
// recharge_promo_handler.go - 充值优惠活动 API
package api

import (
	"strconv"
	"time"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 管理员 API ====================

// AdminGetRechargePromos 管理员获取充值优惠活动列表
func AdminGetRechargePromos(c *gin.Context) {
	if RechargePromoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
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

	promos, total, err := RechargePromoSvc.GetAllPromos(page, pageSize, status)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取列表失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    promos,
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// AdminCreateRechargePromo 管理员创建充值优惠活动
func AdminCreateRechargePromo(c *gin.Context) {
	if RechargePromoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Name         string   `json:"name" binding:"required"`
		Description  string   `json:"description"`
		PromoType    string   `json:"promo_type" binding:"required"`
		MinAmount    float64  `json:"min_amount"`
		MaxAmount    float64  `json:"max_amount"`
		Value        float64  `json:"value" binding:"required"`
		MaxBonus     float64  `json:"max_bonus"`
		Priority     int      `json:"priority"`
		PerUserLimit int      `json:"per_user_limit"`
		TotalLimit   int      `json:"total_limit"`
		StartAt      *string  `json:"start_at"`
		EndAt        *string  `json:"end_at"`
		Status       int      `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误: " + err.Error()})
		return
	}

	// 验证优惠类型
	validTypes := map[string]bool{
		model.PromoTypeDiscount: true,
		model.PromoTypeBonus:    true,
		model.PromoTypePercent:  true,
	}
	if !validTypes[req.PromoType] {
		c.JSON(400, gin.H{"success": false, "error": "无效的优惠类型"})
		return
	}

	promo := &model.RechargePromo{
		Name:         req.Name,
		Description:  req.Description,
		PromoType:    req.PromoType,
		MinAmount:    req.MinAmount,
		MaxAmount:    req.MaxAmount,
		Value:        req.Value,
		MaxBonus:     req.MaxBonus,
		Priority:     req.Priority,
		PerUserLimit: req.PerUserLimit,
		TotalLimit:   req.TotalLimit,
		Status:       req.Status,
	}

	// 解析时间
	if req.StartAt != nil && *req.StartAt != "" {
		t, err := time.Parse("2006-01-02 15:04:05", *req.StartAt)
		if err == nil {
			promo.StartAt = &t
		}
	}
	if req.EndAt != nil && *req.EndAt != "" {
		t, err := time.Parse("2006-01-02 15:04:05", *req.EndAt)
		if err == nil {
			promo.EndAt = &t
		}
	}

	if err := RechargePromoSvc.CreatePromo(promo); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "创建充值优惠活动", "recharge_promo", strconv.FormatUint(uint64(promo.ID), 10), req, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": promo})
}

// AdminUpdateRechargePromo 管理员更新充值优惠活动
func AdminUpdateRechargePromo(c *gin.Context) {
	if RechargePromoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的ID"})
		return
	}

	promo, err := RechargePromoSvc.GetPromoByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "活动不存在"})
		return
	}

	var req struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		PromoType    string   `json:"promo_type"`
		MinAmount    float64  `json:"min_amount"`
		MaxAmount    float64  `json:"max_amount"`
		Value        float64  `json:"value"`
		MaxBonus     float64  `json:"max_bonus"`
		Priority     int      `json:"priority"`
		PerUserLimit int      `json:"per_user_limit"`
		TotalLimit   int      `json:"total_limit"`
		StartAt      *string  `json:"start_at"`
		EndAt        *string  `json:"end_at"`
		Status       int      `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 更新字段
	if req.Name != "" {
		promo.Name = req.Name
	}
	promo.Description = req.Description
	if req.PromoType != "" {
		promo.PromoType = req.PromoType
	}
	promo.MinAmount = req.MinAmount
	promo.MaxAmount = req.MaxAmount
	if req.Value > 0 {
		promo.Value = req.Value
	}
	promo.MaxBonus = req.MaxBonus
	promo.Priority = req.Priority
	promo.PerUserLimit = req.PerUserLimit
	promo.TotalLimit = req.TotalLimit
	promo.Status = req.Status

	// 解析时间
	if req.StartAt != nil {
		if *req.StartAt == "" {
			promo.StartAt = nil
		} else {
			t, err := time.Parse("2006-01-02 15:04:05", *req.StartAt)
			if err == nil {
				promo.StartAt = &t
			}
		}
	}
	if req.EndAt != nil {
		if *req.EndAt == "" {
			promo.EndAt = nil
		} else {
			t, err := time.Parse("2006-01-02 15:04:05", *req.EndAt)
			if err == nil {
				promo.EndAt = &t
			}
		}
	}

	if err := RechargePromoSvc.UpdatePromo(promo); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "更新充值优惠活动", "recharge_promo", strconv.FormatUint(id, 10), req, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": promo})
}

// AdminDeleteRechargePromo 管理员删除充值优惠活动
func AdminDeleteRechargePromo(c *gin.Context) {
	if RechargePromoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的ID"})
		return
	}

	if err := RechargePromoSvc.DeletePromo(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "删除充值优惠活动", "recharge_promo", strconv.FormatUint(id, 10), nil, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// AdminToggleRechargePromoStatus 管理员切换优惠活动状态
func AdminToggleRechargePromoStatus(c *gin.Context) {
	if RechargePromoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的ID"})
		return
	}

	if err := RechargePromoSvc.TogglePromoStatus(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "操作失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "状态已切换"})
}

// AdminGetRechargePromoUsages 管理员获取优惠使用记录
func AdminGetRechargePromoUsages(c *gin.Context) {
	if RechargePromoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	promoID, _ := strconv.ParseUint(c.Query("promo_id"), 10, 32)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	usages, total, err := RechargePromoSvc.GetPromoUsages(uint(promoID), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取记录失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    usages,
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// AdminGetRechargePromoStats 管理员获取优惠统计
func AdminGetRechargePromoStats(c *gin.Context) {
	if RechargePromoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats, err := RechargePromoSvc.GetPromoStats()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取统计失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": stats})
}

// ==================== 用户端 API ====================

// GetActiveRechargePromos 获取当前有效的充值优惠活动
func GetActiveRechargePromos(c *gin.Context) {
	if RechargePromoSvc == nil {
		c.JSON(200, gin.H{"success": true, "data": []interface{}{}})
		return
	}

	promos, err := RechargePromoSvc.GetActivePromos()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取优惠活动失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": promos})
}

// CalculateRechargePromo 计算充值优惠
func CalculateRechargePromo(c *gin.Context) {
	if RechargePromoSvc == nil {
		c.JSON(200, gin.H{
			"success": true,
			"data": gin.H{
				"original_amount": 0,
				"pay_amount":      0,
				"bonus_amount":    0,
				"discount_amount": 0,
				"total_credit":    0,
			},
		})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请输入有效的充值金额"})
		return
	}

	// 获取用户ID（可选，未登录时为0）
	var userID uint = 0
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	result, err := RechargePromoSvc.CalculatePromo(userID, req.Amount)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "计算优惠失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": result})
}

// GetAllApplicablePromos 获取所有适用的优惠方案
func GetAllApplicablePromos(c *gin.Context) {
	if RechargePromoSvc == nil {
		c.JSON(200, gin.H{"success": true, "data": []interface{}{}})
		return
	}

	amount, err := strconv.ParseFloat(c.Query("amount"), 64)
	if err != nil || amount <= 0 {
		c.JSON(400, gin.H{"success": false, "error": "请输入有效的充值金额"})
		return
	}

	// 获取用户ID（可选）
	var userID uint = 0
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	results, err := RechargePromoSvc.CalculateAllPromos(userID, amount)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取优惠方案失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": results})
}
