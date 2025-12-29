'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Badge, Card, Input } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { cn } from '@/lib/utils'
import { SupportConfig, LiveChat, ChatMessage } from './types'

/**
 * 在线咨询标签页
 */
export function LiveChatTab({
  config,
  isLoggedIn,
  guestToken,
  setGuestToken,
}: {
  config: SupportConfig
  isLoggedIn: boolean
  guestToken: string
  setGuestToken: (token: string) => void
}) {
  const [chat, setChat] = useState<LiveChat | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [inputMessage, setInputMessage] = useState('')
  const [sending, setSending] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const pollingRef = useRef<NodeJS.Timeout | null>(null)
  // 使用 ref 存储最新的消息 ID，避免闭包问题
  const lastMessageIdRef = useRef<number>(0)

  // 滚动到底部
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  // 更新最新消息 ID
  useEffect(() => {
    if (messages.length > 0) {
      lastMessageIdRef.current = messages[messages.length - 1].id
    }
  }, [messages])

  // 轮询消息 - 使用 useCallback 避免重复创建
  const startPolling = useCallback((sessionId: string) => {
    // 清理旧的轮询
    if (pollingRef.current) {
      clearInterval(pollingRef.current)
      pollingRef.current = null
    }
    
    const poll = async () => {
      const res = await apiGet<{ messages: ChatMessage[]; chat: LiveChat }>(
        `/api/chat/${sessionId}/messages?after_id=${lastMessageIdRef.current}&guest_token=${guestToken}`
      )
      if (res.success) {
        if (res.messages && res.messages.length > 0) {
          setMessages(prev => [...prev, ...res.messages])
        }
        if (res.chat) {
          setChat(res.chat)
        }
      }
    }

    pollingRef.current = setInterval(poll, 3000)
  }, [guestToken])

  // 开始聊天
  const startChat = async () => {
    const res = await apiPost<{
      session_id: string
      guest_token: string
      welcome: string
    }>('/api/chat/start', { guest_token: guestToken })

    if (res.success) {
      if (res.guest_token) {
        setGuestToken(res.guest_token)
        localStorage.setItem('guest_token', res.guest_token)
      }
      setChat({ id: 0, session_id: res.session_id, status: 0, staff_name: '' })
      // 添加欢迎消息
      if (res.welcome) {
        setMessages([{
          id: 0,
          sender_type: 'system',
          sender_name: '系统',
          content: res.welcome,
          created_at: new Date().toISOString(),
        }])
      }
      // 开始轮询消息
      startPolling(res.session_id)
    } else {
      toast.error(res.error || '开始聊天失败')
    }
  }

  // 清理轮询
  useEffect(() => {
    return () => {
      if (pollingRef.current) {
        clearInterval(pollingRef.current)
        pollingRef.current = null
      }
    }
  }, [])

  // 发送消息
  const sendMessage = async () => {
    if (!inputMessage.trim() || !chat) return

    setSending(true)
    const res = await apiPost<{ message: ChatMessage }>(
      `/api/chat/${chat.session_id}/send`,
      { content: inputMessage, guest_token: guestToken }
    )

    if (res.success && res.message) {
      setMessages(prev => [...prev, res.message])
      setInputMessage('')
    } else {
      toast.error(res.error || '发送失败')
    }
    setSending(false)
  }

  // 结束聊天
  const endChat = async () => {
    if (!chat) return
    
    const res = await apiPost(`/api/chat/${chat.session_id}/end?guest_token=${guestToken}`)
    if (res.success) {
      if (pollingRef.current) clearInterval(pollingRef.current)
      setChat(null)
      setMessages([])
      toast.success('聊天已结束')
    }
  }

  // 未开始聊天
  if (!chat) {
    return (
      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
        <Card className="text-center py-12">
          <div className="text-6xl mb-4">💬</div>
          <h3 className="text-xl font-semibold text-dark-100 mb-2">在线客服</h3>
          <p className="text-dark-400 mb-6">
            {config.is_online
              ? `当前有 ${config.online_count} 位客服在线`
              : config.offline || '当前客服不在线，请留言或提交工单'}
          </p>
          {!isLoggedIn && !config.allow_guest ? (
            <div>
              <p className="text-amber-400 mb-4">请先登录后再进行咨询</p>
              <Button onClick={() => window.location.href = '/login/'}>
                去登录
              </Button>
            </div>
          ) : (
            <Button onClick={startChat} disabled={!config.is_online}>
              <i className="fas fa-comment-dots mr-2" />
              开始咨询
            </Button>
          )}
        </Card>
      </motion.div>
    )
  }

  // 聊天界面
  return (
    <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
      <Card className="flex flex-col h-[500px]">
        {/* 聊天头部 */}
        <div className="flex items-center justify-between pb-4 border-b border-dark-700/50">
          <div className="flex items-center gap-2">
            <i className="fas fa-headset text-primary-400" />
            <span className="text-dark-100 font-medium">
              {chat.staff_name || '等待客服接入...'}
            </span>
            {chat.status === 0 && (
              <Badge variant="warning">排队中</Badge>
            )}
            {chat.status === 1 && (
              <Badge variant="success">对话中</Badge>
            )}
          </div>
          <Button size="sm" variant="danger" onClick={endChat}>
            结束对话
          </Button>
        </div>

        {/* 消息列表 */}
        <div className="flex-1 overflow-y-auto py-4 space-y-4">
          {messages.map((msg) => (
            <div
              key={msg.id}
              className={cn(
                'flex',
                msg.sender_type === 'user' || msg.sender_type === 'guest'
                  ? 'justify-end'
                  : 'justify-start'
              )}
            >
              <div
                className={cn(
                  'max-w-[70%] rounded-lg px-4 py-2',
                  msg.sender_type === 'user' || msg.sender_type === 'guest'
                    ? 'bg-primary-500 text-white'
                    : msg.sender_type === 'system'
                    ? 'bg-dark-600/50 text-dark-300 text-center w-full max-w-full text-sm'
                    : 'bg-dark-700 text-dark-100'
                )}
              >
                {msg.sender_type !== 'system' && msg.sender_type !== 'user' && msg.sender_type !== 'guest' && (
                  <div className="text-xs text-dark-400 mb-1">{msg.sender_name}</div>
                )}
                <div className="whitespace-pre-wrap break-words">{msg.content}</div>
              </div>
            </div>
          ))}
          <div ref={messagesEndRef} />
        </div>

        {/* 输入框 */}
        <div className="pt-4 border-t border-dark-700/50 flex items-center gap-2">
          <div className="flex-1">
            <Input
              value={inputMessage}
              onChange={(e) => setInputMessage(e.target.value)}
              placeholder="输入消息..."
              onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && sendMessage()}
            />
          </div>
          <Button 
            onClick={sendMessage} 
            loading={sending} 
            disabled={!inputMessage.trim()}
            className="h-10 px-4 shrink-0"
          >
            <i className="fas fa-paper-plane" />
          </Button>
        </div>
      </Card>
    </motion.div>
  )
}
