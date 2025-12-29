'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, PromptModal, ConfirmModal } from '@/components/ui'
import { apiGet, apiPost, apiDelete } from '@/lib/api'
import { Backup } from './types'

export function BackupsPage() {
  const [backups, setBackups] = useState<Backup[]>([])
  const [dbType, setDbType] = useState('')
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [createLoading, setCreateLoading] = useState(false)
  const [deleteLoading, setDeleteLoading] = useState(false)

  const loadBackups = useCallback(async () => {
    const [backupsRes, infoRes] = await Promise.all([
      apiGet<{ backups: Backup[] }>('/api/admin/backups'),
      apiGet<{ db_type: string }>('/api/admin/backup/info')
    ])
    if (backupsRes.success) setBackups(backupsRes.backups || [])
    if (infoRes.success) setDbType(infoRes.db_type || '')
    setLoading(false)
  }, [])

  useEffect(() => { loadBackups() }, [loadBackups])

  const handleCreate = async (remark: string) => {
    setCreateLoading(true)
    const res = await apiPost<{ backup: Backup }>('/api/admin/backup', { remark: remark.trim() })
    setCreateLoading(false)
    setShowCreateModal(false)
    if (res.success) { toast.success(`备份创建成功：${res.backup?.filename}`); loadBackups() }
    else toast.error(res.error || '备份失败')
  }

  const handleDownload = (id: number) => {
    window.open(`/api/admin/backup/${id}/download`, '_blank')
  }

  const handleDeleteClick = (id: number) => {
    setDeleteId(id)
    setShowDeleteModal(true)
  }

  const handleDelete = async () => {
    if (!deleteId) return
    setDeleteLoading(true)
    const res = await apiDelete(`/api/admin/backup/${deleteId}`)
    setDeleteLoading(false)
    setShowDeleteModal(false)
    setDeleteId(null)
    if (res.success) { toast.success('备份已删除'); loadBackups() }
    else toast.error(res.error || '删除失败')
  }

  const dbTypeText: Record<string, string> = { sqlite: 'SQLite', mysql: 'MySQL', postgres: 'PostgreSQL' }

  if (loading) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>


  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-medium text-dark-100">数据备份</h2>
        <Button size="sm" onClick={() => setShowCreateModal(true)}>创建备份</Button>
      </div>
      <Card>
        <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg mb-4">
          <p className="text-blue-400 text-sm">当前数据库类型：<strong>{dbTypeText[dbType] || dbType}</strong></p>
          <p className="text-blue-400/70 text-xs mt-1">备份文件存储在程序目录的 backups 文件夹中</p>
        </div>
        {backups.length === 0 ? (
          <div className="text-center py-12 text-dark-500">暂无备份</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-dark-700">
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">文件名</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">大小</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">类型</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">备注</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">创建者</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">创建时间</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">操作</th>
                </tr>
              </thead>
              <tbody>
                {backups.map((backup) => (
                  <tr key={backup.id} className="border-b border-dark-700/50 hover:bg-dark-700/30">
                    <td className="py-3 px-4 text-dark-100 font-mono text-sm">{backup.filename}</td>
                    <td className="py-3 px-4 text-dark-300">{backup.file_size_text}</td>
                    <td className="py-3 px-4 text-dark-300">{backup.db_type.toUpperCase()}</td>
                    <td className="py-3 px-4 text-dark-300">{backup.remark || '-'}</td>
                    <td className="py-3 px-4 text-dark-300">{backup.created_by}</td>
                    <td className="py-3 px-4 text-dark-300 text-sm">{backup.created_at}</td>
                    <td className="py-3 px-4">
                      <div className="flex gap-2">
                        <Button size="sm" variant="ghost" onClick={() => handleDownload(backup.id)}>下载</Button>
                        <Button size="sm" variant="ghost" onClick={() => handleDeleteClick(backup.id)}>删除</Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <Card title="备份说明">
        <ul className="text-dark-400 text-sm space-y-2">
          <li>• <strong>SQLite</strong>：直接复制数据库文件并压缩为ZIP格式</li>
          <li>• <strong>MySQL/PostgreSQL</strong>：导出SQL语句文件</li>
          <li>• 建议定期备份数据，并将备份文件下载到本地或其他安全位置</li>
          <li>• 恢复数据需要手动操作，请参考技术文档</li>
        </ul>
      </Card>

      {/* 创建备份弹窗 */}
      <PromptModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        title="创建备份"
        message="请输入备份备注（可选）"
        placeholder="备份备注"
        confirmText="创建"
        onConfirm={handleCreate}
        loading={createLoading}
      />

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteModal}
        onClose={() => { setShowDeleteModal(false); setDeleteId(null) }}
        title="删除备份"
        message="确定要删除此备份吗？删除后无法恢复！"
        confirmText="删除"
        variant="danger"
        onConfirm={handleDelete}
        loading={deleteLoading}
      />
    </div>
  )
}
