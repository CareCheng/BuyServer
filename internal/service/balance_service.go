package service

import (
	"errors"
	"fmt"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// BalanceService 余额服务
type BalanceService struct {
	repo      *repository.Repository
	configSvc *ConfigService        // 配置服务引用
	promoSvc  *RechargePromoService // 充值优惠服务引用
}

// OperatorInfo 操作者信息（用于余额变动日志）
type OperatorInfo struct {
	OperatorID   uint   // 操作者ID（管理员ID或用户ID）
	OperatorType string // 操作者类型：user/admin/system
	ClientIP     string // 客户端IP
}

// 余额系统安全限制默认值（当配置服务不可用时使用）
const (
	DefaultMaxRechargeAmount = 50000.0  // 单笔充值最大金额（元）
	DefaultMinRechargeAmount = 1.0      // 单笔充值最小金额（元）
	DefaultMaxBalanceLimit   = 100000.0 // 用户余额上限（元）
	DefaultMaxDailyRecharge  = 100000.0 // 每日充值上限（元）
)

// NewBalanceService 创建余额服务
func NewBalanceService(repo *repository.Repository) *BalanceService {
	return &BalanceService{repo: repo}
}

// SetConfigService 设置配置服务引用
func (s *BalanceService) SetConfigService(configSvc *ConfigService) {
	s.configSvc = configSvc
}

// SetPromoService 设置充值优惠服务引用
func (s *BalanceService) SetPromoService(promoSvc *RechargePromoService) {
	s.promoSvc = promoSvc
}

// getBalanceLimits 获取余额限制配置
func (s *BalanceService) getBalanceLimits() (minRecharge, maxRecharge, maxDaily, maxBalance float64) {
	if s.configSvc != nil {
		return s.configSvc.GetBalanceLimits()
	}
	return DefaultMinRechargeAmount, DefaultMaxRechargeAmount, DefaultMaxDailyRecharge, DefaultMaxBalanceLimit
}

// GetUserBalance 获取用户余额
func (s *BalanceService) GetUserBalance(userID uint) (*model.UserBalance, error) {
	db := s.repo.GetDB()
	var balance model.UserBalance
	err := db.Where("user_id = ?", userID).First(&balance).Error
	if err == gorm.ErrRecordNotFound {
		// 如果不存在，创建一个新的余额记录
		balance = model.UserBalance{
			UserID:  userID,
			Balance: 0,
			Frozen:  0,
		}
		if err := db.Create(&balance).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetBalanceLogs 获取余额变动记录
func (s *BalanceService) GetBalanceLogs(userID uint, page, pageSize int, logType string) ([]model.BalanceLog, int64, error) {
	db := s.repo.GetDB()
	var logs []model.BalanceLog
	var total int64

	query := db.Model(&model.BalanceLog{}).Where("user_id = ?", userID)
	if logType != "" {
		query = query.Where("type = ?", logType)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// Recharge 充值（增加余额）
// 使用 FOR UPDATE 锁定行防止并发问题，并检查余额上限
func (s *BalanceService) Recharge(userID uint, amount float64, rechargeNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("充值金额必须大于0")
	}

	// 获取余额限制配置
	_, _, _, maxBalance := s.getBalanceLimits()

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// 使用 FOR UPDATE 锁定行
		var balance model.UserBalance
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).First(&balance).Error
		if err == gorm.ErrRecordNotFound {
			balance = model.UserBalance{
				UserID:  userID,
				Balance: 0,
				Frozen:  0,
			}
			if err := tx.Create(&balance).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// 检查余额上限
		if balance.Balance+amount > maxBalance {
			return fmt.Errorf("充值后余额将超过上限 %.2f 元", maxBalance)
		}

		beforeBalance := balance.Balance

		// 使用原子更新
		if err := tx.Model(&model.UserBalance{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"balance":  gorm.Expr("balance + ?", amount),
				"total_in": gorm.Expr("total_in + ?", amount),
			}).Error; err != nil {
			return err
		}

		// 记录变动日志
		log := &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeRecharge,
			Amount:        amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  beforeBalance + amount,
			RechargeNo:    rechargeNo,
			Remark:        remark,
		}
		// 填充操作者信息
		if operator != nil {
			log.OperatorID = operator.OperatorID
			log.OperatorType = operator.OperatorType
			log.ClientIP = operator.ClientIP
		} else {
			log.OperatorType = "system"
		}
		return tx.Create(log).Error
	})
}

// Consume 消费（扣减余额）
// 使用原子更新防止并发竞态条件
func (s *BalanceService) Consume(userID uint, amount float64, orderNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("消费金额必须大于0")
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// 使用 FOR UPDATE 锁定行，防止并发读取
		var balance model.UserBalance
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).First(&balance).Error
		if err != nil {
			return errors.New("余额记录不存在")
		}

		if balance.Balance < amount {
			return errors.New("余额不足")
		}

		beforeBalance := balance.Balance

		// 使用原子更新，确保余额充足时才扣减
		result := tx.Model(&model.UserBalance{}).
			Where("user_id = ? AND balance >= ?", userID, amount).
			Updates(map[string]interface{}{
				"balance":   gorm.Expr("balance - ?", amount),
				"total_out": gorm.Expr("total_out + ?", amount),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("余额不足或更新失败")
		}

		// 记录变动日志
		log := &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeConsume,
			Amount:        -amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  beforeBalance - amount,
			OrderNo:       orderNo,
			Remark:        remark,
		}
		// 填充操作者信息
		if operator != nil {
			log.OperatorID = operator.OperatorID
			log.OperatorType = operator.OperatorType
			log.ClientIP = operator.ClientIP
		} else {
			log.OperatorType = "user"
		}
		return tx.Create(log).Error
	})
}

// Refund 退款（返还余额）
// 使用 FOR UPDATE 锁定行防止并发问题
func (s *BalanceService) Refund(userID uint, amount float64, orderNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("退款金额必须大于0")
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// 使用 FOR UPDATE 锁定行
		var balance model.UserBalance
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).First(&balance).Error
		if err == gorm.ErrRecordNotFound {
			balance = model.UserBalance{
				UserID:  userID,
				Balance: 0,
				Frozen:  0,
			}
			if err := tx.Create(&balance).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		beforeBalance := balance.Balance

		// 使用原子更新
		if err := tx.Model(&model.UserBalance{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"balance":   gorm.Expr("balance + ?", amount),
				"total_out": gorm.Expr("total_out - ?", amount),
			}).Error; err != nil {
			return err
		}

		log := &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeRefund,
			Amount:        amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  beforeBalance + amount,
			OrderNo:       orderNo,
			Remark:        remark,
		}
		// 填充操作者信息
		if operator != nil {
			log.OperatorID = operator.OperatorID
			log.OperatorType = operator.OperatorType
			log.ClientIP = operator.ClientIP
		} else {
			log.OperatorType = "system"
		}
		return tx.Create(log).Error
	})
}

// Freeze 冻结余额
// 使用原子更新防止并发竞态条件
func (s *BalanceService) Freeze(userID uint, amount float64, orderNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("冻结金额必须大于0")
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// 使用 FOR UPDATE 锁定行
		var balance model.UserBalance
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).First(&balance).Error
		if err != nil {
			return errors.New("余额记录不存在")
		}

		if balance.Balance < amount {
			return errors.New("可用余额不足")
		}

		beforeBalance := balance.Balance

		// 使用原子更新
		result := tx.Model(&model.UserBalance{}).
			Where("user_id = ? AND balance >= ?", userID, amount).
			Updates(map[string]interface{}{
				"balance": gorm.Expr("balance - ?", amount),
				"frozen":  gorm.Expr("frozen + ?", amount),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("可用余额不足或更新失败")
		}

		log := &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeFreeze,
			Amount:        -amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  beforeBalance - amount,
			OrderNo:       orderNo,
			Remark:        remark,
		}
		// 填充操作者信息
		if operator != nil {
			log.OperatorID = operator.OperatorID
			log.OperatorType = operator.OperatorType
			log.ClientIP = operator.ClientIP
		} else {
			log.OperatorType = "user"
		}
		return tx.Create(log).Error
	})
}

// Unfreeze 解冻余额
// 使用原子更新防止并发竞态条件
func (s *BalanceService) Unfreeze(userID uint, amount float64, orderNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("解冻金额必须大于0")
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// 使用 FOR UPDATE 锁定行
		var balance model.UserBalance
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).First(&balance).Error
		if err != nil {
			return errors.New("余额记录不存在")
		}

		if balance.Frozen < amount {
			return errors.New("冻结余额不足")
		}

		beforeBalance := balance.Balance

		// 使用原子更新
		result := tx.Model(&model.UserBalance{}).
			Where("user_id = ? AND frozen >= ?", userID, amount).
			Updates(map[string]interface{}{
				"balance": gorm.Expr("balance + ?", amount),
				"frozen":  gorm.Expr("frozen - ?", amount),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("冻结余额不足或更新失败")
		}

		log := &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeUnfreeze,
			Amount:        amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  beforeBalance + amount,
			OrderNo:       orderNo,
			Remark:        remark,
		}
		// 填充操作者信息
		if operator != nil {
			log.OperatorID = operator.OperatorID
			log.OperatorType = operator.OperatorType
			log.ClientIP = operator.ClientIP
		} else {
			log.OperatorType = "user"
		}
		return tx.Create(log).Error
	})
}

// DeductFrozen 扣除冻结金额（用于确认消费）
// 使用原子更新防止并发竞态条件
func (s *BalanceService) DeductFrozen(userID uint, amount float64, orderNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("扣除金额必须大于0")
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// 使用 FOR UPDATE 锁定行
		var balance model.UserBalance
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).First(&balance).Error
		if err != nil {
			return errors.New("余额记录不存在")
		}

		if balance.Frozen < amount {
			return errors.New("冻结余额不足")
		}

		currentBalance := balance.Balance

		// 使用原子更新
		result := tx.Model(&model.UserBalance{}).
			Where("user_id = ? AND frozen >= ?", userID, amount).
			Updates(map[string]interface{}{
				"frozen":    gorm.Expr("frozen - ?", amount),
				"total_out": gorm.Expr("total_out + ?", amount),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("冻结余额不足或更新失败")
		}

		log := &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeConsume,
			Amount:        -amount,
			BeforeBalance: currentBalance,
			AfterBalance:  currentBalance,
			OrderNo:       orderNo,
			Remark:        remark + "（从冻结扣除）",
		}
		// 填充操作者信息
		if operator != nil {
			log.OperatorID = operator.OperatorID
			log.OperatorType = operator.OperatorType
			log.ClientIP = operator.ClientIP
		} else {
			log.OperatorType = "user"
		}
		return tx.Create(log).Error
	})
}

// AdjustBalance 调整余额（管理员操作）
// 使用 FOR UPDATE 锁定行防止并发问题
func (s *BalanceService) AdjustBalance(userID uint, amount float64, remark string, operator *OperatorInfo) error {
	// 获取余额限制配置
	_, _, _, maxBalance := s.getBalanceLimits()

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// 使用 FOR UPDATE 锁定行
		var balance model.UserBalance
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).First(&balance).Error
		if err == gorm.ErrRecordNotFound {
			balance = model.UserBalance{
				UserID:  userID,
				Balance: 0,
				Frozen:  0,
			}
			if err := tx.Create(&balance).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		beforeBalance := balance.Balance
		newBalance := balance.Balance + amount
		if newBalance < 0 {
			return errors.New("调整后余额不能为负数")
		}

		// 如果是增加余额，检查余额上限
		if amount > 0 && newBalance > maxBalance {
			return fmt.Errorf("调整后余额将超过上限 %.2f 元", maxBalance)
		}

		// 使用原子更新
		updates := map[string]interface{}{
			"balance": gorm.Expr("balance + ?", amount),
		}
		if amount > 0 {
			updates["total_in"] = gorm.Expr("total_in + ?", amount)
		} else {
			updates["total_out"] = gorm.Expr("total_out - ?", amount)
		}

		if err := tx.Model(&model.UserBalance{}).
			Where("user_id = ?", userID).
			Updates(updates).Error; err != nil {
			return err
		}

		log := &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeAdjust,
			Amount:        amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  newBalance,
			Remark:        remark,
		}
		// 填充操作者信息（管理员操作）
		if operator != nil {
			log.OperatorID = operator.OperatorID
			log.OperatorType = operator.OperatorType
			log.ClientIP = operator.ClientIP
		} else {
			log.OperatorType = "admin"
		}
		return tx.Create(log).Error
	})
}

// GiftBalance 赠送余额
// 使用 FOR UPDATE 锁定行防止并发问题，并检查余额上限
func (s *BalanceService) GiftBalance(userID uint, amount float64, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("赠送金额必须大于0")
	}

	// 获取余额限制配置
	_, _, _, maxBalance := s.getBalanceLimits()

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// 使用 FOR UPDATE 锁定行
		var balance model.UserBalance
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).First(&balance).Error
		if err == gorm.ErrRecordNotFound {
			balance = model.UserBalance{
				UserID:  userID,
				Balance: 0,
				Frozen:  0,
			}
			if err := tx.Create(&balance).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// 检查余额上限
		if balance.Balance+amount > maxBalance {
			return fmt.Errorf("赠送后余额将超过上限 %.2f 元", maxBalance)
		}

		beforeBalance := balance.Balance

		// 使用原子更新
		if err := tx.Model(&model.UserBalance{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"balance": gorm.Expr("balance + ?", amount),
			}).Error; err != nil {
			return err
		}

		log := &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeGift,
			Amount:        amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  beforeBalance + amount,
			Remark:        remark,
		}
		// 填充操作者信息（管理员操作）
		if operator != nil {
			log.OperatorID = operator.OperatorID
			log.OperatorType = operator.OperatorType
			log.ClientIP = operator.ClientIP
		} else {
			log.OperatorType = "admin"
		}
		return tx.Create(log).Error
	})
}

// ==================== 充值订单管理 ====================

// GenerateRechargeNo 生成充值单号
// 使用时间戳+随机数，提高不可预测性
func (s *BalanceService) GenerateRechargeNo() string {
	return fmt.Sprintf("RC%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)
}

// CreateRechargeOrder 创建充值订单
// 安全限制：单笔金额限制、余额上限检查、每日充值上限检查
// 自动计算并应用最优充值优惠
func (s *BalanceService) CreateRechargeOrder(userID uint, amount float64, paymentMethod string) (*model.RechargeOrder, error) {
	// 获取余额限制配置
	minRecharge, maxRecharge, maxDaily, maxBalance := s.getBalanceLimits()

	// 验证充值金额范围
	if amount < minRecharge {
		return nil, fmt.Errorf("充值金额不能低于 %.2f 元", minRecharge)
	}
	if amount > maxRecharge {
		return nil, fmt.Errorf("单笔充值金额不能超过 %.2f 元", maxRecharge)
	}

	db := s.repo.GetDB()

	// 检查用户当前余额，确保充值后不超过余额上限
	balance, err := s.GetUserBalance(userID)
	if err != nil {
		return nil, errors.New("获取用户余额失败")
	}
	if balance.Balance+amount > maxBalance {
		return nil, fmt.Errorf("充值后余额将超过上限 %.2f 元，当前余额 %.2f 元", maxBalance, balance.Balance)
	}

	// 检查今日充值总额
	var todayRechargeTotal float64
	todayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	db.Model(&model.RechargeOrder{}).
		Where("user_id = ? AND status = ? AND paid_at >= ?", userID, model.RechargeStatusPaid, todayStart).
		Select("COALESCE(SUM(amount), 0)").Scan(&todayRechargeTotal)
	
	if todayRechargeTotal+amount > maxDaily {
		return nil, fmt.Errorf("今日充值总额将超过上限 %.2f 元，今日已充值 %.2f 元", maxDaily, todayRechargeTotal)
	}

	// 计算充值优惠
	var promoID uint = 0
	var promoName string = ""
	var payAmount float64 = amount
	var bonusAmount float64 = 0
	var totalCredit float64 = amount

	if s.promoSvc != nil {
		promoResult, err := s.promoSvc.CalculatePromo(userID, amount)
		if err == nil && promoResult != nil && promoResult.PromoID > 0 {
			promoID = promoResult.PromoID
			promoName = promoResult.PromoName
			payAmount = promoResult.PayAmount
			bonusAmount = promoResult.BonusAmount
			totalCredit = promoResult.TotalCredit
		}
	}

	order := &model.RechargeOrder{
		RechargeNo:    s.GenerateRechargeNo(),
		UserID:        userID,
		Amount:        amount,
		PayAmount:     payAmount,
		BonusAmount:   bonusAmount,
		TotalCredit:   totalCredit,
		PromoID:       promoID,
		PromoName:     promoName,
		PaymentMethod: paymentMethod,
		Status:        model.RechargeStatusPending,
		ExpireAt:      time.Now().Add(30 * time.Minute), // 30分钟过期
	}

	err = db.Create(order).Error
	return order, err
}

// GetRechargeOrder 获取充值订单
func (s *BalanceService) GetRechargeOrder(rechargeNo string) (*model.RechargeOrder, error) {
	var order model.RechargeOrder
	err := s.repo.GetDB().Where("recharge_no = ?", rechargeNo).First(&order).Error
	return &order, err
}

// GetUserRechargeOrders 获取用户充值订单列表
func (s *BalanceService) GetUserRechargeOrders(userID uint, page, pageSize int) ([]model.RechargeOrder, int64, error) {
	var orders []model.RechargeOrder
	var total int64

	db := s.repo.GetDB()
	db.Model(&model.RechargeOrder{}).Where("user_id = ?", userID).Count(&total)

	offset := (page - 1) * pageSize
	err := db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// CompleteRechargeOrder 完成充值订单
// 使用原子更新防止重复完成
// 支持充值优惠：到账金额 = 充值金额 + 赠送金额
func (s *BalanceService) CompleteRechargeOrder(rechargeNo, paymentNo string) error {
	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// 使用原子更新确保只有一个请求能成功更新订单状态
		now := time.Now()
		result := tx.Model(&model.RechargeOrder{}).
			Where("recharge_no = ? AND status = ?", rechargeNo, model.RechargeStatusPending).
			Updates(map[string]interface{}{
				"status":     model.RechargeStatusPaid,
				"payment_no": paymentNo,
				"paid_at":    now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("订单已处理或不存在")
		}

		// 获取订单信息用于后续操作
		var order model.RechargeOrder
		if err := tx.Where("recharge_no = ?", rechargeNo).First(&order).Error; err != nil {
			return errors.New("获取订单信息失败")
		}

		// 计算实际到账金额（充值金额 + 赠送金额）
		creditAmount := order.Amount + order.BonusAmount
		if order.TotalCredit > 0 {
			creditAmount = order.TotalCredit
		}

		// 使用 FOR UPDATE 锁定余额记录
		var balance model.UserBalance
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", order.UserID).First(&balance).Error
		if err == gorm.ErrRecordNotFound {
			// 创建新的余额记录
			balance = model.UserBalance{
				UserID:  order.UserID,
				Balance: creditAmount,
				Frozen:  0,
				TotalIn: creditAmount,
			}
			if err := tx.Create(&balance).Error; err != nil {
				return err
			}
			// 记录变动日志（系统自动完成充值）
			remark := "在线充值"
			if order.BonusAmount > 0 {
				remark = fmt.Sprintf("在线充值（含赠金 %.2f 元）", order.BonusAmount)
			}
			log := &model.BalanceLog{
				UserID:        order.UserID,
				Type:          model.BalanceTypeRecharge,
				Amount:        creditAmount,
				BeforeBalance: 0,
				AfterBalance:  creditAmount,
				RechargeNo:    rechargeNo,
				Remark:        remark,
				OperatorType:  "system",
			}
			if err := tx.Create(log).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			beforeBalance := balance.Balance

			// 使用原子更新增加余额
			if err := tx.Model(&model.UserBalance{}).
				Where("user_id = ?", order.UserID).
				Updates(map[string]interface{}{
					"balance":  gorm.Expr("balance + ?", creditAmount),
					"total_in": gorm.Expr("total_in + ?", creditAmount),
				}).Error; err != nil {
				return err
			}

			// 记录变动日志（系统自动完成充值）
			remark := "在线充值"
			if order.BonusAmount > 0 {
				remark = fmt.Sprintf("在线充值（含赠金 %.2f 元）", order.BonusAmount)
			}
			log := &model.BalanceLog{
				UserID:        order.UserID,
				Type:          model.BalanceTypeRecharge,
				Amount:        creditAmount,
				BeforeBalance: beforeBalance,
				AfterBalance:  beforeBalance + creditAmount,
				RechargeNo:    rechargeNo,
				Remark:        remark,
				OperatorType:  "system",
			}
			if err := tx.Create(log).Error; err != nil {
				return err
			}
		}

		// 记录优惠使用（如果有使用优惠）
		if order.PromoID > 0 && s.promoSvc != nil {
			discountAmount := order.Amount - order.PayAmount
			if discountAmount < 0 {
				discountAmount = 0
			}
			_ = s.promoSvc.RecordPromoUsage(order.PromoID, order.UserID, rechargeNo, order.Amount, order.BonusAmount, discountAmount)
		}

		return nil
	})
}

// CancelRechargeOrder 取消充值订单
func (s *BalanceService) CancelRechargeOrder(rechargeNo string) error {
	return s.repo.GetDB().Model(&model.RechargeOrder{}).
		Where("recharge_no = ? AND status = ?", rechargeNo, model.RechargeStatusPending).
		Update("status", model.RechargeStatusCancelled).Error
}

// CancelExpiredRechargeOrders 取消过期的充值订单
func (s *BalanceService) CancelExpiredRechargeOrders() error {
	return s.repo.GetDB().Model(&model.RechargeOrder{}).
		Where("status = ? AND expire_at < ?", model.RechargeStatusPending, time.Now()).
		Update("status", model.RechargeStatusCancelled).Error
}

// ==================== 管理员功能 ====================

// AdminGetAllBalances 管理员获取所有用户余额
func (s *BalanceService) AdminGetAllBalances(page, pageSize int, keyword string) ([]map[string]interface{}, int64, error) {
	db := s.repo.GetDB()
	var total int64

	query := db.Table("user_balances").
		Select("user_balances.*, users.username, users.email").
		Joins("LEFT JOIN users ON users.id = user_balances.user_id")

	if keyword != "" {
		query = query.Where("users.username LIKE ? OR users.email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	var results []map[string]interface{}
	offset := (page - 1) * pageSize
	err := query.Order("user_balances.balance DESC").Offset(offset).Limit(pageSize).Find(&results).Error
	return results, total, err
}

// AdminGetBalanceLogs 管理员获取余额变动记录
func (s *BalanceService) AdminGetBalanceLogs(page, pageSize int, userID uint, logType string) ([]model.BalanceLog, int64, error) {
	db := s.repo.GetDB()
	var logs []model.BalanceLog
	var total int64

	query := db.Model(&model.BalanceLog{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if logType != "" {
		query = query.Where("type = ?", logType)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// AdminGetRechargeOrders 管理员获取充值订单
func (s *BalanceService) AdminGetRechargeOrders(page, pageSize int, status int, keyword string) ([]model.RechargeOrder, int64, error) {
	db := s.repo.GetDB()
	var orders []model.RechargeOrder
	var total int64

	query := db.Model(&model.RechargeOrder{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("recharge_no LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// GetBalanceStats 获取余额统计
func (s *BalanceService) GetBalanceStats() (map[string]interface{}, error) {
	db := s.repo.GetDB()
	var stats struct {
		TotalBalance float64
		TotalFrozen  float64
		TotalIn      float64
		TotalOut     float64
		UserCount    int64
	}

	db.Model(&model.UserBalance{}).Select("SUM(balance) as total_balance, SUM(frozen) as total_frozen, SUM(total_in) as total_in, SUM(total_out) as total_out, COUNT(*) as user_count").Scan(&stats)

	// 今日充值统计
	var todayRecharge float64
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	db.Model(&model.RechargeOrder{}).Where("status = ? AND paid_at >= ?", model.RechargeStatusPaid, todayStart).Select("COALESCE(SUM(amount), 0)").Scan(&todayRecharge)

	return map[string]interface{}{
		"total_balance":  stats.TotalBalance,
		"total_frozen":   stats.TotalFrozen,
		"total_in":       stats.TotalIn,
		"total_out":      stats.TotalOut,
		"user_count":     stats.UserCount,
		"today_recharge": todayRecharge,
	}, nil
}
