package service

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// AutoReplyService 智能客服服务
type AutoReplyService struct {
	repo *repository.Repository
}

// NewAutoReplyService 创建智能客服服务实例
func NewAutoReplyService(repo *repository.Repository) *AutoReplyService {
	return &AutoReplyService{repo: repo}
}

// GetConfig 获取智能客服配置
// 返回：
//   - 配置信息
//   - 错误信息（如有）
func (s *AutoReplyService) GetConfig() (*model.AutoReplyConfig, error) {
	var config model.AutoReplyConfig
	result := s.repo.GetDB().First(&config)
	if result.Error != nil {
		// 返回默认配置
		return &model.AutoReplyConfig{
			Enabled:          false,
			WelcomeMessage:   "您好！我是智能客服助手，请问有什么可以帮助您的？",
			NoMatchReply:     "抱歉，我暂时无法理解您的问题。您可以输入\"转人工\"联系人工客服。",
			TransferKeywords: `["转人工","人工客服","人工服务"]`,
			TransferMessage:  "正在为您转接人工客服，请稍候...",
		}, nil
	}
	return &config, nil
}

// SaveConfig 保存智能客服配置
// 参数：
//   - config: 配置信息
// 返回：
//   - 错误信息（如有）
func (s *AutoReplyService) SaveConfig(config *model.AutoReplyConfig) error {
	var existing model.AutoReplyConfig
	result := s.repo.GetDB().First(&existing)
	if result.Error != nil {
		return s.repo.GetDB().Create(config).Error
	}
	config.ID = existing.ID
	return s.repo.GetDB().Save(config).Error
}

// AutoReplyResult 自动回复结果
type AutoReplyResult struct {
	Reply       string `json:"reply"`        // 回复内容
	RuleID      uint   `json:"rule_id"`      // 匹配的规则ID
	RuleName    string `json:"rule_name"`    // 规则名称
	Transferred bool   `json:"transferred"`  // 是否转人工
	Matched     bool   `json:"matched"`      // 是否匹配到规则
}

// ProcessMessage 处理用户消息
// 参数：
//   - sessionID: 会话ID
//   - userID: 用户ID
//   - message: 用户消息
// 返回：
//   - 回复结果
//   - 错误信息（如有）
func (s *AutoReplyService) ProcessMessage(sessionID string, userID uint, message string) (*AutoReplyResult, error) {
	config, _ := s.GetConfig()
	if !config.Enabled {
		return nil, nil // 智能客服未启用
	}

	// 检查工作时间
	if config.WorkingHoursOnly && !s.isWorkingHours(config) {
		return nil, nil // 非工作时间不启用
	}

	result := &AutoReplyResult{}

	// 检查是否是转人工关键词
	var transferKeywords []string
	json.Unmarshal([]byte(config.TransferKeywords), &transferKeywords)
	for _, keyword := range transferKeywords {
		if strings.Contains(message, keyword) {
			result.Reply = config.TransferMessage
			result.Transferred = true
			s.logReply(sessionID, userID, message, 0, "转人工", result.Reply, true)
			return result, nil
		}
	}

	// 查找匹配的规则
	var rules []model.AutoReplyRule
	s.repo.GetDB().Where("status = 1").Order("priority DESC").Find(&rules)

	for _, rule := range rules {
		if s.matchRule(&rule, message) {
			result.Reply = rule.Reply
			result.RuleID = rule.ID
			result.RuleName = rule.Name
			result.Matched = true

			// 更新命中次数
			s.repo.GetDB().Model(&rule).Update("hit_count", rule.HitCount+1)

			// 记录日志
			s.logReply(sessionID, userID, message, rule.ID, rule.Name, result.Reply, false)
			return result, nil
		}
	}

	// 无匹配，返回默认回复
	result.Reply = config.NoMatchReply
	result.Matched = false
	s.logReply(sessionID, userID, message, 0, "", result.Reply, false)
	return result, nil
}

// matchRule 检查消息是否匹配规则
func (s *AutoReplyService) matchRule(rule *model.AutoReplyRule, message string) bool {
	var keywords []string
	if err := json.Unmarshal([]byte(rule.Keywords), &keywords); err != nil {
		return false
	}

	message = strings.ToLower(message)

	for _, keyword := range keywords {
		keyword = strings.ToLower(keyword)
		switch rule.MatchType {
		case "exact":
			if message == keyword {
				return true
			}
		case "regex":
			if matched, _ := regexp.MatchString(keyword, message); matched {
				return true
			}
		default: // contains
			if strings.Contains(message, keyword) {
				return true
			}
		}
	}
	return false
}

// isWorkingHours 检查是否在工作时间内
func (s *AutoReplyService) isWorkingHours(config *model.AutoReplyConfig) bool {
	if config.WorkingHoursStart == "" || config.WorkingHoursEnd == "" {
		return true
	}

	now := time.Now()
	currentTime := now.Format("15:04")

	return currentTime >= config.WorkingHoursStart && currentTime <= config.WorkingHoursEnd
}

// logReply 记录自动回复日志
func (s *AutoReplyService) logReply(sessionID string, userID uint, userMessage string, ruleID uint, ruleName, botReply string, transferred bool) {
	log := &model.AutoReplyLog{
		SessionID:   sessionID,
		UserID:      userID,
		UserMessage: userMessage,
		RuleID:      ruleID,
		RuleName:    ruleName,
		BotReply:    botReply,
		Transferred: transferred,
	}
	s.repo.GetDB().Create(log)
}

// GetWelcomeMessage 获取欢迎语
// 返回：
//   - 欢迎语
func (s *AutoReplyService) GetWelcomeMessage() string {
	config, _ := s.GetConfig()
	if !config.Enabled {
		return ""
	}
	return config.WelcomeMessage
}

// ========== 规则管理 ==========

// GetRules 获取规则列表
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 规则列表
//   - 总数
//   - 错误信息（如有）
func (s *AutoReplyService) GetRules(page, pageSize int) ([]model.AutoReplyRule, int64, error) {
	var total int64
	s.repo.GetDB().Model(&model.AutoReplyRule{}).Count(&total)

	var rules []model.AutoReplyRule
	offset := (page - 1) * pageSize
	err := s.repo.GetDB().Order("priority DESC, created_at DESC").Offset(offset).Limit(pageSize).Find(&rules).Error

	return rules, total, err
}

// GetRule 获取单个规则
// 参数：
//   - ruleID: 规则ID
// 返回：
//   - 规则信息
//   - 错误信息（如有）
func (s *AutoReplyService) GetRule(ruleID uint) (*model.AutoReplyRule, error) {
	var rule model.AutoReplyRule
	err := s.repo.GetDB().First(&rule, ruleID).Error
	return &rule, err
}

// CreateRule 创建规则
// 参数：
//   - rule: 规则信息
// 返回：
//   - 错误信息（如有）
func (s *AutoReplyService) CreateRule(rule *model.AutoReplyRule) error {
	return s.repo.GetDB().Create(rule).Error
}

// UpdateRule 更新规则
// 参数：
//   - rule: 规则信息
// 返回：
//   - 错误信息（如有）
func (s *AutoReplyService) UpdateRule(rule *model.AutoReplyRule) error {
	return s.repo.GetDB().Save(rule).Error
}

// DeleteRule 删除规则
// 参数：
//   - ruleID: 规则ID
// 返回：
//   - 错误信息（如有）
func (s *AutoReplyService) DeleteRule(ruleID uint) error {
	return s.repo.GetDB().Delete(&model.AutoReplyRule{}, ruleID).Error
}

// ========== 日志查询 ==========

// GetLogs 获取自动回复日志
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 日志列表
//   - 总数
//   - 错误信息（如有）
func (s *AutoReplyService) GetLogs(page, pageSize int) ([]model.AutoReplyLog, int64, error) {
	var total int64
	s.repo.GetDB().Model(&model.AutoReplyLog{}).Count(&total)

	var logs []model.AutoReplyLog
	offset := (page - 1) * pageSize
	err := s.repo.GetDB().Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

// GetStats 获取智能客服统计
// 返回：
//   - 统计信息
func (s *AutoReplyService) GetStats() map[string]interface{} {
	var totalMessages, matchedMessages, transferredMessages int64
	var topRules []struct {
		RuleID   uint   `json:"rule_id"`
		RuleName string `json:"rule_name"`
		HitCount int    `json:"hit_count"`
	}

	s.repo.GetDB().Model(&model.AutoReplyLog{}).Count(&totalMessages)
	s.repo.GetDB().Model(&model.AutoReplyLog{}).Where("rule_id > 0").Count(&matchedMessages)
	s.repo.GetDB().Model(&model.AutoReplyLog{}).Where("transferred = true").Count(&transferredMessages)

	// 获取命中最多的规则
	s.repo.GetDB().Model(&model.AutoReplyRule{}).
		Select("id as rule_id, name as rule_name, hit_count").
		Order("hit_count DESC").
		Limit(5).
		Scan(&topRules)

	matchRate := float64(0)
	if totalMessages > 0 {
		matchRate = float64(matchedMessages) / float64(totalMessages) * 100
	}

	return map[string]interface{}{
		"total_messages":      totalMessages,
		"matched_messages":    matchedMessages,
		"transferred_messages": transferredMessages,
		"match_rate":          matchRate,
		"top_rules":           topRules,
	}
}

// InitDefaultRules 初始化默认规则
func (s *AutoReplyService) InitDefaultRules() error {
	var count int64
	s.repo.GetDB().Model(&model.AutoReplyRule{}).Count(&count)
	if count > 0 {
		return nil // 已有规则，不需要初始化
	}

	defaultRules := []model.AutoReplyRule{
		{
			Name:      "订单查询",
			Keywords:  `["订单","查订单","订单状态","订单号"]`,
			MatchType: "contains",
			Reply:     "您可以在【用户中心】-【我的订单】中查看订单状态。如需查询特定订单，请提供订单号。",
			Priority:  10,
			Status:    1,
		},
		{
			Name:      "支付问题",
			Keywords:  `["支付","付款","付不了","支付失败","怎么付款"]`,
			MatchType: "contains",
			Reply:     "我们支持多种支付方式：PayPal、支付宝、微信支付等。如果支付遇到问题，请检查网络连接或尝试更换支付方式。",
			Priority:  10,
			Status:    1,
		},
		{
			Name:      "退款问题",
			Keywords:  `["退款","退钱","申请退款","怎么退款"]`,
			MatchType: "contains",
			Reply:     "如需申请退款，请联系人工客服并提供订单号。我们会在1-3个工作日内处理您的退款申请。",
			Priority:  10,
			Status:    1,
		},
		{
			Name:      "卡密问题",
			Keywords:  `["卡密","激活码","兑换码","怎么用","如何使用"]`,
			MatchType: "contains",
			Reply:     "购买成功后，卡密会显示在订单详情中。您可以复制卡密到对应软件中进行激活。",
			Priority:  10,
			Status:    1,
		},
		{
			Name:      "联系方式",
			Keywords:  `["联系","客服","电话","邮箱","怎么联系"]`,
			MatchType: "contains",
			Reply:     "您可以通过以下方式联系我们：\n1. 在线客服（工作时间）\n2. 提交工单\n3. 发送邮件至客服邮箱",
			Priority:  5,
			Status:    1,
		},
	}

	for _, rule := range defaultRules {
		s.repo.GetDB().Create(&rule)
	}

	return nil
}
