package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"user-frontend/internal/config"

	"github.com/gin-gonic/gin"
)

// PayPalTestConnection 测试PayPal连接
// POST /api/admin/paypal/test
func PayPalTestConnection(c *gin.Context) {
	cfg := config.GlobalConfig
	if cfg == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "配置未加载"})
		return
	}

	paypalCfg := cfg.PaymentConfig.PayPal
	if !paypalCfg.Enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "PayPal支付未启用"})
		return
	}

	if paypalCfg.ClientID == "" || paypalCfg.ClientSecret == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "PayPal Client ID 或 Client Secret 未配置"})
		return
	}

	// 测试获取访问令牌
	baseURL := "https://api-m.paypal.com"
	if paypalCfg.Sandbox {
		baseURL = "https://api-m.sandbox.paypal.com"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/v1/oauth2/token", bytes.NewBufferString("grant_type=client_credentials"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "创建请求失败: " + err.Error()})
		return
	}

	req.SetBasicAuth(paypalCfg.ClientID, paypalCfg.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "连接PayPal失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "PayPal认证失败: " + string(body)})
		return
	}

	// 解析响应
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "解析响应失败: " + err.Error()})
		return
	}

	if tokenResp.AccessToken == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "获取访问令牌失败"})
		return
	}

	mode := "正式环境"
	if paypalCfg.Sandbox {
		mode = "沙盒环境"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "PayPal连接测试成功（" + mode + "）",
	})
}

// AlipayTestConnection 测试支付宝连接
// POST /api/admin/alipay/test
func AlipayTestConnection(c *gin.Context) {
	cfg := config.GlobalConfig
	if cfg == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "配置未加载"})
		return
	}

	alipayCfg := cfg.PaymentConfig.AlipayF2F
	if !alipayCfg.Enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "支付宝当面付未启用"})
		return
	}

	if alipayCfg.AppID == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "支付宝 App ID 未配置"})
		return
	}

	if alipayCfg.PrivateKey == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "支付宝应用私钥未配置"})
		return
	}

	if alipayCfg.PublicKey == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "支付宝公钥未配置"})
		return
	}

	// 验证私钥格式（简单检查）
	if len(alipayCfg.PrivateKey) < 100 {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "支付宝应用私钥格式不正确"})
		return
	}

	// 验证公钥格式（简单检查）
	if len(alipayCfg.PublicKey) < 100 {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "支付宝公钥格式不正确"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "支付宝配置验证通过（App ID: " + alipayCfg.AppID + "）",
	})
}

// WechatPayTestConnection 测试微信支付连接
// POST /api/admin/wechat/test
func WechatPayTestConnection(c *gin.Context) {
	cfg := config.GlobalConfig
	if cfg == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "配置未加载"})
		return
	}

	wechatCfg := cfg.PaymentConfig.WechatPay
	if !wechatCfg.Enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "微信支付未启用"})
		return
	}

	if wechatCfg.AppID == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "微信支付 App ID 未配置"})
		return
	}

	if wechatCfg.MchID == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "微信支付商户号未配置"})
		return
	}

	if wechatCfg.APIKey == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "微信支付 API 密钥未配置"})
		return
	}

	// 验证API密钥长度（微信支付API密钥为32位）
	if len(wechatCfg.APIKey) != 32 {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "微信支付 API 密钥长度不正确（应为32位）"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "微信支付配置验证通过（商户号: " + wechatCfg.MchID + "）",
	})
}

// YiPayTestConnection 测试易支付连接
// POST /api/admin/yipay/test
func YiPayTestConnection(c *gin.Context) {
	cfg := config.GlobalConfig
	if cfg == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "配置未加载"})
		return
	}

	yipayCfg := cfg.PaymentConfig.YiPay
	if !yipayCfg.Enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "易支付未启用"})
		return
	}

	if yipayCfg.APIURL == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "易支付 API 地址未配置"})
		return
	}

	if yipayCfg.PID == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "易支付商户ID未配置"})
		return
	}

	if yipayCfg.Key == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "易支付商户密钥未配置"})
		return
	}

	// 测试API地址是否可访问
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(yipayCfg.APIURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "无法连接易支付服务器: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// 只要能连接上就算成功（易支付接口通常返回HTML页面）
	if resp.StatusCode >= 500 {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "易支付服务器错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "易支付连接测试成功（商户ID: " + yipayCfg.PID + "）",
	})
}
