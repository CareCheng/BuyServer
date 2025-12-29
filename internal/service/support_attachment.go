// Package service 提供业务逻辑服务
// support_attachment.go - 工单附件管理相关方法
package service

import (
	"time"

	"user-frontend/internal/model"
)

// SaveAttachment 保存附件信息
func (s *SupportService) SaveAttachment(ticketID, messageID uint, fileName, filePath, mimeType string, fileSize int64) (*model.SupportAttachment, error) {
	attachment := &model.SupportAttachment{
		TicketID:  ticketID,
		MessageID: messageID,
		FileName:  fileName,
		FilePath:  filePath,
		FileSize:  fileSize,
		MimeType:  mimeType,
	}

	if err := s.repo.GetDB().Create(attachment).Error; err != nil {
		return nil, err
	}

	return attachment, nil
}

// GetTicketAttachments 获取工单附件列表
func (s *SupportService) GetTicketAttachments(ticketID uint) ([]model.SupportAttachment, error) {
	var attachments []model.SupportAttachment
	if err := s.repo.GetDB().Where("ticket_id = ?", ticketID).Order("created_at ASC").Find(&attachments).Error; err != nil {
		return nil, err
	}
	return attachments, nil
}

// ReplyTicketWithAttachment 带附件回复工单
func (s *SupportService) ReplyTicketWithAttachment(ticketID uint, senderType string, senderID uint, senderName, content string, isInternal bool, fileURL, fileName string, fileSize int64) (*model.SupportMessage, error) {
	msgType := "text"
	if fileURL != "" {
		// 根据文件类型判断消息类型
		if isImageFile(fileName) {
			msgType = "image"
		} else {
			msgType = "file"
		}
	}

	msg := &model.SupportMessage{
		TicketID:   ticketID,
		SenderType: senderType,
		SenderID:   senderID,
		SenderName: senderName,
		Content:    content,
		MsgType:    msgType,
		FileURL:    fileURL,
		FileName:   fileName,
		FileSize:   fileSize,
		IsInternal: isInternal,
	}

	if err := s.repo.GetDB().Create(msg).Error; err != nil {
		return nil, err
	}

	// 更新工单状态和最后回复信息
	now := time.Now()
	updates := map[string]interface{}{
		"last_reply_at": now,
		"last_reply_by": senderName,
	}

	if senderType == "staff" && !isInternal {
		updates["status"] = model.TicketStatusReplied
	}
	if senderType == "user" || senderType == "guest" {
		var ticket model.SupportTicket
		s.repo.GetDB().First(&ticket, ticketID)
		if ticket.Status == model.TicketStatusReplied {
			updates["status"] = model.TicketStatusProcessing
		}
	}

	s.repo.GetDB().Model(&model.SupportTicket{}).Where("id = ?", ticketID).Updates(updates)

	return msg, nil
}

// isImageFile 判断是否为图片文件
func isImageFile(fileName string) bool {
	ext := ""
	if idx := len(fileName) - 1; idx > 0 {
		for i := idx; i >= 0; i-- {
			if fileName[i] == '.' {
				ext = fileName[i:]
				break
			}
		}
	}
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		return true
	}
	return false
}
