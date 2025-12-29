package service

import (
	"errors"

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

	return category, nil
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(id uint) error {
	return s.repo.DeleteCategory(id)
}

// GetAllCategories 获取所有分类
func (s *CategoryService) GetAllCategories(onlyActive bool) ([]model.ProductCategory, error) {
	return s.repo.GetAllCategories(onlyActive)
}

// GetCategoryByID 获取分类详情
func (s *CategoryService) GetCategoryByID(id uint) (*model.ProductCategory, error) {
	return s.repo.GetCategoryByID(id)
}
