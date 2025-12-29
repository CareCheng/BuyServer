package service

import (
	"errors"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
	"user-frontend/internal/utils"
)

type UserService struct {
	repo *repository.Repository
}

func NewUserService(repo *repository.Repository) *UserService {
	return &UserService{repo: repo}
}

// Register 用户注册
func (s *UserService) Register(username, email, password, phone string) (*model.User, error) {
	// 检查用户名是否存在
	if _, err := s.repo.GetUserByUsername(username); err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	if email != "" {
		if _, err := s.repo.GetUserByEmail(email); err == nil {
			return nil, errors.New("邮箱已被注册")
		}
	}

	// 密码加密
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Phone:        phone,
		Status:       1,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(username, password, clientIP string) (*model.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新登录信息
	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = clientIP
	s.repo.UpdateUser(user)

	return user, nil
}

// GetUserByID 获取用户信息
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.repo.GetUserByID(id)
}

// UpdatePassword 修改密码
func (s *UserService) UpdatePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if !utils.CheckPassword(oldPassword, user.PasswordHash) {
		return errors.New("原密码错误")
	}

	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	user.PasswordHash = passwordHash
	return s.repo.UpdateUser(user)
}

// Enable2FA 启用两步验证
func (s *UserService) Enable2FA(userID uint, secret string) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.Enable2FA = true
	user.TOTPSecret = secret
	return s.repo.UpdateUser(user)
}

// Disable2FA 禁用两步验证
func (s *UserService) Disable2FA(userID uint) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.Enable2FA = false
	user.TOTPSecret = ""
	return s.repo.UpdateUser(user)
}

// GetUser2FAStatus 获取用户2FA状态
func (s *UserService) GetUser2FAStatus(userID uint) (bool, string, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return false, "", err
	}
	return user.Enable2FA, user.TOTPSecret, nil
}


// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers(page, pageSize int) ([]model.User, int64, error) {
	return s.repo.GetAllUsers(page, pageSize)
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(id uint, status int) error {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return err
	}
	user.Status = status
	return s.repo.UpdateUser(user)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *model.User) error {
	return s.repo.UpdateUser(user)
}

// SetPreferEmailAuth 设置登录验证方式偏好
func (s *UserService) SetPreferEmailAuth(userID uint, preferEmail bool) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	user.PreferEmailAuth = preferEmail
	return s.repo.UpdateUser(user)
}

// BindEmail 绑定邮箱
func (s *UserService) BindEmail(userID uint, email string) error {
	// 检查邮箱是否已被使用
	existingUser, err := s.repo.GetUserByEmail(email)
	if err == nil && existingUser.ID != userID {
		return errors.New("该邮箱已被其他用户绑定")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.Email = email
	user.EmailVerified = true
	return s.repo.UpdateUser(user)
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	return s.repo.GetUserByEmail(email)
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	return s.repo.GetUserByUsername(username)
}

// RegisterWithVerifiedEmail 使用已验证邮箱注册
func (s *UserService) RegisterWithVerifiedEmail(username, email, password, phone string) (*model.User, error) {
	// 检查用户名是否存在
	if _, err := s.repo.GetUserByUsername(username); err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	if email != "" {
		if _, err := s.repo.GetUserByEmail(email); err == nil {
			return nil, errors.New("邮箱已被注册")
		}
	}

	// 密码加密
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &model.User{
		Username:      username,
		Email:         email,
		EmailVerified: true, // 邮箱已验证
		PasswordHash:  passwordHash,
		Phone:         phone,
		Status:        1,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangeEmail 更换邮箱
func (s *UserService) ChangeEmail(userID uint, newEmail string) error {
	// 检查邮箱是否已被使用
	existingUser, err := s.repo.GetUserByEmail(newEmail)
	if err == nil && existingUser.ID != userID {
		return errors.New("该邮箱已被其他用户绑定")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.Email = newEmail
	user.EmailVerified = true
	return s.repo.UpdateUser(user)
}
