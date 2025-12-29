# KamiServer

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/React-18.3-61DAFB?style=flat-square&logo=react" alt="React">
  <img src="https://img.shields.io/badge/Next.js-14.2-000000?style=flat-square&logo=next.js" alt="Next.js">
  <img src="https://img.shields.io/badge/License-GPL%20v3-blue?style=flat-square" alt="License">
</p>

<p align="center">
  A full-featured card key sales management system with multiple payment methods, online customer service, and multi-admin permission management.
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#tech-stack">Tech Stack</a> •
  <a href="#project-structure">Project Structure</a> •
  <a href="#documentation">Documentation</a> •
  <a href="#license">License</a>
</p>

---

## Features

### 🛒 Product System
- Product category management
- Multi-image product display
- Manual card key import and management
- Product favorites

### 💳 Payment Integration
- Alipay (Face-to-face / Web payment)
- WeChat Pay (Native / JSAPI)
- PayPal International Payment
- Stripe Credit Card Payment
- USDT Cryptocurrency Payment
- YiPay Aggregated Payment
- Balance Payment

### 👤 User System
- User registration / login
- Email verification
- Two-factor authentication (TOTP + Email)
- Device management and login alerts
- User balance and points system
- Shopping cart

### 🎫 Order System
- Order creation and management
- Coupon system
- Auto-cancel expired orders
- Public order query
- Invoice management

### 💬 Customer Service
- Ticket system
- WebSocket real-time chat
- Staff workbench
- Guest support
- Smart auto-reply
- Knowledge base
- Satisfaction rating

### 🔧 Admin Panel
- Dashboard statistics
- Multi-admin role permissions
- Homepage customization
- Operation log audit
- Database backup
- System monitoring

### 🔒 Security Features
- Login failure lockout
- IP blacklist
- Tiered rate limiting
- CSRF protection
- Security headers
- bcrypt password hashing
- AES-GCM sensitive data encryption

## Quick Start

### Requirements

- Go 1.21+
- Node.js 18+ (for frontend build)

### Build

```bash
# Windows
.\build.ps1

# Linux
./build.sh

# Embed mode (single file deployment)
.\build.ps1 --embed
```

### Run

```bash
# Windows
.\dist\windows\UserFrontend.exe

# Linux
./dist/linux/UserFrontend
```

### Access URLs

| Page | URL |
|------|-----|
| User Frontend | http://localhost:8080/ |
| Admin Panel | http://localhost:8080/manage |
| Staff Workbench | http://localhost:8080/staff/login |

### Default Account

On first access to the admin panel, the system will guide you to set the administrator password.

- Default username: `admin`
- Password: Set on first startup

> ⚠️ **Security Notice**: Please set a strong password with letters, numbers, and special characters!

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend Framework | Go + Gin |
| ORM | GORM |
| Database | MySQL / PostgreSQL / SQLite |
| Frontend Framework | React + Next.js + TypeScript |
| Styling | Tailwind CSS |
| State Management | Zustand |
| Real-time Communication | WebSocket |
| Authentication | Session + Cookie + TOTP |
| Encryption | bcrypt + AES-GCM |

## Project Structure

```
Server/
├── cmd/server/main.go       # Entry point
├── internal/
│   ├── api/                 # HTTP API handlers
│   ├── config/              # Configuration
│   ├── model/               # Data models
│   ├── repository/          # Data access layer
│   ├── service/             # Business logic
│   └── utils/               # Utilities
├── web/                     # Frontend (React + Next.js)
│   ├── src/
│   │   ├── app/             # Page routes
│   │   ├── components/      # React components
│   │   ├── hooks/           # Custom hooks
│   │   └── lib/             # Utilities
│   └── package.json
├── Product/                 # Product images
├── user_config/             # Runtime config
├── build.ps1                # Windows build script
└── build.sh                 # Linux build script
```

## Payment Configuration

| Payment Method | Description |
|----------------|-------------|
| Alipay | Face-to-face / Web payment, requires app keys |
| WeChat Pay | Native / JSAPI, requires merchant ID and certificates |
| PayPal | International payment, supports sandbox and production |
| Stripe | Credit card payment, supports Webhook |
| USDT | Supports TRC20 / ERC20 / BEP20 networks |
| YiPay | Third-party aggregated payment |

## Documentation

- 📖 [Chinese Technical Documentation](Technical_Documentation_CN.md)
- 📖 [English Technical Documentation](Technical_Documentation_EN.md)

## Contributing

Contributions are welcome! Please feel free to submit Issues and Pull Requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the [GNU General Public License v3.0](LICENSE).

```
KamiServer - Card Key Sales Management System
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

## Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web Framework
- [GORM](https://gorm.io/) - ORM Library
- [Next.js](https://nextjs.org/) - React Framework
- [Tailwind CSS](https://tailwindcss.com/) - CSS Framework

---

<p align="center">
  If this project helps you, please give it a ⭐ Star!
</p>
