//go:build !embed
// +build !embed

// Package static 提供静态资源支持
// 此文件用于非嵌入模式（默认模式）
package static

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// IsEmbedded 检查是否使用嵌入式资源
// 非嵌入模式下始终返回 false
func IsEmbedded() bool {
	return false
}

// SetupStaticRoutes 设置静态文件路由
// 非嵌入模式：从文件系统加载
func SetupStaticRoutes(r *gin.Engine) {
	r.Static("/static", "./web")
	r.Static("/_next", "./web/_next")
	r.Static("/product-files", "./Product")
	r.Static("/uploads", "./uploads")
}

// ServeEmbeddedPage 从文件系统服务页面
func ServeEmbeddedPage(pagePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.File(filepath.Join("./web", pagePath))
	}
}

// GetWebFS 获取前端静态资源文件系统（非嵌入模式不可用）
func GetWebFS() (fs.FS, error) {
	return os.DirFS("./web"), nil
}

// ServeEmbeddedFS 返回嵌入式文件系统的 HTTP 处理器（非嵌入模式不可用）
func ServeEmbeddedFS() http.FileSystem {
	return http.Dir("./web")
}
