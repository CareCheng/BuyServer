'use client'

import { useState, useEffect, useCallback } from 'react'
import { getLocale, setLocale, t, getTranslations, Locale, locales, localeNames } from '@/lib/i18n'

/**
 * 国际化 Hook
 * 提供翻译函数和语言切换功能
 */
export function useI18n() {
  const [locale, setLocaleState] = useState<Locale>('zh')
  const [isClient, setIsClient] = useState(false)

  // 客户端初始化
  useEffect(() => {
    setIsClient(true)
    setLocaleState(getLocale())
  }, [])

  // 翻译函数
  const translate = useCallback((key: string): string => {
    return t(key, locale)
  }, [locale])

  // 切换语言
  const changeLocale = useCallback((newLocale: Locale) => {
    setLocale(newLocale)
  }, [])

  // 获取翻译字典
  const translations = getTranslations(locale)

  return {
    locale,
    locales,
    localeNames,
    t: translate,
    setLocale: changeLocale,
    translations,
    isClient,
  }
}

export default useI18n
