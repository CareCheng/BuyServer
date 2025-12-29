package api

import (
	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminGetHomepageConfig 获取首页配置
func AdminGetHomepageConfig(c *gin.Context) {
	if HomepageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	config, err := HomepageSvc.GetActiveConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取配置失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"config":  config,
	})
}

// AdminUpdateHomepageConfig 更新首页配置
func AdminUpdateHomepageConfig(c *gin.Context) {
	if HomepageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var config model.HomepageFullConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误: " + err.Error()})
		return
	}

	if err := HomepageSvc.UpdateConfig(&config); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存配置失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "配置已保存",
	})
}

// AdminGetTemplateList 获取模板列表
func AdminGetTemplateList(c *gin.Context) {
	if HomepageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	templates := HomepageSvc.GetTemplateList()
	c.JSON(200, gin.H{
		"success":   true,
		"templates": templates,
	})
}

// AdminGetTemplateDefault 获取模板默认配置
func AdminGetTemplateDefault(c *gin.Context) {
	if HomepageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	template := c.Query("template")
	if template == "" {
		template = "modern"
	}

	config := HomepageSvc.GetDefaultConfigByTemplate(template)
	c.JSON(200, gin.H{
		"success": true,
		"config":  config,
	})
}

// AdminResetHomepage 重置首页为默认配置
func AdminResetHomepage(c *gin.Context) {
	if HomepageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Template string `json:"template"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Template = "modern"
	}

	if err := HomepageSvc.ResetToDefault(req.Template); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "重置失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "已重置为默认配置",
	})
}

// GetPublicHomepageConfig 获取公开的首页配置（用于前端渲染）
func GetPublicHomepageConfig(c *gin.Context) {
	if HomepageSvc == nil {
		// 返回默认配置
		config := model.GetDefaultConfig("modern")
		c.JSON(200, gin.H{
			"success": true,
			"config":  config,
		})
		return
	}

	config, err := HomepageSvc.GetActiveConfig()
	if err != nil {
		// 返回默认配置
		defaultConfig := model.GetDefaultConfig("modern")
		c.JSON(200, gin.H{
			"success": true,
			"config":  defaultConfig,
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"config":  config,
	})
}
