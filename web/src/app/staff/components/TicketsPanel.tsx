'use client'

import { useState, useEffect } from 'react'
import toast from 'react-hot-toast'
import { motion } from 'framer-motion'
import { Button, Badge, Card, Modal } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime, cn } from '@/lib/utils'
import { StaffInfo, Ticket } from './types'
import { StaffTicketModal } from './StaffTicketModal'

/**
 * 工单管理面板
 */
export function TicketsPanel({ staff }: { staff: StaffInfo }) {
  const [tickets, setTickets] = useState<Ticket[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState({ status: -1, myOnly: false })
  const [selectedTicket, setSelectedTicket] = useState<Ticket | null>(null)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  
  // 合并功能相关状态
  const [mergeMode, setMergeMode] = useState(false)
  const [selectedTickets, setSelectedTickets] = useState<Set<string>>(new Set())
  const [showMergeModal, setShowMergeModal] = useState(false)
  const [targetTicketNo, setTargetTicketNo] = useState('')
  const [merging, setMerging] = useState(false)

  // 加载工单列表
  const loadTickets = async () => {
    setLoading(true)
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: '20',
      status: filter.status.toString(),
      my_only: filter.myOnly.toString(),
    })
    const res = await apiGet<{ tickets: Ticket[]; total: number }>(
      `/api/staff/tickets?${params}`
    )
    if (res.success) {
      setTickets(res.tickets || [])
      setTotal(res.total || 0)
    }
    setLoading(false)
  }

  useEffect(() => {
    loadTickets()
  }, [page, filter])

  // 获取状态徽章
  const getStatusBadge = (status: number) => {
    const map: Record<number, { text: string; variant: 'warning' | 'info' | 'success' | 'default' }> = {
      0: { text: '待处理', variant: 'warning' },
      1: { text: '处理中', variant: 'info' },
      2: { text: '已回复', variant: 'success' },
      3: { text: '已解决', variant: 'success' },
      4: { text: '已关闭', variant: 'default' },
      5: { text: '已合并', variant: 'default' },
    }
    const s = map[status] || { text: '未知', variant: 'default' as const }
    return <Badge variant={s.variant}>{s.text}</Badge>
  }

  // 获取优先级徽章
  const getPriorityBadge = (priority: number) => {
    const map: Record<number, { text: string; variant: 'default' | 'warning' | 'danger' }> = {
      1: { text: '普通', variant: 'default' },
      2: { text: '紧急', variant: 'warning' },
      3: { text: '非常紧急', variant: 'danger' },
    }
    const p = map[priority] || { text: '普通', variant: 'default' as const }
    return <Badge variant={p.variant}>{p.text}</Badge>
  }

  // 切换工单选择
  const toggleTicketSelection = (ticketNo: string) => {
    const newSelected = new Set(selectedTickets)
    if (newSelected.has(ticketNo)) {
      newSelected.delete(ticketNo)
    } else {
      newSelected.add(ticketNo)
    }
    setSelectedTickets(newSelected)
  }

  // 全选/取消全选
  const toggleSelectAll = () => {
    if (selectedTickets.size === tickets.length) {
      setSelectedTickets(new Set())
    } else {
      setSelectedTickets(new Set(tickets.map(t => t.ticket_no)))
    }
  }

  // 打开合并弹窗
  const openMergeModal = () => {
    if (selectedTickets.size < 2) {
      toast.error('请至少选择2个工单进行合并')
      return
    }
    // 默认选择第一个作为目标工单
    const firstSelected = Array.from(selectedTickets)[0]
    setTargetTicketNo(firstSelected)
    setShowMergeModal(true)
  }

  // 执行合并
  const handleMerge = async () => {
    if (!targetTicketNo) {
      toast.error('请选择目标工单')
      return
    }

    const sourceTicketNos = Array.from(selectedTickets).filter(no => no !== targetTicketNo)
    if (sourceTicketNos.length === 0) {
      toast.error('请选择要合并的源工单')
      return
    }

    setMerging(true)
    const res = await apiPost<{ merged_count: number }>('/api/staff/tickets/merge', {
      target_ticket_no: targetTicketNo,
      source_ticket_nos: sourceTicketNos,
    })

    if (res.success) {
      toast.success(`成功合并 ${res.merged_count} 个工单`)
      setShowMergeModal(false)
      setMergeMode(false)
      setSelectedTickets(new Set())
      loadTickets()
    } else {
      toast.error(res.error || '合并失败')
    }
    setMerging(false)
  }

  // 退出合并模式
  const exitMergeMode = () => {
    setMergeMode(false)
    setSelectedTickets(new Set())
  }

  return (
    <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
      {/* 筛选栏 */}
      <div className="flex flex-wrap items-center gap-4 mb-4">
        <select
          value={filter.status}
          onChange={(e) => setFilter({ ...filter, status: Number(e.target.value) })}
          className="input w-40"
        >
          <option value={-1}>全部状态</option>
          <option value={0}>待处理</option>
          <option value={1}>处理中</option>
          <option value={2}>已回复</option>
          <option value={3}>已解决</option>
          <option value={4}>已关闭</option>
          <option value={5}>已合并</option>
        </select>
        <label className="flex items-center gap-2 text-dark-300 cursor-pointer p-2 bg-dark-700/30 rounded-lg hover:bg-dark-700/50 transition-colors">
          <div
            className={`relative w-10 h-5 rounded-full transition-colors ${
              filter.myOnly ? 'bg-primary-500' : 'bg-dark-600'
            }`}
            onClick={() => setFilter({ ...filter, myOnly: !filter.myOnly })}
          >
            <div
              className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform ${
                filter.myOnly ? 'translate-x-5' : 'translate-x-0.5'
              }`}
            />
          </div>
          <span>只看我的</span>
        </label>
        <Button size="sm" variant="secondary" onClick={loadTickets}>
          <i className="fas fa-sync-alt mr-2" />
          刷新
        </Button>
        
        {/* 合并模式按钮 */}
        <div className="ml-auto flex items-center gap-2">
          {mergeMode ? (
            <>
              <span className="text-dark-400 text-sm">
                已选择 {selectedTickets.size} 个工单
              </span>
              <Button 
                size="sm" 
                onClick={openMergeModal}
                disabled={selectedTickets.size < 2}
              >
                <i className="fas fa-compress-arrows-alt mr-2" />
                合并选中
              </Button>
              <Button size="sm" variant="secondary" onClick={exitMergeMode}>
                取消
              </Button>
            </>
          ) : (
            <Button size="sm" variant="secondary" onClick={() => setMergeMode(true)}>
              <i className="fas fa-compress-arrows-alt mr-2" />
              合并工单
            </Button>
          )}
        </div>
      </div>

      {/* 工单列表 */}
      <Card>
        {loading ? (
          <div className="text-center py-8">
            <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
          </div>
        ) : tickets.length === 0 ? (
          <div className="text-center py-12 text-dark-400">暂无工单</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  {mergeMode && (
                    <th className="pb-3 pr-2 w-10">
                      <input
                        type="checkbox"
                        checked={selectedTickets.size === tickets.length && tickets.length > 0}
                        onChange={toggleSelectAll}
                        className="w-4 h-4 rounded border-dark-600 bg-dark-700 text-primary-500 focus:ring-primary-500"
                      />
                    </th>
                  )}
                  <th className="pb-3 pr-4">工单号</th>
                  <th className="pb-3 pr-4">用户</th>
                  <th className="pb-3 pr-4">主题</th>
                  <th className="pb-3 pr-4">分类</th>
                  <th className="pb-3 pr-4">优先级</th>
                  <th className="pb-3 pr-4">状态</th>
                  <th className="pb-3 pr-4">处理人</th>
                  <th className="pb-3">创建时间</th>
                </tr>
              </thead>
              <tbody>
                {tickets.map((ticket) => (
                  <tr
                    key={ticket.id}
                    onClick={() => {
                      if (mergeMode) {
                        toggleTicketSelection(ticket.ticket_no)
                      } else {
                        setSelectedTicket(ticket)
                      }
                    }}
                    className={cn(
                      'border-b border-dark-700/50 hover:bg-dark-700/30 cursor-pointer',
                      mergeMode && selectedTickets.has(ticket.ticket_no) && 'bg-primary-500/10'
                    )}
                  >
                    {mergeMode && (
                      <td className="py-3 pr-2">
                        <input
                          type="checkbox"
                          checked={selectedTickets.has(ticket.ticket_no)}
                          onChange={() => toggleTicketSelection(ticket.ticket_no)}
                          onClick={(e) => e.stopPropagation()}
                          className="w-4 h-4 rounded border-dark-600 bg-dark-700 text-primary-500 focus:ring-primary-500"
                        />
                      </td>
                    )}
                    <td className="py-3 pr-4 font-mono text-sm text-primary-400">
                      #{ticket.ticket_no}
                    </td>
                    <td className="py-3 pr-4 text-dark-200">{ticket.username}</td>
                    <td className="py-3 pr-4 text-dark-100 max-w-xs truncate">
                      {ticket.subject}
                    </td>
                    <td className="py-3 pr-4 text-dark-300">{ticket.category}</td>
                    <td className="py-3 pr-4">{getPriorityBadge(ticket.priority)}</td>
                    <td className="py-3 pr-4">{getStatusBadge(ticket.status)}</td>
                    <td className="py-3 pr-4 text-dark-300">
                      {ticket.assigned_name || '-'}
                    </td>
                    <td className="py-3 text-dark-400 text-sm">
                      {formatDateTime(ticket.created_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* 分页 */}
        {total > 20 && (
          <div className="flex justify-center gap-2 mt-4 pt-4 border-t border-dark-700">
            <Button
              size="sm"
              variant="secondary"
              disabled={page === 1}
              onClick={() => setPage(page - 1)}
            >
              上一页
            </Button>
            <span className="text-dark-400 py-1">
              第 {page} 页 / 共 {Math.ceil(total / 20)} 页
            </span>
            <Button
              size="sm"
              variant="secondary"
              disabled={page >= Math.ceil(total / 20)}
              onClick={() => setPage(page + 1)}
            >
              下一页
            </Button>
          </div>
        )}
      </Card>

      {/* 工单详情弹窗 */}
      {selectedTicket && (
        <StaffTicketModal
          isOpen={!!selectedTicket}
          onClose={() => setSelectedTicket(null)}
          ticket={selectedTicket}
          staff={staff}
          onUpdate={loadTickets}
        />
      )}

      {/* 合并工单弹窗 */}
      <Modal
        isOpen={showMergeModal}
        onClose={() => setShowMergeModal(false)}
        title="合并工单"
        size="sm"
      >
        <div className="space-y-4">
          <div className="bg-amber-500/10 border border-amber-500/30 rounded-lg p-3 text-amber-400 text-sm">
            <i className="fas fa-exclamation-triangle mr-2" />
            合并后，源工单的所有消息将迁移到目标工单，源工单将被标记为"已合并"状态。此操作不可撤销。
          </div>
          
          <div>
            <label className="block text-sm text-dark-300 mb-2">选择目标工单（保留的工单）</label>
            <select
              value={targetTicketNo}
              onChange={(e) => setTargetTicketNo(e.target.value)}
              className="input w-full"
            >
              {Array.from(selectedTickets).map((ticketNo) => {
                const ticket = tickets.find(t => t.ticket_no === ticketNo)
                return (
                  <option key={ticketNo} value={ticketNo}>
                    #{ticketNo} - {ticket?.subject || ''}
                  </option>
                )
              })}
            </select>
          </div>

          <div>
            <label className="block text-sm text-dark-300 mb-2">将被合并的工单</label>
            <div className="space-y-1 max-h-32 overflow-y-auto">
              {Array.from(selectedTickets)
                .filter(no => no !== targetTicketNo)
                .map((ticketNo) => {
                  const ticket = tickets.find(t => t.ticket_no === ticketNo)
                  return (
                    <div key={ticketNo} className="text-dark-400 text-sm py-1 px-2 bg-dark-700/30 rounded">
                      #{ticketNo} - {ticket?.subject || ''}
                    </div>
                  )
                })}
            </div>
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <Button variant="secondary" onClick={() => setShowMergeModal(false)}>
              取消
            </Button>
            <Button onClick={handleMerge} loading={merging}>
              确认合并
            </Button>
          </div>
        </div>
      </Modal>
    </motion.div>
  )
}
