package service

import (
	"errors"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// FAQService FAQ服务
type FAQService struct {
	repo *repository.Repository
	db   *gorm.DB
}

// NewFAQService 创建FAQ服务实例
func NewFAQService(repo *repository.Repository) *FAQService {
	return &FAQService{
		repo: repo,
		db:   repo.GetDB(),
	}
}

// ==================== FAQ分类管理 ====================

// GetCategories 获取所有启用的FAQ分类
func (s *FAQService) GetCategories() ([]model.FAQCategory, error) {
	var categories []model.FAQCategory
	err := s.db.Where("status = ?", 1).Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, err
}

// GetAllCategories 获取所有FAQ分类（管理后台）
func (s *FAQService) GetAllCategories() ([]model.FAQCategory, error) {
	var categories []model.FAQCategory
	err := s.db.Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, err
}

// GetCategoryByID 根据ID获取分类
func (s *FAQService) GetCategoryByID(id uint) (*model.FAQCategory, error) {
	var category model.FAQCategory
	err := s.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// CreateCategory 创建FAQ分类
func (s *FAQService) CreateCategory(category *model.FAQCategory) error {
	return s.db.Create(category).Error
}

// UpdateCategory 更新FAQ分类
func (s *FAQService) UpdateCategory(category *model.FAQCategory) error {
	return s.db.Save(category).Error
}

// DeleteCategory 删除FAQ分类
func (s *FAQService) DeleteCategory(id uint) error {
	// 检查分类下是否有FAQ
	var count int64
	s.db.Model(&model.FAQ{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该分类下还有FAQ，无法删除")
	}
	return s.db.Delete(&model.FAQCategory{}, id).Error
}

// ==================== FAQ管理 ====================

// GetFAQsByCategory 根据分类获取FAQ列表
func (s *FAQService) GetFAQsByCategory(categoryID uint) ([]model.FAQ, error) {
	var faqs []model.FAQ
	query := s.db.Where("status = ?", 1)
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	err := query.Order("sort_order ASC, id ASC").Find(&faqs).Error
	return faqs, err
}

// GetAllFAQs 获取所有FAQ（管理后台）
func (s *FAQService) GetAllFAQs(page, pageSize int, categoryID uint, keyword string) ([]model.FAQ, int64, error) {
	var faqs []model.FAQ
	var total int64

	query := s.db.Model(&model.FAQ{})
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		query = query.Where("question LIKE ? OR answer LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("sort_order ASC, id DESC").Offset(offset).Limit(pageSize).Find(&faqs).Error
	return faqs, total, err
}

// GetFAQByID 根据ID获取FAQ
func (s *FAQService) GetFAQByID(id uint) (*model.FAQ, error) {
	var faq model.FAQ
	err := s.db.First(&faq, id).Error
	if err != nil {
		return nil, err
	}
	return &faq, nil
}

// CreateFAQ 创建FAQ
func (s *FAQService) CreateFAQ(faq *model.FAQ) error {
	return s.db.Create(faq).Error
}

// UpdateFAQ 更新FAQ
func (s *FAQService) UpdateFAQ(faq *model.FAQ) error {
	return s.db.Save(faq).Error
}

// DeleteFAQ 删除FAQ
func (s *FAQService) DeleteFAQ(id uint) error {
	return s.db.Delete(&model.FAQ{}, id).Error
}

// IncrementViewCount 增加浏览次数
func (s *FAQService) IncrementViewCount(id uint) error {
	return s.db.Model(&model.FAQ{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// ==================== FAQ反馈 ====================

// SubmitFeedback 提交FAQ反馈
func (s *FAQService) SubmitFeedback(faqID uint, userID uint, sessionID string, helpful bool) error {
	// 检查是否已经反馈过
	var existing model.FAQFeedback
	query := s.db.Where("faq_id = ?", faqID)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("session_id = ?", sessionID)
	}
	
	if err := query.First(&existing).Error; err == nil {
		// 已经反馈过，更新反馈
		if existing.Helpful != helpful {
			// 更新反馈类型
			existing.Helpful = helpful
			s.db.Save(&existing)
			
			// 更新FAQ统计
			if helpful {
				s.db.Model(&model.FAQ{}).Where("id = ?", faqID).
					UpdateColumns(map[string]interface{}{
						"helpful":     gorm.Expr("helpful + 1"),
						"not_helpful": gorm.Expr("CASE WHEN not_helpful > 0 THEN not_helpful - 1 ELSE 0 END"),
					})
			} else {
				s.db.Model(&model.FAQ{}).Where("id = ?", faqID).
					UpdateColumns(map[string]interface{}{
						"helpful":     gorm.Expr("CASE WHEN helpful > 0 THEN helpful - 1 ELSE 0 END"),
						"not_helpful": gorm.Expr("not_helpful + 1"),
					})
			}
		}
		return nil
	}

	// 创建新反馈
	feedback := &model.FAQFeedback{
		FAQID:     faqID,
		UserID:    userID,
		SessionID: sessionID,
		Helpful:   helpful,
	}
	if err := s.db.Create(feedback).Error; err != nil {
		return err
	}

	// 更新FAQ统计
	if helpful {
		s.db.Model(&model.FAQ{}).Where("id = ?", faqID).
			UpdateColumn("helpful", gorm.Expr("helpful + 1"))
	} else {
		s.db.Model(&model.FAQ{}).Where("id = ?", faqID).
			UpdateColumn("not_helpful", gorm.Expr("not_helpful + 1"))
	}

	return nil
}

// GetUserFeedback 获取用户对某FAQ的反馈状态
func (s *FAQService) GetUserFeedback(faqID uint, userID uint, sessionID string) (*model.FAQFeedback, error) {
	var feedback model.FAQFeedback
	query := s.db.Where("faq_id = ?", faqID)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("session_id = ?", sessionID)
	}
	
	err := query.First(&feedback).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

// SearchFAQs 搜索FAQ
func (s *FAQService) SearchFAQs(keyword string) ([]model.FAQ, error) {
	var faqs []model.FAQ
	err := s.db.Where("status = ?", 1).
		Where("question LIKE ? OR answer LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Order("view_count DESC, helpful DESC").
		Limit(20).
		Find(&faqs).Error
	return faqs, err
}

// GetHotFAQs 获取热门FAQ
func (s *FAQService) GetHotFAQs(limit int) ([]model.FAQ, error) {
	var faqs []model.FAQ
	err := s.db.Where("status = ?", 1).
		Order("view_count DESC, helpful DESC").
		Limit(limit).
		Find(&faqs).Error
	return faqs, err
}
