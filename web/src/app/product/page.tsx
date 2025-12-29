'use client'

import { useState, useEffect, Suspense } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Navbar, Footer } from '@/components/layout'
import { Button, Modal } from '@/components/ui'
import { apiGet, apiPost, apiDelete } from '@/lib/api'
import { formatMoney, copyToClipboard } from '@/lib/utils'

/**
 * 商品接口
 */
interface Product {
  id: number
  name: string
  description: string
  detail: string           // 详细介绍（Markdown）
  specs: string            // 规格参数（JSON）
  features: string         // 特性列表（JSON）
  tags: string             // 标签（逗号分隔）
  price: number
  duration: number
  duration_unit: string
  stock: number
  image_url: string
  category_id: number
  category_name?: string
}

/**
 * 规格参数项接口
 */
interface SpecItem {
  key: string
  value: string
}

/**
 * 商品图片接口
 */
interface ProductImage {
  id: number
  product_id: number
  image_url: string
  sort_order: number
  is_primary: boolean
}

/**
 * 商品评价接口
 */
interface ProductReview {
  id: number
  product_id: number
  user_id: number
  username: string
  rating: number
  content: string
  images: string
  admin_reply: string
  created_at: string
}

/**
 * 评价统计接口
 */
interface ReviewStats {
  total: number
  average_rating: number
  rating_distribution: { [key: string]: number }
}

/**
 * 商品详情内容组件
 */
function ProductDetailContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const productId = searchParams.get('id')

  const [product, setProduct] = useState<Product | null>(null)
  const [images, setImages] = useState<ProductImage[]>([])
  const [reviews, setReviews] = useState<ProductReview[]>([])
  const [reviewStats, setReviewStats] = useState<ReviewStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [currentImageIndex, setCurrentImageIndex] = useState(0)
  const [showPurchaseModal, setShowPurchaseModal] = useState(false)
  const [showResultModal, setShowResultModal] = useState(false)
  const [purchaseResult, setPurchaseResult] = useState<{ order_no: string; kami_code: string } | null>(null)
  const [purchasing, setPurchasing] = useState(false)
  const [quantity, setQuantity] = useState(1)
  const [isFavorite, setIsFavorite] = useState(false)
  const [favoriteLoading, setFavoriteLoading] = useState(false)

  // 检查收藏状态
  const checkFavorite = async (productId: string) => {
    const res = await apiGet<{ is_favorite: boolean }>(`/api/user/favorite/${productId}/check`)
    if (res.success) {
      setIsFavorite(res.is_favorite || false)
    }
  }

  // 切换收藏状态
  const toggleFavorite = async () => {
    if (!product) return
    setFavoriteLoading(true)
    
    if (isFavorite) {
      // 取消收藏
      const res = await apiDelete(`/api/user/favorite/${product.id}`)
      if (res.success) {
        setIsFavorite(false)
        toast.success('已取消收藏')
      } else {
        if (res.error === '请先登录') {
          window.location.href = '/login/'
        } else {
          toast.error(res.error || '操作失败')
        }
      }
    } else {
      // 添加收藏
      const res = await apiPost('/api/user/favorite', { product_id: product.id })
      if (res.success) {
        setIsFavorite(true)
        toast.success('已添加到收藏')
      } else {
        if (res.error === '请先登录') {
          window.location.href = '/login/'
        } else {
          toast.error(res.error || '操作失败')
        }
      }
    }
    setFavoriteLoading(false)
  }

  // 加载商品详情
  useEffect(() => {
    const loadProduct = async () => {
      if (!productId) {
        router.push('/products')
        return
      }

      setLoading(true)
      try {
        // 先加载商品基本信息（必须成功）
        const productRes = await apiGet<{ product: Product }>(`/api/product/${productId}`)
        
        if (!productRes.success || !productRes.product) {
          toast.error('商品不存在')
          router.push('/products')
          return
        }
        
        setProduct(productRes.product)

        // 并行加载其他信息（允许部分失败）
        const [imagesRes, reviewsRes, statsRes] = await Promise.all([
          apiGet<{ data: ProductImage[] }>(`/api/product/${productId}/images`).catch(() => ({ success: false, data: [] as ProductImage[] })),
          apiGet<{ data: { reviews: ProductReview[]; total: number } }>(`/api/product/${productId}/reviews?page=1&page_size=10`).catch(() => ({ success: false, data: { reviews: [] as ProductReview[], total: 0 } })),
          apiGet<{ data: ReviewStats }>(`/api/product/${productId}/review-stats`).catch(() => ({ success: false, data: null })),
        ])

        if (imagesRes.success && imagesRes.data) {
          setImages(Array.isArray(imagesRes.data) ? imagesRes.data : [])
        }

        if (reviewsRes.success && reviewsRes.data?.reviews) {
          setReviews(reviewsRes.data.reviews)
        }

        if (statsRes.success && statsRes.data) {
          setReviewStats(statsRes.data)
        }

        // 检查收藏状态（允许失败）
        checkFavorite(productId)
      } catch (err) {
        console.error('加载商品信息失败:', err)
        toast.error('加载商品信息失败')
      }
      setLoading(false)
    }

    loadProduct()
  }, [productId, router])

  // 获取所有图片
  const allImages = product ? [
    { id: 0, image_url: product.image_url, is_primary: true },
    ...images.filter(img => img.image_url !== product.image_url)
  ].filter(img => img.image_url) : []

  // 确认购买
  const handlePurchase = async () => {
    if (!product) return
    setPurchasing(true)
    const res = await apiPost<{ order_no: string }>('/api/order/create', {
      product_id: product.id,
      quantity: quantity,
    })
    setPurchasing(false)

    if (res.success && res.order_no) {
      setShowPurchaseModal(false)
      toast.success('订单创建成功，正在跳转支付页面...')
      window.location.href = `/payment?order_no=${res.order_no}`
    } else {
      if (res.error === '请先登录') {
        window.location.href = '/login/'
      } else {
        toast.error(res.error || '创建订单失败')
      }
    }
  }

  // 添加到购物车
  const handleAddToCart = async () => {
    if (!product) return
    const res = await apiPost('/api/user/cart', {
      product_id: product.id,
      quantity: quantity,
    })

    if (res.success) {
      toast.success('已添加到购物车')
    } else {
      if (res.error === '请先登录') {
        window.location.href = '/login/'
      } else {
        toast.error(res.error || '添加失败')
      }
    }
  }

  // 复制卡密
  const handleCopyKami = async () => {
    if (purchaseResult?.kami_code) {
      const success = await copyToClipboard(purchaseResult.kami_code)
      if (success) {
        toast.success('已复制到剪贴板')
      }
    }
  }

  // 渲染星级
  const renderStars = (rating: number) => (
    <div className="flex items-center gap-0.5">
      {[1, 2, 3, 4, 5].map((star) => (
        <i key={star} className={`fas fa-star text-sm ${star <= rating ? 'text-yellow-400' : 'text-dark-600'}`} />
      ))}
    </div>
  )


  if (loading) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 py-8 px-4">
          <div className="max-w-6xl mx-auto">
            <div className="animate-pulse">
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                <div className="h-96 bg-dark-700/50 rounded-xl" />
                <div className="space-y-4">
                  <div className="h-8 bg-dark-700/50 rounded w-3/4" />
                  <div className="h-4 bg-dark-700/50 rounded w-1/2" />
                  <div className="h-20 bg-dark-700/50 rounded" />
                </div>
              </div>
            </div>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  if (!product) return null

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <main className="flex-1 py-8 px-4">
        <div className="max-w-6xl mx-auto">
          {/* 面包屑 */}
          <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="mb-6 flex items-center gap-2 text-sm text-dark-400">
            <a href="/products" className="hover:text-primary-400 transition-colors">商品列表</a>
            <i className="fas fa-chevron-right text-xs" />
            <span className="text-dark-200">{product.name}</span>
          </motion.div>

          {/* 商品信息 */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-12">
            {/* 图片区域 */}
            <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }} className="space-y-4">
              <div className="aspect-square bg-dark-800/50 rounded-2xl overflow-hidden border border-dark-700/50">
                {allImages.length > 0 && allImages[currentImageIndex]?.image_url ? (
                  <img src={allImages[currentImageIndex].image_url} alt={product.name} className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full flex items-center justify-center"><span className="text-8xl">📦</span></div>
                )}
              </div>
              {allImages.length > 1 && (
                <div className="flex gap-2 overflow-x-auto pb-2">
                  {allImages.map((img, index) => (
                    <button key={img.id} onClick={() => setCurrentImageIndex(index)}
                      className={`flex-shrink-0 w-20 h-20 rounded-lg overflow-hidden border-2 transition-colors ${index === currentImageIndex ? 'border-primary-500' : 'border-dark-700/50 hover:border-dark-600'}`}>
                      <img src={img.image_url} alt="" className="w-full h-full object-cover" />
                    </button>
                  ))}
                </div>
              )}
            </motion.div>

            {/* 商品详情 */}
            <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} className="space-y-6">
              <div>
                <h1 className="text-3xl font-bold text-dark-100 mb-2">{product.name}</h1>
                {product.category_name && <span className="inline-block px-3 py-1 bg-primary-500/20 text-primary-400 text-sm rounded-full">{product.category_name}</span>}
              </div>

              {reviewStats && reviewStats.total > 0 && (
                <div className="flex items-center gap-3">
                  {renderStars(Math.round(reviewStats.average_rating))}
                  <span className="text-dark-300">{reviewStats.average_rating.toFixed(1)} 分</span>
                  <span className="text-dark-500">|</span>
                  <span className="text-dark-400">{reviewStats.total} 条评价</span>
                </div>
              )}

              <div className="flex items-baseline gap-2">
                <span className="text-4xl font-bold text-primary-400">{formatMoney(product.price)}</span>
                <span className="text-dark-500">/ {product.duration}{product.duration_unit}</span>
              </div>

              <div className="bg-dark-800/30 rounded-xl p-4">
                <h3 className="text-dark-300 text-sm mb-2">商品描述</h3>
                <p className="text-dark-200 whitespace-pre-wrap">{product.description || '暂无描述'}</p>
              </div>

              {/* 商品标签 */}
              {product.tags && (
                <div className="flex flex-wrap gap-2">
                  {product.tags.split(',').filter(t => t.trim()).map((tag, i) => (
                    <span key={i} className="px-3 py-1 bg-primary-500/20 text-primary-400 text-sm rounded-full">
                      {tag.trim()}
                    </span>
                  ))}
                </div>
              )}

              <div className="flex items-center gap-4">
                <span className="text-dark-400">库存状态：</span>
                {product.stock === -1 ? (
                  <span className="text-emerald-400"><i className="fas fa-check-circle mr-1" />库存充足</span>
                ) : product.stock > 0 ? (
                  <span className="text-amber-400"><i className="fas fa-exclamation-circle mr-1" />剩余 {product.stock} 件</span>
                ) : (
                  <span className="text-red-400"><i className="fas fa-times-circle mr-1" />已售罄</span>
                )}
              </div>

              {product.stock !== 0 && (
                <div className="flex items-center gap-4">
                  <span className="text-dark-400">购买数量：</span>
                  <div className="flex items-center gap-2">
                    <button onClick={() => setQuantity(Math.max(1, quantity - 1))} className="w-10 h-10 rounded-lg bg-dark-700/50 text-dark-300 hover:bg-dark-700 transition-colors flex items-center justify-center">
                      <i className="fas fa-minus" />
                    </button>
                    <input type="number" value={quantity} onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value) || 1))}
                      className="w-16 h-10 text-center bg-dark-800/50 border border-dark-700/50 rounded-lg text-dark-100" />
                    <button onClick={() => setQuantity(quantity + 1)} className="w-10 h-10 rounded-lg bg-dark-700/50 text-dark-300 hover:bg-dark-700 transition-colors flex items-center justify-center">
                      <i className="fas fa-plus" />
                    </button>
                  </div>
                </div>
              )}

              <div className="flex flex-col sm:flex-row gap-3 pt-4">
                <Button variant="primary" size="lg" className="flex-1" onClick={() => setShowPurchaseModal(true)} disabled={product.stock === 0}>
                  <i className="fas fa-shopping-bag mr-2" />立即购买
                </Button>
                <Button variant="secondary" size="lg" className="flex-1" onClick={handleAddToCart} disabled={product.stock === 0}>
                  <i className="fas fa-cart-plus mr-2" />加入购物车
                </Button>
                <Button 
                  variant={isFavorite ? 'danger' : 'ghost'} 
                  size="lg" 
                  onClick={toggleFavorite}
                  loading={favoriteLoading}
                  title={isFavorite ? '取消收藏' : '添加收藏'}
                >
                  <i className={`fas fa-heart ${isFavorite ? '' : 'text-dark-400'}`} />
                </Button>
              </div>
            </motion.div>
          </div>

          {/* 商品详情（Markdown） */}
          {product.detail && (
            <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="card p-6">
              <h2 className="text-xl font-bold text-dark-100 mb-6">
                <i className="fas fa-file-alt mr-2 text-primary-400" />商品详情
              </h2>
              <div className="prose prose-invert prose-sm max-w-none">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {product.detail}
                </ReactMarkdown>
              </div>
            </motion.div>
          )}

          {/* 规格参数 */}
          {product.specs && (() => {
            try {
              const specs: SpecItem[] = JSON.parse(product.specs)
              if (specs.length > 0) {
                return (
                  <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.15 }} className="card p-6">
                    <h2 className="text-xl font-bold text-dark-100 mb-6">
                      <i className="fas fa-list-ul mr-2 text-primary-400" />规格参数
                    </h2>
                    <div className="overflow-x-auto">
                      <table className="w-full">
                        <tbody>
                          {specs.map((spec, index) => (
                            <tr key={index} className={index % 2 === 0 ? 'bg-dark-700/30' : ''}>
                              <td className="py-3 px-4 text-dark-400 font-medium w-1/3">{spec.key}</td>
                              <td className="py-3 px-4 text-dark-200">{spec.value}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </motion.div>
                )
              }
            } catch { /* 解析失败忽略 */ }
            return null
          })()}

          {/* 特性/卖点 */}
          {product.features && (() => {
            try {
              const features: string[] = JSON.parse(product.features)
              if (features.length > 0) {
                return (
                  <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.18 }} className="card p-6">
                    <h2 className="text-xl font-bold text-dark-100 mb-6">
                      <i className="fas fa-star mr-2 text-primary-400" />产品特性
                    </h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {features.map((feature, index) => (
                        <div key={index} className="flex items-start gap-3 p-3 bg-dark-700/30 rounded-lg">
                          <i className="fas fa-check-circle text-emerald-400 mt-0.5" />
                          <span className="text-dark-200">{feature}</span>
                        </div>
                      ))}
                    </div>
                  </motion.div>
                )
              }
            } catch { /* 解析失败忽略 */ }
            return null
          })()}

          {/* 商品评价 */}
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="card p-6">
            <h2 className="text-xl font-bold text-dark-100 mb-6">
              <i className="fas fa-comments mr-2 text-primary-400" />商品评价
              {reviewStats && reviewStats.total > 0 && <span className="text-dark-400 text-base font-normal ml-2">({reviewStats.total} 条)</span>}
            </h2>

            {reviews.length === 0 ? (
              <div className="text-center py-12">
                <div className="text-5xl mb-4">💬</div>
                <p className="text-dark-400">暂无评价，购买后可以发表评价</p>
              </div>
            ) : (
              <div className="space-y-6">
                {reviews.map((review) => (
                  <div key={review.id} className="border-b border-dark-700/50 pb-6 last:border-0">
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-primary-500/20 flex items-center justify-center">
                          <span className="text-primary-400 font-medium">{review.username.charAt(0).toUpperCase()}</span>
                        </div>
                        <div>
                          <p className="text-dark-200 font-medium">{review.username}</p>
                          <p className="text-dark-500 text-sm">{new Date(review.created_at).toLocaleDateString()}</p>
                        </div>
                      </div>
                      {renderStars(review.rating)}
                    </div>
                    <p className="text-dark-300 mb-3">{review.content}</p>
                    {review.images && (
                      <div className="flex gap-2 flex-wrap mb-3">
                        {review.images.split(',').filter(Boolean).map((img, idx) => (
                          <img key={idx} src={img} alt="" className="w-20 h-20 object-cover rounded-lg" />
                        ))}
                      </div>
                    )}
                    {review.admin_reply && (
                      <div className="bg-dark-700/30 rounded-lg p-3 mt-3">
                        <p className="text-dark-400 text-sm mb-1"><i className="fas fa-reply mr-1" />商家回复</p>
                        <p className="text-dark-300 text-sm">{review.admin_reply}</p>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </motion.div>
        </div>
      </main>
      <Footer />

      {/* 购买确认弹窗 */}
      <Modal isOpen={showPurchaseModal} onClose={() => setShowPurchaseModal(false)} title="确认购买" size="sm">
        <div className="space-y-4">
          <div className="bg-dark-700/30 rounded-xl p-4 space-y-2">
            <div className="flex justify-between"><span className="text-dark-400">商品名称</span><span className="text-dark-100">{product.name}</span></div>
            <div className="flex justify-between"><span className="text-dark-400">有效期</span><span className="text-dark-100">{product.duration}{product.duration_unit}</span></div>
            <div className="flex justify-between"><span className="text-dark-400">数量</span><span className="text-dark-100">{quantity}</span></div>
            <div className="flex justify-between border-t border-dark-600/50 pt-2 mt-2">
              <span className="text-dark-400">总价</span>
              <span className="text-primary-400 font-bold text-lg">{formatMoney(product.price * quantity)}</span>
            </div>
          </div>
          <div className="flex gap-3">
            <Button variant="secondary" className="flex-1" onClick={() => setShowPurchaseModal(false)}>取消</Button>
            <Button variant="primary" className="flex-1" onClick={handlePurchase} loading={purchasing}>确认购买</Button>
          </div>
        </div>
      </Modal>

      {/* 购买结果弹窗 */}
      <Modal isOpen={showResultModal} onClose={() => setShowResultModal(false)} title="购买成功" size="sm">
        {purchaseResult && (
          <div className="space-y-4">
            <div className="text-center py-4"><div className="text-5xl mb-4">🎉</div><p className="text-dark-300">恭喜您，购买成功！</p></div>
            <div className="bg-dark-700/30 rounded-xl p-4 space-y-3">
              <div><span className="text-dark-400 text-sm">订单号</span><p className="text-dark-100 font-mono">{purchaseResult.order_no}</p></div>
              <div><span className="text-dark-400 text-sm">卡密</span><p className="text-primary-400 font-mono text-lg break-all">{purchaseResult.kami_code}</p></div>
            </div>
            <div className="flex gap-3">
              <Button variant="secondary" className="flex-1" onClick={() => setShowResultModal(false)}>关闭</Button>
              <Button variant="primary" className="flex-1" onClick={handleCopyKami}><i className="fas fa-copy mr-2" />复制卡密</Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}

/**
 * 商品详情页面
 */
export default function ProductDetailPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 py-8 px-4">
          <div className="max-w-6xl mx-auto">
            <div className="animate-pulse">
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                <div className="h-96 bg-dark-700/50 rounded-xl" />
                <div className="space-y-4">
                  <div className="h-8 bg-dark-700/50 rounded w-3/4" />
                  <div className="h-4 bg-dark-700/50 rounded w-1/2" />
                </div>
              </div>
            </div>
          </div>
        </main>
        <Footer />
      </div>
    }>
      <ProductDetailContent />
    </Suspense>
  )
}
