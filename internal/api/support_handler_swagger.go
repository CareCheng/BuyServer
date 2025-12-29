package api

// ==================== 客服工单 API 文档注解 ====================

// CreateTicket 创建工单
// @Summary      创建客服工单
// @Description  创建新的客服工单
// @Tags         客服工单
// @Accept       json
// @Produce      json
// @Param        request body SwaggerCreateTicketRequest true "工单信息"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerSupportTicket} "创建成功"
// @Failure      400 {object} SwaggerErrorResponse "参数错误"
// @Security     CookieAuth
// @Router       /api/support/tickets [post]

// GetUserTickets 获取用户工单列表
// @Summary      获取我的工单
// @Description  获取当前用户的工单列表
// @Tags         客服工单
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        status query string false "状态筛选(open/in_progress/closed)"
// @Success      200 {object} SwaggerPagedResponse{data=[]SwaggerSupportTicket} "工单列表"
// @Security     CookieAuth
// @Router       /api/support/tickets [get]

// GetTicketDetail 获取工单详情
// @Summary      获取工单详情
// @Description  获取指定工单的详细信息和对话记录
// @Tags         客服工单
// @Produce      json
// @Param        ticket_no path string true "工单号"
// @Success      200 {object} SwaggerSuccessResponse{data=object{ticket=SwaggerSupportTicket,messages=[]object{id=int,content=string,sender_type=string,created_at=string}}} "工单详情"
// @Failure      400 {object} SwaggerErrorResponse "工单不存在"
// @Security     CookieAuth
// @Router       /api/support/tickets/{ticket_no} [get]

// ReplyTicket 回复工单
// @Summary      回复工单
// @Description  向工单添加回复消息
// @Tags         客服工单
// @Accept       json
// @Produce      json
// @Param        ticket_no path string true "工单号"
// @Param        request body object{content=string} true "回复内容"
// @Success      200 {object} SwaggerSuccessResponse "回复成功"
// @Failure      400 {object} SwaggerErrorResponse "工单不存在或已关闭"
// @Security     CookieAuth
// @Router       /api/support/tickets/{ticket_no}/reply [post]

// CloseTicket 关闭工单
// @Summary      关闭工单
// @Description  用户主动关闭工单
// @Tags         客服工单
// @Produce      json
// @Param        ticket_no path string true "工单号"
// @Success      200 {object} SwaggerSuccessResponse "关闭成功"
// @Failure      400 {object} SwaggerErrorResponse "工单不存在或已关闭"
// @Security     CookieAuth
// @Router       /api/support/tickets/{ticket_no}/close [post]

// RateTicket 评价工单
// @Summary      评价客服
// @Description  对已关闭的工单进行评价
// @Tags         客服工单
// @Accept       json
// @Produce      json
// @Param        ticket_no path string true "工单号"
// @Param        request body object{rating=int,comment=string} true "评价信息(rating: 1-5)"
// @Success      200 {object} SwaggerSuccessResponse "评价成功"
// @Failure      400 {object} SwaggerErrorResponse "工单未关闭或已评价"
// @Security     CookieAuth
// @Router       /api/support/tickets/{ticket_no}/rate [post]

// ==================== FAQ 常见问题 API 文档注解 ====================

// GetFAQCategories 获取FAQ分类
// @Summary      获取FAQ分类
// @Description  获取所有FAQ分类列表
// @Tags         FAQ
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=[]SwaggerFAQCategory} "分类列表"
// @Router       /api/faq/categories [get]

// GetFAQList 获取FAQ列表
// @Summary      获取FAQ列表
// @Description  获取指定分类下的FAQ列表
// @Tags         FAQ
// @Produce      json
// @Param        category_id query int false "分类ID"
// @Param        keyword query string false "搜索关键词"
// @Success      200 {object} SwaggerSuccessResponse{data=[]SwaggerFAQ} "FAQ列表"
// @Router       /api/faq [get]

// GetFAQDetail 获取FAQ详情
// @Summary      获取FAQ详情
// @Description  获取指定FAQ的详细内容
// @Tags         FAQ
// @Produce      json
// @Param        id path int true "FAQ ID"
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerFAQ} "FAQ详情"
// @Failure      400 {object} SwaggerErrorResponse "FAQ不存在"
// @Router       /api/faq/{id} [get]

// MarkFAQHelpful 标记FAQ有帮助
// @Summary      标记FAQ有帮助
// @Description  标记某个FAQ对用户有帮助
// @Tags         FAQ
// @Produce      json
// @Param        id path int true "FAQ ID"
// @Success      200 {object} SwaggerSuccessResponse "标记成功"
// @Router       /api/faq/{id}/helpful [post]

// ==================== 公告 API 文档注解 ====================

// GetAnnouncements 获取公告列表
// @Summary      获取公告列表
// @Description  获取系统公告列表
// @Tags         公告
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(10)
// @Success      200 {object} SwaggerPagedResponse{data=[]object{id=int,title=string,content=string,type=string,created_at=string}} "公告列表"
// @Router       /api/announcements [get]

// GetAnnouncementDetail 获取公告详情
// @Summary      获取公告详情
// @Description  获取指定公告的详细内容
// @Tags         公告
// @Produce      json
// @Param        id path int true "公告ID"
// @Success      200 {object} SwaggerSuccessResponse{data=object{id=int,title=string,content=string,type=string,created_at=string}} "公告详情"
// @Failure      400 {object} SwaggerErrorResponse "公告不存在"
// @Router       /api/announcements/{id} [get]

// ==================== 设备管理 API 文档注解 ====================

// GetLoginDevices 获取登录设备
// @Summary      获取登录设备列表
// @Description  获取当前用户的所有登录设备
// @Tags         设备管理
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=[]object{id=int,device_name=string,ip=string,location=string,last_active=string,is_current=bool}} "设备列表"
// @Security     CookieAuth
// @Router       /api/user/devices [get]

// LogoutDevice 登出设备
// @Summary      登出指定设备
// @Description  使指定设备的登录状态失效
// @Tags         设备管理
// @Produce      json
// @Param        id path int true "设备记录ID"
// @Success      200 {object} SwaggerSuccessResponse "登出成功"
// @Failure      400 {object} SwaggerErrorResponse "设备不存在"
// @Security     CookieAuth
// @Router       /api/user/devices/{id}/logout [post]

// LogoutAllDevices 登出所有设备
// @Summary      登出所有设备
// @Description  使当前用户的所有其他设备登录状态失效
// @Tags         设备管理
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse "登出成功"
// @Security     CookieAuth
// @Router       /api/user/devices/logout-all [post]

// ==================== 通知 API 文档注解 ====================

// GetNotifications 获取通知列表
// @Summary      获取通知列表
// @Description  获取当前用户的通知列表
// @Tags         通知
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        unread_only query bool false "仅显示未读"
// @Success      200 {object} SwaggerPagedResponse{data=[]object{id=int,title=string,content=string,type=string,read=bool,created_at=string}} "通知列表"
// @Security     CookieAuth
// @Router       /api/notifications [get]

// MarkNotificationRead 标记通知已读
// @Summary      标记通知已读
// @Description  标记指定通知为已读状态
// @Tags         通知
// @Produce      json
// @Param        id path int true "通知ID"
// @Success      200 {object} SwaggerSuccessResponse "标记成功"
// @Security     CookieAuth
// @Router       /api/notifications/{id}/read [post]

// MarkAllNotificationsRead 标记全部已读
// @Summary      标记全部通知已读
// @Description  将所有未读通知标记为已读
// @Tags         通知
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse "标记成功"
// @Security     CookieAuth
// @Router       /api/notifications/read-all [post]

// GetUnreadCount 获取未读数量
// @Summary      获取未读通知数量
// @Description  获取当前用户的未读通知数量
// @Tags         通知
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{count=int}} "未读数量"
// @Security     CookieAuth
// @Router       /api/notifications/unread-count [get]

// ==================== 系统配置 API 文档注解 ====================

// GetSystemConfig 获取系统配置
// @Summary      获取系统公开配置
// @Description  获取面向用户的系统配置信息
// @Tags         系统
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{site_name=string,site_logo=string,contact_email=string,payment_methods=[]string}} "系统配置"
// @Router       /api/system/config [get]

// GetCaptcha 获取验证码
// @Summary      获取图形验证码
// @Description  获取图形验证码图片
// @Tags         系统
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{captcha_id=string,captcha_image=string}} "验证码信息"
// @Router       /api/captcha [get]

// VerifyCaptcha 验证验证码
// @Summary      验证图形验证码
// @Description  验证图形验证码是否正确
// @Tags         系统
// @Accept       json
// @Produce      json
// @Param        request body object{captcha_id=string,captcha_code=string} true "验证信息"
// @Success      200 {object} SwaggerSuccessResponse "验证成功"
// @Failure      400 {object} SwaggerErrorResponse "验证码错误"
// @Router       /api/captcha/verify [post]

// ==================== WebSocket API 文档注解 ====================

// WebSocketNotification WebSocket通知连接
// @Summary      WebSocket通知连接
// @Description  建立WebSocket连接接收实时通知（支付状态、工单回复等）
// @Tags         WebSocket
// @Produce      json
// @Security     CookieAuth
// @Router       /api/ws/notifications [get]

// WebSocketChat WebSocket在线客服
// @Summary      WebSocket在线客服
// @Description  建立WebSocket连接进行实时客服对话
// @Tags         WebSocket
// @Produce      json
// @Param        ticket_no query string true "工单号"
// @Security     CookieAuth
// @Router       /api/ws/chat [get]
