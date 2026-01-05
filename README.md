# KamiServer

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/React-19.2.3-61DAFB?style=flat-square&logo=react" alt="React">
  <img src="https://img.shields.io/badge/Next.js-16.1.1-000000?style=flat-square&logo=next.js" alt="Next.js">
  <img src="https://img.shields.io/badge/License-GPL%20v3-blue?style=flat-square" alt="License">
</p>

<p align="center">
  一个功能完整的卡密销售管理系统，支持多种支付方式、在线客服、多管理员权限管理。
</p>

<p align="center">
  <a href="#功能特性">功能特性</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#技术栈">技术栈</a> •
  <a href="#项目结构">项目结构</a> •
  <a href="#文档">文档</a> •
  <a href="#许可证">许可证</a>
</p>

---

## 功能特性

### 🛒 商品系统
- 商品分类管理
- 商品多图展示
- 手动卡密导入与管理
- 商品收藏功能

### 💳 支付集成
- 支付宝（当面付/网页支付）
- 微信支付（Native/JSAPI）
- PayPal 国际支付
- Stripe 信用卡支付
- USDT 加密货币支付
- 易支付聚合支付
- 余额支付

### 👤 用户系统
- 用户注册/登录
- 邮箱验证
- 两步验证（TOTP + 邮箱）
- 设备管理与登录提醒
- 用户余额与积分系统
- 购物车功能

### 🎫 订单系统
- 订单创建与管理
- 优惠券系统
- 订单超时自动取消
- 公开订单查询
- 发票管理

### 💬 客服系统
- 工单系统
- WebSocket 实时聊天
- 客服工作台
- 游客支持
- 智能自动回复
- 知识库管理
- 满意度评价

### 🔧 管理后台
- 仪表盘数据统计
- 多管理员角色权限
- 首页自定义配置
- 操作日志审计
- 数据库备份
- 系统监控

### 🔒 安全特性
- 登录失败锁定
- IP 黑名单
- 分级速率限制
- CSRF 保护
- 安全响应头
- 密码 bcrypt 加密
- 敏感数据 AES-GCM 加密

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
| 前端框架 | React 19 + Next.js 16 + TypeScript 5 |
| 样式 | Tailwind CSS 3.4 |
| 状态管理 | Zustand 5 |
| 实时通信 | WebSocket |
| 认证 | Session + Cookie + TOTP |
| 加密 | bcrypt + AES-GCM |

## 项目结构

```
Server/
├── cmd/server/main.go       # 程序入口
├── internal/
│   ├── api/                 # HTTP API 处理层
│   ├── config/              # 配置定义
│   ├── model/               # 数据模型
│   ├── repository/          # 数据访问层
│   ├── service/             # 业务逻辑层
│   └── utils/               # 工具函数
├── web/                     # 前端源码 (React + Next.js)
│   ├── src/
│   │   ├── app/             # 页面路由
│   │   ├── components/      # React 组件
│   │   ├── hooks/           # 自定义 Hooks
│   │   └── lib/             # 工具库
│   └── package.json
├── Product/                 # 商品图片存储
├── user_config/             # 运行时配置
├── build.ps1                # Windows 构建脚本
└── build.sh                 # Linux 构建脚本
```

## 支付配置

| 支付方式 | 说明 |
|----------|------|
| 支付宝 | 当面付/网页支付，需配置应用公私钥 |
| 微信支付 | Native/JSAPI，需配置商户号和证书 |
| PayPal | 国际支付，支持沙盒和生产环境 |
| Stripe | 信用卡支付，支持 Webhook |
| USDT | 支持 TRC20/ERC20/BEP20 网络 |
| 易支付 | 第三方聚合支付 |

## 文档

- 📖 [中文技术文档](Technical_Documentation_CN.md)
- 📖 [English Documentation](Technical_Documentation_EN.md)
- 📖 [用户使用说明书](用户端程序使用说明书.md)

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

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [GORM](https://gorm.io/) - ORM 库
- [Next.js](https://nextjs.org/) - React 框架
- [Tailwind CSS](https://tailwindcss.com/) - CSS 框架

---

<p align="center">
  如果这个项目对你有帮助，请给一个 ⭐ Star！
</p>
