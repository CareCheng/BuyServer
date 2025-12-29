'use client'

import { useState, useEffect, useCallback } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Modal, Badge } from '@/components/ui'
import Toggle from '@/components/common/Toggle'
import { apiGet, apiPost } from '@/lib/api'
import { formatMoney } from '@/lib/utils'

/**
 * 发票接口
 */
interface Invoice {
  id: number
  user_id: number
  username: string
  order_no: string
  title_type: number  // 1: 个人, 2: 企业
  title_name: string
  tax_number: string
  amount: number
  status: number  // 0: 待开票, 1: 已开票, 2: 已拒绝
  invoice_url: string
  invoice_no: string
  reject_reason: string
  created_at: string
  issued_at: string
}

/**
 * 发票配置接口
 */
interface InvoiceConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  tax_rate: number
  company_name: string
  company_tax_number: string
  company_address: string
  company_phone: string
  company_bank: string
  company_account: string
}

/**
 * 发票管理页面
 */
export function InvoicesPage() {
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [config, setConfig] = useState<InvoiceConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [statusFilter, setStatusFilter] = useState<number | ''>('')
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [showIssueModal, setShowIssueModal] = useState(false)
  const [showRejectModal, setShowRejectModal] = useState(false)
  const [selectedInvoice, setSelectedInvoice] = useState<Invoice | null>(null)
  const [issueForm, setIssueForm] = useState({ invoice_no: '', invoice_url: '' })
  const [rejectReason, setRejectReason] = useState('')
  const [configForm, setConfigForm] = useState<InvoiceConfig>({
    enabled: false,
    min_amount: 0,
    max_amount: 0,
    tax_rate: 0,
    company_name: '',
    company_tax_number: '',
    company_address: '',
    company_phone: '',
    company_bank: '',
    company_account: '',
  })
  const pageSize = 20

  // 加载发票列表
  const loadInvoices = useCallback(async () => {
    setLoading(true)
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    })
    if (statusFilter !== '') {
      params.append('status', statusFilter.toString())
    }

    const res = await apiGet<{ invoices: Invoice[]; total: number }>(
      `/api/admin/invoices?${params}`
    )
    if (res.success) {
      setInvoices(res.invoices || [])
      setTotal(res.total || 0)
    }
    setLoading(false)
  }, [page, statusFilter])

  // 加载发票配置
  const loadConfig = useCallback(async () => {
    const res = await apiGet<{ config: InvoiceConfig }>('/api/admin/invoice/config')
    if (res.success && res.config) {
      setConfig(res.config)
      setConfigForm(res.config)
    }
  }, [])

  useEffect(() => {
    loadInvoices()
    loadConfig()
  }, [loadInvoices, loadConfig])

  // 开票
  const handleIssue = async () => {
    if (!selectedInvoice) return
    if (!issueForm.invoice_no.trim()) {
      toast.error('请输入发票号码')
      return
    }

    const res = await apiPost(`/api/admin/invoice/${selectedInvoice.id}/issue`, issueForm)
    if (res.success) {
      toast.success('开票成功')
      setShowIssueModal(false)
      setSelectedInvoice(null)
      setIssueForm({ invoice_no: '', invoice_url: '' })
      loadInvoices()
    } else {
      toast.error(res.error || '开票失败')
    }
  }

  // 拒绝开票
  const handleReject = async () => {
    if (!selectedInvoice) return
    if (!rejectReason.trim()) {
      toast.error('请输入拒绝原因')
      return
    }

    const res = await apiPost(`/api/admin/invoice/${selectedInvoice.id}/reject`, {
      reason: rejectReason,
    })
    if (res.success) {
      toast.success('已拒绝')
      setShowRejectModal(false)
      setSelectedInvoice(null)
      setRejectReason('')
      loadInvoices()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 保存配置
  const handleSaveConfig = async () => {
    const res = await apiPost('/api/admin/invoice/config', configForm as unknown as Record<string, unknown>)
    if (res.success) {
      toast.success('配置已保存')
      setShowConfigModal(false)
      loadConfig()
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  // 获取状态标签
  const getStatusBadge = (status: number) => {
    switch (status) {
      case 0:
        return <Badge variant="warning">待开票</Badge>
      case 1:
        return <Badge variant="success">已开票</Badge>
      case 2:
        return <Badge variant="danger">已拒绝</Badge>
      default:
        return <Badge variant="default">未知</Badge>
    }
  }

  // 获取抬头类型
  const getTitleType = (type: number) => {
    return type === 1 ? '个人' : '企业'
  }

  const totalPages = Math.ceil(total / pageSize)

  return (
    <div className="space-y-6">
      {/* 头部操作栏 */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
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
            <option value="0">待开票</option>
            <option value="1">已开票</option>
            <option value="2">已拒绝</option>
          </select>
        </div>
        <Button variant="secondary" onClick={() => setShowConfigModal(true)}>
          <i className="fas fa-cog mr-2" />
          发票配置
        </Button>
      </div>

      {/* 统计卡片 */}
      {config && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="card p-4">
            <div className="text-dark-400 text-sm mb-1">发票功能</div>
            <div className={`text-lg font-bold ${config.enabled ? 'text-emerald-400' : 'text-red-400'}`}>
              {config.enabled ? '已启用' : '已禁用'}
            </div>
          </div>
          <div className="card p-4">
            <div className="text-dark-400 text-sm mb-1">最低开票金额</div>
            <div className="text-lg font-bold text-dark-100">{formatMoney(config.min_amount)}</div>
          </div>
          <div className="card p-4">
            <div className="text-dark-400 text-sm mb-1">税率</div>
            <div className="text-lg font-bold text-dark-100">{config.tax_rate}%</div>
          </div>
          <div className="card p-4">
            <div className="text-dark-400 text-sm mb-1">待处理</div>
            <div className="text-lg font-bold text-amber-400">
              {invoices.filter(i => i.status === 0).length}
            </div>
          </div>
        </div>
      )}

      {/* 发票列表 */}
      <div className="card overflow-hidden">
        {loading ? (
          <div className="p-8 text-center">
            <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
          </div>
        ) : invoices.length === 0 ? (
          <div className="p-8 text-center text-dark-400">
            <i className="fas fa-file-invoice text-4xl mb-4 opacity-50" />
            <p>暂无发票申请</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-dark-700/50">
                <tr>
                  <th className="px-4 py-3 text-left text-dark-300 font-medium">ID</th>
                  <th className="px-4 py-3 text-left text-dark-300 font-medium">用户</th>
                  <th className="px-4 py-3 text-left text-dark-300 font-medium">订单号</th>
                  <th className="px-4 py-3 text-left text-dark-300 font-medium">抬头类型</th>
                  <th className="px-4 py-3 text-left text-dark-300 font-medium">抬头名称</th>
                  <th className="px-4 py-3 text-left text-dark-300 font-medium">金额</th>
                  <th className="px-4 py-3 text-left text-dark-300 font-medium">状态</th>
                  <th className="px-4 py-3 text-left text-dark-300 font-medium">申请时间</th>
                  <th className="px-4 py-3 text-left text-dark-300 font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-dark-700/50">
                {invoices.map((invoice) => (
                  <motion.tr
                    key={invoice.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    className="hover:bg-dark-700/30"
                  >
                    <td className="px-4 py-3 text-dark-300">{invoice.id}</td>
                    <td className="px-4 py-3 text-dark-200">{invoice.username}</td>
                    <td className="px-4 py-3 text-dark-300 font-mono text-sm">{invoice.order_no}</td>
                    <td className="px-4 py-3 text-dark-300">{getTitleType(invoice.title_type)}</td>
                    <td className="px-4 py-3 text-dark-200">{invoice.title_name}</td>
                    <td className="px-4 py-3 text-primary-400 font-medium">
                      {formatMoney(invoice.amount)}
                    </td>
                    <td className="px-4 py-3">{getStatusBadge(invoice.status)}</td>
                    <td className="px-4 py-3 text-dark-400 text-sm">
                      {new Date(invoice.created_at).toLocaleString()}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        {invoice.status === 0 && (
                          <>
                            <Button
                              size="sm"
                              variant="primary"
                              onClick={() => {
                                setSelectedInvoice(invoice)
                                setShowIssueModal(true)
                              }}
                            >
                              开票
                            </Button>
                            <Button
                              size="sm"
                              variant="danger"
                              onClick={() => {
                                setSelectedInvoice(invoice)
                                setShowRejectModal(true)
                              }}
                            >
                              拒绝
                            </Button>
                          </>
                        )}
                        {invoice.status === 1 && invoice.invoice_url && (
                          <a
                            href={invoice.invoice_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-primary-400 hover:text-primary-300"
                          >
                            <i className="fas fa-download mr-1" />
                            下载
                          </a>
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
          <div className="flex justify-center items-center gap-2 p-4 border-t border-dark-700/50">
            <Button
              size="sm"
              variant="secondary"
              disabled={page === 1}
              onClick={() => setPage(page - 1)}
            >
              上一页
            </Button>
            <span className="text-dark-400 text-sm">
              {page} / {totalPages}
            </span>
            <Button
              size="sm"
              variant="secondary"
              disabled={page === totalPages}
              onClick={() => setPage(page + 1)}
            >
              下一页
            </Button>
          </div>
        )}
      </div>

      {/* 开票弹窗 */}
      <Modal
        isOpen={showIssueModal}
        onClose={() => {
          setShowIssueModal(false)
          setSelectedInvoice(null)
          setIssueForm({ invoice_no: '', invoice_url: '' })
        }}
        title="开具发票"
      >
        {selectedInvoice && (
          <div className="space-y-4">
            <div className="bg-dark-700/30 rounded-lg p-4 space-y-2">
              <div className="flex justify-between">
                <span className="text-dark-400">抬头名称</span>
                <span className="text-dark-200">{selectedInvoice.title_name}</span>
              </div>
              {selectedInvoice.tax_number && (
                <div className="flex justify-between">
                  <span className="text-dark-400">税号</span>
                  <span className="text-dark-200 font-mono">{selectedInvoice.tax_number}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-dark-400">金额</span>
                <span className="text-primary-400 font-medium">
                  {formatMoney(selectedInvoice.amount)}
                </span>
              </div>
            </div>

            <div>
              <label className="block text-dark-300 text-sm mb-2">发票号码 *</label>
              <input
                type="text"
                value={issueForm.invoice_no}
                onChange={(e) => setIssueForm({ ...issueForm, invoice_no: e.target.value })}
                className="input w-full"
                placeholder="请输入发票号码"
              />
            </div>

            <div>
              <label className="block text-dark-300 text-sm mb-2">发票文件URL</label>
              <input
                type="text"
                value={issueForm.invoice_url}
                onChange={(e) => setIssueForm({ ...issueForm, invoice_url: e.target.value })}
                className="input w-full"
                placeholder="电子发票下载链接（可选）"
              />
            </div>

            <div className="flex gap-3 pt-2">
              <Button
                variant="secondary"
                className="flex-1"
                onClick={() => setShowIssueModal(false)}
              >
                取消
              </Button>
              <Button variant="primary" className="flex-1" onClick={handleIssue}>
                确认开票
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* 拒绝弹窗 */}
      <Modal
        isOpen={showRejectModal}
        onClose={() => {
          setShowRejectModal(false)
          setSelectedInvoice(null)
          setRejectReason('')
        }}
        title="拒绝开票"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-dark-300 text-sm mb-2">拒绝原因 *</label>
            <textarea
              value={rejectReason}
              onChange={(e) => setRejectReason(e.target.value)}
              className="input w-full h-24 resize-none"
              placeholder="请输入拒绝原因"
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

      {/* 配置弹窗 */}
      <Modal
        isOpen={showConfigModal}
        onClose={() => setShowConfigModal(false)}
        title="发票配置"
        size="lg"
      >
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-dark-300">启用发票功能</span>
            <Toggle
              checked={configForm.enabled}
              onChange={(checked) => setConfigForm({ ...configForm, enabled: checked })}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-dark-300 text-sm mb-2">最低开票金额</label>
              <input
                type="number"
                value={configForm.min_amount}
                onChange={(e) => setConfigForm({ ...configForm, min_amount: Number(e.target.value) })}
                className="input w-full"
              />
            </div>
            <div>
              <label className="block text-dark-300 text-sm mb-2">最高开票金额</label>
              <input
                type="number"
                value={configForm.max_amount}
                onChange={(e) => setConfigForm({ ...configForm, max_amount: Number(e.target.value) })}
                className="input w-full"
              />
            </div>
          </div>

          <div>
            <label className="block text-dark-300 text-sm mb-2">税率 (%)</label>
            <input
              type="number"
              value={configForm.tax_rate}
              onChange={(e) => setConfigForm({ ...configForm, tax_rate: Number(e.target.value) })}
              className="input w-full"
              step="0.01"
            />
          </div>

          <div className="border-t border-dark-700/50 pt-4 mt-4">
            <h4 className="text-dark-200 font-medium mb-4">开票方信息</h4>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-dark-300 text-sm mb-2">公司名称</label>
                  <input
                    type="text"
                    value={configForm.company_name}
                    onChange={(e) => setConfigForm({ ...configForm, company_name: e.target.value })}
                    className="input w-full"
                  />
                </div>
                <div>
                  <label className="block text-dark-300 text-sm mb-2">税号</label>
                  <input
                    type="text"
                    value={configForm.company_tax_number}
                    onChange={(e) => setConfigForm({ ...configForm, company_tax_number: e.target.value })}
                    className="input w-full"
                  />
                </div>
              </div>
              <div>
                <label className="block text-dark-300 text-sm mb-2">公司地址</label>
                <input
                  type="text"
                  value={configForm.company_address}
                  onChange={(e) => setConfigForm({ ...configForm, company_address: e.target.value })}
                  className="input w-full"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-dark-300 text-sm mb-2">联系电话</label>
                  <input
                    type="text"
                    value={configForm.company_phone}
                    onChange={(e) => setConfigForm({ ...configForm, company_phone: e.target.value })}
                    className="input w-full"
                  />
                </div>
                <div>
                  <label className="block text-dark-300 text-sm mb-2">开户银行</label>
                  <input
                    type="text"
                    value={configForm.company_bank}
                    onChange={(e) => setConfigForm({ ...configForm, company_bank: e.target.value })}
                    className="input w-full"
                  />
                </div>
              </div>
              <div>
                <label className="block text-dark-300 text-sm mb-2">银行账号</label>
                <input
                  type="text"
                  value={configForm.company_account}
                  onChange={(e) => setConfigForm({ ...configForm, company_account: e.target.value })}
                  className="input w-full"
                />
              </div>
            </div>
          </div>

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
    </div>
  )
}
