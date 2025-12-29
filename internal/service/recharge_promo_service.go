// Package service 提供业务逻辑服务
// recharge_promo_service.go - 充值优惠活动服务
package service

import (
	"errors"
	"sort"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// RechargePromoService 充值优惠服务
type RechargePromoService struct {
	repo *repository.Repository
}

// NewRechargePromoService 创建充值优惠服务
func NewRechargePromoService(repo *repository.Repository) *RechargePromoService {
	return &RechargePromoService{repo: repo}
}

// PromoResult 优惠计算结果
type PromoResult struct {
	PromoID        uint    `json:"promo_id"`         // 优惠活动ID
	PromoName      string  `json:"promo_name"`       // 优惠活动名称
	PromoType      string  `json:"promo_type"`       // 优惠类型
	OriginalAmount float64 `json:"original_amount"`  // 原始充值金额
	PayAmount      float64 `json:"pay_amount"`       // 实际支付金额
	BonusAmount    float64 `json:"bonus_amount"`     // 赠送金额
	DiscountAmount float64 `json:"discount_amount"`  // 折扣金额
	TotalCredit    float64 `json:"total_credit"`     // 总到账金额
}

// ==================== 管理员功能 ====================

// CreatePromo 创建优惠活动
func (s *RechargePromoService) CreatePromo(promo *model.RechargePromo) error {
	// 验证参数
	if promo.Name == "" {
		return errors.New("活动名称不能为空")
	}
	if promo.PromoType == "" {
		return errors.New("优惠类型不能为空")
	}
	if promo.MinAmount < 0 {
		return errors.New("最低充值金额不能为负数")
	}
	if promo.Value <= 0 {
		return errors.New("优惠值必须大于0")
	}
	// 折扣类型验证
	if promo.PromoType == model.PromoTypeDiscount {
		if promo.Value <= 0 || promo.Value >= 1 {
			return errors.New("折扣率必须在0到1之间（如0.9表示9折）")
		}
	}

	return s.repo.GetDB().Create(promo).Error
}

// UpdatePromo 更新优惠活动
func (s *RechargePromoService) UpdatePromo(promo *model.RechargePromo) error {
	// 验证参数
	if promo.Name == "" {
		return errors.New("活动名称不能为空")
	}
	if promo.PromoType == model.PromoTypeDiscount {
		if promo.Value <= 0 || promo.Value >= 1 {
			return errors.New("折扣率必须在0到1之间（如0.9表示9折）")
		}
	}

	return s.repo.GetDB().Save(promo).Error
}

// DeletePromo 删除优惠活动
func (s *RechargePromoService) DeletePromo(id uint) error {
	return s.repo.GetDB().Delete(&model.RechargePromo{}, id).Error
}

// GetPromoByID 根据ID获取优惠活动
func (s *RechargePromoService) GetPromoByID(id uint) (*model.RechargePromo, error) {
	var promo model.RechargePromo
	err := s.repo.GetDB().First(&promo, id).Error
	return &promo, err
}

// GetAllPromos 获取所有优惠活动（管理员）
func (s *RechargePromoService) GetAllPromos(page, pageSize int, status int) ([]model.RechargePromo, int64, error) {
	db := s.repo.GetDB()
	var promos []model.RechargePromo
	var total int64

	query := db.Model(&model.RechargePromo{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("priority DESC, id DESC").Offset(offset).Limit(pageSize).Find(&promos).Error
	return promos, total, err
}

// TogglePromoStatus 切换优惠活动状态
func (s *RechargePromoService) TogglePromoStatus(id uint) error {
	var promo model.RechargePromo
	if err := s.repo.GetDB().First(&promo, id).Error; err != nil {
		return err
	}
	newStatus := 1
	if promo.Status == 1 {
		newStatus = 0
	}
	return s.repo.GetDB().Model(&promo).Update("status", newStatus).Error
}

// GetPromoUsages 获取优惠使用记录
func (s *RechargePromoService) GetPromoUsages(promoID uint, page, pageSize int) ([]model.RechargePromoUsage, int64, error) {
	db := s.repo.GetDB()
	var usages []model.RechargePromoUsage
	var total int64

	query := db.Model(&model.RechargePromoUsage{})
	if promoID > 0 {
		query = query.Where("promo_id = ?", promoID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&usages).Error
	return usages, total, err
}

// GetPromoStats 获取优惠统计
func (s *RechargePromoService) GetPromoStats() (map[string]interface{}, error) {
	db := s.repo.GetDB()

	// 活动总数
	var totalPromos int64
	db.Model(&model.RechargePromo{}).Count(&totalPromos)

	// 启用中的活动数
	var activePromos int64
	now := time.Now()
	db.Model(&model.RechargePromo{}).
		Where("status = ? AND (start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", 1, now, now).
		Count(&activePromos)

	// 总使用次数
	var totalUsages int64
	db.Model(&model.RechargePromoUsage{}).Count(&totalUsages)

	// 总赠送金额
	var totalBonus float64
	db.Model(&model.RechargePromoUsage{}).Select("COALESCE(SUM(bonus_amount), 0)").Scan(&totalBonus)

	// 总折扣金额
	var totalDiscount float64
	db.Model(&model.RechargePromoUsage{}).Select("COALESCE(SUM(discount_amount), 0)").Scan(&totalDiscount)

	return map[string]interface{}{
		"total_promos":   totalPromos,
		"active_promos":  activePromos,
		"total_usages":   totalUsages,
		"total_bonus":    totalBonus,
		"total_discount": totalDiscount,
	}, nil
}

// ==================== 用户端功能 ====================

// GetActivePromos 获取当前有效的优惠活动（用户端）
func (s *RechargePromoService) GetActivePromos() ([]model.RechargePromo, error) {
	db := s.repo.GetDB()
	var promos []model.RechargePromo
	now := time.Now()

	err := db.Where("status = ? AND (start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", 1, now, now).
		Where("total_limit = 0 OR used_count < total_limit").
		Order("priority DESC, min_amount ASC").
		Find(&promos).Error

	return promos, err
}

// CalculatePromo 计算充值优惠
// 返回最优的优惠方案
func (s *RechargePromoService) CalculatePromo(userID uint, rechargeAmount float64) (*PromoResult, error) {
	// 获取所有有效的优惠活动
	promos, err := s.GetActivePromos()
	if err != nil {
		return nil, err
	}

	if len(promos) == 0 {
		// 没有优惠活动
		return &PromoResult{
			OriginalAmount: rechargeAmount,
			PayAmount:      rechargeAmount,
			BonusAmount:    0,
			DiscountAmount: 0,
			TotalCredit:    rechargeAmount,
		}, nil
	}

	// 筛选符合条件的优惠
	var validPromos []model.RechargePromo
	for _, promo := range promos {
		// 检查金额门槛
		if rechargeAmount < promo.MinAmount {
			continue
		}
		if promo.MaxAmount > 0 && rechargeAmount > promo.MaxAmount {
			continue
		}
		// 检查用户使用次数限制
		if promo.PerUserLimit > 0 {
			usageCount := s.getUserPromoUsageCount(userID, promo.ID)
			if usageCount >= promo.PerUserLimit {
				continue
			}
		}
		validPromos = append(validPromos, promo)
	}

	if len(validPromos) == 0 {
		return &PromoResult{
			OriginalAmount: rechargeAmount,
			PayAmount:      rechargeAmount,
			BonusAmount:    0,
			DiscountAmount: 0,
			TotalCredit:    rechargeAmount,
		}, nil
	}

	// 计算每个优惠的收益，选择最优的
	var bestResult *PromoResult
	var bestBenefit float64 = 0

	for _, promo := range validPromos {
		bonus := promo.CalculateBonus(rechargeAmount)
		discount := promo.CalculateDiscount(rechargeAmount)
		payAmount := rechargeAmount - discount
		totalCredit := rechargeAmount + bonus // 到账金额 = 充值金额 + 赠金（折扣不影响到账）

		// 计算总收益 = 赠金 + 折扣
		benefit := bonus + discount

		if benefit > bestBenefit {
			bestBenefit = benefit
			bestResult = &PromoResult{
				PromoID:        promo.ID,
				PromoName:      promo.Name,
				PromoType:      promo.PromoType,
				OriginalAmount: rechargeAmount,
				PayAmount:      payAmount,
				BonusAmount:    bonus,
				DiscountAmount: discount,
				TotalCredit:    totalCredit,
			}
		}
	}

	if bestResult == nil {
		return &PromoResult{
			OriginalAmount: rechargeAmount,
			PayAmount:      rechargeAmount,
			BonusAmount:    0,
			DiscountAmount: 0,
			TotalCredit:    rechargeAmount,
		}, nil
	}

	return bestResult, nil
}

// CalculateAllPromos 计算所有适用的优惠（用于展示）
func (s *RechargePromoService) CalculateAllPromos(userID uint, rechargeAmount float64) ([]PromoResult, error) {
	promos, err := s.GetActivePromos()
	if err != nil {
		return nil, err
	}

	var results []PromoResult
	for _, promo := range promos {
		// 检查金额门槛
		if rechargeAmount < promo.MinAmount {
			continue
		}
		if promo.MaxAmount > 0 && rechargeAmount > promo.MaxAmount {
			continue
		}
		// 检查用户使用次数限制
		if promo.PerUserLimit > 0 {
			usageCount := s.getUserPromoUsageCount(userID, promo.ID)
			if usageCount >= promo.PerUserLimit {
				continue
			}
		}

		bonus := promo.CalculateBonus(rechargeAmount)
		discount := promo.CalculateDiscount(rechargeAmount)
		payAmount := rechargeAmount - discount
		totalCredit := rechargeAmount + bonus

		results = append(results, PromoResult{
			PromoID:        promo.ID,
			PromoName:      promo.Name,
			PromoType:      promo.PromoType,
			OriginalAmount: rechargeAmount,
			PayAmount:      payAmount,
			BonusAmount:    bonus,
			DiscountAmount: discount,
			TotalCredit:    totalCredit,
		})
	}

	// 按收益排序（赠金+折扣）
	sort.Slice(results, func(i, j int) bool {
		benefitI := results[i].BonusAmount + results[i].DiscountAmount
		benefitJ := results[j].BonusAmount + results[j].DiscountAmount
		return benefitI > benefitJ
	})

	return results, nil
}

// RecordPromoUsage 记录优惠使用
func (s *RechargePromoService) RecordPromoUsage(promoID, userID uint, rechargeNo string, amount, bonusAmount, discountAmount float64) error {
	db := s.repo.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		// 创建使用记录
		usage := &model.RechargePromoUsage{
			PromoID:        promoID,
			UserID:         userID,
			RechargeNo:     rechargeNo,
			Amount:         amount,
			BonusAmount:    bonusAmount,
			DiscountAmount: discountAmount,
		}
		if err := tx.Create(usage).Error; err != nil {
			return err
		}

		// 更新活动使用次数
		return tx.Model(&model.RechargePromo{}).
			Where("id = ?", promoID).
			Update("used_count", gorm.Expr("used_count + 1")).Error
	})
}

// getUserPromoUsageCount 获取用户对某活动的使用次数
func (s *RechargePromoService) getUserPromoUsageCount(userID, promoID uint) int {
	var count int64
	s.repo.GetDB().Model(&model.RechargePromoUsage{}).
		Where("user_id = ? AND promo_id = ?", userID, promoID).
		Count(&count)
	return int(count)
}

// GetUserPromoUsages 获取用户的优惠使用记录
func (s *RechargePromoService) GetUserPromoUsages(userID uint, page, pageSize int) ([]model.RechargePromoUsage, int64, error) {
	db := s.repo.GetDB()
	var usages []model.RechargePromoUsage
	var total int64

	query := db.Model(&model.RechargePromoUsage{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&usages).Error
	return usages, total, err
}
