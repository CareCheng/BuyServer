'use client'

import { cn } from '@/lib/utils'

interface CardProps {
  children?: React.ReactNode
  className?: string
  title?: string
  icon?: React.ReactNode
  action?: React.ReactNode
}

/**
 * 卡片组件
 */
export function Card({ children, className, title, icon, action }: CardProps) {
  return (
    <div className={cn('card p-6', className)}>
      {title && (
        <div className="flex items-center justify-between gap-2 mb-4 pb-4" style={{ borderBottom: '1px solid var(--border-light)' }}>
          <div className="flex items-center gap-2">
            {icon && <span className="text-primary-400">{icon}</span>}
            <h3 className="text-lg font-semibold" style={{ color: 'var(--text-primary)' }}>{title}</h3>
          </div>
          {action && <div>{action}</div>}
        </div>
      )}
      {children}
    </div>
  )
}

interface FeatureCardProps {
  icon: string
  title: string
  description: string
}

/**
 * 特性卡片组件
 */
export function FeatureCard({ icon, title, description }: FeatureCardProps) {
  return (
    <div className="feature-card text-center">
      <div className="text-4xl mb-4">{icon}</div>
      <h3 className="text-lg font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>{title}</h3>
      <p className="text-sm" style={{ color: 'var(--text-muted)' }}>{description}</p>
    </div>
  )
}
