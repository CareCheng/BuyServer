'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { Product } from '../types'
import { ProductFormData, DetailFormData, SpecItem } from './types'
import { ProductList } from './ProductList'
import { ProductFormModal } from './ProductFormModal'
import { ProductDetailModal } from './ProductDetailModal'
import { KamiManager } from './KamiManager'

/**
 * 商品管理页面
 * 支持手动卡密商品类型
 */
export function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingProduct, setEditingProduct] = useState<Product | null>(null)
  const [form, setForm] = useState<ProductFormData>({
    name: '', description: '', price: '', stock: '0', duration: '30', duration_unit: '天',
    status: '1', category_id: ''
  })
  
  // 商品详情扩展字段
  const [detailForm, setDetailForm] = useState<DetailFormData>({
    detail: '',
    specs: [] as SpecItem[],
    features: [] as string[],
    tags: ''
  })
  
  // 详情编辑弹窗
  const [showDetailModal, setShowDetailModal] = useState(false)
  const [detailProduct, setDetailProduct] = useState<Product | null>(null)
  const [detailSaving, setDetailSaving] = useState(false)
  
  // 手动卡密管理状态
  const [showKamiModal, setShowKamiModal] = useState(false)
  const [kamiProduct, setKamiProduct] = useState<Product | null>(null)

  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteProductId, setDeleteProductId] = useState<number | null>(null)

  const loadData = useCallback(async () => {
    const productsRes = await apiGet<{ products: Product[] }>('/api/admin/products')
    if (productsRes.success) setProducts(productsRes.products || [])
    setLoading(false)
  }, [])

  useEffect(() => { loadData() }, [loadData])

  const openAddModal = () => {
    setEditingProduct(null)
    setForm({ name: '', description: '', price: '', stock: '0', duration: '30', duration_unit: '天', status: '1', category_id: '' })
    setShowModal(true)
  }

  const openEditModal = (product: Product) => {
    setEditingProduct(product)
    setForm({
      name: product.name, description: product.description || '', price: String(product.price),
      stock: String(product.stock), duration: String(product.duration), duration_unit: product.duration_unit,
      status: String(product.status), category_id: ''
    })
    setShowModal(true)
  }

  const handleSave = async () => {
    if (!form.name.trim()) { toast.error('请输入商品名称'); return }
    if (!form.price || isNaN(Number(form.price))) { toast.error('请输入有效价格'); return }

    const data = {
      name: form.name.trim(), description: form.description.trim(),
      price: parseFloat(form.price), 
      stock: 0, // 手动卡密模式库存由卡密数量决定
      duration: parseInt(form.duration), duration_unit: form.duration_unit,
      status: parseInt(form.status)
    }

    const res = editingProduct
      ? await apiPut(`/api/admin/product/${editingProduct.id}`, data)
      : await apiPost('/api/admin/product', data)

    if (res.success) { toast.success('保存成功'); setShowModal(false); loadData() }
    else toast.error(res.error || '保存失败')
  }

  // 打开删除确认弹窗
  const handleDelete = (id: number) => {
    setDeleteProductId(id)
    setShowDeleteConfirm(true)
  }

  // 确认删除商品
  const confirmDelete = async () => {
    if (!deleteProductId) return
    const res = await apiDelete(`/api/admin/product/${deleteProductId}`)
    if (res.success) { 
      toast.success('删除成功')
      loadData() 
    } else {
      toast.error(res.error || '删除失败')
    }
    setShowDeleteConfirm(false)
    setDeleteProductId(null)
  }

  // 打开卡密管理弹窗
  const openKamiModal = (product: Product) => {
    setKamiProduct(product)
    setShowKamiModal(true)
  }

  // 打开详情编辑弹窗
  const openDetailModal = (product: Product) => {
    setDetailProduct(product)
    let specs: SpecItem[] = []
    let features: string[] = []
    try {
      if (product.specs) specs = JSON.parse(product.specs)
    } catch { specs = [] }
    try {
      if (product.features) features = JSON.parse(product.features)
    } catch { features = [] }
    
    setDetailForm({
      detail: product.detail || '',
      specs: specs.length > 0 ? specs : [{ key: '', value: '' }],
      features: features.length > 0 ? features : [''],
      tags: product.tags || ''
    })
    setShowDetailModal(true)
  }

  // 保存商品详情
  const handleSaveDetail = async () => {
    if (!detailProduct) return
    setDetailSaving(true)
    
    const validSpecs = detailForm.specs.filter(s => s.key.trim() && s.value.trim())
    const validFeatures = detailForm.features.filter(f => f.trim())
    
    const data = {
      ...detailProduct,
      detail: detailForm.detail,
      specs: JSON.stringify(validSpecs),
      features: JSON.stringify(validFeatures),
      tags: detailForm.tags.trim()
    }
    
    const res = await apiPut(`/api/admin/product/${detailProduct.id}`, data)
    setDetailSaving(false)
    
    if (res.success) {
      toast.success('详情保存成功')
      setShowDetailModal(false)
      loadData()
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  if (loading) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>

  return (
    <>
      <ProductList
        products={products}
        onAdd={openAddModal}
        onEdit={openEditModal}
        onDelete={handleDelete}
        onOpenDetail={openDetailModal}
        onOpenKami={openKamiModal}
      />

      <ProductFormModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        editingProduct={editingProduct}
        form={form}
        setForm={setForm}
        onSave={handleSave}
      />

      <ProductDetailModal
        isOpen={showDetailModal}
        onClose={() => setShowDetailModal(false)}
        product={detailProduct}
        form={detailForm}
        setForm={setDetailForm}
        saving={detailSaving}
        onSave={handleSaveDetail}
      />

      <KamiManager
        isOpen={showKamiModal}
        onClose={() => setShowKamiModal(false)}
        product={kamiProduct}
        onDataChange={loadData}
      />

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setDeleteProductId(null) }}
        title="删除商品"
        message="确定要删除该商品吗？此操作不可恢复。"
        confirmText="删除"
        variant="danger"
        onConfirm={confirmDelete}
      />
    </>
  )
}
