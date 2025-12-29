'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Card, Badge, Modal } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { formatMoney } from '@/lib/utils'

/**
 * 购物车商品接口
 */
interface CartItem {
  id: number
  product_id: number
  product_name: string
  product_price: number
  product_image: string
  product_stock: number
  quantity: number
  created_at: string
}

/**
 * 购物车标签页
 */
export function CartTab() {
  const [items, setItems] = useState<CartItem[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [showCheckoutModal, setShowCheckoutModal] = useState(false)
  const [checkoutLoading, setCheckoutLoading] = useState(false)
  // 清空购物车确认弹窗状态
  const [showClearConfirm, setShowClearConfirm] = useState(false)

  // 加载购物车
  const loadCart = async () => {
    setLoading(true)
    const res = await apiGet<{ items: CartItem[] }>('/api/user/cart')
    if (res.success && res.items) {
      setItems(res.items)
      // 默认全选
      setSelectedIds(res.items.map(item => item.id))
    }
    setLoading(false)
  }

  useEffect(() => {
    loadCart()
  }, [])

  // 更新数量
  const updateQuantity = async (id: number, quantity: number) => {
    if (quantity < 1) return
    const res = await apiPut(`/api/user/cart/${id}`, { quantity })
    if (res.success) {
      setItems(items.map(item => item.id === id ? { ...item, quantity } : item))
    } else {
      toast.error(res.error || '更新失败')
    }
  }

  // 删除商品
  const removeItem = async (id: number) => {
    const res = await apiDelete(`/api/user/cart/${id}`)
    if (res.success) {
      setItems(items.filter(item => item.id !== id))
      setSelectedIds(selectedIds.filter(sid => sid !== id))
      toast.success('已移除')
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  // 确认清空购物车
  const confirmClearCart = async () => {
    const res = await apiDelete('/api/user/cart')
    if (res.success) {
      setItems([])
      setSelectedIds([])
      toast.success('购物车已清空')
    } else {
      toast.error(res.error || '清空失败')
    }
    setShowClearConfirm(false)
  }

  // 切换选中状态
  const toggleSelect = (id: number) => {
    if (selectedIds.includes(id)) {
      setSelectedIds(selectedIds.filter(sid => sid !== id))
    } else {
      setSelectedIds([...selectedIds, id])
    }
  }

  // 全选/取消全选
  const toggleSelectAll = () => {
    if (selectedIds.length === items.length) {
      setSelectedIds([])
    } else {
      setSelectedIds(items.map(item => item.id))
    }
  }

  // 计算选中商品总价
  const selectedTotal = items
    .filter(item => selectedIds.includes(item.id))
    .reduce((sum, item) => sum + item.product_price * item.quantity, 0)

  // 结算
  const handleCheckout = async () => {
    if (selectedIds.length === 0) {
      toast.error('请选择要结算的商品')
      return
    }

    setCheckoutLoading(true)
    // 验证购物车
    const validateRes = await apiPost<{ valid: boolean; errors: string[] }>('/api/user/cart/validate', {
      item_ids: selectedIds,
    })

    if (!validateRes.success || !validateRes.valid) {
      toast.error(validateRes.errors?.[0] || validateRes.error || '部分商品无法购买')
      setCheckoutLoading(false)
      return
    }

    // 创建订单（批量）
    const selectedItems = items.filter(item => selectedIds.includes(item.id))
    
    // 逐个创建订单
    const orderNos: string[] = []
    for (const item of selectedItems) {
      const res = await apiPost<{ order_no: string }>('/api/order/create', {
        product_id: item.product_id,
        quantity: item.quantity,
      })
      if (res.success && res.order_no) {
        orderNos.push(res.order_no)
        // 从购物车移除
        await apiDelete(`/api/user/cart/${item.id}`)
      }
    }

    setCheckoutLoading(false)
    setShowCheckoutModal(false)

    if (orderNos.length > 0) {
      toast.success(`已创建 ${orderNos.length} 个订单`)
      // 跳转到第一个订单的支付页面
      window.location.href = `/payment?order_no=${orderNos[0]}`
    } else {
      toast.error('创建订单失败')
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
      <Card title="我的购物车" icon={<i className="fas fa-cart-shopping" />}>
        {items.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-5xl mb-4">🛒</div>
            <p className="text-dark-400 mb-4">购物车是空的</p>
            <a href="/products" className="text-primary-400 hover:text-primary-300">
              去逛逛 <i className="fas fa-arrow-right ml-1" />
            </a>
          </div>
        ) : (
          <>
            {/* 操作栏 */}
            <div className="flex items-center justify-between mb-4 pb-4 border-b border-dark-700/50">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedIds.length === items.length}
                  onChange={toggleSelectAll}
                  className="w-4 h-4 rounded border-dark-600 bg-dark-700 text-primary-500 focus:ring-primary-500"
                />
                <span className="text-dark-300">全选</span>
              </label>
              <Button size="sm" variant="ghost" onClick={() => setShowClearConfirm(true)}>
                <i className="fas fa-trash mr-1" />清空购物车
              </Button>
            </div>

            {/* 商品列表 */}
            <div className="space-y-4">
              {items.map((item) => (
                <div
                  key={item.id}
                  className={`flex items-center gap-4 p-4 rounded-xl border transition-colors ${
                    selectedIds.includes(item.id)
                      ? 'bg-primary-500/5 border-primary-500/30'
                      : 'bg-dark-700/30 border-dark-600/50'
                  }`}
                >
                  {/* 选择框 */}
                  <input
                    type="checkbox"
                    checked={selectedIds.includes(item.id)}
                    onChange={() => toggleSelect(item.id)}
                    className="w-4 h-4 rounded border-dark-600 bg-dark-700 text-primary-500 focus:ring-primary-500"
                  />

                  {/* 商品图片 */}
                  <div className="w-20 h-20 rounded-lg overflow-hidden bg-dark-700/50 flex-shrink-0">
                    {item.product_image ? (
                      <img src={item.product_image} alt={item.product_name} className="w-full h-full object-cover" />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-3xl">📦</div>
                    )}
                  </div>

                  {/* 商品信息 */}
                  <div className="flex-1 min-w-0">
                    <a
                      href={`/product?id=${item.product_id}`}
                      className="text-dark-100 font-medium hover:text-primary-400 transition-colors line-clamp-1"
                    >
                      {item.product_name}
                    </a>
                    <div className="text-primary-400 font-medium mt-1">
                      {formatMoney(item.product_price)}
                    </div>
                    {item.product_stock === 0 && (
                      <Badge variant="danger" className="mt-1">已售罄</Badge>
                    )}
                  </div>

                  {/* 数量控制 */}
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => updateQuantity(item.id, item.quantity - 1)}
                      disabled={item.quantity <= 1}
                      className="w-8 h-8 rounded-lg bg-dark-700/50 text-dark-300 hover:bg-dark-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    >
                      <i className="fas fa-minus text-xs" />
                    </button>
                    <span className="w-10 text-center text-dark-100">{item.quantity}</span>
                    <button
                      onClick={() => updateQuantity(item.id, item.quantity + 1)}
                      className="w-8 h-8 rounded-lg bg-dark-700/50 text-dark-300 hover:bg-dark-700 transition-colors"
                    >
                      <i className="fas fa-plus text-xs" />
                    </button>
                  </div>

                  {/* 小计 */}
                  <div className="text-right w-24">
                    <div className="text-dark-400 text-xs">小计</div>
                    <div className="text-dark-100 font-medium">
                      {formatMoney(item.product_price * item.quantity)}
                    </div>
                  </div>

                  {/* 删除按钮 */}
                  <button
                    onClick={() => removeItem(item.id)}
                    className="p-2 text-dark-500 hover:text-red-400 transition-colors"
                    title="移除"
                  >
                    <i className="fas fa-times" />
                  </button>
                </div>
              ))}
            </div>

            {/* 结算栏 */}
            <div className="mt-6 pt-4 border-t border-dark-700/50 flex items-center justify-between">
              <div className="text-dark-400">
                已选 <span className="text-primary-400 font-medium">{selectedIds.length}</span> 件商品
              </div>
              <div className="flex items-center gap-4">
                <div className="text-right">
                  <span className="text-dark-400">合计：</span>
                  <span className="text-2xl font-bold text-primary-400 ml-2">
                    {formatMoney(selectedTotal)}
                  </span>
                </div>
                <Button
                  size="lg"
                  onClick={() => setShowCheckoutModal(true)}
                  disabled={selectedIds.length === 0}
                >
                  <i className="fas fa-credit-card mr-2" />
                  结算
                </Button>
              </div>
            </div>
          </>
        )}
      </Card>

      {/* 结算确认弹窗 */}
      <Modal isOpen={showCheckoutModal} onClose={() => setShowCheckoutModal(false)} title="确认结算" size="sm">
        <div className="space-y-4">
          <div className="bg-dark-700/30 rounded-xl p-4">
            <div className="flex justify-between mb-2">
              <span className="text-dark-400">商品数量</span>
              <span className="text-dark-100">{selectedIds.length} 件</span>
            </div>
            <div className="flex justify-between border-t border-dark-600/50 pt-2 mt-2">
              <span className="text-dark-400">应付金额</span>
              <span className="text-primary-400 font-bold text-xl">{formatMoney(selectedTotal)}</span>
            </div>
          </div>
          <p className="text-dark-500 text-sm">
            <i className="fas fa-info-circle mr-1" />
            将为每件商品创建独立订单，您可以分别支付
          </p>
          <div className="flex gap-3">
            <Button variant="secondary" className="flex-1" onClick={() => setShowCheckoutModal(false)}>
              取消
            </Button>
            <Button className="flex-1" onClick={handleCheckout} loading={checkoutLoading}>
              确认结算
            </Button>
          </div>
        </div>
      </Modal>

      {/* 清空购物车确认弹窗 */}
      <ConfirmModal
        isOpen={showClearConfirm}
        onClose={() => setShowClearConfirm(false)}
        title="清空购物车"
        message="确定要清空购物车吗？此操作不可恢复。"
        confirmText="清空"
        variant="danger"
        onConfirm={confirmClearCart}
      />
    </motion.div>
  )
}
