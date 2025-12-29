package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"user-frontend/internal/config"
)

// PayPalService PayPal支付服务
type PayPalService struct {
	config     *config.PayPalConfig
	httpClient *http.Client
}

// PayPalOrder PayPal订单响应
type PayPalOrder struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Links  []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}

// PayPalCaptureResponse PayPal捕获支付响应
type PayPalCaptureResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PaymentSource struct {
		PayPal struct {
			EmailAddress string `json:"email_address"`
			AccountID    string `json:"account_id"`
		} `json:"paypal"`
	} `json:"payment_source"`
	PurchaseUnits []struct {
		Payments struct {
			Captures []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Amount struct {
					CurrencyCode string `json:"currency_code"`
					Value        string `json:"value"`
				} `json:"amount"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
}

// NewPayPalService 创建PayPal服务
func NewPayPalService(cfg *config.PayPalConfig) *PayPalService {
	return &PayPalService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getBaseURL 获取API基础URL
func (s *PayPalService) getBaseURL() string {
	if s.config.Sandbox {
		return "https://api-m.sandbox.paypal.com"
	}
	return "https://api-m.paypal.com"
}

// getAccessToken 获取访问令牌
func (s *PayPalService) getAccessToken() (string, error) {
	url := s.getBaseURL() + "/v1/oauth2/token"

	req, err := http.NewRequest("POST", url, bytes.NewBufferString("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}

	// 基础认证
	auth := base64.StdEncoding.EncodeToString([]byte(s.config.ClientID + ":" + s.config.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求PayPal失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("获取访问令牌失败: %s", string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	return result.AccessToken, nil
}

// CreateOrder 创建PayPal订单
func (s *PayPalService) CreateOrder(orderNo string, amount float64, description string) (*PayPalOrder, error) {
	if !s.config.Enabled {
		return nil, errors.New("PayPal支付未启用")
	}

	accessToken, err := s.getAccessToken()
	if err != nil {
		return nil, err
	}

	currency := s.config.Currency
	if currency == "" {
		currency = "USD"
	}

	// 构建订单请求
	orderData := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"reference_id": orderNo,
				"description":  description,
				"amount": map[string]interface{}{
					"currency_code": currency,
					"value":         fmt.Sprintf("%.2f", amount),
				},
			},
		},
		"application_context": map[string]interface{}{
			"brand_name":          "卡密购买系统",
			"landing_page":        "LOGIN",
			"user_action":         "PAY_NOW",
			"return_url":          s.config.ReturnURL,
			"cancel_url":          s.config.CancelURL,
			"shipping_preference": "NO_SHIPPING",
		},
	}

	jsonData, _ := json.Marshal(orderData)

	url := s.getBaseURL() + "/v2/checkout/orders"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PayPal-Request-Id", orderNo) // 幂等性

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("创建PayPal订单失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("创建订单失败: %s", string(body))
	}

	var order PayPalOrder
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("解析订单响应失败: %v", err)
	}

	return &order, nil
}

// CaptureOrder 捕获PayPal订单（完成支付）
func (s *PayPalService) CaptureOrder(paypalOrderID string) (*PayPalCaptureResponse, error) {
	if !s.config.Enabled {
		return nil, errors.New("PayPal支付未启用")
	}

	accessToken, err := s.getAccessToken()
	if err != nil {
		return nil, err
	}

	url := s.getBaseURL() + "/v2/checkout/orders/" + paypalOrderID + "/capture"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("捕获PayPal订单失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return nil, fmt.Errorf("捕获订单失败: %s", string(body))
	}

	var captureResp PayPalCaptureResponse
	if err := json.Unmarshal(body, &captureResp); err != nil {
		return nil, fmt.Errorf("解析捕获响应失败: %v", err)
	}

	return &captureResp, nil
}

// GetOrderDetails 获取订单详情
func (s *PayPalService) GetOrderDetails(paypalOrderID string) (*PayPalOrder, error) {
	if !s.config.Enabled {
		return nil, errors.New("PayPal支付未启用")
	}

	accessToken, err := s.getAccessToken()
	if err != nil {
		return nil, err
	}

	url := s.getBaseURL() + "/v2/checkout/orders/" + paypalOrderID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取PayPal订单详情失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("获取订单详情失败: %s", string(body))
	}

	var order PayPalOrder
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("解析订单详情失败: %v", err)
	}

	return &order, nil
}
