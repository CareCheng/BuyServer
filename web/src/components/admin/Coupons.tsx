'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Input, Modal } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { Coupon } from './types'

export function CouponsPage() {
  const [coupons, setCoupons] = useState<Coupon[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingCoupon, setEditingCoupon] = useState<Coupon | null>(null)
  const [form, setForm] = useState({
    name: '', code: '', type: 'percent', value: '', min_amount: '0', max_discount: '0',
    total_count: '-1', per_user_limit: '1', status: '1', start_at: '', end_at: ''
  })
  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Coupon | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  const loadCoupons = useCallback(async () => {
    const res = await apiGet<{ coupons: Coupon[] }>('/api/admin/coupons')
    if (res.success) setCoupons(res.coupons || [])
    setLoading(false)
  }, [])

  useEffect(() => { loadCoupons() }, [loadCoupons])

  const openAddModal = () => {
    setEditingCoupon(null)
    setForm({ name: '', code: '', type: 'percent', value: '', min_amount: '0', max_discount: '0', total_count: '-1', per_user_limit: '1', status: '1', start_at: '', end_at: '' })
    setShowModal(true)
  }

  const openEditModal = (coupon: Coupon) => {
    setEditingCoupon(coupon)
    setForm({
      name: coupon.name, code: coupon.code, type: coupon.type, value: String(coupon.value),
      min_amount: String(coupon.min_amount), max_discount: String(coupon.max_discount),
      total_count: String(coupon.total_count), per_user_limit: String(coupon.per_user_limit),
      status: String(coupon.status),
      start_at: coupon.start_at ? new Date(coupon.start_at).toISOString().slice(0, 16) : '',
      end_at: coupon.end_at ? new Date(coupon.end_at).toISOString().slice(0, 16) : ''
    })
    setShowModal(true)
  }

  const handleSave = async () => {
    if (!form.name.trim()) { toast.error('请输入优惠券名称'); return }
    if (!form.value) { toast.error('请输入优惠值'); return }
    const data = {
      name: form.name.trim(), code: form.code.toUpperCase().trim(), type: form.type,
      value: parseFloat(form.value), min_amount: parseFloat(form.min_amount) || 0,
      max_discount: parseFloat(form.max_discount) || 0, total_count: parseInt(form.total_count),
      per_user_limit: parseInt(form.per_user_limit), status: parseInt(form.status),
      start_at: form.start_at ? new Date(form.start_at).toISOString().replace('T', ' ').slice(0, 19) : '',
      end_at: form.end_at ? new Date(form.end_at).toISOString().replace('T', ' ').slice(0, 19) : ''
    }
    const res = editingCoupon
      ? await apiPut(`/api/admin/coupon/${editingCoupon.id}`, data)
      : await apiPost('/api/admin/coupon', data)
    if (res.success) { toast.success('保存成功'); setShowModal(false); loadCoupons() }
    else toast.error(res.error || '保存失败')
  }

  // 打开删除确认弹窗
  const openDeleteConfirm = (coupon: Coupon) => {
    setDeleteTarget(coupon)
    setShowDeleteConfirm(true)
  }

  // 执行删除
  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    const res = await apiDelete(`/api/admin/coupon/${deleteTarget.id}`)
    setDeleteLoading(false)
    if (res.success) {
      toast.success('删除成功')
      setShowDeleteConfirm(false)
      setDeleteTarget(null)
      loadCoupons()
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  const formatCouponValue = (coupon: Coupon) => {
    switch (coupon.type) {
      case 'percent': return `${coupon.value}%折扣`
      case 'fixed': return `减¥${coupon.value.toFixed(2)}`
      case 'minus': return coupon.min_amount > 0 ? `满${coupon.min_amount}减${coupon.value}` : `减¥${coupon.value.toFixed(2)}`
      default: return ''
    }
  }

  if (loading) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-medium text-dark-100">优惠券列表</h2>
        <Button size="sm" onClick={openAddModal}>创建优惠券</Button>
      </div>
      <Card>
        {coupons.length === 0 ? (
          <div className="text-center py-12 text-dark-500">暂无优惠券</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-dark-700">
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">优惠码</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">名称</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">优惠</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">使用/总量</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">状态</th>
                  <th className="text-left py-3 px-4 text-dark-400 font-medium">操作</th>
                </tr>
              </thead>
              <tbody>
                {coupons.map((coupon) => (
                  <tr key={coupon.id} className="border-b border-dark-700/50 hover:bg-dark-700/30">
                    <td className="py-3 px-4 text-dark-100 font-mono">{coupon.code}</td>
                    <td className="py-3 px-4 text-dark-300">{coupon.name}</td>
                    <td className="py-3 px-4 text-dark-300">{formatCouponValue(coupon)}</td>
                    <td className="py-3 px-4 text-dark-300">{coupon.used_count}/{coupon.total_count === -1 ? '∞' : coupon.total_count}</td>
                    <td className="py-3 px-4">
                      <span className={`px-2 py-1 rounded text-xs ${coupon.status === 1 ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                        {coupon.status === 1 ? '启用' : '禁用'}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex gap-2">
                        <Button size="sm" variant="ghost" onClick={() => openEditModal(coupon)}>编辑</Button>
                        <Button size="sm" variant="ghost" onClick={() => openDeleteConfirm(coupon)}>删除</Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <Modal isOpen={showModal} onClose={() => setShowModal(false)} title={editingCoupon ? '编辑优惠券' : '创建优惠券'}>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Input label="优惠券名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
            <Input label="优惠券码 (留空自动生成)" value={form.code} onChange={(e) => setForm({ ...form, code: e.target.value })} />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-1">优惠类型</label>
              <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })}>
                <option value="percent">折扣（百分比）</option>
                <option value="fixed">固定金额</option>
                <option value="minus">满减</option>
              </select>
            </div>
            <Input label={form.type === 'percent' ? '折扣百分比' : '优惠金额'} type="number" step="0.01" value={form.value} onChange={(e) => setForm({ ...form, value: e.target.value })} required />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input label="最低消费金额" type="number" step="0.01" value={form.min_amount} onChange={(e) => setForm({ ...form, min_amount: e.target.value })} />
            <Input label="最大优惠金额 (0=无限)" type="number" step="0.01" value={form.max_discount} onChange={(e) => setForm({ ...form, max_discount: e.target.value })} />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input label="发放总量 (-1=无限)" type="number" value={form.total_count} onChange={(e) => setForm({ ...form, total_count: e.target.value })} />
            <Input label="每人限用次数" type="number" value={form.per_user_limit} onChange={(e) => setForm({ ...form, per_user_limit: e.target.value })} />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input label="生效时间" type="datetime-local" value={form.start_at} onChange={(e) => setForm({ ...form, start_at: e.target.value })} />
            <Input label="失效时间" type="datetime-local" value={form.end_at} onChange={(e) => setForm({ ...form, end_at: e.target.value })} />
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">状态</label>
            <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
              <option value="1">启用</option>
              <option value="0">禁用</option>
            </select>
          </div>
          <div className="flex justify-end gap-2 pt-4">
            <Button variant="secondary" onClick={() => setShowModal(false)}>取消</Button>
            <Button onClick={handleSave}>保存</Button>
          </div>
        </div>
      </Modal>

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setDeleteTarget(null) }}
        title="删除优惠券"
        message={`确定要删除优惠券 "${deleteTarget?.name}" 吗？`}
        confirmText="删除"
        variant="danger"
        onConfirm={handleDelete}
        loading={deleteLoading}
      />
    </div>
  )
}
