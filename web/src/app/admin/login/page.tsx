'use client'

import { useState, useEffect, FormEvent, ChangeEvent } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Input, Switch } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'

/**
 * 管理员登录页面
 */
export default function AdminLoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [captchaCode, setCaptchaCode] = useState('')
  const [captchaId, setCaptchaId] = useState('')
  const [captchaImage, setCaptchaImage] = useState('')
  const [remember, setRemember] = useState(false)
  const [loading, setLoading] = useState(false)
  const [checking, setChecking] = useState(true)

  // 获取当前管理后台路径前缀
  const getAdminBasePath = () => {
    const path = window.location.pathname
    const parts = path.split('/')
    // 移除最后的 login 部分
    if (parts[parts.length - 1] === '' || parts[parts.length - 1] === 'login') {
      parts.pop()
    }
    if (parts[parts.length - 1] === 'login') {
      parts.pop()
    }
    return parts.join('/') || '/'
  }

  // 检查是否需要初始化设置
  const checkSetup = async () => {
    const basePath = getAdminBasePath()
    try {
      const res = await apiGet<{ needs_setup: boolean }>(`${basePath}/check-setup`)
      if (res.success && res.needs_setup) {
        // 需要初始化设置，跳转到设置页面
        window.location.href = `${basePath}/setup/`
        return
      }
    } catch {
      // 忽略错误，继续显示登录页面
    }
    setChecking(false)
  }

  // 加载验证码
  const loadCaptcha = async () => {
    const res = await apiGet<{ captcha_id: string; image: string }>('/api/captcha')
    if (res.success && res.captcha_id) {
      setCaptchaId(res.captcha_id)
      setCaptchaImage(res.image || '')
    }
  }

  useEffect(() => {
    checkSetup()
    loadCaptcha()
  }, [])

  // 提交登录
  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!username || !password) {
      toast.error('请输入用户名和密码')
      return
    }
    if (!captchaCode) {
      toast.error('请输入验证码')
      return
    }

    const basePath = getAdminBasePath()
    
    setLoading(true)
    const res = await apiPost<{ require_totp?: boolean }>(`${basePath}/login`, {
      username,
      password,
      captcha_id: captchaId,
      captcha_code: captchaCode,
      remember,
    })
    setLoading(false)

    if (res.require_totp) {
      // 需要TOTP验证
      window.location.href = `${basePath}/totp/`
    } else if (res.success) {
      toast.success('登录成功')
      setTimeout(() => {
        window.location.href = `${basePath}/`
      }, 1000)
    } else {
      toast.error(res.error || '登录失败')
      loadCaptcha()
      setCaptchaCode('')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-dark-900 via-dark-800 to-dark-900 p-4">
      {checking ? (
        <div className="text-dark-400">加载中...</div>
      ) : (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="w-full max-w-md"
        >
          <div className="glass-card p-8">
            <div className="text-center mb-8">
              <div className="text-4xl mb-4">🔐</div>
              <h1 className="text-2xl font-bold text-dark-100">管理员登录</h1>
            </div>

            <form onSubmit={handleSubmit} className="space-y-5">
              <Input
                label="用户名"
                placeholder="请输入用户名"
                value={username}
                onChange={(e: ChangeEvent<HTMLInputElement>) => setUsername(e.target.value)}
                autoComplete="username"
              />

              <Input
                label="密码"
                type="password"
                placeholder="请输入密码"
                value={password}
                onChange={(e: ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
                autoComplete="current-password"
              />

              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-dark-300">验证码</label>
                <div className="flex items-center gap-3">
                  <Input
                    placeholder="请输入验证码"
                    value={captchaCode}
                    onChange={(e: ChangeEvent<HTMLInputElement>) => setCaptchaCode(e.target.value)}
                  />
                  {captchaImage && (
                    <img
                      src={captchaImage}
                      alt="验证码"
                      onClick={loadCaptcha}
                      className="h-12 rounded-lg cursor-pointer hover:opacity-80 transition-opacity shrink-0"
                    />
                  )}
                </div>
              </div>

              <Switch 
                checked={remember} 
                onChange={(checked) => setRemember(checked)} 
                label="记住我"
                size="sm"
              />

              <Button type="submit" className="w-full" loading={loading}>
                登录
              </Button>
            </form>
          </div>
        </motion.div>
      )}
    </div>
  )
}
