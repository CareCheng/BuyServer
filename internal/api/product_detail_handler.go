package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetProductDetailFile 获取商品详情文件内容
// GET /api/product/:id/detail-file
func GetProductDetailFile(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	productID := c.Param("id")
	
	// 构建文件路径
	filePath := fmt.Sprintf("./Product/%s/detail.md", productID)
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(200, gin.H{"success": true, "content": "", "exists": false})
		return
	}
	
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "读取文件失败"})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "content": string(content), "exists": true})
}

// SaveProductDetailFile 保存商品详情文件
// POST /api/admin/product/:id/detail-file
func SaveProductDetailFile(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	productID := c.Param("id")
	
	var req struct {
		Content string `json:"content"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	// 创建目录
	dirPath := fmt.Sprintf("./Product/%s", productID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建目录失败"})
		return
	}
	
	// 保存文件
	filePath := filepath.Join(dirPath, "detail.md")
	if err := os.WriteFile(filePath, []byte(req.Content), 0644); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存文件失败"})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "message": "保存成功"})
}

// UploadProductDetailImage 上传商品详情图片
// POST /api/admin/product/:id/detail-image
func UploadProductDetailImage(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	productID := c.Param("id")
	
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请选择要上传的图片"})
		return
	}
	defer file.Close()
	
	// 验证文件扩展名
	ext := filepath.Ext(header.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(400, gin.H{"success": false, "error": "不支持的图片格式，仅支持 jpg/jpeg/png/gif/webp"})
		return
	}
	
	// 创建目录
	dirPath := fmt.Sprintf("./Product/%s/detail_images", productID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建目录失败"})
		return
	}
	
	// 生成文件名
	filename := fmt.Sprintf("%d%s", generateTimestamp(), ext)
	filePath := filepath.Join(dirPath, filename)
	
	// 保存文件
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建文件失败"})
		return
	}
	defer dst.Close()
	
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存文件失败"})
		return
	}
	
	// 返回图片URL
	imageURL := fmt.Sprintf("/product-files/%s/detail_images/%s", productID, filename)
	c.JSON(200, gin.H{"success": true, "url": imageURL})
}

// generateTimestamp 生成时间戳
func generateTimestamp() int64 {
	return time.Now().UnixNano()
}
