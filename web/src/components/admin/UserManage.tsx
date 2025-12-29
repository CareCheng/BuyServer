'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { UsersPage } from './Users'
import { RolesPage } from './Roles'
import { BalancePage } from './Balance'
import { PointsPage } from './Points'
import { DeletionsPage } from './Deletions'

/**
 * 用户管理页面标签配置
 */
const TABS = [
  { id: 'users', label: '用户列表', icon: 'fa-users' },
  { id: 'roles', label: '权限管理', icon: 'fa-user-shield' },
  { id: 'balance', label: '余额管理', icon: 'fa-wallet' },
  { id: 'points', label: '积分管理', icon: 'fa-coins' },
  { id: 'deletions', label: '注销管理', icon: 'fa-user-slash' },
]

/**
 * 用户管理组合页面
 * 合并：用户列表、权限管理、余额管理、积分管理、注销管理
 */
export function UserManagePage() {
  const [activeTab, setActiveTab] = useState('users')

  // 渲染当前标签内容
  const renderContent = () => {
    switch (activeTab) {
      case 'users':
        return <UsersPage />
      case 'roles':
        return <RolesPage />
      case 'balance':
        return <BalancePage />
      case 'points':
        return <PointsPage />
      case 'deletions':
        return <DeletionsPage />
      default:
        return <UsersPage />
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
