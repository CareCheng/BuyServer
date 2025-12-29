'use client'

import { useState } from 'react'
import { TabButton } from './components'
import { StaffManagement } from './StaffManagement'
import { ConfigManagement } from './ConfigManagement'
import { StatsPanel } from './StatsPanel'

/**
 * 客服管理组件
 */
export function Support() {
  const [activeTab, setActiveTab] = useState<'staff' | 'config' | 'stats'>('staff')

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-dark-100">
        <i className="fas fa-headset mr-2 text-primary-400" />
        客服管理
      </h2>

      {/* 标签页 */}
      <div className="flex items-center gap-2 flex-wrap">
        <TabButton
          active={activeTab === 'staff'}
          onClick={() => setActiveTab('staff')}
          icon="fa-users"
          label="客服人员"
        />
        <TabButton
          active={activeTab === 'config'}
          onClick={() => setActiveTab('config')}
          icon="fa-cog"
          label="系统配置"
        />
        <TabButton
          active={activeTab === 'stats'}
          onClick={() => setActiveTab('stats')}
          icon="fa-chart-bar"
          label="数据统计"
        />
      </div>

      {activeTab === 'staff' && <StaffManagement />}
      {activeTab === 'config' && <ConfigManagement />}
      {activeTab === 'stats' && <StatsPanel />}
    </div>
  )
}
