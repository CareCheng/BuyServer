'use client'

import { useState, useEffect, ChangeEvent } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Badge, Modal, Input } from '@/components/ui'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { formatDateTime, formatMoney } from '@/lib/utils'

/**
 * 用户余额接口
 */
interface UserBalance {
  id: number
  user_id: number
  username: string
  balance: number
  frozen_balance: number
  total_recharge: number
  total_consume: number
  status: number
  updated_at: string
}

/**
 * 余额日志接口
 */
interface BalanceLog {
  id: number
  user_id: number
  username: string
  type: string
  amount: number
  balance_before: number
  balance_after: number
  remark: string
  created_at: string
}

/**
 * 余额配置接口
 */
interface BalanceConfig {
  min_recharge_amount: number
  max_recharge_amount: number
  max_daily_recharge: number
  max_balance_limit: number
  large_recharge_threshold: number
  large_consume_threshold: number
  frequent_recharge_count: number
  frequent_consume_count: number
  large_admin_adjust_threshold: number
}

/**
 * 充值优惠活动接口
 */
interface RechargePromo {
  id: number
  name: string
  description: string
  promo_type: 'discount' | 'bonus' | 'percent'
  min_amount: number
  max_amount: number
  value: number
  max_bonus: number
  priority: number
  per_user_limit: number
  total_limit: number
  used_count: number
  start_at: string | null
  end_at: string | null
  status: number
  created_at: string
  updated_at: string
}

/**
 * 优惠使用记录接口
 */
interface PromoUsage {
  id: number
  promo_id: number
  user_id: number
  recharge_no: string
  amount: number
  bonus_amount: number
  discount_amount: number
  created_at: string
}

/**
 * 余额管理页面
 */
export function BalancePage() {
  const [activeTab, setActiveTab] = useState<'balances' | 'logs' | 'promos' | 'config'>('balances')
  const [balances, setBalances] = useState<UserBalance[]>([])
  const [logs, setLogs] = useState<BalanceLog[]>([])
  const [promos, setPromos] = useState<RechargePromo[]>([])
  const [loading, setLoading] = useState(true)
  const [showAdjustModal, setShowAdjustModal] = useState(false)
  const [showPromoModal, setShowPromoModal] = useState(false)
  const [selectedUser, setSelectedUser] = useState<UserBalance | null>(null)
  const [selectedPromo, setSelectedPromo] = useState<RechargePromo | null>(null)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const pageSize = 20

  // 调整表单
  const [adjustForm, setAdjustForm] = useState({
    amount: '',
    type: 'add',
    remark: '',
  })

  // 搜索
  const [searchKeyword, setSearchKeyword] = useState('')

  // 配置表单
  const [config, setConfig] = useState<BalanceConfig>({
    min_recharge_amount: 1,
    max_recharge_amount: 50000,
    max_daily_recharge: 100000,
    max_balance_limit: 100000,
    large_recharge_threshold: 1000,
    large_consume_threshold: 500,
    frequent_recharge_count: 5,
    frequent_consume_count: 10,
    large_admin_adjust_threshold: 1000,
  })
  const [configLoading, setConfigLoading] = useState(false)
  const [configSaving, setConfigSaving] = useState(false)

  // 优惠活动表单
  const [promoForm, setPromoForm] = useState({
    name: '',
    description: '',
    promo_type: 'bonus' as 'discount' | 'bonus' | 'percent',
    min_amount: 0,
    max_amount: 0,
    value: 0,
    max_bonus: 0,
    priority: 0,
    per_user_limit: 0,
    total_limit: 0,
    start_at: '',
    end_at: '',
    status: 1,
  })

  useEffect(() => {
    loadData()
  }, [activeTab, page])

  const loadData = async () => {
    setLoading(true)
    if (activeTab === 'balances') {
      await loadBalances()
    } else if (activeTab === 'logs') {
      await loadLogs()
    } else if (activeTab === 'promos') {
      await loadPromos()
    } else if (activeTab === 'config') {
      await loadConfig()
    }
    setLoading(false)
  }

  const loadBalances = async () => {
    const res = await apiGet<{ data: UserBalance[]; total: number }>(
      `/api/admin/balances?page=${page}&page_size=${pageSize}&keyword=${searchKeyword}`
    )
    if (res.success) {
      setBalances(res.data || [])
      setTotal(res.total || 0)
    }
  }

  const loadLogs = async () => {
    const res = await apiGet<{ data: BalanceLog[]; total: number }>(
      `/api/admin/balance/logs?page=${page}&page_size=${pageSize}`
    )
    if (res.success) {
      setLogs(res.data || [])
      setTotal(res.total || 0)
    }
  }

  const loadPromos = async () => {
    const res = await apiGet<{ data: RechargePromo[]; total: number }>(
      `/api/admin/recharge-promos?page=${page}&page_size=${pageSize}`
    )
    if (res.success) {
      setPromos(res.data || [])
      setTotal(res.total || 0)
    }
  }

  const loadConfig = async () => {
    setConfigLoading(true)
    const res = await apiGet<{ data: BalanceConfig }>('/api/admin/balance/config')
    if (res.success && res.data) {
      setConfig(res.data)
    }
    setConfigLoading(false)
  }

  const saveConfig = async () => {
    setConfigSaving(true)
    const res = await apiPost('/api/admin/balance/config', config as unknown as Record<string, unknown>)
    if (res.success) {
      toast.success('配置保存成功')
    } else {
      toast.error(res.error || '保存失败')
    }
    setConfigSaving(false)
  }

  // 打开调整弹窗
  const openAdjustModal = (user: UserBalance) => {
    setSelectedUser(user)
    setAdjustForm({ amount: '', type: 'add', remark: '' })
    setShowAdjustModal(true)
  }

  // 打开优惠活动弹窗（新建）
  const openCreatePromoModal = () => {
    setSelectedPromo(null)
    setPromoForm({
      name: '',
      description: '',
      promo_type: 'bonus',
      min_amount: 0,
      max_amount: 0,
      value: 0,
      max_bonus: 0,
      priority: 0,
      per_user_limit: 0,
      total_limit: 0,
      start_at: '',
      end_at: '',
      status: 1,
    })
    setShowPromoModal(true)
  }

  // 打开优惠活动弹窗（编辑）
  const openEditPromoModal = (promo: RechargePromo) => {
    setSelectedPromo(promo)
    setPromoForm({
      name: promo.name,
      description: promo.description,
      promo_type: promo.promo_type,
      min_amount: promo.min_amount,
      max_amount: promo.max_amount,
      value: promo.value,
      max_bonus: promo.max_bonus,
      priority: promo.priority,
      per_user_limit: promo.per_user_limit,
      total_limit: promo.total_limit,
      start_at: promo.start_at ? promo.start_at.slice(0, 16).replace('T', ' ') : '',
      end_at: promo.end_at ? promo.end_at.slice(0, 16).replace('T', ' ') : '',
      status: promo.status,
    })
    setShowPromoModal(true)
  }

  // 调整余额
  const handleAdjust = async () => {
    if (!selectedUser) return
    const amount = parseFloat(adjustForm.amount)
    if (isNaN(amount) || amount <= 0) {
      toast.error('请输入有效金额')
      return
    }
    const res = await apiPost('/api/admin/balance/adjust', {
      user_id: selectedUser.user_id,
      amount: adjustForm.type === 'add' ? amount : -amount,
      remark: adjustForm.remark || (adjustForm.type === 'add' ? '管理员充值' : '管理员扣款'),
    })
    if (res.success) {
      toast.success('余额调整成功')
      setShowAdjustModal(false)
      loadBalances()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 保存优惠活动
  const handleSavePromo = async () => {
    if (!promoForm.name) {
      toast.error('请输入活动名称')
      return
    }
    if (promoForm.value <= 0) {
      toast.error('请输入有效的优惠值')
      return
    }
    if (promoForm.promo_type === 'discount' && (promoForm.value <= 0 || promoForm.value >= 1)) {
      toast.error('折扣率必须在0到1之间（如0.9表示9折）')
      return
    }

    const payload = {
      ...promoForm,
      start_at: promoForm.start_at ? promoForm.start_at + ':00' : null,
      end_at: promoForm.end_at ? promoForm.end_at + ':00' : null,
    }

    let res
    if (selectedPromo) {
      res = await apiPut(`/api/admin/recharge-promo/${selectedPromo.id}`, payload as unknown as Record<string, unknown>)
    } else {
      res = await apiPost('/api/admin/recharge-promo', payload as unknown as Record<string, unknown>)
    }

    if (res.success) {
      toast.success(selectedPromo ? '更新成功' : '创建成功')
      setShowPromoModal(false)
      loadPromos()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 删除优惠活动
  const handleDeletePromo = async (id: number) => {
    if (!confirm('确定要删除此优惠活动吗？')) return
    const res = await apiDelete(`/api/admin/recharge-promo/${id}`)
    if (res.success) {
      toast.success('删除成功')
      loadPromos()
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  // 切换优惠活动状态
  const handleTogglePromoStatus = async (id: number) => {
    const res = await apiPost(`/api/admin/recharge-promo/${id}/toggle`, {})
    if (res.success) {
      toast.success('状态已切换')
      loadPromos()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 搜索
  const handleSearch = () => {
    setPage(1)
    loadBalances()
  }

  // 获取类型标签
  const getTypeLabel = (type: string) => {
    const types: Record<string, { label: string; variant: 'success' | 'danger' | 'warning' | 'info' }> = {
      recharge: { label: '充值', variant: 'success' },
      consume: { label: '消费', variant: 'danger' },
      refund: { label: '退款', variant: 'warning' },
      adjust: { label: '调整', variant: 'info' },
      reward: { label: '奖励', variant: 'success' },
    }
    return types[type] || { label: type, variant: 'info' as const }
  }

  // 获取优惠类型标签
  const getPromoTypeLabel = (type: string) => {
    const types: Record<string, { label: string; variant: 'success' | 'warning' | 'info' }> = {
      discount: { label: '充值折扣', variant: 'warning' },
      bonus: { label: '固定赠金', variant: 'success' },
      percent: { label: '百分比赠送', variant: 'info' },
    }
    return types[type] || { label: type, variant: 'info' as const }
  }

  // 格式化优惠值显示
  const formatPromoValue = (promo: RechargePromo) => {
    switch (promo.promo_type) {
      case 'discount':
        return `${(promo.value * 10).toFixed(1)}折`
      case 'bonus':
        return `赠${formatMoney(promo.value)}`
      case 'percent':
        return `赠${promo.value}%${promo.max_bonus > 0 ? `(最高${formatMoney(promo.max_bonus)})` : ''}`
      default:
        return String(promo.value)
    }
  }

  const totalPages = Math.ceil(total / pageSize)

  if (loading && balances.length === 0 && logs.length === 0 && activeTab !== 'config') {
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
        <button
          onClick={() => { setActiveTab('balances'); setPage(1) }}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'balances'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-wallet mr-2" />
          用户余额
        </button>
        <button
          onClick={() => { setActiveTab('logs'); setPage(1) }}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'logs'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-history mr-2" />
          余额记录
        </button>
        <button
          onClick={() => { setActiveTab('promos'); setPage(1) }}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'promos'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-gift mr-2" />
          充值优惠
        </button>
        <button
          onClick={() => { setActiveTab('config'); setPage(1) }}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'config'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-cog mr-2" />
          余额配置
        </button>
      </div>

      {/* 用户余额列表 */}
      {activeTab === 'balances' && (
        <Card title="用户余额" icon={<i className="fas fa-wallet" />}>
          {/* 搜索栏 */}
          <div className="flex gap-2 mb-4">
            <Input
              placeholder="搜索用户名"
              value={searchKeyword}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setSearchKeyword(e.target.value)}
              className="flex-1"
            />
            <Button onClick={handleSearch}>
              <i className="fas fa-search mr-1" />
              搜索
            </Button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">用户</th>
                  <th className="pb-3 font-medium">可用余额</th>
                  <th className="pb-3 font-medium">冻结余额</th>
                  <th className="pb-3 font-medium">累计充值</th>
                  <th className="pb-3 font-medium">累计消费</th>
                  <th className="pb-3 font-medium">状态</th>
                  <th className="pb-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {balances.map((item) => (
                  <tr key={item.id} className="border-b border-dark-700/50">
                    <td className="py-3">{item.username}</td>
                    <td className="py-3 text-green-400">{formatMoney(item.balance)}</td>
                    <td className="py-3 text-yellow-400">{formatMoney(item.frozen_balance)}</td>
                    <td className="py-3">{formatMoney(item.total_recharge)}</td>
                    <td className="py-3">{formatMoney(item.total_consume)}</td>
                    <td className="py-3">
                      <Badge variant={item.status === 1 ? 'success' : 'danger'}>
                        {item.status === 1 ? '正常' : '冻结'}
                      </Badge>
                    </td>
                    <td className="py-3">
                      <Button size="sm" variant="ghost" onClick={() => openAdjustModal(item)}>
                        <i className="fas fa-edit mr-1" />
                        调整
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* 分页 */}
          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-4">
              <Button size="sm" variant="ghost" disabled={page === 1} onClick={() => setPage(p => p - 1)}>
                上一页
              </Button>
              <span className="px-4 py-2 text-dark-400">{page} / {totalPages}</span>
              <Button size="sm" variant="ghost" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
                下一页
              </Button>
            </div>
          )}
        </Card>
      )}

      {/* 余额记录 */}
      {activeTab === 'logs' && (
        <Card title="余额记录" icon={<i className="fas fa-history" />}>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">用户</th>
                  <th className="pb-3 font-medium">类型</th>
                  <th className="pb-3 font-medium">金额</th>
                  <th className="pb-3 font-medium">变动前</th>
                  <th className="pb-3 font-medium">变动后</th>
                  <th className="pb-3 font-medium">备注</th>
                  <th className="pb-3 font-medium">时间</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {logs.map((log) => {
                  const typeInfo = getTypeLabel(log.type)
                  return (
                    <tr key={log.id} className="border-b border-dark-700/50">
                      <td className="py-3">{log.username}</td>
                      <td className="py-3">
                        <Badge variant={typeInfo.variant}>{typeInfo.label}</Badge>
                      </td>
                      <td className={`py-3 ${log.amount >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                        {log.amount >= 0 ? '+' : ''}{formatMoney(log.amount)}
                      </td>
                      <td className="py-3">{formatMoney(log.balance_before)}</td>
                      <td className="py-3">{formatMoney(log.balance_after)}</td>
                      <td className="py-3 text-dark-400 max-w-xs truncate">{log.remark || '-'}</td>
                      <td className="py-3 text-sm text-dark-400">{formatDateTime(log.created_at)}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>

          {/* 分页 */}
          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-4">
              <Button size="sm" variant="ghost" disabled={page === 1} onClick={() => setPage(p => p - 1)}>
                上一页
              </Button>
              <span className="px-4 py-2 text-dark-400">{page} / {totalPages}</span>
              <Button size="sm" variant="ghost" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
                下一页
              </Button>
            </div>
          )}
        </Card>
      )}

      {/* 充值优惠活动 */}
      {activeTab === 'promos' && (
        <Card 
          title="充值优惠活动" 
          icon={<i className="fas fa-gift" />}
          action={
            <Button size="sm" onClick={openCreatePromoModal}>
              <i className="fas fa-plus mr-1" />
              新建活动
            </Button>
          }
        >
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">活动名称</th>
                  <th className="pb-3 font-medium">类型</th>
                  <th className="pb-3 font-medium">优惠内容</th>
                  <th className="pb-3 font-medium">充值门槛</th>
                  <th className="pb-3 font-medium">使用次数</th>
                  <th className="pb-3 font-medium">有效期</th>
                  <th className="pb-3 font-medium">状态</th>
                  <th className="pb-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {promos.map((promo) => {
                  const typeInfo = getPromoTypeLabel(promo.promo_type)
                  const isExpired = promo.end_at && new Date(promo.end_at) < new Date()
                  return (
                    <tr key={promo.id} className="border-b border-dark-700/50">
                      <td className="py-3">
                        <div className="font-medium">{promo.name}</div>
                        {promo.description && (
                          <div className="text-xs text-dark-400 mt-1">{promo.description}</div>
                        )}
                      </td>
                      <td className="py-3">
                        <Badge variant={typeInfo.variant}>{typeInfo.label}</Badge>
                      </td>
                      <td className="py-3 text-green-400 font-medium">
                        {formatPromoValue(promo)}
                      </td>
                      <td className="py-3">
                        {promo.min_amount > 0 ? `≥${formatMoney(promo.min_amount)}` : '无门槛'}
                        {promo.max_amount > 0 && ` / ≤${formatMoney(promo.max_amount)}`}
                      </td>
                      <td className="py-3">
                        {promo.used_count}
                        {promo.total_limit > 0 && ` / ${promo.total_limit}`}
                      </td>
                      <td className="py-3 text-sm">
                        {promo.start_at || promo.end_at ? (
                          <div>
                            {promo.start_at && <div>{formatDateTime(promo.start_at).slice(0, 10)}</div>}
                            {promo.end_at && <div className={isExpired ? 'text-red-400' : ''}>至 {formatDateTime(promo.end_at).slice(0, 10)}</div>}
                          </div>
                        ) : (
                          <span className="text-dark-400">长期有效</span>
                        )}
                      </td>
                      <td className="py-3">
                        <Badge variant={promo.status === 1 ? 'success' : 'danger'}>
                          {promo.status === 1 ? '启用' : '禁用'}
                        </Badge>
                      </td>
                      <td className="py-3">
                        <div className="flex gap-1">
                          <Button size="sm" variant="ghost" onClick={() => openEditPromoModal(promo)}>
                            <i className="fas fa-edit" />
                          </Button>
                          <Button size="sm" variant="ghost" onClick={() => handleTogglePromoStatus(promo.id)}>
                            <i className={`fas fa-${promo.status === 1 ? 'pause' : 'play'}`} />
                          </Button>
                          <Button size="sm" variant="ghost" className="text-red-400" onClick={() => handleDeletePromo(promo.id)}>
                            <i className="fas fa-trash" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
                {promos.length === 0 && (
                  <tr>
                    <td colSpan={8} className="py-8 text-center text-dark-400">
                      暂无优惠活动
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {/* 分页 */}
          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-4">
              <Button size="sm" variant="ghost" disabled={page === 1} onClick={() => setPage(p => p - 1)}>
                上一页
              </Button>
              <span className="px-4 py-2 text-dark-400">{page} / {totalPages}</span>
              <Button size="sm" variant="ghost" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
                下一页
              </Button>
            </div>
          )}
        </Card>
      )}

      {/* 余额配置 */}
      {activeTab === 'config' && (
        <div className="space-y-6">
          {/* 充值限制配置 */}
          <Card title="充值限制配置" icon={<i className="fas fa-sliders-h" />}>
            {configLoading ? (
              <div className="flex items-center justify-center py-8">
                <i className="fas fa-spinner fa-spin text-xl text-primary-400" />
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="单笔充值最小金额（元）"
                  type="number"
                  value={config.min_recharge_amount}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfig({ ...config, min_recharge_amount: parseFloat(e.target.value) || 0 })}
                />
                <Input
                  label="单笔充值最大金额（元）"
                  type="number"
                  value={config.max_recharge_amount}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfig({ ...config, max_recharge_amount: parseFloat(e.target.value) || 0 })}
                />
                <Input
                  label="每日充值上限（元）"
                  type="number"
                  value={config.max_daily_recharge}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfig({ ...config, max_daily_recharge: parseFloat(e.target.value) || 0 })}
                />
                <Input
                  label="用户余额上限（元）"
                  type="number"
                  value={config.max_balance_limit}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfig({ ...config, max_balance_limit: parseFloat(e.target.value) || 0 })}
                />
              </div>
            )}
          </Card>

          {/* 告警阈值配置 */}
          <Card title="告警阈值配置" icon={<i className="fas fa-bell" />}>
            {configLoading ? (
              <div className="flex items-center justify-center py-8">
                <i className="fas fa-spinner fa-spin text-xl text-primary-400" />
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="大额充值告警阈值（元）"
                  type="number"
                  value={config.large_recharge_threshold}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfig({ ...config, large_recharge_threshold: parseFloat(e.target.value) || 0 })}
                />
                <Input
                  label="大额消费告警阈值（元）"
                  type="number"
                  value={config.large_consume_threshold}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfig({ ...config, large_consume_threshold: parseFloat(e.target.value) || 0 })}
                />
                <Input
                  label="频繁充值告警次数（每小时）"
                  type="number"
                  value={config.frequent_recharge_count}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfig({ ...config, frequent_recharge_count: parseInt(e.target.value) || 0 })}
                />
                <Input
                  label="频繁消费告警次数（每小时）"
                  type="number"
                  value={config.frequent_consume_count}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfig({ ...config, frequent_consume_count: parseInt(e.target.value) || 0 })}
                />
                <Input
                  label="管理员大额调整告警阈值（元）"
                  type="number"
                  value={config.large_admin_adjust_threshold}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfig({ ...config, large_admin_adjust_threshold: parseFloat(e.target.value) || 0 })}
                />
              </div>
            )}
          </Card>

          {/* 保存按钮 */}
          <div className="flex justify-end">
            <Button onClick={saveConfig} disabled={configSaving}>
              {configSaving ? (
                <>
                  <i className="fas fa-spinner fa-spin mr-2" />
                  保存中...
                </>
              ) : (
                <>
                  <i className="fas fa-save mr-2" />
                  保存配置
                </>
              )}
            </Button>
          </div>
        </div>
      )}

      {/* 调整余额弹窗 */}
      <Modal
        isOpen={showAdjustModal}
        onClose={() => setShowAdjustModal(false)}
        title={`调整余额 - ${selectedUser?.username}`}
      >
        <div className="space-y-4">
          <div className="p-3 bg-dark-700/30 rounded-lg">
            <div className="text-sm text-dark-400">当前余额</div>
            <div className="text-xl font-bold text-green-400">{formatMoney(selectedUser?.balance || 0)}</div>
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">操作类型</label>
            <div className="flex gap-2">
              <button
                onClick={() => setAdjustForm({ ...adjustForm, type: 'add' })}
                className={`flex-1 py-2 rounded-lg text-sm font-medium transition-colors ${
                  adjustForm.type === 'add'
                    ? 'bg-green-500/20 text-green-400 border border-green-500/50'
                    : 'bg-dark-700/50 text-dark-400 border border-dark-600'
                }`}
              >
                <i className="fas fa-plus mr-1" />
                充值
              </button>
              <button
                onClick={() => setAdjustForm({ ...adjustForm, type: 'subtract' })}
                className={`flex-1 py-2 rounded-lg text-sm font-medium transition-colors ${
                  adjustForm.type === 'subtract'
                    ? 'bg-red-500/20 text-red-400 border border-red-500/50'
                    : 'bg-dark-700/50 text-dark-400 border border-dark-600'
                }`}
              >
                <i className="fas fa-minus mr-1" />
                扣款
              </button>
            </div>
          </div>
          <Input
            label="金额"
            type="number"
            placeholder="请输入金额"
            value={adjustForm.amount}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setAdjustForm({ ...adjustForm, amount: e.target.value })}
          />
          <Input
            label="备注"
            placeholder="请输入备注（可选）"
            value={adjustForm.remark}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setAdjustForm({ ...adjustForm, remark: e.target.value })}
          />
          <Button className="w-full" onClick={handleAdjust}>
            确认调整
          </Button>
        </div>
      </Modal>

      {/* 优惠活动编辑弹窗 */}
      <Modal
        isOpen={showPromoModal}
        onClose={() => setShowPromoModal(false)}
        title={selectedPromo ? '编辑优惠活动' : '新建优惠活动'}
      >
        <div className="space-y-4 max-h-[70vh] overflow-y-auto">
          <Input
            label="活动名称"
            placeholder="请输入活动名称"
            value={promoForm.name}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, name: e.target.value })}
          />
          <Input
            label="活动描述"
            placeholder="请输入活动描述（可选）"
            value={promoForm.description}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, description: e.target.value })}
          />
          
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">优惠类型</label>
            <div className="flex gap-2">
              <button
                onClick={() => setPromoForm({ ...promoForm, promo_type: 'bonus' })}
                className={`flex-1 py-2 rounded-lg text-sm font-medium transition-colors ${
                  promoForm.promo_type === 'bonus'
                    ? 'bg-green-500/20 text-green-400 border border-green-500/50'
                    : 'bg-dark-700/50 text-dark-400 border border-dark-600'
                }`}
              >
                固定赠金
              </button>
              <button
                onClick={() => setPromoForm({ ...promoForm, promo_type: 'percent' })}
                className={`flex-1 py-2 rounded-lg text-sm font-medium transition-colors ${
                  promoForm.promo_type === 'percent'
                    ? 'bg-blue-500/20 text-blue-400 border border-blue-500/50'
                    : 'bg-dark-700/50 text-dark-400 border border-dark-600'
                }`}
              >
                百分比赠送
              </button>
              <button
                onClick={() => setPromoForm({ ...promoForm, promo_type: 'discount' })}
                className={`flex-1 py-2 rounded-lg text-sm font-medium transition-colors ${
                  promoForm.promo_type === 'discount'
                    ? 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/50'
                    : 'bg-dark-700/50 text-dark-400 border border-dark-600'
                }`}
              >
                充值折扣
              </button>
            </div>
            <div className="mt-2 text-xs text-dark-400">
              {promoForm.promo_type === 'bonus' && '固定赠金：充值达到门槛后赠送固定金额'}
              {promoForm.promo_type === 'percent' && '百分比赠送：按充值金额的百分比赠送'}
              {promoForm.promo_type === 'discount' && '充值折扣：充值时享受折扣（如0.9表示9折）'}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <Input
              label="最低充值金额"
              type="number"
              placeholder="0表示无门槛"
              value={promoForm.min_amount}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, min_amount: parseFloat(e.target.value) || 0 })}
            />
            <Input
              label="最高充值金额"
              type="number"
              placeholder="0表示不限"
              value={promoForm.max_amount}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, max_amount: parseFloat(e.target.value) || 0 })}
            />
          </div>

          <Input
            label={promoForm.promo_type === 'discount' ? '折扣率（如0.9表示9折）' : promoForm.promo_type === 'percent' ? '赠送百分比' : '赠送金额'}
            type="number"
            step={promoForm.promo_type === 'discount' ? '0.01' : '1'}
            placeholder={promoForm.promo_type === 'discount' ? '0.9' : promoForm.promo_type === 'percent' ? '10' : '10'}
            value={promoForm.value}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, value: parseFloat(e.target.value) || 0 })}
          />

          {promoForm.promo_type === 'percent' && (
            <Input
              label="最大赠送金额"
              type="number"
              placeholder="0表示不限"
              value={promoForm.max_bonus}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, max_bonus: parseFloat(e.target.value) || 0 })}
            />
          )}

          <div className="grid grid-cols-2 gap-4">
            <Input
              label="每用户限用次数"
              type="number"
              placeholder="0表示不限"
              value={promoForm.per_user_limit}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, per_user_limit: parseInt(e.target.value) || 0 })}
            />
            <Input
              label="总使用次数限制"
              type="number"
              placeholder="0表示不限"
              value={promoForm.total_limit}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, total_limit: parseInt(e.target.value) || 0 })}
            />
          </div>

          <Input
            label="优先级"
            type="number"
            placeholder="数字越大优先级越高"
            value={promoForm.priority}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, priority: parseInt(e.target.value) || 0 })}
          />

          <div className="grid grid-cols-2 gap-4">
            <Input
              label="开始时间"
              type="datetime-local"
              value={promoForm.start_at}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, start_at: e.target.value })}
            />
            <Input
              label="结束时间"
              type="datetime-local"
              value={promoForm.end_at}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setPromoForm({ ...promoForm, end_at: e.target.value })}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">状态</label>
            <div className="flex gap-2">
              <button
                onClick={() => setPromoForm({ ...promoForm, status: 1 })}
                className={`flex-1 py-2 rounded-lg text-sm font-medium transition-colors ${
                  promoForm.status === 1
                    ? 'bg-green-500/20 text-green-400 border border-green-500/50'
                    : 'bg-dark-700/50 text-dark-400 border border-dark-600'
                }`}
              >
                启用
              </button>
              <button
                onClick={() => setPromoForm({ ...promoForm, status: 0 })}
                className={`flex-1 py-2 rounded-lg text-sm font-medium transition-colors ${
                  promoForm.status === 0
                    ? 'bg-red-500/20 text-red-400 border border-red-500/50'
                    : 'bg-dark-700/50 text-dark-400 border border-dark-600'
                }`}
              >
                禁用
              </button>
            </div>
          </div>

          <Button className="w-full" onClick={handleSavePromo}>
            {selectedPromo ? '保存修改' : '创建活动'}
          </Button>
        </div>
      </Modal>
    </div>
  )
}
