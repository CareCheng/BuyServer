'use client'

import { useState, useEffect, useCallback, ChangeEvent } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Modal, Badge, Card, Input } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import Toggle from '@/components/common/Toggle'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * 工单模板接口
 */
interface TicketTemplate {
  id: number
  name: string
  category: string
  title_template: string
  content_template: string
  priority: number
  sort_order: number
  use_count: number
  status: number
  created_at: string
  updated_at: string
}

/**
 * 工单模板管理页面
 */
export function TicketTemplatesPage() {
  const [templates, setTemplates] = useState<TicketTemplate[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<TicketTemplate | null>(null)
  const [form, setForm] = useState({
    name: '',
    category: 'general',
    title_template: '',
    content_template: '',
    priority: 1,
    sort_order: 0,
    status: 1,
  })
  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<TicketTemplate | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  // 加载模板列表
  const loadTemplates = useCallback(async () => {
    setLoading(true)
    const res = await apiGet<{ templates: TicketTemplate[] }>('/api/admin/ticket-templates')
    if (res.success) {
      setTemplates(res.templates || [])
    }
    setLoading(false)
  }, [])

  useEffect(() => {
    loadTemplates()
  }, [loadTemplates])

  // 打开编辑弹窗
  const openModal = (template?: TicketTemplate) => {
    if (template) {
      setEditingTemplate(template)
      setForm({
        name: template.name,
        category: template.category,
        title_template: template.title_template,
        content_template: template.content_template,
        priority: template.priority,
        sort_order: template.sort_order,
        status: template.status,
      })
    } else {
      setEditingTemplate(null)
      setForm({
        name: '',
        category: 'general',
        title_template: '',
        content_template: '',
        priority: 1,
        sort_order: 0,
        status: 1,
      })
    }
    setShowModal(true)
  }

  // 保存模板
  const handleSave = async () => {
    if (!form.name.trim()) {
      toast.error('请输入模板名称')
      return
    }
    if (!form.title_template.trim()) {
      toast.error('请输入标题模板')
      return
    }
    const res = editingTemplate
      ? await apiPut(`/api/admin/ticket-template/${editingTemplate.id}`, form)
      : await apiPost('/api/admin/ticket-template', form)
    if (res.success) {
      toast.success(editingTemplate ? '模板已更新' : '模板已创建')
      setShowModal(false)
      loadTemplates()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 打开删除确认弹窗
  const openDeleteConfirm = (template: TicketTemplate) => {
    setDeleteTarget(template)
    setShowDeleteConfirm(true)
  }

  // 删除模板
  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    const res = await apiDelete(`/api/admin/ticket-template/${deleteTarget.id}`)
    setDeleteLoading(false)
    if (res.success) {
      toast.success('模板已删除')
      setShowDeleteConfirm(false)
      setDeleteTarget(null)
      loadTemplates()
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  // 切换状态
  const handleToggleStatus = async (template: TicketTemplate) => {
    const res = await apiPut(`/api/admin/ticket-template/${template.id}`, {
      ...template,
      status: template.status === 1 ? 0 : 1,
    })
    if (res.success) {
      toast.success(template.status === 1 ? '模板已禁用' : '模板已启用')
      loadTemplates()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 获取分类标签
  const getCategoryLabel = (category: string) => {
    const categories: Record<string, { label: string; variant: 'info' | 'warning' | 'success' | 'danger' }> = {
      general: { label: '常规问题', variant: 'info' },
      payment: { label: '支付问题', variant: 'warning' },
      product: { label: '商品问题', variant: 'success' },
      account: { label: '账户问题', variant: 'danger' },
      technical: { label: '技术问题', variant: 'info' },
      suggestion: { label: '建议反馈', variant: 'success' },
    }
    return categories[category] || { label: category, variant: 'info' as const }
  }

  // 获取优先级标签
  const getPriorityLabel = (priority: number) => {
    const priorities: Record<number, { label: string; variant: 'success' | 'warning' | 'danger' }> = {
      0: { label: '低', variant: 'success' },
      1: { label: '中', variant: 'warning' },
      2: { label: '高', variant: 'danger' },
    }
    return priorities[priority] || { label: '中', variant: 'warning' as const }
  }

  if (loading && templates.length === 0) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <Card
        title="工单模板"
        icon={<i className="fas fa-clipboard-list" />}
        action={
          <Button size="sm" onClick={() => openModal()}>
            <i className="fas fa-plus mr-1" />
            添加模板
          </Button>
        }
      >
        {templates.length === 0 ? (
          <div className="p-8 text-center text-dark-400">
            <i className="fas fa-clipboard text-4xl mb-4 opacity-50" />
            <p>暂无工单模板</p>
            <Button className="mt-4" onClick={() => openModal()}>
              创建第一个模板
            </Button>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">模板名称</th>
                  <th className="pb-3 font-medium">分类</th>
                  <th className="pb-3 font-medium">标题模板</th>
                  <th className="pb-3 font-medium">优先级</th>
                  <th className="pb-3 font-medium">使用次数</th>
                  <th className="pb-3 font-medium">排序</th>
                  <th className="pb-3 font-medium">状态</th>
                  <th className="pb-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {templates.map((template) => {
                  const categoryInfo = getCategoryLabel(template.category)
                  const priorityInfo = getPriorityLabel(template.priority)
                  return (
                    <motion.tr
                      key={template.id}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      className="border-b border-dark-700/50"
                    >
                      <td className="py-3">{template.name}</td>
                      <td className="py-3">
                        <Badge variant={categoryInfo.variant}>{categoryInfo.label}</Badge>
                      </td>
                      <td className="py-3 text-dark-400 max-w-xs truncate">
                        {template.title_template}
                      </td>
                      <td className="py-3">
                        <Badge variant={priorityInfo.variant}>{priorityInfo.label}</Badge>
                      </td>
                      <td className="py-3 text-green-400">{template.use_count}</td>
                      <td className="py-3">{template.sort_order}</td>
                      <td className="py-3">
                        <Badge variant={template.status === 1 ? 'success' : 'danger'}>
                          {template.status === 1 ? '启用' : '禁用'}
                        </Badge>
                      </td>
                      <td className="py-3">
                        <div className="flex gap-1">
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => handleToggleStatus(template)}
                            title={template.status === 1 ? '禁用' : '启用'}
                          >
                            <i className={`fas fa-${template.status === 1 ? 'pause' : 'play'} text-yellow-400`} />
                          </Button>
                          <Button size="sm" variant="ghost" onClick={() => openModal(template)}>
                            <i className="fas fa-edit" />
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            className="text-red-400"
                            onClick={() => openDeleteConfirm(template)}
                          >
                            <i className="fas fa-trash" />
                          </Button>
                        </div>
                      </td>
                    </motion.tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* 编辑弹窗 */}
      <Modal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        title={editingTemplate ? '编辑模板' : '添加模板'}
        size="lg"
      >
        <div className="space-y-4">
          <Input
            label="模板名称"
            placeholder="请输入模板名称"
            value={form.name}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setForm({ ...form, name: e.target.value })}
          />
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">问题分类</label>
            <select
              value={form.category}
              onChange={(e) => setForm({ ...form, category: e.target.value })}
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200"
            >
              <option value="general">常规问题</option>
              <option value="payment">支付问题</option>
              <option value="product">商品问题</option>
              <option value="account">账户问题</option>
              <option value="technical">技术问题</option>
              <option value="suggestion">建议反馈</option>
            </select>
          </div>
          <Input
            label="标题模板"
            placeholder="例如：关于{商品名称}的问题"
            value={form.title_template}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setForm({ ...form, title_template: e.target.value })}
          />
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">内容模板</label>
            <textarea
              value={form.content_template}
              onChange={(e) => setForm({ ...form, content_template: e.target.value })}
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200 h-32 resize-none"
              placeholder="请输入内容模板，可使用变量如 {订单号}、{商品名称} 等"
            />
            <div className="mt-1 text-xs text-dark-400">
              可用变量：{'{订单号}'}, {'{商品名称}'}, {'{用户名}'}, {'{日期}'}
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-2">优先级</label>
              <select
                value={form.priority}
                onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })}
                className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200"
              >
                <option value={0}>低</option>
                <option value={1}>中</option>
                <option value={2}>高</option>
              </select>
            </div>
            <Input
              label="排序"
              type="number"
              placeholder="数字越小越靠前"
              value={form.sort_order.toString()}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setForm({ ...form, sort_order: parseInt(e.target.value) || 0 })}
            />
          </div>
          <Toggle
            checked={form.status === 1}
            onChange={(checked) => setForm({ ...form, status: checked ? 1 : 0 })}
            label="启用模板"
          />
          <Button className="w-full" onClick={handleSave}>
            {editingTemplate ? '保存修改' : '创建模板'}
          </Button>
        </div>
      </Modal>

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setDeleteTarget(null) }}
        title="删除模板"
        message={`确定要删除模板 "${deleteTarget?.name}" 吗？`}
        confirmText="删除"
        variant="danger"
        onConfirm={handleDelete}
        loading={deleteLoading}
      />
    </div>
  )
}
