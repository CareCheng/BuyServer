'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Card, Badge } from '@/components/ui'
import { apiGet, apiDelete } from '@/lib/api'

/**
 * 收藏商品接口
 */
interface FavoriteProduct {
  id: number
  product_id: number
  product_name: string
  product_price: number
  product_image: string
  product_status: number
  created_at: string
}

/**
 * 我的收藏标签页
 */
export function FavoritesTab() {
  const [favorites, setFavorites] = useState<FavoriteProduct[]>([])
  const [loading, setLoading] = useState(true)

  // 加载收藏列表
  const loadFavorites = async () => {
    setLoading(true)
    const res = await apiGet<{ favorites: FavoriteProduct[] }>('/api/user/favorites')
    if (res.success && res.favorites) {
      setFavorites(res.favorites)
    }
    setLoading(false)
  }

  useEffect(() => {
    loadFavorites()
  }, [])

  // 取消收藏
  const handleRemoveFavorite = async (productId: number) => {
    const res = await apiDelete(`/api/user/favorite/${productId}`)
    if (res.success) {
      toast.success('已取消收藏')
      setFavorites(favorites.filter(f => f.product_id !== productId))
    } else {
      toast.error(res.error || '操作失败')
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
    <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
      <Card title="我的收藏" icon={<i className="fas fa-heart" />}>
        {favorites.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-5xl mb-4">💔</div>
            <p className="text-dark-400 mb-4">暂无收藏的商品</p>
            <a href="/products/">
              <Button>去逛逛</Button>
            </a>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {favorites.map((item) => (
              <div
                key={item.id}
                className="bg-dark-700/30 rounded-xl border border-dark-600/50 overflow-hidden group"
              >
                {/* 商品图片 */}
                <div className="aspect-video bg-dark-800 relative">
                  {item.product_image ? (
                    <img
                      src={item.product_image}
                      alt={item.product_name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-4xl">
                      📦
                    </div>
                  )}
                  {item.product_status !== 1 && (
                    <div className="absolute inset-0 bg-black/60 flex items-center justify-center">
                      <Badge variant="danger">已下架</Badge>
                    </div>
                  )}
                </div>
                {/* 商品信息 */}
                <div className="p-4">
                  <h3 className="font-medium text-dark-100 mb-2 truncate">{item.product_name}</h3>
                  <div className="flex items-center justify-between">
                    <span className="text-primary-400 font-bold">¥{item.product_price.toFixed(2)}</span>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleRemoveFavorite(item.product_id)}
                        title="取消收藏"
                      >
                        <i className="fas fa-heart text-red-400" />
                      </Button>
                      {item.product_status === 1 && (
                        <a href={`/product?id=${item.product_id}`}>
                          <Button size="sm">查看</Button>
                        </a>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>
    </motion.div>
  )
}
