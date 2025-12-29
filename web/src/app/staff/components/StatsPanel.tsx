'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Card } from '@/components/ui'
import { apiGet } from '@/lib/api'
import { TicketStats } from './types'

/**
 * 统计卡片组件
 */
function StatCard({
  title,
  value,
  icon,
  color,
}: {
  title: string
  value: number
  icon: string
  color: string
}) {
  return (
    <Card className="text-center">
      <i className={`fas ${icon} text-2xl ${color} mb-2`} />
      <div className="text-2xl font-bold text-dark-100">{value}</div>
      <div className="text-dark-400 text-sm">{title}</div>
    </Card>
  )
}

/**
 * 数据统计面板
 */
export function StatsPanel() {
  const [stats, setStats] = useState<{ tickets: TicketStats; online_staff: number; total_staff: number } | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadStats = async () => {
      const res = await apiGet<{ stats: typeof stats }>('/api/staff/tickets/stats')
      if (res.success && res.stats) {
        // 兼容两种返回格式
        const ticketStats = await apiGet<{ stats: TicketStats }>('/api/staff/tickets/stats')
        if (ticketStats.success) {
          setStats({
            tickets: ticketStats.stats as TicketStats,
            online_staff: 0,
            total_staff: 0,
          })
        }
      }
      setLoading(false)
    }
    loadStats()
  }, [])

  if (loading) {
    return (
      <div className="text-center py-12">
        <i className="fas fa-spinner fa-spin text-3xl text-primary-400" />
      </div>
    )
  }

  const ticketStats = stats?.tickets

  return (
    <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <StatCard
          title="待处理"
          value={ticketStats?.pending || 0}
          icon="fa-clock"
          color="text-amber-400"
        />
        <StatCard
          title="处理中"
          value={ticketStats?.processing || 0}
          icon="fa-spinner"
          color="text-blue-400"
        />
        <StatCard
          title="已回复"
          value={ticketStats?.replied || 0}
          icon="fa-reply"
          color="text-emerald-400"
        />
        <StatCard
          title="今日新增"
          value={ticketStats?.today || 0}
          icon="fa-calendar-day"
          color="text-primary-400"
        />
      </div>

      <Card title="工单统计" icon={<i className="fas fa-chart-pie" />}>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 text-center">
          <div className="bg-dark-700/30 rounded-lg p-4">
            <div className="text-2xl font-bold text-dark-100">{ticketStats?.total || 0}</div>
            <div className="text-dark-400 text-sm">总工单</div>
          </div>
          <div className="bg-dark-700/30 rounded-lg p-4">
            <div className="text-2xl font-bold text-amber-400">{ticketStats?.pending || 0}</div>
            <div className="text-dark-400 text-sm">待处理</div>
          </div>
          <div className="bg-dark-700/30 rounded-lg p-4">
            <div className="text-2xl font-bold text-blue-400">{ticketStats?.processing || 0}</div>
            <div className="text-dark-400 text-sm">处理中</div>
          </div>
          <div className="bg-dark-700/30 rounded-lg p-4">
            <div className="text-2xl font-bold text-emerald-400">{ticketStats?.resolved || 0}</div>
            <div className="text-dark-400 text-sm">已解决</div>
          </div>
          <div className="bg-dark-700/30 rounded-lg p-4">
            <div className="text-2xl font-bold text-dark-400">{ticketStats?.closed || 0}</div>
            <div className="text-dark-400 text-sm">已关闭</div>
          </div>
        </div>
      </Card>
    </motion.div>
  )
}
