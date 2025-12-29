package service

import (
	"errors"
	"regexp"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// PayPasswordService 支付密码服务
type PayPasswordService struct {
	repo *repository.Repository
}

// NewPayPasswordService 创建支付密码服务
func NewPayPasswordService(repo *repository.Repository) *PayPasswordService {
	return &PayPasswordService{repo: repo}
}

// 支付密码配置常量
const (
	PayPasswordMaxErrors = 5               // 最大错误次数
	PayPasswordLockTime  = 30 * time.Minute // 锁定时间
)

// SetPayPassword 设置支付密码
// 参数：
//   - userID: 用户ID
//   - password: 6位数字支付密码
//   - loginPassword: 登录密码（用于验证不能与支付密码相同）
func (s *PayPasswordService) SetPayPassword(userID uint, password, loginPassword string) error {
	// 验证支付密码格式（6位纯数字）
	if !s.validatePayPasswordFormat(password) {
		return errors.New("支付密码必须为6位纯数字")
	}

	// 获取用户信息
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 检查是否已设置支付密码
	if user.PayPasswordSet {
		return errors.New("支付密码已设置，请使用修改功能")
	}

	// 验证支付密码不能与登录密码相同
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil {
		return errors.New("支付密码不能与登录密码相同")
	}

	// 加密支付密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新用户支付密码
	user.PayPassword = string(hashedPassword)
	user.PayPasswordSet = true
	user.PayPasswordErrors = 0
	user.PayPasswordLockAt = nil

	return s.repo.UpdateUser(user)
}

// UpdatePayPassword 修改支付密码
// 参数：
//   - userID: 用户ID
//   - oldPassword: 旧支付密码
//   - newPassword: 新支付密码
func (s *PayPasswordService) UpdatePayPassword(userID uint, oldPassword, newPassword string) error {
	// 验证新密码格式
	if !s.validatePayPasswordFormat(newPassword) {
		return errors.New("新支付密码必须为6位纯数字")
	}

	// 获取用户信息
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 检查是否已设置支付密码
	if !user.PayPasswordSet {
		return errors.New("请先设置支付密码")
	}

	// 检查是否被锁定
	if s.isPayPasswordLocked(user) {
		return errors.New("支付密码已锁定，请30分钟后再试")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PayPassword), []byte(oldPassword)); err != nil {
		// 记录错误次数
		s.recordPayPasswordError(user)
		return errors.New("原支付密码错误")
	}

	// 验证新旧密码不能相同
	if oldPassword == newPassword {
		return errors.New("新支付密码不能与原密码相同")
	}

	// 验证新密码不能与登录密码相同
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newPassword)) == nil {
		return errors.New("支付密码不能与登录密码相同")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新支付密码
	user.PayPassword = string(hashedPassword)
	user.PayPasswordErrors = 0
	user.PayPasswordLockAt = nil

	return s.repo.UpdateUser(user)
}

// ResetPayPassword 重置支付密码（通过邮箱验证后调用）
// 参数：
//   - userID: 用户ID
//   - newPassword: 新支付密码
func (s *PayPasswordService) ResetPayPassword(userID uint, newPassword string) error {
	// 验证新密码格式
	if !s.validatePayPasswordFormat(newPassword) {
		return errors.New("支付密码必须为6位纯数字")
	}

	// 获取用户信息
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证新密码不能与登录密码相同
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newPassword)) == nil {
		return errors.New("支付密码不能与登录密码相同")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新支付密码
	user.PayPassword = string(hashedPassword)
	user.PayPasswordSet = true
	user.PayPasswordErrors = 0
	user.PayPasswordLockAt = nil

	return s.repo.UpdateUser(user)
}

// VerifyPayPassword 验证支付密码
// 参数：
//   - userID: 用户ID
//   - password: 支付密码
// 返回：
//   - error: 验证失败的错误信息
func (s *PayPasswordService) VerifyPayPassword(userID uint, password string) error {
	// 获取用户信息
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 检查是否已设置支付密码
	if !user.PayPasswordSet {
		return errors.New("请先设置支付密码")
	}

	// 检查是否被锁定
	if s.isPayPasswordLocked(user) {
		return errors.New("支付密码已锁定，请30分钟后再试")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PayPassword), []byte(password)); err != nil {
		// 记录错误次数
		s.recordPayPasswordError(user)
		remainingAttempts := PayPasswordMaxErrors - user.PayPasswordErrors - 1
		if remainingAttempts <= 0 {
			return errors.New("支付密码错误次数过多，已锁定30分钟")
		}
		return errors.New("支付密码错误")
	}

	// 验证成功，重置错误次数
	if user.PayPasswordErrors > 0 {
		user.PayPasswordErrors = 0
		user.PayPasswordLockAt = nil
		s.repo.UpdateUser(user)
	}

	return nil
}

// GetPayPasswordStatus 获取支付密码状态
// 返回：
//   - isSet: 是否已设置
//   - isLocked: 是否被锁定
//   - lockRemainingSeconds: 锁定剩余秒数
func (s *PayPasswordService) GetPayPasswordStatus(userID uint) (isSet bool, isLocked bool, lockRemainingSeconds int, err error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return false, false, 0, errors.New("用户不存在")
	}

	isSet = user.PayPasswordSet
	isLocked = s.isPayPasswordLocked(user)

	if isLocked && user.PayPasswordLockAt != nil {
		unlockTime := user.PayPasswordLockAt.Add(PayPasswordLockTime)
		lockRemainingSeconds = int(unlockTime.Sub(time.Now()).Seconds())
		if lockRemainingSeconds < 0 {
			lockRemainingSeconds = 0
		}
	}

	return isSet, isLocked, lockRemainingSeconds, nil
}

// CheckPayPasswordRequired 检查是否需要支付密码
// 用于前端判断是否显示支付密码输入框
func (s *PayPasswordService) CheckPayPasswordRequired(userID uint) (required bool, isSet bool, err error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return false, false, errors.New("用户不存在")
	}

	// 必须设置支付密码才能使用余额功能
	return true, user.PayPasswordSet, nil
}

// validatePayPasswordFormat 验证支付密码格式（6位纯数字）
func (s *PayPasswordService) validatePayPasswordFormat(password string) bool {
	matched, _ := regexp.MatchString(`^\d{6}$`, password)
	return matched
}

// isPayPasswordLocked 检查支付密码是否被锁定
func (s *PayPasswordService) isPayPasswordLocked(user *model.User) bool {
	if user.PayPasswordLockAt == nil {
		return false
	}
	unlockTime := user.PayPasswordLockAt.Add(PayPasswordLockTime)
	return time.Now().Before(unlockTime)
}

// recordPayPasswordError 记录支付密码错误
func (s *PayPasswordService) recordPayPasswordError(user *model.User) {
	user.PayPasswordErrors++
	if user.PayPasswordErrors >= PayPasswordMaxErrors {
		now := time.Now()
		user.PayPasswordLockAt = &now
	}
	s.repo.UpdateUser(user)
}
