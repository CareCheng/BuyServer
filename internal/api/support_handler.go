package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==========================================
//         用户端工单 API
// ==========================================

// CreateTicket 创建工单
func CreateTicket(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Subject      string `json:"subject" binding:"required"`
		Category     string `json:"category" binding:"required"`
		Content      string `json:"content" binding:"required"`
		Email        string `json:"email"`
		Priority     int    `json:"priority"`
		RelatedOrder string `json:"related_order"`
		GuestToken   string `json:"guest_token"` // 游客令牌（用于关联多个工单）
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	// 检查是否允许游客
	config, _ := SupportSvc.GetSupportConfig()
	
	var userID uint
	var username string
	
	// 尝试获取登录用户信息
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
		if uname, ok := c.Get("username"); ok {
			username = uname.(string)
		}
	} else if !config.AllowGuest {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "请先登录后再提交工单"})
		return
	}
	
	if req.Priority == 0 {
		req.Priority = model.TicketPriorityNormal
	}
	
	ticket, err := SupportSvc.CreateTicket(userID, username, req.Email, req.Subject, req.Category, req.Content, req.RelatedOrder, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "创建工单失败"})
		return
	}
	
	// 自动分配工单
	SupportSvc.AutoAssignTicket(ticket.ID)
	
	// 通知在线客服（邮件）
	SupportSvc.NotifyStaffOnNewTicket(ticket)
	
	// WebSocket 通知所有客服有新工单
	NotifyNewTicket(ticket)
	
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"ticket_no":   ticket.TicketNo,
		"guest_token": ticket.GuestToken,
	})
}

// GetUserTickets 获取用户工单列表
func GetUserTickets(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	tickets, total, err := SupportSvc.GetUserTickets(userID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "获取工单失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tickets": tickets,
		"total":   total,
		"page":    page,
	})
}

// GetGuestTickets 游客获取工单列表
func GetGuestTickets(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	guestToken := c.Query("guest_token")
	if guestToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "缺少游客令牌"})
		return
	}
	
	tickets, err := SupportSvc.GetGuestTickets(guestToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "获取工单失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "tickets": tickets})
}

// GetTicketDetail 获取工单详情
func GetTicketDetail(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	guestToken := c.Query("guest_token")
	
	var ticket *model.SupportTicket
	var err error
	
	// 检查是否登录用户
	if userID, exists := c.Get("user_id"); exists {
		ticket, err = SupportSvc.GetTicketByNo(ticketNo)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "工单不存在"})
			return
		}
		// 验证工单归属
		if ticket.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权访问此工单"})
			return
		}
	} else if guestToken != "" {
		// 游客访问
		ticket, err = SupportSvc.GetTicketByGuestToken(ticketNo, guestToken)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "工单不存在或令牌无效"})
			return
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "请登录或提供游客令牌"})
		return
	}
	
	// 获取消息
	messages, _ := SupportSvc.GetTicketMessages(ticket.ID, false)
	
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"ticket":   ticket,
		"messages": messages,
	})
}

// ReplyTicket 用户回复工单
func ReplyTicket(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	
	var req struct {
		Content    string `json:"content" binding:"required"`
		GuestToken string `json:"guest_token"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	var ticket *model.SupportTicket
	var err error
	var senderType string
	var senderID uint
	var senderName string
	
	// 检查是否登录用户
	if uid, exists := c.Get("user_id"); exists {
		senderID = uid.(uint)
		if uname, ok := c.Get("username"); ok {
			senderName = uname.(string)
		}
		senderType = "user"
		
		ticket, err = SupportSvc.GetTicketByNo(ticketNo)
		if err != nil || ticket.UserID != senderID {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权操作此工单"})
			return
		}
	} else if req.GuestToken != "" {
		senderType = "guest"
		ticket, err = SupportSvc.GetTicketByGuestToken(ticketNo, req.GuestToken)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "工单不存在或令牌无效"})
			return
		}
		senderName = ticket.Username
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "请登录或提供游客令牌"})
		return
	}
	
	// 检查工单状态
	if ticket.Status == model.TicketStatusClosed {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "工单已关闭，无法回复"})
		return
	}
	
	msg, err := SupportSvc.ReplyTicket(ticket.ID, senderType, senderID, senderName, req.Content, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "回复失败"})
		return
	}
	
	// WebSocket 通知工单订阅者有新消息
	NotifyTicketMessage(ticket.ID, msg)
	
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}

// CloseTicket 用户关闭工单
func CloseTicket(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	
	var req struct {
		GuestToken string `json:"guest_token"`
	}
	c.ShouldBindJSON(&req)
	
	var ticket *model.SupportTicket
	var err error
	var operatorName string
	
	if uid, exists := c.Get("user_id"); exists {
		ticket, err = SupportSvc.GetTicketByNo(ticketNo)
		if err != nil || ticket.UserID != uid.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权操作此工单"})
			return
		}
		if uname, ok := c.Get("username"); ok {
			operatorName = uname.(string)
		}
	} else if req.GuestToken != "" {
		ticket, err = SupportSvc.GetTicketByGuestToken(ticketNo, req.GuestToken)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "工单不存在或令牌无效"})
			return
		}
		operatorName = ticket.Username
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "请登录或提供游客令牌"})
		return
	}
	
	if err := SupportSvc.UpdateTicketStatus(ticket.ID, model.TicketStatusClosed, operatorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "关闭失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ==========================================
//         实时聊天 API
// ==========================================

// StartLiveChat 开始实时聊天
func StartLiveChat(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	config, _ := SupportSvc.GetSupportConfig()
	
	var userID uint
	var username string
	var guestToken string
	
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
		if uname, ok := c.Get("username"); ok {
			username = uname.(string)
		}
	} else if !config.AllowGuest {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "请先登录"})
		return
	} else {
		// 尝试从请求获取游客令牌
		var req struct {
			GuestToken string `json:"guest_token"`
		}
		c.ShouldBindJSON(&req)
		guestToken = req.GuestToken
	}
	
	chat, err := SupportSvc.CreateLiveChat(userID, username, guestToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "创建聊天失败"})
		return
	}
	
	// WebSocket 通知所有客服有新聊天
	NotifyNewChat(chat)
	
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"session_id":  chat.SessionID,
		"guest_token": chat.GuestToken,
		"welcome":     config.WelcomeMessage,
	})
}

// SendChatMessage 发送聊天消息
func SendChatMessage(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	sessionID := c.Param("session_id")
	
	var req struct {
		Content    string `json:"content" binding:"required"`
		MsgType    string `json:"msg_type"`
		GuestToken string `json:"guest_token"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	chat, err := SupportSvc.GetLiveChatBySession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "聊天会话不存在"})
		return
	}
	
	// 验证权限
	var senderType string
	var senderID uint
	var senderName string
	
	if uid, exists := c.Get("user_id"); exists {
		if chat.UserID != uid.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权访问此聊天"})
			return
		}
		senderID = uid.(uint)
		if uname, ok := c.Get("username"); ok {
			senderName = uname.(string)
		}
		senderType = "user"
	} else if req.GuestToken != "" && chat.GuestToken == req.GuestToken {
		senderType = "guest"
		senderName = chat.Username
	} else {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权访问此聊天"})
		return
	}
	
	if req.MsgType == "" {
		req.MsgType = "text"
	}
	
	msg, err := SupportSvc.SendChatMessage(chat.ID, senderType, senderID, senderName, req.Content, req.MsgType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "发送失败"})
		return
	}
	
	// WebSocket 通知聊天订阅者有新消息
	NotifyChatMessage(chat.ID, msg)
	
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}

// GetChatMessages 获取聊天消息
func GetChatMessages(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	sessionID := c.Param("session_id")
	afterID, _ := strconv.ParseUint(c.Query("after_id"), 10, 64)
	guestToken := c.Query("guest_token")
	
	chat, err := SupportSvc.GetLiveChatBySession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "聊天会话不存在"})
		return
	}
	
	// 验证权限
	if uid, exists := c.Get("user_id"); exists {
		if chat.UserID != uid.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权访问此聊天"})
			return
		}
	} else if guestToken == "" || chat.GuestToken != guestToken {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权访问此聊天"})
		return
	}
	
	messages, err := SupportSvc.GetChatMessages(chat.ID, uint(afterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "获取消息失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"messages": messages,
		"chat":     chat,
	})
}

// EndLiveChat 结束聊天
func EndLiveChat(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	sessionID := c.Param("session_id")
	guestToken := c.Query("guest_token")
	
	chat, err := SupportSvc.GetLiveChatBySession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "聊天会话不存在"})
		return
	}
	
	// 验证权限
	if uid, exists := c.Get("user_id"); exists {
		if chat.UserID != uid.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权操作"})
			return
		}
	} else if guestToken == "" || chat.GuestToken != guestToken {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权操作"})
		return
	}
	
	if err := SupportSvc.EndChat(chat.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "结束聊天失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ==========================================
//         文件上传 API
// ==========================================

// UploadTicketAttachment 上传工单附件
func UploadTicketAttachment(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	guestToken := c.PostForm("guest_token")
	
	var ticket *model.SupportTicket
	var err error
	var senderType string
	var senderID uint
	var senderName string
	
	// 验证权限
	if uid, exists := c.Get("user_id"); exists {
		senderID = uid.(uint)
		if uname, ok := c.Get("username"); ok {
			senderName = uname.(string)
		}
		senderType = "user"
		
		ticket, err = SupportSvc.GetTicketByNo(ticketNo)
		if err != nil || ticket.UserID != senderID {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权操作此工单"})
			return
		}
	} else if guestToken != "" {
		senderType = "guest"
		ticket, err = SupportSvc.GetTicketByGuestToken(ticketNo, guestToken)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "工单不存在或令牌无效"})
			return
		}
		senderName = ticket.Username
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "请登录或提供游客令牌"})
		return
	}
	
	// 检查工单状态
	if ticket.Status == model.TicketStatusClosed || ticket.Status == model.TicketStatusMerged {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "工单已关闭，无法上传附件"})
		return
	}
	
	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请选择要上传的文件"})
		return
	}
	
	// 获取配置检查文件大小限制
	config, _ := SupportSvc.GetSupportConfig()
	maxSize := int64(config.MaxAttachmentSize) * 1024 * 1024 // MB转字节
	if maxSize == 0 {
		maxSize = 5 * 1024 * 1024 // 默认5MB
	}
	
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "文件大小超过限制"})
		return
	}
	
	// 生成存储路径
	uploadDir := "./uploads/tickets/" + ticketNo
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	filePath := uploadDir + "/" + fileName
	
	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "文件保存失败"})
		return
	}
	
	// 生成访问URL
	fileURL := "/uploads/tickets/" + ticketNo + "/" + fileName
	
	// 创建带附件的消息
	content := c.PostForm("content")
	if content == "" {
		content = "[附件]"
	}
	
	msg, err := SupportSvc.ReplyTicketWithAttachment(ticket.ID, senderType, senderID, senderName, content, false, fileURL, file.Filename, file.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存消息失败"})
		return
	}
	
	// 保存附件记录
	SupportSvc.SaveAttachment(ticket.ID, msg.ID, file.Filename, filePath, file.Header.Get("Content-Type"), file.Size)
	
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  msg,
		"file_url": fileURL,
	})
}

// StaffUploadTicketAttachment 客服上传工单附件
func StaffUploadTicketAttachment(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	
	ticket, err := SupportSvc.GetTicketByNo(ticketNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "工单不存在"})
		return
	}
	
	staffID := c.GetUint("staff_id")
	staff, _ := SupportSvc.GetStaffByID(staffID)
	staffName := staff.Nickname
	if staffName == "" {
		staffName = staff.Username
	}
	
	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请选择要上传的文件"})
		return
	}
	
	// 获取配置检查文件大小限制
	config, _ := SupportSvc.GetSupportConfig()
	maxSize := int64(config.MaxAttachmentSize) * 1024 * 1024
	if maxSize == 0 {
		maxSize = 5 * 1024 * 1024
	}
	
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "文件大小超过限制"})
		return
	}
	
	// 生成存储路径
	uploadDir := "./uploads/tickets/" + ticketNo
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	filePath := uploadDir + "/" + fileName
	
	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "文件保存失败"})
		return
	}
	
	fileURL := "/uploads/tickets/" + ticketNo + "/" + fileName
	
	content := c.PostForm("content")
	if content == "" {
		content = "[附件]"
	}
	isInternal := c.PostForm("is_internal") == "true"
	
	msg, err := SupportSvc.ReplyTicketWithAttachment(ticket.ID, "staff", staffID, staffName, content, isInternal, fileURL, file.Filename, file.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存消息失败"})
		return
	}
	
	SupportSvc.SaveAttachment(ticket.ID, msg.ID, file.Filename, filePath, file.Header.Get("Content-Type"), file.Size)
	
	// 非内部备注时通知用户
	if !isInternal {
		SupportSvc.NotifyUserOnReply(ticket.ID, content)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  msg,
		"file_url": fileURL,
	})
}

// GetSupportConfig 获取客服配置（公开）
func GetSupportConfig(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusOK, gin.H{
			"success":         true,
			"enabled":         false,
			"allow_guest":     false,
			"welcome":         "",
			"offline":         "",
			"categories":      []string{},
			"online_count":    0,
			"is_online":       false,
			"is_working_time": false,
		})
		return
	}

	config, _ := SupportSvc.GetSupportConfig()
	
	// 检查是否有在线客服
	onlineStaff, _ := SupportSvc.GetOnlineStaff()
	
	// 检查是否在工作时间
	isWorkingTime := SupportSvc.IsWorkingTime()
	
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"enabled":         config.Enabled,
		"allow_guest":     config.AllowGuest,
		"welcome":         config.WelcomeMessage,
		"offline":         config.OfflineMessage,
		"categories":      config.TicketCategories,
		"online_count":    len(onlineStaff),
		"is_online":       len(onlineStaff) > 0,
		"is_working_time": isWorkingTime,
		"working_hours":   gin.H{
			"start": config.WorkingHoursStart,
			"end":   config.WorkingHoursEnd,
			"days":  config.WorkingDays,
		},
	})
}
