# KamiServer

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/React-19.2.3-61DAFB?style=flat-square&logo=react" alt="React">
  <img src="https://img.shields.io/badge/Next.js-16.1.1-000000?style=flat-square&logo=next.js" alt="Next.js">
  <img src="https://img.shields.io/badge/Redis-7.0+-DC382D?style=flat-square&logo=redis" alt="Redis">
  <img src="https://img.shields.io/badge/License-GPL%20v3-blue?style=flat-square" alt="License">
</p>

<p align="center">
  一个功能完整的卡密销售管理系统，支持多种支付方式、Redis 缓存、在线客服、多管理员权限管理。
</p>

<p align="center">
  <a href="#功能特性">功能特性</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#技术栈">技术栈</a> •
  <a href="#配置说明">配置说明</a> •
  <a href="#项目结构">项目结构</a> •
  <a href="#文档">文档</a> •
  <a href="#许可证">许可证</a>
</p>

---

## 功能特性

### 🛒 商品系统
- 商品分类管理与多级分类
- 商品多图展示与轮播
- 手动卡密导入与批量管理
- 商品收藏与浏览历史
- 商品库存预警
- 商品评价与评分

### 💳 支付集成
- **支付宝** - 当面付/网页支付
- **微信支付** - Native/JSAPI
- **PayPal** - 国际支付
- **Stripe** - 信用卡支付 + Webhook
- **USDT** - 加密货币支付（TRC20/ERC20/BEP20）
- **易支付** - 聚合支付
- **余额支付** - 账户余额

### 👤 用户系统
- 用户注册/登录（多种方式）
- 邮箱验证与找回密码
- 两步验证（TOTP + 邮箱验证码）
- 设备管理与异地登录提醒
- 用户余额系统（充值/提现）
- 积分系统与积分商城
- 购物车与收藏夹

### 🎫 订单系统
- 订单创建与全生命周期管理
- 多类型优惠券（满减/折扣/固定金额）
- 订单超时自动取消与库存回滚
- 公开订单查询（无需登录）
- 发票管理与电子发票
- 订单导出（Excel/CSV）

### 💬 客服系统
- 工单系统（多状态流转）
- WebSocket 实时聊天
- 客服工作台与排队管理
- 游客支持（无需注册）
- 智能自动回复（关键词匹配）
- 知识库管理
- 满意度评价与数据统计

### 🔧 管理后台
- 仪表盘数据统计（实时/日/周/月）
- 多管理员角色权限（RBAC）
- 首页自定义配置（轮播图/公告/导航）
- 操作日志审计与追溯
- 数据库备份与恢复
- 系统监控与健康检查

### 🚀 缓存系统
- **Redis 支持** - 单机/哨兵/集群模式
- **自动故障转移** - Redis 不可用时自动降级到本地缓存
- **缓存仪表盘** - 命中率、内存使用、键管理
- **键管理** - 搜索、查看、删除缓存键
- **统计监控** - 实时监控缓存性能

### 🔒 安全特性
- 登录失败锁定（渐进式）
- IP 黑名单与白名单
- 分级速率限制（API/登录/注册）
- CSRF 保护
- 安全响应头（HSTS/XSS/CSP）
- 密码 bcrypt 加密
- 敏感数据 AES-GCM 加密
- 支付密码与二次验证

## 快速开始

### 环境要求

- Go 1.25+
- Node.js 18+（前端构建）

### 构建

```bash
# Windows
.\build.ps1

# Linux
./build.sh

# 嵌入模式（单文件部署）
.\build.ps1 --embed
```

### 运行

```bash
# Windows
.\dist\windows\UserFrontend.exe

# Linux
./dist/linux/UserFrontend
```

### 访问地址

| 页面 | 地址 |
|------|------|
| 用户前台 | http://localhost:8080/ |
| 管理后台 | http://localhost:8080/manage |
| 客服工作台 | http://localhost:8080/staff/login |

### 默认账户

首次访问管理后台时，系统会引导您设置管理员密码。

- 默认用户名：`admin`
- 密码：首次启动时设置

> ⚠️ **安全提示**：请设置强密码，建议包含字母、数字和特殊字符！

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端框架 | Go 1.25 + Gin 1.11 |
| ORM | GORM 1.31 |
| 数据库 | MySQL / PostgreSQL / SQLite |
| 缓存 | Redis 7.0+ (go-redis/v9) + 本地内存缓存 |
| 前端框架 | React 19 + Next.js 16 + TypeScript 5 |
| 样式 | Tailwind CSS 3.4 |
| 状态管理 | Zustand 5 |
| 实时通信 | WebSocket (gorilla/websocket) |
| 认证 | Session + Cookie + TOTP |
| 加密 | bcrypt + AES-GCM |

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `PORT` | 服务端口 | `8080` |
| `DB_TYPE` | 数据库类型 | `sqlite` |
| `DB_HOST` | 数据库地址 | `localhost` |
| `DB_PORT` | 数据库端口 | `3306` |
| `DB_USER` | 数据库用户名 | `root` |
| `DB_PASSWORD` | 数据库密码 | - |
| `DB_NAME` | 数据库名称 | `kamiserver` |
| `REDIS_ENABLED` | 启用 Redis | `false` |
| `REDIS_ADDRESS` | Redis 地址 | `localhost:6379` |
| `REDIS_PASSWORD` | Redis 密码 | - |
| `REDIS_DB` | Redis 数据库 | `0` |

### Redis 配置

系统支持三种 Redis 模式：

**单机模式**
```yaml
redis:
  enabled: true
  mode: standalone
  address: localhost:6379
  password: ""
  database: 0
```

**哨兵模式**
```yaml
redis:
  enabled: true
  mode: sentinel
  master_name: mymaster
  sentinel_addrs:
    - 192.168.1.1:26379
    - 192.168.1.2:26379
```

**集群模式**
```yaml
redis:
  enabled: true
  mode: cluster
  address: 192.168.1.1:6379,192.168.1.2:6379
```

> 💡 **提示**：Redis 不是必需的。未启用 Redis 时，系统会自动使用本地内存缓存。

## 项目结构

```
Server/
├── cmd/server/main.go       # 程序入口
├── internal/
│   ├── api/                 # HTTP API 处理层
│   │   ├── router.go        # 路由注册
│   │   ├── middleware.go    # 中间件（安全/限流）
│   │   ├── user_*.go        # 用户相关 API
│   │   ├── admin_*.go       # 管理后台 API
│   │   ├── order_*.go       # 订单相关 API
│   │   ├── payment_*.go     # 支付相关 API
│   │   ├── support_*.go     # 客服相关 API
│   │   └── redis_handler.go # Redis 缓存管理 API
│   ├── cache/               # 缓存层
│   │   ├── cache_manager.go # 缓存管理器
│   │   ├── redis_cache.go   # Redis 缓存实现
│   │   ├── local_cache.go   # 本地内存缓存实现
│   │   └── cache_metrics.go # 缓存统计指标
│   ├── config/              # 配置定义
│   ├── model/               # 数据模型
│   ├── repository/          # 数据访问层
│   ├── service/             # 业务逻辑层
│   │   ├── user_service.go  # 用户服务
│   │   ├── order_service.go # 订单服务
│   │   ├── product_service.go # 商品服务
│   │   ├── balance_service.go # 余额服务
│   │   ├── support_service.go # 客服服务
│   │   └── ...              # 其他服务
│   └── utils/               # 工具函数
│       ├── crypto.go        # 加密工具
│       ├── logger.go        # 日志系统
│       └── order.go         # 订单号生成
├── web/                     # 前端源码 (React + Next.js)
│   ├── src/
│   │   ├── app/             # 页面路由
│   │   │   ├── page.tsx     # 用户首页
│   │   │   ├── manage/      # 管理后台页面
│   │   │   └── staff/       # 客服工作台页面
│   │   ├── components/      # React 组件
│   │   │   ├── ui/          # 基础 UI 组件
│   │   │   ├── admin/       # 管理后台组件
│   │   │   └── user/        # 用户端组件
│   │   ├── hooks/           # 自定义 Hooks
│   │   └── lib/             # 工具库、API 封装
│   └── package.json
├── Product/                 # 商品图片存储
├── user_config/             # 运行时配置
├── server_log/              # 操作日志存储
├── backups/                 # 数据库备份
├── build.ps1                # Windows 构建脚本
└── build.sh                 # Linux 构建脚本
```

## 支付配置

| 支付方式 | 说明 | 配置项 |
|----------|------|--------|
| 支付宝 | 当面付/网页支付 | 应用ID、应用公私钥、支付宝公钥 |
| 微信支付 | Native/JSAPI | 商户号、API密钥、证书文件 |
| PayPal | 国际支付 | Client ID、Client Secret、环境模式 |
| Stripe | 信用卡支付 | Secret Key、Publishable Key、Webhook密钥 |
| USDT | 加密货币 | 钱包地址、网络类型（TRC20/ERC20/BEP20） |
| 易支付 | 聚合支付 | 商户ID、商户密钥、网关地址 |

### 支付配置示例

管理后台 → 系统设置 → 支付配置中进行配置。

```yaml
# 支付宝配置示例
alipay:
  app_id: "2021000000000000"
  private_key: "MIIEvQIBADANBg..."
  alipay_public_key: "MIIBIjANBg..."
  notify_url: "https://example.com/api/callback/alipay"
```

## API 接口

系统提供完整的 RESTful API：

| 模块 | 路径前缀 | 说明 |
|------|----------|------|
| 用户认证 | `/api/user/auth` | 注册、登录、登出 |
| 用户资料 | `/api/user/profile` | 个人信息管理 |
| 商品 | `/api/products` | 商品列表、详情 |
| 订单 | `/api/orders` | 订单管理 |
| 支付 | `/api/payment` | 支付创建、回调 |
| 客服 | `/api/support` | 工单、聊天 |
| 管理后台 | `/api/admin/*` | 后台管理 API |
| 缓存管理 | `/api/admin/redis/*` | Redis 仪表盘、键管理 |

详细 API 文档请参阅 [技术文档](Technical_Documentation_CN.md)。

## 文档

- 📖 [中文技术文档](Technical_Documentation_CN.md) - 详细的技术实现文档
- 📖 [English Documentation](Technical_Documentation_EN.md) - Technical documentation in English
- 📖 [用户使用说明书](用户端程序使用说明书.md) - 面向用户的操作指南
- 📖 [Redis 集成方案](Redis集成实施方案.md) - Redis 缓存系统集成指南

## 部署方式

### 方式一：二进制部署

```bash
# 1. 构建
.\build.ps1 --embed

# 2. 运行
.\dist\windows\UserFrontend.exe
```

### 方式二：Docker 部署

```dockerfile
# Dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/web/out ./web/out
EXPOSE 8080
CMD ["./server"]
```

```bash
# 构建并运行
docker build -t kamiserver .
docker run -d -p 8080:8080 \
  -e REDIS_ENABLED=true \
  -e REDIS_ADDRESS=redis:6379 \
  kamiserver
```

### 方式三：Docker Compose

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_TYPE=mysql
      - DB_HOST=db
      - DB_NAME=kamiserver
      - REDIS_ENABLED=true
      - REDIS_ADDRESS=redis:6379
    depends_on:
      - db
      - redis
  
  db:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: kamiserver
    volumes:
      - mysql_data:/var/lib/mysql
  
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  mysql_data:
  redis_data:
```

## 截图

<details>
<summary>点击展开截图</summary>

### 用户前台
- 首页商品展示
- 商品详情页
- 用户中心
- 订单管理

### 管理后台
- 仪表盘
- 商品管理
- 订单管理
- 系统设置

</details>

## 常见问题

<details>
<summary>如何重置管理员密码？</summary>

删除 `user_config/db-config.db` 文件，重启程序后会重新进入初始化向导。

</details>

<details>
<summary>Redis 连接失败会影响系统运行吗？</summary>

不会。系统具备自动故障转移能力，Redis 不可用时会自动切换到本地内存缓存，不影响正常使用。

</details>

<details>
<summary>如何备份数据？</summary>

1. **自动备份**：管理后台 → 系统设置 → 数据备份，可配置定时自动备份
2. **手动备份**：点击"立即备份"按钮
3. **数据库文件**：SQLite 模式下，直接复制数据库文件

</details>

<details>
<summary>支持哪些数据库？</summary>

- **SQLite**（默认）：适合小型部署，无需额外配置
- **MySQL 5.7+**：推荐用于生产环境
- **PostgreSQL 12+**：高级功能支持

</details>

## 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 许可证

本项目采用 [GNU General Public License v3.0](LICENSE) 开源许可证。

```
KamiServer - 卡密销售管理系统
Copyright (C) 2025

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```

## 致谢

- [Gin](https://github.com/gin-gonic/gin) - 高性能 HTTP Web 框架
- [GORM](https://gorm.io/) - Go 语言 ORM 库
- [go-redis](https://github.com/redis/go-redis) - Redis Go 客户端
- [Next.js](https://nextjs.org/) - React 全栈框架
- [Tailwind CSS](https://tailwindcss.com/) - 实用优先的 CSS 框架
- [Zustand](https://github.com/pmndrs/zustand) - 轻量级状态管理

---

<p align="center">
  如果这个项目对你有帮助，请给一个 ⭐ Star！
</p>
<p align="center">
  Made with ❤️ by CareCheng with 海绵
</p>
