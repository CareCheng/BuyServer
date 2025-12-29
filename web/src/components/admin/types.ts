/**
 * 管理后台类型定义
 */

// 商品
export interface Product {
  id: number
  name: string
  description: string
  detail: string           // 详细介绍（Markdown/HTML）
  specs: string            // 规格参数（JSON格式）
  features: string         // 特性/卖点列表（JSON格式）
  tags: string             // 商品标签（逗号分隔）
  category_name: string
  price: number
  stock: number
  duration: number
  duration_unit: string
  status: number
  image_url: string
  product_type: number  // 1: 手动卡密
  created_at: string
}

// 手动卡密
export interface ManualKami {
  id: number
  product_id: number
  kami_code: string
  status: number  // 0: 可用, 1: 已售出, 2: 已禁用
  order_id: number
  order_no: string
  sold_at: string
  created_at: string
}

// 卡密统计
export interface KamiStats {
  total: number
  available: number
  sold: number
  disabled: number
}

// 分类
export interface Category {
  id: number
  name: string
  icon: string
  sort_order: number
  status: number
}

// 优惠券
export interface Coupon {
  id: number
  code: string
  name: string
  type: string
  value: number
  min_amount: number
  max_discount: number
  total_count: number
  used_count: number
  per_user_limit: number
  status: number
  start_at: string
  end_at: string
}

// 订单
export interface Order {
  id: number
  order_no: string
  username: string
  product_name: string
  quantity: number
  price: number
  status: number
  created_at: string
  paid_at: string
  card_info: string
}

// 用户
export interface User {
  id: number
  username: string
  email: string
  phone: string
  status: number
  created_at: string
}

// 公告
export interface Announcement {
  id: number
  title: string
  content: string
  type: string
  status: number
  sort_order: number
  created_at: string
}

// 备份
export interface Backup {
  id: number
  filename: string
  file_size_text: string
  db_type: string
  remark: string
  created_by: string
  created_at: string
}

// 日志（文件存储版本，使用AES-256-GCM加密）
export interface Log {
  id: number
  user_type: string    // user, admin, security
  user_id: number
  username: string
  action: string
  target: string
  target_id: string
  detail: string
  ip: string
  user_agent: string
  created_at: string
}

// 支付配置
export interface PaymentConfig {
  alipay_f2f?: { enabled: boolean; app_id: string; has_private_key: boolean; has_public_key: boolean; notify_url: string }
  wechat_pay?: { enabled: boolean; app_id: string; mch_id: string; has_api_key: boolean; notify_url: string }
  yi_pay?: { enabled: boolean; api_url: string; pid: string; has_key: boolean; notify_url: string; return_url: string }
  paypal?: { enabled: boolean; sandbox: boolean; client_id: string; has_client_secret: boolean; currency: string; return_url: string; cancel_url: string }
  stripe?: { enabled: boolean; publishable_key: string; has_secret_key: boolean; has_webhook_secret: boolean; currency: string }
  usdt?: { enabled: boolean; network: string; wallet_address: string; api_provider: string; has_api_key: boolean; has_api_secret: boolean; has_webhook_secret: boolean; exchange_rate: number; min_amount: number; confirmations: number }
}

// 邮箱配置
export interface EmailConfig {
  enabled: boolean
  smtp_host: string
  smtp_port: number
  smtp_user: string
  has_password: boolean
  from_name: string
  from_email: string
  encryption: string  // 加密方式：none/ssl/starttls
  code_length: number
}

// 数据库配置
export interface DBConfig {
  connected: boolean
  type: string
  host: string
  port: number
  user: string
  database: string
  key_length: number
  encryption_key: string
}

// 系统设置
export interface Settings {
  system_title: string
  admin_suffix: string
  server_port: number
  enable_login: boolean
  admin_username: string
  enable_2fa: boolean
  totp_secret: string
}

// 页面配置（精简版 - 合并相关功能）
export const PAGE_CONFIG: Record<string, { title: string; icon: string; permissions?: string[] }> = {
  dashboard: { title: '仪表盘', icon: '📊', permissions: ['dashboard:view'] },
  products: { title: '商品管理', icon: '📦', permissions: ['product:view'] },
  categories: { title: '分类管理', icon: '📁', permissions: ['category:view'] },
  coupons: { title: '优惠券', icon: '🎫', permissions: ['coupon:view'] },
  orders: { title: '订单管理', icon: '📋', permissions: ['order:view'] },
  users: { title: '用户管理', icon: '👥', permissions: ['user:view', 'admin:view', 'role:view'] },
  support: { title: '客服管理', icon: '🎧', permissions: ['support:view'] },
  content: { title: '内容管理', icon: '📢', permissions: ['announcement:view', 'faq:view', 'knowledge:view', 'review:view'] },
  homepage: { title: '首页配置', icon: '🏠', permissions: ['settings:view'] },
  system: { title: '系统管理', icon: '🖥️', permissions: ['log:view', 'backup:view', 'stats:view', 'monitor:view'] },
  config: { title: '系统配置', icon: '⚙️', permissions: ['settings:view', 'settings:payment', 'settings:email', 'settings:database'] },
}
