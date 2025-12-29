'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Input, Switch, PromptModal } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { EmailConfig } from './types'

export function EmailPage() {
  const [config, setConfig] = useState<EmailConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [showTestModal, setShowTestModal] = useState(false)
  const [testLoading, setTestLoading] = useState(false)
  const [form, setForm] = useState({
    enabled: false, smtp_host: '', smtp_port: '465', smtp_user: '', smtp_password: '',
    from_name: '', from_email: '', encryption: 'ssl', code_length: '6'
  })

  const loadConfig = useCallback(async () => {
    const res = await apiGet<{ config: EmailConfig }>('/api/admin/email/config')
    if (res.success && res.config) {
      setConfig(res.config)
      setForm({
        enabled: res.config.enabled, smtp_host: res.config.smtp_host || '',
        smtp_port: String(res.config.smtp_port || 465), smtp_user: res.config.smtp_user || '',
        smtp_password: '', from_name: res.config.from_name || '', from_email: res.config.from_email || '',
        encryption: res.config.encryption || 'ssl', code_length: String(res.config.code_length || 6)
      })
    }
    setLoading(false)
  }, [])

  useEffect(() => { loadConfig() }, [loadConfig])

  const handleSave = async () => {
    const data: Record<string, unknown> = {
      enabled: form.enabled, smtp_host: form.smtp_host.trim(),
      smtp_port: parseInt(form.smtp_port) || 465, smtp_user: form.smtp_user.trim(),
      from_name: form.from_name.trim(), from_email: form.from_email.trim(),
      encryption: form.encryption, code_length: parseInt(form.code_length) || 6
    }
    if (form.smtp_password) data.smtp_password = form.smtp_password
    const res = await apiPost('/api/admin/email/config', data)
    if (res.success) { toast.success('配置已保存'); loadConfig() }
    else toast.error(res.error || '保存失败')
  }

  const handleTest = async (email: string) => {
    if (!email) return
    setTestLoading(true)
    const res = await apiPost('/api/admin/email/test', { email })
    setTestLoading(false)
    setShowTestModal(false)
    if (res.success) toast.success('测试邮件已发送，请检查收件箱')
    else toast.error(res.error || '发送失败')
  }

  // 根据加密方式自动设置推荐端口
  const handleEncryptionChange = (encryption: string) => {
    let port = form.smtp_port
    if (encryption === 'ssl' && form.smtp_port === '587') port = '465'
    else if (encryption === 'starttls' && form.smtp_port === '465') port = '587'
    else if (encryption === 'none' && (form.smtp_port === '465' || form.smtp_port === '587')) port = '25'
    setForm({ ...form, encryption, smtp_port: port })
  }

  if (loading) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-medium text-dark-100">邮箱配置</h2>
      <Card>
        <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg mb-4">
          <p className="text-blue-400 text-sm">配置SMTP邮箱后，系统可以发送验证码邮件用于用户注册、登录验证等功能。</p>
        </div>
        <div className="space-y-4">
          <Switch checked={form.enabled} onChange={(checked) => setForm({ ...form, enabled: checked })} label="启用邮箱服务" />
          <div className="grid grid-cols-2 gap-4">
            <Input label="SMTP服务器" value={form.smtp_host} onChange={(e) => setForm({ ...form, smtp_host: e.target.value })} placeholder="smtp.example.com" />
            <Input label="端口" type="number" value={form.smtp_port} onChange={(e) => setForm({ ...form, smtp_port: e.target.value })} />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input label="SMTP用户名" value={form.smtp_user} onChange={(e) => setForm({ ...form, smtp_user: e.target.value })} placeholder="your@email.com" />
            <Input label="SMTP密码" type="password" value={form.smtp_password} onChange={(e) => setForm({ ...form, smtp_password: e.target.value })} placeholder={config?.has_password ? '******(已配置)' : '授权码或密码'} />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input label="发件人名称" value={form.from_name} onChange={(e) => setForm({ ...form, from_name: e.target.value })} placeholder="系统通知" />
            <Input label="发件人邮箱" value={form.from_email} onChange={(e) => setForm({ ...form, from_email: e.target.value })} placeholder="留空则使用SMTP用户名" />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input label="验证码长度" type="number" value={form.code_length} onChange={(e) => setForm({ ...form, code_length: e.target.value })} />
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-2">加密方式</label>
              <select
                value={form.encryption}
                onChange={(e) => handleEncryptionChange(e.target.value)}
                className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100 focus:outline-none focus:border-primary-500"
              >
                <option value="ssl">SSL/TLS (端口465)</option>
                <option value="starttls">STARTTLS (端口587)</option>
                <option value="none">无加密 (端口25)</option>
              </select>
            </div>
          </div>
          <div className="flex gap-2 pt-4">
            <Button onClick={handleSave}>保存配置</Button>
            <Button variant="secondary" onClick={() => setShowTestModal(true)}>发送测试邮件</Button>
          </div>
        </div>
      </Card>

      <Card title="常用SMTP配置参考">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-dark-700">
                <th className="text-left py-2 px-4 text-dark-400">邮箱服务</th>
                <th className="text-left py-2 px-4 text-dark-400">SMTP服务器</th>
                <th className="text-left py-2 px-4 text-dark-400">端口</th>
                <th className="text-left py-2 px-4 text-dark-400">加密方式</th>
              </tr>
            </thead>
            <tbody className="text-dark-300">
              <tr className="border-b border-dark-700/50"><td className="py-2 px-4">QQ邮箱</td><td className="py-2 px-4">smtp.qq.com</td><td className="py-2 px-4">465</td><td className="py-2 px-4">SSL/TLS</td></tr>
              <tr className="border-b border-dark-700/50"><td className="py-2 px-4">163邮箱</td><td className="py-2 px-4">smtp.163.com</td><td className="py-2 px-4">465</td><td className="py-2 px-4">SSL/TLS</td></tr>
              <tr className="border-b border-dark-700/50"><td className="py-2 px-4">Gmail</td><td className="py-2 px-4">smtp.gmail.com</td><td className="py-2 px-4">587</td><td className="py-2 px-4">STARTTLS</td></tr>
              <tr className="border-b border-dark-700/50"><td className="py-2 px-4">Outlook</td><td className="py-2 px-4">smtp.office365.com</td><td className="py-2 px-4">587</td><td className="py-2 px-4">STARTTLS</td></tr>
              <tr><td className="py-2 px-4">阿里云企业邮箱</td><td className="py-2 px-4">smtp.qiye.aliyun.com</td><td className="py-2 px-4">465</td><td className="py-2 px-4">SSL/TLS</td></tr>
            </tbody>
          </table>
        </div>
        <p className="text-dark-500 text-xs mt-2">注意：QQ邮箱和163邮箱需要使用授权码而非登录密码</p>
      </Card>

      {/* 测试邮件弹窗 */}
      <PromptModal
        isOpen={showTestModal}
        onClose={() => setShowTestModal(false)}
        title="发送测试邮件"
        message="请输入接收测试邮件的邮箱地址"
        placeholder="test@example.com"
        inputType="email"
        required
        confirmText="发送"
        onConfirm={handleTest}
        loading={testLoading}
      />
    </div>
  )
}
