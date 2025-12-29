package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
	"user-frontend/internal/utils"

	"github.com/pquerna/otp/totp"
)

// SensitiveService 敏感操作验证服务
type SensitiveService struct {
	repo     *repository.Repository
	emailSvc *EmailService
}

// NewSensitiveService 创建敏感操作验证服务
func NewSensitiveService(repo *repository.Repository, emailSvc *EmailService) *SensitiveService {
	return &SensitiveService{
		repo:     repo,
		emailSvc: emailSvc,
	}
}

// RequestVerification 请求敏感操作验证
// 返回验证令牌，用于后续验证
func (s *SensitiveService) RequestVerification(userID uint, operationType string) (string, error) {
	// 生成随机令牌
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", errors.New("生成令牌失败")
	}
	token := hex.EncodeToString(tokenBytes)

	// 创建验证令牌记录
	opToken := &model.SensitiveOperationToken{
		UserID:        userID,
		Token:         token,
		OperationType: operationType,
		Verified:      false,
		ExpiresAt:     time.Now().Add(10 * time.Minute), // 10分钟有效期
	}

	if err := s.repo.GetDB().Create(opToken).Error; err != nil {
		return "", errors.New("创建验证令牌失败")
	}

	return token, nil
}

// VerifyWithTOTP 使用TOTP验证敏感操作
func (s *SensitiveService) VerifyWithTOTP(userID uint, token, totpCode string) error {
	// 获取令牌记录
	var opToken model.SensitiveOperationToken
	if err := s.repo.GetDB().Where("token = ? AND user_id = ?", token, userID).First(&opToken).Error; err != nil {
		return errors.New("无效的验证令牌")
	}

	// 检查是否过期
	if time.Now().After(opToken.ExpiresAt) {
		s.repo.GetDB().Delete(&opToken)
		return errors.New("验证令牌已过期")
	}

	// 获取用户信息
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证TOTP
	if user.TOTPSecret == "" {
		return errors.New("未设置动态口令")
	}

	if !totp.Validate(totpCode, user.TOTPSecret) {
		return errors.New("动态口令错误")
	}

	// 标记为已验证
	opToken.Verified = true
	if err := s.repo.GetDB().Save(&opToken).Error; err != nil {
		return errors.New("更新验证状态失败")
	}

	return nil
}

// VerifyWithEmail 使用邮箱验证码验证敏感操作
func (s *SensitiveService) VerifyWithEmail(userID uint, token, emailCode string) error {
	// 获取令牌记录
	var opToken model.SensitiveOperationToken
	if err := s.repo.GetDB().Where("token = ? AND user_id = ?", token, userID).First(&opToken).Error; err != nil {
		return errors.New("无效的验证令牌")
	}

	// 检查是否过期
	if time.Now().After(opToken.ExpiresAt) {
		s.repo.GetDB().Delete(&opToken)
		return errors.New("验证令牌已过期")
	}

	// 获取用户信息
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证邮箱验证码
	if s.emailSvc == nil {
		return errors.New("邮箱服务未初始化")
	}

	// 根据操作类型使用不同的验证码类型
	codeType := "sensitive_" + opToken.OperationType
	if !s.emailSvc.VerifyCode(user.Email, emailCode, codeType) {
		return errors.New("邮箱验证码错误或已过期")
	}

	// 标记为已验证
	opToken.Verified = true
	if err := s.repo.GetDB().Save(&opToken).Error; err != nil {
		return errors.New("更新验证状态失败")
	}

	return nil
}

// VerifyWithPassword 使用密码验证敏感操作
func (s *SensitiveService) VerifyWithPassword(userID uint, token, password string) error {
	// 获取令牌记录
	var opToken model.SensitiveOperationToken
	if err := s.repo.GetDB().Where("token = ? AND user_id = ?", token, userID).First(&opToken).Error; err != nil {
		return errors.New("无效的验证令牌")
	}

	// 检查是否过期
	if time.Now().After(opToken.ExpiresAt) {
		s.repo.GetDB().Delete(&opToken)
		return errors.New("验证令牌已过期")
	}

	// 获取用户信息
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证密码
	if !checkPassword(password, user.PasswordHash) {
		return errors.New("密码错误")
	}

	// 标记为已验证
	opToken.Verified = true
	if err := s.repo.GetDB().Save(&opToken).Error; err != nil {
		return errors.New("更新验证状态失败")
	}

	return nil
}

// CheckVerified 检查敏感操作是否已验证
func (s *SensitiveService) CheckVerified(userID uint, token, operationType string) (bool, error) {
	var opToken model.SensitiveOperationToken
	if err := s.repo.GetDB().Where("token = ? AND user_id = ? AND operation_type = ?", token, userID, operationType).First(&opToken).Error; err != nil {
		return false, errors.New("无效的验证令牌")
	}

	// 检查是否过期
	if time.Now().After(opToken.ExpiresAt) {
		s.repo.GetDB().Delete(&opToken)
		return false, errors.New("验证令牌已过期")
	}

	return opToken.Verified, nil
}

// ConsumeToken 消费验证令牌（执行敏感操作后调用）
func (s *SensitiveService) ConsumeToken(userID uint, token string) error {
	result := s.repo.GetDB().Where("token = ? AND user_id = ? AND verified = ?", token, userID, true).Delete(&model.SensitiveOperationToken{})
	if result.RowsAffected == 0 {
		return errors.New("无效或未验证的令牌")
	}
	return nil
}

// CleanupExpiredTokens 清理过期的验证令牌
func (s *SensitiveService) CleanupExpiredTokens() {
	s.repo.GetDB().Where("expires_at < ?", time.Now()).Delete(&model.SensitiveOperationToken{})
}

// GetUserVerificationMethods 获取用户可用的验证方式
func (s *SensitiveService) GetUserVerificationMethods(userID uint) (map[string]bool, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	methods := map[string]bool{
		"password": true,                                  // 密码验证始终可用
		"email":    user.EmailVerified && user.Email != "", // 邮箱验证需要已验证邮箱
		"totp":     user.TOTPSecret != "",                 // TOTP验证需要已设置
	}

	return methods, nil
}

// SendVerificationEmail 发送敏感操作验证邮件
func (s *SensitiveService) SendVerificationEmail(userID uint, operationType string) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if !user.EmailVerified || user.Email == "" {
		return errors.New("邮箱未验证")
	}

	if s.emailSvc == nil {
		return errors.New("邮箱服务未初始化")
	}

	// 发送验证码
	codeType := "sensitive_" + operationType
	return s.emailSvc.SendVerifyCode(user.Email, codeType)
}

// checkPassword 验证密码（内部函数）
func checkPassword(password, hash string) bool {
	return utils.CheckPassword(password, hash)
}
