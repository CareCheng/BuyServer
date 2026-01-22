package service

import (
	"encoding/json"
	"errors"
	"log"

	"user-frontend/internal/cache"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// CategoryService 分类服务
type CategoryService struct {
	repo *repository.Repository
}

func NewCategoryService(repo *repository.Repository) *CategoryService {
	return &CategoryService{repo: repo}
}

// ==================== 缓存辅助方法 ====================

// cacheCategory 缓存分类信息
func (s *CategoryService) cacheCategory(category *model.ProductCategory) {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	key := cache.CategoryKey(category.ID)
	data, err := json.Marshal(category)
	if err != nil {
		log.Printf("[CategoryService] 序列化分类缓存失败: %v", err)
		return
	}

	if err := cm.Set(key, string(data), cache.CategoryTTL); err != nil {
		log.Printf("[CategoryService] 缓存分类失败: %v", err)
	}
}

// getCategoryFromCache 从缓存获取分类
func (s *CategoryService) getCategoryFromCache(categoryID uint) *model.ProductCategory {
	cm := cache.GetManager()
	if cm == nil {
		return nil
	}

	key := cache.CategoryKey(categoryID)
	data, ok := cm.Get(key)
	if !ok {
		return nil
	}

	dataStr, ok := data.(string)
	if !ok {
		return nil
	}

	var category model.ProductCategory
	if err := json.Unmarshal([]byte(dataStr), &category); err != nil {
		log.Printf("[CategoryService] 反序列化分类缓存失败: %v", err)
		return nil
	}

	return &category
}

// cacheCategoryList 缓存分类列表
func (s *CategoryService) cacheCategoryList(categories []model.ProductCategory) {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	key := cache.CategoryListKey()
	data, err := json.Marshal(categories)
	if err != nil {
		log.Printf("[CategoryService] 序列化分类列表缓存失败: %v", err)
		return
	}

	if err := cm.Set(key, string(data), cache.CategoryTTL); err != nil {
		log.Printf("[CategoryService] 缓存分类列表失败: %v", err)
	}
}

// getCategoryListFromCache 从缓存获取分类列表
func (s *CategoryService) getCategoryListFromCache() []model.ProductCategory {
	cm := cache.GetManager()
	if cm == nil {
		return nil
	}

	key := cache.CategoryListKey()
	data, ok := cm.Get(key)
	if !ok {
		return nil
	}

	dataStr, ok := data.(string)
	if !ok {
		return nil
	}

	var categories []model.ProductCategory
	if err := json.Unmarshal([]byte(dataStr), &categories); err != nil {
		log.Printf("[CategoryService] 反序列化分类列表缓存失败: %v", err)
		return nil
	}

	return categories
}

// invalidateCategoryCache 使分类缓存失效
func (s *CategoryService) invalidateCategoryCache(categoryID uint) {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	// 删除单个分类缓存
	key := cache.CategoryKey(categoryID)
	if err := cm.Delete(key); err != nil {
		log.Printf("[CategoryService] 删除分类缓存失败: %v", err)
	}

	// 同时清除分类列表缓存
	s.invalidateCategoryListCache()
}

// invalidateCategoryListCache 使分类列表缓存失效
func (s *CategoryService) invalidateCategoryListCache() {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	key := cache.CategoryListKey()
	if err := cm.Delete(key); err != nil {
		log.Printf("[CategoryService] 删除分类列表缓存失败: %v", err)
	}

	// 同时清除分类树缓存
	treeKey := cache.CategoryTreeKey()
	if err := cm.Delete(treeKey); err != nil {
		log.Printf("[CategoryService] 删除分类树缓存失败: %v", err)
	}
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(name, icon string, sortOrder int) (*model.ProductCategory, error) {
	if name == "" {
		return nil, errors.New("分类名称不能为空")
	}

	category := &model.ProductCategory{
		Name:      name,
		Icon:      icon,
		SortOrder: sortOrder,
		Status:    1,
	}

	if err := s.repo.CreateCategory(category); err != nil {
		return nil, err
	}

	// 创建成功后使列表缓存失效
	s.invalidateCategoryListCache()

	return category, nil
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(id uint, name, icon string, sortOrder, status int) (*model.ProductCategory, error) {
	category, err := s.repo.GetCategoryByID(id)
	if err != nil {
		return nil, errors.New("分类不存在")
	}

	if name != "" {
		category.Name = name
	}
	category.Icon = icon
	category.SortOrder = sortOrder
	category.Status = status

	if err := s.repo.UpdateCategory(category); err != nil {
		return nil, err
	}

	// 更新成功后使缓存失效
	s.invalidateCategoryCache(id)

	return category, nil
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(id uint) error {
	err := s.repo.DeleteCategory(id)
	if err == nil {
		s.invalidateCategoryCache(id)
	}
	return err
}

// GetAllCategories 获取所有分类（支持缓存）
func (s *CategoryService) GetAllCategories(onlyActive bool) ([]model.ProductCategory, error) {
	// 只缓存全量数据（包含非活跃的）
	if !onlyActive {
		if categories := s.getCategoryListFromCache(); categories != nil {
			return categories, nil
		}
	}

	// 从数据库获取
	categories, err := s.repo.GetAllCategories(onlyActive)
	if err != nil {
		return nil, err
	}

	// 只缓存全量数据
	if !onlyActive {
		s.cacheCategoryList(categories)
	}

	return categories, nil
}

// GetCategoryByID 获取分类详情（支持缓存）
func (s *CategoryService) GetCategoryByID(id uint) (*model.ProductCategory, error) {
	// 先从缓存获取
	if category := s.getCategoryFromCache(id); category != nil {
		return category, nil
	}

	// 从数据库获取
	category, err := s.repo.GetCategoryByID(id)
	if err != nil {
		return nil, err
	}

	// 缓存分类
	s.cacheCategory(category)

	return category, nil
}
