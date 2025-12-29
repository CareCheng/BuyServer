'use client'

import { useState, useEffect, useCallback } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Input } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { isValidEmail, cn } from '@/lib/utils'
import { useI18n } from '@/hooks/useI18n'

/**
 * 验证码验证状态
 */
type CodeStatus = 'idle' | 'checking' | 'valid' | 'invalid'

/**
 * 注册页面
 */
export default function RegisterPage() {
  const router = useRouter()
  const { t } = useI18n()
  const [loading, setLoading] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [captchaId, setCaptchaId] = useState('')
  const [captchaImage, setCaptchaImage] = useState('')
  const [codeLength, setCodeLength] = useState(6) // 验证码长度
  const [codeStatus, setCodeStatus] = useState<CodeStatus>('idle') // 验证码验证状态
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    emailCode: '',
    phone: '',
    password: '',
    confirmPassword: '',
    captcha: '',
  })

  // 刷新验证码
  const refreshCaptcha = async () => {
    const res = await apiGet<{ captcha_id: string; image: string }>('/api/captcha')
    if (res.success) {
      setCaptchaId(res.captcha_id)
      setCaptchaImage(res.image)
    }
  }

  // 获取验证码长度
  const loadCodeLength = useCallback(async () => {
    const res = await apiGet<{ code_length: number }>('/api/user/email/code_length')
    if (res.success && res.code_length) {
      setCodeLength(res.code_length)
    }
  }, [])

  useEffect(() => {
    refreshCaptcha()
    loadCodeLength()
  }, [loadCodeLength])

  // 倒计时
  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  // 实时验证验证码
  const verifyCodeRealtime = useCallback(async (code: string) => {
    if (!formData.email || code.length !== codeLength) {
      setCodeStatus('idle')
      return
    }

    setCodeStatus('checking')
    const res = await apiPost<{ valid: boolean }>('/api/user/email/verify_only', {
      email: formData.email,
      code: code,
      code_type: 'register',
    })

    if (res.success) {
      setCodeStatus(res.valid ? 'valid' : 'invalid')
    } else {
      setCodeStatus('invalid')
    }
  }, [formData.email, codeLength])

  // 当验证码输入达到指定长度时自动验证
  useEffect(() => {
    if (formData.emailCode.length === codeLength && formData.email) {
      verifyCodeRealtime(formData.emailCode)
    } else if (formData.emailCode.length < codeLength) {
      setCodeStatus('idle')
    }
  }, [formData.emailCode, formData.email, codeLength, verifyCodeRealtime])

  // 发送邮箱验证码
  const sendEmailCode = async () => {
    if (!formData.email) {
      toast.error(t('user.emailFirst'))
      return
    }
    if (!isValidEmail(formData.email)) {
      toast.error(t('user.emailInvalid'))
      return
    }

    setSendingCode(true)
    const res = await apiPost('/api/user/email/send_code', {
      email: formData.email,
      code_type: 'register',
    })
    setSendingCode(false)

    if (res.success) {
      toast.success(t('user.codeSent'))
      setCountdown(60)
      // 重置验证状态
      setCodeStatus('idle')
      setFormData(prev => ({ ...prev, emailCode: '' }))
    } else {
      toast.error(res.error || t('user.codeSendFailed'))
    }
  }

  // 提交注册
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!formData.username || !formData.email || !formData.emailCode || !formData.password) {
      toast.error(t('common.fillComplete'))
      return
    }

    if (formData.password !== formData.confirmPassword) {
      toast.error(t('user.passwordMismatch'))
      return
    }

    if (formData.password.length < 6) {
      toast.error(t('user.passwordTooShort'))
      return
    }

    // 检查验证码状态
    if (codeStatus === 'invalid') {
      toast.error(t('user.codeIncorrect'))
      return
    }

    setLoading(true)
    const res = await apiPost('/api/user/register', {
      username: formData.username,
      email: formData.email,
      email_code: formData.emailCode,
      phone: formData.phone,
      password: formData.password,
      confirm_password: formData.confirmPassword,
      captcha_id: captchaId,
      captcha_code: formData.captcha,
    })
    setLoading(false)

    if (res.success) {
      toast.success(t('auth.registerSuccess'))
      setTimeout(() => router.push('/login/'), 1500)
    } else {
      toast.error(res.error || t('auth.registerFailed'))
      refreshCaptcha()
    }
  }

  // 获取验证码状态图标和样式
  const getCodeStatusIcon = () => {
    switch (codeStatus) {
      case 'checking':
        return <i className="fas fa-spinner fa-spin text-primary-400" />
      case 'valid':
        return <i className="fas fa-check-circle text-green-400" />
      case 'invalid':
        return <i className="fas fa-times-circle text-red-400" />
      default:
        return null
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 py-12">
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
            <h1 className="text-2xl font-bold text-dark-100 mb-2">{t('auth.registerTitle')}</h1>
            <p className="text-dark-400">{t('auth.registerSubtitle')}</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label={t('user.username')}
              placeholder={t('user.usernamePlaceholder')}
              value={formData.username}
              onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              icon={<i className="fas fa-user" />}
            />

            <Input
              label={t('user.email')}
              type="email"
              placeholder={t('user.emailPlaceholder')}
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              icon={<i className="fas fa-envelope" />}
            />

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-dark-300">{t('user.emailCode')}</label>
              <div className="flex items-center gap-3">
                <div className="relative flex-1">
                  <Input
                    placeholder={t('user.emailCodePlaceholder').replace('{length}', String(codeLength))}
                    value={formData.emailCode}
                    onChange={(e) => {
                      // 只允许输入数字
                      const value = e.target.value.replace(/\D/g, '').slice(0, codeLength)
                      setFormData({ ...formData, emailCode: value })
                    }}
                    maxLength={codeLength}
                    className={cn(
                      codeStatus === 'valid' && 'border-green-500/50 focus:border-green-500',
                      codeStatus === 'invalid' && 'border-red-500/50 focus:border-red-500'
                    )}
                  />
                  {/* 验证状态图标 */}
                  <div className="absolute right-3 top-1/2 -translate-y-1/2">
                    {getCodeStatusIcon()}
                  </div>
                </div>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={sendEmailCode}
                  disabled={countdown > 0 || sendingCode}
                >
                  {countdown > 0 ? `${countdown}${t('common.seconds')}` : sendingCode ? t('common.sending') : t('user.sendCode')}
                </Button>
              </div>
              {/* 验证状态提示 */}
              {codeStatus === 'valid' && (
                <p className="text-green-400 text-xs mt-1">
                  <i className="fas fa-check mr-1" />{t('user.codeCorrect')}
                </p>
              )}
              {codeStatus === 'invalid' && (
                <p className="text-red-400 text-xs mt-1">
                  <i className="fas fa-times mr-1" />{t('user.codeIncorrect')}
                </p>
              )}
            </div>

            <Input
              label={t('user.phoneOptional')}
              type="tel"
              placeholder={t('user.phonePlaceholder')}
              value={formData.phone}
              onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
              icon={<i className="fas fa-phone" />}
            />

            <Input
              label={t('user.password')}
              type="password"
              placeholder={t('user.passwordMinLength')}
              value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              icon={<i className="fas fa-lock" />}
            />

            <Input
              label={t('user.confirmPassword')}
              type="password"
              placeholder={t('user.confirmPasswordPlaceholder')}
              value={formData.confirmPassword}
              onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
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

            <Button type="submit" className="w-full" loading={loading}>
              {t('auth.register')}
            </Button>
          </form>

          <div className="mt-6 flex justify-between text-sm">
            <Link href="/login/" className="text-primary-400 hover:text-primary-300 transition-colors">
              {t('auth.hasAccount')}
            </Link>
            <Link href="/" className="text-dark-400 hover:text-dark-300 transition-colors">
              {t('auth.backToHome')}
            </Link>
          </div>
        </div>
      </motion.div>
    </div>
  )
}
