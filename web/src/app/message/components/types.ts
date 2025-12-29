/**
 * 客服支持页面类型定义
 */

/**
 * 工单接口
 */
export interface Ticket {
  id: number
  ticket_no: string
  subject: string
  category: string
  priority: number
  status: number
  created_at: string
  last_reply_at: string | null
  last_reply_by: string
}

/**
 * 工单消息接口
 */
export interface TicketMessage {
  id: number
  sender_type: string
  sender_name: string
  content: string
  created_at: string
}

/**
 * 聊天会话接口
 */
export interface LiveChat {
  id: number
  session_id: string
  status: number
  staff_name: string
}

/**
 * 聊天消息接口
 */
export interface ChatMessage {
  id: number
  sender_type: string
  sender_name: string
  content: string
  created_at: string
}

/**
 * 客服配置接口
 */
export interface SupportConfig {
  enabled: boolean
  allow_guest: boolean
  welcome: string
  offline: string
  categories: string
  online_count: number
  is_online: boolean
}
