import { syncFromNativeHost } from '@/shared/nativeHostSync'
import { isNative, nativeHost } from '@/shared/nativeHost'
import { useUserStore } from '@/stores/user'

/**
 * 小程序：看 storage 里的 accessToken。
 * Rider WebView：token 只在原生侧，须先 syncFromNativeHost，再认 `nativeHost` / profile.loggedIn。
 * 不要再用 `loginInfo.accessToken` 单独判断 —— native 下永远为空。
 */
export async function ensureLoggedIn(): Promise<boolean> {
  if (isNative()) {
    await syncFromNativeHost()
    const user = useUserStore()
    user.hydrateFromStorage()
    if (user.isLoggedIn) return true
    try {
      const profile = await nativeHost()?.getProfile()
      if (profile?.loggedIn) {
        user.setLoginInfo({ ...user.loginInfo, nativeHost: true })
        return true
      }
    } catch {
      /* bridge not ready */
    }
    return false
  }

  const user = useUserStore()
  user.hydrateFromStorage()
  return user.isLoggedIn
}
