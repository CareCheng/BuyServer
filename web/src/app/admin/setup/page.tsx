'use client'

import { useState, FormEvent, ChangeEvent } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Input } from '@/components/ui'
import { apiPost } from '@/lib/api'

/**
 * 管理员初始化设置页面
 * 首次启动时设置管理员密码
 */
export default function AdminSetupPage() {
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)

  // 获取当前管理后台路径前缀
  const getAdminBasePath = () => {
    const path = window.location.pathname
    const parts = path.split('/')
    // 移除最后的 setup 部分
    if (parts[parts.length - 1] === '' || parts[parts.length - 1] === 'setup') {
      parts.pop()
    }
    if (parts[parts.length - 1] === 'setup') {
      parts.pop()
    }
    return parts.join('/') || '/'
  }

  // 提交设置
  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    
    if (!password || !confirmPassword) {
      toast.error('请输入密码')
      return
    }
    
    if (password.length < 6) {
      toast.error('密码长度至少6位')
      return
    }
    
    if (password !== confirmPassword) {
      toast.error('两次输入的密码不一致')
      return
    }

    const basePath = getAdminBasePath()
    
    setLoading(true)
    const res = await apiPost(`${basePath}/setup`, {
      password,
      confirm_password: confirmPassword,
    })
    setLoading(false)

    if (res.success) {
      toast.success('密码设置成功，即将跳转到登录页面')
      setTimeout(() => {
        window.location.href = `${basePath}/login/`
      }, 1500)
    } else {
      toast.error(res.error || '设置失败')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-dark-900 via-dark-800 to-dark-900 p-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md"
      >
        <div className="glass-card p-8">
          <div className="text-center mb-8">
            <div className="text-4xl mb-4">🔧</div>
            <h1 className="text-2xl font-bold text-dark-100">初始化设置</h1>
            <p className="text-dark-400 mt-2">首次使用，请设置管理员密码</p>
          </div>

          <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-4 mb-6">
            <div className="flex items-start gap-3">
              <span className="text-yellow-500 text-xl">⚠️</span>
              <div className="text-sm text-yellow-200">
                <p className="font-medium mb-1">安全提示</p>
                <p className="text-yellow-300/80">
                  请设置一个强密码，建议包含字母、数字和特殊字符。
                  此密码将用于管理后台登录。
                </p>
              </div>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <Input
              label="管理员密码"
              type="password"
              placeholder="请输入密码（至少6位）"
              value={password}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
              autoComplete="new-password"
            />

            <Input
              label="确认密码"
              type="password"
              placeholder="请再次输入密码"
              value={confirmPassword}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setConfirmPassword(e.target.value)}
              autoComplete="new-password"
            />

            <div className="text-sm text-dark-400 space-y-1">
              <p>• 默认用户名：<span className="text-dark-200 font-mono">admin</span></p>
              <p>• 密码设置后可在系统设置中修改</p>
            </div>

            <Button type="submit" className="w-full" loading={loading}>
              完成设置
            </Button>
          </form>
        </div>
      </motion.div>
    </div>
  )
}
