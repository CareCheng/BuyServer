package model

import (
	"time"

	"gorm.io/gorm"
)

// ==========================================
//         客服支持系统模型
// ==========================================

// SupportTicket 工单模型
type SupportTicket struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TicketNo      string         `gorm:"type:varchar(32);uniqueIndex" json:"ticket_no"` // 工单编号
	UserID        uint           `gorm:"index" json:"user_id"`                          // 用户ID（0表示游客）
	Username      string         `gorm:"type:varchar(100)" json:"username"`             // 用户名或游客标识
	Email         string         `gorm:"type:varchar(255)" json:"email"`                // 联系邮箱
	Subject       string         `gorm:"type:varchar(200)" json:"subject"`              // 工单主题
	Category      string         `gorm:"type:varchar(50)" json:"category"`              // 分类：order/product/payment/account/other
	Priority      int            `gorm:"default:1" json:"priority"`                     // 优先级：1普通 2紧急 3非常紧急
	Status        int            `gorm:"default:0" json:"status"`                       // 状态：0待处理 1处理中 2已回复 3已解决 4已关闭 5已合并
	AssignedTo    uint           `gorm:"default:0" json:"assigned_to"`                  // 分配给客服ID
	AssignedName  string         `gorm:"type:varchar(100)" json:"assigned_name"`        // 客服名称
	RelatedOrder  string         `gorm:"type:varchar(64)" json:"related_order"`         // 关联订单号
	GuestToken    string         `gorm:"type:varchar(64);index" json:"guest_token"`     // 游客访问令牌
	MergedTo      uint           `gorm:"default:0" json:"merged_to"`                    // 合并到的工单ID
	MergedFrom    string         `gorm:"type:text" json:"merged_from"`                  // 被合并的工单ID列表（JSON数组）
	TransferCount int            `gorm:"default:0" json:"transfer_count"`               // 转接次数
	TransferLog   string         `gorm:"type:text" json:"transfer_log"`                 // 转接记录（JSON数组）
	Rating        int            `gorm:"default:0" json:"rating"`                       // 满意度评分 1-5星
	RatingComment string         `gorm:"type:text" json:"rating_comment"`               // 评价内容
	RatedAt       *time.Time     `json:"rated_at"`                                      // 评价时间
	LastReplyAt   *time.Time     `json:"last_reply_at"`                                 // 最后回复时间
	LastReplyBy   string         `gorm:"type:varchar(100)" json:"last_reply_by"`        // 最后回复人
	ClosedAt      *time.Time     `json:"closed_at"`                                     // 关闭时间
	ClosedBy      string         `gorm:"type:varchar(100)" json:"closed_by"`            // 关闭人
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// SupportMessage 工单消息/聊天消息
type SupportMessage struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	TicketID   uint       `gorm:"index" json:"ticket_id"`                     // 关联工单ID
	SenderType string     `gorm:"type:varchar(20)" json:"sender_type"`        // 发送者类型：user/guest/staff/system
	SenderID   uint       `gorm:"default:0" json:"sender_id"`                 // 发送者ID
	SenderName string     `gorm:"type:varchar(100)" json:"sender_name"`       // 发送者名称
	Content    string     `gorm:"type:text" json:"content"`                   // 消息内容
	MsgType    string     `gorm:"type:varchar(20);default:'text'" json:"msg_type"` // 消息类型：text/image/file
	FileURL    string     `gorm:"type:varchar(500)" json:"file_url"`          // 附件URL
	FileName   string     `gorm:"type:varchar(255)" json:"file_name"`         // 附件文件名
	FileSize   int64      `gorm:"default:0" json:"file_size"`                 // 附件大小（字节）
	IsInternal bool       `gorm:"default:false" json:"is_internal"`           // 是否内部备注（用户不可见）
	ReadAt     *time.Time `json:"read_at"`                                    // 已读时间
	CreatedAt  time.Time  `json:"created_at"`
}

// SupportAttachment 工单附件
type SupportAttachment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TicketID  uint      `gorm:"index" json:"ticket_id"`                // 关联工单ID
	MessageID uint      `gorm:"index" json:"message_id"`               // 关联消息ID
	FileName  string    `gorm:"type:varchar(255)" json:"file_name"`    // 原始文件名
	FilePath  string    `gorm:"type:varchar(500)" json:"file_path"`    // 存储路径
	FileSize  int64     `gorm:"default:0" json:"file_size"`            // 文件大小（字节）
	MimeType  string    `gorm:"type:varchar(100)" json:"mime_type"`    // MIME类型
	CreatedAt time.Time `json:"created_at"`
}

func (SupportAttachment) TableName() string {
	return "support_attachments"
}

// SupportStaff 客服人员
type SupportStaff struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"type:varchar(100);uniqueIndex" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255)" json:"-"`
	Nickname     string         `gorm:"type:varchar(100)" json:"nickname"`    // 显示名称
	Avatar       string         `gorm:"type:varchar(500)" json:"avatar"`      // 头像URL
	Email        string         `gorm:"type:varchar(255)" json:"email"`
	Role         string         `gorm:"type:varchar(50);default:'staff'" json:"role"` // staff/supervisor
	Status       int            `gorm:"default:1" json:"status"`              // 1在线 0离线 -1禁用
	MaxTickets   int            `gorm:"default:10" json:"max_tickets"`        // 最大同时处理工单数
	CurrentLoad  int            `gorm:"default:0" json:"current_load"`        // 当前处理工单数
	Enable2FA    bool           `gorm:"default:false" json:"enable_2fa"`      // 是否启用二步验证
	TOTPSecret   string         `gorm:"type:varchar(64)" json:"-"`            // TOTP密钥
	LastActiveAt *time.Time     `json:"last_active_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// SupportStaffSession 客服会话
type SupportStaffSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"type:varchar(64);uniqueIndex" json:"session_id"`
	StaffID   uint      `gorm:"index" json:"staff_id"`
	Username  string    `gorm:"type:varchar(100)" json:"username"`
	Role      string    `gorm:"type:varchar(50)" json:"role"`
	Verified  bool      `gorm:"default:false" json:"verified"` // 二步验证是否通过
	IP        string    `gorm:"type:varchar(50)" json:"ip"`
	UserAgent string    `gorm:"type:varchar(500)" json:"user_agent"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SupportConfig 客服系统配置
type SupportConfigDB struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	Enabled             bool      `gorm:"default:true" json:"enabled"`                    // 是否启用客服系统
	AllowGuest          bool      `gorm:"default:true" json:"allow_guest"`                // 是否允许游客咨询
	StaffPortalSuffix   string    `gorm:"type:varchar(50);default:'staff'" json:"staff_portal_suffix"` // 客服后台路径后缀
	EnableStaff2FA      bool      `gorm:"default:false" json:"enable_staff_2fa"`          // 客服登录是否启用二步验证
	WorkingHoursStart   string    `gorm:"type:varchar(10)" json:"working_hours_start"`    // 工作时间开始 如 "09:00"
	WorkingHoursEnd     string    `gorm:"type:varchar(10)" json:"working_hours_end"`      // 工作时间结束 如 "18:00"
	WorkingDays         string    `gorm:"type:varchar(50)" json:"working_days"`           // 工作日 如 "1,2,3,4,5"
	OfflineMessage      string    `gorm:"type:text" json:"offline_message"`               // 离线提示消息
	WelcomeMessage      string    `gorm:"type:text" json:"welcome_message"`               // 欢迎消息
	AutoCloseHours      int       `gorm:"default:72" json:"auto_close_hours"`             // 自动关闭无回复工单（小时）
	TicketCategories    string    `gorm:"type:text" json:"ticket_categories"`             // 工单分类（JSON数组）
	EnableAutoAssign    bool      `gorm:"default:false" json:"enable_auto_assign"`        // 是否启用自动分配
	EnableEmailNotify   bool      `gorm:"default:false" json:"enable_email_notify"`       // 是否启用邮件通知
	NotifyOnNewTicket   bool      `gorm:"default:true" json:"notify_on_new_ticket"`       // 新工单时通知客服
	NotifyOnReply       bool      `gorm:"default:true" json:"notify_on_reply"`            // 有新回复时通知用户
	MaxAttachmentSize   int       `gorm:"default:5" json:"max_attachment_size"`           // 最大附件大小（MB）
	AllowedFileTypes    string    `gorm:"type:text" json:"allowed_file_types"`            // 允许的文件类型（JSON数组）
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// LiveChat 实时聊天会话
type LiveChat struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	SessionID   string     `gorm:"type:varchar(64);uniqueIndex" json:"session_id"` // 聊天会话ID
	UserID      uint       `gorm:"index" json:"user_id"`                           // 用户ID（0表示游客）
	Username    string     `gorm:"type:varchar(100)" json:"username"`
	GuestToken  string     `gorm:"type:varchar(64);index" json:"guest_token"`      // 游客令牌
	StaffID     uint       `gorm:"default:0" json:"staff_id"`                      // 接待客服ID
	StaffName   string     `gorm:"type:varchar(100)" json:"staff_name"`
	Status      int        `gorm:"default:0" json:"status"`                        // 0等待接入 1进行中 2已结束
	Rating      int        `gorm:"default:0" json:"rating"`                        // 评分 1-5
	Feedback    string     `gorm:"type:text" json:"feedback"`                      // 评价内容
	EndedAt     *time.Time `json:"ended_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// LiveChatMessage 实时聊天消息
type LiveChatMessage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ChatID     uint      `gorm:"index" json:"chat_id"`                       // 关联聊天会话
	SenderType string    `gorm:"type:varchar(20)" json:"sender_type"`        // user/guest/staff/system
	SenderID   uint      `gorm:"default:0" json:"sender_id"`
	SenderName string    `gorm:"type:varchar(100)" json:"sender_name"`
	Content    string    `gorm:"type:text" json:"content"`
	MsgType    string    `gorm:"type:varchar(20);default:'text'" json:"msg_type"` // text/image/file
	FileURL    string    `gorm:"type:varchar(500)" json:"file_url"`
	ReadAt     *time.Time `json:"read_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// TableName 设置表名
func (SupportTicket) TableName() string {
	return "support_tickets"
}

func (SupportMessage) TableName() string {
	return "support_messages"
}

func (SupportStaff) TableName() string {
	return "support_staff"
}

func (SupportStaffSession) TableName() string {
	return "support_staff_sessions"
}

func (SupportConfigDB) TableName() string {
	return "support_configs"
}

func (LiveChat) TableName() string {
	return "live_chats"
}

func (LiveChatMessage) TableName() string {
	return "live_chat_messages"
}

// ==========================================
//         工单状态常量
// ==========================================

const (
	TicketStatusPending    = 0 // 待处理
	TicketStatusProcessing = 1 // 处理中
	TicketStatusReplied    = 2 // 已回复
	TicketStatusResolved   = 3 // 已解决
	TicketStatusClosed     = 4 // 已关闭
	TicketStatusMerged     = 5 // 已合并
)

const (
	TicketPriorityNormal  = 1 // 普通
	TicketPriorityUrgent  = 2 // 紧急
	TicketPriorityCritical = 3 // 非常紧急
)

const (
	ChatStatusWaiting  = 0 // 等待接入
	ChatStatusActive   = 1 // 进行中
	ChatStatusEnded    = 2 // 已结束
)

// GetTicketStatusText 获取工单状态文本
func GetTicketStatusText(status int) string {
	switch status {
	case TicketStatusPending:
		return "待处理"
	case TicketStatusProcessing:
		return "处理中"
	case TicketStatusReplied:
		return "已回复"
	case TicketStatusResolved:
		return "已解决"
	case TicketStatusClosed:
		return "已关闭"
	case TicketStatusMerged:
		return "已合并"
	default:
		return "未知"
	}
}

// GetTicketPriorityText 获取优先级文本
func GetTicketPriorityText(priority int) string {
	switch priority {
	case TicketPriorityNormal:
		return "普通"
	case TicketPriorityUrgent:
		return "紧急"
	case TicketPriorityCritical:
		return "非常紧急"
	default:
		return "普通"
	}
}
