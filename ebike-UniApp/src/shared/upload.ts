import { getApiBaseUrl, getTenantConfig } from '@/shared/config'
import { storage } from '@/shared/storage'
import { t } from '@/locales'
import { logger } from '@/shared/logger'
import { isNative, nativeHost } from '@/shared/nativeHost'

type LoginInfo = {
  accessToken?: string
  tokenType?: string
}

export type PhotoSource = 'camera' | 'album'

/**
 * 原生壳内选图并上传：一次桥调用完成（拍照/相册 → 签名 multipart → CDN URL）。
 * 不要先 `uni.chooseImage` 再调本函数，否则会弹两次选图。
 */
export async function pickAndUploadPhoto(
  source: PhotoSource = 'camera',
): Promise<string | undefined> {
  const host = nativeHost()
  if (!host) {
    uni.showModal({
      title: '提示',
      content: '原生拍照桥不可用（未检测到 App 壳）',
      showCancel: false,
    })
    return undefined
  }
  try {
    // 拍照过程不要 mask loading，部分机型会挡住相机/权限弹窗。
    const res = await host.capturePhoto({ source })
    const uri = String(res?.uri || '').trim()
    if (!uri) {
      uni.showModal({
        title: '提示',
        content: t('support.objectionUploadFail'),
        showCancel: false,
      })
      return undefined
    }
    return uri
  } catch (e) {
    logger.error('native pickAndUploadPhoto fail', e)
    const msg = e instanceof Error ? e.message : ''
    if (/cancel|PHOTO_CANCELLED|CANCELLED/i.test(msg)) {
      uni.showToast({ title: t('common.cancel'), icon: 'none' })
      return undefined
    }
    uni.showModal({
      title: '拍照失败',
      content: msg || t('support.objectionUploadFail'),
      showCancel: false,
    })
    return undefined
  }
}

/**
 * Upload a local file to `/client/file/upload` (multipart), matching legacy commonFun.uploadFile.
 * Returns the remote URL string on success.
 *
 * Rider WebView（isNative）：忽略 filePath，走原生 `capturePhoto`（选图 + 签名上传）。
 * 新代码请优先用 [pickAndUploadPhoto]，避免 chooseImage + upload 双弹窗。
 */
export async function uploadFile(filePath: string): Promise<string | undefined> {
  const path = String(filePath || '').trim()
  // 原生 chooseImage polyfill / 已上传 CDN 地址，勿再弹一次相机。
  if (/^https?:\/\//i.test(path) || path.startsWith('//')) {
    return path.startsWith('//') ? `https:${path}` : path
  }
  if (isNative()) {
    return pickAndUploadPhoto('camera')
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
