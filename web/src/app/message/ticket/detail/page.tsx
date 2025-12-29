'use client'

import { useState, useEffect, useRef, Suspense } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Navbar, Footer } from '@/components/layout'
import { Button, Badge, Card, Input } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime, cn } from '@/lib/utils'

/**
 * 工单接口
 */
interface Ticket {
  id: number
  ticket_no: string
  subject: string
  category: string
  priority: number
  status: number
  user_id: number
  username: string
  guest_token: string
  email: string
  rating: number
  rating_comment: string
  rated_at: string
  created_at: string
  updated_at: string
  closed_at: string
}

/**
 * 消息接口
 */
interface Message {
  id: number
  ticket_id: number
  sender_type: string
  sender_id: number
  sender_name: string
  content: string
  attachment_url: string
  attachment_name: string
  created_at: string
}

/**
 * 工单详情内容组件
 */
function TicketDetailContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const ticketNo = searchParams.get('ticket_no')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  
  const [ticket, setTicket] = useState<Ticket | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [replyContent, setReplyContent] = useState('')
  const [sending, setSending] = useState(false)
  const [closing, setClosing] = useState(false)
  const [needToken, setNeedToken] = useState(false)
  const [guestToken, setGuestToken] = useState('')
  const [showRating, setShowRating] = useState(false)
  const [rating, setRating] = useState(0)
  const [ratingComment, setRatingComment] = useState('')
  const [submittingRating, setSubmittingRating] = useState(false)
  // 关闭工单确认弹窗状态
  const [showCloseConfirm, setShowCloseConfirm] = useState(false)
  // 附件上传状态
  const [uploading, setUploading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // 滚动到底部
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  // 加载工单详情
  const loadTicket = async (token?: string) => {
    if (!ticketNo) {
      setError('工单号不能为空')
      setLoading(false)
      return
    }

    setLoading(true)
    const url = token 
      ? `/api/support/ticket/${ticketNo}?guest_token=${token}`
      : `/api/support/ticket/${ticketNo}`
    
    const res = await apiGet<{ ticket: Ticket; messages: Message[] }>(url)
    
    if (res.success) {
      setTicket(res.ticket)
      setMessages(res.messages || [])
      setNeedToken(false)
      // 保存游客令牌
      if (token) {
        localStorage.setItem('guest_token', token)
      }
    } else if (res.error === '请提供游客令牌') {
      setNeedToken(true)
      // 尝试从本地存储获取
      const savedToken = localStorage.getItem('guest_token')
      if (savedToken && !token) {
        loadTicket(savedToken)
        return
      }
    } else {
      setError(res.error || '获取工单详情失败')
    }
    setLoading(false)
  }

  useEffect(() => {
    loadTicket()
  }, [ticketNo])

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  // 提交游客令牌
  const handleSubmitToken = () => {
    if (!guestToken.trim()) {
      toast.error('请输入游客令牌')
      return
    }
    loadTicket(guestToken.trim())
  }

  // 发送回复
  const handleReply = async () => {
    if (!replyContent.trim()) {
      toast.error('请输入回复内容')
      return
    }

    setSending(true)
    const savedToken = localStorage.getItem('guest_token')
    const res = await apiPost(`/api/support/ticket/${ticketNo}/reply`, {
      content: replyContent,
      guest_token: savedToken,
    })
    setSending(false)

    if (res.success) {
      setReplyContent('')
      // 重新加载消息
      loadTicket(savedToken || undefined)
      toast.success('回复成功')
    } else {
      toast.error(res.error || '回复失败')
    }
  }

  // 上传附件
  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // 检查文件大小（默认最大 5MB）
    if (file.size > 5 * 1024 * 1024) {
      toast.error('文件大小不能超过 5MB')
      return
    }

    setUploading(true)
    const formData = new FormData()
    formData.append('file', file)
    
    const savedToken = localStorage.getItem('guest_token')
    if (savedToken) {
      formData.append('guest_token', savedToken)
    }

    try {
      const res = await fetch(`/api/support/ticket/${ticketNo}/upload`, {
        method: 'POST',
        body: formData,
        credentials: 'include',
      })
      const data = await res.json()
      
      if (data.success) {
        toast.success('附件上传成功')
        loadTicket(savedToken || undefined)
      } else {
        toast.error(data.error || '上传失败')
      }
    } catch {
      toast.error('上传失败')
    }
    setUploading(false)
    // 清空文件选择
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  // 确认关闭工单
  const confirmCloseTicket = async () => {
    setClosing(true)
    const savedToken = localStorage.getItem('guest_token')
    const res = await apiPost(`/api/support/ticket/${ticketNo}/close`, {
      guest_token: savedToken,
    })
    setClosing(false)
    setShowCloseConfirm(false)

    if (res.success) {
      toast.success('工单已关闭')
      setTicket(prev => prev ? { ...prev, status: 3 } : null)
    } else {
      toast.error(res.error || '关闭工单失败')
    }
  }

  // 提交满意度评价
  const handleSubmitRating = async () => {
    if (rating === 0) {
      toast.error('请选择评分')
      return
    }

    setSubmittingRating(true)
    const savedToken = localStorage.getItem('guest_token')

    const res = await apiPost(`/api/support/ticket/${ticketNo}/rate`, {
      rating,
      comment: ratingComment,
      guest_token: savedToken,
    })
    setSubmittingRating(false)

    if (res.success) {
      toast.success('感谢您的评价！')
      setShowRating(false)
      // 更新工单评价状态
      setTicket(prev => prev ? { ...prev, rating, rating_comment: ratingComment } : null)
    } else {
      toast.error(res.error || '提交评价失败')
    }
  }

  // 获取状态信息
  const getStatusInfo = (status: number) => {
    const statusMap: Record<number, { text: string; variant: string }> = {
      0: { text: '待处理', variant: 'warning' },
      1: { text: '处理中', variant: 'info' },
      2: { text: '已回复', variant: 'success' },
      3: { text: '已关闭', variant: 'default' },
    }
    return statusMap[status] || { text: '未知', variant: 'default' }
  }

  // 获取优先级信息
  const getPriorityInfo = (priority: number) => {
    const priorityMap: Record<number, { text: string; color: string }> = {
      0: { text: '低', color: 'text-dark-400' },
      1: { text: '中', color: 'text-amber-400' },
      2: { text: '高', color: 'text-red-400' },
    }
    return priorityMap[priority] || { text: '普通', color: 'text-dark-400' }
  }

  // 加载中
  if (loading) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 flex items-center justify-center">
          <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
        </main>
        <Footer />
      </div>
    )
  }

  // 需要游客令牌
  if (needToken) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 py-8 px-4">
          <div className="max-w-md mx-auto">
            <Card>
              <div className="text-center mb-6">
                <div className="text-5xl mb-4">🔐</div>
                <h2 className="text-xl font-bold text-dark-100 mb-2">需要验证</h2>
                <p className="text-dark-400">请输入创建工单时获得的游客令牌</p>
              </div>
              <div className="space-y-4">
                <Input
                  placeholder="请输入游客令牌"
                  value={guestToken}
                  onChange={(e) => setGuestToken(e.target.value)}
                />
                <Button variant="primary" className="w-full" onClick={handleSubmitToken}>
                  验证
                </Button>
                <Button variant="secondary" className="w-full" onClick={() => router.push('/message')}>
                  返回客服中心
                </Button>
              </div>
            </Card>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  // 错误或工单不存在
  if (error || !ticket) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 py-8 px-4">
          <div className="max-w-md mx-auto text-center">
            <div className="text-6xl mb-4">😕</div>
            <h1 className="text-2xl font-bold text-dark-100 mb-2">工单不存在</h1>
            <p className="text-dark-400 mb-6">{error || '无法找到该工单'}</p>
            <Button variant="primary" onClick={() => router.push('/message')}>
              返回客服中心
            </Button>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  const statusInfo = getStatusInfo(ticket.status)
  const priorityInfo = getPriorityInfo(ticket.priority)

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />

      <main className="flex-1 py-8 px-4">
        <div className="max-w-3xl mx-auto">
          {/* 页面标题 */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6"
          >
            <button
              onClick={() => router.push('/message')}
              className="text-dark-400 hover:text-dark-200 mb-4 flex items-center"
            >
              <i className="fas fa-arrow-left mr-2" />
              返回客服中心
            </button>
            <div className="flex items-center justify-between">
              <h1 className="text-2xl font-bold text-dark-100">{ticket.subject}</h1>
              <Badge variant={statusInfo.variant as 'success' | 'warning' | 'info' | 'default'}>
                {statusInfo.text}
              </Badge>
            </div>
            <div className="flex items-center gap-4 mt-2 text-sm text-dark-400">
              <span>工单号: {ticket.ticket_no}</span>
              <span>分类: {ticket.category}</span>
              <span className={priorityInfo.color}>优先级: {priorityInfo.text}</span>
            </div>
          </motion.div>

          {/* 消息列表 */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <Card className="mb-6">
              <div className="space-y-4 max-h-96 overflow-y-auto">
                {messages.length === 0 ? (
                  <div className="text-center py-8 text-dark-400">
                    暂无消息
                  </div>
                ) : (
                  messages.map((msg) => (
                    <div
                      key={msg.id}
                      className={cn(
                        'p-4 rounded-lg',
                        msg.sender_type === 'staff'
                          ? 'bg-primary-500/10 ml-8'
                          : 'bg-dark-700/50 mr-8'
                      )}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className={cn(
                          'font-medium',
                          msg.sender_type === 'staff' ? 'text-primary-400' : 'text-dark-200'
                        )}>
                          {msg.sender_type === 'staff' ? '客服' : '我'}
                          {msg.sender_name && ` (${msg.sender_name})`}
                        </span>
                        <span className="text-dark-500 text-sm">
                          {formatDateTime(msg.created_at)}
                        </span>
                      </div>
                      <p className="text-dark-300 whitespace-pre-wrap">{msg.content}</p>
                      {msg.attachment_url && (
                        <a
                          href={msg.attachment_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center mt-2 text-primary-400 hover:text-primary-300 text-sm"
                        >
                          <i className="fas fa-paperclip mr-1" />
                          {msg.attachment_name || '附件'}
                        </a>
                      )}
                    </div>
                  ))
                )}
                <div ref={messagesEndRef} />
              </div>
            </Card>
          </motion.div>

          {/* 回复区域 */}
          {ticket.status !== 3 && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              <Card>
                <div className="space-y-4">
                  <textarea
                    placeholder="请输入回复内容..."
                    value={replyContent}
                    onChange={(e) => setReplyContent(e.target.value)}
                    className="w-full h-32 px-4 py-3 bg-dark-800/50 border border-dark-700/50 rounded-xl text-dark-100 placeholder-dark-500 focus:outline-none focus:border-primary-500/50 resize-none"
                  />
                  <div className="flex justify-between items-center">
                    <div className="flex items-center gap-2">
                      <Button
                        variant="secondary"
                        onClick={() => setShowCloseConfirm(true)}
                        loading={closing}
                      >
                        关闭工单
                      </Button>
                      {/* 附件上传按钮 */}
                      <input
                        ref={fileInputRef}
                        type="file"
                        onChange={handleFileSelect}
                        className="hidden"
                        accept="image/*,.pdf,.doc,.docx,.xls,.xlsx,.txt,.zip,.rar"
                      />
                      <Button
                        variant="secondary"
                        onClick={() => fileInputRef.current?.click()}
                        loading={uploading}
                      >
                        <i className="fas fa-paperclip mr-1" />
                        附件
                      </Button>
                    </div>
                    <Button
                      variant="primary"
                      onClick={handleReply}
                      loading={sending}
                    >
                      <i className="fas fa-paper-plane mr-2" />
                      发送回复
                    </Button>
                  </div>
                </div>
              </Card>
            </motion.div>
          )}

          {/* 工单已关闭提示和评价 */}
          {ticket.status === 3 && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              <Card>
                {/* 已有评价 */}
                {ticket.rating > 0 ? (
                  <div className="text-center">
                    <div className="text-dark-400 mb-2">
                      <i className="fas fa-check-circle text-green-400 mr-2" />
                      您已评价此工单
                    </div>
                    <div className="flex justify-center gap-1 mb-2">
                      {[1, 2, 3, 4, 5].map((star) => (
                        <i
                          key={star}
                          className={cn(
                            'fas fa-star text-xl',
                            star <= ticket.rating ? 'text-yellow-400' : 'text-dark-600'
                          )}
                        />
                      ))}
                    </div>
                    {ticket.rating_comment && (
                      <p className="text-dark-400 text-sm">{ticket.rating_comment}</p>
                    )}
                  </div>
                ) : showRating ? (
                  /* 评价表单 */
                  <div className="space-y-4">
                    <div className="text-center">
                      <h3 className="text-lg font-medium text-dark-100 mb-2">请对本次服务进行评价</h3>
                      <p className="text-dark-400 text-sm">您的反馈将帮助我们改进服务质量</p>
                    </div>
                    {/* 星级评分 */}
                    <div className="flex justify-center gap-2">
                      {[1, 2, 3, 4, 5].map((star) => (
                        <button
                          key={star}
                          onClick={() => setRating(star)}
                          className="p-1 transition-transform hover:scale-110"
                        >
                          <i
                            className={cn(
                              'fas fa-star text-3xl transition-colors',
                              star <= rating ? 'text-yellow-400' : 'text-dark-600 hover:text-yellow-400/50'
                            )}
                          />
                        </button>
                      ))}
                    </div>
                    <div className="text-center text-sm text-dark-400">
                      {rating === 1 && '非常不满意'}
                      {rating === 2 && '不满意'}
                      {rating === 3 && '一般'}
                      {rating === 4 && '满意'}
                      {rating === 5 && '非常满意'}
                    </div>
                    {/* 评价内容 */}
                    <textarea
                      placeholder="请输入您的评价内容（可选）"
                      value={ratingComment}
                      onChange={(e) => setRatingComment(e.target.value)}
                      className="w-full h-24 px-4 py-3 bg-dark-800/50 border border-dark-700/50 rounded-xl text-dark-100 placeholder-dark-500 focus:outline-none focus:border-primary-500/50 resize-none"
                    />
                    {/* 按钮 */}
                    <div className="flex justify-center gap-4">
                      <Button
                        variant="secondary"
                        onClick={() => setShowRating(false)}
                      >
                        取消
                      </Button>
                      <Button
                        variant="primary"
                        onClick={handleSubmitRating}
                        loading={submittingRating}
                        disabled={rating === 0}
                      >
                        提交评价
                      </Button>
                    </div>
                  </div>
                ) : (
                  /* 关闭提示和评价按钮 */
                  <div className="text-center space-y-4">
                    <div className="text-dark-400">
                      <i className="fas fa-lock mr-2" />
                      此工单已关闭，如需帮助请创建新工单
                    </div>
                    <Button
                      variant="primary"
                      onClick={() => setShowRating(true)}
                    >
                      <i className="fas fa-star mr-2" />
                      评价此次服务
                    </Button>
                  </div>
                )}
              </Card>
            </motion.div>
          )}
        </div>
      </main>

      <Footer />

      {/* 关闭工单确认弹窗 */}
      <ConfirmModal
        isOpen={showCloseConfirm}
        onClose={() => setShowCloseConfirm(false)}
        title="关闭工单"
        message="确定要关闭此工单吗？关闭后将无法继续回复。"
        confirmText="关闭工单"
        variant="warning"
        onConfirm={confirmCloseTicket}
        loading={closing}
      />
    </div>
  )
}

/**
 * 工单详情页面
 * 使用 Suspense 包裹以支持 useSearchParams
 */
export default function TicketDetailPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 flex items-center justify-center">
          <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
        </main>
        <Footer />
      </div>
    }>
      <TicketDetailContent />
    </Suspense>
  )
}
