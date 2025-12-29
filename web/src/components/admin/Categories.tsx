'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Input, Modal } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { Category } from './types'

export function CategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingCategory, setEditingCategory] = useState<Category | null>(null)
  const [form, setForm] = useState({ name: '', icon: '', sort_order: '0', status: '1' })
  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Category | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  const loadCategories = useCallback(async () => {
    const res = await apiGet<{ categories: Category[] }>('/api/admin/categories')
    if (res.success) setCategories(res.categories || [])
    setLoading(false)
  }, [])

  useEffect(() => { loadCategories() }, [loadCategories])

  const openAddModal = () => {
    setEditingCategory(null)
    setForm({ name: '', icon: '', sort_order: '0', status: '1' })
    setShowModal(true)
  }

  const openEditModal = (cat: Category) => {
    setEditingCategory(cat)
    setForm({ name: cat.name, icon: cat.icon || '', sort_order: String(cat.sort_order), status: String(cat.status) })
    setShowModal(true)
  }

  const handleSave = async () => {
    if (!form.name.trim()) { toast.error('请输入分类名称'); return }
    const data = { name: form.name.trim(), icon: form.icon.trim(), sort_order: parseInt(form.sort_order) || 0, status: parseInt(form.status) }
    const res = editingCategory
      ? await apiPut(`/api/admin/category/${editingCategory.id}`, data)
      : await apiPost('/api/admin/category', data)
    if (res.success) { toast.success('保存成功'); setShowModal(false); loadCategories() }
    else toast.error(res.error || '保存失败')
  }

  // 打开删除确认弹窗
  const openDeleteConfirm = (cat: Category) => {
    setDeleteTarget(cat)
    setShowDeleteConfirm(true)
  }

  // 执行删除
  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    const res = await apiDelete(`/api/admin/category/${deleteTarget.id}`)
    setDeleteLoading(false)
    if (res.success) {
      toast.success('删除成功')
      setShowDeleteConfirm(false)
      setDeleteTarget(null)
      loadCategories()
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  if (loading) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-medium text-dark-100">分类列表</h2>
        <Button size="sm" onClick={openAddModal}>添加分类</Button>
      </div>
      <Card>
        {categories.length === 0 ? (
          <div className="text-center py-12 text-dark-500">暂无分类</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-dark-700">
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">ID</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">图标</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">名称</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">排序</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">状态</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">操作</th>
                </tr>
              </thead>
              <tbody>
                {categories.map((cat) => (
                  <tr key={cat.id} className="border-b border-dark-700/50 hover:bg-dark-700/30">
                    <td className="py-3 px-4 text-dark-300">{cat.id}</td>
                    <td className="py-3 px-4 text-dark-300">{cat.icon || '-'}</td>
                    <td className="py-3 px-4 text-dark-100">{cat.name}</td>
                    <td className="py-3 px-4 text-dark-300">{cat.sort_order}</td>
                    <td className="py-3 px-4">
                      <span className={`px-2 py-1 rounded text-xs ${cat.status === 1 ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                        {cat.status === 1 ? '启用' : '禁用'}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex gap-2">
                        <Button size="sm" variant="ghost" onClick={() => openEditModal(cat)}>编辑</Button>
                        <Button size="sm" variant="ghost" onClick={() => openDeleteConfirm(cat)}>删除</Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <Modal isOpen={showModal} onClose={() => setShowModal(false)} title={editingCategory ? '编辑分类' : '添加分类'}>
        <div className="space-y-4">
          <Input label="分类名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <Input label="图标 (Emoji)" value={form.icon} onChange={(e) => setForm({ ...form, icon: e.target.value })} placeholder="📦" />
          <div className="grid grid-cols-2 gap-4">
            <Input label="排序" type="number" value={form.sort_order} onChange={(e) => setForm({ ...form, sort_order: e.target.value })} />
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-1">状态</label>
              <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
                <option value="1">启用</option>
                <option value="0">禁用</option>
              </select>
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-4">
            <Button variant="secondary" onClick={() => setShowModal(false)}>取消</Button>
            <Button onClick={handleSave}>保存</Button>
          </div>
        </div>
      </Modal>

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setDeleteTarget(null) }}
        title="删除分类"
        message={`确定要删除分类 "${deleteTarget?.name}" 吗？`}
        confirmText="删除"
        variant="danger"
        onConfirm={handleDelete}
        loading={deleteLoading}
      />
    </div>
  )
}
