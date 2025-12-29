package service

import (
	"errors"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// FavoriteService 商品收藏服务
type FavoriteService struct {
	repo *repository.Repository
}

// NewFavoriteService 创建收藏服务实例
func NewFavoriteService(repo *repository.Repository) *FavoriteService {
	return &FavoriteService{repo: repo}
}

// AddFavorite 添加收藏
// 参数：
//   - userID: 用户ID
//   - productID: 商品ID
// 返回：
//   - 错误信息（如有）
func (s *FavoriteService) AddFavorite(userID, productID uint) error {
	// 检查商品是否存在
	var product model.Product
	if err := s.repo.GetDB().First(&product, productID).Error; err != nil {
		return errors.New("商品不存在")
	}

	// 检查是否已收藏
	var existing model.ProductFavorite
	result := s.repo.GetDB().Where("user_id = ? AND product_id = ?", userID, productID).First(&existing)
	if result.Error == nil {
		return errors.New("已收藏该商品")
	}

	// 创建收藏记录
	favorite := model.ProductFavorite{
		UserID:    userID,
		ProductID: productID,
	}
	return s.repo.GetDB().Create(&favorite).Error
}

// RemoveFavorite 取消收藏
// 参数：
//   - userID: 用户ID
//   - productID: 商品ID
// 返回：
//   - 错误信息（如有）
func (s *FavoriteService) RemoveFavorite(userID, productID uint) error {
	result := s.repo.GetDB().Where("user_id = ? AND product_id = ?", userID, productID).Delete(&model.ProductFavorite{})
	if result.RowsAffected == 0 {
		return errors.New("未收藏该商品")
	}
	return result.Error
}

// IsFavorite 检查是否已收藏
// 参数：
//   - userID: 用户ID
//   - productID: 商品ID
// 返回：
//   - 是否已收藏
func (s *FavoriteService) IsFavorite(userID, productID uint) bool {
	var count int64
	s.repo.GetDB().Model(&model.ProductFavorite{}).Where("user_id = ? AND product_id = ?", userID, productID).Count(&count)
	return count > 0
}

// FavoriteProductInfo 收藏商品信息
type FavoriteProductInfo struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
	Status      int     `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

// GetUserFavorites 获取用户收藏列表
// 参数：
//   - userID: 用户ID
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 收藏列表
//   - 总数
//   - 错误信息（如有）
func (s *FavoriteService) GetUserFavorites(userID uint, page, pageSize int) ([]FavoriteProductInfo, int64, error) {
	var total int64
	s.repo.GetDB().Model(&model.ProductFavorite{}).Where("user_id = ?", userID).Count(&total)

	var favorites []model.ProductFavorite
	offset := (page - 1) * pageSize
	if err := s.repo.GetDB().Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&favorites).Error; err != nil {
		return nil, 0, err
	}

	// 获取商品详情
	result := make([]FavoriteProductInfo, 0, len(favorites))
	for _, fav := range favorites {
		var product model.Product
		if err := s.repo.GetDB().First(&product, fav.ProductID).Error; err != nil {
			continue // 商品可能已删除，跳过
		}
		result = append(result, FavoriteProductInfo{
			ID:          fav.ID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Price:       product.Price,
			ImageURL:    product.ImageURL,
			Status:      product.Status,
			CreatedAt:   fav.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return result, total, nil
}

// GetFavoriteCount 获取用户收藏数量
// 参数：
//   - userID: 用户ID
// 返回：
//   - 收藏数量
func (s *FavoriteService) GetFavoriteCount(userID uint) int64 {
	var count int64
	s.repo.GetDB().Model(&model.ProductFavorite{}).Where("user_id = ?", userID).Count(&count)
	return count
}

// BatchCheckFavorites 批量检查收藏状态
// 参数：
//   - userID: 用户ID
//   - productIDs: 商品ID列表
// 返回：
//   - 商品ID到收藏状态的映射
func (s *FavoriteService) BatchCheckFavorites(userID uint, productIDs []uint) map[uint]bool {
	result := make(map[uint]bool)
	if len(productIDs) == 0 {
		return result
	}

	var favorites []model.ProductFavorite
	s.repo.GetDB().Where("user_id = ? AND product_id IN ?", userID, productIDs).Find(&favorites)

	for _, fav := range favorites {
		result[fav.ProductID] = true
	}
	return result
}
