package api

import (
	"net/http"
	"strconv"
	"time"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==========================================
//         客服后台 API
// ==========================================

// StaffLogin 客服登录
func StaffLogin(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	staff, sessionID, needs2FA, err := SupportSvc.StaffLogin(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	// 设置 Cookie
	c.SetCookie("staff_session", sessionID, 86400, "/", "", false, true)
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"needs_2fa": needs2FA,
		"staff": gin.H{
			"id":       staff.ID,
			"username": staff.Username,
			"nickname": staff.Nickname,
			"role":     staff.Role,
		},
	})
}

// StaffVerify2FA 客服二步验证
func StaffVerify2FA(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	sessionID, err := c.Cookie("staff_session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "请先登录"})
		return
	}
	
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	if err := SupportSvc.StaffVerify2FA(sessionID, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StaffGenerate2FASecret 生成客服二步验证密钥
func StaffGenerate2FASecret(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	staffID := c.GetUint("staff_id")
	
	secret, url, err := SupportSvc.StaffGenerate2FASecret(staffID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"secret":  secret,
		"url":     url,
	})
}

// StaffEnable2FA 启用客服二步验证
func StaffEnable2FA(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	staffID := c.GetUint("staff_id")
	
	var req struct {
		Secret string `json:"secret" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	if err := SupportSvc.StaffEnable2FA(staffID, req.Secret, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StaffDisable2FA 禁用客服二步验证
func StaffDisable2FA(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	staffID := c.GetUint("staff_id")
	
	if err := SupportSvc.StaffDisable2FA(staffID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StaffGet2FAStatus 获取客服二步验证状态
func StaffGet2FAStatus(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	staffID := c.GetUint("staff_id")
	
	staff, err := SupportSvc.GetStaffByID(staffID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "客服不存在"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"enabled":    staff.Enable2FA,
		"has_secret": staff.TOTPSecret != "",
	})
}

// StaffLogout 客服登出
func StaffLogout(c *gin.Context) {
	sessionID, _ := c.Cookie("staff_session")
	if sessionID != "" {
		SupportSvc.StaffLogout(sessionID)
	}
	c.SetCookie("staff_session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StaffInfo 获取客服信息
func StaffInfo(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	staffID := c.GetUint("staff_id")
	staff, err := SupportSvc.GetStaffByID(staffID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "客服不存在"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"staff": gin.H{
			"id":            staff.ID,
			"username":      staff.Username,
			"nickname":      staff.Nickname,
			"email":         staff.Email,
			"role":          staff.Role,
			"status":        staff.Status,
			"max_tickets":   staff.MaxTickets,
			"current_load":  staff.CurrentLoad,
			"last_active":   staff.LastActiveAt,
		},
	})
}

// StaffAuthRequired 客服认证中间件
func StaffAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if SupportSvc == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
			c.Abort()
			return
		}

		sessionID, err := c.Cookie("staff_session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "请先登录"})
			c.Abort()
			return
		}
		
		session, err := SupportSvc.GetStaffSession(sessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "登录已过期"})
			c.Abort()
			return
		}
		
		// 检查二步验证状态
		if !session.Verified {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "请完成二步验证", "needs_2fa": true})
			c.Abort()
			return
		}
		
		c.Set("staff_id", session.StaffID)
		c.Set("staff_username", session.Username)
		c.Set("staff_role", session.Role)
		c.Next()
	}
}

// ==========================================
//         客服工单管理
// ==========================================

// StaffGetTickets 客服获取工单列表
func StaffGetTickets(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	priority, _ := strconv.Atoi(c.DefaultQuery("priority", "0"))
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	myOnly := c.Query("my_only") == "true"
	
	var tickets []model.SupportTicket
	var total int64
	var err error
	
	if myOnly {
		staffID := c.GetUint("staff_id")
		tickets, total, err = SupportSvc.GetStaffTickets(staffID, status, page, pageSize)
	} else {
		tickets, total, err = SupportSvc.GetAllTickets(status, priority, category, page, pageSize)
	}
	
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

// StaffGetTicketDetail 客服获取工单详情
func StaffGetTicketDetail(c *gin.Context) {
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
	
	// 获取消息（包含内部备注）
	messages, _ := SupportSvc.GetTicketMessages(ticket.ID, true)
	
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"ticket":   ticket,
		"messages": messages,
	})
}

// StaffReplyTicket 客服回复工单
func StaffReplyTicket(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	
	var req struct {
		Content    string `json:"content" binding:"required"`
		IsInternal bool   `json:"is_internal"` // 是否内部备注
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
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
	
	msg, err := SupportSvc.ReplyTicket(ticket.ID, "staff", staffID, staffName, req.Content, req.IsInternal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "回复失败"})
		return
	}
	
	// 非内部备注时通知用户
	if !req.IsInternal {
		SupportSvc.NotifyUserOnReply(ticket.ID, req.Content)
		// WebSocket 通知工单订阅者有新消息
		NotifyTicketMessage(ticket.ID, msg)
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}

// StaffUpdateTicketStatus 客服更新工单状态
func StaffUpdateTicketStatus(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	ticket, err := SupportSvc.GetTicketByNo(ticketNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "工单不存在"})
		return
	}
	
	staffID := c.GetUint("staff_id")
	staff, _ := SupportSvc.GetStaffByID(staffID)
	
	if err := SupportSvc.UpdateTicketStatus(ticket.ID, req.Status, staff.Nickname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "更新失败"})
		return
	}
	
	// WebSocket 通知工单状态变更
	NotifyTicketStatusChange(ticket.ID, req.Status, staff.Nickname)
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StaffAssignTicket 分配工单
func StaffAssignTicket(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	
	var req struct {
		StaffID uint `json:"staff_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	ticket, err := SupportSvc.GetTicketByNo(ticketNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "工单不存在"})
		return
	}
	
	staff, err := SupportSvc.GetStaffByID(req.StaffID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "客服不存在"})
		return
	}
	
	if err := SupportSvc.AssignTicket(ticket.ID, staff.ID, staff.Nickname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "分配失败"})
		return
	}
	
	// WebSocket 通知工单分配
	NotifyTicketAssigned(ticket.ID, staff.ID, staff.Nickname)
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StaffGetTicketStats 获取工单统计
func StaffGetTicketStats(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats, err := SupportSvc.GetTicketStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "获取统计失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "stats": stats})
}

// ==========================================
//         客服实时聊天管理
// ==========================================

// StaffGetWaitingChats 获取等待接入的聊天
func StaffGetWaitingChats(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	chats, err := SupportSvc.GetWaitingChats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "获取失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "chats": chats})
}

// StaffAcceptChat 客服接入聊天
func StaffAcceptChat(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	chatID, _ := strconv.ParseUint(c.Param("chat_id"), 10, 64)
	
	staffID := c.GetUint("staff_id")
	staff, _ := SupportSvc.GetStaffByID(staffID)
	
	if err := SupportSvc.AcceptChat(uint(chatID), staffID, staff.Nickname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "接入失败"})
		return
	}
	
	// 发送系统消息
	sysMsg, _ := SupportSvc.SendChatMessage(uint(chatID), "system", 0, "系统", "客服 "+staff.Nickname+" 已接入对话", "text")
	
	// WebSocket 通知聊天已被接入
	NotifyChatAccepted(uint(chatID), staffID, staff.Nickname)
	NotifyChatMessage(uint(chatID), sysMsg)
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StaffSendChatMessage 客服发送聊天消息
func StaffSendChatMessage(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	chatID, _ := strconv.ParseUint(c.Param("chat_id"), 10, 64)
	
	var req struct {
		Content string `json:"content" binding:"required"`
		MsgType string `json:"msg_type"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	staffID := c.GetUint("staff_id")
	staff, _ := SupportSvc.GetStaffByID(staffID)
	
	if req.MsgType == "" {
		req.MsgType = "text"
	}
	
	msg, err := SupportSvc.SendChatMessage(uint(chatID), "staff", staffID, staff.Nickname, req.Content, req.MsgType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "发送失败"})
		return
	}
	
	// WebSocket 通知聊天订阅者有新消息
	NotifyChatMessage(uint(chatID), msg)
	
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}

// StaffGetChatMessages 客服获取聊天消息
func StaffGetChatMessages(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	chatID, _ := strconv.ParseUint(c.Param("chat_id"), 10, 64)
	afterID, _ := strconv.ParseUint(c.Query("after_id"), 10, 64)
	
	messages, err := SupportSvc.GetChatMessages(uint(chatID), uint(afterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "获取消息失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"messages": messages,
		"chat_id":  chatID,
	})
}

// StaffEndChat 客服结束聊天
func StaffEndChat(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	chatID, _ := strconv.ParseUint(c.Param("chat_id"), 10, 64)
	
	staffID := c.GetUint("staff_id")
	staff, _ := SupportSvc.GetStaffByID(staffID)
	
	// 发送系统消息
	sysMsg, _ := SupportSvc.SendChatMessage(uint(chatID), "system", 0, "系统", "客服 "+staff.Nickname+" 已结束对话", "text")
	
	if err := SupportSvc.EndChat(uint(chatID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "结束聊天失败"})
		return
	}
	
	// WebSocket 通知聊天已结束
	NotifyChatMessage(uint(chatID), sysMsg)
	NotifyChatEnded(uint(chatID))
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ==========================================
//         管理后台 - 客服管理
// ==========================================

// AdminGetSupportConfig 获取客服配置
func AdminGetSupportConfig(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	config, _ := SupportSvc.GetSupportConfig()
	c.JSON(http.StatusOK, gin.H{"success": true, "config": config})
}

// AdminSaveSupportConfig 保存客服配置
func AdminSaveSupportConfig(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var config model.SupportConfigDB
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	config.UpdatedAt = time.Now()
	if err := SupportSvc.SaveSupportConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminGetStaffList 获取客服列表
func AdminGetStaffList(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	staff, err := SupportSvc.GetAllStaff()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "获取失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "staff": staff})
}

// AdminCreateStaff 创建客服账号
func AdminCreateStaff(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	if req.Role == "" {
		req.Role = "staff"
	}
	if req.Nickname == "" {
		req.Nickname = req.Username
	}
	
	staff, err := SupportSvc.CreateStaff(req.Username, req.Password, req.Nickname, req.Email, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "staff": staff})
}

// AdminUpdateStaff 更新客服信息
func AdminUpdateStaff(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	
	var req struct {
		Nickname   string `json:"nickname"`
		Email      string `json:"email"`
		MaxTickets int    `json:"max_tickets"`
		Status     int    `json:"status"`
		Password   string `json:"password"` // 可选，不为空则更新密码
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	if err := SupportSvc.UpdateStaff(uint(id), req.Nickname, req.Email, req.MaxTickets, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "更新失败"})
		return
	}
	
	// 更新密码
	if req.Password != "" {
		if err := SupportSvc.UpdateStaffPassword(uint(id), req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "更新密码失败"})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminDeleteStaff 删除客服
func AdminDeleteStaff(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	
	if err := SupportSvc.DeleteStaff(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "删除失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminGetSupportStats 获取客服系统统计
func AdminGetSupportStats(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketStats, _ := SupportSvc.GetTicketStats()
	onlineStaff, _ := SupportSvc.GetOnlineStaff()
	allStaff, _ := SupportSvc.GetAllStaff()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats": gin.H{
			"tickets":      ticketStats,
			"online_staff": len(onlineStaff),
			"total_staff":  len(allStaff),
		},
	})
}

// ==========================================
//         工单转接与合并
// ==========================================

// StaffTransferTicket 客服转接工单
func StaffTransferTicket(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	ticketNo := c.Param("ticket_no")
	
	var req struct {
		ToStaffID uint   `json:"to_staff_id" binding:"required"`
		Reason    string `json:"reason"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	ticket, err := SupportSvc.GetTicketByNo(ticketNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "工单不存在"})
		return
	}
	
	fromStaffID := c.GetUint("staff_id")
	
	if err := SupportSvc.TransferTicket(ticket.ID, fromStaffID, req.ToStaffID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StaffMergeTickets 客服合并工单
func StaffMergeTickets(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		TargetTicketNo  string   `json:"target_ticket_no" binding:"required"`
		SourceTicketNos []string `json:"source_ticket_nos" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	if len(req.SourceTicketNos) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请选择要合并的工单"})
		return
	}
	
	// 获取目标工单
	targetTicket, err := SupportSvc.GetTicketByNo(req.TargetTicketNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "目标工单不存在"})
		return
	}
	
	// 获取源工单ID列表
	var sourceIDs []uint
	for _, ticketNo := range req.SourceTicketNos {
		if ticketNo == req.TargetTicketNo {
			continue
		}
		ticket, err := SupportSvc.GetTicketByNo(ticketNo)
		if err != nil {
			continue
		}
		sourceIDs = append(sourceIDs, ticket.ID)
	}
	
	if len(sourceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "没有有效的源工单"})
		return
	}
	
	staffID := c.GetUint("staff_id")
	staff, _ := SupportSvc.GetStaffByID(staffID)
	operatorName := staff.Nickname
	if operatorName == "" {
		operatorName = staff.Username
	}
	
	if err := SupportSvc.MergeTickets(targetTicket.ID, sourceIDs, operatorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "merged_count": len(sourceIDs)})
}

// StaffGetOnlineStaff 获取在线客服列表（用于转接选择）
func StaffGetOnlineStaff(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	staff, err := SupportSvc.GetOnlineStaff()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "获取失败"})
		return
	}
	
	// 排除当前客服
	currentStaffID := c.GetUint("staff_id")
	var result []gin.H
	for _, s := range staff {
		if s.ID != currentStaffID {
			result = append(result, gin.H{
				"id":           s.ID,
				"nickname":     s.Nickname,
				"username":     s.Username,
				"current_load": s.CurrentLoad,
				"max_tickets":  s.MaxTickets,
			})
		}
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "staff": result})
}

// StaffGetTicketAttachments 获取工单附件列表
func StaffGetTicketAttachments(c *gin.Context) {
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
	
	attachments, err := SupportSvc.GetTicketAttachments(ticket.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "获取附件失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "attachments": attachments})
}
