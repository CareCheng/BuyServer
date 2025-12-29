package service

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"user-frontend/internal/config"
)

// AlipayService 支付宝当面付服务
type AlipayService struct {
	config *config.AlipayF2FConfig
}

// NewAlipayService 创建支付宝服务
func NewAlipayService(cfg *config.AlipayF2FConfig) *AlipayService {
	return &AlipayService{config: cfg}
}

// CreatePreOrder 创建支付宝预下单（当面付）
// 返回二维码内容（用于生成二维码供用户扫描）
// 参数：
//   - orderNo: 商户订单号
//   - amount: 订单金额（元）
//   - subject: 订单标题
//
// 返回：
//   - 二维码内容字符串
//   - 错误信息
func (s *AlipayService) CreatePreOrder(orderNo string, amount float64, subject string) (string, error) {
	if !s.config.Enabled {
		return "", errors.New("支付宝当面付未启用")
	}

	if s.config.AppID == "" || s.config.PrivateKey == "" {
		return "", errors.New("支付宝配置不完整")
	}

	// 构建业务参数
	bizContent := fmt.Sprintf(`{"out_trade_no":"%s","total_amount":"%.2f","subject":"%s"}`,
		orderNo, amount, subject)

	// 构建请求参数
	params := map[string]string{
		"app_id":      s.config.AppID,
		"method":      "alipay.trade.precreate",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  s.config.NotifyURL,
		"biz_content": bizContent,
	}

	// 签名
	sign, err := s.sign(params)
	if err != nil {
		return "", fmt.Errorf("签名失败: %v", err)
	}
	params["sign"] = sign

	// 注意：这里只是返回构建好的参数，实际调用需要使用HTTP客户端
	// 由于没有实际的支付宝SDK，这里返回一个占位符
	// 实际使用时需要：
	// 1. 发送POST请求到 https://openapi.alipay.com/gateway.do
	// 2. 解析响应获取 qr_code
	return fmt.Sprintf("alipay://pay?orderNo=%s&amount=%.2f", orderNo, amount), nil
}

// VerifyNotify 验证支付宝异步通知
// 参数：
//   - params: 通知参数
//
// 返回：
//   - 商户订单号
//   - 支付宝交易号
//   - 错误信息
func (s *AlipayService) VerifyNotify(params url.Values) (string, string, error) {
	if !s.config.Enabled {
		return "", "", errors.New("支付宝当面付未启用")
	}

	// 获取签名
	sign := params.Get("sign")
	if sign == "" {
		return "", "", errors.New("签名为空")
	}

	// 获取签名类型
	signType := params.Get("sign_type")
	if signType != "RSA2" {
		return "", "", errors.New("不支持的签名类型")
	}

	// 验证签名
	if err := s.verifySign(params, sign); err != nil {
		return "", "", fmt.Errorf("签名验证失败: %v", err)
	}

	// 验证交易状态
	tradeStatus := params.Get("trade_status")
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		return "", "", errors.New("交易未成功")
	}

	orderNo := params.Get("out_trade_no")
	tradeNo := params.Get("trade_no")

	return orderNo, tradeNo, nil
}

// QueryOrder 查询支付宝订单状态
// 参数：
//   - orderNo: 商户订单号
//
// 返回：
//   - 是否已支付
//   - 支付宝交易号
//   - 错误信息
func (s *AlipayService) QueryOrder(orderNo string) (bool, string, error) {
	if !s.config.Enabled {
		return false, "", errors.New("支付宝当面付未启用")
	}

	// 注意：这里需要实际调用支付宝查询接口
	// 由于没有实际的SDK，返回false
	return false, "", nil
}

// sign 生成签名
func (s *AlipayService) sign(params map[string]string) (string, error) {
	// 获取排序后的参数字符串
	signStr := s.getSignString(params)

	// 解析私钥
	block, _ := pem.Decode([]byte(s.config.PrivateKey))
	if block == nil {
		// 尝试直接解析（不带PEM头尾）
		keyBytes, err := base64.StdEncoding.DecodeString(s.config.PrivateKey)
		if err != nil {
			return "", errors.New("私钥格式错误")
		}
		block = &pem.Block{Bytes: keyBytes}
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS1格式
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("解析私钥失败: %v", err)
		}
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("私钥类型错误")
	}

	// SHA256签名
	hashed := sha256.Sum256([]byte(signStr))
	signature, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("签名失败: %v", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// verifySign 验证签名
func (s *AlipayService) verifySign(params url.Values, sign string) error {
	if s.config.PublicKey == "" {
		return errors.New("未配置支付宝公钥")
	}

	// 构建待验签字符串
	signParams := make(map[string]string)
	for key := range params {
		if key != "sign" && key != "sign_type" {
			signParams[key] = params.Get(key)
		}
	}
	signStr := s.getSignString(signParams)

	// 解析公钥
	block, _ := pem.Decode([]byte(s.config.PublicKey))
	if block == nil {
		keyBytes, err := base64.StdEncoding.DecodeString(s.config.PublicKey)
		if err != nil {
			return errors.New("公钥格式错误")
		}
		block = &pem.Block{Bytes: keyBytes}
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析公钥失败: %v", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("公钥类型错误")
	}

	// 解码签名
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return errors.New("签名解码失败")
	}

	// 验证签名
	hashed := sha256.Sum256([]byte(signStr))
	return rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hashed[:], signBytes)
}

// getSignString 获取待签名字符串
func (s *AlipayService) getSignString(params map[string]string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		if params[k] != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
		}
	}

	return strings.Join(parts, "&")
}
