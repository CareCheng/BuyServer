'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { AnnouncementsPage } from './Announcements'
import { ReviewsPage } from './Reviews'
import { InvoicesPage } from './Invoices'
import { UndoPage } from './Undo'
import { FAQPage } from './FAQ'

/**
 * 内容管理页面标签配置
 */
const TABS = [
  { id: 'announcements', label: '公告管理', icon: 'fa-bullhorn' },
  { id: 'reviews', label: '评价管理', icon: 'fa-star' },
  { id: 'faq', label: 'FAQ管理', icon: 'fa-question-circle' },
  { id: 'invoices', label: '发票管理', icon: 'fa-file-invoice' },
  { id: 'undo', label: '操作撤销', icon: 'fa-undo' },
]

/**
 * 内容管理组合页面
 * 合并：公告管理、评价管理、发票管理、操作撤销
 */
export function ContentManagePage() {
  const [activeTab, setActiveTab] = useState('announcements')

  // 渲染当前标签内容
  const renderContent = () => {
    switch (activeTab) {
      case 'announcements':
        return <AnnouncementsPage />
      case 'reviews':
        return <ReviewsPage />
      case 'faq':
        return <FAQPage />
      case 'invoices':
        return <InvoicesPage />
      case 'undo':
        return <UndoPage />
      default:
        return <AnnouncementsPage />
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
