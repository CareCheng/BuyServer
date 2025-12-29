'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Card, Badge, Modal } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * 账户注销状态接口
 */
interface DeletionStatus {
  has_request: boolean
  status: string
  scheduled_at: string
  reason: string
  created_at: string
}

/**
 * 账户设置标签页（账户注销）
 */
export function AccountTab() {
  const [deletionStatus, setDeletionStatus] = useState<DeletionStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [showCancelModal, setShowCancelModal] = useState(false)
  const [deleteReason, setDeleteReason] = useState('')
  const [confirmText, setConfirmText] = useState('')

  // 加载注销状态
  const loadDeletionStatus = async () => {
    setLoading(true)
    const res = await apiGet<{ data: { has_request: boolean; request: DeletionStatus | null } }>('/api/user/account/delete/status')
    if (res.success && res.data) {
      if (res.data.has_request && res.data.request) {
        setDeletionStatus({
          has_request: true,
          status: res.data.request.status,
          scheduled_at: res.data.request.scheduled_at,
          reason: res.data.request.reason,
          created_at: res.data.request.created_at,
        })
      } else {
        setDeletionStatus({ has_request: false, status: '', scheduled_at: '', reason: '', created_at: '' })
      }
    }
    setLoading(false)
  }

  useEffect(() => {
    loadDeletionStatus()
  }, [])

  // 申请注销
  const handleRequestDeletion = async () => {
    if (confirmText !== '确认注销') {
      toast.error('请输入"确认注销"以继续')
      return
    }
    const res = await apiPost('/api/user/account/delete', { reason: deleteReason })
    if (res.success) {
      toast.success('注销申请已提交，将在7天后执行')
      setShowDeleteModal(false)
      setDeleteReason('')
      setConfirmText('')
      loadDeletionStatus()
    } else {
      toast.error(res.error || '申请失败')
    }
  }

  // 取消注销
  const handleCancelDeletion = async () => {
    const res = await apiPost('/api/user/account/delete/cancel', {})
    if (res.success) {
      toast.success('已取消注销申请')
      setShowCancelModal(false)
      loadDeletionStatus()
    } else {
      toast.error(res.error || '取消失败')
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
      {/* 账户注销 */}
      <Card title="账户注销" icon={<i className="fas fa-user-times" />}>
        {deletionStatus?.has_request ? (
          <div className="space-y-4">
            <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-xl">
              <div className="flex items-center gap-2 mb-2">
                <i className="fas fa-exclamation-triangle text-red-400" />
                <span className="font-medium text-red-400">账户注销申请中</span>
                <Badge variant="danger">{deletionStatus.status === 'pending' ? '待执行' : '处理中'}</Badge>
              </div>
              <div className="text-sm text-dark-400 space-y-1">
                <div>申请时间: {formatDateTime(deletionStatus.created_at)}</div>
                <div>计划执行时间: {formatDateTime(deletionStatus.scheduled_at)}</div>
                {deletionStatus.reason && <div>注销原因: {deletionStatus.reason}</div>}
              </div>
              <p className="text-sm text-dark-500 mt-3">
                在执行时间之前，您可以随时取消注销申请。一旦执行，您的账户和所有数据将被永久删除，无法恢复。
              </p>
            </div>
            <Button variant="secondary" onClick={() => setShowCancelModal(true)}>
              取消注销申请
            </Button>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="p-4 bg-dark-700/30 rounded-xl">
              <h4 className="font-medium text-dark-100 mb-2">注销账户须知</h4>
              <ul className="text-sm text-dark-400 space-y-2">
                <li className="flex items-start gap-2">
                  <i className="fas fa-check text-green-400 mt-1" />
                  <span>注销申请提交后，将有7天的冷静期</span>
                </li>
                <li className="flex items-start gap-2">
                  <i className="fas fa-check text-green-400 mt-1" />
                  <span>冷静期内您可以随时取消注销申请</span>
                </li>
                <li className="flex items-start gap-2">
                  <i className="fas fa-exclamation text-yellow-400 mt-1" />
                  <span>注销后，您的账户信息、订单记录、卡密等数据将被永久删除</span>
                </li>
                <li className="flex items-start gap-2">
                  <i className="fas fa-exclamation text-yellow-400 mt-1" />
                  <span>账户余额和积分将被清零，无法退还</span>
                </li>
                <li className="flex items-start gap-2">
                  <i className="fas fa-times text-red-400 mt-1" />
                  <span>注销操作不可逆，请谨慎操作</span>
                </li>
              </ul>
            </div>
            <Button variant="danger" onClick={() => setShowDeleteModal(true)}>
              <i className="fas fa-user-times mr-2" />
              申请注销账户
            </Button>
          </div>
        )}
      </Card>

      {/* 申请注销弹窗 */}
      <Modal isOpen={showDeleteModal} onClose={() => setShowDeleteModal(false)} title="申请注销账户" size="md">
        <div className="space-y-4">
          <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-xl">
            <div className="flex items-center gap-2 text-red-400 mb-2">
              <i className="fas fa-exclamation-triangle" />
              <span className="font-medium">警告：此操作不可逆</span>
            </div>
            <p className="text-sm text-dark-400">
              注销账户后，您的所有数据将被永久删除，包括：个人信息、订单记录、卡密、余额、积分等。
            </p>
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">注销原因（可选）</label>
            <textarea
              className="w-full px-4 py-3 bg-dark-700/50 border border-dark-600 rounded-xl text-dark-100 placeholder-dark-500 focus:outline-none focus:border-primary-500"
              rows={3}
              placeholder="请告诉我们您注销的原因，帮助我们改进服务"
              value={deleteReason}
              onChange={(e) => setDeleteReason(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              请输入 <span className="text-red-400">确认注销</span> 以继续
            </label>
            <input
              type="text"
              className="w-full px-4 py-3 bg-dark-700/50 border border-dark-600 rounded-xl text-dark-100 placeholder-dark-500 focus:outline-none focus:border-primary-500"
              placeholder="确认注销"
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
            />
          </div>
          <div className="flex gap-3">
            <Button variant="secondary" className="flex-1" onClick={() => setShowDeleteModal(false)}>
              取消
            </Button>
            <Button
              variant="danger"
              className="flex-1"
              onClick={handleRequestDeletion}
              disabled={confirmText !== '确认注销'}
            >
              确认申请注销
            </Button>
          </div>
        </div>
      </Modal>

      {/* 取消注销弹窗 */}
      <Modal isOpen={showCancelModal} onClose={() => setShowCancelModal(false)} title="取消注销申请" size="sm">
        <div className="space-y-4">
          <p className="text-dark-400">确定要取消账户注销申请吗？取消后您的账户将恢复正常使用。</p>
          <div className="flex gap-3">
            <Button variant="secondary" className="flex-1" onClick={() => setShowCancelModal(false)}>
              返回
            </Button>
            <Button className="flex-1" onClick={handleCancelDeletion}>
              确认取消
            </Button>
          </div>
        </div>
      </Modal>
    </motion.div>
  )
}
