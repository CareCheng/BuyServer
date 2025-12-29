package api

import (
	"strconv"
	"user-frontend/internal/model"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// ========== 用户端发票 API ==========

// GetInvoiceConfig 获取发票配置（公开）
func GetInvoiceConfig(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(200, gin.H{"success": true, "enabled": false})
		return
	}

	config, _ := InvoiceSvc.GetConfig()
	c.JSON(200, gin.H{
		"success":          true,
		"enabled":          config.Enabled,
		"min_amount":       config.MinAmount,
		"allow_personal":   config.AllowPersonal,
		"allow_enterprise": config.AllowEnterprise,
		"default_content":  config.DefaultContent,
	})
}

// ApplyInvoice 申请开票
func ApplyInvoice(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	var req struct {
		OrderNo     string `json:"order_no" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Title       string `json:"title" binding:"required"`
		TaxNo       string `json:"tax_no"`
		Email       string `json:"email" binding:"required"`
		Phone       string `json:"phone"`
		Address     string `json:"address"`
		BankName    string `json:"bank_name"`
		BankAccount string `json:"bank_account"`
		Remark      string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 企业发票必须填写税号
	if req.Type == model.InvoiceTypeEnterprise && req.TaxNo == "" {
		c.JSON(400, gin.H{"success": false, "error": "企业发票必须填写税号"})
		return
	}

	applyReq := &service.InvoiceApplyRequest{
		Type:        req.Type,
		Title:       req.Title,
		TaxNo:       req.TaxNo,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		Remark:      req.Remark,
	}

	invoice, err := InvoiceSvc.ApplyInvoice(userID.(uint), req.OrderNo, applyReq)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "发票申请已提交",
		"invoice": invoice,
	})
}

// GetMyInvoices 获取我的发票列表
func GetMyInvoices(c *gin.Context) {
	if InvoiceSvc == nil {
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

	invoices, total, err := InvoiceSvc.GetUserInvoices(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取发票列表失败"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(200, gin.H{
		"success":     true,
		"invoices":    invoices,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetInvoiceDetail 获取发票详情
func GetInvoiceDetail(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	invoiceNo := c.Param("invoice_no")
	invoice, err := InvoiceSvc.GetInvoiceDetail(userID.(uint), invoiceNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "发票不存在"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"invoice": invoice,
	})
}

// CancelInvoice 取消发票申请
func CancelInvoice(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	invoiceNo := c.Param("invoice_no")
	if err := InvoiceSvc.CancelInvoice(userID.(uint), invoiceNo); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "发票申请已取消",
	})
}

// GetMyInvoiceTitles 获取我的发票抬头
func GetMyInvoiceTitles(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	titles, err := InvoiceSvc.GetUserInvoiceTitles(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取抬头列表失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"titles":  titles,
	})
}

// SaveInvoiceTitle 保存发票抬头
func SaveInvoiceTitle(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	var req struct {
		ID          uint   `json:"id"`
		Type        string `json:"type" binding:"required"`
		Title       string `json:"title" binding:"required"`
		TaxNo       string `json:"tax_no"`
		Address     string `json:"address"`
		Phone       string `json:"phone"`
		BankName    string `json:"bank_name"`
		BankAccount string `json:"bank_account"`
		IsDefault   bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	title := &model.InvoiceTitle{
		Type:        req.Type,
		Title:       req.Title,
		TaxNo:       req.TaxNo,
		Address:     req.Address,
		Phone:       req.Phone,
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		IsDefault:   req.IsDefault,
	}
	title.ID = req.ID

	if err := InvoiceSvc.SaveInvoiceTitle(userID.(uint), title); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "保存成功",
		"title":   title,
	})
}

// DeleteInvoiceTitle 删除发票抬头
func DeleteInvoiceTitle(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	titleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := InvoiceSvc.DeleteInvoiceTitle(userID.(uint), uint(titleID)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

// ========== 管理端发票 API ==========

// AdminGetInvoices 管理员获取发票列表
func AdminGetInvoices(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	invoices, total, err := InvoiceSvc.AdminGetInvoices(status, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取发票列表失败"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(200, gin.H{
		"success":     true,
		"invoices":    invoices,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// AdminIssueInvoice 管理员开具发票
func AdminIssueInvoice(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	invoiceNo := c.Param("invoice_no")
	var req struct {
		InvoiceURL string `json:"invoice_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请提供电子发票URL"})
		return
	}

	adminUsername, _ := c.Get("admin_username")
	if err := InvoiceSvc.AdminIssueInvoice(invoiceNo, req.InvoiceURL, adminUsername.(string)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	LogSvc.LogAdminActionSimple(adminUsername.(string), "开具发票", "invoice", invoiceNo, nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{
		"success": true,
		"message": "发票已开具",
	})
}

// AdminRejectInvoice 管理员拒绝发票申请
func AdminRejectInvoice(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	invoiceNo := c.Param("invoice_no")
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请提供拒绝原因"})
		return
	}

	if err := InvoiceSvc.AdminRejectInvoice(invoiceNo, req.Reason); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "拒绝发票", "invoice", invoiceNo, req.Reason, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{
		"success": true,
		"message": "已拒绝发票申请",
	})
}

// AdminGetInvoiceConfig 管理员获取发票配置
func AdminGetInvoiceConfig(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	config, err := InvoiceSvc.GetConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取配置失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"config":  config,
	})
}

// AdminSaveInvoiceConfig 管理员保存发票配置
func AdminSaveInvoiceConfig(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var config model.InvoiceConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := InvoiceSvc.SaveConfig(&config); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存配置失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "更新发票配置", "invoice_config", "", nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{
		"success": true,
		"message": "配置已保存",
	})
}

// AdminGetInvoiceStats 管理员获取发票统计
func AdminGetInvoiceStats(c *gin.Context) {
	if InvoiceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats := InvoiceSvc.GetInvoiceStats()
	c.JSON(200, gin.H{
		"success": true,
		"stats":   stats,
	})
}
