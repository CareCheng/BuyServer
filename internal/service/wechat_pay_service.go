package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"user-frontend/internal/config"
)

// WechatPayService 微信支付服务
type WechatPayService struct {
	config *config.WechatPayConfig
}

// NewWechatPayService 创建微信支付服务
func NewWechatPayService(cfg *config.WechatPayConfig) *WechatPayService {
	return &WechatPayService{config: cfg}
}

// WechatNotifyResult 微信支付通知结果
type WechatNotifyResult struct {
	XMLName       xml.Name `xml:"xml"`
	ReturnCode    string   `xml:"return_code"`
	ReturnMsg     string   `xml:"return_msg"`
	AppID         string   `xml:"appid"`
	MchID         string   `xml:"mch_id"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	ResultCode    string   `xml:"result_code"`
	OutTradeNo    string   `xml:"out_trade_no"`
	TransactionID string   `xml:"transaction_id"`
	TradeType     string   `xml:"trade_type"`
	TotalFee      int      `xml:"total_fee"`
}

// CreateNativeOrder 创建微信Native支付订单（扫码支付）
// 返回二维码内容（code_url）
// 参数：
//   - orderNo: 商户订单号
//   - amount: 订单金额（元）
//   - description: 商品描述
//
// 返回：
//   - 二维码内容字符串
//   - 错误信息
func (s *WechatPayService) CreateNativeOrder(orderNo string, amount float64, description string) (string, error) {
	if !s.config.Enabled {
		return "", errors.New("微信支付未启用")
	}

	if s.config.AppID == "" || s.config.MchID == "" || s.config.APIKey == "" {
		return "", errors.New("微信支付配置不完整")
	}

	// 金额转为分
	totalFee := int(amount * 100)

	// 构建请求参数
	params := map[string]string{
		"appid":            s.config.AppID,
		"mch_id":           s.config.MchID,
		"nonce_str":        generateNonceStr(),
		"body":             description,
		"out_trade_no":     orderNo,
		"total_fee":        fmt.Sprintf("%d", totalFee),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       s.config.NotifyURL,
		"trade_type":       "NATIVE",
	}

	// 签名
	params["sign"] = s.sign(params)

	// 注意：这里只是构建参数，实际需要发送到微信支付接口
	// 由于没有实际调用，返回占位符
	return fmt.Sprintf("weixin://wxpay/bizpayurl?orderNo=%s&amount=%.2f", orderNo, amount), nil
}

// VerifyNotify 验证微信支付异步通知
// 参数：
//   - request: HTTP请求
//
// 返回：
//   - 商户订单号
//   - 微信交易号
//   - 错误信息
func (s *WechatPayService) VerifyNotify(request *http.Request) (string, string, error) {
	if !s.config.Enabled {
		return "", "", errors.New("微信支付未启用")
	}

	// 读取请求体
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取请求体失败: %v", err)
	}

	// 解析XML
	var result WechatNotifyResult
	if err := xml.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("解析XML失败: %v", err)
	}

	// 验证返回码
	if result.ReturnCode != "SUCCESS" {
		return "", "", errors.New("微信返回失败: " + result.ReturnMsg)
	}

	if result.ResultCode != "SUCCESS" {
		return "", "", errors.New("交易失败")
	}

	// 验证签名
	params := map[string]string{
		"return_code":    result.ReturnCode,
		"return_msg":     result.ReturnMsg,
		"appid":          result.AppID,
		"mch_id":         result.MchID,
		"nonce_str":      result.NonceStr,
		"result_code":    result.ResultCode,
		"out_trade_no":   result.OutTradeNo,
		"transaction_id": result.TransactionID,
		"trade_type":     result.TradeType,
		"total_fee":      fmt.Sprintf("%d", result.TotalFee),
	}

	expectedSign := s.sign(params)
	if expectedSign != result.Sign {
		return "", "", errors.New("签名验证失败")
	}

	return result.OutTradeNo, result.TransactionID, nil
}

// QueryOrder 查询微信支付订单状态
// 参数：
//   - orderNo: 商户订单号
//
// 返回：
//   - 是否已支付
//   - 微信交易号
//   - 错误信息
func (s *WechatPayService) QueryOrder(orderNo string) (bool, string, error) {
	if !s.config.Enabled {
		return false, "", errors.New("微信支付未启用")
	}

	// 注意：这里需要实际调用微信查询接口
	// 由于没有实际的SDK，返回false
	return false, "", nil
}

// sign 生成签名
func (s *WechatPayService) sign(params map[string]string) string {
	// 获取排序后的参数字符串
	var keys []string
	for k := range params {
		if params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	// 拼接API密钥
	signStr := strings.Join(parts, "&") + "&key=" + s.config.APIKey

	// MD5签名
	hash := md5.Sum([]byte(signStr))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// generateNonceStr 生成随机字符串
func generateNonceStr() string {
	hash := md5.Sum([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:])[:32]
}
