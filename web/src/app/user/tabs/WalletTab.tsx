'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Card, Badge, Modal, Input } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * 余额信息接口
 */
interface BalanceInfo {
  balance: number
  frozen: number
  total_in: number
  total_out: number
}

/**
 * 余额记录接口
 */
interface BalanceLog {
  id: number
  type: string
  amount: number
  before_balance: number
  after_balance: number
  order_no: string
  remark: string
  created_at: string
}

/**
 * 积分信息接口
 */
interface PointsInfo {
  points: number
  total_earn: number
  total_used: number
}

/**
 * 积分记录接口
 */
interface PointsLog {
  id: number
  type: string
  points: number
  balance: number
  order_no: string
  remark: string
  created_at: string
}

/**
 * 可兑换项目接口
 */
interface ExchangeItem {
  id: number
  name: string
  type: string
  points_cost: number
  value: number
  stock: number
}

/**
 * 支付密码状态接口
 */
interface PayPasswordStatus {
  is_set: boolean
  is_locked: boolean
  lock_remaining_seconds: number
}

/**
 * 充值优惠计算结果接口
 */
interface PromoResult {
  promo_id: number
  promo_name: string
  promo_type: string
  original_amount: number
  pay_amount: number
  bonus_amount: number
  discount_amount: number
  total_credit: number
}

/**
 * 充值优惠活动接口
 */
interface RechargePromo {
  id: number
  name: string
  description: string
  promo_type: string
  min_amount: number
  max_amount: number
  value: number
  max_bonus: number
}

/**
 * 我的钱包标签页（余额+积分）
 */
export function WalletTab() {
  const [activeSection, setActiveSection] = useState<'balance' | 'points'>('balance')
  const [balanceInfo, setBalanceInfo] = useState<BalanceInfo | null>(null)
  const [balanceLogs, setBalanceLogs] = useState<BalanceLog[]>([])
  const [pointsInfo, setPointsInfo] = useState<PointsInfo | null>(null)
  const [pointsLogs, setPointsLogs] = useState<PointsLog[]>([])
  const [exchangeList, setExchangeList] = useState<ExchangeItem[]>([])
  const [loading, setLoading] = useState(true)
  const [showRechargeModal, setShowRechargeModal] = useState(false)
  const [showExchangeModal, setShowExchangeModal] = useState(false)
  const [selectedExchange, setSelectedExchange] = useState<ExchangeItem | null>(null)
  const [rechargeAmount, setRechargeAmount] = useState('')
  
  // 支付密码相关状态
  const [payPasswordStatus, setPayPasswordStatus] = useState<PayPasswordStatus | null>(null)
  const [showSetPayPasswordModal, setShowSetPayPasswordModal] = useState(false)
  const [showUpdatePayPasswordModal, setShowUpdatePayPasswordModal] = useState(false)
  const [showResetPayPasswordModal, setShowResetPayPasswordModal] = useState(false)
  const [payPassword, setPayPassword] = useState('')
  const [confirmPayPassword, setConfirmPayPassword] = useState('')
  const [loginPassword, setLoginPassword] = useState('')
  const [oldPayPassword, setOldPayPassword] = useState('')
  const [emailCode, setEmailCode] = useState('')
  const [rechargePayPassword, setRechargePayPassword] = useState('')
  const [sendingCode, setSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)

  // 充值优惠相关状态
  const [promoResult, setPromoResult] = useState<PromoResult | null>(null)
  const [activePromos, setActivePromos] = useState<RechargePromo[]>([])
  const [calculatingPromo, setCalculatingPromo] = useState(false)

  // 加载余额信息
  const loadBalance = async () => {
    const res = await apiGet<{ data: BalanceInfo }>('/api/user/balance')
    if (res.success && res.data) {
      setBalanceInfo(res.data)
    }
  }

  // 加载余额记录
  const loadBalanceLogs = async () => {
    const res = await apiGet<{ data: BalanceLog[] }>('/api/user/balance/logs')
    if (res.success && res.data) {
      setBalanceLogs(res.data)
    }
  }

  // 加载支付密码状态
  const loadPayPasswordStatus = async () => {
    const res = await apiGet<{ data: PayPasswordStatus }>('/api/user/pay-password/status')
    if (res.success && res.data) {
      setPayPasswordStatus(res.data)
    }
  }

  // 加载有效的充值优惠活动
  const loadActivePromos = async () => {
    const res = await apiGet<{ data: RechargePromo[] }>('/api/user/balance/promos')
    if (res.success && res.data) {
      setActivePromos(res.data)
    }
  }

  // 计算充值优惠
  const calculatePromo = async (amount: number) => {
    if (amount <= 0) {
      setPromoResult(null)
      return
    }
    setCalculatingPromo(true)
    const res = await apiPost<{ data: PromoResult }>('/api/user/balance/promo/calculate', { amount })
    if (res.success && res.data) {
      setPromoResult(res.data)
    } else {
      setPromoResult(null)
    }
    setCalculatingPromo(false)
  }

  // 加载积分信息
  const loadPoints = async () => {
    const res = await apiGet<{ points: number; total_earn: number; total_used: number }>('/api/user/points')
    if (res.success) {
      setPointsInfo({
        points: res.points || 0,
        total_earn: res.total_earn || 0,
        total_used: res.total_used || 0,
      })
    }
  }

  // 加载积分记录
  const loadPointsLogs = async () => {
    const res = await apiGet<{ logs: PointsLog[] }>('/api/user/points/logs')
    if (res.success && res.logs) {
      setPointsLogs(res.logs)
    }
  }

  // 加载可兑换列表
  const loadExchangeList = async () => {
    const res = await apiGet<{ list: ExchangeItem[] }>('/api/user/points/exchange/list')
    if (res.success && res.list) {
      setExchangeList(res.list)
    }
  }

  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      await Promise.all([loadBalance(), loadBalanceLogs(), loadPayPasswordStatus(), loadActivePromos()])
      setLoading(false)
    }
    loadData()
  }, [])

  // 切换到积分时加载数据
  useEffect(() => {
    if (activeSection === 'points' && !pointsInfo) {
      loadPoints()
      loadPointsLogs()
      loadExchangeList()
    }
  }, [activeSection])

  // 创建充值订单
  const handleRecharge = async () => {
    const amount = parseFloat(rechargeAmount)
    if (isNaN(amount) || amount <= 0) {
      toast.error('请输入有效金额')
      return
    }
    if (!rechargePayPassword || rechargePayPassword.length !== 6) {
      toast.error('请输入6位支付密码')
      return
    }
    const res = await apiPost<{ data: { recharge_no: string; pay_amount: number; bonus_amount: number; total_credit: number } }>('/api/user/balance/recharge', { 
      amount,
      payment_method: 'yipay',
      pay_password: rechargePayPassword
    })
    if (res.success && res.data?.recharge_no) {
      toast.success('充值订单已创建，正在跳转支付页面...')
      setShowRechargeModal(false)
      setRechargeAmount('')
      setRechargePayPassword('')
      setPromoResult(null)
      // 跳转到支付页面
      window.location.href = `/payment?type=recharge&recharge_no=${res.data.recharge_no}`
    } else {
      toast.error(res.error || '创建失败')
    }
  }

  // 处理充值金额变化
  const handleRechargeAmountChange = (value: string) => {
    setRechargeAmount(value)
    const amount = parseFloat(value)
    if (!isNaN(amount) && amount > 0) {
      calculatePromo(amount)
    } else {
      setPromoResult(null)
    }
  }

  // 设置支付密码
  const handleSetPayPassword = async () => {
    if (payPassword.length !== 6 || !/^\d{6}$/.test(payPassword)) {
      toast.error('支付密码必须为6位纯数字')
      return
    }
    if (payPassword !== confirmPayPassword) {
      toast.error('两次输入的密码不一致')
      return
    }
    if (!loginPassword) {
      toast.error('请输入登录密码')
      return
    }
    const res = await apiPost('/api/user/pay-password/set', {
      password: payPassword,
      login_password: loginPassword
    })
    if (res.success) {
      toast.success('支付密码设置成功')
      setShowSetPayPasswordModal(false)
      setPayPassword('')
      setConfirmPayPassword('')
      setLoginPassword('')
      loadPayPasswordStatus()
    } else {
      toast.error(res.error || '设置失败')
    }
  }

  // 修改支付密码
  const handleUpdatePayPassword = async () => {
    if (payPassword.length !== 6 || !/^\d{6}$/.test(payPassword)) {
      toast.error('新支付密码必须为6位纯数字')
      return
    }
    if (payPassword !== confirmPayPassword) {
      toast.error('两次输入的密码不一致')
      return
    }
    if (!oldPayPassword) {
      toast.error('请输入原支付密码')
      return
    }
    const res = await apiPost('/api/user/pay-password/update', {
      old_password: oldPayPassword,
      new_password: payPassword
    })
    if (res.success) {
      toast.success('支付密码修改成功')
      setShowUpdatePayPasswordModal(false)
      setPayPassword('')
      setConfirmPayPassword('')
      setOldPayPassword('')
      loadPayPasswordStatus()
    } else {
      toast.error(res.error || '修改失败')
    }
  }

  // 发送重置验证码
  const handleSendResetCode = async () => {
    setSendingCode(true)
    const res = await apiPost('/api/user/pay-password/send-reset-code', {})
    setSendingCode(false)
    if (res.success) {
      toast.success('验证码已发送到您的邮箱')
      setCountdown(60)
      const timer = setInterval(() => {
        setCountdown(prev => {
          if (prev <= 1) {
            clearInterval(timer)
            return 0
          }
          return prev - 1
        })
      }, 1000)
    } else {
      toast.error(res.error || '发送失败')
    }
  }

  // 重置支付密码
  const handleResetPayPassword = async () => {
    if (payPassword.length !== 6 || !/^\d{6}$/.test(payPassword)) {
      toast.error('新支付密码必须为6位纯数字')
      return
    }
    if (payPassword !== confirmPayPassword) {
      toast.error('两次输入的密码不一致')
      return
    }
    if (!emailCode) {
      toast.error('请输入邮箱验证码')
      return
    }
    const res = await apiPost('/api/user/pay-password/reset', {
      new_password: payPassword,
      email_code: emailCode
    })
    if (res.success) {
      toast.success('支付密码重置成功')
      setShowResetPayPasswordModal(false)
      setPayPassword('')
      setConfirmPayPassword('')
      setEmailCode('')
      loadPayPasswordStatus()
    } else {
      toast.error(res.error || '重置失败')
    }
  }

  // 兑换优惠券
  const handleExchange = async () => {
    if (!selectedExchange) return
    const res = await apiPost('/api/user/points/exchange/coupon', { exchange_id: selectedExchange.id })
    if (res.success) {
      toast.success('兑换成功！优惠券已发放到您的账户')
      setShowExchangeModal(false)
      setSelectedExchange(null)
      loadPoints()
      loadPointsLogs()
    } else {
      toast.error(res.error || '兑换失败')
    }
  }

  // 获取余额类型文本
  const getBalanceTypeText = (type: string) => {
    const types: Record<string, string> = {
      recharge: '充值',
      consume: '消费',
      refund: '退款',
      gift: '赠送',
      adjust: '调整',
    }
    return types[type] || type
  }

  // 获取积分类型文本
  const getPointsTypeText = (type: string) => {
    const types: Record<string, string> = {
      earn: '获得',
      use: '使用',
      admin: '管理员调整',
      expire: '过期',
    }
    return types[type] || type
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
          onClick={() => setActiveSection('balance')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeSection === 'balance'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-wallet mr-2" />
          我的余额
        </button>
        <button
          onClick={() => setActiveSection('points')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeSection === 'points'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-coins mr-2" />
          我的积分
        </button>
      </div>

      {/* 余额部分 */}
      {activeSection === 'balance' && balanceInfo && (
        <>
          {/* 支付密码提示 */}
          {payPasswordStatus && !payPasswordStatus.is_set && (
            <div className="bg-amber-500/10 border border-amber-500/30 rounded-xl p-4 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <i className="fas fa-exclamation-triangle text-amber-400" />
                <span className="text-amber-200">请先设置支付密码才能使用余额功能</span>
              </div>
              <Button size="sm" onClick={() => setShowSetPayPasswordModal(true)}>
                立即设置
              </Button>
            </div>
          )}

          {/* 支付密码管理卡片 */}
          {payPasswordStatus && payPasswordStatus.is_set && (
            <Card className="bg-dark-800/50">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-green-500/20 flex items-center justify-center">
                    <i className="fas fa-shield-alt text-green-400" />
                  </div>
                  <div>
                    <div className="text-dark-100 font-medium">支付密码</div>
                    <div className="text-sm text-dark-400">
                      {payPasswordStatus.is_locked 
                        ? `已锁定，${Math.ceil(payPasswordStatus.lock_remaining_seconds / 60)}分钟后解锁`
                        : '已设置'}
                    </div>
                  </div>
                </div>
                <div className="flex gap-2">
                  <Button size="sm" variant="secondary" onClick={() => setShowUpdatePayPasswordModal(true)}>
                    修改
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => setShowResetPayPasswordModal(true)}>
                    忘记密码
                  </Button>
                </div>
              </div>
            </Card>
          )}

          {/* 余额卡片 */}
          <div className="bg-gradient-to-r from-primary-600 to-primary-500 rounded-2xl p-6 text-white">
            <div className="flex items-center justify-between mb-4">
              <span className="text-white/80">可用余额</span>
              <Button
                size="sm"
                variant="secondary"
                className="bg-white/20 hover:bg-white/30 text-white border-0"
                onClick={() => setShowRechargeModal(true)}
                disabled={!payPasswordStatus?.is_set}
                title={!payPasswordStatus?.is_set ? '请先设置支付密码' : ''}
              >
                <i className="fas fa-plus mr-1" />
                充值
              </Button>
            </div>
            <div className="text-4xl font-bold mb-4">¥{balanceInfo.balance.toFixed(2)}</div>
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div>
                <div className="text-white/60">冻结金额</div>
                <div className="font-medium">¥{balanceInfo.frozen.toFixed(2)}</div>
              </div>
              <div>
                <div className="text-white/60">累计充值</div>
                <div className="font-medium">¥{balanceInfo.total_in.toFixed(2)}</div>
              </div>
              <div>
                <div className="text-white/60">累计消费</div>
                <div className="font-medium">¥{balanceInfo.total_out.toFixed(2)}</div>
              </div>
            </div>
          </div>

          {/* 余额记录 */}
          <Card title="余额明细" icon={<i className="fas fa-list" />}>
            {balanceLogs.length === 0 ? (
              <div className="text-center py-8 text-dark-400">暂无记录</div>
            ) : (
              <div className="space-y-3">
                {balanceLogs.map((log) => (
                  <div key={log.id} className="flex items-center justify-between p-3 bg-dark-700/30 rounded-lg">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-dark-100">{getBalanceTypeText(log.type)}</span>
                        {log.remark && <span className="text-dark-500 text-sm">- {log.remark}</span>}
                      </div>
                      <div className="text-sm text-dark-500">{formatDateTime(log.created_at)}</div>
                    </div>
                    <div className={`font-medium ${log.amount >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                      {log.amount >= 0 ? '+' : ''}{log.amount.toFixed(2)}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </>
      )}

      {/* 积分部分 */}
      {activeSection === 'points' && pointsInfo && (
        <>
          {/* 积分卡片 */}
          <div className="bg-gradient-to-r from-amber-600 to-amber-500 rounded-2xl p-6 text-white">
            <div className="flex items-center justify-between mb-4">
              <span className="text-white/80">当前积分</span>
              <Button
                size="sm"
                variant="secondary"
                className="bg-white/20 hover:bg-white/30 text-white border-0"
                onClick={() => loadExchangeList()}
              >
                <i className="fas fa-gift mr-1" />
                兑换
              </Button>
            </div>
            <div className="text-4xl font-bold mb-4">{pointsInfo.points}</div>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-white/60">累计获得</div>
                <div className="font-medium">{pointsInfo.total_earn}</div>
              </div>
              <div>
                <div className="text-white/60">累计使用</div>
                <div className="font-medium">{pointsInfo.total_used}</div>
              </div>
            </div>
          </div>

          {/* 可兑换列表 */}
          {exchangeList.length > 0 && (
            <Card title="积分兑换" icon={<i className="fas fa-gift" />}>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {exchangeList.map((item) => (
                  <div key={item.id} className="p-4 bg-dark-700/30 rounded-xl border border-dark-600/50">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium text-dark-100">{item.name}</span>
                      <Badge variant="warning">{item.points_cost} 积分</Badge>
                    </div>
                    <div className="text-sm text-dark-400 mb-3">
                      {item.type === 'coupon' ? `优惠券 ¥${item.value}` : `商品价值 ¥${item.value}`}
                      {item.stock > 0 && <span className="ml-2">库存: {item.stock}</span>}
                    </div>
                    <Button
                      size="sm"
                      className="w-full"
                      disabled={pointsInfo.points < item.points_cost || item.stock === 0}
                      onClick={() => {
                        setSelectedExchange(item)
                        setShowExchangeModal(true)
                      }}
                    >
                      {item.stock === 0 ? '已兑完' : pointsInfo.points < item.points_cost ? '积分不足' : '立即兑换'}
                    </Button>
                  </div>
                ))}
              </div>
            </Card>
          )}

          {/* 积分记录 */}
          <Card title="积分明细" icon={<i className="fas fa-list" />}>
            {pointsLogs.length === 0 ? (
              <div className="text-center py-8 text-dark-400">暂无记录</div>
            ) : (
              <div className="space-y-3">
                {pointsLogs.map((log) => (
                  <div key={log.id} className="flex items-center justify-between p-3 bg-dark-700/30 rounded-lg">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-dark-100">{getPointsTypeText(log.type)}</span>
                        {log.remark && <span className="text-dark-500 text-sm">- {log.remark}</span>}
                      </div>
                      <div className="text-sm text-dark-500">{formatDateTime(log.created_at)}</div>
                    </div>
                    <div className={`font-medium ${log.points >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                      {log.points >= 0 ? '+' : ''}{log.points}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </>
      )}

      {/* 充值弹窗 */}
      <Modal isOpen={showRechargeModal} onClose={() => { setShowRechargeModal(false); setPromoResult(null); setRechargeAmount(''); }} title="余额充值" size="sm">
        <div className="space-y-4">
          {/* 当前有效的优惠活动提示 */}
          {activePromos.length > 0 && (
            <div className="p-3 bg-green-500/10 border border-green-500/30 rounded-lg">
              <div className="flex items-center gap-2 text-green-400 text-sm mb-2">
                <i className="fas fa-gift" />
                <span>当前充值优惠</span>
              </div>
              <div className="space-y-1 text-xs text-dark-300">
                {activePromos.slice(0, 3).map((promo) => (
                  <div key={promo.id}>
                    • {promo.name}
                    {promo.min_amount > 0 && `（满${promo.min_amount}元）`}
                  </div>
                ))}
              </div>
            </div>
          )}

          <Input
            label="充值金额"
            type="number"
            placeholder="请输入充值金额"
            value={rechargeAmount}
            onChange={(e) => handleRechargeAmountChange(e.target.value)}
          />
          <div className="flex gap-2">
            {[10, 50, 100, 200].map((amount) => (
              <button
                key={amount}
                onClick={() => handleRechargeAmountChange(amount.toString())}
                className={`flex-1 py-2 rounded-lg border transition-colors ${
                  rechargeAmount === amount.toString()
                    ? 'bg-primary-500/20 border-primary-500 text-primary-400'
                    : 'border-dark-600 text-dark-400 hover:border-dark-500'
                }`}
              >
                ¥{amount}
              </button>
            ))}
          </div>

          {/* 优惠计算结果 */}
          {calculatingPromo && (
            <div className="text-center py-2 text-dark-400 text-sm">
              <i className="fas fa-spinner fa-spin mr-2" />
              计算优惠中...
            </div>
          )}
          {promoResult && !calculatingPromo && (
            <div className="p-4 bg-dark-700/50 rounded-xl space-y-2">
              {promoResult.promo_id > 0 && (
                <div className="flex items-center gap-2 text-green-400 text-sm mb-2">
                  <i className="fas fa-check-circle" />
                  <span>已享受优惠：{promoResult.promo_name}</span>
                </div>
              )}
              <div className="flex justify-between text-sm">
                <span className="text-dark-400">充值金额</span>
                <span className="text-dark-200">¥{promoResult.original_amount.toFixed(2)}</span>
              </div>
              {promoResult.discount_amount > 0 && (
                <div className="flex justify-between text-sm">
                  <span className="text-dark-400">折扣优惠</span>
                  <span className="text-green-400">-¥{promoResult.discount_amount.toFixed(2)}</span>
                </div>
              )}
              <div className="flex justify-between text-sm">
                <span className="text-dark-400">实际支付</span>
                <span className="text-primary-400 font-medium">¥{promoResult.pay_amount.toFixed(2)}</span>
              </div>
              {promoResult.bonus_amount > 0 && (
                <div className="flex justify-between text-sm">
                  <span className="text-dark-400">赠送金额</span>
                  <span className="text-green-400">+¥{promoResult.bonus_amount.toFixed(2)}</span>
                </div>
              )}
              <div className="border-t border-dark-600 pt-2 mt-2">
                <div className="flex justify-between">
                  <span className="text-dark-300 font-medium">实际到账</span>
                  <span className="text-xl font-bold text-green-400">¥{promoResult.total_credit.toFixed(2)}</span>
                </div>
              </div>
            </div>
          )}

          <Input
            label="支付密码"
            type="password"
            placeholder="请输入6位支付密码"
            maxLength={6}
            value={rechargePayPassword}
            onChange={(e) => setRechargePayPassword(e.target.value.replace(/\D/g, ''))}
          />
          <Button className="w-full" onClick={handleRecharge}>
            确认充值
          </Button>
        </div>
      </Modal>

      {/* 兑换确认弹窗 */}
      <Modal isOpen={showExchangeModal} onClose={() => setShowExchangeModal(false)} title="确认兑换" size="sm">
        {selectedExchange && (
          <div className="space-y-4">
            <div className="p-4 bg-dark-700/30 rounded-xl">
              <div className="text-lg font-medium text-dark-100 mb-2">{selectedExchange.name}</div>
              <div className="text-dark-400">
                消耗积分: <span className="text-amber-400 font-medium">{selectedExchange.points_cost}</span>
              </div>
            </div>
            <div className="flex gap-3">
              <Button variant="secondary" className="flex-1" onClick={() => setShowExchangeModal(false)}>
                取消
              </Button>
              <Button className="flex-1" onClick={handleExchange}>
                确认兑换
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* 设置支付密码弹窗 */}
      <Modal 
        isOpen={showSetPayPasswordModal} 
        onClose={() => {
          setShowSetPayPasswordModal(false)
          setPayPassword('')
          setConfirmPayPassword('')
          setLoginPassword('')
        }} 
        title="设置支付密码" 
        size="sm"
      >
        <div className="space-y-4">
          <div className="text-sm text-dark-400 mb-2">
            支付密码用于余额支付和充值操作，请设置6位纯数字密码
          </div>
          <Input
            label="支付密码"
            type="password"
            placeholder="请输入6位数字"
            maxLength={6}
            value={payPassword}
            onChange={(e) => setPayPassword(e.target.value.replace(/\D/g, ''))}
          />
          <Input
            label="确认支付密码"
            type="password"
            placeholder="请再次输入"
            maxLength={6}
            value={confirmPayPassword}
            onChange={(e) => setConfirmPayPassword(e.target.value.replace(/\D/g, ''))}
          />
          <Input
            label="登录密码"
            type="password"
            placeholder="请输入登录密码验证身份"
            value={loginPassword}
            onChange={(e) => setLoginPassword(e.target.value)}
          />
          <Button className="w-full" onClick={handleSetPayPassword}>
            确认设置
          </Button>
        </div>
      </Modal>

      {/* 修改支付密码弹窗 */}
      <Modal 
        isOpen={showUpdatePayPasswordModal} 
        onClose={() => {
          setShowUpdatePayPasswordModal(false)
          setPayPassword('')
          setConfirmPayPassword('')
          setOldPayPassword('')
        }} 
        title="修改支付密码" 
        size="sm"
      >
        <div className="space-y-4">
          <Input
            label="原支付密码"
            type="password"
            placeholder="请输入原支付密码"
            maxLength={6}
            value={oldPayPassword}
            onChange={(e) => setOldPayPassword(e.target.value.replace(/\D/g, ''))}
          />
          <Input
            label="新支付密码"
            type="password"
            placeholder="请输入6位数字"
            maxLength={6}
            value={payPassword}
            onChange={(e) => setPayPassword(e.target.value.replace(/\D/g, ''))}
          />
          <Input
            label="确认新密码"
            type="password"
            placeholder="请再次输入"
            maxLength={6}
            value={confirmPayPassword}
            onChange={(e) => setConfirmPayPassword(e.target.value.replace(/\D/g, ''))}
          />
          <Button className="w-full" onClick={handleUpdatePayPassword}>
            确认修改
          </Button>
        </div>
      </Modal>

      {/* 重置支付密码弹窗 */}
      <Modal 
        isOpen={showResetPayPasswordModal} 
        onClose={() => {
          setShowResetPayPasswordModal(false)
          setPayPassword('')
          setConfirmPayPassword('')
          setEmailCode('')
        }} 
        title="重置支付密码" 
        size="sm"
      >
        <div className="space-y-4">
          <div className="text-sm text-dark-400 mb-2">
            通过邮箱验证码重置支付密码
          </div>
          <div className="flex gap-2">
            <Input
              label="邮箱验证码"
              placeholder="请输入验证码"
              value={emailCode}
              onChange={(e) => setEmailCode(e.target.value)}
              className="flex-1"
            />
            <Button 
              variant="secondary" 
              className="mt-6"
              onClick={handleSendResetCode}
              disabled={sendingCode || countdown > 0}
            >
              {countdown > 0 ? `${countdown}s` : '发送验证码'}
            </Button>
          </div>
          <Input
            label="新支付密码"
            type="password"
            placeholder="请输入6位数字"
            maxLength={6}
            value={payPassword}
            onChange={(e) => setPayPassword(e.target.value.replace(/\D/g, ''))}
          />
          <Input
            label="确认新密码"
            type="password"
            placeholder="请再次输入"
            maxLength={6}
            value={confirmPayPassword}
            onChange={(e) => setConfirmPayPassword(e.target.value.replace(/\D/g, ''))}
          />
          <Button className="w-full" onClick={handleResetPayPassword}>
            确认重置
          </Button>
        </div>
      </Modal>
    </motion.div>
  )
}
