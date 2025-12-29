'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Support } from './support/index'
import { BotPage } from './Bot'
import { KnowledgePage } from './Knowledge'
import { TicketTemplatesPage } from './TicketTemplates'

/**
 * 客服管理页面标签配置
 */
const TABS = [
  { id: 'support', label: '工单管理', icon: 'fa-headset' },
  { id: 'bot', label: '智能客服', icon: 'fa-robot' },
  { id: 'knowledge', label: '知识库', icon: 'fa-book' },
  { id: 'templates', label: '工单模板', icon: 'fa-clipboard-list' },
]

/**
 * 客服管理组合页面
 * 合并：工单管理、智能客服、知识库、工单模板
 */
export function SupportManagePage() {
  const [activeTab, setActiveTab] = useState('support')

  // 渲染当前标签内容
  const renderContent = () => {
    switch (activeTab) {
      case 'support':
        return <Support />
      case 'bot':
        return <BotPage />
      case 'knowledge':
        return <KnowledgePage />
      case 'templates':
        return <TicketTemplatesPage />
      default:
        return <Support />
    }
  }

  return (
    <div className="space-y-6">
      {/* 顶部标签切换 */}
      <div className="flex flex-wrap gap-2 border-b border-dark-700/50 pb-4">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? 'bg-primary-500/20 text-primary-400'
                : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
            }`}
          >
            <i className={`fas ${tab.icon} mr-2`} />
            {tab.label}
          </button>
        ))}
      </div>

      {/* 内容区域 */}
      <AnimatePresence mode="wait">
        <motion.div
          key={activeTab}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -10 }}
          transition={{ duration: 0.15 }}
        >
          {renderContent()}
        </motion.div>
      </AnimatePresence>
    </div>
  )
}
