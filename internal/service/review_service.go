package service

import (
	"encoding/json"
	"errors"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// ReviewService 商品评价服务
type ReviewService struct {
	repo *repository.Repository
}

// NewReviewService 创建商品评价服务实例
func NewReviewService(repo *repository.Repository) *ReviewService {
	return &ReviewService{repo: repo}
}

// CreateReview 创建商品评价
// 参数：
//   - userID: 用户ID
//   - username: 用户名
//   - orderNo: 订单号
//   - productID: 商品ID
//   - rating: 评分（1-5）
//   - content: 评价内容
//   - images: 评价图片
//   - isAnon: 是否匿名
// 返回：
//   - 创建的评价
//   - 错误信息
func (s *ReviewService) CreateReview(userID uint, username, orderNo string, productID uint, rating int, content string, images []string, isAnon bool) (*model.ProductReview, error) {
	// 验证评分范围
	if rating < 1 || rating > 5 {
		return nil, errors.New("评分必须在1-5之间")
	}

	// 验证订单是否存在且属于该用户
	var order model.Order
	if err := s.repo.GetDB().Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		return nil, err
	}

	// 验证订单状态（必须是已完成的订单）
	if order.Status != 2 {
		return nil, errors.New("只能评价已完成的订单")
	}

	// 验证商品ID是否匹配
	if order.ProductID != productID {
		return nil, errors.New("商品ID与订单不匹配")
	}

	// 检查是否已评价
	var existingReview model.ProductReview
	if err := s.repo.GetDB().Where("order_no = ?", orderNo).First(&existingReview).Error; err == nil {
		return nil, errors.New("该订单已评价")
	}

	// 序列化图片数组
	imagesJSON := "[]"
	if len(images) > 0 {
		if data, err := json.Marshal(images); err == nil {
			imagesJSON = string(data)
		}
	}

	// 创建评价
	review := &model.ProductReview{
		ProductID: productID,
		UserID:    userID,
		Username:  username,
		OrderNo:   orderNo,
		Rating:    rating,
		Content:   content,
		Images:    imagesJSON,
		IsAnon:    isAnon,
		Status:    1,
	}

	if err := s.repo.GetDB().Create(review).Error; err != nil {
		return nil, err
	}

	return review, nil
}

// GetProductReviews 获取商品评价列表
// 参数：
//   - productID: 商品ID
//   - page: 页码
//   - pageSize: 每页数量
//   - rating: 筛选评分（0表示全部）
// 返回：
//   - 评价列表
//   - 总数
//   - 错误信息
func (s *ReviewService) GetProductReviews(productID uint, page, pageSize, rating int) ([]model.ProductReview, int64, error) {
	var reviews []model.ProductReview
	var total int64

	query := s.repo.GetDB().Model(&model.ProductReview{}).Where("product_id = ? AND status = 1", productID)

	// 筛选评分
	if rating > 0 && rating <= 5 {
		query = query.Where("rating = ?", rating)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	// 处理匿名用户名
	for i := range reviews {
		if reviews[i].IsAnon {
			reviews[i].Username = maskUsername(reviews[i].Username)
		}
	}

	return reviews, total, nil
}

// GetProductReviewStats 获取商品评价统计
// 参数：
//   - productID: 商品ID
// 返回：
//   - 评价统计
//   - 错误信息
func (s *ReviewService) GetProductReviewStats(productID uint) (*model.ProductReviewStats, error) {
	stats := &model.ProductReviewStats{ProductID: productID}

	// 统计总数和平均分
	var result struct {
		TotalCount int64
		AvgRating  float64
	}
	if err := s.repo.GetDB().Model(&model.ProductReview{}).
		Where("product_id = ? AND status = 1", productID).
		Select("COUNT(*) as total_count, COALESCE(AVG(rating), 0) as avg_rating").
		Scan(&result).Error; err != nil {
		return nil, err
	}
	stats.TotalCount = result.TotalCount
	stats.AvgRating = result.AvgRating

	// 统计各评分数量
	var ratingCounts []struct {
		Rating int
		Count  int64
	}
	if err := s.repo.GetDB().Model(&model.ProductReview{}).
		Where("product_id = ? AND status = 1", productID).
		Select("rating, COUNT(*) as count").
		Group("rating").
		Scan(&ratingCounts).Error; err != nil {
		return nil, err
	}

	for _, rc := range ratingCounts {
		switch rc.Rating {
		case 5:
			stats.Rating5Count = rc.Count
		case 4:
			stats.Rating4Count = rc.Count
		case 3:
			stats.Rating3Count = rc.Count
		case 2:
			stats.Rating2Count = rc.Count
		case 1:
			stats.Rating1Count = rc.Count
		}
	}

	return stats, nil
}

// GetUserReviews 获取用户的评价列表
// 参数：
//   - userID: 用户ID
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 评价列表
//   - 总数
//   - 错误信息
func (s *ReviewService) GetUserReviews(userID uint, page, pageSize int) ([]model.ProductReview, int64, error) {
	var reviews []model.ProductReview
	var total int64

	query := s.repo.GetDB().Model(&model.ProductReview{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// CheckCanReview 检查用户是否可以评价订单
// 参数：
//   - userID: 用户ID
//   - orderNo: 订单号
// 返回：
//   - 是否可以评价
//   - 原因
func (s *ReviewService) CheckCanReview(userID uint, orderNo string) (bool, string) {
	// 检查订单是否存在
	var order model.Order
	if err := s.repo.GetDB().Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		return false, "订单不存在"
	}

	// 检查订单状态
	if order.Status != 2 {
		return false, "只能评价已完成的订单"
	}

	// 检查是否已评价
	var count int64
	s.repo.GetDB().Model(&model.ProductReview{}).Where("order_no = ?", orderNo).Count(&count)
	if count > 0 {
		return false, "该订单已评价"
	}

	return true, ""
}

// ReplyReview 商家回复评价（管理员功能）
// 参数：
//   - reviewID: 评价ID
//   - reply: 回复内容
// 返回：
//   - 错误信息
func (s *ReviewService) ReplyReview(reviewID uint, reply string) error {
	now := time.Now()
	return s.repo.GetDB().Model(&model.ProductReview{}).
		Where("id = ?", reviewID).
		Updates(map[string]interface{}{
			"reply":    reply,
			"reply_at": &now,
		}).Error
}

// UpdateReviewStatus 更新评价状态（管理员功能）
// 参数：
//   - reviewID: 评价ID
//   - status: 状态（1显示 0隐藏）
// 返回：
//   - 错误信息
func (s *ReviewService) UpdateReviewStatus(reviewID uint, status int) error {
	return s.repo.GetDB().Model(&model.ProductReview{}).
		Where("id = ?", reviewID).
		Update("status", status).Error
}

// DeleteReview 删除评价（管理员功能）
// 参数：
//   - reviewID: 评价ID
// 返回：
//   - 错误信息
func (s *ReviewService) DeleteReview(reviewID uint) error {
	return s.repo.GetDB().Delete(&model.ProductReview{}, reviewID).Error
}

// GetAllReviews 获取所有评价（管理员功能）
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//   - productID: 商品ID筛选（0表示全部）
//   - status: 状态筛选（-1表示全部）
// 返回：
//   - 评价列表
//   - 总数
//   - 错误信息
func (s *ReviewService) GetAllReviews(page, pageSize int, productID uint, status int) ([]model.ProductReview, int64, error) {
	var reviews []model.ProductReview
	var total int64

	query := s.repo.GetDB().Model(&model.ProductReview{})

	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// maskUsername 隐藏用户名中间部分
func maskUsername(username string) string {
	runes := []rune(username)
	length := len(runes)
	if length <= 2 {
		return string(runes[0]) + "***"
	}
	return string(runes[0]) + "***" + string(runes[length-1])
}
