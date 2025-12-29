'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { apiGet, apiPost } from '@/lib/api'
import { useAppStore } from '@/lib/store'
import { useTheme } from '@/lib/theme'
import { LanguageSwitcher } from '@/components/LanguageSwitcher'
import { useI18n } from '@/hooks/useI18n'

/**
 * 导航栏组件
 */
export function Navbar() {
  const pathname = usePathname()
  const { user, setUser, isLoggedIn, setIsLoggedIn } = useAppStore()
  const { theme, toggleTheme } = useTheme()
  const { t, locale } = useI18n()
  const [loading, setLoading] = useState(true)
  const [cartCount, setCartCount] = useState(0)

  // 加载用户信息
  useEffect(() => {
    const loadUser = async () => {
      const res = await apiGet<{ user: typeof user }>('/api/user/info')
      if (res.success && res.user) {
        setUser(res.user)
        setIsLoggedIn(true)
        // 加载购物车数量
        loadCartCount()
      } else {
        setUser(null)
        setIsLoggedIn(false)
      }
      setLoading(false)
    }
    loadUser()
  }, [setUser, setIsLoggedIn])

  // 加载购物车数量
  const loadCartCount = async () => {
    const res = await apiGet<{ count: number }>('/api/user/cart/count')
    if (res.success) {
      setCartCount(res.count || 0)
    }
  }

  // 退出登录
  const handleLogout = async () => {
    await apiPost('/api/user/logout')
    setUser(null)
    setIsLoggedIn(false)
    window.location.href = '/'
  }

  const navLinks = [
    { href: '/', label: t('nav.home'), icon: 'fa-home' },
    { href: '/products/', label: t('nav.products'), icon: 'fa-box' },
    { href: '/faq/', label: t('nav.faq'), icon: 'fa-circle-question' },
    { href: '/message/', label: t('nav.support'), icon: 'fa-headset' },
  ]

  return (
    <nav className="navbar">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-2">
            <span className="text-2xl">🔐</span>
            <span className="text-lg font-bold" style={{ color: 'var(--text-primary)' }}>
              {locale === 'en' ? 'License Store' : '卡密购买系统'}
            </span>
          </Link>

          {/* 导航链接 */}
          <div className="flex items-center gap-4 sm:gap-6">
            {navLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className={cn(
                  'text-sm font-medium transition-colors duration-200 flex items-center gap-1.5',
                  pathname === link.href || pathname === link.href.slice(0, -1)
                    ? 'text-primary-400'
                    : 'text-dark-400 hover:text-dark-200'
                )}
              >
                <i className={`fas ${link.icon} text-base`} />
                <span className="hidden sm:inline">{link.label}</span>
              </Link>
            ))}

            {/* 语言切换 */}
            <LanguageSwitcher />

            {/* 购物车图标 */}
            {isLoggedIn && (
              <Link
                href="/user/#cart"
                className="relative p-2 rounded-lg transition-all duration-200 hover:bg-primary-500/10 group"
                title={locale === 'en' ? 'Shopping Cart' : '购物车'}
              >
                <i className="fas fa-cart-shopping text-lg text-dark-400 group-hover:text-primary-400 transition-colors" />
                {cartCount > 0 && (
                  <span className="absolute -top-1 -right-1 w-5 h-5 bg-primary-500 text-white text-xs rounded-full flex items-center justify-center font-medium">
                    {cartCount > 99 ? '99+' : cartCount}
                  </span>
                )}
              </Link>
            )}

            {/* 主题切换按钮 */}
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg transition-all duration-200 hover:bg-primary-500/10 group"
              title={theme === 'dark' ? (locale === 'en' ? 'Switch to Light Theme' : '切换到浅色主题') : (locale === 'en' ? 'Switch to Dark Theme' : '切换到深色主题')}
            >
              {theme === 'dark' ? (
                <i className="fas fa-sun text-lg text-amber-400 group-hover:text-amber-300 transition-colors" />
              ) : (
                <i className="fas fa-moon text-lg text-indigo-400 group-hover:text-indigo-300 transition-colors" />
              )}
            </button>

            {/* 用户菜单 */}
            {loading ? (
              <div className="w-20 h-8 rounded-lg animate-pulse" style={{ background: 'var(--bg-tertiary)' }} />
            ) : isLoggedIn && user ? (
              <div className="flex items-center gap-4">
                <Link
                  href="/user/"
                  className="flex items-center gap-2 text-sm hover:text-primary-400 transition-colors"
                  style={{ color: 'var(--text-secondary)' }}
                >
                  <i className="fas fa-user-circle text-lg" />
                  <span className="hidden sm:inline">{user.username}</span>
                </Link>
                <button
                  onClick={handleLogout}
                  className="flex items-center gap-1.5 text-sm hover:text-red-400 transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                >
                  <i className="fas fa-right-from-bracket" />
                  <span className="hidden sm:inline">{t('nav.logout')}</span>
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-3">
                <Link
                  href="/login/"
                  className="flex items-center gap-1.5 text-sm transition-colors hover:text-primary-400"
                  style={{ color: 'var(--text-muted)' }}
                >
                  <i className="fas fa-right-to-bracket" />
                  <span className="hidden sm:inline">{t('nav.login')}</span>
                </Link>
                <Link
                  href="/register/"
                  className="btn btn-primary btn-sm"
                >
                  <i className="fas fa-user-plus mr-1" />
                  <span className="hidden sm:inline">{t('nav.register')}</span>
                </Link>
              </div>
            )}
          </div>
        </div>
      </div>
    </nav>
  )
}

/**
 * 页脚组件
 */
export function Footer() {
  const { t, locale } = useI18n()
  
  return (
    <footer className="py-8 mt-auto" style={{ borderTop: '1px solid var(--border-light)' }}>
      <div className="max-w-7xl mx-auto px-4 text-center text-sm" style={{ color: 'var(--text-muted)' }}>
        <p>
          &copy; {new Date().getFullYear()} {locale === 'en' ? 'License Store' : '卡密购买系统'}. {t('footer.allRightsReserved')}.
        </p>
      </div>
    </footer>
  )
}
