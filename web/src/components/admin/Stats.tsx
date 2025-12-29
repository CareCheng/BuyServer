'use client'

import { useState, useEffect } from 'react'
import { Card } from '@/components/ui'
import { apiGet } from '@/lib/api'

/**
 * 销售统计接口（匹配后端 SalesStats 结构）
 */
interface SalesStats {
  total_revenue: number      // 总收入
  total_orders: number       // 总订单数
  paid_orders: number        // 已支付订单数
  completed_orders: number   // 已完成订单数
  cancelled_orders: number   // 已取消订单数
  refunded_orders: number    // 已退款订单数
  avg_order_value: number    // 平均订单金额
  conversion_rate: number    // 转化率
}

/**
 * 商品排行接口（匹配后端 ProductSalesData 结构）
 */
interface ProductRanking {
  product_id: number
  product_name: string
  sales_count: number
  revenue: number
}

/**
 * 支付方式统计接口（匹配后端 PaymentMethodStats 结构）
 */
interface PaymentStats {
  method: string
  count: number
  revenue: number
  percent: number
}

/**
 * 用户统计接口（匹配后端 UserStats 结构）
 */
interface UserStats {
  total_users: number        // 总用户数
  active_users: number       // 活跃用户数
  new_users_today: number    // 今日新增
  new_users_week: number     // 本周新增
  new_users_month: number    // 本月新增
  verified_users: number     // 已验证邮箱用户
  enable_2fa_users: number   // 启用2FA用户
  retention_rate: number     // 留存率
}

/**
 * 统计报表页面
 */
export function StatsPage() {
  const [salesStats, setSalesStats] = useState<SalesStats | null>(null)
  const [productRanking, setProductRanking] = useState<ProductRanking[]>([])
  const [paymentStats, setPaymentStats] = useState<PaymentStats[]>([])
  const [userStats, setUserStats] = useState<UserStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'sales' | 'products' | 'payments' | 'users'>('sales')

  // 加载数据
  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      await Promise.all([
        loadSalesStats(),
        loadProductRanking(),
        loadPaymentStats(),
        loadUserStats(),
      ])
      setLoading(false)
    }
    loadData()
  }, [])

  const loadSalesStats = async () => {
    const res = await apiGet<{ data: SalesStats }>('/api/admin/stats/sales')
    console.log('销售统计 API 响应:', res)
    if (res.success && res.data) {
      setSalesStats(res.data)
    } else {
      // 即使没有数据也设置默认值，确保页面能显示
      setSalesStats({
        total_revenue: 0, total_orders: 0, paid_orders: 0, completed_orders: 0,
        cancelled_orders: 0, refunded_orders: 0, avg_order_value: 0, conversion_rate: 0
      })
    }
  }

  const loadProductRanking = async () => {
    const res = await apiGet<{ data: ProductRanking[] }>('/api/admin/stats/product-ranking')
    console.log('商品排行 API 响应:', res)
    if (res.success && res.data) {
      setProductRanking(res.data)
    }
  }

  const loadPaymentStats = async () => {
    const res = await apiGet<{ data: PaymentStats[] }>('/api/admin/stats/payment')
    console.log('支付统计 API 响应:', res)
    if (res.success && res.data) {
      setPaymentStats(res.data)
    }
  }

  const loadUserStats = async () => {
    const res = await apiGet<{ data: UserStats }>('/api/admin/stats/users')
    console.log('用户统计 API 响应:', res)
    if (res.success && res.data) {
      setUserStats(res.data)
    } else {
      // 即使没有数据也设置默认值，确保页面能显示
      setUserStats({
        total_users: 0, active_users: 0, new_users_today: 0, new_users_week: 0,
        new_users_month: 0, verified_users: 0, enable_2fa_users: 0, retention_rate: 0
      })
    }
  }

  // 获取支付方式名称
  const getPaymentMethodName = (method: string) => {
    const names: Record<string, string> = {
      paypal: 'PayPal',
      alipay: '支付宝',
      wechat: '微信支付',
      stripe: 'Stripe',
      usdt: 'USDT',
      balance: '余额支付',
    }
    return names[method] || method
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 标签切换 */}
      <div className="flex gap-2 border-b border-dark-700/50 pb-4">
        {[
          { id: 'sales', label: '销售统计', icon: 'fa-chart-line' },
          { id: 'products', label: '商品排行', icon: 'fa-trophy' },
          { id: 'payments', label: '支付分析', icon: 'fa-credit-card' },
          { id: 'users', label: '用户统计', icon: 'fa-users' },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id as typeof activeTab)}
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

      {/* 销售统计 */}
      {activeTab === 'sales' && salesStats && (
        <div className="space-y-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard title="总收入" value={`¥${(salesStats.total_revenue || 0).toFixed(2)}`} subValue={`${salesStats.total_orders || 0} 单`} color="blue" />
            <StatCard title="已支付订单" value={`${salesStats.paid_orders || 0}`} subValue={`转化率 ${(salesStats.conversion_rate || 0).toFixed(1)}%`} color="green" />
            <StatCard title="已完成订单" value={`${salesStats.completed_orders || 0}`} subValue={`已取消 ${salesStats.cancelled_orders || 0}`} color="purple" />
            <StatCard title="平均订单金额" value={`¥${(salesStats.avg_order_value || 0).toFixed(2)}`} subValue={`已退款 ${salesStats.refunded_orders || 0}`} color="amber" />
          </div>
        </div>
      )}

      {/* 商品排行 */}
      {activeTab === 'products' && (
        <Card title="商品销量排行" icon={<i className="fas fa-trophy" />}>
          {productRanking.length === 0 ? (
            <div className="text-center py-8 text-dark-400">暂无数据</div>
          ) : (
            <div className="space-y-3">
              {productRanking.map((product, index) => (
                <div key={product.product_id} className="flex items-center gap-4 p-3 bg-dark-700/30 rounded-lg">
                  <div className={`w-8 h-8 rounded-full flex items-center justify-center font-bold ${
                    index === 0 ? 'bg-yellow-500/20 text-yellow-400' :
                    index === 1 ? 'bg-gray-400/20 text-gray-400' :
                    index === 2 ? 'bg-amber-600/20 text-amber-600' :
                    'bg-dark-600/50 text-dark-400'
                  }`}>
                    {index + 1}
                  </div>
                  <div className="flex-1">
                    <div className="font-medium text-dark-100">{product.product_name}</div>
                    <div className="text-sm text-dark-500">销量: {product.sales_count}</div>
                  </div>
                  <div className="text-right">
                    <div className="text-primary-400 font-medium">¥{(product.revenue || 0).toFixed(2)}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      )}

      {/* 支付分析 */}
      {activeTab === 'payments' && (
        <Card title="支付方式分析" icon={<i className="fas fa-credit-card" />}>
          {paymentStats.length === 0 ? (
            <div className="text-center py-8 text-dark-400">暂无数据</div>
          ) : (
            <div className="space-y-4">
              {paymentStats.map((stat) => (
                <div key={stat.method} className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-dark-100">{getPaymentMethodName(stat.method)}</span>
                    <span className="text-dark-400">{stat.count || 0} 笔 · ¥{(stat.revenue || 0).toFixed(2)}</span>
                  </div>
                  <div className="h-2 bg-dark-700 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-primary-500 rounded-full transition-all"
                      style={{ width: `${stat.percent || 0}%` }}
                    />
                  </div>
                  <div className="text-right text-sm text-dark-500">{(stat.percent || 0).toFixed(1)}%</div>
                </div>
              ))}
            </div>
          )}
        </Card>
      )}

      {/* 用户统计 */}
      {activeTab === 'users' && userStats && (
        <div className="space-y-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard title="总用户数" value={(userStats.total_users || 0).toString()} color="blue" />
            <StatCard title="今日新增" value={(userStats.new_users_today || 0).toString()} color="green" />
            <StatCard title="本周新增" value={(userStats.new_users_week || 0).toString()} color="purple" />
            <StatCard title="本月新增" value={(userStats.new_users_month || 0).toString()} color="amber" />
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard title="活跃用户" value={(userStats.active_users || 0).toString()} subValue="30天内登录" color="cyan" />
            <StatCard title="已验证邮箱" value={(userStats.verified_users || 0).toString()} color="blue" />
            <StatCard title="启用2FA" value={(userStats.enable_2fa_users || 0).toString()} color="green" />
            <StatCard title="留存率" value={`${(userStats.retention_rate || 0).toFixed(1)}%`} color="purple" />
          </div>
        </div>
      )}
    </div>
  )
}

/**
 * 统计卡片组件
 */
function StatCard({ title, value, subValue, color }: { title: string; value: string; subValue?: string; color: string }) {
  const colorClasses: Record<string, string> = {
    blue: 'from-blue-500/20 to-blue-600/10 border-blue-500/30',
    green: 'from-green-500/20 to-green-600/10 border-green-500/30',
    purple: 'from-purple-500/20 to-purple-600/10 border-purple-500/30',
    amber: 'from-amber-500/20 to-amber-600/10 border-amber-500/30',
    cyan: 'from-cyan-500/20 to-cyan-600/10 border-cyan-500/30',
  }

  return (
    <div className={`bg-gradient-to-br ${colorClasses[color]} rounded-xl p-4 border`}>
      <div className="text-dark-400 text-sm mb-1">{title}</div>
      <div className="text-2xl font-bold text-dark-100">{value}</div>
      {subValue && <div className="text-dark-500 text-sm mt-1">{subValue}</div>}
    </div>
  )
}
