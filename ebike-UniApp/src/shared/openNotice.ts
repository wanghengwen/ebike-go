import { getLoginPath, navigate } from '@/shared/navigate'
import { useUserStore } from '@/stores/user'
import { mapLegacyRedEnvelopePath } from '@/features/bike/redEnvelope'
import { mapLegacyPagePath } from '@/shared/mapLegacyPath'
import { t } from '@/locales'

type NoticeLike = {
  skipUrl?: string
  linkUrl?: string
  appid?: string
  appId?: string
  params?: unknown
  param?: unknown
  title?: string
  linkTitle?: string
  type?: number | string
}

function ensureLogin(): boolean {
  const user = useUserStore()
  user.hydrateFromStorage()
  if (user.isLoggedIn) return true
  uni.showModal({
    title: t('auth.loginTitle'),
    content: t('auth.loginTitle'),
    showCancel: false,
    confirmText: t('auth.loginTitle'),
    success: (r) => {
      if (r.confirm) navigate('to', getLoginPath())
    },
  })
  return false
}

function openInternalOrH5(linkUrl: string, title?: string) {
  let url = linkUrl
  if (url.startsWith('miniapp')) url = url.split(':/')[1] || url
  if (/^https?:\/\//i.test(url)) {
    navigate(
      'to',
      `/pages/webview/webview?url=${encodeURIComponent(url)}${
        title ? `&title=${encodeURIComponent(title)}` : ''
      }`,
    )
    return
  }
  const mapped = mapLegacyPagePath(mapLegacyRedEnvelopePath(url))
  if (mapped.startsWith('/pages') || mapped.startsWith('/pages-sub')) {
    navigate('to', mapped)
  } else if (/^https?:\/\//i.test(mapped)) {
    navigate('to', `/pages/webview/webview?url=${encodeURIComponent(mapped)}`)
  }
}

/**
 * Legacy UserClickAction(chainType): 0 internal, 1 H5, 2 other mini program, 3 noop.
 */
export function openUserClickAction(
  chainType: number,
  jumpPage: Record<string, unknown> = {},
) {
  if (chainType === 0) {
    openJumpAction(1, jumpPage)
    return
  }
  if (chainType === 1) {
    openJumpAction(3, jumpPage)
    return
  }
  if (chainType === 2) {
    openJumpAction(2, jumpPage)
    return
  }
}

/**
 * Ops popup clickEvent: 0 close-only (caller), 1 internal, 2 other mini program, 3 H5.
 */
export function openJumpAction(clickEvent: number, jumpPage: Record<string, unknown> = {}) {
  const linkUrl = String(jumpPage.linkUrl || jumpPage.skipUrl || '')
  const appId = String(jumpPage.appId || jumpPage.appid || '')
  const title = String(jumpPage.linkTitle || jumpPage.title || '')
  if (clickEvent === 1) {
    if (!ensureLogin()) return
    if (linkUrl) openInternalOrH5(linkUrl, title)
    return
  }
  if (clickEvent === 2 && appId) {
    // #ifdef MP-WEIXIN
    uni.navigateToMiniProgram({
      appId,
      path: linkUrl || undefined,
      extraData: (jumpPage.param || jumpPage.params || {}) as Record<string, unknown>,
      fail: () => uni.showToast({ title: t('common.networkError'), icon: 'none' }),
    })
    // #endif
    // #ifndef MP-WEIXIN
    uni.showToast({ title: t('common.networkError'), icon: 'none' })
    // #endif
    return
  }
  if (clickEvent === 3 && linkUrl) {
    openInternalOrH5(linkUrl, title)
  }
}

/**
 * Legacy home viewNotice → UserClickAction(type - 1).
 * type: 1 internal, 2 H5, 3 other mini program, 4 noop.
 */
export function openScrollerNotice(item?: NoticeLike | null) {
  if (!item) return
  if (!ensureLogin()) return

  const linkUrl = String(item.skipUrl || item.linkUrl || '')
  const appId = String(item.appid || item.appId || '')
  const chainType = Number(item.type || 0) - 1
  const title = String(item.title || item.linkTitle || '')

  if (chainType === 2 && appId) {
    openJumpAction(2, { appId, linkUrl, param: item.params ?? item.param })
    return
  }
  if (!linkUrl) return
  if (chainType === 1 || /^https?:\/\//i.test(linkUrl) || linkUrl.startsWith('miniapp')) {
    openJumpAction(3, { linkUrl, linkTitle: title })
    return
  }
  openJumpAction(1, { linkUrl, linkTitle: title })
}
