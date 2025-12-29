'use client'

import { useState, useEffect } from 'react'
import { Button, Card } from '@/components/ui'
import { apiGet } from '@/lib/api'

/**
 * 统计卡片组件
 * 支持移动端响应式布局
 */
function StatCard({ icon, value, label }: { icon: string; value: string | number; label: string }) {
  return (
    <div className="bg-dark-800/50 rounded-xl p-4 sm:p-6 border border-dark-700/50">
      <div className="text-2xl sm:text-3xl mb-2">{icon}</div>
      <div className="text-xl sm:text-2xl font-bold text-dark-100">{value}</div>
      <div className="text-dark-500 text-xs sm:text-sm">{label}</div>
    </div>
  )
}

/**
 * 仪表盘页面
 * 支持移动端响应式布局
 */
export function DashboardPage() {
  const [data, setData] = useState<{
    db_connected: boolean
    stats: { total_orders: number; paid_orders: number; total_revenue: number; today_orders: number }
  } | null>(null)
  const [loading, setLoading] = useState(true)
  const [chartDays, setChartDays] = useState(7)
  const [chartData, setChartData] = useState<{ date: string; order_count: number; revenue: number }[]>([])

  useEffect(() => {
    loadDashboard()
  }, [])

  useEffect(() => {
    loadChart()
  }, [chartDays])

  const loadDashboard = async () => {
    const res = await apiGet<typeof data>('/api/admin/dashboard')
    if (res.success) setData(res as typeof data)
    setLoading(false)
  }

  const loadChart = async () => {
    const res = await apiGet<{ stats: typeof chartData }>(`/api/admin/stats/chart?days=${chartDays}`)
    if (res.success && res.stats) setChartData(res.stats)
  }

  if (loading) {
    return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>
  }

  if (!data?.db_connected) {
    return (
      <Card>
        <div className="text-center py-8">
          <div className="text-4xl mb-4">⚠️</div>
          <h3 className="text-lg font-medium text-dark-100 mb-2">数据库未连接</h3>
          <p className="text-dark-400">请先前往数据库配置页面配置数据库连接</p>
        </div>
      </Card>
    )
  }

  const stats = data?.stats || { total_orders: 0, paid_orders: 0, total_revenue: 0, today_orders: 0 }

  return (
    <div className="space-y-4 sm:space-y-6">
      {/* 主要统计 - 移动端2列，桌面端4列 */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
        <StatCard icon="📦" value={stats.total_orders} label="总订单数" />
        <StatCard icon="✅" value={stats.paid_orders} label="已完成订单" />
        <StatCard icon="💰" value={`¥${stats.total_revenue.toFixed(2)}`} label="总收入" />
        <StatCard icon="📈" value={stats.today_orders} label="今日订单" />
      </div>

      {/* 订单趋势图表 */}
      <Card title="📊 订单趋势">
        <div className="flex flex-wrap gap-2 mb-4">
          {[7, 14, 30].map((days) => (
            <Button key={days} size="sm" variant={chartDays === days ? 'primary' : 'secondary'} onClick={() => setChartDays(days)}>
              近{days}天
            </Button>
          ))}
        </div>
        <div className="h-48 sm:h-64">
          {chartData.length === 0 ? (
            <div className="h-full flex items-center justify-center text-dark-500">暂无数据</div>
          ) : (
            <div className="h-full flex items-end gap-1 sm:gap-2">
              {chartData.map((item, index) => {
                const maxOrders = Math.max(...chartData.map((d) => d.order_count), 1)
                const height = (item.order_count / maxOrders) * 100
                return (
                  <div key={index} className="flex-1 flex flex-col items-center min-w-0">
                    <div 
                      className="w-full bg-primary-500/50 rounded-t transition-all hover:bg-primary-500/70" 
                      style={{ height: `${Math.max(height, 2)}%` }} 
                      title={`订单: ${item.order_count}, 收入: ¥${item.revenue.toFixed(2)}`} 
                    />
                    <div className="text-[10px] sm:text-xs text-dark-500 mt-1 sm:mt-2 truncate w-full text-center">
                      {item.date.slice(5)}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </Card>

      {/* 快捷操作 */}
      <Card title="快捷操作">
        <div className="flex flex-wrap gap-2 sm:gap-3">
          <Button onClick={() => (window.location.hash = 'products')} className="flex-1 sm:flex-none">
            <i className="fas fa-box mr-2 hidden sm:inline" />管理商品
          </Button>
          <Button variant="secondary" onClick={() => (window.location.hash = 'orders')} className="flex-1 sm:flex-none">
            <i className="fas fa-list mr-2 hidden sm:inline" />查看订单
          </Button>
          <Button variant="secondary" onClick={() => (window.location.hash = 'config')} className="flex-1 sm:flex-none">
            <i className="fas fa-cog mr-2 hidden sm:inline" />系统配置
          </Button>
        </div>
      </Card>
    </div>
  )
}
