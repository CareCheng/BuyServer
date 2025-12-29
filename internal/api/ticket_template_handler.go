package api

import (
	"strconv"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetTicketTemplates 获取工单模板列表（用户端）
// GET /api/support/templates
func GetTicketTemplates(c *gin.Context) {
	category := c.Query("category")

	if TicketTemplateSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	templates, err := TicketTemplateSvc.GetTemplates(category, true)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取模板失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": templates})
}

// GetTicketTemplatesByCategory 按分类获取工单模板
// GET /api/support/templates/by-category
func GetTicketTemplatesByCategory(c *gin.Context) {
	if TicketTemplateSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	templates, err := TicketTemplateSvc.GetTemplatesByCategory()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取模板失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": templates})
}

// GetHotTicketTemplates 获取热门工单模板
// GET /api/support/templates/hot
func GetHotTicketTemplates(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit < 1 || limit > 20 {
		limit = 5
	}

	if TicketTemplateSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	templates, err := TicketTemplateSvc.GetHotTemplates(limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取模板失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": templates})
}

// GetTicketTemplateDetail 获取工单模板详情
// GET /api/support/template/:id
func GetTicketTemplateDetail(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "模板ID无效"})
		return
	}

	if TicketTemplateSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	template, err := TicketTemplateSvc.GetTemplate(uint(templateID))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 解析自定义字段
	fields, _ := TicketTemplateSvc.ParseTemplateFields(template.Fields)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"template": template,
			"fields":   fields,
		},
	})
}

// UseTicketTemplate 使用工单模板（增加使用次数）
// POST /api/support/template/:id/use
func UseTicketTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "模板ID无效"})
		return
	}

	if TicketTemplateSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := TicketTemplateSvc.IncrementUseCount(uint(templateID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "操作失败"})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

// AdminGetTicketTemplates 管理员获取工单模板列表
// GET /api/admin/support/templates
func AdminGetTicketTemplates(c *gin.Context) {
	category := c.Query("category")

	if TicketTemplateSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	templates, err := TicketTemplateSvc.GetTemplates(category, false)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取模板失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": templates})
}

// AdminCreateTicketTemplate 管理员创建工单模板
// POST /api/admin/support/template
func AdminCreateTicketTemplate(c *gin.Context) {
	var req struct {
		Name        string                       `json:"name" binding:"required"`
		Description string                       `json:"description"`
		Category    string                       `json:"category"`
		Subject     string                       `json:"subject"`
		Content     string                       `json:"content"`
		Fields      []model.TicketTemplateField  `json:"fields"`
		Icon        string                       `json:"icon"`
		SortOrder   int                          `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if TicketTemplateSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	template, err := TicketTemplateSvc.CreateTemplate(
		req.Name, req.Description, req.Category, req.Subject, req.Content,
		req.Fields, req.Icon, req.SortOrder,
	)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "create_template", "ticket_template", strconv.Itoa(int(template.ID)), "创建工单模板: "+req.Name, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": template})
}

// AdminUpdateTicketTemplate 管理员更新工单模板
// PUT /api/admin/support/template/:id
func AdminUpdateTicketTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "模板ID无效"})
		return
	}

	var req struct {
		Name        string                       `json:"name"`
		Description string                       `json:"description"`
		Category    string                       `json:"category"`
		Subject     string                       `json:"subject"`
		Content     string                       `json:"content"`
		Fields      []model.TicketTemplateField  `json:"fields"`
		Icon        string                       `json:"icon"`
		SortOrder   int                          `json:"sort_order"`
		Status      int                          `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if TicketTemplateSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"category":    req.Category,
		"subject":     req.Subject,
		"content":     req.Content,
		"icon":        req.Icon,
		"sort_order":  req.SortOrder,
		"status":      req.Status,
	}

	if len(req.Fields) > 0 {
		updates["fields"] = req.Fields
	}

	if err := TicketTemplateSvc.UpdateTemplate(uint(templateID), updates); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "update_template", "ticket_template", strconv.Itoa(int(templateID)), "更新工单模板", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// AdminDeleteTicketTemplate 管理员删除工单模板
// DELETE /api/admin/support/template/:id
func AdminDeleteTicketTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "模板ID无效"})
		return
	}

	if TicketTemplateSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := TicketTemplateSvc.DeleteTemplate(uint(templateID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername := c.GetString("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "delete_template", "ticket_template", strconv.Itoa(int(templateID)), "删除工单模板", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}
