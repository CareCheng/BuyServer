package api

import (
	"strconv"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminBatchDeleteProducts 批量删除商品
// POST /api/admin/products/batch-delete
func AdminBatchDeleteProducts(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请选择要删除的商品"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 批量删除
	result := model.DB.Where("id IN ?", req.IDs).Delete(&model.Product{})
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "batch_delete_products", "product", "", "批量删除商品: "+strconv.Itoa(len(req.IDs))+"个", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "成功删除 " + strconv.Itoa(int(result.RowsAffected)) + " 个商品",
		"count":   result.RowsAffected,
	})
}

// AdminBatchUpdateProductStatus 批量更新商品状态
// POST /api/admin/products/batch-status
func AdminBatchUpdateProductStatus(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Status int    `json:"status" binding:"oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请选择要操作的商品"})
		return
	}

	result := model.DB.Model(&model.Product{}).Where("id IN ?", req.IDs).Update("status", req.Status)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	action := "batch_disable_products"
	if req.Status == 1 {
		action = "batch_enable_products"
	}
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, action, "product", "", "批量更新商品状态: "+strconv.Itoa(len(req.IDs))+"个", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "成功更新 " + strconv.Itoa(int(result.RowsAffected)) + " 个商品",
		"count":   result.RowsAffected,
	})
}

// AdminBatchDeleteUsers 批量删除用户
// POST /api/admin/users/batch-delete
func AdminBatchDeleteUsers(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请选择要删除的用户"})
		return
	}

	// 批量删除
	result := model.DB.Where("id IN ?", req.IDs).Delete(&model.User{})
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "batch_delete_users", "user", "", "批量删除用户: "+strconv.Itoa(len(req.IDs))+"个", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "成功删除 " + strconv.Itoa(int(result.RowsAffected)) + " 个用户",
		"count":   result.RowsAffected,
	})
}

// AdminBatchUpdateUserStatus 批量更新用户状态
// POST /api/admin/users/batch-status
func AdminBatchUpdateUserStatus(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Status int    `json:"status" binding:"oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请选择要操作的用户"})
		return
	}

	result := model.DB.Model(&model.User{}).Where("id IN ?", req.IDs).Update("status", req.Status)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	action := "batch_disable_users"
	if req.Status == 1 {
		action = "batch_enable_users"
	}
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, action, "user", "", "批量更新用户状态: "+strconv.Itoa(len(req.IDs))+"个", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "成功更新 " + strconv.Itoa(int(result.RowsAffected)) + " 个用户",
		"count":   result.RowsAffected,
	})
}

// AdminBatchDeleteOrders 批量删除订单
// POST /api/admin/orders/batch-delete
func AdminBatchDeleteOrders(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请选择要删除的订单"})
		return
	}

	// 批量删除
	result := model.DB.Where("id IN ?", req.IDs).Delete(&model.Order{})
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "batch_delete_orders", "order", "", "批量删除订单: "+strconv.Itoa(len(req.IDs))+"个", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "成功删除 " + strconv.Itoa(int(result.RowsAffected)) + " 个订单",
		"count":   result.RowsAffected,
	})
}

// AdminBatchDeleteCoupons 批量删除优惠券
// POST /api/admin/coupons/batch-delete
func AdminBatchDeleteCoupons(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请选择要删除的优惠券"})
		return
	}

	// 批量删除
	result := model.DB.Where("id IN ?", req.IDs).Delete(&model.Coupon{})
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "batch_delete_coupons", "coupon", "", "批量删除优惠券: "+strconv.Itoa(len(req.IDs))+"个", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "成功删除 " + strconv.Itoa(int(result.RowsAffected)) + " 个优惠券",
		"count":   result.RowsAffected,
	})
}

// AdminBatchUpdateCouponStatus 批量更新优惠券状态
// POST /api/admin/coupons/batch-status
func AdminBatchUpdateCouponStatus(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Status int    `json:"status" binding:"oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请选择要操作的优惠券"})
		return
	}

	result := model.DB.Model(&model.Coupon{}).Where("id IN ?", req.IDs).Update("status", req.Status)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	action := "batch_disable_coupons"
	if req.Status == 1 {
		action = "batch_enable_coupons"
	}
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, action, "coupon", "", "批量更新优惠券状态: "+strconv.Itoa(len(req.IDs))+"个", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "成功更新 " + strconv.Itoa(int(result.RowsAffected)) + " 个优惠券",
		"count":   result.RowsAffected,
	})
}

// AdminBatchDeleteAnnouncements 批量删除公告
// POST /api/admin/announcements/batch-delete
func AdminBatchDeleteAnnouncements(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请选择要删除的公告"})
		return
	}

	// 批量删除
	result := model.DB.Where("id IN ?", req.IDs).Delete(&model.Announcement{})
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "batch_delete_announcements", "announcement", "", "批量删除公告: "+strconv.Itoa(len(req.IDs))+"个", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "成功删除 " + strconv.Itoa(int(result.RowsAffected)) + " 个公告",
		"count":   result.RowsAffected,
	})
}

// AdminBatchDeleteCategories 批量删除分类
// POST /api/admin/categories/batch-delete
func AdminBatchDeleteCategories(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请选择要删除的分类"})
		return
	}

	// 检查分类下是否有商品
	var productCount int64
	model.DB.Model(&model.Product{}).Where("category_id IN ?", req.IDs).Count(&productCount)
	if productCount > 0 {
		c.JSON(400, gin.H{"success": false, "error": "所选分类下存在商品，无法删除"})
		return
	}

	// 批量删除
	result := model.DB.Where("id IN ?", req.IDs).Delete(&model.ProductCategory{})
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "batch_delete_categories", "category", "", "批量删除分类: "+strconv.Itoa(len(req.IDs))+"个", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "成功删除 " + strconv.Itoa(int(result.RowsAffected)) + " 个分类",
		"count":   result.RowsAffected,
	})
}
