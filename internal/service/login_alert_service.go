package service

import (
	"strings"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// LoginAlertService 异地登录提醒服务
type LoginAlertService struct {
	repo     *repository.Repository
	emailSvc *EmailService
}

// NewLoginAlertService 创建异地登录提醒服务实例
func NewLoginAlertService(repo *repository.Repository, emailSvc *EmailService) *LoginAlertService {
	return &LoginAlertService{
		repo:     repo,
		emailSvc: emailSvc,
	}
}

// CheckAndAlertLogin 检查登录并发送异地登录提醒
// 参数：
//   - userID: 用户ID
//   - username: 用户名
//   - email: 用户邮箱
//   - ip: 登录IP
//   - location: IP归属地
//   - deviceInfo: 设备信息
// 返回：
//   - 是否为异地登录
//   - 错误信息
func (s *LoginAlertService) CheckAndAlertLogin(userID uint, username, email, ip, location, deviceInfo string) (bool, error) {
	// 获取用户常用登录地点
	var locations []model.UserLoginLocation
	s.repo.GetDB().Where("user_id = ?", userID).Order("login_count DESC").Limit(10).Find(&locations)

	// 检查是否为新地点
	isNewLocation := true
	var previousLocation *model.UserLoginLocation

	for i := range locations {
		// 检查IP前缀是否相同（同一网段）或归属地是否相同
		if isSameNetwork(locations[i].IP, ip) || isSameLocation(locations[i].Location, location) {
			isNewLocation = false
			previousLocation = &locations[i]
			break
		}
	}

	// 更新或创建登录地点记录
	if isNewLocation {
		// 创建新的登录地点记录
		newLocation := &model.UserLoginLocation{
			UserID:      userID,
			IP:          ip,
			Location:    location,
			LoginCount:  1,
			LastLoginAt: time.Now(),
			IsTrusted:   false,
		}
		s.repo.GetDB().Create(newLocation)

		// 如果有历史登录记录，发送异地登录提醒
		if len(locations) > 0 && email != "" {
			prevIP := locations[0].IP
			prevLoc := locations[0].Location

			// 创建提醒记录
			alert := &model.LoginAlert{
				UserID:       userID,
				Username:     username,
				IP:           ip,
				Location:     location,
				PreviousIP:   prevIP,
				PrevLocation: prevLoc,
				DeviceInfo:   deviceInfo,
				AlertType:    model.AlertTypeNewLocation,
			}
			s.repo.GetDB().Create(alert)

			// 发送邮件提醒
			if s.emailSvc != nil {
				go func() {
					if err := s.emailSvc.SendLoginAlertEmail(email, username, ip, location, deviceInfo, time.Now()); err == nil {
						now := time.Now()
						s.repo.GetDB().Model(alert).Updates(map[string]interface{}{
							"email_sent":    true,
							"email_sent_at": &now,
						})
					}
				}()
			}
		}
	} else if previousLocation != nil {
		// 更新已有地点的登录次数
		s.repo.GetDB().Model(previousLocation).Updates(map[string]interface{}{
			"login_count":   previousLocation.LoginCount + 1,
			"last_login_at": time.Now(),
		})
	}

	return isNewLocation, nil
}

// GetUserAlerts 获取用户的登录提醒记录
// 参数：
//   - userID: 用户ID
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 提醒列表
//   - 总数
//   - 错误信息
func (s *LoginAlertService) GetUserAlerts(userID uint, page, pageSize int) ([]model.LoginAlert, int64, error) {
	var alerts []model.LoginAlert
	var total int64

	query := s.repo.GetDB().Model(&model.LoginAlert{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&alerts).Error; err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

// AcknowledgeAlert 确认登录提醒
// 参数：
//   - userID: 用户ID
//   - alertID: 提醒ID
// 返回：
//   - 错误信息
func (s *LoginAlertService) AcknowledgeAlert(userID uint, alertID uint) error {
	return s.repo.GetDB().Model(&model.LoginAlert{}).
		Where("id = ? AND user_id = ?", alertID, userID).
		Update("acknowledged", true).Error
}

// TrustLocation 将地点标记为可信
// 参数：
//   - userID: 用户ID
//   - locationID: 地点ID
// 返回：
//   - 错误信息
func (s *LoginAlertService) TrustLocation(userID uint, locationID uint) error {
	return s.repo.GetDB().Model(&model.UserLoginLocation{}).
		Where("id = ? AND user_id = ?", locationID, userID).
		Update("is_trusted", true).Error
}

// GetUserLocations 获取用户的登录地点列表
// 参数：
//   - userID: 用户ID
// 返回：
//   - 地点列表
//   - 错误信息
func (s *LoginAlertService) GetUserLocations(userID uint) ([]model.UserLoginLocation, error) {
	var locations []model.UserLoginLocation
	if err := s.repo.GetDB().Where("user_id = ?", userID).
		Order("login_count DESC").Find(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

// RemoveLocation 移除登录地点
// 参数：
//   - userID: 用户ID
//   - locationID: 地点ID
// 返回：
//   - 错误信息
func (s *LoginAlertService) RemoveLocation(userID uint, locationID uint) error {
	return s.repo.GetDB().Where("id = ? AND user_id = ?", locationID, userID).
		Delete(&model.UserLoginLocation{}).Error
}

// GetUnacknowledgedCount 获取未确认的提醒数量
// 参数：
//   - userID: 用户ID
// 返回：
//   - 未确认数量
func (s *LoginAlertService) GetUnacknowledgedCount(userID uint) int64 {
	var count int64
	s.repo.GetDB().Model(&model.LoginAlert{}).
		Where("user_id = ? AND acknowledged = ?", userID, false).
		Count(&count)
	return count
}

// isSameNetwork 检查两个IP是否在同一网段
func isSameNetwork(ip1, ip2 string) bool {
	// 简单实现：比较IP的前三段
	parts1 := strings.Split(ip1, ".")
	parts2 := strings.Split(ip2, ".")

	if len(parts1) < 3 || len(parts2) < 3 {
		return ip1 == ip2
	}

	return parts1[0] == parts2[0] && parts1[1] == parts2[1] && parts1[2] == parts2[2]
}

// isSameLocation 检查两个归属地是否相同
func isSameLocation(loc1, loc2 string) bool {
	// 简单实现：检查城市级别是否相同
	// 归属地格式通常为：国家 省份 城市 运营商
	if loc1 == "" || loc2 == "" {
		return false
	}

	// 提取城市信息进行比较
	city1 := extractCity(loc1)
	city2 := extractCity(loc2)

	return city1 != "" && city1 == city2
}

// extractCity 从归属地字符串中提取城市
func extractCity(location string) string {
	// 简单实现：按空格分割，取第三个部分（城市）
	parts := strings.Fields(location)
	if len(parts) >= 3 {
		return parts[2]
	}
	if len(parts) >= 2 {
		return parts[1]
	}
	return location
}
