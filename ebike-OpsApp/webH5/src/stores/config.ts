import { defineStore } from 'pinia'
import { computed, reactive, readonly } from 'vue'
import { storage, StorageKey } from '@/utils/storage'
import type { TokenPayload } from '@/api/types'

/**
 * App 通过 URL query 注入的运行时配置，字段名与 `shared/.../core/config/H5ScreenUrls.kt`
 * 的 `resolve()` 一一对应。改这里必须同步改 Kotlin 侧。
 */
export interface RuntimeConfig {
  apiHost: string
  accessToken: string
  refreshToken: string
  tenantId: string
  tenantSecret: string
  sign: string
  platform: string
  deviceId: string
  opMan: string
}

const EMPTY_CONFIG: RuntimeConfig = {
  apiHost: '',
  accessToken: '',
  refreshToken: '',
  tenantId: '',
  tenantSecret: '',
  sign: '',
  platform: '',
  deviceId: '',
  opMan: '',
}

/** 与 vue-router 的 `LocationQuery` 兼容：数组元素也可能是 null。 */
type QueryValue = string | null | undefined | Array<string | null>
type Query = Record<string, QueryValue>

function first(value: QueryValue): string {
  if (Array.isArray(value)) return value[0] ?? ''
  return value ?? ''
}

/** 仅在开发模式下生效，避免每次调试都要手工拼一长串 query。 */
function devFallback(): Partial<RuntimeConfig> {
  if (!import.meta.env.DEV) return {}
  return {
    apiHost: import.meta.env.VITE_DEV_API_HOST ?? '',
    tenantId: import.meta.env.VITE_DEV_TENANT_ID ?? '',
    sign: import.meta.env.VITE_DEV_SIGN ?? '',
    tenantSecret: import.meta.env.VITE_DEV_TENANT_SECRET ?? '',
    accessToken: import.meta.env.VITE_DEV_ACCESS_TOKEN ?? '',
    refreshToken: import.meta.env.VITE_DEV_REFRESH_TOKEN ?? '',
  }
}

export const useConfigStore = defineStore('config', () => {
  const state = reactive<RuntimeConfig>({
    ...EMPTY_CONFIG,
    ...devFallback(),
    ...(storage.get<Partial<RuntimeConfig>>(StorageKey.Config) ?? {}),
  })

  const baseUrl = computed(() => state.apiHost.replace(/\/+$/, ''))
  const isReady = computed(() => Boolean(state.apiHost && state.tenantId))

  function persist(): void {
    storage.set(StorageKey.Config, { ...state })
  }

  /**
   * 从路由 query 装载配置。沿用遗留判定：四个关键字段齐了才认为是一次新的注入，
   * 否则保留已缓存的会话（App 内二次进入大屏时 query 可能已被清掉）。
   */
  function hydrateFromQuery(query: Query): boolean {
    const incoming = {
      apiHost: first(query.apiHost),
      accessToken: first(query.accessToken),
      tenantId: first(query.tenantId),
      sign: first(query.sign),
    }
    if (!incoming.apiHost || !incoming.accessToken || !incoming.tenantId || !incoming.sign) {
      return false
    }

    Object.assign(state, {
      ...incoming,
      refreshToken: first(query.refreshToken),
      tenantSecret: first(query.tenantSecret),
      platform: first(query.platform),
      deviceId: first(query.deviceId),
      opMan: first(query.opMan),
    } satisfies RuntimeConfig)
    persist()
    return true
  }

  function setTokens(payload: TokenPayload): void {
    state.accessToken = payload.accessToken
    state.refreshToken = payload.refreshToken
    persist()
  }

  function clearTokens(): void {
    state.accessToken = ''
    state.refreshToken = ''
    persist()
  }

  return { config: readonly(state), baseUrl, isReady, hydrateFromQuery, setTokens, clearTokens }
})
