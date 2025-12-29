package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"user-frontend/internal/config"
)

// USDTService USDT支付服务
type USDTService struct {
	cfg *config.Config
}

// NewUSDTService 创建USDT支付服务
func NewUSDTService(cfg *config.Config) *USDTService {
	return &USDTService{cfg: cfg}
}

// USDTConfig USDT支付配置
type USDTConfig struct {
	Enabled       bool    `json:"enabled"`         // 是否启用
	Network       string  `json:"network"`         // 网络类型：TRC20, ERC20, BEP20
	WalletAddress string  `json:"wallet_address"`  // 收款钱包地址
	APIProvider   string  `json:"api_provider"`    // API提供商：nowpayments, coingate, manual
	APIKey        string  `json:"api_key"`         // API密钥
	APISecret     string  `json:"api_secret"`      // API密钥（部分提供商需要）
	WebhookSecret string  `json:"webhook_secret"`  // Webhook签名密钥
	ExchangeRate  float64 `json:"exchange_rate"`   // 汇率（手动模式使用）
	MinAmount     float64 `json:"min_amount"`      // 最小支付金额（USDT）
	Confirmations int     `json:"confirmations"`   // 需要的确认数
}

// USDTPaymentRequest USDT支付请求
type USDTPaymentRequest struct {
	OrderNo     string  `json:"order_no"`
	Amount      float64 `json:"amount"`       // 原始金额（法币）
	Currency    string  `json:"currency"`     // 原始货币
	Description string  `json:"description"`
}

// USDTPaymentResponse USDT支付响应
type USDTPaymentResponse struct {
	PaymentID     string  `json:"payment_id"`      // 支付ID
	WalletAddress string  `json:"wallet_address"`  // 收款地址
	Amount        float64 `json:"amount"`          // USDT金额
	Network       string  `json:"network"`         // 网络类型
	ExpiresAt     int64   `json:"expires_at"`      // 过期时间戳
	QRCode        string  `json:"qr_code"`         // 二维码内容
	PaymentURL    string  `json:"payment_url"`     // 支付页面URL（如果有）
}

// USDTPaymentStatus USDT支付状态
type USDTPaymentStatus struct {
	PaymentID     string  `json:"payment_id"`
	Status        string  `json:"status"`        // waiting, confirming, confirmed, expired, failed
	Confirmations int     `json:"confirmations"` // 当前确认数
	TxHash        string  `json:"tx_hash"`       // 交易哈希
	AmountPaid    float64 `json:"amount_paid"`   // 实际支付金额
}

// CreatePayment 创建USDT支付
// 参数：
//   - req: 支付请求
// 返回：
//   - 支付响应
//   - 错误信息
func (s *USDTService) CreatePayment(req *USDTPaymentRequest) (*USDTPaymentResponse, error) {
	cfg := s.getUSDTConfig()
	if !cfg.Enabled {
		return nil, errors.New("USDT支付未启用")
	}

	switch cfg.APIProvider {
	case "nowpayments":
		return s.createNowPaymentsPayment(req, cfg)
	case "coingate":
		return s.createCoinGatePayment(req, cfg)
	case "manual":
		return s.createManualPayment(req, cfg)
	default:
		return nil, errors.New("不支持的API提供商")
	}
}

// createNowPaymentsPayment 通过NOWPayments创建支付
func (s *USDTService) createNowPaymentsPayment(req *USDTPaymentRequest, cfg *USDTConfig) (*USDTPaymentResponse, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("NOWPayments API密钥未配置")
	}

	// 构建请求
	payload := map[string]interface{}{
		"price_amount":      req.Amount,
		"price_currency":    strings.ToLower(req.Currency),
		"pay_currency":      "usdttrc20", // 默认TRC20
		"order_id":          req.OrderNo,
		"order_description": req.Description,
	}

	// 根据网络类型选择币种
	switch cfg.Network {
	case "ERC20":
		payload["pay_currency"] = "usdterc20"
	case "BEP20":
		payload["pay_currency"] = "usdtbsc"
	}

	jsonData, _ := json.Marshal(payload)

	httpReq, err := http.NewRequest("POST", "https://api.nowpayments.io/v1/payment", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("x-api-key", cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("API错误: %s", string(body))
	}

	var result struct {
		PaymentID      string  `json:"payment_id"`
		PayAddress     string  `json:"pay_address"`
		PayAmount      float64 `json:"pay_amount"`
		PayCurrency    string  `json:"pay_currency"`
		ExpirationTime string  `json:"expiration_estimate_date"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// 解析过期时间
	expiresAt := time.Now().Add(30 * time.Minute).Unix()
	if result.ExpirationTime != "" {
		if t, err := time.Parse(time.RFC3339, result.ExpirationTime); err == nil {
			expiresAt = t.Unix()
		}
	}

	return &USDTPaymentResponse{
		PaymentID:     result.PaymentID,
		WalletAddress: result.PayAddress,
		Amount:        result.PayAmount,
		Network:       cfg.Network,
		ExpiresAt:     expiresAt,
		QRCode:        result.PayAddress,
	}, nil
}

// createCoinGatePayment 通过CoinGate创建支付
func (s *USDTService) createCoinGatePayment(req *USDTPaymentRequest, cfg *USDTConfig) (*USDTPaymentResponse, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("CoinGate API密钥未配置")
	}

	// 构建请求
	payload := map[string]interface{}{
		"order_id":         req.OrderNo,
		"price_amount":     req.Amount,
		"price_currency":   strings.ToUpper(req.Currency),
		"receive_currency": "USDT",
		"title":            req.Description,
	}

	jsonData, _ := json.Marshal(payload)

	httpReq, err := http.NewRequest("POST", "https://api.coingate.com/v2/orders", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Token "+cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("API错误: %s", string(body))
	}

	var result struct {
		ID         int     `json:"id"`
		PayAmount  float64 `json:"pay_amount"`
		PayAddress string  `json:"pay_address"`
		PaymentURL string  `json:"payment_url"`
		ExpireAt   string  `json:"expire_at"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(30 * time.Minute).Unix()
	if result.ExpireAt != "" {
		if t, err := time.Parse(time.RFC3339, result.ExpireAt); err == nil {
			expiresAt = t.Unix()
		}
	}

	return &USDTPaymentResponse{
		PaymentID:     strconv.Itoa(result.ID),
		WalletAddress: result.PayAddress,
		Amount:        result.PayAmount,
		Network:       cfg.Network,
		ExpiresAt:     expiresAt,
		QRCode:        result.PayAddress,
		PaymentURL:    result.PaymentURL,
	}, nil
}

// createManualPayment 创建手动支付（显示钱包地址）
func (s *USDTService) createManualPayment(req *USDTPaymentRequest, cfg *USDTConfig) (*USDTPaymentResponse, error) {
	if cfg.WalletAddress == "" {
		return nil, errors.New("收款钱包地址未配置")
	}

	// 计算USDT金额
	usdtAmount := req.Amount
	if cfg.ExchangeRate > 0 {
		usdtAmount = req.Amount / cfg.ExchangeRate
	}

	// 检查最小金额
	if cfg.MinAmount > 0 && usdtAmount < cfg.MinAmount {
		return nil, fmt.Errorf("支付金额不能小于 %.2f USDT", cfg.MinAmount)
	}

	return &USDTPaymentResponse{
		PaymentID:     req.OrderNo,
		WalletAddress: cfg.WalletAddress,
		Amount:        usdtAmount,
		Network:       cfg.Network,
		ExpiresAt:     time.Now().Add(30 * time.Minute).Unix(),
		QRCode:        cfg.WalletAddress,
	}, nil
}

// GetPaymentStatus 获取支付状态
func (s *USDTService) GetPaymentStatus(paymentID string) (*USDTPaymentStatus, error) {
	cfg := s.getUSDTConfig()
	if !cfg.Enabled {
		return nil, errors.New("USDT支付未启用")
	}

	switch cfg.APIProvider {
	case "nowpayments":
		return s.getNowPaymentsStatus(paymentID, cfg)
	case "coingate":
		return s.getCoinGateStatus(paymentID, cfg)
	case "manual":
		// 手动模式需要管理员确认
		return &USDTPaymentStatus{
			PaymentID: paymentID,
			Status:    "waiting",
		}, nil
	default:
		return nil, errors.New("不支持的API提供商")
	}
}

// getNowPaymentsStatus 获取NOWPayments支付状态
func (s *USDTService) getNowPaymentsStatus(paymentID string, cfg *USDTConfig) (*USDTPaymentStatus, error) {
	httpReq, err := http.NewRequest("GET", "https://api.nowpayments.io/v1/payment/"+paymentID, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("x-api-key", cfg.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API错误: %s", string(body))
	}

	var result struct {
		PaymentID      string  `json:"payment_id"`
		PaymentStatus  string  `json:"payment_status"`
		ActuallyPaid   float64 `json:"actually_paid"`
		PayinHash      string  `json:"payin_hash"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// 转换状态
	status := "waiting"
	switch result.PaymentStatus {
	case "waiting":
		status = "waiting"
	case "confirming":
		status = "confirming"
	case "confirmed", "sending", "partially_paid", "finished":
		status = "confirmed"
	case "expired":
		status = "expired"
	case "failed", "refunded":
		status = "failed"
	}

	return &USDTPaymentStatus{
		PaymentID:  result.PaymentID,
		Status:     status,
		TxHash:     result.PayinHash,
		AmountPaid: result.ActuallyPaid,
	}, nil
}

// getCoinGateStatus 获取CoinGate支付状态
func (s *USDTService) getCoinGateStatus(paymentID string, cfg *USDTConfig) (*USDTPaymentStatus, error) {
	httpReq, err := http.NewRequest("GET", "https://api.coingate.com/v2/orders/"+paymentID, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Token "+cfg.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API错误: %s", string(body))
	}

	var result struct {
		ID            int     `json:"id"`
		Status        string  `json:"status"`
		ReceiveAmount float64 `json:"receive_amount"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// 转换状态
	status := "waiting"
	switch result.Status {
	case "new", "pending":
		status = "waiting"
	case "confirming":
		status = "confirming"
	case "paid":
		status = "confirmed"
	case "expired", "canceled":
		status = "expired"
	case "invalid":
		status = "failed"
	}

	return &USDTPaymentStatus{
		PaymentID:  strconv.Itoa(result.ID),
		Status:     status,
		AmountPaid: result.ReceiveAmount,
	}, nil
}

// VerifyWebhook 验证Webhook签名
// 安全特性：强制验证签名（如果未配置密钥则拒绝所有请求）
func (s *USDTService) VerifyWebhook(payload []byte, signature string) error {
	cfg := s.getUSDTConfig()
	
	// 【安全检查】如果未配置Webhook密钥，拒绝所有请求
	// 这可防止攻击者伪造支付回调
	if cfg.WebhookSecret == "" {
		return errors.New("Webhook密钥未配置，无法验证回调请求")
	}
	
	// 签名不能为空
	if signature == "" {
		return errors.New("缺少签名头")
	}

	// 计算HMAC签名
	mac := hmac.New(sha256.New, []byte(cfg.WebhookSecret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return errors.New("签名验证失败")
	}

	return nil
}

// ParseWebhookEvent 解析Webhook事件
func (s *USDTService) ParseWebhookEvent(payload []byte) (orderNo string, status string, err error) {
	orderNo, status, _, err = s.ParseWebhookEventWithAmount(payload)
	return
}

// ParseWebhookEventWithAmount 解析Webhook事件（包含支付金额）
// 返回值：
//   - orderNo: 订单号
//   - status: 支付状态
//   - paidAmount: 实际支付金额（法币）
//   - err: 错误信息
func (s *USDTService) ParseWebhookEventWithAmount(payload []byte) (orderNo string, status string, paidAmount float64, err error) {
	cfg := s.getUSDTConfig()

	switch cfg.APIProvider {
	case "nowpayments":
		var event struct {
			PaymentID      string  `json:"payment_id"`
			PaymentStatus  string  `json:"payment_status"`
			OrderID        string  `json:"order_id"`
			PriceAmount    float64 `json:"price_amount"`    // 原始订单金额（法币）
			ActuallyPaid   float64 `json:"actually_paid"`   // 实际支付的USDT金额
			PayCurrency    string  `json:"pay_currency"`
			OutcomeAmount  float64 `json:"outcome_amount"`  // 结算金额
		}
		if err := json.Unmarshal(payload, &event); err != nil {
			return "", "", 0, err
		}
		// 返回订单的原始法币金额（用于验证）
		return event.OrderID, event.PaymentStatus, event.PriceAmount, nil

	case "coingate":
		var event struct {
			ID            int     `json:"id"`
			OrderID       string  `json:"order_id"`
			Status        string  `json:"status"`
			PriceAmount   float64 `json:"price_amount"`   // 订单金额
			ReceiveAmount float64 `json:"receive_amount"` // 实际收到金额
		}
		if err := json.Unmarshal(payload, &event); err != nil {
			return "", "", 0, err
		}
		return event.OrderID, event.Status, event.PriceAmount, nil

	default:
		return "", "", 0, errors.New("不支持的API提供商")
	}
}

// getUSDTConfig 获取USDT配置
func (s *USDTService) getUSDTConfig() *USDTConfig {
	if s.cfg == nil {
		return &USDTConfig{}
	}

	return &USDTConfig{
		Enabled:       s.cfg.PaymentConfig.USDTEnabled,
		Network:       s.cfg.PaymentConfig.USDTNetwork,
		WalletAddress: s.cfg.PaymentConfig.USDTWalletAddress,
		APIProvider:   s.cfg.PaymentConfig.USDTAPIProvider,
		APIKey:        s.cfg.PaymentConfig.USDTAPIKey,
		APISecret:     s.cfg.PaymentConfig.USDTAPISecret,
		WebhookSecret: s.cfg.PaymentConfig.USDTWebhookSecret,
		ExchangeRate:  s.cfg.PaymentConfig.USDTExchangeRate,
		MinAmount:     s.cfg.PaymentConfig.USDTMinAmount,
		Confirmations: s.cfg.PaymentConfig.USDTConfirmations,
	}
}

// IsEnabled 检查USDT支付是否启用
func (s *USDTService) IsEnabled() bool {
	cfg := s.getUSDTConfig()
	return cfg.Enabled
}

// GetConfig 获取USDT配置（前端使用，隐藏敏感信息）
func (s *USDTService) GetConfig() map[string]interface{} {
	cfg := s.getUSDTConfig()
	return map[string]interface{}{
		"enabled":        cfg.Enabled,
		"network":        cfg.Network,
		"wallet_address": cfg.WalletAddress,
		"api_provider":   cfg.APIProvider,
		"min_amount":     cfg.MinAmount,
		"exchange_rate":  cfg.ExchangeRate,
	}
}

// TestConnection 测试连接
func (s *USDTService) TestConnection() error {
	cfg := s.getUSDTConfig()
	if !cfg.Enabled {
		return errors.New("USDT支付未启用")
	}

	switch cfg.APIProvider {
	case "nowpayments":
		httpReq, _ := http.NewRequest("GET", "https://api.nowpayments.io/v1/status", nil)
		httpReq.Header.Set("x-api-key", cfg.APIKey)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			return fmt.Errorf("连接失败: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New("API连接失败")
		}
		return nil

	case "coingate":
		httpReq, _ := http.NewRequest("GET", "https://api.coingate.com/v2/ping", nil)
		httpReq.Header.Set("Authorization", "Token "+cfg.APIKey)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			return fmt.Errorf("连接失败: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New("API连接失败")
		}
		return nil

	case "manual":
		if cfg.WalletAddress == "" {
			return errors.New("钱包地址未配置")
		}
		return nil

	default:
		return errors.New("不支持的API提供商")
	}
}
