'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Card, Badge } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime, copyToClipboard } from '@/lib/utils'

/**
 * 卡密接口
 */
interface Kami {
  id: number
  order_no: string
  product_name: string
  kami_code: string
  status: string
  expire_at: string
  created_at: string
  days_remaining: number
}

/**
 * 续费统计接口
 */
interface RenewalStats {
  total_kamis: number
  active_kamis: number
  expiring_soon: number
  expired: number
}

/**
 * 我的卡密标签页
 */
export function KamisTab() {
  const [activeSection, setActiveSection] = useState<'all' | 'expiring' | 'expired'>('all')
  const [kamis, setKamis] = useState<Kami[]>([])
  const [expiringKamis, setExpiringKamis] = useState<Kami[]>([])
  const [expiredKamis, setExpiredKamis] = useState<Kami[]>([])
  const [stats, setStats] = useState<RenewalStats | null>(null)
  const [loading, setLoading] = useState(true)

  // 加载所有卡密
  const loadKamis = async () => {
    const res = await apiGet<{ kamis: Kami[] }>('/api/user/kamis')
    if (res.success && res.kamis) {
      setKamis(res.kamis)
    }
  }

  // 加载即将过期的卡密
  const loadExpiringKamis = async () => {
    const res = await apiGet<{ kamis: Kami[] }>('/api/user/kamis/expiring')
    if (res.success && res.kamis) {
      setExpiringKamis(res.kamis)
    }
  }

  // 加载已过期的卡密
  const loadExpiredKamis = async () => {
    const res = await apiGet<{ kamis: Kami[] }>('/api/user/kamis/expired')
    if (res.success && res.kamis) {
      setExpiredKamis(res.kamis)
    }
  }

  // 加载续费统计
  const loadStats = async () => {
    const res = await apiGet<{ stats: { total: number; active: number; expiring: number; expired: number } }>('/api/user/renewal/stats')
    if (res.success && res.stats) {
      setStats({
        total_kamis: res.stats.total,
        active_kamis: res.stats.active,
        expiring_soon: res.stats.expiring,
        expired: res.stats.expired,
      })
    }
  }

  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      await Promise.all([loadKamis(), loadStats()])
      setLoading(false)
    }
    loadData()
  }, [])

  // 切换分区时加载数据
  useEffect(() => {
    if (activeSection === 'expiring' && expiringKamis.length === 0) {
      loadExpiringKamis()
    } else if (activeSection === 'expired' && expiredKamis.length === 0) {
      loadExpiredKamis()
    }
  }, [activeSection])

  // 复制卡密
  const handleCopyKami = async (code: string) => {
    const success = await copyToClipboard(code)
    if (success) toast.success('已复制到剪贴板')
  }

  // 导出卡密为 CSV
  const handleExportKami = (kami: Kami) => {
    const kamiCodes = kami.kami_code.split('\n').filter(code => code.trim())
    if (kamiCodes.length <= 1) return
    
    // 生成 CSV 内容
    const csvContent = [
      '序号,卡密,商品名称,订单号,到期时间',
      ...kamiCodes.map((code, index) => 
        `${index + 1},"${code.trim()}","${kami.product_name}","${kami.order_no}","${formatDateTime(kami.expire_at)}"`
      )
    ].join('\n')
    
    // 添加 BOM 以支持中文
    const BOM = '\uFEFF'
    const blob = new Blob([BOM + csvContent], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `卡密_${kami.order_no}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    toast.success('卡密已导出')
  }

  // 获取卡密数量
  const getKamiCount = (kamiCode: string) => {
    return kamiCode.split('\n').filter(code => code.trim()).length
  }

  // 请求续费提醒
  const handleRequestReminder = async (orderNo: string) => {
    const res = await apiPost('/api/user/renewal/remind', { order_no: orderNo })
    if (res.success) {
      toast.success('已设置续费提醒，将在到期前通过邮件通知您')
    } else {
      toast.error(res.error || '设置失败')
    }
  }

  // 获取状态徽章
  const getStatusBadge = (kami: Kami) => {
    if (kami.status === 'expired') {
      return <Badge variant="danger">已过期</Badge>
    }
    if (kami.days_remaining <= 7) {
      return <Badge variant="warning">即将过期</Badge>
    }
    return <Badge variant="success">有效</Badge>
  }

  // 获取当前显示的卡密列表
  const getCurrentKamis = () => {
    switch (activeSection) {
      case 'expiring': return expiringKamis
      case 'expired': return expiredKamis
      default: return kamis
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
      {/* 统计卡片 */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-dark-700/30 rounded-xl p-4 border border-dark-600/50">
            <div className="text-2xl font-bold text-dark-100">{stats.total_kamis}</div>
            <div className="text-sm text-dark-400">全部卡密</div>
          </div>
          <div className="bg-dark-700/30 rounded-xl p-4 border border-dark-600/50">
            <div className="text-2xl font-bold text-green-400">{stats.active_kamis}</div>
            <div className="text-sm text-dark-400">有效卡密</div>
          </div>
          <div className="bg-dark-700/30 rounded-xl p-4 border border-dark-600/50">
            <div className="text-2xl font-bold text-yellow-400">{stats.expiring_soon}</div>
            <div className="text-sm text-dark-400">即将过期</div>
          </div>
          <div className="bg-dark-700/30 rounded-xl p-4 border border-dark-600/50">
            <div className="text-2xl font-bold text-red-400">{stats.expired}</div>
            <div className="text-sm text-dark-400">已过期</div>
          </div>
        </div>
      )}

      {/* 分区切换 */}
      <div className="flex gap-2 border-b border-dark-700/50 pb-4">
        <button
          onClick={() => setActiveSection('all')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeSection === 'all'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          全部卡密
        </button>
        <button
          onClick={() => setActiveSection('expiring')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors relative ${
            activeSection === 'expiring'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          即将过期
          {stats && stats.expiring_soon > 0 && (
            <span className="absolute -top-1 -right-1 w-5 h-5 bg-yellow-500 text-white text-xs rounded-full flex items-center justify-center">
              {stats.expiring_soon}
            </span>
          )}
        </button>
        <button
          onClick={() => setActiveSection('expired')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeSection === 'expired'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          已过期
        </button>
      </div>

      {/* 卡密列表 */}
      <Card
        title={activeSection === 'all' ? '全部卡密' : activeSection === 'expiring' ? '即将过期' : '已过期'}
        icon={<i className="fas fa-key" />}
      >
        {getCurrentKamis().length === 0 ? (
          <div className="text-center py-8 text-dark-400">
            {activeSection === 'all' ? '暂无卡密' : activeSection === 'expiring' ? '暂无即将过期的卡密' : '暂无已过期的卡密'}
          </div>
        ) : (
          <div className="space-y-4">
            {getCurrentKamis().map((kami) => {
              const kamiCount = getKamiCount(kami.kami_code)
              const kamiCodes = kami.kami_code.split('\n').filter(code => code.trim())
              
              return (
                <div
                  key={kami.id}
                  className={`p-4 rounded-xl border ${
                    kami.status === 'expired'
                      ? 'bg-red-500/5 border-red-500/20'
                      : kami.days_remaining <= 7
                      ? 'bg-yellow-500/5 border-yellow-500/20'
                      : 'bg-dark-700/30 border-dark-600/50'
                  }`}
                >
                  <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <span className="font-medium text-dark-100">{kami.product_name}</span>
                        {getStatusBadge(kami)}
                        {kamiCount > 1 && (
                          <Badge variant="info">{kamiCount}个卡密</Badge>
                        )}
                      </div>
                      <div className="bg-dark-800/50 rounded-lg p-3 mb-3">
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex-1 space-y-2 max-h-40 overflow-y-auto">
                            {kamiCodes.map((code, index) => (
                              <div key={index} className="flex items-center gap-2">
                                {kamiCount > 1 && (
                                  <span className="text-dark-500 text-sm w-6 flex-shrink-0">{index + 1}.</span>
                                )}
                                <span className="font-mono text-primary-400 break-all text-sm flex-1">{code.trim()}</span>
                              </div>
                            ))}
                          </div>
                          <div className="flex flex-col gap-1 flex-shrink-0">
                            <Button size="sm" variant="ghost" onClick={() => handleCopyKami(kami.kami_code)} title="复制全部">
                              <i className="fas fa-copy" />
                            </Button>
                            {kamiCount > 1 && (
                              <Button size="sm" variant="ghost" onClick={() => handleExportKami(kami)} title="导出CSV">
                                <i className="fas fa-download" />
                              </Button>
                            )}
                          </div>
                        </div>
                      </div>
                      <div className="flex flex-wrap gap-4 text-sm text-dark-400">
                        <span>
                          <i className="fas fa-receipt mr-1" />
                          订单: {kami.order_no}
                        </span>
                        <span>
                          <i className="fas fa-calendar mr-1" />
                          到期: {formatDateTime(kami.expire_at)}
                        </span>
                        {kami.status !== 'expired' && (
                          <span className={kami.days_remaining <= 7 ? 'text-yellow-400' : ''}>
                            <i className="fas fa-clock mr-1" />
                            剩余: {kami.days_remaining} 天
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="flex gap-2">
                      {kami.status !== 'expired' && kami.days_remaining <= 30 && (
                        <Button size="sm" variant="secondary" onClick={() => handleRequestReminder(kami.order_no)}>
                          <i className="fas fa-bell mr-1" />
                          设置提醒
                        </Button>
                      )}
                      <a href={`/products/`}>
                        <Button size="sm">
                          <i className="fas fa-redo mr-1" />
                          续费
                        </Button>
                      </a>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </Card>
    </motion.div>
  )
}
