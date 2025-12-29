package service

import (
	"errors"
	"fmt"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// RenewalService 续费服务
type RenewalService struct {
	repo     *repository.Repository
	emailSvc *EmailService
}

// NewRenewalService 创建续费服务
func NewRenewalService(repo *repository.Repository, emailSvc *EmailService) *RenewalService {
	return &RenewalService{
		repo:     repo,
		emailSvc: emailSvc,
	}
}

// UserKamiInfo 用户卡密信息（用于续费展示）
type UserKamiInfo struct {
	OrderID      uint      `json:"order_id"`
	OrderNo      string    `json:"order_no"`
	ProductID    uint      `json:"product_id"`
	ProductName  string    `json:"product_name"`
	KamiCode     string    `json:"kami_code"`
	Duration     int       `json:"duration"`
	DurationUnit string    `json:"duration_unit"`
	PurchaseTime time.Time `json:"purchase_time"`
	ExpireTime   time.Time `json:"expire_time"`
	DaysLeft     int       `json:"days_left"`     // 剩余天数
	IsExpired    bool      `json:"is_expired"`    // 是否已过期
	CanRenew     bool      `json:"can_renew"`     // 是否可续费
}

// GetUserKamis 获取用户的卡密列表（用于续费页面）
func (s *RenewalService) GetUserKamis(userID uint) ([]UserKamiInfo, error) {
	// 获取用户已完成的订单（有卡密的）
	var orders []model.Order
	err := s.repo.GetDB().Where("user_id = ? AND status = ? AND kami_code != ''", userID, 2).
		Order("created_at DESC").Find(&orders).Error
	if err != nil {
		return nil, err
	}

	var kamis []UserKamiInfo
	now := time.Now()

	for _, order := range orders {
		// 计算过期时间
		expireTime := s.calculateExpireTime(order.PaymentTime, order.Duration, order.DurationUnit)
		daysLeft := int(expireTime.Sub(now).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}

		// 检查商品是否仍然存在且可购买
		canRenew := false
		if product, err := s.repo.GetProductByID(order.ProductID); err == nil && product.Status == 1 {
			canRenew = true
		}

		kamis = append(kamis, UserKamiInfo{
			OrderID:      order.ID,
			OrderNo:      order.OrderNo,
			ProductID:    order.ProductID,
			ProductName:  order.ProductName,
			KamiCode:     order.KamiCode,
			Duration:     order.Duration,
			DurationUnit: order.DurationUnit,
			PurchaseTime: *order.PaymentTime,
			ExpireTime:   expireTime,
			DaysLeft:     daysLeft,
			IsExpired:    now.After(expireTime),
			CanRenew:     canRenew,
		})
	}

	return kamis, nil
}

// GetExpiringKamis 获取即将过期的卡密（用于提醒）
// daysBeforeExpire: 过期前多少天内的卡密
func (s *RenewalService) GetExpiringKamis(daysBeforeExpire int) ([]UserKamiInfo, error) {
	// 获取所有已完成的订单
	var orders []model.Order
	err := s.repo.GetDB().Where("status = ? AND kami_code != ''", 2).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	var expiringKamis []UserKamiInfo
	now := time.Now()
	deadline := now.AddDate(0, 0, daysBeforeExpire)

	for _, order := range orders {
		if order.PaymentTime == nil {
			continue
		}

		expireTime := s.calculateExpireTime(order.PaymentTime, order.Duration, order.DurationUnit)

		// 检查是否在指定天数内过期
		if expireTime.After(now) && expireTime.Before(deadline) {
			daysLeft := int(expireTime.Sub(now).Hours() / 24)

			expiringKamis = append(expiringKamis, UserKamiInfo{
				OrderID:      order.ID,
				OrderNo:      order.OrderNo,
				ProductID:    order.ProductID,
				ProductName:  order.ProductName,
				KamiCode:     order.KamiCode,
				Duration:     order.Duration,
				DurationUnit: order.DurationUnit,
				PurchaseTime: *order.PaymentTime,
				ExpireTime:   expireTime,
				DaysLeft:     daysLeft,
				IsExpired:    false,
				CanRenew:     true,
			})
		}
	}

	return expiringKamis, nil
}

// calculateExpireTime 计算过期时间
func (s *RenewalService) calculateExpireTime(paymentTime *time.Time, duration int, durationUnit string) time.Time {
	if paymentTime == nil {
		return time.Now()
	}

	switch durationUnit {
	case "天":
		return paymentTime.AddDate(0, 0, duration)
	case "周":
		return paymentTime.AddDate(0, 0, duration*7)
	case "月":
		return paymentTime.AddDate(0, duration, 0)
	case "年":
		return paymentTime.AddDate(duration, 0, 0)
	default:
		return paymentTime.AddDate(0, 0, duration)
	}
}

// SendRenewalReminder 发送续费提醒邮件
func (s *RenewalService) SendRenewalReminder(userID uint, orderNo string, remindType string) error {
	// 获取用户信息
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if user.Email == "" || !user.EmailVerified {
		return errors.New("用户邮箱未验证")
	}

	// 获取订单信息
	order, err := s.repo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return errors.New("订单不存在")
	}

	if order.PaymentTime == nil {
		return errors.New("订单未支付")
	}

	// 计算过期时间
	expireTime := s.calculateExpireTime(order.PaymentTime, order.Duration, order.DurationUnit)
	daysLeft := int(expireTime.Sub(time.Now()).Hours() / 24)

	// 检查是否已发送过该类型的提醒
	var existingReminder model.RenewalReminder
	err = s.repo.GetDB().Where("order_no = ? AND remind_type = ?", orderNo, remindType).First(&existingReminder).Error
	if err == nil {
		return errors.New("已发送过该类型的提醒")
	}

	// 发送邮件
	if s.emailSvc == nil {
		return errors.New("邮箱服务未初始化")
	}

	subject := s.getReminderSubject(remindType, order.ProductName)
	body := s.getReminderBody(remindType, user.Username, order.ProductName, order.KamiCode, expireTime, daysLeft)

	if err := s.emailSvc.SendEmail(user.Email, subject, body); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	// 记录提醒
	reminder := &model.RenewalReminder{
		UserID:     userID,
		OrderID:    order.ID,
		OrderNo:    orderNo,
		KamiCode:   order.KamiCode,
		ExpireAt:   expireTime,
		RemindAt:   time.Now(),
		RemindType: remindType,
	}
	s.repo.GetDB().Create(reminder)

	return nil
}

// getReminderSubject 获取提醒邮件主题
func (s *RenewalService) getReminderSubject(remindType, productName string) string {
	switch remindType {
	case model.RemindType7Day:
		return fmt.Sprintf("【续费提醒】您的 %s 将在7天后到期", productName)
	case model.RemindType3Day:
		return fmt.Sprintf("【续费提醒】您的 %s 将在3天后到期", productName)
	case model.RemindType1Day:
		return fmt.Sprintf("【紧急提醒】您的 %s 明天即将到期", productName)
	case model.RemindTypeExpired:
		return fmt.Sprintf("【到期通知】您的 %s 已到期", productName)
	default:
		return fmt.Sprintf("【续费提醒】您的 %s 即将到期", productName)
	}
}

// getReminderBody 获取提醒邮件内容
func (s *RenewalService) getReminderBody(remindType, username, productName, kamiCode string, expireTime time.Time, daysLeft int) string {
	var urgency string
	switch remindType {
	case model.RemindType7Day:
		urgency = "将在7天后"
	case model.RemindType3Day:
		urgency = "将在3天后"
	case model.RemindType1Day:
		urgency = "明天即将"
	case model.RemindTypeExpired:
		urgency = "已"
	default:
		urgency = "即将"
	}

	return fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h2 style="color: #333;">尊敬的 %s：</h2>
    <p style="color: #666; line-height: 1.6;">
        您购买的 <strong>%s</strong> %s到期，请及时续费以确保服务不中断。
    </p>
    <div style="background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0;">
        <p style="margin: 5px 0;"><strong>商品名称：</strong>%s</p>
        <p style="margin: 5px 0;"><strong>卡密：</strong>%s</p>
        <p style="margin: 5px 0;"><strong>到期时间：</strong>%s</p>
        <p style="margin: 5px 0;"><strong>剩余天数：</strong>%d 天</p>
    </div>
    <p style="color: #666; line-height: 1.6;">
        为了避免服务中断，建议您尽快登录系统进行续费。
    </p>
    <p style="color: #999; font-size: 12px; margin-top: 30px;">
        此邮件由系统自动发送，请勿直接回复。
    </p>
</div>
`, username, productName, urgency, productName, kamiCode, expireTime.Format("2006-01-02 15:04:05"), daysLeft)
}

// CheckAndSendReminders 检查并发送续费提醒（定时任务调用）
func (s *RenewalService) CheckAndSendReminders() {
	// 获取所有已完成的订单
	var orders []model.Order
	err := s.repo.GetDB().Where("status = ? AND kami_code != ''", 2).Find(&orders).Error
	if err != nil {
		return
	}

	now := time.Now()

	for _, order := range orders {
		if order.PaymentTime == nil {
			continue
		}

		// 获取用户信息
		user, err := s.repo.GetUserByID(order.UserID)
		if err != nil || user.Email == "" || !user.EmailVerified {
			continue
		}

		expireTime := s.calculateExpireTime(order.PaymentTime, order.Duration, order.DurationUnit)
		daysLeft := int(expireTime.Sub(now).Hours() / 24)

		// 根据剩余天数发送不同类型的提醒
		var remindType string
		if daysLeft <= 0 {
			remindType = model.RemindTypeExpired
		} else if daysLeft <= 1 {
			remindType = model.RemindType1Day
		} else if daysLeft <= 3 {
			remindType = model.RemindType3Day
		} else if daysLeft <= 7 {
			remindType = model.RemindType7Day
		} else {
			continue // 不需要提醒
		}

		// 检查是否已发送过该类型的提醒
		var existingReminder model.RenewalReminder
		err = s.repo.GetDB().Where("order_no = ? AND remind_type = ?", order.OrderNo, remindType).First(&existingReminder).Error
		if err == nil {
			continue // 已发送过
		}

		// 发送提醒
		s.SendRenewalReminder(order.UserID, order.OrderNo, remindType)
	}
}

// GetRenewalHistory 获取用户的续费提醒历史
func (s *RenewalService) GetRenewalHistory(userID uint) ([]model.RenewalReminder, error) {
	var reminders []model.RenewalReminder
	err := s.repo.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&reminders).Error
	return reminders, err
}
