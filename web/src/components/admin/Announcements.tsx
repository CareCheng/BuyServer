'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Input, Modal } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { Announcement } from './types'

export function AnnouncementsPage() {
  const [announcements, setAnnouncements] = useState<Announcement[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingAnnouncement, setEditingAnnouncement] = useState<Announcement | null>(null)
  const [form, setForm] = useState({ title: '', content: '', type: 'info', status: '1', sort_order: '0' })
  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Announcement | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  const loadAnnouncements = useCallback(async () => {
    const res = await apiGet<{ announcements: Announcement[] }>('/api/admin/announcements')
    if (res.success) setAnnouncements(res.announcements || [])
    setLoading(false)
  }, [])

  useEffect(() => { loadAnnouncements() }, [loadAnnouncements])

  const openAddModal = () => {
    setEditingAnnouncement(null)
    setForm({ title: '', content: '', type: 'info', status: '1', sort_order: '0' })
    setShowModal(true)
  }

  const openEditModal = (ann: Announcement) => {
    setEditingAnnouncement(ann)
    setForm({ title: ann.title, content: ann.content || '', type: ann.type, status: String(ann.status), sort_order: String(ann.sort_order) })
    setShowModal(true)
  }

  const handleSave = async () => {
    if (!form.title.trim()) { toast.error('请输入公告标题'); return }
    const data = { title: form.title.trim(), content: form.content.trim(), type: form.type, status: parseInt(form.status), sort_order: parseInt(form.sort_order) || 0 }
    const res = editingAnnouncement
      ? await apiPut(`/api/admin/announcement/${editingAnnouncement.id}`, data)
      : await apiPost('/api/admin/announcement', data)
    if (res.success) { toast.success('保存成功'); setShowModal(false); loadAnnouncements() }
    else toast.error(res.error || '保存失败')
  }

  // 打开删除确认弹窗
  const openDeleteConfirm = (ann: Announcement) => {
    setDeleteTarget(ann)
    setShowDeleteConfirm(true)
  }

  // 执行删除
  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    const res = await apiDelete(`/api/admin/announcement/${deleteTarget.id}`)
    setDeleteLoading(false)
    if (res.success) {
      toast.success('删除成功')
      setShowDeleteConfirm(false)
      setDeleteTarget(null)
      loadAnnouncements()
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  const getTypeText = (type: string) => {
    const types: Record<string, string> = { info: '信息', success: '成功', warning: '警告', danger: '危险' }
    return types[type] || type
  }

  if (loading) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-medium text-dark-100">公告列表</h2>
        <Button size="sm" onClick={openAddModal}>添加公告</Button>
      </div>
      <Card>
        {announcements.length === 0 ? (
          <div className="text-center py-12 text-dark-500">暂无公告</div>
        ) : (
          <div className="space-y-4">
            {announcements.map((ann) => (
              <div key={ann.id} className="p-4 bg-dark-700/30 rounded-lg">
                <div className="flex justify-between items-start mb-2">
                  <div className="flex items-center gap-2">
                    <h3 className="font-medium text-dark-100">{ann.title}</h3>
                    <span className={`px-2 py-0.5 rounded text-xs ${ann.type === 'info' ? 'bg-blue-500/20 text-blue-400' : ann.type === 'success' ? 'bg-green-500/20 text-green-400' : ann.type === 'warning' ? 'bg-yellow-500/20 text-yellow-400' : 'bg-red-500/20 text-red-400'}`}>
                      {getTypeText(ann.type)}
                    </span>
                  </div>
                  <span className={`px-2 py-1 rounded text-xs ${ann.status === 1 ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                    {ann.status === 1 ? '显示' : '隐藏'}
                  </span>
                </div>
                <p className="text-dark-400 text-sm mb-2">{ann.content}</p>
                <div className="flex justify-between items-center">
                  <span className="text-dark-500 text-xs">{ann.created_at}</span>
                  <div className="flex gap-2">
                    <Button size="sm" variant="ghost" onClick={() => openEditModal(ann)}>编辑</Button>
                    <Button size="sm" variant="ghost" onClick={() => openDeleteConfirm(ann)}>删除</Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>

      <Modal isOpen={showModal} onClose={() => setShowModal(false)} title={editingAnnouncement ? '编辑公告' : '添加公告'}>
        <div className="space-y-4">
          <Input label="公告标题" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required />
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">公告内容</label>
            <textarea className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100 h-32" value={form.content} onChange={(e) => setForm({ ...form, content: e.target.value })} />
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-1">类型</label>
              <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })}>
                <option value="info">信息</option>
                <option value="success">成功</option>
                <option value="warning">警告</option>
                <option value="danger">危险</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-1">状态</label>
              <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
                <option value="1">显示</option>
                <option value="0">隐藏</option>
              </select>
            </div>
            <Input label="排序" type="number" value={form.sort_order} onChange={(e) => setForm({ ...form, sort_order: e.target.value })} />
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
        title="删除公告"
        message={`确定要删除公告 "${deleteTarget?.title}" 吗？`}
        confirmText="删除"
        variant="danger"
        onConfirm={handleDelete}
        loading={deleteLoading}
      />
    </div>
  )
}
