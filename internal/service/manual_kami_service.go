package service

import (
	"encoding/json"
	"errors"
	"regexp"
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

// parseKamiCodes 智能解析卡密文本，支持多种格式
// 支持的格式：
//  1. 换行分隔：每行一个卡密
//  2. 逗号分隔：code1,code2,code3
//  3. 分号分隔：code1;code2;code3
//  4. 空格/Tab分隔：code1 code2 code3
//  5. JSON数组：["code1","code2","code3"]
//  6. 账号密码格式：account----password 或 account:password 或 account|password
//  7. CSV格式：自动跳过表头行
//  8. 混合格式：自动识别并处理
func parseKamiCodes(codesText string) []string {
	codesText = strings.TrimSpace(codesText)
	if codesText == "" {
		return nil
	}

	var codes []string

	// 尝试解析 JSON 数组格式
	if strings.HasPrefix(codesText, "[") {
		var jsonCodes []string
		if err := json.Unmarshal([]byte(codesText), &jsonCodes); err == nil {
			for _, code := range jsonCodes {
				code = strings.TrimSpace(code)
				if code != "" && !isHeaderLine(code) {
					codes = append(codes, code)
				}
			}
			if len(codes) > 0 {
				return codes
			}
		}
	}

	// 统一换行符
	codesText = strings.ReplaceAll(codesText, "\r\n", "\n")
	codesText = strings.ReplaceAll(codesText, "\r", "\n")

	// 按行分割
	lines := strings.Split(codesText, "\n")

	// 检测分隔符类型（基于第一个非空行）
	var primaryDelimiter string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || isHeaderLine(line) {
			continue
		}
		// 检测主要分隔符
		if strings.Contains(line, ",") {
			primaryDelimiter = ","
		} else if strings.Contains(line, ";") {
			primaryDelimiter = ";"
		} else if strings.Contains(line, "\t") {
			primaryDelimiter = "\t"
		}
		break
	}

	// 解析每一行
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || isHeaderLine(line) {
			continue
		}

		// 如果检测到分隔符，按分隔符拆分
		if primaryDelimiter != "" {
			parts := strings.Split(line, primaryDelimiter)
			for _, part := range parts {
				code := strings.TrimSpace(part)
				if code != "" && !isHeaderLine(code) {
					codes = append(codes, code)
				}
			}
		} else {
			// 单行单卡密
			codes = append(codes, line)
		}
	}

	return codes
}

// isHeaderLine 判断是否为表头行（CSV/Excel导出常见）
func isHeaderLine(line string) bool {
	lower := strings.ToLower(line)
	headers := []string{
		"kami", "code", "卡密", "密码", "password", "key", "serial",
		"序列号", "激活码", "兑换码", "cdkey", "cd-key", "license",
		"账号", "account", "username", "id", "编号",
	}
	for _, h := range headers {
		if lower == h || strings.HasPrefix(lower, h+",") || strings.HasPrefix(lower, h+"\t") {
			return true
		}
	}
	return false
}

// detectKamiFormat 检测卡密格式类型
// 返回格式描述，用于前端显示
func detectKamiFormat(code string) string {
	// 账号密码格式检测
	if strings.Contains(code, "----") {
		return "账号----密码"
	}
	if matched, _ := regexp.MatchString(`^[^:]+:[^:]+$`, code); matched && !strings.Contains(code, "://") {
		return "账号:密码"
	}
	if matched, _ := regexp.MatchString(`^[^|]+\|[^|]+$`, code); matched {
		return "账号|密码"
	}
	// 常见卡密格式
	if matched, _ := regexp.MatchString(`^[A-Z0-9]{4,5}-[A-Z0-9]{4,5}-[A-Z0-9]{4,5}`, code); matched {
		return "标准卡密 (XXXX-XXXX-XXXX)"
	}
	if matched, _ := regexp.MatchString(`^[A-Fa-f0-9]{32}$`, code); matched {
		return "MD5格式"
	}
	if matched, _ := regexp.MatchString(`^[A-Za-z0-9+/]{20,}={0,2}$`, code); matched {
		return "Base64格式"
	}
	if matched, _ := regexp.MatchString(`^[A-Za-z0-9]{16,}$`, code); matched {
		return "纯字母数字"
	}
	return "自定义格式"
}

// ImportKamiCodes 批量导入卡密
// 参数：
//   - productID: 商品ID
//   - codesText: 卡密文本（支持多种格式自动识别）
//
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

	// 使用智能解析函数
	codes := parseKamiCodes(codesText)

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
