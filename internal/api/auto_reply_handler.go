package api

import (
	"strconv"
	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ========== 用户端智能客服 API ==========

// GetAutoReplyWelcome 获取智能客服欢迎语
func GetAutoReplyWelcome(c *gin.Context) {
	if AutoReplySvc == nil {
		c.JSON(200, gin.H{"success": true, "enabled": false})
		return
	}

	config, _ := AutoReplySvc.GetConfig()
	c.JSON(200, gin.H{
		"success":         true,
		"enabled":         config.Enabled,
		"welcome_message": config.WelcomeMessage,
	})
}

// SendAutoReplyMessage 发送消息给智能客服
func SendAutoReplyMessage(c *gin.Context) {
	if AutoReplySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Message   string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取用户ID（可能是游客）
	var userID uint
	if id, exists := c.Get("user_id"); exists {
		userID = id.(uint)
	}

	result, err := AutoReplySvc.ProcessMessage(req.SessionID, userID, req.Message)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "处理消息失败"})
		return
	}

	if result == nil {
		c.JSON(200, gin.H{
			"success": true,
			"enabled": false,
			"message": "智能客服未启用",
		})
		return
	}

	c.JSON(200, gin.H{
		"success":     true,
		"reply":       result.Reply,
		"matched":     result.Matched,
		"transferred": result.Transferred,
	})
}

// ========== 管理端智能客服 API ==========

// AdminGetAutoReplyConfig 获取智能客服配置
func AdminGetAutoReplyConfig(c *gin.Context) {
	if AutoReplySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	config, err := AutoReplySvc.GetConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取配置失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"config":  config,
	})
}

// AdminSaveAutoReplyConfig 保存智能客服配置
func AdminSaveAutoReplyConfig(c *gin.Context) {
	if AutoReplySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var config model.AutoReplyConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := AutoReplySvc.SaveConfig(&config); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存配置失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "更新智能客服配置", "auto_reply_config", "", nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{
		"success": true,
		"message": "配置已保存",
	})
}

// AdminGetAutoReplyRules 获取自动回复规则列表
func AdminGetAutoReplyRules(c *gin.Context) {
	if AutoReplySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
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

	rules, total, err := AutoReplySvc.GetRules(page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取规则列表失败"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(200, gin.H{
		"success":     true,
		"rules":       rules,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// AdminCreateAutoReplyRule 创建自动回复规则
func AdminCreateAutoReplyRule(c *gin.Context) {
	if AutoReplySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		Keywords  string `json:"keywords" binding:"required"`
		MatchType string `json:"match_type"`
		Reply     string `json:"reply" binding:"required"`
		Priority  int    `json:"priority"`
		Status    int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if req.MatchType == "" {
		req.MatchType = "contains"
	}

	rule := &model.AutoReplyRule{
		Name:      req.Name,
		Keywords:  req.Keywords,
		MatchType: req.MatchType,
		Reply:     req.Reply,
		Priority:  req.Priority,
		Status:    req.Status,
	}

	if err := AutoReplySvc.CreateRule(rule); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建规则失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "创建自动回复规则", "auto_reply_rule", "", req.Name, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{
		"success": true,
		"message": "规则已创建",
		"rule":    rule,
	})
}

// AdminUpdateAutoReplyRule 更新自动回复规则
func AdminUpdateAutoReplyRule(c *gin.Context) {
	if AutoReplySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	rule, err := AutoReplySvc.GetRule(uint(ruleID))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "规则不存在"})
		return
	}

	var req struct {
		Name      string `json:"name"`
		Keywords  string `json:"keywords"`
		MatchType string `json:"match_type"`
		Reply     string `json:"reply"`
		Priority  int    `json:"priority"`
		Status    int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.Keywords != "" {
		rule.Keywords = req.Keywords
	}
	if req.MatchType != "" {
		rule.MatchType = req.MatchType
	}
	if req.Reply != "" {
		rule.Reply = req.Reply
	}
	rule.Priority = req.Priority
	rule.Status = req.Status

	if err := AutoReplySvc.UpdateRule(rule); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新规则失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "更新自动回复规则", "auto_reply_rule", c.Param("id"), nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{
		"success": true,
		"message": "规则已更新",
	})
}

// AdminDeleteAutoReplyRule 删除自动回复规则
func AdminDeleteAutoReplyRule(c *gin.Context) {
	if AutoReplySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := AutoReplySvc.DeleteRule(uint(ruleID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除规则失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "删除自动回复规则", "auto_reply_rule", c.Param("id"), nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{
		"success": true,
		"message": "规则已删除",
	})
}

// AdminGetAutoReplyLogs 获取自动回复日志
func AdminGetAutoReplyLogs(c *gin.Context) {
	if AutoReplySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
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

	logs, total, err := AutoReplySvc.GetLogs(page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取日志失败"})
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

// AdminGetAutoReplyStats 获取智能客服统计
func AdminGetAutoReplyStats(c *gin.Context) {
	if AutoReplySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats := AutoReplySvc.GetStats()
	c.JSON(200, gin.H{
		"success": true,
		"stats":   stats,
	})
}
