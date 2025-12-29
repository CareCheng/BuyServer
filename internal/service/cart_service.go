package service

import (
	"errors"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// CartService 购物车服务
type CartService struct {
	repo *repository.Repository
}

// NewCartService 创建购物车服务
func NewCartService(repo *repository.Repository) *CartService {
	return &CartService{repo: repo}
}

// GetCart 获取用户购物车
func (s *CartService) GetCart(userID uint) (*model.CartSummary, error) {
	db := s.repo.GetDB()
	var items []model.CartItem

	err := db.Preload("Product").Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
	if err != nil {
		return nil, err
	}

	// 计算汇总
	summary := &model.CartSummary{
		Items:      items,
		TotalCount: 0,
		TotalPrice: 0,
	}

	for _, item := range items {
		if item.Product != nil && item.Product.Status == 1 {
			summary.TotalCount += item.Quantity
			summary.TotalPrice += item.Product.Price * float64(item.Quantity)
		}
	}

	return summary, nil
}

// AddToCart 添加商品到购物车
func (s *CartService) AddToCart(userID, productID uint, quantity int) (*model.CartItem, error) {
	if quantity <= 0 {
		return nil, errors.New("数量必须大于0")
	}

	db := s.repo.GetDB()

	// 检查商品是否存在且上架
	var product model.Product
	if err := db.First(&product, productID).Error; err != nil {
		return nil, errors.New("商品不存在")
	}
	if product.Status != 1 {
		return nil, errors.New("商品已下架")
	}

	// 检查库存
	if product.Stock != -1 && product.Stock < quantity {
		return nil, errors.New("库存不足")
	}

	// 检查购物车是否已有该商品
	var existingItem model.CartItem
	err := db.Where("user_id = ? AND product_id = ?", userID, productID).First(&existingItem).Error
	if err == nil {
		// 已存在，更新数量
		newQuantity := existingItem.Quantity + quantity
		if product.Stock != -1 && product.Stock < newQuantity {
			return nil, errors.New("库存不足")
		}
		existingItem.Quantity = newQuantity
		if err := db.Save(&existingItem).Error; err != nil {
			return nil, err
		}
		// 加载商品信息
		db.Preload("Product").First(&existingItem, existingItem.ID)
		return &existingItem, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 不存在，创建新项
	item := &model.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	}
	if err := db.Create(item).Error; err != nil {
		return nil, err
	}

	// 加载商品信息
	db.Preload("Product").First(item, item.ID)
	return item, nil
}

// UpdateCartItem 更新购物车项数量
func (s *CartService) UpdateCartItem(userID, itemID uint, quantity int) error {
	if quantity <= 0 {
		return s.RemoveFromCart(userID, itemID)
	}

	db := s.repo.GetDB()

	var item model.CartItem
	if err := db.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error; err != nil {
		return errors.New("购物车项不存在")
	}

	// 检查库存
	var product model.Product
	if err := db.First(&product, item.ProductID).Error; err != nil {
		return errors.New("商品不存在")
	}
	if product.Stock != -1 && product.Stock < quantity {
		return errors.New("库存不足")
	}

	item.Quantity = quantity
	return db.Save(&item).Error
}

// RemoveFromCart 从购物车移除商品
func (s *CartService) RemoveFromCart(userID, itemID uint) error {
	return s.repo.GetDB().Where("id = ? AND user_id = ?", itemID, userID).Delete(&model.CartItem{}).Error
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(userID uint) error {
	return s.repo.GetDB().Where("user_id = ?", userID).Delete(&model.CartItem{}).Error
}

// GetCartItemCount 获取购物车商品数量
func (s *CartService) GetCartItemCount(userID uint) int64 {
	var count int64
	s.repo.GetDB().Model(&model.CartItem{}).Where("user_id = ?", userID).Count(&count)
	return count
}

// ValidateCart 验证购物车（检查商品状态和库存）
func (s *CartService) ValidateCart(userID uint) ([]model.CartItem, []string, error) {
	db := s.repo.GetDB()
	var items []model.CartItem
	var warnings []string

	err := db.Preload("Product").Where("user_id = ?", userID).Find(&items).Error
	if err != nil {
		return nil, nil, err
	}

	validItems := make([]model.CartItem, 0)
	for _, item := range items {
		if item.Product == nil {
			// 商品已删除
			db.Delete(&item)
			warnings = append(warnings, "部分商品已下架，已自动移除")
			continue
		}
		if item.Product.Status != 1 {
			// 商品已下架
			db.Delete(&item)
			warnings = append(warnings, item.Product.Name+" 已下架，已自动移除")
			continue
		}
		if item.Product.Stock != -1 && item.Product.Stock < item.Quantity {
			// 库存不足
			if item.Product.Stock > 0 {
				item.Quantity = item.Product.Stock
				db.Save(&item)
				warnings = append(warnings, item.Product.Name+" 库存不足，已调整数量")
			} else {
				db.Delete(&item)
				warnings = append(warnings, item.Product.Name+" 已售罄，已自动移除")
				continue
			}
		}
		validItems = append(validItems, item)
	}

	return validItems, warnings, nil
}

// GetCartTotal 获取购物车总价
func (s *CartService) GetCartTotal(userID uint) (float64, error) {
	summary, err := s.GetCart(userID)
	if err != nil {
		return 0, err
	}
	return summary.TotalPrice, nil
}

// RemoveProductFromAllCarts 从所有购物车移除指定商品（商品下架时调用）
func (s *CartService) RemoveProductFromAllCarts(productID uint) error {
	return s.repo.GetDB().Where("product_id = ?", productID).Delete(&model.CartItem{}).Error
}
