package service

import (
	"encoding/json"
	"errors"
	"time"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// UndoService 操作撤销服务
type UndoService struct {
	repo *repository.Repository
}

// NewUndoService 创建撤销服务实例
func NewUndoService(repo *repository.Repository) *UndoService {
	return &UndoService{repo: repo}
}

// GetConfig 获取撤销配置
// 返回：
//   - 撤销配置
//   - 错误信息（如有）
func (s *UndoService) GetConfig() (*model.UndoConfig, error) {
	var config model.UndoConfig
	result := s.repo.GetDB().First(&config)
	if result.Error != nil {
		// 返回默认配置
		return &model.UndoConfig{
			Enabled:        true,
			RetentionHours: 24,
			AllowedTypes:   `["product_delete","product_disable","user_disable","coupon_delete","coupon_disable","category_delete","announcement_delete"]`,
		}, nil
	}
	return &config, nil
}

// SaveConfig 保存撤销配置
// 参数：
//   - config: 配置信息
// 返回：
//   - 错误信息（如有）
func (s *UndoService) SaveConfig(config *model.UndoConfig) error {
	var existing model.UndoConfig
	result := s.repo.GetDB().First(&existing)
	if result.Error != nil {
		return s.repo.GetDB().Create(config).Error
	}
	config.ID = existing.ID
	return s.repo.GetDB().Save(config).Error
}

// RecordOperation 记录可撤销操作
// 参数：
//   - operationType: 操作类型
//   - targetType: 目标类型
//   - targetID: 目标ID
//   - targetName: 目标名称
//   - originalData: 原始数据
//   - adminName: 管理员名称
// 返回：
//   - 操作记录
//   - 错误信息（如有）
func (s *UndoService) RecordOperation(operationType, targetType string, targetID uint, targetName string, originalData interface{}, adminName string) (*model.UndoOperation, error) {
	config, _ := s.GetConfig()
	if !config.Enabled {
		return nil, nil // 未启用撤销功能
	}

	// 检查操作类型是否允许撤销
	var allowedTypes []string
	json.Unmarshal([]byte(config.AllowedTypes), &allowedTypes)
	allowed := false
	for _, t := range allowedTypes {
		if t == operationType {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, nil // 该操作类型不支持撤销
	}

	// 序列化原始数据
	dataJSON, err := json.Marshal(originalData)
	if err != nil {
		return nil, err
	}

	// 计算过期时间
	expireAt := time.Now().Add(time.Duration(config.RetentionHours) * time.Hour)

	operation := &model.UndoOperation{
		OperationType: operationType,
		TargetType:    targetType,
		TargetID:      targetID,
		TargetName:    targetName,
		OriginalData:  string(dataJSON),
		AdminName:     adminName,
		Status:        0,
		ExpireAt:      expireAt,
	}

	if err := s.repo.GetDB().Create(operation).Error; err != nil {
		return nil, err
	}

	return operation, nil
}

// GetUndoableOperations 获取可撤销操作列表
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 操作列表
//   - 总数
//   - 错误信息（如有）
func (s *UndoService) GetUndoableOperations(page, pageSize int) ([]model.UndoOperation, int64, error) {
	var total int64
	now := time.Now()
	s.repo.GetDB().Model(&model.UndoOperation{}).Where("status = 0 AND expire_at > ?", now).Count(&total)

	var operations []model.UndoOperation
	offset := (page - 1) * pageSize
	err := s.repo.GetDB().Where("status = 0 AND expire_at > ?", now).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&operations).Error

	return operations, total, err
}

// GetAllOperations 获取所有操作记录（包括已撤销和已过期）
// 参数：
//   - status: 状态筛选（-1表示全部）
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 操作列表
//   - 总数
//   - 错误信息（如有）
func (s *UndoService) GetAllOperations(status, page, pageSize int) ([]model.UndoOperation, int64, error) {
	var total int64
	query := s.repo.GetDB().Model(&model.UndoOperation{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	var operations []model.UndoOperation
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&operations).Error

	return operations, total, err
}

// UndoOperation 撤销操作
// 参数：
//   - operationID: 操作记录ID
//   - adminName: 撤销人
// 返回：
//   - 错误信息（如有）
func (s *UndoService) UndoOperation(operationID uint, adminName string) error {
	var operation model.UndoOperation
	if err := s.repo.GetDB().First(&operation, operationID).Error; err != nil {
		return errors.New("操作记录不存在")
	}

	if operation.Status != 0 {
		return errors.New("该操作已撤销或已过期")
	}

	if time.Now().After(operation.ExpireAt) {
		operation.Status = 2 // 标记为已过期
		s.repo.GetDB().Save(&operation)
		return errors.New("该操作已过期，无法撤销")
	}

	// 根据操作类型执行撤销
	var err error
	switch operation.OperationType {
	case model.UndoTypeProductDelete:
		err = s.undoProductDelete(&operation)
	case model.UndoTypeProductDisable:
		err = s.undoProductDisable(&operation)
	case model.UndoTypeUserDisable:
		err = s.undoUserDisable(&operation)
	case model.UndoTypeCouponDelete:
		err = s.undoCouponDelete(&operation)
	case model.UndoTypeCouponDisable:
		err = s.undoCouponDisable(&operation)
	case model.UndoTypeCategoryDelete:
		err = s.undoCategoryDelete(&operation)
	case model.UndoTypeAnnouncementDelete:
		err = s.undoAnnouncementDelete(&operation)
	default:
		return errors.New("不支持的操作类型")
	}

	if err != nil {
		return err
	}

	// 更新操作状态
	now := time.Now()
	operation.Status = 1
	operation.UndoneAt = &now
	operation.UndoneBy = adminName
	return s.repo.GetDB().Save(&operation).Error
}

// undoProductDelete 撤销商品删除
func (s *UndoService) undoProductDelete(operation *model.UndoOperation) error {
	var product model.Product
	if err := json.Unmarshal([]byte(operation.OriginalData), &product); err != nil {
		return errors.New("数据解析失败")
	}

	// 恢复商品（软删除恢复）
	return s.repo.GetDB().Unscoped().Model(&model.Product{}).Where("id = ?", operation.TargetID).Update("deleted_at", nil).Error
}

// undoProductDisable 撤销商品禁用
func (s *UndoService) undoProductDisable(operation *model.UndoOperation) error {
	return s.repo.GetDB().Model(&model.Product{}).Where("id = ?", operation.TargetID).Update("status", 1).Error
}

// undoUserDisable 撤销用户禁用
func (s *UndoService) undoUserDisable(operation *model.UndoOperation) error {
	return s.repo.GetDB().Model(&model.User{}).Where("id = ?", operation.TargetID).Update("status", 1).Error
}

// undoCouponDelete 撤销优惠券删除
func (s *UndoService) undoCouponDelete(operation *model.UndoOperation) error {
	return s.repo.GetDB().Unscoped().Model(&model.Coupon{}).Where("id = ?", operation.TargetID).Update("deleted_at", nil).Error
}

// undoCouponDisable 撤销优惠券禁用
func (s *UndoService) undoCouponDisable(operation *model.UndoOperation) error {
	return s.repo.GetDB().Model(&model.Coupon{}).Where("id = ?", operation.TargetID).Update("status", 1).Error
}

// undoCategoryDelete 撤销分类删除
func (s *UndoService) undoCategoryDelete(operation *model.UndoOperation) error {
	return s.repo.GetDB().Unscoped().Model(&model.ProductCategory{}).Where("id = ?", operation.TargetID).Update("deleted_at", nil).Error
}

// undoAnnouncementDelete 撤销公告删除
func (s *UndoService) undoAnnouncementDelete(operation *model.UndoOperation) error {
	return s.repo.GetDB().Unscoped().Model(&model.Announcement{}).Where("id = ?", operation.TargetID).Update("deleted_at", nil).Error
}

// CleanupExpiredOperations 清理过期的操作记录
func (s *UndoService) CleanupExpiredOperations() {
	now := time.Now()
	// 标记过期的操作
	s.repo.GetDB().Model(&model.UndoOperation{}).
		Where("status = 0 AND expire_at < ?", now).
		Update("status", 2)

	// 删除7天前的已撤销或已过期记录
	sevenDaysAgo := now.AddDate(0, 0, -7)
	s.repo.GetDB().Where("status != 0 AND created_at < ?", sevenDaysAgo).Delete(&model.UndoOperation{})
}

// GetUndoStats 获取撤销统计
// 返回：
//   - 统计信息
func (s *UndoService) GetUndoStats() map[string]interface{} {
	var undoable, undone, expired int64
	now := time.Now()

	s.repo.GetDB().Model(&model.UndoOperation{}).Where("status = 0 AND expire_at > ?", now).Count(&undoable)
	s.repo.GetDB().Model(&model.UndoOperation{}).Where("status = 1").Count(&undone)
	s.repo.GetDB().Model(&model.UndoOperation{}).Where("status = 2").Count(&expired)

	return map[string]interface{}{
		"undoable": undoable,
		"undone":   undone,
		"expired":  expired,
	}
}
