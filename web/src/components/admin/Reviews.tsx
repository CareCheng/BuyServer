'use client'

import { useState, useEffect, useCallback } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Modal, Badge, Card } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * 商品评价接口
 */
interface ProductReview {
  id: number
  user_id: number
  username: string
  product_id: number
  product_name: string
  order_no: string
  rating: number
  content: string
  images: string  // JSON字符串或逗号分隔的URL
  reply: string
  reply_at: string
  status: number  // 0: 待审核, 1: 已通过, 2: 已拒绝
  is_anonymous: boolean
  created_at: string
}

/**
 * 商品评价管理页面
 */
export function ReviewsPage() {
  const [reviews, setReviews] = useState<ProductReview[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [statusFilter, setStatusFilter] = useState<number | ''>('')
  const [showReplyModal, setShowReplyModal] = useState(false)
  const [selectedReview, setSelectedReview] = useState<ProductReview | null>(null)
  const [replyContent, setReplyContent] = useState('')
  // 确认弹窗状态
  const [showRejectConfirm, setShowRejectConfirm] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [confirmTarget, setConfirmTarget] = useState<ProductReview | null>(null)
  const [confirmLoading, setConfirmLoading] = useState(false)
  const pageSize = 20

  // 加载评价列表
  const loadReviews = useCallback(async () => {
    setLoading(true)
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    })
    if (statusFilter !== '') {
      params.append('status', statusFilter.toString())
    }

    const res = await apiGet<{ reviews: ProductReview[]; total: number }>(
      `/api/admin/reviews?${params}`
    )
    if (res.success) {
      setReviews(res.reviews || [])
      setTotal(res.total || 0)
    }
    setLoading(false)
  }, [page, statusFilter])

  useEffect(() => {
    loadReviews()
  }, [loadReviews])

  // 审核通过
  const handleApprove = async (id: number) => {
    const res = await apiPut(`/api/admin/review/${id}/status`, { status: 1 })
    if (res.success) {
      toast.success('评价已通过')
      loadReviews()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 打开拒绝确认弹窗
  const openRejectConfirm = (review: ProductReview) => {
    setConfirmTarget(review)
    setShowRejectConfirm(true)
  }

  // 审核拒绝
  const handleReject = async () => {
    if (!confirmTarget) return
    setConfirmLoading(true)
    const res = await apiPut(`/api/admin/review/${confirmTarget.id}/status`, { status: 2 })
    setConfirmLoading(false)
    if (res.success) {
      toast.success('评价已拒绝')
      setShowRejectConfirm(false)
      setConfirmTarget(null)
      loadReviews()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 打开删除确认弹窗
  const openDeleteConfirm = (review: ProductReview) => {
    setConfirmTarget(review)
    setShowDeleteConfirm(true)
  }

  // 删除评价
  const handleDelete = async () => {
    if (!confirmTarget) return
    setConfirmLoading(true)
    const res = await apiDelete(`/api/admin/review/${confirmTarget.id}`)
    setConfirmLoading(false)
    if (res.success) {
      toast.success('评价已删除')
      setShowDeleteConfirm(false)
      setConfirmTarget(null)
      loadReviews()
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  // 回复评价
  const handleReply = async () => {
    if (!selectedReview) return
    if (!replyContent.trim()) {
      toast.error('请输入回复内容')
      return
    }
    const res = await apiPost(`/api/admin/review/${selectedReview.id}/reply`, {
      reply: replyContent,
    })
    if (res.success) {
      toast.success('回复成功')
      setShowReplyModal(false)
      setSelectedReview(null)
      setReplyContent('')
      loadReviews()
    } else {
      toast.error(res.error || '回复失败')
    }
  }

  // 获取状态标签
  const getStatusBadge = (status: number) => {
    switch (status) {
      case 0:
        return <Badge variant="warning">待审核</Badge>
      case 1:
        return <Badge variant="success">已通过</Badge>
      case 2:
        return <Badge variant="danger">已拒绝</Badge>
      default:
        return <Badge variant="default">未知</Badge>
    }
  }

  // 渲染星级
  const renderStars = (rating: number) => {
    return (
      <div className="flex gap-0.5">
        {[1, 2, 3, 4, 5].map((star) => (
          <i
            key={star}
            className={`fas fa-star text-sm ${
              star <= rating ? 'text-yellow-400' : 'text-dark-600'
            }`}
          />
        ))}
      </div>
    )
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
            <option value="0">待审核</option>
            <option value="1">已通过</option>
            <option value="2">已拒绝</option>
          </select>
        </div>
        <div className="text-dark-400 text-sm">
          共 {total} 条评价
        </div>
      </div>

      {/* 评价列表 */}
      <Card title="商品评价" icon={<i className="fas fa-star" />}>
        {loading ? (
          <div className="p-8 text-center">
            <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
          </div>
        ) : reviews.length === 0 ? (
          <div className="p-8 text-center text-dark-400">
            <i className="fas fa-comment-slash text-4xl mb-4 opacity-50" />
            <p>暂无评价</p>
          </div>
        ) : (
          <div className="space-y-4">
            {reviews.map((review) => (
              <motion.div
                key={review.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                className="p-4 bg-dark-700/30 rounded-lg"
              >
                <div className="flex flex-col sm:flex-row justify-between gap-4">
                  <div className="flex-1 space-y-2">
                    {/* 头部信息 */}
                    <div className="flex flex-wrap items-center gap-3">
                      <span className="text-dark-200 font-medium">
                        {review.is_anonymous ? '匿名用户' : review.username}
                      </span>
                      {renderStars(review.rating)}
                      {getStatusBadge(review.status)}
                      <span className="text-dark-400 text-sm">
                        {formatDateTime(review.created_at)}
                      </span>
                    </div>

                    {/* 商品信息 */}
                    <div className="text-sm text-dark-400">
                      商品：<span className="text-dark-300">{review.product_name}</span>
                      <span className="mx-2">|</span>
                      订单：<span className="text-dark-300 font-mono">{review.order_no}</span>
                    </div>

                    {/* 评价内容 */}
                    <p className="text-dark-200">{review.content}</p>

                    {/* 评价图片 */}
                    {review.images && (
                      <div className="flex gap-2 flex-wrap">
                        {(typeof review.images === 'string' ? review.images.split(',').filter(Boolean) : []).map((img, idx) => (
                          <img
                            key={idx}
                            src={img}
                            alt={`评价图片${idx + 1}`}
                            className="w-16 h-16 object-cover rounded cursor-pointer hover:opacity-80"
                            onClick={() => window.open(img, '_blank')}
                          />
                        ))}
                      </div>
                    )}

                    {/* 商家回复 */}
                    {review.reply && (
                      <div className="mt-3 p-3 bg-dark-600/30 rounded-lg">
                        <div className="text-sm text-primary-400 mb-1">
                          <i className="fas fa-store mr-1" />
                          商家回复
                          <span className="text-dark-500 ml-2">
                            {review.reply_at && formatDateTime(review.reply_at)}
                          </span>
                        </div>
                        <p className="text-dark-300 text-sm">{review.reply}</p>
                      </div>
                    )}
                  </div>

                  {/* 操作按钮 */}
                  <div className="flex sm:flex-col gap-2">
                    {review.status === 0 && (
                      <>
                        <Button
                          size="sm"
                          variant="primary"
                          onClick={() => handleApprove(review.id)}
                        >
                          <i className="fas fa-check mr-1" />
                          通过
                        </Button>
                        <Button
                          size="sm"
                          variant="danger"
                          onClick={() => openRejectConfirm(review)}
                        >
                          <i className="fas fa-times mr-1" />
                          拒绝
                        </Button>
                      </>
                    )}
                    {!review.reply && review.status === 1 && (
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={() => {
                          setSelectedReview(review)
                          setReplyContent('')
                          setShowReplyModal(true)
                        }}
                      >
                        <i className="fas fa-reply mr-1" />
                        回复
                      </Button>
                    )}
                    <Button
                      size="sm"
                      variant="ghost"
                      className="text-red-400"
                      onClick={() => openDeleteConfirm(review)}
                    >
                      <i className="fas fa-trash mr-1" />
                      删除
                    </Button>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        )}

        {/* 分页 */}
        {totalPages > 1 && (
          <div className="flex justify-center items-center gap-2 mt-4 pt-4 border-t border-dark-700/50">
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
      </Card>

      {/* 回复弹窗 */}
      <Modal
        isOpen={showReplyModal}
        onClose={() => {
          setShowReplyModal(false)
          setSelectedReview(null)
          setReplyContent('')
        }}
        title="回复评价"
      >
        {selectedReview && (
          <div className="space-y-4">
            {/* 原评价 */}
            <div className="p-3 bg-dark-700/30 rounded-lg">
              <div className="flex items-center gap-2 mb-2">
                <span className="text-dark-300">{selectedReview.username}</span>
                {renderStars(selectedReview.rating)}
              </div>
              <p className="text-dark-400 text-sm">{selectedReview.content}</p>
            </div>

            {/* 回复内容 */}
            <div>
              <label className="block text-dark-300 text-sm mb-2">回复内容</label>
              <textarea
                value={replyContent}
                onChange={(e) => setReplyContent(e.target.value)}
                className="input w-full h-32 resize-none"
                placeholder="请输入回复内容..."
              />
            </div>

            <div className="flex gap-3">
              <Button
                variant="secondary"
                className="flex-1"
                onClick={() => setShowReplyModal(false)}
              >
                取消
              </Button>
              <Button variant="primary" className="flex-1" onClick={handleReply}>
                发送回复
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* 拒绝确认弹窗 */}
      <ConfirmModal
        isOpen={showRejectConfirm}
        onClose={() => { setShowRejectConfirm(false); setConfirmTarget(null) }}
        title="拒绝评价"
        message={`确定要拒绝用户 "${confirmTarget?.username}" 的评价吗？`}
        confirmText="拒绝"
        variant="warning"
        onConfirm={handleReject}
        loading={confirmLoading}
      />

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setConfirmTarget(null) }}
        title="删除评价"
        message={`确定要删除用户 "${confirmTarget?.username}" 的评价吗？此操作不可恢复。`}
        confirmText="删除"
        variant="danger"
        onConfirm={handleDelete}
        loading={confirmLoading}
      />
    </div>
  )
}
