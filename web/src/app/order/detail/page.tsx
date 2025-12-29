'use client'

import { useState, useEffect, Suspense } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Navbar, Footer } from '@/components/layout'
import { Button, Badge, Card } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime, getOrderStatus, copyToClipboard } from '@/lib/utils'

/**
 * 订单详情接口
 */
interface OrderDetail {
  id: number
  order_no: string
  user_id: number
  username: string
  product_id: number
  product_name: string
  quantity: number
  price: number
  original_price: number
  coupon_code: string
  coupon_discount: number
  status: number
  is_test: boolean
  kami_code: string
  payment_method: string
  payment_no: string
  payment_time: string
  created_at: string
  updated_at: string
  duration: number
  duration_unit: string
}

/**
 * 订单详情内容组件
 */
function OrderDetailContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const orderNo = searchParams.get('order_no')
  
  const [order, setOrder] = useState<OrderDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [canceling, setCanceling] = useState(false)
  // 取消订单确认弹窗状态
  const [showCancelConfirm, setShowCancelConfirm] = useState(false)

  // 加载订单详情
  useEffect(() => {
    if (!orderNo) {
      setError('订单号不能为空')
      setLoading(false)
      return
    }

    const loadOrder = async () => {
      setLoading(true)
      const res = await apiGet<{ order: OrderDetail }>(`/api/order/detail/${orderNo}`)
      if (res.success && res.order) {
        setOrder(res.order)
      } else {
        setError(res.error || '获取订单详情失败')
      }
      setLoading(false)
    }
    loadOrder()
  }, [orderNo])

  // 确认取消订单
  const confirmCancelOrder = async () => {
    if (!order) return
    
    setCanceling(true)
    const res = await apiPost('/api/order/cancel', { order_no: order.order_no })
    setCanceling(false)
    setShowCancelConfirm(false)
    
    if (res.success) {
      toast.success('订单已取消')
      // 刷新订单状态
      setOrder({ ...order, status: 3 })
    } else {
      toast.error(res.error || '取消订单失败')
    }
  }

  // 继续支付
  const handlePay = () => {
    if (order) {
      router.push(`/payment?order_no=${order.order_no}`)
    }
  }

  // 复制卡密
  const handleCopyKami = async () => {
    if (order?.kami_code) {
      const success = await copyToClipboard(order.kami_code)
      if (success) {
        toast.success('卡密已复制到剪贴板')
      }
    }
  }

  // 获取状态信息
  const statusInfo = order ? getOrderStatus(order.status) : null

  // 获取支付方式名称
  const getPaymentMethodName = (method: string) => {
    const methods: Record<string, string> = {
      'paypal': 'PayPal',
      'alipay': '支付宝',
      'wechat': '微信支付',
      'yipay': '易支付',
      'stripe': 'Stripe',
      'usdt': 'USDT',
      'balance': '余额支付',
      'test': '测试支付',
    }
    return methods[method] || method || '未支付'
  }

  if (loading) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 flex items-center justify-center">
          <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
        </main>
        <Footer />
      </div>
    )
  }

  if (error || !order) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 py-8 px-4">
          <div className="max-w-lg mx-auto text-center">
            <div className="text-6xl mb-4">😕</div>
            <h1 className="text-2xl font-bold text-dark-100 mb-2">订单不存在</h1>
            <p className="text-dark-400 mb-6">{error || '无法找到该订单'}</p>
            <Button variant="primary" onClick={() => router.push('/user')}>
              返回用户中心
            </Button>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />

      <main className="flex-1 py-8 px-4">
        <div className="max-w-3xl mx-auto">
          {/* 页面标题 */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6"
          >
            <button
              onClick={() => router.back()}
              className="text-dark-400 hover:text-dark-200 mb-4 flex items-center"
            >
              <i className="fas fa-arrow-left mr-2" />
              返回
            </button>
            <h1 className="text-2xl font-bold text-dark-100">订单详情</h1>
          </motion.div>

          {/* 订单状态卡片 */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <Card className="mb-6">
              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <div className={`w-12 h-12 rounded-full flex items-center justify-center mr-4 ${
                    order.status === 2 ? 'bg-emerald-500/20' :
                    order.status === 0 ? 'bg-amber-500/20' :
                    order.status === 3 ? 'bg-red-500/20' :
                    'bg-dark-700/50'
                  }`}>
                    <i className={`fas ${
                      order.status === 2 ? 'fa-check text-emerald-400' :
                      order.status === 0 ? 'fa-clock text-amber-400' :
                      order.status === 3 ? 'fa-times text-red-400' :
                      'fa-receipt text-dark-400'
                    } text-xl`} />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      {statusInfo && (
                        <Badge variant={statusInfo.variant as 'success' | 'warning' | 'danger' | 'info'}>
                          {statusInfo.text}
                        </Badge>
                      )}
                      {order.is_test && (
                        <Badge variant="warning">测试订单</Badge>
                      )}
                    </div>
                    <p className="text-dark-400 text-sm mt-1">
                      订单号: {order.order_no}
                    </p>
                  </div>
                </div>
                
                {/* 操作按钮 */}
                <div className="flex gap-2">
                  {order.status === 0 && (
                    <>
                      <Button variant="secondary" size="sm" onClick={() => setShowCancelConfirm(true)} loading={canceling}>
                        取消订单
                      </Button>
                      <Button variant="primary" size="sm" onClick={handlePay}>
                        继续支付
                      </Button>
                    </>
                  )}
                </div>
              </div>
            </Card>
          </motion.div>

          {/* 卡密信息（已完成订单显示） */}
          {order.status === 2 && order.kami_code && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.15 }}
            >
              <Card className="mb-6 border-emerald-500/30">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-dark-200 font-medium mb-2">
                      <i className="fas fa-key text-emerald-400 mr-2" />
                      卡密信息
                      {order.quantity > 1 && (
                        <span className="text-dark-400 text-sm ml-2">({order.quantity}个)</span>
                      )}
                    </h3>
                    <div className="space-y-2">
                      {order.kami_code.split('\n').map((code, index) => (
                        <p key={index} className="text-emerald-400 font-mono text-lg break-all">
                          {order.quantity > 1 && <span className="text-dark-400 text-sm mr-2">{index + 1}.</span>}
                          {code}
                        </p>
                      ))}
                    </div>
                  </div>
                  <Button variant="primary" size="sm" onClick={handleCopyKami} className="ml-4 flex-shrink-0">
                    <i className="fas fa-copy mr-2" />
                    复制
                  </Button>
                </div>
              </Card>
            </motion.div>
          )}

          {/* 商品信息 */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
          >
            <Card className="mb-6">
              <h3 className="text-dark-200 font-medium mb-4">
                <i className="fas fa-box text-primary-400 mr-2" />
                商品信息
              </h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-dark-400">商品名称</span>
                  <span className="text-dark-100">{order.product_name}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-dark-400">购买数量</span>
                  <span className="text-dark-100">{order.quantity || 1}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-dark-400">有效期</span>
                  <span className="text-dark-100">{order.duration} {order.duration_unit}</span>
                </div>
                {order.original_price > order.price && (
                  <div className="flex justify-between">
                    <span className="text-dark-400">原价</span>
                    <span className="text-dark-500 line-through">¥{order.original_price.toFixed(2)}</span>
                  </div>
                )}
                {order.coupon_code && (
                  <div className="flex justify-between">
                    <span className="text-dark-400">优惠券</span>
                    <span className="text-emerald-400">-¥{order.coupon_discount.toFixed(2)} ({order.coupon_code})</span>
                  </div>
                )}
                <div className="flex justify-between pt-3 border-t border-dark-700/50">
                  <span className="text-dark-200 font-medium">实付金额</span>
                  <span className="text-primary-400 font-bold text-xl">¥{order.price.toFixed(2)}</span>
                </div>
              </div>
            </Card>
          </motion.div>

          {/* 支付信息 */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.25 }}
          >
            <Card className="mb-6">
              <h3 className="text-dark-200 font-medium mb-4">
                <i className="fas fa-credit-card text-primary-400 mr-2" />
                支付信息
              </h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-dark-400">支付方式</span>
                  <span className="text-dark-100">{getPaymentMethodName(order.payment_method)}</span>
                </div>
                {order.payment_no && (
                  <div className="flex justify-between">
                    <span className="text-dark-400">支付单号</span>
                    <span className="text-dark-100 font-mono text-sm">{order.payment_no}</span>
                  </div>
                )}
                {order.payment_time && (
                  <div className="flex justify-between">
                    <span className="text-dark-400">支付时间</span>
                    <span className="text-dark-100">{formatDateTime(order.payment_time)}</span>
                  </div>
                )}
              </div>
            </Card>
          </motion.div>

          {/* 订单时间 */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
          >
            <Card>
              <h3 className="text-dark-200 font-medium mb-4">
                <i className="fas fa-clock text-primary-400 mr-2" />
                订单时间
              </h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-dark-400">创建时间</span>
                  <span className="text-dark-100">{formatDateTime(order.created_at)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-dark-400">更新时间</span>
                  <span className="text-dark-100">{formatDateTime(order.updated_at)}</span>
                </div>
              </div>
            </Card>
          </motion.div>
        </div>
      </main>

      <Footer />

      {/* 取消订单确认弹窗 */}
      <ConfirmModal
        isOpen={showCancelConfirm}
        onClose={() => setShowCancelConfirm(false)}
        title="取消订单"
        message="确定要取消此订单吗？取消后将无法恢复。"
        confirmText="取消订单"
        variant="danger"
        onConfirm={confirmCancelOrder}
        loading={canceling}
      />
    </div>
  )
}

/**
 * 订单详情页面
 * 使用 Suspense 包裹以支持 useSearchParams
 */
export default function OrderDetailPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 flex items-center justify-center">
          <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
        </main>
        <Footer />
      </div>
    }>
      <OrderDetailContent />
    </Suspense>
  )
}
