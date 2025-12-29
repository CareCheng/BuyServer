'use client'

import { useState, FormEvent, ChangeEvent } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button } from '@/components/ui'
import { apiPost } from '@/lib/api'

/**
 * 管理员TOTP验证页面
 */
export default function AdminTOTPPage() {
  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)

  // 获取当前管理后台路径前缀
  const getAdminBasePath = () => {
    const path = window.location.pathname
    const parts = path.split('/')
    // 移除最后的 totp 部分
    if (parts[parts.length - 1] === '' || parts[parts.length - 1] === 'totp') {
      parts.pop()
    }
    if (parts[parts.length - 1] === 'totp') {
      parts.pop()
    }
    return parts.join('/') || '/'
  }

  // 提交验证
  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!code || code.length !== 6) {
      toast.error('请输入6位验证码')
      return
    }

    const basePath = getAdminBasePath()
    
    setLoading(true)
    const res = await apiPost(`${basePath}/totp`, { code })
    setLoading(false)

    if (res.success) {
      toast.success('验证成功')
      setTimeout(() => {
        window.location.href = `${basePath}/`
      }, 1000)
    } else {
      toast.error(res.error || '验证码错误')
      setCode('')
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
            <div className="text-4xl mb-4">🔐</div>
            <h1 className="text-2xl font-bold text-dark-100">两步验证</h1>
            <p className="text-dark-400 mt-2">请输入验证器APP中的动态口令</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <input
                type="text"
                maxLength={6}
                placeholder="000000"
                value={code}
                onChange={(e: ChangeEvent<HTMLInputElement>) => setCode(e.target.value.replace(/\D/g, ''))}
                className="verify-code-input"
                autoFocus
              />
            </div>

            <Button type="submit" className="w-full" loading={loading}>
              验证
            </Button>

            <div className="text-center">
              <button
                type="button"
                onClick={() => window.location.href = `${getAdminBasePath()}/login/`}
                className="text-sm text-dark-400 hover:text-primary-400 transition-colors"
              >
                返回登录
              </button>
            </div>
          </form>
        </div>
      </motion.div>
    </div>
  )
}
