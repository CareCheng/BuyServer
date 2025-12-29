package api

import (
	"encoding/json"
	"net/http"
	"time"

	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ==========================================
//         WebSocket 处理器
// ==========================================

// WebSocket升级器配置
var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许跨域（生产环境应该限制）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	// 写入超时
	writeWait = 10 * time.Second
	// Pong等待时间
	pongWait = 60 * time.Second
	// Ping间隔（必须小于pongWait）
	pingPeriod = (pongWait * 9) / 10
	// 最大消息大小
	maxMessageSize = 512 * 1024 // 512KB
)

// WSUserConnect 用户端WebSocket连接
func WSUserConnect(c *gin.Context) {
	// 升级HTTP连接为WebSocket
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 获取用户信息
	var userID uint
	var userType string
	var guestToken string

	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
		userType = "user"
	} else {
		// 游客连接
		guestToken = c.Query("guest_token")
		if guestToken == "" {
			conn.WriteMessage(websocket.CloseMessage, []byte("缺少游客令牌"))
			conn.Close()
			return
		}
		userType = "guest"
	}

	hub := service.GetWSHub()
	client := &service.WSClient{
		ID:         uuid.New().String(),
		Conn:       conn,
		UserID:     userID,
		UserType:   userType,
		GuestToken: guestToken,
		Send:       make(chan []byte, 256),
		Hub:        hub,
	}

	// 注册客户端
	hub.Register(client)

	// 启动读写协程
	go wsWritePump(client)
	go wsReadPump(client)

	// 发送连接成功消息
	welcomeMsg := &service.WSMessage{
		Type:      "connected",
		Data:      map[string]interface{}{"client_id": client.ID},
		Timestamp: time.Now().Unix(),
	}
	data, _ := json.Marshal(welcomeMsg)
	client.Send <- data
}

// WSStaffConnect 客服端WebSocket连接
func WSStaffConnect(c *gin.Context) {
	if SupportSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 验证客服身份
	sessionID, err := c.Cookie("staff_session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "请先登录"})
		return
	}

	session, err := SupportSvc.GetStaffSession(sessionID)
	if err != nil || !session.Verified {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "登录已过期"})
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	hub := service.GetWSHub()
	client := &service.WSClient{
		ID:       uuid.New().String(),
		Conn:     conn,
		UserType: "staff",
		StaffID:  session.StaffID,
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	// 注册客户端
	hub.Register(client)

	// 启动读写协程
	go wsWritePump(client)
	go wsReadPump(client)

	// 发送连接成功消息
	welcomeMsg := &service.WSMessage{
		Type:      "connected",
		Data:      map[string]interface{}{"client_id": client.ID, "staff_id": session.StaffID},
		Timestamp: time.Now().Unix(),
	}
	data, _ := json.Marshal(welcomeMsg)
	client.Send <- data
}

// wsReadPump 读取客户端消息
func wsReadPump(client *service.WSClient) {
	defer func() {
		client.Hub.Unregister(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		// 解析消息
		var msg service.WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// 处理消息
		handleWSMessage(client, &msg)
	}
}

// wsWritePump 向客户端发送消息
func wsWritePump(client *service.WSClient) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub关闭了通道
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送队列中的消息
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleWSMessage 处理WebSocket消息
func handleWSMessage(client *service.WSClient, msg *service.WSMessage) {
	hub := client.Hub

	switch msg.Type {
	case "subscribe_ticket":
		// 订阅工单消息
		if ticketID, ok := msg.Data.(map[string]interface{})["ticket_id"].(float64); ok {
			hub.SubscribeTicket(client, uint(ticketID))
			// 发送订阅成功消息
			response := &service.WSMessage{
				Type:      "subscribed",
				TicketID:  uint(ticketID),
				Timestamp: time.Now().Unix(),
			}
			data, _ := json.Marshal(response)
			client.Send <- data
		}

	case "unsubscribe_ticket":
		// 取消订阅工单消息
		if ticketID, ok := msg.Data.(map[string]interface{})["ticket_id"].(float64); ok {
			hub.UnsubscribeTicket(client, uint(ticketID))
		}

	case "subscribe_chat":
		// 订阅聊天消息
		if chatID, ok := msg.Data.(map[string]interface{})["chat_id"].(float64); ok {
			hub.SubscribeChat(client, uint(chatID))
			response := &service.WSMessage{
				Type:      "subscribed",
				ChatID:   uint(chatID),
				Timestamp: time.Now().Unix(),
			}
			data, _ := json.Marshal(response)
			client.Send <- data
		}

	case "unsubscribe_chat":
		// 取消订阅聊天消息
		if chatID, ok := msg.Data.(map[string]interface{})["chat_id"].(float64); ok {
			hub.UnsubscribeChat(client, uint(chatID))
		}

	case "typing":
		// 正在输入提示
		if msg.TicketID > 0 {
			hub.BroadcastToTicket(msg.TicketID, &service.WSMessage{
				Type:      "typing",
				TicketID:  msg.TicketID,
				Data:      map[string]interface{}{"user_type": client.UserType, "user_id": client.UserID, "staff_id": client.StaffID},
				Timestamp: time.Now().Unix(),
			}, client)
		} else if msg.ChatID > 0 {
			hub.BroadcastToChat(msg.ChatID, &service.WSMessage{
				Type:      "typing",
				ChatID:   msg.ChatID,
				Data:      map[string]interface{}{"user_type": client.UserType, "user_id": client.UserID, "staff_id": client.StaffID},
				Timestamp: time.Now().Unix(),
			}, client)
		}

	case "read":
		// 消息已读
		if msg.TicketID > 0 {
			hub.BroadcastToTicket(msg.TicketID, &service.WSMessage{
				Type:      "read",
				TicketID:  msg.TicketID,
				Data:      map[string]interface{}{"user_type": client.UserType, "user_id": client.UserID, "staff_id": client.StaffID},
				Timestamp: time.Now().Unix(),
			}, client)
		} else if msg.ChatID > 0 {
			hub.BroadcastToChat(msg.ChatID, &service.WSMessage{
				Type:      "read",
				ChatID:   msg.ChatID,
				Data:      map[string]interface{}{"user_type": client.UserType, "user_id": client.UserID, "staff_id": client.StaffID},
				Timestamp: time.Now().Unix(),
			}, client)
		}

	case "ping":
		// 心跳响应
		response := &service.WSMessage{
			Type:      "pong",
			Timestamp: time.Now().Unix(),
		}
		data, _ := json.Marshal(response)
		client.Send <- data
	}
}

// ==========================================
//         WebSocket 消息推送辅助函数
// ==========================================

// NotifyTicketMessage 通知工单新消息
func NotifyTicketMessage(ticketID uint, message interface{}) {
	hub := service.GetWSHub()
	hub.BroadcastToTicket(ticketID, &service.WSMessage{
		Type:      "new_message",
		TicketID:  ticketID,
		Data:      message,
		Timestamp: time.Now().Unix(),
	}, nil)
}

// NotifyChatMessage 通知聊天新消息
func NotifyChatMessage(chatID uint, message interface{}) {
	hub := service.GetWSHub()
	hub.BroadcastToChat(chatID, &service.WSMessage{
		Type:      "new_message",
		ChatID:   chatID,
		Data:      message,
		Timestamp: time.Now().Unix(),
	}, nil)
}

// NotifyTicketStatusChange 通知工单状态变更
func NotifyTicketStatusChange(ticketID uint, status int, operatorName string) {
	hub := service.GetWSHub()
	hub.BroadcastToTicket(ticketID, &service.WSMessage{
		Type:     "status_change",
		TicketID: ticketID,
		Data: map[string]interface{}{
			"status":   status,
			"operator": operatorName,
		},
		Timestamp: time.Now().Unix(),
	}, nil)
}

// NotifyTicketAssigned 通知工单分配
func NotifyTicketAssigned(ticketID uint, staffID uint, staffName string) {
	hub := service.GetWSHub()
	// 通知工单订阅者
	hub.BroadcastToTicket(ticketID, &service.WSMessage{
		Type:     "assigned",
		TicketID: ticketID,
		Data: map[string]interface{}{
			"staff_id":   staffID,
			"staff_name": staffName,
		},
		Timestamp: time.Now().Unix(),
	}, nil)
	// 通知被分配的客服
	hub.SendToStaff(staffID, &service.WSMessage{
		Type:     "new_assignment",
		TicketID: ticketID,
		Data: map[string]interface{}{
			"ticket_id": ticketID,
		},
		Timestamp: time.Now().Unix(),
	})
}

// NotifyNewTicket 通知所有客服有新工单
func NotifyNewTicket(ticket interface{}) {
	hub := service.GetWSHub()
	hub.BroadcastToAllStaff(&service.WSMessage{
		Type:      "new_ticket",
		Data:      ticket,
		Timestamp: time.Now().Unix(),
	})
}

// NotifyNewChat 通知所有客服有新聊天
func NotifyNewChat(chat interface{}) {
	hub := service.GetWSHub()
	hub.BroadcastToAllStaff(&service.WSMessage{
		Type:      "new_chat",
		Data:      chat,
		Timestamp: time.Now().Unix(),
	})
}

// NotifyChatAccepted 通知聊天已被接入
func NotifyChatAccepted(chatID uint, staffID uint, staffName string) {
	hub := service.GetWSHub()
	hub.BroadcastToChat(chatID, &service.WSMessage{
		Type:   "chat_accepted",
		ChatID: chatID,
		Data: map[string]interface{}{
			"staff_id":   staffID,
			"staff_name": staffName,
		},
		Timestamp: time.Now().Unix(),
	}, nil)
}

// NotifyChatEnded 通知聊天已结束
func NotifyChatEnded(chatID uint) {
	hub := service.GetWSHub()
	hub.BroadcastToChat(chatID, &service.WSMessage{
		Type:      "chat_ended",
		ChatID:   chatID,
		Timestamp: time.Now().Unix(),
	}, nil)
}
