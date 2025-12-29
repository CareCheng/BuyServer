package service

import (
	"errors"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

type ProductService struct {
	repo *repository.Repository
}

func NewProductService(repo *repository.Repository) *ProductService {
	return &ProductService{repo: repo}
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(name, description string, price float64, duration int, durationUnit string, stock int, imageURL string) (*model.Product, error) {
	if name == "" {
		return nil, errors.New("商品名称不能为空")
	}
	if price < 0 {
		return nil, errors.New("价格不能为负数")
	}
	if duration <= 0 {
		return nil, errors.New("时长必须大于0")
	}

	product := &model.Product{
		Name:         name,
		Description:  description,
		Price:        price,
		Duration:     duration,
		DurationUnit: durationUnit,
		Stock:        stock,
		Status:       1,
		ImageURL:     imageURL,
		ProductType:  model.ProductTypeManual, // 默认手动卡密类型
	}

	if err := s.repo.CreateProduct(product); err != nil {
		return nil, err
	}

	return product, nil
}

// CreateProductFull 创建商品（完整字段）
func (s *ProductService) CreateProductFull(product *model.Product) error {
	if product.Name == "" {
		return errors.New("商品名称不能为空")
	}
	if product.Price < 0 {
		return errors.New("价格不能为负数")
	}
	if product.Duration <= 0 {
		return errors.New("时长必须大于0")
	}
	product.Status = 1
	return s.repo.CreateProduct(product)
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(id uint, name, description string, price float64, duration int, durationUnit string, stock int, status int, imageURL string) (*model.Product, error) {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return nil, errors.New("商品不存在")
	}

	if name != "" {
		product.Name = name
	}
	product.Description = description
	if price >= 0 {
		product.Price = price
	}
	if duration > 0 {
		product.Duration = duration
	}
	if durationUnit != "" {
		product.DurationUnit = durationUnit
	}
	// 手动卡密类型的商品库存由卡密数量决定，不允许手动修改
	if product.ProductType != model.ProductTypeManual {
		product.Stock = stock
	}
	product.Status = status
	product.ImageURL = imageURL

	if err := s.repo.UpdateProduct(product); err != nil {
		return nil, err
	}

	return product, nil
}

// UpdateProductFull 更新商品（完整字段）
func (s *ProductService) UpdateProductFull(product *model.Product) error {
	existing, err := s.repo.GetProductByID(product.ID)
	if err != nil {
		return errors.New("商品不存在")
	}
	// 手动卡密类型的商品库存由卡密数量决定，不允许手动修改
	if existing.ProductType == model.ProductTypeManual {
		product.Stock = existing.Stock
	}
	return s.repo.UpdateProduct(product)
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(id uint) error {
	return s.repo.DeleteProduct(id)
}

// GetProductByID 获取商品
func (s *ProductService) GetProductByID(id uint) (*model.Product, error) {
	return s.repo.GetProductByID(id)
}

// GetAllProducts 获取所有商品
func (s *ProductService) GetAllProducts(onlyActive bool) ([]model.Product, error) {
	return s.repo.GetAllProducts(onlyActive)
}

// GetProductsWithPagination 分页获取商品
func (s *ProductService) GetProductsWithPagination(page, pageSize int, onlyActive bool) ([]model.Product, int64, error) {
	return s.repo.GetProductsWithPagination(page, pageSize, onlyActive)
}

// UpdateProductStatus 更新商品状态
func (s *ProductService) UpdateProductStatus(id uint, status int) error {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return errors.New("商品不存在")
	}

	product.Status = status
	return s.repo.UpdateProduct(product)
}

// UpdateProductStock 更新商品库存
func (s *ProductService) UpdateProductStock(id uint, stock int) error {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return errors.New("商品不存在")
	}

	product.Stock = stock
	return s.repo.UpdateProduct(product)
}

// UpdateProductImageURL 更新商品图片URL
func (s *ProductService) UpdateProductImageURL(id uint, imageURL string) error {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return errors.New("商品不存在")
	}

	product.ImageURL = imageURL
	return s.repo.UpdateProduct(product)
}
