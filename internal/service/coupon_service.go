package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"user-frontend/internal/cache"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

type CouponService struct {
	repo *repository.Repository
}

func NewCouponService(repo *repository.Repository) *CouponService {
	return &CouponService{repo: repo}
}

// ==================== 缓存辅助方法 ====================

// cacheUserCoupons 缓存用户优惠券列表
func (s *CouponService) cacheUserCoupons(userID uint, coupons []model.UserCoupon) {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	key := cache.UserCouponsKey(userID)
	data, err := json.Marshal(coupons)
	if err != nil {
		log.Printf("[CouponService] 序列化用户优惠券缓存失败: %v", err)
		return
	}

	if err := cm.Set(key, string(data), cache.CouponTTL); err != nil {
		log.Printf("[CouponService] 缓存用户优惠券失败: %v", err)
	}
}

// getUserCouponsFromCache 从缓存获取用户优惠券
func (s *CouponService) getUserCouponsFromCache(userID uint) []model.UserCoupon {
	cm := cache.GetManager()
	if cm == nil {
		return nil
	}

	key := cache.UserCouponsKey(userID)
	data, ok := cm.Get(key)
	if !ok {
		return nil
	}

	dataStr, ok := data.(string)
	if !ok {
		return nil
	}

	var coupons []model.UserCoupon
	if err := json.Unmarshal([]byte(dataStr), &coupons); err != nil {
		log.Printf("[CouponService] 反序列化用户优惠券缓存失败: %v", err)
		return nil
	}

	return coupons
}

// invalidateUserCouponsCache 使用户优惠券缓存失效
func (s *CouponService) invalidateUserCouponsCache(userID uint) {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	key := cache.UserCouponsKey(userID)
	if err := cm.Delete(key); err != nil {
		log.Printf("[CouponService] 删除用户优惠券缓存失败: %v", err)
	}
}

// invalidateAvailableCouponsCache 使可用优惠券缓存失效
func (s *CouponService) invalidateAvailableCouponsCache() {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	key := cache.AvailableCouponsKey()
	if err := cm.Delete(key); err != nil {
		log.Printf("[CouponService] 删除可用优惠券缓存失败: %v", err)
	}
}

// GenerateCouponCode 生成优惠券码
func (s *CouponService) GenerateCouponCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

// CreateCoupon 创建优惠券
func (s *CouponService) CreateCoupon(name, code, couponType string, value, minAmount, maxDiscount float64, totalCount, perUserLimit int, productIDs, categoryIDs string, startAt, endAt *time.Time) (*model.Coupon, error) {
	// 验证类型
	if couponType != "percent" && couponType != "fixed" && couponType != "minus" {
		return nil, errors.New("无效的优惠券类型")
	}

	// 验证值
	if couponType == "percent" && (value <= 0 || value > 100) {
		return nil, errors.New("折扣百分比必须在1-100之间")
	}
	if (couponType == "fixed" || couponType == "minus") && value <= 0 {
		return nil, errors.New("优惠金额必须大于0")
	}

	// 如果没有提供优惠券码，自动生成
	if code == "" {
		code = s.GenerateCouponCode(8)
	}

	// 检查优惠券码是否已存在
	existing, _ := s.repo.GetCouponByCode(code)
	if existing != nil {
		return nil, errors.New("优惠券码已存在")
	}

	coupon := &model.Coupon{
		Code:         strings.ToUpper(code),
		Name:         name,
		Type:         couponType,
		Value:        value,
		MinAmount:    minAmount,
		MaxDiscount:  maxDiscount,
		TotalCount:   totalCount,
		PerUserLimit: perUserLimit,
		ProductIDs:   productIDs,
		CategoryIDs:  categoryIDs,
		StartAt:      startAt,
		EndAt:        endAt,
		Status:       1,
	}

	if err := s.repo.CreateCoupon(coupon); err != nil {
		return nil, err
	}

	// 使可用优惠券缓存失效
	s.invalidateAvailableCouponsCache()

	return coupon, nil
}

// UpdateCoupon 更新优惠券
func (s *CouponService) UpdateCoupon(id uint, name, couponType string, value, minAmount, maxDiscount float64, totalCount, perUserLimit int, productIDs, categoryIDs string, startAt, endAt *time.Time, status int) (*model.Coupon, error) {
	coupon, err := s.repo.GetCouponByID(id)
	if err != nil {
		return nil, errors.New("优惠券不存在")
	}

	if name != "" {
		coupon.Name = name
	}
	if couponType != "" {
		coupon.Type = couponType
	}
	if value > 0 {
		coupon.Value = value
	}
	coupon.MinAmount = minAmount
	coupon.MaxDiscount = maxDiscount
	coupon.TotalCount = totalCount
	coupon.PerUserLimit = perUserLimit
	coupon.ProductIDs = productIDs
	coupon.CategoryIDs = categoryIDs
	coupon.StartAt = startAt
	coupon.EndAt = endAt
	coupon.Status = status

	if err := s.repo.UpdateCoupon(coupon); err != nil {
		return nil, err
	}

	// 使可用优惠券缓存失效
	s.invalidateAvailableCouponsCache()

	return coupon, nil
}

// DeleteCoupon 删除优惠券
func (s *CouponService) DeleteCoupon(id uint) error {
	err := s.repo.DeleteCoupon(id)
	if err == nil {
		s.invalidateAvailableCouponsCache()
	}
	return err
}

// GetAllCoupons 获取所有优惠券
func (s *CouponService) GetAllCoupons() ([]model.Coupon, error) {
	return s.repo.GetAllCoupons()
}

// GetCouponByID 根据ID获取优惠券
func (s *CouponService) GetCouponByID(id uint) (*model.Coupon, error) {
	return s.repo.GetCouponByID(id)
}

// ValidateCoupon 验证优惠券是否可用
func (s *CouponService) ValidateCoupon(code string, userID uint, productID uint, categoryID uint, orderAmount float64) (*model.Coupon, float64, error) {
	coupon, err := s.repo.GetCouponByCode(strings.ToUpper(code))
	if err != nil {
		return nil, 0, errors.New("优惠券不存在")
	}

	// 检查状态
	if coupon.Status != 1 {
		return nil, 0, errors.New("优惠券已禁用")
	}

	// 检查时间
	now := time.Now()
	if coupon.StartAt != nil && now.Before(*coupon.StartAt) {
		return nil, 0, errors.New("优惠券尚未生效")
	}
	if coupon.EndAt != nil && now.After(*coupon.EndAt) {
		return nil, 0, errors.New("优惠券已过期")
	}

	// 检查数量
	if coupon.TotalCount != -1 && coupon.UsedCount >= coupon.TotalCount {
		return nil, 0, errors.New("优惠券已被领完")
	}

	// 检查用户使用次数
	usageCount, _ := s.repo.GetUserCouponUsageCount(userID, coupon.ID)
	if int(usageCount) >= coupon.PerUserLimit {
		return nil, 0, errors.New("您已达到该优惠券的使用次数上限")
	}

	// 检查最低消费
	if orderAmount < coupon.MinAmount {
		return nil, 0, fmt.Errorf("订单金额需满%.2f元才能使用此优惠券", coupon.MinAmount)
	}

	// 检查适用商品
	if coupon.ProductIDs != "" {
		productIDList := strings.Split(coupon.ProductIDs, ",")
		found := false
		for _, pid := range productIDList {
			if pid == strconv.Itoa(int(productID)) {
				found = true
				break
			}
		}
		if !found {
			return nil, 0, errors.New("该优惠券不适用于此商品")
		}
	}

	// 检查适用分类
	if coupon.CategoryIDs != "" {
		categoryIDList := strings.Split(coupon.CategoryIDs, ",")
		found := false
		for _, cid := range categoryIDList {
			if cid == strconv.Itoa(int(categoryID)) {
				found = true
				break
			}
		}
		if !found {
			return nil, 0, errors.New("该优惠券不适用于此分类商品")
		}
	}

	// 计算优惠金额
	var discount float64
	switch coupon.Type {
	case "percent":
		discount = orderAmount * coupon.Value / 100
	case "fixed":
		discount = coupon.Value
	case "minus":
		discount = coupon.Value
	}

	// 限制最大优惠金额
	if coupon.MaxDiscount > 0 && discount > coupon.MaxDiscount {
		discount = coupon.MaxDiscount
	}

	// 优惠金额不能超过订单金额
	if discount > orderAmount {
		discount = orderAmount
	}

	return coupon, discount, nil
}

// UseCoupon 使用优惠券
// 安全特性：使用事务保护，防止并发使用同一优惠券
func (s *CouponService) UseCoupon(couponID, userID, orderID uint, orderNo string, discount float64) error {
	err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		// 加锁获取优惠券，防止并发使用
		var coupon model.Coupon
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&coupon, couponID).Error; err != nil {
			return errors.New("优惠券不存在")
		}

		// 验证优惠券可用性
		if coupon.Status != 1 {
			return errors.New("优惠券已禁用")
		}

		if coupon.TotalCount > 0 && coupon.UsedCount >= coupon.TotalCount {
			return errors.New("优惠券已用完")
		}

		if coupon.EndAt != nil && time.Now().After(*coupon.EndAt) {
			return errors.New("优惠券已过期")
		}

		// 创建使用记录
		usage := &model.CouponUsage{
			CouponID: couponID,
			UserID:   userID,
			OrderID:  orderID,
			OrderNo:  orderNo,
			Discount: discount,
		}
		if err := tx.Create(usage).Error; err != nil {
			return errors.New("创建使用记录失败")
		}

		// 增加使用次数
		if err := tx.Model(&model.Coupon{}).
			Where("id = ?", couponID).
			UpdateColumn("used_count", gorm.Expr("used_count + ?", 1)).Error; err != nil {
			return errors.New("更新使用次数失败")
		}

		return nil
	})

	if err == nil {
		// 使用户优惠券缓存失效
		s.invalidateUserCouponsCache(userID)
		s.invalidateAvailableCouponsCache()
	}

	return err
}

// GetCouponUsages 获取优惠券使用记录
func (s *CouponService) GetCouponUsages(couponID uint, page, pageSize int) ([]model.CouponUsage, int64, error) {
	return s.repo.GetCouponUsages(couponID, page, pageSize)
}

// GetCouponTypeText 获取优惠券类型文本
func GetCouponTypeText(couponType string) string {
	switch couponType {
	case "percent":
		return "折扣"
	case "fixed":
		return "固定金额"
	case "minus":
		return "满减"
	default:
		return "未知"
	}
}

// FormatCouponValue 格式化优惠券值显示
func FormatCouponValue(coupon *model.Coupon) string {
	switch coupon.Type {
	case "percent":
		return fmt.Sprintf("%.0f%%折扣", coupon.Value)
	case "fixed":
		return fmt.Sprintf("减¥%.2f", coupon.Value)
	case "minus":
		if coupon.MinAmount > 0 {
			return fmt.Sprintf("满%.0f减%.0f", coupon.MinAmount, coupon.Value)
		}
		return fmt.Sprintf("减¥%.2f", coupon.Value)
	default:
		return ""
	}
}

// GetUserCoupons 获取用户优惠券列表
// 参数：
//   - userID: 用户ID
//   - status: 状态筛选（-1表示全部）
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 用户优惠券列表
//   - 总数
//   - 错误信息（如有）
func (s *CouponService) GetUserCoupons(userID uint, status int, page, pageSize int) ([]model.UserCoupon, int64, error) {
	var total int64
	db := s.repo.GetDB().Model(&model.UserCoupon{}).Where("user_id = ?", userID)
	
	if status >= 0 {
		db = db.Where("status = ?", status)
	}
	
	db.Count(&total)

	var coupons []model.UserCoupon
	offset := (page - 1) * pageSize
	err := db.Preload("Coupon").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&coupons).Error

	return coupons, total, err
}

// GetUserCouponByID 根据ID获取用户优惠券
// 参数：
//   - userID: 用户ID
//   - couponID: 用户优惠券ID
// 返回：
//   - 用户优惠券信息
//   - 错误信息（如有）
func (s *CouponService) GetUserCouponByID(userID, couponID uint) (*model.UserCoupon, error) {
	var userCoupon model.UserCoupon
	err := s.repo.GetDB().Preload("Coupon").
		Where("id = ? AND user_id = ?", couponID, userID).
		First(&userCoupon).Error
	return &userCoupon, err
}

// UseUserCoupon 使用用户优惠券
// 参数：
//   - userCouponID: 用户优惠券ID
//   - userID: 用户ID
//   - orderNo: 订单号
// 返回：
//   - 错误信息（如有）
func (s *CouponService) UseUserCoupon(userCouponID, userID uint, orderNo string) error {
	var userCoupon model.UserCoupon
	err := s.repo.GetDB().Where("id = ? AND user_id = ?", userCouponID, userID).First(&userCoupon).Error
	if err != nil {
		return errors.New("优惠券不存在")
	}

	if userCoupon.Status != model.UserCouponStatusUnused {
		return errors.New("优惠券不可用")
	}

	// 检查是否过期
	if userCoupon.ExpireAt != nil && userCoupon.ExpireAt.Before(time.Now()) {
		s.repo.GetDB().Model(&userCoupon).Updates(map[string]interface{}{
			"status": model.UserCouponStatusExpired,
		})
		s.invalidateUserCouponsCache(userID)
		return errors.New("优惠券已过期")
	}

	// 更新状态为已使用
	now := time.Now()
	err = s.repo.GetDB().Model(&userCoupon).Updates(map[string]interface{}{
		"status":     model.UserCouponStatusUsed,
		"used_at":    &now,
		"used_order": orderNo,
	}).Error

	if err == nil {
		s.invalidateUserCouponsCache(userID)
	}

	return err
}

// GetUserAvailableCoupons 获取用户可用优惠券列表（未使用且未过期）
// 参数：
//   - userID: 用户ID
//   - orderAmount: 订单金额（用于筛选满足最低消费的优惠券）
// 返回：
//   - 可用优惠券列表
//   - 错误信息（如有）
func (s *CouponService) GetUserAvailableCoupons(userID uint, orderAmount float64) ([]model.UserCoupon, error) {
	now := time.Now()
	var coupons []model.UserCoupon
	
	err := s.repo.GetDB().Preload("Coupon").
		Where("user_id = ? AND status = ?", userID, model.UserCouponStatusUnused).
		Where("expire_at IS NULL OR expire_at > ?", now).
		Find(&coupons).Error
	
	if err != nil {
		return nil, err
	}

	// 筛选满足最低消费金额的优惠券
	if orderAmount > 0 {
		var availableCoupons []model.UserCoupon
		for _, uc := range coupons {
			if uc.Coupon != nil && orderAmount >= uc.Coupon.MinAmount {
				availableCoupons = append(availableCoupons, uc)
			}
		}
		return availableCoupons, nil
	}

	return coupons, nil
}

// ExpireUserCoupons 过期用户优惠券（定时任务调用）
// 将所有已过期但状态仍为未使用的优惠券标记为已过期
// 返回：
//   - 过期数量
//   - 错误信息（如有）
func (s *CouponService) ExpireUserCoupons() (int64, error) {
	result := s.repo.GetDB().Model(&model.UserCoupon{}).
		Where("status = ? AND expire_at IS NOT NULL AND expire_at < ?", 
			model.UserCouponStatusUnused, time.Now()).
		Updates(map[string]interface{}{
			"status": model.UserCouponStatusExpired,
		})
	return result.RowsAffected, result.Error
}
