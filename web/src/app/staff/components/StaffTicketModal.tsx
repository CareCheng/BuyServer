'use client'

import { useState, useEffect, useRef } from 'react'
import toast from 'react-hot-toast'
import { Button, Badge, Card, Modal } from '@/components/ui'
import { apiGet, apiPost, apiPut } from '@/lib/api'
import { formatDateTime, cn, formatFileSize } from '@/lib/utils'
import { StaffInfo, Ticket, TicketMessage, OnlineStaff, TicketAttachment } from './types'

/**
 * 客服工单详情弹窗
 */
export function StaffTicketModal({
  isOpen,
  onClose,
  ticket,
  staff,
  onUpdate,
}: {
  isOpen: boolean
  onClose: () => void
  ticket: Ticket
  staff: StaffInfo
  onUpdate: () => void
}) {
  const [messages, setMessages] = useState<TicketMessage[]>([])
  const [attachments, setAttachments] = useState<TicketAttachment[]>([])
  const [loading, setLoading] = useState(true)
  const [replyContent, setReplyContent] = useState('')
  const [isInternal, setIsInternal] = useState(false)
  const [sending, setSending] = useState(false)
  const [newStatus, setNewStatus] = useState(ticket.status)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  
  // 转接相关状态
  const [showTransferModal, setShowTransferModal] = useState(false)
  const [onlineStaff, setOnlineStaff] = useState<OnlineStaff[]>([])
  const [transferTarget, setTransferTarget] = useState<number>(0)
  const [transferReason, setTransferReason] = useState('')
  const [transferring, setTransferring] = useState(false)
  
  // 附件上传相关状态
  const [uploading, setUploading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // 加载工单详情
  const loadDetail = async () => {
    setLoading(true)
    const res = await apiGet<{ ticket: Ticket; messages: TicketMessage[] }>(
      `/api/staff/ticket/${ticket.ticket_no}`
    )
    if (res.success && res.messages) {
      setMessages(res.messages)
    }
    // 加载附件列表
    const attachRes = await apiGet<{ attachments: TicketAttachment[] }>(
      `/api/staff/ticket/${ticket.ticket_no}/attachments`
    )
    if (attachRes.success && attachRes.attachments) {
      setAttachments(attachRes.attachments)
    }
    setLoading(false)
  }

  useEffect(() => {
    if (isOpen) {
      loadDetail()
      setNewStatus(ticket.status)
    }
  }, [isOpen, ticket.ticket_no])

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
    const res = await apiPost(`/api/staff/ticket/${ticket.ticket_no}/reply`, {
      content: replyContent,
      is_internal: isInternal,
    })

    if (res.success) {
      toast.success(isInternal ? '内部备注已添加' : '回复成功')
      setReplyContent('')
      setIsInternal(false)
      loadDetail()
      onUpdate()
    } else {
      toast.error(res.error || '回复失败')
    }
    setSending(false)
  }

  // 更新状态
  const handleUpdateStatus = async () => {
    if (newStatus === ticket.status) return

    const res = await apiPut(`/api/staff/ticket/${ticket.ticket_no}/status`, {
      status: newStatus,
    })

    if (res.success) {
      toast.success('状态已更新')
      onUpdate()
    } else {
      toast.error(res.error || '更新失败')
    }
  }

  // 分配给自己
  const handleAssignToMe = async () => {
    const res = await apiPost(`/api/staff/ticket/${ticket.ticket_no}/assign`, {
      staff_id: staff.id,
    })

    if (res.success) {
      toast.success('已分配给您')
      onUpdate()
    } else {
      toast.error(res.error || '分配失败')
    }
  }

  // 打开转接弹窗
  const openTransferModal = async () => {
    // 获取在线客服列表
    const res = await apiGet<{ staff: OnlineStaff[] }>('/api/staff/staff/online')
    if (res.success && res.staff) {
      // 过滤掉自己
      const filtered = res.staff.filter(s => s.id !== staff.id)
      setOnlineStaff(filtered)
    }
    setShowTransferModal(true)
  }

  // 执行转接
  const handleTransfer = async () => {
    if (!transferTarget) {
      toast.error('请选择转接目标')
      return
    }

    setTransferring(true)
    const res = await apiPost(`/api/staff/ticket/${ticket.ticket_no}/transfer`, {
      to_staff_id: transferTarget,
      reason: transferReason,
    })

    if (res.success) {
      toast.success('工单已转接')
      setShowTransferModal(false)
      setTransferTarget(0)
      setTransferReason('')
      onUpdate()
      onClose()
    } else {
      toast.error(res.error || '转接失败')
    }
    setTransferring(false)
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

    try {
      const res = await fetch(`/api/staff/ticket/${ticket.ticket_no}/upload`, {
        method: 'POST',
        body: formData,
        credentials: 'include',
      })
      const data = await res.json()
      
      if (data.success) {
        toast.success('附件上传成功')
        loadDetail()
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

  const isClosed = ticket.status === 4 || ticket.status === 5

  // 获取消息类型图标
  const getMsgTypeIcon = (msgType?: string) => {
    if (msgType === 'image') return 'fa-image'
    if (msgType === 'file') return 'fa-file'
    return null
  }

  return (
    <>
      <Modal isOpen={isOpen} onClose={onClose} title={`工单 #${ticket.ticket_no}`} size="xl">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          {/* 左侧：消息列表 */}
          <div className="lg:col-span-2">
            <div className="bg-dark-700/30 rounded-lg p-4 mb-4">
              <h4 className="text-dark-100 font-medium mb-2">{ticket.subject}</h4>
              <div className="text-dark-400 text-sm">
                用户: {ticket.username} | 邮箱: {ticket.email}
              </div>
            </div>

            <div className="h-64 overflow-y-auto mb-4 space-y-3 bg-dark-800/50 rounded-lg p-3">
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
                      msg.is_internal
                        ? 'bg-amber-500/10 border border-amber-500/30'
                        : msg.sender_type === 'staff'
                        ? 'bg-primary-500/20 ml-8'
                        : 'bg-dark-700/50 mr-8'
                    )}
                  >
                    <div className="flex justify-between items-center mb-1">
                      <span className="text-dark-300 text-sm font-medium">
                        {msg.sender_name}
                        {msg.is_internal && (
                          <Badge variant="warning" className="ml-2 text-xs">内部备注</Badge>
                        )}
                      </span>
                      <span className="text-dark-500 text-xs">
                        {formatDateTime(msg.created_at)}
                      </span>
                    </div>
                    <div className="text-dark-100 whitespace-pre-wrap">{msg.content}</div>
                    {/* 附件显示 */}
                    {msg.file_url && (
                      <div className="mt-2">
                        {msg.msg_type === 'image' ? (
                          <a href={msg.file_url} target="_blank" rel="noopener noreferrer">
                            <img 
                              src={msg.file_url} 
                              alt={msg.file_name || '图片'} 
                              className="max-w-xs max-h-48 rounded-lg"
                            />
                          </a>
                        ) : (
                          <a 
                            href={msg.file_url} 
                            target="_blank" 
                            rel="noopener noreferrer"
                            className="inline-flex items-center gap-2 px-3 py-2 bg-dark-600/50 rounded-lg text-primary-400 hover:text-primary-300"
                          >
                            <i className={`fas ${getMsgTypeIcon(msg.msg_type) || 'fa-paperclip'}`} />
                            <span>{msg.file_name || '附件'}</span>
                            {msg.file_size && (
                              <span className="text-dark-500 text-xs">
                                ({formatFileSize(msg.file_size)})
                              </span>
                            )}
                          </a>
                        )}
                      </div>
                    )}
                  </div>
                ))
              )}
              <div ref={messagesEndRef} />
            </div>

            {/* 附件列表 */}
            {attachments.length > 0 && (
              <div className="mb-4 p-3 bg-dark-700/30 rounded-lg">
                <h5 className="text-dark-300 text-sm mb-2">
                  <i className="fas fa-paperclip mr-2" />
                  附件 ({attachments.length})
                </h5>
                <div className="flex flex-wrap gap-2">
                  {attachments.map((att) => (
                    <a
                      key={att.id}
                      href={att.file_path}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 px-2 py-1 bg-dark-600/50 rounded text-sm text-primary-400 hover:text-primary-300"
                    >
                      <i className="fas fa-file text-xs" />
                      <span className="max-w-32 truncate">{att.file_name}</span>
                    </a>
                  ))}
                </div>
              </div>
            )}

            {/* 回复区域 */}
            {!isClosed && (
              <div className="space-y-3">
                <textarea
                  value={replyContent}
                  onChange={(e) => setReplyContent(e.target.value)}
                  placeholder="输入回复内容..."
                  className="input w-full h-24 resize-none"
                />
                <div className="flex flex-col sm:flex-row justify-between items-stretch sm:items-center gap-3">
                  <div className="flex items-center gap-3">
                    <label className="flex items-center gap-2 text-dark-300 text-sm cursor-pointer p-2 bg-dark-700/30 rounded-lg hover:bg-dark-700/50 transition-colors">
                      <div
                        className={`relative w-10 h-5 rounded-full transition-colors ${
                          isInternal ? 'bg-amber-500' : 'bg-dark-600'
                        }`}
                        onClick={() => setIsInternal(!isInternal)}
                      >
                        <div
                          className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform ${
                            isInternal ? 'translate-x-5' : 'translate-x-0.5'
                          }`}
                        />
                      </div>
                      <span>内部备注</span>
                    </label>
                    {/* 附件上传按钮 */}
                    <input
                      ref={fileInputRef}
                      type="file"
                      onChange={handleFileSelect}
                      className="hidden"
                      accept="image/*,.pdf,.doc,.docx,.xls,.xlsx,.txt,.zip,.rar"
                    />
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => fileInputRef.current?.click()}
                      loading={uploading}
                    >
                      <i className="fas fa-paperclip mr-1" />
                      附件
                    </Button>
                  </div>
                  <Button onClick={handleReply} loading={sending} className="shrink-0">
                    <i className="fas fa-reply mr-2" />
                    发送
                  </Button>
                </div>
              </div>
            )}
          </div>

          {/* 右侧：工单信息 */}
          <div className="space-y-4">
            <Card title="工单信息" className="text-sm">
              <div className="space-y-3">
                <div>
                  <span className="text-dark-400">分类:</span>
                  <span className="text-dark-100 ml-2">{ticket.category}</span>
                </div>
                <div>
                  <span className="text-dark-400">优先级:</span>
                  <span className="ml-2">
                    {ticket.priority === 3 ? (
                      <Badge variant="danger">非常紧急</Badge>
                    ) : ticket.priority === 2 ? (
                      <Badge variant="warning">紧急</Badge>
                    ) : (
                      <Badge>普通</Badge>
                    )}
                  </span>
                </div>
                <div>
                  <span className="text-dark-400">处理人:</span>
                  <span className="text-dark-100 ml-2">
                    {ticket.assigned_name || '未分配'}
                  </span>
                </div>
                {ticket.transfer_count && ticket.transfer_count > 0 && (
                  <div>
                    <span className="text-dark-400">转接次数:</span>
                    <span className="text-amber-400 ml-2">{ticket.transfer_count}</span>
                  </div>
                )}
                <div>
                  <span className="text-dark-400">创建时间:</span>
                  <span className="text-dark-100 ml-2 block">
                    {formatDateTime(ticket.created_at)}
                  </span>
                </div>
              </div>
            </Card>

            {/* 操作 */}
            <Card title="操作" className="text-sm">
              <div className="space-y-3">
                {/* 状态更新 */}
                <div>
                  <label className="text-dark-400 block mb-1">更新状态</label>
                  <div className="flex items-center gap-2">
                    <select
                      value={newStatus}
                      onChange={(e) => setNewStatus(Number(e.target.value))}
                      className="input flex-1 h-10"
                    >
                      <option value={0}>待处理</option>
                      <option value={1}>处理中</option>
                      <option value={2}>已回复</option>
                      <option value={3}>已解决</option>
                      <option value={4}>已关闭</option>
                    </select>
                    <Button size="sm" onClick={handleUpdateStatus}>
                      更新
                    </Button>
                  </div>
                </div>

                {/* 分配 */}
                {!ticket.assigned_to && (
                  <Button size="sm" className="w-full" onClick={handleAssignToMe}>
                    <i className="fas fa-hand-paper mr-2" />
                    分配给我
                  </Button>
                )}

                {/* 转接按钮 */}
                {!isClosed && ticket.assigned_to === staff.id && (
                  <Button 
                    size="sm" 
                    variant="secondary" 
                    className="w-full"
                    onClick={openTransferModal}
                  >
                    <i className="fas fa-exchange-alt mr-2" />
                    转接工单
                  </Button>
                )}
              </div>
            </Card>
          </div>
        </div>
      </Modal>

      {/* 转接弹窗 */}
      <Modal
        isOpen={showTransferModal}
        onClose={() => setShowTransferModal(false)}
        title="转接工单"
        size="sm"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-dark-300 mb-2">选择转接目标</label>
            {onlineStaff.length === 0 ? (
              <div className="text-dark-400 text-sm py-4 text-center">
                暂无其他在线客服
              </div>
            ) : (
              <select
                value={transferTarget}
                onChange={(e) => setTransferTarget(Number(e.target.value))}
                className="input w-full"
              >
                <option value={0}>请选择客服</option>
                {onlineStaff.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.nickname || s.username}
                    {s.role === 'supervisor' && ' (主管)'}
                    {` - 当前负载: ${s.current_load}/${s.max_tickets}`}
                  </option>
                ))}
              </select>
            )}
          </div>
          <div>
            <label className="block text-sm text-dark-300 mb-2">转接原因（可选）</label>
            <textarea
              value={transferReason}
              onChange={(e) => setTransferReason(e.target.value)}
              placeholder="请输入转接原因..."
              className="input w-full h-24 resize-none"
            />
          </div>
          <div className="flex justify-end gap-3">
            <Button variant="secondary" onClick={() => setShowTransferModal(false)}>
              取消
            </Button>
            <Button 
              onClick={handleTransfer} 
              loading={transferring}
              disabled={!transferTarget || onlineStaff.length === 0}
            >
              确认转接
            </Button>
          </div>
        </div>
      </Modal>
    </>
  )
}
