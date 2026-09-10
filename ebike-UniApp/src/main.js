import { createSSRApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { i18n } from './locales'
import { setupDayjsLocale } from './shared/dayjs'
import '@/styles/app.scss'
import '@/shared/mapLegacyPath'

export function createApp() {
  const app = createSSRApp(App)
  const pinia = createPinia()
  app.use(pinia)
  app.use(i18n)
  setupDayjsLocale(i18n.global.locale.value)
  return { app }
}
