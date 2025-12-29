// Package api 提供 HTTP API 处理器
// handler.go - 认证中间件和页面渲染
package api

import (
	"user-frontend/internal/config"

	"github.com/gin-gonic/gin"
)

// PageAuthRequired 页面认证中间件 - 会话过期时重定向到登录页
func PageAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("user_session")
		if err != nil {
			c.Redirect(302, "/login")
			c.Abort()
			return
		}

		if SessionSvc == nil {
			c.Redirect(302, "/login")
			c.Abort()
			return
		}

		session, err := SessionSvc.GetUserSession(sessionID)
		if err != nil {
			c.Redirect(302, "/login")
			c.Abort()
			return
		}

		c.Set("user_id", session.UserID)
		c.Set("username", session.Username)
		c.Next()
	}
}

// AuthRequired 用户认证中间件 - 用于API请求
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("user_session")
		if err != nil {
			c.JSON(401, gin.H{"success": false, "error": "请先登录"})
			c.Abort()
			return
		}

		if SessionSvc == nil {
			c.JSON(401, gin.H{"success": false, "error": "服务未初始化"})
			c.Abort()
			return
		}

		session, err := SessionSvc.GetUserSession(sessionID)
		if err != nil {
			c.JSON(401, gin.H{"success": false, "error": "登录已过期"})
			c.Abort()
			return
		}

		c.Set("user_id", session.UserID)
		c.Set("username", session.Username)
		c.Next()
	}
}

// OptionalAuth 可选认证中间件 - 支持游客和登录用户
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("user_session")
		if err == nil && SessionSvc != nil {
			session, err := SessionSvc.GetUserSession(sessionID)
			if err == nil {
				c.Set("user_id", session.UserID)
				c.Set("username", session.Username)
			}
		}
		c.Next()
	}
}

// AdminAuthRequired 管理员认证中间件
func AdminAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果禁用了登录验证，直接放行
		if !config.GlobalConfig.ServerConfig.EnableLogin {
			c.Set("admin_username", config.GlobalConfig.ServerConfig.AdminUsername)
			c.Set("admin_role", "super_admin")
			c.Next()
			return
		}

		sessionID, err := c.Cookie("admin_session")
		if err != nil {
			c.JSON(401, gin.H{"success": false, "error": "请先登录"})
			c.Abort()
			return
		}

		if SessionSvc == nil {
			c.JSON(401, gin.H{"success": false, "error": "服务未初始化"})
			c.Abort()
			return
		}

		session, err := SessionSvc.GetAdminSession(sessionID)
		if err != nil {
			c.JSON(401, gin.H{"success": false, "error": "登录已过期"})
			c.Abort()
			return
		}

		if !session.Verified {
			c.JSON(401, gin.H{"success": false, "error": "请完成两步验证"})
			c.Abort()
			return
		}

		c.Set("admin_username", session.Username)
		c.Set("admin_role", session.Role)
		c.Next()
	}
}

// AdminPortalAccess 管理后台入口
func AdminPortalAccess(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置访问权限Cookie
		c.SetCookie("admin_portal_access", "true", 3600, "/", "", false, true)

		// 如果禁用了登录验证，直接进入管理后台
		if !cfg.ServerConfig.EnableLogin {
			c.HTML(200, "admin_index.html", gin.H{
				"title": cfg.ServerConfig.SystemTitle + " - 管理后台",
			})
			return
		}

		// 检查会话
		sessionID, _ := c.Cookie("admin_session")
		if SessionSvc == nil {
			c.Redirect(302, "/"+cfg.ServerConfig.AdminSuffix+"/login")
			return
		}

		session, err := SessionSvc.GetAdminSession(sessionID)
		if err != nil {
			c.Redirect(302, "/"+cfg.ServerConfig.AdminSuffix+"/login")
			return
		}

		if !session.Verified {
			c.Redirect(302, "/"+cfg.ServerConfig.AdminSuffix+"/totp")
			return
		}

		c.HTML(200, "admin_index.html", gin.H{
			"title": cfg.ServerConfig.SystemTitle + " - 管理后台",
		})
	}
}
