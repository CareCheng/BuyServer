package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ==================== 满意度评价 API ====================

// RateTicket 对工单进行满意度评价
// 用户在工单关闭后可以对服务进行评价
func RateTicket(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	if ticketNo == "" {
		c.JSON(400, gin.H{"success": false, "error": "工单编号不能为空"})
		return
	}

	// 获取工单
	ticket, err := SupportSvc.GetTicketByNo(ticketNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "工单不存在"})
		return
	}

	// 解析请求参数
	var req struct {
		Rating     int    `json:"rating" binding:"required,min=1,max=5"`
		Comment    string `json:"comment"`
		GuestToken string `json:"guest_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误，评分必须在1-5之间"})
		return
	}

	// 验证权限（用户只能评价自己的工单）
	userID, exists := c.Get("user_id")
	guestToken := req.GuestToken
	if guestToken == "" {
		guestToken = c.GetHeader("X-Guest-Token")
	}

	if ticket.UserID > 0 {
		// 注册用户的工单
		if !exists || userID.(uint) != ticket.UserID {
			c.JSON(403, gin.H{"success": false, "error": "无权评价此工单"})
			return
		}
	} else {
		// 游客工单，验证令牌
		if guestToken == "" || guestToken != ticket.GuestToken {
			c.JSON(403, gin.H{"success": false, "error": "无权评价此工单"})
			return
		}
	}

	// 提交评价
	if err := SupportSvc.RateTicket(ticket.ID, req.Rating, req.Comment); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "感谢您的评价"})
}

// RateLiveChat 对实时聊天进行满意度评价
func RateLiveChat(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(400, gin.H{"success": false, "error": "会话ID不能为空"})
		return
	}

	// 获取聊天会话
	chat, err := SupportSvc.GetLiveChatBySession(sessionID)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "会话不存在"})
		return
	}

	// 解析请求参数
	var req struct {
		Rating     int    `json:"rating" binding:"required,min=1,max=5"`
		Feedback   string `json:"feedback"`
		GuestToken string `json:"guest_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误，评分必须在1-5之间"})
		return
	}

	// 验证权限
	userID, exists := c.Get("user_id")
	guestToken := req.GuestToken
	if guestToken == "" {
		guestToken = c.GetHeader("X-Guest-Token")
	}

	if chat.UserID > 0 {
		// 注册用户的会话
		if !exists || userID.(uint) != chat.UserID {
			c.JSON(403, gin.H{"success": false, "error": "无权评价此会话"})
			return
		}
	} else {
		// 游客会话，验证令牌
		if guestToken == "" || guestToken != chat.GuestToken {
			c.JSON(403, gin.H{"success": false, "error": "无权评价此会话"})
			return
		}
	}

	// 提交评价
	if err := SupportSvc.RateLiveChat(chat.ID, req.Rating, req.Feedback); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "感谢您的评价"})
}

// GetTicketRating 获取工单评价信息
func GetTicketRating(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	if ticketNo == "" {
		c.JSON(400, gin.H{"success": false, "error": "工单编号不能为空"})
		return
	}

	ticket, err := SupportSvc.GetTicketByNo(ticketNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "工单不存在"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"rating":         ticket.Rating,
			"rating_comment": ticket.RatingComment,
			"rated_at":       ticket.RatedAt,
			"can_rate":       ticket.Rating == 0 && (ticket.Status == 3 || ticket.Status == 4),
		},
	})
}

// ==================== 管理后台满意度统计 API ====================

// AdminGetRatingStats 获取满意度统计（管理后台）
func AdminGetRatingStats(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats, err := SupportSvc.GetTicketRatingStats()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取统计数据失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": stats})
}

// AdminGetStaffRatingStats 获取客服个人满意度统计
func AdminGetStaffRatingStats(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	staffID, err := strconv.ParseUint(c.Param("staff_id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的客服ID"})
		return
	}

	stats, err := SupportSvc.GetStaffRatingStats(uint(staffID))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取统计数据失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": stats})
}

// StaffGetMyRatingStats 客服获取自己的满意度统计
func StaffGetMyRatingStats(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	staffID, exists := c.Get("staff_id")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "未登录"})
		return
	}

	stats, err := SupportSvc.GetStaffRatingStats(staffID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取统计数据失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": stats})
}
