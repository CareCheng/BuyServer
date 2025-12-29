/**
 * 首页配置类型定义
 */

// 特性项
export interface FeatureItem {
  icon: string
  title: string
  description: string
}

// 统计项
export interface StatItem {
  value: string
  label: string
  icon: string
}

// 页脚链接
export interface FooterLink {
  text: string
  url: string
}

// 完整首页配置
export interface HomepageConfig {
  // 基础设置
  template: string
  primary_color: string
  secondary_color: string

  // 高级模式（自定义 HTML）
  advanced_mode: boolean
  custom_html: string
  custom_css: string
  custom_js: string

  // Logo 设置
  logo_type: 'text' | 'image' | 'emoji'
  logo_text: string
  logo_image: string
  logo_emoji: string

  // Hero 区块
  hero_enabled: boolean
  hero_title: string
  hero_subtitle: string
  hero_button_text: string
  hero_button_link: string
  hero_background: 'gradient' | 'image' | 'solid'
  hero_bg_image: string
  hero_bg_color: string

  // 特性区块
  features_enabled: boolean
  features_title: string
  features: FeatureItem[]

  // 公告区块
  announcement_enabled: boolean
  announcement_title: string
  announcement_content: string
  announcement_type: 'info' | 'warning' | 'success'

  // 商品展示区块
  products_enabled: boolean
  products_title: string
  products_count: number

  // 统计区块
  stats_enabled: boolean
  stats: StatItem[]

  // CTA 区块
  cta_enabled: boolean
  cta_title: string
  cta_subtitle: string
  cta_button_text: string
  cta_button_link: string

  // 页脚设置
  footer_text: string
  footer_links: FooterLink[]

  // 浮动按钮
  floating_button_enabled: boolean
  floating_button_icon: string
  floating_button_link: string
}

// 模板信息
export interface TemplateInfo {
  id: string
  name: string
  description: string
  preview: string
}

// 默认配置
export const defaultHomepageConfig: HomepageConfig = {
  template: 'modern',
  primary_color: '#6366f1',
  secondary_color: '#8b5cf6',

  // 高级模式默认关闭
  advanced_mode: false,
  custom_html: '',
  custom_css: '',
  custom_js: '',

  logo_type: 'emoji',
  logo_text: '卡密购买系统',
  logo_image: '',
  logo_emoji: '🔐',

  hero_enabled: true,
  hero_title: '欢迎使用卡密购买系统',
  hero_subtitle: '安全、便捷的卡密购买平台',
  hero_button_text: '浏览商品',
  hero_button_link: '/products/',
  hero_background: 'gradient',
  hero_bg_image: '',
  hero_bg_color: '',

  features_enabled: true,
  features_title: '为什么选择我们',
  features: [
    { icon: '🔒', title: '安全可靠', description: '采用ECC加密通信，保障交易安全' },
    { icon: '⚡', title: '即时发货', description: '支付成功后立即获取卡密' },
    { icon: '💬', title: '售后保障', description: '专业客服团队，随时为您服务' },
  ],

  announcement_enabled: false,
  announcement_title: '系统公告',
  announcement_content: '',
  announcement_type: 'info',

  products_enabled: true,
  products_title: '热门商品',
  products_count: 6,

  stats_enabled: true,
  stats: [
    { value: '10000+', label: '用户数量', icon: '👥' },
    { value: '50000+', label: '成交订单', icon: '📦' },
    { value: '99.9%', label: '好评率', icon: '⭐' },
    { value: '24/7', label: '在线客服', icon: '💬' },
  ],

  cta_enabled: true,
  cta_title: '准备好开始了吗？',
  cta_subtitle: '立即注册，享受便捷的购买体验',
  cta_button_text: '立即注册',
  cta_button_link: '/register/',

  footer_text: '卡密购买系统',
  footer_links: [
    { text: '常见问题', url: '/faq/' },
    { text: '联系客服', url: '/message/' },
  ],

  floating_button_enabled: true,
  floating_button_icon: 'fa-headset',
  floating_button_link: '/message/',
}
