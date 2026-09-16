import { getApiBaseUrl, getTenantConfig } from '@/shared/config'
import { storage } from '@/shared/storage'
import { t } from '@/locales'
import { logger } from '@/shared/logger'
import { isNative, nativeHost } from '@/shared/nativeHost'

type LoginInfo = {
  accessToken?: string
  tokenType?: string
}

/**
 * Upload a local file to `/client/file/upload` (multipart), matching legacy commonFun.uploadFile.
 * Returns the remote URL string on success.
 *
 * Rider WebView（isNative）：H5 没有 Bearer，改走原生 `capturePhoto`（选图 + 签名上传）。
 * 调用方若先 `chooseImage` 再 upload，在 native 下会再弹一次选图 —— 可接受；后续可统一改成本函数。
 */
export async function uploadFile(filePath: string): Promise<string | undefined> {
  if (isNative()) {
    const host = nativeHost()
    if (!host) {
      uni.showToast({ title: t('common.networkError'), icon: 'none' })
      return undefined
    }
    uni.showLoading({ title: t('common.loading'), mask: true })
    try {
      const res = await host.capturePhoto()
      const uri = String(res?.uri || '').trim()
      if (!uri) {
        uni.showToast({ title: t('common.networkError'), icon: 'none' })
        return undefined
      }
      return uri
    } catch (e) {
      logger.error('native upload fail', e)
      uni.showToast({ title: t('common.networkError'), icon: 'none' })
      return undefined
    } finally {
      uni.hideLoading()
    }
  }

  const tenant = getTenantConfig()
  const base = getApiBaseUrl().replace(/\/$/, '')
  const timestamp = Date.now()
  const login = storage.get<LoginInfo>('loginInfo', {}) || {}
  const token = login.accessToken || ''
  const authType = login.tokenType || 'Bearer'

  uni.showLoading({ title: t('common.loading'), mask: true })
  return new Promise((resolve) => {
    uni.uploadFile({
      url: `${base}/client/file/upload`,
      filePath,
      name: 'file',
      header: {
        _t: String(timestamp),
        Authorization: token ? `${authType} ${token}` : '',
      },
      formData: {
        traceId: `${timestamp}${(Math.random() * 1e11).toFixed(0)}`,
        tenantId: tenant.platformTenantId || '',
      },
      success(res) {
        uni.hideLoading()
        let parsed: { success?: boolean; data?: string; msg?: string } = {}
        try {
          parsed = JSON.parse(String(res.data || '{}'))
        } catch (e) {
          logger.warn('upload parse fail', e)
        }
        if (parsed.success && parsed.data) {
          resolve(String(parsed.data))
          return
        }
        uni.showToast({ title: parsed.msg || t('common.networkError'), icon: 'none' })
        resolve(undefined)
      },
      fail(err) {
        uni.hideLoading()
        logger.error('upload fail', err)
        uni.showToast({ title: t('common.networkError'), icon: 'none' })
        resolve(undefined)
      },
    })
  })
}
