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
 * 修改密码弹窗
 */
export function ChangePasswordModal({
  isOpen,
  onClose,
  userEmail,
  twoFAStatus,
}: {
  isOpen: boolean
  onClose: () => void
  userEmail: string
  twoFAStatus: TwoFAStatus
}) {
  const [step, setStep] = useState(1)
  const [verifyCode, setVerifyCode] = useState('')
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [sending, setSending] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [verified, setVerified] = useState(false)

  const isUsingTOTP = twoFAStatus.enabled && twoFAStatus.has_totp && !twoFAStatus.prefer_email_auth

  useEffect(() => {
    if (!isOpen) {
      setStep(1)
      setVerifyCode('')
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
      setVerified(false)
    }
  }, [isOpen])

  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  // 发送邮箱验证码
  const sendCode = async () => {
    setSending(true)
    const res = await apiPost('/api/user/email/send_code', { email: userEmail, code_type: 'change_password' })
    setSending(false)
    if (res.success) {
      toast.success('验证码已发送')
      setCountdown(60)
    } else {
      toast.error(res.error || '发送失败')
    }
  }

  // 验证身份
  const verifyIdentity = async () => {
    if (isUsingTOTP) {
      if (!verifyCode || verifyCode.length !== 6) {
        toast.error('请输入6位动态口令')
        return
      }
      setSubmitting(true)
      const res = await apiPost('/api/user/2fa/verify_totp', { code: verifyCode })
      setSubmitting(false)
      if (res.success) {
        toast.success('验证成功')
        setVerified(true)
        setStep(2)
      } else {
        toast.error(res.error || '动态口令错误')
      }
    } else {
      if (!verifyCode) {
        toast.error('请输入验证码')
        return
      }
      setSubmitting(true)
      const res = await apiPost('/api/user/email/verify', { email: userEmail, code: verifyCode, code_type: 'change_password' })
      setSubmitting(false)
      if (res.success) {
        toast.success('验证成功')
        setVerified(true)
        setStep(2)
      } else {
        toast.error(res.error || '验证码错误')
      }
    }
  }

  // 提交修改密码
  const handleSubmit = async () => {
    if (!verified) {
      toast.error('请先完成身份验证')
      return
    }
    if (!oldPassword || !newPassword) {
      toast.error('请填写完整信息')
      return
    }
    if (newPassword !== confirmPassword) {
      toast.error('两次密码不一致')
      return
    }
    if (newPassword.length < 6) {
      toast.error('新密码长度至少6位')
      return
    }
    setSubmitting(true)
    const res = await apiPost('/api/user/password', { old_password: oldPassword, new_password: newPassword })
    setSubmitting(false)
    if (res.success) {
      toast.success('密码修改成功')
      onClose()
    } else {
      toast.error(res.error || '修改失败')
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="修改密码" size="sm">
      {step === 1 ? (
        <div className="space-y-4">
          <p className="text-dark-400 text-sm">为确保是您本人操作，请先完成验证</p>
          {isUsingTOTP ? (
            <div className="setup-step">
              <div className="setup-step-title">🔐 输入动态口令</div>
              <input
                type="text"
                maxLength={6}
                placeholder="000000"
                value={verifyCode}
                onChange={(e: ChangeEvent<HTMLInputElement>) => setVerifyCode(e.target.value.replace(/\D/g, ''))}
                className="verify-code-input"
              />
            </div>
          ) : (
            <div className="setup-step">
              <div className="setup-step-title">📧 验证邮箱</div>
              <p className="text-dark-500 text-sm mb-3">验证码将发送到 {userEmail}</p>
              <div className="flex items-center gap-3">
                <Input
                  placeholder="输入验证码"
                  value={verifyCode}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setVerifyCode(e.target.value)}
                />
                <Button
                  variant="secondary"
                  onClick={sendCode}
                  disabled={countdown > 0 || sending}
                >
                  {countdown > 0 ? `${countdown}秒` : sending ? '发送中...' : '发送验证码'}
                </Button>
              </div>
            </div>
          )}
          <Button className="w-full" onClick={verifyIdentity} loading={submitting}>
            验证并继续
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          <Input
            label="当前密码"
            type="password"
            placeholder="请输入当前密码"
            value={oldPassword}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setOldPassword(e.target.value)}
          />
          <Input
            label="新密码"
            type="password"
            placeholder="至少6位"
            value={newPassword}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setNewPassword(e.target.value)}
          />
          <Input
            label="确认新密码"
            type="password"
            placeholder="再次输入新密码"
            value={confirmPassword}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setConfirmPassword(e.target.value)}
          />
          <div className="flex flex-col sm:flex-row gap-3">
            <Button variant="secondary" className="flex-1 sm:flex-none" onClick={() => setStep(1)}>
              上一步
            </Button>
            <Button className="flex-1" onClick={handleSubmit} loading={submitting}>
              修改密码
            </Button>
          </div>
        </div>
      )}
    </Modal>
  )
}
