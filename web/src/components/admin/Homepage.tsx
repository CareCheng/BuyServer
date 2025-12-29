'use client'

import { useState, useEffect } from 'react'
import toast from 'react-hot-toast'
import { apiGet, apiPost } from '@/lib/api'
import { Button, Modal } from '@/components/ui'
import Toggle from '@/components/common/Toggle'
import type { HomepageConfig, TemplateInfo, FeatureItem, StatItem, FooterLink } from '@/types/homepage'
import { defaultHomepageConfig } from '@/types/homepage'

/**
 * 首页配置管理组件
 */
export function Homepage() {
  const [config, setConfig] = useState<HomepageConfig>(defaultHomepageConfig)
  const [templates, setTemplates] = useState<TemplateInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [activeTab, setActiveTab] = useState('template')
  const [showResetModal, setShowResetModal] = useState(false)

  // 加载配置
  useEffect(() => {
    loadConfig()
    loadTemplates()
  }, [])

  const loadConfig = async () => {
    const res = await apiGet<{ config: HomepageConfig }>('/api/admin/homepage/config')
    if (res.success && res.config) {
      setConfig(res.config)
    }
    setLoading(false)
  }

  const loadTemplates = async () => {
    const res = await apiGet<{ templates: TemplateInfo[] }>('/api/admin/homepage/templates')
    if (res.success && res.templates) {
      setTemplates(res.templates)
    }
  }

  // 保存配置
  const handleSave = async () => {
    setSaving(true)
    const res = await apiPost('/api/admin/homepage/config', config as unknown as Record<string, unknown>)
    setSaving(false)
    if (res.success) {
      toast.success('配置已保存')
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  // 选择模板
  const handleSelectTemplate = async (templateId: string) => {
    const res = await apiGet<{ config: HomepageConfig }>(`/api/admin/homepage/template/default?template=${templateId}`)
    if (res.success && res.config) {
      setConfig(res.config)
      toast.success(`已切换到 ${templates.find(t => t.id === templateId)?.name || templateId} 模板`)
    }
  }

  // 重置配置
  const handleReset = async () => {
    const res = await apiPost('/api/admin/homepage/reset', { template: config.template })
    if (res.success) {
      await loadConfig()
      toast.success('已重置为默认配置')
      setShowResetModal(false)
    } else {
      toast.error(res.error || '重置失败')
    }
  }

  // 更新特性项
  const updateFeature = (index: number, field: keyof FeatureItem, value: string) => {
    const newFeatures = [...config.features]
    newFeatures[index] = { ...newFeatures[index], [field]: value }
    setConfig({ ...config, features: newFeatures })
  }

  // 添加特性项
  const addFeature = () => {
    setConfig({
      ...config,
      features: [...config.features, { icon: '⭐', title: '新特性', description: '特性描述' }],
    })
  }

  // 删除特性项
  const removeFeature = (index: number) => {
    setConfig({
      ...config,
      features: config.features.filter((_, i) => i !== index),
    })
  }

  // 更新统计项
  const updateStat = (index: number, field: keyof StatItem, value: string) => {
    const newStats = [...config.stats]
    newStats[index] = { ...newStats[index], [field]: value }
    setConfig({ ...config, stats: newStats })
  }

  // 添加统计项
  const addStat = () => {
    setConfig({
      ...config,
      stats: [...config.stats, { value: '0', label: '新统计', icon: '📊' }],
    })
  }

  // 删除统计项
  const removeStat = (index: number) => {
    setConfig({
      ...config,
      stats: config.stats.filter((_, i) => i !== index),
    })
  }

  // 更新页脚链接
  const updateFooterLink = (index: number, field: keyof FooterLink, value: string) => {
    const newLinks = [...config.footer_links]
    newLinks[index] = { ...newLinks[index], [field]: value }
    setConfig({ ...config, footer_links: newLinks })
  }

  // 添加页脚链接
  const addFooterLink = () => {
    setConfig({
      ...config,
      footer_links: [...config.footer_links, { text: '新链接', url: '/' }],
    })
  }

  // 删除页脚链接
  const removeFooterLink = (index: number) => {
    setConfig({
      ...config,
      footer_links: config.footer_links.filter((_, i) => i !== index),
    })
  }

  const tabs = [
    { id: 'template', label: '模板选择', icon: 'fa-palette' },
    { id: 'hero', label: 'Hero区块', icon: 'fa-image' },
    { id: 'features', label: '特性区块', icon: 'fa-star' },
    { id: 'announcement', label: '公告区块', icon: 'fa-bullhorn' },
    { id: 'products', label: '商品展示', icon: 'fa-box' },
    { id: 'stats', label: '统计区块', icon: 'fa-chart-bar' },
    { id: 'cta', label: 'CTA区块', icon: 'fa-rocket' },
    { id: 'footer', label: '页脚设置', icon: 'fa-shoe-prints' },
    { id: 'advanced', label: '高级模式', icon: 'fa-code' },
  ]

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="text-center">
          <div className="w-10 h-10 border-4 border-primary-500 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-dark-400">加载配置中...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 标题栏 */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-dark-100">首页配置</h2>
          <p className="text-sm text-dark-400 mt-1">自定义用户端首页的显示内容和样式</p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="secondary" onClick={() => setShowResetModal(true)}>
            <i className="fas fa-undo mr-2" />重置
          </Button>
          <Button variant="primary" onClick={handleSave} loading={saving}>
            <i className="fas fa-save mr-2" />保存配置
          </Button>
        </div>
      </div>

      {/* 标签页导航 */}
      <div className="flex flex-wrap gap-2 p-1 bg-dark-800/50 rounded-xl">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              activeTab === tab.id
                ? 'bg-primary-500 text-white'
                : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
            }`}
          >
            <i className={`fas ${tab.icon}`} />
            {tab.label}
          </button>
        ))}
      </div>

      {/* 配置面板 */}
      <div className="card p-6">
        {/* 模板选择 */}
        {activeTab === 'template' && (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold text-dark-100">选择模板</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {templates.map((template) => (
                <div
                  key={template.id}
                  onClick={() => handleSelectTemplate(template.id)}
                  className={`cursor-pointer rounded-xl border-2 p-4 transition-all hover:shadow-lg ${
                    config.template === template.id
                      ? 'border-primary-500 bg-primary-500/10'
                      : 'border-dark-700 hover:border-dark-600'
                  }`}
                >
                  <div className="h-24 rounded-lg mb-3 flex items-center justify-center text-4xl"
                    style={{ backgroundColor: 'var(--bg-tertiary)' }}
                  >
                    {template.id === 'modern' && '🎨'}
                    {template.id === 'gradient' && '🌈'}
                    {template.id === 'minimal' && '⬜'}
                    {template.id === 'card' && '🃏'}
                    {template.id === 'hero' && '🖼️'}
                    {template.id === 'business' && '💼'}
                  </div>
                  <h4 className="font-semibold text-dark-100">{template.name}</h4>
                  <p className="text-sm text-dark-400 mt-1">{template.description}</p>
                  {config.template === template.id && (
                    <div className="mt-2 text-xs text-primary-400 flex items-center gap-1">
                      <i className="fas fa-check-circle" />当前使用
                    </div>
                  )}
                </div>
              ))}
            </div>

            {/* 颜色设置 */}
            <div className="border-t border-dark-700 pt-6 mt-6">
              <h4 className="font-semibold text-dark-100 mb-4">颜色设置</h4>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-dark-400 mb-2">主色调</label>
                  <div className="flex items-center gap-3">
                    <input
                      type="color"
                      value={config.primary_color}
                      onChange={(e) => setConfig({ ...config, primary_color: e.target.value })}
                      className="w-12 h-12 rounded-lg cursor-pointer"
                    />
                    <input
                      type="text"
                      value={config.primary_color}
                      onChange={(e) => setConfig({ ...config, primary_color: e.target.value })}
                      className="flex-1 px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm text-dark-400 mb-2">次色调</label>
                  <div className="flex items-center gap-3">
                    <input
                      type="color"
                      value={config.secondary_color}
                      onChange={(e) => setConfig({ ...config, secondary_color: e.target.value })}
                      className="w-12 h-12 rounded-lg cursor-pointer"
                    />
                    <input
                      type="text"
                      value={config.secondary_color}
                      onChange={(e) => setConfig({ ...config, secondary_color: e.target.value })}
                      className="flex-1 px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Hero 区块配置 */}
        {activeTab === 'hero' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-dark-100">Hero 区块</h3>
              <Toggle
                checked={config.hero_enabled}
                onChange={(checked) => setConfig({ ...config, hero_enabled: checked })}
                label="启用"
                labelPosition="left"
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm text-dark-400 mb-2">标题</label>
                <input
                  type="text"
                  value={config.hero_title}
                  onChange={(e) => setConfig({ ...config, hero_title: e.target.value })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
              <div>
                <label className="block text-sm text-dark-400 mb-2">副标题</label>
                <input
                  type="text"
                  value={config.hero_subtitle}
                  onChange={(e) => setConfig({ ...config, hero_subtitle: e.target.value })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
              <div>
                <label className="block text-sm text-dark-400 mb-2">按钮文字</label>
                <input
                  type="text"
                  value={config.hero_button_text}
                  onChange={(e) => setConfig({ ...config, hero_button_text: e.target.value })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
              <div>
                <label className="block text-sm text-dark-400 mb-2">按钮链接</label>
                <input
                  type="text"
                  value={config.hero_button_link}
                  onChange={(e) => setConfig({ ...config, hero_button_link: e.target.value })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm text-dark-400 mb-2">背景类型</label>
              <div className="flex gap-4">
                {['gradient', 'image', 'solid'].map((type) => (
                  <label key={type} className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="hero_background"
                      checked={config.hero_background === type}
                      onChange={() => setConfig({ ...config, hero_background: type as HomepageConfig['hero_background'] })}
                      className="text-primary-500 focus:ring-primary-500"
                    />
                    <span className="text-sm text-dark-300">
                      {type === 'gradient' ? '渐变' : type === 'image' ? '图片' : '纯色'}
                    </span>
                  </label>
                ))}
              </div>
            </div>

            {config.hero_background === 'image' && (
              <div>
                <label className="block text-sm text-dark-400 mb-2">背景图片 URL</label>
                <input
                  type="text"
                  value={config.hero_bg_image}
                  onChange={(e) => setConfig({ ...config, hero_bg_image: e.target.value })}
                  placeholder="https://example.com/image.jpg"
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
            )}

            {config.hero_background === 'solid' && (
              <div>
                <label className="block text-sm text-dark-400 mb-2">背景颜色</label>
                <div className="flex items-center gap-3">
                  <input
                    type="color"
                    value={config.hero_bg_color || config.primary_color}
                    onChange={(e) => setConfig({ ...config, hero_bg_color: e.target.value })}
                    className="w-12 h-12 rounded-lg cursor-pointer"
                  />
                  <input
                    type="text"
                    value={config.hero_bg_color}
                    onChange={(e) => setConfig({ ...config, hero_bg_color: e.target.value })}
                    className="flex-1 px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                  />
                </div>
              </div>
            )}
          </div>
        )}

        {/* 特性区块配置 */}
        {activeTab === 'features' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-dark-100">特性区块</h3>
              <Toggle
                checked={config.features_enabled}
                onChange={(checked) => setConfig({ ...config, features_enabled: checked })}
                label="启用"
                labelPosition="left"
              />
            </div>

            <div>
              <label className="block text-sm text-dark-400 mb-2">区块标题</label>
              <input
                type="text"
                value={config.features_title}
                onChange={(e) => setConfig({ ...config, features_title: e.target.value })}
                className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
              />
            </div>

            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <label className="text-sm text-dark-400">特性列表</label>
                <Button variant="secondary" size="sm" onClick={addFeature}>
                  <i className="fas fa-plus mr-1" />添加
                </Button>
              </div>

              {config.features.map((feature, index) => (
                <div key={index} className="p-4 bg-dark-700/30 rounded-lg space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-dark-400">特性 {index + 1}</span>
                    <button
                      onClick={() => removeFeature(index)}
                      className="text-red-400 hover:text-red-300 text-sm"
                    >
                      <i className="fas fa-trash" />
                    </button>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                    <div>
                      <label className="block text-xs text-dark-500 mb-1">图标</label>
                      <input
                        type="text"
                        value={feature.icon}
                        onChange={(e) => updateFeature(index, 'icon', e.target.value)}
                        placeholder="🔒 或 fa-lock"
                        className="w-full px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs text-dark-500 mb-1">标题</label>
                      <input
                        type="text"
                        value={feature.title}
                        onChange={(e) => updateFeature(index, 'title', e.target.value)}
                        className="w-full px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs text-dark-500 mb-1">描述</label>
                      <input
                        type="text"
                        value={feature.description}
                        onChange={(e) => updateFeature(index, 'description', e.target.value)}
                        className="w-full px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 text-sm"
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* 公告区块配置 */}
        {activeTab === 'announcement' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-dark-100">公告区块</h3>
              <Toggle
                checked={config.announcement_enabled}
                onChange={(checked) => setConfig({ ...config, announcement_enabled: checked })}
                label="启用"
                labelPosition="left"
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm text-dark-400 mb-2">公告标题</label>
                <input
                  type="text"
                  value={config.announcement_title}
                  onChange={(e) => setConfig({ ...config, announcement_title: e.target.value })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
              <div>
                <label className="block text-sm text-dark-400 mb-2">公告类型</label>
                <select
                  value={config.announcement_type}
                  onChange={(e) => setConfig({ ...config, announcement_type: e.target.value as HomepageConfig['announcement_type'] })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                >
                  <option value="info">信息（蓝色）</option>
                  <option value="warning">警告（黄色）</option>
                  <option value="success">成功（绿色）</option>
                </select>
              </div>
            </div>

            <div>
              <label className="block text-sm text-dark-400 mb-2">公告内容</label>
              <textarea
                value={config.announcement_content}
                onChange={(e) => setConfig({ ...config, announcement_content: e.target.value })}
                rows={4}
                className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 resize-none"
                placeholder="输入公告内容..."
              />
            </div>
          </div>
        )}

        {/* 商品展示配置 */}
        {activeTab === 'products' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-dark-100">商品展示区块</h3>
              <Toggle
                checked={config.products_enabled}
                onChange={(checked) => setConfig({ ...config, products_enabled: checked })}
                label="启用"
                labelPosition="left"
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm text-dark-400 mb-2">区块标题</label>
                <input
                  type="text"
                  value={config.products_title}
                  onChange={(e) => setConfig({ ...config, products_title: e.target.value })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
              <div>
                <label className="block text-sm text-dark-400 mb-2">展示数量</label>
                <input
                  type="number"
                  min={1}
                  max={12}
                  value={config.products_count}
                  onChange={(e) => setConfig({ ...config, products_count: parseInt(e.target.value) || 6 })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
            </div>

            <div className="p-4 bg-dark-700/30 rounded-lg">
              <p className="text-sm text-dark-400">
                <i className="fas fa-info-circle mr-2 text-primary-400" />
                商品将自动从商品列表中获取，按照排序显示前 {config.products_count} 个商品。
              </p>
            </div>
          </div>
        )}

        {/* 统计区块配置 */}
        {activeTab === 'stats' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-dark-100">统计区块</h3>
              <Toggle
                checked={config.stats_enabled}
                onChange={(checked) => setConfig({ ...config, stats_enabled: checked })}
                label="启用"
                labelPosition="left"
              />
            </div>

            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <label className="text-sm text-dark-400">统计项列表</label>
                <Button variant="secondary" size="sm" onClick={addStat}>
                  <i className="fas fa-plus mr-1" />添加
                </Button>
              </div>

              {config.stats.map((stat, index) => (
                <div key={index} className="p-4 bg-dark-700/30 rounded-lg space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-dark-400">统计项 {index + 1}</span>
                    <button
                      onClick={() => removeStat(index)}
                      className="text-red-400 hover:text-red-300 text-sm"
                    >
                      <i className="fas fa-trash" />
                    </button>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                    <div>
                      <label className="block text-xs text-dark-500 mb-1">图标</label>
                      <input
                        type="text"
                        value={stat.icon}
                        onChange={(e) => updateStat(index, 'icon', e.target.value)}
                        placeholder="👥 或 fa-users"
                        className="w-full px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs text-dark-500 mb-1">数值</label>
                      <input
                        type="text"
                        value={stat.value}
                        onChange={(e) => updateStat(index, 'value', e.target.value)}
                        placeholder="10000+"
                        className="w-full px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs text-dark-500 mb-1">标签</label>
                      <input
                        type="text"
                        value={stat.label}
                        onChange={(e) => updateStat(index, 'label', e.target.value)}
                        placeholder="用户数量"
                        className="w-full px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 text-sm"
                      />
                    </div>
                  </div>
                </div>
              ))}

              {config.stats.length === 0 && (
                <div className="text-center py-8 text-dark-500">
                  <i className="fas fa-chart-bar text-3xl mb-2" />
                  <p>暂无统计项，点击上方按钮添加</p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* CTA 区块配置 */}
        {activeTab === 'cta' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-dark-100">CTA 区块</h3>
              <Toggle
                checked={config.cta_enabled}
                onChange={(checked) => setConfig({ ...config, cta_enabled: checked })}
                label="启用"
                labelPosition="left"
              />
            </div>

            <div className="p-4 bg-dark-700/30 rounded-lg mb-4">
              <p className="text-sm text-dark-400">
                <i className="fas fa-info-circle mr-2 text-primary-400" />
                CTA（Call to Action）区块用于引导用户进行特定操作，如注册、购买等。
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm text-dark-400 mb-2">标题</label>
                <input
                  type="text"
                  value={config.cta_title}
                  onChange={(e) => setConfig({ ...config, cta_title: e.target.value })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
              <div>
                <label className="block text-sm text-dark-400 mb-2">副标题</label>
                <input
                  type="text"
                  value={config.cta_subtitle}
                  onChange={(e) => setConfig({ ...config, cta_subtitle: e.target.value })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
              <div>
                <label className="block text-sm text-dark-400 mb-2">按钮文字</label>
                <input
                  type="text"
                  value={config.cta_button_text}
                  onChange={(e) => setConfig({ ...config, cta_button_text: e.target.value })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
              <div>
                <label className="block text-sm text-dark-400 mb-2">按钮链接</label>
                <input
                  type="text"
                  value={config.cta_button_link}
                  onChange={(e) => setConfig({ ...config, cta_button_link: e.target.value })}
                  className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                />
              </div>
            </div>
          </div>
        )}

        {/* 页脚设置 */}
        {activeTab === 'footer' && (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold text-dark-100">页脚设置</h3>

            <div>
              <label className="block text-sm text-dark-400 mb-2">页脚文字</label>
              <input
                type="text"
                value={config.footer_text}
                onChange={(e) => setConfig({ ...config, footer_text: e.target.value })}
                className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
              />
            </div>

            {/* 页脚链接管理 */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <label className="text-sm text-dark-400">页脚链接</label>
                <Button variant="secondary" size="sm" onClick={addFooterLink}>
                  <i className="fas fa-plus mr-1" />添加链接
                </Button>
              </div>

              {config.footer_links.map((link, index) => (
                <div key={index} className="p-4 bg-dark-700/30 rounded-lg space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-dark-400">链接 {index + 1}</span>
                    <button
                      onClick={() => removeFooterLink(index)}
                      className="text-red-400 hover:text-red-300 text-sm"
                    >
                      <i className="fas fa-trash" />
                    </button>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs text-dark-500 mb-1">链接文字</label>
                      <input
                        type="text"
                        value={link.text}
                        onChange={(e) => updateFooterLink(index, 'text', e.target.value)}
                        className="w-full px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs text-dark-500 mb-1">链接地址</label>
                      <input
                        type="text"
                        value={link.url}
                        onChange={(e) => updateFooterLink(index, 'url', e.target.value)}
                        className="w-full px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 text-sm"
                      />
                    </div>
                  </div>
                </div>
              ))}

              {config.footer_links.length === 0 && (
                <div className="text-center py-8 text-dark-500">
                  <i className="fas fa-link text-3xl mb-2" />
                  <p>暂无页脚链接，点击上方按钮添加</p>
                </div>
              )}
            </div>

            {/* 浮动按钮设置 */}
            <div className="border-t border-dark-700 pt-6 mt-6">
              <div className="flex items-center justify-between mb-4">
                <h4 className="font-semibold text-dark-100">浮动按钮</h4>
                <Toggle
                  checked={config.floating_button_enabled}
                  onChange={(checked) => setConfig({ ...config, floating_button_enabled: checked })}
                  label="启用"
                  labelPosition="left"
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-dark-400 mb-2">图标</label>
                  <input
                    type="text"
                    value={config.floating_button_icon}
                    onChange={(e) => setConfig({ ...config, floating_button_icon: e.target.value })}
                    placeholder="fa-headset"
                    className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                  />
                  <p className="text-xs text-dark-500 mt-1">使用 Font Awesome 图标类名，如 fa-headset</p>
                </div>
                <div>
                  <label className="block text-sm text-dark-400 mb-2">链接地址</label>
                  <input
                    type="text"
                    value={config.floating_button_link}
                    onChange={(e) => setConfig({ ...config, floating_button_link: e.target.value })}
                    className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100"
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {/* 高级模式配置 */}
        {activeTab === 'advanced' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold text-dark-100">高级模式</h3>
                <p className="text-sm text-dark-400 mt-1">使用自定义 HTML/CSS/JS 完全控制首页设计</p>
              </div>
              <Toggle
                checked={config.advanced_mode}
                onChange={(checked) => setConfig({ ...config, advanced_mode: checked })}
                label="启用高级模式"
                labelPosition="left"
              />
            </div>

            {/* 警告提示 */}
            <div className="p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
              <div className="flex items-start gap-3">
                <i className="fas fa-exclamation-triangle text-yellow-500 mt-0.5" />
                <div>
                  <p className="text-sm text-yellow-200 font-medium">注意事项</p>
                  <ul className="text-sm text-yellow-200/80 mt-1 space-y-1 list-disc list-inside">
                    <li>启用高级模式后，上方的模板和区块配置将被忽略</li>
                    <li>自定义代码中的错误可能导致页面无法正常显示</li>
                    <li>请确保代码安全，避免引入恶意脚本</li>
                    <li>建议先在本地测试后再保存</li>
                  </ul>
                </div>
              </div>
            </div>

            {config.advanced_mode && (
              <>
                {/* 自定义 HTML */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <label className="text-sm text-dark-400">自定义 HTML</label>
                    <span className="text-xs text-dark-500">支持完整的 HTML 结构</span>
                  </div>
                  <textarea
                    value={config.custom_html}
                    onChange={(e) => setConfig({ ...config, custom_html: e.target.value })}
                    rows={15}
                    className="w-full px-4 py-3 bg-dark-900 border border-dark-600 rounded-lg text-dark-100 font-mono text-sm resize-y"
                    placeholder={`<!-- 自定义首页 HTML -->
<div class="custom-hero">
  <h1>欢迎来到我的网站</h1>
  <p>这是一个完全自定义的首页</p>
  <a href="/products/" class="btn">浏览商品</a>
</div>

<section class="custom-features">
  <div class="feature">
    <i class="fas fa-shield-alt"></i>
    <h3>安全可靠</h3>
  </div>
</section>`}
                  />
                </div>

                {/* 自定义 CSS */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <label className="text-sm text-dark-400">自定义 CSS</label>
                    <span className="text-xs text-dark-500">样式将自动注入到页面</span>
                  </div>
                  <textarea
                    value={config.custom_css}
                    onChange={(e) => setConfig({ ...config, custom_css: e.target.value })}
                    rows={12}
                    className="w-full px-4 py-3 bg-dark-900 border border-dark-600 rounded-lg text-dark-100 font-mono text-sm resize-y"
                    placeholder={`/* 自定义样式 */
.custom-hero {
  padding: 100px 20px;
  text-align: center;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
}

.custom-hero h1 {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.custom-hero .btn {
  display: inline-block;
  padding: 12px 24px;
  background: white;
  color: #6366f1;
  border-radius: 8px;
  text-decoration: none;
  margin-top: 20px;
}

.custom-features {
  padding: 60px 20px;
  display: flex;
  justify-content: center;
  gap: 40px;
}

.feature {
  text-align: center;
}`}
                  />
                </div>

                {/* 自定义 JavaScript */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <label className="text-sm text-dark-400">自定义 JavaScript</label>
                    <span className="text-xs text-dark-500">可访问 container 和 config 变量</span>
                  </div>
                  <textarea
                    value={config.custom_js}
                    onChange={(e) => setConfig({ ...config, custom_js: e.target.value })}
                    rows={10}
                    className="w-full px-4 py-3 bg-dark-900 border border-dark-600 rounded-lg text-dark-100 font-mono text-sm resize-y"
                    placeholder={`// 自定义 JavaScript
// 可用变量：
//   container - 首页内容容器 DOM 元素
//   config - 当前首页配置对象

console.log('自定义首页已加载');

// 示例：添加动画效果
const hero = container.querySelector('.custom-hero');
if (hero) {
  hero.style.opacity = '0';
  hero.style.transform = 'translateY(20px)';
  hero.style.transition = 'all 0.6s ease';
  
  setTimeout(() => {
    hero.style.opacity = '1';
    hero.style.transform = 'translateY(0)';
  }, 100);
}`}
                  />
                </div>

                {/* 可用变量说明 */}
                <div className="p-4 bg-dark-700/30 rounded-lg">
                  <h4 className="text-sm font-medium text-dark-200 mb-2">
                    <i className="fas fa-info-circle mr-2 text-primary-400" />
                    JavaScript 可用变量
                  </h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                    <div>
                      <code className="text-primary-400">container</code>
                      <p className="text-dark-400 mt-1">首页内容的 DOM 容器元素，可用于操作自定义 HTML</p>
                    </div>
                    <div>
                      <code className="text-primary-400">config</code>
                      <p className="text-dark-400 mt-1">当前首页配置对象，包含所有配置项</p>
                    </div>
                  </div>
                </div>
              </>
            )}

            {!config.advanced_mode && (
              <div className="text-center py-12 text-dark-500">
                <i className="fas fa-code text-4xl mb-4" />
                <p>启用高级模式后，可以使用自定义 HTML/CSS/JS 完全控制首页设计</p>
                <p className="text-sm mt-2">适合有前端开发经验的用户</p>
              </div>
            )}
          </div>
        )}
      </div>

      {/* 重置确认弹窗 */}
      <Modal
        isOpen={showResetModal}
        onClose={() => setShowResetModal(false)}
        title="确认重置"
      >
        <div className="space-y-4">
          <p className="text-dark-300">
            确定要将首页配置重置为当前模板的默认设置吗？此操作不可撤销。
          </p>
          <div className="flex justify-end gap-3">
            <Button variant="secondary" onClick={() => setShowResetModal(false)}>
              取消
            </Button>
            <Button variant="danger" onClick={handleReset}>
              确认重置
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default Homepage
