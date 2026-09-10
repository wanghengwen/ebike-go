import { authState } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { navigate } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

export type VerifyGateResult =
  | { ok: true }
  | { ok: false; reason: 'reviewing' | 'need_auth' }

/**
 * Legacy checkVerifyStatusAndGo:
 * authState 0/2 → verified, 1 → bind-card-state, 3 / izAuth → ok.
 * Returns whether caller may continue (verified).
 */
export async function ensureVerifiedOrNavigate(
  opts: { free?: boolean; directReturn?: boolean } = {},
): Promise<VerifyGateResult> {
  const user = useUserStore()
  user.hydrateFromStorage()
  const freeQs = opts.free ? '&free=1' : ''
  const directQs = `izDirectReturn=${opts.directReturn ? 'true' : 'false'}`
  try {
    const res = await authState()
    if (!res.success || !res.data) {
      navigate('to', `/pages/auth/verified?${directQs}${freeQs}`)
      return { ok: false, reason: 'need_auth' }
    }
    const data = res.data as {
      izAuth?: boolean
      authState?: number
      authName?: string
      authNo?: string
    }
    if (data.izAuth || Number(data.authState) === 3) {
      if (data.izAuth) user.setUserInfo({ ...(user.userInfo as object), izAuth: true } as never)
      return { ok: true }
    }
    const state = Number(data.authState)
    if (state === 1) {
      const info = (user.userInfo || storage.get<Record<string, unknown>>('userInfo', {}) || {}) as Record<
        string,
        unknown
      >
      const authName = encodeURIComponent(String(data.authName || info.authName || ''))
      const authNo = encodeURIComponent(String(data.authNo || info.authNo || ''))
      navigate(
        'to',
        `/pages-sub/account/bind-card/bind-card-state?authName=${authName}&authNo=${authNo}`,
      )
      return { ok: false, reason: 'reviewing' }
    }
    navigate('to', `/pages/auth/verified?${directQs}${freeQs}`)
    return { ok: false, reason: 'need_auth' }
  } catch (e) {
    logger.warn('ensureVerifiedOrNavigate soft fail', e)
    navigate('to', `/pages/auth/verified?${directQs}${freeQs}`)
    return { ok: false, reason: 'need_auth' }
  }
}

/** Profile / menu entry — fire-and-forget navigate. */
export async function checkVerifyAndGo(opts: { free?: boolean; directReturn?: boolean } = {}) {
  await ensureVerifiedOrNavigate(opts)
}
