/**
 * localStorage 封装。遗留实现直接 `JSON.parse(localStorage.getItem(key))`，
 * 值被外部写脏时会整页抛错，这里统一吞掉并回落到 null。
 */
export const storage = {
  get<T>(key: string): T | null {
    try {
      const raw = localStorage.getItem(key)
      return raw === null ? null : (JSON.parse(raw) as T)
    } catch {
      return null
    }
  },

  set(key: string, value: unknown): void {
    try {
      localStorage.setItem(key, JSON.stringify(value))
    } catch {
      /* 隐私模式或配额耗尽时忽略 */
    }
  },

  remove(key: string): void {
    try {
      localStorage.removeItem(key)
    } catch {
      /* ignore */
    }
  },
}

export const StorageKey = {
  Config: 'config',
  ThemeColor: 'themeColor',
  LightTxtColor: 'lightTxtColor',
  Language: 'language',
} as const
