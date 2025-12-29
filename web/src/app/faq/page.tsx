'use client'

import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import toast from 'react-hot-toast'
import { Navbar, Footer } from '@/components/layout'
import { Button } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'

/**
 * FAQ分类接口
 */
interface FAQCategory {
  id: number
  name: string
  icon: string
  sort_order: number
}

/**
 * FAQ接口
 */
interface FAQ {
  id: number
  category_id: number
  question: string
  answer: string
  view_count: number
  helpful: number
  not_helpful: number
}

/**
 * FAQ常见问题页面
 */
export default function FAQPage() {
  const [categories, setCategories] = useState<FAQCategory[]>([])
  const [faqs, setFaqs] = useState<FAQ[]>([])
  const [hotFaqs, setHotFaqs] = useState<FAQ[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedCategory, setSelectedCategory] = useState<number>(0)
  const [expandedFaq, setExpandedFaq] = useState<number | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<FAQ[]>([])
  const [searching, setSearching] = useState(false)
  const [feedbackStatus, setFeedbackStatus] = useState<Record<number, boolean | null>>({})

  // 加载分类和热门FAQ
  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      
      // 并行加载分类和热门FAQ
      const [catRes, hotRes] = await Promise.all([
        apiGet<{ data: FAQCategory[] }>('/api/faq/categories'),
        apiGet<{ data: FAQ[] }>('/api/faq/hot?limit=5')
      ])
      
      if (catRes.success && catRes.data) {
        setCategories(catRes.data)
      }
      if (hotRes.success && hotRes.data) {
        setHotFaqs(hotRes.data)
      }
      
      setLoading(false)
    }
    loadData()
  }, [])

  // 加载分类下的FAQ
  useEffect(() => {
    const loadFaqs = async () => {
      if (searchQuery) return // 搜索模式下不加载分类FAQ
      
      const res = await apiGet<{ data: FAQ[] }>(`/api/faq/list?category_id=${selectedCategory}`)
      if (res.success && res.data) {
        setFaqs(res.data)
      }
    }
    loadFaqs()
  }, [selectedCategory, searchQuery])

  // 搜索FAQ
  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      setSearchResults([])
      return
    }
    
    setSearching(true)
    const res = await apiGet<{ data: FAQ[] }>(`/api/faq/search?keyword=${encodeURIComponent(searchQuery)}`)
    if (res.success && res.data) {
      setSearchResults(res.data)
    }
    setSearching(false)
  }

  // 搜索防抖
  useEffect(() => {
    const timer = setTimeout(() => {
      if (searchQuery.trim()) {
        handleSearch()
      } else {
        setSearchResults([])
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [searchQuery])

  // 展开FAQ详情
  const handleExpandFaq = async (faq: FAQ) => {
    if (expandedFaq === faq.id) {
      setExpandedFaq(null)
    } else {
      setExpandedFaq(faq.id)
      // 获取详情（增加浏览次数）
      const res = await apiGet<{ feedback: boolean | null }>(`/api/faq/detail/${faq.id}`)
      if (res.success && res.feedback !== undefined) {
        setFeedbackStatus(prev => ({ ...prev, [faq.id]: res.feedback }))
      }
    }
  }

  // 提交反馈
  const handleFeedback = async (faqId: number, helpful: boolean) => {
    const res = await apiPost(`/api/faq/feedback/${faqId}`, { helpful })
    if (res.success) {
      setFeedbackStatus(prev => ({ ...prev, [faqId]: helpful }))
      toast.success('感谢您的反馈！')
    } else {
      toast.error(res.error || '提交反馈失败')
    }
  }

  // 显示的FAQ列表
  const displayFaqs = searchQuery ? searchResults : faqs

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />

      <main className="flex-1 py-8 px-4">
        <div className="max-w-4xl mx-auto">
          {/* 页面标题 */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-center mb-8"
          >
            <h1 className="text-3xl font-bold text-dark-100 mb-2">
              <i className="fas fa-question-circle mr-3 text-primary-400" />
              常见问题
            </h1>
            <p className="text-dark-400">
              在这里找到您需要的答案，如果没有找到，可以联系客服
            </p>
          </motion.div>

          {/* 搜索框 */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="mb-8"
          >
            <div className="relative">
              <i className="fas fa-search absolute left-4 top-1/2 -translate-y-1/2 text-dark-500" />
              <input
                type="text"
                placeholder="搜索问题..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-12 pr-4 py-4 bg-dark-800/50 border border-dark-700/50 rounded-xl text-dark-100 placeholder-dark-500 focus:outline-none focus:border-primary-500/50 transition-colors text-lg"
              />
              {searchQuery && (
                <button
                  onClick={() => setSearchQuery('')}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-dark-500 hover:text-dark-300"
                >
                  <i className="fas fa-times" />
                </button>
              )}
            </div>
            {searching && (
              <div className="mt-2 text-dark-400 text-sm">
                <i className="fas fa-spinner fa-spin mr-2" />
                搜索中...
              </div>
            )}
          </motion.div>

          {/* 热门问题 */}
          {!searchQuery && hotFaqs.length > 0 && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="mb-8"
            >
              <h2 className="text-lg font-semibold text-dark-200 mb-4 flex items-center">
                <i className="fas fa-fire text-orange-400 mr-2" />
                热门问题
              </h2>
              <div className="space-y-2">
                {hotFaqs.map((faq) => (
                  <button
                    key={faq.id}
                    onClick={() => {
                      setSearchQuery('')
                      setExpandedFaq(faq.id)
                      // 滚动到对应FAQ
                      setTimeout(() => {
                        document.getElementById(`faq-${faq.id}`)?.scrollIntoView({ behavior: 'smooth' })
                      }, 100)
                    }}
                    className="w-full text-left px-4 py-3 bg-dark-800/30 hover:bg-dark-800/50 rounded-lg text-dark-300 hover:text-dark-100 transition-colors flex items-center"
                  >
                    <i className="fas fa-chevron-right text-primary-400 mr-3 text-sm" />
                    <span className="flex-1 truncate">{faq.question}</span>
                    <span className="text-dark-500 text-sm ml-2">
                      <i className="fas fa-eye mr-1" />
                      {faq.view_count}
                    </span>
                  </button>
                ))}
              </div>
            </motion.div>
          )}

          {/* 分类标签 */}
          {!searchQuery && categories.length > 0 && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3 }}
              className="mb-6"
            >
              <div className="flex flex-wrap gap-2">
                <button
                  onClick={() => setSelectedCategory(0)}
                  className={`px-4 py-2 rounded-lg transition-colors ${
                    selectedCategory === 0
                      ? 'bg-primary-500 text-white'
                      : 'bg-dark-800/50 text-dark-300 hover:bg-dark-700/50'
                  }`}
                >
                  全部
                </button>
                {categories.map((cat) => (
                  <button
                    key={cat.id}
                    onClick={() => setSelectedCategory(cat.id)}
                    className={`px-4 py-2 rounded-lg transition-colors ${
                      selectedCategory === cat.id
                        ? 'bg-primary-500 text-white'
                        : 'bg-dark-800/50 text-dark-300 hover:bg-dark-700/50'
                    }`}
                  >
                    {cat.icon && <i className={`${cat.icon} mr-2`} />}
                    {cat.name}
                  </button>
                ))}
              </div>
            </motion.div>
          )}

          {/* 搜索结果提示 */}
          {searchQuery && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="mb-4 text-dark-400"
            >
              找到 {searchResults.length} 个相关问题
              {searchResults.length === 0 && (
                <span className="ml-2">- 试试其他关键词？</span>
              )}
            </motion.div>
          )}

          {/* FAQ列表 */}
          {loading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="card p-6 animate-pulse">
                  <div className="h-6 bg-dark-700/50 rounded w-3/4 mb-2" />
                  <div className="h-4 bg-dark-700/50 rounded w-1/2" />
                </div>
              ))}
            </div>
          ) : displayFaqs.length === 0 ? (
            <div className="text-center py-16">
              <div className="text-6xl mb-4">🤔</div>
              <p className="text-dark-400 mb-4">
                {searchQuery ? '没有找到相关问题' : '暂无常见问题'}
              </p>
              <Button
                variant="secondary"
                onClick={() => window.location.href = '/message'}
              >
                <i className="fas fa-headset mr-2" />
                联系客服
              </Button>
            </div>
          ) : (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="space-y-3"
            >
              <AnimatePresence>
                {displayFaqs.map((faq, index) => (
                  <motion.div
                    key={faq.id}
                    id={`faq-${faq.id}`}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -10 }}
                    transition={{ delay: index * 0.05 }}
                    className="card overflow-hidden"
                  >
                    {/* 问题标题 */}
                    <button
                      onClick={() => handleExpandFaq(faq)}
                      className="w-full px-6 py-4 flex items-center justify-between text-left hover:bg-dark-700/30 transition-colors"
                    >
                      <div className="flex items-center flex-1 min-w-0">
                        <i className={`fas fa-${expandedFaq === faq.id ? 'minus' : 'plus'} text-primary-400 mr-3`} />
                        <span className="text-dark-100 font-medium truncate">{faq.question}</span>
                      </div>
                      <div className="flex items-center text-dark-500 text-sm ml-4 shrink-0">
                        <span className="mr-3">
                          <i className="fas fa-eye mr-1" />
                          {faq.view_count}
                        </span>
                        <span className="text-emerald-400">
                          <i className="fas fa-thumbs-up mr-1" />
                          {faq.helpful}
                        </span>
                      </div>
                    </button>

                    {/* 答案内容 */}
                    <AnimatePresence>
                      {expandedFaq === faq.id && (
                        <motion.div
                          initial={{ height: 0, opacity: 0 }}
                          animate={{ height: 'auto', opacity: 1 }}
                          exit={{ height: 0, opacity: 0 }}
                          transition={{ duration: 0.2 }}
                          className="overflow-hidden"
                        >
                          <div className="px-6 pb-4 border-t border-dark-700/50">
                            <div className="pt-4 text-dark-300 whitespace-pre-wrap leading-relaxed">
                              {faq.answer}
                            </div>
                            
                            {/* 反馈区域 */}
                            <div className="mt-6 pt-4 border-t border-dark-700/30 flex items-center justify-between">
                              <span className="text-dark-500 text-sm">这个回答对您有帮助吗？</span>
                              <div className="flex items-center gap-2">
                                <button
                                  onClick={() => handleFeedback(faq.id, true)}
                                  className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                                    feedbackStatus[faq.id] === true
                                      ? 'bg-emerald-500/20 text-emerald-400'
                                      : 'bg-dark-700/50 text-dark-400 hover:bg-dark-700'
                                  }`}
                                >
                                  <i className="fas fa-thumbs-up mr-1" />
                                  有帮助
                                </button>
                                <button
                                  onClick={() => handleFeedback(faq.id, false)}
                                  className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                                    feedbackStatus[faq.id] === false
                                      ? 'bg-red-500/20 text-red-400'
                                      : 'bg-dark-700/50 text-dark-400 hover:bg-dark-700'
                                  }`}
                                >
                                  <i className="fas fa-thumbs-down mr-1" />
                                  没帮助
                                </button>
                              </div>
                            </div>
                          </div>
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </motion.div>
                ))}
              </AnimatePresence>
            </motion.div>
          )}

          {/* 底部提示 */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5 }}
            className="mt-12 text-center"
          >
            <div className="inline-flex items-center px-6 py-4 bg-dark-800/30 rounded-xl">
              <i className="fas fa-info-circle text-primary-400 mr-3" />
              <span className="text-dark-400">
                没有找到答案？
                <a href="/message" className="text-primary-400 hover:text-primary-300 ml-2">
                  联系客服
                </a>
              </span>
            </div>
          </motion.div>
        </div>
      </main>

      <Footer />
    </div>
  )
}
