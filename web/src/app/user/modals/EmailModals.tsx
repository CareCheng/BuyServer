'use client'

import { useState, useEffect, ChangeEvent } from 'react'
import toast from 'react-hot-toast'
import { Button, Input, Modal } from '@/components/ui'
import { apiPost } from '@/lib/api'
import { isValidEmail } from '@/lib/utils'

/**
 * 绑定邮箱弹窗
 */
export function BindEmailModal({
  isOpen,
  onClose,
  onSuccess,
}: {
  isOpen: boolean
  onClose: () => void
  onSuccess: () => void
}) {
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [sending, setSending] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [countdown, setCountdown] = useState(0)

  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  const sendCode = async () => {
    if (!email || !isValidEmail(email)) {
      toast.error('请输入有效的邮箱地址')
      return
    }
    setSending(true)
    const res = await apiPost('/api/user/email/send_code', { email, code_type: 'register' })
    setSending(false)
    if (res.success) {
      toast.success('验证码已发送')
      setCountdown(60)
    } else {
      toast.error(res.error || '发送失败')
    }
  }

  const handleSubmit = async () => {
    if (!email || !code) {
      toast.error('请填写完整信息')
      return
    }
    setSubmitting(true)
    const res = await apiPost('/api/user/email/bind', { email, code })
    setSubmitting(false)
    if (res.success) {
      toast.success('邮箱绑定成功')
      onClose()
      onSuccess()
    } else {
      toast.error(res.error || '绑定失败')
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="绑定邮箱" size="sm">
      <div className="space-y-4">
        <Input
          label="邮箱地址"
          type="email"
          placeholder="请输入邮箱地址"
          value={email}
          onChange={(e: ChangeEvent<HTMLInputElement>) => setEmail(e.target.value)}
        />
        <div className="space-y-1.5">
          <label className="block text-sm font-medium text-dark-300">验证码</label>
          <div className="flex items-center gap-3">
            <Input
              placeholder="请输入验证码"
              value={code}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setCode(e.target.value)}
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
        <Button className="w-full" onClick={handleSubmit} loading={submitting}>
          确认绑定
        </Button>
      </div>
    </Modal>
  )
}

/**
 * 更换邮箱弹窗
 */
export function ChangeEmailModal({
  isOpen,
  onClose,
  currentEmail,
  onSuccess,
}: {
  isOpen: boolean
  onClose: () => void
  currentEmail: string
  onSuccess: () => void
}) {
  const [step, setStep] = useState(1)
  const [verifyCode, setVerifyCode] = useState('')
  const [newEmail, setNewEmail] = useState('')
  const [newCode, setNewCode] = useState('')
  const [sending, setSending] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [verified, setVerified] = useState(false)

  useEffect(() => {
    if (!isOpen) {
      setStep(1)
      setVerifyCode('')
      setNewEmail('')
      setNewCode('')
      setVerified(false)
    }
  }, [isOpen])

  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  // 发送原邮箱验证码
  const sendOldCode = async () => {
    setSending(true)
    const res = await apiPost('/api/user/email/send_code', { email: currentEmail, code_type: 'change_email' })
    setSending(false)
    if (res.success) {
      toast.success('验证码已发送')
      setCountdown(60)
    } else {
      toast.error(res.error || '发送失败')
    }
  }

  // 验证原邮箱
  const verifyOldEmail = async () => {
    if (!verifyCode) {
      toast.error('请输入验证码')
      return
    }
    setSubmitting(true)
    const res = await apiPost('/api/user/email/verify', { email: currentEmail, code: verifyCode, code_type: 'change_email' })
    setSubmitting(false)
    if (res.success) {
      toast.success('验证成功')
      setVerified(true)
      setStep(2)
      setCountdown(0)
    } else {
      toast.error(res.error || '验证码错误')
    }
  }

  // 发送新邮箱验证码
  const sendNewCode = async () => {
    if (!newEmail || !isValidEmail(newEmail)) {
      toast.error('请输入有效的邮箱地址')
      return
    }
    setSending(true)
    const res = await apiPost('/api/user/email/send_code', { email: newEmail, code_type: 'register' })
    setSending(false)
    if (res.success) {
      toast.success('验证码已发送')
      setCountdown(60)
    } else {
      toast.error(res.error || '发送失败')
    }
  }

  // 确认更换
  const handleSubmit = async () => {
    if (!verified) {
      toast.error('请先完成身份验证')
      return
    }
    if (!newEmail || !newCode) {
      toast.error('请填写完整信息')
      return
    }
    setSubmitting(true)
    const res = await apiPost('/api/user/email/bind', { email: newEmail, code: newCode })
    setSubmitting(false)
    if (res.success) {
      toast.success('邮箱更换成功')
      onClose()
      onSuccess()
    } else {
      toast.error(res.error || '更换失败')
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="更换邮箱" size="sm">
      {step === 1 ? (
        <div className="space-y-4">
          <p className="text-dark-400 text-sm">为确保是您本人操作，请先验证原邮箱</p>
          <div className="setup-step">
            <div className="setup-step-title">📧 验证原邮箱</div>
            <p className="text-dark-500 text-sm mb-3">验证码将发送到 {currentEmail}</p>
            <div className="flex items-center gap-3">
              <Input
                placeholder="输入验证码"
                value={verifyCode}
                onChange={(e: ChangeEvent<HTMLInputElement>) => setVerifyCode(e.target.value)}
              />
              <Button
                variant="secondary"
                onClick={sendOldCode}
                disabled={countdown > 0 || sending}
              >
                {countdown > 0 ? `${countdown}秒` : sending ? '发送中...' : '发送验证码'}
              </Button>
            </div>
          </div>
          <Button className="w-full" onClick={verifyOldEmail} loading={submitting}>
            验证并继续
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          <Input
            label="新邮箱地址"
            type="email"
            placeholder="请输入新邮箱地址"
            value={newEmail}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setNewEmail(e.target.value)}
          />
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">新邮箱验证码</label>
            <div className="flex items-center gap-3">
              <Input
                placeholder="请输入验证码"
                value={newCode}
                onChange={(e: ChangeEvent<HTMLInputElement>) => setNewCode(e.target.value)}
              />
              <Button
                variant="secondary"
                onClick={sendNewCode}
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
            <Button className="flex-1" onClick={handleSubmit} loading={submitting}>
              确认更换
            </Button>
          </div>
        </div>
      )}
    </Modal>
  )
}
