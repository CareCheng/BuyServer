'use client'

import { useState, useEffect, ChangeEvent } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Card, Badge, Modal, Input } from '@/components/ui'
import Toggle from '@/components/common/Toggle'
import { apiGet, apiPost, apiDelete } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * 发票接口
 */
interface Invoice {
  id: number
  invoice_no: string
  order_no: string
  type: string
  title: string
  tax_no: string
  amount: number
  email: string
  status: number
  invoice_url: string
  reject_reason: string
  created_at: string
}

/**
 * 发票抬头接口
 */
interface InvoiceTitle {
  id: number
  type: string
  title: string
  tax_no: string
  is_default: boolean
}

/**
 * 我的发票标签页
 */
export function InvoicesTab() {
  const [activeSection, setActiveSection] = useState<'list' | 'titles'>('list')
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [titles, setTitles] = useState<InvoiceTitle[]>([])
  const [loading, setLoading] = useState(true)
  const [showApplyModal, setShowApplyModal] = useState(false)
  const [showTitleModal, setShowTitleModal] = useState(false)
  
  // 申请发票表单
  const [applyForm, setApplyForm] = useState({
    order_no: '',
    type: 'personal',
    title: '',
    tax_no: '',
    email: '',
  })
  
  // 发票抬头表单
  const [titleForm, setTitleForm] = useState({
    type: 'personal',
    title: '',
    tax_no: '',
    is_default: false,
  })

  // 加载发票列表
  const loadInvoices = async () => {
    const res = await apiGet<{ invoices: Invoice[] }>('/api/user/invoices')
    if (res.success && res.invoices) {
      setInvoices(res.invoices)
    }
  }

  // 加载发票抬头
  const loadTitles = async () => {
    const res = await apiGet<{ titles: InvoiceTitle[] }>('/api/user/invoice/titles')
    if (res.success && res.titles) {
      setTitles(res.titles)
    }
  }

  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      await Promise.all([loadInvoices(), loadTitles()])
      setLoading(false)
    }
    loadData()
  }, [])

  // 申请发票
  const handleApplyInvoice = async () => {
    if (!applyForm.order_no || !applyForm.title || !applyForm.email) {
      toast.error('请填写完整信息')
      return
    }
    if (applyForm.type === 'enterprise' && !applyForm.tax_no) {
      toast.error('企业发票需要填写税号')
      return
    }
    const res = await apiPost('/api/user/invoice', applyForm)
    if (res.success) {
      toast.success('发票申请已提交')
      setShowApplyModal(false)
      setApplyForm({ order_no: '', type: 'personal', title: '', tax_no: '', email: '' })
      loadInvoices()
    } else {
      toast.error(res.error || '申请失败')
    }
  }

  // 取消发票申请
  const handleCancelInvoice = async (invoiceNo: string) => {
    const res = await apiPost(`/api/user/invoice/${invoiceNo}/cancel`, {})
    if (res.success) {
      toast.success('已取消申请')
      loadInvoices()
    } else {
      toast.error(res.error || '取消失败')
    }
  }

  // 保存发票抬头
  const handleSaveTitle = async () => {
    if (!titleForm.title) {
      toast.error('请填写抬头名称')
      return
    }
    if (titleForm.type === 'enterprise' && !titleForm.tax_no) {
      toast.error('企业抬头需要填写税号')
      return
    }
    const res = await apiPost('/api/user/invoice/title', titleForm)
    if (res.success) {
      toast.success('抬头已保存')
      setShowTitleModal(false)
      setTitleForm({ type: 'personal', title: '', tax_no: '', is_default: false })
      loadTitles()
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  // 删除发票抬头
  const handleDeleteTitle = async (id: number) => {
    const res = await apiDelete(`/api/user/invoice/title/${id}`)
    if (res.success) {
      toast.success('已删除')
      loadTitles()
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  // 使用抬头填充申请表单
  const useTitle = (title: InvoiceTitle) => {
    setApplyForm({
      ...applyForm,
      type: title.type,
      title: title.title,
      tax_no: title.tax_no,
    })
  }

  // 获取发票状态
  const getInvoiceStatus = (status: number) => {
    const statusMap: Record<number, { text: string; variant: 'warning' | 'success' | 'danger' | 'default' }> = {
      0: { text: '待开具', variant: 'warning' },
      1: { text: '已开具', variant: 'success' },
      2: { text: '已拒绝', variant: 'danger' },
      3: { text: '已取消', variant: 'default' },
    }
    return statusMap[status] || { text: '未知', variant: 'default' }
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
      {/* 分区切换 */}
      <div className="flex gap-2 border-b border-dark-700/50 pb-4">
        <button
          onClick={() => setActiveSection('list')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeSection === 'list'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-file-invoice mr-2" />
          发票列表
        </button>
        <button
          onClick={() => setActiveSection('titles')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeSection === 'titles'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-building mr-2" />
          发票抬头
        </button>
      </div>

      {/* 发票列表 */}
      {activeSection === 'list' && (
        <Card
          title="我的发票"
          icon={<i className="fas fa-file-invoice" />}
          action={
            <Button size="sm" onClick={() => setShowApplyModal(true)}>
              <i className="fas fa-plus mr-1" />
              申请发票
            </Button>
          }
        >
          {invoices.length === 0 ? (
            <div className="text-center py-8 text-dark-400">暂无发票记录</div>
          ) : (
            <div className="space-y-4">
              {invoices.map((invoice) => {
                const status = getInvoiceStatus(invoice.status)
                return (
                  <div key={invoice.id} className="p-4 bg-dark-700/30 rounded-xl border border-dark-600/50">
                    <div className="flex items-start justify-between mb-3">
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-dark-100">{invoice.title}</span>
                          <Badge variant={status.variant}>{status.text}</Badge>
                        </div>
                        <div className="text-sm text-dark-500 mt-1">
                          发票号: {invoice.invoice_no}
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-primary-400 font-bold">¥{invoice.amount.toFixed(2)}</div>
                        <div className="text-sm text-dark-500">{formatDateTime(invoice.created_at)}</div>
                      </div>
                    </div>
                    <div className="text-sm text-dark-400 space-y-1">
                      <div>订单号: {invoice.order_no}</div>
                      <div>类型: {invoice.type === 'personal' ? '个人发票' : '企业发票'}</div>
                      {invoice.tax_no && <div>税号: {invoice.tax_no}</div>}
                      <div>接收邮箱: {invoice.email}</div>
                    </div>
                    {invoice.status === 2 && invoice.reject_reason && (
                      <div className="mt-3 p-2 bg-red-500/10 rounded text-sm text-red-400">
                        拒绝原因: {invoice.reject_reason}
                      </div>
                    )}
                    <div className="mt-3 flex gap-2">
                      {invoice.status === 0 && (
                        <Button size="sm" variant="danger" onClick={() => handleCancelInvoice(invoice.invoice_no)}>
                          取消申请
                        </Button>
                      )}
                      {invoice.status === 1 && invoice.invoice_url && (
                        <a href={invoice.invoice_url} target="_blank" rel="noopener noreferrer">
                          <Button size="sm">
                            <i className="fas fa-download mr-1" />
                            下载发票
                          </Button>
                        </a>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </Card>
      )}

      {/* 发票抬头 */}
      {activeSection === 'titles' && (
        <Card
          title="发票抬头"
          icon={<i className="fas fa-building" />}
          action={
            <Button size="sm" onClick={() => setShowTitleModal(true)}>
              <i className="fas fa-plus mr-1" />
              添加抬头
            </Button>
          }
        >
          {titles.length === 0 ? (
            <div className="text-center py-8 text-dark-400">暂无保存的发票抬头</div>
          ) : (
            <div className="space-y-3">
              {titles.map((title) => (
                <div key={title.id} className="p-4 bg-dark-700/30 rounded-xl border border-dark-600/50 flex items-center justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-dark-100">{title.title}</span>
                      <Badge variant={title.type === 'personal' ? 'default' : 'info'}>
                        {title.type === 'personal' ? '个人' : '企业'}
                      </Badge>
                      {title.is_default && <Badge variant="success">默认</Badge>}
                    </div>
                    {title.tax_no && <div className="text-sm text-dark-500 mt-1">税号: {title.tax_no}</div>}
                  </div>
                  <div className="flex gap-2">
                    <Button size="sm" variant="ghost" onClick={() => handleDeleteTitle(title.id)}>
                      <i className="fas fa-trash text-red-400" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      )}

      {/* 申请发票弹窗 */}
      <Modal isOpen={showApplyModal} onClose={() => setShowApplyModal(false)} title="申请发票" size="md">
        <div className="space-y-4">
          <Input
            label="订单号"
            placeholder="请输入订单号"
            value={applyForm.order_no}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setApplyForm({ ...applyForm, order_no: e.target.value })}
          />
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">发票类型</label>
            <div className="flex gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  checked={applyForm.type === 'personal'}
                  onChange={() => setApplyForm({ ...applyForm, type: 'personal', tax_no: '' })}
                  className="text-primary-500"
                />
                <span className="text-dark-300">个人发票</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  checked={applyForm.type === 'enterprise'}
                  onChange={() => setApplyForm({ ...applyForm, type: 'enterprise' })}
                  className="text-primary-500"
                />
                <span className="text-dark-300">企业发票</span>
              </label>
            </div>
          </div>
          {titles.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-2">选择已保存的抬头</label>
              <div className="flex flex-wrap gap-2">
                {titles.map((title) => (
                  <button
                    key={title.id}
                    onClick={() => useTitle(title)}
                    className="px-3 py-1 text-sm rounded-lg border border-dark-600 text-dark-400 hover:border-primary-500 hover:text-primary-400 transition-colors"
                  >
                    {title.title}
                  </button>
                ))}
              </div>
            </div>
          )}
          <Input
            label="发票抬头"
            placeholder={applyForm.type === 'personal' ? '请输入个人姓名' : '请输入公司名称'}
            value={applyForm.title}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setApplyForm({ ...applyForm, title: e.target.value })}
          />
          {applyForm.type === 'enterprise' && (
            <Input
              label="税号"
              placeholder="请输入纳税人识别号"
              value={applyForm.tax_no}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setApplyForm({ ...applyForm, tax_no: e.target.value })}
            />
          )}
          <Input
            label="接收邮箱"
            type="email"
            placeholder="电子发票将发送到此邮箱"
            value={applyForm.email}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setApplyForm({ ...applyForm, email: e.target.value })}
          />
          <Button className="w-full" onClick={handleApplyInvoice}>
            提交申请
          </Button>
        </div>
      </Modal>

      {/* 添加抬头弹窗 */}
      <Modal isOpen={showTitleModal} onClose={() => setShowTitleModal(false)} title="添加发票抬头" size="sm">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">抬头类型</label>
            <div className="flex gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  checked={titleForm.type === 'personal'}
                  onChange={() => setTitleForm({ ...titleForm, type: 'personal', tax_no: '' })}
                  className="text-primary-500"
                />
                <span className="text-dark-300">个人</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  checked={titleForm.type === 'enterprise'}
                  onChange={() => setTitleForm({ ...titleForm, type: 'enterprise' })}
                  className="text-primary-500"
                />
                <span className="text-dark-300">企业</span>
              </label>
            </div>
          </div>
          <Input
            label="抬头名称"
            placeholder={titleForm.type === 'personal' ? '请输入个人姓名' : '请输入公司名称'}
            value={titleForm.title}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setTitleForm({ ...titleForm, title: e.target.value })}
          />
          {titleForm.type === 'enterprise' && (
            <Input
              label="税号"
              placeholder="请输入纳税人识别号"
              value={titleForm.tax_no}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setTitleForm({ ...titleForm, tax_no: e.target.value })}
            />
          )}
          <Toggle
            checked={titleForm.is_default}
            onChange={(checked) => setTitleForm({ ...titleForm, is_default: checked })}
            label="设为默认抬头"
          />
          <Button className="w-full" onClick={handleSaveTitle}>
            保存
          </Button>
        </div>
      </Modal>
    </motion.div>
  )
}
