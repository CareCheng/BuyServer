package service

import (
	"fmt"
	"time"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// NotificationService 通知服务
// 处理邮箱订阅通知、订单状态变更通知等
type NotificationService struct {
	repo     *repository.Repository
	emailSvc *EmailService
}

// NewNotificationService 创建通知服务
func NewNotificationService(repo *repository.Repository, emailSvc *EmailService) *NotificationService {
	return &NotificationService{
		repo:     repo,
		emailSvc: emailSvc,
	}
}

// SetEmailService 设置邮件服务引用
func (s *NotificationService) SetEmailService(emailSvc *EmailService) {
	s.emailSvc = emailSvc
}

// ==================== 订单状态通知 ====================

// NotifyOrderCreated 通知订单创建
// 参数：
//   - order: 订单信息
//   - email: 用户邮箱
func (s *NotificationService) NotifyOrderCreated(order *model.Order, email string) error {
	if s.emailSvc == nil || email == "" {
		return nil
	}

	subject := fmt.Sprintf("[%s] 订单创建成功 - %s", config.GlobalConfig.ServerConfig.SystemTitle, order.OrderNo)
	body := s.buildOrderCreatedEmail(order)

	return s.emailSvc.SendEmail(email, subject, body)
}

// NotifyOrderPaid 通知订单支付成功
// 参数：
//   - order: 订单信息
//   - email: 用户邮箱
//   - kamiCode: 卡密（如果有）
func (s *NotificationService) NotifyOrderPaid(order *model.Order, email, kamiCode string) error {
	if s.emailSvc == nil || email == "" {
		return nil
	}

	subject := fmt.Sprintf("[%s] 订单支付成功 - %s", config.GlobalConfig.ServerConfig.SystemTitle, order.OrderNo)
	body := s.buildOrderPaidEmail(order, kamiCode)

	return s.emailSvc.SendEmail(email, subject, body)
}

// NotifyOrderCancelled 通知订单取消
// 参数：
//   - order: 订单信息
//   - email: 用户邮箱
//   - reason: 取消原因
func (s *NotificationService) NotifyOrderCancelled(order *model.Order, email, reason string) error {
	if s.emailSvc == nil || email == "" {
		return nil
	}

	subject := fmt.Sprintf("[%s] 订单已取消 - %s", config.GlobalConfig.ServerConfig.SystemTitle, order.OrderNo)
	body := s.buildOrderCancelledEmail(order, reason)

	return s.emailSvc.SendEmail(email, subject, body)
}

// NotifyOrderRefunded 通知订单退款
// 参数：
//   - order: 订单信息
//   - email: 用户邮箱
func (s *NotificationService) NotifyOrderRefunded(order *model.Order, email string) error {
	if s.emailSvc == nil || email == "" {
		return nil
	}

	subject := fmt.Sprintf("[%s] 订单已退款 - %s", config.GlobalConfig.ServerConfig.SystemTitle, order.OrderNo)
	body := s.buildOrderRefundedEmail(order)

	return s.emailSvc.SendEmail(email, subject, body)
}

// ==================== 卡密到期提醒 ====================

// NotifyKamiExpiring 通知卡密即将过期
// 参数：
//   - email: 用户邮箱
//   - productName: 商品名称
//   - kamiCode: 卡密
//   - expireTime: 过期时间
//   - daysLeft: 剩余天数
func (s *NotificationService) NotifyKamiExpiring(email, productName, kamiCode string, expireTime time.Time, daysLeft int) error {
	if s.emailSvc == nil || email == "" {
		return nil
	}

	subject := fmt.Sprintf("[%s] 卡密即将过期提醒", config.GlobalConfig.ServerConfig.SystemTitle)
	body := s.buildKamiExpiringEmail(productName, kamiCode, expireTime, daysLeft)

	return s.emailSvc.SendEmail(email, subject, body)
}

// ==================== 安全通知 ====================

// NotifyNewDeviceLogin 通知新设备登录
// 参数：
//   - email: 用户邮箱
//   - device: 设备信息
//   - ip: 登录IP
//   - location: IP归属地
//   - loginTime: 登录时间
func (s *NotificationService) NotifyNewDeviceLogin(email, device, ip, location string, loginTime time.Time) error {
	if s.emailSvc == nil || email == "" {
		return nil
	}

	subject := fmt.Sprintf("[%s] 新设备登录提醒", config.GlobalConfig.ServerConfig.SystemTitle)
	body := s.buildNewDeviceLoginEmail(device, ip, location, loginTime)

	return s.emailSvc.SendEmail(email, subject, body)
}

// NotifyPasswordChanged 通知密码已修改
// 参数：
//   - email: 用户邮箱
//   - ip: 操作IP
//   - changeTime: 修改时间
func (s *NotificationService) NotifyPasswordChanged(email, ip string, changeTime time.Time) error {
	if s.emailSvc == nil || email == "" {
		return nil
	}

	subject := fmt.Sprintf("[%s] 密码修改通知", config.GlobalConfig.ServerConfig.SystemTitle)
	body := s.buildPasswordChangedEmail(ip, changeTime)

	return s.emailSvc.SendEmail(email, subject, body)
}

// ==================== 邮件模板 ====================

func (s *NotificationService) buildOrderCreatedEmail(order *model.Order) string {
	systemTitle := config.GlobalConfig.ServerConfig.SystemTitle
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #667eea;">%s</h2>
        <p>您好，</p>
        <p>您的订单已创建成功，订单详情如下：</p>
        <div style="background: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 8px;">
            <p><strong>订单号：</strong>%s</p>
            <p><strong>商品名称：</strong>%s</p>
            <p><strong>时长：</strong>%d %s</p>
            <p><strong>订单金额：</strong>¥%.2f</p>
            <p><strong>创建时间：</strong>%s</p>
        </div>
        <p>请在30分钟内完成支付，超时订单将自动取消。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, systemTitle, order.OrderNo, order.ProductName, order.Duration, order.DurationUnit, order.Price, order.CreatedAt.Format("2006-01-02 15:04:05"))
}

func (s *NotificationService) buildOrderPaidEmail(order *model.Order, kamiCode string) string {
	systemTitle := config.GlobalConfig.ServerConfig.SystemTitle
	kamiSection := ""
	if kamiCode != "" {
		kamiSection = fmt.Sprintf(`
            <div style="background: #e8f5e9; padding: 15px; margin: 15px 0; border-radius: 8px; border-left: 4px solid #4caf50;">
                <p style="margin: 0;"><strong>您的卡密：</strong></p>
                <p style="font-size: 18px; font-family: monospace; color: #2e7d32; margin: 10px 0;">%s</p>
                <p style="color: #666; font-size: 12px; margin: 0;">请妥善保管您的卡密，不要泄露给他人。</p>
            </div>
`, kamiCode)
	}

	paymentTimeStr := ""
	if order.PaymentTime != nil {
		paymentTimeStr = order.PaymentTime.Format("2006-01-02 15:04:05")
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #4caf50;">%s - 支付成功</h2>
        <p>您好，</p>
        <p>您的订单已支付成功！</p>
        <div style="background: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 8px;">
            <p><strong>订单号：</strong>%s</p>
            <p><strong>商品名称：</strong>%s</p>
            <p><strong>时长：</strong>%d %s</p>
            <p><strong>支付金额：</strong>¥%.2f</p>
            <p><strong>支付时间：</strong>%s</p>
        </div>
        %s
        <p>感谢您的购买！</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, systemTitle, order.OrderNo, order.ProductName, order.Duration, order.DurationUnit, order.Price, paymentTimeStr, kamiSection)
}

func (s *NotificationService) buildOrderCancelledEmail(order *model.Order, reason string) string {
	systemTitle := config.GlobalConfig.ServerConfig.SystemTitle
	reasonText := "用户取消"
	if reason != "" {
		reasonText = reason
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #f44336;">%s - 订单已取消</h2>
        <p>您好，</p>
        <p>您的订单已取消。</p>
        <div style="background: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 8px;">
            <p><strong>订单号：</strong>%s</p>
            <p><strong>商品名称：</strong>%s</p>
            <p><strong>订单金额：</strong>¥%.2f</p>
            <p><strong>取消原因：</strong>%s</p>
        </div>
        <p>如有疑问，请联系客服。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, systemTitle, order.OrderNo, order.ProductName, order.Price, reasonText)
}

func (s *NotificationService) buildOrderRefundedEmail(order *model.Order) string {
	systemTitle := config.GlobalConfig.ServerConfig.SystemTitle
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #ff9800;">%s - 订单已退款</h2>
        <p>您好，</p>
        <p>您的订单已退款成功。</p>
        <div style="background: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 8px;">
            <p><strong>订单号：</strong>%s</p>
            <p><strong>商品名称：</strong>%s</p>
            <p><strong>退款金额：</strong>¥%.2f</p>
        </div>
        <p>退款将在1-3个工作日内原路返回。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, systemTitle, order.OrderNo, order.ProductName, order.Price)
}

func (s *NotificationService) buildKamiExpiringEmail(productName, kamiCode string, expireTime time.Time, daysLeft int) string {
	systemTitle := config.GlobalConfig.ServerConfig.SystemTitle
	urgencyColor := "#ff9800"
	if daysLeft <= 3 {
		urgencyColor = "#f44336"
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: %s;">%s - 卡密即将过期</h2>
        <p>您好，</p>
        <p>您的卡密即将过期，请及时续费。</p>
        <div style="background: #fff3e0; padding: 20px; margin: 20px 0; border-radius: 8px; border-left: 4px solid %s;">
            <p><strong>商品名称：</strong>%s</p>
            <p><strong>卡密：</strong><code style="background: #f5f5f5; padding: 2px 6px;">%s</code></p>
            <p><strong>过期时间：</strong>%s</p>
            <p style="color: %s; font-weight: bold;">剩余 %d 天</p>
        </div>
        <p>为避免服务中断，建议您尽快续费。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, urgencyColor, systemTitle, urgencyColor, productName, kamiCode, expireTime.Format("2006-01-02 15:04:05"), urgencyColor, daysLeft)
}

func (s *NotificationService) buildNewDeviceLoginEmail(device, ip, location string, loginTime time.Time) string {
	systemTitle := config.GlobalConfig.ServerConfig.SystemTitle
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2196f3;">%s - 新设备登录提醒</h2>
        <p>您好，</p>
        <p>您的账号在新设备上登录：</p>
        <div style="background: #e3f2fd; padding: 20px; margin: 20px 0; border-radius: 8px; border-left: 4px solid #2196f3;">
            <p><strong>设备信息：</strong>%s</p>
            <p><strong>登录IP：</strong>%s</p>
            <p><strong>IP归属地：</strong>%s</p>
            <p><strong>登录时间：</strong>%s</p>
        </div>
        <p style="color: #f44336;">如果这不是您本人的操作，请立即修改密码并检查账号安全。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, systemTitle, device, ip, location, loginTime.Format("2006-01-02 15:04:05"))
}

func (s *NotificationService) buildPasswordChangedEmail(ip string, changeTime time.Time) string {
	systemTitle := config.GlobalConfig.ServerConfig.SystemTitle
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #ff9800;">%s - 密码修改通知</h2>
        <p>您好，</p>
        <p>您的账号密码已被修改。</p>
        <div style="background: #fff3e0; padding: 20px; margin: 20px 0; border-radius: 8px; border-left: 4px solid #ff9800;">
            <p><strong>操作IP：</strong>%s</p>
            <p><strong>修改时间：</strong>%s</p>
        </div>
        <p style="color: #f44336;">如果这不是您本人的操作，请立即联系客服。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, systemTitle, ip, changeTime.Format("2006-01-02 15:04:05"))
}
