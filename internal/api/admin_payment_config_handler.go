package api

import (
	"user-frontend/internal/config"

	"github.com/gin-gonic/gin"
)

// ==================== 支付配置相关 API ====================

// AdminGetPaymentConfig 获取支付配置（从数据库）
func AdminGetPaymentConfig(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	paymentCfg, err := ConfigSvc.GetPaymentConfig()
	if err != nil {
		paymentCfg = &config.PaymentConfig{}
	}

	c.JSON(200, gin.H{
		"success": true,
		"config": gin.H{
			"alipay_f2f": gin.H{
				"enabled":         paymentCfg.AlipayF2F.Enabled,
				"app_id":          paymentCfg.AlipayF2F.AppID,
				"has_private_key": paymentCfg.AlipayF2F.PrivateKey != "",
				"has_public_key":  paymentCfg.AlipayF2F.PublicKey != "",
				"notify_url":      paymentCfg.AlipayF2F.NotifyURL,
			},
			"wechat_pay": gin.H{
				"enabled":     paymentCfg.WechatPay.Enabled,
				"app_id":      paymentCfg.WechatPay.AppID,
				"mch_id":      paymentCfg.WechatPay.MchID,
				"has_api_key": paymentCfg.WechatPay.APIKey != "",
				"notify_url":  paymentCfg.WechatPay.NotifyURL,
			},
			"yi_pay": gin.H{
				"enabled":    paymentCfg.YiPay.Enabled,
				"api_url":    paymentCfg.YiPay.APIURL,
				"pid":        paymentCfg.YiPay.PID,
				"has_key":    paymentCfg.YiPay.Key != "",
				"notify_url": paymentCfg.YiPay.NotifyURL,
				"return_url": paymentCfg.YiPay.ReturnURL,
			},
			"paypal": gin.H{
				"enabled":           paymentCfg.PayPal.Enabled,
				"client_id":         paymentCfg.PayPal.ClientID,
				"has_client_secret": paymentCfg.PayPal.ClientSecret != "",
				"sandbox":           paymentCfg.PayPal.Sandbox,
				"currency":          paymentCfg.PayPal.Currency,
				"return_url":        paymentCfg.PayPal.ReturnURL,
				"cancel_url":        paymentCfg.PayPal.CancelURL,
			},
			"stripe": gin.H{
				"enabled":            paymentCfg.StripeEnabled,
				"publishable_key":    paymentCfg.StripePublishableKey,
				"has_secret_key":     paymentCfg.StripeSecretKey != "",
				"has_webhook_secret": paymentCfg.StripeWebhookSecret != "",
				"currency":           paymentCfg.StripeCurrency,
			},
			"usdt": gin.H{
				"enabled":            paymentCfg.USDTEnabled,
				"network":            paymentCfg.USDTNetwork,
				"wallet_address":     paymentCfg.USDTWalletAddress,
				"api_provider":       paymentCfg.USDTAPIProvider,
				"has_api_key":        paymentCfg.USDTAPIKey != "",
				"has_api_secret":     paymentCfg.USDTAPISecret != "",
				"has_webhook_secret": paymentCfg.USDTWebhookSecret != "",
				"exchange_rate":      paymentCfg.USDTExchangeRate,
				"min_amount":         paymentCfg.USDTMinAmount,
				"confirmations":      paymentCfg.USDTConfirmations,
			},
		},
	})
}

// AdminSavePaymentConfig 保存支付配置（到数据库）
func AdminSavePaymentConfig(c *gin.Context) {
	var req struct {
		PaymentType string `json:"payment_type" binding:"required"`
		// 支付宝当面付
		AlipayEnabled    bool   `json:"alipay_enabled"`
		AlipayAppID      string `json:"alipay_app_id"`
		AlipayPrivateKey string `json:"alipay_private_key"`
		AlipayPublicKey  string `json:"alipay_public_key"`
		AlipayNotifyURL  string `json:"alipay_notify_url"`
		// 微信支付
		WechatEnabled   bool   `json:"wechat_enabled"`
		WechatAppID     string `json:"wechat_app_id"`
		WechatMchID     string `json:"wechat_mch_id"`
		WechatAPIKey    string `json:"wechat_api_key"`
		WechatNotifyURL string `json:"wechat_notify_url"`
		// 易支付
		YiPayEnabled   bool   `json:"yipay_enabled"`
		YiPayAPIURL    string `json:"yipay_api_url"`
		YiPayPID       string `json:"yipay_pid"`
		YiPayKey       string `json:"yipay_key"`
		YiPayNotifyURL string `json:"yipay_notify_url"`
		YiPayReturnURL string `json:"yipay_return_url"`
		// PayPal 配置
		PayPalEnabled      bool   `json:"paypal_enabled"`
		PayPalClientID     string `json:"paypal_client_id"`
		PayPalClientSecret string `json:"paypal_client_secret"`
		PayPalSandbox      bool   `json:"paypal_sandbox"`
		PayPalCurrency     string `json:"paypal_currency"`
		PayPalReturnURL    string `json:"paypal_return_url"`
		PayPalCancelURL    string `json:"paypal_cancel_url"`
		// Stripe 配置
		StripeEnabled        bool   `json:"stripe_enabled"`
		StripePublishableKey string `json:"stripe_publishable_key"`
		StripeSecretKey      string `json:"stripe_secret_key"`
		StripeWebhookSecret  string `json:"stripe_webhook_secret"`
		StripeCurrency       string `json:"stripe_currency"`
		// USDT 配置
		USDTEnabled       bool    `json:"usdt_enabled"`
		USDTNetwork       string  `json:"usdt_network"`
		USDTWalletAddress string  `json:"usdt_wallet_address"`
		USDTAPIProvider   string  `json:"usdt_api_provider"`
		USDTAPIKey        string  `json:"usdt_api_key"`
		USDTAPISecret     string  `json:"usdt_api_secret"`
		USDTWebhookSecret string  `json:"usdt_webhook_secret"`
		USDTExchangeRate  float64 `json:"usdt_exchange_rate"`
		USDTMinAmount     float64 `json:"usdt_min_amount"`
		USDTConfirmations int     `json:"usdt_confirmations"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 优先保存到数据库
	if ConfigSvc != nil {
		var saveErr error
		switch req.PaymentType {
		case "alipay_f2f":
			// 获取现有配置以保留密钥
			existingCfg, _ := ConfigSvc.GetPaymentConfig()
			alipayCfg := &config.AlipayF2FConfig{
				Enabled:   req.AlipayEnabled,
				AppID:     req.AlipayAppID,
				NotifyURL: req.AlipayNotifyURL,
			}
			if req.AlipayPrivateKey != "" {
				alipayCfg.PrivateKey = req.AlipayPrivateKey
			} else if existingCfg != nil {
				alipayCfg.PrivateKey = existingCfg.AlipayF2F.PrivateKey
			}
			if req.AlipayPublicKey != "" {
				alipayCfg.PublicKey = req.AlipayPublicKey
			} else if existingCfg != nil {
				alipayCfg.PublicKey = existingCfg.AlipayF2F.PublicKey
			}
			saveErr = ConfigSvc.SaveAlipayConfig(alipayCfg)
			if saveErr == nil {
				config.GlobalConfig.PaymentConfig.AlipayF2F = *alipayCfg
			}

		case "wechat_pay":
			existingCfg, _ := ConfigSvc.GetPaymentConfig()
			wechatCfg := &config.WechatPayConfig{
				Enabled:   req.WechatEnabled,
				AppID:     req.WechatAppID,
				MchID:     req.WechatMchID,
				NotifyURL: req.WechatNotifyURL,
			}
			if req.WechatAPIKey != "" {
				wechatCfg.APIKey = req.WechatAPIKey
			} else if existingCfg != nil {
				wechatCfg.APIKey = existingCfg.WechatPay.APIKey
			}
			saveErr = ConfigSvc.SaveWechatPayConfig(wechatCfg)
			if saveErr == nil {
				config.GlobalConfig.PaymentConfig.WechatPay = *wechatCfg
			}

		case "yi_pay":
			existingCfg, _ := ConfigSvc.GetPaymentConfig()
			yipayCfg := &config.YiPayConfig{
				Enabled:   req.YiPayEnabled,
				APIURL:    req.YiPayAPIURL,
				PID:       req.YiPayPID,
				NotifyURL: req.YiPayNotifyURL,
				ReturnURL: req.YiPayReturnURL,
			}
			if req.YiPayKey != "" {
				yipayCfg.Key = req.YiPayKey
			} else if existingCfg != nil {
				yipayCfg.Key = existingCfg.YiPay.Key
			}
			saveErr = ConfigSvc.SaveYiPayConfig(yipayCfg)
			if saveErr == nil {
				config.GlobalConfig.PaymentConfig.YiPay = *yipayCfg
			}

		case "paypal":
			existingCfg, _ := ConfigSvc.GetPaymentConfig()
			paypalCfg := &config.PayPalConfig{
				Enabled:   req.PayPalEnabled,
				ClientID:  req.PayPalClientID,
				Sandbox:   req.PayPalSandbox,
				Currency:  req.PayPalCurrency,
				ReturnURL: req.PayPalReturnURL,
				CancelURL: req.PayPalCancelURL,
			}
			if paypalCfg.Currency == "" {
				paypalCfg.Currency = "USD"
			}
			if req.PayPalClientSecret != "" {
				paypalCfg.ClientSecret = req.PayPalClientSecret
			} else if existingCfg != nil {
				paypalCfg.ClientSecret = existingCfg.PayPal.ClientSecret
			}
			saveErr = ConfigSvc.SavePayPalConfig(paypalCfg)
			if saveErr == nil {
				config.GlobalConfig.PaymentConfig.PayPal = *paypalCfg
			}

		case "stripe":
			existingCfg, _ := ConfigSvc.GetPaymentConfig()
			// 保存Stripe配置到全局配置
			config.GlobalConfig.PaymentConfig.StripeEnabled = req.StripeEnabled
			config.GlobalConfig.PaymentConfig.StripePublishableKey = req.StripePublishableKey
			if req.StripeSecretKey != "" {
				config.GlobalConfig.PaymentConfig.StripeSecretKey = req.StripeSecretKey
			} else if existingCfg != nil {
				config.GlobalConfig.PaymentConfig.StripeSecretKey = existingCfg.StripeSecretKey
			}
			if req.StripeWebhookSecret != "" {
				config.GlobalConfig.PaymentConfig.StripeWebhookSecret = req.StripeWebhookSecret
			} else if existingCfg != nil {
				config.GlobalConfig.PaymentConfig.StripeWebhookSecret = existingCfg.StripeWebhookSecret
			}
			config.GlobalConfig.PaymentConfig.StripeCurrency = req.StripeCurrency
			if config.GlobalConfig.PaymentConfig.StripeCurrency == "" {
				config.GlobalConfig.PaymentConfig.StripeCurrency = "usd"
			}
			// 保存到数据库
			saveErr = ConfigSvc.SaveStripeConfig(req.StripeEnabled, req.StripePublishableKey,
				config.GlobalConfig.PaymentConfig.StripeSecretKey,
				config.GlobalConfig.PaymentConfig.StripeWebhookSecret,
				config.GlobalConfig.PaymentConfig.StripeCurrency)
			// 重新初始化Stripe服务
			if saveErr == nil {
				InitStripeService(config.GlobalConfig)
			}

		case "usdt":
			existingCfg, _ := ConfigSvc.GetPaymentConfig()
			// 保存USDT配置到全局配置
			config.GlobalConfig.PaymentConfig.USDTEnabled = req.USDTEnabled
			config.GlobalConfig.PaymentConfig.USDTNetwork = req.USDTNetwork
			config.GlobalConfig.PaymentConfig.USDTWalletAddress = req.USDTWalletAddress
			config.GlobalConfig.PaymentConfig.USDTAPIProvider = req.USDTAPIProvider
			if req.USDTAPIKey != "" {
				config.GlobalConfig.PaymentConfig.USDTAPIKey = req.USDTAPIKey
			} else if existingCfg != nil {
				config.GlobalConfig.PaymentConfig.USDTAPIKey = existingCfg.USDTAPIKey
			}
			if req.USDTAPISecret != "" {
				config.GlobalConfig.PaymentConfig.USDTAPISecret = req.USDTAPISecret
			} else if existingCfg != nil {
				config.GlobalConfig.PaymentConfig.USDTAPISecret = existingCfg.USDTAPISecret
			}
			if req.USDTWebhookSecret != "" {
				config.GlobalConfig.PaymentConfig.USDTWebhookSecret = req.USDTWebhookSecret
			} else if existingCfg != nil {
				config.GlobalConfig.PaymentConfig.USDTWebhookSecret = existingCfg.USDTWebhookSecret
			}
			config.GlobalConfig.PaymentConfig.USDTExchangeRate = req.USDTExchangeRate
			config.GlobalConfig.PaymentConfig.USDTMinAmount = req.USDTMinAmount
			config.GlobalConfig.PaymentConfig.USDTConfirmations = req.USDTConfirmations
			// 保存到数据库
			saveErr = ConfigSvc.SaveUSDTConfig(
				req.USDTEnabled, req.USDTNetwork, req.USDTWalletAddress, req.USDTAPIProvider,
				config.GlobalConfig.PaymentConfig.USDTAPIKey,
				config.GlobalConfig.PaymentConfig.USDTAPISecret,
				config.GlobalConfig.PaymentConfig.USDTWebhookSecret,
				req.USDTExchangeRate, req.USDTMinAmount, req.USDTConfirmations)
			// 重新初始化USDT服务
			if saveErr == nil {
				InitUSDTService(config.GlobalConfig)
			}

		default:
			c.JSON(400, gin.H{"success": false, "error": "未知的支付类型"})
			return
		}

		if saveErr != nil {
			c.JSON(500, gin.H{"success": false, "error": "保存配置失败: " + saveErr.Error()})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": "支付配置已保存"})
		return
	}

	c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
}
