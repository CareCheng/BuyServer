package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateProductReview 创建商品评价
// POST /api/review
func CreateProductReview(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	var req struct {
		OrderNo   string   `json:"order_no" binding:"required"`
		ProductID uint     `json:"product_id" binding:"required"`
		Rating    int      `json:"rating" binding:"required,min=1,max=5"`
		Content   string   `json:"content"`
		Images    []string `json:"images"`
		IsAnon    bool     `json:"is_anon"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if ReviewSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	review, err := ReviewSvc.CreateReview(userID, username, req.OrderNo, req.ProductID, req.Rating, req.Content, req.Images, req.IsAnon)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogUserActionSimple(userID, username, "create_review", "product_review", strconv.Itoa(int(review.ID)), "创建商品评价", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": review})
}

// GetProductReviews 获取商品评价列表
// GET /api/product/:id/reviews
func GetProductReviews(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "商品ID无效"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	rating, _ := strconv.Atoi(c.DefaultQuery("rating", "0"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	if ReviewSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	reviews, total, err := ReviewSvc.GetProductReviews(uint(productID), page, pageSize, rating)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取评价失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"reviews":   reviews,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetProductReviewStats 获取商品评价统计
// GET /api/product/:id/review-stats
func GetProductReviewStats(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "商品ID无效"})
		return
	}

	if ReviewSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats, err := ReviewSvc.GetProductReviewStats(uint(productID))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取统计失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": stats})
}

// GetUserReviews 获取用户的评价列表
// GET /api/user/reviews
func GetUserReviews(c *gin.Context) {
	userID := c.GetUint("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	if ReviewSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	reviews, total, err := ReviewSvc.GetUserReviews(userID, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取评价失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"reviews":   reviews,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// CheckCanReview 检查是否可以评价订单
// GET /api/order/:order_no/can-review
func CheckCanReview(c *gin.Context) {
	userID := c.GetUint("user_id")
	orderNo := c.Param("order_no")

	if ReviewSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	canReview, reason := ReviewSvc.CheckCanReview(userID, orderNo)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"can_review": canReview,
			"reason":     reason,
		},
	})
}

// AdminGetReviews 管理员获取评价列表
// GET /api/admin/reviews
func AdminGetReviews(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	productID, _ := strconv.ParseUint(c.DefaultQuery("product_id", "0"), 10, 32)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	if ReviewSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	reviews, total, err := ReviewSvc.GetAllReviews(page, pageSize, uint(productID), status)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取评价失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"reviews":   reviews,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// AdminReplyReview 管理员回复评价
// POST /api/admin/review/:id/reply
func AdminReplyReview(c *gin.Context) {
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "评价ID无效"})
		return
	}

	var req struct {
		Reply string `json:"reply" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if ReviewSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := ReviewSvc.ReplyReview(uint(reviewID), req.Reply); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "回复失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "reply_review", "product_review", strconv.Itoa(int(reviewID)), "回复商品评价", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "回复成功"})
}

// AdminUpdateReviewStatus 管理员更新评价状态
// PUT /api/admin/review/:id/status
func AdminUpdateReviewStatus(c *gin.Context) {
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "评价ID无效"})
		return
	}

	var req struct {
		Status int `json:"status" binding:"oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if ReviewSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := ReviewSvc.UpdateReviewStatus(uint(reviewID), req.Status); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	action := "hide_review"
	if req.Status == 1 {
		action = "show_review"
	}
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, action, "product_review", strconv.Itoa(int(reviewID)), "更新评价状态", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// AdminDeleteReview 管理员删除评价
// DELETE /api/admin/review/:id
func AdminDeleteReview(c *gin.Context) {
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "评价ID无效"})
		return
	}

	if ReviewSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := ReviewSvc.DeleteReview(uint(reviewID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "delete_review", "product_review", strconv.Itoa(int(reviewID)), "删除商品评价", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}
