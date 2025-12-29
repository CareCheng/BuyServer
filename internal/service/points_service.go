package service

import (
	"errors"
	"time"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// PointsService 积分服务
type PointsService struct {
	repo *repository.Repository
}

// NewPointsService 创建积分服务实例
func NewPointsService(repo *repository.Repository) *PointsService {
	return &PointsService{repo: repo}
}

// GetUserPoints 获取用户积分
// 参数：
//   - userID: 用户ID
// 返回：
//   - 用户积分信息
//   - 错误信息（如有）
func (s *PointsService) GetUserPoints(userID uint) (*model.UserPoints, error) {
	var points model.UserPoints
	result := s.repo.GetDB().Where("user_id = ?", userID).First(&points)
	if result.Error != nil {
		// 如果不存在，创建新记录
		points = model.UserPoints{
			UserID:    userID,
			Points:    0,
			TotalEarn: 0,
			TotalUsed: 0,
		}
		if err := s.repo.GetDB().Create(&points).Error; err != nil {
			return nil, err
		}
	}
	return &points, nil
}

// AddPoints 增加积分
// 参数：
//   - userID: 用户ID
//   - points: 积分数量
//   - orderNo: 关联订单号
//   - remark: 备注
// 返回：
//   - 错误信息（如有）
func (s *PointsService) AddPoints(userID uint, points int, orderNo, remark string) error {
	if points <= 0 {
		return errors.New("积分数量必须大于0")
	}

	// 获取或创建用户积分记录
	userPoints, err := s.GetUserPoints(userID)
	if err != nil {
		return err
	}

	// 更新积分
	userPoints.Points += points
	userPoints.TotalEarn += points
	if err := s.repo.GetDB().Save(userPoints).Error; err != nil {
		return err
	}

	// 记录积分变动
	log := model.PointsLog{
		UserID:  userID,
		Type:    model.PointsTypeEarn,
		Points:  points,
		Balance: userPoints.Points,
		OrderNo: orderNo,
		Remark:  remark,
	}
	return s.repo.GetDB().Create(&log).Error
}

// UsePoints 使用积分
// 参数：
//   - userID: 用户ID
//   - points: 积分数量
//   - remark: 备注
// 返回：
//   - 错误信息（如有）
func (s *PointsService) UsePoints(userID uint, points int, remark string) error {
	if points <= 0 {
		return errors.New("积分数量必须大于0")
	}

	// 获取用户积分
	userPoints, err := s.GetUserPoints(userID)
	if err != nil {
		return err
	}

	if userPoints.Points < points {
		return errors.New("积分不足")
	}

	// 扣除积分
	userPoints.Points -= points
	userPoints.TotalUsed += points
	if err := s.repo.GetDB().Save(userPoints).Error; err != nil {
		return err
	}

	// 记录积分变动
	log := model.PointsLog{
		UserID:  userID,
		Type:    model.PointsTypeUse,
		Points:  -points,
		Balance: userPoints.Points,
		Remark:  remark,
	}
	return s.repo.GetDB().Create(&log).Error
}

// AdminAdjustPoints 管理员调整积分
// 参数：
//   - userID: 用户ID
//   - points: 积分数量（正数增加，负数减少）
//   - remark: 备注
// 返回：
//   - 错误信息（如有）
func (s *PointsService) AdminAdjustPoints(userID uint, points int, remark string) error {
	// 获取用户积分
	userPoints, err := s.GetUserPoints(userID)
	if err != nil {
		return err
	}

	// 检查扣除后是否为负
	if userPoints.Points+points < 0 {
		return errors.New("调整后积分不能为负数")
	}

	// 更新积分
	userPoints.Points += points
	if points > 0 {
		userPoints.TotalEarn += points
	} else {
		userPoints.TotalUsed += -points
	}
	if err := s.repo.GetDB().Save(userPoints).Error; err != nil {
		return err
	}

	// 记录积分变动
	log := model.PointsLog{
		UserID:  userID,
		Type:    model.PointsTypeAdmin,
		Points:  points,
		Balance: userPoints.Points,
		Remark:  remark,
	}
	return s.repo.GetDB().Create(&log).Error
}

// GetPointsLogs 获取积分变动记录
// 参数：
//   - userID: 用户ID
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 积分记录列表
//   - 总数
//   - 错误信息（如有）
func (s *PointsService) GetPointsLogs(userID uint, page, pageSize int) ([]model.PointsLog, int64, error) {
	var total int64
	s.repo.GetDB().Model(&model.PointsLog{}).Where("user_id = ?", userID).Count(&total)

	var logs []model.PointsLog
	offset := (page - 1) * pageSize
	err := s.repo.GetDB().Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// CalculateOrderPoints 计算订单可获得的积分
// 参数：
//   - amount: 订单金额
// 返回：
//   - 可获得积分
func (s *PointsService) CalculateOrderPoints(amount float64) int {
	// 获取订单积分规则
	var rule model.PointsRule
	result := s.repo.GetDB().Where("type = ? AND status = 1", model.PointsRuleOrder).First(&rule)
	if result.Error != nil {
		return 0
	}

	// 检查最低消费金额
	if amount < rule.MinAmount {
		return 0
	}

	// 计算积分
	points := int(amount * rule.Ratio)
	if rule.Points > 0 {
		points += rule.Points // 加上固定积分
	}

	// 检查最高积分限制
	if rule.MaxPoints > 0 && points > rule.MaxPoints {
		points = rule.MaxPoints
	}

	return points
}

// ProcessOrderPoints 处理订单积分（订单完成后调用）
// 参数：
//   - userID: 用户ID
//   - orderNo: 订单号
//   - amount: 订单金额
// 返回：
//   - 获得的积分
//   - 错误信息（如有）
func (s *PointsService) ProcessOrderPoints(userID uint, orderNo string, amount float64) (int, error) {
	points := s.CalculateOrderPoints(amount)
	if points <= 0 {
		return 0, nil
	}

	err := s.AddPoints(userID, points, orderNo, "订单消费奖励")
	return points, err
}

// PointsRuleInfo 积分规则信息
type PointsRuleInfo struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Points      int     `json:"points"`
	Ratio       float64 `json:"ratio"`
	MinAmount   float64 `json:"min_amount"`
	MaxPoints   int     `json:"max_points"`
	Status      int     `json:"status"`
	Description string  `json:"description"`
}

// GetPointsRules 获取积分规则列表
// 返回：
//   - 规则列表
//   - 错误信息（如有）
func (s *PointsService) GetPointsRules() ([]PointsRuleInfo, error) {
	var rules []model.PointsRule
	if err := s.repo.GetDB().Order("id ASC").Find(&rules).Error; err != nil {
		return nil, err
	}

	result := make([]PointsRuleInfo, len(rules))
	for i, r := range rules {
		result[i] = PointsRuleInfo{
			ID:          r.ID,
			Name:        r.Name,
			Type:        r.Type,
			Points:      r.Points,
			Ratio:       r.Ratio,
			MinAmount:   r.MinAmount,
			MaxPoints:   r.MaxPoints,
			Status:      r.Status,
			Description: r.Description,
		}
	}
	return result, nil
}

// CreatePointsRule 创建积分规则
// 参数：
//   - rule: 规则信息
// 返回：
//   - 错误信息（如有）
func (s *PointsService) CreatePointsRule(rule *model.PointsRule) error {
	return s.repo.GetDB().Create(rule).Error
}

// UpdatePointsRule 更新积分规则
// 参数：
//   - rule: 规则信息
// 返回：
//   - 错误信息（如有）
func (s *PointsService) UpdatePointsRule(rule *model.PointsRule) error {
	return s.repo.GetDB().Save(rule).Error
}

// DeletePointsRule 删除积分规则
// 参数：
//   - ruleID: 规则ID
// 返回：
//   - 错误信息（如有）
func (s *PointsService) DeletePointsRule(ruleID uint) error {
	return s.repo.GetDB().Delete(&model.PointsRule{}, ruleID).Error
}

// ExchangeInfo 可兑换商品信息
type ExchangeInfo struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Points      int    `json:"points"`
	Description string `json:"description"`
	Stock       int    `json:"stock"`
}

// GetExchangeList 获取可兑换列表
// 返回：
//   - 可兑换商品列表
//   - 错误信息（如有）
func (s *PointsService) GetExchangeList() ([]ExchangeInfo, error) {
	// 获取可用积分兑换的优惠券
	var coupons []model.Coupon
	s.repo.GetDB().Where("status = 1 AND points_price > 0").Find(&coupons)

	result := make([]ExchangeInfo, 0)
	for _, c := range coupons {
		result = append(result, ExchangeInfo{
			ID:          c.ID,
			Type:        "coupon",
			Name:        c.Name,
			Points:      c.PointsPrice,
			Description: c.Description,
			Stock:       c.Stock,
		})
	}

	return result, nil
}

// ExchangeCoupon 积分兑换优惠券
// 参数：
//   - userID: 用户ID
//   - couponID: 优惠券ID
// 返回：
//   - 兑换记录
//   - 错误信息（如有）
//
// 安全特性：使用事务保护，防止并发兑换导致的积分透支或库存超卖
func (s *PointsService) ExchangeCoupon(userID, couponID uint) (*model.PointsExchange, error) {
	var exchange *model.PointsExchange

	// 使用事务保证原子性
	err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		// 获取优惠券信息（加锁防止并发）
		var coupon model.Coupon
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&coupon, couponID).Error; err != nil {
			return errors.New("优惠券不存在")
		}

		if coupon.Status != 1 {
			return errors.New("优惠券已下架")
		}

		if coupon.PointsPrice <= 0 {
			return errors.New("该优惠券不支持积分兑换")
		}

		if coupon.Stock == 0 {
			return errors.New("优惠券库存不足")
		}

		// 检查是否已过期
		if coupon.ExpireAt != nil && coupon.ExpireAt.Before(time.Now()) {
			return errors.New("优惠券已过期")
		}

		// 获取用户积分（加锁防止并发）
		var userPoints model.UserPoints
		result := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&userPoints)
		if result.Error != nil {
			// 如果不存在，创建新记录
			userPoints = model.UserPoints{
				UserID:    userID,
				Points:    0,
				TotalEarn: 0,
				TotalUsed: 0,
			}
			if err := tx.Create(&userPoints).Error; err != nil {
				return errors.New("获取用户积分失败")
			}
		}

		// 检查积分是否充足
		if userPoints.Points < coupon.PointsPrice {
			return errors.New("积分不足")
		}

		// 扣除积分
		userPoints.Points -= coupon.PointsPrice
		userPoints.TotalUsed += coupon.PointsPrice
		if err := tx.Save(&userPoints).Error; err != nil {
			return errors.New("扣除积分失败")
		}

		// 记录积分变动
		pointsLog := model.PointsLog{
			UserID:  userID,
			Type:    model.PointsTypeUse,
			Points:  -coupon.PointsPrice,
			Balance: userPoints.Points,
			Remark:  "兑换优惠券: " + coupon.Name,
		}
		if err := tx.Create(&pointsLog).Error; err != nil {
			return errors.New("记录积分变动失败")
		}

		// 减少库存
		if coupon.Stock > 0 {
			coupon.Stock--
			if err := tx.Save(&coupon).Error; err != nil {
				return errors.New("更新优惠券库存失败")
			}
		}

		// 创建兑换记录
		exchange = &model.PointsExchange{
			UserID:     userID,
			Points:     coupon.PointsPrice,
			Type:       "coupon",
			TargetID:   couponID,
			TargetName: coupon.Name,
			Status:     1,
		}
		if err := tx.Create(exchange).Error; err != nil {
			return errors.New("创建兑换记录失败")
		}

		// 计算优惠券过期时间
		var expireAt *time.Time
		if coupon.ExpireAt != nil {
			expireAt = coupon.ExpireAt
		}

		// 发放优惠券给用户
		userCoupon := &model.UserCoupon{
			UserID:     userID,
			CouponID:   couponID,
			CouponCode: coupon.Code,
			CouponName: coupon.Name,
			Source:     model.UserCouponSourceExchange,
			SourceID:   exchange.ID,
			Status:     model.UserCouponStatusUnused,
			ExpireAt:   expireAt,
		}
		if err := tx.Create(userCoupon).Error; err != nil {
			return errors.New("发放优惠券失败")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return exchange, nil
}

// GetUserExchanges 获取用户兑换记录
// 参数：
//   - userID: 用户ID
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 兑换记录列表
//   - 总数
//   - 错误信息（如有）
func (s *PointsService) GetUserExchanges(userID uint, page, pageSize int) ([]model.PointsExchange, int64, error) {
	var total int64
	s.repo.GetDB().Model(&model.PointsExchange{}).Where("user_id = ?", userID).Count(&total)

	var exchanges []model.PointsExchange
	offset := (page - 1) * pageSize
	err := s.repo.GetDB().Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&exchanges).Error

	return exchanges, total, err
}

// UserPointsInfo 用户积分信息（管理员视图）
type UserPointsInfo struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Points    int    `json:"points"`
	TotalEarn int    `json:"total_earn"`
	TotalUsed int    `json:"total_used"`
	UpdatedAt string `json:"updated_at"`
}

// AdminGetAllUserPoints 管理员获取所有用户积分列表
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//   - keyword: 搜索关键词（用户名）
// 返回：
//   - 用户积分列表
//   - 总数
//   - 错误信息（如有）
func (s *PointsService) AdminGetAllUserPoints(page, pageSize int, keyword string) ([]UserPointsInfo, int64, error) {
	var total int64
	query := s.repo.GetDB().Model(&model.UserPoints{})

	// 如果有关键词，需要关联用户表搜索
	if keyword != "" {
		query = query.Joins("JOIN users ON users.id = user_points.user_id").
			Where("users.username LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)

	var userPoints []model.UserPoints
	offset := (page - 1) * pageSize

	if keyword != "" {
		err := s.repo.GetDB().Joins("JOIN users ON users.id = user_points.user_id").
			Where("users.username LIKE ?", "%"+keyword+"%").
			Order("user_points.updated_at DESC").
			Offset(offset).Limit(pageSize).
			Find(&userPoints).Error
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := s.repo.GetDB().Order("updated_at DESC").
			Offset(offset).Limit(pageSize).
			Find(&userPoints).Error
		if err != nil {
			return nil, 0, err
		}
	}

	// 获取用户名
	result := make([]UserPointsInfo, len(userPoints))
	for i, up := range userPoints {
		var user model.User
		s.repo.GetDB().Select("username").First(&user, up.UserID)
		result[i] = UserPointsInfo{
			ID:        up.ID,
			UserID:    up.UserID,
			Username:  user.Username,
			Points:    up.Points,
			TotalEarn: up.TotalEarn,
			TotalUsed: up.TotalUsed,
			UpdatedAt: up.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return result, total, nil
}

// PointsLogInfo 积分记录信息（管理员视图）
type PointsLogInfo struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Type      string `json:"type"`
	Points    int    `json:"points"`
	Balance   int    `json:"balance"`
	OrderNo   string `json:"order_no"`
	Remark    string `json:"remark"`
	CreatedAt string `json:"created_at"`
}

// AdminGetPointsLogs 管理员获取积分变动记录
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//   - userID: 用户ID（可选，0表示全部）
// 返回：
//   - 积分记录列表
//   - 总数
//   - 错误信息（如有）
func (s *PointsService) AdminGetPointsLogs(page, pageSize int, userID uint) ([]PointsLogInfo, int64, error) {
	var total int64
	query := s.repo.GetDB().Model(&model.PointsLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	query.Count(&total)

	var logs []model.PointsLog
	offset := (page - 1) * pageSize

	logQuery := s.repo.GetDB().Order("created_at DESC").Offset(offset).Limit(pageSize)
	if userID > 0 {
		logQuery = logQuery.Where("user_id = ?", userID)
	}

	if err := logQuery.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	// 获取用户名
	result := make([]PointsLogInfo, len(logs))
	userCache := make(map[uint]string)

	for i, log := range logs {
		username, ok := userCache[log.UserID]
		if !ok {
			var user model.User
			s.repo.GetDB().Select("username").First(&user, log.UserID)
			username = user.Username
			userCache[log.UserID] = username
		}

		result[i] = PointsLogInfo{
			ID:        log.ID,
			UserID:    log.UserID,
			Username:  username,
			Type:      log.Type,
			Points:    log.Points,
			Balance:   log.Balance,
			OrderNo:   log.OrderNo,
			Remark:    log.Remark,
			CreatedAt: log.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return result, total, nil
}
