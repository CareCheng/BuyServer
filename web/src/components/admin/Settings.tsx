'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Input, Switch } from '@/components/ui'
import { apiGet, apiPost, apiDelete } from '@/lib/api'
import { useTheme, Theme } from '@/lib/theme'
import { Settings } from './types'

// 黑名单条目类型
interface BlacklistEntry {
  ip: string
  expires_at: string
  remaining: number
}

// 白名单配置类型
interface WhitelistConfig {
  enabled: boolean
  whitelist: string[]
}

export function SettingsPage() {
  const { theme, setTheme } = useTheme()
  const [loading, setLoading] = useState(true)
  const [basicForm, setBasicForm] = useState({ system_title: '', admin_suffix: 'manage', server_port: '8080' })
  const [securityForm, setSecurityForm] = useState({
    enable_login: true, admin_username: 'admin', admin_password: '', enable_2fa: false, totp_secret: ''
  })
  const [totpTestCode, setTotpTestCode] = useState('')
  const [totpTestResult, setTotpTestResult] = useState<boolean | null>(null)
  const [blacklist, setBlacklist] = useState<BlacklistEntry[]>([])
  const [blacklistLoading, setBlacklistLoading] = useState(false)
  const [whitelistEnabled, setWhitelistEnabled] = useState(false)
  const [whitelist, setWhitelist] = useState<string[]>([])
  const [whitelistLoading, setWhitelistLoading] = useState(false)
  const [newWhitelistIP, setNewWhitelistIP] = useState('')

  const loadSettings = useCallback(async () => {
    const res = await apiGet<{ settings: Settings }>('/api/admin/settings')
    if (res.success && res.settings) {
      setBasicForm({
        system_title: res.settings.system_title || '',
        admin_suffix: res.settings.admin_suffix || 'manage',
        server_port: String(res.settings.server_port || 8080)
      })
      setSecurityForm({
        enable_login: res.settings.enable_login,
        admin_username: res.settings.admin_username || 'admin',
        admin_password: '',
        enable_2fa: res.settings.enable_2fa,
        totp_secret: res.settings.totp_secret || ''
      })
    }
    setLoading(false)
  }, [])

  const loadBlacklist = useCallback(async () => {
    setBlacklistLoading(true)
    const res = await apiGet<{ blacklist: BlacklistEntry[] }>('/api/admin/blacklist')
    if (res.success && res.blacklist) {
      setBlacklist(res.blacklist)
    }
    setBlacklistLoading(false)
  }, [])

  const loadWhitelist = useCallback(async () => {
    setWhitelistLoading(true)
    const res = await apiGet<WhitelistConfig>('/api/admin/whitelist')
    if (res.success) {
      setWhitelistEnabled(res.enabled || false)
      setWhitelist(res.whitelist || [])
    }
    setWhitelistLoading(false)
  }, [])

  useEffect(() => { loadSettings(); loadBlacklist(); loadWhitelist() }, [loadSettings, loadBlacklist, loadWhitelist])

  // 从黑名单移除IP
  const handleRemoveFromBlacklist = async (ip: string) => {
    if (!confirm(`确定要将 ${ip} 从黑名单中移除吗？`)) return
    const res = await apiDelete(`/api/admin/blacklist/${encodeURIComponent(ip)}`)
    if (res.success) {
      toast.success('已从黑名单中移除')
      loadBlacklist()
    } else {
      toast.error(res.error || '移除失败')
    }
  }

  // 清空黑名单
  const handleClearBlacklist = async () => {
    if (!confirm('确定要清空所有黑名单吗？此操作不可撤销。')) return
    const res = await apiDelete('/api/admin/blacklist')
    if (res.success) {
      toast.success('已清空黑名单')
      loadBlacklist()
    } else {
      toast.error(res.error || '清空失败')
    }
  }

  // 保存白名单配置
  const handleSaveWhitelist = async (enabled: boolean, list: string[]) => {
    const res = await apiPost('/api/admin/whitelist', { enabled, whitelist: list })
    if (res.success) {
      toast.success('白名单配置已保存')
      setWhitelistEnabled(enabled)
      setWhitelist(list)
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  // 切换白名单开关
  const handleToggleWhitelist = async (enabled: boolean) => {
    await handleSaveWhitelist(enabled, whitelist)
  }

  // 添加IP到白名单
  const handleAddWhitelistIP = async () => {
    const ip = newWhitelistIP.trim()
    if (!ip) {
      toast.error('请输入IP地址')
      return
    }
    // 简单的IP格式验证
    const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/
    if (!ipRegex.test(ip)) {
      toast.error('请输入有效的IP地址格式')
      return
    }
    if (whitelist.includes(ip)) {
      toast.error('该IP已在白名单中')
      return
    }
    const newList = [...whitelist, ip]
    await handleSaveWhitelist(whitelistEnabled, newList)
    setNewWhitelistIP('')
  }

  // 从白名单移除IP
  const handleRemoveWhitelistIP = async (ip: string) => {
    if (!confirm(`确定要将 ${ip} 从白名单中移除吗？`)) return
    const newList = whitelist.filter(item => item !== ip)
    await handleSaveWhitelist(whitelistEnabled, newList)
  }

  // 清空白名单
  const handleClearWhitelist = async () => {
    if (!confirm('确定要清空所有白名单吗？')) return
    await handleSaveWhitelist(whitelistEnabled, [])
  }

  // 格式化剩余时间
  const formatRemaining = (seconds: number) => {
    if (seconds < 60) return `${seconds}秒`
    if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`
    return `${Math.floor(seconds / 3600)}时${Math.floor((seconds % 3600) / 60)}分`
  }

  const handleSaveBasic = async () => {
    const suffix = basicForm.admin_suffix.trim()
    if (suffix && !/^[a-zA-Z0-9_-]+$/.test(suffix)) {
      toast.error('管理后台入口后缀只能包含字母、数字、下划线和横线')
      return
    }
    const port = parseInt(basicForm.server_port) || 8080
    if (port < 1 || port > 65535) {
      toast.error('端口号必须在 1-65535 之间')
      return
    }
    const res = await apiPost('/api/admin/settings', {
      system_title: basicForm.system_title.trim(),
      admin_suffix: suffix || 'manage',
      server_port: port
    })
    if (res.success) {
      toast.success('设置已保存')
      // 从服务器重新加载以确保显示最新数据
      await loadSettings()
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  const handleSaveSecurity = async () => {
    const data: Record<string, unknown> = {
      enable_login: securityForm.enable_login,
      admin_username: securityForm.admin_username.trim() || 'admin',
      enable_2fa: securityForm.enable_2fa,
      totp_secret: securityForm.totp_secret
    }
    if (securityForm.admin_password) data.admin_password = securityForm.admin_password
    const res = await apiPost('/api/admin/settings/security', data)
    if (res.success) {
      toast.success('安全设置已保存')
      // 从服务器重新加载以确保显示最新数据
      await loadSettings()
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  const generateTotp = async () => {
    const res = await apiPost<{ secret: string }>('/api/admin/2fa/generate', {})
    if (res.success && res.secret) {
      setSecurityForm({ ...securityForm, totp_secret: res.secret })
      toast.success('新密钥已生成，请扫描二维码并验证')
    } else {
      toast.error(res.error || '生成失败')
    }
  }

  const testTotp = async () => {
    if (!totpTestCode || totpTestCode.length !== 6) {
      toast.error('请输入6位验证码')
      return
    }
    const res = await apiPost('/api/admin/2fa/verify', { code: totpTestCode, secret: securityForm.totp_secret })
    setTotpTestResult(res.success)
  }

  const getTotpQrUrl = () => {
    const title = encodeURIComponent(basicForm.system_title || '卡密购买系统')
    const user = encodeURIComponent(securityForm.admin_username || 'admin')
    const uri = `otpauth://totp/${title}:${user}?secret=${securityForm.totp_secret}&issuer=${title}`
    return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(uri)}`
  }

  const handleThemeChange = (newTheme: Theme) => {
    setTheme(newTheme)
    toast.success(`已切换到${newTheme === 'dark' ? '深色' : '浅色'}主题`)
  }

  if (loading) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-medium" style={{ color: 'var(--text-primary)' }}>系统设置</h2>

      <Card title="主题设置">
        <div className="space-y-4">
          <p className="text-sm" style={{ color: 'var(--text-muted)' }}>选择您喜欢的界面主题风格，设置将自动保存并应用于所有页面</p>
          <div className="grid grid-cols-2 gap-4">
            <button
              onClick={() => handleThemeChange('dark')}
              className={`p-4 rounded-xl border-2 transition-all duration-200 ${
                theme === 'dark' 
                  ? 'border-primary-500 bg-primary-500/10' 
                  : 'border-dark-600 hover:border-dark-500'
              }`}
            >
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-lg bg-dark-800 border border-dark-600 flex items-center justify-center">
                  <i className="fas fa-moon text-primary-400" />
                </div>
                <div className="text-left">
                  <div className="font-medium" style={{ color: 'var(--text-primary)' }}>深色主题</div>
                  <div className="text-xs" style={{ color: 'var(--text-muted)' }}>护眼暗色风格</div>
                </div>
              </div>
              <div className="flex gap-1">
                <div className="w-6 h-6 rounded bg-slate-900 border border-dark-600" />
                <div className="w-6 h-6 rounded bg-slate-800 border border-dark-600" />
                <div className="w-6 h-6 rounded bg-purple-900 border border-dark-600" />
                <div className="w-6 h-6 rounded bg-primary-500 border border-dark-600" />
              </div>
            </button>
            <button
              onClick={() => handleThemeChange('light')}
              className={`p-4 rounded-xl border-2 transition-all duration-200 ${
                theme === 'light' 
                  ? 'border-primary-500 bg-primary-500/10' 
                  : 'border-dark-600 hover:border-dark-500'
              }`}
            >
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-lg bg-white border border-gray-200 flex items-center justify-center">
                  <i className="fas fa-sun text-amber-500" />
                </div>
                <div className="text-left">
                  <div className="font-medium" style={{ color: 'var(--text-primary)' }}>浅色主题</div>
                  <div className="text-xs" style={{ color: 'var(--text-muted)' }}>明亮清爽风格</div>
                </div>
              </div>
              <div className="flex gap-1">
                <div className="w-6 h-6 rounded bg-gray-50 border border-gray-200" />
                <div className="w-6 h-6 rounded bg-white border border-gray-200" />
                <div className="w-6 h-6 rounded bg-indigo-50 border border-gray-200" />
                <div className="w-6 h-6 rounded bg-primary-500 border border-gray-200" />
              </div>
            </button>
          </div>
        </div>
      </Card>
      
      <Card title="基本设置">
        <div className="space-y-4">
          <Input label="系统标题" value={basicForm.system_title} onChange={(e) => setBasicForm({ ...basicForm, system_title: e.target.value })} />
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Input label="管理后台入口后缀" value={basicForm.admin_suffix} onChange={(e) => setBasicForm({ ...basicForm, admin_suffix: e.target.value })} placeholder="manage" />
              <p className="text-dark-500 text-xs mt-1">管理后台访问地址: {typeof window !== 'undefined' ? window.location.origin : ''}/{basicForm.admin_suffix || 'manage'}</p>
            </div>
            <div>
              <Input label="服务器端口" type="number" value={basicForm.server_port} onChange={(e) => setBasicForm({ ...basicForm, server_port: e.target.value })} />
              <p className="text-dark-500 text-xs mt-1">修改端口后需要重启程序生效</p>
            </div>
          </div>
          <div className="flex justify-end">
            <Button onClick={handleSaveBasic}>保存设置</Button>
          </div>
        </div>
      </Card>

      <Card title="登录安全设置">
        <div className="space-y-4">
          <Switch 
            checked={securityForm.enable_login} 
            onChange={(checked) => setSecurityForm({ ...securityForm, enable_login: checked })} 
            label="启用登录验证"
            description="关闭后无需登录即可访问管理后台，不推荐"
          />

          {securityForm.enable_login && (
            <div className="p-4 bg-dark-700/30 rounded-lg space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <Input label="管理员用户名" value={securityForm.admin_username} onChange={(e) => setSecurityForm({ ...securityForm, admin_username: e.target.value })} />
                <Input label="新密码" type="password" value={securityForm.admin_password} onChange={(e) => setSecurityForm({ ...securityForm, admin_password: e.target.value })} placeholder="留空保持不变" />
              </div>

              <Switch 
                checked={securityForm.enable_2fa} 
                onChange={(checked) => setSecurityForm({ ...securityForm, enable_2fa: checked })} 
                label="启用两步验证"
                description="使用 TOTP 验证器增强账户安全"
              />

              {securityForm.enable_2fa && (
                <div className="p-4 bg-dark-700/30 rounded-lg space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-dark-300 mb-1">TOTP密钥</label>
                    <div className="flex gap-2">
                      <input type="text" value={securityForm.totp_secret} readOnly className="flex-1 px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-400 font-mono text-sm" />
                      <Button variant="secondary" onClick={generateTotp}>生成新密钥</Button>
                    </div>
                  </div>

                  {securityForm.totp_secret && (
                    <>
                      <div>
                        <label className="block text-sm font-medium text-dark-300 mb-2">扫描二维码</label>
                        <div className="inline-block p-2 bg-white rounded-lg">
                          <img src={getTotpQrUrl()} alt="TOTP QR Code" className="w-48 h-48" />
                        </div>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-dark-300 mb-1">验证码测试</label>
                        <div className="flex gap-2 items-center">
                          <input type="text" value={totpTestCode} onChange={(e) => { setTotpTestCode(e.target.value); setTotpTestResult(null) }} placeholder="输入6位验证码" maxLength={6} className="w-32 px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" />
                          <Button variant="secondary" onClick={testTotp}>验证</Button>
                          {totpTestResult !== null && (
                            <span className={totpTestResult ? 'text-green-400' : 'text-red-400'}>
                              {totpTestResult ? '✓ 验证通过' : '✗ 验证失败'}
                            </span>
                          )}
                        </div>
                      </div>
                    </>
                  )}
                </div>
              )}
            </div>
          )}

          <div className="flex justify-end">
            <Button onClick={handleSaveSecurity}>保存安全设置</Button>
          </div>
        </div>
      </Card>

      <Card title="IP黑名单管理">
        <div className="space-y-4">
          <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
            连续登录失败10次的IP会被自动加入临时黑名单30分钟。您可以在此查看和管理被封禁的IP。
          </p>
          
          <div className="flex justify-between items-center">
            <div className="flex items-center gap-2">
              <Button variant="secondary" onClick={loadBlacklist} disabled={blacklistLoading}>
                {blacklistLoading ? <i className="fas fa-spinner fa-spin mr-2" /> : <i className="fas fa-sync-alt mr-2" />}
                刷新
              </Button>
              <span className="text-sm" style={{ color: 'var(--text-muted)' }}>
                共 {blacklist.length} 个IP被封禁
              </span>
            </div>
            {blacklist.length > 0 && (
              <Button variant="danger" onClick={handleClearBlacklist}>
                <i className="fas fa-trash-alt mr-2" />清空全部
              </Button>
            )}
          </div>

          {blacklist.length === 0 ? (
            <div className="text-center py-8" style={{ color: 'var(--text-muted)' }}>
              <i className="fas fa-shield-alt text-4xl mb-3 opacity-50" />
              <p>暂无被封禁的IP</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b" style={{ borderColor: 'var(--border-color)' }}>
                    <th className="text-left py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>IP地址</th>
                    <th className="text-left py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>过期时间</th>
                    <th className="text-left py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>剩余时间</th>
                    <th className="text-right py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {blacklist.map((entry) => (
                    <tr key={entry.ip} className="border-b" style={{ borderColor: 'var(--border-color)' }}>
                      <td className="py-3 px-4">
                        <code className="px-2 py-1 rounded text-sm" style={{ backgroundColor: 'var(--bg-tertiary)' }}>
                          {entry.ip}
                        </code>
                      </td>
                      <td className="py-3 px-4 text-sm" style={{ color: 'var(--text-secondary)' }}>
                        {entry.expires_at}
                      </td>
                      <td className="py-3 px-4">
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-500/20 text-red-400">
                          <i className="fas fa-clock mr-1" />{formatRemaining(entry.remaining)}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-right">
                        <Button variant="secondary" size="sm" onClick={() => handleRemoveFromBlacklist(entry.ip)}>
                          <i className="fas fa-unlock mr-1" />解封
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </Card>

      <Card title="IP白名单管理">
        <div className="space-y-4">
          <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
            启用后，只有白名单中的IP才能访问管理后台。请确保将您当前的IP添加到白名单中，否则可能无法访问。
          </p>

          <Switch 
            checked={whitelistEnabled} 
            onChange={handleToggleWhitelist} 
            label="启用IP白名单"
            description="开启后只有白名单中的IP可以访问管理后台"
          />

          {whitelistEnabled && (
            <div className="p-4 bg-dark-700/30 rounded-lg space-y-4">
              <div className="flex gap-2">
                <Input 
                  placeholder="输入IP地址，如 192.168.1.1" 
                  value={newWhitelistIP} 
                  onChange={(e) => setNewWhitelistIP(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleAddWhitelistIP()}
                />
                <Button onClick={handleAddWhitelistIP} disabled={whitelistLoading}>
                  <i className="fas fa-plus mr-2" />添加
                </Button>
              </div>

              <div className="flex justify-between items-center">
                <div className="flex items-center gap-2">
                  <Button variant="secondary" onClick={loadWhitelist} disabled={whitelistLoading}>
                    {whitelistLoading ? <i className="fas fa-spinner fa-spin mr-2" /> : <i className="fas fa-sync-alt mr-2" />}
                    刷新
                  </Button>
                  <span className="text-sm" style={{ color: 'var(--text-muted)' }}>
                    共 {whitelist.length} 个IP
                  </span>
                </div>
                {whitelist.length > 0 && (
                  <Button variant="danger" onClick={handleClearWhitelist}>
                    <i className="fas fa-trash-alt mr-2" />清空全部
                  </Button>
                )}
              </div>

              {whitelist.length === 0 ? (
                <div className="text-center py-8" style={{ color: 'var(--text-muted)' }}>
                  <i className="fas fa-list text-4xl mb-3 opacity-50" />
                  <p>白名单为空，请添加允许访问的IP</p>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b" style={{ borderColor: 'var(--border-color)' }}>
                        <th className="text-left py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>IP地址</th>
                        <th className="text-right py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {whitelist.map((ip) => (
                        <tr key={ip} className="border-b" style={{ borderColor: 'var(--border-color)' }}>
                          <td className="py-3 px-4">
                            <code className="px-2 py-1 rounded text-sm" style={{ backgroundColor: 'var(--bg-tertiary)' }}>
                              {ip}
                            </code>
                          </td>
                          <td className="py-3 px-4 text-right">
                            <Button variant="secondary" size="sm" onClick={() => handleRemoveWhitelistIP(ip)}>
                              <i className="fas fa-trash-alt mr-1" />移除
                            </Button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}
        </div>
      </Card>
    </div>
  )
}
