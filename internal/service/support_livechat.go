// Package service 提供业务逻辑服务
// support_livechat.go - 实时聊天相关方法
package service

import (
	"time"

	"user-frontend/internal/model"

	"github.com/google/uuid"
)

// CreateLiveChat 创建实时聊天会话
func (s *SupportService) CreateLiveChat(userID uint, username, guestToken string) (*model.LiveChat, error) {
	sessionID := uuid.New().String()

	if userID == 0 && guestToken == "" {
		guestToken = generateToken(32)
	}
	if username == "" && userID == 0 {
		username = "游客_" + guestToken[:8]
	}

	chat := &model.LiveChat{
		SessionID:  sessionID,
		UserID:     userID,
		Username:   username,
		GuestToken: guestToken,
		Status:     model.ChatStatusWaiting,
	}

	if err := s.repo.GetDB().Create(chat).Error; err != nil {
		return nil, err
	}

	return chat, nil
}

// GetLiveChatBySession 根据会话ID获取聊天
func (s *SupportService) GetLiveChatBySession(sessionID string) (*model.LiveChat, error) {
	var chat model.LiveChat
	if err := s.repo.GetDB().Where("session_id = ?", sessionID).First(&chat).Error; err != nil {
		return nil, err
	}
	return &chat, nil
}

// GetWaitingChats 获取等待接入的聊天
func (s *SupportService) GetWaitingChats() ([]model.LiveChat, error) {
	var chats []model.LiveChat
	if err := s.repo.GetDB().Where("status = ?", model.ChatStatusWaiting).Order("created_at ASC").Find(&chats).Error; err != nil {
		return nil, err
	}
	return chats, nil
}

// AcceptChat 客服接入聊天
func (s *SupportService) AcceptChat(chatID, staffID uint, staffName string) error {
	return s.repo.GetDB().Model(&model.LiveChat{}).Where("id = ?", chatID).Updates(map[string]interface{}{
		"staff_id":   staffID,
		"staff_name": staffName,
		"status":     model.ChatStatusActive,
	}).Error
}

// EndChat 结束聊天
func (s *SupportService) EndChat(chatID uint) error {
	now := time.Now()
	return s.repo.GetDB().Model(&model.LiveChat{}).Where("id = ?", chatID).Updates(map[string]interface{}{
		"status":   model.ChatStatusEnded,
		"ended_at": now,
	}).Error
}

// SendChatMessage 发送聊天消息
func (s *SupportService) SendChatMessage(chatID uint, senderType string, senderID uint, senderName, content, msgType string) (*model.LiveChatMessage, error) {
	msg := &model.LiveChatMessage{
		ChatID:     chatID,
		SenderType: senderType,
		SenderID:   senderID,
		SenderName: senderName,
		Content:    content,
		MsgType:    msgType,
	}

	if err := s.repo.GetDB().Create(msg).Error; err != nil {
		return nil, err
	}

	return msg, nil
}

// GetChatMessages 获取聊天消息
func (s *SupportService) GetChatMessages(chatID uint, afterID uint) ([]model.LiveChatMessage, error) {
	var messages []model.LiveChatMessage
	query := s.repo.GetDB().Where("chat_id = ?", chatID)
	if afterID > 0 {
		query = query.Where("id > ?", afterID)
	}
	if err := query.Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// IsWorkingTime 检查当前是否在工作时间内
func (s *SupportService) IsWorkingTime() bool {
	config, _ := s.GetSupportConfig()

	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // 周日改为7
	}

	// 检查是否工作日
	workingDays := config.WorkingDays
	if workingDays == "" {
		workingDays = "1,2,3,4,5"
	}

	isWorkDay := false
	for _, d := range workingDays {
		if d >= '1' && d <= '7' && int(d-'0') == weekday {
			isWorkDay = true
			break
		}
	}
	if !isWorkDay {
		return false
	}

	// 检查是否在工作时间段内
	startTime := config.WorkingHoursStart
	endTime := config.WorkingHoursEnd
	if startTime == "" {
		startTime = "09:00"
	}
	if endTime == "" {
		endTime = "18:00"
	}

	currentTime := now.Format("15:04")
	return currentTime >= startTime && currentTime <= endTime
}
