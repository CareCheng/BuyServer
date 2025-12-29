// Package service 提供业务逻辑服务
// support_service.go - 客服支持服务（结构体定义和配置管理）
package service

import (
	"crypto/rand"
	"encoding/hex"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// SupportService 客服支持服务
type SupportService struct {
	repo *repository.Repository
}

// NewSupportService 创建客服支持服务
func NewSupportService(repo *repository.Repository) *SupportService {
	return &SupportService{repo: repo}
}

// ==========================================
//         配置管理
// ==========================================

// GetSupportConfig 获取客服配置
func (s *SupportService) GetSupportConfig() (*model.SupportConfigDB, error) {
	var config model.SupportConfigDB
	if err := s.repo.GetDB().First(&config).Error; err != nil {
		// 返回默认配置
		return &model.SupportConfigDB{
			Enabled:           true,
			AllowGuest:        true,
			WorkingHoursStart: "09:00",
			WorkingHoursEnd:   "18:00",
			WorkingDays:       "1,2,3,4,5",
			WelcomeMessage:    "您好，欢迎咨询！请问有什么可以帮助您的？",
			OfflineMessage:    "当前客服不在线，请留言或提交工单，我们会尽快回复您。",
			AutoCloseHours:    72,
			TicketCategories:  `["订单问题","商品咨询","支付问题","账户问题","其他"]`,
		}, nil
	}
	return &config, nil
}

// SaveSupportConfig 保存客服配置
func (s *SupportService) SaveSupportConfig(config *model.SupportConfigDB) error {
	var existing model.SupportConfigDB
	if err := s.repo.GetDB().First(&existing).Error; err != nil {
		// 创建新配置
		return s.repo.GetDB().Create(config).Error
	}
	// 更新现有配置
	config.ID = existing.ID
	return s.repo.GetDB().Save(config).Error
}

// ==========================================
//         辅助函数
// ==========================================

// generateToken 生成随机令牌
func generateToken(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
