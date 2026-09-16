/**
 * RiderApp WebView 桥。微信小程序 / 普通浏览器里 `isNative()` 为 false，走原 uni 路径。
 *
 * Android：`window.__riderNative.invoke(json)`
 * iOS：`webkit.messageHandlers.riderNative.postMessage(json)`
 *
 * 不要用 #ifdef 拆掉小程序路径；一律运行时判断。
 */

export type NativeRequestConfig = {
  url: string
  method?: string
  data?: Record<string, unknown>
  header?: Record<string, string>
  auth?: boolean
}

export type NativePayPayload = {
  sale_type?: string
  sale_info?: Record<string, unknown>
  order?: Record<string, unknown>
  [key: string]: unknown
}

export type NativeProfile = {
  loggedIn?: boolean
  userId?: string
  pin?: string
  displayName?: string
  phone?: string
  avatar?: string
  balance?: number
  tenantId?: string
  serviceAreaId?: string
}

export type NativeHost = {
  request(config: NativeRequestConfig): Promise<Record<string, unknown>>
  getProfile(): Promise<NativeProfile>
  pay(payload?: NativePayPayload): Promise<Record<string, unknown>>
  scanCode(): Promise<{ code?: string }>
  capturePhoto(): Promise<{ uri?: string }>
  currentLocation(): Promise<{ latitude?: number; longitude?: number }>
  openNavigation(lat: number, lng: number, name?: string): Promise<void>
  navigate(opts: { type: string; url?: string; delta?: number }): Promise<void>
  close(): Promise<void>
  setTitle(title: string): Promise<void>
  toast(message: string): Promise<void>
  getLanguage(): Promise<string>
  setLanguage(language: string): Promise<string>
  platform(): string
}

type Pending = {
  resolve: (value: unknown) => void
  reject: (err: Error) => void
  timer: ReturnType<typeof setTimeout>
}

type RiderNativeWindow = {
  __riderNative?: {
    invoke?: (payload: string) => void
    platform?: () => string
  }
  __riderNativePlatform?: string
  __riderNativeQueue?: Array<{ id: string; result: unknown }>
  __riderNativeOnResult?: (id: string, result: unknown) => void
  webkit?: { messageHandlers?: { riderNative?: { postMessage: (body: unknown) => void } } }
}

let seq = 0
const pending = new Map<string, Pending>()
let installed = false

function getWin(): RiderNativeWindow | undefined {
  const g = globalThis as { window?: RiderNativeWindow }
  return g.window
}

export function isNative(): boolean {
  const w = getWin()
  if (!w) return false
  if (w.__riderNative || w.webkit?.messageHandlers?.riderNative) return true
  // 原生注入可能略早于 __riderNative 接口挂载
  if (w.__riderNativePlatform === 'android' || w.__riderNativePlatform === 'ios') return true
  if (typeof document !== 'undefined' && document.documentElement.classList.contains('rider-native-host')) {
    return true
  }
  return false
}

function installCallback() {
  const w = getWin()
  if (!w || installed) return
  installed = true
  const queued = w.__riderNativeQueue || []
  w.__riderNativeOnResult = (id: string, result: unknown) => {
    const item = pending.get(String(id))
    if (!item) return
    pending.delete(String(id))
    clearTimeout(item.timer)
    if (result && typeof result === 'object' && (result as { __error?: boolean }).__error) {
      const msg = String((result as { msg?: string }).msg || 'nativeHost error')
      item.reject(new Error(msg))
      return
    }
    item.resolve(result)
  }
  queued.forEach((row) => {
    w.__riderNativeOnResult?.(row.id, row.result)
  })
  w.__riderNativeQueue = []
}

function timeoutMs(method: string): number {
  if (method === 'scanCode' || method === 'capturePhoto') return 120000
  if (method === 'request') return 60000
  return 20000
}

function invoke(method: string, args: Record<string, unknown> = {}): Promise<unknown> {
  installCallback()
  const w = getWin()
  const id = String(++seq)
  const payload = { id, method, args }
  return new Promise((resolve, reject) => {
    if (!w || !isNative()) {
      reject(new Error('nativeHost unavailable'))
      return
    }
    const timer = setTimeout(() => {
      pending.delete(id)
      reject(new Error(`nativeHost timeout: ${method}`))
    }, timeoutMs(method))
    pending.set(id, { resolve, reject, timer })
    try {
      if (w.__riderNative?.invoke) {
        w.__riderNative.invoke(JSON.stringify(payload))
      } else if (w.webkit?.messageHandlers?.riderNative) {
        w.webkit.messageHandlers.riderNative.postMessage(JSON.stringify(payload))
      } else {
        pending.delete(id)
        clearTimeout(timer)
        reject(new Error('nativeHost unavailable'))
      }
    } catch (e) {
      pending.delete(id)
      clearTimeout(timer)
      reject(e instanceof Error ? e : new Error(String(e)))
    }
  })
}

const hostImpl: NativeHost = {
  request(config) {
    return invoke('request', {
      url: config.url,
      method: config.method || 'POST',
      data: config.data || {},
      header: config.header,
      auth: config.auth,
    }) as Promise<Record<string, unknown>>
  },
  getProfile() {
    return invoke('getProfile') as Promise<NativeProfile>
  },
  pay(payload) {
    return invoke('pay', payload || {}) as Promise<Record<string, unknown>>
  },
  scanCode() {
    return invoke('scanCode') as Promise<{ code?: string }>
  },
  capturePhoto() {
    return invoke('capturePhoto') as Promise<{ uri?: string }>
  },
  currentLocation() {
    return invoke('currentLocation') as Promise<{ latitude?: number; longitude?: number }>
  },
  async openNavigation(lat, lng, name) {
    await invoke('openNavigation', { latitude: lat, longitude: lng, name: name || '' })
  },
  async navigate(opts) {
    await invoke('navigate', opts)
  },
  async close() {
    await invoke('close')
  },
  async setTitle(title) {
    await invoke('setTitle', { title })
  },
  async toast(message) {
    await invoke('toast', { title: message })
  },
  async getLanguage() {
    const raw = (await invoke('getLanguage')) as { language?: string }
    return String(raw?.language || '')
  },
  async setLanguage(language) {
    const raw = (await invoke('setLanguage', { language })) as { language?: string }
    return String(raw?.language || language)
  },
  platform() {
    const w = getWin()
    try {
      if (typeof w?.__riderNative?.platform === 'function') {
        return String(w.__riderNative.platform() || 'android')
      }
    } catch {
      /* ignore */
    }
    return String(w?.__riderNativePlatform || 'android')
  },
}

export function nativeHost(): NativeHost | null {
  return isNative() ? hostImpl : null
}

export function sanitizeLoginInfo(value: unknown): Record<string, unknown> {
  const src = value && typeof value === 'object' && !Array.isArray(value)
    ? { ...(value as Record<string, unknown>) }
    : {}
  delete src.accessToken
  delete src.refreshToken
  delete src.token
  delete src.tokenType
  delete src.Authorization
  return src
}
