import sha256 from 'crypto-js/sha256'
import encUtf8 from 'crypto-js/enc-utf8'
import encBase64 from 'crypto-js/enc-base64'
import { getAcceptLanguage, t } from '@/locales'
import { getApiBaseUrl, getTenantConfig } from '@/shared/config'
import { logger } from '@/shared/logger'
import { storage } from '@/shared/storage'

export type RequestConfig = {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: Record<string, unknown> | unknown
  header?: Record<string, string>
  /** When false, skip Bearer; /oauth/token still uses Basic. */
  auth?: boolean
  showError?: boolean
}

export type ApiResult<T = unknown> = {
  success: boolean
  code?: string
  msg?: string
  data?: T
  isHideApiError?: boolean
}

type LoginInfo = {
  accessToken?: string
  refreshToken?: string
  tokenType?: string
  openid?: string
  [key: string]: unknown
}

const REFRESH_URL = '/oauth/token'
/** Codes that must not resolve immediately — handled specially. */
const WHITE_CODES = new Set(['00005', '00006', '00013', '00015'])
const WHITE_APIS = new Set(['/oauth/logout', '/client/order/config/get'])

let refreshing = false
let pending: Array<{
  config: RequestConfig
  resolve: (v: ApiResult) => void
}> = []

function platformTag(): string {
  // Legacy generateAjaxParams always sent platform: 'wechat'
  return 'wechat'
}

function ensureDeviceId(): string {
  const existing = storage.get<unknown>('deviceId', '')
  const id = existing === undefined || existing === null || existing === '' ? '' : String(existing)
  if (id) {
    // Migrate legacy bare/numeric storage values to quoted JSON strings
    if (typeof existing !== 'string') storage.set('deviceId', id)
    return id
  }
  const next = `${Date.now()}${(Math.random() * 1e11).toFixed(0)}`
  storage.set('deviceId', next)
  return next
}

/**
 * Align with legacy generateAjaxParams: public body fields + t/s sign + Basic/Bearer.
 */
function buildSignedPayload(
  url: string,
  data: Record<string, unknown> = {},
  auth = true,
): { data: Record<string, unknown>; header: Record<string, string> } {
  const tenant = getTenantConfig()
  const timestamp = Date.now()
  const platformConfig = {
    tenantId: String(tenant.platformTenantId || ''),
    secret: String(tenant.platformSecret || ''),
    sign: String(tenant.platformSign || ''),
  }
  // Auth service binds these as Go strings — numbers fail with 00004 Invalid request body format
  const publicParams: Record<string, unknown> = {
    platform: platformTag(),
    traceId: `${timestamp}${(Math.random() * 1e11).toFixed(0)}`,
    deviceId: ensureDeviceId(),
    tenantId: platformConfig.tenantId,
  }
  const body = { ...publicParams, ...data }
  if (!platformConfig.tenantId || !platformConfig.sign) {
    logger.warn('request missing platform keys', url, {
      tenantId: platformConfig.tenantId,
      hasSign: Boolean(platformConfig.sign),
      hasSecret: Boolean(platformConfig.secret),
      alias: tenant.alias,
    })
  }

  // Production gateway expects `_t`/`_s` and signs with `_t=` (not `t`/`s`).
  // Verified against client.luopingtech.com: t/s → 缺少签名头; t= in sign → 无效签名.
  const header: Record<string, string> = {
    'Content-Type': 'application/json',
    _t: String(timestamp),
    _s: sha256(JSON.stringify(body) + `_t=${timestamp}${platformConfig.sign}`).toString(),
    'Accept-Language': getAcceptLanguage(),
    'accept-language': getAcceptLanguage(),
  }

  if (url === REFRESH_URL) {
    const wordArr = encUtf8.parse(`${platformConfig.tenantId}:${platformConfig.secret}`)
    header.Authorization = `Basic ${encBase64.stringify(wordArr)}`
  } else if (auth !== false) {
    const login = storage.get<LoginInfo>('loginInfo', {}) || {}
    if (login.accessToken) {
      // Legacy always uses Bearer prefix (ignore tokenType casing from oauth)
      header.Authorization = `Bearer ${login.accessToken}`
    }
  }

  return { data: body, header }
}

function clearSession(keepServiceId = false) {
  storage.remove('loginInfo')
  storage.remove('userInfo')
  if (!keepServiceId) storage.remove('serviceId')
  pending = []
  refreshing = false
}

function forceRelogin(msg?: string, keepServiceId = false) {
  clearSession(keepServiceId)
  uni.showModal({
    title: t('error.unauthorized'),
    content: msg || t('error.unauthorized'),
    showCancel: false,
    success: () => {
      uni.reLaunch({ url: '/pages/auth/quick-login' })
    },
  })
}

async function refreshToken(): Promise<boolean> {
  const login = storage.get<LoginInfo>('loginInfo', {}) || {}
  if (!login.refreshToken) return false
  const res = await rawRequest<LoginInfo>({
    url: REFRESH_URL,
    method: 'POST',
    data: {
      grant_type: 'refresh_token',
      refresh_token: login.refreshToken,
    },
    auth: false,
    showError: false,
  })
  if (res.success && res.data) {
    const next: LoginInfo = { ...login, ...res.data }
    next.openid = next.openid || login.openid || ''
    delete next.pin
    storage.set('loginInfo', next)
    return true
  }
  return false
}

function rawRequest<T = unknown>(config: RequestConfig): Promise<ApiResult<T>> {
  const tenant = getTenantConfig()
  const base = getApiBaseUrl().replace(/\/$/, '')
  const url = config.url.startsWith('http') ? config.url : `${base}${config.url}`
  const timeout = tenant.networkTimeout?.request || 60000
  const body =
    config.data && typeof config.data === 'object' && !Array.isArray(config.data)
      ? (config.data as Record<string, unknown>)
      : {}
  const signed = buildSignedPayload(config.url, body, config.auth !== false)
  const header: Record<string, string> = {
    ...signed.header,
    ...(config.header || {}),
  }

  return new Promise((resolve) => {
    uni.request({
      url,
      method: config.method || 'GET',
      data: signed.data as UniApp.RequestOptions['data'],
      header,
      timeout,
      success: (res) => {
        const raw = (res.data || {}) as ApiResult<T>
        resolve({
          success: Boolean(raw.success),
          code: raw.code,
          msg: raw.msg,
          data: raw.data,
        })
      },
      fail: (err) => {
        logger.error('request fail', url, err)
        if (config.showError !== false) {
          uni.showToast({ title: t('common.networkError'), icon: 'none' })
        }
        resolve({ success: false, code: 'NETWORK', msg: t('common.networkError') })
      },
    })
  })
}

function showBizError(config: RequestConfig, res: ApiResult) {
  if (config.showError === false || !res.msg || res.isHideApiError) return
  const mapped = t(`error.${res.code}`)
  uni.showToast({
    title: mapped !== `error.${res.code}` ? mapped : res.msg,
    icon: 'none',
  })
}

export async function request<T = unknown>(config: RequestConfig): Promise<ApiResult<T>> {
  const res = await rawRequest<T>(config)

  // Refresh token itself failed while logged in → force logout
  if (!res.success && config.url === REFRESH_URL) {
    const login = storage.get<LoginInfo>('loginInfo', {}) || {}
    if (login.accessToken) {
      forceRelogin(res.msg)
      return res
    }
  }

  if (WHITE_APIS.has(config.url) || !res.code || !WHITE_CODES.has(res.code)) {
    if (!res.success) showBizError(config, res)
    return res
  }

  switch (res.code) {
    case '00006': {
      // Missing/invalid Authorization. Guest hits on auth-required APIs also return
      // 00006 — only clear session when we actually had a token to reject.
      const login = storage.get<LoginInfo>('loginInfo', {}) || {}
      if (config.auth !== false && login.accessToken) {
        const pages = getCurrentPages()
        const route = pages.length ? (pages[pages.length - 1] as { route?: string }).route : ''
        if (route !== 'pages/account/profile' && route !== 'pages/userInfo/userInfo') {
          clearSession(true)
        }
      }
      return { ...res, isHideApiError: true, success: false }
    }
    case '00015': {
      forceRelogin(res.msg)
      return res
    }
    case '00005':
    case '00013': {
      // Token expired / invalid — queue and refresh once
      return new Promise((resolve) => {
        pending.push({ config, resolve: resolve as (v: ApiResult) => void })
        if (refreshing) return
        refreshing = true
        refreshToken()
          .then(async (ok) => {
            const queue = pending
            pending = []
            refreshing = false
            if (!ok) {
              forceRelogin()
              queue.forEach((item) => item.resolve({ success: false, code: '00005' }))
              return
            }
            for (const item of queue) {
              item.resolve(await rawRequest(item.config))
            }
          })
          .catch(() => {
            refreshing = false
            pending = []
            forceRelogin()
            resolve({ success: false, code: '00005' })
          })
      })
    }
    default:
      if (!res.success) showBizError(config, res)
      return res
  }
}
