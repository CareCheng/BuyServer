package api

import (
	"strconv"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ==================== 用户端FAQ API ====================

// GetFAQCategories 获取FAQ分类列表
func GetFAQCategories(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	categories, err := FAQSvc.GetCategories()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取分类失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": categories})
}

// GetFAQList 获取FAQ列表
func GetFAQList(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	categoryID, _ := strconv.ParseUint(c.Query("category_id"), 10, 32)
	
	faqs, err := FAQSvc.GetFAQsByCategory(uint(categoryID))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取FAQ列表失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": faqs})
}

// GetFAQDetail 获取FAQ详情
func GetFAQDetail(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的FAQ ID"})
		return
	}

	faq, err := FAQSvc.GetFAQByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "FAQ不存在"})
		return
	}

	// 增加浏览次数
	FAQSvc.IncrementViewCount(uint(id))

	// 获取用户反馈状态
	var feedbackStatus *bool
	userID, _ := c.Get("user_id")
	sessionID := getOrCreateSessionID(c)
	
	if feedback, err := FAQSvc.GetUserFeedback(uint(id), getUserIDFromContext(userID), sessionID); err == nil {
		feedbackStatus = &feedback.Helpful
	}

	c.JSON(200, gin.H{
		"success":  true,
		"data":     faq,
		"feedback": feedbackStatus,
	})
}

// SearchFAQs 搜索FAQ
func SearchFAQs(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(400, gin.H{"success": false, "error": "请输入搜索关键词"})
		return
	}

	faqs, err := FAQSvc.SearchFAQs(keyword)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "搜索失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": faqs})
}

// GetHotFAQs 获取热门FAQ
func GetHotFAQs(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	faqs, err := FAQSvc.GetHotFAQs(limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取热门FAQ失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": faqs})
}

// SubmitFAQFeedback 提交FAQ反馈
func SubmitFAQFeedback(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的FAQ ID"})
		return
	}

	var req struct {
		Helpful bool `json:"helpful"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	userID, _ := c.Get("user_id")
	sessionID := getOrCreateSessionID(c)

	err = FAQSvc.SubmitFeedback(uint(id), getUserIDFromContext(userID), sessionID, req.Helpful)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "提交反馈失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "感谢您的反馈"})
}

// ==================== 管理后台FAQ API ====================

// AdminGetFAQCategories 管理后台获取FAQ分类
func AdminGetFAQCategories(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	categories, err := FAQSvc.GetAllCategories()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取分类失败"})
		return
	}

	// 返回 categories 字段以匹配前端期望
	c.JSON(200, gin.H{"success": true, "categories": categories})
}

// AdminCreateFAQCategory 创建FAQ分类
func AdminCreateFAQCategory(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var category model.FAQCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if category.Name == "" {
		c.JSON(400, gin.H{"success": false, "error": "分类名称不能为空"})
		return
	}

	if err := FAQSvc.CreateCategory(&category); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建分类失败"})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		username, _ := c.Get("admin_username")
		LogSvc.LogAdminActionSimple(username.(string), "create", "faq_category", strconv.Itoa(int(category.ID)),
			"创建FAQ分类: "+category.Name, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": category})
}

// AdminUpdateFAQCategory 更新FAQ分类
func AdminUpdateFAQCategory(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的分类ID"})
		return
	}

	category, err := FAQSvc.GetCategoryByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "分类不存在"})
		return
	}

	var req struct {
		Name      string `json:"name"`
		Icon      string `json:"icon"`
		SortOrder int    `json:"sort_order"`
		Status    int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	category.Icon = req.Icon
	category.SortOrder = req.SortOrder
	category.Status = req.Status

	if err := FAQSvc.UpdateCategory(category); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新分类失败"})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		username, _ := c.Get("admin_username")
		LogSvc.LogAdminActionSimple(username.(string), "update", "faq_category", strconv.Itoa(int(id)),
			"更新FAQ分类: "+category.Name, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": category})
}

// AdminDeleteFAQCategory 删除FAQ分类
func AdminDeleteFAQCategory(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的分类ID"})
		return
	}

	if err := FAQSvc.DeleteCategory(uint(id)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		username, _ := c.Get("admin_username")
		LogSvc.LogAdminActionSimple(username.(string), "delete", "faq_category", strconv.Itoa(int(id)),
			"删除FAQ分类", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// AdminGetFAQs 管理后台获取FAQ列表
func AdminGetFAQs(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	categoryID, _ := strconv.ParseUint(c.Query("category_id"), 10, 32)
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	faqs, total, err := FAQSvc.GetAllFAQs(page, pageSize, uint(categoryID), keyword)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取FAQ列表失败"})
		return
	}

	// 返回 faqs 字段以匹配前端期望
	c.JSON(200, gin.H{
		"success":  true,
		"faqs":     faqs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// AdminCreateFAQ 创建FAQ
func AdminCreateFAQ(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var faq model.FAQ
	if err := c.ShouldBindJSON(&faq); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if faq.Question == "" || faq.Answer == "" {
		c.JSON(400, gin.H{"success": false, "error": "问题和答案不能为空"})
		return
	}

	if err := FAQSvc.CreateFAQ(&faq); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建FAQ失败"})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		username, _ := c.Get("admin_username")
		LogSvc.LogAdminActionSimple(username.(string), "create", "faq", strconv.Itoa(int(faq.ID)),
			"创建FAQ: "+faq.Question, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": faq})
}

// AdminUpdateFAQ 更新FAQ
func AdminUpdateFAQ(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的FAQ ID"})
		return
	}

	faq, err := FAQSvc.GetFAQByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "FAQ不存在"})
		return
	}

	var req struct {
		CategoryID uint   `json:"category_id"`
		Question   string `json:"question"`
		Answer     string `json:"answer"`
		SortOrder  int    `json:"sort_order"`
		Status     int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	faq.CategoryID = req.CategoryID
	if req.Question != "" {
		faq.Question = req.Question
	}
	if req.Answer != "" {
		faq.Answer = req.Answer
	}
	faq.SortOrder = req.SortOrder
	faq.Status = req.Status

	if err := FAQSvc.UpdateFAQ(faq); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新FAQ失败"})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		username, _ := c.Get("admin_username")
		LogSvc.LogAdminActionSimple(username.(string), "update", "faq", strconv.Itoa(int(id)),
			"更新FAQ: "+faq.Question, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": faq})
}

// AdminDeleteFAQ 删除FAQ
func AdminDeleteFAQ(c *gin.Context) {
	if FAQSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的FAQ ID"})
		return
	}

	if err := FAQSvc.DeleteFAQ(uint(id)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除FAQ失败"})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		username, _ := c.Get("admin_username")
		LogSvc.LogAdminActionSimple(username.(string), "delete", "faq", strconv.Itoa(int(id)),
			"删除FAQ", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// ==================== 辅助函数 ====================

// getOrCreateSessionID 获取或创建会话ID（用于游客反馈）
func getOrCreateSessionID(c *gin.Context) string {
	sessionID, err := c.Cookie("faq_session")
	if err != nil || sessionID == "" {
		sessionID = uuid.New().String()
		c.SetCookie("faq_session", sessionID, 86400*30, "/", "", false, true)
	}
	return sessionID
}

// getUserIDFromContext 从上下文获取用户ID
func getUserIDFromContext(userID interface{}) uint {
	if userID == nil {
		return 0
	}
	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}
