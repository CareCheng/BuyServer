'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Input, Switch } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { PaymentConfig } from './types'

/**
 * 支付配置页面
 * 支持：支付宝当面付、微信支付、易支付、PayPal、Stripe、USDT
 */
export function PaymentPage() {
  const [config, setConfig] = useState<PaymentConfig>({})
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('paypal')
  const [form, setForm] = useState<Record<string, string | boolean | number>>({})

  // 加载支付配置
  const loadConfig = useCallback(async () => {
    const res = await apiGet<{ config: PaymentConfig }>('/api/admin/payment/config')
    if (res.success && res.config) setConfig(res.config)
    setLoading(false)
  }, [])

  useEffect(() => { loadConfig() }, [loadConfig])

  // 切换标签时更新表单
  useEffect(() => {
    const cfg = config[activeTab as keyof PaymentConfig] || {}
    setForm(cfg as Record<string, string | boolean | number>)
  }, [activeTab, config])

  // 保存配置
  const handleSave = async () => {
    const data: Record<string, unknown> = { payment_type: activeTab }
    
    if (activeTab === 'paypal') {
      data.paypal_enabled = form.enabled
      data.paypal_sandbox = form.sandbox
      data.paypal_client_id = form.client_id
      if (form.client_secret) data.paypal_client_secret = form.client_secret
      data.paypal_currency = form.currency || 'USD'
      data.paypal_return_url = form.return_url
      data.paypal_cancel_url = form.cancel_url
    } else if (activeTab === 'alipay_f2f') {
      data.alipay_enabled = form.enabled
      data.alipay_app_id = form.app_id
      if (form.private_key) data.alipay_private_key = form.private_key
      if (form.public_key) data.alipay_public_key = form.public_key
      data.alipay_notify_url = form.notify_url
    } else if (activeTab === 'wechat_pay') {
      data.wechat_enabled = form.enabled
      data.wechat_app_id = form.app_id
      data.wechat_mch_id = form.mch_id
      if (form.api_key) data.wechat_api_key = form.api_key
      data.wechat_notify_url = form.notify_url
    } else if (activeTab === 'yi_pay') {
      data.yipay_enabled = form.enabled
      data.yipay_api_url = form.api_url
      data.yipay_pid = form.pid
      if (form.key) data.yipay_key = form.key
      data.yipay_notify_url = form.notify_url
      data.yipay_return_url = form.return_url
    } else if (activeTab === 'stripe') {
      data.stripe_enabled = form.enabled
      data.stripe_publishable_key = form.publishable_key
      if (form.secret_key) data.stripe_secret_key = form.secret_key
      if (form.webhook_secret) data.stripe_webhook_secret = form.webhook_secret
      data.stripe_currency = form.currency || 'usd'
    } else if (activeTab === 'usdt') {
      data.usdt_enabled = form.enabled
      data.usdt_network = form.network || 'TRC20'
      data.usdt_wallet_address = form.wallet_address
      data.usdt_api_provider = form.api_provider || 'manual'
      if (form.api_key) data.usdt_api_key = form.api_key
      if (form.api_secret) data.usdt_api_secret = form.api_secret
      if (form.webhook_secret) data.usdt_webhook_secret = form.webhook_secret
      data.usdt_exchange_rate = Number(form.exchange_rate) || 7.2
      data.usdt_min_amount = Number(form.min_amount) || 1
      data.usdt_confirmations = Number(form.confirmations) || 1
    }

    const res = await apiPost('/api/admin/payment/config', data)
    if (res.success) { toast.success('配置已保存'); loadConfig() }
    else toast.error(res.error || '保存失败')
  }

  // 测试连接函数
  const testPayPal = async () => {
    const res = await apiPost('/api/admin/paypal/test', {})
    if (res.success) toast.success('PayPal连接测试成功')
    else toast.error(res.error || 'PayPal连接测试失败')
  }

  const testStripe = async () => {
    const res = await apiPost('/api/admin/stripe/test', {})
    if (res.success) toast.success('Stripe连接测试成功')
    else toast.error(res.error || 'Stripe连接测试失败')
  }

  const testUSDT = async () => {
    const res = await apiPost('/api/admin/usdt/test', {})
    if (res.success) toast.success('USDT连接测试成功')
    else toast.error(res.error || 'USDT连接测试失败')
  }

  const testAlipay = async () => {
    const res = await apiPost('/api/admin/alipay/test', {})
    if (res.success) toast.success('支付宝连接测试成功')
    else toast.error(res.error || '支付宝连接测试失败')
  }

  const testWechatPay = async () => {
    const res = await apiPost('/api/admin/wechat/test', {})
    if (res.success) toast.success('微信支付连接测试成功')
    else toast.error(res.error || '微信支付连接测试失败')
  }

  const testYiPay = async () => {
    const res = await apiPost('/api/admin/yipay/test', {})
    if (res.success) toast.success('易支付连接测试成功')
    else toast.error(res.error || '易支付连接测试失败')
  }

  if (loading) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>

  const tabs = [
    { key: 'paypal', label: 'PayPal', icon: 'fab fa-paypal' },
    { key: 'stripe', label: 'Stripe', icon: 'fab fa-stripe' },
    { key: 'usdt', label: 'USDT', icon: 'fab fa-bitcoin' },
    { key: 'alipay_f2f', label: '支付宝', icon: 'fab fa-alipay' },
    { key: 'wechat_pay', label: '微信支付', icon: 'fab fa-weixin' },
    { key: 'yi_pay', label: '易支付', icon: 'fas fa-credit-card' },
  ]

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-medium text-dark-100">支付配置</h2>
      
      {/* 标签切换 */}
      <div className="flex flex-wrap gap-2 border-b border-dark-700 pb-2">
        {tabs.map((tab) => (
          <button key={tab.key} onClick={() => setActiveTab(tab.key)} className={`px-4 py-2 rounded-t transition-colors flex items-center gap-2 ${activeTab === tab.key ? 'bg-primary-500/20 text-primary-400' : 'text-dark-400 hover:text-dark-200'}`}>
            <i className={tab.icon} />
            {tab.label}
          </button>
        ))}
      </div>

      <Card>
        {/* PayPal 配置 */}
        {activeTab === 'paypal' && (
          <div className="space-y-4">
            <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
              <p className="text-blue-400 text-sm">PayPal 是全球领先的在线支付平台。请前往 <a href="https://developer.paypal.com/" target="_blank" rel="noopener" className="underline">PayPal Developer</a> 创建应用获取 Client ID 和 Secret。</p>
            </div>
            <Switch checked={!!form.enabled} onChange={(checked) => setForm({ ...form, enabled: checked })} label="启用 PayPal 支付" description="开启后用户可使用 PayPal 进行支付" />
            <Switch checked={!!form.sandbox} onChange={(checked) => setForm({ ...form, sandbox: checked })} label="使用沙盒环境（测试）" description="开发测试时启用，正式上线请关闭" />
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Input label="Client ID" value={String(form.client_id || '')} onChange={(e) => setForm({ ...form, client_id: e.target.value })} />
              <Input label="Client Secret" type="password" value={String(form.client_secret || '')} onChange={(e) => setForm({ ...form, client_secret: e.target.value })} placeholder={config.paypal?.has_client_secret ? '******(已配置)' : ''} />
            </div>
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-1">货币类型</label>
              <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={String(form.currency || 'USD')} onChange={(e) => setForm({ ...form, currency: e.target.value })}>
                {['USD', 'CNY', 'EUR', 'GBP', 'CAD', 'AUD', 'JPY', 'HKD', 'SGD'].map(c => <option key={c} value={c}>{c}</option>)}
              </select>
            </div>
            <Input label="支付成功返回地址" value={String(form.return_url || '')} onChange={(e) => setForm({ ...form, return_url: e.target.value })} />
            <Input label="支付取消返回地址" value={String(form.cancel_url || '')} onChange={(e) => setForm({ ...form, cancel_url: e.target.value })} />
            <Button variant="secondary" onClick={testPayPal}><i className="fas fa-plug mr-2" />测试连接</Button>
          </div>
        )}

        {/* Stripe 配置 */}
        {activeTab === 'stripe' && (
          <div className="space-y-4">
            <div className="p-4 bg-purple-500/10 border border-purple-500/20 rounded-lg">
              <p className="text-purple-400 text-sm">Stripe 是国际领先的支付平台，支持信用卡、借记卡等多种支付方式。请前往 <a href="https://dashboard.stripe.com/apikeys" target="_blank" rel="noopener" className="underline">Stripe Dashboard</a> 获取 API 密钥。</p>
            </div>
            <Switch checked={!!form.enabled} onChange={(checked) => setForm({ ...form, enabled: checked })} label="启用 Stripe 支付" description="开启后用户可使用信用卡进行支付" />
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Input label="Publishable Key" value={String(form.publishable_key || '')} onChange={(e) => setForm({ ...form, publishable_key: e.target.value })} placeholder="pk_live_..." />
              <Input label="Secret Key" type="password" value={String(form.secret_key || '')} onChange={(e) => setForm({ ...form, secret_key: e.target.value })} placeholder={config.stripe?.has_secret_key ? '******(已配置)' : 'sk_live_...'} />
            </div>
            <Input label="Webhook Secret" type="password" value={String(form.webhook_secret || '')} onChange={(e) => setForm({ ...form, webhook_secret: e.target.value })} placeholder={config.stripe?.has_webhook_secret ? '******(已配置)' : 'whsec_...'} />
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-1">货币类型</label>
              <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={String(form.currency || 'usd')} onChange={(e) => setForm({ ...form, currency: e.target.value })}>
                {['usd', 'cny', 'eur', 'gbp', 'cad', 'aud', 'jpy', 'hkd', 'sgd'].map(c => <option key={c} value={c}>{c.toUpperCase()}</option>)}
              </select>
            </div>
            <Button variant="secondary" onClick={testStripe}><i className="fas fa-plug mr-2" />测试连接</Button>
          </div>
        )}

        {/* USDT 配置 */}
        {activeTab === 'usdt' && (
          <div className="space-y-4">
            <div className="p-4 bg-green-500/10 border border-green-500/20 rounded-lg">
              <p className="text-green-400 text-sm">USDT 是稳定币支付方式，支持 TRC20、ERC20 等网络。可选择手动模式或接入 NOWPayments、CoinGate 等支付网关。</p>
            </div>
            <Switch checked={!!form.enabled} onChange={(checked) => setForm({ ...form, enabled: checked })} label="启用 USDT 支付" description="开启后用户可使用 USDT 进行支付" />
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-dark-300 mb-1">网络类型</label>
                <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={String(form.network || 'TRC20')} onChange={(e) => setForm({ ...form, network: e.target.value })}>
                  <option value="TRC20">TRC20 (波场)</option>
                  <option value="ERC20">ERC20 (以太坊)</option>
                  <option value="BEP20">BEP20 (币安链)</option>
                  <option value="POLYGON">Polygon</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-dark-300 mb-1">API 提供商</label>
                <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={String(form.api_provider || 'manual')} onChange={(e) => setForm({ ...form, api_provider: e.target.value })}>
                  <option value="manual">手动模式（无自动确认）</option>
                  <option value="nowpayments">NOWPayments</option>
                  <option value="coingate">CoinGate</option>
                </select>
              </div>
            </div>
            <Input label="钱包地址" value={String(form.wallet_address || '')} onChange={(e) => setForm({ ...form, wallet_address: e.target.value })} placeholder="T..." />
            {form.api_provider !== 'manual' && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input label="API Key" type="password" value={String(form.api_key || '')} onChange={(e) => setForm({ ...form, api_key: e.target.value })} placeholder={config.usdt?.has_api_key ? '******(已配置)' : ''} />
                <Input label="API Secret" type="password" value={String(form.api_secret || '')} onChange={(e) => setForm({ ...form, api_secret: e.target.value })} placeholder={config.usdt?.has_api_secret ? '******(已配置)' : ''} />
              </div>
            )}
            {form.api_provider !== 'manual' && (
              <Input label="Webhook Secret" type="password" value={String(form.webhook_secret || '')} onChange={(e) => setForm({ ...form, webhook_secret: e.target.value })} placeholder={config.usdt?.has_webhook_secret ? '******(已配置)' : ''} />
            )}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Input label="汇率 (1 USDT = ? CNY)" type="number" value={String(form.exchange_rate || 7.2)} onChange={(e) => setForm({ ...form, exchange_rate: parseFloat(e.target.value) })} />
              <Input label="最小支付金额 (USDT)" type="number" value={String(form.min_amount || 1)} onChange={(e) => setForm({ ...form, min_amount: parseFloat(e.target.value) })} />
              <Input label="确认数" type="number" value={String(form.confirmations || 1)} onChange={(e) => setForm({ ...form, confirmations: parseInt(e.target.value) })} />
            </div>
            <Button variant="secondary" onClick={testUSDT}><i className="fas fa-plug mr-2" />测试连接</Button>
          </div>
        )}

        {/* 支付宝当面付配置 */}
        {activeTab === 'alipay_f2f' && (
          <div className="space-y-4">
            <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
              <p className="text-blue-400 text-sm">支付宝当面付适用于线下扫码支付场景。请前往 <a href="https://open.alipay.com/" target="_blank" rel="noopener" className="underline">支付宝开放平台</a> 创建应用并获取密钥。</p>
            </div>
            <Switch checked={!!form.enabled} onChange={(checked) => setForm({ ...form, enabled: checked })} label="启用支付宝当面付" description="开启后用户可使用支付宝扫码支付" />
            <Input label="应用 ID (App ID)" value={String(form.app_id || '')} onChange={(e) => setForm({ ...form, app_id: e.target.value })} />
            <div className="space-y-2">
              <label className="block text-sm font-medium text-dark-300">应用私钥</label>
              <textarea className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100 h-24 font-mono text-xs" value={String(form.private_key || '')} onChange={(e) => setForm({ ...form, private_key: e.target.value })} placeholder={config.alipay_f2f?.has_private_key ? '******(已配置，留空保持不变)' : '请输入 RSA2 私钥'} />
            </div>
            <div className="space-y-2">
              <label className="block text-sm font-medium text-dark-300">支付宝公钥</label>
              <textarea className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100 h-24 font-mono text-xs" value={String(form.public_key || '')} onChange={(e) => setForm({ ...form, public_key: e.target.value })} placeholder={config.alipay_f2f?.has_public_key ? '******(已配置，留空保持不变)' : '请输入支付宝公钥'} />
            </div>
            <Input label="异步通知地址" value={String(form.notify_url || '')} onChange={(e) => setForm({ ...form, notify_url: e.target.value })} placeholder="https://your-domain.com/api/payment/alipay/notify" />
            <Button variant="secondary" onClick={testAlipay}><i className="fas fa-plug mr-2" />测试连接</Button>
          </div>
        )}

        {/* 微信支付配置 */}
        {activeTab === 'wechat_pay' && (
          <div className="space-y-4">
            <div className="p-4 bg-green-500/10 border border-green-500/20 rounded-lg">
              <p className="text-green-400 text-sm">微信支付 Native 支付适用于 PC 网站扫码支付。请前往 <a href="https://pay.weixin.qq.com/" target="_blank" rel="noopener" className="underline">微信支付商户平台</a> 获取商户号和 API 密钥。</p>
            </div>
            <Switch checked={!!form.enabled} onChange={(checked) => setForm({ ...form, enabled: checked })} label="启用微信支付" description="开启后用户可使用微信扫码支付" />
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Input label="应用 ID (App ID)" value={String(form.app_id || '')} onChange={(e) => setForm({ ...form, app_id: e.target.value })} />
              <Input label="商户号 (Mch ID)" value={String(form.mch_id || '')} onChange={(e) => setForm({ ...form, mch_id: e.target.value })} />
            </div>
            <Input label="API 密钥" type="password" value={String(form.api_key || '')} onChange={(e) => setForm({ ...form, api_key: e.target.value })} placeholder={config.wechat_pay?.has_api_key ? '******(已配置，留空保持不变)' : ''} />
            <Input label="异步通知地址" value={String(form.notify_url || '')} onChange={(e) => setForm({ ...form, notify_url: e.target.value })} placeholder="https://your-domain.com/api/payment/wechat/notify" />
            <Button variant="secondary" onClick={testWechatPay}><i className="fas fa-plug mr-2" />测试连接</Button>
          </div>
        )}

        {/* 易支付配置 */}
        {activeTab === 'yi_pay' && (
          <div className="space-y-4">
            <div className="p-4 bg-orange-500/10 border border-orange-500/20 rounded-lg">
              <p className="text-orange-400 text-sm">易支付是第三方聚合支付平台，支持支付宝、微信等多种支付方式。请联系您的易支付服务商获取接口地址和密钥。</p>
            </div>
            <Switch checked={!!form.enabled} onChange={(checked) => setForm({ ...form, enabled: checked })} label="启用易支付" description="开启后用户可使用易支付进行支付" />
            <Input label="接口地址" value={String(form.api_url || '')} onChange={(e) => setForm({ ...form, api_url: e.target.value })} placeholder="https://pay.example.com/" />
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Input label="商户 ID (PID)" value={String(form.pid || '')} onChange={(e) => setForm({ ...form, pid: e.target.value })} />
              <Input label="商户密钥" type="password" value={String(form.key || '')} onChange={(e) => setForm({ ...form, key: e.target.value })} placeholder={config.yi_pay?.has_key ? '******(已配置，留空保持不变)' : ''} />
            </div>
            <Input label="异步通知地址" value={String(form.notify_url || '')} onChange={(e) => setForm({ ...form, notify_url: e.target.value })} placeholder="https://your-domain.com/api/payment/yipay/notify" />
            <Input label="同步返回地址" value={String(form.return_url || '')} onChange={(e) => setForm({ ...form, return_url: e.target.value })} placeholder="https://your-domain.com/order/success" />
            <Button variant="secondary" onClick={testYiPay}><i className="fas fa-plug mr-2" />测试连接</Button>
          </div>
        )}

        {/* 保存按钮 */}
        <div className="mt-6 pt-4 border-t border-dark-700">
          <Button onClick={handleSave}><i className="fas fa-save mr-2" />保存配置</Button>
        </div>
      </Card>
    </div>
  )
}
