package service

import (
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"github.com/google/uuid"
)

// SessionService 会话服务（数据库持久化）
type SessionService struct {
	repo *repository.Repository
}

// 会话过期时间
const (
	UserSessionDuration  = 2 * time.Hour  // 用户会话2小时
	AdminSessionDuration = 1 * time.Hour  // 管理员会话1小时
	RememberMeDuration   = 7 * 24 * time.Hour // 记住我7天
)

func NewSessionService(repo *repository.Repository) *SessionService {
	return &SessionService{repo: repo}
}

// CreateUserSession 创建用户会话
func (s *SessionService) CreateUserSession(userID uint, username, ip, userAgent string, remember bool) (string, error) {
	sessionID := uuid.New().String()
	duration := UserSessionDuration
	if remember {
		duration = RememberMeDuration
	}

	session := &model.UserSession{
		SessionID: sessionID,
		UserID:    userID,
		Username:  username,
		IP:        ip,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(duration),
	}

	if err := s.repo.CreateUserSession(session); err != nil {
		return "", err
	}

	return sessionID, nil
}

// GetUserSession 获取用户会话
func (s *SessionService) GetUserSession(sessionID string) (*model.UserSession, error) {
	return s.repo.GetUserSession(sessionID)
}

// RefreshUserSession 刷新用户会话
func (s *SessionService) RefreshUserSession(sessionID string) error {
	session, err := s.repo.GetUserSession(sessionID)
	if err != nil {
		return err
	}

	// 如果会话还有超过一半的时间，不刷新
	remaining := time.Until(session.ExpiresAt)
	if remaining > UserSessionDuration/2 {
		return nil
	}

	session.ExpiresAt = time.Now().Add(UserSessionDuration)
	return s.repo.UpdateUserSession(session)
}

// DeleteUserSession 删除用户会话
func (s *SessionService) DeleteUserSession(sessionID string) error {
	return s.repo.DeleteUserSession(sessionID)
}

// DeleteUserSessionsByUserID 删除用户的所有会话
func (s *SessionService) DeleteUserSessionsByUserID(userID uint) error {
	return s.repo.DeleteUserSessionsByUserID(userID)
}

// CreateAdminSession 创建管理员会话
func (s *SessionService) CreateAdminSession(username, role, ip, userAgent string, remember bool) (string, error) {
	sessionID := uuid.New().String()
	duration := AdminSessionDuration
	if remember {
		duration = 24 * time.Hour // 管理员记住我24小时
	}

	session := &model.AdminSession{
		SessionID: sessionID,
		Username:  username,
		Role:      role,
		IP:        ip,
		UserAgent: userAgent,
		Verified:  false,
		ExpiresAt: time.Now().Add(duration),
	}

	if err := s.repo.CreateAdminSession(session); err != nil {
		return "", err
	}

	return sessionID, nil
}

// GetAdminSession 获取管理员会话
func (s *SessionService) GetAdminSession(sessionID string) (*model.AdminSession, error) {
	return s.repo.GetAdminSession(sessionID)
}

// SetAdminSessionVerified 设置管理员会话已验证
func (s *SessionService) SetAdminSessionVerified(sessionID string) error {
	session, err := s.repo.GetAdminSession(sessionID)
	if err != nil {
		return err
	}
	session.Verified = true
	return s.repo.UpdateAdminSession(session)
}

// DeleteAdminSession 删除管理员会话
func (s *SessionService) DeleteAdminSession(sessionID string) error {
	return s.repo.DeleteAdminSession(sessionID)
}

// CleanupExpiredSessions 清理过期会话
func (s *SessionService) CleanupExpiredSessions() {
	s.repo.DeleteExpiredUserSessions()
	s.repo.DeleteExpiredAdminSessions()
}
