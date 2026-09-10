import { loginByPassword, getPersonInfo, sendSmsCode, getOpenIdByJsCode } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { t } from '@/locales'
import { navigate } from '@/shared/navigate'
import { getTenantConfig } from '@/shared/config'
import { logger } from '@/shared/logger'

function normalizePhone(phone: string): string {
  const raw = String(phone || '').trim()
  if (!raw) return raw
  if (raw.startsWith('+')) return raw
  return `+86-${raw.replace(/^86-?/, '')}`
}

export function uniLoginCode(): Promise<string> {
  return new Promise((resolve) => {
    uni.login({
      provider: 'weixin',
      success: (res) => resolve(res.code || ''),
      fail: (err) => {
        logger.warn('uni.login fail', err)
        resolve('')
      },
    })
  })
}

export function useAuth() {
  const user = useUserStore()

  async function loginWithSms(
    phone: string,
    code: string,
    opts: { skipNavigate?: boolean } = {},
  ) {
    uni.showLoading({ title: t('common.loading'), mask: true })
    try {
      const phoneNorm = normalizePhone(phone)
      const res = await loginByPassword({
        grant_type: 'phone_code',
        phone: phoneNorm,
        messageCode: code,
      })
      if (!res.success || !res.data) {
        const banned = res.code && Number(res.code) === 15012
        uni.showModal({
          title: t('auth.loginFail'),
          content: banned ? res.msg || t('auth.loginFail') : t('auth.loginFail'),
          showCancel: false,
        })
        return res
      }
      // Temporary save, then backfill openid like legacy SMS login
      user.setLoginInfo(res.data as never)
      const openId = await ensureOpenId()
      if (openId) {
        user.setLoginInfo({ ...(res.data as object), openid: openId } as never)
      }
      const profile = await getPersonInfo()
      if (profile.success && profile.data) user.setUserInfo(profile.data as never)
      if (!opts.skipNavigate) navigate('reLaunch', '/pages/home/home')
      return res
    } finally {
      uni.hideLoading()
    }
  }

  /**
   * WeChat mini-program one-tap phone login.
   * grant_type / partner code come from tenant.auth (backend may still use yudaoxing_app).
   */
  async function loginWithWechatPhone(detail: {
    encryptedData?: string
    iv?: string
    code?: string
  }) {
    uni.showLoading({ title: t('common.loading'), mask: true })
    try {
      const jsCode = await uniLoginCode()
      const tenant = getTenantConfig()
      const grantType = tenant.auth?.wechatGrantType || 'wechat_miniapp'
      const params: Record<string, unknown> = {
        grant_type: grantType,
        appId: tenant.appId || '',
        js_code: jsCode,
        encryptedData: detail.encryptedData,
        iv: detail.iv,
      }
      if (detail.code) params.phoneCode = detail.code
      // Legacy backend field name retained when partner code is configured
      if (tenant.auth?.wechatPartnerCode) {
        params.yudaoxingCode = tenant.auth.wechatPartnerCode
      }
      const res = await loginByPassword(params)
      if (!res.success || !res.data) {
        const banned = res.code && Number(res.code) === 15012
        uni.showToast({
          title: banned ? res.msg || t('auth.loginFail') : t('auth.loginFail'),
          icon: 'none',
        })
        return res
      }
      user.setLoginInfo(res.data as never)
      const profile = await getPersonInfo()
      if (profile.success && profile.data) user.setUserInfo(profile.data as never)
      navigate('reLaunch', '/pages/home/home')
      return res
    } finally {
      uni.hideLoading()
    }
  }

  async function ensureOpenId() {
    const jsCode = await uniLoginCode()
    if (!jsCode) return null
    const res = await getOpenIdByJsCode({ code: jsCode, type: 'WXLITE' })
    return res.success ? res.data : null
  }

  async function requestCode(phone: string) {
    return sendSmsCode({
      phone: normalizePhone(phone),
      scene: 1,
    })
  }

  function logout() {
    user.logout()
    // Match legacy setUserLogout: clear session; land on home
    navigate('reLaunch', '/pages/home/home')
  }

  return { loginWithSms, loginWithWechatPhone, ensureOpenId, requestCode, logout, uniLoginCode }
}
