'use client'

import { Button, Card } from '@/components/ui'
import { Product } from '../types'

interface ProductListProps {
  products: Product[]
  onAdd: () => void
  onEdit: (product: Product) => void
  onDelete: (id: number) => void
  onOpenDetail: (product: Product) => void
  onOpenKami: (product: Product) => void
}

/**
 * 商品列表组件
 * 支持桌面端表格视图和移动端卡片视图
 */
export function ProductList({
  products,
  onAdd,
  onEdit,
  onDelete,
  onOpenDetail,
  onOpenKami,
}: ProductListProps) {
  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-3">
        <h2 className="text-lg font-medium text-dark-100">商品列表</h2>
        <Button size="sm" onClick={onAdd}>添加商品</Button>
      </div>
      <Card>
        {products.length === 0 ? (
          <div className="text-center py-12 text-dark-500">暂无商品</div>
        ) : (
          <>
            {/* 桌面端表格视图 */}
            <div className="hidden lg:block overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-dark-700">
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">ID</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">商品名称</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">价格</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">时长</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">卡密管理</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">库存</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">状态</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {products.map((product) => (
                    <tr key={product.id} className="border-b border-dark-700/50 hover:bg-dark-700/30">
                      <td className="py-3 px-4 text-dark-300">{product.id}</td>
                      <td className="py-3 px-4 text-dark-100">{product.name}</td>
                      <td className="py-3 px-4 text-dark-300">¥{product.price.toFixed(2)}</td>
                      <td className="py-3 px-4 text-dark-300">{product.duration}{product.duration_unit}</td>
                      <td className="py-3 px-4 text-dark-300">
                        <Button size="sm" variant="ghost" onClick={() => onOpenKami(product)}>
                          管理卡密
                        </Button>
                      </td>
                      <td className="py-3 px-4 text-dark-300">{product.stock === -1 ? '无限' : product.stock}</td>
                      <td className="py-3 px-4">
                        <span className={`px-2 py-1 rounded text-xs ${product.status === 1 ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                          {product.status === 1 ? '上架' : '下架'}
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex gap-2">
                          <Button size="sm" variant="ghost" onClick={() => onOpenDetail(product)}>详情</Button>
                          <Button size="sm" variant="ghost" onClick={() => onEdit(product)}>编辑</Button>
                          <Button size="sm" variant="ghost" onClick={() => onDelete(product.id)}>删除</Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* 移动端卡片视图 */}
            <div className="lg:hidden space-y-3">
              {products.map((product) => (
                <div key={product.id} className="bg-dark-700/30 rounded-lg p-4 space-y-3">
                  {/* 商品名称和状态 */}
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 min-w-0">
                      <h3 className="text-dark-100 font-medium truncate">{product.name}</h3>
                      <div className="flex items-center gap-2 mt-1">
                        <span className={`px-2 py-0.5 rounded text-xs ${product.status === 1 ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                          {product.status === 1 ? '上架' : '下架'}
                        </span>
                      </div>
                    </div>
                    <span className="text-primary-400 font-bold whitespace-nowrap">¥{product.price.toFixed(2)}</span>
                  </div>
                  {/* 详细信息 */}
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <div className="text-dark-400">
                      时长：<span className="text-dark-300">{product.duration}{product.duration_unit}</span>
                    </div>
                    <div className="text-dark-400">
                      库存：<span className="text-dark-300">{product.stock === -1 ? '无限' : product.stock}</span>
                    </div>
                  </div>
                  {/* 操作按钮 */}
                  <div className="flex gap-2 pt-2 border-t border-dark-600/50">
                    <Button size="sm" variant="secondary" onClick={() => onOpenKami(product)} className="flex-1">
                      <i className="fas fa-key mr-1" />卡密
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => onOpenDetail(product)} className="flex-1">
                      <i className="fas fa-file-alt mr-1" />详情
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => onEdit(product)} className="flex-1">
                      <i className="fas fa-edit mr-1" />编辑
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => onDelete(product.id)} className="flex-1">
                      <i className="fas fa-trash mr-1" />删除
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </>
        )}
      </Card>
    </div>
  )
}
