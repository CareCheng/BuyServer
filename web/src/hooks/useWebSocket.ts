/**
 * WebSocket React Hook
 * 提供便捷的 WebSocket 连接和消息处理
 */

import { useEffect, useCallback, useRef, useState } from 'react';
import { wsClient, WSMessage } from '@/lib/websocket';

interface UseWebSocketOptions {
  type: 'user' | 'staff';
  guestToken?: string;
  autoConnect?: boolean;
  onMessage?: (message: WSMessage) => void;
  onConnected?: () => void;
  onDisconnected?: () => void;
}

interface UseWebSocketReturn {
  isConnected: boolean;
  connect: () => Promise<void>;
  disconnect: () => void;
  subscribeTicket: (ticketId: number) => void;
  unsubscribeTicket: (ticketId: number) => void;
  subscribeChat: (chatId: number) => void;
  unsubscribeChat: (chatId: number) => void;
  sendTyping: (ticketId?: number, chatId?: number) => void;
  sendRead: (ticketId?: number, chatId?: number) => void;
}

export function useWebSocket(options: UseWebSocketOptions): UseWebSocketReturn {
  const { type, guestToken, autoConnect = true, onMessage, onConnected, onDisconnected } = options;
  const [isConnected, setIsConnected] = useState(false);
  const callbacksRef = useRef({ onMessage, onConnected, onDisconnected });

  // 更新回调引用
  useEffect(() => {
    callbacksRef.current = { onMessage, onConnected, onDisconnected };
  }, [onMessage, onConnected, onDisconnected]);

  // 连接处理
  const connect = useCallback(async () => {
    try {
      await wsClient.connect(type, guestToken);
      setIsConnected(true);
    } catch (error) {
      console.error('WebSocket 连接失败:', error);
    }
  }, [type, guestToken]);

  // 断开连接
  const disconnect = useCallback(() => {
    wsClient.disconnect();
    setIsConnected(false);
  }, []);

  // 设置事件监听
  useEffect(() => {
    const handleMessage = (msg: WSMessage) => {
      callbacksRef.current.onMessage?.(msg);
    };

    const handleConnected = () => {
      setIsConnected(true);
      callbacksRef.current.onConnected?.();
    };

    const handleDisconnected = () => {
      setIsConnected(false);
      callbacksRef.current.onDisconnected?.();
    };

    wsClient.on('message', handleMessage);
    wsClient.on('connected', handleConnected);
    wsClient.on('disconnected', handleDisconnected);

    // 自动连接
    if (autoConnect) {
      connect();
    }

    return () => {
      wsClient.off('message', handleMessage);
      wsClient.off('connected', handleConnected);
      wsClient.off('disconnected', handleDisconnected);
    };
  }, [autoConnect, connect]);

  return {
    isConnected,
    connect,
    disconnect,
    subscribeTicket: wsClient.subscribeTicket.bind(wsClient),
    unsubscribeTicket: wsClient.unsubscribeTicket.bind(wsClient),
    subscribeChat: wsClient.subscribeChat.bind(wsClient),
    unsubscribeChat: wsClient.unsubscribeChat.bind(wsClient),
    sendTyping: wsClient.sendTyping.bind(wsClient),
    sendRead: wsClient.sendRead.bind(wsClient),
  };
}

/**
 * 工单消息 Hook
 * 订阅特定工单的实时消息
 */
interface UseTicketMessagesOptions {
  ticketId: number;
  type: 'user' | 'staff';
  guestToken?: string;
  onNewMessage?: (message: any) => void;
  onTyping?: (data: any) => void;
  onStatusChange?: (status: number, operator: string) => void;
}

export function useTicketMessages(options: UseTicketMessagesOptions) {
  const { ticketId, type, guestToken, onNewMessage, onTyping, onStatusChange } = options;
  const [isTyping, setIsTyping] = useState(false);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleMessage = useCallback((msg: WSMessage) => {
    if (msg.type === 'new_message' && msg.ticket_id === ticketId) {
      onNewMessage?.(msg.data);
    } else if (msg.type === 'typing' && msg.ticket_id === ticketId) {
      setIsTyping(true);
      onTyping?.(msg.data);
      // 3秒后自动清除输入状态
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
      typingTimeoutRef.current = setTimeout(() => {
        setIsTyping(false);
      }, 3000);
    } else if (msg.type === 'status_change' && msg.ticket_id === ticketId) {
      const data = msg.data as { status: number; operator: string };
      onStatusChange?.(data.status, data.operator);
    }
  }, [ticketId, onNewMessage, onTyping, onStatusChange]);

  const ws = useWebSocket({
    type,
    guestToken,
    onMessage: handleMessage,
  });

  // 订阅工单
  useEffect(() => {
    if (ws.isConnected && ticketId) {
      ws.subscribeTicket(ticketId);
      return () => {
        ws.unsubscribeTicket(ticketId);
      };
    }
  }, [ws.isConnected, ticketId]);

  // 清理
  useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
    };
  }, []);

  return {
    ...ws,
    isTyping,
    sendTyping: () => ws.sendTyping(ticketId),
    sendRead: () => ws.sendRead(ticketId),
  };
}

/**
 * 聊天消息 Hook
 * 订阅特定聊天的实时消息
 */
interface UseChatMessagesOptions {
  chatId: number;
  type: 'user' | 'staff';
  guestToken?: string;
  onNewMessage?: (message: any) => void;
  onTyping?: (data: any) => void;
  onChatAccepted?: (staffId: number, staffName: string) => void;
  onChatEnded?: () => void;
}

export function useChatMessages(options: UseChatMessagesOptions) {
  const { chatId, type, guestToken, onNewMessage, onTyping, onChatAccepted, onChatEnded } = options;
  const [isTyping, setIsTyping] = useState(false);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleMessage = useCallback((msg: WSMessage) => {
    if (msg.type === 'new_message' && msg.chat_id === chatId) {
      onNewMessage?.(msg.data);
    } else if (msg.type === 'typing' && msg.chat_id === chatId) {
      setIsTyping(true);
      onTyping?.(msg.data);
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
      typingTimeoutRef.current = setTimeout(() => {
        setIsTyping(false);
      }, 3000);
    } else if (msg.type === 'chat_accepted' && msg.chat_id === chatId) {
      const data = msg.data as { staff_id: number; staff_name: string };
      onChatAccepted?.(data.staff_id, data.staff_name);
    } else if (msg.type === 'chat_ended' && msg.chat_id === chatId) {
      onChatEnded?.();
    }
  }, [chatId, onNewMessage, onTyping, onChatAccepted, onChatEnded]);

  const ws = useWebSocket({
    type,
    guestToken,
    onMessage: handleMessage,
  });

  // 订阅聊天
  useEffect(() => {
    if (ws.isConnected && chatId) {
      ws.subscribeChat(chatId);
      return () => {
        ws.unsubscribeChat(chatId);
      };
    }
  }, [ws.isConnected, chatId]);

  // 清理
  useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
    };
  }, []);

  return {
    ...ws,
    isTyping,
    sendTyping: () => ws.sendTyping(undefined, chatId),
    sendRead: () => ws.sendRead(undefined, chatId),
  };
}

/**
 * 客服通知 Hook
 * 接收新工单、新聊天等通知
 */
interface UseStaffNotificationsOptions {
  onNewTicket?: (ticket: any) => void;
  onNewChat?: (chat: any) => void;
  onNewAssignment?: (ticketId: number) => void;
  onStaffOnline?: (staffId: number) => void;
  onStaffOffline?: (staffId: number) => void;
}

export function useStaffNotifications(options: UseStaffNotificationsOptions) {
  const { onNewTicket, onNewChat, onNewAssignment, onStaffOnline, onStaffOffline } = options;

  const handleMessage = useCallback((msg: WSMessage) => {
    switch (msg.type) {
      case 'new_ticket':
        onNewTicket?.(msg.data);
        break;
      case 'new_chat':
        onNewChat?.(msg.data);
        break;
      case 'new_assignment':
        onNewAssignment?.(msg.data?.ticket_id);
        break;
      case 'staff_online':
        onStaffOnline?.(msg.data?.staff_id);
        break;
      case 'staff_offline':
        onStaffOffline?.(msg.data?.staff_id);
        break;
    }
  }, [onNewTicket, onNewChat, onNewAssignment, onStaffOnline, onStaffOffline]);

  return useWebSocket({
    type: 'staff',
    onMessage: handleMessage,
  });
}
