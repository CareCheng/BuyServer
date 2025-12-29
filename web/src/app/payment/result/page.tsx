'use client'

import { useState, useEffect, Suspense } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Navbar, Footer } from '@/components/layout'
import { Button, Card } from '@/components/ui'
import { apiPost, apiGet } from '@/lib/api'
import { copyToClipboard } from '@/lib/utils'

/**
 * 支付结果接口
 */
interface PaymentResult {
  order_no: string
  kami_code: string
  product_name?: string
  quantity?: number
}

/**
 * 充值结果接口
 */
interface RechargeResult {
  recharge_no: string
  amount: number
  bonus_amount: number
  total_credit: number
}

/**
 * 支付结果页面内容组件
 */
function PaymentResultContent() {
  const searchParams = useSearchParams()
  const router = useRouter()

  const paypalOrderId = searchParams.get('paypal_order_id') || searchParams.get('token')
  const yipayOrderNo = searchParams.get('out_trade_no')
  const yipayTradeNo = searchParams.get('trade_no')
  const directOrderNo = searchParams.get('order_no')
  const directRechargeNo = searchParams.get('recharge_no')
  const directStatus = searchParams.get('status')
  const stripeSessionId = searchParams.get('session_id')
  const paymentType = searchParams.get('type') // 'recharge' 表示充值订单

  const [loading, setLoading] = useState(true)
  const [result, setResult] = useState<PaymentResult | null>(null)
  const [rechargeResult, setRechargeResult] = useState<RechargeResult | null>(null)
  const [error, setError] = useState<string>('')
  const [isRecharge, setIsRecharge] = useState(false)

  useEffect(() => {
    const processPayment = async () => {
      // 从 sessionStorage 获取订单号
      const pendingOrderNo = sessionStorage.getItem('pending_order_no')
      const pendingRechargeNo = sessionStorage.getItem('pending_recharge_no')
      
      // 判断是否为充值订单
      const isRechargeOrder = paymentType === 'recharge' || !!directRechargeNo || !!pendingRechargeNo || 
        !!(yipayOrderNo && yipayOrderNo.startsWith('RC'))
      setIsRecharge(isRechargeOrder)

      // 检查是否是从二维码支付页面跳转过来
      const fromQRCode = searchParams.get('from') === 'qrcode'
      if (fromQRCode) {
        const paymentResult = sessionStorage.getItem('payment_result')
        if (paymentResult) {
          try {
            const parsed = JSON.parse(paymentResult)
            setResult({
              order_no: parsed.order_no,
              kami_code: parsed.kami_code,
            })
            sessionStorage.removeItem('payment_result')
            setLoading(false)
            return
          } catch {
            // 解析失败，继续其他处理
          }
        }
      }

      // 处理 Stripe 支付回调（带 session_id）
      if (stripeSessionId && directOrderNo) {
        await processStripe(stripeSessionId, directOrderNo)
        return
      }

      // 处理直接跳转的支付结果（余额支付等）
      if (directOrderNo && directStatus === 'success') {
        await loadOrderResult(directOrderNo)
        return
      }

      // 处理直接跳转的充值结果
      if (directRechargeNo && directStatus === 'success') {
        await loadRechargeResult(directRechargeNo)
        return
      }

      // 处理 PayPal 支付回调
      if (paypalOrderId) {
        await processPayPal(paypalOrderId, pendingOrderNo)
        return
      }

      // 处理易支付回调
      if (yipayOrderNo && yipayTradeNo) {
        if (isRechargeOrder) {
          await processYiPayRecharge(yipayOrderNo, yipayTradeNo)
        } else {
          await processYiPay(yipayOrderNo, yipayTradeNo)
        }
        return
      }

      // 没有有效的支付信息
      setError('无效的支付信息')
      setLoading(false)
    }

    processPayment()
  }, [paypalOrderId, yipayOrderNo, yipayTradeNo, directOrderNo, directRechargeNo, directStatus, stripeSessionId, paymentType, searchParams])

  // 加载订单结果（用于余额支付等直接跳转的情况）
  const loadOrderResult = async (orderNo: string) => {
    const res = await apiGet<{ order: { order_no: string; kami_code: string; product_name: string; quantity: number } }>(`/api/order/detail/${orderNo}`)
    
    if (res.success && res.order) {
      setResult({
        order_no: res.order.order_no,
        kami_code: res.order.kami_code,
        product_name: res.order.product_name,
        quantity: res.order.quantity,
      })
    } else {
      setError(res.error || '获取订单信息失败')
    }
    setLoading(false)
  }

  // 加载充值结果
  const loadRechargeResult = async (rechargeNo: string) => {
    const res = await apiGet<{ data: { recharge_no: string; amount: number; bonus_amount: number; total_credit: number; status: number } }>(`/api/user/balance/recharge/${rechargeNo}`)
    
    if (res.success && res.data) {
      if (res.data.status === 1) { // 已支付
        setRechargeResult({
          recharge_no: res.data.recharge_no,
          amount: res.data.amount,
          bonus_amount: res.data.bonus_amount,
          total_credit: res.data.total_credit,
        })
      } else {
        setError('充值订单未完成支付')
      }
    } else {
      setError(res.error || '获取充值订单信息失败')
    }
    setLoading(false)
  }

  // 处理 Stripe 支付
  const processStripe = async (sessionId: string, orderNo: string) => {
    // 验证 Stripe 支付状态
    const verifyRes = await apiGet<{ data: { status: string; order_no: string } }>(`/api/stripe/verify/${sessionId}`)
    
    if (verifyRes.success && verifyRes.data?.status === 'complete') {
      // 支付成功，加载订单信息（Webhook 应该已经处理了订单）
      await new Promise(resolve => setTimeout(resolve, 1000))
      await loadOrderResult(orderNo)
      sessionStorage.removeItem('pending_order_no')
    } else if (verifyRes.success && verifyRes.data?.status === 'open') {
      setError('支付未完成，请重新支付')
      setLoading(false)
    } else {
      await loadOrderResult(orderNo)
      sessionStorage.removeItem('pending_order_no')
    }
  }

  // 处理 PayPal 支付
  const processPayPal = async (paypalOrderId: string, orderNo: string | null) => {
    if (!orderNo) {
      setError('订单信息丢失，请返回订单页面重新支付')
      setLoading(false)
      return
    }

    const res = await apiPost<PaymentResult>('/api/paypal/capture', {
      order_no: orderNo,
      paypal_order_id: paypalOrderId,
    })

    if (res.success) {
      setResult({
        order_no: res.order_no,
        kami_code: res.kami_code,
      })
      sessionStorage.removeItem('pending_order_no')
    } else {
      setError(res.error || '支付确认失败')
    }
    setLoading(false)
  }

  // 处理易支付（商品订单）
  const processYiPay = async (orderNo: string, tradeNo: string) => {
    const res = await apiPost<PaymentResult>('/api/yipay/callback', {
      out_trade_no: orderNo,
      trade_no: tradeNo,
    })

    if (res.success) {
      setResult({
        order_no: res.order_no,
        kami_code: res.kami_code,
      })
      sessionStorage.removeItem('pending_order_no')
    } else {
      setError(res.error || '支付确认失败')
    }
    setLoading(false)
  }

  // 处理易支付（充值订单）
  const processYiPayRecharge = async (rechargeNo: string, tradeNo: string) => {
    const res = await apiPost<{ recharge_no: string; amount: number; bonus_amount: number; total_credit: number }>('/api/yipay/recharge/callback', {
      out_trade_no: rechargeNo,
      trade_no: tradeNo,
    })

    if (res.success) {
      setRechargeResult({
        recharge_no: res.recharge_no,
        amount: res.amount,
        bonus_amount: res.bonus_amount,
        total_credit: res.total_credit,
      })
      sessionStorage.removeItem('pending_recharge_no')
    } else {
      setError(res.error || '充值确认失败')
    }
    setLoading(false)
  }

  // 复制卡密
  const handleCopyKami = async () => {
    if (result?.kami_code) {
      const success = await copyToClipboard(result.kami_code)
      if (success) {
        toast.success('卡密已复制到剪贴板')
      }
    }
  }

  // 导出卡密为 CSV
  const handleExportKami = () => {
    if (!result?.kami_code) return
    
    const kamiCodes = result.kami_code.split('\n').filter(code => code.trim())
    if (kamiCodes.length <= 1) return
    
    // 生成 CSV 内容
    const csvContent = [
      '序号,卡密',
      ...kamiCodes.map((code, index) => `${index + 1},"${code.trim()}"`)
    ].join('\n')
    
    // 添加 BOM 以支持中文
    const BOM = '\uFEFF'
    const blob = new Blob([BOM + csvContent], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `卡密_${result.order_no}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    toast.success('卡密已导出')
  }

  // 获取卡密数量
  const getKamiCount = () => {
    if (!result?.kami_code) return 0
    return result.kami_code.split('\n').filter(code => code.trim()).length
  }

  // 加载中
  if (loading) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 flex items-center justify-center">
          <div className="text-center">
            <i className="fas fa-spinner fa-spin text-5xl text-primary-400 mb-4" />
            <p className="text-dark-300 text-lg">正在确认支付结果...</p>
            <p className="text-dark-500 text-sm mt-2">请稍候，不要关闭此页面</p>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  // 支付失败
  if (error) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 py-8 px-4">
          <div className="max-w-lg mx-auto">
            <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
              <Card>
                <div className="text-center py-8">
                  <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-red-500/20 flex items-center justify-center">
                    <i className="fas fa-times text-4xl text-red-400" />
                  </div>
                  <h2 className="text-2xl font-bold text-dark-100 mb-2">
                    {isRecharge ? '充值失败' : '支付失败'}
                  </h2>
                  <p className="text-dark-400 mb-6">{error}</p>
                  <div className="flex flex-col sm:flex-row gap-4 justify-center">
                    <Button variant="secondary" onClick={() => router.push('/user')}>
                      {isRecharge ? '返回钱包' : '查看订单'}
                    </Button>
                    <Button variant="primary" onClick={() => router.push(isRecharge ? '/user' : '/products')}>
                      {isRecharge ? '重新充值' : '继续购买'}
                    </Button>
                  </div>
                </div>
              </Card>
            </motion.div>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  // 充值成功
  if (isRecharge && rechargeResult) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 py-8 px-4">
          <div className="max-w-lg mx-auto">
            <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
              <Card>
                <div className="text-center py-6">
                  <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-emerald-500/20 flex items-center justify-center">
                    <i className="fas fa-check text-4xl text-emerald-400" />
                  </div>
                  <h2 className="text-2xl font-bold text-dark-100 mb-2">🎉 充值成功</h2>
                  <p className="text-dark-400 mb-6">余额已到账！</p>

                  {/* 充值信息 */}
                  <div className="bg-dark-700/30 rounded-xl p-4 mb-6 text-left">
                    <div className="space-y-3">
                      <div className="flex justify-between">
                        <span className="text-dark-500">充值单号</span>
                        <span className="text-dark-100 font-mono text-sm">{rechargeResult.recharge_no}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-dark-500">充值金额</span>
                        <span className="text-dark-100">¥{rechargeResult.amount.toFixed(2)}</span>
                      </div>
                      {rechargeResult.bonus_amount > 0 && (
                        <div className="flex justify-between text-green-400">
                          <span>赠送金额</span>
                          <span>+¥{rechargeResult.bonus_amount.toFixed(2)}</span>
                        </div>
                      )}
                      <div className="border-t border-dark-600 pt-3 flex justify-between">
                        <span className="text-dark-300 font-medium">实际到账</span>
                        <span className="text-green-400 text-xl font-bold">
                          ¥{(rechargeResult.total_credit > 0 ? rechargeResult.total_credit : rechargeResult.amount).toFixed(2)}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* 操作按钮 */}
                  <div className="flex flex-col sm:flex-row gap-4">
                    <Button variant="secondary" className="flex-1" onClick={() => router.push('/user')}>
                      <i className="fas fa-wallet mr-2" />
                      查看钱包
                    </Button>
                    <Button variant="primary" className="flex-1" onClick={() => router.push('/products')}>
                      <i className="fas fa-cart-shopping mr-2" />
                      去购物
                    </Button>
                  </div>
                </div>
              </Card>
            </motion.div>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  // 商品订单支付成功
  const kamiCount = getKamiCount()
  
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <main className="flex-1 py-8 px-4">
        <div className="max-w-lg mx-auto">
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
            <Card>
              <div className="text-center py-6">
                <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-emerald-500/20 flex items-center justify-center">
                  <i className="fas fa-check text-4xl text-emerald-400" />
                </div>
                <h2 className="text-2xl font-bold text-dark-100 mb-2">🎉 支付成功</h2>
                <p className="text-dark-400 mb-6">感谢您的购买！</p>

                {/* 订单信息 */}
                <div className="bg-dark-700/30 rounded-xl p-4 mb-6 text-left">
                  <div className="space-y-3">
                    <div>
                      <span className="text-dark-500 text-sm">订单号</span>
                      <p className="text-dark-100 font-mono">{result?.order_no}</p>
                    </div>
                    <div>
                      <div className="flex items-center justify-between mb-1">
                        <span className="text-dark-500 text-sm">卡密 {kamiCount > 1 && `(${kamiCount}个)`}</span>
                        {kamiCount > 1 && (
                          <Button size="sm" variant="ghost" onClick={handleExportKami} className="text-xs">
                            <i className="fas fa-download mr-1" />
                            导出CSV
                          </Button>
                        )}
                      </div>
                      <div className="mt-1 bg-dark-800/50 rounded-lg p-3">
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex-1 space-y-2 max-h-60 overflow-y-auto">
                            {result?.kami_code?.split('\n').filter(code => code.trim()).map((code, index) => (
                              <div key={index} className="flex items-center gap-2">
                                {kamiCount > 1 && (
                                  <span className="text-dark-500 text-sm w-6 flex-shrink-0">{index + 1}.</span>
                                )}
                                <p className="font-mono text-primary-400 break-all text-lg flex-1">
                                  {code.trim()}
                                </p>
                              </div>
                            ))}
                          </div>
                          <Button size="sm" variant="ghost" onClick={handleCopyKami} className="flex-shrink-0">
                            <i className="fas fa-copy" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* 提示信息 */}
                <div className="bg-blue-500/10 border border-blue-500/30 rounded-xl p-4 mb-6 text-left">
                  <div className="flex items-start gap-3">
                    <i className="fas fa-info-circle text-blue-400 mt-0.5" />
                    <div className="text-sm">
                      <p className="text-blue-300 font-medium">请妥善保管您的卡密</p>
                      <p className="text-blue-400/70 mt-1">
                        卡密是您使用服务的凭证，请勿泄露给他人。您可以在用户中心随时查看。
                      </p>
                    </div>
                  </div>
                </div>

                {/* 操作按钮 */}
                <div className="flex flex-col sm:flex-row gap-4">
                  <Button variant="secondary" className="flex-1" onClick={() => router.push('/user')}>
                    <i className="fas fa-user mr-2" />
                    用户中心
                  </Button>
                  <Button variant="primary" className="flex-1" onClick={handleCopyKami}>
                    <i className="fas fa-copy mr-2" />
                    复制卡密
                  </Button>
                </div>
              </div>
            </Card>
          </motion.div>
        </div>
      </main>
      <Footer />
    </div>
  )
}

/**
 * 支付结果页面
 * 处理各种支付方式的回调，支持商品订单和充值订单
 */
export default function PaymentResultPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center bg-dark-900">
          <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
        </div>
      }
    >
      <PaymentResultContent />
    </Suspense>
  )
}
