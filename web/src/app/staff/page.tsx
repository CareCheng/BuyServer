'use client'

import { useState, useEffect } from 'react'
import toast from 'react-hot-toast'
import { Button, Badge } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { cn } from '@/lib/utils'
import { StaffInfo, TicketsPanel, ChatsPanel, StatsPanel } from './components'

/**
 * 客服工作台页面
 */
export default function StaffDashboardPage() {
  const [staff, setStaff] = useState<StaffInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'tickets' | 'chats' | 'stats'>('tickets')

  // 检查登录状态
  useEffect(() => {
    const checkAuth = async () => {
      const res = await apiGet<{ staff: StaffInfo }>('/api/staff/info')
      if (res.success && res.staff) {
        setStaff(res.staff)
      } else {
        window.location.href = '/staff/login/'
      }
      setLoading(false)
    }
    checkAuth()
  }, [])

  // 登出
  const handleLogout = async () => {
    await apiPost('/api/staff/logout')
    window.location.href = '/staff/login/'
  }

  const tabs = [
    { id: 'tickets' as const, label: '工单管理', icon: 'fa-ticket-alt' },
    { id: 'chats' as const, label: '在线咨询', icon: 'fa-comments' },
    { id: 'stats' as const, label: '数据统计', icon: 'fa-chart-bar' },
  ]

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-dark-900">
        <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-dark-900">
      {/* 顶部导航 */}
      <header className="bg-dark-800 border-b border-dark-700">
        <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <i className="fas fa-headset text-2xl text-primary-400" />
            <h1 className="text-xl font-bold text-dark-100">客服工作台</h1>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-dark-300">
              <i className="fas fa-user mr-2" />
              {staff?.nickname || staff?.username}
              {staff?.role === 'supervisor' && (
                <Badge variant="info" className="ml-2">主管</Badge>
              )}
            </span>
            <Button size="sm" variant="ghost" onClick={handleLogout}>
              <i className="fas fa-sign-out-alt mr-1" />
              退出
            </Button>
          </div>
        </div>
      </header>

      {/* 主内容 */}
      <main className="max-w-7xl mx-auto px-4 py-6">
        {/* 标签页导航 */}
        <div className="flex border-b border-dark-700/50 mb-6">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn('tab', activeTab === tab.id && 'active')}
            >
              <i className={`fas ${tab.icon} mr-2`} />
              {tab.label}
            </button>
          ))}
        </div>

        {/* 工单管理 */}
        {activeTab === 'tickets' && staff && <TicketsPanel staff={staff} />}

        {/* 在线咨询 */}
        {activeTab === 'chats' && staff && <ChatsPanel staff={staff} />}

        {/* 数据统计 */}
        {activeTab === 'stats' && <StatsPanel />}
      </main>
    </div>
  )
}
