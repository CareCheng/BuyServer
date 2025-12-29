import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

/**
 * 合并 Tailwind CSS 类名
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * 格式化日期时间
 */
export function formatDateTime(dateStr: string | Date | null | undefined): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

/**
 * 格式化日期
 */
export function formatDate(dateStr: string | Date | null | undefined): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

/**
 * 格式化金额
 */
export function formatMoney(amount: number | string, currency = '¥'): string {
  const num = parseFloat(String(amount))
  if (isNaN(num)) return currency + '0.00'
  return currency + num.toFixed(2)
}

/**
 * 复制文本到剪贴板
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch {
      // 降级到传统方法
    }
  }

  // 降级方案
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()

  try {
    document.execCommand('copy')
    return true
  } catch {
    return false
  } finally {
    document.body.removeChild(textarea)
  }
}

/**
 * 验证邮箱格式
 */
export function isValidEmail(email: string): boolean {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return re.test(email)
}

/**
 * 验证手机号格式（中国大陆）
 */
export function isValidPhone(phone: string): boolean {
  const re = /^1[3-9]\d{9}$/
  return re.test(phone)
}

/**
 * 获取订单状态信息
 */
export function getOrderStatus(status: number): { text: string; variant: 'warning' | 'success' | 'danger' | 'info' } {
  const statusMap: Record<number, { text: string; variant: 'warning' | 'success' | 'danger' | 'info' }> = {
    0: { text: '待支付', variant: 'warning' },
    1: { text: '已支付', variant: 'info' },
    2: { text: '已完成', variant: 'success' },
    3: { text: '已取消', variant: 'danger' },
    4: { text: '已退款', variant: 'danger' },
    5: { text: '已过期', variant: 'danger' },
  }
  return statusMap[status] || { text: '未知', variant: 'info' }
}

/**
 * 遮蔽邮箱
 */
export function maskEmail(email: string): string {
  if (!email) return ''
  const [name, domain] = email.split('@')
  if (!domain) return email
  const maskedName = name.length > 2 ? name[0] + '***' + name[name.length - 1] : name[0] + '***'
  return maskedName + '@' + domain
}

/**
 * 格式化文件大小
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
