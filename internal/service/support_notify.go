// Package service 提供业务逻辑服务
// support_notify.go - 工单邮件通知相关方法
package service

import (
	"fmt"

	"user-frontend/internal/model"
)

// emailService 邮件服务引用（需要在初始化时设置）
var supportEmailSvc *EmailService

// SetEmailService 设置邮件服务
func (s *SupportService) SetEmailService(emailSvc *EmailService) {
	supportEmailSvc = emailSvc
}

// NotifyUserOnReply 工单有新回复时通知用户
func (s *SupportService) NotifyUserOnReply(ticketID uint, replyContent string) {
	config, _ := s.GetSupportConfig()
	if !config.EnableEmailNotify || !config.NotifyOnReply {
		return
	}

	ticket, err := s.GetTicketByID(ticketID)
	if err != nil || ticket.Email == "" {
		return
	}

	if supportEmailSvc == nil {
		return
	}

	subject := fmt.Sprintf("您的工单 [%s] 有新回复", ticket.TicketNo)
	body := fmt.Sprintf(`
<h3>您好，%s</h3>
<p>您的工单 <strong>%s</strong> 有新的回复：</p>
<div style="background:#f5f5f5;padding:15px;border-radius:5px;margin:15px 0;">
<p><strong>工单主题：</strong>%s</p>
<p><strong>回复内容：</strong></p>
<p>%s</p>
</div>
<p>请登录系统查看详情并回复。</p>
<p style="color:#999;font-size:12px;">此邮件由系统自动发送，请勿直接回复。</p>
`, ticket.Username, ticket.TicketNo, ticket.Subject, replyContent)

	go supportEmailSvc.SendEmail(ticket.Email, subject, body)
}

// NotifyStaffOnNewTicket 新工单时通知在线客服
func (s *SupportService) NotifyStaffOnNewTicket(ticket *model.SupportTicket) {
	config, _ := s.GetSupportConfig()
	if !config.EnableEmailNotify || !config.NotifyOnNewTicket {
		return
	}

	if supportEmailSvc == nil {
		return
	}

	// 获取在线客服的邮箱
	onlineStaff, _ := s.GetOnlineStaff()
	if len(onlineStaff) == 0 {
		return
	}

	subject := fmt.Sprintf("新工单提醒 [%s] %s", ticket.TicketNo, ticket.Subject)
	body := fmt.Sprintf(`
<h3>有新的工单需要处理</h3>
<div style="background:#f5f5f5;padding:15px;border-radius:5px;margin:15px 0;">
<p><strong>工单编号：</strong>%s</p>
<p><strong>用户：</strong>%s</p>
<p><strong>主题：</strong>%s</p>
<p><strong>分类：</strong>%s</p>
<p><strong>优先级：</strong>%s</p>
</div>
<p>请登录客服后台处理。</p>
`, ticket.TicketNo, ticket.Username, ticket.Subject, ticket.Category, model.GetTicketPriorityText(ticket.Priority))

	for _, staff := range onlineStaff {
		if staff.Email != "" {
			go supportEmailSvc.SendEmail(staff.Email, subject, body)
		}
	}
}
