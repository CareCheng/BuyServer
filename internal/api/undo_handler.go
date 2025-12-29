package api

import (
	"strconv"
	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminGetUndoableOperations 获取可撤销操作列表
func AdminGetUndoableOperations(c *gin.Context) {
	if UndoSvc == nil {
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

	operations, total, err := UndoSvc.GetUndoableOperations(page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取操作列表失败"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(200, gin.H{
		"success":     true,
		"operations":  operations,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// AdminGetAllUndoOperations 获取所有操作记录
func AdminGetAllUndoOperations(c *gin.Context) {
	if UndoSvc == nil {
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

	operations, total, err := UndoSvc.GetAllOperations(status, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取操作列表失败"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(200, gin.H{
		"success":     true,
		"operations":  operations,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// AdminUndoOperation 撤销操作
func AdminUndoOperation(c *gin.Context) {
	if UndoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	operationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	adminUsername, _ := c.Get("admin_username")
	if err := UndoSvc.UndoOperation(uint(operationID), adminUsername.(string)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	LogSvc.LogAdminActionSimple(adminUsername.(string), "撤销操作", "undo_operation", c.Param("id"), nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{
		"success": true,
		"message": "操作已撤销",
	})
}

// AdminGetUndoConfig 获取撤销配置
func AdminGetUndoConfig(c *gin.Context) {
	if UndoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	config, err := UndoSvc.GetConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取配置失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"config":  config,
	})
}

// AdminSaveUndoConfig 保存撤销配置
func AdminSaveUndoConfig(c *gin.Context) {
	if UndoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var config model.UndoConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := UndoSvc.SaveConfig(&config); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存配置失败"})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	LogSvc.LogAdminActionSimple(adminUsername.(string), "更新撤销配置", "undo_config", "", nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(200, gin.H{
		"success": true,
		"message": "配置已保存",
	})
}

// AdminGetUndoStats 获取撤销统计
func AdminGetUndoStats(c *gin.Context) {
	if UndoSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats := UndoSvc.GetUndoStats()
	c.JSON(200, gin.H{
		"success": true,
		"stats":   stats,
	})
}
