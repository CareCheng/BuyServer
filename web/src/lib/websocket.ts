/**
 * WebSocket 实时通信客户端
 * 用于工单和聊天的实时消息推送
 */

// WebSocket 消息类型
export interface WSMessage {
  type: string;
  ticket_id?: number;
  chat_id?: number;
  data?: any;
  timestamp: number;
}

// WebSocket 事件回调类型
type WSEventCallback = (message: WSMessage) => void;

// WebSocket 客户端类
class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string = '';
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 5;
  private reconnectDelay: number = 3000;
  private pingInterval: NodeJS.Timeout | null = null;
  private eventListeners: Map<string, Set<WSEventCallback>> = new Map();
  private isConnecting: boolean = false;
  private guestToken: string = '';

  // 连接 WebSocket
  connect(type: 'user' | 'staff', guestToken?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      if (this.isConnecting) {
        reject(new Error('正在连接中'));
        return;
      }

      this.isConnecting = true;
      this.guestToken = guestToken || '';

      // 构建 WebSocket URL
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      let wsUrl = `${protocol}//${host}/ws/${type}`;
      
      if (type === 'user' && guestToken) {
        wsUrl += `?guest_token=${encodeURIComponent(guestToken)}`;
      }

      this.url = wsUrl;

      try {
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          console.log('[WebSocket] 连接成功');
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          this.startPing();
          this.emit('connected', { type: 'connected', timestamp: Date.now() });
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            // 处理多条消息（批量发送时用换行分隔）
            const messages = event.data.split('\n').filter((m: string) => m.trim());
            messages.forEach((msgStr: string) => {
              const message: WSMessage = JSON.parse(msgStr);
              this.handleMessage(message);
            });
          } catch (e) {
            console.error('[WebSocket] 消息解析失败:', e);
          }
        };

        this.ws.onclose = (event) => {
          console.log('[WebSocket] 连接关闭:', event.code, event.reason);
          this.isConnecting = false;
          this.stopPing();
          this.emit('disconnected', { type: 'disconnected', timestamp: Date.now() });
          
          // 尝试重连
          if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`[WebSocket] ${this.reconnectDelay / 1000}秒后尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            setTimeout(() => {
              this.connect(type, guestToken).catch(() => {});
            }, this.reconnectDelay);
          }
        };

        this.ws.onerror = (error) => {
          console.error('[WebSocket] 连接错误:', error);
          this.isConnecting = false;
          reject(error);
        };
      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  // 断开连接
  disconnect() {
    this.stopPing();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.reconnectAttempts = this.maxReconnectAttempts; // 阻止重连
  }

  // 发送消息
  send(message: Partial<WSMessage>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        ...message,
        timestamp: Date.now()
      }));
    } else {
      console.warn('[WebSocket] 连接未就绪，无法发送消息');
    }
  }

  // 订阅工单消息
  subscribeTicket(ticketId: number) {
    this.send({
      type: 'subscribe_ticket',
      data: { ticket_id: ticketId }
    });
  }

  // 取消订阅工单消息
  unsubscribeTicket(ticketId: number) {
    this.send({
      type: 'unsubscribe_ticket',
      data: { ticket_id: ticketId }
    });
  }

  // 订阅聊天消息
  subscribeChat(chatId: number) {
    this.send({
      type: 'subscribe_chat',
      data: { chat_id: chatId }
    });
  }

  // 取消订阅聊天消息
  unsubscribeChat(chatId: number) {
    this.send({
      type: 'unsubscribe_chat',
      data: { chat_id: chatId }
    });
  }

  // 发送正在输入提示
  sendTyping(ticketId?: number, chatId?: number) {
    this.send({
      type: 'typing',
      ticket_id: ticketId,
      chat_id: chatId
    });
  }

  // 发送已读回执
  sendRead(ticketId?: number, chatId?: number) {
    this.send({
      type: 'read',
      ticket_id: ticketId,
      chat_id: chatId
    });
  }

  // 添加事件监听
  on(event: string, callback: WSEventCallback) {
    if (!this.eventListeners.has(event)) {
      this.eventListeners.set(event, new Set());
    }
    this.eventListeners.get(event)!.add(callback);
  }

  // 移除事件监听
  off(event: string, callback: WSEventCallback) {
    this.eventListeners.get(event)?.delete(callback);
  }

  // 触发事件
  private emit(event: string, message: WSMessage) {
    this.eventListeners.get(event)?.forEach(callback => {
      try {
        callback(message);
      } catch (e) {
        console.error('[WebSocket] 事件处理错误:', e);
      }
    });
  }

  // 处理收到的消息
  private handleMessage(message: WSMessage) {
    // 触发通用消息事件
    this.emit('message', message);
    
    // 触发特定类型事件
    this.emit(message.type, message);

    // 根据消息类型触发特定事件
    switch (message.type) {
      case 'new_message':
        if (message.ticket_id) {
          this.emit(`ticket_message_${message.ticket_id}`, message);
        }
        if (message.chat_id) {
          this.emit(`chat_message_${message.chat_id}`, message);
        }
        break;
      case 'typing':
        if (message.ticket_id) {
          this.emit(`ticket_typing_${message.ticket_id}`, message);
        }
        if (message.chat_id) {
          this.emit(`chat_typing_${message.chat_id}`, message);
        }
        break;
      case 'read':
        if (message.ticket_id) {
          this.emit(`ticket_read_${message.ticket_id}`, message);
        }
        if (message.chat_id) {
          this.emit(`chat_read_${message.chat_id}`, message);
        }
        break;
      case 'status_change':
        if (message.ticket_id) {
          this.emit(`ticket_status_${message.ticket_id}`, message);
        }
        break;
      case 'assigned':
        if (message.ticket_id) {
          this.emit(`ticket_assigned_${message.ticket_id}`, message);
        }
        break;
      case 'pong':
        // 心跳响应，不需要处理
        break;
    }
  }

  // 开始心跳
  private startPing() {
    this.stopPing();
    this.pingInterval = setInterval(() => {
      this.send({ type: 'ping' });
    }, 30000); // 30秒发送一次心跳
  }

  // 停止心跳
  private stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  // 获取连接状态
  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// 导出单例实例
export const wsClient = new WebSocketClient();

// 导出类（用于需要多个实例的场景）
export { WebSocketClient };
