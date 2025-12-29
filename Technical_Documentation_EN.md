# User - Client Portal Technical Documentation

> Version: 3.9 | Updated: December 2025

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [System Architecture](#2-system-architecture)
3. [Directory Structure](#3-directory-structure)
4. [Data Models](#4-data-models)
5. [Configuration Management](#5-configuration-management)
6. [API Reference](#6-api-reference)
7. [Security Features](#7-security-features)
8. [Order Management](#8-order-management)
9. [Payment Integration](#9-payment-integration)
10. [Support System](#10-support-system)
11. [Frontend Architecture](#11-frontend-architecture)
12. [Deployment Guide](#12-deployment-guide)
13. [Changelog](#13-changelog)
21. [Role-Based Access Control](#21-role-based-access-control)
22. [Balance System](#22-balance-system)
23. [Product Multi-Image System](#23-product-multi-image-system)
24. [Shopping Cart System](#24-shopping-cart-system)
25. [Product Favorites System](#25-product-favorites-system)
26. [Points System](#26-points-system)
27. [Scheduled Tasks System](#27-scheduled-tasks-system)
28. [Invoice System](#28-invoice-system)
29. [Operation Undo System](#29-operation-undo-system)
30. [Smart Customer Service System](#30-smart-customer-service-system)

---

## 1. System Overview

User is a client-facing card key purchase system frontend, developed in Go using the Gin framework. The system provides user registration/login, product browsing, order management, multiple payment methods, and communicates securely with the Server via ECC signatures.

### 1.1 Technology Stack

| Component | Technology |
|-----------|------------|
| Backend Framework | Gin v1.9.1 |
| ORM | GORM v1.25.5 |
| Database | MySQL / PostgreSQL / SQLite |
| Authentication | Session + Cookie |
| 2FA | TOTP (pquerna/otp) |
| Captcha | base64Captcha |
| Encryption | ECC (ECDSA P-256) |
| Frontend | React + Next.js + TypeScript |
| Real-time | WebSocket (gorilla/websocket) |

### 1.2 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      User Client Portal                      │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │  User   │  │ Product │  │  Order  │  │ Payment │        │
│  │ Module  │  │ Module  │  │ Module  │  │ Module  │        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
│       │            │            │            │              │
│  ┌────┴────────────┴────────────┴────────────┴────┐        │
│  │                  Service Layer                  │        │
│  │  SecuritySvc | LogSvc | SupportSvc | ...       │        │
│  └────────────────────────┬─────────────────────────┘        │
│                           │                                  │
│  ┌────────────────────────┴─────────────────────────┐        │
│  │                Repository Layer                   │        │
│  └────────────────────────┬─────────────────────────┘        │
│                           │                                  │
│  ┌────────────────────────┴─────────────────────────┐        │
│  │           Database (MySQL/PostgreSQL/SQLite)      │        │
│  └──────────────────────────────────────────────────┘        │
│                           │                                  │
│  ┌────────────────────────┴─────────────────────────┐        │
│  │              BackendClient (ECC Signature)        │        │
│  └────────────────────────┬─────────────────────────┘        │
└───────────────────────────┼─────────────────────────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │  Server API   │
                    └───────────────┘
```

---

## 2. System Architecture

### 2.1 Dual Database Architecture

The system uses a dual database architecture:

1. **Config Database (SQLite)**: Stores database connection configuration, located at `user_config/config.db`
2. **Main Database (MySQL/PostgreSQL/SQLite)**: Stores business data

### 2.2 Module Overview

| Module | Description |
|--------|-------------|
| User Module | Registration, login, profile, 2FA |
| Product Module | Product listing, categories, images |
| Order Module | Order creation, payment, status tracking |
| Payment Module | PayPal, Alipay, WeChat Pay integration |
| Support Module | Ticket system, live chat, staff management |
| Admin Module | Dashboard, configuration, user management |

---

## 3. Directory Structure

```
User/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── api/
│   │   ├── handler.go           # Route registration
│   │   ├── middleware.go        # Security middleware
│   │   ├── user_handler.go      # User APIs
│   │   ├── admin_handler.go     # Admin APIs
│   │   ├── order_handler.go     # Order APIs
│   │   ├── support_handler.go   # Support APIs (tickets, chat)
│   │   ├── support_staff_handler.go # Staff backend APIs
│   │   └── websocket_handler.go # WebSocket handler
│   ├── config/
│   │   └── config.go            # Configuration
│   ├── model/
│   │   ├── models.go            # Data models
│   │   ├── support.go           # Support system models
│   │   └── db.go                # Database initialization
│   ├── repository/
│   │   └── repository.go        # Data access layer
│   ├── service/
│   │   ├── user_service.go      # User business logic
│   │   ├── order_service.go     # Order business logic
│   │   ├── support_service.go   # Support business logic
│   │   ├── websocket_service.go # WebSocket hub manager
│   │   └── backend_client.go    # Server communication
│   └── utils/
│       ├── crypto.go            # Password encryption
│       ├── ecc.go               # ECC encryption/signing
│       └── order.go             # Order number generation
├── web/                         # Frontend (React + Next.js)
├── Product/                     # Product images storage
├── user_config/                 # Config directory
├── go.mod
├── build_windows.ps1
└── build_linux.ps1
```

---

## 4. Data Models

### 4.1 User Model

```go
type User struct {
    ID              uint           // Primary key
    Username        string         // Username (unique)
    Email           string         // Email (unique)
    PasswordHash    string         // Password hash
    Phone           string         // Phone number
    EmailVerified   bool           // Email verified
    Enable2FA       bool           // 2FA enabled
    TOTPSecret      string         // TOTP secret
    Status          int            // Status: 1=active, 0=disabled
    LastLoginAt     *time.Time     // Last login time
    LastLoginIP     string         // Last login IP
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### 4.2 Product Model

```go
type Product struct {
    ID           uint           // Primary key
    Name         string         // Product name
    Description  string         // Description
    Price        float64        // Price
    Duration     int            // Duration value
    DurationUnit string         // Duration unit: day/week/month/year
    Stock        int            // Stock, -1 = unlimited
    Status       int            // Status: 1=active, 0=inactive
    AllowTest    bool           // Allow test purchase
    SortOrder    int            // Sort order
    ImageURL     string         // Product image
    CategoryID   uint           // Category ID
    ProductType  int            // Product type: 1=manual kami (default)
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Product Type Description**:
| Value | Name | Description |
|-------|------|-------------|
| 1 | Manual Kami | Default mode, admin imports kami manually, allocate from local pool when order completes |

### 4.2.1 Manual Kami Model

```go
type ManualKami struct {
    ID        uint           // Primary key
    ProductID uint           // Associated product ID
    KamiCode  string         // Kami content
    Status    int            // Status: 0=available, 1=sold, 2=disabled
    OrderID   uint           // Associated order ID (filled after sold)
    OrderNo   string         // Associated order number
    SoldAt    *time.Time     // Sold time
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Kami Status Description**:
| Value | Name | Description |
|-------|------|-------------|
| 0 | Available | Kami can be allocated to new orders |
| 1 | Sold | Kami has been allocated to an order |
| 2 | Disabled | Kami is disabled by admin, cannot be allocated |

### 4.3 Order Model

```go
type Order struct {
    ID            uint           // Primary key
    OrderNo       string         // Order number (unique)
    PaymentNo     string         // Payment order number
    UserID        uint           // User ID
    Username      string         // Username
    Email         string         // User email
    ProductID     uint           // Product ID
    ProductName   string         // Product name
    Price         float64        // Price
    Status        int            // Status: 0=pending, 1=paid, 2=completed, 3=cancelled, 4=refunded, 5=expired
    PaymentMethod string         // Payment method
    PaymentTime   *time.Time     // Payment time
    KamiCode      string         // Generated card key
    IsTest        bool           // Is test order
    ExpireAt      *time.Time     // Order expiration
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### 4.4 Support Ticket Model

```go
type SupportTicket struct {
    ID           uint       // Primary key
    TicketNo     string     // Ticket number (unique)
    UserID       uint       // User ID (0 = guest)
    Username     string     // Username
    Email        string     // Contact email
    Subject      string     // Ticket subject
    Category     string     // Category
    Priority     int        // Priority: 1=normal, 2=urgent, 3=critical
    Status       int        // Status: 0=pending, 1=processing, 2=replied, 3=resolved, 4=closed
    AssignedTo   uint       // Assigned staff ID
    GuestToken   string     // Guest access token
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

---

## 5. Configuration Management

### 5.1 Configuration Storage

| Config Type | Storage Location |
|-------------|------------------|
| Database Connection | SQLite config database (DBConfigDB) |
| System Config | Main database (SystemConfigDB) |
| Email Config | Main database (EmailConfigDB) |
| Payment Config | Main database (PaymentConfigDB) |

### 5.2 Encryption Key Management

- **Algorithm**: AES-GCM
- **Key Length**: 128/192/256 bits (default 256)
- **Storage**: SQLite config database
- **Auto-generation**: Generated on first startup

---

## 6. API Reference

### 6.1 Public APIs

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/health | Health check |
| GET | /api/announcements | Get active announcements |
| POST | /api/order/query | Public order query |

### 6.2 User APIs

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/user/register | User registration | No |
| POST | /api/user/login | User login | No |
| POST | /api/user/logout | User logout | Yes |
| GET | /api/user/info | Get user info | Yes |
| PUT | /api/user/info | Update user info | Yes |
| POST | /api/user/password | Change password | Yes |
| GET | /api/user/orders | Get order list | Yes |
| POST | /api/user/2fa/enable | Enable 2FA | Yes |
| POST | /api/user/2fa/disable | Disable 2FA | Yes |

### 6.3 Order APIs

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/order/create | Create order | Yes |
| POST | /api/order/test | Create test order | Yes |
| GET | /api/order/detail/:order_no | Order detail | Yes |
| POST | /api/order/cancel | Cancel order | Yes |

### 6.4 Support APIs

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/support/ticket | Create ticket | Optional |
| GET | /api/support/tickets | Get user tickets | Yes |
| GET | /api/support/ticket/:ticket_no | Get ticket detail | Optional |
| POST | /api/support/ticket/:ticket_no/reply | Reply to ticket | Optional |
| POST | /api/chat/start | Start live chat | Optional |
| POST | /api/chat/:session_id/send | Send chat message | Optional |

### 6.5 WebSocket APIs

| Path | Description |
|------|-------------|
| /ws/user | User WebSocket connection |
| /ws/staff | Staff WebSocket connection |

---

## 7. Security Features

### 7.1 Login Security

- **Login Failure Lock**: 5 consecutive failures = 15 min lock (persisted to database)
- **Auto Blacklist**: 10 consecutive failures = 30 min IP blacklist
- **Session Management**: Cookie-based, persisted to database
- **Session Timeout**: User 2 hours, Admin 1 hour
- **2FA**: TOTP and email verification support

### 7.2 API Security

- **Tiered Rate Limiting**:
  - Login: 10/min
  - Registration: 5/min
  - Email verification: 3/min
  - Admin API: 60/min
  - Payment: 20/min
  - Default: 120/min
- **CSRF Protection**: Token verification, 2-hour validity
- **IP Blacklist**: Temporary and permanent support

### 7.3 Security Headers

| Header | Value | Purpose |
|--------|-------|---------|
| X-Frame-Options | SAMEORIGIN | Prevent clickjacking |
| X-Content-Type-Options | nosniff | Prevent MIME sniffing |
| X-XSS-Protection | 1; mode=block | XSS protection |
| Content-Security-Policy | default-src 'self'... | Content security |

---

## 8. Order Management

### 8.1 Order Status

| Code | Description |
|------|-------------|
| 0 | Pending payment |
| 1 | Paid |
| 2 | Completed |
| 3 | Cancelled |
| 4 | Refunded |
| 5 | Expired |

### 8.2 Auto Expiration

- Orders expire 30 minutes after creation
- Background task checks every minute
- Unpaid expired orders marked as "Expired"

---

## 9. Payment Integration

### 9.1 Supported Payment Methods

| Method | Config | Status |
|--------|--------|--------|
| PayPal | PayPalConfig | Implemented |
| Alipay F2F | AlipayF2FConfig | Config ready |
| WeChat Pay | WechatPayConfig | Config ready |
| YiPay | YiPayConfig | Config ready |
| Stripe | StripeConfig | Implemented |
| USDT | USDTConfig | Implemented |

### 9.2 PayPal Flow

```
User -> Create Order -> Create PayPal Payment -> Redirect to PayPal
                                                      ↓
User <- Show Card Key <- Generate Key <- Capture Payment <- PayPal Callback
```

### 9.3 Stripe Integration

**API Endpoints:**
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/stripe/config | Get Stripe config (public key) | No |
| POST | /api/stripe/create | Create Checkout Session | Yes |
| GET | /api/stripe/verify/:session_id | Verify payment status | Yes |
| POST | /stripe/webhook | Stripe Webhook callback | No |
| POST | /api/admin/stripe/test | Test Stripe connection | Admin |

**Configuration:**
| Field | Description |
|-------|-------------|
| stripe_enabled | Enable Stripe payment |
| stripe_publishable_key | Stripe public key (frontend) |
| stripe_secret_key | Stripe secret key (backend) |
| stripe_webhook_secret | Webhook signing secret |
| stripe_currency | Currency code (default: usd) |

### 9.4 USDT Integration

**API Endpoints:**
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/usdt/config | Get USDT config | No |
| POST | /api/usdt/create | Create USDT payment | Yes |
| GET | /api/usdt/status/:payment_id | Get payment status | Yes |
| POST | /usdt/webhook | USDT Webhook callback | No |
| POST | /api/admin/usdt/test | Test USDT connection | Admin |
| POST | /api/admin/usdt/confirm | Manual confirm payment | Admin |

**Configuration:**
| Field | Description |
|-------|-------------|
| usdt_enabled | Enable USDT payment |
| usdt_network | Network type: TRC20, ERC20, BEP20 |
| usdt_wallet_address | Receiving wallet address |
| usdt_api_provider | API provider: nowpayments, coingate, manual |
| usdt_api_key | API key |
| usdt_api_secret | API secret (some providers) |
| usdt_webhook_secret | Webhook signing secret |
| usdt_exchange_rate | Exchange rate (manual mode) |
| usdt_min_amount | Minimum payment amount (USDT) |
| usdt_confirmations | Required confirmations |

**Supported API Providers:**
- **NOWPayments**: Auto-create payment address, auto-confirm
- **CoinGate**: Auto-create payment address, auto-confirm
- **Manual**: Display fixed wallet address, admin manual confirm

---

## 10. Support System

### 10.1 Features

- **Ticket System**: Async problem handling
- **Live Chat**: Real-time customer service
- **Staff Backend**: Multi-staff support
- **Guest Support**: Token-based guest access
- **WebSocket**: Real-time message push

### 10.2 Ticket Status

| Code | Description |
|------|-------------|
| 0 | Pending |
| 1 | Processing |
| 2 | Replied |
| 3 | Resolved |
| 4 | Closed |
| 5 | Merged |

### 10.3 WebSocket Message Types

| Type | Description |
|------|-------------|
| connected | Connection successful |
| new_message | New message |
| typing | Typing indicator |
| read | Read receipt |
| status_change | Ticket status change |
| assigned | Ticket assigned |
| new_ticket | New ticket notification |
| staff_online | Staff online |
| staff_offline | Staff offline |

### 10.4 Satisfaction Rating System

#### 10.4.1 Overview

The satisfaction rating system allows users to rate the service after a ticket is closed or a chat session ends.

#### 10.4.2 Data Model Extensions

**SupportTicket New Fields**:
| Field | Type | Description |
|-------|------|-------------|
| rating | INT | Satisfaction rating (1-5 stars) |
| rating_comment | TEXT | Rating comment |
| rated_at | DATETIME | Rating time |

#### 10.4.3 Rating APIs

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/support/ticket/:ticket_no/rate | Rate ticket | Optional |
| GET | /api/support/ticket/:ticket_no/rating | Get ticket rating | Optional |
| POST | /api/chat/:session_id/rate | Rate chat | Optional |
| GET | /api/admin/support/rating/stats | Get rating statistics | Admin |
| GET | /api/admin/support/rating/staff/:staff_id | Get staff rating stats | Admin |
| GET | /api/staff/rating/stats | Get own rating stats | Staff |

---

## 10.5 FAQ System

### 10.5.1 Overview

The FAQ system provides frequently asked questions and answers to help users solve problems independently.

### 10.5.2 Data Models

**FAQCategory**:
```go
type FAQCategory struct {
    ID        uint      // Primary key
    Name      string    // Category name
    Icon      string    // Category icon
    SortOrder int       // Sort order
    Status    int       // Status: 1=active, 0=disabled
}
```

**FAQ**:
```go
type FAQ struct {
    ID         uint      // Primary key
    CategoryID uint      // Category ID
    Question   string    // Question
    Answer     string    // Answer
    SortOrder  int       // Sort order
    ViewCount  int       // View count
    HelpfulYes int       // Helpful count
    HelpfulNo  int       // Not helpful count
    Status     int       // Status: 1=active, 0=disabled
}
```

**FAQFeedback**:
```go
type FAQFeedback struct {
    ID        uint      // Primary key
    FAQID     uint      // FAQ ID
    UserID    uint      // User ID (0 = guest)
    SessionID string    // Guest session ID
    Helpful   bool      // Is helpful
}
```

### 10.5.3 FAQ APIs

**User APIs**:
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/faq/categories | Get FAQ categories |
| GET | /api/faq/list | Get FAQ list |
| GET | /api/faq/detail/:id | Get FAQ detail |
| GET | /api/faq/search | Search FAQs |
| GET | /api/faq/hot | Get hot FAQs |
| POST | /api/faq/feedback/:id | Submit feedback |

**Admin APIs**:
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/admin/faq/categories | Get all categories |
| POST | /api/admin/faq/category | Create category |
| PUT | /api/admin/faq/category/:id | Update category |
| DELETE | /api/admin/faq/category/:id | Delete category |
| GET | /api/admin/faqs | Get FAQ list (paginated) |
| POST | /api/admin/faq | Create FAQ |
| PUT | /api/admin/faq/:id | Update FAQ |
| DELETE | /api/admin/faq/:id | Delete FAQ |

---

## 11. Frontend Architecture

### 11.1 Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Framework | Next.js | 14.2.x |
| UI Library | React | 18.3.x |
| State Management | Zustand | 5.x |
| Styling | Tailwind CSS | 3.4.x |
| Animation | Framer Motion | 11.x |
| Language | TypeScript | 5.x |

### 11.2 Directory Structure

```
web/
├── src/
│   ├── app/                    # Next.js App Router pages
│   │   ├── layout.tsx          # Root layout
│   │   ├── page.tsx            # Home page
│   │   ├── login/              # Login page
│   │   ├── register/           # Registration page
│   │   ├── products/           # Product list
│   │   ├── user/               # User center
│   │   ├── message/            # Support page
│   │   ├── staff/              # Staff backend
│   │   └── admin/              # Admin backend
│   ├── components/
│   │   ├── ui/                 # Base UI components
│   │   └── layout/             # Layout components
│   ├── hooks/
│   │   └── useWebSocket.ts     # WebSocket hooks
│   └── lib/
│       ├── api.ts              # API wrapper
│       ├── store.ts            # Zustand state
│       ├── websocket.ts        # WebSocket client
│       └── utils.ts            # Utilities
├── package.json
├── next.config.js              # Static export config
└── tailwind.config.ts
```

---

## 12. Deployment Guide

### 12.1 Build

```powershell
# Windows
.\build_windows.ps1

# Linux (PowerShell)
.\build_linux.ps1
```

### 12.2 Run

```bash
# Direct run
./user

# Background (Linux)
nohup ./user > user.log 2>&1 &
```

### 12.3 Access URLs

- User Frontend: `http://localhost:8080/`
- Admin Backend: `http://localhost:8080/manage`
- Staff Backend: `http://localhost:8080/staff`

### 12.4 Default Credentials

- Admin Username: `admin`
- Admin Password: `admin123`

**Note**: Change default password immediately after deployment!

---

## 13. Changelog

### v3.9 (2025-12-28)
- **Ticket Advanced Features Frontend Implementation**:
  - Staff dashboard: Added ticket transfer feature with online staff selection and transfer reason
  - Staff dashboard: Added ticket merge feature with multi-select and target ticket selection
  - Staff dashboard: Added attachment upload feature for images and files
  - User ticket detail page: Added attachment upload feature
  - Added `formatFileSize` utility function for file size formatting
- **Modified Files**:
  - `web/src/app/staff/components/types.ts`: Added OnlineStaff, TicketAttachment types
  - `web/src/app/staff/components/TicketsPanel.tsx`: Added merge mode and merge modal
  - `web/src/app/staff/components/StaffTicketModal.tsx`: Added transfer modal and attachment upload
  - `web/src/app/message/ticket/detail/page.tsx`: User-side attachment upload
  - `web/src/lib/utils.ts`: Added formatFileSize function

### v3.8 (2025-12-26)
- **Permission System Enhancement**:
  - Backend `/api/admin/info` now returns user permissions array and `is_super_admin` flag
  - Frontend permission context (PermissionContext) for centralized permission management
  - Dynamic sidebar menu filtering based on user permissions
  - PermissionGuard component for button-level permission control
  - Features without permission are automatically hidden
  - Permission templates for quick role assignment (5 presets: viewer, support, operator, admin, super_admin)
  - Template selector in admin creation dialog for quick permission assignment
- **Permission Templates**:
  | Template | Description | Permissions |
  |----------|-------------|-------------|
  | super_admin | Super Administrator | All permissions |
  | admin | Administrator | Basic management |
  | operator | Operator | Products, content, coupons |
  | support | Support Staff | Order view, ticket reply |
  | viewer | Read-only | All view permissions |
- **New Files**:
  - `web/src/contexts/PermissionContext.tsx`: Permission context and guard component
- **Modified Files**:
  - `internal/api/admin_handler.go`: AdminInfo returns permissions
  - `internal/api/role_handler.go`: AdminGetPermissions returns templates
  - `internal/model/admin_role.go`: PermissionTemplates definition
  - `web/src/components/admin/types.ts`: PAGE_CONFIG with permissions
  - `web/src/app/admin/page.tsx`: Menu filtering by permissions
  - `web/src/components/admin/Roles.tsx`: Template selector UI

### v3.7 (2025-12-26)
- **Security Enhancement - Admin Username Uniqueness Check**:
  - Added check for system config account username conflict when creating multi-admin accounts
  - Added check for admins table username conflict when modifying system config account username
  - Prevents username collision between dual admin account systems
- **Modified Files**:
  - `internal/service/role_service.go`: CreateAdmin function now checks system config account username
  - `internal/api/admin_handler.go`: AdminSaveSecuritySettings function now checks admins table for conflicts

### v3.6 (2025-12-26)
- **Operation Log File Storage**:
  - Changed operation log storage from database to encrypted CSV files
  - Log files stored in `server_log/YYYY-MM-DD.csv` format
  - Each field encrypted with AES-256-GCM using database encryption key
  - Added date picker for viewing logs by date
  - New API endpoint: `GET /api/admin/logs/dates` for available log dates
  - Improved security: logs are encrypted at rest
  - Better performance: no database overhead for logging
- **Modified Files**:
  - `internal/service/log_service.go`: Rewritten for file-based storage
  - `internal/api/admin_extra_handler.go`: Added date parameter support
  - `internal/api/handler.go`: Added new route for log dates
  - `internal/model/db.go`: Removed OperationLog from auto-migration
  - `internal/repository/repository.go`: Removed database log functions
  - `web/src/components/admin/Logs.tsx`: Added date selector UI

### v3.5 (2025-12-25)
- **FAQ System**:
  - FAQ category management (FAQCategory model)
  - FAQ question management (FAQ model)
  - FAQ feedback feature (FAQFeedback model)
  - Category filtering, keyword search, hot questions
  - View count tracking and helpful/not helpful feedback
  - User FAQ page (/faq)
  - Admin FAQ management
- **Satisfaction Rating System**:
  - Ticket rating (1-5 stars + comment)
  - Chat rating (1-5 stars + comment)
  - Rating statistics (star distribution, average, satisfaction rate)
  - Staff personal rating statistics
  - Rating component in ticket detail page
- **Order Detail Page**:
  - Independent order detail page (/order/detail?order_no=xxx)
  - Display complete order info, product info, kami info
  - Copy kami functionality
- **Product Search**:
  - Search box on product list page
  - Search by name and description
  - Sort by price and sales
- **New Backend Files**:
  - `internal/model/faq.go` - FAQ data models
  - `internal/service/faq_service.go` - FAQ business logic
  - `internal/api/faq_handler.go` - FAQ API handlers
  - `internal/api/rating_handler.go` - Rating API handlers
- **New Frontend Pages**:
  - `web/src/app/faq/page.tsx` - FAQ page
  - `web/src/app/order/detail/page.tsx` - Order detail page
  - `web/src/app/message/ticket/detail/page.tsx` - Ticket detail with rating
- **Data Model Changes**:
  - `SupportTicket` added rating, rating_comment, rated_at fields

### v3.4 (2025-12-17)
  - Added ProductType field to Product model (1=manual kami, default mode)
  - New ManualKami model for storing admin-imported kami codes
  - Batch import kami codes (one per line, auto deduplication)
  - Auto update product stock based on available kami count
  - Auto allocate kami from pool when order completes
  - Local order number generation, no external dependency
- **New Files**:
  - `model/manual_kami.go`: Manual kami data model
  - `service/manual_kami_service.go`: Manual kami business logic
  - `api/manual_kami_handler.go`: Manual kami API handlers
- **New API Endpoints**:
  - `POST /api/admin/product/:id/kami/import`: Import kami codes
  - `GET /api/admin/product/:id/kami`: Get kami list
  - `GET /api/admin/product/:id/kami/stats`: Get kami statistics
  - `DELETE /api/admin/kami/:id`: Delete kami
  - `POST /api/admin/kami/:id/disable`: Disable kami
  - `POST /api/admin/kami/:id/enable`: Enable kami
  - `POST /api/admin/kami/batch-delete`: Batch delete kami

## 21. Role-Based Access Control

### 21.1 Overview

Multi-admin support with role-based permission management.

### 21.2 Data Models

#### AdminRole

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| name | string | Role name (unique) |
| description | string | Role description |
| permissions | text | JSON permission list |
| is_system | bool | System built-in role |
| status | int | Status: 1=enabled 0=disabled |

#### Admin

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| username | string | Username (unique) |
| password_hash | string | Password hash |
| role_id | uint | Role ID |
| email | string | Email |
| nickname | string | Nickname |
| enable_2fa | bool | 2FA enabled |
| status | int | Status: 1=enabled 0=disabled |

### 21.3 Predefined Roles

| Role | Description | Permissions |
|------|-------------|-------------|
| super_admin | Super Administrator | All permissions |
| admin | Administrator | Basic management |
| support | Support Staff | Support related |
| operator | Operator | Products and campaigns |

### 21.4 API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/admin/roles | Get role list | Admin |
| POST | /api/admin/role | Create role | Admin |
| PUT | /api/admin/role/:id | Update role | Admin |
| DELETE | /api/admin/role/:id | Delete role | Admin |
| GET | /api/admin/permissions | Get permissions | Admin |
| GET | /api/admin/admins | Get admin list | Admin |
| POST | /api/admin/admin | Create admin | Admin |
| PUT | /api/admin/admin/:id | Update admin | Admin |
| DELETE | /api/admin/admin/:id | Delete admin | Admin |

## 22. Balance System

### 22.1 Overview

User balance system for recharge and payment.

### 22.2 Data Models

#### UserBalance

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| user_id | uint | User ID (unique) |
| balance | decimal | Available balance |
| frozen | decimal | Frozen amount |
| total_in | decimal | Total recharged |
| total_out | decimal | Total consumed |

#### BalanceLog

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| user_id | uint | User ID |
| type | string | Type: recharge/consume/refund/gift/adjust |
| amount | decimal | Change amount |
| before_balance | decimal | Balance before |
| after_balance | decimal | Balance after |

### 22.3 API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/user/balance | Get my balance | User |
| GET | /api/user/balance/logs | Get balance logs | User |
| POST | /api/user/balance/recharge | Create recharge order | User |
| GET | /api/admin/balances | Get user balances | Admin |
| POST | /api/admin/balance/adjust | Adjust balance | Admin |
| POST | /api/admin/balance/gift | Gift balance | Admin |

## 23. Product Multi-Image System

### 23.1 Overview

Products support multiple images with primary image and sorting.

### 23.2 Data Model

#### ProductImage

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| product_id | uint | Product ID |
| url | string | Image URL |
| sort_order | int | Sort order |
| is_primary | bool | Is primary image |

### 23.3 API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/product/:id/images | Get product images | Public |
| POST | /api/admin/product/:id/images | Upload image | Admin |
| DELETE | /api/admin/product/:id/image/:image_id | Delete image | Admin |
| POST | /api/admin/product/:id/images/primary | Set primary | Admin |

## 24. Shopping Cart System

### 24.1 Overview

Shopping cart for batch purchasing.

### 24.2 Data Model

#### CartItem

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| user_id | uint | User ID |
| product_id | uint | Product ID |
| quantity | int | Quantity |

### 24.3 API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/user/cart | Get cart | User |
| POST | /api/user/cart | Add to cart | User |
| PUT | /api/user/cart/:id | Update quantity | User |
| DELETE | /api/user/cart/:id | Remove item | User |
| DELETE | /api/user/cart | Clear cart | User |

## 25. Product Favorites System

### 25.1 Overview

Users can favorite products for later viewing and purchasing.

### 25.2 Data Model

#### ProductFavorite

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| user_id | uint | User ID |
| product_id | uint | Product ID |
| created_at | time | Favorite time |

### 25.3 API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/user/favorites | Get favorites list | User |
| POST | /api/user/favorite | Add favorite | User |
| DELETE | /api/user/favorite/:product_id | Remove favorite | User |
| GET | /api/user/favorite/:product_id/check | Check favorite status | User |
| GET | /api/user/favorites/count | Get favorites count | User |

## 26. Points System

### 26.1 Overview

Users earn points from purchases, which can be exchanged for coupons or products.

### 26.2 Data Model

#### UserPoints

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| user_id | uint | User ID (unique) |
| points | int | Current points |
| total_earn | int | Total earned |
| total_used | int | Total used |

#### PointsLog

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| user_id | uint | User ID |
| type | string | Type: earn/use/admin/expire |
| points | int | Points change |
| balance | int | Balance after change |
| order_no | string | Related order |
| remark | string | Remark |

#### PointsRule

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| name | string | Rule name |
| type | string | Type: order/register/daily/invite |
| points | int | Fixed points |
| ratio | float64 | Points ratio |
| min_amount | float64 | Minimum amount |
| max_points | int | Maximum points |
| status | int | Status: 1=enabled 0=disabled |

### 26.3 API Endpoints

#### User

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/user/points | Get my points | User |
| GET | /api/user/points/logs | Get points logs | User |
| GET | /api/user/points/exchange/list | Get exchange list | User |
| POST | /api/user/points/exchange/coupon | Exchange coupon | User |
| GET | /api/user/points/exchanges | Get exchange records | User |

#### Admin

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/admin/points/rules | Get points rules | Admin |
| POST | /api/admin/points/rule | Create rule | Admin |
| PUT | /api/admin/points/rule/:id | Update rule | Admin |
| DELETE | /api/admin/points/rule/:id | Delete rule | Admin |
| POST | /api/admin/points/adjust | Adjust user points | Admin |

## 27. Scheduled Tasks System

### 27.1 Overview

Scheduled task system for automatic cleanup, report generation, etc.

### 27.2 Data Model

#### ScheduledTask

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| name | string | Task name |
| type | string | Task type |
| cron_expr | string | Cron expression |
| config | string | Task config (JSON) |
| status | int | Status: 1=enabled 0=disabled |
| last_run_at | time | Last run time |
| next_run_at | time | Next run time |
| run_count | int | Run count |
| fail_count | int | Fail count |

#### TaskLog

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| task_id | uint | Task ID |
| task_name | string | Task name |
| status | string | Status: success/failed |
| duration | int | Duration (ms) |
| result | string | Result |
| error | string | Error message |

### 27.3 Built-in Task Types

| Type | Name | Description |
|------|------|-------------|
| clean_expired_orders | Clean Expired Orders | Cancel unpaid orders |
| clean_expired_sessions | Clean Expired Sessions | Clean inactive devices |
| clean_old_logs | Clean Old Logs | Clean old log records |
| send_daily_report | Daily Report | Send daily sales report |
| backup_database | Database Backup | Scheduled backup |

### 27.4 API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/admin/tasks | Get tasks | Admin |
| GET | /api/admin/tasks/types | Get task types | Admin |
| GET | /api/admin/tasks/stats | Get task stats | Admin |
| GET | /api/admin/tasks/logs | Get task logs | Admin |
| POST | /api/admin/task | Create task | Admin |
| PUT | /api/admin/task/:id | Update task | Admin |
| DELETE | /api/admin/task/:id | Delete task | Admin |
| POST | /api/admin/task/:id/run | Run now | Admin |
| POST | /api/admin/task/:id/toggle | Toggle status | Admin |

## 28. Invoice System

### 28.1 Overview

Electronic invoice application, issuance and management, supporting personal and enterprise invoices.

### 28.2 Data Models

#### Invoice

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| invoice_no | string | Invoice number |
| user_id | uint | User ID |
| order_no | string | Related order number |
| type | string | Type: personal/enterprise |
| title | string | Invoice title |
| tax_no | string | Tax number (enterprise) |
| amount | float64 | Invoice amount |
| email | string | Receiving email |
| status | int | Status: 0=pending 1=issued 2=rejected 3=cancelled |
| invoice_url | string | Electronic invoice URL |

#### InvoiceTitle

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| user_id | uint | User ID |
| type | string | Type |
| title | string | Title name |
| tax_no | string | Tax number |
| is_default | bool | Is default |

### 28.3 API Endpoints

#### User

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/invoice/config | Get invoice config | Public |
| GET | /api/user/invoices | Get invoice list | User |
| GET | /api/user/invoice/:invoice_no | Get invoice detail | User |
| POST | /api/user/invoice | Apply for invoice | User |
| POST | /api/user/invoice/:invoice_no/cancel | Cancel application | User |
| GET | /api/user/invoice/titles | Get title list | User |
| POST | /api/user/invoice/title | Save title | User |
| DELETE | /api/user/invoice/title/:id | Delete title | User |

#### Admin

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/admin/invoices | Get invoice list | Admin |
| GET | /api/admin/invoice/config | Get config | Admin |
| POST | /api/admin/invoice/config | Save config | Admin |
| GET | /api/admin/invoice/stats | Get statistics | Admin |
| POST | /api/admin/invoice/:invoice_no/issue | Issue invoice | Admin |
| POST | /api/admin/invoice/:invoice_no/reject | Reject application | Admin |

## 29. Operation Undo System

### 29.1 Overview

Support undoing important operations like accidental product deletion, user disabling, etc. Configurable retention period.

### 29.2 Data Models

#### UndoOperation

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| operation_type | string | Operation type |
| target_type | string | Target type |
| target_id | uint | Target ID |
| target_name | string | Target name |
| original_data | string | Original data (JSON) |
| admin_name | string | Admin who performed |
| status | int | Status: 0=undoable 1=undone 2=expired |
| expire_at | time | Expiration time |

### 29.3 Supported Operation Types

| Type | Description |
|------|-------------|
| product_delete | Delete product |
| product_disable | Disable product |
| user_disable | Disable user |
| coupon_delete | Delete coupon |
| coupon_disable | Disable coupon |
| category_delete | Delete category |
| announcement_delete | Delete announcement |

### 29.4 API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/admin/undo/operations | Get undoable operations | Admin |
| GET | /api/admin/undo/all | Get all operation records | Admin |
| POST | /api/admin/undo/:id | Undo operation | Admin |
| GET | /api/admin/undo/config | Get config | Admin |
| POST | /api/admin/undo/config | Save config | Admin |
| GET | /api/admin/undo/stats | Get statistics | Admin |

## 30. Smart Customer Service System

### 30.1 Overview

Keyword-based auto-reply system with support for transfer to human agents and auto-reply logging.

### 30.2 Data Models

#### AutoReplyRule

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| name | string | Rule name |
| keywords | string | Keywords (JSON array) |
| match_type | string | Match type: contains/exact/regex |
| reply | string | Reply content |
| priority | int | Priority |
| status | int | Status |
| hit_count | int | Hit count |

#### AutoReplyConfig

| Field | Type | Description |
|-------|------|-------------|
| enabled | bool | Is enabled |
| welcome_message | string | Welcome message |
| no_match_reply | string | No match reply |
| transfer_keywords | string | Transfer to human keywords |
| transfer_message | string | Transfer prompt message |

### 30.3 API Endpoints

#### Public

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/bot/welcome | Get welcome message | Public |
| POST | /api/bot/message | Send message | Optional |

#### Admin

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/admin/bot/config | Get config | Admin |
| POST | /api/admin/bot/config | Save config | Admin |
| GET | /api/admin/bot/rules | Get rule list | Admin |
| POST | /api/admin/bot/rule | Create rule | Admin |
| PUT | /api/admin/bot/rule/:id | Update rule | Admin |
| DELETE | /api/admin/bot/rule/:id | Delete rule | Admin |
| GET | /api/admin/bot/logs | Get logs | Admin |
| GET | /api/admin/bot/stats | Get statistics | Admin |

---

## 31. Email Configuration

### 31.1 Email Config Model

```go
type EmailConfigDB struct {
    ID           uint      // Primary key
    Enabled      bool      // Enable email service
    SMTPHost     string    // SMTP server address
    SMTPPort     int       // SMTP port
    SMTPUser     string    // SMTP username
    SMTPPassword string    // SMTP password (encrypted)
    FromName     string    // Sender name
    FromEmail    string    // Sender email
    Encryption   string    // Encryption: none/ssl/starttls
    CodeLength   int       // Verification code length (4-8)
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 31.2 Encryption Methods

| Method | Description | Recommended Port | Use Case |
|--------|-------------|------------------|----------|
| ssl | SSL/TLS encryption | 465 | QQ Mail, 163 Mail, Aliyun Enterprise |
| starttls | STARTTLS encryption | 587 | Gmail, Outlook, Office365 |
| none | No encryption | 25 | Internal mail servers (not recommended) |

### 31.3 Common Email Configurations

| Service | SMTP Server | Port | Encryption | Notes |
|---------|-------------|------|------------|-------|
| QQ Mail | smtp.qq.com | 465 | SSL | Use authorization code |
| 163 Mail | smtp.163.com | 465 | SSL | Use authorization code |
| Gmail | smtp.gmail.com | 587 | STARTTLS | Use app password |
| Outlook | smtp.office365.com | 587 | STARTTLS | - |
| Aliyun Enterprise | smtp.qiye.aliyun.com | 465 | SSL | - |

### 31.4 API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/admin/email/config | Get email config |
| POST | /api/admin/email/config | Save email config |
| POST | /api/admin/email/test | Send test email |

### 31.5 Request Parameters

```json
{
    "enabled": true,
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "smtp_user": "your@gmail.com",
    "smtp_password": "app_password",
    "from_name": "System Notification",
    "from_email": "your@gmail.com",
    "encryption": "starttls",
    "code_length": 6
}
```

## 32. Multi-Admin System

### 32.1 Dual Admin Table Architecture

The system supports two admin storage methods:

| Table | Description | Purpose |
|-------|-------------|---------|
| system_configs | System config table | Stores default admin (admin_username/admin_password) |
| admins | Multi-admin table | Stores role-based multi-admin accounts |

### 32.2 Admin Model

```go
type Admin struct {
    ID           uint       // Primary key
    Username     string     // Username (unique)
    PasswordHash string     // Password hash
    Nickname     string     // Nickname
    Email        string     // Email
    RoleID       uint       // Role ID
    Status       int        // Status: 1=enabled 0=disabled
    LastLoginAt  *time.Time // Last login time
    LastLoginIP  string     // Last login IP
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 32.3 Role Model

```go
type AdminRole struct {
    ID          uint      // Primary key
    Name        string    // Role name (unique)
    Description string    // Role description
    Permissions string    // Permission list (JSON array)
    IsSystem    bool      // System role (cannot delete)
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 32.4 Permission List

| Permission | Description |
|------------|-------------|
| dashboard:view | View dashboard |
| product:view/create/edit/delete | Product management |
| category:view/create/edit/delete | Category management |
| order:view/edit/refund | Order management |
| user:view/edit/delete | User management |
| admin:view/create/edit/delete | Admin management |
| role:view/create/edit/delete | Role management |
| coupon:view/create/edit/delete | Coupon management |
| announcement:view/create/edit/delete | Announcement management |
| support:view/manage | Support management |
| settings:view/edit/payment/email/database | System settings |
| log:view | View logs |
| backup:view/create/delete | Backup management |

### 32.5 Permission Templates

| Template | Description | Scope |
|----------|-------------|-------|
| super_admin | Super Administrator | All permissions |
| admin | Administrator | All except role/admin management |
| operator | Operator | Products, orders, coupons, announcements |
| support | Support Staff | Order view, user view, support management |
| readonly | Read-only User | All view permissions |

### 32.6 Login Authentication Flow

When admin logs in, the system validates in this order:

1. **Check admins table first**: Find admin account matching username
2. **Fallback to system config**: If not in admins table, check default admin in system_configs
