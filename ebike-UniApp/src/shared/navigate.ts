import { isNative, nativeHost } from '@/shared/nativeHost'

function isNativeLoginUrl(url?: string): boolean {
  if (!url) return false
  return (
    url === 'login' ||
    url.startsWith('native://login') ||
    url.includes('/pages/auth/phone-login') ||
    url.includes('/pages/auth/quick-login')
  )
}

export function navigate(
  type: 'to' | 'redirect' | 'reLaunch' | 'back' | 'switchTab',
  url?: string,
  delta = 1,
) {
  if (isNative() && isNativeLoginUrl(url)) {
    nativeHost()?.navigate({ type, url: 'login' })
    return
  }
  if (type === 'back') {
    if (isNative()) {
      // Rider 从首页打开的长尾页是 WebView 栈底：再 back 应关闭容器回原生首页，
      // 不能 uni.reLaunch 到 H5 的 home/profile（会留在 WebView 里）。
      const pages = getCurrentPages()
      if (pages.length <= 1) {
        void nativeHost()?.close()
        return
      }
      uni.navigateBack({
        delta,
        animationType: 'none',
        animationDuration: 0,
        fail: () => {
          void nativeHost()?.close()
        },
      })
      return
    }
    uni.navigateBack({
      delta,
      fail: () => {
        uni.reLaunch({ url: '/pages/home/home' })
      },
    })
    return
  }
  if (!url) return
  const map = {
    to: uni.navigateTo,
    redirect: uni.redirectTo,
    reLaunch: uni.reLaunch,
    switchTab: uni.switchTab,
  } as const
  if (isNative() && (type === 'to' || type === 'redirect')) {
    // H5 下 animation* 可能被忽略，仍传入；App.vue CSS 会关掉转场
    ;(map[type] as (opts: Record<string, unknown>) => void)({
      url,
      animationType: 'none',
      animationDuration: 0,
    })
    return
  }
  map[type]({ url })
}

/** WeChat mini program → one-tap login; App / H5 → SMS phone login. */
export function getLoginPath(): string {
  if (isNative()) return 'native://login'
  // #ifdef MP-WEIXIN
  return '/pages/auth/quick-login'
  // #endif
  // #ifndef MP-WEIXIN
  return '/pages/auth/phone-login'
  // #endif
}

export function setNavTitle(title: string) {
  if (isNative()) {
    nativeHost()?.setTitle(title)
  }
  uni.setNavigationBarTitle({ title })
}
