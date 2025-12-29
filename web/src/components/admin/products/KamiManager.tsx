'use client'

import { useState, useEffect, useRef } from 'react'
import toast from 'react-hot-toast'
import { Button, Modal } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { apiGet, apiPost, apiDelete } from '@/lib/api'
import { Product, ManualKami, KamiStats } from '../types'

interface KamiManagerProps {
  isOpen: boolean
  onClose: () => void
  product: Product | null
  onDataChange: () => void
}

/**
 * 卡密管理组件
 * 用于管理手动卡密类型商品的卡密
 */
export function KamiManager({ isOpen, onClose, product, onDataChange }: KamiManagerProps) {
  const [kamis, setKamis] = useState<ManualKami[]>([])
  const [kamiStats, setKamiStats] = useState<KamiStats | null>(null)
  const [kamiLoading, setKamiLoading] = useState(false)
  const [kamiPage, setKamiPage] = useState(1)
  const [kamiTotal, setKamiTotal] = useState(0)
  const [kamiFilter, setKamiFilter] = useState<string>('')
  const [showImportModal, setShowImportModal] = useState(false)
  const [importCodes, setImportCodes] = useState('')
  const [importing, setImporting] = useState(false)
  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteKamiId, setDeleteKamiId] = useState<number | null>(null)
  // 用于追踪是否已加载过数据，防止重复请求
  const [dataLoaded, setDataLoaded] = useState(false)
  // 用于追踪当前加载的商品ID，防止切换商品时数据混乱
  const loadedProductIdRef = useRef<number | null>(null)

  // 加载卡密列表
  const loadKamis = async (productId: number, page: number, status: string) => {
    setKamiLoading(true)
    const params = new URLSearchParams({ page: String(page), page_size: '20' })
    if (status) params.append('status', status)
    
    const res = await apiGet<{ kamis: ManualKami[]; total: number; stats: KamiStats }>(
      `/api/admin/product/${productId}/kami?${params}`
    )
    if (res.success) {
      setKamis(res.kamis || [])
      setKamiTotal(res.total || 0)
      setKamiStats(res.stats || null)
    }
    setKamiLoading(false)
    setDataLoaded(true)
  }

  // 使用 useEffect 处理弹窗打开时的数据加载
  useEffect(() => {
    if (isOpen && product) {
      // 检查是否需要重新加载（新打开或切换了商品）
      if (!dataLoaded || loadedProductIdRef.current !== product.id) {
        loadedProductIdRef.current = product.id
        setKamiPage(1)
        setKamiFilter('')
        setDataLoaded(false)
        loadKamis(product.id, 1, '')
      }
    }
  }, [isOpen, product?.id])

  // 弹窗关闭时重置状态
  useEffect(() => {
    if (!isOpen) {
      setDataLoaded(false)
      loadedProductIdRef.current = null
      setKamis([])
      setKamiStats(null)
      setKamiPage(1)
      setKamiFilter('')
    }
  }, [isOpen])

  // 导入卡密
  const handleImportKami = async () => {
    if (!importCodes.trim()) { toast.error('请输入卡密内容'); return }
    if (!product) return

    setImporting(true)
    const res = await apiPost<{ imported: number; duplicates: number }>(
      `/api/admin/product/${product.id}/kami/import`,
      { codes: importCodes }
    )
    setImporting(false)

    if (res.success) {
      toast.success(`导入成功：${res.imported} 个，重复跳过：${res.duplicates} 个`)
      setShowImportModal(false)
      setImportCodes('')
      loadKamis(product.id, kamiPage, kamiFilter)
      onDataChange()
    } else {
      toast.error(res.error || '导入失败')
    }
  }

  // 打开删除确认弹窗
  const handleDeleteKami = (kamiId: number) => {
    setDeleteKamiId(kamiId)
    setShowDeleteConfirm(true)
  }

  // 确认删除卡密
  const confirmDeleteKami = async () => {
    if (!deleteKamiId) return
    const res = await apiDelete(`/api/admin/kami/${deleteKamiId}`)
    if (res.success) {
      toast.success('删除成功')
      if (product) {
        loadKamis(product.id, kamiPage, kamiFilter)
        onDataChange()
      }
    } else {
      toast.error(res.error || '删除失败')
    }
    setShowDeleteConfirm(false)
    setDeleteKamiId(null)
  }

  // 禁用/启用卡密
  const handleToggleKami = async (kami: ManualKami) => {
    const action = kami.status === 0 ? 'disable' : 'enable'
    const res = await apiPost(`/api/admin/kami/${kami.id}/${action}`, {})
    if (res.success) {
      toast.success(action === 'disable' ? '已禁用' : '已启用')
      if (product) {
        loadKamis(product.id, kamiPage, kamiFilter)
        onDataChange()
      }
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 获取卡密状态标签
  const getKamiStatusBadge = (status: number) => {
    switch (status) {
      case 0: return <span className="px-2 py-1 rounded text-xs bg-green-500/20 text-green-400">可用</span>
      case 1: return <span className="px-2 py-1 rounded text-xs bg-blue-500/20 text-blue-400">已售出</span>
      case 2: return <span className="px-2 py-1 rounded text-xs bg-gray-500/20 text-gray-400">已禁用</span>
      default: return null
    }
  }

  // 处理文件导入
  const handleFileImport = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = (event) => {
      const content = event.target?.result as string
      if (content) {
        // 解析 CSV/TXT 文件，支持逗号、换行分隔
        const codes = content
          .split(/[\r\n,]+/)
          .map(line => line.trim())
          .filter(line => line && !line.toLowerCase().startsWith('kami') && !line.toLowerCase().startsWith('code') && !line.toLowerCase().startsWith('卡密'))
        setImportCodes(codes.join('\n'))
        toast.success(`已读取 ${codes.length} 个卡密`)
      }
    }
    reader.readAsText(file)
    e.target.value = ''
  }

  return (
    <>
      <Modal isOpen={isOpen} onClose={onClose} title={`卡密管理 - ${product?.name || ''}`} size="lg">
        <div className="space-y-4">
          {/* 统计信息 */}
          {kamiStats && (
            <div className="grid grid-cols-4 gap-4">
              <div className="bg-dark-700 rounded-lg p-3 text-center">
                <div className="text-2xl font-bold text-dark-100">{kamiStats.total}</div>
                <div className="text-sm text-dark-400">总数</div>
              </div>
              <div className="bg-dark-700 rounded-lg p-3 text-center">
                <div className="text-2xl font-bold text-green-400">{kamiStats.available}</div>
                <div className="text-sm text-dark-400">可用</div>
              </div>
              <div className="bg-dark-700 rounded-lg p-3 text-center">
                <div className="text-2xl font-bold text-blue-400">{kamiStats.sold}</div>
                <div className="text-sm text-dark-400">已售出</div>
              </div>
              <div className="bg-dark-700 rounded-lg p-3 text-center">
                <div className="text-2xl font-bold text-gray-400">{kamiStats.disabled}</div>
                <div className="text-sm text-dark-400">已禁用</div>
              </div>
            </div>
          )}

          {/* 操作栏 */}
          <div className="flex justify-between items-center">
            <div className="flex gap-2">
              <select 
                className="px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100 text-sm"
                value={kamiFilter}
                onChange={(e) => {
                  setKamiFilter(e.target.value)
                  setKamiPage(1)
                  if (product) loadKamis(product.id, 1, e.target.value)
                }}
              >
                <option value="">全部状态</option>
                <option value="0">可用</option>
                <option value="1">已售出</option>
                <option value="2">已禁用</option>
              </select>
            </div>
            <Button size="sm" onClick={() => setShowImportModal(true)}>导入卡密</Button>
          </div>

          {/* 卡密列表 */}
          {kamiLoading ? (
            <div className="text-center py-8"><i className="fas fa-spinner fa-spin text-xl text-primary-400" /></div>
          ) : kamis.length === 0 ? (
            <div className="text-center py-8 text-dark-500">暂无卡密，请点击"导入卡密"添加</div>
          ) : (
            <div className="overflow-x-auto max-h-96">
              <table className="w-full">
                <thead className="sticky top-0 bg-dark-800">
                  <tr className="border-b border-dark-700">
                    <th className="text-left py-2 px-3 text-dark-400 font-medium text-sm">ID</th>
                    <th className="text-left py-2 px-3 text-dark-400 font-medium text-sm">卡密</th>
                    <th className="text-left py-2 px-3 text-dark-400 font-medium text-sm">状态</th>
                    <th className="text-left py-2 px-3 text-dark-400 font-medium text-sm">订单号</th>
                    <th className="text-left py-2 px-3 text-dark-400 font-medium text-sm">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {kamis.map((kami) => (
                    <tr key={kami.id} className="border-b border-dark-700/50 hover:bg-dark-700/30">
                      <td className="py-2 px-3 text-dark-300 text-sm">{kami.id}</td>
                      <td className="py-2 px-3 text-dark-100 text-sm font-mono">{kami.kami_code}</td>
                      <td className="py-2 px-3">{getKamiStatusBadge(kami.status)}</td>
                      <td className="py-2 px-3 text-dark-300 text-sm">{kami.order_no || '-'}</td>
                      <td className="py-2 px-3">
                        <div className="flex gap-1">
                          {kami.status !== 1 && (
                            <>
                              <Button size="sm" variant="ghost" onClick={() => handleToggleKami(kami)}>
                                {kami.status === 0 ? '禁用' : '启用'}
                              </Button>
                              <Button size="sm" variant="ghost" onClick={() => handleDeleteKami(kami.id)}>删除</Button>
                            </>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* 分页 */}
          {kamiTotal > 20 && (
            <div className="flex justify-center gap-2">
              <Button 
                size="sm" 
                variant="secondary" 
                disabled={kamiPage <= 1}
                onClick={() => {
                  const newPage = kamiPage - 1
                  setKamiPage(newPage)
                  if (product) loadKamis(product.id, newPage, kamiFilter)
                }}
              >
                上一页
              </Button>
              <span className="px-3 py-1 text-dark-300">
                {kamiPage} / {Math.ceil(kamiTotal / 20)}
              </span>
              <Button 
                size="sm" 
                variant="secondary" 
                disabled={kamiPage >= Math.ceil(kamiTotal / 20)}
                onClick={() => {
                  const newPage = kamiPage + 1
                  setKamiPage(newPage)
                  if (product) loadKamis(product.id, newPage, kamiFilter)
                }}
              >
                下一页
              </Button>
            </div>
          )}
        </div>
      </Modal>

      {/* 导入卡密弹窗 */}
      <Modal isOpen={showImportModal} onClose={() => setShowImportModal(false)} title="导入卡密">
        <div className="space-y-4">
          {/* 文件导入 */}
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">从文件导入</label>
            <div className="flex gap-2">
              <input
                type="file"
                accept=".csv,.txt"
                className="hidden"
                id="kami-file-input"
                onChange={handleFileImport}
              />
              <label
                htmlFor="kami-file-input"
                className="flex-1 px-4 py-3 bg-dark-700 border border-dark-600 border-dashed rounded-lg text-dark-300 text-center cursor-pointer hover:bg-dark-600 hover:border-dark-500 transition-colors"
              >
                <i className="fas fa-file-upload mr-2" />
                点击选择 CSV 或 TXT 文件
              </label>
            </div>
            <p className="text-xs text-dark-500 mt-1">
              支持 CSV 和 TXT 格式，每行一个卡密或逗号分隔
            </p>
          </div>

          {/* 分隔线 */}
          <div className="flex items-center gap-3">
            <div className="flex-1 border-t border-dark-600" />
            <span className="text-dark-500 text-sm">或手动输入</span>
            <div className="flex-1 border-t border-dark-600" />
          </div>

          {/* 手动输入 */}
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">卡密内容</label>
            <textarea 
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100 h-40 font-mono text-sm"
              placeholder="每行一个卡密，例如：&#10;KAMI-XXXX-XXXX-XXXX&#10;KAMI-YYYY-YYYY-YYYY&#10;KAMI-ZZZZ-ZZZZ-ZZZZ"
              value={importCodes}
              onChange={(e) => setImportCodes(e.target.value)}
            />
            {importCodes && (
              <p className="text-xs text-dark-400 mt-1">
                已输入 {importCodes.split('\n').filter(l => l.trim()).length} 个卡密
              </p>
            )}
          </div>

          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => { setShowImportModal(false); setImportCodes('') }}>取消</Button>
            <Button onClick={handleImportKami} disabled={importing || !importCodes.trim()}>
              {importing ? '导入中...' : '导入'}
            </Button>
          </div>
        </div>
      </Modal>

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setDeleteKamiId(null) }}
        title="删除卡密"
        message="确定要删除该卡密吗？此操作不可恢复。"
        confirmText="删除"
        variant="danger"
        onConfirm={confirmDeleteKami}
      />
    </>
  )
}
