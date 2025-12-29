package api

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 商品管理相关 API ====================

// AdminGetProducts 获取商品列表
func AdminGetProducts(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	products, err := ProductSvc.GetAllProducts(false)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "products": products})
}

// AdminCreateProduct 创建商品
func AdminCreateProduct(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Name         string  `json:"name" binding:"required"`
		Description  string  `json:"description"`
		Detail       string  `json:"detail"`       // 详细介绍（Markdown/HTML）
		Specs        string  `json:"specs"`        // 规格参数（JSON格式）
		Features     string  `json:"features"`     // 特性/卖点列表（JSON格式）
		Tags         string  `json:"tags"`         // 商品标签（逗号分隔）
		Price        float64 `json:"price"`
		Duration     int     `json:"duration" binding:"required"`
		DurationUnit string  `json:"duration_unit"`
		Stock        int     `json:"stock"`
		ImageURL     string  `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if req.DurationUnit == "" {
		req.DurationUnit = "天"
	}

	// 手动卡密类型初始库存为0，由导入的卡密数量决定
	stock := 0

	// 使用完整字段创建商品
	product := &model.Product{
		Name:         req.Name,
		Description:  req.Description,
		Detail:       req.Detail,
		Specs:        req.Specs,
		Features:     req.Features,
		Tags:         req.Tags,
		Price:        req.Price,
		Duration:     req.Duration,
		DurationUnit: req.DurationUnit,
		Stock:        stock,
		ImageURL:     req.ImageURL,
		ProductType:  model.ProductTypeManual,
	}

	if err := ProductSvc.CreateProductFull(product); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "product": product})
}

// AdminUpdateProduct 更新商品
func AdminUpdateProduct(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var req struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		Detail       string  `json:"detail"`       // 详细介绍（Markdown/HTML）
		Specs        string  `json:"specs"`        // 规格参数（JSON格式）
		Features     string  `json:"features"`     // 特性/卖点列表（JSON格式）
		Tags         string  `json:"tags"`         // 商品标签（逗号分隔）
		Price        float64 `json:"price"`
		Duration     int     `json:"duration"`
		DurationUnit string  `json:"duration_unit"`
		Stock        int     `json:"stock"`
		Status       int     `json:"status"`
		ImageURL     string  `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取现有商品
	existing, err := ProductSvc.GetProductByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "商品不存在"})
		return
	}

	// 更新字段
	if req.Name != "" {
		existing.Name = req.Name
	}
	existing.Description = req.Description
	existing.Detail = req.Detail
	existing.Specs = req.Specs
	existing.Features = req.Features
	existing.Tags = req.Tags
	if req.Price >= 0 {
		existing.Price = req.Price
	}
	if req.Duration > 0 {
		existing.Duration = req.Duration
	}
	if req.DurationUnit != "" {
		existing.DurationUnit = req.DurationUnit
	}
	existing.Stock = req.Stock
	existing.Status = req.Status
	existing.ImageURL = req.ImageURL

	if err := ProductSvc.UpdateProductFull(existing); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "product": existing})
}

// AdminDeleteProduct 删除商品
func AdminDeleteProduct(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := ProductSvc.DeleteProduct(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "商品已删除"})
}

// AdminUploadProductImage 上传商品图片
func AdminUploadProductImage(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	// 获取商品信息
	product, err := ProductSvc.GetProductByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "商品不存在"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请选择要上传的图片"})
		return
	}

	// 验证文件类型
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	contentType := file.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		c.JSON(400, gin.H{"success": false, "error": "只支持 JPG、PNG、GIF、WebP 格式的图片"})
		return
	}

	// 限制文件大小 (5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(400, gin.H{"success": false, "error": "图片大小不能超过5MB"})
		return
	}

	// 创建商品专属目录: Product/{product_id}/
	productDir := fmt.Sprintf("Product/%d", product.ID)
	if err := os.MkdirAll(productDir, 0755); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建目录失败"})
		return
	}

	// 生成文件名（使用时间戳避免缓存问题）
	ext := ".jpg"
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	}
	filename := fmt.Sprintf("image_%d%s", time.Now().Unix(), ext)
	filepath := fmt.Sprintf("%s/%s", productDir, filename)

	// 删除旧图片（如果存在）
	if product.ImageURL != "" {
		oldPath := "." + product.ImageURL // 转换为相对路径
		os.Remove(oldPath)
	}

	// 保存文件
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存图片失败"})
		return
	}

	// 更新数据库中的图片路径
	imageURL := "/product-files/" + filepath[8:] // 存储为 /product-files/{id}/image_xxx.jpg，去掉 Product/ 前缀
	product.ImageURL = imageURL
	if err := ProductSvc.UpdateProductImageURL(uint(id), imageURL); err != nil {
		// 如果数据库更新失败，删除已上传的文件
		os.Remove(filepath)
		c.JSON(500, gin.H{"success": false, "error": "更新商品信息失败"})
		return
	}

	c.JSON(200, gin.H{
		"success":   true,
		"message":   "图片上传成功",
		"image_url": imageURL,
	})
}

// AdminDeleteProductImage 删除商品图片
func AdminDeleteProductImage(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	// 获取商品信息
	product, err := ProductSvc.GetProductByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "商品不存在"})
		return
	}

	// 删除图片文件
	if product.ImageURL != "" {
		oldPath := "." + product.ImageURL
		os.Remove(oldPath)
	}

	// 清空数据库中的图片路径
	if err := ProductSvc.UpdateProductImageURL(uint(id), ""); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新商品信息失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "图片已删除"})
}
