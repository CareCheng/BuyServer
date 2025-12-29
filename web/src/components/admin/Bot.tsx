'use client'

import { useState, useEffect, useCallback, ChangeEvent } from 'react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Modal, Badge, Card, Input } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import Toggle from '@/components/common/Toggle'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

/**
 * 自动回复规则接口
 */
interface AutoReplyRule {
  id: number
  name: string
  keywords: string
  match_type: string  // exact, contains, regex
  reply_content: string
  reply_type: string  // text, transfer
  priority: number
  status: number
  hit_count: number
  created_at: string
}

/**
 * 自动回复配置接口
 */
interface AutoReplyConfig {
  enabled: boolean
  welcome_message: string
  default_reply: string
  transfer_message: string
  work_start_time: string
  work_end_time: string
}

/**
 * 自动回复日志接口
 */
interface AutoReplyLog {
  id: number
  user_id: number
  username: string
  question: string
  answer: string
  rule_id: number
  rule_name: string
  is_transferred: boolean
  created_at: string
}

/**
 * 智能客服管理页面
 */
export function BotPage() {
  const [activeTab, setActiveTab] = useState<'rules' | 'logs' | 'config'>('rules')
  const [rules, setRules] = useState<AutoReplyRule[]>([])
  const [logs, setLogs] = useState<AutoReplyLog[]>([])
  const [config, setConfig] = useState<AutoReplyConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const pageSize = 20

  // 规则编辑弹窗
  const [showRuleModal, setShowRuleModal] = useState(false)
  const [editingRule, setEditingRule] = useState<AutoReplyRule | null>(null)
  const [ruleForm, setRuleForm] = useState({
    name: '',
    keywords: '',
    match_type: 'contains',
    reply_content: '',
    reply_type: 'text',
    priority: 0,
    status: 1,
  })

  // 配置表单
  const [configForm, setConfigForm] = useState<AutoReplyConfig>({
    enabled: false,
    welcome_message: '',
    default_reply: '',
    transfer_message: '',
    work_start_time: '09:00',
    work_end_time: '18:00',
  })

  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<AutoReplyRule | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  // 加载数据
  const loadData = useCallback(async () => {
    setLoading(true)
    if (activeTab === 'rules') {
      await loadRules()
    } else if (activeTab === 'logs') {
      await loadLogs()
    } else {
      await loadConfig()
    }
    setLoading(false)
  }, [activeTab, page])

  useEffect(() => {
    loadData()
  }, [loadData])

  const loadRules = async () => {
    const res = await apiGet<{ rules: AutoReplyRule[] }>('/api/admin/auto-reply/rules')
    if (res.success) {
      setRules(res.rules || [])
    }
  }

  const loadLogs = async () => {
    const res = await apiGet<{ logs: AutoReplyLog[]; total: number }>(
      `/api/admin/auto-reply/logs?page=${page}&page_size=${pageSize}`
    )
    if (res.success) {
      setLogs(res.logs || [])
      setTotal(res.total || 0)
    }
  }

  const loadConfig = async () => {
    const res = await apiGet<{ config: AutoReplyConfig }>('/api/admin/auto-reply/config')
    if (res.success && res.config) {
      setConfig(res.config)
      setConfigForm(res.config)
    }
  }

  // 打开规则编辑弹窗
  const openRuleModal = (rule?: AutoReplyRule) => {
    if (rule) {
      setEditingRule(rule)
      setRuleForm({
        name: rule.name,
        keywords: rule.keywords,
        match_type: rule.match_type,
        reply_content: rule.reply_content,
        reply_type: rule.reply_type,
        priority: rule.priority,
        status: rule.status,
      })
    } else {
      setEditingRule(null)
      setRuleForm({
        name: '',
        keywords: '',
        match_type: 'contains',
        reply_content: '',
        reply_type: 'text',
        priority: 0,
        status: 1,
      })
    }
    setShowRuleModal(true)
  }

  // 保存规则
  const handleSaveRule = async () => {
    if (!ruleForm.name.trim()) {
      toast.error('请输入规则名称')
      return
    }
    if (!ruleForm.keywords.trim()) {
      toast.error('请输入关键词')
      return
    }
    if (!ruleForm.reply_content.trim() && ruleForm.reply_type === 'text') {
      toast.error('请输入回复内容')
      return
    }
    const res = editingRule
      ? await apiPut(`/api/admin/auto-reply/rule/${editingRule.id}`, ruleForm)
      : await apiPost('/api/admin/auto-reply/rule', ruleForm)
    if (res.success) {
      toast.success(editingRule ? '规则已更新' : '规则已创建')
      setShowRuleModal(false)
      loadRules()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 打开删除确认弹窗
  const openDeleteConfirm = (rule: AutoReplyRule) => {
    setDeleteTarget(rule)
    setShowDeleteConfirm(true)
  }

  // 删除规则
  const handleDeleteRule = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    const res = await apiDelete(`/api/admin/auto-reply/rule/${deleteTarget.id}`)
    setDeleteLoading(false)
    if (res.success) {
      toast.success('规则已删除')
      setShowDeleteConfirm(false)
      setDeleteTarget(null)
      loadRules()
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  // 切换规则状态
  const handleToggleStatus = async (rule: AutoReplyRule) => {
    const res = await apiPut(`/api/admin/auto-reply/rule/${rule.id}`, {
      ...rule,
      status: rule.status === 1 ? 0 : 1,
    })
    if (res.success) {
      toast.success(rule.status === 1 ? '规则已禁用' : '规则已启用')
      loadRules()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 保存配置
  const handleSaveConfig = async () => {
    const res = await apiPost('/api/admin/auto-reply/config', configForm as unknown as Record<string, unknown>)
    if (res.success) {
      toast.success('配置已保存')
      loadConfig()
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  // 获取匹配类型标签
  const getMatchTypeLabel = (type: string) => {
    const types: Record<string, string> = {
      exact: '精确匹配',
      contains: '包含匹配',
      regex: '正则匹配',
    }
    return types[type] || type
  }

  const totalPages = Math.ceil(total / pageSize)

  if (loading && rules.length === 0 && logs.length === 0 && !config) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 标签切换 */}
      <div className="flex gap-2 border-b border-dark-700/50 pb-4">
        <button
          onClick={() => { setActiveTab('rules'); setPage(1) }}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'rules'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-robot mr-2" />
          自动回复规则
        </button>
        <button
          onClick={() => { setActiveTab('logs'); setPage(1) }}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'logs'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-history mr-2" />
          回复日志
        </button>
        <button
          onClick={() => { setActiveTab('config'); setPage(1) }}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'config'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-cog mr-2" />
          基础配置
        </button>
      </div>

      {/* 自动回复规则 */}
      {activeTab === 'rules' && (
        <Card
          title="自动回复规则"
          icon={<i className="fas fa-robot" />}
          action={
            <Button size="sm" onClick={() => openRuleModal()}>
              <i className="fas fa-plus mr-1" />
              添加规则
            </Button>
          }
        >
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">规则名称</th>
                  <th className="pb-3 font-medium">关键词</th>
                  <th className="pb-3 font-medium">匹配方式</th>
                  <th className="pb-3 font-medium">回复类型</th>
                  <th className="pb-3 font-medium">优先级</th>
                  <th className="pb-3 font-medium">命中次数</th>
                  <th className="pb-3 font-medium">状态</th>
                  <th className="pb-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {rules.map((rule) => (
                  <motion.tr
                    key={rule.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    className="border-b border-dark-700/50"
                  >
                    <td className="py-3">{rule.name}</td>
                    <td className="py-3 text-sm text-dark-400 max-w-xs truncate">{rule.keywords}</td>
                    <td className="py-3 text-sm">{getMatchTypeLabel(rule.match_type)}</td>
                    <td className="py-3">
                      <Badge variant={rule.reply_type === 'transfer' ? 'warning' : 'info'}>
                        {rule.reply_type === 'transfer' ? '转人工' : '文本回复'}
                      </Badge>
                    </td>
                    <td className="py-3">{rule.priority}</td>
                    <td className="py-3 text-green-400">{rule.hit_count}</td>
                    <td className="py-3">
                      <Badge variant={rule.status === 1 ? 'success' : 'danger'}>
                        {rule.status === 1 ? '启用' : '禁用'}
                      </Badge>
                    </td>
                    <td className="py-3">
                      <div className="flex gap-1">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleToggleStatus(rule)}
                          title={rule.status === 1 ? '禁用' : '启用'}
                        >
                          <i className={`fas fa-${rule.status === 1 ? 'pause' : 'play'} text-yellow-400`} />
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => openRuleModal(rule)}>
                          <i className="fas fa-edit" />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          className="text-red-400"
                          onClick={() => openDeleteConfirm(rule)}
                        >
                          <i className="fas fa-trash" />
                        </Button>
                      </div>
                    </td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* 回复日志 */}
      {activeTab === 'logs' && (
        <Card title="回复日志" icon={<i className="fas fa-history" />}>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">用户</th>
                  <th className="pb-3 font-medium">问题</th>
                  <th className="pb-3 font-medium">回复</th>
                  <th className="pb-3 font-medium">匹配规则</th>
                  <th className="pb-3 font-medium">是否转人工</th>
                  <th className="pb-3 font-medium">时间</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {logs.map((log) => (
                  <tr key={log.id} className="border-b border-dark-700/50">
                    <td className="py-3">{log.username || '访客'}</td>
                    <td className="py-3 text-dark-400 max-w-xs truncate">{log.question}</td>
                    <td className="py-3 text-dark-400 max-w-xs truncate">{log.answer}</td>
                    <td className="py-3 text-sm">{log.rule_name || '-'}</td>
                    <td className="py-3">
                      <Badge variant={log.is_transferred ? 'warning' : 'info'}>
                        {log.is_transferred ? '是' : '否'}
                      </Badge>
                    </td>
                    <td className="py-3 text-sm text-dark-400">{formatDateTime(log.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* 分页 */}
          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-4">
              <Button size="sm" variant="ghost" disabled={page === 1} onClick={() => setPage(p => p - 1)}>
                上一页
              </Button>
              <span className="px-4 py-2 text-dark-400">{page} / {totalPages}</span>
              <Button size="sm" variant="ghost" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
                下一页
              </Button>
            </div>
          )}
        </Card>
      )}

      {/* 基础配置 */}
      {activeTab === 'config' && (
        <Card title="基础配置" icon={<i className="fas fa-cog" />}>
          <div className="space-y-6 max-w-2xl">
            <div className="flex items-center justify-between p-4 bg-dark-700/30 rounded-lg">
              <div>
                <div className="text-dark-200 font-medium">启用智能客服</div>
                <div className="text-sm text-dark-400">开启后将自动回复用户消息</div>
              </div>
              <Toggle
                checked={configForm.enabled}
                onChange={(checked) => setConfigForm({ ...configForm, enabled: checked })}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-dark-300 mb-2">欢迎语</label>
              <textarea
                value={configForm.welcome_message}
                onChange={(e) => setConfigForm({ ...configForm, welcome_message: e.target.value })}
                className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200 h-24 resize-none"
                placeholder="用户首次发起对话时的欢迎语"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-dark-300 mb-2">默认回复</label>
              <textarea
                value={configForm.default_reply}
                onChange={(e) => setConfigForm({ ...configForm, default_reply: e.target.value })}
                className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200 h-24 resize-none"
                placeholder="无法匹配任何规则时的默认回复"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-dark-300 mb-2">转人工提示</label>
              <textarea
                value={configForm.transfer_message}
                onChange={(e) => setConfigForm({ ...configForm, transfer_message: e.target.value })}
                className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200 h-24 resize-none"
                placeholder="转接人工客服时的提示语"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-dark-300 mb-2">工作开始时间</label>
                <Input
                  type="time"
                  value={configForm.work_start_time}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfigForm({ ...configForm, work_start_time: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-dark-300 mb-2">工作结束时间</label>
                <Input
                  type="time"
                  value={configForm.work_end_time}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setConfigForm({ ...configForm, work_end_time: e.target.value })}
                />
              </div>
            </div>

            <Button onClick={handleSaveConfig}>
              <i className="fas fa-save mr-2" />
              保存配置
            </Button>
          </div>
        </Card>
      )}

      {/* 规则编辑弹窗 */}
      <Modal
        isOpen={showRuleModal}
        onClose={() => setShowRuleModal(false)}
        title={editingRule ? '编辑规则' : '添加规则'}
      >
        <div className="space-y-4">
          <Input
            label="规则名称"
            placeholder="请输入规则名称"
            value={ruleForm.name}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setRuleForm({ ...ruleForm, name: e.target.value })}
          />
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">关键词</label>
            <textarea
              value={ruleForm.keywords}
              onChange={(e) => setRuleForm({ ...ruleForm, keywords: e.target.value })}
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200 h-20 resize-none"
              placeholder="多个关键词用逗号分隔"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">匹配方式</label>
            <select
              value={ruleForm.match_type}
              onChange={(e) => setRuleForm({ ...ruleForm, match_type: e.target.value })}
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200"
            >
              <option value="contains">包含匹配</option>
              <option value="exact">精确匹配</option>
              <option value="regex">正则匹配</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">回复类型</label>
            <select
              value={ruleForm.reply_type}
              onChange={(e) => setRuleForm({ ...ruleForm, reply_type: e.target.value })}
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200"
            >
              <option value="text">文本回复</option>
              <option value="transfer">转人工</option>
            </select>
          </div>
          {ruleForm.reply_type === 'text' && (
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-2">回复内容</label>
              <textarea
                value={ruleForm.reply_content}
                onChange={(e) => setRuleForm({ ...ruleForm, reply_content: e.target.value })}
                className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200 h-24 resize-none"
                placeholder="请输入回复内容"
              />
            </div>
          )}
          <Input
            label="优先级"
            type="number"
            placeholder="数字越大优先级越高"
            value={ruleForm.priority.toString()}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setRuleForm({ ...ruleForm, priority: parseInt(e.target.value) || 0 })}
          />
          <Toggle
            checked={ruleForm.status === 1}
            onChange={(checked) => setRuleForm({ ...ruleForm, status: checked ? 1 : 0 })}
            label="启用规则"
          />
          <Button className="w-full" onClick={handleSaveRule}>
            {editingRule ? '保存修改' : '创建规则'}
          </Button>
        </div>
      </Modal>

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setDeleteTarget(null) }}
        title="删除规则"
        message={`确定要删除规则 "${deleteTarget?.name}" 吗？`}
        confirmText="删除"
        variant="danger"
        onConfirm={handleDeleteRule}
        loading={deleteLoading}
      />
    </div>
  )
}
