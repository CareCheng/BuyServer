//go:build embed
// +build embed

// Package static 提供静态资源支持
// 此文件用于嵌入模式（单文件模式）
package static

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:web
var embeddedWeb embed.FS

// IsEmbedded 检查是否使用嵌入式资源
// 嵌入模式下始终返回 true
func IsEmbedded() bool {
	return true
}

// SetupStaticRoutes 设置静态文件路由
// 嵌入模式：从嵌入的文件系统加载
func SetupStaticRoutes(r *gin.Engine) {
	// 获取嵌入的 web 子目录
	webFS, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		panic("无法获取嵌入的 web 目录: " + err.Error())
	}

	// 静态资源路由
	r.GET("/static/*filepath", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Param("filepath"), "/")
		serveEmbeddedFile(c, webFS, path)
	})

	r.GET("/_next/*filepath", func(c *gin.Context) {
		path := "_next" + c.Param("filepath")
		serveEmbeddedFile(c, webFS, path)
	})

	// 嵌入模式下，product-files 和 uploads 仍从外部加载
	r.Static("/product-files", "./Product")
	r.Static("/uploads", "./uploads")
}

// serveEmbeddedFile 从嵌入的文件系统服务文件
func serveEmbeddedFile(c *gin.Context, webFS fs.FS, path string) {
	// 尝试读取文件
	data, err := fs.ReadFile(webFS, path)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// 设置 Content-Type
	contentType := getContentType(path)
	c.Data(http.StatusOK, contentType, data)
}

// getContentType 根据文件扩展名获取 Content-Type
func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".map":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

// ServeEmbeddedPage 从嵌入的文件系统服务页面
func ServeEmbeddedPage(pagePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		webFS, err := fs.Sub(embeddedWeb, "web")
		if err != nil {
			c.String(http.StatusInternalServerError, "内部错误")
			return
		}

		data, err := fs.ReadFile(webFS, pagePath)
		if err != nil {
			// 尝试 index.html
			data, err = fs.ReadFile(webFS, "index.html")
			if err != nil {
				c.String(http.StatusNotFound, "页面未找到")
				return
			}
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	}
}

// GetWebFS 获取前端静态资源文件系统
func GetWebFS() (fs.FS, error) {
	return fs.Sub(embeddedWeb, "web")
}

// ServeEmbeddedFS 返回嵌入式文件系统的 HTTP 处理器
func ServeEmbeddedFS() http.FileSystem {
	webFS, _ := fs.Sub(embeddedWeb, "web")
	return http.FS(webFS)
}
