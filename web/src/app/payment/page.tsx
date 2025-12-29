'use client'

import { useState, useEffect, Suspense, useCallback } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Navbar, Footer } from '@/components/layout'
import { Button, Card, Input } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { formatMoney, formatDateTime } from '@/lib/utils'

/**
 * 订单接口
 */
interface Order {
  id: number
  order_no: string
  product_id: number
  category_id: number
  product_name: string
  quantity: number
  price: number
  duration: number
  duration_unit: string
  status: number
  created_at: string
}

/**
 * 充值订单接口
 */
interface RechargeOrder {
  id: number
  recharge_no: string
  user_id: number
  amount: number
  pay_amount: number
  bonus_amount: number
  total_credit: number
  promo_id: number
  promo_name: string
  payment_method: string
  payment_no: string
  status: number
  expire_at: string
  created_at: string
}

/**
 * 优惠券验证结果接口
 */
interface CouponValidateResult {
  coupon_id: number
  coupon_name: string
  coupon_type: string
  discount: number
  final_amount: number
}

/**
 * 支付配置接口
 */
interface PaymentMethods {
  paypal: { enabled: boolean; sandbox: boolean }
  alipay_f2f: { enabled: boolean }
  wechat_pay: { enabled: boolean }
  yi_pay: { enabled: boolean }
  stripe: { enabled: boolean }
  usdt: { enabled: boolean; network: string }
  balance: { enabled: boolean }
}

/**
 * 支付页面内容组件
 */
function PaymentPageContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const orderNo = searchParams.get('order_no')
  const rechargeNo = searchParams.get('recharge_no')
  const paymentType = searchParams.get('type') // 'order' 或 'recharge'

  const [order, setOrder] = useState<Order | null>(null)
  const [rechargeOrder, setRechargeOrder] = useState<RechargeOrder | null>(null)
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethods | null>(null)
  const [loading, setLoading] = useState(true)
  const [paying, setPaying] = useState(false)
  const [selectedMethod, setSelectedMethod] = useState<string>('')
  
  // 优惠券相关状态（仅商品订单使用）
  const [couponCode, setCouponCode] = useState('')
  const [couponValidating, setCouponValidating] = useState(false)
  const [appliedCoupon, setAppliedCoupon] = useState<CouponValidateResult | null>(null)
  const [finalPrice, setFinalPrice] = useState<number>(0)
  
  // 倒计时相关状态
  const [countdown, setCountdown] = useState<number>(0)
  
  // 用户余额（仅商品订单使用）
  const [userBalance, setUserBalance] = useState<number>(0)
  
  // 支付密码相关（仅余额支付使用）
  const [payPasswordSet, setPayPasswordSet] = useState<boolean>(false)
  const [balancePayPassword, setBalancePayPassword] = useState<string>('')
  const [showPayPasswordInput, setShowPayPasswordInput] = useState<boolean>(false)

  // 判断是否为充值订单
  const isRecharge = paymentType === 'recharge' || !!rechargeNo

  // 加载充值订单数据
  const loadRechargeOrder = async () => {
    const targetNo = rechargeNo
    if (!targetNo) return false

    const res = await apiGet<{ data: RechargeOrder }>(`/api/user/balance/recharge/${targetNo}`)
    if (!res.success) {
      if (res.error === '请先登录') {
        router.push('/login')
      } else {
        toast.error(res.error || '充值订单不存在')
        router.push('/user')
      }
      return false
    }

    if (res.data.status !== 0) {
      toast.error('订单状态异常，无法支付')
      router.push('/user')
      return false
    }

    setRechargeOrder(res.data)
    // 充值订单使用实际支付金额
    const payAmount = res.data.pay_amount > 0 ? res.data.pay_amount : res.data.amount
    setFinalPrice(payAmount)
    
    // 计算倒计时（充值订单使用 expire_at）
    if (res.data.expire_at) {
      const expireAt = new Date(res.data.expire_at).getTime()
      const remaining = Math.floor((expireAt - Date.now()) / 1000)
      setCountdown(remaining > 0 ? remaining : 0)
    } else {
      // 默认30分钟
      const createdAt = new Date(res.data.created_at).getTime()
      const expireAt = createdAt + 30 * 60 * 1000
      const remaining = Math.floor((expireAt - Date.now()) / 1000)
      setCountdown(remaining > 0 ? remaining : 0)
    }

    return true
  }

  // 加载商品订单数据
  const loadOrder = async () => {
    if (!orderNo) return false

    const orderRes = await apiGet<{ order: Order }>(`/api/order/detail/${orderNo}`)
    if (!orderRes.success) {
      if (orderRes.error === '请先登录') {
        router.push('/login')
      } else {
        toast.error(orderRes.error || '订单不存在')
        router.push('/products')
      }
      return false
    }

    if (orderRes.order.status !== 0) {
      toast.error('订单状态异常，无法支付')
      router.push('/user')
      return false
    }

    setOrder(orderRes.order)
    setFinalPrice(orderRes.order.price)
    
    // 计算倒计时（订单创建后30分钟过期）
    const createdAt = new Date(orderRes.order.created_at).getTime()
    const expireAt = createdAt + 30 * 60 * 1000
    const remaining = Math.floor((expireAt - Date.now()) / 1000)
    setCountdown(remaining > 0 ? remaining : 0)

    return true
  }

  // 加载订单和支付配置
  useEffect(() => {
    // 充值订单和商品订单至少需要一个
    if (!orderNo && !rechargeNo) {
      toast.error('订单号无效')
      router.push('/products')
      return
    }

    const loadData = async () => {
      let success = false

      if (isRecharge) {
        // 加载充值订单
        success = await loadRechargeOrder()
      } else {
        // 加载商品订单
        success = await loadOrder()
      }

      if (!success) return

      // 加载支付方式配置
      const paymentRes = await apiGet<{ methods: PaymentMethods }>('/api/payment/methods')
      if (paymentRes.success && paymentRes.methods) {
        setPaymentMethods(paymentRes.methods)
        // 充值订单不支持余额支付，自动选择其他方式
        if (isRecharge) {
          if (paymentRes.methods.yi_pay?.enabled) {
            setSelectedMethod('yi_pay')
          } else if (paymentRes.methods.paypal?.enabled) {
            setSelectedMethod('paypal')
          } else if (paymentRes.methods.stripe?.enabled) {
            setSelectedMethod('stripe')
          } else if (paymentRes.methods.alipay_f2f?.enabled) {
            setSelectedMethod('alipay_f2f')
          } else if (paymentRes.methods.wechat_pay?.enabled) {
            setSelectedMethod('wechat_pay')
          } else if (paymentRes.methods.usdt?.enabled) {
            setSelectedMethod('usdt')
          }
        } else {
          // 商品订单可以使用余额支付
          if (paymentRes.methods.balance?.enabled) {
            setSelectedMethod('balance')
          } else if (paymentRes.methods.yi_pay?.enabled) {
            setSelectedMethod('yi_pay')
          } else if (paymentRes.methods.paypal?.enabled) {
            setSelectedMethod('paypal')
          } else if (paymentRes.methods.stripe?.enabled) {
            setSelectedMethod('stripe')
          } else if (paymentRes.methods.alipay_f2f?.enabled) {
            setSelectedMethod('alipay_f2f')
          } else if (paymentRes.methods.wechat_pay?.enabled) {
            setSelectedMethod('wechat_pay')
          } else if (paymentRes.methods.usdt?.enabled) {
            setSelectedMethod('usdt')
          }
        }
      }
      
      // 商品订单需要加载余额和支付密码状态
      if (!isRecharge) {
        const balanceRes = await apiGet<{ data: { balance: number } }>('/api/user/balance')
        if (balanceRes.success && balanceRes.data) {
          setUserBalance(balanceRes.data.balance || 0)
        }
        
        const payPwdRes = await apiGet<{ data: { is_set: boolean } }>('/api/user/pay-password/status')
        if (payPwdRes.success && payPwdRes.data) {
          setPayPasswordSet(payPwdRes.data.is_set)
        }
      }

      setLoading(false)
    }

    loadData()
  }, [orderNo, rechargeNo, router, isRecharge])

  // 倒计时定时器
  useEffect(() => {
    if (countdown <= 0) return
    
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(timer)
          toast.error('订单已过期，请重新下单')
          router.push(isRecharge ? '/user' : '/products')
          return 0
        }
        return prev - 1
      })
    }, 1000)
    
    return () => clearInterval(timer)
  }, [countdown, router, isRecharge])
  
  // 格式化倒计时显示
  const formatCountdown = useCallback(() => {
    const minutes = Math.floor(countdown / 60)
    const seconds = countdown % 60
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }, [countdown])
  
  // 验证优惠券（仅商品订单）
  const handleValidateCoupon = async () => {
    if (!couponCode.trim()) {
      toast.error('请输入优惠券码')
      return
    }
    if (!order) return
    
    setCouponValidating(true)
    const res = await apiPost<{ coupon_id: number; coupon_name: string; coupon_type: string; discount: number; final_amount: number }>('/api/coupon/validate', {
      code: couponCode.trim(),
      product_id: order.product_id,
      category_id: order.category_id,
      amount: order.price,
    })
    
    if (res.success) {
      setAppliedCoupon({
        coupon_id: res.coupon_id,
        coupon_name: res.coupon_name,
        coupon_type: res.coupon_type,
        discount: res.discount,
        final_amount: res.final_amount,
      })
      setFinalPrice(res.final_amount)
      toast.success(`优惠券已应用，优惠 ${formatMoney(res.discount)}`)
    } else {
      toast.error(res.error || '优惠券无效')
    }
    setCouponValidating(false)
  }
  
  // 移除优惠券
  const handleRemoveCoupon = () => {
    setAppliedCoupon(null)
    setCouponCode('')
    if (order) {
      setFinalPrice(order.price)
    }
    toast.success('已移除优惠券')
  }

  // 处理支付
  const handlePay = async () => {
    if (!selectedMethod) {
      toast.error('请选择支付方式')
      return
    }

    // 充值订单必须有充值订单数据，商品订单必须有订单数据
    if (isRecharge && !rechargeOrder) {
      toast.error('充值订单数据异常')
      return
    }
    if (!isRecharge && !order) {
      toast.error('订单数据异常')
      return
    }

    setPaying(true)

    try {
      if (isRecharge) {
        // 充值订单支付
        await handleRechargePayment()
      } else {
        // 商品订单支付
        switch (selectedMethod) {
          case 'balance':
            await handleBalance()
            break
          case 'paypal':
            await handlePayPal()
            break
          case 'stripe':
            await handleStripe()
            break
          case 'alipay_f2f':
            await handleAlipay()
            break
          case 'wechat_pay':
            await handleWechatPay()
            break
          case 'yi_pay':
            await handleYiPay()
            break
          case 'usdt':
            await handleUSDT()
            break
          default:
            toast.error('不支持的支付方式')
        }
      }
    } catch {
      toast.error('支付失败，请重试')
    } finally {
      setPaying(false)
    }
  }

  // 充值订单支付处理
  const handleRechargePayment = async () => {
    if (!rechargeOrder) return

    switch (selectedMethod) {
      case 'yi_pay':
        await handleRechargeYiPay()
        break
      case 'paypal':
        await handleRechargePayPal()
        break
      case 'stripe':
        await handleRechargeStripe()
        break
      case 'alipay_f2f':
        await handleRechargeAlipay()
        break
      case 'wechat_pay':
        await handleRechargeWechatPay()
        break
      case 'usdt':
        await handleRechargeUSDT()
        break
      default:
        toast.error('不支持的支付方式')
    }
  }

  // 充值订单易支付
  const handleRechargeYiPay = async () => {
    const res = await apiPost<{ pay_url: string }>('/api/yipay/recharge/create', {
      recharge_no: rechargeOrder?.recharge_no,
    })

    if (res.success && res.pay_url) {
      sessionStorage.setItem('pending_recharge_no', rechargeOrder?.recharge_no || '')
      window.location.href = res.pay_url
    } else {
      toast.error(res.error || '创建易支付订单失败')
    }
  }

  // 充值订单 PayPal 支付
  const handleRechargePayPal = async () => {
    const res = await apiPost<{ paypal_order_id: string; approve_url: string }>('/api/paypal/recharge/create', {
      recharge_no: rechargeOrder?.recharge_no,
    })

    if (res.success && res.approve_url) {
      sessionStorage.setItem('pending_recharge_no', rechargeOrder?.recharge_no || '')
      window.location.href = res.approve_url
    } else {
      toast.error(res.error || '创建 PayPal 订单失败')
    }
  }

  // 充值订单 Stripe 支付
  const handleRechargeStripe = async () => {
    const res = await apiPost<{ session_id: string; url: string }>('/api/stripe/recharge/create', {
      recharge_no: rechargeOrder?.recharge_no,
    })

    if (res.success && res.url) {
      sessionStorage.setItem('pending_recharge_no', rechargeOrder?.recharge_no || '')
      window.location.href = res.url
    } else {
      toast.error(res.error || '创建 Stripe 订单失败')
    }
  }

  // 充值订单支付宝支付
  const handleRechargeAlipay = async () => {
    const res = await apiPost<{ qr_code: string }>('/api/alipay/recharge/create', {
      recharge_no: rechargeOrder?.recharge_no,
    })

    if (res.success && res.qr_code) {
      router.push(`/payment/qrcode?type=recharge&recharge_no=${rechargeOrder?.recharge_no}&pay_type=alipay&qr=${encodeURIComponent(res.qr_code)}`)
    } else {
      toast.error(res.error || '创建支付宝订单失败')
    }
  }

  // 充值订单微信支付
  const handleRechargeWechatPay = async () => {
    const res = await apiPost<{ qr_code: string }>('/api/wechat/recharge/create', {
      recharge_no: rechargeOrder?.recharge_no,
    })

    if (res.success && res.qr_code) {
      router.push(`/payment/qrcode?type=recharge&recharge_no=${rechargeOrder?.recharge_no}&pay_type=wechat&qr=${encodeURIComponent(res.qr_code)}`)
    } else {
      toast.error(res.error || '创建微信支付订单失败')
    }
  }

  // 充值订单 USDT 支付
  const handleRechargeUSDT = async () => {
    const res = await apiPost<{ payment_id: string; wallet_address: string; amount_usdt: number; network: string }>('/api/usdt/recharge/create', {
      recharge_no: rechargeOrder?.recharge_no,
    })

    if (res.success && res.wallet_address) {
      router.push(`/payment/qrcode?type=recharge&recharge_no=${rechargeOrder?.recharge_no}&pay_type=usdt&payment_id=${res.payment_id}&address=${encodeURIComponent(res.wallet_address)}&amount=${res.amount_usdt}&network=${res.network}`)
    } else {
      toast.error(res.error || '创建 USDT 订单失败')
    }
  }

  // PayPal 支付
  const handlePayPal = async () => {
    const res = await apiPost<{ paypal_order_id: string; approve_url: string }>('/api/paypal/create', {
      order_no: order?.order_no,
    })

    if (res.success && res.approve_url) {
      sessionStorage.setItem('pending_order_no', order?.order_no || '')
      window.location.href = res.approve_url
    } else {
      toast.error(res.error || '创建 PayPal 订单失败')
    }
  }

  // 支付宝当面付
  const handleAlipay = async () => {
    const res = await apiPost<{ qr_code: string }>('/api/alipay/create', {
      order_no: order?.order_no,
    })

    if (res.success && res.qr_code) {
      router.push(`/payment/qrcode?order_no=${order?.order_no}&type=alipay&qr=${encodeURIComponent(res.qr_code)}`)
    } else {
      toast.error(res.error || '创建支付宝订单失败')
    }
  }

  // 微信支付
  const handleWechatPay = async () => {
    const res = await apiPost<{ qr_code: string }>('/api/wechat/create', {
      order_no: order?.order_no,
    })

    if (res.success && res.qr_code) {
      router.push(`/payment/qrcode?order_no=${order?.order_no}&type=wechat&qr=${encodeURIComponent(res.qr_code)}`)
    } else {
      toast.error(res.error || '创建微信支付订单失败')
    }
  }

  // 易支付
  const handleYiPay = async () => {
    const res = await apiPost<{ pay_url: string }>('/api/yipay/create', {
      order_no: order?.order_no,
    })

    if (res.success && res.pay_url) {
      sessionStorage.setItem('pending_order_no', order?.order_no || '')
      window.location.href = res.pay_url
    } else {
      toast.error(res.error || '创建易支付订单失败')
    }
  }

  // Stripe 支付
  const handleStripe = async () => {
    const res = await apiPost<{ session_id: string; url: string }>('/api/stripe/create', {
      order_no: order?.order_no,
    })

    if (res.success && res.url) {
      sessionStorage.setItem('pending_order_no', order?.order_no || '')
      window.location.href = res.url
    } else {
      toast.error(res.error || '创建 Stripe 订单失败')
    }
  }

  // USDT 支付
  const handleUSDT = async () => {
    const res = await apiPost<{ payment_id: string; wallet_address: string; amount_usdt: number; network: string }>('/api/usdt/create', {
      order_no: order?.order_no,
    })

    if (res.success && res.wallet_address) {
      router.push(`/payment/qrcode?order_no=${order?.order_no}&type=usdt&payment_id=${res.payment_id}&address=${encodeURIComponent(res.wallet_address)}&amount=${res.amount_usdt}&network=${res.network}`)
    } else {
      toast.error(res.error || '创建 USDT 订单失败')
    }
  }

  // 余额支付
  const handleBalance = async () => {
    if (userBalance < finalPrice) {
      toast.error('余额不足，请先充值')
      return
    }
    
    if (!payPasswordSet) {
      toast.error('请先在用户中心设置支付密码')
      return
    }
    
    if (!balancePayPassword || balancePayPassword.length !== 6) {
      setShowPayPasswordInput(true)
      toast.error('请输入6位支付密码')
      return
    }
    
    const res = await apiPost<{ success: boolean }>('/api/order/pay/balance', {
      order_no: order?.order_no,
      pay_password: balancePayPassword,
    })

    if (res.success) {
      toast.success('支付成功！')
      router.push(`/payment/result?order_no=${order?.order_no}&status=success`)
    } else {
      toast.error(res.error || '余额支付失败')
      setBalancePayPassword('')
    }
  }

  // 取消订单
  const handleCancel = async () => {
    if (isRecharge && rechargeOrder) {
      const res = await apiPost(`/api/user/balance/recharge/${rechargeOrder.recharge_no}/cancel`, {})
      if (res.success) {
        toast.success('充值订单已取消')
        router.push('/user')
      } else {
        toast.error(res.error || '取消订单失败')
      }
    } else if (order) {
      const res = await apiPost('/api/order/cancel', { order_no: order.order_no })
      if (res.success) {
        toast.success('订单已取消')
        router.push('/products')
      } else {
        toast.error(res.error || '取消订单失败')
      }
    }
  }

  // 支付方式配置
  const getPaymentOptions = () => {
    const options = []

    // 充值订单不支持余额支付（不能用余额充值余额）
    if (!isRecharge) {
      options.push({
        id: 'balance',
        name: '余额支付',
        icon: '\u{1F4B0}',
        description: payPasswordSet 
          ? `可用余额 ¥${userBalance.toFixed(2)}` 
          : '请先设置支付密码',
        enabled: paymentMethods?.balance?.enabled && userBalance >= finalPrice && payPasswordSet,
      })
    }

    // 所有其他支付方式都支持充值订单
    options.push(
      {
        id: 'yi_pay',
        name: '易支付',
        icon: '\u{1F517}',
        description: '聚合支付',
        enabled: paymentMethods?.yi_pay?.enabled,
      },
      {
        id: 'paypal',
        name: 'PayPal',
        icon: '\u{1F4B3}',
        description: '全球安全支付',
        enabled: paymentMethods?.paypal?.enabled,
      },
      {
        id: 'stripe',
        name: 'Stripe',
        icon: '\u{1F48E}',
        description: '信用卡/借记卡',
        enabled: paymentMethods?.stripe?.enabled,
      },
      {
        id: 'alipay_f2f',
        name: '支付宝',
        icon: '\u{1F4F1}',
        description: '扫码支付',
        enabled: paymentMethods?.alipay_f2f?.enabled,
      },
      {
        id: 'wechat_pay',
        name: '微信支付',
        icon: '\u{1F4AC}',
        description: '扫码支付',
        enabled: paymentMethods?.wechat_pay?.enabled,
      },
      {
        id: 'usdt',
        name: 'USDT',
        icon: '\u{1FA99}',
        description: `虚拟货币 (${paymentMethods?.usdt?.network || 'TRC20'})`,
        enabled: paymentMethods?.usdt?.enabled,
      }
    )

    return options
  }

  const paymentOptions = getPaymentOptions()
  const enabledMethods = paymentOptions.filter((m) => m.enabled)

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-dark-900">
        <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
      </div>
    )
  }

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />

      <main className="flex-1 py-8 px-4">
        <div className="max-w-2xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="space-y-6"
          >
            {/* 订单信息 */}
            <Card 
              title={isRecharge ? "充值订单信息" : "订单信息"} 
              icon={<i className={isRecharge ? "fas fa-wallet" : "fas fa-cart-shopping"} />}
            >
              {isRecharge && rechargeOrder ? (
                // 充值订单信息
                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">充值单号</span>
                    <span className="text-dark-100 font-mono text-sm">{rechargeOrder.recharge_no}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">充值金额</span>
                    <span className="text-dark-100">¥{rechargeOrder.amount.toFixed(2)}</span>
                  </div>
                  {rechargeOrder.bonus_amount > 0 && (
                    <div className="flex justify-between items-center text-green-400">
                      <span>赠送金额</span>
                      <span>+¥{rechargeOrder.bonus_amount.toFixed(2)}</span>
                    </div>
                  )}
                  {rechargeOrder.promo_name && (
                    <div className="flex justify-between items-center text-amber-400">
                      <span>优惠活动</span>
                      <span>{rechargeOrder.promo_name}</span>
                    </div>
                  )}
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">创建时间</span>
                    <span className="text-dark-100">{formatDateTime(rechargeOrder.created_at)}</span>
                  </div>
                  <div className="border-t border-dark-700 pt-3">
                    <div className="flex justify-between items-center mb-2">
                      <span className="text-dark-400">实际支付</span>
                      <span className="text-primary-400 text-xl font-bold">
                        ¥{(rechargeOrder.pay_amount > 0 ? rechargeOrder.pay_amount : rechargeOrder.amount).toFixed(2)}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-dark-400 text-lg">实际到账</span>
                      <span className="text-green-400 text-2xl font-bold">
                        ¥{(rechargeOrder.total_credit > 0 ? rechargeOrder.total_credit : rechargeOrder.amount).toFixed(2)}
                      </span>
                    </div>
                  </div>
                </div>
              ) : order ? (
                // 商品订单信息
                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">订单号</span>
                    <span className="text-dark-100 font-mono text-sm">{order.order_no}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">商品名称</span>
                    <span className="text-dark-100">{order.product_name}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">购买数量</span>
                    <span className="text-dark-100">{order.quantity || 1}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">有效期</span>
                    <span className="text-dark-100">
                      {order.duration}
                      {order.duration_unit}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">创建时间</span>
                    <span className="text-dark-100">{formatDateTime(order.created_at)}</span>
                  </div>
                  {appliedCoupon && (
                    <div className="flex justify-between items-center text-emerald-400">
                      <span>优惠券 ({appliedCoupon.coupon_name})</span>
                      <span>-{formatMoney(appliedCoupon.discount)}</span>
                    </div>
                  )}
                  <div className="border-t border-dark-700 pt-3 flex justify-between items-center">
                    <span className="text-dark-400 text-lg">应付金额</span>
                    <div className="text-right">
                      {appliedCoupon && (
                        <span className="text-dark-500 line-through text-sm mr-2">
                          {formatMoney(order.price)}
                        </span>
                      )}
                      <span className="text-primary-400 text-2xl font-bold">
                        {formatMoney(finalPrice)}
                      </span>
                    </div>
                  </div>
                </div>
              ) : null}
            </Card>

            {/* 优惠券输入（仅商品订单显示） */}
            {!isRecharge && order && (
              <Card title="优惠券" icon={<i className="fas fa-ticket-alt" />}>
                {appliedCoupon ? (
                  <div className="flex items-center justify-between bg-emerald-500/10 border border-emerald-500/30 rounded-xl p-4">
                    <div className="flex items-center gap-3">
                      <i className="fas fa-check-circle text-emerald-400 text-xl" />
                      <div>
                        <p className="text-emerald-300 font-medium">{appliedCoupon.coupon_name}</p>
                        <p className="text-emerald-400/70 text-sm">已优惠 {formatMoney(appliedCoupon.discount)}</p>
                      </div>
                    </div>
                    <Button size="sm" variant="ghost" onClick={handleRemoveCoupon}>
                      <i className="fas fa-times" />
                    </Button>
                  </div>
                ) : (
                  <div className="flex gap-3">
                    <div className="flex-1">
                      <Input
                        value={couponCode}
                        onChange={(e) => setCouponCode(e.target.value)}
                        placeholder="输入优惠券码"
                        onKeyDown={(e) => e.key === 'Enter' && handleValidateCoupon()}
                      />
                    </div>
                    <Button
                      variant="secondary"
                      onClick={handleValidateCoupon}
                      loading={couponValidating}
                      disabled={!couponCode.trim()}
                    >
                      验证
                    </Button>
                  </div>
                )}
              </Card>
            )}

            {/* 支付方式选择 */}
            <Card title="选择支付方式" icon={<i className="fas fa-credit-card" />}>
              {enabledMethods.length === 0 ? (
                <div className="text-center py-8">
                  <div className="text-5xl mb-4">{'\u{1F622}'}</div>
                  <p className="text-dark-400">暂无可用的支付方式</p>
                  <p className="text-dark-500 text-sm mt-2">请联系管理员配置支付方式</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  {enabledMethods.map((method) => (
                    <button
                      key={method.id}
                      onClick={() => {
                        setSelectedMethod(method.id)
                        if (method.id === 'balance') {
                          setShowPayPasswordInput(true)
                        } else {
                          setShowPayPasswordInput(false)
                          setBalancePayPassword('')
                        }
                      }}
                      className={`p-4 rounded-xl border-2 transition-all text-left ${
                        selectedMethod === method.id
                          ? 'border-primary-500 bg-primary-500/10'
                          : 'border-dark-600 bg-dark-700/30 hover:border-dark-500'
                      }`}
                    >
                      <div className="flex items-center gap-3">
                        <span className="text-3xl">{method.icon}</span>
                        <div>
                          <h4 className="text-dark-100 font-medium">{method.name}</h4>
                          <p className="text-dark-500 text-sm">{method.description}</p>
                        </div>
                        {selectedMethod === method.id && (
                          <i className="fas fa-check-circle text-primary-400 ml-auto" />
                        )}
                      </div>
                    </button>
                  ))}
                </div>
              )}
              
              {/* 余额支付密码输入（仅商品订单） */}
              {!isRecharge && selectedMethod === 'balance' && showPayPasswordInput && (
                <div className="mt-4 p-4 bg-dark-700/30 rounded-xl border border-dark-600">
                  <label className="block text-sm text-dark-300 mb-2">
                    <i className="fas fa-lock mr-2" />
                    请输入支付密码
                  </label>
                  <Input
                    type="password"
                    placeholder="请输入6位支付密码"
                    maxLength={6}
                    value={balancePayPassword}
                    onChange={(e) => setBalancePayPassword(e.target.value.replace(/\D/g, ''))}
                    className="text-center text-lg tracking-widest"
                  />
                </div>
              )}
            </Card>

            {/* 操作按钮 */}
            <div className="flex flex-col sm:flex-row gap-4">
              <Button
                variant="secondary"
                className="flex-1"
                onClick={handleCancel}
                disabled={paying}
              >
                取消订单
              </Button>
              <Button
                variant="primary"
                className="flex-1"
                onClick={handlePay}
                loading={paying}
                disabled={!selectedMethod || enabledMethods.length === 0}
              >
                <i className="fas fa-lock mr-2" />
                立即支付
              </Button>
            </div>

            {/* 倒计时提示 */}
            <div className={`text-center text-sm flex items-center justify-center gap-2 ${
              countdown < 300 ? 'text-red-400' : 'text-dark-500'
            }`}>
              <i className="fas fa-clock" />
              <span>支付剩余时间：</span>
              <span className="font-mono font-bold text-lg">{formatCountdown()}</span>
            </div>
          </motion.div>
        </div>
      </main>

      <Footer />
    </div>
  )
}

/**
 * 支付页面
 * 用户创建订单后跳转到此页面选择支付方式并完成支付
 * 支持商品订单和充值订单两种类型
 */
export default function PaymentPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center bg-dark-900">
          <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
        </div>
      }
    >
      <PaymentPageContent />
    </Suspense>
  )
}
