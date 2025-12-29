/**
 * 客服管理模块类型定义
 */

/**
 * 客服人员接口
 */
export interface Staff {
  id: number
  username: string
  nickname: string
  email: string
  role: string
  status: number
  max_tickets: number
  current_load: number
  last_active_at: string | null
  created_at: string
}

/**
 * 客服配置接口
 */
export interface SupportConfig {
  id: number
  enabled: boolean
  allow_guest: boolean
  staff_portal_suffix: string
  enable_staff_2fa: boolean
  enable_auto_assign: boolean
  enable_email_notify: boolean
  notify_on_new_ticket: boolean
  notify_on_reply: boolean
  max_attachment_size: number
  working_hours_start: string
  working_hours_end: string
  working_days: string
  offline_message: string
  welcome_message: string
  auto_close_hours: number
  ticket_categories: string
}

/**
 * 统计数据接口
 */
export interface SupportStats {
  tickets: {
    pending: number
    processing: number
    replied: number
    resolved: number
    closed: number
    total: number
    today: number
  }
  online_staff: number
  total_staff: number
}
