'use client'

import { useState, useEffect, useRef } from 'react'
import toast from 'react-hot-toast'
import { Button, Badge, Modal } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime, cn } from '@/lib/utils'
import { Ticket, TicketMessage } from './types'

/**
 * 工单详情弹窗
 */
export function TicketDetailModal({
  isOpen,
  onClose,
  ticket,
  guestToken,
  isLoggedIn,
  onUpdate,
}: {
  isOpen: boolean
  onClose: () => void
  ticket: Ticket
  guestToken: string
  isLoggedIn: boolean
  onUpdate: () => void
}) {
  const [messages, setMessages] = useState<TicketMessage[]>([])
  const [loading, setLoading] = useState(true)
  const [replyContent, setReplyContent] = useState('')
  const [sending, setSending] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  // 加载工单详情和消息
  const loadDetail = async () => {
    setLoading(true)
    const url = isLoggedIn
      ? `/api/support/ticket/${ticket.ticket_no}`
      : `/api/support/ticket/${ticket.ticket_no}?guest_token=${guestToken}`
    
    const res = await apiGet<{ ticket: Ticket; messages: TicketMessage[] }>(url)
    if (res.success && res.messages) {
      setMessages(res.messages)
    }
    setLoading(false)
  }

  useEffect(() => {
    if (isOpen) {
      loadDetail()
    }
  }, [isOpen, ticket.ticket_no])

  // 滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  // 回复工单
  const handleReply = async () => {
    if (!replyContent.trim()) {
      toast.error('请输入回复内容')
      return
    }

    setSending(true)
    const res = await apiPost(`/api/support/ticket/${ticket.ticket_no}/reply`, {
      content: replyContent,
      guest_token: guestToken,
    })

    if (res.success) {
      toast.success('回复成功')
      setReplyContent('')
      loadDetail()
      onUpdate()
    } else {
      toast.error(res.error || '回复失败')
    }
    setSending(false)
  }

  // 关闭工单
  const handleClose = async () => {
    const res = await apiPost(`/api/support/ticket/${ticket.ticket_no}/close`, {
      guest_token: guestToken,
    })

    if (res.success) {
      toast.success('工单已关闭')
      onUpdate()
      onClose()
    } else {
      toast.error(res.error || '关闭失败')
    }
  }

  // 获取状态文本
  const getStatusText = (status: number) => {
    const map: Record<number, string> = {
      0: '待处理',
      1: '处理中',
      2: '已回复',
      3: '已解决',
      4: '已关闭',
    }
    return map[status] || '未知'
  }

  const isClosed = ticket.status === 4

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={`工单 #${ticket.ticket_no}`} size="lg">
      {/* 工单信息 */}
      <div className="bg-dark-700/30 rounded-lg p-4 mb-4">
        <h4 className="text-dark-100 font-medium mb-2">{ticket.subject}</h4>
        <div className="text-dark-400 text-sm space-y-1">
          <div>状态: {getStatusText(ticket.status)} | 分类: {ticket.category}</div>
          <div>创建时间: {formatDateTime(ticket.created_at)}</div>
        </div>
      </div>

      {/* 消息列表 */}
      <div className="h-64 overflow-y-auto mb-4 space-y-3">
        {loading ? (
          <div className="text-center py-8">
            <i className="fas fa-spinner fa-spin text-xl text-primary-400" />
          </div>
        ) : (
          messages.map((msg) => (
            <div
              key={msg.id}
              className={cn(
                'rounded-lg p-3',
                msg.sender_type === 'user' || msg.sender_type === 'guest'
                  ? 'bg-primary-500/20 ml-8'
                  : msg.sender_type === 'system'
                  ? 'bg-dark-600/30 text-center text-sm'
                  : 'bg-dark-700/50 mr-8'
              )}
            >
              <div className="flex justify-between items-center mb-1">
                <span className="text-dark-300 text-sm font-medium">
                  {msg.sender_name}
                  {msg.sender_type === 'staff' && (
                    <Badge variant="info" className="ml-2 text-xs">客服</Badge>
                  )}
                </span>
                <span className="text-dark-500 text-xs">
                  {formatDateTime(msg.created_at)}
                </span>
              </div>
              <div className="text-dark-100 whitespace-pre-wrap">{msg.content}</div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* 回复区域 */}
      {!isClosed ? (
        <div className="space-y-3">
          <textarea
            value={replyContent}
            onChange={(e) => setReplyContent(e.target.value)}
            placeholder="输入回复内容..."
            className="input w-full h-24 resize-none"
          />
          <div className="flex flex-col sm:flex-row justify-between items-stretch sm:items-center gap-3">
            <Button variant="danger" size="sm" onClick={handleClose} className="order-2 sm:order-1">
              <i className="fas fa-times mr-2" />
              关闭工单
            </Button>
            <Button onClick={handleReply} loading={sending} className="order-1 sm:order-2">
              <i className="fas fa-reply mr-2" />
              发送回复
            </Button>
          </div>
        </div>
      ) : (
        <div className="text-center text-dark-400 py-4 bg-dark-700/30 rounded-lg">
          <i className="fas fa-lock mr-2" />
          工单已关闭，无法继续回复
        </div>
      )}
    </Modal>
  )
}
