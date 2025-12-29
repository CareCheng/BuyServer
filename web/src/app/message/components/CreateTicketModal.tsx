'use client'

import { useState } from 'react'
import toast from 'react-hot-toast'
import { Button, Input, Modal } from '@/components/ui'
import { apiPost } from '@/lib/api'
import { SupportConfig } from './types'

/**
 * 创建工单弹窗
 */
export function CreateTicketModal({
  isOpen,
  onClose,
  config,
  isLoggedIn,
  guestToken,
  setGuestToken,
  onSuccess,
}: {
  isOpen: boolean
  onClose: () => void
  config: SupportConfig
  isLoggedIn: boolean
  guestToken: string
  setGuestToken: (token: string) => void
  onSuccess: () => void
}) {
  const [subject, setSubject] = useState('')
  const [category, setCategory] = useState('')
  const [content, setContent] = useState('')
  const [email, setEmail] = useState('')
  const [priority, setPriority] = useState(1)
  const [relatedOrder, setRelatedOrder] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // 解析分类
  const categories = (() => {
    try {
      return JSON.parse(config.categories || '[]') as string[]
    } catch {
      return ['订单问题', '商品咨询', '支付问题', '账户问题', '其他']
    }
  })()

  const handleSubmit = async () => {
    if (!subject.trim()) {
      toast.error('请输入工单主题')
      return
    }
    if (!category) {
      toast.error('请选择问题分类')
      return
    }
    if (!content.trim()) {
      toast.error('请输入问题描述')
      return
    }
    if (!isLoggedIn && !email.trim()) {
      toast.error('请输入联系邮箱')
      return
    }

    setSubmitting(true)
    const res = await apiPost<{ ticket_no: string; guest_token: string }>('/api/support/ticket', {
      subject,
      category,
      content,
      email,
      priority,
      related_order: relatedOrder,
      guest_token: guestToken,
    })

    if (res.success) {
      toast.success(`工单已提交，编号: ${res.ticket_no}`)
      if (res.guest_token) {
        setGuestToken(res.guest_token)
        localStorage.setItem('guest_token', res.guest_token)
        toast.success('请保存您的访问令牌以便后续查看工单', { duration: 5000 })
      }
      // 重置表单
      setSubject('')
      setCategory('')
      setContent('')
      setEmail('')
      setPriority(1)
      setRelatedOrder('')
      onSuccess()
    } else {
      toast.error(res.error || '提交失败')
    }
    setSubmitting(false)
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="提交工单" size="lg">
      <div className="space-y-4">
        <Input
          label="工单主题"
          value={subject}
          onChange={(e) => setSubject(e.target.value)}
          placeholder="简要描述您的问题"
        />

        <div className="space-y-1.5">
          <label className="block text-sm font-medium text-dark-300">问题分类</label>
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            className="input w-full"
          >
            <option value="">请选择分类</option>
            {categories.map((cat) => (
              <option key={cat} value={cat}>{cat}</option>
            ))}
          </select>
        </div>

        <div className="space-y-1.5">
          <label className="block text-sm font-medium text-dark-300">优先级</label>
          <select
            value={priority}
            onChange={(e) => setPriority(Number(e.target.value))}
            className="input w-full"
          >
            <option value={1}>普通</option>
            <option value={2}>紧急</option>
            <option value={3}>非常紧急</option>
          </select>
        </div>

        <div className="space-y-1.5">
          <label className="block text-sm font-medium text-dark-300">问题描述</label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="详细描述您遇到的问题..."
            className="input w-full h-32 resize-none"
          />
        </div>

        {!isLoggedIn && (
          <Input
            label="联系邮箱"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="用于接收工单回复通知"
          />
        )}

        <Input
          label="关联订单号（可选）"
          value={relatedOrder}
          onChange={(e) => setRelatedOrder(e.target.value)}
          placeholder="如有相关订单请填写"
        />

        <div className="flex flex-col sm:flex-row justify-end gap-3 pt-4">
          <Button variant="secondary" className="w-full sm:w-auto" onClick={onClose}>
            取消
          </Button>
          <Button className="w-full sm:w-auto" onClick={handleSubmit} loading={submitting}>
            提交工单
          </Button>
        </div>
      </div>
    </Modal>
  )
}
