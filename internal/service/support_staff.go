// Package service 提供业务逻辑服务
// support_staff.go - 客服人员管理相关方法
package service

import (
	"errors"
	"time"

	"user-frontend/internal/model"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// CreateStaff 创建客服账号
func (s *SupportService) CreateStaff(username, password, nickname, email, role string) (*model.SupportStaff, error) {
	// 检查用户名是否存在
	var count int64
	s.repo.GetDB().Model(&model.SupportStaff{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	staff := &model.SupportStaff{
		Username:     username,
		PasswordHash: string(hash),
		Nickname:     nickname,
		Email:        email,
		Role:         role,
		Status:       0, // 默认离线
		MaxTickets:   10,
	}

	if err := s.repo.GetDB().Create(staff).Error; err != nil {
		return nil, err
	}

	return staff, nil
}

// StaffLogin 客服登录
func (s *SupportService) StaffLogin(username, password string) (*model.SupportStaff, string, bool, error) {
	var staff model.SupportStaff
	if err := s.repo.GetDB().Where("username = ?", username).First(&staff).Error; err != nil {
		return nil, "", false, errors.New("用户名或密码错误")
	}

	if staff.Status == -1 {
		return nil, "", false, errors.New("账号已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(password)); err != nil {
		return nil, "", false, errors.New("用户名或密码错误")
	}

	// 检查是否需要二步验证
	config, _ := s.GetSupportConfig()
	needs2FA := config.EnableStaff2FA && staff.Enable2FA && staff.TOTPSecret != ""

	// 创建会话
	sessionID := uuid.New().String()
	session := &model.SupportStaffSession{
		SessionID: sessionID,
		StaffID:   staff.ID,
		Username:  staff.Username,
		Role:      staff.Role,
		Verified:  !needs2FA, // 如果不需要2FA，直接标记为已验证
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	s.repo.GetDB().Create(session)

	// 更新在线状态
	now := time.Now()
	s.repo.GetDB().Model(&staff).Updates(map[string]interface{}{
		"status":         1,
		"last_active_at": now,
	})

	return &staff, sessionID, needs2FA, nil
}

// StaffVerify2FA 客服二步验证
func (s *SupportService) StaffVerify2FA(sessionID, code string) error {
	session, err := s.GetStaffSession(sessionID)
	if err != nil {
		return errors.New("会话不存在或已过期")
	}

	staff, err := s.GetStaffByID(session.StaffID)
	if err != nil {
		return errors.New("客服不存在")
	}

	// 验证TOTP
	if !verifyTOTP(staff.TOTPSecret, code) {
		return errors.New("验证码错误")
	}

	// 更新会话验证状态
	s.repo.GetDB().Model(&model.SupportStaffSession{}).Where("session_id = ?", sessionID).Update("verified", true)

	return nil
}

// StaffGenerate2FASecret 生成客服二步验证密钥
func (s *SupportService) StaffGenerate2FASecret(staffID uint) (string, string, error) {
	staff, err := s.GetStaffByID(staffID)
	if err != nil {
		return "", "", errors.New("客服不存在")
	}

	// 生成TOTP密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "客服系统",
		AccountName: staff.Username,
	})
	if err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

// StaffEnable2FA 启用客服二步验证
func (s *SupportService) StaffEnable2FA(staffID uint, secret, code string) error {
	// 验证TOTP
	if !verifyTOTP(secret, code) {
		return errors.New("验证码错误")
	}

	// 保存密钥并启用2FA
	return s.repo.GetDB().Model(&model.SupportStaff{}).Where("id = ?", staffID).Updates(map[string]interface{}{
		"enable_2fa":  true,
		"totp_secret": secret,
	}).Error
}

// StaffDisable2FA 禁用客服二步验证
func (s *SupportService) StaffDisable2FA(staffID uint) error {
	return s.repo.GetDB().Model(&model.SupportStaff{}).Where("id = ?", staffID).Updates(map[string]interface{}{
		"enable_2fa":  false,
		"totp_secret": "",
	}).Error
}

// verifyTOTP 验证TOTP码
func verifyTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// StaffLogout 客服登出
func (s *SupportService) StaffLogout(sessionID string) error {
	var session model.SupportStaffSession
	if err := s.repo.GetDB().Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		return err
	}

	// 更新离线状态
	s.repo.GetDB().Model(&model.SupportStaff{}).Where("id = ?", session.StaffID).Update("status", 0)

	// 删除会话
	return s.repo.GetDB().Delete(&session).Error
}

// GetStaffSession 获取客服会话
func (s *SupportService) GetStaffSession(sessionID string) (*model.SupportStaffSession, error) {
	var session model.SupportStaffSession
	if err := s.repo.GetDB().Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// GetStaffByID 根据ID获取客服
func (s *SupportService) GetStaffByID(id uint) (*model.SupportStaff, error) {
	var staff model.SupportStaff
	if err := s.repo.GetDB().First(&staff, id).Error; err != nil {
		return nil, err
	}
	return &staff, nil
}

// GetAllStaff 获取所有客服
func (s *SupportService) GetAllStaff() ([]model.SupportStaff, error) {
	var staff []model.SupportStaff
	if err := s.repo.GetDB().Find(&staff).Error; err != nil {
		return nil, err
	}
	return staff, nil
}

// GetOnlineStaff 获取在线客服
func (s *SupportService) GetOnlineStaff() ([]model.SupportStaff, error) {
	var staff []model.SupportStaff
	if err := s.repo.GetDB().Where("status = 1").Find(&staff).Error; err != nil {
		return nil, err
	}
	return staff, nil
}

// UpdateStaff 更新客服信息
func (s *SupportService) UpdateStaff(id uint, nickname, email string, maxTickets int, status int) error {
	return s.repo.GetDB().Model(&model.SupportStaff{}).Where("id = ?", id).Updates(map[string]interface{}{
		"nickname":    nickname,
		"email":       email,
		"max_tickets": maxTickets,
		"status":      status,
	}).Error
}

// UpdateStaffPassword 更新客服密码
func (s *SupportService) UpdateStaffPassword(id uint, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.GetDB().Model(&model.SupportStaff{}).Where("id = ?", id).Update("password_hash", string(hash)).Error
}

// DeleteStaff 删除客服
func (s *SupportService) DeleteStaff(id uint) error {
	return s.repo.GetDB().Delete(&model.SupportStaff{}, id).Error
}

// UpdateStaffLoad 更新客服工单负载
func (s *SupportService) UpdateStaffLoad(staffID uint) {
	var count int64
	s.repo.GetDB().Model(&model.SupportTicket{}).
		Where("assigned_to = ? AND status IN (?, ?)", staffID, model.TicketStatusProcessing, model.TicketStatusReplied).
		Count(&count)

	s.repo.GetDB().Model(&model.SupportStaff{}).Where("id = ?", staffID).
		Update("current_load", count)
}

// CleanupExpiredStaffSessions 清理过期的客服会话
func (s *SupportService) CleanupExpiredStaffSessions() {
	s.repo.GetDB().Where("expires_at < ?", time.Now()).Delete(&model.SupportStaffSession{})
}
