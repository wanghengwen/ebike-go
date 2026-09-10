import { t } from '@/locales'
import { logger } from '@/shared/logger'

export type LocationAuthResult = {
  ok: boolean
  latitude?: number
  longitude?: number
  systemAuth?: boolean
  miniAppAuth?: boolean
  isFuzzy?: boolean
}

/**
 * Legacy checkIsAuthLocation + toCheckLocationTips.
 * Returns ok=true with coords when usable; otherwise shows guidance and ok=false.
 */
export async function ensureLocationAuthorized(
  tipText?: string,
): Promise<LocationAuthResult> {
  const text = tipText || t('ride.needLocationAuth')

  // Probe setting for scope.userLocation when available (MP)
  let miniAppAuth: boolean | undefined
  try {
    const setting = await new Promise<UniApp.GetSettingSuccessResult>((resolve, reject) => {
      uni.getSetting({ success: resolve, fail: reject })
    })
    miniAppAuth = Boolean(setting.authSetting?.['scope.userLocation'])
  } catch {
    miniAppAuth = undefined
  }

  return new Promise((resolve) => {
    uni.getLocation({
      type: 'gcj02',
      isHighAccuracy: true,
      success: (res) => {
        resolve({
          ok: true,
          latitude: res.latitude,
          longitude: res.longitude,
          systemAuth: true,
          miniAppAuth: true,
          isFuzzy: Boolean((res as { isFuzzy?: boolean }).isFuzzy),
        })
      },
      fail: (err) => {
        logger.warn('ensureLocationAuthorized fail', err)
        const errMsg = String((err as { errMsg?: string })?.errMsg || '')
        const denied =
          /auth deny|authorize|permission|隐私|privacy/i.test(errMsg) || miniAppAuth === false

        // #ifdef MP-WEIXIN
        if (!denied && /system|locate|定位/i.test(errMsg)) {
          uni.showModal({
            showCancel: false,
            title: t('ride.needSystemLocation'),
          })
          resolve({ ok: false, systemAuth: false, miniAppAuth: true })
          return
        }
        // #endif

        if (denied || miniAppAuth === false) {
          uni.showModal({
            content: text,
            confirmText: t('ride.getLocation'),
            cancelText: t('common.cancel'),
            success: (r) => {
              if (r.confirm) uni.openSetting({})
            },
          })
          resolve({ ok: false, systemAuth: true, miniAppAuth: false })
          return
        }

        uni.showModal({
          showCancel: false,
          title: t('ride.needPreciseLocation'),
          content: t('ride.preciseLocationHint'),
          confirmText: t('common.confirm'),
        })
        resolve({ ok: false, isFuzzy: true })
      },
    })
  })
}
