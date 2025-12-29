'use client'

import { useState, useEffect, useRef } from 'react'
import { apiGet } from '@/lib/api'
import { Navbar, Footer } from '@/components/layout'
import { HeroSection } from './HeroSection'
import { FeaturesSection } from './FeaturesSection'
import { AnnouncementSection } from './AnnouncementSection'
import { ProductsSection } from './ProductsSection'
import { StatsSection } from './StatsSection'
import { CTASection } from './CTASection'
import { FloatingButton } from './FloatingButton'
import type { HomepageConfig } from '@/types/homepage'
import { defaultHomepageConfig } from '@/types/homepage'

/**
 * 高级模式自定义 HTML 渲染组件
 */
function CustomHTMLRenderer({ config }: { config: HomepageConfig }) {
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!containerRef.current) return

    // 注入自定义 CSS
    if (config.custom_css) {
      const styleId = 'homepage-custom-css'
      let styleEl = document.getElementById(styleId) as HTMLStyleElement
      if (!styleEl) {
        styleEl = document.createElement('style')
        styleEl.id = styleId
        document.head.appendChild(styleEl)
      }
      styleEl.textContent = config.custom_css
    }

    // 注入自定义 JS（在 HTML 渲染后执行）
    if (config.custom_js) {
      try {
        // 使用 Function 构造器执行自定义 JS，提供一些有用的上下文
        const customFn = new Function('container', 'config', config.custom_js)
        customFn(containerRef.current, config)
      } catch (err) {
        console.error('自定义 JS 执行错误:', err)
      }
    }

    // 清理函数
    return () => {
      const styleEl = document.getElementById('homepage-custom-css')
      if (styleEl) {
        styleEl.remove()
      }
    }
  }, [config.custom_css, config.custom_js])

  return (
    <div 
      ref={containerRef}
      className="custom-homepage-content"
      dangerouslySetInnerHTML={{ __html: config.custom_html }}
    />
  )
}

/**
 * 动态首页组件
 * 根据后台配置渲染不同的首页样式
 */
export function DynamicHomepage() {
  const [config, setConfig] = useState<HomepageConfig>(defaultHomepageConfig)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadConfig = async () => {
      try {
        const res = await apiGet<{ config: HomepageConfig }>('/api/homepage/config')
        if (res.success && res.config) {
          setConfig(res.config)
        }
      } catch {
        // 使用默认配置
      }
      setLoading(false)
    }
    loadConfig()
  }, [])

  if (loading) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1 flex items-center justify-center">
          <div className="text-center">
            <div className="w-12 h-12 border-4 border-primary-500 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
            <p style={{ color: 'var(--text-muted)' }}>加载中...</p>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  // 高级模式：渲染自定义 HTML
  if (config.advanced_mode && config.custom_html) {
    return (
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1">
          <CustomHTMLRenderer config={config} />
        </main>
        <Footer />
        {/* 高级模式下浮动按钮仍然可用 */}
        <FloatingButton config={config} />
      </div>
    )
  }

  // 普通模式：使用组件化渲染
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />

      <main className="flex-1">
        {/* Hero 区块 */}
        <HeroSection config={config} />

        {/* 公告区块 */}
        <AnnouncementSection config={config} />

        {/* 特性区块 */}
        <FeaturesSection config={config} />

        {/* 商品展示区块 */}
        <ProductsSection config={config} />

        {/* 统计区块 */}
        <StatsSection config={config} />

        {/* CTA 区块 */}
        <CTASection config={config} />
      </main>

      <Footer />

      {/* 浮动按钮 */}
      <FloatingButton config={config} />
    </div>
  )
}

// 导出所有子组件
export { HeroSection } from './HeroSection'
export { FeaturesSection } from './FeaturesSection'
export { AnnouncementSection } from './AnnouncementSection'
export { ProductsSection } from './ProductsSection'
export { StatsSection } from './StatsSection'
export { CTASection } from './CTASection'
export { FloatingButton } from './FloatingButton'
