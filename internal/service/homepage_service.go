package service

import (
	"encoding/json"
	"user-frontend/internal/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// HomepageService 首页配置服务
type HomepageService struct {
	db *gorm.DB
}

// NewHomepageService 创建首页配置服务
func NewHomepageService(db *gorm.DB) *HomepageService {
	return &HomepageService{db: db}
}

// GetActiveConfig 获取当前启用的首页配置
func (s *HomepageService) GetActiveConfig() (*model.HomepageFullConfig, error) {
	var config model.HomepageConfig
	err := s.db.Where("is_active = ?", true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 返回默认配置
			defaultConfig := model.GetDefaultConfig("modern")
			return &defaultConfig, nil
		}
		return nil, err
	}

	var fullConfig model.HomepageFullConfig
	if err := json.Unmarshal(config.Config, &fullConfig); err != nil {
		return nil, err
	}
	fullConfig.Template = config.Template

	return &fullConfig, nil
}

// SaveConfig 保存首页配置
func (s *HomepageService) SaveConfig(fullConfig *model.HomepageFullConfig) error {
	configJSON, err := json.Marshal(fullConfig)
	if err != nil {
		return err
	}

	// 先将所有配置设为非活动
	s.db.Model(&model.HomepageConfig{}).Where("is_active = ?", true).Update("is_active", false)

	// 创建新配置
	config := model.HomepageConfig{
		Template: fullConfig.Template,
		Config:   datatypes.JSON(configJSON),
		IsActive: true,
	}

	return s.db.Create(&config).Error
}

// UpdateConfig 更新首页配置
func (s *HomepageService) UpdateConfig(fullConfig *model.HomepageFullConfig) error {
	configJSON, err := json.Marshal(fullConfig)
	if err != nil {
		return err
	}

	var config model.HomepageConfig
	err = s.db.Where("is_active = ?", true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 不存在则创建
			return s.SaveConfig(fullConfig)
		}
		return err
	}

	config.Template = fullConfig.Template
	config.Config = datatypes.JSON(configJSON)

	return s.db.Save(&config).Error
}

// GetTemplateList 获取可用模板列表
func (s *HomepageService) GetTemplateList() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":          "modern",
			"name":        "现代简约",
			"description": "简洁大气的现代风格，适合大多数场景",
			"preview":     "/templates/modern.png",
		},
		{
			"id":          "gradient",
			"name":        "渐变炫彩",
			"description": "丰富的渐变色彩，视觉冲击力强",
			"preview":     "/templates/gradient.png",
		},
		{
			"id":          "minimal",
			"name":        "极简风格",
			"description": "极简设计，突出内容本身",
			"preview":     "/templates/minimal.png",
		},
		{
			"id":          "card",
			"name":        "卡片风格",
			"description": "卡片式布局，层次分明",
			"preview":     "/templates/card.png",
		},
		{
			"id":          "hero",
			"name":        "大图展示",
			"description": "全屏大图背景，适合品牌展示",
			"preview":     "/templates/hero.png",
		},
		{
			"id":          "business",
			"name":        "商务专业",
			"description": "专业商务风格，适合企业用户",
			"preview":     "/templates/business.png",
		},
	}
}

// GetDefaultConfigByTemplate 根据模板获取默认配置
func (s *HomepageService) GetDefaultConfigByTemplate(template string) model.HomepageFullConfig {
	config := model.GetDefaultConfig(template)
	
	// 根据不同模板调整默认配置
	switch template {
	case "gradient":
		config.PrimaryColor = "#ec4899"
		config.SecondaryColor = "#8b5cf6"
		config.HeroBackground = "gradient"
	case "minimal":
		config.PrimaryColor = "#374151"
		config.SecondaryColor = "#6b7280"
		config.StatsEnabled = false
		config.CTAEnabled = false
	case "card":
		config.PrimaryColor = "#3b82f6"
		config.SecondaryColor = "#06b6d4"
	case "hero":
		config.HeroBackground = "image"
		config.HeroBgImage = "/images/hero-bg.jpg"
	case "business":
		config.PrimaryColor = "#1e40af"
		config.SecondaryColor = "#3b82f6"
		config.Features = []model.FeatureItem{
			{Icon: "fa-shield-halved", Title: "安全保障", Description: "企业级安全防护，数据加密传输"},
			{Icon: "fa-clock", Title: "高效服务", Description: "7x24小时自动发货，即买即用"},
			{Icon: "fa-headset", Title: "专业支持", Description: "专业技术团队，全程服务保障"},
			{Icon: "fa-certificate", Title: "正版授权", Description: "官方正版授权，品质有保证"},
		}
	}
	
	return config
}

// ResetToDefault 重置为默认配置
func (s *HomepageService) ResetToDefault(template string) error {
	defaultConfig := s.GetDefaultConfigByTemplate(template)
	return s.UpdateConfig(&defaultConfig)
}
