package service

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ==========================================
//         WebSocket 实时通信服务
// ==========================================

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type      string      `json:"type"`       // 消息类型：message/typing/read/online/offline/system
	TicketID  uint        `json:"ticket_id"`  // 工单ID
	ChatID    uint        `json:"chat_id"`    // 聊天会话ID
	Data      interface{} `json:"data"`       // 消息数据
	Timestamp int64       `json:"timestamp"`  // 时间戳
}

// WSClient WebSocket客户端连接
type WSClient struct {
	ID         string          // 客户端唯一ID
	Conn       *websocket.Conn // WebSocket连接
	UserID     uint            // 用户ID（0表示游客）
	UserType   string          // 用户类型：user/guest/staff
	GuestToken string          // 游客令牌
	StaffID    uint            // 客服ID（仅客服有效）
	Send       chan []byte     // 发送消息通道
	Hub        *WSHub          // 所属Hub
}

// WSHub WebSocket连接管理中心
type WSHub struct {
	// 所有客户端连接
	clients map[*WSClient]bool
	// 用户ID到客户端的映射（用户端）
	userClients map[uint]*WSClient
	// 游客令牌到客户端的映射
	guestClients map[string]*WSClient
	// 客服ID到客户端的映射
	staffClients map[uint]*WSClient
	// 工单订阅：工单ID -> 订阅的客户端列表
	ticketSubscribers map[uint]map[*WSClient]bool
	// 聊天订阅：聊天ID -> 订阅的客户端列表
	chatSubscribers map[uint]map[*WSClient]bool
	// 注册客户端
	register chan *WSClient
	// 注销客户端
	unregister chan *WSClient
	// 广播消息
	broadcast chan *WSBroadcast
	// 互斥锁
	mu sync.RWMutex
}

// WSBroadcast 广播消息结构
type WSBroadcast struct {
	TicketID uint        // 目标工单ID（0表示不限）
	ChatID   uint        // 目标聊天ID（0表示不限）
	StaffAll bool        // 是否广播给所有客服
	Message  *WSMessage  // 消息内容
	Exclude  *WSClient   // 排除的客户端
}

// 全局WebSocket Hub实例
var wsHub *WSHub

// GetWSHub 获取WebSocket Hub实例
func GetWSHub() *WSHub {
	if wsHub == nil {
		wsHub = NewWSHub()
		go wsHub.Run()
	}
	return wsHub
}

// NewWSHub 创建新的WebSocket Hub
func NewWSHub() *WSHub {
	return &WSHub{
		clients:           make(map[*WSClient]bool),
		userClients:       make(map[uint]*WSClient),
		guestClients:      make(map[string]*WSClient),
		staffClients:      make(map[uint]*WSClient),
		ticketSubscribers: make(map[uint]map[*WSClient]bool),
		chatSubscribers:   make(map[uint]map[*WSClient]bool),
		register:          make(chan *WSClient),
		unregister:        make(chan *WSClient),
		broadcast:         make(chan *WSBroadcast),
	}
}

// Run 运行WebSocket Hub
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case broadcast := <-h.broadcast:
			h.broadcastMessage(broadcast)
		}
	}
}

// registerClient 注册客户端
func (h *WSHub) registerClient(client *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// 根据用户类型注册到对应映射
	switch client.UserType {
	case "user":
		if client.UserID > 0 {
			// 如果已有连接，关闭旧连接
			if old, ok := h.userClients[client.UserID]; ok {
				close(old.Send)
				delete(h.clients, old)
			}
			h.userClients[client.UserID] = client
		}
	case "guest":
		if client.GuestToken != "" {
			if old, ok := h.guestClients[client.GuestToken]; ok {
				close(old.Send)
				delete(h.clients, old)
			}
			h.guestClients[client.GuestToken] = client
		}
	case "staff":
		if client.StaffID > 0 {
			if old, ok := h.staffClients[client.StaffID]; ok {
				close(old.Send)
				delete(h.clients, old)
			}
			h.staffClients[client.StaffID] = client
			// 通知其他客服该客服上线
			h.notifyStaffOnline(client.StaffID, true)
		}
	}
}

// unregisterClient 注销客户端
func (h *WSHub) unregisterClient(client *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)

		// 从用户映射中移除
		switch client.UserType {
		case "user":
			if h.userClients[client.UserID] == client {
				delete(h.userClients, client.UserID)
			}
		case "guest":
			if h.guestClients[client.GuestToken] == client {
				delete(h.guestClients, client.GuestToken)
			}
		case "staff":
			if h.staffClients[client.StaffID] == client {
				delete(h.staffClients, client.StaffID)
				// 通知其他客服该客服下线
				h.notifyStaffOnline(client.StaffID, false)
			}
		}

		// 从所有订阅中移除
		for ticketID, subscribers := range h.ticketSubscribers {
			delete(subscribers, client)
			if len(subscribers) == 0 {
				delete(h.ticketSubscribers, ticketID)
			}
		}
		for chatID, subscribers := range h.chatSubscribers {
			delete(subscribers, client)
			if len(subscribers) == 0 {
				delete(h.chatSubscribers, chatID)
			}
		}
	}
}

// broadcastMessage 广播消息
func (h *WSHub) broadcastMessage(broadcast *WSBroadcast) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(broadcast.Message)
	if err != nil {
		return
	}

	// 广播给所有客服
	if broadcast.StaffAll {
		for _, client := range h.staffClients {
			if client != broadcast.Exclude {
				select {
				case client.Send <- data:
				default:
					// 发送失败，跳过
				}
			}
		}
		return
	}

	// 广播给工单订阅者
	if broadcast.TicketID > 0 {
		if subscribers, ok := h.ticketSubscribers[broadcast.TicketID]; ok {
			for client := range subscribers {
				if client != broadcast.Exclude {
					select {
					case client.Send <- data:
					default:
					}
				}
			}
		}
		return
	}

	// 广播给聊天订阅者
	if broadcast.ChatID > 0 {
		if subscribers, ok := h.chatSubscribers[broadcast.ChatID]; ok {
			for client := range subscribers {
				if client != broadcast.Exclude {
					select {
					case client.Send <- data:
					default:
					}
				}
			}
		}
		return
	}
}

// notifyStaffOnline 通知客服上下线
func (h *WSHub) notifyStaffOnline(staffID uint, online bool) {
	msgType := "staff_offline"
	if online {
		msgType = "staff_online"
	}

	msg := &WSMessage{
		Type:      msgType,
		Data:      map[string]interface{}{"staff_id": staffID},
		Timestamp: time.Now().Unix(),
	}

	data, _ := json.Marshal(msg)
	for id, client := range h.staffClients {
		if id != staffID {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

// SubscribeTicket 订阅工单消息
func (h *WSHub) SubscribeTicket(client *WSClient, ticketID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.ticketSubscribers[ticketID] == nil {
		h.ticketSubscribers[ticketID] = make(map[*WSClient]bool)
	}
	h.ticketSubscribers[ticketID][client] = true
}

// UnsubscribeTicket 取消订阅工单消息
func (h *WSHub) UnsubscribeTicket(client *WSClient, ticketID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subscribers, ok := h.ticketSubscribers[ticketID]; ok {
		delete(subscribers, client)
		if len(subscribers) == 0 {
			delete(h.ticketSubscribers, ticketID)
		}
	}
}

// SubscribeChat 订阅聊天消息
func (h *WSHub) SubscribeChat(client *WSClient, chatID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.chatSubscribers[chatID] == nil {
		h.chatSubscribers[chatID] = make(map[*WSClient]bool)
	}
	h.chatSubscribers[chatID][client] = true
}

// UnsubscribeChat 取消订阅聊天消息
func (h *WSHub) UnsubscribeChat(client *WSClient, chatID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subscribers, ok := h.chatSubscribers[chatID]; ok {
		delete(subscribers, client)
		if len(subscribers) == 0 {
			delete(h.chatSubscribers, chatID)
		}
	}
}

// SendToUser 发送消息给指定用户
func (h *WSHub) SendToUser(userID uint, msg *WSMessage) {
	h.mu.RLock()
	client, ok := h.userClients[userID]
	h.mu.RUnlock()

	if ok {
		data, _ := json.Marshal(msg)
		select {
		case client.Send <- data:
		default:
		}
	}
}

// SendToGuest 发送消息给指定游客
func (h *WSHub) SendToGuest(guestToken string, msg *WSMessage) {
	h.mu.RLock()
	client, ok := h.guestClients[guestToken]
	h.mu.RUnlock()

	if ok {
		data, _ := json.Marshal(msg)
		select {
		case client.Send <- data:
		default:
		}
	}
}

// SendToStaff 发送消息给指定客服
func (h *WSHub) SendToStaff(staffID uint, msg *WSMessage) {
	h.mu.RLock()
	client, ok := h.staffClients[staffID]
	h.mu.RUnlock()

	if ok {
		data, _ := json.Marshal(msg)
		select {
		case client.Send <- data:
		default:
		}
	}
}

// BroadcastToTicket 广播消息到工单
func (h *WSHub) BroadcastToTicket(ticketID uint, msg *WSMessage, exclude *WSClient) {
	h.broadcast <- &WSBroadcast{
		TicketID: ticketID,
		Message:  msg,
		Exclude:  exclude,
	}
}

// BroadcastToChat 广播消息到聊天
func (h *WSHub) BroadcastToChat(chatID uint, msg *WSMessage, exclude *WSClient) {
	h.broadcast <- &WSBroadcast{
		ChatID:  chatID,
		Message: msg,
		Exclude: exclude,
	}
}

// BroadcastToAllStaff 广播消息给所有客服
func (h *WSHub) BroadcastToAllStaff(msg *WSMessage) {
	h.broadcast <- &WSBroadcast{
		StaffAll: true,
		Message:  msg,
	}
}

// GetOnlineStaffCount 获取在线客服数量
func (h *WSHub) GetOnlineStaffCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.staffClients)
}

// GetOnlineStaffIDs 获取在线客服ID列表
func (h *WSHub) GetOnlineStaffIDs() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]uint, 0, len(h.staffClients))
	for id := range h.staffClients {
		ids = append(ids, id)
	}
	return ids
}

// Register 注册客户端（公开方法）
func (h *WSHub) Register(client *WSClient) {
	h.register <- client
}

// Unregister 注销客户端（公开方法）
func (h *WSHub) Unregister(client *WSClient) {
	h.unregister <- client
}
