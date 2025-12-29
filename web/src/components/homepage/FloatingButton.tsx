'use client'

import Link from 'next/link'
import type { HomepageConfig } from '@/types/homepage'

interface FloatingButtonProps {
  config: HomepageConfig
}

/**
 * 浮动按钮组件
 */
export function FloatingButton({ config }: FloatingButtonProps) {
  if (!config.floating_button_enabled) return null

  // 判断图标是否是 Font Awesome 图标
  const renderIcon = () => {
    const icon = config.floating_button_icon
    if (icon.startsWith('fa-')) {
      return <i className={`fas ${icon} text-xl`} />
    }
    return <span className="text-xl">{icon}</span>
  }

  return (
    <Link
      href={config.floating_button_link}
      className="fixed bottom-6 right-6 w-14 h-14 rounded-full shadow-lg flex items-center justify-center text-white transition-all hover:scale-110 hover:shadow-xl z-50"
      style={{
        background: `linear-gradient(135deg, ${config.primary_color}, ${config.secondary_color})`,
        boxShadow: `0 10px 30px -10px ${config.primary_color}80`,
      }}
      title="联系客服"
    >
      {renderIcon()}
    </Link>
  )
}
