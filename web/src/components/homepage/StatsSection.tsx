'use client'

import { motion } from 'framer-motion'
import type { HomepageConfig } from '@/types/homepage'

interface StatsSectionProps {
  config: HomepageConfig
}

/**
 * 统计区块组件
 */
export function StatsSection({ config }: StatsSectionProps) {
  if (!config.stats_enabled || !config.stats?.length) return null

  // 判断图标是否是 Font Awesome 图标
  const renderIcon = (icon: string) => {
    if (icon.startsWith('fa-')) {
      return <i className={`fas ${icon} text-2xl`} style={{ color: config.primary_color }} />
    }
    return <span className="text-3xl">{icon}</span>
  }

  return (
    <section className="py-16 px-4">
      <div className="max-w-6xl mx-auto">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
          {config.stats.map((stat, index) => (
            <motion.div
              key={index}
              initial={{ opacity: 0, scale: 0.9 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ delay: index * 0.1 }}
              className="text-center p-6"
            >
              <div className="mb-3">{renderIcon(stat.icon)}</div>
              <div
                className="text-3xl md:text-4xl font-bold mb-2"
                style={{ color: config.primary_color }}
              >
                {stat.value}
              </div>
              <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
                {stat.label}
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}
