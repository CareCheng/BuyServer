'use client'

import { motion } from 'framer-motion'
import type { HomepageConfig } from '@/types/homepage'

interface AnnouncementSectionProps {
  config: HomepageConfig
}

/**
 * 公告区块组件
 */
export function AnnouncementSection({ config }: AnnouncementSectionProps) {
  if (!config.announcement_enabled || !config.announcement_content) return null

  // 根据类型获取样式
  const getTypeStyles = () => {
    switch (config.announcement_type) {
      case 'warning':
        return {
          bg: 'rgba(245, 158, 11, 0.1)',
          border: 'rgba(245, 158, 11, 0.3)',
          icon: 'fa-triangle-exclamation',
          iconColor: '#f59e0b',
        }
      case 'success':
        return {
          bg: 'rgba(16, 185, 129, 0.1)',
          border: 'rgba(16, 185, 129, 0.3)',
          icon: 'fa-circle-check',
          iconColor: '#10b981',
        }
      default:
        return {
          bg: `${config.primary_color}15`,
          border: `${config.primary_color}30`,
          icon: 'fa-bullhorn',
          iconColor: config.primary_color,
        }
    }
  }

  const styles = getTypeStyles()

  return (
    <section className="py-8 px-4">
      <div className="max-w-4xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="rounded-xl p-6"
          style={{
            backgroundColor: styles.bg,
            border: `1px solid ${styles.border}`,
          }}
        >
          <div className="flex items-start gap-4">
            <div
              className="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0"
              style={{ backgroundColor: `${styles.iconColor}20` }}
            >
              <i className={`fas ${styles.icon}`} style={{ color: styles.iconColor }} />
            </div>
            <div className="flex-1">
              {config.announcement_title && (
                <h3 className="font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
                  {config.announcement_title}
                </h3>
              )}
              <p className="text-sm leading-relaxed" style={{ color: 'var(--text-secondary)' }}>
                {config.announcement_content}
              </p>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
