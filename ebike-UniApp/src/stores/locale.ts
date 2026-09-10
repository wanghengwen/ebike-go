import { defineStore } from 'pinia'
import { setLocale, type AppLocale, i18n, SUPPORTED_LOCALES } from '@/locales'

export const useLocaleStore = defineStore('locale', {
  state: () => ({
    locale: i18n.global.locale.value as AppLocale,
  }),
  getters: {
    options: () =>
      SUPPORTED_LOCALES.map((code) => ({
        code,
        labelKey: code === 'zh-CN' ? 'common.zhCN' : 'common.enUS',
      })),
  },
  actions: {
    change(locale: AppLocale) {
      setLocale(locale)
      this.locale = locale
    },
  },
})
