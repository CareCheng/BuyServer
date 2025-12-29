'use client'

import { useState, useEffect } from 'react'
import { Card, Badge } from '@/components/ui'
import { apiGet } from '@/lib/api'

/**
 * 系统信息接口（匹配后端 SystemInfo 结构）
 */
interface SystemInfo {
  go_version: string
  goos: string
  goarch: string
  num_cpu: number
  num_goroutine: number
  uptime: number
  uptime_str: string
}

/**
 * 内存统计接口（匹配后端 MemoryStats 结构）
 */
interface MemoryStats {
  alloc: number
  total_alloc: number
  sys: number
  num_gc: number
  heap_alloc: number
  heap_sys: number
  heap_idle: number
  heap_inuse: number
  stack_inuse: number
  alloc_mb: number
  sys_mb: number
  heap_alloc_mb: number
}

/**
 * 数据库统计接口（匹配后端 DatabaseStats 结构）
 */
interface DatabaseStats {
  total_users: number
  total_orders: number
  total_products: number
  total_tickets: number
  pending_orders: number
  active_sessions: number
  today_orders: number
  today_users: number
  today_revenue: number
}

/**
 * 健康状态接口（匹配后端 GetHealthStatus 返回结构）
 */
interface HealthStatus {
  status: string
  timestamp: string
  database: string
  memory_mb: number
  memory_warning?: boolean
  goroutines: number
  goroutine_warning?: boolean
}

/**
 * 系统监控页面
 */
export function MonitorPage() {
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null)
  const [memoryStats, setMemoryStats] = useState<MemoryStats | null>(null)
  const [databaseStats, setDatabaseStats] = useState<DatabaseStats | null>(null)
  const [healthStatus, setHealthStatus] = useState<HealthStatus | null>(null)
  const [loading, setLoading] = useState(true)

  // 加载数据
  useEffect(() => {
    loadData()
    // 每30秒刷新一次
    const interval = setInterval(loadData, 30000)
    return () => clearInterval(interval)
  }, [])

  const loadData = async () => {
    setLoading(true)
    await Promise.all([
      loadSystemInfo(),
      loadMemoryStats(),
      loadDatabaseStats(),
      loadHealthStatus(),
    ])
    setLoading(false)
  }

  const loadSystemInfo = async () => {
    const res = await apiGet<{ data: SystemInfo }>('/api/admin/monitor/system')
    console.log('系统信息 API 响应:', res)
    if (res.success && res.data) {
      setSystemInfo(res.data)
    }
  }

  const loadMemoryStats = async () => {
    const res = await apiGet<{ data: MemoryStats }>('/api/admin/monitor/memory')
    console.log('内存统计 API 响应:', res)
    if (res.success && res.data) {
      setMemoryStats(res.data)
    }
  }

  const loadDatabaseStats = async () => {
    const res = await apiGet<{ data: DatabaseStats }>('/api/admin/monitor/database')
    console.log('数据库统计 API 响应:', res)
    if (res.success && res.data) {
      setDatabaseStats(res.data)
    }
  }

  const loadHealthStatus = async () => {
    const res = await apiGet<{ data: HealthStatus }>('/api/admin/monitor/health')
    console.log('健康状态 API 响应:', res)
    if (res.success && res.data) {
      setHealthStatus(res.data)
    }
  }

  // 格式化字节
  const formatBytes = (bytes: number | undefined | null) => {
    if (!bytes || bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // 获取状态徽章
  const getStatusBadge = (status: string) => {
    if (status === 'healthy' || status === 'ok' || status === 'connected') {
      return <Badge variant="success">正常</Badge>
    }
    if (status === 'warning') {
      return <Badge variant="warning">警告</Badge>
    }
    return <Badge variant="danger">异常</Badge>
  }

  if (loading && !systemInfo) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 健康状态 */}
      {healthStatus && (
        <Card title="服务状态" icon={<i className="fas fa-heartbeat" />}>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="p-4 bg-dark-700/30 rounded-xl">
              <div className="flex items-center justify-between mb-2">
                <span className="text-dark-400">系统状态</span>
                {getStatusBadge(healthStatus.status)}
              </div>
              {systemInfo && (
                <div className="text-sm text-dark-500">运行时间: {systemInfo.uptime_str || '-'}</div>
              )}
            </div>
            <div className="p-4 bg-dark-700/30 rounded-xl">
              <div className="flex items-center justify-between mb-2">
                <span className="text-dark-400">数据库</span>
                {getStatusBadge(healthStatus.database)}
              </div>
            </div>
            <div className="p-4 bg-dark-700/30 rounded-xl">
              <div className="flex items-center justify-between mb-2">
                <span className="text-dark-400">内存使用</span>
                {healthStatus.memory_warning ? (
                  <Badge variant="warning">警告</Badge>
                ) : (
                  <Badge variant="success">正常</Badge>
                )}
              </div>
              <div className="text-sm text-dark-500">{healthStatus.memory_mb?.toFixed(2) || 0} MB</div>
            </div>
            <div className="p-4 bg-dark-700/30 rounded-xl">
              <div className="flex items-center justify-between mb-2">
                <span className="text-dark-400">Goroutine</span>
                {healthStatus.goroutine_warning ? (
                  <Badge variant="warning">警告</Badge>
                ) : (
                  <Badge variant="success">正常</Badge>
                )}
              </div>
              <div className="text-sm text-dark-500">{healthStatus.goroutines || 0} 个</div>
            </div>
          </div>
        </Card>
      )}

      {/* 系统信息 */}
      {systemInfo && (
        <Card title="系统信息" icon={<i className="fas fa-server" />}>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            <InfoItem label="操作系统" value={systemInfo.goos || '-'} />
            <InfoItem label="架构" value={systemInfo.goarch || '-'} />
            <InfoItem label="Go 版本" value={systemInfo.go_version || '-'} />
            <InfoItem label="CPU 核心" value={(systemInfo.num_cpu || 0).toString()} />
            <InfoItem label="Goroutine 数" value={(systemInfo.num_goroutine || 0).toString()} />
            <InfoItem label="运行时间" value={systemInfo.uptime_str || '-'} />
          </div>
        </Card>
      )}

      {/* 内存统计 */}
      {memoryStats && (
        <Card title="内存统计" icon={<i className="fas fa-memory" />}>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <InfoItem label="已分配内存" value={`${memoryStats.alloc_mb?.toFixed(2) || 0} MB`} />
            <InfoItem label="累计分配" value={formatBytes(memoryStats.total_alloc)} />
            <InfoItem label="系统内存" value={`${memoryStats.sys_mb?.toFixed(2) || 0} MB`} />
            <InfoItem label="GC 次数" value={(memoryStats.num_gc || 0).toString()} />
            <InfoItem label="堆内存分配" value={`${memoryStats.heap_alloc_mb?.toFixed(2) || 0} MB`} />
            <InfoItem label="堆内存系统" value={formatBytes(memoryStats.heap_sys)} />
            <InfoItem label="堆内存空闲" value={formatBytes(memoryStats.heap_idle)} />
            <InfoItem label="堆内存使用" value={formatBytes(memoryStats.heap_inuse)} />
          </div>
        </Card>
      )}

      {/* 数据库统计 */}
      {databaseStats && (
        <Card title="数据库统计" icon={<i className="fas fa-database" />}>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <InfoItem label="总用户数" value={(databaseStats.total_users || 0).toLocaleString()} />
            <InfoItem label="总订单数" value={(databaseStats.total_orders || 0).toLocaleString()} />
            <InfoItem label="总商品数" value={(databaseStats.total_products || 0).toLocaleString()} />
            <InfoItem label="总工单数" value={(databaseStats.total_tickets || 0).toLocaleString()} />
            <InfoItem label="待支付订单" value={(databaseStats.pending_orders || 0).toLocaleString()} />
            <InfoItem label="活跃会话" value={(databaseStats.active_sessions || 0).toLocaleString()} />
            <InfoItem label="今日订单" value={(databaseStats.today_orders || 0).toLocaleString()} />
            <InfoItem label="今日新用户" value={(databaseStats.today_users || 0).toLocaleString()} />
          </div>
          {/* 今日收入单独显示 */}
          <div className="mt-4 p-4 bg-gradient-to-r from-green-500/10 to-emerald-500/10 rounded-xl border border-green-500/20">
            <div className="flex items-center justify-between">
              <span className="text-dark-400">今日收入</span>
              <span className="text-2xl font-bold text-green-400">¥{(databaseStats.today_revenue || 0).toFixed(2)}</span>
            </div>
          </div>
        </Card>
      )}
    </div>
  )
}

/**
 * 信息项组件
 */
function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="p-3 bg-dark-700/30 rounded-lg">
      <div className="text-dark-500 text-sm mb-1">{label}</div>
      <div className="text-dark-100 font-medium">{value}</div>
    </div>
  )
}
