package service

import (
	"fmt"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// BalanceAlertService 余额告警服务
type BalanceAlertService struct {
	repo      *repository.Repository
	configSvc *ConfigService // 配置服务引用
}

// AlertConfig 告警配置（保留用于兼容）
type AlertConfig struct {
	LargeRechargeThreshold  float64 // 大额充值阈值
	LargeConsumeThreshold   float64 // 大额消费阈值
	FrequentRechargeCount   int     // 频繁充值次数阈值（每小时）
	FrequentConsumeCount    int     // 频繁消费次数阈值（每小时）
	LargeAdminAdjustThreshold float64 // 管理员大额调整阈值
}

// DefaultAlertConfig 默认告警配置
var DefaultAlertConfig = AlertConfig{
	LargeRechargeThreshold:    1000,  // 单笔充值超过1000元
	LargeConsumeThreshold:     500,   // 单笔消费超过500元
	FrequentRechargeCount:     5,     // 每小时充值超过5次
	FrequentConsumeCount:      10,    // 每小时消费超过10次
	LargeAdminAdjustThreshold: 1000,  // 管理员单次调整超过1000元
}

// NewBalanceAlertService 创建余额告警服务
func NewBalanceAlertService(repo *repository.Repository) *BalanceAlertService {
	return &BalanceAlertService{repo: repo}
}

// SetConfigService 设置配置服务引用
func (s *BalanceAlertService) SetConfigService(configSvc *ConfigService) {
	s.configSvc = configSvc
}

// getAlertConfig 获取告警配置
func (s *BalanceAlertService) getAlertConfig() AlertConfig {
	if s.configSvc != nil {
		largeRecharge, largeConsume, largeAdminAdjust, freqRecharge, freqConsume := s.configSvc.GetBalanceAlertThresholds()
		return AlertConfig{
			LargeRechargeThreshold:    largeRecharge,
			LargeConsumeThreshold:     largeConsume,
			FrequentRechargeCount:     freqRecharge,
			FrequentConsumeCount:      freqConsume,
			LargeAdminAdjustThreshold: largeAdminAdjust,
		}
	}
	return DefaultAlertConfig
}

// CreateAlert 创建告警记录
func (s *BalanceAlertService) CreateAlert(alert *model.BalanceAlert) error {
	return s.repo.GetDB().Create(alert).Error
}

// CheckLargeRecharge 检查大额充值
func (s *BalanceAlertService) CheckLargeRecharge(userID uint, amount float64, rechargeNo, clientIP string) {
	cfg := s.getAlertConfig()
	if amount >= cfg.LargeRechargeThreshold {
		alert := &model.BalanceAlert{
			UserID:    userID,
			AlertType: model.AlertTypeLargeRecharge,
			Level:     model.AlertLevelWarning,
			Title:     "大额充值告警",
			Content:   fmt.Sprintf("用户ID %d 进行大额充值，金额: %.2f 元，充值单号: %s", userID, amount, rechargeNo),
			Amount:    amount,
			RelatedID: rechargeNo,
			ClientIP:  clientIP,
		}
		s.CreateAlert(alert)
	}
}

// CheckLargeConsume 检查大额消费
func (s *BalanceAlertService) CheckLargeConsume(userID uint, amount float64, orderNo, clientIP string) {
	cfg := s.getAlertConfig()
	if amount >= cfg.LargeConsumeThreshold {
		alert := &model.BalanceAlert{
			UserID:    userID,
			AlertType: model.AlertTypeLargeConsume,
			Level:     model.AlertLevelWarning,
			Title:     "大额消费告警",
			Content:   fmt.Sprintf("用户ID %d 进行大额消费，金额: %.2f 元，订单号: %s", userID, amount, orderNo),
			Amount:    amount,
			RelatedID: orderNo,
			ClientIP:  clientIP,
		}
		s.CreateAlert(alert)
	}
}

// CheckFrequentRecharge 检查频繁充值
func (s *BalanceAlertService) CheckFrequentRecharge(userID uint, clientIP string) {
	cfg := s.getAlertConfig()
	db := s.repo.GetDB()
	var count int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	
	db.Model(&model.BalanceLog{}).
		Where("user_id = ? AND type = ? AND created_at >= ?", userID, model.BalanceTypeRecharge, oneHourAgo).
		Count(&count)
	
	if int(count) >= cfg.FrequentRechargeCount {
		// 检查是否已有未处理的同类告警（避免重复告警）
		var existingAlert model.BalanceAlert
		err := db.Where("user_id = ? AND alert_type = ? AND status = ? AND created_at >= ?",
			userID, model.AlertTypeFrequentRecharge, model.AlertStatusPending, oneHourAgo).
			First(&existingAlert).Error
		if err == gorm.ErrRecordNotFound {
			alert := &model.BalanceAlert{
				UserID:    userID,
				AlertType: model.AlertTypeFrequentRecharge,
				Level:     model.AlertLevelWarning,
				Title:     "频繁充值告警",
				Content:   fmt.Sprintf("用户ID %d 在1小时内充值 %d 次，超过阈值 %d 次", userID, count, cfg.FrequentRechargeCount),
				Amount:    0,
				ClientIP:  clientIP,
			}
			s.CreateAlert(alert)
		}
	}
}

// CheckFrequentConsume 检查频繁消费
func (s *BalanceAlertService) CheckFrequentConsume(userID uint, clientIP string) {
	cfg := s.getAlertConfig()
	db := s.repo.GetDB()
	var count int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	
	db.Model(&model.BalanceLog{}).
		Where("user_id = ? AND type = ? AND created_at >= ?", userID, model.BalanceTypeConsume, oneHourAgo).
		Count(&count)
	
	if int(count) >= cfg.FrequentConsumeCount {
		// 检查是否已有未处理的同类告警
		var existingAlert model.BalanceAlert
		err := db.Where("user_id = ? AND alert_type = ? AND status = ? AND created_at >= ?",
			userID, model.AlertTypeFrequentConsume, model.AlertStatusPending, oneHourAgo).
			First(&existingAlert).Error
		if err == gorm.ErrRecordNotFound {
			alert := &model.BalanceAlert{
				UserID:    userID,
				AlertType: model.AlertTypeFrequentConsume,
				Level:     model.AlertLevelWarning,
				Title:     "频繁消费告警",
				Content:   fmt.Sprintf("用户ID %d 在1小时内消费 %d 次，超过阈值 %d 次", userID, count, cfg.FrequentConsumeCount),
				Amount:    0,
				ClientIP:  clientIP,
			}
			s.CreateAlert(alert)
		}
	}
}

// CheckNegativeBalance 检查余额异常（负数）
func (s *BalanceAlertService) CheckNegativeBalance(userID uint, balance float64, clientIP string) {
	if balance < 0 {
		alert := &model.BalanceAlert{
			UserID:    userID,
			AlertType: model.AlertTypeNegativeBalance,
			Level:     model.AlertLevelCritical,
			Title:     "余额异常告警",
			Content:   fmt.Sprintf("用户ID %d 余额出现负数: %.2f 元，需要立即检查", userID, balance),
			Amount:    balance,
			ClientIP:  clientIP,
		}
		s.CreateAlert(alert)
	}
}

// CheckBalanceMismatch 检查余额不一致
// 比较 balance + frozen 与 total_in - total_out 是否一致
func (s *BalanceAlertService) CheckBalanceMismatch(userID uint) {
	db := s.repo.GetDB()
	var balance model.UserBalance
	if err := db.Where("user_id = ?", userID).First(&balance).Error; err != nil {
		return
	}
	
	// 计算理论余额：累计充值 - 累计消费
	theoreticalBalance := balance.TotalIn - balance.TotalOut
	actualBalance := balance.Balance + balance.Frozen
	
	// 允许0.01的误差（浮点数精度问题）
	diff := theoreticalBalance - actualBalance
	if diff > 0.01 || diff < -0.01 {
		alert := &model.BalanceAlert{
			UserID:    userID,
			AlertType: model.AlertTypeBalanceMismatch,
			Level:     model.AlertLevelCritical,
			Title:     "余额不一致告警",
			Content:   fmt.Sprintf("用户ID %d 余额不一致，理论余额: %.2f，实际余额: %.2f（可用: %.2f + 冻结: %.2f），差额: %.2f",
				userID, theoreticalBalance, actualBalance, balance.Balance, balance.Frozen, diff),
			Amount:    diff,
		}
		s.CreateAlert(alert)
	}
}

// CheckAdminLargeAdjust 检查管理员大额调整
func (s *BalanceAlertService) CheckAdminLargeAdjust(userID uint, amount float64, adminID uint, remark, clientIP string) {
	cfg := s.getAlertConfig()
	absAmount := amount
	if absAmount < 0 {
		absAmount = -absAmount
	}
	
	if absAmount >= cfg.LargeAdminAdjustThreshold {
		level := model.AlertLevelWarning
		if absAmount >= cfg.LargeAdminAdjustThreshold*5 {
			level = model.AlertLevelCritical
		}
		
		alert := &model.BalanceAlert{
			UserID:    userID,
			AlertType: model.AlertTypeAdminAdjust,
			Level:     level,
			Title:     "管理员大额调整告警",
			Content:   fmt.Sprintf("管理员ID %d 对用户ID %d 进行大额余额调整，金额: %.2f 元，备注: %s", adminID, userID, amount, remark),
			Amount:    amount,
			ClientIP:  clientIP,
		}
		s.CreateAlert(alert)
	}
}

// RecordUnfreezeFailure 记录解冻失败
func (s *BalanceAlertService) RecordUnfreezeFailure(userID uint, amount float64, orderNo, errorMsg, clientIP string) {
	alert := &model.BalanceAlert{
		UserID:    userID,
		AlertType: model.AlertTypeUnfreezeFailure,
		Level:     model.AlertLevelCritical,
		Title:     "解冻失败告警",
		Content:   fmt.Sprintf("用户ID %d 解冻失败，金额: %.2f 元，订单号: %s，错误: %s", userID, amount, orderNo, errorMsg),
		Amount:    amount,
		RelatedID: orderNo,
		ClientIP:  clientIP,
	}
	s.CreateAlert(alert)
}

// RecordRefundAnomaly 记录退款异常
func (s *BalanceAlertService) RecordRefundAnomaly(userID uint, amount float64, orderNo, reason, clientIP string) {
	alert := &model.BalanceAlert{
		UserID:    userID,
		AlertType: model.AlertTypeRefundAnomaly,
		Level:     model.AlertLevelWarning,
		Title:     "退款异常告警",
		Content:   fmt.Sprintf("用户ID %d 退款异常，金额: %.2f 元，订单号: %s，原因: %s", userID, amount, orderNo, reason),
		Amount:    amount,
		RelatedID: orderNo,
		ClientIP:  clientIP,
	}
	s.CreateAlert(alert)
}


// ==================== 管理员功能 ====================

// GetAlerts 获取告警列表
func (s *BalanceAlertService) GetAlerts(page, pageSize int, alertType, level string, status int) ([]model.BalanceAlert, int64, error) {
	db := s.repo.GetDB()
	var alerts []model.BalanceAlert
	var total int64

	query := db.Model(&model.BalanceAlert{})
	if alertType != "" {
		query = query.Where("alert_type = ?", alertType)
	}
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&alerts).Error
	return alerts, total, err
}

// GetAlertByID 获取告警详情
func (s *BalanceAlertService) GetAlertByID(id uint) (*model.BalanceAlert, error) {
	var alert model.BalanceAlert
	err := s.repo.GetDB().First(&alert, id).Error
	return &alert, err
}

// HandleAlert 处理告警
func (s *BalanceAlertService) HandleAlert(id uint, adminID uint, status int, remark string) error {
	now := time.Now()
	return s.repo.GetDB().Model(&model.BalanceAlert{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        status,
			"handled_by":    adminID,
			"handled_at":    now,
			"handle_remark": remark,
		}).Error
}

// GetAlertStats 获取告警统计
func (s *BalanceAlertService) GetAlertStats() (map[string]interface{}, error) {
	db := s.repo.GetDB()
	
	// 未处理告警数量
	var pendingCount int64
	db.Model(&model.BalanceAlert{}).Where("status = ?", model.AlertStatusPending).Count(&pendingCount)
	
	// 严重告警数量
	var criticalCount int64
	db.Model(&model.BalanceAlert{}).Where("status = ? AND level = ?", model.AlertStatusPending, model.AlertLevelCritical).Count(&criticalCount)
	
	// 今日告警数量
	todayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	var todayCount int64
	db.Model(&model.BalanceAlert{}).Where("created_at >= ?", todayStart).Count(&todayCount)
	
	// 按类型统计未处理告警
	type TypeCount struct {
		AlertType string
		Count     int64
	}
	var typeCounts []TypeCount
	db.Model(&model.BalanceAlert{}).
		Select("alert_type, COUNT(*) as count").
		Where("status = ?", model.AlertStatusPending).
		Group("alert_type").
		Scan(&typeCounts)
	
	typeStats := make(map[string]int64)
	for _, tc := range typeCounts {
		typeStats[tc.AlertType] = tc.Count
	}
	
	return map[string]interface{}{
		"pending_count":  pendingCount,
		"critical_count": criticalCount,
		"today_count":    todayCount,
		"type_stats":     typeStats,
	}, nil
}

// BatchCheckBalanceMismatch 批量检查余额不一致
// 用于定时任务
func (s *BalanceAlertService) BatchCheckBalanceMismatch() (int, error) {
	db := s.repo.GetDB()
	var balances []model.UserBalance
	
	// 获取所有有余额的用户
	if err := db.Where("balance > 0 OR frozen > 0 OR total_in > 0").Find(&balances).Error; err != nil {
		return 0, err
	}
	
	alertCount := 0
	for _, balance := range balances {
		theoreticalBalance := balance.TotalIn - balance.TotalOut
		actualBalance := balance.Balance + balance.Frozen
		
		diff := theoreticalBalance - actualBalance
		if diff > 0.01 || diff < -0.01 {
			// 检查是否已有未处理的同类告警
			var existingAlert model.BalanceAlert
			err := db.Where("user_id = ? AND alert_type = ? AND status = ?",
				balance.UserID, model.AlertTypeBalanceMismatch, model.AlertStatusPending).
				First(&existingAlert).Error
			if err == gorm.ErrRecordNotFound {
				alert := &model.BalanceAlert{
					UserID:    balance.UserID,
					AlertType: model.AlertTypeBalanceMismatch,
					Level:     model.AlertLevelCritical,
					Title:     "余额不一致告警",
					Content:   fmt.Sprintf("用户ID %d 余额不一致，理论余额: %.2f，实际余额: %.2f（可用: %.2f + 冻结: %.2f），差额: %.2f",
						balance.UserID, theoreticalBalance, actualBalance, balance.Balance, balance.Frozen, diff),
					Amount:    diff,
				}
				if err := s.CreateAlert(alert); err == nil {
					alertCount++
				}
			}
		}
	}
	
	return alertCount, nil
}

// CleanOldAlerts 清理旧告警（已处理超过30天的）
func (s *BalanceAlertService) CleanOldAlerts(days int) (int64, error) {
	threshold := time.Now().AddDate(0, 0, -days)
	result := s.repo.GetDB().
		Where("status != ? AND handled_at < ?", model.AlertStatusPending, threshold).
		Delete(&model.BalanceAlert{})
	return result.RowsAffected, result.Error
}
