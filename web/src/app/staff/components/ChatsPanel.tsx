'use client'

import { useState, useEffect, useRef } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Badge, Card, Input } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime, cn } from '@/lib/utils'
import { StaffInfo, LiveChat, ChatMessage } from './types'

/**
 * 在线咨询面板
 */
export function ChatsPanel({ staff }: { staff: StaffInfo }) {
  const [waitingChats, setWaitingChats] = useState<LiveChat[]>([])
  const [activeChat, setActiveChat] = useState<LiveChat | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [inputMessage, setInputMessage] = useState('')
  const [sending, setSending] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const pollingRef = useRef<NodeJS.Timeout | null>(null)

  // 加载等待中的聊天
  const loadWaitingChats = async () => {
    const res = await apiGet<{ chats: LiveChat[] }>('/api/staff/chats/waiting')
    if (res.success) {
      setWaitingChats(res.chats || [])
    }
  }

  useEffect(() => {
    loadWaitingChats()
    const interval = setInterval(loadWaitingChats, 5000)
    return () => clearInterval(interval)
  }, [])

  // 滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  // 接入聊天
  const acceptChat = async (chat: LiveChat) => {
    const res = await apiPost(`/api/staff/chat/${chat.id}/accept`)
    if (res.success) {
      setActiveChat({ ...chat, status: 1 })
      loadWaitingChats()
      startPolling(chat.session_id)
      toast.success('已接入对话')
    } else {
      toast.error(res.error || '接入失败')
    }
  }

  // 轮询消息
  const startPolling = (sessionId: string) => {
    if (pollingRef.current) clearInterval(pollingRef.current)

    const poll = async () => {
      const lastId = messages.length > 0 ? messages[messages.length - 1].id : 0
      const res = await apiGet<{ messages: ChatMessage[]; chat: LiveChat }>(
        `/api/staff/chat/${sessionId}/messages?after_id=${lastId}`
      )
      if (res.success) {
        if (res.messages && res.messages.length > 0) {
          setMessages(prev => [...prev, ...res.messages])
        }
        if (res.chat) {
          setActiveChat(res.chat)
          if (res.chat.status === 2) {
            // 聊天已结束
            if (pollingRef.current) clearInterval(pollingRef.current)
          }
        }
      }
    }

    poll()
    pollingRef.current = setInterval(poll, 2000)
  }

  useEffect(() => {
    return () => {
      if (pollingRef.current) clearInterval(pollingRef.current)
    }
  }, [])

  // 发送消息
  const sendMessage = async () => {
    if (!inputMessage.trim() || !activeChat) return

    setSending(true)
    const res = await apiPost<{ message: ChatMessage }>(
      `/api/staff/chat/${activeChat.session_id}/send`,
      { content: inputMessage }
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
    if (!activeChat) return

    const res = await apiPost(`/api/staff/chat/${activeChat.session_id}/end`)
    if (res.success) {
      if (pollingRef.current) clearInterval(pollingRef.current)
      setActiveChat(null)
      setMessages([])
      toast.success('对话已结束')
    }
  }

  return (
    <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* 左侧：等待列表 */}
        <Card title="等待接入" icon={<i className="fas fa-clock" />}>
          {waitingChats.length === 0 ? (
            <div className="text-center py-8 text-dark-400">
              <i className="fas fa-inbox text-3xl mb-2" />
              <p>暂无等待中的咨询</p>
            </div>
          ) : (
            <div className="space-y-2">
              {waitingChats.map((chat) => (
                <div
                  key={chat.id}
                  className="bg-dark-700/30 rounded-lg p-3 flex justify-between items-center"
                >
                  <div>
                    <div className="text-dark-100">{chat.username}</div>
                    <div className="text-dark-500 text-xs">
                      {formatDateTime(chat.created_at)}
                    </div>
                  </div>
                  <Button size="sm" onClick={() => acceptChat(chat)}>
                    接入
                  </Button>
                </div>
              ))}
            </div>
          )}
        </Card>

        {/* 右侧：聊天窗口 */}
        <div className="lg:col-span-2">
          <Card className="flex flex-col h-[500px]">
            {!activeChat ? (
              <div className="flex-1 flex items-center justify-center text-dark-400">
                <div className="text-center">
                  <i className="fas fa-comments text-4xl mb-3" />
                  <p>选择一个对话开始接待</p>
                </div>
              </div>
            ) : (
              <>
                {/* 聊天头部 */}
                <div className="flex items-center justify-between pb-4 border-b border-dark-700/50">
                  <div className="flex items-center gap-2">
                    <i className="fas fa-user text-primary-400" />
                    <span className="text-dark-100 font-medium">
                      {activeChat.username}
                    </span>
                    {activeChat.status === 2 && (
                      <Badge variant="default">已结束</Badge>
                    )}
                  </div>
                  {activeChat.status !== 2 && (
                    <Button size="sm" variant="danger" onClick={endChat}>
                      结束对话
                    </Button>
                  )}
                </div>

                {/* 消息列表 */}
                <div className="flex-1 overflow-y-auto py-4 space-y-3">
                  {messages.map((msg) => (
                    <div
                      key={msg.id}
                      className={cn(
                        'flex',
                        msg.sender_type === 'staff' ? 'justify-end' : 'justify-start'
                      )}
                    >
                      <div
                        className={cn(
                          'max-w-[70%] rounded-lg px-4 py-2',
                          msg.sender_type === 'staff'
                            ? 'bg-primary-500 text-white'
                            : msg.sender_type === 'system'
                            ? 'bg-dark-600/50 text-dark-300 text-center w-full max-w-full text-sm'
                            : 'bg-dark-700 text-dark-100'
                        )}
                      >
                        {msg.sender_type !== 'system' && msg.sender_type !== 'staff' && (
                          <div className="text-xs text-dark-400 mb-1">{msg.sender_name}</div>
                        )}
                        <div className="whitespace-pre-wrap break-words">{msg.content}</div>
                      </div>
                    </div>
                  ))}
                  <div ref={messagesEndRef} />
                </div>

                {/* 输入框 */}
                {activeChat.status !== 2 && (
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
                )}
              </>
            )}
          </Card>
        </div>
      </div>
    </motion.div>
  )
}
