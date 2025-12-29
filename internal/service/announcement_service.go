package service

import (
	"errors"
	"time"

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

	return announcement, nil
}

// DeleteAnnouncement 删除公告
func (s *AnnouncementService) DeleteAnnouncement(id uint) error {
	return s.repo.DeleteAnnouncement(id)
}

// GetAllAnnouncements 获取所有公告（管理后台）
func (s *AnnouncementService) GetAllAnnouncements() ([]model.Announcement, error) {
	return s.repo.GetAllAnnouncements()
}

// GetActiveAnnouncements 获取有效公告（前台展示）
func (s *AnnouncementService) GetActiveAnnouncements() ([]model.Announcement, error) {
	return s.repo.GetActiveAnnouncements()
}

// GetAnnouncementByID 获取公告详情
func (s *AnnouncementService) GetAnnouncementByID(id uint) (*model.Announcement, error) {
	return s.repo.GetAnnouncementByID(id)
}
