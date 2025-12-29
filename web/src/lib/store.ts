import { create } from 'zustand'

/**
 * 用户信息接口
 */
interface UserInfo {
  id: number
  username: string
  email: string
  email_verified: boolean
  phone: string
  created_at: string
}

/**
 * 2FA 状态接口
 */
interface TwoFAStatus {
  enabled: boolean
  has_totp: boolean
  prefer_email_auth: boolean
}

/**
 * 应用状态存储
 */
interface AppState {
  // 用户信息
  user: UserInfo | null
  setUser: (user: UserInfo | null) => void

  // 2FA 状态
  twoFAStatus: TwoFAStatus | null
  setTwoFAStatus: (status: TwoFAStatus | null) => void

  // 登录状态
  isLoggedIn: boolean
  setIsLoggedIn: (value: boolean) => void

  // 加载状态
  isLoading: boolean
  setIsLoading: (value: boolean) => void
}

export const useAppStore = create<AppState>((set) => ({
  // 用户信息
  user: null,
  setUser: (user) => set({ user, isLoggedIn: !!user }),

  // 2FA 状态
  twoFAStatus: null,
  setTwoFAStatus: (twoFAStatus) => set({ twoFAStatus }),

  // 登录状态
  isLoggedIn: false,
  setIsLoggedIn: (isLoggedIn) => set({ isLoggedIn }),

  // 加载状态
  isLoading: false,
  setIsLoading: (isLoading) => set({ isLoading }),
}))
