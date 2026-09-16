import { isNative, sanitizeLoginInfo } from '@/shared/nativeHost'

const PREFIX = 'ebike_'

/**
 * Persist with JSON.stringify for every value (including strings).
 *
 * Previously strings were stored raw; `JSON.parse('1789…')` then turned
 * deviceId into a number and /oauth/token failed with 00004
 * (Go binds deviceId as string → "Invalid request body format").
 */
export const storage = {
  get<T = unknown>(key: string, fallback: T | null = null): T | null {
    try {
      const raw = uni.getStorageSync(PREFIX + key)
      if (raw === '' || raw === undefined || raw === null) return fallback
      if (typeof raw !== 'string') return raw as T
      const trimmed = raw.trim()
      // Objects, arrays, and JSON-encoded strings (new writes)
      if (
        trimmed.startsWith('{') ||
        trimmed.startsWith('[') ||
        trimmed.startsWith('"')
      ) {
        try {
          const parsed = JSON.parse(trimmed) as T
          if (isNative() && key === 'loginInfo') {
            return sanitizeLoginInfo(parsed) as T
          }
          return parsed
        } catch {
          return raw as unknown as T
        }
      }
      if (trimmed === 'true' || trimmed === 'false' || trimmed === 'null') {
        try {
          return JSON.parse(trimmed) as T
        } catch {
          return raw as unknown as T
        }
      }
      // Bare digits / other legacy raw strings (deviceId) — keep as string
      return raw as unknown as T
    } catch {
      return fallback
    }
  },
  set(key: string, value: unknown) {
    let next = value
    if (isNative() && key === 'loginInfo') {
      const src =
        value && typeof value === 'object' && !Array.isArray(value)
          ? (value as Record<string, unknown>)
          : {}
      const sanitized = sanitizeLoginInfo(value)
      // 显式 nativeHost=false 表示未登录；省略时默认 true（宿主写入登录态）
      const flagged = Object.prototype.hasOwnProperty.call(src, 'nativeHost')
        ? Boolean(src.nativeHost)
        : true
      next = { ...sanitized, nativeHost: flagged }
    }
    uni.setStorageSync(PREFIX + key, JSON.stringify(next))
  },
  remove(key: string) {
    uni.removeStorageSync(PREFIX + key)
  },
}
