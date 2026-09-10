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
          return JSON.parse(trimmed) as T
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
    uni.setStorageSync(PREFIX + key, JSON.stringify(value))
  },
  remove(key: string) {
    uni.removeStorageSync(PREFIX + key)
  },
}
