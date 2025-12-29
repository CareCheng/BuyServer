package service

import (
	"errors"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// AccountDeletionService 账户注销服务
type AccountDeletionService struct {
	repo     *repository.Repository
	emailSvc *EmailService
}

// NewAccountDeletionService 创建账户注销服务实例
func NewAccountDeletionService(repo *repository.Repository, emailSvc *EmailService) *AccountDeletionService {
	return &AccountDeletionService{
		repo:     repo,
		emailSvc: emailSvc,
	}
}

// RequestDeletion 申请账户注销
// 参数：
//   - userID: 用户ID
//   - username: 用户名
//   - email: 邮箱
//   - reason: 注销原因
// 返回：
//   - 注销申请
//   - 错误信息
func (s *AccountDeletionService) RequestDeletion(userID uint, username, email, reason string) (*model.AccountDeletionRequest, error) {
	// 检查是否已有待处理的申请
	var existing model.AccountDeletionRequest
	if err := s.repo.GetDB().Where("user_id = ? AND status = ?", userID, model.DeletionStatusPending).First(&existing).Error; err == nil {
		return nil, errors.New("您已有待处理的注销申请")
	}

	// 创建注销申请
	request := &model.AccountDeletionRequest{
		UserID:   userID,
		Username: username,
		Email:    email,
		Reason:   reason,
		Status:   model.DeletionStatusPending,
	}

	if err := s.repo.GetDB().Create(request).Error; err != nil {
		return nil, err
	}

	// 发送确认邮件
	if s.emailSvc != nil && email != "" {
		go s.emailSvc.SendAccountDeletionRequestEmail(email, username)
	}

	return request, nil
}

// CancelDeletion 取消注销申请
// 参数：
//   - userID: 用户ID
// 返回：
//   - 错误信息
func (s *AccountDeletionService) CancelDeletion(userID uint) error {
	result := s.repo.GetDB().Model(&model.AccountDeletionRequest{}).
		Where("user_id = ? AND status = ?", userID, model.DeletionStatusPending).
		Update("status", model.DeletionStatusCancelled)

	if result.RowsAffected == 0 {
		return errors.New("没有待处理的注销申请")
	}

	return result.Error
}

// GetUserDeletionRequest 获取用户的注销申请
// 参数：
//   - userID: 用户ID
// 返回：
//   - 注销申请
//   - 错误信息
func (s *AccountDeletionService) GetUserDeletionRequest(userID uint) (*model.AccountDeletionRequest, error) {
	var request model.AccountDeletionRequest
	if err := s.repo.GetDB().Where("user_id = ? AND status IN ?", userID, []int{model.DeletionStatusPending, model.DeletionStatusApproved}).
		Order("created_at DESC").First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &request, nil
}

// ApproveDeletion 批准注销申请（管理员功能）
// 参数：
//   - requestID: 申请ID
//   - adminUsername: 管理员用户名
// 返回：
//   - 错误信息
func (s *AccountDeletionService) ApproveDeletion(requestID uint, adminUsername string) error {
	var request model.AccountDeletionRequest
	if err := s.repo.GetDB().First(&request, requestID).Error; err != nil {
		return errors.New("申请不存在")
	}

	if request.Status != model.DeletionStatusPending {
		return errors.New("该申请已处理")
	}

	// 设置计划删除时间（7天后）
	scheduledAt := time.Now().AddDate(0, 0, 7)
	now := time.Now()

	if err := s.repo.GetDB().Model(&request).Updates(map[string]interface{}{
		"status":       model.DeletionStatusApproved,
		"processed_by": adminUsername,
		"processed_at": &now,
		"scheduled_at": &scheduledAt,
	}).Error; err != nil {
		return err
	}

	// 发送通知邮件
	if s.emailSvc != nil && request.Email != "" {
		go s.emailSvc.SendAccountDeletionApprovedEmail(request.Email, request.Username, scheduledAt)
	}

	return nil
}

// RejectDeletion 拒绝注销申请（管理员功能）
// 参数：
//   - requestID: 申请ID
//   - adminUsername: 管理员用户名
//   - reason: 拒绝原因
// 返回：
//   - 错误信息
func (s *AccountDeletionService) RejectDeletion(requestID uint, adminUsername, reason string) error {
	var request model.AccountDeletionRequest
	if err := s.repo.GetDB().First(&request, requestID).Error; err != nil {
		return errors.New("申请不存在")
	}

	if request.Status != model.DeletionStatusPending {
		return errors.New("该申请已处理")
	}

	now := time.Now()
	if err := s.repo.GetDB().Model(&request).Updates(map[string]interface{}{
		"status":        model.DeletionStatusRejected,
		"reject_reason": reason,
		"processed_by":  adminUsername,
		"processed_at":  &now,
	}).Error; err != nil {
		return err
	}

	// 发送通知邮件
	if s.emailSvc != nil && request.Email != "" {
		go s.emailSvc.SendAccountDeletionRejectedEmail(request.Email, request.Username, reason)
	}

	return nil
}

// GetPendingRequests 获取待处理的注销申请（管理员功能）
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 申请列表
//   - 总数
//   - 错误信息
func (s *AccountDeletionService) GetPendingRequests(page, pageSize int) ([]model.AccountDeletionRequest, int64, error) {
	var requests []model.AccountDeletionRequest
	var total int64

	query := s.repo.GetDB().Model(&model.AccountDeletionRequest{}).Where("status = ?", model.DeletionStatusPending)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&requests).Error; err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

// GetAllRequests 获取所有注销申请（管理员功能）
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//   - status: 状态筛选（-1表示全部）
// 返回：
//   - 申请列表
//   - 总数
//   - 错误信息
func (s *AccountDeletionService) GetAllRequests(page, pageSize, status int) ([]model.AccountDeletionRequest, int64, error) {
	var requests []model.AccountDeletionRequest
	var total int64

	query := s.repo.GetDB().Model(&model.AccountDeletionRequest{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&requests).Error; err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

// ExecuteScheduledDeletions 执行计划中的账户删除
// 返回：
//   - 删除的账户数量
//   - 错误信息
func (s *AccountDeletionService) ExecuteScheduledDeletions() (int, error) {
	var requests []model.AccountDeletionRequest
	now := time.Now()

	// 查找已到期的批准申请
	if err := s.repo.GetDB().Where("status = ? AND scheduled_at <= ?", model.DeletionStatusApproved, now).
		Find(&requests).Error; err != nil {
		return 0, err
	}

	deletedCount := 0
	for _, request := range requests {
		// 执行账户删除
		if err := s.deleteUserAccount(request.UserID); err != nil {
			continue
		}

		// 更新申请状态
		s.repo.GetDB().Model(&request).Update("status", model.DeletionStatusCompleted)

		// 发送删除完成邮件
		if s.emailSvc != nil && request.Email != "" {
			go s.emailSvc.SendAccountDeletionCompletedEmail(request.Email, request.Username)
		}

		deletedCount++
	}

	return deletedCount, nil
}

// deleteUserAccount 删除用户账户及相关数据
func (s *AccountDeletionService) deleteUserAccount(userID uint) error {
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		// 删除用户会话
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserSession{}).Error; err != nil {
			return err
		}

		// 删除登录设备记录
		if err := tx.Where("user_id = ?", userID).Delete(&model.LoginDevice{}).Error; err != nil {
			return err
		}

		// 删除登录历史
		if err := tx.Where("user_id = ?", userID).Delete(&model.LoginHistory{}).Error; err != nil {
			return err
		}

		// 匿名化订单数据（保留订单记录但移除用户关联）
		if err := tx.Model(&model.Order{}).Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"username": "[已注销用户]",
			}).Error; err != nil {
			return err
		}

		// 匿名化评价数据
		if err := tx.Model(&model.ProductReview{}).Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"username": "[已注销用户]",
				"is_anon":  true,
			}).Error; err != nil {
			return err
		}

		// 软删除用户
		if err := tx.Delete(&model.User{}, userID).Error; err != nil {
			return err
		}

		return nil
	})
}
