package service

import (
	"encoding/json"
	"errors"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// TicketTemplateService 工单模板服务
type TicketTemplateService struct {
	repo *repository.Repository
}

// NewTicketTemplateService 创建工单模板服务实例
func NewTicketTemplateService(repo *repository.Repository) *TicketTemplateService {
	return &TicketTemplateService{repo: repo}
}

// GetTemplates 获取工单模板列表
// 参数：
//   - category: 分类筛选（空表示全部）
//   - onlyEnabled: 是否只返回启用的模板
// 返回：
//   - 模板列表
//   - 错误信息
func (s *TicketTemplateService) GetTemplates(category string, onlyEnabled bool) ([]model.TicketTemplate, error) {
	var templates []model.TicketTemplate

	query := s.repo.GetDB().Model(&model.TicketTemplate{})

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if onlyEnabled {
		query = query.Where("status = 1")
	}

	if err := query.Order("sort_order ASC, use_count DESC").Find(&templates).Error; err != nil {
		return nil, err
	}

	return templates, nil
}

// GetTemplate 获取单个模板详情
// 参数：
//   - templateID: 模板ID
// 返回：
//   - 模板详情
//   - 错误信息
func (s *TicketTemplateService) GetTemplate(templateID uint) (*model.TicketTemplate, error) {
	var template model.TicketTemplate
	if err := s.repo.GetDB().First(&template, templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模板不存在")
		}
		return nil, err
	}
	return &template, nil
}

// CreateTemplate 创建工单模板（管理员功能）
// 参数：
//   - name: 模板名称
//   - description: 模板描述
//   - category: 分类
//   - subject: 预设主题
//   - content: 预设内容
//   - fields: 自定义字段
//   - icon: 图标
//   - sortOrder: 排序
// 返回：
//   - 创建的模板
//   - 错误信息
func (s *TicketTemplateService) CreateTemplate(name, description, category, subject, content string, fields []model.TicketTemplateField, icon string, sortOrder int) (*model.TicketTemplate, error) {
	// 验证必填字段
	if name == "" {
		return nil, errors.New("模板名称不能为空")
	}

	// 序列化自定义字段
	fieldsJSON := "[]"
	if len(fields) > 0 {
		if data, err := json.Marshal(fields); err == nil {
			fieldsJSON = string(data)
		}
	}

	template := &model.TicketTemplate{
		Name:        name,
		Description: description,
		Category:    category,
		Subject:     subject,
		Content:     content,
		Fields:      fieldsJSON,
		Icon:        icon,
		SortOrder:   sortOrder,
		Status:      1,
	}

	if err := s.repo.GetDB().Create(template).Error; err != nil {
		return nil, err
	}

	return template, nil
}

// UpdateTemplate 更新工单模板（管理员功能）
// 参数：
//   - templateID: 模板ID
//   - updates: 更新字段
// 返回：
//   - 错误信息
func (s *TicketTemplateService) UpdateTemplate(templateID uint, updates map[string]interface{}) error {
	// 检查模板是否存在
	var template model.TicketTemplate
	if err := s.repo.GetDB().First(&template, templateID).Error; err != nil {
		return errors.New("模板不存在")
	}

	// 处理自定义字段
	if fields, ok := updates["fields"]; ok {
		if fieldsList, ok := fields.([]model.TicketTemplateField); ok {
			if data, err := json.Marshal(fieldsList); err == nil {
				updates["fields"] = string(data)
			}
		}
	}

	return s.repo.GetDB().Model(&template).Updates(updates).Error
}

// DeleteTemplate 删除工单模板（管理员功能）
// 参数：
//   - templateID: 模板ID
// 返回：
//   - 错误信息
func (s *TicketTemplateService) DeleteTemplate(templateID uint) error {
	return s.repo.GetDB().Delete(&model.TicketTemplate{}, templateID).Error
}

// IncrementUseCount 增加模板使用次数
// 参数：
//   - templateID: 模板ID
// 返回：
//   - 错误信息
func (s *TicketTemplateService) IncrementUseCount(templateID uint) error {
	return s.repo.GetDB().Model(&model.TicketTemplate{}).
		Where("id = ?", templateID).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}

// GetHotTemplates 获取热门模板
// 参数：
//   - limit: 数量限制
// 返回：
//   - 模板列表
//   - 错误信息
func (s *TicketTemplateService) GetHotTemplates(limit int) ([]model.TicketTemplate, error) {
	var templates []model.TicketTemplate

	if err := s.repo.GetDB().Model(&model.TicketTemplate{}).
		Where("status = 1").
		Order("use_count DESC").
		Limit(limit).
		Find(&templates).Error; err != nil {
		return nil, err
	}

	return templates, nil
}

// GetTemplatesByCategory 按分类获取模板
// 返回：
//   - 按分类分组的模板
//   - 错误信息
func (s *TicketTemplateService) GetTemplatesByCategory() (map[string][]model.TicketTemplate, error) {
	var templates []model.TicketTemplate

	if err := s.repo.GetDB().Model(&model.TicketTemplate{}).
		Where("status = 1").
		Order("category ASC, sort_order ASC").
		Find(&templates).Error; err != nil {
		return nil, err
	}

	// 按分类分组
	result := make(map[string][]model.TicketTemplate)
	for _, t := range templates {
		result[t.Category] = append(result[t.Category], t)
	}

	return result, nil
}

// ParseTemplateFields 解析模板自定义字段
// 参数：
//   - fieldsJSON: JSON格式的字段定义
// 返回：
//   - 字段列表
//   - 错误信息
func (s *TicketTemplateService) ParseTemplateFields(fieldsJSON string) ([]model.TicketTemplateField, error) {
	var fields []model.TicketTemplateField
	if fieldsJSON == "" || fieldsJSON == "[]" {
		return fields, nil
	}

	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return nil, err
	}

	return fields, nil
}

// InitDefaultTemplates 初始化默认工单模板
// 返回：
//   - 错误信息
func (s *TicketTemplateService) InitDefaultTemplates() error {
	// 检查是否已有模板
	var count int64
	s.repo.GetDB().Model(&model.TicketTemplate{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 创建默认模板
	defaultTemplates := []model.TicketTemplate{
		{
			Name:        "订单问题",
			Description: "关于订单的咨询或问题",
			Category:    model.TemplateOrderCategory,
			Subject:     "订单问题咨询",
			Content:     "订单号：\n问题描述：\n",
			Icon:        "📦",
			SortOrder:   1,
			Status:      1,
			Fields:      `[{"name":"order_no","label":"订单号","type":"text","required":true,"placeholder":"请输入订单号"}]`,
		},
		{
			Name:        "支付问题",
			Description: "支付失败、退款等问题",
			Category:    model.TemplatePaymentCategory,
			Subject:     "支付问题反馈",
			Content:     "支付方式：\n问题描述：\n",
			Icon:        "💳",
			SortOrder:   2,
			Status:      1,
			Fields:      `[{"name":"payment_method","label":"支付方式","type":"select","required":true,"options":["支付宝","微信","PayPal","其他"]}]`,
		},
		{
			Name:        "商品咨询",
			Description: "商品功能、使用方法等咨询",
			Category:    model.TemplateProductCategory,
			Subject:     "商品咨询",
			Content:     "商品名称：\n咨询内容：\n",
			Icon:        "🛍️",
			SortOrder:   3,
			Status:      1,
			Fields:      `[{"name":"product_name","label":"商品名称","type":"text","required":false,"placeholder":"请输入商品名称"}]`,
		},
		{
			Name:        "账户问题",
			Description: "账户登录、密码、安全等问题",
			Category:    model.TemplateAccountCategory,
			Subject:     "账户问题",
			Content:     "问题类型：\n详细描述：\n",
			Icon:        "👤",
			SortOrder:   4,
			Status:      1,
			Fields:      `[{"name":"issue_type","label":"问题类型","type":"select","required":true,"options":["无法登录","忘记密码","账户安全","其他"]}]`,
		},
		{
			Name:        "其他问题",
			Description: "其他类型的问题或建议",
			Category:    model.TemplateOtherCategory,
			Subject:     "问题反馈",
			Content:     "",
			Icon:        "❓",
			SortOrder:   5,
			Status:      1,
			Fields:      `[]`,
		},
	}

	for _, t := range defaultTemplates {
		if err := s.repo.GetDB().Create(&t).Error; err != nil {
			return err
		}
	}

	return nil
}
