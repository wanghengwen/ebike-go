import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
// #ifndef MP-WEIXIN
import 'dayjs/locale/en'
// #endif

export function setupDayjsLocale(locale: string) {
  // #ifdef MP-WEIXIN
  dayjs.locale('zh-cn')
  // #endif
  // #ifndef MP-WEIXIN
  dayjs.locale(locale.toLowerCase().startsWith('zh') ? 'zh-cn' : 'en')
  // #endif
}

export { dayjs }
