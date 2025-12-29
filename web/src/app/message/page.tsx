'use client'

import { useState, useEffect } from 'react'
import { Navbar, Footer } from '@/components/layout'
import { apiGet } from '@/lib/api'
import { cn } from '@/lib/utils'
import { SupportConfig, LiveChatTab, TicketsTab } from './components'

/**
 * 客服支持页面
 */
export default function MessagePage() {
  const [activeTab, setActiveTab] = useState<'chat' | 'tickets'>('chat')
  const [config, setConfig] = useState<SupportConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [isLoggedIn, setIsLoggedIn] = useState(false)
  const [guestToken, setGuestToken] = useState<string>('')

  // 加载配置
  useEffect(() => {
    const loadConfig = async () => {
      // 检查登录状态
      const userRes = await apiGet<{ user: { username: string } }>('/api/user/info')
      setIsLoggedIn(userRes.success)

      // 加载客服配置
      const configRes = await apiGet<{ enabled: boolean; allow_guest: boolean; welcome: string; offline: string; categories: string; online_count: number; is_online: boolean }>('/api/support/config')
      if (configRes.success) {
        setConfig({
          enabled: configRes.enabled,
          allow_guest: configRes.allow_guest,
          welcome: configRes.welcome,
          offline: configRes.offline,
          categories: configRes.categories,
          online_count: configRes.online_count,
          is_online: configRes.is_online,
        })
      }

      // 从 localStorage 获取游客令牌
      const savedToken = localStorage.getItem('guest_token')
      if (savedToken) setGuestToken(savedToken)

      setLoading(false)
    }
    loadConfig()
  }, [])

  const tabs = [
    { id: 'chat' as const, label: '在线咨询', icon: 'fa-comments' },
    { id: 'tickets' as const, label: '工单中心', icon: 'fa-ticket-alt' },
  ]

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
      </div>
    )
  }

  if (!config?.enabled) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 flex items-center justify-center">
          <div className="text-center">
            <div className="text-6xl mb-4">🔒</div>
            <h2 className="text-xl font-semibold text-dark-100 mb-2">客服系统暂未开放</h2>
            <p className="text-dark-400">请稍后再试</p>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />

      <main className="flex-1 py-8 px-4">
        <div className="max-w-4xl mx-auto">
          <h1 className="text-2xl font-bold text-dark-100 mb-6">
            <i className="fas fa-headset mr-3 text-primary-400" />
            客服支持
          </h1>

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

          {/* 在线咨询 */}
          {activeTab === 'chat' && (
            <LiveChatTab
              config={config}
              isLoggedIn={isLoggedIn}
              guestToken={guestToken}
              setGuestToken={setGuestToken}
            />
          )}

          {/* 工单中心 */}
          {activeTab === 'tickets' && (
            <TicketsTab
              config={config}
              isLoggedIn={isLoggedIn}
              guestToken={guestToken}
              setGuestToken={setGuestToken}
            />
          )}
        </div>
      </main>

      <Footer />
    </div>
  )
}
