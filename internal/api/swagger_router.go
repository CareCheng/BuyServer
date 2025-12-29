package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ==================== Swagger 配置和路由 ====================
// 本文件包含 Swagger UI 的配置和路由注册
// 使用 swaggo/swag 生成 API 文档
//
// 安装步骤:
// 1. 安装 swag 命令行工具:
//    go install github.com/swaggo/swag/cmd/swag@latest
//
// 2. 在项目根目录生成文档:
//    swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
//
// 3. 安装 gin-swagger:
//    go get -u github.com/swaggo/gin-swagger
//    go get -u github.com/swaggo/files
//
// 4. 在 handler.go 中导入生成的 docs 包:
//    _ "user-frontend/docs"
//
// 5. 调用 RegisterSwaggerRoutes 注册 Swagger 路由

// SwaggerInfo Swagger 基本信息结构
// 用于动态配置 Swagger 文档信息
type SwaggerInfo struct {
	Title       string
	Description string
	Version     string
	Host        string
	BasePath    string
}

// DefaultSwaggerInfo 默认 Swagger 配置
var DefaultSwaggerInfo = SwaggerInfo{
	Title:       "KamiServer 用户端 API",
	Description: "卡密购买系统用户端 API 接口文档",
	Version:     "1.0",
	Host:        "localhost:8080",
	BasePath:    "/api",
}

// RegisterSwaggerRoutes 注册 Swagger 路由
// 需要在生成 docs 后启用此功能
// 参数:
//   - r: Gin 路由器实例
//
// 使用方法:
// 在 handler.go 的 RegisterRoutes 函数中添加:
//
//	api.RegisterSwaggerRoutes(r)
func RegisterSwaggerRoutes(r *gin.Engine) {
	// 注意: 实际使用时需要导入以下包并取消注释
	// import (
	//     swaggerFiles "github.com/swaggo/files"
	//     ginSwagger "github.com/swaggo/gin-swagger"
	//     _ "user-frontend/docs"  // 导入生成的 docs 包
	// )
	//
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 当前提供一个简单的 API 文档页面作为占位
	r.GET("/api/docs", ServeAPIDocs)
	r.GET("/api/docs/json", ServeAPIDocsJSON)
}

// ServeAPIDocs 提供简单的 API 文档 HTML 页面
// 这是一个临时的文档页面，完整功能需要使用 swaggo 生成
func ServeAPIDocs(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>KamiServer 用户端 API 文档</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; color: #333; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        h1 { color: #1976d2; margin-bottom: 10px; }
        .version { color: #666; margin-bottom: 20px; }
        .section { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h2 { color: #333; border-bottom: 2px solid #1976d2; padding-bottom: 10px; margin-bottom: 15px; }
        h3 { color: #1976d2; margin: 15px 0 10px 0; }
        .endpoint { background: #f8f9fa; border-radius: 4px; padding: 15px; margin: 10px 0; border-left: 4px solid #1976d2; }
        .method { display: inline-block; padding: 2px 8px; border-radius: 4px; font-weight: bold; font-size: 12px; margin-right: 10px; }
        .get { background: #4caf50; color: white; }
        .post { background: #2196f3; color: white; }
        .put { background: #ff9800; color: white; }
        .delete { background: #f44336; color: white; }
        .path { font-family: monospace; color: #666; }
        .desc { margin-top: 8px; color: #555; }
        .auth { font-size: 12px; color: #ff9800; margin-top: 5px; }
        .note { background: #fff3e0; border-radius: 4px; padding: 15px; margin: 20px 0; border-left: 4px solid #ff9800; }
    </style>
</head>
<body>
    <div class="container">
        <h1>KamiServer 用户端 API 文档</h1>
        <p class="version">版本: 1.0 | 基础路径: /api</p>

        <div class="note">
            <strong>注意:</strong> 这是简化版 API 文档。完整的交互式文档请安装 swaggo 并生成 Swagger UI。<br>
            运行命令: <code>swag init -g cmd/server/main.go -o docs</code>
        </div>

        <div class="section">
            <h2>用户认证</h2>
            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/api/user/register</span>
                <div class="desc">用户注册，需要邮箱验证码</div>
            </div>
            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/api/user/login</span>
                <div class="desc">用户登录，如启用2FA需要额外验证</div>
            </div>
            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/api/user/logout</span>
                <div class="desc">用户登出</div>
                <div class="auth">🔐 需要登录</div>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/user/info</span>
                <div class="desc">获取当前用户信息</div>
                <div class="auth">🔐 需要登录</div>
            </div>
        </div>

        <div class="section">
            <h2>商品</h2>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/products</span>
                <div class="desc">获取商品列表，支持分页和筛选</div>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/products/:id</span>
                <div class="desc">获取商品详情</div>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/products/categories</span>
                <div class="desc">获取商品分类列表</div>
            </div>
        </div>

        <div class="section">
            <h2>订单</h2>
            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/api/orders</span>
                <div class="desc">创建订单</div>
                <div class="auth">🔐 需要登录</div>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/orders</span>
                <div class="desc">获取用户订单列表</div>
                <div class="auth">🔐 需要登录</div>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/orders/:order_no</span>
                <div class="desc">获取订单详情</div>
                <div class="auth">🔐 需要登录</div>
            </div>
        </div>

        <div class="section">
            <h2>支付</h2>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/payment/methods</span>
                <div class="desc">获取可用支付方式</div>
            </div>
            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/api/payment/paypal/create</span>
                <div class="desc">创建 PayPal 支付</div>
                <div class="auth">🔐 需要登录</div>
            </div>
            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/api/payment/balance/pay</span>
                <div class="desc">使用余额支付</div>
                <div class="auth">🔐 需要登录</div>
            </div>
        </div>

        <div class="section">
            <h2>客服工单</h2>
            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/api/support/tickets</span>
                <div class="desc">创建客服工单</div>
                <div class="auth">🔐 需要登录</div>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/support/tickets</span>
                <div class="desc">获取用户工单列表</div>
                <div class="auth">🔐 需要登录</div>
            </div>
        </div>

        <div class="section">
            <h2>管理员 API</h2>
            <p style="color: #666; margin-bottom: 15px;">管理员 API 需要管理员身份认证</p>
            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/api/admin/login</span>
                <div class="desc">管理员登录</div>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/admin/dashboard</span>
                <div class="desc">获取仪表盘数据</div>
                <div class="auth">🔐 需要管理员权限</div>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/admin/products</span>
                <div class="desc">获取商品列表（管理）</div>
                <div class="auth">🔐 需要管理员权限</div>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/admin/orders</span>
                <div class="desc">获取订单列表（管理）</div>
                <div class="auth">🔐 需要管理员权限</div>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/admin/users</span>
                <div class="desc">获取用户列表</div>
                <div class="auth">🔐 需要管理员权限</div>
            </div>
        </div>

        <div class="section">
            <h2>错误码说明</h2>
            <table style="width: 100%; border-collapse: collapse;">
                <tr style="background: #f5f5f5;"><th style="padding: 10px; text-align: left; border-bottom: 1px solid #ddd;">错误码范围</th><th style="padding: 10px; text-align: left; border-bottom: 1px solid #ddd;">说明</th></tr>
                <tr><td style="padding: 10px; border-bottom: 1px solid #eee;">0</td><td style="padding: 10px; border-bottom: 1px solid #eee;">成功</td></tr>
                <tr><td style="padding: 10px; border-bottom: 1px solid #eee;">1000-1999</td><td style="padding: 10px; border-bottom: 1px solid #eee;">通用错误</td></tr>
                <tr><td style="padding: 10px; border-bottom: 1px solid #eee;">2000-2999</td><td style="padding: 10px; border-bottom: 1px solid #eee;">用户相关错误</td></tr>
                <tr><td style="padding: 10px; border-bottom: 1px solid #eee;">3000-3999</td><td style="padding: 10px; border-bottom: 1px solid #eee;">订单相关错误</td></tr>
                <tr><td style="padding: 10px; border-bottom: 1px solid #eee;">4000-4999</td><td style="padding: 10px; border-bottom: 1px solid #eee;">商品相关错误</td></tr>
                <tr><td style="padding: 10px; border-bottom: 1px solid #eee;">5000-5999</td><td style="padding: 10px; border-bottom: 1px solid #eee;">支付相关错误</td></tr>
                <tr><td style="padding: 10px; border-bottom: 1px solid #eee;">6000-6999</td><td style="padding: 10px; border-bottom: 1px solid #eee;">客服相关错误</td></tr>
                <tr><td style="padding: 10px; border-bottom: 1px solid #eee;">7000-7999</td><td style="padding: 10px; border-bottom: 1px solid #eee;">管理员相关错误</td></tr>
                <tr><td style="padding: 10px; border-bottom: 1px solid #eee;">8000-8999</td><td style="padding: 10px; border-bottom: 1px solid #eee;">配置相关错误</td></tr>
            </table>
        </div>
    </div>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// ServeAPIDocsJSON 提供 API 文档的 JSON 格式
// 返回简化的 OpenAPI 规范格式
func ServeAPIDocsJSON(c *gin.Context) {
	// 简化的 OpenAPI 3.0 规范
	openAPI := gin.H{
		"openapi": "3.0.0",
		"info": gin.H{
			"title":       DefaultSwaggerInfo.Title,
			"description": DefaultSwaggerInfo.Description,
			"version":     DefaultSwaggerInfo.Version,
		},
		"servers": []gin.H{
			{
				"url":         "http://localhost:8080",
				"description": "开发服务器",
			},
		},
		"tags": []gin.H{
			{"name": "用户认证", "description": "用户注册、登录、登出等"},
			{"name": "用户信息", "description": "用户信息管理"},
			{"name": "两步验证", "description": "2FA 相关操作"},
			{"name": "商品", "description": "商品浏览"},
			{"name": "订单", "description": "订单管理"},
			{"name": "支付", "description": "支付相关"},
			{"name": "优惠券", "description": "优惠券相关"},
			{"name": "余额", "description": "余额系统"},
			{"name": "积分", "description": "积分系统"},
			{"name": "购物车", "description": "购物车管理"},
			{"name": "收藏夹", "description": "收藏管理"},
			{"name": "客服工单", "description": "客服支持"},
			{"name": "FAQ", "description": "常见问题"},
			{"name": "通知", "description": "用户通知"},
			{"name": "管理员-认证", "description": "管理员认证"},
			{"name": "管理员-商品管理", "description": "商品管理"},
			{"name": "管理员-订单管理", "description": "订单管理"},
			{"name": "管理员-用户管理", "description": "用户管理"},
		},
		"paths": gin.H{
			"/api/user/register": gin.H{
				"post": gin.H{
					"tags":        []string{"用户认证"},
					"summary":     "用户注册",
					"description": "新用户注册，需要邮箱验证码",
				},
			},
			"/api/user/login": gin.H{
				"post": gin.H{
					"tags":        []string{"用户认证"},
					"summary":     "用户登录",
					"description": "用户登录，如启用2FA需要额外验证",
				},
			},
			"/api/products": gin.H{
				"get": gin.H{
					"tags":        []string{"商品"},
					"summary":     "获取商品列表",
					"description": "获取所有上架商品的列表，支持分页和筛选",
				},
			},
			"/api/orders": gin.H{
				"get": gin.H{
					"tags":        []string{"订单"},
					"summary":     "获取订单列表",
					"description": "获取当前用户的订单列表",
				},
				"post": gin.H{
					"tags":        []string{"订单"},
					"summary":     "创建订单",
					"description": "创建新的商品订单",
				},
			},
		},
		"components": gin.H{
			"securitySchemes": gin.H{
				"CookieAuth": gin.H{
					"type": "apiKey",
					"in":   "cookie",
					"name": "user_session",
				},
				"AdminCookieAuth": gin.H{
					"type": "apiKey",
					"in":   "cookie",
					"name": "admin_session",
				},
			},
		},
	}

	c.JSON(http.StatusOK, openAPI)
}

// GetSwaggerSetupInstructions 返回 Swagger 完整安装说明
func GetSwaggerSetupInstructions() string {
	return `
========================================
Swagger API 文档完整安装指南
========================================

1. 安装 swag 命令行工具:
   go install github.com/swaggo/swag/cmd/swag@latest

2. 安装 gin-swagger 依赖:
   go get -u github.com/swaggo/gin-swagger
   go get -u github.com/swaggo/files

3. 在项目根目录生成文档:
   cd User
   swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

4. 在 main.go 或 handler.go 中添加导入:
   import (
       swaggerFiles "github.com/swaggo/files"
       ginSwagger "github.com/swaggo/gin-swagger"
       _ "user-frontend/docs"
   )

5. 在路由注册中添加 Swagger 路由:
   r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

6. 访问文档:
   http://localhost:8080/swagger/index.html

========================================
`
}
