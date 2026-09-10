import { getApiBaseUrl, getTenantConfig } from '@/shared/config'
import { storage } from '@/shared/storage'
import { t } from '@/locales'
import { logger } from '@/shared/logger'

type LoginInfo = {
  accessToken?: string
  tokenType?: string
}

/**
 * Upload a local file to `/client/file/upload` (multipart), matching legacy commonFun.uploadFile.
 * Returns the remote URL string on success.
 */
export function uploadFile(filePath: string): Promise<string | undefined> {
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
