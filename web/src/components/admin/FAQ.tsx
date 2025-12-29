'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Card, Badge, Modal, Input } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import Toggle from '@/components/common/Toggle'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * FAQ 分类接口
 */
interface FAQCategory {
  id: number
  name: string
  sort_order: number
  is_active: boolean
  created_at: string
}

/**
 * FAQ 条目接口
 */
interface FAQItem {
  id: number
  category_id: number
  category_name?: string
  question: string
  answer: string
  sort_order: number
  is_active: boolean
  view_count: number
  created_at: string
  updated_at: string
}

/**
 * FAQ 管理页面
 */
export function FAQPage() {
  const [activeSubTab, setActiveSubTab] = useState<'categories' | 'items'>('items')
  const [categories, setCategories] = useState<FAQCategory[]>([])
  const [items, setItems] = useState<FAQItem[]>([])
  const [loading, setLoading] = useState(true)
  const [showCategoryModal, setShowCategoryModal] = useState(false)
  const [showItemModal, setShowItemModal] = useState(false)
  const [editingCategory, setEditingCategory] = useState<FAQCategory | null>(null)
  const [editingItem, setEditingItem] = useState<FAQItem | null>(null)
  const [categoryForm, setCategoryForm] = useState({ name: '', sort_order: 0, is_active: true })
  const [itemForm, setItemForm] = useState({ category_id: 0, question: '', answer: '', sort_order: 0, is_active: true })
  const [saving, setSaving] = useState(false)
  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<{ type: 'category' | 'item'; id: number; name: string } | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  // 加载分类列表
  const loadCategories = async () => {
    const res = await apiGet<{ categories: FAQCategory[] }>('/api/admin/faq/categories')
    if (res.success && res.categories) {
      setCategories(res.categories)
    }
  }

  // 加载 FAQ 列表
  const loadItems = async () => {
    const res = await apiGet<{ faqs: FAQItem[] }>('/api/admin/faqs')
    if (res.success && res.faqs) {
      setItems(res.faqs)
    }
  }

  // 初始加载
  useEffect(() => {
    const load = async () => {
      setLoading(true)
      await Promise.all([loadCategories(), loadItems()])
      setLoading(false)
    }
    load()
  }, [])

  // 打开分类编辑弹窗
  const openCategoryModal = (category?: FAQCategory) => {
    if (category) {
      setEditingCategory(category)
      setCategoryForm({
        name: category.name,
        sort_order: category.sort_order,
        is_active: category.is_active,
      })
    } else {
      setEditingCategory(null)
      setCategoryForm({ name: '', sort_order: 0, is_active: true })
    }
    setShowCategoryModal(true)
  }

  // 保存分类
  const saveCategory = async () => {
    if (!categoryForm.name.trim()) {
      toast.error('请输入分类名称')
      return
    }

    setSaving(true)
    const url = editingCategory ? `/api/admin/faq/category/${editingCategory.id}` : '/api/admin/faq/category'
    const method = editingCategory ? apiPut : apiPost
    const res = await method(url, categoryForm as Record<string, unknown>)
    setSaving(false)

    if (res.success) {
      toast.success(editingCategory ? '分类已更新' : '分类已创建')
      setShowCategoryModal(false)
      loadCategories()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 打开删除确认弹窗
  const openDeleteConfirm = (type: 'category' | 'item', id: number, name: string) => {
    setDeleteTarget({ type, id, name })
    setShowDeleteConfirm(true)
  }

  // 执行删除
  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    const url = deleteTarget.type === 'category' 
      ? `/api/admin/faq/category/${deleteTarget.id}`
      : `/api/admin/faq/${deleteTarget.id}`
    const res = await apiDelete(url)
    setDeleteLoading(false)
    if (res.success) {
      toast.success(deleteTarget.type === 'category' ? '分类已删除' : 'FAQ已删除')
      setShowDeleteConfirm(false)
      setDeleteTarget(null)
      if (deleteTarget.type === 'category') {
        loadCategories()
        loadItems()
      } else {
        loadItems()
      }
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  // 删除分类（打开确认弹窗）
  const deleteCategory = (id: number) => {
    const cat = categories.find(c => c.id === id)
    openDeleteConfirm('category', id, cat?.name || '')
  }

  // 删除 FAQ（打开确认弹窗）
  const deleteItem = (id: number) => {
    const item = items.find(i => i.id === id)
    openDeleteConfirm('item', id, item?.question || '')
  }

  // 打开 FAQ 编辑弹窗
  const openItemModal = (item?: FAQItem) => {
    if (item) {
      setEditingItem(item)
      setItemForm({
        category_id: item.category_id,
        question: item.question,
        answer: item.answer,
        sort_order: item.sort_order,
        is_active: item.is_active,
      })
    } else {
      setEditingItem(null)
      setItemForm({
        category_id: categories.length > 0 ? categories[0].id : 0,
        question: '',
        answer: '',
        sort_order: 0,
        is_active: true,
      })
    }
    setShowItemModal(true)
  }

  // 保存 FAQ
  const saveItem = async () => {
    if (!itemForm.question.trim()) {
      toast.error('请输入问题')
      return
    }
    if (!itemForm.answer.trim()) {
      toast.error('请输入答案')
      return
    }
    if (!itemForm.category_id) {
      toast.error('请选择分类')
      return
    }

    setSaving(true)
    const url = editingItem ? `/api/admin/faq/${editingItem.id}` : '/api/admin/faq'
    const method = editingItem ? apiPut : apiPost
    const res = await method(url, itemForm as Record<string, unknown>)
    setSaving(false)

    if (res.success) {
      toast.success(editingItem ? 'FAQ已更新' : 'FAQ已创建')
      setShowItemModal(false)
      loadItems()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 切换状态
  const toggleItemStatus = async (item: FAQItem) => {
    const res = await apiPut(`/api/admin/faq/${item.id}`, {
      ...item,
      is_active: !item.is_active,
    } as Record<string, unknown>)
    if (res.success) {
      toast.success(item.is_active ? '已禁用' : '已启用')
      loadItems()
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
    <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
      {/* 子标签切换 */}
      <div className="flex gap-2">
        <button
          onClick={() => setActiveSubTab('items')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeSubTab === 'items'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-question-circle mr-2" />
          FAQ 列表
        </button>
        <button
          onClick={() => setActiveSubTab('categories')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeSubTab === 'categories'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-folder mr-2" />
          分类管理
        </button>
      </div>

      {/* 分类管理 */}
      {activeSubTab === 'categories' && (
        <Card
          title="FAQ 分类"
          icon={<i className="fas fa-folder" />}
          action={
            <Button size="sm" onClick={() => openCategoryModal()}>
              <i className="fas fa-plus mr-1" />新增分类
            </Button>
          }
        >
          {categories.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-5xl mb-4">📁</div>
              <p className="text-dark-400">暂无分类，请先创建分类</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-dark-700/50">
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">ID</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">名称</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">排序</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">状态</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">创建时间</th>
                    <th className="text-right py-3 px-4 text-dark-400 font-medium">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {categories.map((category) => (
                    <tr key={category.id} className="border-b border-dark-700/30 hover:bg-dark-700/20">
                      <td className="py-3 px-4 text-dark-300">{category.id}</td>
                      <td className="py-3 px-4 text-dark-100">{category.name}</td>
                      <td className="py-3 px-4 text-dark-300">{category.sort_order}</td>
                      <td className="py-3 px-4">
                        <Badge variant={category.is_active ? 'success' : 'default'}>
                          {category.is_active ? '启用' : '禁用'}
                        </Badge>
                      </td>
                      <td className="py-3 px-4 text-dark-400 text-sm">{formatDateTime(category.created_at)}</td>
                      <td className="py-3 px-4 text-right">
                        <div className="flex justify-end gap-2">
                          <Button size="sm" variant="ghost" onClick={() => openCategoryModal(category)}>
                            <i className="fas fa-edit" />
                          </Button>
                          <Button size="sm" variant="ghost" onClick={() => deleteCategory(category.id)}>
                            <i className="fas fa-trash text-red-400" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      )}

      {/* FAQ 列表 */}
      {activeSubTab === 'items' && (
        <Card
          title="FAQ 列表"
          icon={<i className="fas fa-question-circle" />}
          action={
            <Button size="sm" onClick={() => openItemModal()} disabled={categories.length === 0}>
              <i className="fas fa-plus mr-1" />新增FAQ
            </Button>
          }
        >
          {categories.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-5xl mb-4">📁</div>
              <p className="text-dark-400 mb-4">请先创建FAQ分类</p>
              <Button size="sm" onClick={() => setActiveSubTab('categories')}>
                去创建分类
              </Button>
            </div>
          ) : items.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-5xl mb-4">❓</div>
              <p className="text-dark-400">暂无FAQ</p>
            </div>
          ) : (
            <div className="space-y-4">
              {items.map((item) => (
                <div
                  key={item.id}
                  className="bg-dark-700/30 rounded-xl p-4 border border-dark-600/50"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-2">
                        <Badge variant="info">{item.category_name || '未分类'}</Badge>
                        <Badge variant={item.is_active ? 'success' : 'default'}>
                          {item.is_active ? '启用' : '禁用'}
                        </Badge>
                        <span className="text-dark-500 text-sm">
                          <i className="fas fa-eye mr-1" />{item.view_count} 次浏览
                        </span>
                      </div>
                      <h4 className="text-dark-100 font-medium mb-2">
                        <i className="fas fa-question text-primary-400 mr-2" />
                        {item.question}
                      </h4>
                      <p className="text-dark-400 text-sm line-clamp-2">{item.answer}</p>
                    </div>
                    <div className="flex gap-2 flex-shrink-0">
                      <Button size="sm" variant="ghost" onClick={() => toggleItemStatus(item)}>
                        <i className={`fas ${item.is_active ? 'fa-eye-slash' : 'fa-eye'}`} />
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => openItemModal(item)}>
                        <i className="fas fa-edit" />
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => deleteItem(item.id)}>
                        <i className="fas fa-trash text-red-400" />
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      )}

      {/* 分类编辑弹窗 */}
      <Modal
        isOpen={showCategoryModal}
        onClose={() => setShowCategoryModal(false)}
        title={editingCategory ? '编辑分类' : '新增分类'}
        size="sm"
      >
        <div className="space-y-4">
          <Input
            label="分类名称"
            value={categoryForm.name}
            onChange={(e) => setCategoryForm({ ...categoryForm, name: e.target.value })}
            placeholder="请输入分类名称"
          />
          <Input
            label="排序"
            type="number"
            value={categoryForm.sort_order}
            onChange={(e) => setCategoryForm({ ...categoryForm, sort_order: parseInt(e.target.value) || 0 })}
            placeholder="数字越小越靠前"
          />
          <Toggle
            checked={categoryForm.is_active}
            onChange={(checked) => setCategoryForm({ ...categoryForm, is_active: checked })}
            label="启用"
          />
          <div className="flex gap-3 pt-2">
            <Button variant="secondary" className="flex-1" onClick={() => setShowCategoryModal(false)}>
              取消
            </Button>
            <Button className="flex-1" onClick={saveCategory} loading={saving}>
              保存
            </Button>
          </div>
        </div>
      </Modal>

      {/* FAQ 编辑弹窗 */}
      <Modal
        isOpen={showItemModal}
        onClose={() => setShowItemModal(false)}
        title={editingItem ? '编辑FAQ' : '新增FAQ'}
        size="md"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-dark-300 text-sm mb-2">分类</label>
            <select
              value={itemForm.category_id}
              onChange={(e) => setItemForm({ ...itemForm, category_id: parseInt(e.target.value) })}
              className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600/50 rounded-lg text-dark-100 focus:outline-none focus:border-primary-500"
            >
              <option value={0}>请选择分类</option>
              {categories.filter(c => c.is_active).map((category) => (
                <option key={category.id} value={category.id}>{category.name}</option>
              ))}
            </select>
          </div>
          <Input
            label="问题"
            value={itemForm.question}
            onChange={(e) => setItemForm({ ...itemForm, question: e.target.value })}
            placeholder="请输入问题"
          />
          <div>
            <label className="block text-dark-300 text-sm mb-2">答案</label>
            <textarea
              value={itemForm.answer}
              onChange={(e) => setItemForm({ ...itemForm, answer: e.target.value })}
              placeholder="请输入答案"
              rows={5}
              className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600/50 rounded-lg text-dark-100 placeholder-dark-500 focus:outline-none focus:border-primary-500 resize-none"
            />
          </div>
          <Input
            label="排序"
            type="number"
            value={itemForm.sort_order}
            onChange={(e) => setItemForm({ ...itemForm, sort_order: parseInt(e.target.value) || 0 })}
            placeholder="数字越小越靠前"
          />
          <Toggle
            checked={itemForm.is_active}
            onChange={(checked) => setItemForm({ ...itemForm, is_active: checked })}
            label="启用"
          />
          <div className="flex gap-3 pt-2">
            <Button variant="secondary" className="flex-1" onClick={() => setShowItemModal(false)}>
              取消
            </Button>
            <Button className="flex-1" onClick={saveItem} loading={saving}>
              保存
            </Button>
          </div>
        </div>
      </Modal>

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setDeleteTarget(null) }}
        title={deleteTarget?.type === 'category' ? '删除分类' : '删除FAQ'}
        message={deleteTarget?.type === 'category' 
          ? `确定要删除分类 "${deleteTarget?.name}" 吗？该分类下的所有FAQ也会被删除。`
          : `确定要删除FAQ "${deleteTarget?.name}" 吗？`}
        confirmText="删除"
        variant="danger"
        onConfirm={handleDelete}
        loading={deleteLoading}
      />
    </motion.div>
  )
}
