'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Input, Switch } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { useI18n } from '@/hooks/useI18n'

/**
 * 登录页面
 */
export default function LoginPage() {
  const router = useRouter()
  const { t } = useI18n()
  const [loading, setLoading] = useState(false)
  const [captchaId, setCaptchaId] = useState('')
  const [captchaImage, setCaptchaImage] = useState('')
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    captcha: '',
    remember: false,
  })

  // 刷新验证码
  const refreshCaptcha = async () => {
    const res = await apiGet<{ captcha_id: string; image: string }>('/api/captcha')
    if (res.success) {
      setCaptchaId(res.captcha_id)
      setCaptchaImage(res.image)
    }
  }

  useEffect(() => {
    refreshCaptcha()
  }, [])

  // 提交登录
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.username || !formData.password || !formData.captcha) {
      toast.error(t('common.fillComplete'))
      return
    }

    setLoading(true)
    const res = await apiPost<{ require_2fa?: boolean; verify_token?: string }>('/api/user/login', {
      username: formData.username,
      password: formData.password,
      captcha_id: captchaId,
      captcha_code: formData.captcha,
      remember: formData.remember,
    })
    setLoading(false)

    if (res.require_2fa && res.verify_token) {
      // 需要二次验证
      router.push(`/verify/?token=${res.verify_token}`)
    } else if (res.success) {
      toast.success(t('auth.loginSuccess'))
      setTimeout(() => {
        window.location.href = '/products/'
      }, 1000)
    } else {
      toast.error(res.error || t('auth.loginFailed'))
      refreshCaptcha()
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      {/* 背景装饰 */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-primary-500/20 rounded-full blur-3xl" />
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-purple-500/20 rounded-full blur-3xl" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative w-full max-w-md"
      >
        <div className="card p-8">
          <div className="text-center mb-8">
            <h1 className="text-2xl font-bold text-dark-100 mb-2">{t('auth.loginTitle')}</h1>
            <p className="text-dark-400">{t('auth.loginSubtitle')}</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <Input
              label={t('user.username')}
              placeholder={t('user.usernamePlaceholder')}
              value={formData.username}
              onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              icon={<i className="fas fa-user" />}
            />

            <Input
              label={t('user.password')}
              type="password"
              placeholder={t('user.passwordPlaceholder')}
              value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              icon={<i className="fas fa-lock" />}
            />

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-dark-300">{t('user.captcha')}</label>
              <div className="flex items-center gap-3">
                <Input
                  placeholder={t('user.captchaPlaceholder')}
                  value={formData.captcha}
                  onChange={(e) => setFormData({ ...formData, captcha: e.target.value })}
                />
                {captchaImage && (
                  <img
                    src={captchaImage}
                    alt={t('user.captcha')}
                    className="h-12 rounded-lg cursor-pointer hover:opacity-80 transition-opacity shrink-0"
                    onClick={refreshCaptcha}
                    title={t('common.clickRefresh')}
                  />
                )}
              </div>
            </div>

            <Switch 
              checked={formData.remember} 
              onChange={(checked) => setFormData({ ...formData, remember: checked })} 
              label={t('user.rememberMe')}
              size="sm"
            />

            <Button type="submit" className="w-full" loading={loading}>
              {t('auth.login')}
            </Button>
          </form>

          <div className="mt-6 flex justify-between text-sm">
            <Link href="/register/" className="text-primary-400 hover:text-primary-300 transition-colors">
              {t('auth.noAccount')}
            </Link>
            <Link href="/forgot/" className="text-dark-400 hover:text-dark-300 transition-colors">
              {t('auth.forgotPassword')}
            </Link>
          </div>
        </div>
      </motion.div>
    </div>
  )
}
