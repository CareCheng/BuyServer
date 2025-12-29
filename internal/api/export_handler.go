package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// ==================== 数据导出 API ====================

// AdminExportOrders 导出订单数据
// 管理员可导出指定时间范围内的订单数据为Excel格式
func AdminExportOrders(c *gin.Context) {
	if ExportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 解析参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	status := c.Query("status")

	// 默认导出最近30天
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = t.Add(24*time.Hour - time.Second) // 包含当天
		}
	}

	data, err := ExportSvc.ExportOrders(startDate, endDate, status)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "导出失败: " + err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		adminUsername, _ := c.Get("admin_username")
		LogSvc.LogAdminActionSimple(adminUsername.(string), "export_orders", "order", "",
			fmt.Sprintf("导出订单数据 %s 至 %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
			c.ClientIP(), c.GetHeader("User-Agent"))
	}

	// 设置响应头
	filename := fmt.Sprintf("orders_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// AdminExportUsers 导出用户数据
// 管理员可导出指定时间范围内的用户数据为Excel格式
func AdminExportUsers(c *gin.Context) {
	if ExportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 解析参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// 默认导出最近30天
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = t.Add(24*time.Hour - time.Second)
		}
	}

	data, err := ExportSvc.ExportUsers(startDate, endDate)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "导出失败: " + err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		adminUsername, _ := c.Get("admin_username")
		LogSvc.LogAdminActionSimple(adminUsername.(string), "export_users", "user", "",
			fmt.Sprintf("导出用户数据 %s 至 %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
			c.ClientIP(), c.GetHeader("User-Agent"))
	}

	filename := fmt.Sprintf("users_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// AdminExportLogs 导出操作日志
// 管理员可导出指定时间范围内的操作日志为Excel格式
func AdminExportLogs(c *gin.Context) {
	if ExportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 解析参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	userType := c.Query("user_type")

	// 默认导出最近7天
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = t.Add(24*time.Hour - time.Second)
		}
	}

	data, err := ExportSvc.ExportOperationLogs(startDate, endDate, userType)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "导出失败: " + err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		adminUsername, _ := c.Get("admin_username")
		LogSvc.LogAdminActionSimple(adminUsername.(string), "export_logs", "log", "",
			fmt.Sprintf("导出操作日志 %s 至 %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
			c.ClientIP(), c.GetHeader("User-Agent"))
	}

	filename := fmt.Sprintf("logs_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// AdminExportLoginHistory 导出登录历史
// 管理员可导出指定时间范围内的登录历史为Excel格式
func AdminExportLoginHistory(c *gin.Context) {
	if ExportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 解析参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// 默认导出最近7天
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = t.Add(24*time.Hour - time.Second)
		}
	}

	data, err := ExportSvc.ExportLoginHistory(0, startDate, endDate)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "导出失败: " + err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		adminUsername, _ := c.Get("admin_username")
		LogSvc.LogAdminActionSimple(adminUsername.(string), "export_login_history", "login_history", "",
			fmt.Sprintf("导出登录历史 %s 至 %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
			c.ClientIP(), c.GetHeader("User-Agent"))
	}

	filename := fmt.Sprintf("login_history_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// UserExportOrders 用户导出自己的订单
// 用户可导出自己的订单数据为Excel格式
func UserExportOrders(c *gin.Context) {
	if ExportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	data, err := ExportSvc.ExportUserOrders(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "导出失败: " + err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		username, _ := c.Get("username")
		LogSvc.LogUserActionSimple(userID.(uint), username.(string), "export_orders", "order", "",
			"导出个人订单数据", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	filename := fmt.Sprintf("my_orders_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
