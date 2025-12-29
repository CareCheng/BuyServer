package api

import (
	"image/color"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

var captchaStore = base64Captcha.DefaultMemStore

// CaptchaHandler 生成验证码
func CaptchaHandler(c *gin.Context) {
	// 使用高对比度的字母数字验证码，更清晰易读
	driver := &base64Captcha.DriverString{
		Height:          80,                                      // 高度
		Width:           240,                                     // 宽度
		NoiseCount:      30,                                      // 干扰点数量（减少以提高可读性）
		ShowLineOptions: base64Captcha.OptionShowHollowLine,      // 显示空心线条
		Length:          4,                                       // 验证码长度
		Source:          "23456789abcdefghjkmnpqrstuvwxyz",       // 排除易混淆字符(0,1,i,l,o)
		BgColor:         &color.RGBA{255, 255, 255, 255},         // 白色背景
		Fonts:           []string{"wqy-microhei.ttc"},            // 字体
	}
	captcha := base64Captcha.NewCaptcha(driver, captchaStore)

	id, b64s, _, err := captcha.Generate()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "生成验证码失败"})
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"captcha_id": id,
		"image":      b64s,
	})
}

// VerifyCaptcha 验证验证码
func VerifyCaptcha(c *gin.Context) {
	var req struct {
		CaptchaID   string `json:"captcha_id" binding:"required"`
		CaptchaCode string `json:"captcha_code" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}
	
	if captchaStore.Verify(req.CaptchaID, req.CaptchaCode, true) {
		c.JSON(200, gin.H{"success": true})
	} else {
		c.JSON(400, gin.H{"success": false, "error": "验证码错误"})
	}
}

// VerifyCaptchaCode 验证验证码（内部使用）
func VerifyCaptchaCode(id, code string) bool {
	return captchaStore.Verify(id, code, true)
}
