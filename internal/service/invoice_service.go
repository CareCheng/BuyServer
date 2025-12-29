package service

import (
	"errors"
	"fmt"
	"time"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// InvoiceService 发票服务
type InvoiceService struct {
	repo         *repository.Repository
	emailService *EmailService
}

// NewInvoiceService 创建发票服务实例
func NewInvoiceService(repo *repository.Repository, emailService *EmailService) *InvoiceService {
	return &InvoiceService{
		repo:         repo,
		emailService: emailService,
	}
}

// generateInvoiceNo 生成发票编号
func (s *InvoiceService) generateInvoiceNo() string {
	return fmt.Sprintf("INV%s%04d", time.Now().Format("20060102150405"), time.Now().Nanosecond()%10000)
}

// GetConfig 获取发票配置
// 返回：
//   - 发票配置
//   - 错误信息（如有）
func (s *InvoiceService) GetConfig() (*model.InvoiceConfig, error) {
	var config model.InvoiceConfig
	result := s.repo.GetDB().First(&config)
	if result.Error != nil {
		// 返回默认配置
		return &model.InvoiceConfig{
			Enabled:         false,
			MinAmount:       0,
			AutoIssue:       false,
			AllowPersonal:   true,
			AllowEnterprise: true,
			DefaultContent:  "信息服务费",
		}, nil
	}
	return &config, nil
}

// SaveConfig 保存发票配置
// 参数：
//   - config: 配置信息
// 返回：
//   - 错误信息（如有）
func (s *InvoiceService) SaveConfig(config *model.InvoiceConfig) error {
	var existing model.InvoiceConfig
	result := s.repo.GetDB().First(&existing)
	if result.Error != nil {
		// 创建新配置
		return s.repo.GetDB().Create(config).Error
	}
	// 更新现有配置
	config.ID = existing.ID
	return s.repo.GetDB().Save(config).Error
}

// ApplyInvoice 申请开票
// 参数：
//   - userID: 用户ID
//   - orderNo: 订单号
//   - req: 开票请求
// 返回：
//   - 发票记录
//   - 错误信息（如有）
func (s *InvoiceService) ApplyInvoice(userID uint, orderNo string, req *InvoiceApplyRequest) (*model.Invoice, error) {
	// 检查发票功能是否启用
	config, _ := s.GetConfig()
	if !config.Enabled {
		return nil, errors.New("发票功能未启用")
	}

	// 检查订单是否存在且已支付
	var order model.Order
	if err := s.repo.GetDB().Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != 1 && order.Status != 2 {
		return nil, errors.New("订单未支付，无法开票")
	}

	// 检查是否已申请过发票
	var existing model.Invoice
	if err := s.repo.GetDB().Where("order_no = ? AND status != ?", orderNo, model.InvoiceStatusCanceled).First(&existing).Error; err == nil {
		return nil, errors.New("该订单已申请过发票")
	}

	// 检查最低开票金额
	if order.Price < config.MinAmount {
		return nil, fmt.Errorf("订单金额不足，最低开票金额为 %.2f 元", config.MinAmount)
	}

	// 检查发票类型是否允许
	if req.Type == model.InvoiceTypePersonal && !config.AllowPersonal {
		return nil, errors.New("不支持个人发票")
	}
	if req.Type == model.InvoiceTypeEnterprise && !config.AllowEnterprise {
		return nil, errors.New("不支持企业发票")
	}

	// 创建发票记录
	invoice := &model.Invoice{
		InvoiceNo:   s.generateInvoiceNo(),
		UserID:      userID,
		OrderNo:     orderNo,
		Type:        req.Type,
		TitleType:   req.Type,
		Title:       req.Title,
		TaxNo:       req.TaxNo,
		Amount:      order.Price,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		Content:     config.DefaultContent,
		Remark:      req.Remark,
		Status:      model.InvoiceStatusPending,
	}

	if err := s.repo.GetDB().Create(invoice).Error; err != nil {
		return nil, err
	}

	return invoice, nil
}

// InvoiceApplyRequest 开票申请请求
type InvoiceApplyRequest struct {
	Type        string `json:"type" binding:"required"`  // personal/enterprise
	Title       string `json:"title" binding:"required"` // 发票抬头
	TaxNo       string `json:"tax_no"`                   // 税号（企业发票必填）
	Email       string `json:"email" binding:"required"` // 接收邮箱
	Phone       string `json:"phone"`                    // 联系电话
	Address     string `json:"address"`                  // 企业地址
	BankName    string `json:"bank_name"`                // 开户银行
	BankAccount string `json:"bank_account"`             // 银行账号
	Remark      string `json:"remark"`                   // 备注
}

// GetUserInvoices 获取用户发票列表
// 参数：
//   - userID: 用户ID
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 发票列表
//   - 总数
//   - 错误信息（如有）
func (s *InvoiceService) GetUserInvoices(userID uint, page, pageSize int) ([]model.Invoice, int64, error) {
	var total int64
	s.repo.GetDB().Model(&model.Invoice{}).Where("user_id = ?", userID).Count(&total)

	var invoices []model.Invoice
	offset := (page - 1) * pageSize
	err := s.repo.GetDB().Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&invoices).Error

	return invoices, total, err
}

// GetInvoiceDetail 获取发票详情
// 参数：
//   - userID: 用户ID
//   - invoiceNo: 发票编号
// 返回：
//   - 发票详情
//   - 错误信息（如有）
func (s *InvoiceService) GetInvoiceDetail(userID uint, invoiceNo string) (*model.Invoice, error) {
	var invoice model.Invoice
	err := s.repo.GetDB().Where("invoice_no = ? AND user_id = ?", invoiceNo, userID).First(&invoice).Error
	return &invoice, err
}

// CancelInvoice 取消发票申请
// 参数：
//   - userID: 用户ID
//   - invoiceNo: 发票编号
// 返回：
//   - 错误信息（如有）
func (s *InvoiceService) CancelInvoice(userID uint, invoiceNo string) error {
	var invoice model.Invoice
	if err := s.repo.GetDB().Where("invoice_no = ? AND user_id = ?", invoiceNo, userID).First(&invoice).Error; err != nil {
		return errors.New("发票不存在")
	}

	if invoice.Status != model.InvoiceStatusPending {
		return errors.New("只能取消待开具的发票")
	}

	invoice.Status = model.InvoiceStatusCanceled
	return s.repo.GetDB().Save(&invoice).Error
}

// AdminGetInvoices 管理员获取发票列表
// 参数：
//   - status: 状态筛选（-1表示全部）
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 发票列表
//   - 总数
//   - 错误信息（如有）
func (s *InvoiceService) AdminGetInvoices(status, page, pageSize int) ([]model.Invoice, int64, error) {
	var total int64
	query := s.repo.GetDB().Model(&model.Invoice{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	var invoices []model.Invoice
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&invoices).Error

	return invoices, total, err
}

// AdminIssueInvoice 管理员开具发票
// 参数：
//   - invoiceNo: 发票编号
//   - invoiceURL: 电子发票URL
//   - issuedBy: 开具人
// 返回：
//   - 错误信息（如有）
func (s *InvoiceService) AdminIssueInvoice(invoiceNo, invoiceURL, issuedBy string) error {
	var invoice model.Invoice
	if err := s.repo.GetDB().Where("invoice_no = ?", invoiceNo).First(&invoice).Error; err != nil {
		return errors.New("发票不存在")
	}

	if invoice.Status != model.InvoiceStatusPending {
		return errors.New("只能开具待处理的发票")
	}

	now := time.Now()
	invoice.Status = model.InvoiceStatusIssued
	invoice.InvoiceURL = invoiceURL
	invoice.IssuedAt = &now
	invoice.IssuedBy = issuedBy

	if err := s.repo.GetDB().Save(&invoice).Error; err != nil {
		return err
	}

	// 发送邮件通知
	if s.emailService != nil && invoice.Email != "" {
		subject := "您的电子发票已开具"
		body := fmt.Sprintf(`
			<h2>电子发票通知</h2>
			<p>尊敬的用户，您申请的电子发票已开具成功。</p>
			<p><strong>发票编号：</strong>%s</p>
			<p><strong>发票抬头：</strong>%s</p>
			<p><strong>发票金额：</strong>%.2f 元</p>
			<p><strong>开具时间：</strong>%s</p>
			<p>请登录系统下载电子发票。</p>
		`, invoice.InvoiceNo, invoice.Title, invoice.Amount, now.Format("2006-01-02 15:04:05"))
		s.emailService.SendEmail(invoice.Email, subject, body)
	}

	return nil
}

// AdminRejectInvoice 管理员拒绝发票申请
// 参数：
//   - invoiceNo: 发票编号
//   - reason: 拒绝原因
// 返回：
//   - 错误信息（如有）
func (s *InvoiceService) AdminRejectInvoice(invoiceNo, reason string) error {
	var invoice model.Invoice
	if err := s.repo.GetDB().Where("invoice_no = ?", invoiceNo).First(&invoice).Error; err != nil {
		return errors.New("发票不存在")
	}

	if invoice.Status != model.InvoiceStatusPending {
		return errors.New("只能拒绝待处理的发票")
	}

	invoice.Status = model.InvoiceStatusRejected
	invoice.RejectReason = reason

	if err := s.repo.GetDB().Save(&invoice).Error; err != nil {
		return err
	}

	// 发送邮件通知
	if s.emailService != nil && invoice.Email != "" {
		subject := "您的发票申请已被拒绝"
		body := fmt.Sprintf(`
			<h2>发票申请通知</h2>
			<p>尊敬的用户，很抱歉，您的发票申请已被拒绝。</p>
			<p><strong>发票编号：</strong>%s</p>
			<p><strong>拒绝原因：</strong>%s</p>
			<p>如有疑问，请联系客服。</p>
		`, invoice.InvoiceNo, reason)
		s.emailService.SendEmail(invoice.Email, subject, body)
	}

	return nil
}

// SaveInvoiceTitle 保存发票抬头
// 参数：
//   - userID: 用户ID
//   - title: 抬头信息
// 返回：
//   - 错误信息（如有）
func (s *InvoiceService) SaveInvoiceTitle(userID uint, title *model.InvoiceTitle) error {
	title.UserID = userID

	// 如果设为默认，取消其他默认
	if title.IsDefault {
		s.repo.GetDB().Model(&model.InvoiceTitle{}).Where("user_id = ?", userID).Update("is_default", false)
	}

	if title.ID > 0 {
		// 更新
		var existing model.InvoiceTitle
		if err := s.repo.GetDB().Where("id = ? AND user_id = ?", title.ID, userID).First(&existing).Error; err != nil {
			return errors.New("抬头不存在")
		}
		return s.repo.GetDB().Save(title).Error
	}
	// 创建
	return s.repo.GetDB().Create(title).Error
}

// GetUserInvoiceTitles 获取用户发票抬头列表
// 参数：
//   - userID: 用户ID
// 返回：
//   - 抬头列表
//   - 错误信息（如有）
func (s *InvoiceService) GetUserInvoiceTitles(userID uint) ([]model.InvoiceTitle, error) {
	var titles []model.InvoiceTitle
	err := s.repo.GetDB().Where("user_id = ?", userID).Order("is_default DESC, created_at DESC").Find(&titles).Error
	return titles, err
}

// DeleteInvoiceTitle 删除发票抬头
// 参数：
//   - userID: 用户ID
//   - titleID: 抬头ID
// 返回：
//   - 错误信息（如有）
func (s *InvoiceService) DeleteInvoiceTitle(userID uint, titleID uint) error {
	result := s.repo.GetDB().Where("id = ? AND user_id = ?", titleID, userID).Delete(&model.InvoiceTitle{})
	if result.RowsAffected == 0 {
		return errors.New("抬头不存在")
	}
	return result.Error
}

// GetInvoiceStats 获取发票统计
// 返回：
//   - 统计信息
func (s *InvoiceService) GetInvoiceStats() map[string]interface{} {
	var pending, issued, rejected int64
	var totalAmount float64

	s.repo.GetDB().Model(&model.Invoice{}).Where("status = ?", model.InvoiceStatusPending).Count(&pending)
	s.repo.GetDB().Model(&model.Invoice{}).Where("status = ?", model.InvoiceStatusIssued).Count(&issued)
	s.repo.GetDB().Model(&model.Invoice{}).Where("status = ?", model.InvoiceStatusRejected).Count(&rejected)
	s.repo.GetDB().Model(&model.Invoice{}).Where("status = ?", model.InvoiceStatusIssued).Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount)

	return map[string]interface{}{
		"pending":      pending,
		"issued":       issued,
		"rejected":     rejected,
		"total_amount": totalAmount,
	}
}
