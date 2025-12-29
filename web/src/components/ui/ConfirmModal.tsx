'use client'

import { useState, useCallback, createContext, useContext, ReactNode } from 'react'
import { Modal } from './Modal'
import { Button } from './Button'

/**
 * 确认弹窗配置
 */
interface ConfirmConfig {
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  variant?: 'primary' | 'danger' | 'warning'
  onConfirm: () => void | Promise<void>
  onCancel?: () => void
}

/**
 * 输入弹窗配置
 */
interface PromptConfig {
  title: string
  message?: string
  placeholder?: string
  defaultValue?: string
  confirmText?: string
  cancelText?: string
  inputType?: 'text' | 'email' | 'password' | 'number'
  required?: boolean
  onConfirm: (value: string) => void | Promise<void>
  onCancel?: () => void
}

/**
 * 弹窗上下文类型
 */
interface ModalContextType {
  showConfirm: (config: ConfirmConfig) => void
  showPrompt: (config: PromptConfig) => void
  hideModal: () => void
}

const ModalContext = createContext<ModalContextType | null>(null)

/**
 * 使用弹窗 Hook
 */
export function useModal() {
  const context = useContext(ModalContext)
  if (!context) {
    throw new Error('useModal 必须在 ModalProvider 内使用')
  }
  return context
}

/**
 * 弹窗提供者组件
 */
export function ModalProvider({ children }: { children: ReactNode }) {
  const [confirmConfig, setConfirmConfig] = useState<ConfirmConfig | null>(null)
  const [promptConfig, setPromptConfig] = useState<PromptConfig | null>(null)
  const [promptValue, setPromptValue] = useState('')
  const [loading, setLoading] = useState(false)

  const showConfirm = useCallback((config: ConfirmConfig) => {
    setConfirmConfig(config)
  }, [])

  const showPrompt = useCallback((config: PromptConfig) => {
    setPromptConfig(config)
    setPromptValue(config.defaultValue || '')
  }, [])

  const hideModal = useCallback(() => {
    setConfirmConfig(null)
    setPromptConfig(null)
    setPromptValue('')
    setLoading(false)
  }, [])

  // 处理确认弹窗确认
  const handleConfirm = async () => {
    if (!confirmConfig) return
    setLoading(true)
    try {
      await confirmConfig.onConfirm()
      hideModal()
    } catch {
      setLoading(false)
    }
  }

  // 处理确认弹窗取消
  const handleConfirmCancel = () => {
    confirmConfig?.onCancel?.()
    hideModal()
  }

  // 处理输入弹窗确认
  const handlePromptConfirm = async () => {
    if (!promptConfig) return
    if (promptConfig.required && !promptValue.trim()) return
    setLoading(true)
    try {
      await promptConfig.onConfirm(promptValue)
      hideModal()
    } catch {
      setLoading(false)
    }
  }

  // 处理输入弹窗取消
  const handlePromptCancel = () => {
    promptConfig?.onCancel?.()
    hideModal()
  }

  // 获取按钮变体样式
  const getButtonVariant = (variant?: string) => {
    switch (variant) {
      case 'danger': return 'danger'
      case 'warning': return 'warning'
      default: return 'primary'
    }
  }

  return (
    <ModalContext.Provider value={{ showConfirm, showPrompt, hideModal }}>
      {children}

      {/* 确认弹窗 */}
      <Modal
        isOpen={!!confirmConfig}
        onClose={handleConfirmCancel}
        title={confirmConfig?.title || '确认'}
        size="sm"
      >
        <div className="space-y-4">
          <p className="text-dark-300">{confirmConfig?.message}</p>
          <div className="flex justify-end gap-3">
            <Button variant="secondary" onClick={handleConfirmCancel} disabled={loading}>
              {confirmConfig?.cancelText || '取消'}
            </Button>
            <Button
              variant={getButtonVariant(confirmConfig?.variant)}
              onClick={handleConfirm}
              loading={loading}
            >
              {confirmConfig?.confirmText || '确认'}
            </Button>
          </div>
        </div>
      </Modal>

      {/* 输入弹窗 */}
      <Modal
        isOpen={!!promptConfig}
        onClose={handlePromptCancel}
        title={promptConfig?.title || '输入'}
        size="sm"
      >
        <div className="space-y-4">
          {promptConfig?.message && (
            <p className="text-dark-300 text-sm">{promptConfig.message}</p>
          )}
          <input
            type={promptConfig?.inputType || 'text'}
            value={promptValue}
            onChange={(e) => setPromptValue(e.target.value)}
            placeholder={promptConfig?.placeholder}
            className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 placeholder-dark-500 focus:outline-none focus:border-primary-500"
            autoFocus
            onKeyDown={(e) => {
              if (e.key === 'Enter') handlePromptConfirm()
              if (e.key === 'Escape') handlePromptCancel()
            }}
          />
          <div className="flex justify-end gap-3">
            <Button variant="secondary" onClick={handlePromptCancel} disabled={loading}>
              {promptConfig?.cancelText || '取消'}
            </Button>
            <Button
              variant="primary"
              onClick={handlePromptConfirm}
              loading={loading}
              disabled={promptConfig?.required && !promptValue.trim()}
            >
              {promptConfig?.confirmText || '确认'}
            </Button>
          </div>
        </div>
      </Modal>
    </ModalContext.Provider>
  )
}

/**
 * 独立的确认弹窗组件（用于不使用 Provider 的场景）
 */
interface ConfirmModalProps {
  isOpen: boolean
  onClose: () => void
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  variant?: 'primary' | 'danger' | 'warning'
  onConfirm: () => void | Promise<void>
  loading?: boolean
}

export function ConfirmModal({
  isOpen,
  onClose,
  title,
  message,
  confirmText = '确认',
  cancelText = '取消',
  variant = 'primary',
  onConfirm,
  loading = false,
}: ConfirmModalProps) {
  const getButtonVariant = () => {
    switch (variant) {
      case 'danger': return 'danger'
      case 'warning': return 'warning'
      default: return 'primary'
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={title} size="sm">
      <div className="space-y-4">
        <p className="text-dark-300">{message}</p>
        <div className="flex justify-end gap-3">
          <Button variant="secondary" onClick={onClose} disabled={loading}>
            {cancelText}
          </Button>
          <Button variant={getButtonVariant()} onClick={onConfirm} loading={loading}>
            {confirmText}
          </Button>
        </div>
      </div>
    </Modal>
  )
}

/**
 * 独立的输入弹窗组件
 */
interface PromptModalProps {
  isOpen: boolean
  onClose: () => void
  title: string
  message?: string
  placeholder?: string
  defaultValue?: string
  confirmText?: string
  cancelText?: string
  inputType?: 'text' | 'email' | 'password' | 'number'
  required?: boolean
  onConfirm: (value: string) => void | Promise<void>
  loading?: boolean
}

export function PromptModal({
  isOpen,
  onClose,
  title,
  message,
  placeholder,
  defaultValue = '',
  confirmText = '确认',
  cancelText = '取消',
  inputType = 'text',
  required = false,
  onConfirm,
  loading = false,
}: PromptModalProps) {
  const [value, setValue] = useState(defaultValue)

  // 当弹窗打开时重置值
  const handleClose = () => {
    setValue(defaultValue)
    onClose()
  }

  const handleConfirm = async () => {
    if (required && !value.trim()) return
    await onConfirm(value)
    setValue(defaultValue)
  }

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title={title} size="sm">
      <div className="space-y-4">
        {message && <p className="text-dark-300 text-sm">{message}</p>}
        <input
          type={inputType}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder={placeholder}
          className="w-full px-4 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-100 placeholder-dark-500 focus:outline-none focus:border-primary-500"
          autoFocus
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleConfirm()
            if (e.key === 'Escape') handleClose()
          }}
        />
        <div className="flex justify-end gap-3">
          <Button variant="secondary" onClick={handleClose} disabled={loading}>
            {cancelText}
          </Button>
          <Button
            variant="primary"
            onClick={handleConfirm}
            loading={loading}
            disabled={required && !value.trim()}
          >
            {confirmText}
          </Button>
        </div>
      </div>
    </Modal>
  )
}
