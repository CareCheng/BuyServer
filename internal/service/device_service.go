package service

import (
	"regexp"
	"strings"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// DeviceService 设备管理服务
type DeviceService struct {
	repo *repository.Repository
}

// NewDeviceService 创建设备管理服务
func NewDeviceService(repo *repository.Repository) *DeviceService {
	return &DeviceService{repo: repo}
}

// RecordLoginDevice 记录登录设备
// 参数：
//   - userID: 用户ID
//   - sessionID: 会话ID
//   - ip: 登录IP
//   - userAgent: User-Agent字符串
// 返回：
//   - 设备记录
//   - 错误信息
func (s *DeviceService) RecordLoginDevice(userID uint, sessionID, ip, userAgent string) (*model.LoginDevice, error) {
	// 解析设备信息
	deviceName, deviceType, browser, os := parseUserAgent(userAgent)

	// 检查是否已存在相同设备
	var existing model.LoginDevice
	err := s.repo.GetDB().Where("user_id = ? AND session_id = ?", userID, sessionID).First(&existing).Error
	if err == nil {
		// 更新现有记录
		existing.LastActive = time.Now()
		existing.IP = ip
		s.repo.GetDB().Save(&existing)
		return &existing, nil
	}

	// 创建新设备记录
	device := &model.LoginDevice{
		UserID:     userID,
		SessionID:  sessionID,
		DeviceName: deviceName,
		DeviceType: deviceType,
		Browser:    browser,
		OS:         os,
		IP:         ip,
		Location:   getIPLocation(ip),
		LastActive: time.Now(),
	}

	if err := s.repo.GetDB().Create(device).Error; err != nil {
		return nil, err
	}

	return device, nil
}

// GetUserDevices 获取用户的登录设备列表
// 参数：
//   - userID: 用户ID
//   - currentSessionID: 当前会话ID（用于标记当前设备）
// 返回：
//   - 设备列表
//   - 错误信息
func (s *DeviceService) GetUserDevices(userID uint, currentSessionID string) ([]model.LoginDevice, error) {
	var devices []model.LoginDevice
	
	// 获取所有有效设备（关联有效会话的设备）
	err := s.repo.GetDB().
		Where("user_id = ?", userID).
		Order("last_active DESC").
		Find(&devices).Error
	
	if err != nil {
		return nil, err
	}

	// 过滤掉已过期会话的设备
	var validDevices []model.LoginDevice
	for _, device := range devices {
		var session model.UserSession
		if err := s.repo.GetDB().Where("session_id = ? AND expires_at > ?", device.SessionID, time.Now()).First(&session).Error; err == nil {
			// 标记当前设备
			if device.SessionID == currentSessionID {
				device.IsCurrent = true
			}
			validDevices = append(validDevices, device)
		}
	}

	return validDevices, nil
}

// RemoveDevice 移除设备（踢出登录）
// 参数：
//   - userID: 用户ID
//   - deviceID: 设备ID
//   - currentSessionID: 当前会话ID（不能踢出自己）
// 返回：
//   - 错误信息
func (s *DeviceService) RemoveDevice(userID uint, deviceID uint, currentSessionID string) error {
	var device model.LoginDevice
	if err := s.repo.GetDB().Where("id = ? AND user_id = ?", deviceID, userID).First(&device).Error; err != nil {
		return err
	}

	// 不能踢出当前设备
	if device.SessionID == currentSessionID {
		return nil
	}

	// 删除关联的会话
	s.repo.GetDB().Where("session_id = ?", device.SessionID).Delete(&model.UserSession{})

	// 删除设备记录
	return s.repo.GetDB().Delete(&device).Error
}

// RemoveAllOtherDevices 移除所有其他设备
// 参数：
//   - userID: 用户ID
//   - currentSessionID: 当前会话ID
// 返回：
//   - 移除的设备数量
//   - 错误信息
func (s *DeviceService) RemoveAllOtherDevices(userID uint, currentSessionID string) (int64, error) {
	// 获取所有其他设备
	var devices []model.LoginDevice
	s.repo.GetDB().Where("user_id = ? AND session_id != ?", userID, currentSessionID).Find(&devices)

	var count int64
	for _, device := range devices {
		// 删除关联的会话
		s.repo.GetDB().Where("session_id = ?", device.SessionID).Delete(&model.UserSession{})
		// 删除设备记录
		s.repo.GetDB().Delete(&device)
		count++
	}

	return count, nil
}

// UpdateDeviceActivity 更新设备活跃时间
// 参数：
//   - sessionID: 会话ID
func (s *DeviceService) UpdateDeviceActivity(sessionID string) {
	s.repo.GetDB().Model(&model.LoginDevice{}).
		Where("session_id = ?", sessionID).
		Update("last_active", time.Now())
}

// RecordLoginHistory 记录登录历史
// 参数：
//   - userID: 用户ID
//   - username: 用户名
//   - ip: 登录IP
//   - userAgent: User-Agent
//   - success: 是否成功
//   - failReason: 失败原因
func (s *DeviceService) RecordLoginHistory(userID uint, username, ip, userAgent string, success bool, failReason string) {
	_, _, browser, os := parseUserAgent(userAgent)
	
	status := 1
	if !success {
		status = 0
	}

	history := &model.LoginHistory{
		UserID:     userID,
		Username:   username,
		IP:         ip,
		Location:   getIPLocation(ip),
		Device:     getDeviceSummary(userAgent),
		Browser:    browser,
		OS:         os,
		Status:     status,
		FailReason: failReason,
	}

	s.repo.GetDB().Create(history)
}

// GetLoginHistory 获取登录历史
// 参数：
//   - userID: 用户ID
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 登录历史列表
//   - 总数
//   - 错误信息
func (s *DeviceService) GetLoginHistory(userID uint, page, pageSize int) ([]model.LoginHistory, int64, error) {
	var histories []model.LoginHistory
	var total int64

	query := s.repo.GetDB().Model(&model.LoginHistory{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&histories).Error

	return histories, total, err
}

// CleanupExpiredDevices 清理过期设备记录
func (s *DeviceService) CleanupExpiredDevices() {
	// 删除关联会话已过期的设备记录
	s.repo.GetDB().Exec(`
		DELETE FROM login_devices 
		WHERE session_id NOT IN (
			SELECT session_id FROM user_sessions WHERE expires_at > ?
		)
	`, time.Now())
}

// ==================== 辅助函数 ====================

// parseUserAgent 解析User-Agent字符串
// 返回：设备名称、设备类型、浏览器、操作系统
func parseUserAgent(ua string) (deviceName, deviceType, browser, os string) {
	ua = strings.ToLower(ua)

	// 检测设备类型
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") && !strings.Contains(ua, "tablet") {
		deviceType = "Mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		deviceType = "Tablet"
	} else {
		deviceType = "PC"
	}

	// 检测操作系统
	switch {
	case strings.Contains(ua, "windows nt 10"):
		os = "Windows 10/11"
	case strings.Contains(ua, "windows nt 6.3"):
		os = "Windows 8.1"
	case strings.Contains(ua, "windows nt 6.2"):
		os = "Windows 8"
	case strings.Contains(ua, "windows nt 6.1"):
		os = "Windows 7"
	case strings.Contains(ua, "windows"):
		os = "Windows"
	case strings.Contains(ua, "mac os x"):
		os = "macOS"
	case strings.Contains(ua, "iphone"):
		os = "iOS"
	case strings.Contains(ua, "ipad"):
		os = "iPadOS"
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "linux"):
		os = "Linux"
	default:
		os = "Unknown"
	}

	// 检测浏览器
	switch {
	case strings.Contains(ua, "edg/"):
		browser = "Edge"
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg"):
		browser = "Chrome"
	case strings.Contains(ua, "firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		browser = "Safari"
	case strings.Contains(ua, "opera") || strings.Contains(ua, "opr"):
		browser = "Opera"
	case strings.Contains(ua, "msie") || strings.Contains(ua, "trident"):
		browser = "IE"
	default:
		browser = "Unknown"
	}

	// 生成设备名称
	deviceName = browser + " on " + os

	return
}

// getDeviceSummary 获取设备摘要信息
func getDeviceSummary(ua string) string {
	deviceName, deviceType, _, _ := parseUserAgent(ua)
	return deviceType + " - " + deviceName
}

// getIPLocation 获取IP归属地
// 注意：这里使用简单实现，实际项目中应该调用IP归属地API
func getIPLocation(ip string) string {
	// 本地IP
	if ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return "本地网络"
	}

	// 简单的IP段判断（实际应该使用IP库或API）
	// 这里只是示例，实际项目中建议使用 ip2region 或类似的IP库
	ipPattern := regexp.MustCompile(`^(\d+)\.(\d+)\.`)
	matches := ipPattern.FindStringSubmatch(ip)
	if len(matches) >= 3 {
		// 简单返回，实际应该查询IP库
		return "未知位置"
	}

	return "未知位置"
}
