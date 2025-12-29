'use client'

import { useState, useEffect } from 'react'
import toast from 'react-hot-toast'
import { Button, Card } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { SupportConfig } from './types'
import { Toggle } from './components'

/**
 * 配置管理组件
 */
export function ConfigManagement() {
  const [config, setConfig] = useState<SupportConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const loadConfig = async () => {
    setLoading(true)
    const res = await apiGet<{ config: SupportConfig }>('/api/admin/support/config')
    if (res.success && res.config) {
      setConfig(res.config)
    }
    setLoading(false)
  }

  useEffect(() => {
    loadConfig()
  }, [])

  const handleSave = async () => {
    if (!config) return

    setSaving(true)
    const res = await apiPost('/api/admin/support/config', config as unknown as Record<string, unknown>)
    if (res.success) {
      toast.success('保存成功')
    } else {
      toast.error(res.error || '保存失败')
    }
    setSaving(false)
  }

  if (loading) {
    return (
      <div className="text-center py-8">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  if (!config) {
    return <div className="text-center py-8 text-dark-400">加载配置失败</div>
  }

  return (
    <div className="space-y-6">
      {/* 基本设置 */}
      <Card>
        <h3 className="text-lg font-medium text-dark-100 mb-4 flex items-center gap-2">
          <i className="fas fa-sliders-h text-primary-400" />
          基本设置
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <Toggle
            checked={config.enabled}
            onChange={(v) => setConfig({ ...config, enabled: v })}
            label="启用客服系统"
            description="关闭后用户将无法访问客服功能"
          />
          <Toggle
            checked={config.allow_guest}
            onChange={(v) => setConfig({ ...config, allow_guest: v })}
            label="允许游客咨询"
            description="未登录用户是否可以提交工单"
          />
          <Toggle
            checked={config.enable_staff_2fa}
            onChange={(v) => setConfig({ ...config, enable_staff_2fa: v })}
            label="客服二步验证"
            description="客服登录时需要二步验证"
          />
          <Toggle
            checked={config.enable_auto_assign || false}
            onChange={(v) => setConfig({ ...config, enable_auto_assign: v })}
            label="自动分配工单"
            description="新工单自动分配给负载最低的客服"
          />
        </div>
      </Card>

      {/* 通知设置 */}
      <Card>
        <h3 className="text-lg font-medium text-dark-100 mb-4 flex items-center gap-2">
          <i className="fas fa-bell text-primary-400" />
          通知设置
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <Toggle
            checked={config.enable_email_notify || false}
            onChange={(v) => setConfig({ ...config, enable_email_notify: v })}
            label="启用邮件通知"
            description="开启后将发送邮件通知"
          />
          <Toggle
            checked={config.notify_on_new_ticket || false}
            onChange={(v) => setConfig({ ...config, notify_on_new_ticket: v })}
            label="新工单通知客服"
            description="有新工单时邮件通知在线客服"
          />
          <Toggle
            checked={config.notify_on_reply || false}
            onChange={(v) => setConfig({ ...config, notify_on_reply: v })}
            label="回复通知用户"
            description="客服回复时邮件通知用户"
          />
        </div>
      </Card>

      {/* 客服后台入口 */}
      <Card>
        <h3 className="text-lg font-medium text-dark-100 mb-4 flex items-center gap-2">
          <i className="fas fa-door-open text-primary-400" />
          客服后台入口
        </h3>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">后台路径后缀</label>
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2 flex-1 min-w-0">
                <span className="text-dark-400 text-lg shrink-0">/</span>
                <input
                  type="text"
                  value={config.staff_portal_suffix || 'staff'}
                  onChange={(e) => setConfig({ ...config, staff_portal_suffix: e.target.value })}
                  className="input flex-1 min-w-0"
                  placeholder="staff"
                />
              </div>
              <a
                href={`/${config.staff_portal_suffix || 'staff'}/login`}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center gap-2 h-12 px-4 bg-dark-700 hover:bg-dark-600 text-dark-200 rounded-lg transition-colors whitespace-nowrap shrink-0"
              >
                <i className="fas fa-external-link-alt" />
                <span>打开客服后台</span>
              </a>
            </div>
            <p className="text-dark-500 text-xs">
              访问地址：{typeof window !== 'undefined' ? window.location.origin : ''}/{config.staff_portal_suffix || 'staff'}/login
            </p>
          </div>
        </div>
      </Card>


      {/* 工作时间 */}
      <Card>
        <h3 className="text-lg font-medium text-dark-100 mb-4 flex items-center gap-2">
          <i className="fas fa-clock text-primary-400" />
          工作时间
        </h3>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">开始时间</label>
            <input
              type="time"
              value={config.working_hours_start || '09:00'}
              onChange={(e) => setConfig({ ...config, working_hours_start: e.target.value })}
              className="input w-full"
            />
          </div>
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">结束时间</label>
            <input
              type="time"
              value={config.working_hours_end || '18:00'}
              onChange={(e) => setConfig({ ...config, working_hours_end: e.target.value })}
              className="input w-full"
            />
          </div>
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">工作日</label>
            <input
              type="text"
              value={config.working_days || '1,2,3,4,5'}
              onChange={(e) => setConfig({ ...config, working_days: e.target.value })}
              className="input w-full"
              placeholder="1,2,3,4,5"
            />
            <p className="text-dark-500 text-xs">1-7 表示周一到周日</p>
          </div>
        </div>
      </Card>

      {/* 消息设置 */}
      <Card>
        <h3 className="text-lg font-medium text-dark-100 mb-4 flex items-center gap-2">
          <i className="fas fa-comment-alt text-primary-400" />
          消息设置
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">欢迎消息</label>
            <textarea
              value={config.welcome_message || ''}
              onChange={(e) => setConfig({ ...config, welcome_message: e.target.value })}
              className="input w-full h-24 resize-none"
              placeholder="用户开始咨询时显示的欢迎消息"
            />
          </div>
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">离线提示</label>
            <textarea
              value={config.offline_message || ''}
              onChange={(e) => setConfig({ ...config, offline_message: e.target.value })}
              className="input w-full h-24 resize-none"
              placeholder="客服不在线时显示的提示消息"
            />
          </div>
        </div>
      </Card>

      {/* 工单设置 */}
      <Card>
        <h3 className="text-lg font-medium text-dark-100 mb-4 flex items-center gap-2">
          <i className="fas fa-ticket-alt text-primary-400" />
          工单设置
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">自动关闭时间（小时）</label>
            <input
              type="number"
              value={config.auto_close_hours || 72}
              onChange={(e) => setConfig({ ...config, auto_close_hours: Number(e.target.value) })}
              className="input w-full"
              min={0}
            />
            <p className="text-dark-500 text-xs">工单无回复超过此时间后自动关闭，0表示不自动关闭</p>
          </div>
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">最大附件大小（MB）</label>
            <input
              type="number"
              value={config.max_attachment_size || 5}
              onChange={(e) => setConfig({ ...config, max_attachment_size: Number(e.target.value) })}
              className="input w-full"
              min={1}
              max={50}
            />
          </div>
          <div className="space-y-1.5 md:col-span-2">
            <label className="block text-sm font-medium text-dark-300">工单分类</label>
            <textarea
              value={config.ticket_categories || ''}
              onChange={(e) => setConfig({ ...config, ticket_categories: e.target.value })}
              className="input w-full h-20 resize-none font-mono text-sm"
              placeholder='["订单问题","商品咨询","支付问题","账户问题","其他"]'
            />
            <p className="text-dark-500 text-xs">JSON数组格式，如：["分类1","分类2","分类3"]</p>
          </div>
        </div>
      </Card>

      {/* 保存按钮 */}
      <div className="flex justify-end">
        <Button onClick={handleSave} loading={saving} size="lg">
          <i className="fas fa-save mr-2" />
          保存配置
        </Button>
      </div>
    </div>
  )
}
