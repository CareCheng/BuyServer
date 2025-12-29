'use client'

import { useState, useEffect } from 'react'
import toast from 'react-hot-toast'
import { Button, Badge, Card, Input, Modal } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { apiGet, apiPut, apiPost, apiDelete } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'
import { Staff } from './types'

/**
 * 客服人员管理组件
 */
export function StaffManagement() {
  const [staffList, setStaffList] = useState<Staff[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [editStaff, setEditStaff] = useState<Staff | null>(null)
  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteStaffId, setDeleteStaffId] = useState<number | null>(null)

  const loadStaff = async () => {
    setLoading(true)
    const res = await apiGet<{ staff: Staff[] }>('/api/admin/support/staff')
    if (res.success) {
      setStaffList(res.staff || [])
    }
    setLoading(false)
  }

  useEffect(() => {
    loadStaff()
  }, [])

  // 打开删除确认弹窗
  const handleDelete = (id: number) => {
    setDeleteStaffId(id)
    setShowDeleteConfirm(true)
  }

  // 确认删除客服
  const confirmDelete = async () => {
    if (!deleteStaffId) return
    const res = await apiDelete(`/api/admin/support/staff/${deleteStaffId}`)
    if (res.success) {
      toast.success('删除成功')
      loadStaff()
    } else {
      toast.error(res.error || '删除失败')
    }
    setShowDeleteConfirm(false)
    setDeleteStaffId(null)
  }

  const getStatusBadge = (status: number) => {
    if (status === 1) return <Badge variant="success">在线</Badge>
    if (status === 0) return <Badge variant="default">离线</Badge>
    return <Badge variant="danger">禁用</Badge>
  }

  return (
    <Card>
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 mb-4">
        <h3 className="text-lg font-medium text-dark-100">客服列表</h3>
        <Button size="sm" onClick={() => setShowCreate(true)}>
          <i className="fas fa-plus mr-2" />
          添加客服
        </Button>
      </div>

      {loading ? (
        <div className="text-center py-8">
          <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
        </div>
      ) : staffList.length === 0 ? (
        <div className="text-center py-8 text-dark-400">暂无客服人员</div>
      ) : (
        <div className="overflow-x-auto -mx-4 sm:mx-0">
          <table className="w-full min-w-[700px]">
            <thead>
              <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                <th className="pb-3 px-4 sm:px-0 sm:pr-4">用户名</th>
                <th className="pb-3 pr-4">昵称</th>
                <th className="pb-3 pr-4 hidden md:table-cell">邮箱</th>
                <th className="pb-3 pr-4">角色</th>
                <th className="pb-3 pr-4">状态</th>
                <th className="pb-3 pr-4">负载</th>
                <th className="pb-3 pr-4 hidden lg:table-cell">最后活跃</th>
                <th className="pb-3 px-4 sm:px-0">操作</th>
              </tr>
            </thead>
            <tbody>
              {staffList.map((staff) => (
                <tr key={staff.id} className="border-b border-dark-700/50">
                  <td className="py-3 px-4 sm:px-0 sm:pr-4 text-dark-100 font-medium">{staff.username}</td>
                  <td className="py-3 pr-4 text-dark-200">{staff.nickname || '-'}</td>
                  <td className="py-3 pr-4 text-dark-300 hidden md:table-cell">{staff.email || '-'}</td>
                  <td className="py-3 pr-4">
                    {staff.role === 'supervisor' ? (
                      <Badge variant="info">主管</Badge>
                    ) : (
                      <Badge>客服</Badge>
                    )}
                  </td>
                  <td className="py-3 pr-4">{getStatusBadge(staff.status)}</td>
                  <td className="py-3 pr-4 text-dark-300">
                    <span className="font-mono">{staff.current_load}/{staff.max_tickets}</span>
                  </td>
                  <td className="py-3 pr-4 text-dark-400 text-sm hidden lg:table-cell">
                    {staff.last_active_at ? formatDateTime(staff.last_active_at) : '-'}
                  </td>
                  <td className="py-3 px-4 sm:px-0">
                    <div className="flex gap-1">
                      <Button size="sm" variant="ghost" onClick={() => setEditStaff(staff)} title="编辑">
                        <i className="fas fa-edit" />
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => handleDelete(staff.id)} title="删除">
                        <i className="fas fa-trash text-red-400" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <StaffModal
        isOpen={showCreate}
        onClose={() => setShowCreate(false)}
        onSuccess={() => { loadStaff(); setShowCreate(false) }}
      />

      {editStaff && (
        <StaffModal
          isOpen={!!editStaff}
          onClose={() => setEditStaff(null)}
          staff={editStaff}
          onSuccess={() => { loadStaff(); setEditStaff(null) }}
        />
      )}

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setDeleteStaffId(null) }}
        title="删除客服"
        message="确定要删除该客服账号吗？此操作不可恢复。"
        confirmText="删除"
        variant="danger"
        onConfirm={confirmDelete}
      />
    </Card>
  )
}


/**
 * 客服编辑弹窗
 */
function StaffModal({
  isOpen,
  onClose,
  staff,
  onSuccess,
}: {
  isOpen: boolean
  onClose: () => void
  staff?: Staff
  onSuccess: () => void
}) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [nickname, setNickname] = useState('')
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('staff')
  const [maxTickets, setMaxTickets] = useState(10)
  const [status, setStatus] = useState(0)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (staff) {
      setUsername(staff.username)
      setNickname(staff.nickname)
      setEmail(staff.email || '')
      setRole(staff.role)
      setMaxTickets(staff.max_tickets)
      setStatus(staff.status)
    } else {
      setUsername('')
      setPassword('')
      setNickname('')
      setEmail('')
      setRole('staff')
      setMaxTickets(10)
      setStatus(0)
    }
  }, [staff, isOpen])

  const handleSubmit = async () => {
    if (!staff && !username.trim()) {
      toast.error('请输入用户名')
      return
    }
    if (!staff && !password) {
      toast.error('请输入密码')
      return
    }

    setSubmitting(true)

    let res
    if (staff) {
      res = await apiPut(`/api/admin/support/staff/${staff.id}`, {
        nickname,
        email,
        max_tickets: maxTickets,
        status,
        password: password || undefined,
      })
    } else {
      res = await apiPost('/api/admin/support/staff', {
        username,
        password,
        nickname: nickname || username,
        email,
        role,
      })
    }

    if (res.success) {
      toast.success(staff ? '更新成功' : '创建成功')
      onSuccess()
    } else {
      toast.error(res.error || '操作失败')
    }
    setSubmitting(false)
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={staff ? '编辑客服' : '添加客服'}>
      <div className="space-y-4">
        {!staff && (
          <Input
            label="用户名"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="登录用户名"
          />
        )}

        <Input
          label={staff ? '新密码（留空不修改）' : '密码'}
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder={staff ? '留空则不修改密码' : '登录密码'}
        />

        <Input
          label="昵称"
          value={nickname}
          onChange={(e) => setNickname(e.target.value)}
          placeholder="显示名称"
        />

        <Input
          label="邮箱"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="联系邮箱"
        />

        {!staff && (
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">角色</label>
            <select value={role} onChange={(e) => setRole(e.target.value)} className="input w-full">
              <option value="staff">客服</option>
              <option value="supervisor">主管</option>
            </select>
          </div>
        )}

        {staff && (
          <>
            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-dark-300">最大工单数</label>
              <input
                type="number"
                value={maxTickets}
                onChange={(e) => setMaxTickets(Number(e.target.value))}
                className="input w-full"
                min={1}
                max={100}
              />
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-dark-300">状态</label>
              <select value={status} onChange={(e) => setStatus(Number(e.target.value))} className="input w-full">
                <option value={0}>离线</option>
                <option value={1}>在线</option>
                <option value={-1}>禁用</option>
              </select>
            </div>
          </>
        )}

        <div className="flex justify-end gap-3 pt-4">
          <Button variant="secondary" onClick={onClose}>取消</Button>
          <Button onClick={handleSubmit} loading={submitting}>{staff ? '保存' : '创建'}</Button>
        </div>
      </div>
    </Modal>
  )
}
