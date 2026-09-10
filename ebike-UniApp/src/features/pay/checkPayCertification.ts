import { getPersonInfo } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { navigate } from '@/shared/navigate'
import { t } from '@/locales'
import { logger } from '@/shared/logger'

/**
 * Legacy payMixins.checkUserCertificationStatus:
 * if izNeedAuth && !izAuth → warn under-16, allow continue or go verified.
 */
export function checkPayCertification(): Promise<boolean> {
  return new Promise((resolve) => {
    const user = useUserStore()
    user.hydrateFromStorage()
    const run = async () => {
      let profile = (user.userInfo || {}) as Record<string, unknown>
      try {
        const res = await getPersonInfo()
        if (res.success && res.data) {
          user.setUserInfo(res.data as never)
          profile = res.data as Record<string, unknown>
        }
      } catch (e) {
        logger.warn('checkPayCertification soft fail', e)
      }
      const need = profile.izNeedAuth === true || profile.izNeedAuth === 1
      const authed = Boolean(profile.izAuth || profile.authName)
      if (!profile.pin || !need || authed) {
        resolve(true)
        return
      }
      uni.showModal({
        title: t('auth.verifiedTitle'),
        content: t('pay.needAuthAgeTip'),
        confirmText: t('auth.verifiedTitle'),
        cancelText: t('pay.continuePay'),
        success: (r) => {
          if (r.confirm) {
            navigate('to', '/pages/auth/verified')
            resolve(false)
          } else {
            resolve(true)
          }
        },
        fail: () => resolve(true),
      })
    }
    void run()
  })
}
