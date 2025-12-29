'use client'

import { useState, useEffect, useCallback } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Modal, Badge, Card } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * 账户注销申请接口
 */
interface AccountDeletion {
  id: number
  user_id: number
  username: string
  email: string
  reason: string
  status: number  // 0: 待处理, 1: 已批准, 2: 已拒绝, 3: 已完成
  reject_reason: string
  scheduled_at: string
  completed_at: string
  created_at: string
}

/**
 * 账户注销管理页面
 */
export function DeletionsPage() {
  const [deletions, setDeletions] = useState<AccountDeletion[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [statusFilter, setStatusFilter] = useState<number | ''>('')
  const [showRejectModal, setShowRejectModal] = useState(false)
  const [showDetailModal, setShowDetailModal] = useState(false)
  const [selectedDeletion, setSelectedDeletion] = useState<AccountDeletion | null>(null)
  const [rejectReason, setRejectReason] = useState('')
  // 确认弹窗状态
  const [showApproveConfirm, setShowApproveConfirm] = useState(false)
  const [showExecuteConfirm, setShowExecuteConfirm] = useState(false)
  const [confirmTarget, setConfirmTarget] = useState<AccountDeletion | null>(null)
  const [confirmLoading, setConfirmLoading] = useState(false)
  const pageSize = 20

  // 加载注销申请列表
  const loadDeletions = useCallback(async () => {
    setLoading(true)
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    })
    if (statusFilter !== '') {
      params.append('status', statusFilter.toString())
    }

    const res = await apiGet<{ deletions: AccountDeletion[]; total: number }>(
      `/api/admin/account/deletions?${params}`
    )
    if (res.success) {
      setDeletions(res.deletions || [])
      setTotal(res.total || 0)
    }
    setLoading(false)
  }, [page, statusFilter])

  useEffect(() => {
    loadDeletions()
  }, [loadDeletions])

  // 打开批准确认弹窗
  const openApproveConfirm = (deletion: AccountDeletion) => {
    setConfirmTarget(deletion)
    setShowApproveConfirm(true)
  }

  // 批准注销
  const handleApprove = async () => {
    if (!confirmTarget) return
    setConfirmLoading(true)
    const res = await apiPost(`/api/admin/account/deletion/${confirmTarget.id}/approve`, {})
    setConfirmLoading(false)
    if (res.success) {
      toast.success('已批准注销申请')
      setShowApproveConfirm(false)
      setConfirmTarget(null)
      loadDeletions()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 拒绝注销
  const handleReject = async () => {
    if (!selectedDeletion) return
    if (!rejectReason.trim()) {
      toast.error('请输入拒绝原因')
      return
    }
    const res = await apiPost(`/api/admin/account/deletion/${selectedDeletion.id}/reject`, {
      reason: rejectReason,
    })
    if (res.success) {
      toast.success('已拒绝注销申请')
      setShowRejectModal(false)
      setSelectedDeletion(null)
      setRejectReason('')
      loadDeletions()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 打开执行确认弹窗
  const openExecuteConfirm = (deletion: AccountDeletion) => {
    setConfirmTarget(deletion)
    setShowExecuteConfirm(true)
  }

  // 立即执行注销
  const handleExecute = async () => {
    if (!confirmTarget) return
    setConfirmLoading(true)
    const res = await apiPost(`/api/admin/account/deletion/${confirmTarget.id}/execute`, {})
    setConfirmLoading(false)
    if (res.success) {
      toast.success('账户已注销')
      setShowExecuteConfirm(false)
      setConfirmTarget(null)
      loadDeletions()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 获取状态标签
  const getStatusBadge = (status: number) => {
    switch (status) {
      case 0:
        return <Badge variant="warning">待处理</Badge>
      case 1:
        return <Badge variant="info">已批准</Badge>
      case 2:
        return <Badge variant="danger">已拒绝</Badge>
      case 3:
        return <Badge variant="success">已完成</Badge>
      default:
        return <Badge variant="default">未知</Badge>
    }
  }

  // 计算剩余冷静期
  const getRemainingDays = (scheduledAt: string) => {
    if (!scheduledAt) return '-'
    const diff = new Date(scheduledAt).getTime() - Date.now()
    if (diff <= 0) return '可执行'
    const days = Math.ceil(diff / (24 * 60 * 60 * 1000))
    return `${days} 天`
  }

  const totalPages = Math.ceil(total / pageSize)

  // 统计数据
  const stats = {
    pending: deletions.filter(d => d.status === 0).length,
    approved: deletions.filter(d => d.status === 1).length,
    completed: deletions.filter(d => d.status === 3).length,
  }

  return (
    <div className="space-y-6">
      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="card p-4">
          <div className="text-dark-400 text-sm mb-1">待处理</div>
          <div className="text-2xl font-bold text-amber-400">{stats.pending}</div>
        </div>
        <div className="card p-4">
          <div className="text-dark-400 text-sm mb-1">已批准（冷静期）</div>
          <div className="text-2xl font-bold text-blue-400">{stats.approved}</div>
        </div>
        <div className="card p-4">
          <div className="text-dark-400 text-sm mb-1">已完成</div>
          <div className="text-2xl font-bold text-emerald-400">{stats.completed}</div>
        </div>
        <div className="card p-4">
          <div className="text-dark-400 text-sm mb-1">总申请数</div>
          <div className="text-2xl font-bold text-dark-100">{total}</div>
        </div>
      </div>

      {/* 筛选 */}
      <div className="flex items-center gap-4">
        <select
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value === '' ? '' : Number(e.target.value))
            setPage(1)
          }}
          className="input w-40"
        >
          <option value="">全部状态</option>
          <option value="0">待处理</option>
          <option value="1">已批准</option>
          <option value="2">已拒绝</option>
          <option value="3">已完成</option>
        </select>
      </div>

      {/* 注销申请列表 */}
      <Card title="账户注销申请" icon={<i className="fas fa-user-slash" />}>
        {loading ? (
          <div className="p-8 text-center">
            <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
          </div>
        ) : deletions.length === 0 ? (
          <div className="p-8 text-center text-dark-400">
            <i className="fas fa-user-check text-4xl mb-4 opacity-50" />
            <p>暂无注销申请</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">用户</th>
                  <th className="pb-3 font-medium">邮箱</th>
                  <th className="pb-3 font-medium">注销原因</th>
                  <th className="pb-3 font-medium">申请时间</th>
                  <th className="pb-3 font-medium">冷静期剩余</th>
                  <th className="pb-3 font-medium">状态</th>
                  <th className="pb-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {deletions.map((deletion) => (
                  <motion.tr
                    key={deletion.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    className="border-b border-dark-700/50"
                  >
                    <td className="py-3">{deletion.username}</td>
                    <td className="py-3 text-dark-400">{deletion.email}</td>
                    <td className="py-3 text-dark-400 max-w-xs truncate">
                      {deletion.reason || '-'}
                    </td>
                    <td className="py-3 text-sm text-dark-400">
                      {formatDateTime(deletion.created_at)}
                    </td>
                    <td className="py-3">
                      {deletion.status === 1 ? (
                        <span className="text-yellow-400">
                          {getRemainingDays(deletion.scheduled_at)}
                        </span>
                      ) : (
                        <span className="text-dark-500">-</span>
                      )}
                    </td>
                    <td className="py-3">{getStatusBadge(deletion.status)}</td>
                    <td className="py-3">
                      <div className="flex gap-1">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => {
                            setSelectedDeletion(deletion)
                            setShowDetailModal(true)
                          }}
                        >
                          <i className="fas fa-eye" />
                        </Button>
                        {deletion.status === 0 && (
                          <>
                            <Button
                              size="sm"
                              variant="primary"
                              onClick={() => openApproveConfirm(deletion)}
                            >
                              批准
                            </Button>
                            <Button
                              size="sm"
                              variant="danger"
                              onClick={() => {
                                setSelectedDeletion(deletion)
                                setRejectReason('')
                                setShowRejectModal(true)
                              }}
                            >
                              拒绝
                            </Button>
                          </>
                        )}
                        {deletion.status === 1 && getRemainingDays(deletion.scheduled_at) === '可执行' && (
                          <Button
                            size="sm"
                            variant="danger"
                            onClick={() => openExecuteConfirm(deletion)}
                          >
                            执行注销
                          </Button>
                        )}
                      </div>
                    </td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* 分页 */}
        {totalPages > 1 && (
          <div className="flex justify-center gap-2 mt-4">
            <Button size="sm" variant="ghost" disabled={page === 1} onClick={() => setPage(p => p - 1)}>
              上一页
            </Button>
            <span className="px-4 py-2 text-dark-400">{page} / {totalPages}</span>
            <Button size="sm" variant="ghost" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
              下一页
            </Button>
          </div>
        )}
      </Card>

      {/* 拒绝弹窗 */}
      <Modal
        isOpen={showRejectModal}
        onClose={() => {
          setShowRejectModal(false)
          setSelectedDeletion(null)
          setRejectReason('')
        }}
        title="拒绝注销申请"
      >
        <div className="space-y-4">
          <div className="p-3 bg-dark-700/30 rounded-lg">
            <div className="text-dark-400 text-sm">用户</div>
            <div className="text-dark-200">{selectedDeletion?.username}</div>
          </div>
          <div>
            <label className="block text-dark-300 text-sm mb-2">拒绝原因 *</label>
            <textarea
              value={rejectReason}
              onChange={(e) => setRejectReason(e.target.value)}
              className="input w-full h-24 resize-none"
              placeholder="请输入拒绝原因，将通知用户"
            />
          </div>
          <div className="flex gap-3">
            <Button
              variant="secondary"
              className="flex-1"
              onClick={() => setShowRejectModal(false)}
            >
              取消
            </Button>
            <Button variant="danger" className="flex-1" onClick={handleReject}>
              确认拒绝
            </Button>
          </div>
        </div>
      </Modal>

      {/* 详情弹窗 */}
      <Modal
        isOpen={showDetailModal}
        onClose={() => {
          setShowDetailModal(false)
          setSelectedDeletion(null)
        }}
        title="注销申请详情"
      >
        {selectedDeletion && (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-dark-400 text-sm">用户名</div>
                <div className="text-dark-200">{selectedDeletion.username}</div>
              </div>
              <div>
                <div className="text-dark-400 text-sm">邮箱</div>
                <div className="text-dark-200">{selectedDeletion.email}</div>
              </div>
              <div>
                <div className="text-dark-400 text-sm">申请时间</div>
                <div className="text-dark-200">{formatDateTime(selectedDeletion.created_at)}</div>
              </div>
              <div>
                <div className="text-dark-400 text-sm">状态</div>
                <div>{getStatusBadge(selectedDeletion.status)}</div>
              </div>
            </div>

            <div>
              <div className="text-dark-400 text-sm mb-1">注销原因</div>
              <div className="p-3 bg-dark-700/30 rounded-lg text-dark-200">
                {selectedDeletion.reason || '用户未填写原因'}
              </div>
            </div>

            {selectedDeletion.status === 1 && (
              <div className="p-3 bg-blue-500/10 border border-blue-500/30 rounded-lg">
                <div className="text-blue-400 text-sm">
                  <i className="fas fa-clock mr-2" />
                  计划执行时间：{formatDateTime(selectedDeletion.scheduled_at)}
                </div>
              </div>
            )}

            {selectedDeletion.status === 2 && selectedDeletion.reject_reason && (
              <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
                <div className="text-red-400 text-sm mb-1">
                  <i className="fas fa-times-circle mr-2" />
                  拒绝原因
                </div>
                <div className="text-dark-300">{selectedDeletion.reject_reason}</div>
              </div>
            )}

            {selectedDeletion.status === 3 && (
              <div className="p-3 bg-green-500/10 border border-green-500/30 rounded-lg">
                <div className="text-green-400 text-sm">
                  <i className="fas fa-check-circle mr-2" />
                  账户已于 {formatDateTime(selectedDeletion.completed_at)} 完成注销
                </div>
              </div>
            )}
          </div>
        )}
      </Modal>

      {/* 批准确认弹窗 */}
      <ConfirmModal
        isOpen={showApproveConfirm}
        onClose={() => { setShowApproveConfirm(false); setConfirmTarget(null) }}
        title="批准注销申请"
        message={`确定要批准用户 "${confirmTarget?.username}" 的注销申请吗？用户账户将在冷静期后被永久删除。`}
        confirmText="批准"
        variant="warning"
        onConfirm={handleApprove}
        loading={confirmLoading}
      />

      {/* 执行确认弹窗 */}
      <ConfirmModal
        isOpen={showExecuteConfirm}
        onClose={() => { setShowExecuteConfirm(false); setConfirmTarget(null) }}
        title="立即执行注销"
        message={`确定要立即执行用户 "${confirmTarget?.username}" 的账户注销吗？此操作不可恢复！`}
        confirmText="执行注销"
        variant="danger"
        onConfirm={handleExecute}
        loading={confirmLoading}
      />
    </div>
  )
}
