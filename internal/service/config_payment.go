// Package service 提供业务逻辑服务
// config_payment.go - 支付配置管理方法
package service

import (
	"encoding/json"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
)

// GetPaymentConfig 获取所有支付配置
func (s *ConfigService) GetPaymentConfig() (*config.PaymentConfig, error) {
	result := &config.PaymentConfig{}

	// 获取支付宝配置
	if alipay, err := s.repo.GetPaymentConfig("alipay_f2f"); err == nil {
		var cfg config.AlipayF2FConfig
		json.Unmarshal([]byte(alipay.ConfigJSON), &cfg)
		cfg.Enabled = alipay.Enabled
		result.AlipayF2F = cfg
	}

	// 获取微信支付配置
	if wechat, err := s.repo.GetPaymentConfig("wechat_pay"); err == nil {
		var cfg config.WechatPayConfig
		json.Unmarshal([]byte(wechat.ConfigJSON), &cfg)
		cfg.Enabled = wechat.Enabled
		result.WechatPay = cfg
	}

	// 获取易支付配置
	if yipay, err := s.repo.GetPaymentConfig("yi_pay"); err == nil {
		var cfg config.YiPayConfig
		json.Unmarshal([]byte(yipay.ConfigJSON), &cfg)
		cfg.Enabled = yipay.Enabled
		result.YiPay = cfg
	}

	// 获取PayPal配置
	if paypal, err := s.repo.GetPaymentConfig("paypal"); err == nil {
		var cfg config.PayPalConfig
		json.Unmarshal([]byte(paypal.ConfigJSON), &cfg)
		cfg.Enabled = paypal.Enabled
		result.PayPal = cfg
	}

	// 获取Stripe配置
	if stripe, err := s.repo.GetPaymentConfig("stripe"); err == nil {
		var cfg map[string]interface{}
		json.Unmarshal([]byte(stripe.ConfigJSON), &cfg)
		result.StripeEnabled = stripe.Enabled
		if v, ok := cfg["publishable_key"].(string); ok {
			result.StripePublishableKey = v
		}
		if v, ok := cfg["secret_key"].(string); ok {
			result.StripeSecretKey = v
		}
		if v, ok := cfg["webhook_secret"].(string); ok {
			result.StripeWebhookSecret = v
		}
		if v, ok := cfg["currency"].(string); ok {
			result.StripeCurrency = v
		}
	}

	// 获取USDT配置
	if usdt, err := s.repo.GetPaymentConfig("usdt"); err == nil {
		var cfg map[string]interface{}
		json.Unmarshal([]byte(usdt.ConfigJSON), &cfg)
		result.USDTEnabled = usdt.Enabled
		if v, ok := cfg["network"].(string); ok {
			result.USDTNetwork = v
		}
		if v, ok := cfg["wallet_address"].(string); ok {
			result.USDTWalletAddress = v
		}
		if v, ok := cfg["api_provider"].(string); ok {
			result.USDTAPIProvider = v
		}
		if v, ok := cfg["api_key"].(string); ok {
			result.USDTAPIKey = v
		}
		if v, ok := cfg["api_secret"].(string); ok {
			result.USDTAPISecret = v
		}
		if v, ok := cfg["webhook_secret"].(string); ok {
			result.USDTWebhookSecret = v
		}
		if v, ok := cfg["exchange_rate"].(float64); ok {
			result.USDTExchangeRate = v
		}
		if v, ok := cfg["min_amount"].(float64); ok {
			result.USDTMinAmount = v
		}
		if v, ok := cfg["confirmations"].(float64); ok {
			result.USDTConfirmations = int(v)
		}
	}

	return result, nil
}

// SaveAlipayConfig 保存支付宝配置
func (s *ConfigService) SaveAlipayConfig(cfg *config.AlipayF2FConfig) error {
	jsonData, _ := json.Marshal(cfg)
	dbConfig := &model.PaymentConfigDB{
		PaymentType: "alipay_f2f",
		Enabled:     cfg.Enabled,
		ConfigJSON:  string(jsonData),
	}
	return s.repo.SavePaymentConfig(dbConfig)
}

// SaveWechatPayConfig 保存微信支付配置
func (s *ConfigService) SaveWechatPayConfig(cfg *config.WechatPayConfig) error {
	jsonData, _ := json.Marshal(cfg)
	dbConfig := &model.PaymentConfigDB{
		PaymentType: "wechat_pay",
		Enabled:     cfg.Enabled,
		ConfigJSON:  string(jsonData),
	}
	return s.repo.SavePaymentConfig(dbConfig)
}

// SaveYiPayConfig 保存易支付配置
func (s *ConfigService) SaveYiPayConfig(cfg *config.YiPayConfig) error {
	jsonData, _ := json.Marshal(cfg)
	dbConfig := &model.PaymentConfigDB{
		PaymentType: "yi_pay",
		Enabled:     cfg.Enabled,
		ConfigJSON:  string(jsonData),
	}
	return s.repo.SavePaymentConfig(dbConfig)
}

// SavePayPalConfig 保存PayPal配置
func (s *ConfigService) SavePayPalConfig(cfg *config.PayPalConfig) error {
	jsonData, _ := json.Marshal(cfg)
	dbConfig := &model.PaymentConfigDB{
		PaymentType: "paypal",
		Enabled:     cfg.Enabled,
		ConfigJSON:  string(jsonData),
	}
	return s.repo.SavePaymentConfig(dbConfig)
}

// SaveStripeConfig 保存Stripe配置
func (s *ConfigService) SaveStripeConfig(enabled bool, publishableKey, secretKey, webhookSecret, currency string) error {
	cfgData := map[string]interface{}{
		"enabled":         enabled,
		"publishable_key": publishableKey,
		"secret_key":      secretKey,
		"webhook_secret":  webhookSecret,
		"currency":        currency,
	}
	jsonData, _ := json.Marshal(cfgData)
	dbConfig := &model.PaymentConfigDB{
		PaymentType: "stripe",
		Enabled:     enabled,
		ConfigJSON:  string(jsonData),
	}
	return s.repo.SavePaymentConfig(dbConfig)
}

// SaveUSDTConfig 保存USDT配置
func (s *ConfigService) SaveUSDTConfig(enabled bool, network, walletAddress, apiProvider, apiKey, apiSecret, webhookSecret string, exchangeRate, minAmount float64, confirmations int) error {
	cfgData := map[string]interface{}{
		"enabled":        enabled,
		"network":        network,
		"wallet_address": walletAddress,
		"api_provider":   apiProvider,
		"api_key":        apiKey,
		"api_secret":     apiSecret,
		"webhook_secret": webhookSecret,
		"exchange_rate":  exchangeRate,
		"min_amount":     minAmount,
		"confirmations":  confirmations,
	}
	jsonData, _ := json.Marshal(cfgData)
	dbConfig := &model.PaymentConfigDB{
		PaymentType: "usdt",
		Enabled:     enabled,
		ConfigJSON:  string(jsonData),
	}
	return s.repo.SavePaymentConfig(dbConfig)
}
