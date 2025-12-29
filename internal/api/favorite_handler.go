package api

import "github.com/gin-gonic/gin"

// GetFavorites 获取收藏列表
func GetFavorites(c *gin.Context) {
	if FavoriteSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	page := 1
	pageSize := 20
	if p, err := parseInt(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if ps, err := parseInt(c.DefaultQuery("page_size", "20")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	favorites, total, err := FavoriteSvc.GetUserFavorites(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取收藏列表失败"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(200, gin.H{
		"success":     true,
		"favorites":   favorites,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// AddFavorite 添加收藏
func AddFavorite(c *gin.Context) {
	if FavoriteSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := FavoriteSvc.AddFavorite(userID.(uint), req.ProductID); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "收藏成功"})
}

// RemoveFavorite 取消收藏
func RemoveFavorite(c *gin.Context) {
	if FavoriteSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	productID, err := parseUint(c.Param("product_id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := FavoriteSvc.RemoveFavorite(userID.(uint), uint(productID)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "已取消收藏"})
}

// CheckFavorite 检查收藏状态
func CheckFavorite(c *gin.Context) {
	if FavoriteSvc == nil {
		c.JSON(200, gin.H{"success": true, "is_favorite": false})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, gin.H{"success": true, "is_favorite": false})
		return
	}

	productID, err := parseUint(c.Param("product_id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	isFavorite := FavoriteSvc.IsFavorite(userID.(uint), uint(productID))
	c.JSON(200, gin.H{"success": true, "is_favorite": isFavorite})
}

// GetFavoriteCount 获取收藏数量
func GetFavoriteCount(c *gin.Context) {
	if FavoriteSvc == nil {
		c.JSON(200, gin.H{"success": true, "count": 0})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, gin.H{"success": true, "count": 0})
		return
	}

	count := FavoriteSvc.GetFavoriteCount(userID.(uint))
	c.JSON(200, gin.H{"success": true, "count": count})
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	var i int
	_, err := parseIntHelper(s, &i)
	return i, err
}

// parseUint 解析无符号整数
func parseUint(s string) (uint64, error) {
	var i uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errInvalidNumber
		}
		i = i*10 + uint64(c-'0')
	}
	return i, nil
}

// parseIntHelper 辅助函数
func parseIntHelper(s string, i *int) (bool, error) {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, errInvalidNumber
		}
		*i = *i*10 + int(c-'0')
	}
	return true, nil
}

var errInvalidNumber = &invalidNumberError{}

type invalidNumberError struct{}

func (e *invalidNumberError) Error() string {
	return "无效的数字"
}
