'use client'

import { useState, useEffect, useCallback } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Modal, Badge, Card } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import Toggle from '@/components/common/Toggle'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * 可撤销操作接口
 */
interface UndoOperation {
  id: number
  operation_type: string
  target_type: string
  target_id: number
  target_name: string
  operator_id: number
  operator_name: string
  original_data: string
  can_undo: boolean
  undo_deadline: string
  is_undone: boolean
  undone_at: string
  created_at: string
}

/**
 * 撤销配置接口
 */
interface UndoConfig {
  enabled: boolean
  undo_window_minutes: number
  supported_operations: string[]
}

/**
 * 操作撤销管理页面
 */
export function UndoPage() {
  const [operations, setOperations] = useState<UndoOperation[]>([])
  const [config, setConfig] = useState<UndoConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [showDetailModal, setShowDetailModal] = useState(false)
  const [selectedOperation, setSelectedOperation] = useState<UndoOperation | null>(null)
  // 确认弹窗状态
  const [showUndoConfirm, setShowUndoConfirm] = useState(false)
  const [undoTarget, setUndoTarget] = useState<UndoOperation | null>(null)
  const [undoLoading, setUndoLoading] = useState(false)
  const [configForm, setConfigForm] = useState({
    enabled: false,
    undo_window_minutes: 30,
  })
  const pageSize = 20

  // 加载操作列表
  const loadOperations = useCallback(async () => {
    setLoading(true)
    const res = await apiGet<{ operations: UndoOperation[]; total: number }>(
      `/api/admin/undo/operations?page=${page}&page_size=${pageSize}`
    )
    if (res.success) {
      setOperations(res.operations || [])
      setTotal(res.total || 0)
    }
    setLoading(false)
  }, [page])

  // 加载配置
  const loadConfig = useCallback(async () => {
    const res = await apiGet<{ config: UndoConfig }>('/api/admin/undo/config')
    if (res.success && res.config) {
      setConfig(res.config)
      setConfigForm({
        enabled: res.config.enabled,
        undo_window_minutes: res.config.undo_window_minutes,
      })
    }
  }, [])

  useEffect(() => {
    loadOperations()
    loadConfig()
  }, [loadOperations, loadConfig])

  // 打开撤销确认弹窗
  const openUndoConfirm = (operation: UndoOperation) => {
    setUndoTarget(operation)
    setShowUndoConfirm(true)
  }

  // 撤销操作
  const handleUndo = async () => {
    if (!undoTarget) return
    setUndoLoading(true)
    const res = await apiPost(`/api/admin/undo/${undoTarget.id}`, {})
    setUndoLoading(false)
    if (res.success) {
      toast.success('操作已撤销')
      setShowUndoConfirm(false)
      setUndoTarget(null)
      loadOperations()
    } else {
      toast.error(res.error || '撤销失败')
    }
  }

  // 保存配置
  const handleSaveConfig = async () => {
    const res = await apiPost('/api/admin/undo/config', configForm)
    if (res.success) {
      toast.success('配置已保存')
      setShowConfigModal(false)
      loadConfig()
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  // 获取操作类型标签
  const getOperationTypeLabel = (type: string) => {
    const types: Record<string, { label: string; variant: 'danger' | 'warning' | 'info' }> = {
      delete: { label: '删除', variant: 'danger' },
      disable: { label: '禁用', variant: 'warning' },
      update: { label: '修改', variant: 'info' },
    }
    return types[type] || { label: type, variant: 'info' as const }
  }

  // 获取目标类型标签
  const getTargetTypeLabel = (type: string) => {
    const types: Record<string, string> = {
      product: '商品',
      user: '用户',
      coupon: '优惠券',
      category: '分类',
      announcement: '公告',
      order: '订单',
    }
    return types[type] || type
  }

  // 检查是否可撤销
  const canUndo = (operation: UndoOperation) => {
    if (operation.is_undone || !operation.can_undo) return false
    const deadline = new Date(operation.undo_deadline)
    return deadline > new Date()
  }

  // 计算剩余时间
  const getRemainingTime = (deadline: string) => {
    const diff = new Date(deadline).getTime() - Date.now()
    if (diff <= 0) return '已过期'
    const minutes = Math.floor(diff / 60000)
    const seconds = Math.floor((diff % 60000) / 1000)
    if (minutes > 0) return `${minutes}分${seconds}秒`
    return `${seconds}秒`
  }

  const totalPages = Math.ceil(total / pageSize)

  if (loading && operations.length === 0) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 配置卡片 */}
      {config && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="card p-4">
            <div className="text-dark-400 text-sm mb-1">撤销功能</div>
            <div className={`text-lg font-bold ${config.enabled ? 'text-emerald-400' : 'text-red-400'}`}>
              {config.enabled ? '已启用' : '已禁用'}
            </div>
          </div>
          <div className="card p-4">
            <div className="text-dark-400 text-sm mb-1">撤销时间窗口</div>
            <div className="text-lg font-bold text-dark-100">{config.undo_window_minutes} 分钟</div>
          </div>
          <div className="card p-4 flex items-center justify-between">
            <div>
              <div className="text-dark-400 text-sm mb-1">支持的操作</div>
              <div className="text-lg font-bold text-dark-100">
                {config.supported_operations?.length || 0} 种
              </div>
            </div>
            <Button size="sm" variant="secondary" onClick={() => setShowConfigModal(true)}>
              <i className="fas fa-cog mr-1" />
              配置
            </Button>
          </div>
        </div>
      )}

      {/* 操作列表 */}
      <Card title="可撤销操作" icon={<i className="fas fa-undo" />}>
        {operations.length === 0 ? (
          <div className="p-8 text-center text-dark-400">
            <i className="fas fa-history text-4xl mb-4 opacity-50" />
            <p>暂无可撤销的操作记录</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">操作类型</th>
                  <th className="pb-3 font-medium">目标类型</th>
                  <th className="pb-3 font-medium">目标名称</th>
                  <th className="pb-3 font-medium">操作人</th>
                  <th className="pb-3 font-medium">操作时间</th>
                  <th className="pb-3 font-medium">剩余时间</th>
                  <th className="pb-3 font-medium">状态</th>
                  <th className="pb-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {operations.map((operation) => {
                  const opTypeInfo = getOperationTypeLabel(operation.operation_type)
                  const isUndoable = canUndo(operation)
                  return (
                    <motion.tr
                      key={operation.id}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      className="border-b border-dark-700/50"
                    >
                      <td className="py-3">
                        <Badge variant={opTypeInfo.variant}>{opTypeInfo.label}</Badge>
                      </td>
                      <td className="py-3">{getTargetTypeLabel(operation.target_type)}</td>
                      <td className="py-3">{operation.target_name}</td>
                      <td className="py-3">{operation.operator_name}</td>
                      <td className="py-3 text-sm text-dark-400">
                        {formatDateTime(operation.created_at)}
                      </td>
                      <td className="py-3">
                        {operation.is_undone ? (
                          <span className="text-dark-500">-</span>
                        ) : (
                          <span className={isUndoable ? 'text-yellow-400' : 'text-dark-500'}>
                            {getRemainingTime(operation.undo_deadline)}
                          </span>
                        )}
                      </td>
                      <td className="py-3">
                        {operation.is_undone ? (
                          <Badge variant="success">已撤销</Badge>
                        ) : isUndoable ? (
                          <Badge variant="warning">可撤销</Badge>
                        ) : (
                          <Badge variant="default">已过期</Badge>
                        )}
                      </td>
                      <td className="py-3">
                        <div className="flex gap-1">
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => {
                              setSelectedOperation(operation)
                              setShowDetailModal(true)
                            }}
                          >
                            <i className="fas fa-eye" />
                          </Button>
                          {isUndoable && (
                            <Button
                              size="sm"
                              variant="primary"
                              onClick={() => openUndoConfirm(operation)}
                            >
                              <i className="fas fa-undo mr-1" />
                              撤销
                            </Button>
                          )}
                        </div>
                      </td>
                    </motion.tr>
                  )
                })}
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

      {/* 配置弹窗 */}
      <Modal
        isOpen={showConfigModal}
        onClose={() => setShowConfigModal(false)}
        title="撤销配置"
      >
        <div className="space-y-4">
          <div className="flex items-center justify-between p-4 bg-dark-700/30 rounded-lg">
            <div>
              <div className="text-dark-200 font-medium">启用撤销功能</div>
              <div className="text-sm text-dark-400">开启后可撤销指定时间内的操作</div>
            </div>
            <Toggle
              checked={configForm.enabled}
              onChange={(checked) => setConfigForm({ ...configForm, enabled: checked })}
            />
          </div>

          <div>
            <label className="block text-dark-300 text-sm mb-2">撤销时间窗口（分钟）</label>
            <input
              type="number"
              value={configForm.undo_window_minutes}
              onChange={(e) => setConfigForm({ ...configForm, undo_window_minutes: Number(e.target.value) })}
              className="input w-full"
              min={1}
              max={1440}
            />
            <div className="mt-1 text-xs text-dark-400">
              操作后在此时间内可以撤销，建议设置 15-60 分钟
            </div>
          </div>

          {config?.supported_operations && (
            <div>
              <label className="block text-dark-300 text-sm mb-2">支持撤销的操作</label>
              <div className="flex flex-wrap gap-2">
                {config.supported_operations.map((op) => (
                  <Badge key={op} variant="info">{op}</Badge>
                ))}
              </div>
            </div>
          )}

          <div className="flex gap-3 pt-2">
            <Button
              variant="secondary"
              className="flex-1"
              onClick={() => setShowConfigModal(false)}
            >
              取消
            </Button>
            <Button variant="primary" className="flex-1" onClick={handleSaveConfig}>
              保存配置
            </Button>
          </div>
        </div>
      </Modal>

      {/* 详情弹窗 */}
      <Modal
        isOpen={showDetailModal}
        onClose={() => {
          setShowDetailModal(false)
          setSelectedOperation(null)
        }}
        title="操作详情"
      >
        {selectedOperation && (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-dark-400 text-sm">操作类型</div>
                <div className="text-dark-200">
                  {getOperationTypeLabel(selectedOperation.operation_type).label}
                </div>
              </div>
              <div>
                <div className="text-dark-400 text-sm">目标类型</div>
                <div className="text-dark-200">
                  {getTargetTypeLabel(selectedOperation.target_type)}
                </div>
              </div>
              <div>
                <div className="text-dark-400 text-sm">目标名称</div>
                <div className="text-dark-200">{selectedOperation.target_name}</div>
              </div>
              <div>
                <div className="text-dark-400 text-sm">操作人</div>
                <div className="text-dark-200">{selectedOperation.operator_name}</div>
              </div>
              <div>
                <div className="text-dark-400 text-sm">操作时间</div>
                <div className="text-dark-200">{formatDateTime(selectedOperation.created_at)}</div>
              </div>
              <div>
                <div className="text-dark-400 text-sm">撤销截止</div>
                <div className="text-dark-200">{formatDateTime(selectedOperation.undo_deadline)}</div>
              </div>
            </div>

            {selectedOperation.original_data && (
              <div>
                <div className="text-dark-400 text-sm mb-2">原始数据</div>
                <pre className="p-3 bg-dark-700/50 rounded-lg text-sm text-dark-300 overflow-x-auto max-h-48">
                  {JSON.stringify(JSON.parse(selectedOperation.original_data), null, 2)}
                </pre>
              </div>
            )}

            {selectedOperation.is_undone && (
              <div className="p-3 bg-green-500/10 border border-green-500/30 rounded-lg">
                <div className="text-green-400 text-sm">
                  <i className="fas fa-check-circle mr-2" />
                  此操作已于 {formatDateTime(selectedOperation.undone_at)} 撤销
                </div>
              </div>
            )}
          </div>
        )}
      </Modal>

      {/* 撤销确认弹窗 */}
      <ConfirmModal
        isOpen={showUndoConfirm}
        onClose={() => { setShowUndoConfirm(false); setUndoTarget(null) }}
        title="撤销操作"
        message={`确定要撤销对 "${undoTarget?.target_name}" 的${getOperationTypeLabel(undoTarget?.operation_type || '').label}操作吗？`}
        confirmText="撤销"
        variant="warning"
        onConfirm={handleUndo}
        loading={undoLoading}
      />
    </div>
  )
}
