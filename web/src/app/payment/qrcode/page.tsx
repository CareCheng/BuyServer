'use client'

import { useState, useEffect, Suspense, useCallback } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Navbar, Footer } from '@/components/layout'
import { Button, Card } from '@/components/ui'
import { apiGet } from '@/lib/api'

/**
 * 二维码支付页面内容组件
 * 用于支付宝和微信扫码支付
 */
function QRCodePaymentContent() {
  const searchParams = useSearchParams()
  const router = useRouter()

  const orderNo = searchParams.get('order_no')
  const payType = searchParams.get('type') // alipay 或 wechat
  const qrCode = searchParams.get('qr')

  const [checking, setChecking] = useState(false)
  const [countdown, setCountdown] = useState(300) // 5分钟倒计时

  // 检查支付状态
  const checkPaymentStatus = useCallback(async () => {
    if (!orderNo) return

    setChecking(true)
    const endpoint = payType === 'wechat' 
      ? `/api/wechat/status/${orderNo}`
      : `/api/alipay/status/${orderNo}`

    const res = await apiGet<{ paid: boolean; order: { kami_code: string } }>(endpoint)
    setChecking(false)

    if (res.success && res.paid) {
      toast.success('支付成功！')
      // 保存订单信息到 sessionStorage
      sessionStorage.setItem('payment_result', JSON.stringify({
        order_no: orderNo,
        kami_code: res.order?.kami_code || '',
      }))
      router.push('/payment/result?from=qrcode')
    }
  }, [orderNo, payType, router])

  // 轮询检查支付状态
  useEffect(() => {
    if (!orderNo || !qrCode) {
      toast.error('参数错误')
      router.push('/products')
      return
    }

    // 每3秒检查一次支付状态
    const pollInterval = setInterval(checkPaymentStatus, 3000)

    // 倒计时
    const countdownInterval = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(pollInterval)
          clearInterval(countdownInterval)
          toast.error('支付超时，请重新下单')
          router.push('/user')
          return 0
        }
        return prev - 1
      })
    }, 1000)

    return () => {
      clearInterval(pollInterval)
      clearInterval(countdownInterval)
    }
  }, [orderNo, qrCode, router, checkPaymentStatus])

  // 格式化倒计时
  const formatCountdown = () => {
    const minutes = Math.floor(countdown / 60)
    const seconds = countdown % 60
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }

  // 获取支付方式信息
  const getPaymentInfo = () => {
    if (payType === 'wechat') {
      return {
        name: '微信支付',
        icon: '💬',
        color: 'text-green-400',
        bgColor: 'bg-green-500/10',
        app: '微信',
      }
    }
    return {
      name: '支付宝',
      icon: '📱',
      color: 'text-blue-400',
      bgColor: 'bg-blue-500/10',
      app: '支付宝',
    }
  }

  const paymentInfo = getPaymentInfo()

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />

      <main className="flex-1 py-8 px-4">
        <div className="max-w-lg mx-auto">
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
            <Card>
              <div className="text-center py-6">
                {/* 支付方式标识 */}
                <div className={`w-16 h-16 mx-auto mb-4 rounded-full ${paymentInfo.bgColor} flex items-center justify-center`}>
                  <span className="text-4xl">{paymentInfo.icon}</span>
                </div>
                <h2 className={`text-xl font-bold ${paymentInfo.color} mb-2`}>
                  {paymentInfo.name}
                </h2>
                <p className="text-dark-400 text-sm mb-6">
                  请使用{paymentInfo.app}扫描下方二维码完成支付
                </p>

                {/* 二维码区域 */}
                <div className="bg-white p-4 rounded-xl inline-block mb-6">
                  {/* 这里应该使用二维码组件生成二维码 */}
                  <div className="w-48 h-48 bg-gray-100 flex items-center justify-center">
                    {qrCode ? (
                      <img
                        src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(qrCode)}`}
                        alt="支付二维码"
                        className="w-full h-full"
                      />
                    ) : (
                      <span className="text-gray-400">加载中...</span>
                    )}
                  </div>
                </div>

                {/* 订单信息 */}
                <div className="bg-dark-700/30 rounded-xl p-4 mb-6 text-left">
                  <div className="flex justify-between items-center mb-2">
                    <span className="text-dark-400">订单号</span>
                    <span className="text-dark-100 font-mono text-sm">{orderNo}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">剩余时间</span>
                    <span className={`font-mono text-lg ${countdown < 60 ? 'text-red-400' : 'text-primary-400'}`}>
                      {formatCountdown()}
                    </span>
                  </div>
                </div>

                {/* 状态提示 */}
                <div className="flex items-center justify-center gap-2 text-dark-400 mb-6">
                  {checking ? (
                    <>
                      <i className="fas fa-spinner fa-spin" />
                      <span>正在检查支付状态...</span>
                    </>
                  ) : (
                    <>
                      <i className="fas fa-clock" />
                      <span>等待支付中，支付完成后自动跳转</span>
                    </>
                  )}
                </div>

                {/* 操作按钮 */}
                <div className="flex flex-col sm:flex-row gap-4">
                  <Button
                    variant="secondary"
                    className="flex-1"
                    onClick={() => router.push('/user')}
                  >
                    取消支付
                  </Button>
                  <Button
                    variant="primary"
                    className="flex-1"
                    onClick={checkPaymentStatus}
                    loading={checking}
                  >
                    <i className="fas fa-sync-alt mr-2" />
                    我已支付
                  </Button>
                </div>
              </div>
            </Card>

            {/* 帮助提示 */}
            <div className="mt-6 text-center text-dark-500 text-sm space-y-2">
              <p>
                <i className="fas fa-info-circle mr-1" />
                如果二维码无法显示，请刷新页面重试
              </p>
              <p>
                <i className="fas fa-question-circle mr-1" />
                支付遇到问题？<a href="/message" className="text-primary-400 hover:underline">联系客服</a>
              </p>
            </div>
          </motion.div>
        </div>
      </main>

      <Footer />
    </div>
  )
}

/**
 * 二维码支付页面
 */
export default function QRCodePaymentPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center bg-dark-900">
          <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
        </div>
      }
    >
      <QRCodePaymentContent />
    </Suspense>
  )
}
