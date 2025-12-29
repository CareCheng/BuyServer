/**
 * 客服工作台类型定义
 */

/**
 * 客服信息接口
 */
export interface StaffInfo {
  id: number
  username: string
  nickname: string
  role: string
  status: number
  current_load: number
  max_tickets: number
}

/**
 * 工单接口
 */
export interface Ticket {
  id: number
  ticket_no: string
  username: string
  email: string
  subject: string
  category: string
  priority: number
  status: number
  assigned_to: number
  assigned_name: string
  created_at: string
  last_reply_at: string | null
  transfer_count?: number
  merged_to?: number
}

/**
 * 工单消息接口
 */
export interface TicketMessage {
  id: number
  sender_type: string
  sender_name: string
  content: string
  is_internal: boolean
  created_at: string
  msg_type?: string
  file_url?: string
  file_name?: string
  file_size?: number
}

/**
 * 工单附件接口
 */
export interface TicketAttachment {
  id: number
  ticket_id: number
  file_name: string
  file_path: string
  file_size: number
  mime_type: string
  uploader_type: string
  uploader_name: string
  created_at: string
}

/**
 * 在线客服接口（用于转接选择）
 */
export interface OnlineStaff {
  id: number
  username: string
  nickname: string
  role: string
  current_load: number
  max_tickets: number
}

/**
 * 聊天会话接口
 */
export interface LiveChat {
  id: number
  session_id: string
  username: string
  status: number
  created_at: string
}

/**
 * 统计数据接口
 */
export interface TicketStats {
  pending: number
  processing: number
  replied: number
  resolved: number
  closed: number
  total: number
  today: number
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
