// Package service 提供业务逻辑服务
// support_ticket.go - 工单管理相关方法
package service

import (
	"fmt"
	"time"

	"user-frontend/internal/model"
)

// CreateTicket 创建工单
// 参数：
//   - userID: 用户ID（游客为0）
//   - username: 用户名
//   - email: 邮箱
//   - subject: 工单主题
//   - category: 工单分类
//   - content: 工单内容
//   - relatedOrder: 关联订单号
//   - priority: 优先级
// 返回：
//   - 创建的工单
//   - 错误信息（如有）
func (s *SupportService) CreateTicket(userID uint, username, email, subject, category, content, relatedOrder string, priority int) (*model.SupportTicket, error) {
	// 生成工单编号
	ticketNo := fmt.Sprintf("TK%s%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)

	// 游客令牌
	var guestToken string
	if userID == 0 {
		guestToken = generateToken(32)
		if username == "" {
			username = "游客_" + guestToken[:8]
		}
	}

	ticket := &model.SupportTicket{
		TicketNo:     ticketNo,
		UserID:       userID,
		Username:     username,
		Email:        email,
		Subject:      subject,
		Category:     category,
		Priority:     priority,
		Status:       model.TicketStatusPending,
		RelatedOrder: relatedOrder,
		GuestToken:   guestToken,
	}

	if err := s.repo.GetDB().Create(ticket).Error; err != nil {
		return nil, err
	}

	// 创建初始消息
	senderType := "user"
	if userID == 0 {
		senderType = "guest"
	}

	msg := &model.SupportMessage{
		TicketID:   ticket.ID,
		SenderType: senderType,
		SenderID:   userID,
		SenderName: username,
		Content:    content,
	}
	s.repo.GetDB().Create(msg)

	return ticket, nil
}

// GetTicketByNo 根据工单号获取工单
func (s *SupportService) GetTicketByNo(ticketNo string) (*model.SupportTicket, error) {
	var ticket model.SupportTicket
	if err := s.repo.GetDB().Where("ticket_no = ?", ticketNo).First(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

// GetTicketByID 根据ID获取工单
func (s *SupportService) GetTicketByID(ticketID uint) (*model.SupportTicket, error) {
	var ticket model.SupportTicket
	if err := s.repo.GetDB().First(&ticket, ticketID).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

// GetTicketByGuestToken 游客通过令牌获取工单
func (s *SupportService) GetTicketByGuestToken(ticketNo, guestToken string) (*model.SupportTicket, error) {
	var ticket model.SupportTicket
	if err := s.repo.GetDB().Where("ticket_no = ? AND guest_token = ?", ticketNo, guestToken).First(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

// GetUserTickets 获取用户工单列表
func (s *SupportService) GetUserTickets(userID uint, status int, page, pageSize int) ([]model.SupportTicket, int64, error) {
	var tickets []model.SupportTicket
	var total int64

	query := s.repo.GetDB().Model(&model.SupportTicket{}).Where("user_id = ?", userID)
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// GetGuestTickets 获取游客工单列表
func (s *SupportService) GetGuestTickets(guestToken string) ([]model.SupportTicket, error) {
	var tickets []model.SupportTicket
	if err := s.repo.GetDB().Where("guest_token = ?", guestToken).Order("created_at DESC").Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

// GetTicketMessages 获取工单消息
func (s *SupportService) GetTicketMessages(ticketID uint, includeInternal bool) ([]model.SupportMessage, error) {
	var messages []model.SupportMessage
	query := s.repo.GetDB().Where("ticket_id = ?", ticketID)
	if !includeInternal {
		query = query.Where("is_internal = ?", false)
	}
	if err := query.Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// ReplyTicket 回复工单
func (s *SupportService) ReplyTicket(ticketID uint, senderType string, senderID uint, senderName, content string, isInternal bool) (*model.SupportMessage, error) {
	msg := &model.SupportMessage{
		TicketID:   ticketID,
		SenderType: senderType,
		SenderID:   senderID,
		SenderName: senderName,
		Content:    content,
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

	// 客服回复时更新状态为已回复
	if senderType == "staff" && !isInternal {
		updates["status"] = model.TicketStatusReplied
	}
	// 用户回复时如果状态是已回复，改为处理中
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

// UpdateTicketStatus 更新工单状态
func (s *SupportService) UpdateTicketStatus(ticketID uint, status int, operatorName string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == model.TicketStatusClosed || status == model.TicketStatusResolved {
		now := time.Now()
		updates["closed_at"] = now
		updates["closed_by"] = operatorName
	}

	return s.repo.GetDB().Model(&model.SupportTicket{}).Where("id = ?", ticketID).Updates(updates).Error
}

// AssignTicket 分配工单给客服
func (s *SupportService) AssignTicket(ticketID, staffID uint, staffName string) error {
	return s.repo.GetDB().Model(&model.SupportTicket{}).Where("id = ?", ticketID).Updates(map[string]interface{}{
		"assigned_to":   staffID,
		"assigned_name": staffName,
		"status":        model.TicketStatusProcessing,
	}).Error
}

// GetAllTickets 获取所有工单（客服后台）
func (s *SupportService) GetAllTickets(status, priority int, category string, page, pageSize int) ([]model.SupportTicket, int64, error) {
	var tickets []model.SupportTicket
	var total int64

	query := s.repo.GetDB().Model(&model.SupportTicket{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if priority > 0 {
		query = query.Where("priority = ?", priority)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("priority DESC, created_at ASC").Offset(offset).Limit(pageSize).Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// GetStaffTickets 获取分配给指定客服的工单
func (s *SupportService) GetStaffTickets(staffID uint, status int, page, pageSize int) ([]model.SupportTicket, int64, error) {
	var tickets []model.SupportTicket
	var total int64

	query := s.repo.GetDB().Model(&model.SupportTicket{}).Where("assigned_to = ?", staffID)
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("priority DESC, created_at ASC").Offset(offset).Limit(pageSize).Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// GetTicketStats 获取工单统计
func (s *SupportService) GetTicketStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// 各状态数量
	var pending, processing, replied, resolved, closed int64
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("status = ?", model.TicketStatusPending).Count(&pending)
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("status = ?", model.TicketStatusProcessing).Count(&processing)
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("status = ?", model.TicketStatusReplied).Count(&replied)
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("status = ?", model.TicketStatusResolved).Count(&resolved)
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("status = ?", model.TicketStatusClosed).Count(&closed)

	stats["pending"] = pending
	stats["processing"] = processing
	stats["replied"] = replied
	stats["resolved"] = resolved
	stats["closed"] = closed
	stats["total"] = pending + processing + replied + resolved + closed

	// 今日新增
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var todayCount int64
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("created_at >= ?", todayStart).Count(&todayCount)
	stats["today"] = todayCount

	return stats, nil
}

// TransferTicket 转接工单给其他客服
func (s *SupportService) TransferTicket(ticketID, fromStaffID, toStaffID uint, reason string) error {
	ticket, err := s.GetTicketByID(ticketID)
	if err != nil {
		return fmt.Errorf("工单不存在")
	}

	toStaff, err := s.GetStaffByID(toStaffID)
	if err != nil {
		return fmt.Errorf("目标客服不存在")
	}

	fromStaff, _ := s.GetStaffByID(fromStaffID)
	fromName := ""
	if fromStaff != nil {
		fromName = fromStaff.Nickname
	}

	// 记录转接日志
	transferLog := fmt.Sprintf(`{"time":"%s","from_id":%d,"from_name":"%s","to_id":%d,"to_name":"%s","reason":"%s"}`,
		time.Now().Format("2006-01-02 15:04:05"), fromStaffID, fromName, toStaffID, toStaff.Nickname, reason)

	existingLog := ticket.TransferLog
	if existingLog == "" {
		existingLog = "[]"
	}
	newLog := existingLog[:len(existingLog)-1]
	if len(newLog) > 1 {
		newLog += ","
	}
	newLog += transferLog + "]"

	// 更新工单
	err = s.repo.GetDB().Model(&model.SupportTicket{}).Where("id = ?", ticketID).Updates(map[string]interface{}{
		"assigned_to":    toStaffID,
		"assigned_name":  toStaff.Nickname,
		"transfer_count": ticket.TransferCount + 1,
		"transfer_log":   newLog,
	}).Error

	if err != nil {
		return err
	}

	// 添加系统消息
	s.ReplyTicket(ticketID, "system", 0, "系统",
		fmt.Sprintf("工单已从 %s 转接给 %s，原因：%s", fromName, toStaff.Nickname, reason), true)

	// 更新双方客服负载
	if fromStaffID > 0 {
		s.UpdateStaffLoad(fromStaffID)
	}
	s.UpdateStaffLoad(toStaffID)

	return nil
}

// MergeTickets 合并工单
func (s *SupportService) MergeTickets(targetTicketID uint, sourceTicketIDs []uint, operatorName string) error {
	targetTicket, err := s.GetTicketByID(targetTicketID)
	if err != nil {
		return fmt.Errorf("目标工单不存在")
	}

	for _, sourceID := range sourceTicketIDs {
		if sourceID == targetTicketID {
			continue
		}

		sourceTicket, err := s.GetTicketByID(sourceID)
		if err != nil {
			continue
		}

		// 将源工单的消息复制到目标工单
		var messages []model.SupportMessage
		s.repo.GetDB().Where("ticket_id = ?", sourceID).Find(&messages)

		for _, msg := range messages {
			newMsg := model.SupportMessage{
				TicketID:   targetTicketID,
				SenderType: msg.SenderType,
				SenderID:   msg.SenderID,
				SenderName: msg.SenderName,
				Content:    fmt.Sprintf("[来自工单 %s] %s", sourceTicket.TicketNo, msg.Content),
				MsgType:    msg.MsgType,
				FileURL:    msg.FileURL,
				FileName:   msg.FileName,
				FileSize:   msg.FileSize,
				IsInternal: msg.IsInternal,
				CreatedAt:  msg.CreatedAt,
			}
			s.repo.GetDB().Create(&newMsg)
		}

		// 标记源工单为已合并
		s.repo.GetDB().Model(&model.SupportTicket{}).Where("id = ?", sourceID).Updates(map[string]interface{}{
			"status":    model.TicketStatusMerged,
			"merged_to": targetTicketID,
			"closed_at": time.Now(),
			"closed_by": operatorName,
		})

		// 添加系统消息到源工单
		s.ReplyTicket(sourceID, "system", 0, "系统",
			fmt.Sprintf("此工单已合并到工单 %s", targetTicket.TicketNo), false)
	}

	// 更新目标工单的合并来源
	mergedFrom := fmt.Sprintf("%v", sourceTicketIDs)
	existingMerged := targetTicket.MergedFrom
	if existingMerged != "" && existingMerged != "[]" {
		mergedFrom = existingMerged[:len(existingMerged)-1] + "," + mergedFrom[1:]
	}

	s.repo.GetDB().Model(&model.SupportTicket{}).Where("id = ?", targetTicketID).
		Update("merged_from", mergedFrom)

	// 添加系统消息到目标工单
	s.ReplyTicket(targetTicketID, "system", 0, "系统",
		fmt.Sprintf("已合并 %d 个工单到此工单", len(sourceTicketIDs)), true)

	return nil
}

// AutoAssignTicket 自动分配工单给负载最低的在线客服
func (s *SupportService) AutoAssignTicket(ticketID uint) error {
	config, _ := s.GetSupportConfig()
	if !config.EnableAutoAssign {
		return nil
	}

	// 获取在线且未满负载的客服，按当前负载排序
	var staff model.SupportStaff
	err := s.repo.GetDB().
		Where("status = 1 AND current_load < max_tickets").
		Order("current_load ASC").
		First(&staff).Error

	if err != nil {
		return nil // 没有可用客服，不分配
	}

	// 分配工单
	if err := s.AssignTicket(ticketID, staff.ID, staff.Nickname); err != nil {
		return err
	}

	// 更新客服负载
	s.repo.GetDB().Model(&model.SupportStaff{}).Where("id = ?", staff.ID).
		Update("current_load", staff.CurrentLoad+1)

	return nil
}

// AutoCloseInactiveTickets 自动关闭长时间无回复的工单
func (s *SupportService) AutoCloseInactiveTickets(hours int) {
	threshold := time.Now().Add(-time.Duration(hours) * time.Hour)
	s.repo.GetDB().Model(&model.SupportTicket{}).
		Where("status IN (?, ?) AND last_reply_at < ?", model.TicketStatusReplied, model.TicketStatusProcessing, threshold).
		Updates(map[string]interface{}{
			"status":    model.TicketStatusClosed,
			"closed_at": time.Now(),
			"closed_by": "系统自动关闭",
		})
}
