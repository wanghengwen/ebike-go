import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
// #ifndef MP-WEIXIN
import enUS from './en-US'
// #endif
import { setupDayjsLocale } from '@/shared/dayjs'
import { storage } from '@/shared/storage'

export const LOCALE_KEY = 'app_locale'

// WeChat mini program is China-only — ship zh-CN only to shrink package size.
// #ifdef MP-WEIXIN
export const SUPPORTED_LOCALES = ['zh-CN'] as const
// #endif
// #ifndef MP-WEIXIN
export const SUPPORTED_LOCALES = ['zh-CN', 'en-US'] as const
// #endif

export type AppLocale = (typeof SUPPORTED_LOCALES)[number]

function mapSystemLanguage(raw?: string): AppLocale {
  // #ifdef MP-WEIXIN
  return 'zh-CN'
  // #endif
  // #ifndef MP-WEIXIN
  const lang = (raw || '').toLowerCase()
  if (lang.startsWith('en') && SUPPORTED_LOCALES.includes('en-US')) return 'en-US'
  return 'zh-CN'
  // #endif
}

export function resolveInitialLocale(): AppLocale {
  const saved = storage.get<string>(LOCALE_KEY)
  if (saved && (SUPPORTED_LOCALES as readonly string[]).includes(saved)) {
    return saved as AppLocale
  }
  try {
    const sys = uni.getSystemInfoSync()
    return mapSystemLanguage(sys.language || (sys as { appLanguage?: string }).appLanguage)
  } catch {
    return 'zh-CN'
  }
}

const messages: Record<string, typeof zhCN> = {
  'zh-CN': zhCN,
}
// #ifndef MP-WEIXIN
messages['en-US'] = enUS
// #endif

/** Fill `{name}` placeholders when vue-i18n runtime leaves them untouched (mp builds). */
function fillNamedPlaceholders(text: string, params?: Record<string, unknown>): string {
  if (!params || typeof text !== 'string' || !text.includes('{')) return text
  return text.replace(/\{(\w+)\}/g, (all, name: string) => {
    if (!Object.prototype.hasOwnProperty.call(params, name)) return all
    const v = params[name]
    return v === undefined || v === null ? '' : String(v)
  })
}

function namedParams(arg: unknown): Record<string, unknown> | undefined {
  if (!arg || typeof arg !== 'object' || Array.isArray(arg)) return undefined
  return arg as Record<string, unknown>
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages,
})

// Patch composer.t so both useI18n().t and locales.t interpolate reliably on MP.
const composerT = i18n.global.t.bind(i18n.global)
;(i18n.global as { t: typeof i18n.global.t }).t = ((key: string, ...args: unknown[]) => {
  const out = composerT(key as never, ...(args as never[]))
  return fillNamedPlaceholders(String(out), namedParams(args[0])) as never
}) as typeof i18n.global.t

export function initLocale() {
  // #ifdef MP-WEIXIN
  setLocale('zh-CN')
  // #endif
  // #ifndef MP-WEIXIN
  setLocale(resolveInitialLocale())
  // #endif
}

export function setLocale(locale: AppLocale) {
  const next = (SUPPORTED_LOCALES as readonly string[]).includes(locale) ? locale : 'zh-CN'
  i18n.global.locale.value = next as AppLocale
  storage.set(LOCALE_KEY, next)
  setupDayjsLocale(next)
  // #ifdef H5
  if (typeof document !== 'undefined') {
    document.documentElement.lang = next
  }
  // #endif
}

export function getAcceptLanguage(): string {
  return i18n.global.locale.value || 'zh-CN'
}

export function t(key: string, params?: Record<string, unknown>) {
  return fillNamedPlaceholders(String(i18n.global.t(key, params as never)), params)
}
