'use client'

/**
 * 客服管理模块通用组件
 */

/**
 * 开关组件
 */
export function Toggle({
  checked,
  onChange,
  label,
  description,
}: {
  checked: boolean
  onChange: (checked: boolean) => void
  label: string
  description?: string
}) {
  return (
    <label className="flex items-center justify-between p-3 bg-dark-700/30 rounded-lg cursor-pointer hover:bg-dark-700/50 transition-colors">
      <div className="flex-1">
        <span className="text-dark-200 font-medium">{label}</span>
        {description && (
          <p className="text-dark-500 text-xs mt-0.5">{description}</p>
        )}
      </div>
      <div
        className={`relative w-11 h-6 rounded-full transition-colors ${
          checked ? 'bg-primary-500' : 'bg-dark-600'
        }`}
        onClick={() => onChange(!checked)}
      >
        <div
          className={`absolute top-1 w-4 h-4 bg-white rounded-full transition-transform ${
            checked ? 'translate-x-6' : 'translate-x-1'
          }`}
        />
      </div>
    </label>
  )
}

/**
 * 标签按钮组件
 */
export function TabButton({
  active,
  onClick,
  icon,
  label,
}: {
  active: boolean
  onClick: () => void
  icon: string
  label: string
}) {
  return (
    <button
      onClick={onClick}
      className={`
        inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg font-medium text-sm
        transition-all duration-200 whitespace-nowrap
        ${active 
          ? 'bg-primary-500 text-white shadow-lg shadow-primary-500/25' 
          : 'bg-dark-700/50 text-dark-300 hover:bg-dark-700 hover:text-dark-100 border border-dark-600/50'
        }
      `}
    >
      <i className={`fas ${icon}`} />
      <span>{label}</span>
    </button>
  )
}
