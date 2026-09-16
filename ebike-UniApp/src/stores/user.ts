import { defineStore } from 'pinia'
import { storage } from '@/shared/storage'

export type UserInfo = {
  userId?: string
  /** Backend user pin used by ride/pay APIs */
  pin?: string
  phone?: string
  nickName?: string
  avatar?: string
  realNameStatus?: number
  [key: string]: unknown
}

export type LoginInfo = {
  accessToken?: string
  refreshToken?: string
  tokenType?: string
  expiresIn?: number
  openid?: string
  /** Rider WebView：已登录但无 token（token 只在原生侧）。 */
  nativeHost?: boolean
}

export const useUserStore = defineStore('user', {
  state: () => ({
    userInfo: {} as UserInfo,
    loginInfo: {} as LoginInfo,
  }),
  getters: {
    isLoggedIn: (s) => Boolean(s.loginInfo.accessToken || s.loginInfo.nativeHost),
  },
  actions: {
    hydrateFromStorage() {
      this.loginInfo = storage.get<LoginInfo>('loginInfo', {}) || {}
      this.userInfo = storage.get<UserInfo>('userInfo', {}) || {}
    },
    setLoginInfo(info: LoginInfo) {
      this.loginInfo = info
      storage.set('loginInfo', info)
    },
    setUserInfo(info: UserInfo) {
      this.userInfo = info
      storage.set('userInfo', info)
    },
    logout(keepServiceId = false) {
      this.loginInfo = {}
      this.userInfo = {}
      storage.remove('loginInfo')
      storage.remove('userInfo')
      if (!keepServiceId) storage.remove('serviceId')
    },
  },
})
