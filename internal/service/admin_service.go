package service

import (
	"errors"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
	"user-frontend/internal/utils"
)

type AdminService struct {
	repo *repository.Repository
}

func NewAdminService(repo *repository.Repository) *AdminService {
	return &AdminService{repo: repo}
}

// Login 管理员登录
func (s *AdminService) Login(username, password, clientIP string) (*model.AdminUser, error) {
	admin, err := s.repo.GetAdminByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if !utils.CheckPassword(password, admin.PasswordHash) {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新登录信息
	now := time.Now()
	admin.LastLoginAt = &now
	admin.LastLoginIP = clientIP
	s.repo.UpdateAdminUser(admin)

	return admin, nil
}

// CreateAdmin 创建管理员
func (s *AdminService) CreateAdmin(username, password, role string) (*model.AdminUser, error) {
	// 检查用户名是否存在
	if _, err := s.repo.GetAdminByUsername(username); err == nil {
		return nil, errors.New("用户名已存在")
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	admin := &model.AdminUser{
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
	}

	if err := s.repo.CreateAdminUser(admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// UpdatePassword 修改密码
func (s *AdminService) UpdatePassword(username, oldPassword, newPassword string) error {
	admin, err := s.repo.GetAdminByUsername(username)
	if err != nil {
		return errors.New("管理员不存在")
	}

	if !utils.CheckPassword(oldPassword, admin.PasswordHash) {
		return errors.New("原密码错误")
	}

	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	admin.PasswordHash = passwordHash
	return s.repo.UpdateAdminUser(admin)
}

// Enable2FA 启用两步验证
func (s *AdminService) Enable2FA(username, secret string) error {
	admin, err := s.repo.GetAdminByUsername(username)
	if err != nil {
		return errors.New("管理员不存在")
	}

	admin.Enable2FA = true
	admin.TOTPSecret = secret
	return s.repo.UpdateAdminUser(admin)
}

// Disable2FA 禁用两步验证
func (s *AdminService) Disable2FA(username string) error {
	admin, err := s.repo.GetAdminByUsername(username)
	if err != nil {
		return errors.New("管理员不存在")
	}

	admin.Enable2FA = false
	admin.TOTPSecret = ""
	return s.repo.UpdateAdminUser(admin)
}

// GetAdmin2FAStatus 获取管理员2FA状态
func (s *AdminService) GetAdmin2FAStatus(username string) (bool, string, error) {
	admin, err := s.repo.GetAdminByUsername(username)
	if err != nil {
		return false, "", err
	}
	return admin.Enable2FA, admin.TOTPSecret, nil
}

// GetAllAdmins 获取所有管理员
func (s *AdminService) GetAllAdmins() ([]model.AdminUser, error) {
	return s.repo.GetAllAdmins()
}

// InitDefaultAdmin 初始化默认管理员
func (s *AdminService) InitDefaultAdmin(username, password string) error {
	// 检查是否已存在管理员
	admins, err := s.repo.GetAllAdmins()
	if err == nil && len(admins) > 0 {
		return nil // 已存在管理员，不需要初始化
	}

	_, err = s.CreateAdmin(username, password, "super_admin")
	return err
}


// GetAdminByUsername 获取管理员
func (s *AdminService) GetAdminByUsername(username string) (*model.AdminUser, error) {
	return s.repo.GetAdminByUsername(username)
}
