'use client';

import React from 'react';

interface ToggleProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
  size?: 'sm' | 'md' | 'lg';
  label?: string;
  labelPosition?: 'left' | 'right';
  className?: string;
}

/**
 * 通用滑块开关组件
 * @param checked - 是否选中
 * @param onChange - 状态变化回调
 * @param disabled - 是否禁用
 * @param size - 尺寸：sm/md/lg
 * @param label - 标签文本
 * @param labelPosition - 标签位置：left/right
 * @param className - 额外样式类
 */
export default function Toggle({
  checked,
  onChange,
  disabled = false,
  size = 'md',
  label,
  labelPosition = 'right',
  className = '',
}: ToggleProps) {
  // 根据尺寸设置样式
  const sizeStyles = {
    sm: {
      track: 'w-8 h-4',
      thumb: 'w-3 h-3',
      translate: 'translate-x-4',
      thumbOffset: 'translate-x-0.5',
    },
    md: {
      track: 'w-11 h-6',
      thumb: 'w-5 h-5',
      translate: 'translate-x-5',
      thumbOffset: 'translate-x-0.5',
    },
    lg: {
      track: 'w-14 h-7',
      thumb: 'w-6 h-6',
      translate: 'translate-x-7',
      thumbOffset: 'translate-x-0.5',
    },
  };

  const styles = sizeStyles[size];

  const handleClick = () => {
    if (!disabled) {
      onChange(!checked);
    }
  };

  const toggle = (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={handleClick}
      className={`
        relative inline-flex shrink-0 cursor-pointer rounded-full
        transition-colors duration-200 ease-in-out
        focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
        ${styles.track}
        ${checked ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'}
        ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
      `}
    >
      <span
        className={`
          pointer-events-none inline-block rounded-full bg-white shadow-lg
          transform transition-transform duration-200 ease-in-out
          ${styles.thumb}
          ${checked ? styles.translate : styles.thumbOffset}
          ${disabled ? '' : ''}
        `}
        style={{ marginTop: size === 'sm' ? '2px' : '2px' }}
      />
    </button>
  );

  if (!label) {
    return <div className={className}>{toggle}</div>;
  }

  return (
    <label
      className={`
        inline-flex items-center gap-2 cursor-pointer
        ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
        ${className}
      `}
    >
      {labelPosition === 'left' && (
        <span className="text-sm text-gray-700 dark:text-gray-300">{label}</span>
      )}
      {toggle}
      {labelPosition === 'right' && (
        <span className="text-sm text-gray-700 dark:text-gray-300">{label}</span>
      )}
    </label>
  );
}
