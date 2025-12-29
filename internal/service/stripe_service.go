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
	"net/url"
	"strconv"
	"strings"
	"time"

	"user-frontend/internal/config"
)

// StripeService Stripe支付服务
type StripeService struct {
	cfg *config.Config
}

// NewStripeService 创建Stripe支付服务
func NewStripeService(cfg *config.Config) *StripeService {
	return &StripeService{cfg: cfg}
}

// StripeConfig Stripe配置
type StripeConfig struct {
	Enabled       bool   `json:"enabled"`
	PublishableKey string `json:"publishable_key"` // 公钥（前端使用）
	SecretKey     string `json:"secret_key"`       // 私钥（后端使用）
	WebhookSecret string `json:"webhook_secret"`   // Webhook签名密钥
	Currency      string `json:"currency"`         // 货币代码，默认usd
}

// StripeCheckoutSession Stripe Checkout会话
type StripeCheckoutSession struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	PaymentIntent string `json:"payment_intent"`
	Status     string `json:"status"`
}

// StripePaymentIntent Stripe支付意图
type StripePaymentIntent struct {
	ID            string `json:"id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	ClientSecret  string `json:"client_secret"`
}

// StripeWebhookEvent Stripe Webhook事件
type StripeWebhookEvent struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
	Created int64           `json:"created"`
}

// CreateCheckoutSession 创建Stripe Checkout会话
// 参数：
//   - orderNo: 订单号
//   - amount: 金额（单位：分）
//   - productName: 商品名称
//   - successURL: 支付成功跳转URL
//   - cancelURL: 支付取消跳转URL
func (s *StripeService) CreateCheckoutSession(orderNo string, amount int64, productName, successURL, cancelURL string) (*StripeCheckoutSession, error) {
	stripeCfg := s.getStripeConfig()
	if !stripeCfg.Enabled || stripeCfg.SecretKey == "" {
		return nil, errors.New("Stripe支付未启用或未配置")
	}

	currency := stripeCfg.Currency
	if currency == "" {
		currency = "usd"
	}

	// 构建请求参数
	data := url.Values{}
	data.Set("mode", "payment")
	data.Set("success_url", successURL+"?session_id={CHECKOUT_SESSION_ID}")
	data.Set("cancel_url", cancelURL)
	data.Set("client_reference_id", orderNo)
	data.Set("line_items[0][price_data][currency]", currency)
	data.Set("line_items[0][price_data][product_data][name]", productName)
	data.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(amount, 10))
	data.Set("line_items[0][quantity]", "1")
	data.Set("metadata[order_no]", orderNo)

	// 发送请求
	req, err := http.NewRequest("POST", "https://api.stripe.com/v1/checkout/sessions", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+stripeCfg.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Stripe API错误: %s", string(body))
	}

	var session StripeCheckoutSession
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// CreatePaymentIntent 创建支付意图（用于自定义支付流程）
func (s *StripeService) CreatePaymentIntent(orderNo string, amount int64, description string) (*StripePaymentIntent, error) {
	stripeCfg := s.getStripeConfig()
	if !stripeCfg.Enabled || stripeCfg.SecretKey == "" {
		return nil, errors.New("Stripe支付未启用或未配置")
	}

	currency := stripeCfg.Currency
	if currency == "" {
		currency = "usd"
	}

	data := url.Values{}
	data.Set("amount", strconv.FormatInt(amount, 10))
	data.Set("currency", currency)
	data.Set("description", description)
	data.Set("metadata[order_no]", orderNo)
	data.Set("automatic_payment_methods[enabled]", "true")

	req, err := http.NewRequest("POST", "https://api.stripe.com/v1/payment_intents", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+stripeCfg.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Stripe API错误: %s", string(body))
	}

	var intent StripePaymentIntent
	if err := json.Unmarshal(body, &intent); err != nil {
		return nil, err
	}

	return &intent, nil
}

// RetrieveCheckoutSession 获取Checkout会话详情
func (s *StripeService) RetrieveCheckoutSession(sessionID string) (*StripeCheckoutSession, error) {
	stripeCfg := s.getStripeConfig()
	if stripeCfg.SecretKey == "" {
		return nil, errors.New("Stripe未配置")
	}

	req, err := http.NewRequest("GET", "https://api.stripe.com/v1/checkout/sessions/"+sessionID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+stripeCfg.SecretKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Stripe API错误: %s", string(body))
	}

	var session StripeCheckoutSession
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// RetrievePaymentIntent 获取支付意图详情
func (s *StripeService) RetrievePaymentIntent(intentID string) (*StripePaymentIntent, error) {
	stripeCfg := s.getStripeConfig()
	if stripeCfg.SecretKey == "" {
		return nil, errors.New("Stripe未配置")
	}

	req, err := http.NewRequest("GET", "https://api.stripe.com/v1/payment_intents/"+intentID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+stripeCfg.SecretKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Stripe API错误: %s", string(body))
	}

	var intent StripePaymentIntent
	if err := json.Unmarshal(body, &intent); err != nil {
		return nil, err
	}

	return &intent, nil
}

// VerifyWebhookSignature 验证Webhook签名
func (s *StripeService) VerifyWebhookSignature(payload []byte, signature string) (*StripeWebhookEvent, error) {
	stripeCfg := s.getStripeConfig()
	if stripeCfg.WebhookSecret == "" {
		return nil, errors.New("Webhook密钥未配置")
	}

	// 解析签名头
	parts := strings.Split(signature, ",")
	var timestamp string
	var signatures []string

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			signatures = append(signatures, kv[1])
		}
	}

	if timestamp == "" || len(signatures) == 0 {
		return nil, errors.New("无效的签名格式")
	}

	// 验证时间戳（5分钟内有效）
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, errors.New("无效的时间戳")
	}
	if time.Now().Unix()-ts > 300 {
		return nil, errors.New("签名已过期")
	}

	// 计算预期签名
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(stripeCfg.WebhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// 验证签名
	valid := false
	for _, sig := range signatures {
		if hmac.Equal([]byte(sig), []byte(expectedSig)) {
			valid = true
			break
		}
	}

	if !valid {
		return nil, errors.New("签名验证失败")
	}

	// 解析事件
	var event StripeWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// ParseCheckoutSessionCompleted 解析checkout.session.completed事件
func (s *StripeService) ParseCheckoutSessionCompleted(data json.RawMessage) (string, error) {
	orderNo, _, err := s.ParseCheckoutSessionCompletedWithAmount(data)
	return orderNo, err
}

// ParseCheckoutSessionCompletedWithAmount 解析checkout.session.completed事件（包含金额）
// 返回值：
//   - orderNo: 订单号
//   - paidAmount: 实际支付金额（法币，元）
//   - err: 错误信息
func (s *StripeService) ParseCheckoutSessionCompletedWithAmount(data json.RawMessage) (string, float64, error) {
	var wrapper struct {
		Object struct {
			ClientReferenceID string `json:"client_reference_id"`
			PaymentStatus     string `json:"payment_status"`
			AmountTotal       int64  `json:"amount_total"` // 总金额（分）
			Currency          string `json:"currency"`
			Metadata          struct {
				OrderNo string `json:"order_no"`
			} `json:"metadata"`
		} `json:"object"`
	}

	if err := json.Unmarshal(data, &wrapper); err != nil {
		return "", 0, err
	}

	// 优先使用metadata中的订单号
	orderNo := wrapper.Object.Metadata.OrderNo
	if orderNo == "" {
		orderNo = wrapper.Object.ClientReferenceID
	}

	if wrapper.Object.PaymentStatus != "paid" {
		return "", 0, errors.New("支付未完成")
	}

	// 将分转换为元
	paidAmount := centsToAmount(wrapper.Object.AmountTotal)

	return orderNo, paidAmount, nil
}

// ParsePaymentIntentSucceeded 解析payment_intent.succeeded事件
func (s *StripeService) ParsePaymentIntentSucceeded(data json.RawMessage) (string, error) {
	orderNo, _, err := s.ParsePaymentIntentSucceededWithAmount(data)
	return orderNo, err
}

// ParsePaymentIntentSucceededWithAmount 解析payment_intent.succeeded事件（包含金额）
// 返回值：
//   - orderNo: 订单号
//   - paidAmount: 实际支付金额（法币，元）
//   - err: 错误信息
func (s *StripeService) ParsePaymentIntentSucceededWithAmount(data json.RawMessage) (string, float64, error) {
	var wrapper struct {
		Object struct {
			ID       string `json:"id"`
			Amount   int64  `json:"amount"`   // 金额（分）
			Currency string `json:"currency"`
			Status   string `json:"status"`
			Metadata struct {
				OrderNo string `json:"order_no"`
			} `json:"metadata"`
		} `json:"object"`
	}

	if err := json.Unmarshal(data, &wrapper); err != nil {
		return "", 0, err
	}

	if wrapper.Object.Status != "succeeded" {
		return "", 0, errors.New("支付未成功")
	}

	// 将分转换为元
	paidAmount := centsToAmount(wrapper.Object.Amount)

	return wrapper.Object.Metadata.OrderNo, paidAmount, nil
}

// getStripeConfig 获取Stripe配置
func (s *StripeService) getStripeConfig() *StripeConfig {
	// 从支付配置中获取Stripe配置
	if s.cfg == nil {
		return &StripeConfig{}
	}

	return &StripeConfig{
		Enabled:        s.cfg.PaymentConfig.StripeEnabled,
		PublishableKey: s.cfg.PaymentConfig.StripePublishableKey,
		SecretKey:      s.cfg.PaymentConfig.StripeSecretKey,
		WebhookSecret:  s.cfg.PaymentConfig.StripeWebhookSecret,
		Currency:       s.cfg.PaymentConfig.StripeCurrency,
	}
}

// GetPublishableKey 获取公钥（前端使用）
func (s *StripeService) GetPublishableKey() string {
	cfg := s.getStripeConfig()
	return cfg.PublishableKey
}

// IsEnabled 检查Stripe是否启用
func (s *StripeService) IsEnabled() bool {
	cfg := s.getStripeConfig()
	return cfg.Enabled && cfg.SecretKey != ""
}

// CreateRefund 创建退款
func (s *StripeService) CreateRefund(paymentIntentID string, amount int64, reason string) error {
	stripeCfg := s.getStripeConfig()
	if stripeCfg.SecretKey == "" {
		return errors.New("Stripe未配置")
	}

	data := url.Values{}
	data.Set("payment_intent", paymentIntentID)
	if amount > 0 {
		data.Set("amount", strconv.FormatInt(amount, 10))
	}
	if reason != "" {
		data.Set("reason", reason)
	}

	req, err := http.NewRequest("POST", "https://api.stripe.com/v1/refunds", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+stripeCfg.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("退款失败: %s", string(body))
	}

	return nil
}

// TestConnection 测试Stripe连接
func (s *StripeService) TestConnection() error {
	stripeCfg := s.getStripeConfig()
	if stripeCfg.SecretKey == "" {
		return errors.New("Stripe密钥未配置")
	}

	req, err := http.NewRequest("GET", "https://api.stripe.com/v1/balance", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+stripeCfg.SecretKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API错误: %s", string(body))
	}

	return nil
}

// amountToCents 将金额转换为分（Stripe使用最小货币单位）
func amountToCents(amount float64) int64 {
	return int64(amount * 100)
}

// centsToAmount 将分转换为金额
func centsToAmount(cents int64) float64 {
	return float64(cents) / 100
}

// CreateCheckoutSessionForOrder 为订单创建Checkout会话（便捷方法）
func (s *StripeService) CreateCheckoutSessionForOrder(orderNo string, amount float64, productName, baseURL string) (*StripeCheckoutSession, error) {
	cents := amountToCents(amount)
	successURL := baseURL + "/payment/result?order_no=" + orderNo
	cancelURL := baseURL + "/payment/cancel?order_no=" + orderNo

	return s.CreateCheckoutSession(orderNo, cents, productName, successURL, cancelURL)
}

// StripePaymentResult Stripe支付结果
type StripePaymentResult struct {
	Success   bool   `json:"success"`
	OrderNo   string `json:"order_no"`
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// VerifyPayment 验证支付状态
func (s *StripeService) VerifyPayment(sessionID string) (*StripePaymentResult, error) {
	session, err := s.RetrieveCheckoutSession(sessionID)
	if err != nil {
		return nil, err
	}

	result := &StripePaymentResult{
		SessionID: sessionID,
		Status:    session.Status,
	}

	if session.Status == "complete" {
		result.Success = true
		result.Message = "支付成功"
	} else {
		result.Success = false
		result.Message = "支付未完成"
	}

	return result, nil
}
