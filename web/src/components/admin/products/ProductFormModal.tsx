'use client'

import { Button, Input, Modal } from '@/components/ui'
import { Product } from '../types'
import { ProductFormData } from './types'

interface ProductFormModalProps {
  isOpen: boolean
  onClose: () => void
  editingProduct: Product | null
  form: ProductFormData
  setForm: (form: ProductFormData) => void
  onSave: () => void
}

/**
 * 商品表单弹窗组件
 * 用于创建和编辑商品（手动卡密模式）
 */
export function ProductFormModal({
  isOpen,
  onClose,
  editingProduct,
  form,
  setForm,
  onSave,
}: ProductFormModalProps) {
  return (
    <Modal isOpen={isOpen} onClose={onClose} title={editingProduct ? '编辑商品' : '添加商品'}>
      <div className="space-y-4">
        <Input label="商品名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
        
        <div className="grid grid-cols-2 gap-4">
          <Input label="价格" type="number" step="0.01" value={form.price} onChange={(e) => setForm({ ...form, price: e.target.value })} required />
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">库存</label>
            <div className="px-3 py-2 bg-dark-800 border border-dark-600 rounded-lg text-dark-400">
              由卡密数量自动计算
            </div>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <Input label="时长" type="number" value={form.duration} onChange={(e) => setForm({ ...form, duration: e.target.value })} />
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">单位</label>
            <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={form.duration_unit} onChange={(e) => setForm({ ...form, duration_unit: e.target.value })}>
              <option value="天">天</option>
              <option value="周">周</option>
              <option value="月">月</option>
              <option value="年">年</option>
            </select>
          </div>
        </div>
        
        <div>
          <label className="block text-sm font-medium text-dark-300 mb-1">状态</label>
          <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
            <option value="1">上架</option>
            <option value="0">下架</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-dark-300 mb-1">描述</label>
          <textarea className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100 h-20" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
        </div>
        <div className="flex justify-end gap-2 pt-4">
          <Button variant="secondary" onClick={onClose}>取消</Button>
          <Button onClick={onSave}>保存</Button>
        </div>
      </div>
    </Modal>
  )
}
