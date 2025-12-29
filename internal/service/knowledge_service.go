package service

import (
	"errors"
	"strings"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// KnowledgeService 知识库服务
type KnowledgeService struct {
	repo *repository.Repository
}

// NewKnowledgeService 创建知识库服务实例
func NewKnowledgeService(repo *repository.Repository) *KnowledgeService {
	return &KnowledgeService{repo: repo}
}

// ==================== 分类管理 ====================

// GetCategories 获取知识库分类列表
// 参数：
//   - parentID: 父分类ID（0表示顶级分类）
//   - onlyEnabled: 是否只返回启用的分类
// 返回：
//   - 分类列表
//   - 错误信息
func (s *KnowledgeService) GetCategories(parentID uint, onlyEnabled bool) ([]model.KnowledgeCategory, error) {
	var categories []model.KnowledgeCategory

	query := s.repo.GetDB().Model(&model.KnowledgeCategory{}).Where("parent_id = ?", parentID)
	if onlyEnabled {
		query = query.Where("status = 1")
	}

	if err := query.Order("sort_order ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// GetAllCategories 获取所有分类（树形结构）
// 参数：
//   - onlyEnabled: 是否只返回启用的分类
// 返回：
//   - 分类列表
//   - 错误信息
func (s *KnowledgeService) GetAllCategories(onlyEnabled bool) ([]model.KnowledgeCategory, error) {
	var categories []model.KnowledgeCategory

	query := s.repo.GetDB().Model(&model.KnowledgeCategory{})
	if onlyEnabled {
		query = query.Where("status = 1")
	}

	if err := query.Order("parent_id ASC, sort_order ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// CreateCategory 创建知识库分类
// 参数：
//   - name: 分类名称
//   - description: 分类描述
//   - icon: 图标
//   - parentID: 父分类ID
//   - sortOrder: 排序
// 返回：
//   - 创建的分类
//   - 错误信息
func (s *KnowledgeService) CreateCategory(name, description, icon string, parentID uint, sortOrder int) (*model.KnowledgeCategory, error) {
	if name == "" {
		return nil, errors.New("分类名称不能为空")
	}

	category := &model.KnowledgeCategory{
		Name:        name,
		Description: description,
		Icon:        icon,
		ParentID:    parentID,
		SortOrder:   sortOrder,
		Status:      1,
	}

	if err := s.repo.GetDB().Create(category).Error; err != nil {
		return nil, err
	}

	return category, nil
}

// UpdateCategory 更新知识库分类
// 参数：
//   - categoryID: 分类ID
//   - updates: 更新字段
// 返回：
//   - 错误信息
func (s *KnowledgeService) UpdateCategory(categoryID uint, updates map[string]interface{}) error {
	return s.repo.GetDB().Model(&model.KnowledgeCategory{}).Where("id = ?", categoryID).Updates(updates).Error
}

// DeleteCategory 删除知识库分类
// 参数：
//   - categoryID: 分类ID
// 返回：
//   - 错误信息
func (s *KnowledgeService) DeleteCategory(categoryID uint) error {
	// 检查是否有子分类
	var childCount int64
	s.repo.GetDB().Model(&model.KnowledgeCategory{}).Where("parent_id = ?", categoryID).Count(&childCount)
	if childCount > 0 {
		return errors.New("该分类下存在子分类，无法删除")
	}

	// 检查是否有文章
	var articleCount int64
	s.repo.GetDB().Model(&model.KnowledgeArticle{}).Where("category_id = ?", categoryID).Count(&articleCount)
	if articleCount > 0 {
		return errors.New("该分类下存在文章，无法删除")
	}

	return s.repo.GetDB().Delete(&model.KnowledgeCategory{}, categoryID).Error
}

// ==================== 文章管理 ====================

// GetArticles 获取知识库文章列表
// 参数：
//   - categoryID: 分类ID（0表示全部）
//   - keyword: 搜索关键词
//   - status: 状态筛选（-1表示全部）
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 文章列表
//   - 总数
//   - 错误信息
func (s *KnowledgeService) GetArticles(categoryID uint, keyword string, status, page, pageSize int) ([]model.KnowledgeArticle, int64, error) {
	var articles []model.KnowledgeArticle
	var total int64

	query := s.repo.GetDB().Model(&model.KnowledgeArticle{})

	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ? OR tags LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("sort_order ASC, updated_at DESC").
		Offset(offset).Limit(pageSize).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// GetArticle 获取文章详情
// 参数：
//   - articleID: 文章ID
// 返回：
//   - 文章详情
//   - 错误信息
func (s *KnowledgeService) GetArticle(articleID uint) (*model.KnowledgeArticle, error) {
	var article model.KnowledgeArticle
	if err := s.repo.GetDB().First(&article, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文章不存在")
		}
		return nil, err
	}
	return &article, nil
}

// CreateArticle 创建知识库文章
// 参数：
//   - categoryID: 分类ID
//   - title: 标题
//   - content: 内容
//   - summary: 摘要
//   - tags: 标签
//   - sortOrder: 排序
//   - status: 状态
//   - createdBy: 创建人
// 返回：
//   - 创建的文章
//   - 错误信息
func (s *KnowledgeService) CreateArticle(categoryID uint, title, content, summary, tags string, sortOrder, status int, createdBy string) (*model.KnowledgeArticle, error) {
	if title == "" {
		return nil, errors.New("文章标题不能为空")
	}

	// 自动生成摘要
	if summary == "" && content != "" {
		summary = truncateString(stripMarkdown(content), 200)
	}

	article := &model.KnowledgeArticle{
		CategoryID: categoryID,
		Title:      title,
		Content:    content,
		Summary:    summary,
		Tags:       tags,
		SortOrder:  sortOrder,
		Status:     status,
		CreatedBy:  createdBy,
		UpdatedBy:  createdBy,
	}

	if err := s.repo.GetDB().Create(article).Error; err != nil {
		return nil, err
	}

	return article, nil
}

// UpdateArticle 更新知识库文章
// 参数：
//   - articleID: 文章ID
//   - updates: 更新字段
//   - updatedBy: 更新人
// 返回：
//   - 错误信息
func (s *KnowledgeService) UpdateArticle(articleID uint, updates map[string]interface{}, updatedBy string) error {
	updates["updated_by"] = updatedBy
	return s.repo.GetDB().Model(&model.KnowledgeArticle{}).Where("id = ?", articleID).Updates(updates).Error
}

// DeleteArticle 删除知识库文章
// 参数：
//   - articleID: 文章ID
// 返回：
//   - 错误信息
func (s *KnowledgeService) DeleteArticle(articleID uint) error {
	return s.repo.GetDB().Delete(&model.KnowledgeArticle{}, articleID).Error
}

// IncrementViewCount 增加文章浏览次数
// 参数：
//   - articleID: 文章ID
// 返回：
//   - 错误信息
func (s *KnowledgeService) IncrementViewCount(articleID uint) error {
	return s.repo.GetDB().Model(&model.KnowledgeArticle{}).
		Where("id = ?", articleID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// IncrementUseCount 增加文章使用次数（客服引用）
// 参数：
//   - articleID: 文章ID
// 返回：
//   - 错误信息
func (s *KnowledgeService) IncrementUseCount(articleID uint) error {
	return s.repo.GetDB().Model(&model.KnowledgeArticle{}).
		Where("id = ?", articleID).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}

// SubmitFeedback 提交文章反馈
// 参数：
//   - articleID: 文章ID
//   - helpful: 是否有帮助
// 返回：
//   - 错误信息
func (s *KnowledgeService) SubmitFeedback(articleID uint, helpful bool) error {
	column := "not_helpful"
	if helpful {
		column = "helpful"
	}
	return s.repo.GetDB().Model(&model.KnowledgeArticle{}).
		Where("id = ?", articleID).
		UpdateColumn(column, gorm.Expr(column+" + 1")).Error
}

// SearchArticles 搜索文章
// 参数：
//   - keyword: 搜索关键词
//   - limit: 数量限制
// 返回：
//   - 文章列表
//   - 错误信息
func (s *KnowledgeService) SearchArticles(keyword string, limit int) ([]model.KnowledgeArticle, error) {
	var articles []model.KnowledgeArticle

	if keyword == "" {
		return articles, nil
	}

	if err := s.repo.GetDB().Model(&model.KnowledgeArticle{}).
		Where("status = 1").
		Where("title LIKE ? OR content LIKE ? OR tags LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").
		Order("use_count DESC, view_count DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, err
	}

	return articles, nil
}

// GetHotArticles 获取热门文章
// 参数：
//   - limit: 数量限制
// 返回：
//   - 文章列表
//   - 错误信息
func (s *KnowledgeService) GetHotArticles(limit int) ([]model.KnowledgeArticle, error) {
	var articles []model.KnowledgeArticle

	if err := s.repo.GetDB().Model(&model.KnowledgeArticle{}).
		Where("status = 1").
		Order("use_count DESC, view_count DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, err
	}

	return articles, nil
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// stripMarkdown 简单移除Markdown标记
func stripMarkdown(content string) string {
	// 简单处理：移除常见Markdown标记
	content = strings.ReplaceAll(content, "#", "")
	content = strings.ReplaceAll(content, "*", "")
	content = strings.ReplaceAll(content, "_", "")
	content = strings.ReplaceAll(content, "`", "")
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.TrimSpace(content)
	return content
}
