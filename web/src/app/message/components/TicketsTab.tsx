'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Button, Badge, Card } from '@/components/ui'
import { apiGet } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'
import { SupportConfig, Ticket } from './types'
import { CreateTicketModal } from './CreateTicketModal'
import { TicketDetailModal } from './TicketDetailModal'

/**
 * 工单中心标签页
 */
export function TicketsTab({
  config,
  isLoggedIn,
  guestToken,
  setGuestToken,
}: {
  config: SupportConfig
  isLoggedIn: boolean
  guestToken: string
  setGuestToken: (token: string) => void
}) {
  const [tickets, setTickets] = useState<Ticket[]>([])
  const [loading, setLoading] = useState(false)
  const [showCreate, setShowCreate] = useState(false)
  const [selectedTicket, setSelectedTicket] = useState<Ticket | null>(null)

  // 加载工单列表
  const loadTickets = async () => {
    setLoading(true)
    let res
    if (isLoggedIn) {
      res = await apiGet<{ tickets: Ticket[] }>('/api/support/tickets')
    } else if (guestToken) {
      res = await apiGet<{ tickets: Ticket[] }>(`/api/support/tickets/guest?guest_token=${guestToken}`)
    }
    if (res?.success && res.tickets) {
      setTickets(res.tickets)
    }
    setLoading(false)
  }

  useEffect(() => {
    if (isLoggedIn || guestToken) {
      loadTickets()
    }
  }, [isLoggedIn, guestToken])

  // 获取状态徽章
  const getStatusBadge = (status: number) => {
    const statusMap: Record<number, { text: string; variant: 'warning' | 'info' | 'success' | 'default' }> = {
      0: { text: '待处理', variant: 'warning' },
      1: { text: '处理中', variant: 'info' },
      2: { text: '已回复', variant: 'success' },
      3: { text: '已解决', variant: 'success' },
      4: { text: '已关闭', variant: 'default' },
    }
    const s = statusMap[status] || { text: '未知', variant: 'default' as const }
    return <Badge variant={s.variant}>{s.text}</Badge>
  }

  // 获取优先级文本
  const getPriorityText = (priority: number) => {
    const map: Record<number, string> = { 1: '普通', 2: '紧急', 3: '非常紧急' }
    return map[priority] || '普通'
  }

  return (
    <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
      {/* 创建工单按钮 */}
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold text-dark-100">我的工单</h3>
        <Button onClick={() => setShowCreate(true)}>
          <i className="fas fa-plus mr-2" />
          提交工单
        </Button>
      </div>

      {/* 工单列表 */}
      <Card>
        {loading ? (
          <div className="text-center py-8">
            <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
          </div>
        ) : tickets.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-5xl mb-4">📋</div>
            <p className="text-dark-400 mb-4">暂无工单</p>
            {!isLoggedIn && !guestToken && (
              <p className="text-dark-500 text-sm">提交工单后会生成访问令牌，请妥善保存</p>
            )}
          </div>
        ) : (
          <div className="space-y-3">
            {tickets.map((ticket) => (
              <div
                key={ticket.id}
                onClick={() => setSelectedTicket(ticket)}
                className="bg-dark-700/30 rounded-xl p-4 border border-dark-600/50 cursor-pointer hover:border-primary-500/50 transition-colors"
              >
                <div className="flex justify-between items-start mb-2">
                  <div>
                    <span className="text-dark-500 text-sm font-mono mr-2">
                      #{ticket.ticket_no}
                    </span>
                    {getStatusBadge(ticket.status)}
                  </div>
                  <span className="text-dark-500 text-sm">
                    {formatDateTime(ticket.created_at)}
                  </span>
                </div>
                <h4 className="text-dark-100 font-medium mb-1">{ticket.subject}</h4>
                <div className="text-dark-400 text-sm">
                  分类: {ticket.category} | 优先级: {getPriorityText(ticket.priority)}
                  {ticket.last_reply_at && (
                    <span className="ml-2">
                      | 最后回复: {ticket.last_reply_by} ({formatDateTime(ticket.last_reply_at)})
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>

      {/* 创建工单弹窗 */}
      <CreateTicketModal
        isOpen={showCreate}
        onClose={() => setShowCreate(false)}
        config={config}
        isLoggedIn={isLoggedIn}
        guestToken={guestToken}
        setGuestToken={setGuestToken}
        onSuccess={() => {
          loadTickets()
          setShowCreate(false)
        }}
      />

      {/* 工单详情弹窗 */}
      {selectedTicket && (
        <TicketDetailModal
          isOpen={!!selectedTicket}
          onClose={() => setSelectedTicket(null)}
          ticket={selectedTicket}
          guestToken={guestToken}
          isLoggedIn={isLoggedIn}
          onUpdate={loadTickets}
        />
      )}
    </motion.div>
  )
}
