'use client'

import { useState, useEffect, ChangeEvent } from 'react'
import toast from 'react-hot-toast'
import { Button, Input, Modal } from '@/components/ui'
import { apiPost } from '@/lib/api'

/**
 * 2FA 状态接口
 */
interface TwoFAStatus {
  enabled: boolean
  has_totp: boolean
  prefer_email_auth: boolean
}

/**
 * 开启两步验证弹窗
 */
export function Setup2FAModal({
  isOpen,
  onClose,
  userEmail,
  onSuccess,
}: {
  isOpen: boolean
  onClose: () => void
  userEmail: string
  onSuccess: () => void
}) {
  const [step, setStep] = useState(1)
  const [selectedType, setSelectedType] = useState<'totp' | 'email'>('totp')
  const [emailCode, setEmailCode] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [totpSecret, setTotpSecret] = useState('')
  const [totpUrl, setTotpUrl] = useState('')
  const [sending, setSending] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [countdown, setCountdown] = useState(0)

  useEffect(() => {
    if (!isOpen) {
      setStep(1)
      setSelectedType('totp')
      setEmailCode('')
      setTotpCode('')
      setTotpSecret('')
      setTotpUrl('')
    }
  }, [isOpen])

  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  // 发送邮箱验证码
  const sendEmailCode = async () => {
    setSending(true)
    const res = await apiPost('/api/user/email/send_code', { email: userEmail, code_type: 'enable_2fa' })
    setSending(false)
    if (res.success) {
      toast.success('验证码已发送')
      setCountdown(60)
    } else {
      toast.error(res.error || '发送失败')
    }
  }

  // 验证邮箱并进入下一步
  const verifyEmail = async () => {
    if (!emailCode) {
      toast.error('请输入验证码')
      return
    }
    setSubmitting(true)
    const res = await apiPost('/api/user/email/verify', { email: userEmail, code: emailCode, code_type: 'enable_2fa' })
    setSubmitting(false)
    if (res.success) {
      toast.success('验证成功')
      if (selectedType === 'totp') {
        // 生成TOTP密钥
        const genRes = await apiPost<{ secret: string; url: string }>('/api/user/2fa/generate', {})
        if (genRes.success && genRes.secret) {
          setTotpSecret(genRes.secret)
          setTotpUrl(genRes.url || '')
          setStep(3)
        } else {
          toast.error('生成密钥失败')
        }
      } else {
        // 邮箱方式直接启用
        const enableRes = await apiPost('/api/user/2fa/enable_email', {})
        if (enableRes.success) {
          setStep(4)
        } else {
          toast.error(enableRes.error || '启用失败')
        }
      }
    } else {
      toast.error(res.error || '验证码错误')
    }
  }

  // 启用TOTP
  const enableTOTP = async () => {
    if (!totpCode || totpCode.length !== 6) {
      toast.error('请输入6位验证码')
      return
    }
    setSubmitting(true)
    const res = await apiPost('/api/user/2fa/enable', { secret: totpSecret, code: totpCode })
    setSubmitting(false)
    if (res.success) {
      toast.success('两步验证已启用')
      onClose()
      onSuccess()
    } else {
      toast.error(res.error || '验证码错误')
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="开启两步验证" size="md">
      {/* 步骤1: 选择验证方式 */}
      {step === 1 && (
        <div className="space-y-4">
          <p className="text-dark-400 text-sm">请选择您希望使用的验证方式：</p>
          <div className="grid grid-cols-2 gap-4">
            <button
              onClick={() => setSelectedType('totp')}
              className={`setup-option ${selectedType === 'totp' ? 'active' : ''}`}
            >
              <div className="text-3xl mb-2">🔐</div>
              <div className="font-medium">验证器APP</div>
              <div className="text-xs text-dark-500 mt-1">使用Google Authenticator等</div>
            </button>
            <button
              onClick={() => setSelectedType('email')}
              className={`setup-option ${selectedType === 'email' ? 'active' : ''}`}
            >
              <div className="text-3xl mb-2">📧</div>
              <div className="font-medium">邮箱验证码</div>
              <div className="text-xs text-dark-500 mt-1">每次登录发送验证码</div>
            </button>
          </div>
          <Button className="w-full" onClick={() => setStep(2)}>
            下一步
          </Button>
        </div>
      )}

      {/* 步骤2: 验证邮箱 */}
      {step === 2 && (
        <div className="space-y-4">
          <div className="setup-step">
            <div className="setup-step-title">
              <span className="step-num">1</span> 验证您的邮箱
            </div>
            <p className="text-dark-500 text-sm mb-3">为确保是您本人操作，请先验证邮箱</p>
            <div className="flex items-center gap-3">
              <Input
                placeholder="输入邮箱验证码"
                value={emailCode}
                onChange={(e: ChangeEvent<HTMLInputElement>) => setEmailCode(e.target.value)}
              />
              <Button
                variant="secondary"
                onClick={sendEmailCode}
                disabled={countdown > 0 || sending}
              >
                {countdown > 0 ? `${countdown}秒` : sending ? '发送中...' : '发送验证码'}
              </Button>
            </div>
          </div>
          <div className="flex flex-col sm:flex-row gap-3">
            <Button variant="secondary" className="flex-1 sm:flex-none" onClick={() => setStep(1)}>
              上一步
            </Button>
            <Button className="flex-1" onClick={verifyEmail} loading={submitting}>
              验证并继续
            </Button>
          </div>
        </div>
      )}

      {/* 步骤3: TOTP设置 */}
      {step === 3 && (
        <div className="space-y-4">
          <div className="setup-step">
            <div className="setup-step-title">
              <span className="step-num">2</span> 设置验证器
            </div>
            <p className="text-dark-500 text-sm mb-3">使用验证器APP扫描下方二维码</p>
            <div className="qrcode-box">
              {totpUrl && (
                <img
                  src={`https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=${encodeURIComponent(totpUrl)}`}
                  alt="QR Code"
                  className="mx-auto"
                />
              )}
            </div>
            <div className="mt-4">
              <label className="text-sm text-dark-500 block mb-2">或手动输入密钥：</label>
              <input
                type="text"
                readOnly
                value={totpSecret}
                className="secret-input"
              />
            </div>
          </div>
          <div className="setup-step">
            <div className="setup-step-title">
              <span className="step-num">3</span> 输入验证码确认
            </div>
            <input
              type="text"
              maxLength={6}
              placeholder="000000"
              value={totpCode}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setTotpCode(e.target.value.replace(/\D/g, ''))}
              className="verify-code-input"
            />
          </div>
          <Button className="w-full" onClick={enableTOTP} loading={submitting}>
            确认启用
          </Button>
        </div>
      )}

      {/* 步骤4: 邮箱方式完成 */}
      {step === 4 && (
        <div className="space-y-4">
          <div className="text-center py-8">
            <div className="text-5xl mb-4">✅</div>
            <h4 className="text-lg font-medium text-dark-100 mb-2">邮箱验证已开启</h4>
            <p className="text-dark-500">每次登录时，系统将发送验证码到您的邮箱</p>
          </div>
          <Button className="w-full" onClick={() => { onClose(); onSuccess(); }}>
            完成
          </Button>
        </div>
      )}
    </Modal>
  )
}


/**
 * 关闭两步验证弹窗
 */
export function Disable2FAModal({
  isOpen,
  onClose,
  userEmail,
  twoFAStatus,
  onSuccess,
}: {
  isOpen: boolean
  onClose: () => void
  userEmail: string
  twoFAStatus: TwoFAStatus
  onSuccess: () => void
}) {
  const [totpCode, setTotpCode] = useState('')
  const [emailCode, setEmailCode] = useState('')
  const [sending, setSending] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [countdown, setCountdown] = useState(0)

  const isUsingTOTP = twoFAStatus.has_totp && !twoFAStatus.prefer_email_auth

  useEffect(() => {
    if (!isOpen) {
      setTotpCode('')
      setEmailCode('')
    }
  }, [isOpen])

  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  // 发送邮箱验证码
  const sendEmailCode = async () => {
    setSending(true)
    const res = await apiPost('/api/user/email/send_code', { email: userEmail, code_type: 'disable_2fa' })
    setSending(false)
    if (res.success) {
      toast.success('验证码已发送')
      setCountdown(60)
    } else {
      toast.error(res.error || '发送失败')
    }
  }

  // 确认关闭
  const handleDisable = async () => {
    const body: Record<string, string> = {}
    if (isUsingTOTP) {
      if (!totpCode || totpCode.length !== 6) {
        toast.error('请输入6位动态口令')
        return
      }
      body.totp_code = totpCode
    } else {
      if (!emailCode) {
        toast.error('请输入邮箱验证码')
        return
      }
      body.email_code = emailCode
    }

    setSubmitting(true)
    const res = await apiPost('/api/user/2fa/disable', body)
    setSubmitting(false)
    if (res.success) {
      toast.success('两步验证已关闭')
      onClose()
      onSuccess()
    } else {
      toast.error(res.error || '验证失败')
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="关闭两步验证" size="sm">
      <div className="space-y-4">
        <p className="text-dark-400 text-sm">为确保是您本人操作，请完成验证</p>

        {isUsingTOTP ? (
          <div className="setup-step">
            <div className="setup-step-title">🔐 输入动态口令</div>
            <input
              type="text"
              maxLength={6}
              placeholder="000000"
              value={totpCode}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setTotpCode(e.target.value.replace(/\D/g, ''))}
              className="verify-code-input"
            />
          </div>
        ) : (
          <div className="setup-step">
            <div className="setup-step-title">📧 输入邮箱验证码</div>
            <div className="flex items-center gap-3">
              <Input
                placeholder="输入验证码"
                value={emailCode}
                onChange={(e: ChangeEvent<HTMLInputElement>) => setEmailCode(e.target.value)}
              />
              <Button
                variant="secondary"
                onClick={sendEmailCode}
                disabled={countdown > 0 || sending}
              >
                {countdown > 0 ? `${countdown}秒` : sending ? '发送中...' : '发送验证码'}
              </Button>
            </div>
          </div>
        )}

        <div className="flex flex-col-reverse sm:flex-row gap-3">
          <Button variant="secondary" className="flex-1" onClick={onClose}>
            取消
          </Button>
          <Button variant="danger" className="flex-1" onClick={handleDisable} loading={submitting}>
            确认关闭
          </Button>
        </div>
      </div>
    </Modal>
  )
}

/**
 * 更改验证方式弹窗
 */
export function ChangeMethodModal({
  isOpen,
  onClose,
  userEmail,
  isUsingTOTP,
  onSuccess,
}: {
  isOpen: boolean
  onClose: () => void
  userEmail: string
  isUsingTOTP: boolean
  onSuccess: () => void
}) {
  const [step, setStep] = useState(1)
  const [emailCode, setEmailCode] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [totpSecret, setTotpSecret] = useState('')
  const [totpUrl, setTotpUrl] = useState('')
  const [sending, setSending] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [countdown, setCountdown] = useState(0)

  // 目标方式：当前是TOTP则切换到邮箱，反之亦然
  const targetMethod = isUsingTOTP ? 'email' : 'totp'

  useEffect(() => {
    if (!isOpen) {
      setStep(1)
      setEmailCode('')
      setTotpCode('')
      setTotpSecret('')
      setTotpUrl('')
    }
  }, [isOpen])

  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  // 发送邮箱验证码
  const sendEmailCode = async () => {
    setSending(true)
    const res = await apiPost('/api/user/email/send_code', { email: userEmail, code_type: 'enable_2fa' })
    setSending(false)
    if (res.success) {
      toast.success('验证码已发送')
      setCountdown(60)
    } else {
      toast.error(res.error || '发送失败')
    }
  }

  // 验证邮箱
  const verifyEmail = async () => {
    if (!emailCode) {
      toast.error('请输入验证码')
      return
    }
    setSubmitting(true)
    const res = await apiPost('/api/user/email/verify', { email: userEmail, code: emailCode, code_type: 'enable_2fa' })
    setSubmitting(false)
    if (res.success) {
      toast.success('验证成功')
      if (targetMethod === 'totp') {
        // 切换到TOTP，需要设置验证器
        const genRes = await apiPost<{ secret: string; url: string }>('/api/user/2fa/generate', {})
        if (genRes.success && genRes.secret) {
          setTotpSecret(genRes.secret)
          setTotpUrl(genRes.url || '')
          setStep(2)
        } else {
          toast.error('生成密钥失败')
        }
      } else {
        // 切换到邮箱验证
        const enableRes = await apiPost('/api/user/2fa/enable_email', {})
        if (enableRes.success) {
          setStep(3)
        } else {
          toast.error(enableRes.error || '切换失败')
        }
      }
    } else {
      toast.error(res.error || '验证码错误')
    }
  }

  // 确认切换到TOTP
  const confirmChangeToTOTP = async () => {
    if (!totpCode || totpCode.length !== 6) {
      toast.error('请输入6位验证码')
      return
    }
    setSubmitting(true)
    const res = await apiPost('/api/user/2fa/enable', { secret: totpSecret, code: totpCode })
    if (res.success) {
      // 设置偏好为TOTP
      await apiPost('/api/user/2fa/preference', { prefer_email_auth: false })
      toast.success('已切换到动态口令验证')
      onClose()
      onSuccess()
    } else {
      toast.error(res.error || '验证码错误')
    }
    setSubmitting(false)
  }

  const title = isUsingTOTP ? '切换到邮箱验证' : '设置动态口令验证'

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={title} size="md">
      {/* 步骤1: 验证邮箱 */}
      {step === 1 && (
        <div className="space-y-4">
          <div className="setup-step">
            <div className="setup-step-title">
              <span className="step-num">1</span> 验证您的邮箱
            </div>
            <p className="text-dark-500 text-sm mb-3">为确保是您本人操作，请先验证邮箱</p>
            <div className="flex items-center gap-3">
              <Input
                placeholder="输入邮箱验证码"
                value={emailCode}
                onChange={(e: ChangeEvent<HTMLInputElement>) => setEmailCode(e.target.value)}
              />
              <Button
                variant="secondary"
                onClick={sendEmailCode}
                disabled={countdown > 0 || sending}
              >
                {countdown > 0 ? `${countdown}秒` : sending ? '发送中...' : '发送验证码'}
              </Button>
            </div>
          </div>
          <Button className="w-full" onClick={verifyEmail} loading={submitting}>
            验证并继续
          </Button>
        </div>
      )}

      {/* 步骤2: 设置TOTP（从邮箱切换到TOTP时显示） */}
      {step === 2 && (
        <div className="space-y-4">
          <div className="setup-step">
            <div className="setup-step-title">
              <span className="step-num">2</span> 设置验证器
            </div>
            <p className="text-dark-500 text-sm mb-3">使用验证器APP扫描下方二维码</p>
            <div className="qrcode-box">
              {totpUrl && (
                <img
                  src={`https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=${encodeURIComponent(totpUrl)}`}
                  alt="QR Code"
                  className="mx-auto"
                />
              )}
            </div>
            <div className="mt-4">
              <label className="text-sm text-dark-500 block mb-2">或手动输入密钥：</label>
              <input
                type="text"
                readOnly
                value={totpSecret}
                className="secret-input"
              />
            </div>
          </div>
          <div className="setup-step">
            <div className="setup-step-title">
              <span className="step-num">3</span> 输入验证码确认
            </div>
            <input
              type="text"
              maxLength={6}
              placeholder="000000"
              value={totpCode}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setTotpCode(e.target.value.replace(/\D/g, ''))}
              className="verify-code-input"
            />
          </div>
          <Button className="w-full" onClick={confirmChangeToTOTP} loading={submitting}>
            确认更改
          </Button>
        </div>
      )}

      {/* 步骤3: 切换到邮箱完成 */}
      {step === 3 && (
        <div className="space-y-4">
          <div className="text-center py-8">
            <div className="text-5xl mb-4">✅</div>
            <h4 className="text-lg font-medium text-dark-100 mb-2">已切换到邮箱验证</h4>
            <p className="text-dark-500">每次登录时，系统将发送验证码到您的邮箱</p>
          </div>
          <Button className="w-full" onClick={() => { onClose(); onSuccess(); }}>
            完成
          </Button>
        </div>
      )}
    </Modal>
  )
}
