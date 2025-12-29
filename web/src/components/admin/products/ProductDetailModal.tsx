'use client'

import dynamic from 'next/dynamic'
import { Button, Input, Modal } from '@/components/ui'
import { Product } from '../types'
import { DetailFormData } from './types'

// 动态导入 Markdown 编辑器（避免 SSR 问题）
const MDEditor = dynamic(() => import('@uiw/react-md-editor'), { ssr: false })

interface ProductDetailModalProps {
  isOpen: boolean
  onClose: () => void
  product: Product | null
  form: DetailFormData
  setForm: (form: DetailFormData) => void
  saving: boolean
  onSave: () => void
}

/**
 * 商品详情编辑弹窗组件
 * 支持 Markdown 详情、规格参数、特性和标签编辑
 */
export function ProductDetailModal({
  isOpen,
  onClose,
  product,
  form,
  setForm,
  saving,
  onSave,
}: ProductDetailModalProps) {
  // 添加规格项
  const addSpecItem = () => {
    setForm({ ...form, specs: [...form.specs, { key: '', value: '' }] })
  }

  // 删除规格项
  const removeSpecItem = (index: number) => {
    const newSpecs = form.specs.filter((_, i) => i !== index)
    setForm({ ...form, specs: newSpecs.length > 0 ? newSpecs : [{ key: '', value: '' }] })
  }

  // 更新规格项
  const updateSpecItem = (index: number, field: 'key' | 'value', value: string) => {
    const newSpecs = [...form.specs]
    newSpecs[index][field] = value
    setForm({ ...form, specs: newSpecs })
  }

  // 添加特性
  const addFeature = () => {
    setForm({ ...form, features: [...form.features, ''] })
  }

  // 删除特性
  const removeFeature = (index: number) => {
    const newFeatures = form.features.filter((_, i) => i !== index)
    setForm({ ...form, features: newFeatures.length > 0 ? newFeatures : [''] })
  }

  // 更新特性
  const updateFeature = (index: number, value: string) => {
    const newFeatures = [...form.features]
    newFeatures[index] = value
    setForm({ ...form, features: newFeatures })
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={`商品详情 - ${product?.name || ''}`} size="lg">
      <div className="space-y-6 max-h-[70vh] overflow-y-auto">
        {/* Markdown 详情编辑器 */}
        <div>
          <label className="block text-sm font-medium text-dark-300 mb-2">
            <i className="fas fa-file-alt mr-1" />商品详情（支持 Markdown）
          </label>
          <div data-color-mode="dark">
            <MDEditor
              value={form.detail}
              onChange={(val) => setForm({ ...form, detail: val || '' })}
              height={300}
              preview="edit"
            />
          </div>
        </div>

        {/* 规格参数编辑器 */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="text-sm font-medium text-dark-300">
              <i className="fas fa-list-ul mr-1" />规格参数
            </label>
            <Button size="sm" variant="ghost" onClick={addSpecItem}>
              <i className="fas fa-plus mr-1" />添加
            </Button>
          </div>
          <div className="space-y-2">
            {form.specs.map((spec, index) => (
              <div key={index} className="flex gap-2 items-center">
                <Input
                  placeholder="参数名"
                  value={spec.key}
                  onChange={(e) => updateSpecItem(index, 'key', e.target.value)}
                  className="flex-1"
                />
                <Input
                  placeholder="参数值"
                  value={spec.value}
                  onChange={(e) => updateSpecItem(index, 'value', e.target.value)}
                  className="flex-1"
                />
                <Button size="sm" variant="ghost" onClick={() => removeSpecItem(index)}>
                  <i className="fas fa-times text-red-400" />
                </Button>
              </div>
            ))}
          </div>
        </div>

        {/* 特性/卖点编辑器 */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="text-sm font-medium text-dark-300">
              <i className="fas fa-star mr-1" />特性/卖点
            </label>
            <Button size="sm" variant="ghost" onClick={addFeature}>
              <i className="fas fa-plus mr-1" />添加
            </Button>
          </div>
          <div className="space-y-2">
            {form.features.map((feature, index) => (
              <div key={index} className="flex gap-2 items-center">
                <Input
                  placeholder="输入特性描述"
                  value={feature}
                  onChange={(e) => updateFeature(index, e.target.value)}
                  className="flex-1"
                />
                <Button size="sm" variant="ghost" onClick={() => removeFeature(index)}>
                  <i className="fas fa-times text-red-400" />
                </Button>
              </div>
            ))}
          </div>
        </div>

        {/* 标签编辑器 */}
        <div>
          <label className="block text-sm font-medium text-dark-300 mb-2">
            <i className="fas fa-tags mr-1" />商品标签
          </label>
          <Input
            placeholder="多个标签用逗号分隔，如：热销,推荐,限时优惠"
            value={form.tags}
            onChange={(e) => setForm({ ...form, tags: e.target.value })}
          />
          {form.tags && (
            <div className="flex flex-wrap gap-1 mt-2">
              {form.tags.split(',').filter(t => t.trim()).map((tag, i) => (
                <span key={i} className="px-2 py-1 bg-primary-500/20 text-primary-400 text-xs rounded">
                  {tag.trim()}
                </span>
              ))}
            </div>
          )}
        </div>

        {/* 保存按钮 */}
        <div className="flex justify-end gap-2 pt-4 border-t border-dark-700">
          <Button variant="secondary" onClick={onClose}>取消</Button>
          <Button onClick={onSave} loading={saving}>保存详情</Button>
        </div>
      </div>
    </Modal>
  )
}
