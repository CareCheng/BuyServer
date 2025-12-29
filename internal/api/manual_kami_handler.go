package api

import (
	"strconv"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminImportKami 导入手动卡密
// POST /api/admin/product/:id/kami/import
func AdminImportKami(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ManualKamiSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "卡密服务未初始化"})
		return
	}

	idStr := c.Param("id")
	productID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	var req struct {
		Codes string `json:"codes" binding:"required"` // 卡密文本，每行一个
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误：请提供卡密内容"})
		return
	}

	imported, duplicates, err := ManualKamiSvc.ImportKamiCodes(uint(productID), req.Codes)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"message":    "卡密导入完成",
		"imported":   imported,
		"duplicates": duplicates,
	})
}

// AdminGetProductKamis 获取商品的卡密列表
// GET /api/admin/product/:id/kami
func AdminGetProductKamis(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ManualKamiSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "卡密服务未初始化"})
		return
	}

	idStr := c.Param("id")
	productID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 状态筛选
	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = &s
		}
	}

	kamis, total, err := ManualKamiSvc.GetProductKamis(uint(productID), page, pageSize, status)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 获取统计信息
	stats, _ := ManualKamiSvc.GetKamiStats(uint(productID))

	c.JSON(200, gin.H{
		"success": true,
		"kamis":   kamis,
		"total":   total,
		"page":    page,
		"stats":   stats,
	})
}

// AdminGetKamiStats 获取商品的卡密统计
// GET /api/admin/product/:id/kami/stats
func AdminGetKamiStats(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ManualKamiSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "卡密服务未初始化"})
		return
	}

	idStr := c.Param("id")
	productID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	stats, err := ManualKamiSvc.GetKamiStats(uint(productID))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"stats":   stats,
	})
}

// AdminDeleteKami 删除卡密
// DELETE /api/admin/kami/:id
func AdminDeleteKami(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ManualKamiSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "卡密服务未初始化"})
		return
	}

	idStr := c.Param("id")
	kamiID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的卡密ID"})
		return
	}

	if err := ManualKamiSvc.DeleteKami(uint(kamiID)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "卡密已删除"})
}

// AdminDisableKami 禁用卡密
// POST /api/admin/kami/:id/disable
func AdminDisableKami(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ManualKamiSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "卡密服务未初始化"})
		return
	}

	idStr := c.Param("id")
	kamiID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的卡密ID"})
		return
	}

	if err := ManualKamiSvc.DisableKami(uint(kamiID)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "卡密已禁用"})
}

// AdminEnableKami 启用卡密
// POST /api/admin/kami/:id/enable
func AdminEnableKami(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ManualKamiSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "卡密服务未初始化"})
		return
	}

	idStr := c.Param("id")
	kamiID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的卡密ID"})
		return
	}

	if err := ManualKamiSvc.EnableKami(uint(kamiID)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "卡密已启用"})
}

// AdminBatchDeleteKamis 批量删除卡密
// POST /api/admin/kami/batch-delete
func AdminBatchDeleteKamis(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ManualKamiSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "卡密服务未初始化"})
		return
	}

	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请选择要删除的卡密"})
		return
	}

	deleted, skipped, err := ManualKamiSvc.BatchDeleteKamis(req.IDs)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "批量删除完成",
		"deleted": deleted,
		"skipped": skipped,
	})
}
