'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Input, Modal } from '@/components/ui'
import { apiGet, apiPost, apiDelete } from '@/lib/api'

/**
 * Redis 配置接口
 */
interface RedisConfig {
  enabled: boolean
  mode: string
  address: string
  password: string
  database: number
  key_prefix: string
  pool_size: number
  min_idle_conns: number
  dial_timeout: number
  read_timeout: number
  write_timeout: number
  master_name: string
  sentinel_addrs: string
  tls_enabled: boolean
  max_retries: number
  connected: boolean
}

/**
 * 缓存仪表盘数据
 */
interface CacheDashboard {
  mode: string
  status: string
  version: string
  uptime: string
  uptime_seconds: number
  hit_rate: number
  hit_rate_str: string
  hits: number
  misses: number
  total_requests: number
  ops_per_second: number
  memory_used: string
  memory_used_bytes: number
  memory_peak: string
  memory_peak_bytes: number
  memory_limit: string
  memory_policy: string
  keys_count: number
  expiring_keys: number
  expired_keys: number
  evicted_keys: number
  connected_clients: number
  max_clients: number
  blocked_clients: number
  role: string
  connected_slaves: number
  rdb_enabled: boolean
  aof_enabled: boolean
  last_save_time: string
  last_save_status: string
  failovers: number
  last_error: string
  last_error_time: string
  local_cache_size: number
  local_cache_memory: string
}

/**
 * Redis 配置页面组件
 * 提供 Redis 连接配置、仪表盘、缓存统计和缓存管理功能
 */
export function RedisPage() {
  const [config, setConfig] = useState<RedisConfig | null>(null)
  const [dashboard, setDashboard] = useState<CacheDashboard | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [activeTab, setActiveTab] = useState<'dashboard' | 'config' | 'keys'>('dashboard')
  const [form, setForm] = useState({
    enabled: false,
    mode: 'standalone',
    address: 'localhost:6379',
    password: '',
    database: 0,
    key_prefix: 'user:',
    pool_size: 10,
    min_idle_conns: 5,
    dial_timeout: 5,
    read_timeout: 3,
    write_timeout: 3,
    master_name: '',
    sentinel_addrs: '',
    tls_enabled: false,
    max_retries: 3,
  })
  const [showFlushModal, setShowFlushModal] = useState(false)
  const [flushLoading, setFlushLoading] = useState(false)
  const [flushConfirmText, setFlushConfirmText] = useState('')
  
  // 键管理
  const [keys, setKeys] = useState<string[]>([])
  const [keysLoading, setKeysLoading] = useState(false)
  const [keyPattern, setKeyPattern] = useState('*')
  const [selectedKey, setSelectedKey] = useState<string | null>(null)
  const [keyInfo, setKeyInfo] = useState<{ ttl: string; value: unknown } | null>(null)

  // 加载配置
  const loadConfig = useCallback(async () => {
    setLoading(true)
    try {
      const res = await apiGet<{ config: RedisConfig }>('/api/admin/redis/config')
      if (res.success && res.config) {
        setConfig(res.config)
        setForm({
          enabled: res.config.enabled || false,
          mode: res.config.mode || 'standalone',
          address: res.config.address || 'localhost:6379',
          password: '',
          database: res.config.database || 0,
          key_prefix: res.config.key_prefix || 'user:',
          pool_size: res.config.pool_size || 10,
          min_idle_conns: res.config.min_idle_conns || 5,
          dial_timeout: res.config.dial_timeout || 5,
          read_timeout: res.config.read_timeout || 3,
          write_timeout: res.config.write_timeout || 3,
          master_name: res.config.master_name || '',
          sentinel_addrs: res.config.sentinel_addrs || '',
          tls_enabled: res.config.tls_enabled || false,
          max_retries: res.config.max_retries || 3,
        })
      }
    } catch (error) {
      console.error('加载 Redis 配置失败:', error)
    }
    setLoading(false)
  }, [])

  // 加载仪表盘数据
  const loadDashboard = useCallback(async () => {
    try {
      const res = await apiGet<{ dashboard: CacheDashboard }>('/api/admin/redis/dashboard')
      if (res.success && res.dashboard) {
        setDashboard(res.dashboard)
      }
    } catch (error) {
      console.error('加载仪表盘数据失败:', error)
    }
  }, [])

  // 加载键列表
  const loadKeys = useCallback(async () => {
    setKeysLoading(true)
    try {
      const res = await apiGet<{ keys: string[]; count: number }>(`/api/admin/redis/keys?pattern=${encodeURIComponent(keyPattern)}`)
      if (res.success && res.keys) {
        setKeys(res.keys)
      }
    } catch (error) {
      console.error('加载键列表失败:', error)
    }
    setKeysLoading(false)
  }, [keyPattern])

  // 加载键信息
  const loadKeyInfo = useCallback(async (key: string) => {
    try {
      const res = await apiGet<{ ttl: string; value: unknown }>(`/api/admin/redis/key/info?key=${encodeURIComponent(key)}`)
      if (res.success) {
        setKeyInfo({ ttl: res.ttl || 'N/A', value: res.value })
      }
    } catch (error) {
      console.error('加载键信息失败:', error)
    }
  }, [])

  useEffect(() => {
    loadConfig()
    loadDashboard()
  }, [loadConfig, loadDashboard])

  useEffect(() => {
    if (activeTab === 'keys') {
      loadKeys()
    }
  }, [activeTab, loadKeys])

  // 自动刷新仪表盘
  useEffect(() => {
    if (activeTab === 'dashboard') {
      const interval = setInterval(() => {
        loadDashboard()
      }, 10000)
      return () => clearInterval(interval)
    }
  }, [activeTab, loadDashboard])

  // 测试连接
  const handleTest = async () => {
    setTesting(true)
    toast.loading('正在测试连接...')
    try {
      const data = {
        mode: form.mode,
        address: form.address,
        password: form.password || undefined,
        database: form.database,
        key_prefix: form.key_prefix,
        pool_size: form.pool_size,
        dial_timeout: form.dial_timeout,
        master_name: form.mode === 'sentinel' ? form.master_name : undefined,
        sentinel_addrs: form.mode === 'sentinel' ? form.sentinel_addrs : undefined,
        tls_enabled: form.tls_enabled,
      }
      const res = await apiPost<{ latency: string }>('/api/admin/redis/test', data)
      toast.dismiss()
      if (res.success) {
        toast.success(`连接成功！延迟: ${res.latency || 'N/A'}`)
      } else {
        toast.error(res.error || '连接失败')
      }
    } catch {
      toast.dismiss()
      toast.error('连接测试失败')
    }
    setTesting(false)
  }

  // 保存配置
  const handleSave = async () => {
    setSaving(true)
    try {
      const data: Record<string, unknown> = {
        enabled: form.enabled,
        mode: form.mode,
        address: form.address,
        database: form.database,
        key_prefix: form.key_prefix,
        pool_size: form.pool_size,
        min_idle_conns: form.min_idle_conns,
        dial_timeout: form.dial_timeout,
        read_timeout: form.read_timeout,
        write_timeout: form.write_timeout,
        tls_enabled: form.tls_enabled,
        max_retries: form.max_retries,
      }
      if (form.password) {
        data.password = form.password
      }
      if (form.mode === 'sentinel') {
        data.master_name = form.master_name
        data.sentinel_addrs = form.sentinel_addrs
      }
      const res = await apiPost('/api/admin/redis/config', data)
      if (res.success) {
        toast.success('配置已保存，系统将自动重新连接 Redis')
        loadConfig()
        loadDashboard()
      } else {
        toast.error(res.error || '保存失败')
      }
    } catch {
      toast.error('保存配置失败')
    }
    setSaving(false)
  }

  // 刷新缓存连接
  const handleRefreshCache = async () => {
    toast.loading('正在刷新缓存...')
    try {
      const res = await apiPost('/api/admin/redis/refresh', {})
      toast.dismiss()
      if (res.success) {
        toast.success('缓存已刷新')
        loadDashboard()
      } else {
        toast.error(res.error || '刷新失败')
      }
    } catch {
      toast.dismiss()
      toast.error('刷新缓存失败')
    }
  }

  // 清空缓存
  const handleFlushCache = async () => {
    if (flushConfirmText !== '确认清空所有缓存') {
      toast.error('请输入正确的确认文字')
      return
    }
    setFlushLoading(true)
    try {
      const res = await apiPost('/api/admin/redis/flush?confirm=true', {})
      if (res.success) {
        toast.success('缓存已清空')
        setShowFlushModal(false)
        setFlushConfirmText('')
        loadDashboard()
      } else {
        toast.error(res.error || '清空失败')
      }
    } catch {
      toast.error('清空缓存失败')
    }
    setFlushLoading(false)
  }

  // 删除单个键
  const handleDeleteKey = async (key: string) => {
    if (!confirm(`确定要删除键 "${key}" 吗？`)) return
    try {
      const res = await apiDelete(`/api/admin/redis/key?key=${encodeURIComponent(key)}`)
      if (res.success) {
        toast.success('键已删除')
        setKeys(keys.filter(k => k !== key))
        if (selectedKey === key) {
          setSelectedKey(null)
          setKeyInfo(null)
        }
      } else {
        toast.error(res.error || '删除失败')
      }
    } catch {
      toast.error('删除键失败')
    }
  }

  if (loading) {
    return (
      <div className="text-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'connected': return 'text-green-400 bg-green-500/10 border-green-500/20'
      case 'degraded': return 'text-yellow-400 bg-yellow-500/10 border-yellow-500/20'
      default: return 'text-red-400 bg-red-500/10 border-red-500/20'
    }
  }

  // 获取状态文字
  const getStatusText = (status: string) => {
    switch (status) {
      case 'connected': return '已连接'
      case 'degraded': return '已降级'
      default: return '已断开'
    }
  }

  // 获取模式文字
  const getModeText = (mode: string) => {
    switch (mode) {
      case 'redis-standalone': return 'Redis 单机模式'
      case 'redis-sentinel': return 'Redis 哨兵模式'
      case 'redis-cluster': return 'Redis 集群模式'
      default: return '本地内存缓存'
    }
  }

  return (
    <div className="space-y-4">
      {/* 标签栏 */}
      <div className="flex gap-2 border-b border-dark-700 pb-3">
        <button
          onClick={() => setActiveTab('dashboard')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'dashboard'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-tachometer-alt mr-2" />
          仪表盘
        </button>
        <button
          onClick={() => setActiveTab('config')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'config'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-cog mr-2" />
          配置管理
        </button>
        {config?.enabled && config?.connected && (
          <button
            onClick={() => setActiveTab('keys')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === 'keys'
                ? 'bg-primary-500/20 text-primary-400'
                : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
            }`}
          >
            <i className="fas fa-key mr-2" />
            键管理
          </button>
        )}
      </div>

      {/* 仪表盘标签页 */}
      {activeTab === 'dashboard' && dashboard && (
        <div className="space-y-4">
          {/* 状态概览卡片 */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* 缓存模式 */}
            <Card className="!p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-dark-400 text-xs mb-1">缓存模式</p>
                  <p className="text-lg font-medium text-dark-100">{getModeText(dashboard.mode)}</p>
                </div>
                <div className={`w-12 h-12 rounded-lg flex items-center justify-center ${
                  dashboard.mode === 'local' ? 'bg-blue-500/10' : 'bg-purple-500/10'
                }`}>
                  <i className={`fas ${dashboard.mode === 'local' ? 'fa-memory' : 'fa-server'} text-xl ${
                    dashboard.mode === 'local' ? 'text-blue-400' : 'text-purple-400'
                  }`} />
                </div>
              </div>
              {dashboard.version && (
                <p className="text-dark-500 text-xs mt-2">版本: {dashboard.version}</p>
              )}
            </Card>

            {/* 连接状态 */}
            <Card className="!p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-dark-400 text-xs mb-1">连接状态</p>
                  <p className={`text-lg font-medium ${
                    dashboard.status === 'connected' ? 'text-green-400' :
                    dashboard.status === 'degraded' ? 'text-yellow-400' : 'text-red-400'
                  }`}>
                    {getStatusText(dashboard.status)}
                  </p>
                </div>
                <div className={`w-12 h-12 rounded-lg flex items-center justify-center ${getStatusColor(dashboard.status)}`}>
                  <i className={`fas fa-plug text-xl ${
                    dashboard.status === 'connected' ? 'text-green-400' :
                    dashboard.status === 'degraded' ? 'text-yellow-400' : 'text-red-400'
                  }`} />
                </div>
              </div>
              <p className="text-dark-500 text-xs mt-2">运行时间: {dashboard.uptime}</p>
            </Card>

            {/* 命中率 */}
            <Card className="!p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-dark-400 text-xs mb-1">缓存命中率</p>
                  <p className={`text-lg font-medium ${
                    dashboard.hit_rate >= 90 ? 'text-green-400' :
                    dashboard.hit_rate >= 70 ? 'text-yellow-400' : 'text-red-400'
                  }`}>
                    {dashboard.hit_rate_str || '0.00%'}
                  </p>
                </div>
                <div className={`w-12 h-12 rounded-lg flex items-center justify-center ${
                  dashboard.hit_rate >= 90 ? 'bg-green-500/10' :
                  dashboard.hit_rate >= 70 ? 'bg-yellow-500/10' : 'bg-red-500/10'
                }`}>
                  <i className={`fas fa-bullseye text-xl ${
                    dashboard.hit_rate >= 90 ? 'text-green-400' :
                    dashboard.hit_rate >= 70 ? 'text-yellow-400' : 'text-red-400'
                  }`} />
                </div>
              </div>
              <p className="text-dark-500 text-xs mt-2">
                命中: {dashboard.hits.toLocaleString()} / 未命中: {dashboard.misses.toLocaleString()}
              </p>
            </Card>

            {/* 键数量 */}
            <Card className="!p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-dark-400 text-xs mb-1">缓存键数量</p>
                  <p className="text-lg font-medium text-primary-400">
                    {dashboard.mode === 'local' 
                      ? dashboard.local_cache_size.toLocaleString()
                      : dashboard.keys_count.toLocaleString()
                    }
                  </p>
                </div>
                <div className="w-12 h-12 rounded-lg flex items-center justify-center bg-primary-500/10">
                  <i className="fas fa-database text-xl text-primary-400" />
                </div>
              </div>
              <p className="text-dark-500 text-xs mt-2">
                每秒操作: {dashboard.ops_per_second.toFixed(2)}
              </p>
            </Card>
          </div>

          {/* 详细统计 */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* 内存使用 */}
            <Card title="💾 内存使用">
              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 bg-dark-700/50 rounded-lg">
                  <p className="text-dark-400 text-xs">已用内存</p>
                  <p className="text-lg font-medium text-blue-400">
                    {dashboard.mode === 'local' ? dashboard.local_cache_memory : dashboard.memory_used || 'N/A'}
                  </p>
                </div>
                {dashboard.mode !== 'local' && (
                  <>
                    <div className="p-3 bg-dark-700/50 rounded-lg">
                      <p className="text-dark-400 text-xs">峰值内存</p>
                      <p className="text-lg font-medium text-purple-400">{dashboard.memory_peak || 'N/A'}</p>
                    </div>
                    <div className="p-3 bg-dark-700/50 rounded-lg">
                      <p className="text-dark-400 text-xs">内存限制</p>
                      <p className="text-lg font-medium text-dark-200">{dashboard.memory_limit || '无限制'}</p>
                    </div>
                    <div className="p-3 bg-dark-700/50 rounded-lg">
                      <p className="text-dark-400 text-xs">淘汰策略</p>
                      <p className="text-lg font-medium text-dark-200">{dashboard.memory_policy || 'noeviction'}</p>
                    </div>
                  </>
                )}
              </div>
            </Card>

            {/* 键空间统计 */}
            <Card title="🔑 键空间统计">
              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 bg-dark-700/50 rounded-lg">
                  <p className="text-dark-400 text-xs">总键数</p>
                  <p className="text-lg font-medium text-primary-400">
                    {dashboard.mode === 'local' 
                      ? dashboard.local_cache_size.toLocaleString()
                      : dashboard.keys_count.toLocaleString()
                    }
                  </p>
                </div>
                {dashboard.mode !== 'local' && (
                  <>
                    <div className="p-3 bg-dark-700/50 rounded-lg">
                      <p className="text-dark-400 text-xs">已过期删除</p>
                      <p className="text-lg font-medium text-yellow-400">{dashboard.expired_keys.toLocaleString()}</p>
                    </div>
                    <div className="p-3 bg-dark-700/50 rounded-lg">
                      <p className="text-dark-400 text-xs">被淘汰键数</p>
                      <p className="text-lg font-medium text-red-400">{dashboard.evicted_keys.toLocaleString()}</p>
                    </div>
                  </>
                )}
                <div className="p-3 bg-dark-700/50 rounded-lg">
                  <p className="text-dark-400 text-xs">故障转移次数</p>
                  <p className="text-lg font-medium text-orange-400">{dashboard.failovers}</p>
                </div>
              </div>
            </Card>

            {/* 连接信息（仅Redis模式） */}
            {dashboard.mode !== 'local' && (
              <Card title="🔌 连接信息">
                <div className="grid grid-cols-2 gap-4">
                  <div className="p-3 bg-dark-700/50 rounded-lg">
                    <p className="text-dark-400 text-xs">当前连接</p>
                    <p className="text-lg font-medium text-green-400">{dashboard.connected_clients}</p>
                  </div>
                  <div className="p-3 bg-dark-700/50 rounded-lg">
                    <p className="text-dark-400 text-xs">最大连接</p>
                    <p className="text-lg font-medium text-dark-200">{dashboard.max_clients}</p>
                  </div>
                  <div className="p-3 bg-dark-700/50 rounded-lg">
                    <p className="text-dark-400 text-xs">阻塞客户端</p>
                    <p className="text-lg font-medium text-yellow-400">{dashboard.blocked_clients}</p>
                  </div>
                  <div className="p-3 bg-dark-700/50 rounded-lg">
                    <p className="text-dark-400 text-xs">角色</p>
                    <p className="text-lg font-medium text-dark-200">{dashboard.role || 'master'}</p>
                  </div>
                </div>
              </Card>
            )}

            {/* 持久化状态（仅Redis模式） */}
            {dashboard.mode !== 'local' && (
              <Card title="💽 持久化状态">
                <div className="grid grid-cols-2 gap-4">
                  <div className="p-3 bg-dark-700/50 rounded-lg">
                    <p className="text-dark-400 text-xs">RDB</p>
                    <p className={`text-lg font-medium ${dashboard.rdb_enabled ? 'text-green-400' : 'text-dark-500'}`}>
                      {dashboard.rdb_enabled ? '已启用' : '未启用'}
                    </p>
                  </div>
                  <div className="p-3 bg-dark-700/50 rounded-lg">
                    <p className="text-dark-400 text-xs">AOF</p>
                    <p className={`text-lg font-medium ${dashboard.aof_enabled ? 'text-green-400' : 'text-dark-500'}`}>
                      {dashboard.aof_enabled ? '已启用' : '未启用'}
                    </p>
                  </div>
                  <div className="p-3 bg-dark-700/50 rounded-lg">
                    <p className="text-dark-400 text-xs">最后保存时间</p>
                    <p className="text-sm font-medium text-dark-200">{dashboard.last_save_time || 'N/A'}</p>
                  </div>
                  <div className="p-3 bg-dark-700/50 rounded-lg">
                    <p className="text-dark-400 text-xs">保存状态</p>
                    <p className={`text-lg font-medium ${dashboard.last_save_status === 'ok' ? 'text-green-400' : 'text-dark-200'}`}>
                      {dashboard.last_save_status || 'N/A'}
                    </p>
                  </div>
                </div>
              </Card>
            )}
          </div>

          {/* 错误信息 */}
          {dashboard.last_error && (
            <Card className="!border-red-500/20 !bg-red-500/5">
              <div className="flex items-start gap-3">
                <i className="fas fa-exclamation-triangle text-red-400 mt-1" />
                <div>
                  <p className="text-red-400 font-medium">最后错误</p>
                  <p className="text-dark-300 text-sm mt-1">{dashboard.last_error}</p>
                  {dashboard.last_error_time && (
                    <p className="text-dark-500 text-xs mt-1">时间: {dashboard.last_error_time}</p>
                  )}
                </div>
              </div>
            </Card>
          )}

          {/* 操作按钮 */}
          <div className="flex gap-2 pt-2">
            <Button variant="secondary" onClick={loadDashboard}>
              <i className="fas fa-sync-alt mr-2" />刷新数据
            </Button>
            <Button variant="secondary" onClick={handleRefreshCache}>
              <i className="fas fa-redo mr-2" />重新连接
            </Button>
            <Button variant="danger" onClick={() => setShowFlushModal(true)}>
              <i className="fas fa-trash-alt mr-2" />清空缓存
            </Button>
          </div>
        </div>
      )}

      {/* 配置管理标签页 */}
      {activeTab === 'config' && (
        <Card>
          <div className="space-y-4">
            {/* 启用开关 */}
            <div className="flex items-center justify-between">
              <div>
                <label className="text-sm font-medium text-dark-300">启用 Redis</label>
                <p className="text-dark-500 text-xs">关闭后将使用本地内存缓存</p>
              </div>
              <button
                onClick={() => setForm({ ...form, enabled: !form.enabled })}
                className={`relative w-12 h-6 rounded-full transition-colors ${form.enabled ? 'bg-primary-500' : 'bg-dark-600'}`}
              >
                <span className={`absolute top-1 w-4 h-4 bg-white rounded-full transition-transform ${form.enabled ? 'left-7' : 'left-1'}`} />
              </button>
            </div>

            {form.enabled && (
              <>
                {/* 运行模式 */}
                <div>
                  <label className="block text-sm font-medium text-dark-300 mb-1">运行模式</label>
                  <select
                    className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100"
                    value={form.mode}
                    onChange={(e) => setForm({ ...form, mode: e.target.value })}
                  >
                    <option value="standalone">单机模式</option>
                    <option value="sentinel">哨兵模式</option>
                    <option value="cluster">集群模式</option>
                  </select>
                </div>

                {/* 单机/集群模式地址 */}
                {form.mode !== 'sentinel' && (
                  <Input
                    label="Redis 地址"
                    value={form.address}
                    onChange={(e) => setForm({ ...form, address: e.target.value })}
                    placeholder="localhost:6379"
                  />
                )}

                {/* 哨兵模式配置 */}
                {form.mode === 'sentinel' && (
                  <>
                    <Input
                      label="Master 名称"
                      value={form.master_name}
                      onChange={(e) => setForm({ ...form, master_name: e.target.value })}
                      placeholder="mymaster"
                    />
                    <div>
                      <label className="block text-sm font-medium text-dark-300 mb-1">哨兵节点地址</label>
                      <textarea
                        className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100 h-20"
                        value={form.sentinel_addrs}
                        onChange={(e) => setForm({ ...form, sentinel_addrs: e.target.value })}
                        placeholder="每行一个地址，如：&#10;192.168.1.1:26379&#10;192.168.1.2:26379"
                      />
                    </div>
                  </>
                )}

                {/* 认证 */}
                <div className="grid grid-cols-2 gap-4">
                  <Input
                    label="密码"
                    type="password"
                    value={form.password}
                    onChange={(e) => setForm({ ...form, password: e.target.value })}
                    placeholder="留空保持不变"
                  />
                  <Input
                    label="数据库编号"
                    type="number"
                    value={String(form.database)}
                    onChange={(e) => setForm({ ...form, database: parseInt(e.target.value) || 0 })}
                  />
                </div>

                {/* 键前缀 */}
                <Input
                  label="键前缀"
                  value={form.key_prefix}
                  onChange={(e) => setForm({ ...form, key_prefix: e.target.value })}
                  placeholder="user:"
                />

                {/* 连接池配置 */}
                <div className="border-t border-dark-700 pt-4 mt-4">
                  <h3 className="text-sm font-medium text-dark-300 mb-3">连接池配置</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <Input
                      label="连接池大小"
                      type="number"
                      value={String(form.pool_size)}
                      onChange={(e) => setForm({ ...form, pool_size: parseInt(e.target.value) || 10 })}
                    />
                    <Input
                      label="最小空闲连接"
                      type="number"
                      value={String(form.min_idle_conns)}
                      onChange={(e) => setForm({ ...form, min_idle_conns: parseInt(e.target.value) || 5 })}
                    />
                  </div>
                </div>

                {/* 超时配置 */}
                <div className="border-t border-dark-700 pt-4 mt-4">
                  <h3 className="text-sm font-medium text-dark-300 mb-3">超时配置（秒）</h3>
                  <div className="grid grid-cols-3 gap-4">
                    <Input
                      label="连接超时"
                      type="number"
                      value={String(form.dial_timeout)}
                      onChange={(e) => setForm({ ...form, dial_timeout: parseInt(e.target.value) || 5 })}
                    />
                    <Input
                      label="读取超时"
                      type="number"
                      value={String(form.read_timeout)}
                      onChange={(e) => setForm({ ...form, read_timeout: parseInt(e.target.value) || 3 })}
                    />
                    <Input
                      label="写入超时"
                      type="number"
                      value={String(form.write_timeout)}
                      onChange={(e) => setForm({ ...form, write_timeout: parseInt(e.target.value) || 3 })}
                    />
                  </div>
                </div>

                {/* 高级选项 */}
                <div className="border-t border-dark-700 pt-4 mt-4">
                  <h3 className="text-sm font-medium text-dark-300 mb-3">高级选项</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <Input
                      label="最大重试次数"
                      type="number"
                      value={String(form.max_retries)}
                      onChange={(e) => setForm({ ...form, max_retries: parseInt(e.target.value) || 3 })}
                    />
                    <div className="flex items-center justify-between">
                      <label className="text-sm font-medium text-dark-300">启用 TLS</label>
                      <button
                        onClick={() => setForm({ ...form, tls_enabled: !form.tls_enabled })}
                        className={`relative w-12 h-6 rounded-full transition-colors ${form.tls_enabled ? 'bg-primary-500' : 'bg-dark-600'}`}
                      >
                        <span className={`absolute top-1 w-4 h-4 bg-white rounded-full transition-transform ${form.tls_enabled ? 'left-7' : 'left-1'}`} />
                      </button>
                    </div>
                  </div>
                </div>
              </>
            )}

            {/* 操作按钮 */}
            <div className="flex gap-2 pt-4 border-t border-dark-700">
              {form.enabled && (
                <Button variant="secondary" onClick={handleTest} disabled={testing}>
                  {testing ? '测试中...' : '测试连接'}
                </Button>
              )}
              <Button onClick={handleSave} disabled={saving}>
                {saving ? '保存中...' : '保存配置'}
              </Button>
            </div>
          </div>
        </Card>
      )}

      {/* 键管理标签页 */}
      {activeTab === 'keys' && (
        <div className="space-y-4">
          {/* 搜索栏 */}
          <Card>
            <div className="flex gap-2">
              <Input
                value={keyPattern}
                onChange={(e) => setKeyPattern(e.target.value)}
                placeholder="输入匹配模式，如 user:* 或 *"
                className="flex-1"
              />
              <Button onClick={loadKeys} disabled={keysLoading}>
                {keysLoading ? '搜索中...' : '搜索'}
              </Button>
            </div>
            <p className="text-dark-500 text-xs mt-2">
              使用 * 匹配任意字符，如 user:* 匹配所有以 user: 开头的键
            </p>
          </Card>

          {/* 键列表 */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <Card title={`🔑 键列表 (${keys.length})`}>
              <div className="max-h-96 overflow-y-auto space-y-1">
                {keys.length === 0 ? (
                  <p className="text-dark-500 text-center py-4">没有找到匹配的键</p>
                ) : (
                  keys.map((key) => (
                    <div
                      key={key}
                      className={`flex items-center justify-between p-2 rounded-lg cursor-pointer transition-colors ${
                        selectedKey === key ? 'bg-primary-500/20' : 'hover:bg-dark-700/50'
                      }`}
                      onClick={() => {
                        setSelectedKey(key)
                        loadKeyInfo(key)
                      }}
                    >
                      <span className="text-dark-200 text-sm truncate flex-1 mr-2">{key}</span>
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleDeleteKey(key)
                        }}
                        className="text-red-400 hover:text-red-300 p-1"
                        title="删除"
                      >
                        <i className="fas fa-trash-alt text-xs" />
                      </button>
                    </div>
                  ))
                )}
              </div>
            </Card>

            {/* 键详情 */}
            <Card title="📋 键详情">
              {selectedKey && keyInfo ? (
                <div className="space-y-4">
                  <div>
                    <label className="block text-dark-400 text-xs mb-1">键名</label>
                    <p className="text-dark-200 font-mono text-sm bg-dark-700/50 p-2 rounded break-all">{selectedKey}</p>
                  </div>
                  <div>
                    <label className="block text-dark-400 text-xs mb-1">过期时间</label>
                    <p className="text-dark-200 text-sm">{keyInfo.ttl}</p>
                  </div>
                  <div>
                    <label className="block text-dark-400 text-xs mb-1">值</label>
                    <pre className="text-dark-200 font-mono text-xs bg-dark-700/50 p-2 rounded overflow-auto max-h-48">
                      {typeof keyInfo.value === 'object' 
                        ? JSON.stringify(keyInfo.value, null, 2)
                        : String(keyInfo.value)
                      }
                    </pre>
                  </div>
                  <Button variant="danger" size="sm" onClick={() => handleDeleteKey(selectedKey)}>
                    <i className="fas fa-trash-alt mr-2" />删除此键
                  </Button>
                </div>
              ) : (
                <p className="text-dark-500 text-center py-8">选择一个键查看详情</p>
              )}
            </Card>
          </div>
        </div>
      )}

      {/* 本地缓存提示 */}
      {!config?.enabled && activeTab === 'dashboard' && (
        <Card>
          <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
            <p className="text-blue-400 text-sm">
              <i className="fas fa-info-circle mr-2" />
              当前使用本地内存缓存。本地缓存在程序重启后会丢失，适合单节点部署。
              如需多节点部署或持久化缓存，请启用 Redis。
            </p>
          </div>
        </Card>
      )}

      {/* 清空缓存确认弹窗 */}
      <Modal
        isOpen={showFlushModal}
        onClose={() => { setShowFlushModal(false); setFlushConfirmText('') }}
        title="⚠️ 清空缓存"
      >
        <div className="space-y-4">
          <div className="p-4 bg-yellow-500/10 border border-yellow-500/20 rounded-lg">
            <p className="text-yellow-400 text-sm font-medium mb-2">注意</p>
            <ul className="text-yellow-400/80 text-sm space-y-1 list-disc list-inside">
              <li>清空缓存会删除所有缓存数据</li>
              <li>用户会话不会受影响（存储在数据库中）</li>
              <li>缓存会在下次访问时自动重建</li>
              <li>短时间内可能会增加数据库负载</li>
            </ul>
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">
              请输入确认文字：<span className="text-yellow-400">确认清空所有缓存</span>
            </label>
            <input
              type="text"
              value={flushConfirmText}
              onChange={(e) => setFlushConfirmText(e.target.value)}
              placeholder="请输入上方黄色文字"
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100"
            />
          </div>
          <div className="flex gap-2 justify-end pt-2">
            <Button variant="secondary" onClick={() => { setShowFlushModal(false); setFlushConfirmText('') }}>
              取消
            </Button>
            <Button
              variant="danger"
              onClick={handleFlushCache}
              disabled={flushLoading || flushConfirmText !== '确认清空所有缓存'}
            >
              {flushLoading ? '清空中...' : '确认清空'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
