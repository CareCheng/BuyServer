package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"user-frontend/internal/config"
)

// YiPayService 易支付服务
type YiPayService struct {
	config *config.YiPayConfig
}

// NewYiPayService 创建易支付服务
func NewYiPayService(cfg *config.YiPayConfig) *YiPayService {
	return &YiPayService{config: cfg}
}

// CreateOrder 创建易支付订单
// 返回支付页面URL
// 参数：
//   - orderNo: 商户订单号
//   - amount: 订单金额（元）
//   - productName: 商品名称
//
// 返回：
//   - 支付页面URL
//   - 错误信息
func (s *YiPayService) CreateOrder(orderNo string, amount float64, productName string) (string, error) {
	if !s.config.Enabled {
		return "", errors.New("易支付未启用")
	}

	if s.config.APIURL == "" || s.config.PID == "" || s.config.Key == "" {
		return "", errors.New("易支付配置不完整")
	}

	// 构建请求参数
	params := map[string]string{
		"pid":          s.config.PID,
		"type":         "alipay", // 默认使用支付宝，可根据需求调整
		"out_trade_no": orderNo,
		"notify_url":   s.config.NotifyURL,
		"return_url":   s.config.ReturnURL,
		"name":         productName,
		"money":        fmt.Sprintf("%.2f", amount),
	}

	// 生成签名
	params["sign"] = s.sign(params)
	params["sign_type"] = "MD5"

	// 构建支付URL
	payURL := s.config.APIURL
	if !strings.HasSuffix(payURL, "/") {
		payURL += "/"
	}
	payURL += "submit.php?"

	// 添加参数
	var urlParams []string
	for k, v := range params {
		urlParams = append(urlParams, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
	}
	payURL += strings.Join(urlParams, "&")

	return payURL, nil
}

// VerifyNotify 验证易支付异步通知
// 参数：
//   - request: HTTP请求
//
// 返回：
//   - 商户订单号
//   - 易支付交易号
//   - 错误信息
func (s *YiPayService) VerifyNotify(request *http.Request) (string, string, error) {
	if !s.config.Enabled {
		return "", "", errors.New("易支付未启用")
	}

	// 解析参数
	if err := request.ParseForm(); err != nil {
		return "", "", fmt.Errorf("解析参数失败: %v", err)
	}

	// 获取参数
	outTradeNo := request.Form.Get("out_trade_no")
	tradeNo := request.Form.Get("trade_no")
	tradeStatus := request.Form.Get("trade_status")
	sign := request.Form.Get("sign")

	// 验证交易状态
	if tradeStatus != "TRADE_SUCCESS" {
		return "", "", errors.New("交易未成功")
	}

	// 构建验签参数
	params := map[string]string{
		"pid":          request.Form.Get("pid"),
		"trade_no":     tradeNo,
		"out_trade_no": outTradeNo,
		"type":         request.Form.Get("type"),
		"name":         request.Form.Get("name"),
		"money":        request.Form.Get("money"),
		"trade_status": tradeStatus,
	}

	// 验证签名
	expectedSign := s.sign(params)
	if expectedSign != sign {
		return "", "", errors.New("签名验证失败")
	}

	return outTradeNo, tradeNo, nil
}

// VerifyReturn 验证易支付同步返回
// 参数：
//   - request: HTTP请求
//
// 返回：
//   - 商户订单号
//   - 易支付交易号
//   - 错误信息
func (s *YiPayService) VerifyReturn(request *http.Request) (string, string, error) {
	return s.VerifyNotify(request)
}

// sign 生成签名
// 易支付签名规则：
// 1. 将参数按照参数名ASCII码从小到大排序
// 2. 使用 key=value 格式拼接成字符串
// 3. 在最后拼接 商户密钥
// 4. 对整个字符串进行MD5
func (s *YiPayService) sign(params map[string]string) string {
	// 获取排序后的键
	var keys []string
	for k := range params {
		if params[k] != "" && k != "sign" && k != "sign_type" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	signStr := strings.Join(parts, "&") + s.config.Key

	// MD5签名
	hash := md5.Sum([]byte(signStr))
	return hex.EncodeToString(hash[:])
}
