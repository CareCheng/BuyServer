'use client'

import { useState, useEffect, useCallback, ChangeEvent } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Modal, Badge, Card, Input } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import Toggle from '@/components/common/Toggle'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * 知识库分类接口
 */
interface KnowledgeCategory {
  id: number
  name: string
  icon: string
  sort_order: number
  article_count: number
  status: number
}

/**
 * 知识库文章接口
 */
interface KnowledgeArticle {
  id: number
  category_id: number
  category_name: string
  title: string
  content: string
  keywords: string
  view_count: number
  use_count: number
  status: number
  created_at: string
  updated_at: string
}

/**
 * 知识库管理页面
 */
export function KnowledgePage() {
  const [activeTab, setActiveTab] = useState<'articles' | 'categories'>('articles')
  const [articles, setArticles] = useState<KnowledgeArticle[]>([])
  const [categories, setCategories] = useState<KnowledgeCategory[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [categoryFilter, setCategoryFilter] = useState<number | ''>('')
  const [searchKeyword, setSearchKeyword] = useState('')
  const pageSize = 20

  // 文章编辑弹窗
  const [showArticleModal, setShowArticleModal] = useState(false)
  const [editingArticle, setEditingArticle] = useState<KnowledgeArticle | null>(null)
  const [articleForm, setArticleForm] = useState({
    category_id: 0,
    title: '',
    content: '',
    keywords: '',
    status: 1,
  })

  // 分类编辑弹窗
  const [showCategoryModal, setShowCategoryModal] = useState(false)
  const [editingCategory, setEditingCategory] = useState<KnowledgeCategory | null>(null)
  const [categoryForm, setCategoryForm] = useState({
    name: '',
    icon: '📁',
    sort_order: 0,
    status: 1,
  })

  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<{ type: 'article' | 'category'; id: number; name: string } | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  // 加载数据
  const loadData = useCallback(async () => {
    setLoading(true)
    if (activeTab === 'articles') {
      await loadArticles()
    } else {
      await loadCategories()
    }
    setLoading(false)
  }, [activeTab, page, categoryFilter, searchKeyword])

  useEffect(() => {
    loadData()
  }, [loadData])

  // 初始加载分类（用于文章筛选）
  useEffect(() => {
    loadCategories()
  }, [])

  const loadArticles = async () => {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    })
    if (categoryFilter !== '') {
      params.append('category_id', categoryFilter.toString())
    }
    if (searchKeyword) {
      params.append('keyword', searchKeyword)
    }

    const res = await apiGet<{ articles: KnowledgeArticle[]; total: number }>(
      `/api/admin/knowledge/articles?${params}`
    )
    if (res.success) {
      setArticles(res.articles || [])
      setTotal(res.total || 0)
    }
  }

  const loadCategories = async () => {
    const res = await apiGet<{ categories: KnowledgeCategory[] }>('/api/admin/knowledge/categories')
    if (res.success) {
      setCategories(res.categories || [])
    }
  }

  // 打开文章编辑弹窗
  const openArticleModal = (article?: KnowledgeArticle) => {
    if (article) {
      setEditingArticle(article)
      setArticleForm({
        category_id: article.category_id,
        title: article.title,
        content: article.content,
        keywords: article.keywords,
        status: article.status,
      })
    } else {
      setEditingArticle(null)
      setArticleForm({
        category_id: categories[0]?.id || 0,
        title: '',
        content: '',
        keywords: '',
        status: 1,
      })
    }
    setShowArticleModal(true)
  }

  // 保存文章
  const handleSaveArticle = async () => {
    if (!articleForm.title.trim()) {
      toast.error('请输入文章标题')
      return
    }
    if (!articleForm.content.trim()) {
      toast.error('请输入文章内容')
      return
    }
    if (!articleForm.category_id) {
      toast.error('请选择分类')
      return
    }
    const res = editingArticle
      ? await apiPut(`/api/admin/knowledge/article/${editingArticle.id}`, articleForm)
      : await apiPost('/api/admin/knowledge/article', articleForm)
    if (res.success) {
      toast.success(editingArticle ? '文章已更新' : '文章已创建')
      setShowArticleModal(false)
      loadArticles()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 打开删除确认弹窗
  const openDeleteConfirm = (type: 'article' | 'category', id: number, name: string) => {
    setDeleteTarget({ type, id, name })
    setShowDeleteConfirm(true)
  }

  // 执行删除
  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    const url = deleteTarget.type === 'article'
      ? `/api/admin/knowledge/article/${deleteTarget.id}`
      : `/api/admin/knowledge/category/${deleteTarget.id}`
    const res = await apiDelete(url)
    setDeleteLoading(false)
    if (res.success) {
      toast.success(deleteTarget.type === 'article' ? '文章已删除' : '分类已删除')
      setShowDeleteConfirm(false)
      setDeleteTarget(null)
      if (deleteTarget.type === 'article') {
        loadArticles()
      } else {
        loadCategories()
      }
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  // 删除文章（打开确认弹窗）
  const handleDeleteArticle = (id: number) => {
    const article = articles.find(a => a.id === id)
    openDeleteConfirm('article', id, article?.title || '')
  }

  // 删除分类（打开确认弹窗）
  const handleDeleteCategory = (id: number) => {
    const cat = categories.find(c => c.id === id)
    openDeleteConfirm('category', id, cat?.name || '')
  }

  // 打开分类编辑弹窗
  const openCategoryModal = (category?: KnowledgeCategory) => {
    if (category) {
      setEditingCategory(category)
      setCategoryForm({
        name: category.name,
        icon: category.icon,
        sort_order: category.sort_order,
        status: category.status,
      })
    } else {
      setEditingCategory(null)
      setCategoryForm({
        name: '',
        icon: '📁',
        sort_order: 0,
        status: 1,
      })
    }
    setShowCategoryModal(true)
  }

  // 保存分类
  const handleSaveCategory = async () => {
    if (!categoryForm.name.trim()) {
      toast.error('请输入分类名称')
      return
    }
    const res = editingCategory
      ? await apiPut(`/api/admin/knowledge/category/${editingCategory.id}`, categoryForm)
      : await apiPost('/api/admin/knowledge/category', categoryForm)
    if (res.success) {
      toast.success(editingCategory ? '分类已更新' : '分类已创建')
      setShowCategoryModal(false)
      loadCategories()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 搜索
  const handleSearch = () => {
    setPage(1)
    loadArticles()
  }

  const totalPages = Math.ceil(total / pageSize)

  if (loading && articles.length === 0 && categories.length === 0) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 标签切换 */}
      <div className="flex gap-2 border-b border-dark-700/50 pb-4">
        <button
          onClick={() => { setActiveTab('articles'); setPage(1) }}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'articles'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-file-alt mr-2" />
          知识文章
        </button>
        <button
          onClick={() => { setActiveTab('categories'); setPage(1) }}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'categories'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-folder mr-2" />
          分类管理
        </button>
      </div>

      {/* 知识文章列表 */}
      {activeTab === 'articles' && (
        <Card
          title="知识文章"
          icon={<i className="fas fa-book" />}
          action={
            <Button size="sm" onClick={() => openArticleModal()}>
              <i className="fas fa-plus mr-1" />
              添加文章
            </Button>
          }
        >
          {/* 搜索和筛选 */}
          <div className="flex flex-col sm:flex-row gap-2 mb-4">
            <select
              value={categoryFilter}
              onChange={(e) => {
                setCategoryFilter(e.target.value === '' ? '' : Number(e.target.value))
                setPage(1)
              }}
              className="input w-full sm:w-40"
            >
              <option value="">全部分类</option>
              {categories.map((cat) => (
                <option key={cat.id} value={cat.id}>{cat.icon} {cat.name}</option>
              ))}
            </select>
            <div className="flex gap-2 flex-1">
              <Input
                placeholder="搜索标题或关键词"
                value={searchKeyword}
                onChange={(e: ChangeEvent<HTMLInputElement>) => setSearchKeyword(e.target.value)}
                className="flex-1"
              />
              <Button onClick={handleSearch}>
                <i className="fas fa-search" />
              </Button>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">标题</th>
                  <th className="pb-3 font-medium">分类</th>
                  <th className="pb-3 font-medium">关键词</th>
                  <th className="pb-3 font-medium">浏览/使用</th>
                  <th className="pb-3 font-medium">状态</th>
                  <th className="pb-3 font-medium">更新时间</th>
                  <th className="pb-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {articles.map((article) => (
                  <motion.tr
                    key={article.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    className="border-b border-dark-700/50"
                  >
                    <td className="py-3">
                      <div className="max-w-xs truncate">{article.title}</div>
                    </td>
                    <td className="py-3 text-sm">{article.category_name}</td>
                    <td className="py-3 text-sm text-dark-400 max-w-xs truncate">
                      {article.keywords || '-'}
                    </td>
                    <td className="py-3 text-sm">
                      <span className="text-blue-400">{article.view_count}</span>
                      <span className="text-dark-500 mx-1">/</span>
                      <span className="text-green-400">{article.use_count}</span>
                    </td>
                    <td className="py-3">
                      <Badge variant={article.status === 1 ? 'success' : 'danger'}>
                        {article.status === 1 ? '启用' : '禁用'}
                      </Badge>
                    </td>
                    <td className="py-3 text-sm text-dark-400">
                      {formatDateTime(article.updated_at)}
                    </td>
                    <td className="py-3">
                      <div className="flex gap-1">
                        <Button size="sm" variant="ghost" onClick={() => openArticleModal(article)}>
                          <i className="fas fa-edit" />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          className="text-red-400"
                          onClick={() => handleDeleteArticle(article.id)}
                        >
                          <i className="fas fa-trash" />
                        </Button>
                      </div>
                    </td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </div>

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
      )}

      {/* 分类管理 */}
      {activeTab === 'categories' && (
        <Card
          title="分类管理"
          icon={<i className="fas fa-folder" />}
          action={
            <Button size="sm" onClick={() => openCategoryModal()}>
              <i className="fas fa-plus mr-1" />
              添加分类
            </Button>
          }
        >
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">图标</th>
                  <th className="pb-3 font-medium">分类名称</th>
                  <th className="pb-3 font-medium">文章数</th>
                  <th className="pb-3 font-medium">排序</th>
                  <th className="pb-3 font-medium">状态</th>
                  <th className="pb-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {categories.map((category) => (
                  <motion.tr
                    key={category.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    className="border-b border-dark-700/50"
                  >
                    <td className="py-3 text-2xl">{category.icon}</td>
                    <td className="py-3">{category.name}</td>
                    <td className="py-3">{category.article_count}</td>
                    <td className="py-3">{category.sort_order}</td>
                    <td className="py-3">
                      <Badge variant={category.status === 1 ? 'success' : 'danger'}>
                        {category.status === 1 ? '启用' : '禁用'}
                      </Badge>
                    </td>
                    <td className="py-3">
                      <div className="flex gap-1">
                        <Button size="sm" variant="ghost" onClick={() => openCategoryModal(category)}>
                          <i className="fas fa-edit" />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          className="text-red-400"
                          onClick={() => handleDeleteCategory(category.id)}
                        >
                          <i className="fas fa-trash" />
                        </Button>
                      </div>
                    </td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* 文章编辑弹窗 */}
      <Modal
        isOpen={showArticleModal}
        onClose={() => setShowArticleModal(false)}
        title={editingArticle ? '编辑文章' : '添加文章'}
        size="lg"
      >
        <div className="space-y-4">
          <Input
            label="文章标题"
            placeholder="请输入文章标题"
            value={articleForm.title}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setArticleForm({ ...articleForm, title: e.target.value })}
          />
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">所属分类</label>
            <select
              value={articleForm.category_id}
              onChange={(e) => setArticleForm({ ...articleForm, category_id: Number(e.target.value) })}
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200"
            >
              <option value={0}>请选择分类</option>
              {categories.map((cat) => (
                <option key={cat.id} value={cat.id}>{cat.icon} {cat.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">文章内容</label>
            <textarea
              value={articleForm.content}
              onChange={(e) => setArticleForm({ ...articleForm, content: e.target.value })}
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200 h-48 resize-none"
              placeholder="请输入文章内容..."
            />
          </div>
          <Input
            label="关键词"
            placeholder="多个关键词用逗号分隔"
            value={articleForm.keywords}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setArticleForm({ ...articleForm, keywords: e.target.value })}
          />
          <Toggle
            checked={articleForm.status === 1}
            onChange={(checked) => setArticleForm({ ...articleForm, status: checked ? 1 : 0 })}
            label="启用文章"
          />
          <Button className="w-full" onClick={handleSaveArticle}>
            {editingArticle ? '保存修改' : '创建文章'}
          </Button>
        </div>
      </Modal>

      {/* 分类编辑弹窗 */}
      <Modal
        isOpen={showCategoryModal}
        onClose={() => setShowCategoryModal(false)}
        title={editingCategory ? '编辑分类' : '添加分类'}
      >
        <div className="space-y-4">
          <Input
            label="分类名称"
            placeholder="请输入分类名称"
            value={categoryForm.name}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setCategoryForm({ ...categoryForm, name: e.target.value })}
          />
          <Input
            label="图标"
            placeholder="请输入Emoji图标"
            value={categoryForm.icon}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setCategoryForm({ ...categoryForm, icon: e.target.value })}
          />
          <Input
            label="排序"
            type="number"
            placeholder="数字越小越靠前"
            value={categoryForm.sort_order.toString()}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setCategoryForm({ ...categoryForm, sort_order: parseInt(e.target.value) || 0 })}
          />
          <Toggle
            checked={categoryForm.status === 1}
            onChange={(checked) => setCategoryForm({ ...categoryForm, status: checked ? 1 : 0 })}
            label="启用分类"
          />
          <Button className="w-full" onClick={handleSaveCategory}>
            {editingCategory ? '保存修改' : '创建分类'}
          </Button>
        </div>
      </Modal>

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setDeleteTarget(null) }}
        title={deleteTarget?.type === 'article' ? '删除文章' : '删除分类'}
        message={deleteTarget?.type === 'article'
          ? `确定要删除文章 "${deleteTarget?.name}" 吗？`
          : `确定要删除分类 "${deleteTarget?.name}" 吗？分类下的文章将被移至未分类。`}
        confirmText="删除"
        variant="danger"
        onConfirm={handleDelete}
        loading={deleteLoading}
      />
    </div>
  )
}
