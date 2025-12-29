'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Input, Card } from '@/components/ui'
import { apiPost } from '@/lib/api'

/**
 * 客服登录页面
 */
export default function StaffLoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [needs2FA, setNeeds2FA] = useState(false)
  const [totpCode, setTotpCode] = useState('')

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!username.trim()) {
      toast.error('请输入用户名')
      return
    }
    if (!password) {
      toast.error('请输入密码')
      return
    }

    setLoading(true)
    const res = await apiPost<{ staff: { username: string; nickname: string }; needs_2fa?: boolean }>(
      '/api/staff/login',
      { username, password }
    )

    if (res.success) {
      if (res.needs_2fa) {
        setNeeds2FA(true)
        toast.success('请输入二步验证码')
      } else {
        toast.success('登录成功')
        window.location.href = '/staff/'
      }
    } else {
      toast.error(res.error || '登录失败')
    }
    setLoading(false)
  }

  const handleVerify2FA = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!totpCode || totpCode.length !== 6) {
      toast.error('请输入6位验证码')
      return
    }

    setLoading(true)
    const res = await apiPost('/api/staff/2fa/verify', { code: totpCode })

    if (res.success) {
      toast.success('验证成功')
      window.location.href = '/staff/'
    } else {
      toast.error(res.error || '验证失败')
    }
    setLoading(false)
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-gradient-to-br from-dark-900 via-dark-800 to-dark-900">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md"
      >
        <Card className="p-8">
          {/* Logo */}
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary-500/20 mb-4">
              <i className={`fas ${needs2FA ? 'fa-shield-alt' : 'fa-headset'} text-3xl text-primary-400`} />
            </div>
            <h1 className="text-2xl font-bold text-dark-100">
              {needs2FA ? '二步验证' : '客服工作台'}
            </h1>
            <p className="text-dark-400 mt-2">
              {needs2FA ? '请输入身份验证器中的验证码' : '请登录您的客服账号'}
            </p>
          </div>

          {needs2FA ? (
            /* 二步验证表单 */
            <form onSubmit={handleVerify2FA} className="space-y-4">
              <Input
                label="验证码"
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                placeholder="请输入6位验证码"
                icon={<i className="fas fa-key" />}
                maxLength={6}
              />

              <Button type="submit" className="w-full" loading={loading}>
                <i className="fas fa-check mr-2" />
                验证
              </Button>

              <button
                type="button"
                onClick={() => {
                  setNeeds2FA(false)
                  setTotpCode('')
                  setPassword('')
                }}
                className="w-full text-center text-dark-400 hover:text-primary-400 text-sm"
              >
                <i className="fas fa-arrow-left mr-1" />
                返回登录
              </button>
            </form>
          ) : (
            /* 登录表单 */
            <form onSubmit={handleLogin} className="space-y-4">
              <Input
                label="用户名"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="请输入用户名"
                icon={<i className="fas fa-user" />}
              />

              <Input
                label="密码"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
                icon={<i className="fas fa-lock" />}
              />

              <Button type="submit" className="w-full" loading={loading}>
                <i className="fas fa-sign-in-alt mr-2" />
                登录
              </Button>
            </form>
          )}

          {/* 返回链接 */}
          {!needs2FA && (
            <div className="text-center mt-6">
              <a href="/" className="text-dark-400 hover:text-primary-400 text-sm">
                <i className="fas fa-arrow-left mr-1" />
                返回首页
              </a>
            </div>
          )}
        </Card>
      </motion.div>
    </div>
  )
}
