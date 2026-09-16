import { SUPPORTED_LOCALES, setLocale, type AppLocale } from '@/locales'
import { ensureServiceId } from '@/shared/ensureServiceId'
import { isNative, nativeHost } from '@/shared/nativeHost'
import { storage } from '@/shared/storage'
import { useUserStore } from '@/stores/user'

function mapNativeLang(tag: string): AppLocale {
  const lower = (tag || '').toLowerCase()
  if (lower.startsWith('en') && (SUPPORTED_LOCALES as readonly string[]).includes('en-US')) {
    return 'en-US'
  }
  return 'zh-CN'
}

/**
 * 设置页改语言时：先 `setLocale`，再 `nativeHost.setLanguage`，RiderI18n / SecureStore 同步。
 * 下次打开 H5，这里用 `getLanguage()` 回写 vue-i18n。
 * 同时把原生已定位的服务区写入 `storage.serviceId`（对齐小程序首页逻辑）。
 */
export async function syncFromNativeHost(): Promise<void> {
  if (!isNative()) return
  const host = nativeHost()
  if (!host) return
  try {
    const lang = await host.getLanguage()
    if (lang) {
      const mapped = mapNativeLang(lang)
      if ((SUPPORTED_LOCALES as readonly string[]).includes(mapped)) {
        setLocale(mapped)
      }
    }
    const profile = await host.getProfile()
    const user = useUserStore()
    if (profile.loggedIn) {
      user.setUserInfo({
        userId: profile.userId,
        pin: profile.pin,
        phone: profile.phone,
        nickName: profile.displayName,
        avatar: profile.avatar,
      })
      user.setLoginInfo({ nativeHost: true })
    } else {
      user.logout(true)
    }
    const sid = String(profile.serviceAreaId || '').trim()
    if (sid) storage.set('serviceId', sid)
  } catch {
    /* bridge not injected yet */
  }
  // 无缓存时：定位 → getByLocation（与小程序启动一致）
  await ensureServiceId()
}
