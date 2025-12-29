// Package service 提供业务逻辑服务
// support_rating.go - 满意度评价相关方法
package service

import (
	"errors"
	"fmt"
	"time"

	"user-frontend/internal/model"
)

// RateTicket 对工单进行满意度评价
// 参数：
//   - ticketID: 工单ID
//   - rating: 评分（1-5星）
//   - comment: 评价内容
// 返回：
//   - 错误信息（如有）
func (s *SupportService) RateTicket(ticketID uint, rating int, comment string) error {
	// 验证评分范围
	if rating < 1 || rating > 5 {
		return errors.New("评分必须在1-5之间")
	}

	// 获取工单
	ticket, err := s.GetTicketByID(ticketID)
	if err != nil {
		return errors.New("工单不存在")
	}

	// 检查工单状态（只有已解决或已关闭的工单才能评价）
	if ticket.Status != model.TicketStatusResolved && ticket.Status != model.TicketStatusClosed {
		return errors.New("只能对已解决或已关闭的工单进行评价")
	}

	// 检查是否已评价
	if ticket.Rating > 0 {
		return errors.New("该工单已评价，不能重复评价")
	}

	// 更新评价信息
	now := time.Now()
	return s.repo.GetDB().Model(&model.SupportTicket{}).Where("id = ?", ticketID).Updates(map[string]interface{}{
		"rating":         rating,
		"rating_comment": comment,
		"rated_at":       now,
	}).Error
}

// RateLiveChat 对实时聊天进行满意度评价
// 参数：
//   - chatID: 聊天会话ID
//   - rating: 评分（1-5星）
//   - feedback: 评价内容
// 返回：
//   - 错误信息（如有）
func (s *SupportService) RateLiveChat(chatID uint, rating int, feedback string) error {
	// 验证评分范围
	if rating < 1 || rating > 5 {
		return errors.New("评分必须在1-5之间")
	}

	// 获取聊天会话
	var chat model.LiveChat
	if err := s.repo.GetDB().First(&chat, chatID).Error; err != nil {
		return errors.New("聊天会话不存在")
	}

	// 检查会话状态（只有已结束的会话才能评价）
	if chat.Status != model.ChatStatusEnded {
		return errors.New("只能对已结束的会话进行评价")
	}

	// 检查是否已评价
	if chat.Rating > 0 {
		return errors.New("该会话已评价，不能重复评价")
	}

	// 更新评价信息
	return s.repo.GetDB().Model(&model.LiveChat{}).Where("id = ?", chatID).Updates(map[string]interface{}{
		"rating":   rating,
		"feedback": feedback,
	}).Error
}

// GetTicketRatingStats 获取工单满意度统计
// 返回：
//   - 统计数据（各星级数量、平均分等）
//   - 错误信息（如有）
func (s *SupportService) GetTicketRatingStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 各星级数量
	var star1, star2, star3, star4, star5 int64
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("rating = 1").Count(&star1)
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("rating = 2").Count(&star2)
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("rating = 3").Count(&star3)
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("rating = 4").Count(&star4)
	s.repo.GetDB().Model(&model.SupportTicket{}).Where("rating = 5").Count(&star5)

	stats["star1"] = star1
	stats["star2"] = star2
	stats["star3"] = star3
	stats["star4"] = star4
	stats["star5"] = star5

	// 总评价数
	total := star1 + star2 + star3 + star4 + star5
	stats["total"] = total

	// 平均分
	if total > 0 {
		avgScore := float64(star1*1+star2*2+star3*3+star4*4+star5*5) / float64(total)
		stats["average"] = fmt.Sprintf("%.2f", avgScore)
	} else {
		stats["average"] = "0.00"
	}

	// 满意率（4-5星占比）
	if total > 0 {
		satisfactionRate := float64(star4+star5) / float64(total) * 100
		stats["satisfaction_rate"] = fmt.Sprintf("%.1f%%", satisfactionRate)
	} else {
		stats["satisfaction_rate"] = "0.0%"
	}

	return stats, nil
}

// GetStaffRatingStats 获取客服个人满意度统计
// 参数：
//   - staffID: 客服ID
// 返回：
//   - 统计数据
//   - 错误信息（如有）
func (s *SupportService) GetStaffRatingStats(staffID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 该客服处理的已评价工单
	var tickets []model.SupportTicket
	s.repo.GetDB().Where("assigned_to = ? AND rating > 0", staffID).Find(&tickets)

	if len(tickets) == 0 {
		stats["total"] = 0
		stats["average"] = "0.00"
		stats["satisfaction_rate"] = "0.0%"
		return stats, nil
	}

	// 统计各星级
	var star1, star2, star3, star4, star5 int64
	var totalScore int64
	for _, t := range tickets {
		totalScore += int64(t.Rating)
		switch t.Rating {
		case 1:
			star1++
		case 2:
			star2++
		case 3:
			star3++
		case 4:
			star4++
		case 5:
			star5++
		}
	}

	total := int64(len(tickets))
	stats["star1"] = star1
	stats["star2"] = star2
	stats["star3"] = star3
	stats["star4"] = star4
	stats["star5"] = star5
	stats["total"] = total
	stats["average"] = fmt.Sprintf("%.2f", float64(totalScore)/float64(total))
	stats["satisfaction_rate"] = fmt.Sprintf("%.1f%%", float64(star4+star5)/float64(total)*100)

	return stats, nil
}
