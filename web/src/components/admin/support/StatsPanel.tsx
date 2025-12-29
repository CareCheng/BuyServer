'use client'

import { useState, useEffect } from 'react'
import { Card } from '@/components/ui'
import { apiGet } from '@/lib/api'
import { SupportStats } from './types'

/**
 * 统计面板组件
 */
export function StatsPanel() {
  const [stats, setStats] = useState<SupportStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadStats = async () => {
      const res = await apiGet<{ stats: SupportStats }>('/api/admin/support/stats')
      if (res.success && res.stats) {
        setStats(res.stats)
      }
      setLoading(false)
    }
    loadStats()
  }, [])

  if (loading) {
    return (
      <div className="text-center py-8">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  if (!stats) {
    return <div className="text-center py-8 text-dark-400">加载统计失败</div>
  }

  return (
    <div className="space-y-6">
      {/* 客服统计 */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card className="text-center py-6">
          <i className="fas fa-users text-3xl text-primary-400 mb-3" />
          <div className="text-3xl font-bold text-dark-100">{stats.total_staff}</div>
          <div className="text-dark-400 text-sm mt-1">客服总数</div>
        </Card>
        <Card className="text-center py-6">
          <i className="fas fa-user-check text-3xl text-emerald-400 mb-3" />
          <div className="text-3xl font-bold text-dark-100">{stats.online_staff}</div>
          <div className="text-dark-400 text-sm mt-1">在线客服</div>
        </Card>
        <Card className="text-center py-6">
          <i className="fas fa-ticket-alt text-3xl text-blue-400 mb-3" />
          <div className="text-3xl font-bold text-dark-100">{stats.tickets?.total || 0}</div>
          <div className="text-dark-400 text-sm mt-1">工单总数</div>
        </Card>
        <Card className="text-center py-6">
          <i className="fas fa-calendar-day text-3xl text-amber-400 mb-3" />
          <div className="text-3xl font-bold text-dark-100">{stats.tickets?.today || 0}</div>
          <div className="text-dark-400 text-sm mt-1">今日新增</div>
        </Card>
      </div>

      {/* 工单状态分布 */}
      <Card>
        <h3 className="text-lg font-medium text-dark-100 mb-4 flex items-center gap-2">
          <i className="fas fa-chart-pie text-primary-400" />
          工单状态分布
        </h3>
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3">
          <div className="bg-dark-700/30 rounded-xl p-4 text-center">
            <div className="text-2xl font-bold text-amber-400">{stats.tickets?.pending || 0}</div>
            <div className="text-dark-400 text-sm mt-1">待处理</div>
          </div>
          <div className="bg-dark-700/30 rounded-xl p-4 text-center">
            <div className="text-2xl font-bold text-blue-400">{stats.tickets?.processing || 0}</div>
            <div className="text-dark-400 text-sm mt-1">处理中</div>
          </div>
          <div className="bg-dark-700/30 rounded-xl p-4 text-center">
            <div className="text-2xl font-bold text-cyan-400">{stats.tickets?.replied || 0}</div>
            <div className="text-dark-400 text-sm mt-1">已回复</div>
          </div>
          <div className="bg-dark-700/30 rounded-xl p-4 text-center">
            <div className="text-2xl font-bold text-emerald-400">{stats.tickets?.resolved || 0}</div>
            <div className="text-dark-400 text-sm mt-1">已解决</div>
          </div>
          <div className="bg-dark-700/30 rounded-xl p-4 text-center sm:col-span-1 col-span-2">
            <div className="text-2xl font-bold text-dark-400">{stats.tickets?.closed || 0}</div>
            <div className="text-dark-400 text-sm mt-1">已关闭</div>
          </div>
        </div>
      </Card>
    </div>
  )
}
