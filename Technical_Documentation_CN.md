# KamiServer 技术文档

## 1. 系统概述

KamiServer 是一个功能完整的卡密销售管理系统，基于 Go 语言开发，使用 Gin Web 框架。系统提供用户注册登录、商品浏览、订单管理、多种支付方式、在线客服等功能。

### 1.1 技术栈

| 组件 | 技术 |
|------|------|
| 后端框架 | Gin v1.9.1 |
| ORM | GORM v1.25.5 |
| 数据库 | MySQL / PostgreSQL / SQLite |
| 前端框架 | React + Next.js 14 + TypeScript |
| 样式 | Tailwind CSS |
| 状态管理 | Zustand |
| 实时通信 | WebSocket (gorilla/websocket) |
| 认证 | Session + Cookie |
| 两步验证 | TOTP (pquerna/otp) |
| 验证码 | base64Captcha |
| 加密 | bcrypt + AES-GCM |

### 1.2 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                      KamiServer 系统                         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │ 用户模块 │  │ 商品模块 │  │ 订单模块 │  │ 支付模块 │        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
│       │            │            │            │              │
│  ┌────┴────────────┴────────────┴────────────┴────┐        │
│  │                  Service 层                      │        │
│  │  UserSvc | OrderSvc | PaymentSvc | SupportSvc   │        │
│  └────────────────────────┬─────────────────────────┘        │
│                           │                                  │
│  ┌────────────────────────┴─────────────────────────┐        │
│  │                Repository 层                      │        │
│  └────────────────────────┬─────────────────────────┘        │
│                           │                                  │
│  ┌────────────────────────┴─────────────────────────┐        │
│  │           数据库 (MySQL/PostgreSQL/SQLite)        │        │
│  └──────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## 2. 目录结构

```
Server/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── api/
│   │   ├── doc.go               # API 层包文档
│   │   ├── router.go            # 路由注册
│   │   ├── middleware.go        # 安全中间件（CSRF、限流、安全头、黑名单）
│   │   ├── services.go          # 服务依赖注入
│   │   ├── response_helper.go   # 统一响应辅助函数
│   │   ├── error_codes.go       # 统一错误码定义
│   │   ├── user_auth_handler.go # 用户认证API
│   │   ├── user_profile_handler.go # 用户资料API
│   │   ├── admin_handler.go     # 管理后台API
│   │   ├── order_handler.go     # 订单相关API
│   │   ├── payment_handler.go   # 支付相关API
│   │   ├── support_handler.go   # 用户端客服API
│   │   ├── support_staff_handler.go # 客服后台API
│   │   ├── websocket_handler.go # WebSocket处理
│   │   └── ...                  # 其他API处理器
│   ├── config/
│   │   └── config.go            # 配置结构定义
│   ├── model/
│   │   ├── doc.go               # 数据模型层包文档
│   │   ├── models.go            # 核心数据模型
│   │   ├── db.go                # 数据库初始化
│   │   ├── balance.go           # 余额模型
│   │   ├── points.go            # 积分模型
│   │   ├── cart.go              # 购物车模型
│   │   ├── support.go           # 客服系统模型
│   │   ├── manual_kami.go       # 手动卡密模型
│   │   └── ...                  # 其他模型
│   ├── repository/
│   │   ├── doc.go               # 数据仓库层包文档
│   │   └── repository.go        # 数据访问层
│   ├── service/
│   │   ├── doc.go               # 服务层包文档
│   │   ├── user_service.go      # 用户业务逻辑
│   │   ├── order_service.go     # 订单业务逻辑
│   │   ├── product_service.go   # 商品业务逻辑
│   │   ├── balance_service.go   # 余额业务逻辑
│   │   ├── points_service.go    # 积分业务逻辑
│   │   ├── cart_service.go      # 购物车业务逻辑
│   │   ├── paypal_service.go    # PayPal支付服务
│   │   ├── stripe_service.go    # Stripe支付服务
│   │   ├── usdt_service.go      # USDT支付服务
│   │   ├── support_service.go   # 客服支持服务
│   │   ├── log_service.go       # 操作日志服务
│   │   └── ...                  # 其他服务
│   ├── static/
│   │   └── static.go            # 静态文件处理
│   ├── test/
│   │   └── test_helper.go       # 测试辅助函数
│   └── utils/
│       ├── doc.go               # 工具包文档
│       ├── crypto.go            # 密码加密
│       ├── order.go             # 订单号生成
│       ├── logger.go            # 统一日志系统
│       └── environment.go       # 环境配置管理
├── web/                         # 前端源码 (React + Next.js)
│   ├── src/
│   │   ├── app/                 # Next.js App Router 页面
│   │   ├── components/          # React 组件
│   │   ├── hooks/               # 自定义 Hooks
│   │   ├── lib/                 # 工具库、API 封装
│   │   ├── types/               # TypeScript 类型定义
│   │   └── contexts/            # React Context
│   └── package.json
├── Product/                     # 商品图片存储目录
├── user_config/                 # 配置目录
│   └── db-config.db             # SQLite配置数据库
├── server_log/                  # 操作日志目录
├── backups/                     # 数据库备份目录
├── go.mod
├── go.sum
├── build.ps1                    # Windows 构建脚本
└── build.sh                     # Linux 构建脚本
```

## 3. 数据模型

### 3.1 用户模型 (User)

```go
type User struct {
    ID              uint           // 主键
    Username        string         // 用户名（唯一）
    Email           string         // 邮箱（唯一）
    PasswordHash    string         // 密码哈希
    Phone           string         // 手机号
    EmailVerified   bool           // 邮箱是否已验证
    Enable2FA       bool           // 是否启用两步验证
    TOTPSecret      string         // TOTP密钥
    PreferEmailAuth bool           // 登录时优先使用邮箱验证
    Status          int            // 状态：1正常 0禁用
    LastLoginAt     *time.Time     // 最后登录时间
    LastLoginIP     string         // 最后登录IP
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### 3.2 商品模型 (Product)

```go
type Product struct {
    ID           uint           // 主键
    Name         string         // 商品名称
    Description  string         // 商品描述
    Price        float64        // 价格
    Duration     int            // 时长数值
    DurationUnit string         // 时长单位：天/周/月/年
    Stock        int            // 库存，-1表示无限
    Status       int            // 状态：1上架 0下架
    AllowTest    bool           // 是否允许测试购买
    SortOrder    int            // 排序
    ImageURL     string         // 商品图片
    CategoryID   uint           // 分类ID
    ProductType  int            // 商品类型：1手动卡密（默认）
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**商品类型说明**：
| 类型值 | 名称 | 说明 |
|--------|------|------|
| 1 | 手动卡密 | 默认模式，管理员手动导入卡密，订单完成时从本地卡密池分配 |

### 3.2.1 手动卡密模型 (ManualKami)

```go
type ManualKami struct {
    ID        uint           // 主键
    ProductID uint           // 关联商品ID
    KamiCode  string         // 卡密内容
    Status    int            // 状态：0可用 1已售出 2已禁用
    OrderID   uint           // 关联订单ID（售出后填充）
    OrderNo   string         // 关联订单号
    SoldAt    *time.Time     // 售出时间
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**卡密状态说明**：
| 状态值 | 名称 | 说明 |
|--------|------|------|
| 0 | 可用 | 卡密可被分配给新订单 |
| 1 | 已售出 | 卡密已分配给订单 |
| 2 | 已禁用 | 卡密被管理员禁用，不可分配 |

### 3.3 订单模型 (Order)

```go
type Order struct {
    ID            uint           // 主键
    OrderNo       string         // 订单号（唯一）
    PaymentNo     string         // 支付订单号
    UserID        uint           // 用户ID
    Username      string         // 用户名
    Email         string         // 用户邮箱（用于公开查询）
    ProductID     uint           // 商品ID
    ProductName   string         // 商品名称
    Price         float64        // 价格
    Duration      int            // 时长
    DurationUnit  string         // 时长单位
    Status        int            // 状态：0待支付 1已支付 2已完成 3已取消 4已退款 5已过期
    PaymentMethod string         // 支付方式
    PaymentTime   *time.Time     // 支付时间
    KamiCode      string         // 生成的卡密
    IsTest        bool           // 是否为测试订单
    Remark        string         // 备注
    ClientIP      string         // 客户端IP
    ExpireAt      *time.Time     // 订单过期时间
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### 3.4 登录尝试记录 (LoginAttempt)

```go
type LoginAttempt struct {
    ID           uint      // 主键
    Username     string    // 用户名
    IP           string    // IP地址
    Success      bool      // 是否成功
    FailedCount  int       // 连续失败次数
    LockedUntil  *time.Time // 锁定截止时间
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 3.6 操作日志 (OperationLog) - 文件存储

> **重要变更**：操作日志已从数据库存储改为加密文件存储，提高安全性和性能。

**存储位置**：`server_log/YYYY-MM-DD.csv`

**加密方式**：AES-256-GCM（使用数据库配置中的加密密钥）

**日志条目结构**：
```go
type LogEntry struct {
    ID        uint      // 虚拟ID（用于前端显示）
    UserType  string    // 用户类型：admin/user/security
    UserID    uint      // 用户ID
    Username  string    // 用户名
    Action    string    // 操作类型：login/logout/create/update/delete
    Target    string    // 操作目标：product/order/user/announcement/category
    TargetID  string    // 目标ID
    Detail    string    // 详细信息（JSON格式）
    IP        string    // IP地址
    UserAgent string    // 用户代理
    CreatedAt time.Time // 创建时间
}
```

**CSV文件格式**：
- 表头（明文）：`user_type,user_id,username,action,target,target_id,detail,ip,user_agent,created_at`
- 数据行：每个字段使用AES-256-GCM加密后的Base64字符串

**API接口**：
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/logs | 获取操作日志（支持date参数按日期查询） |
| GET | /api/admin/logs/dates | 获取可用的日志日期列表 |

### 3.7 公告 (Announcement)

```go
type Announcement struct {
    ID        uint       // 主键
    Title     string     // 标题
    Content   string     // 内容
    Type      string     // 类型：info/warning/success/danger
    Status    int        // 状态：1启用 0禁用
    SortOrder int        // 排序
    StartAt   *time.Time // 开始时间
    EndAt     *time.Time // 结束时间
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 3.8 商品分类 (ProductCategory)

```go
type ProductCategory struct {
    ID        uint      // 主键
    Name      string    // 分类名称
    Icon      string    // 图标
    SortOrder int       // 排序
    Status    int       // 状态：1启用 0禁用
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 3.9 用户会话 (UserSession)

```go
type UserSession struct {
    ID        uint      // 主键
    SessionID string    // 会话ID（唯一）
    UserID    uint      // 用户ID
    Username  string    // 用户名
    IP        string    // 登录IP
    UserAgent string    // 用户代理
    ExpiresAt time.Time // 过期时间
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 3.10 管理员会话 (AdminSession)

```go
type AdminSession struct {
    ID        uint      // 主键
    SessionID string    // 会话ID（唯一）
    Username  string    // 管理员用户名
    Role      string    // 角色
    IP        string    // 登录IP
    UserAgent string    // 用户代理
    Verified  bool      // 2FA验证状态
    ExpiresAt time.Time // 过期时间
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 3.11 登录失败记录 (LoginFailureRecord)

```go
type LoginFailureRecord struct {
    ID           uint       // 主键
    Key          string     // IP或用户名（唯一）
    FailureCount int        // 失败次数
    FirstFailAt  time.Time  // 首次失败时间
    LockedAt     *time.Time // 锁定时间
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

## 3.12 数据库表结构

### 3.12.1 数据库架构概览

系统采用双数据库架构：

```
┌─────────────────────────────────────────────────────────────────────┐
│                        数据库架构                                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────┐      ┌─────────────────────────────────┐  │
│  │  配置数据库 (SQLite)  │      │     主数据库 (MySQL/PG/SQLite)    │  │
│  │  user_config/config.db│      │                                 │  │
│  ├─────────────────────┤      ├─────────────────────────────────┤  │
│  │ • db_configs        │      │ 用户相关                         │  │
│  │   (数据库连接配置)    │      │ • users (用户)                   │  │
│  │                     │      │ • user_sessions (用户会话)        │  │
│  │                     │      │ • login_failure_records (登录失败) │  │
│  │                     │      │                                 │  │
│  │                     │      │ 商品订单                         │  │
│  │                     │      │ • products (商品)                │  │
│  │                     │      │ • product_categories (分类)      │  │
│  │                     │      │ • orders (订单)                  │  │
│  │                     │      │ • coupons (优惠券)               │  │
│  │                     │      │ • coupon_usages (优惠券使用)      │  │
│  │                     │      │                                 │  │
│  │                     │      │ 系统配置                         │  │
│  │                     │      │ • system_configs (系统配置)       │  │
│  │                     │      │ • email_configs (邮箱配置)        │  │
│  │                     │      │ • payment_configs (支付配置)      │  │
│  │                     │      │                                 │  │
│  │                     │      │ 管理相关                         │  │
│  │                     │      │ • admin_sessions (管理员会话)     │  │
│  │                     │      │ • operation_logs (操作日志)       │  │
│  │                     │      │ • announcements (公告)           │  │
│  │                     │      │ • backups (备份记录)             │  │
│  │                     │      │                                 │  │
│  │                     │      │ 客服系统                         │  │
│  │                     │      │ • support_tickets (工单)         │  │
│  │                     │      │ • support_messages (工单消息)     │  │
│  │                     │      │ • support_staff (客服人员)        │  │
│  │                     │      │ • support_staff_sessions (客服会话)│  │
│  │                     │      │ • support_configs (客服配置)      │  │
│  │                     │      │ • live_chats (实时聊天)          │  │
│  │                     │      │ • live_chat_messages (聊天消息)   │  │
│  └─────────────────────┘      └─────────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.12.2 主数据库表结构

#### users 用户表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| username | VARCHAR(100) | UNIQUE, NOT NULL | 用户名 |
| email | VARCHAR(255) | UNIQUE | 邮箱 |
| password_hash | VARCHAR(255) | NOT NULL | 密码哈希 |
| phone | VARCHAR(20) | | 手机号 |
| email_verified | BOOLEAN | DEFAULT FALSE | 邮箱是否验证 |
| enable_2fa | BOOLEAN | DEFAULT FALSE | 是否启用2FA |
| totp_secret | VARCHAR(64) | | TOTP密钥 |
| prefer_email_auth | BOOLEAN | DEFAULT FALSE | 优先邮箱验证 |
| status | INT | DEFAULT 1 | 状态：1正常 0禁用 |
| last_login_at | DATETIME | | 最后登录时间 |
| last_login_ip | VARCHAR(50) | | 最后登录IP |
| created_at | DATETIME | | 创建时间 |
| updated_at | DATETIME | | 更新时间 |

#### products 商品表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| name | VARCHAR(200) | NOT NULL | 商品名称 |
| description | TEXT | | 商品描述 |
| price | DECIMAL(10,2) | NOT NULL | 价格 |
| duration | INT | | 时长数值 |
| duration_unit | VARCHAR(20) | | 时长单位 |
| stock | INT | DEFAULT -1 | 库存，-1无限 |
| status | INT | DEFAULT 1 | 状态：1上架 0下架 |
| allow_test | BOOLEAN | DEFAULT FALSE | 允许测试购买 |
| sort_order | INT | DEFAULT 0 | 排序 |
| image_url | VARCHAR(500) | | 商品图片 |
| category_id | BIGINT | FOREIGN KEY | 分类ID |
| product_type | INT | DEFAULT 1 | 商品类型：1手动卡密 |
| created_at | DATETIME | | 创建时间 |
| updated_at | DATETIME | | 更新时间 |

#### orders 订单表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| order_no | VARCHAR(64) | UNIQUE, NOT NULL | 订单号 |
| payment_no | VARCHAR(64) | | 支付订单号 |
| user_id | BIGINT | INDEX | 用户ID |
| username | VARCHAR(100) | | 用户名 |
| email | VARCHAR(255) | | 用户邮箱 |
| product_id | BIGINT | | 商品ID |
| product_name | VARCHAR(200) | | 商品名称 |
| price | DECIMAL(10,2) | | 价格 |
| original_price | DECIMAL(10,2) | | 原价 |
| duration | INT | | 时长 |
| duration_unit | VARCHAR(20) | | 时长单位 |
| status | INT | DEFAULT 0 | 状态：0待支付 1已支付 2已完成 3已取消 4已退款 5已过期 |
| payment_method | VARCHAR(50) | | 支付方式 |
| payment_time | DATETIME | | 支付时间 |
| kami_code | TEXT | | 卡密（加密存储） |
| is_test | BOOLEAN | DEFAULT FALSE | 是否测试订单 |
| remark | TEXT | | 备注 |
| client_ip | VARCHAR(50) | | 客户端IP |
| coupon_id | BIGINT | | 优惠券ID |
| coupon_code | VARCHAR(50) | | 优惠券码 |
| discount_amount | DECIMAL(10,2) | | 优惠金额 |
| expire_at | DATETIME | | 订单过期时间 |
| created_at | DATETIME | | 创建时间 |
| updated_at | DATETIME | | 更新时间 |

#### support_tickets 工单表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| ticket_no | VARCHAR(32) | UNIQUE, NOT NULL | 工单编号 |
| user_id | BIGINT | INDEX | 用户ID（0表示游客） |
| username | VARCHAR(100) | | 用户名 |
| email | VARCHAR(255) | | 联系邮箱 |
| subject | VARCHAR(200) | NOT NULL | 工单主题 |
| category | VARCHAR(50) | | 分类 |
| priority | INT | DEFAULT 1 | 优先级：1普通 2紧急 3非常紧急 |
| status | INT | DEFAULT 0 | 状态：0待处理 1处理中 2已回复 3已解决 4已关闭 |
| assigned_to | BIGINT | | 分配给客服ID |
| assigned_name | VARCHAR(100) | | 客服名称 |
| related_order | VARCHAR(64) | | 关联订单号 |
| guest_token | VARCHAR(64) | INDEX | 游客访问令牌 |
| last_reply_at | DATETIME | | 最后回复时间 |
| last_reply_by | VARCHAR(100) | | 最后回复人 |
| closed_at | DATETIME | | 关闭时间 |
| closed_by | VARCHAR(100) | | 关闭人 |
| created_at | DATETIME | | 创建时间 |
| updated_at | DATETIME | | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

#### support_messages 工单消息表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| ticket_id | BIGINT | INDEX, NOT NULL | 关联工单ID |
| sender_type | VARCHAR(20) | | 发送者类型：user/guest/staff/system |
| sender_id | BIGINT | | 发送者ID |
| sender_name | VARCHAR(100) | | 发送者名称 |
| content | TEXT | | 消息内容 |
| is_internal | BOOLEAN | DEFAULT FALSE | 是否内部备注 |
| read_at | DATETIME | | 已读时间 |
| created_at | DATETIME | | 创建时间 |

#### support_staff 客服人员表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| username | VARCHAR(100) | UNIQUE, NOT NULL | 用户名 |
| password_hash | VARCHAR(255) | NOT NULL | 密码哈希 |
| nickname | VARCHAR(100) | | 显示名称 |
| avatar | VARCHAR(500) | | 头像URL |
| email | VARCHAR(255) | | 邮箱 |
| role | VARCHAR(50) | DEFAULT 'staff' | 角色：staff/supervisor |
| status | INT | DEFAULT 0 | 状态：1在线 0离线 -1禁用 |
| max_tickets | INT | DEFAULT 10 | 最大同时处理工单数 |
| current_load | INT | DEFAULT 0 | 当前处理工单数 |
| enable_2fa | BOOLEAN | DEFAULT FALSE | 是否启用2FA |
| totp_secret | VARCHAR(64) | | TOTP密钥 |
| last_active_at | DATETIME | | 最后活跃时间 |
| created_at | DATETIME | | 创建时间 |
| updated_at | DATETIME | | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

#### support_configs 客服配置表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| enabled | BOOLEAN | DEFAULT TRUE | 是否启用客服系统 |
| allow_guest | BOOLEAN | DEFAULT TRUE | 是否允许游客咨询 |
| staff_portal_suffix | VARCHAR(50) | DEFAULT 'staff' | 客服后台路径后缀 |
| enable_staff_2fa | BOOLEAN | DEFAULT FALSE | 客服是否启用2FA |
| working_hours_start | VARCHAR(10) | | 工作时间开始 |
| working_hours_end | VARCHAR(10) | | 工作时间结束 |
| working_days | VARCHAR(50) | | 工作日 |
| offline_message | TEXT | | 离线提示消息 |
| welcome_message | TEXT | | 欢迎消息 |
| auto_close_hours | INT | DEFAULT 72 | 自动关闭时间（小时） |
| ticket_categories | TEXT | | 工单分类（JSON数组） |
| created_at | DATETIME | | 创建时间 |
| updated_at | DATETIME | | 更新时间 |

#### live_chats 实时聊天表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| session_id | VARCHAR(64) | UNIQUE, NOT NULL | 聊天会话ID |
| user_id | BIGINT | INDEX | 用户ID（0表示游客） |
| username | VARCHAR(100) | | 用户名 |
| guest_token | VARCHAR(64) | INDEX | 游客令牌 |
| staff_id | BIGINT | | 接待客服ID |
| staff_name | VARCHAR(100) | | 客服名称 |
| status | INT | DEFAULT 0 | 状态：0等待接入 1进行中 2已结束 |
| rating | INT | DEFAULT 0 | 评分 1-5 |
| feedback | TEXT | | 评价内容 |
| ended_at | DATETIME | | 结束时间 |
| created_at | DATETIME | | 创建时间 |
| updated_at | DATETIME | | 更新时间 |

### 3.12.3 配置数据库表结构

#### db_configs 数据库配置表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | INTEGER | PRIMARY KEY | 主键 |
| db_type | TEXT | | 数据库类型：mysql/postgres/sqlite |
| host | TEXT | | 主机地址 |
| port | INTEGER | | 端口 |
| user | TEXT | | 用户名 |
| password | TEXT | | 密码（AES-GCM加密） |
| database | TEXT | | 数据库名 |
| server_port | INTEGER | DEFAULT 8080 | 服务器端口 |
| encryption_key | TEXT | | AES加密密钥（Base64） |
| created_at | DATETIME | | 创建时间 |
| updated_at | DATETIME | | 更新时间 |

### 3.12.4 索引设计

| 表名 | 索引名 | 字段 | 类型 | 说明 |
|------|--------|------|------|------|
| users | idx_users_username | username | UNIQUE | 用户名唯一索引 |
| users | idx_users_email | email | UNIQUE | 邮箱唯一索引 |
| orders | idx_orders_order_no | order_no | UNIQUE | 订单号唯一索引 |
| orders | idx_orders_user_id | user_id | INDEX | 用户ID索引 |
| orders | idx_orders_status | status | INDEX | 状态索引 |
| orders | idx_orders_created_at | created_at | INDEX | 创建时间索引 |
| products | idx_products_category | category_id | INDEX | 分类索引 |
| products | idx_products_status | status | INDEX | 状态索引 |
| support_tickets | idx_tickets_user | user_id | INDEX | 用户索引 |
| support_tickets | idx_tickets_guest | guest_token | INDEX | 游客令牌索引 |
| support_tickets | idx_tickets_status | status | INDEX | 状态索引 |
| support_messages | idx_messages_ticket | ticket_id | INDEX | 工单索引 |
| user_sessions | idx_sessions_id | session_id | UNIQUE | 会话ID唯一索引 |
| user_sessions | idx_sessions_expires | expires_at | INDEX | 过期时间索引 |

### 3.12.5 表关系图

```
┌─────────────┐     ┌─────────────┐     ┌─────────────────┐
│    users    │────<│   orders    │     │ product_categories│
└─────────────┘     └─────────────┘     └─────────────────┘
       │                   │                     │
       │                   │                     │
       ▼                   ▼                     ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│user_sessions│     │   coupons   │────<│  products   │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │coupon_usages│
                    └─────────────┘

┌─────────────────┐     ┌─────────────────┐
│ support_tickets │────<│support_messages │
└─────────────────┘     └─────────────────┘
         │
         │
         ▼
┌─────────────────┐     ┌─────────────────────┐
│  support_staff  │────<│support_staff_sessions│
└─────────────────┘     └─────────────────────┘

┌─────────────┐     ┌─────────────────────┐
│  live_chats │────<│ live_chat_messages  │
└─────────────┘     └─────────────────────┘
```

**关系说明**：
- `users` 1:N `orders`：一个用户可以有多个订单
- `users` 1:N `user_sessions`：一个用户可以有多个会话
- `products` N:1 `product_categories`：多个商品属于一个分类
- `orders` N:1 `coupons`：多个订单可以使用同一优惠券
- `support_tickets` 1:N `support_messages`：一个工单有多条消息
- `support_staff` 1:N `support_staff_sessions`：一个客服有多个会话
- `live_chats` 1:N `live_chat_messages`：一个聊天有多条消息

## 4. 配置管理

### 4.1 双数据库架构

系统采用双数据库架构：

1. **配置数据库 (SQLite)**：存储数据库连接配置，位于 `user_config/config.db`
2. **主数据库 (MySQL/PostgreSQL/SQLite)**：存储业务数据

### 4.2 配置存储位置

| 配置类型 | 存储位置 |
|---------|---------|
| 数据库连接配置 | SQLite配置数据库 (DBConfigDB) |
| 系统配置 | 主数据库 (SystemConfigDB) |
| 邮箱配置 | 主数据库 (EmailConfigDB) |
| 支付配置 | 主数据库 (PaymentConfigDB) |

### 4.3 配置结构

```go
// 数据库配置
type DBConfig struct {
    Type     string  // mysql, postgres, sqlite
    Host     string
    Port     int
    User     string
    Password string
    Database string
}

// 服务器配置
type ServerConfig struct {
    Port          int
    UseHTTPS      bool
    CertFile      string
    KeyFile       string
    AdminUsername string
    AdminPassword string
    AdminSuffix   string   // 管理后台路径后缀
    EnableLogin   bool     // 是否启用登录验证
    Enable2FA     bool
    TOTPSecret    string
    SystemTitle   string
}
```

### 4.4 数据加密密钥管理

系统使用 AES-GCM 加密敏感配置数据（如数据库密码）。

#### 4.4.1 密钥特性

| 特性 | 说明 |
|------|------|
| 加密算法 | AES-GCM |
| 支持密钥长度 | 128位 / 192位 / 256位（默认） |
| 存储位置 | SQLite配置数据库 (DBConfigDB.EncryptionKey) |
| 编码格式 | Base64 |

#### 4.4.2 密钥生命周期

1. **自动生成**：首次启动时自动生成256位AES密钥
2. **持久存储**：密钥存储在配置数据库中，重启后自动加载
3. **可视化显示**：在管理后台数据库配置页面以只读方式显示
4. **支持重置**：提供重置功能，需二级确认

#### 4.4.3 重置密钥警告

重置加密密钥是**危险操作**，会导致：
- 所有使用旧密钥加密的数据**永久无法解密**
- 数据库密码等敏感配置将丢失
- 需要重新配置数据库连接

#### 4.4.4 数据迁移

迁移数据时需要：
1. 在管理后台复制当前加密密钥
2. 在新环境配置相同的密钥
3. 导入数据后即可正常解密

## 5. API 接口

### 5.1 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/health | 健康检查 |
| GET | /api/announcements | 获取有效公告列表 |
| POST | /api/order/query | 公开订单查询（订单号+邮箱） |

### 5.2 用户接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/user/register | 用户注册 | 否 |
| POST | /api/user/login | 用户登录 | 否 |
| POST | /api/user/logout | 用户登出 | 是 |
| GET | /api/user/info | 获取用户信息 | 是 |
| PUT | /api/user/info | 更新用户信息 | 是 |
| POST | /api/user/password | 修改密码 | 是 |
| GET | /api/user/orders | 获取订单列表 | 是 |
| POST | /api/user/2fa/enable | 启用两步验证 | 是 |
| POST | /api/user/2fa/disable | 禁用两步验证 | 是 |
| POST | /api/user/email/send_code | 发送邮箱验证码 | 否 |
| POST | /api/user/forgot/check | 检查用户（找回密码） | 否 |
| POST | /api/user/forgot/reset | 重置密码 | 否 |

### 5.3 订单接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/order/create | 创建订单 | 是 |
| POST | /api/order/test | 创建测试订单 | 是 |
| GET | /api/order/detail/:order_no | 订单详情 | 是 |
| POST | /api/order/cancel | 取消订单 | 是 |

### 5.4 支付接口

#### 5.4.1 通用支付接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /api/payment/methods | 获取可用支付方式 | 否 |

**响应示例：**
```json
{
    "success": true,
    "methods": {
        "paypal": { "enabled": true, "sandbox": false },
        "alipay_f2f": { "enabled": true },
        "wechat_pay": { "enabled": false },
        "yi_pay": { "enabled": true }
    }
}
```

#### 5.4.2 PayPal 支付接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/paypal/create | 创建PayPal支付 | 是 |
| POST | /api/paypal/capture | 捕获PayPal支付（用户授权后） | 是 |
| GET | /paypal/return | PayPal支付成功回调 | 否 |
| GET | /paypal/cancel | PayPal支付取消回调 | 否 |

**创建PayPal支付请求：**
```json
{
    "order_no": "ORDER123456"
}
```

**创建PayPal支付响应：**
```json
{
    "success": true,
    "paypal_order_id": "PAYPAL_ORDER_ID",
    "approve_url": "https://www.paypal.com/checkoutnow?token=..."
}
```

**捕获PayPal支付请求：**
```json
{
    "order_no": "ORDER123456",
    "paypal_order_id": "PAYPAL_ORDER_ID"
}
```

**捕获PayPal支付响应：**
```json
{
    "success": true,
    "order_no": "ORDER123456",
    "kami_code": "KAMI-XXXX-XXXX-XXXX"
}
```

#### 5.4.3 支付宝当面付接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/alipay/create | 创建支付宝订单 | 是 |
| GET | /api/alipay/status/:order_no | 查询支付状态 | 是 |
| POST | /alipay/notify | 支付宝异步通知 | 否 |

**创建支付宝订单响应：**
```json
{
    "success": true,
    "qr_code": "https://qr.alipay.com/..."
}
```

#### 5.4.4 微信支付接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/wechat/create | 创建微信支付订单 | 是 |
| GET | /api/wechat/status/:order_no | 查询支付状态 | 是 |
| POST | /wechat/notify | 微信支付异步通知 | 否 |

**创建微信支付订单响应：**
```json
{
    "success": true,
    "qr_code": "weixin://wxpay/bizpayurl?..."
}
```

#### 5.4.5 易支付接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/yipay/create | 创建易支付订单 | 是 |
| POST | /api/yipay/callback | 易支付回调验证（前端） | 否 |
| POST | /yipay/notify | 易支付异步通知 | 否 |
| GET | /yipay/return | 易支付同步返回 | 否 |

**创建易支付订单响应：**
```json
{
    "success": true,
    "pay_url": "https://pay.example.com/submit.php?..."
}
```

### 5.5 管理后台接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/dashboard | 仪表盘数据 |
| GET/POST | /api/admin/products | 商品管理 |
| POST | /api/admin/product/:id/image | 上传商品图片 |
| DELETE | /api/admin/product/:id/image | 删除商品图片 |
| GET | /api/admin/orders | 订单列表 |
| GET | /api/admin/orders/search | 订单搜索（支持筛选） |
| GET | /api/admin/users | 用户管理 |
| GET/POST | /api/admin/settings | 系统设置 |
| GET/POST | /api/admin/db/* | 数据库配置 |
| GET/POST | /api/admin/payment/* | 支付配置 |
| GET/POST | /api/admin/email/* | 邮箱配置 |
| GET/POST/PUT/DELETE | /api/admin/announcements | 公告管理 |
| GET/POST/PUT/DELETE | /api/admin/categories | 分类管理 |
| GET | /api/admin/logs | 操作日志 |
| GET | /api/admin/stats/chart | 统计图表数据 |

### 5.6 手动卡密管理接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/product/:id/kami/import | 导入卡密 |
| GET | /api/admin/product/:id/kami | 获取商品卡密列表 |
| GET | /api/admin/product/:id/kami/stats | 获取卡密统计 |
| DELETE | /api/admin/kami/:id | 删除卡密 |
| POST | /api/admin/kami/:id/disable | 禁用卡密 |
| POST | /api/admin/kami/:id/enable | 启用卡密 |
| POST | /api/admin/kami/batch-delete | 批量删除卡密 |

#### 5.6.1 导入卡密

**接口路径：** `POST /api/admin/product/:id/kami/import`

**功能说明：** 批量导入手动卡密到指定商品

**请求参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| codes | string | 是 | 卡密内容，每行一个卡密 |

**响应示例：**
```json
{
    "success": true,
    "message": "卡密导入完成",
    "imported": 10,
    "duplicates": 2
}
```

#### 5.6.2 获取商品卡密列表

**接口路径：** `GET /api/admin/product/:id/kami`

**功能说明：** 分页获取商品的卡密列表

**请求参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20 |
| status | int | 否 | 状态筛选：0可用 1已售出 2已禁用 |

**响应示例：**
```json
{
    "success": true,
    "kamis": [...],
    "total": 100,
    "page": 1,
    "stats": {
        "total": 100,
        "available": 80,
        "sold": 15,
        "disabled": 5
    }
}
```

#### 5.6.3 获取卡密统计

**接口路径：** `GET /api/admin/product/:id/kami/stats`

**功能说明：** 获取商品的卡密统计信息

**响应示例：**
```json
{
    "success": true,
    "stats": {
        "total": 100,
        "available": 80,
        "sold": 15,
        "disabled": 5
    }
}
```

### 5.7 首页配置接口

首页配置功能允许管理员自定义用户端首页的显示内容和样式，支持多种模板和区块配置。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/homepage/config | 获取首页配置 |
| POST | /api/admin/homepage/config | 更新首页配置 |
| GET | /api/admin/homepage/templates | 获取可用模板列表 |
| GET | /api/admin/homepage/template/default | 获取模板默认配置 |
| POST | /api/admin/homepage/reset | 重置为默认配置 |
| GET | /api/homepage/config | 公开接口：获取首页配置 |

#### 5.7.1 获取首页配置

**接口路径：** `GET /api/admin/homepage/config`

**功能说明：** 获取当前首页配置

**响应示例：**
```json
{
    "success": true,
    "config": {
        "template": "modern",
        "primary_color": "#6366f1",
        "secondary_color": "#8b5cf6",
        "hero_enabled": true,
        "hero_title": "欢迎使用卡密购买系统",
        "hero_subtitle": "安全、便捷的卡密购买平台",
        "features_enabled": true,
        "features": [
            {"icon": "🔒", "title": "安全可靠", "description": "采用ECC加密通信"}
        ],
        "stats_enabled": true,
        "stats": [
            {"value": "10000+", "label": "用户数量", "icon": "👥"}
        ]
    }
}
```

#### 5.7.2 更新首页配置

**接口路径：** `POST /api/admin/homepage/config`

**功能说明：** 更新首页配置

**请求参数：** 完整的 HomepageConfig 对象

**响应示例：**
```json
{
    "success": true,
    "message": "配置已保存"
}
```

#### 5.7.3 获取模板列表

**接口路径：** `GET /api/admin/homepage/templates`

**功能说明：** 获取所有可用的首页模板

**响应示例：**
```json
{
    "success": true,
    "templates": [
        {"id": "modern", "name": "现代简约", "description": "简洁大气，适合大多数场景"},
        {"id": "gradient", "name": "渐变炫彩", "description": "丰富渐变色彩，视觉冲击力强"},
        {"id": "minimal", "name": "极简风格", "description": "极简设计，突出内容本身"},
        {"id": "card", "name": "卡片风格", "description": "卡片式布局，层次分明"},
        {"id": "hero", "name": "大图展示", "description": "全屏大图背景，适合品牌展示"},
        {"id": "business", "name": "商务专业", "description": "专业商务风格，适合企业用户"}
    ]
}
```

#### 5.7.4 获取模板默认配置

**接口路径：** `GET /api/admin/homepage/template/default`

**功能说明：** 获取指定模板的默认配置

**请求参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| template | string | 是 | 模板ID |

**响应示例：**
```json
{
    "success": true,
    "config": { ... }
}
```

#### 5.7.5 重置配置

**接口路径：** `POST /api/admin/homepage/reset`

**功能说明：** 将首页配置重置为当前模板的默认设置

**请求参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| template | string | 否 | 模板ID，不传则使用当前模板 |

**响应示例：**
```json
{
    "success": true,
    "message": "已重置为默认配置"
}
```

#### 5.7.6 首页配置字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| template | string | 模板ID：modern/gradient/minimal/card/hero/business |
| primary_color | string | 主色调（十六进制颜色值） |
| secondary_color | string | 次色调（十六进制颜色值） |
| advanced_mode | bool | 是否启用高级模式（自定义 HTML） |
| custom_html | string | 自定义 HTML 代码（高级模式） |
| custom_css | string | 自定义 CSS 样式（高级模式） |
| custom_js | string | 自定义 JavaScript 代码（高级模式） |
| hero_enabled | bool | 是否启用 Hero 区块 |
| hero_title | string | Hero 标题 |
| hero_subtitle | string | Hero 副标题 |
| hero_button_text | string | Hero 按钮文字 |
| hero_button_link | string | Hero 按钮链接 |
| hero_background | string | 背景类型：gradient/image/solid |
| features_enabled | bool | 是否启用特性区块 |
| features_title | string | 特性区块标题 |
| features | array | 特性列表 [{icon, title, description}] |
| announcement_enabled | bool | 是否启用公告区块 |
| announcement_title | string | 公告标题 |
| announcement_content | string | 公告内容 |
| announcement_type | string | 公告类型：info/warning/success |
| products_enabled | bool | 是否启用商品展示区块 |
| products_title | string | 商品区块标题 |
| products_count | int | 展示商品数量 |
| stats_enabled | bool | 是否启用统计区块 |
| stats | array | 统计项列表 [{value, label, icon}] |
| cta_enabled | bool | 是否启用 CTA 区块 |
| cta_title | string | CTA 标题 |
| cta_subtitle | string | CTA 副标题 |
| cta_button_text | string | CTA 按钮文字 |
| cta_button_link | string | CTA 按钮链接 |
| footer_text | string | 页脚文字 |
| footer_links | array | 页脚链接 [{text, url}] |
| floating_button_enabled | bool | 是否启用浮动按钮 |
| floating_button_icon | string | 浮动按钮图标（Font Awesome 类名） |
| floating_button_link | string | 浮动按钮链接 |

## 6. 安全特性

### 6.1 登录安全

- **登录失败锁定**：连续5次登录失败后锁定账户15分钟（持久化到数据库）
- **自动黑名单**：连续10次登录失败后IP自动加入临时黑名单30分钟
- **Session 管理**：基于 Cookie 的会话管理，会话数据持久化到数据库
- **会话超时**：用户会话2小时，管理员会话1小时，支持"记住我"功能（用户7天，管理员24小时）
- **两步验证**：支持 TOTP 和邮箱验证码两种方式
- **服务重启保持**：会话和登录锁定状态在服务重启后保持有效

### 6.2 API 安全

- **分级速率限制**：
  - 登录接口：每分钟10次
  - 注册接口：每分钟5次
  - 邮箱验证码：每分钟3次
  - 找回密码：每分钟5次
  - 管理后台API：每分钟60次
  - 支付接口：每分钟20次
  - 普通API：每分钟120次
- **CSRF 保护**：
  - 所有状态修改请求需要 CSRF 令牌
  - 令牌通过 Cookie 和请求头双重验证
  - 令牌有效期2小时，自动刷新
- **IP 黑名单**：支持临时和永久黑名单
- **图形验证码**：防止暴力破解
- **邮箱验证码**：有效期限制
- **重置密码令牌**：10分钟过期

### 6.3 安全响应头

系统自动添加以下安全响应头：

| 响应头 | 值 | 作用 |
|--------|-----|------|
| X-Frame-Options | SAMEORIGIN | 防止点击劫持 |
| X-Content-Type-Options | nosniff | 防止MIME类型嗅探 |
| X-XSS-Protection | 1; mode=block | XSS保护 |
| Referrer-Policy | strict-origin-when-cross-origin | 引用策略 |
| Content-Security-Policy | default-src 'self'... | 内容安全策略 |
| Permissions-Policy | geolocation=()... | 权限策略 |
| Strict-Transport-Security | max-age=31536000 | HTTPS强制（仅HTTPS模式） |

### 6.4 Cookie 安全

- **HttpOnly**：会话Cookie设置HttpOnly，防止JavaScript访问
- **SameSite**：设置SameSite=Lax，防止CSRF攻击
- **Secure**：生产环境启用Secure标志，仅HTTPS传输

### 6.5 密码安全

- 使用 bcrypt 进行密码哈希
- 密码最小长度6位
- 找回密码需要邮箱验证
- 数据库连接密码使用 AES-GCM 加密存储

### 6.6 通信安全

- 支持 HTTPS
- 敏感数据使用 AES-GCM 加密存储
- 可配置跳过 TLS 验证（仅用于测试）

### 6.7 操作审计

- 记录管理员和用户的关键操作
- 支持按用户类型、操作类型筛选
- 记录IP地址和User-Agent

### 6.8 前端安全

- **XSS 防护**：
  - 使用 `escapeHtml()` 函数转义所有用户输入
  - 所有表格渲染（商品、订单、用户、公告、日志等）均使用转义
  - 下拉选项值和显示文本均进行转义
- **CSRF 防护**：
  - 自动从 Cookie 获取 CSRF 令牌
  - 所有 POST/PUT/DELETE 请求自动添加 `X-CSRF-Token` 头
  - CSRF 验证失败时自动刷新令牌并重试
- **输入验证**：前端验证邮箱、手机号、密码强度
- **会话过期处理**：自动检测会话过期并跳转登录页
- **安全复制**：使用现代 Clipboard API，避免已弃用方法
- **防抖搜索**：订单搜索输入框使用防抖（500ms），减少不必要的请求

#### 前端 CSRF 使用示例

```javascript
// API请求自动处理CSRF令牌
const result = await apiRequest('/api/admin/product', {
    method: 'POST',
    body: { name: '商品名称', price: 99.99 }
});
// X-CSRF-Token 头会自动添加
```

### 6.9 前端模块架构

管理后台前端采用模块化架构，各模块职责清晰：

```
┌─────────────────────────────────────────────────────────────┐
│                     common.js (公共工具库)                    │
│  escapeHtml | debounce | throttle | apiRequest | dataCache  │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┴─────────────────────────────────┐
│                    admin-core.js (核心模块)                    │
│  AppState | PAGE_CONFIG | ModalManager | loadPage | 路由管理   │
└─────────────────────────────┬─────────────────────────────────┘
                              │
    ┌─────────────┬───────────┼───────────┬─────────────┐
    │             │           │           │             │
    ▼             ▼           ▼           ▼             ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
│dashboard│ │products │ │ orders  │ │  users  │ │ config  │
│  仪表盘  │ │商品管理 │ │订单管理 │ │用户管理 │ │系统配置 │
└─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘
    ┌─────────────┬───────────┬───────────┐
    │             │           │           │
    ▼             ▼           ▼           ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
│ payment │ │ content │ │ system  │ │ support │
│支付配置 │ │内容管理 │ │系统功能 │ │客服管理 │
└─────────┘ └─────────┘ └─────────┘ └─────────┘
```

**模块说明**：

| 模块 | 文件 | 功能 |
|------|------|------|
| 公共工具 | common.js | 安全工具、性能优化、API请求、格式化、验证 |
| 核心模块 | admin-core.js | 路由管理、状态管理、模态框管理、骨架屏 |
| 仪表盘 | admin-dashboard.js | 统计数据、图表渲染 |
| 商品管理 | admin-products.js | 商品CRUD、图片上传 |
| 订单管理 | admin-orders.js | 订单查询、筛选、分页 |
| 用户管理 | admin-users.js | 用户列表、状态管理 |
| 系统配置 | admin-config.js | 数据库、系统设置、安全设置 |
| 支付配置 | admin-payment.js | PayPal、支付宝、微信、易支付 |
| 内容管理 | admin-content.js | 分类、公告、优惠券 |
| 系统功能 | admin-system.js | 操作日志、数据备份 |

**全局状态管理 (AppState)**：

```javascript
const AppState = {
    currentPage: 'dashboard',      // 当前页面
    orderSearchParams: {...},      // 订单搜索参数
    chartDays: 7                   // 图表显示天数
};
```

## 7. 订单管理

### 7.1 订单状态

| 状态码 | 说明 |
|--------|------|
| 0 | 待支付 |
| 1 | 已支付 |
| 2 | 已完成 |
| 3 | 已取消 |
| 4 | 已退款 |
| 5 | 已过期 |

### 7.2 订单超时自动取消

- 订单创建时设置30分钟过期时间
- 后台定时任务每分钟检查过期订单
- 过期未支付订单自动标记为"已过期"状态

### 7.3 订单搜索

支持以下筛选条件：
- 订单号
- 用户名
- 订单状态
- 订单类型（正式/测试）
- 日期范围

## 8. 与Server端通信

### 8.1 BackendClient

## 9. 支付集成

### 9.1 支持的支付方式

| 支付方式 | 配置结构 | 状态 |
|---------|---------|------|
| PayPal | PayPalConfig | 已实现 |
| 支付宝当面付 | AlipayF2FConfig | 配置预留 |
| 微信支付 | WechatPayConfig | 配置预留 |
| 易支付 | YiPayConfig | 配置预留 |

### 9.2 PayPal 支付流程

```
用户 -> 创建订单 -> 创建PayPal支付 -> 跳转PayPal
                                        ↓
用户 <- 显示卡密 <- 生成卡密 <- 捕获支付 <- PayPal回调
```

## 10. 商品图片管理

### 10.1 功能说明

系统支持为每个商品上传图片，图片文件存储在程序根目录的 `Product` 文件夹中，每个商品有独立的子文件夹。

### 10.2 存储结构

```
Product/
├── 1/                          # 商品ID为1的文件夹
│   └── image_1702300000.jpg    # 商品图片（时间戳命名）
├── 2/                          # 商品ID为2的文件夹
│   └── image_1702300100.png
└── ...
```

### 10.3 图片上传API

**上传图片**
- 路径：`POST /api/admin/product/:id/image`
- Content-Type：`multipart/form-data`
- 参数：`image` - 图片文件
- 支持格式：JPG、PNG、GIF、WebP
- 大小限制：5MB

**删除图片**
- 路径：`DELETE /api/admin/product/:id/image`

### 10.4 图片访问

上传的图片通过静态文件服务访问：
- URL格式：`/product/{product_id}/image_xxx.jpg`
- 示例：`http://localhost:8080/product/1/image_1702300000.jpg`

## 11. 公告系统

### 11.1 功能说明

- 支持创建、编辑、删除公告
- 公告类型：info（信息）、warning（警告）、success（成功）、danger（危险）
- 支持设置公告有效期（开始时间、结束时间）
- 支持排序和启用/禁用

### 11.2 公开接口

用户端可通过 `/api/announcements` 获取当前有效的公告列表。

## 12. 分类管理

### 12.1 功能说明

- 支持创建商品分类
- 分类包含名称、图标、排序
- 商品可关联分类

## 13. 统计功能

### 13.1 仪表盘统计

- 总订单数
- 已完成订单数
- 总收入
- 今日订单数

### 13.2 趋势图表

- 支持查看近7天、14天、30天的订单趋势
- 显示每日订单数和收入
- 简洁的柱状图展示

## 14. 启动流程

```
1. 获取可执行文件目录
2. 初始化全局配置（设置默认值）
3. 初始化配置数据库（SQLite）
4. 初始化配置服务
5. 检查并迁移旧的JSON配置文件
6. 从SQLite加载数据库配置
7. 连接主数据库
   - 成功：使用配置的数据库
   - 失败：使用本地SQLite作为默认数据库
8. 初始化各服务（User, Admin, Order, Product, Email, Config, Security, Log, Announcement, Category）
9. 启动定时任务（订单过期检查、登录记录清理）
10. 从数据库加载运行时配置（系统、邮箱、支付）
11. 注册路由
12. 启动HTTP/HTTPS服务器
```

## 15. 定时任务

### 15.1 订单过期检查

- 执行间隔：每分钟
- 功能：检查并标记过期未支付的订单（30分钟未支付）

### 15.2 安全记录清理

- 执行间隔：每分钟
- 功能：清理过期的登录失败记录和API限流记录

### 15.3 会话清理

- 执行间隔：每分钟
- 功能：清理过期的用户会话（2小时）和管理员会话（1小时）

### 15.4 令牌清理

- 执行间隔：每分钟
- 功能：清理过期的重置密码令牌和登录验证令牌（10分钟）

## 16. 部署说明

### 16.1 编译

使用统一的构建脚本 `build.ps1`（Windows）或 `build.sh`（Linux）：

```powershell
# Windows - 默认编译（外部资源模式）
.\build.ps1

# Windows - 编译 Linux 版本
.\build.ps1 --linux

# Windows - 编译所有平台
.\build.ps1 --all

# Windows - 嵌入模式（前端资源打包进程序，生成单文件）
.\build.ps1 --embed

# 清理构建目录
.\build.ps1 --clean

# 强制重新构建（忽略缓存）
.\build.ps1 -Force
```

```bash
# Linux - 默认编译
./build.sh

# Linux - 嵌入模式
./build.sh --embed

# Linux - 编译 Windows 版本
./build.sh --windows
```

#### 构建模式说明

| 模式 | 参数 | 说明 |
|------|------|------|
| 外部资源模式 | （默认） | 前端资源作为独立文件，程序从 `./web/` 目录加载 |
| 嵌入模式 | `--embed` | 前端资源打包进二进制文件，生成单个可执行文件 |

嵌入模式适合需要单文件部署的场景，但会增加可执行文件体积（约 5MB）。

### 16.2 运行

```bash
# 直接运行
./user

# 后台运行 (Linux)
nohup ./user > user.log 2>&1 &
```

### 16.3 配置文件

首次运行会自动创建 `user_config` 目录和默认配置。

### 16.4 访问地址

- 用户前台：`http://localhost:8080/`
- 管理后台：`http://localhost:8080/manage`（默认后缀）

### 16.5 默认账户

- 管理员用户名：`admin`
- 管理员密码：`admin123`

**注意**：首次部署后请立即修改默认密码！

## 17. 优惠券系统

### 17.1 功能说明

系统支持创建和管理优惠券，用户下单时可使用优惠券获得折扣。

### 17.2 优惠券类型

| 类型 | 说明 | 示例 |
|------|------|------|
| percent | 折扣百分比 | 10% 折扣（9折） |
| fixed | 固定金额减免 | 减 5 元 |
| minus | 满减 | 满 100 减 20 |

### 17.3 优惠券属性

- **优惠券码**：唯一标识，用户输入使用
- **发放总量**：-1 表示无限
- **每人限用次数**：默认 1 次
- **最低消费金额**：订单金额需达到此值才能使用
- **最大优惠金额**：限制折扣类优惠券的最大优惠
- **适用商品/分类**：可限制优惠券适用范围
- **有效期**：开始时间和结束时间

### 17.4 API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/coupons | 获取优惠券列表 |
| POST | /api/admin/coupon | 创建优惠券 |
| PUT | /api/admin/coupon/:id | 更新优惠券 |
| DELETE | /api/admin/coupon/:id | 删除优惠券 |
| POST | /api/coupon/validate | 用户验证优惠券 |

## 18. 数据库备份

### 18.1 功能说明

系统支持手动创建数据库备份，备份文件存储在程序目录的 `backups` 文件夹中。

### 18.2 备份方式

| 数据库类型 | 备份方式 | 文件格式 |
|-----------|---------|---------|
| SQLite | 复制数据库文件 | ZIP 压缩包 |
| MySQL | SQL 导出 | .sql 文件 |
| PostgreSQL | SQL 导出 | .sql 文件 |

### 18.3 API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/backups | 获取备份列表 |
| GET | /api/admin/backup/info | 获取数据库信息 |
| POST | /api/admin/backup | 创建备份 |
| GET | /api/admin/backup/:id/download | 下载备份 |
| DELETE | /api/admin/backup/:id | 删除备份 |

### 18.4 备份存储

```
backups/
├── backup_sqlite_20251211_120000.zip
├── backup_mysql_20251211_130000.sql
└── backup_postgres_20251211_140000.sql
```

### 18.5 恢复说明

- **SQLite**：解压 ZIP 文件，替换原数据库文件
- **MySQL**：使用 `mysql -u user -p database < backup.sql` 导入
- **PostgreSQL**：使用 `psql -U user -d database < backup.sql` 导入

## 19. 客服支持系统

### 19.1 功能概述

系统提供完整的客服支持功能，包括工单系统和实时聊天，支持游客和登录用户使用。

### 19.2 数据模型

#### 19.2.1 工单模型 (SupportTicket)

```go
type SupportTicket struct {
    ID           uint       // 主键
    TicketNo     string     // 工单编号（唯一）
    UserID       uint       // 用户ID（0表示游客）
    Username     string     // 用户名或游客标识
    Email        string     // 联系邮箱
    Subject      string     // 工单主题
    Category     string     // 分类：order/product/payment/account/other
    Priority     int        // 优先级：1普通 2紧急 3非常紧急
    Status       int        // 状态：0待处理 1处理中 2已回复 3已解决 4已关闭
    AssignedTo   uint       // 分配给客服ID
    AssignedName string     // 客服名称
    RelatedOrder string     // 关联订单号
    GuestToken   string     // 游客访问令牌
    LastReplyAt  *time.Time // 最后回复时间
    LastReplyBy  string     // 最后回复人
    ClosedAt     *time.Time // 关闭时间
    ClosedBy     string     // 关闭人
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### 19.2.2 工单消息 (SupportMessage)

```go
type SupportMessage struct {
    ID         uint       // 主键
    TicketID   uint       // 关联工单ID
    SenderType string     // 发送者类型：user/guest/staff/system
    SenderID   uint       // 发送者ID
    SenderName string     // 发送者名称
    Content    string     // 消息内容
    IsInternal bool       // 是否内部备注（用户不可见）
    ReadAt     *time.Time // 已读时间
    CreatedAt  time.Time
}
```

#### 19.2.3 客服人员 (SupportStaff)

```go
type SupportStaff struct {
    ID           uint       // 主键
    Username     string     // 用户名（唯一）
    PasswordHash string     // 密码哈希
    Nickname     string     // 显示名称
    Avatar       string     // 头像URL
    Email        string     // 邮箱
    Role         string     // 角色：staff/supervisor
    Status       int        // 状态：1在线 0离线 -1禁用
    MaxTickets   int        // 最大同时处理工单数
    CurrentLoad  int        // 当前处理工单数
    LastActiveAt *time.Time // 最后活跃时间
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### 19.2.4 实时聊天 (LiveChat)

```go
type LiveChat struct {
    ID         uint       // 主键
    SessionID  string     // 聊天会话ID（唯一）
    UserID     uint       // 用户ID（0表示游客）
    Username   string     // 用户名
    GuestToken string     // 游客令牌
    StaffID    uint       // 接待客服ID
    StaffName  string     // 客服名称
    Status     int        // 状态：0等待接入 1进行中 2已结束
    Rating     int        // 评分 1-5
    Feedback   string     // 评价内容
    EndedAt    *time.Time // 结束时间
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

#### 19.2.5 客服配置 (SupportConfigDB)

```go
type SupportConfigDB struct {
    ID                uint      // 主键
    Enabled           bool      // 是否启用客服系统
    AllowGuest        bool      // 是否允许游客咨询
    WorkingHoursStart string    // 工作时间开始（如 "09:00"）
    WorkingHoursEnd   string    // 工作时间结束（如 "18:00"）
    WorkingDays       string    // 工作日（如 "1,2,3,4,5"）
    OfflineMessage    string    // 离线提示消息
    WelcomeMessage    string    // 欢迎消息
    AutoCloseHours    int       // 自动关闭无回复工单（小时）
    TicketCategories  string    // 工单分类（JSON数组）
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

### 19.3 工单状态

| 状态码 | 说明 |
|--------|------|
| 0 | 待处理 |
| 1 | 处理中 |
| 2 | 已回复 |
| 3 | 已解决 |
| 4 | 已关闭 |

### 19.4 工单优先级

| 优先级 | 说明 |
|--------|------|
| 1 | 普通 |
| 2 | 紧急 |
| 3 | 非常紧急 |

### 19.5 API 接口

#### 19.5.1 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/support/config | 获取客服配置（是否启用、是否允许游客等） |

#### 19.5.2 用户端工单接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/support/ticket | 创建工单 | 可选 |
| GET | /api/support/tickets | 获取用户工单列表 | 是 |
| GET | /api/support/tickets/guest | 游客获取工单列表 | 否 |
| GET | /api/support/ticket/:ticket_no | 获取工单详情 | 可选 |
| POST | /api/support/ticket/:ticket_no/reply | 回复工单 | 可选 |
| POST | /api/support/ticket/:ticket_no/close | 关闭工单 | 可选 |

#### 19.5.3 实时聊天接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/chat/start | 开始聊天 | 可选 |
| POST | /api/chat/:session_id/send | 发送消息 | 可选 |
| GET | /api/chat/:session_id/messages | 获取消息 | 可选 |
| POST | /api/chat/:session_id/end | 结束聊天 | 可选 |

#### 19.5.4 客服后台接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/staff/login | 客服登录 |
| POST | /api/staff/logout | 客服登出 |
| GET | /api/staff/info | 获取客服信息 |
| GET | /api/staff/tickets | 获取工单列表 |
| GET | /api/staff/ticket/:ticket_no | 获取工单详情 |
| POST | /api/staff/ticket/:ticket_no/reply | 回复工单 |
| PUT | /api/staff/ticket/:ticket_no/status | 更新工单状态 |
| POST | /api/staff/ticket/:ticket_no/assign | 分配工单 |
| GET | /api/staff/tickets/stats | 获取工单统计 |
| GET | /api/staff/chats/waiting | 获取等待接入的聊天 |
| POST | /api/staff/chat/:chat_id/accept | 接入聊天 |
| POST | /api/staff/chat/:session_id/send | 发送聊天消息 |
| GET | /api/staff/chat/:session_id/messages | 获取聊天消息 |
| POST | /api/staff/chat/:session_id/end | 结束聊天 |

#### 19.5.5 管理后台客服管理接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/support/config | 获取客服配置 |
| POST | /api/admin/support/config | 保存客服配置 |
| GET | /api/admin/support/staff | 获取客服列表 |
| POST | /api/admin/support/staff | 创建客服账号 |
| PUT | /api/admin/support/staff/:id | 更新客服信息 |
| DELETE | /api/admin/support/staff/:id | 删除客服 |
| GET | /api/admin/support/stats | 获取客服系统统计 |

### 19.6 游客访问机制

游客用户通过 `guest_token` 访问自己的工单和聊天：

1. 创建工单/聊天时，系统自动生成 `guest_token`
2. 游客需保存此令牌以便后续访问
3. 访问工单详情或回复时需提供 `guest_token`

### 19.7 页面路由

| 路径 | 说明 |
|------|------|
| /message | 客服支持页面（用户端） |
| /message/ticket/:ticket_no | 工单详情页 |
| /staff | 客服后台主页 |
| /staff/login | 客服登录页 |

### 19.8 定时任务

- **客服会话清理**：每分钟清理过期的客服会话
- **工单自动关闭**：可配置自动关闭长时间无回复的工单（默认72小时）

### 19.9 功能模块详解

#### 19.9.1 工单系统

工单系统是客服支持的核心功能，提供异步的问题处理机制。

**工单生命周期**：

```
用户提交工单 → 待处理(0) → 客服接单 → 处理中(1) → 客服回复 → 已回复(2)
                                                          ↓
                                              用户回复 → 处理中(1)
                                                          ↓
                                              问题解决 → 已解决(3) → 已关闭(4)
```

**工单分类**：
- 订单问题：与订单相关的咨询和问题
- 商品咨询：商品功能、使用方法等咨询
- 支付问题：支付失败、退款等问题
- 账户问题：登录、密码、账户安全等
- 其他：其他类型的问题

**工单功能特性**：
- 支持关联订单号，方便客服快速定位问题
- 支持内部备注，客服之间可以交流但用户不可见
- 支持工单分配，主管可以将工单分配给指定客服
- 支持优先级设置，紧急工单优先处理
- 自动记录最后回复时间和回复人

#### 19.9.2 实时聊天系统

实时聊天提供即时的在线客服功能。

**聊天流程**：

```
用户发起聊天 → 等待接入(0) → 客服接入 → 进行中(1) → 对话结束 → 已结束(2)
```

**聊天功能特性**：
- 支持多客服同时在线
- 客服可以查看等待队列
- 系统自动发送欢迎消息
- 支持消息类型：文本、图片、文件
- 聊天结束后可以评价

**消息轮询机制**：
- 用户端每3秒轮询一次新消息
- 客服端每2秒轮询一次新消息
- 使用 `after_id` 参数实现增量获取

#### 19.9.3 客服工作台

客服工作台是客服人员处理工单和聊天的主要界面。

**工作台功能**：

| 模块 | 功能 |
|------|------|
| 工单管理 | 查看工单列表、筛选、回复、更新状态、分配 |
| 在线咨询 | 查看等待队列、接入聊天、发送消息、结束对话 |
| 数据统计 | 工单统计、今日数据、状态分布 |

**工单筛选条件**：
- 状态筛选：全部/待处理/处理中/已回复/已解决/已关闭
- 优先级筛选：全部/普通/紧急/非常紧急
- 分类筛选：按工单分类筛选
- 只看我的：只显示分配给自己的工单

**客服角色**：
- `staff`：普通客服，可以处理工单和聊天
- `supervisor`：主管，可以管理其他客服、分配工单

#### 19.9.4 管理后台客服管理

管理员可以在管理后台管理客服系统。

**客服人员管理**：
- 创建客服账号（用户名、密码、昵称、邮箱、角色）
- 编辑客服信息（昵称、邮箱、最大工单数、状态）
- 修改客服密码
- 删除客服账号
- 查看客服在线状态和工作负载

**系统配置管理**：
- 启用/禁用客服系统
- 允许/禁止游客咨询
- 设置工作时间（开始时间、结束时间、工作日）
- 设置欢迎消息和离线提示
- 设置工单自动关闭时间
- 配置工单分类

**数据统计**：
- 客服总数和在线数
- 工单总数和今日新增
- 各状态工单数量分布

### 19.10 前端页面说明

#### 19.10.1 用户客服页面 (/message)

用户访问客服支持的主页面，包含两个标签页：

**在线咨询标签页**：
- 显示客服在线状态和在线人数
- 未开始聊天时显示开始咨询按钮
- 聊天中显示消息列表和输入框
- 支持结束对话

**工单中心标签页**：
- 显示用户的工单列表
- 支持创建新工单
- 点击工单查看详情和回复
- 支持关闭工单

#### 19.10.2 客服登录页 (/staff/login)

客服人员登录页面：
- 用户名和密码输入
- 登录成功后跳转到工作台

#### 19.10.3 客服工作台 (/staff)

客服人员的主要工作界面，包含三个标签页：

**工单管理标签页**：
- 工单列表表格（工单号、用户、主题、分类、优先级、状态、处理人、时间）
- 筛选栏（状态、只看我的）
- 点击工单打开详情弹窗
- 详情弹窗支持查看消息、回复、添加内部备注、更新状态、分配

**在线咨询标签页**：
- 左侧显示等待接入的聊天列表
- 右侧显示当前聊天窗口
- 支持接入、发送消息、结束对话

**数据统计标签页**：
- 统计卡片（待处理、处理中、已回复、今日新增）
- 工单状态分布

#### 19.10.4 管理后台客服管理

在管理后台侧边栏点击"客服管理"进入，包含三个子标签页：

**客服人员标签页**：
- 客服列表表格
- 添加客服按钮
- 编辑和删除操作

**系统配置标签页**：
- 基本设置（启用、允许游客）
- 工作时间设置
- 消息设置（欢迎消息、离线提示）
- 工单设置（自动关闭时间、分类配置）

**数据统计标签页**：
- 客服统计（总数、在线数）
- 工单统计（总数、今日、各状态分布）

### 19.11 安全机制

#### 19.11.1 游客令牌机制

游客用户通过 `guest_token` 进行身份验证：
- 32位随机十六进制字符串
- 创建工单/聊天时自动生成
- 存储在 localStorage 中
- 访问工单/聊天时需要提供

#### 19.11.2 客服认证

客服使用独立的认证系统：
- 独立的登录接口 `/api/staff/login`
- 使用 `staff_session` Cookie
- 会话有效期24小时
- 密码使用 bcrypt 加密存储

#### 19.11.3 权限控制

- 用户只能访问自己的工单和聊天
- 游客只能通过 `guest_token` 访问
- 客服可以访问所有工单和聊天
- 管理员可以管理客服和配置

## 20. 前端架构

### 19.1 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 框架 | Next.js | 14.2.x |
| UI库 | React | 18.3.x |
| 状态管理 | Zustand | 5.x |
| 样式 | Tailwind CSS | 3.4.x |
| 动画 | Framer Motion | 11.x |
| 通知 | React Hot Toast | 2.4.x |
| 语言 | TypeScript | 5.x |

### 19.2 目录结构

```
web/
├── src/
│   ├── app/                    # Next.js App Router 页面
│   │   ├── layout.tsx          # 根布局
│   │   ├── page.tsx            # 首页
│   │   ├── login/              # 登录页
│   │   ├── register/           # 注册页
│   │   ├── forgot/             # 找回密码
│   │   ├── verify/             # 2FA验证
│   │   ├── products/           # 商品列表
│   │   ├── user/               # 用户中心
│   │   │   ├── page.tsx        # 用户中心主页
│   │   │   └── modals.tsx      # 用户中心弹窗组件
│   │   └── admin/              # 管理后台
│   │       ├── page.tsx        # 管理后台主页
│   │       ├── login/          # 管理员登录
│   │       └── totp/           # 管理员TOTP验证
│   ├── components/
│   │   ├── ui/                 # 基础UI组件
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Card.tsx
│   │   │   └── Badge.tsx
│   │   └── layout/             # 布局组件
│   │       └── Navbar.tsx      # 导航栏和页脚
│   └── lib/
│       ├── api.ts              # API请求封装
│       ├── store.ts            # Zustand状态管理
│       └── utils.ts            # 工具函数
├── package.json
├── next.config.js              # Next.js配置（静态导出）
├── tailwind.config.ts          # Tailwind配置
└── tsconfig.json               # TypeScript配置
```

### 19.3 构建配置

Next.js 配置为静态导出模式，输出到 `out` 目录：

```javascript
// next.config.js
const nextConfig = {
  output: 'export',
  assetPrefix: '/static',
  trailingSlash: true,
  images: { unoptimized: true },
}
```

### 19.4 状态管理

使用 Zustand 进行全局状态管理：

```typescript
interface AppState {
  user: UserInfo | null
  setUser: (user: UserInfo | null) => void
  twoFAStatus: TwoFAStatus | null
  setTwoFAStatus: (status: TwoFAStatus | null) => void
  isLoggedIn: boolean
  setIsLoggedIn: (value: boolean) => void
}
```

### 19.5 API 封装

统一的 API 请求封装，自动处理 CSRF Token：

```typescript
// GET 请求
const res = await apiGet<{ user: UserInfo }>('/api/user/info')

// POST 请求
const res = await apiPost('/api/user/login', { username, password })
```

## 21. 代码优化与工具

### 21.1 统一响应辅助函数

文件：`internal/api/response_helper.go`

提供统一的 API 响应格式，消除重复代码：

```go
// 成功响应（带数据）
Success(c, data)

// 成功响应（带消息）
SuccessMessage(c, "操作成功")

// 分页响应
SuccessPage(c, list, total, page, pageSize)

// 错误响应
Error(c, http.StatusBadRequest, "参数错误")

// 服务器错误响应
ServerError(c, "服务器内部错误")

// 未授权响应
Unauthorized(c, "请先登录")

// 未找到响应
NotFound(c, "资源不存在")

// 参数绑定错误响应
BindError(c, err)
```

### 21.2 统一错误码系统

文件：`internal/api/error_codes.go`

定义统一的错误码，便于前端处理：

| 错误码 | 说明 | 适用场景 |
|--------|------|----------|
| 0 | 成功 | 操作成功 |
| 1001 | 参数错误 | 请求参数缺失或格式错误 |
| 1002 | 未授权 | 未登录或会话过期 |
| 1003 | 禁止访问 | 无权限访问 |
| 1004 | 资源不存在 | 请求的资源不存在 |
| 2001 | 用户不存在 | 用户查询失败 |
| 2002 | 密码错误 | 登录密码错误 |
| 2003 | 用户名已存在 | 注册时用户名重复 |
| 2004 | 邮箱已存在 | 邮箱已被使用 |
| 2005 | 验证码错误 | 验证码不正确或已过期 |
| 2006 | 账户已锁定 | 登录失败次数过多 |
| 2007 | 两步验证失败 | TOTP 验证失败 |
| 3001 | 商品不存在 | 商品查询失败 |
| 3002 | 库存不足 | 商品库存不足 |
| 3003 | 商品已下架 | 商品状态异常 |
| 4001 | 订单不存在 | 订单查询失败 |
| 4002 | 订单状态错误 | 订单状态不允许操作 |
| 4003 | 支付失败 | 支付处理失败 |
| 4004 | 订单已过期 | 订单超时 |
| 5001 | 优惠券不存在 | 优惠券查询失败 |
| 5002 | 优惠券已过期 | 优惠券已失效 |
| 5003 | 优惠券已使用 | 优惠券已被使用 |
| 5004 | 不满足使用条件 | 金额不满足优惠券要求 |
| 6001 | 工单不存在 | 工单查询失败 |
| 6002 | 工单已关闭 | 工单状态不允许操作 |
| 9001 | 服务器错误 | 服务器内部错误 |
| 9002 | 数据库错误 | 数据库操作失败 |
| 9003 | 网络错误 | 外部服务调用失败 |

### 21.3 环境配置管理

文件：`internal/config/environment.go`

支持开发、测试、生产三种环境：

```go
// 环境类型
const (
    EnvDevelopment = "development"
    EnvTesting     = "testing"
    EnvProduction  = "production"
)

// 设置环境变量 GO_ENV 或 ENV 来切换环境
// development: 开发模式，详细日志，调试信息
// testing: 测试模式，使用测试数据库
// production: 生产模式，优化性能

// 使用示例
env := utils.GetEnv()
if env.IsDevelopment() {
    // 开发环境专用代码
}

// 获取配置
dbType := env.GetDatabaseType()
logLevel := env.GetLogLevel()
```

### 21.4 统一日志系统

文件：`internal/utils/logger.go`

提供结构化日志输出：

```go
// 初始化日志（在 main.go 中调用）
logger := utils.InitLogger()

// 记录日志
utils.LogInfo("用户登录", "user_id", 123, "ip", "192.168.1.1")
utils.LogWarn("库存不足", "product_id", 456, "stock", 0)
utils.LogError("支付失败", "order_no", "ORD123", "error", err)
utils.LogDebug("调试信息", "data", someData)

// 创建带上下文的日志
ctxLogger := utils.WithContext(ctx)
ctxLogger.Info("处理请求")

// API 请求日志
utils.LogRequest(c, startTime)

// 数据库操作日志
utils.LogDB("query", "SELECT * FROM users", duration)
```

### 21.5 Swagger API 文档

#### 21.5.1 安装和配置

1. 安装 swag 命令行工具：
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. 生成文档：
```bash
cd User
swag init -g cmd/server/main.go -o docs
```

3. 访问文档：
- Swagger UI: `http://localhost:8080/swagger/`
- OpenAPI JSON: `http://localhost:8080/swagger/doc.json`

#### 21.5.2 文档文件

| 文件 | 说明 |
|------|------|
| `swagger_models.go` | API 请求/响应模型定义 |
| `user_handler_swagger.go` | 用户认证相关 API 注解 |
| `order_handler_swagger.go` | 订单/商品/支付 API 注解 |
| `support_handler_swagger.go` | 客服支持 API 注解 |
| `admin_handler_swagger.go` | 管理后台 API 注解 |
| `swagger_router.go` | Swagger 路由注册 |

### 21.6 单元测试

#### 21.6.1 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/service/...
go test ./internal/utils/...

# 显示详细输出
go test -v ./internal/service/...

# 运行基准测试
go test -bench=. ./internal/utils/...

# 生成覆盖率报告
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### 21.6.2 测试文件

| 文件 | 说明 |
|------|------|
| `internal/test/test_helper.go` | 测试辅助函数和 Mock 对象 |
| `internal/service/user_service_test.go` | 用户服务测试 |
| `internal/service/order_service_test.go` | 订单服务测试 |
| `internal/utils/utils_test.go` | 工具函数测试 |

#### 21.6.3 测试辅助函数

```go
// 设置测试数据库（SQLite 内存数据库）
db := test.SetupTestDB(t)

// 设置测试服务
services := test.SetupTestServices(t, db)

// 创建测试用户
user := test.CreateTestUser(t, db, "testuser")

// 创建测试商品
product := test.CreateTestProduct(t, db, "Test Product", 99.99)

// 创建测试订单
order := test.CreateTestOrder(t, db, user.ID, product.ID)

// 执行 HTTP 请求测试
resp := test.ExecuteRequest(router, "POST", "/api/user/login", body)

// 断言函数
test.AssertEqual(t, expected, actual, "message")
test.AssertNoError(t, err)
test.AssertHTTPStatus(t, resp, http.StatusOK)
test.AssertJSONSuccess(t, resp)
```

## 22. 邮箱配置

### 22.1 邮箱配置模型

```go
type EmailConfigDB struct {
    ID           uint      // 主键
    Enabled      bool      // 是否启用邮箱服务
    SMTPHost     string    // SMTP服务器地址
    SMTPPort     int       // SMTP端口
    SMTPUser     string    // SMTP用户名
    SMTPPassword string    // SMTP密码（加密存储）
    FromName     string    // 发件人名称
    FromEmail    string    // 发件人邮箱
    Encryption   string    // 加密方式：none/ssl/starttls
    CodeLength   int       // 验证码长度（4-8位）
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 22.2 加密方式说明

| 加密方式 | 说明 | 推荐端口 | 适用场景 |
|---------|------|---------|---------|
| ssl | SSL/TLS 加密 | 465 | QQ邮箱、163邮箱、阿里云企业邮箱 |
| starttls | STARTTLS 加密 | 587 | Gmail、Outlook、Office365 |
| none | 无加密 | 25 | 内网邮件服务器（不推荐） |

### 22.3 常用邮箱配置

| 邮箱服务 | SMTP服务器 | 端口 | 加密方式 | 备注 |
|---------|-----------|------|---------|------|
| QQ邮箱 | smtp.qq.com | 465 | SSL | 需使用授权码 |
| 163邮箱 | smtp.163.com | 465 | SSL | 需使用授权码 |
| Gmail | smtp.gmail.com | 587 | STARTTLS | 需开启应用专用密码 |
| Outlook | smtp.office365.com | 587 | STARTTLS | - |
| 阿里云企业邮箱 | smtp.qiye.aliyun.com | 465 | SSL | - |

### 22.4 API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/email/config | 获取邮箱配置 |
| POST | /api/admin/email/config | 保存邮箱配置 |
| POST | /api/admin/email/test | 发送测试邮件 |

### 22.5 邮箱配置请求参数

```json
{
    "enabled": true,
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "smtp_user": "your@gmail.com",
    "smtp_password": "app_password",
    "from_name": "系统通知",
    "from_email": "your@gmail.com",
    "encryption": "starttls",
    "code_length": 6
}
```

## 23. 多管理员系统

### 23.1 双管理员表架构

系统支持两种管理员存储方式：

| 表名 | 说明 | 用途 |
|------|------|------|
| system_configs | 系统配置表 | 存储默认管理员（admin_username/admin_password） |
| admins | 多管理员表 | 存储基于角色的多管理员账户 |

### 23.2 管理员模型 (Admin)

```go
type Admin struct {
    ID           uint       // 主键
    Username     string     // 用户名（唯一）
    PasswordHash string     // 密码哈希
    Nickname     string     // 昵称
    Email        string     // 邮箱
    RoleID       uint       // 角色ID
    Status       int        // 状态：1启用 0禁用
    LastLoginAt  *time.Time // 最后登录时间
    LastLoginIP  string     // 最后登录IP
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 23.3 角色模型 (AdminRole)

```go
type AdminRole struct {
    ID          uint      // 主键
    Name        string    // 角色名称（唯一）
    Description string    // 角色描述
    Permissions string    // 权限列表（JSON数组）
    IsSystem    bool      // 是否系统角色（不可删除）
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 23.4 权限列表

| 权限标识 | 说明 |
|---------|------|
| dashboard:view | 查看仪表盘 |
| product:view/create/edit/delete | 商品管理 |
| category:view/create/edit/delete | 分类管理 |
| order:view/edit/refund | 订单管理 |
| user:view/edit/delete | 用户管理 |
| admin:view/create/edit/delete | 管理员管理 |
| role:view/create/edit/delete | 角色管理 |
| coupon:view/create/edit/delete | 优惠券管理 |
| announcement:view/create/edit/delete | 公告管理 |
| support:view/manage | 客服管理 |
| settings:view/edit/payment/email/database | 系统设置 |
| log:view | 查看日志 |
| backup:view/create/delete | 备份管理 |

### 23.5 权限模板

| 模板名称 | 说明 | 权限范围 |
|---------|------|---------|
| super_admin | 超级管理员 | 所有权限 |
| admin | 普通管理员 | 除角色/管理员管理外的所有权限 |
| operator | 运营人员 | 商品、订单、优惠券、公告管理 |
| support | 客服人员 | 订单查看、用户查看、客服管理 |
| readonly | 只读用户 | 所有 view 权限 |

### 23.6 登录认证流程

管理员登录时，系统按以下顺序验证：

1. **优先检查 admins 表**：查找用户名匹配的管理员账户
2. **回退到系统配置**：如果 admins 表中不存在，检查 system_configs 中的默认管理员

### 23.7 API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/admins | 获取管理员列表 |
| POST | /api/admin/admin | 创建管理员 |
| PUT | /api/admin/admin/:id | 更新管理员 |
| DELETE | /api/admin/admin/:id | 删除管理员 |
| GET | /api/admin/roles | 获取角色列表 |
| POST | /api/admin/role | 创建角色 |
| PUT | /api/admin/role/:id | 更新角色 |
| DELETE | /api/admin/role/:id | 删除角色 |
| GET | /api/admin/permissions | 获取权限列表和模板 |

## 24. 版本信息

- 文档版本：3.7
- 更新日期：2025-12-28
- Go 版本：1.23

### 更新记录

