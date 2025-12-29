package api

import (
	"strconv"
	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetMyPoints 获取我的积分
func GetMyPoints(c *gin.Context) {
	if PointsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	points, err := PointsSvc.GetUserPoints(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取积分失败"})
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"points":     points.Points,
		"total_earn": points.TotalEarn,
		"total_used": points.TotalUsed,
	})
}

// GetPointsLogs 获取积分记录
func GetPointsLogs(c *gin.Context) {
	if PointsSvc == nil {
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
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := PointsSvc.GetPointsLogs(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取记录失败"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(200, gin.H{
		"success":     true,
		"logs":        logs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetExchangeList 获取可兑换列表
func GetExchangeList(c *gin.Context) {
	if PointsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	list, err := PointsSvc.GetExchangeList()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取列表失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "list": list})
}

// ExchangeCoupon 兑换优惠券
func ExchangeCoupon(c *gin.Context) {
	if PointsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	var req struct {
		CouponID uint `json:"coupon_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	exchange, err := PointsSvc.ExchangeCoupon(userID.(uint), req.CouponID)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":  true,
		"message":  "兑换成功",
		"exchange": exchange,
	})
}

// GetMyExchanges 获取我的兑换记录
func GetMyExchanges(c *gin.Context) {
	if PointsSvc == nil {
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
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	exchanges, total, err := PointsSvc.GetUserExchanges(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取记录失败"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(200, gin.H{
		"success":     true,
		"exchanges":   exchanges,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// AdminGetPointsRules 获取积分规则（管理员）
func AdminGetPointsRules(c *gin.Context) {
	if PointsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	rules, err := PointsSvc.GetPointsRules()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取规则失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "rules": rules})
}

// AdminCreatePointsRule 创建积分规则（管理员）
func AdminCreatePointsRule(c *gin.Context) {
	if PointsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Name        string  `json:"name" binding:"required"`
		Type        string  `json:"type" binding:"required"`
		Points      int     `json:"points"`
		Ratio       float64 `json:"ratio"`
		MinAmount   float64 `json:"min_amount"`
		MaxPoints   int     `json:"max_points"`
		Status      int     `json:"status"`
		Description string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	rule := &model.PointsRule{
		Name:        req.Name,
		Type:        req.Type,
		Points:      req.Points,
		Ratio:       req.Ratio,
		MinAmount:   req.MinAmount,
		MaxPoints:   req.MaxPoints,
		Status:      req.Status,
		Description: req.Description,
	}

	if err := PointsSvc.CreatePointsRule(rule); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "创建积分规则", "points_rule", "", "创建积分规则: "+req.Name, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{"success": true, "message": "创建成功", "rule": rule})
}

// AdminUpdatePointsRule 更新积分规则（管理员）
func AdminUpdatePointsRule(c *gin.Context) {
	if PointsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Points      int     `json:"points"`
		Ratio       float64 `json:"ratio"`
		MinAmount   float64 `json:"min_amount"`
		MaxPoints   int     `json:"max_points"`
		Status      int     `json:"status"`
		Description string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	rule := &model.PointsRule{
		Name:        req.Name,
		Points:      req.Points,
		Ratio:       req.Ratio,
		MinAmount:   req.MinAmount,
		MaxPoints:   req.MaxPoints,
		Status:      req.Status,
		Description: req.Description,
	}
	rule.ID = uint(ruleID)

	if err := PointsSvc.UpdatePointsRule(rule); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "更新积分规则", "points_rule", c.Param("id"), nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// AdminDeletePointsRule 删除积分规则（管理员）
func AdminDeletePointsRule(c *gin.Context) {
	if PointsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := PointsSvc.DeletePointsRule(uint(ruleID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "删除积分规则", "points_rule", c.Param("id"), nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// AdminAdjustPoints 管理员调整用户积分
func AdminAdjustPoints(c *gin.Context) {
	if PointsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		UserID uint   `json:"user_id" binding:"required"`
		Points int    `json:"points" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if req.Remark == "" {
		req.Remark = "管理员调整"
	}

	if err := PointsSvc.AdminAdjustPoints(req.UserID, req.Points, req.Remark); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "调整用户积分", "user_points", "", req.Remark, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{"success": true, "message": "调整成功"})
}


// AdminGetPointsUsers 管理员获取用户积分列表
func AdminGetPointsUsers(c *gin.Context) {
	if PointsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := PointsSvc.AdminGetAllUserPoints(page, pageSize, keyword)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取用户积分列表失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"users":   users,
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// AdminGetPointsLogs 管理员获取积分变动记录
func AdminGetPointsLogs(c *gin.Context) {
	if PointsSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := PointsSvc.AdminGetPointsLogs(page, pageSize, uint(userID))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取积分记录失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"logs":    logs,
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}
