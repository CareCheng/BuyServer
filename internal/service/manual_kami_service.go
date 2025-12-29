package service

import (
	"errors"
	"strings"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// ManualKamiService 手动卡密服务
type ManualKamiService struct {
	repo *repository.Repository
}

// NewManualKamiService 创建手动卡密服务
func NewManualKamiService(repo *repository.Repository) *ManualKamiService {
	return &ManualKamiService{repo: repo}
}

// ImportKamiCodes 批量导入卡密
// 参数：
//   - productID: 商品ID
//   - codesText: 卡密文本（每行一个卡密）
// 返回：
//   - imported: 成功导入数量
//   - duplicates: 重复跳过数量
//   - error: 错误信息
func (s *ManualKamiService) ImportKamiCodes(productID uint, codesText string) (imported int, duplicates int, err error) {
	// 验证商品存在且为手动卡密类型
	product, err := s.repo.GetProductByID(productID)
	if err != nil {
		return 0, 0, errors.New("商品不存在")
	}
	if product.ProductType != model.ProductTypeManual {
		return 0, 0, errors.New("该商品不是手动卡密类型")
	}

	// 解析卡密列表
	lines := strings.Split(codesText, "\n")
	var codes []string
	for _, line := range lines {
		code := strings.TrimSpace(line)
		if code != "" {
			codes = append(codes, code)
		}
	}

	if len(codes) == 0 {
		return 0, 0, errors.New("没有有效的卡密")
	}

	// 获取该商品已有的卡密，用于去重
	existingCodes, err := s.repo.GetManualKamiCodesByProductID(productID)
	if err != nil {
		return 0, 0, err
	}
	existingMap := make(map[string]bool)
	for _, code := range existingCodes {
		existingMap[code] = true
	}

	// 批量导入
	for _, code := range codes {
		if existingMap[code] {
			duplicates++
			continue
		}

		kami := &model.ManualKami{
			ProductID: productID,
			KamiCode:  code,
			Status:    model.ManualKamiStatusAvailable,
		}
		if err := s.repo.CreateManualKami(kami); err != nil {
			continue // 跳过失败的
		}
		imported++
		existingMap[code] = true // 防止同批次重复
	}

	// 更新商品库存（可用卡密数量）
	s.UpdateProductStock(productID)

	return imported, duplicates, nil
}

// GetAvailableKami 获取一个可用的卡密
// 参数：
//   - productID: 商品ID
// 返回：
//   - 卡密对象
//   - 错误信息
func (s *ManualKamiService) GetAvailableKami(productID uint) (*model.ManualKami, error) {
	return s.repo.GetAvailableManualKami(productID)
}

// MarkKamiSold 标记卡密为已售出
// 参数：
//   - kamiID: 卡密ID
//   - orderID: 订单ID
//   - orderNo: 订单号
// 返回：
//   - 错误信息
func (s *ManualKamiService) MarkKamiSold(kamiID uint, orderID uint, orderNo string) error {
	kami, err := s.repo.GetManualKamiByID(kamiID)
	if err != nil {
		return errors.New("卡密不存在")
	}

	if kami.Status != model.ManualKamiStatusAvailable {
		return errors.New("卡密状态异常")
	}

	now := time.Now()
	kami.Status = model.ManualKamiStatusSold
	kami.OrderID = orderID
	kami.OrderNo = orderNo
	kami.SoldAt = &now

	if err := s.repo.UpdateManualKami(kami); err != nil {
		return err
	}

	// 更新商品库存
	s.UpdateProductStock(kami.ProductID)

	return nil
}

// GetKamiStats 获取商品的卡密统计
// 参数：
//   - productID: 商品ID
// 返回：
//   - 统计信息（total, available, sold, disabled）
//   - 错误信息
func (s *ManualKamiService) GetKamiStats(productID uint) (map[string]int64, error) {
	return s.repo.GetManualKamiStats(productID)
}

// GetProductKamis 获取商品的卡密列表（分页）
// 参数：
//   - productID: 商品ID
//   - page: 页码
//   - pageSize: 每页数量
//   - status: 状态筛选（nil表示全部）
// 返回：
//   - 卡密列表
//   - 总数
//   - 错误信息
func (s *ManualKamiService) GetProductKamis(productID uint, page, pageSize int, status *int) ([]model.ManualKami, int64, error) {
	return s.repo.GetManualKamisByProductID(productID, page, pageSize, status)
}

// DeleteKami 删除卡密
// 参数：
//   - kamiID: 卡密ID
// 返回：
//   - 错误信息
func (s *ManualKamiService) DeleteKami(kamiID uint) error {
	kami, err := s.repo.GetManualKamiByID(kamiID)
	if err != nil {
		return errors.New("卡密不存在")
	}

	if kami.Status == model.ManualKamiStatusSold {
		return errors.New("已售出的卡密不能删除")
	}

	productID := kami.ProductID
	if err := s.repo.DeleteManualKami(kamiID); err != nil {
		return err
	}

	// 更新商品库存
	s.UpdateProductStock(productID)

	return nil
}

// DisableKami 禁用卡密
// 参数：
//   - kamiID: 卡密ID
// 返回：
//   - 错误信息
func (s *ManualKamiService) DisableKami(kamiID uint) error {
	kami, err := s.repo.GetManualKamiByID(kamiID)
	if err != nil {
		return errors.New("卡密不存在")
	}

	if kami.Status == model.ManualKamiStatusSold {
		return errors.New("已售出的卡密不能禁用")
	}

	kami.Status = model.ManualKamiStatusDisabled
	if err := s.repo.UpdateManualKami(kami); err != nil {
		return err
	}

	// 更新商品库存
	s.UpdateProductStock(kami.ProductID)

	return nil
}

// EnableKami 启用卡密
// 参数：
//   - kamiID: 卡密ID
// 返回：
//   - 错误信息
func (s *ManualKamiService) EnableKami(kamiID uint) error {
	kami, err := s.repo.GetManualKamiByID(kamiID)
	if err != nil {
		return errors.New("卡密不存在")
	}

	if kami.Status != model.ManualKamiStatusDisabled {
		return errors.New("只能启用已禁用的卡密")
	}

	kami.Status = model.ManualKamiStatusAvailable
	if err := s.repo.UpdateManualKami(kami); err != nil {
		return err
	}

	// 更新商品库存
	s.UpdateProductStock(kami.ProductID)

	return nil
}

// UpdateProductStock 更新商品库存（根据可用卡密数量）
// 参数：
//   - productID: 商品ID
func (s *ManualKamiService) UpdateProductStock(productID uint) {
	stats, err := s.repo.GetManualKamiStats(productID)
	if err != nil {
		return
	}

	product, err := s.repo.GetProductByID(productID)
	if err != nil {
		return
	}

	// 只更新手动卡密类型商品的库存
	if product.ProductType == model.ProductTypeManual {
		product.Stock = int(stats["available"])
		s.repo.UpdateProduct(product)
	}
}

// BatchDeleteKamis 批量删除卡密
// 参数：
//   - kamiIDs: 卡密ID列表
// 返回：
//   - deleted: 成功删除数量
//   - skipped: 跳过数量（已售出的）
//   - error: 错误信息
func (s *ManualKamiService) BatchDeleteKamis(kamiIDs []uint) (deleted int, skipped int, err error) {
	var productID uint
	for _, id := range kamiIDs {
		kami, err := s.repo.GetManualKamiByID(id)
		if err != nil {
			continue
		}
		if productID == 0 {
			productID = kami.ProductID
		}
		if kami.Status == model.ManualKamiStatusSold {
			skipped++
			continue
		}
		if err := s.repo.DeleteManualKami(id); err == nil {
			deleted++
		}
	}

	// 更新商品库存
	if productID > 0 {
		s.UpdateProductStock(productID)
	}

	return deleted, skipped, nil
}
