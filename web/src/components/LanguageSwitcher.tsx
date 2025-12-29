'use client'

import { useState, useEffect, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { getLocale, setLocale, locales, localeNames, Locale } from '@/lib/i18n'

/**
 * 语言切换组件
 * 支持中英文切换
 */
export function LanguageSwitcher() {
  const [currentLocale, setCurrentLocale] = useState<Locale>('zh')
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // 初始化当前语言
  useEffect(() => {
    setCurrentLocale(getLocale())
  }, [])

  // 点击外部关闭下拉菜单
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // 切换语言
  const handleChangeLocale = (locale: Locale) => {
    if (locale !== currentLocale) {
      setLocale(locale)
    }
    setIsOpen(false)
  }

  // 语言图标
  const getLocaleIcon = (locale: Locale) => {
    switch (locale) {
      case 'zh':
        return '🇨🇳'
      case 'en':
        return '🇺🇸'
      default:
        return '🌐'
    }
  }

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 rounded-lg bg-dark-800/50 hover:bg-dark-700/50 border border-dark-700/50 transition-colors"
        aria-label="切换语言"
      >
        <span className="text-lg">{getLocaleIcon(currentLocale)}</span>
        <span className="text-dark-200 text-sm hidden sm:inline">
          {localeNames[currentLocale]}
        </span>
        <i className={`fas fa-chevron-down text-xs text-dark-400 transition-transform ${isOpen ? 'rotate-180' : ''}`} />
      </button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.15 }}
            className="absolute right-0 mt-2 w-36 bg-dark-800 border border-dark-700/50 rounded-xl shadow-xl overflow-hidden z-50"
          >
            {locales.map((locale) => (
              <button
                key={locale}
                onClick={() => handleChangeLocale(locale)}
                className={`w-full flex items-center gap-3 px-4 py-3 text-left transition-colors ${
                  locale === currentLocale
                    ? 'bg-primary-500/20 text-primary-400'
                    : 'text-dark-200 hover:bg-dark-700/50'
                }`}
              >
                <span className="text-lg">{getLocaleIcon(locale)}</span>
                <span className="text-sm">{localeNames[locale]}</span>
                {locale === currentLocale && (
                  <i className="fas fa-check text-xs ml-auto" />
                )}
              </button>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

export default LanguageSwitcher
