import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'amfe-flexible'
import 'vant/lib/index.css'
import App from './App.vue'
import router from './router'
import { i18n } from './i18n'
import { applyTheme } from './utils/theme'
import './styles/global.css'

applyTheme()

createApp(App).use(createPinia()).use(i18n).use(router).mount('#app')
