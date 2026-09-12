import axios, {
  type AxiosError,
  type AxiosInstance,
  type AxiosRequestConfig,
  type InternalAxiosRequestConfig,
} from 'axios'
import CryptoJS from 'crypto-js'
import { stringify } from 'qs'
import { showToast } from 'vant'
import { useConfigStore } from '@/stores/config'
import { currentLocale } from '@/i18n'
import { ApiCode, SILENT_CODES, type ApiResponse, type TokenPayload } from './types'

export const REFRESH_PATH = '/oauth/token'

interface RetryableConfig extends InternalAxiosRequestConfig {
  /** 一个请求最多因 token 过期重放一次，避免刷新链路自身出错时打转。 */
  __retried?: boolean
}

const http: AxiosInstance = axios.create({ timeout: 60_000 })

function traceId(timestamp: number): string {
  return `${timestamp}${Math.floor(Math.random() * 1e11)}`
}

/** 网关要求 GET 的签名按参数名字典序拼接。 */
function sortKeys(params: Record<string, unknown>): Record<string, unknown> {
  return Object.fromEntries(Object.keys(params).sort().map((k) => [k, params[k]]))
}

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const { config: runtime } = useConfigStore()
  const timestamp = Date.now()
  const method = (config.method ?? 'post').toLowerCase()

  const globals: Record<string, string> = {
    platform: runtime.platform,
    traceId: traceId(timestamp),
    tenantId: runtime.tenantId,
    deviceId: runtime.deviceId,
  }

  const isForm = config.data instanceof FormData
  if (isForm) {
    for (const [key, value] of Object.entries(globals)) {
      ;(config.data as FormData).append(key, value)
    }
  } else if (method === 'get') {
    config.params = sortKeys({ ...(config.params ?? {}), ...globals })
  } else {
    // token 过期重放时 axios 已经把 data 序列化成字符串了，直接展开会拆成字符数组。
    const body = typeof config.data === 'string' ? JSON.parse(config.data) : config.data
    config.data = { ...(body ?? {}), ...globals }
  }

  const isRefresh = config.url === REFRESH_PATH
  config.headers.set(
    'Authorization',
    isRefresh
      ? `Basic ${btoa(`${runtime.tenantId}:${runtime.tenantSecret}`)}`
      : runtime.accessToken
        ? `Bearer ${runtime.accessToken}`
        : '',
  )
  config.headers.set('Content-Type', isForm ? 'multipart/form-data' : 'application/json')
  config.headers.set('accept-language', currentLocale())
  config.headers.set('_t', String(timestamp))

  // 签名口径与遗留实现逐字符一致：GET 用 `&_t=`，POST 用 `_t=`。
  const payload =
    method === 'get'
      ? `${stringify(config.params, { encode: false })}&_t=${timestamp}${runtime.sign}`
      : `${JSON.stringify(config.data)}_t=${timestamp}${runtime.sign}`
  config.headers.set('_s', CryptoJS.SHA256(payload).toString())

  return config
})

// --- token 单飞刷新 + 重放队列 ---------------------------------------------
// 遗留实现把请求推进 requestArr 后从不消费（重放代码被注释掉），靠 401 之后
// `window.location.href = path` 整页刷新兜底。这里补成真正的队列。

let refreshing: Promise<boolean> | null = null

async function doRefresh(): Promise<boolean> {
  const store = useConfigStore()
  if (!store.config.refreshToken) return false

  try {
    const { data } = await http.post<ApiResponse<TokenPayload>>(
      REFRESH_PATH,
      { grant_type: 'refresh_token', refresh_token: store.config.refreshToken },
      { baseURL: store.baseUrl },
    )
    if (!data?.success || !data.data?.accessToken) return false
    store.setTokens(data.data)
    return true
  } catch {
    return false
  }
}

function refreshOnce(): Promise<boolean> {
  refreshing ??= doRefresh().finally(() => {
    refreshing = null
  })
  return refreshing
}

function onSessionLost(): void {
  useConfigStore().clearTokens()
  showToast({ message: '身份信息已过期, 请重新进入', duration: 2000 })
}

http.interceptors.response.use(async (response) => {
  const config = response.config as RetryableConfig
  const body = response.data as ApiResponse | undefined
  const code = body?.code

  if (!code || !SILENT_CODES.includes(code)) {
    if (code === ApiCode.Failed && config.url === REFRESH_PATH) onSessionLost()
    return response
  }

  if (config.url === REFRESH_PATH) {
    onSessionLost()
    return response
  }

  if (code === ApiCode.Kicked || code === ApiCode.MissingAuth) {
    onSessionLost()
    return response
  }

  // 00005 / 00013：token 过期，刷新后原样重放一次。
  if (config.__retried) {
    onSessionLost()
    return response
  }
  config.__retried = true

  if (!(await refreshOnce())) {
    onSessionLost()
    return response
  }
  return http.request(config)
})

async function request<T>(config: AxiosRequestConfig): Promise<ApiResponse<T>> {
  const { baseUrl } = useConfigStore()
  const response = await http.request<ApiResponse<T>>({ ...config, baseURL: baseUrl })
  return response.data
}

export function post<T>(url: string, data?: unknown): Promise<ApiResponse<T>> {
  return request<T>({ url, method: 'post', data })
}

export function get<T>(url: string, params?: Record<string, unknown>): Promise<ApiResponse<T>> {
  return request<T>({ url, method: 'get', params })
}

export function isNetworkError(error: unknown): error is AxiosError {
  return axios.isAxiosError(error)
}

export default http
