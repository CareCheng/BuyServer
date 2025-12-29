'use client'

import { useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import { Navbar, Footer } from '@/components/layout'
import { Button, Card } from '@/components/ui'

/**
 * 支付取消页面
 * 用户取消支付后跳转到此页面
 */
export default function PaymentCancelPage() {
  const router = useRouter()

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />

      <main className="flex-1 py-8 px-4">
        <div className="max-w-lg mx-auto">
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
            <Card>
              <div className="text-center py-8">
                <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-yellow-500/20 flex items-center justify-center">
                  <i className="fas fa-exclamation-triangle text-4xl text-yellow-400" />
                </div>
                <h2 className="text-2xl font-bold text-dark-100 mb-2">支付已取消</h2>
                <p className="text-dark-400 mb-6">您已取消本次支付，订单仍保留在您的账户中</p>

                {/* 提示信息 */}
                <div className="bg-dark-700/30 rounded-xl p-4 mb-6 text-left">
                  <div className="flex items-start gap-3">
                    <i className="fas fa-clock text-dark-400 mt-0.5" />
                    <div className="text-sm">
                      <p className="text-dark-300">订单将在 30 分钟后自动取消</p>
                      <p className="text-dark-500 mt-1">
                        如需继续支付，请前往用户中心或重新下单
                      </p>
                    </div>
                  </div>
                </div>

                {/* 操作按钮 */}
                <div className="flex flex-col sm:flex-row gap-4">
                  <Button variant="secondary" className="flex-1" onClick={() => router.push('/user')}>
                    <i className="fas fa-list mr-2" />
                    查看订单
                  </Button>
                  <Button variant="primary" className="flex-1" onClick={() => router.push('/products')}>
                    <i className="fas fa-cart-shopping mr-2" />
                    继续购物
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
