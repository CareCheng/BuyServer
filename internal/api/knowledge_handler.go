package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ==================== 客服端知识库 API ====================

// StaffGetKnowledgeCategories 客服获取知识库分类
// GET /api/staff/knowledge/categories
func StaffGetKnowledgeCategories(c *gin.Context) {
	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	categories, err := KnowledgeSvc.GetAllCategories(true)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取分类失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": categories})
}

// StaffGetKnowledgeArticles 客服获取知识库文章列表
// GET /api/staff/knowledge/articles
func StaffGetKnowledgeArticles(c *gin.Context) {
	categoryID, _ := strconv.ParseUint(c.DefaultQuery("category_id", "0"), 10, 32)
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 只获取已发布的文章
	articles, total, err := KnowledgeSvc.GetArticles(uint(categoryID), keyword, 1, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取文章失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"articles":  articles,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// StaffGetKnowledgeArticle 客服获取知识库文章详情
// GET /api/staff/knowledge/article/:id
func StaffGetKnowledgeArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "文章ID无效"})
		return
	}

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	article, err := KnowledgeSvc.GetArticle(uint(articleID))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 增加浏览次数
	KnowledgeSvc.IncrementViewCount(uint(articleID))

	c.JSON(200, gin.H{"success": true, "data": article})
}

// StaffSearchKnowledge 客服搜索知识库
// GET /api/staff/knowledge/search
func StaffSearchKnowledge(c *gin.Context) {
	keyword := c.Query("keyword")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if limit < 1 || limit > 50 {
		limit = 10
	}

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	articles, err := KnowledgeSvc.SearchArticles(keyword, limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "搜索失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": articles})
}

// StaffGetHotKnowledge 客服获取热门知识
// GET /api/staff/knowledge/hot
func StaffGetHotKnowledge(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if limit < 1 || limit > 50 {
		limit = 10
	}

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	articles, err := KnowledgeSvc.GetHotArticles(limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": articles})
}

// StaffUseKnowledge 客服使用知识库文章（引用到回复）
// POST /api/staff/knowledge/article/:id/use
func StaffUseKnowledge(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "文章ID无效"})
		return
	}

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := KnowledgeSvc.IncrementUseCount(uint(articleID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "操作失败"})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

// ==================== 管理员知识库 API ====================

// AdminGetKnowledgeCategories 管理员获取知识库分类
// GET /api/admin/knowledge/categories
func AdminGetKnowledgeCategories(c *gin.Context) {
	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	categories, err := KnowledgeSvc.GetAllCategories(false)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取分类失败"})
		return
	}

	// 返回 categories 字段以匹配前端期望
	c.JSON(200, gin.H{"success": true, "categories": categories})
}

// AdminCreateKnowledgeCategory 管理员创建知识库分类
// POST /api/admin/knowledge/category
func AdminCreateKnowledgeCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		ParentID    uint   `json:"parent_id"`
		SortOrder   int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	category, err := KnowledgeSvc.CreateCategory(req.Name, req.Description, req.Icon, req.ParentID, req.SortOrder)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "create_knowledge_category", "knowledge_category", strconv.Itoa(int(category.ID)), "创建知识库分类: "+req.Name, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": category})
}

// AdminUpdateKnowledgeCategory 管理员更新知识库分类
// PUT /api/admin/knowledge/category/:id
func AdminUpdateKnowledgeCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "分类ID无效"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		ParentID    uint   `json:"parent_id"`
		SortOrder   int    `json:"sort_order"`
		Status      int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"icon":        req.Icon,
		"parent_id":   req.ParentID,
		"sort_order":  req.SortOrder,
		"status":      req.Status,
	}

	if err := KnowledgeSvc.UpdateCategory(uint(categoryID), updates); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "update_knowledge_category", "knowledge_category", strconv.Itoa(int(categoryID)), "更新知识库分类", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// AdminDeleteKnowledgeCategory 管理员删除知识库分类
// DELETE /api/admin/knowledge/category/:id
func AdminDeleteKnowledgeCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "分类ID无效"})
		return
	}

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := KnowledgeSvc.DeleteCategory(uint(categoryID)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "delete_knowledge_category", "knowledge_category", strconv.Itoa(int(categoryID)), "删除知识库分类", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// AdminGetKnowledgeArticles 管理员获取知识库文章列表
// GET /api/admin/knowledge/articles
func AdminGetKnowledgeArticles(c *gin.Context) {
	categoryID, _ := strconv.ParseUint(c.DefaultQuery("category_id", "0"), 10, 32)
	keyword := c.Query("keyword")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	articles, total, err := KnowledgeSvc.GetArticles(uint(categoryID), keyword, status, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取文章失败"})
		return
	}

	// 返回 articles 和 total 字段以匹配前端期望
	c.JSON(200, gin.H{
		"success":   true,
		"articles":  articles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AdminCreateKnowledgeArticle 管理员创建知识库文章
// POST /api/admin/knowledge/article
func AdminCreateKnowledgeArticle(c *gin.Context) {
	var req struct {
		CategoryID uint   `json:"category_id"`
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content"`
		Summary    string `json:"summary"`
		Tags       string `json:"tags"`
		SortOrder  int    `json:"sort_order"`
		Status     int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	adminUsername := c.GetString("admin_username")

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	article, err := KnowledgeSvc.CreateArticle(req.CategoryID, req.Title, req.Content, req.Summary, req.Tags, req.SortOrder, req.Status, adminUsername)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "create_knowledge_article", "knowledge_article", strconv.Itoa(int(article.ID)), "创建知识库文章: "+req.Title, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": article})
}

// AdminUpdateKnowledgeArticle 管理员更新知识库文章
// PUT /api/admin/knowledge/article/:id
func AdminUpdateKnowledgeArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "文章ID无效"})
		return
	}

	var req struct {
		CategoryID uint   `json:"category_id"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		Summary    string `json:"summary"`
		Tags       string `json:"tags"`
		SortOrder  int    `json:"sort_order"`
		Status     int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	adminUsername := c.GetString("admin_username")

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	updates := map[string]interface{}{
		"category_id": req.CategoryID,
		"title":       req.Title,
		"content":     req.Content,
		"summary":     req.Summary,
		"tags":        req.Tags,
		"sort_order":  req.SortOrder,
		"status":      req.Status,
	}

	if err := KnowledgeSvc.UpdateArticle(uint(articleID), updates, adminUsername); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "update_knowledge_article", "knowledge_article", strconv.Itoa(int(articleID)), "更新知识库文章", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// AdminDeleteKnowledgeArticle 管理员删除知识库文章
// DELETE /api/admin/knowledge/article/:id
func AdminDeleteKnowledgeArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "文章ID无效"})
		return
	}

	if KnowledgeSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := KnowledgeSvc.DeleteArticle(uint(articleID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "delete_knowledge_article", "knowledge_article", strconv.Itoa(int(articleID)), "删除知识库文章", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}
