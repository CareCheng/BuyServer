'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Card, Badge, Modal } from '@/components/ui'
import { apiGet, apiPost, apiDelete } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * 登录设备接口
 */
interface LoginDevice {
  id: number
  device_name: string
  device_type: string
  browser: string
  os: string
  ip: string
  location: string
  last_active: string
  created_at: string
  is_current: boolean
}

/**
 * 登录历史接口
 */
interface LoginHistory {
  id: number
  ip: string
  location: string
  device_name: string
  browser: string
  os: string
  status: string
  created_at: string
}

/**
 * 异地登录提醒接口
 */
interface LoginAlert {
  id: number
  ip: string
  location: string
  device_name: string
  browser: string
  os: string
  acknowledged: boolean
  created_at: string
}

/**
 * 设备管理标签页
 */
export function DevicesTab() {
  const [activeSection, setActiveSection] = useState<'devices' | 'history' | 'alerts'>('devices')
  const [devices, setDevices] = useState<LoginDevice[]>([])
  const [history, setHistory] = useState<LoginHistory[]>([])
  const [alerts, setAlerts] = useState<LoginAlert[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [loading, setLoading] = useState(true)
  const [showRemoveModal, setShowRemoveModal] = useState(false)
  const [selectedDevice, setSelectedDevice] = useState<LoginDevice | null>(null)

  // 加载设备列表
  const loadDevices = async () => {
    const res = await apiGet<{ data: LoginDevice[] }>('/api/user/devices')
    if (res.success && res.data) {
      setDevices(res.data)
    }
  }

  // 加载登录历史
  const loadHistory = async () => {
    const res = await apiGet<{ data: LoginHistory[] }>('/api/user/login-history')
    if (res.success && res.data) {
      setHistory(res.data)
    }
  }

  // 加载异地登录提醒
  const loadAlerts = async () => {
    const res = await apiGet<{ data: { alerts: LoginAlert[] } }>('/api/user/login-alerts')
    if (res.success && res.data?.alerts) {
      setAlerts(res.data.alerts)
    }
  }

  // 加载未读提醒数量
  const loadUnreadCount = async () => {
    const res = await apiGet<{ data: { count: number } }>('/api/user/login-alerts/unread-count')
    if (res.success && res.data) {
      setUnreadCount(res.data.count || 0)
    }
  }

  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      await Promise.all([loadDevices(), loadUnreadCount()])
      setLoading(false)
    }
    loadData()
  }, [])

  // 切换分区时加载数据
  useEffect(() => {
    if (activeSection === 'history' && history.length === 0) {
      loadHistory()
    } else if (activeSection === 'alerts' && alerts.length === 0) {
      loadAlerts()
    }
  }, [activeSection])

  // 移除设备
  const handleRemoveDevice = async () => {
    if (!selectedDevice) return
    const res = await apiDelete(`/api/user/device/${selectedDevice.id}`)
    if (res.success) {
      toast.success('设备已移除')
      setShowRemoveModal(false)
      setSelectedDevice(null)
      loadDevices()
    } else {
      toast.error(res.error || '移除失败')
    }
  }

  // 移除所有其他设备
  const handleRemoveAllOther = async () => {
    const res = await apiPost('/api/user/devices/remove-all', {})
    if (res.success) {
      toast.success('已移除所有其他设备')
      loadDevices()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 确认异地登录提醒
  const handleAcknowledgeAlert = async (alertId: number) => {
    const res = await apiPost(`/api/user/login-alert/${alertId}/acknowledge`, {})
    if (res.success) {
      toast.success('已确认')
      loadAlerts()
      loadUnreadCount()
    }
  }

  // 获取设备图标
  const getDeviceIcon = (deviceType: string) => {
    switch (deviceType?.toLowerCase()) {
      case 'mobile': return 'fa-mobile-alt'
      case 'tablet': return 'fa-tablet-alt'
      default: return 'fa-desktop'
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
      {/* 分区切换 */}
      <div className="flex gap-2 border-b border-dark-700/50 pb-4">
        <button
          onClick={() => setActiveSection('devices')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeSection === 'devices'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-laptop mr-2" />
          登录设备
        </button>
        <button
          onClick={() => setActiveSection('history')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeSection === 'history'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-history mr-2" />
          登录历史
        </button>
        <button
          onClick={() => setActiveSection('alerts')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors relative ${
            activeSection === 'alerts'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-bell mr-2" />
          异地提醒
          {unreadCount > 0 && (
            <span className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
              {unreadCount}
            </span>
          )}
        </button>
      </div>

      {/* 登录设备列表 */}
      {activeSection === 'devices' && (
        <Card
          title="登录设备"
          icon={<i className="fas fa-laptop" />}
          action={
            devices.filter(d => !d.is_current).length > 0 && (
              <Button size="sm" variant="danger" onClick={handleRemoveAllOther}>
                移除其他设备
              </Button>
            )
          }
        >
          {devices.length === 0 ? (
            <div className="text-center py-8 text-dark-400">暂无登录设备</div>
          ) : (
            <div className="space-y-3">
              {devices.map((device) => (
                <div
                  key={device.id}
                  className={`p-4 rounded-xl border ${
                    device.is_current
                      ? 'bg-primary-500/10 border-primary-500/30'
                      : 'bg-dark-700/30 border-dark-600/50'
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start gap-3">
                      <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                        device.is_current ? 'bg-primary-500/20 text-primary-400' : 'bg-dark-600/50 text-dark-400'
                      }`}>
                        <i className={`fas ${getDeviceIcon(device.device_type)} text-lg`} />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-dark-100">{device.device_name || '未知设备'}</span>
                          {device.is_current && (
                            <Badge variant="success">当前设备</Badge>
                          )}
                        </div>
                        <div className="text-sm text-dark-400 mt-1">
                          {device.browser} · {device.os}
                        </div>
                        <div className="text-sm text-dark-500 mt-1">
                          <i className="fas fa-map-marker-alt mr-1" />
                          {device.location || device.ip}
                          <span className="mx-2">·</span>
                          最后活跃: {formatDateTime(device.last_active)}
                        </div>
                      </div>
                    </div>
                    {!device.is_current && (
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          setSelectedDevice(device)
                          setShowRemoveModal(true)
                        }}
                      >
                        <i className="fas fa-times text-red-400" />
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      )}

      {/* 登录历史 */}
      {activeSection === 'history' && (
        <Card title="登录历史" icon={<i className="fas fa-history" />}>
          {history.length === 0 ? (
            <div className="text-center py-8 text-dark-400">暂无登录记录</div>
          ) : (
            <div className="space-y-3">
              {history.map((item) => (
                <div key={item.id} className="p-4 bg-dark-700/30 rounded-xl border border-dark-600/50">
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-dark-100">{item.device_name || '未知设备'}</span>
                        <Badge variant={item.status === 'success' ? 'success' : 'danger'}>
                          {item.status === 'success' ? '成功' : '失败'}
                        </Badge>
                      </div>
                      <div className="text-sm text-dark-400 mt-1">
                        {item.browser} · {item.os}
                      </div>
                      <div className="text-sm text-dark-500 mt-1">
                        <i className="fas fa-map-marker-alt mr-1" />
                        {item.location || item.ip}
                      </div>
                    </div>
                    <div className="text-sm text-dark-500">
                      {formatDateTime(item.created_at)}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      )}

      {/* 异地登录提醒 */}
      {activeSection === 'alerts' && (
        <Card title="异地登录提醒" icon={<i className="fas fa-bell" />}>
          {alerts.length === 0 ? (
            <div className="text-center py-8 text-dark-400">暂无异地登录提醒</div>
          ) : (
            <div className="space-y-3">
              {alerts.map((alert) => (
                <div
                  key={alert.id}
                  className={`p-4 rounded-xl border ${
                    alert.acknowledged
                      ? 'bg-dark-700/30 border-dark-600/50'
                      : 'bg-yellow-500/10 border-yellow-500/30'
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <i className={`fas fa-exclamation-triangle ${alert.acknowledged ? 'text-dark-400' : 'text-yellow-400'}`} />
                        <span className="text-dark-100">检测到异地登录</span>
                        {!alert.acknowledged && <Badge variant="warning">未确认</Badge>}
                      </div>
                      <div className="text-sm text-dark-400 mt-2">
                        设备: {alert.device_name || '未知'} · {alert.browser} · {alert.os}
                      </div>
                      <div className="text-sm text-dark-500 mt-1">
                        <i className="fas fa-map-marker-alt mr-1" />
                        {alert.location || alert.ip}
                        <span className="mx-2">·</span>
                        {formatDateTime(alert.created_at)}
                      </div>
                    </div>
                    {!alert.acknowledged && (
                      <Button size="sm" onClick={() => handleAcknowledgeAlert(alert.id)}>
                        确认是我
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      )}

      {/* 移除设备确认弹窗 */}
      <Modal
        isOpen={showRemoveModal}
        onClose={() => setShowRemoveModal(false)}
        title="移除设备"
        size="sm"
      >
        <div className="space-y-4">
          <p className="text-dark-400">
            确定要移除设备 <span className="text-dark-100">{selectedDevice?.device_name}</span> 吗？
            该设备将被强制下线。
          </p>
          <div className="flex gap-3">
            <Button variant="secondary" className="flex-1" onClick={() => setShowRemoveModal(false)}>
              取消
            </Button>
            <Button variant="danger" className="flex-1" onClick={handleRemoveDevice}>
              确认移除
            </Button>
          </div>
        </div>
      </Modal>
    </motion.div>
  )
}
