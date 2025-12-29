package api

import (
	"strconv"
	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminGetTasks 获取任务列表
func AdminGetTasks(c *gin.Context) {
	if TaskSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	tasks, err := TaskSvc.GetTasks()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取任务列表失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "tasks": tasks})
}

// AdminGetTaskTypes 获取可用任务类型
func AdminGetTaskTypes(c *gin.Context) {
	if TaskSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	types := TaskSvc.GetAvailableTaskTypes()
	c.JSON(200, gin.H{"success": true, "types": types})
}

// AdminGetTaskStats 获取任务统计
func AdminGetTaskStats(c *gin.Context) {
	if TaskSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats := TaskSvc.GetTaskStats()
	c.JSON(200, gin.H{"success": true, "stats": stats})
}

// AdminGetTaskLogs 获取任务日志
func AdminGetTaskLogs(c *gin.Context) {
	if TaskSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	taskID, _ := strconv.ParseUint(c.DefaultQuery("task_id", "0"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := TaskSvc.GetTaskLogs(uint(taskID), page, pageSize)
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

// AdminCreateTask 创建任务
func AdminCreateTask(c *gin.Context) {
	if TaskSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Type        string `json:"type" binding:"required"`
		CronExpr    string `json:"cron_expr"`
		Config      string `json:"config"`
		Status      int    `json:"status"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	task := &model.ScheduledTask{
		Name:        req.Name,
		Type:        req.Type,
		CronExpr:    req.CronExpr,
		Config:      req.Config,
		Status:      req.Status,
		Description: req.Description,
	}

	if err := TaskSvc.CreateTask(task); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "创建定时任务", "scheduled_task", "", "创建任务: "+req.Name, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{"success": true, "message": "创建成功", "task": task})
}

// AdminUpdateTask 更新任务
func AdminUpdateTask(c *gin.Context) {
	if TaskSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取现有任务
	task, err := TaskSvc.GetTask(uint(taskID))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "任务不存在"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		CronExpr    string `json:"cron_expr"`
		Config      string `json:"config"`
		Status      int    `json:"status"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 更新字段
	if req.Name != "" {
		task.Name = req.Name
	}
	if req.CronExpr != "" {
		task.CronExpr = req.CronExpr
	}
	if req.Config != "" {
		task.Config = req.Config
	}
	task.Status = req.Status
	if req.Description != "" {
		task.Description = req.Description
	}

	if err := TaskSvc.UpdateTask(task); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "更新定时任务", "scheduled_task", c.Param("id"), nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// AdminDeleteTask 删除任务
func AdminDeleteTask(c *gin.Context) {
	if TaskSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := TaskSvc.DeleteTask(uint(taskID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "删除定时任务", "scheduled_task", c.Param("id"), nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// AdminRunTaskNow 立即执行任务
func AdminRunTaskNow(c *gin.Context) {
	if TaskSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := TaskSvc.RunTaskNow(uint(taskID)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "手动执行任务", "scheduled_task", c.Param("id"), nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{"success": true, "message": "任务已开始执行"})
}

// AdminToggleTaskStatus 切换任务状态
func AdminToggleTaskStatus(c *gin.Context) {
	if TaskSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	task, err := TaskSvc.GetTask(uint(taskID))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "任务不存在"})
		return
	}

	// 切换状态
	if task.Status == 1 {
		task.Status = 0
	} else {
		task.Status = 1
	}

	if err := TaskSvc.UpdateTask(task); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新失败"})
		return
	}

	statusText := "已禁用"
	if task.Status == 1 {
		statusText = "已启用"
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "切换任务状态", "scheduled_task", c.Param("id"), statusText, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{"success": true, "message": statusText, "status": task.Status})
}
