package api

import (
	"strconv"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 管理员余额告警 API ====================

// AdminGetBalanceAlerts 管理员获取余额告警列表
func AdminGetBalanceAlerts(c *gin.Context) {
	if BalanceAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	alertType := c.Query("alert_type")
	level := c.Query("level")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	alerts, total, err := BalanceAlertSvc.GetAlerts(page, pageSize, alertType, level, status)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取告警列表失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    alerts,
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// AdminGetBalanceAlertDetail 管理员获取告警详情
func AdminGetBalanceAlertDetail(c *gin.Context) {
	if BalanceAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的告警ID"})
		return
	}

	alert, err := BalanceAlertSvc.GetAlertByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "告警不存在"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": alert})
}

// AdminHandleBalanceAlert 管理员处理告警
func AdminHandleBalanceAlert(c *gin.Context) {
	if BalanceAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的告警ID"})
		return
	}

	var req struct {
		Status int    `json:"status" binding:"required,oneof=1 2"` // 1=已处理 2=已忽略
		Remark string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误，状态必须为1(已处理)或2(已忽略)"})
		return
	}

	adminID, _ := c.Get("admin_id")
	if err := BalanceAlertSvc.HandleAlert(uint(id), adminID.(uint), req.Status, req.Remark); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "处理失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		action := "处理余额告警"
		if req.Status == model.AlertStatusIgnored {
			action = "忽略余额告警"
		}
		LogSvc.LogAdminActionSimple(adminUsername.(string), action, "balance_alert", strconv.FormatUint(id, 10), req, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "处理成功"})
}

// AdminGetBalanceAlertStats 管理员获取告警统计
func AdminGetBalanceAlertStats(c *gin.Context) {
	if BalanceAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats, err := BalanceAlertSvc.GetAlertStats()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取统计失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": stats})
}

// AdminBatchCheckBalanceMismatch 管理员触发批量检查余额不一致
func AdminBatchCheckBalanceMismatch(c *gin.Context) {
	if BalanceAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	count, err := BalanceAlertSvc.BatchCheckBalanceMismatch()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "检查失败: " + err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "批量检查余额一致性", "balance_alert", "", map[string]int{"alert_count": count}, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "检查完成",
		"data": map[string]int{
			"alert_count": count,
		},
	})
}

// AdminCleanOldAlerts 管理员清理旧告警
func AdminCleanOldAlerts(c *gin.Context) {
	if BalanceAlertSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Days int `json:"days" binding:"required,min=7"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误，天数必须大于等于7"})
		return
	}

	count, err := BalanceAlertSvc.CleanOldAlerts(req.Days)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "清理失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "清理旧余额告警", "balance_alert", "", map[string]interface{}{"days": req.Days, "deleted_count": count}, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "清理完成",
		"data": map[string]int64{
			"deleted_count": count,
		},
	})
}
