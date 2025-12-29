package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ==================== 购物车 API ====================

// GetCart 获取购物车
func GetCart(c *gin.Context) {
	if CartSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")
	summary, err := CartSvc.GetCart(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取购物车失败"})
		return
	}

	// 返回 items 字段以匹配前端期望
	c.JSON(200, gin.H{"success": true, "items": summary.Items, "total": summary.TotalPrice})
}

// AddToCart 添加商品到购物车
func AddToCart(c *gin.Context) {
	if CartSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	item, err := CartSvc.AddToCart(userID.(uint), req.ProductID, req.Quantity)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": item})
}

// UpdateCartItem 更新购物车项数量
func UpdateCartItem(c *gin.Context) {
	if CartSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的购物车项ID"})
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := CartSvc.UpdateCartItem(userID.(uint), uint(itemID), req.Quantity); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// RemoveFromCart 从购物车移除商品
func RemoveFromCart(c *gin.Context) {
	if CartSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的购物车项ID"})
		return
	}

	if err := CartSvc.RemoveFromCart(userID.(uint), uint(itemID)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "移除成功"})
}

// ClearCart 清空购物车
func ClearCart(c *gin.Context) {
	if CartSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")
	if err := CartSvc.ClearCart(userID.(uint)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "清空成功"})
}

// GetCartCount 获取购物车商品数量
func GetCartCount(c *gin.Context) {
	if CartSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")
	count := CartSvc.GetCartItemCount(userID.(uint))

	c.JSON(200, gin.H{"success": true, "count": count})
}

// ValidateCart 验证购物车
func ValidateCart(c *gin.Context) {
	if CartSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")

	// 解析请求参数
	var req struct {
		ItemIDs []uint `json:"item_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有传参数，验证整个购物车
		req.ItemIDs = nil
	}

	items, warnings, err := CartSvc.ValidateCart(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "验证失败"})
		return
	}

	// 如果指定了 item_ids，只验证这些项
	if len(req.ItemIDs) > 0 {
		itemIDMap := make(map[uint]bool)
		for _, id := range req.ItemIDs {
			itemIDMap[id] = true
		}
		var filteredWarnings []string
		for _, item := range items {
			if itemIDMap[item.ID] && item.Product != nil {
				// 检查库存
				if item.Product.Stock == 0 {
					filteredWarnings = append(filteredWarnings, item.Product.Name+"已售罄")
				} else if item.Product.Stock > 0 && item.Quantity > item.Product.Stock {
					filteredWarnings = append(filteredWarnings, item.Product.Name+"库存不足")
				}
			}
		}
		warnings = filteredWarnings
	}

	// 返回前端期望的字段格式
	c.JSON(200, gin.H{
		"success": true,
		"valid":   len(warnings) == 0,
		"errors":  warnings,
		"items":   items,
	})
}
