package service

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"user-frontend/internal/cache"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// AnnouncementService 公告服务
type AnnouncementService struct {
	repo *repository.Repository
}

func NewAnnouncementService(repo *repository.Repository) *AnnouncementService {
	return &AnnouncementService{repo: repo}
}

// ==================== 缓存辅助方法 ====================

// cacheAnnouncementList 缓存公告列表
func (s *AnnouncementService) cacheAnnouncementList(announcements []model.Announcement) {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	key := cache.AnnouncementListKey()
	data, err := json.Marshal(announcements)
	if err != nil {
		log.Printf("[AnnouncementService] 序列化公告列表缓存失败: %v", err)
		return
	}

	if err := cm.Set(key, string(data), cache.AnnounceTTL); err != nil {
		log.Printf("[AnnouncementService] 缓存公告列表失败: %v", err)
	}
}

// getAnnouncementListFromCache 从缓存获取公告列表
func (s *AnnouncementService) getAnnouncementListFromCache() []model.Announcement {
	cm := cache.GetManager()
	if cm == nil {
		return nil
	}

	key := cache.AnnouncementListKey()
	data, ok := cm.Get(key)
	if !ok {
		return nil
	}

	dataStr, ok := data.(string)
	if !ok {
		return nil
	}

	var announcements []model.Announcement
	if err := json.Unmarshal([]byte(dataStr), &announcements); err != nil {
		log.Printf("[AnnouncementService] 反序列化公告列表缓存失败: %v", err)
		return nil
	}

	return announcements
}

// invalidateAnnouncementCache 使公告缓存失效
func (s *AnnouncementService) invalidateAnnouncementCache() {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	key := cache.AnnouncementListKey()
	if err := cm.Delete(key); err != nil {
		log.Printf("[AnnouncementService] 删除公告列表缓存失败: %v", err)
	}
}

// CreateAnnouncement 创建公告
func (s *AnnouncementService) CreateAnnouncement(title, content, announcementType string, sortOrder int, startAt, endAt *time.Time) (*model.Announcement, error) {
	if title == "" {
		return nil, errors.New("公告标题不能为空")
	}

	if announcementType == "" {
		announcementType = "info"
	}

	announcement := &model.Announcement{
		Title:     title,
		Content:   content,
		Type:      announcementType,
		Status:    1,
		SortOrder: sortOrder,
		StartAt:   startAt,
		EndAt:     endAt,
	}

	if err := s.repo.CreateAnnouncement(announcement); err != nil {
		return nil, err
	}

	// 使缓存失效
	s.invalidateAnnouncementCache()

	return announcement, nil
}

// UpdateAnnouncement 更新公告
func (s *AnnouncementService) UpdateAnnouncement(id uint, title, content, announcementType string, status, sortOrder int, startAt, endAt *time.Time) (*model.Announcement, error) {
	announcement, err := s.repo.GetAnnouncementByID(id)
	if err != nil {
		return nil, errors.New("公告不存在")
	}

	if title != "" {
		announcement.Title = title
	}
	announcement.Content = content
	if announcementType != "" {
		announcement.Type = announcementType
	}
	announcement.Status = status
	announcement.SortOrder = sortOrder
	announcement.StartAt = startAt
	announcement.EndAt = endAt

	if err := s.repo.UpdateAnnouncement(announcement); err != nil {
		return nil, err
	}

	// 使缓存失效
	s.invalidateAnnouncementCache()

	return announcement, nil
}

// DeleteAnnouncement 删除公告
func (s *AnnouncementService) DeleteAnnouncement(id uint) error {
	err := s.repo.DeleteAnnouncement(id)
	if err == nil {
		s.invalidateAnnouncementCache()
	}
	return err
}

// GetAllAnnouncements 获取所有公告（管理后台）
func (s *AnnouncementService) GetAllAnnouncements() ([]model.Announcement, error) {
	return s.repo.GetAllAnnouncements()
}

// GetActiveAnnouncements 获取有效公告（前台展示，支持缓存）
func (s *AnnouncementService) GetActiveAnnouncements() ([]model.Announcement, error) {
	// 先从缓存获取
	if announcements := s.getAnnouncementListFromCache(); announcements != nil {
		return announcements, nil
	}

	// 从数据库获取
	announcements, err := s.repo.GetActiveAnnouncements()
	if err != nil {
		return nil, err
	}

	// 缓存公告列表
	s.cacheAnnouncementList(announcements)

	return announcements, nil
}

// GetAnnouncementByID 获取公告详情
func (s *AnnouncementService) GetAnnouncementByID(id uint) (*model.Announcement, error) {
	return s.repo.GetAnnouncementByID(id)
}
