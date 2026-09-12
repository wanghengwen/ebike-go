import { createI18n } from 'vue-i18n'
import { storage, StorageKey } from '@/utils/storage'
import zhCN from './locales/zh-CN'
import en from './locales/en'

export const SUPPORTED_LOCALES = ['zh-CN', 'en'] as const
export type Locale = (typeof SUPPORTED_LOCALES)[number]

const FALLBACK: Locale = 'zh-CN'

function isSupported(value: string | null | undefined): value is Locale {
  return SUPPORTED_LOCALES.includes(value as Locale)
}

/**
 * 优先级：URL query `lang` > 上次选择 > 浏览器语言 > 中文。
 * `lang` 是新增的可选参数，Kotlin 侧暂未传，缺省时行为与遗留一致（中文）。
 */
export function resolveInitialLocale(): Locale {
  const fromQuery = new URLSearchParams(window.location.search).get('lang')
  const fromHash = new URLSearchParams(window.location.hash.split('?')[1] ?? '').get('lang')
  const candidates = [fromHash, fromQuery, storage.get<string>(StorageKey.Language)]
  for (const candidate of candidates) {
    if (isSupported(candidate)) return candidate
  }
  return navigator.language.toLowerCase().startsWith('zh') ? 'zh-CN' : FALLBACK
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: resolveInitialLocale(),
  fallbackLocale: FALLBACK,
  messages: { 'zh-CN': zhCN, en },
})

export function currentLocale(): Locale {
  return i18n.global.locale.value as Locale
}

export function setLocale(locale: Locale): void {
  i18n.global.locale.value = locale
  storage.set(StorageKey.Language, locale)
  document.documentElement.lang = locale
}

/** 组件外（拦截器、工具函数）取文案用。 */
export const t = i18n.global.t
