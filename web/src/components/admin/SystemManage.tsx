'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { StatsPage } from './Stats'
import { MonitorPage } from './Monitor'
import { TasksPage } from './Tasks'
import { LogsPage } from './Logs'
import { BackupsPage } from './Backups'

/**
 * 系统管理页面标签配置
 */
const TABS = [
  { id: 'stats', label: '统计报表', icon: 'fa-chart-line' },
  { id: 'monitor', label: '系统监控', icon: 'fa-heartbeat' },
  { id: 'tasks', label: '定时任务', icon: 'fa-clock' },
  { id: 'logs', label: '操作日志', icon: 'fa-history' },
  { id: 'backups', label: '数据备份', icon: 'fa-database' },
]

/**
 * 系统管理组合页面
 * 合并：统计报表、系统监控、定时任务、操作日志、数据备份
 */
export function SystemManagePage() {
  const [activeTab, setActiveTab] = useState('stats')

  // 渲染当前标签内容
  const renderContent = () => {
    switch (activeTab) {
      case 'stats':
        return <StatsPage />
      case 'monitor':
        return <MonitorPage />
      case 'tasks':
        return <TasksPage />
      case 'logs':
        return <LogsPage />
      case 'backups':
        return <BackupsPage />
      default:
        return <StatsPage />
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
