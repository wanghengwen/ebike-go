import { getPersonInfo } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { useTempDataStore } from '@/stores/tempData'
import { useFaceCheck } from '@/features/auth/useFaceCheck'
import { navigate } from '@/shared/navigate'
import { logger } from '@/shared/logger'
import { storage } from '@/shared/storage'
import { readLimitRiding, readCreditLimitInfo, creditTipTitle } from '@/features/credit/useCreditLimit'
import { showBeforeUseCarPopup } from '@/features/guide/useGuidePopup'
import { ensureVerifiedOrNavigate } from '@/features/auth/checkVerifyAndGo'
import { checkQrDomain, extractCarId } from '@/features/bike/parseQrCode'
import { ensureLocationAuthorized } from '@/features/map/ensureLocationAuth'
import { t } from '@/locales'

export type ScanPurpose = 'use' | 'fill'

export type ScanGateResult =
  | { ok: true; carId?: string }
  | {
      ok: false
      reason:
        | 'login'
        | 'riding'
        | 'unpaid'
        | 'verified'
        | 'reviewing'
        | 'qualification'
        | 'face'
        | 'blocked'
        | 'cancel'
    }

const AUTO_SCAN_KEY = 'auto_scan_use_bike'
const FILL_CAR_ID_KEY = 'pendingScanCarId'

export function markAutoScanUseBike() {
  storage.set(AUTO_SCAN_KEY, true)
}

export function consumeAutoScanUseBike(): boolean {
  const v = storage.get<boolean | string>(AUTO_SCAN_KEY, false)
  if (v === true || v === 'true') {
    storage.remove(AUTO_SCAN_KEY)
    return true
  }
  return false
}

export function consumePendingScanCarId(): string {
  const id = String(storage.get<string>(FILL_CAR_ID_KEY, '') || '')
  if (id) storage.remove(FILL_CAR_ID_KEY)
  return id
}

function goEnterId(replace = false) {
  navigate(replace ? 'redirect' : 'to', '/pages-sub/ride/enter-id/enter-id')
}

function showUnrecognizedThenEnter(replace = false) {
  uni.showModal({
    title: t('ride.qrUnrecognized'),
    content: t('ride.qrEnterHint'),
    showCancel: false,
    success: () => goEnterId(replace),
  })
}

/** Shared success path for native / custom scan pages. */
export async function applyScanResult(
  raw: string,
  opts: {
    purpose?: ScanPurpose
    failToEnterId?: boolean
    replaceScanPage?: boolean
  } = {},
): Promise<ScanGateResult> {
  const purpose = opts.purpose || 'use'
  const failToEnterId = opts.failToEnterId !== false
  const replace = Boolean(opts.replaceScanPage)
  const text = String(raw || '').trim()

  if (!text) {
    if (purpose === 'fill') {
      if (replace) navigate('back')
      return { ok: false, reason: 'cancel' }
    }
    if (failToEnterId) goEnterId(replace)
    return { ok: false, reason: 'cancel' }
  }

  if (text.includes('http') && !checkQrDomain(text)) {
    if (purpose === 'use' && failToEnterId) {
      showUnrecognizedThenEnter(replace)
    } else {
      uni.showToast({ title: t('ride.qrUnrecognized'), icon: 'none' })
      if (replace) navigate('back')
    }
    return { ok: false, reason: 'cancel' }
  }

  const carId = extractCarId(text)
  if (!carId) {
    if (purpose === 'use' && failToEnterId) {
      showUnrecognizedThenEnter(replace)
    } else {
      uni.showToast({ title: t('ride.qrUnrecognized'), icon: 'none' })
      if (replace) navigate('back')
    }
    return { ok: false, reason: 'cancel' }
  }

  if (purpose === 'fill') {
    storage.set(FILL_CAR_ID_KEY, carId)
    navigate('back')
    return { ok: true, carId }
  }

  storage.remove('scanCarId')
  navigate(
    replace ? 'redirect' : 'to',
    `/pages-sub/ride/precycling/precycling?carId=${encodeURIComponent(carId)}`,
  )
  return { ok: true, carId }
}

function isH5Runtime(): boolean {
  try {
    const info = uni.getSystemInfoSync() as { uniPlatform?: string; platform?: string }
    return info.uniPlatform === 'web' || info.platform === 'web'
  } catch {
    return typeof window !== 'undefined' && typeof document !== 'undefined'
  }
}

function openPlatformScanner(purpose: ScanPurpose): Promise<ScanGateResult> {
  // H5: custom camera page (html5-qrcode + torch). MP/App: native uni.scanCode.
  if (isH5Runtime()) {
    navigate('to', `/pages-sub/ride/scan/scan?purpose=${purpose}`)
    return Promise.resolve({ ok: true })
  }

  return new Promise((resolve) => {
    uni.scanCode({
      onlyFromCamera: false,
      success: async (res) => {
        resolve(
          await applyScanResult(res.result || '', {
            purpose,
            failToEnterId: purpose === 'use',
            replaceScanPage: false,
          }),
        )
      },
      fail: () => {
        if (purpose === 'use') goEnterId(false)
        resolve({ ok: false, reason: 'cancel' })
      },
    })
  })
}

export function useScanGate() {
  const user = useUserStore()
  const temp = useTempDataStore()
  const face = useFaceCheck()

  async function refreshProfile() {
    try {
      const res = await getPersonInfo()
      if (res.success && res.data) {
        user.setUserInfo(res.data as never)
        return res.data as Record<string, unknown>
      }
    } catch (e) {
      logger.warn('scanGate getPersonInfo soft fail', e)
    }
    return (user.userInfo || {}) as Record<string, unknown>
  }

  /** Legacy: only payState === 7 means unpaid gate. */
  function checkUnpaid(profile: Record<string, unknown>): boolean {
    if (Number(profile.payState) === 7) {
      temp.unpaidOrderId = String(temp.unpaidOrderId || '')
      return true
    }
    return false
  }

  /**
   * Legacy getUserRidingType: izRidingType in [5,6,7,9] blocks unlock.
   * 5 platform credit / 6 credit card / 7 none / 9 cancelled
   */
  function checkQualification(profile: Record<string, unknown>): boolean {
    const tpe = Number(profile.izRidingType)
    // Legacy: type 1 → deposit refund page
    if (tpe === 1) {
      uni.showModal({
        title: t('account.noQualification'),
        confirmText: t('account.goDepositRefund'),
        cancelText: t('common.cancel'),
        success: (r) => {
          if (r.confirm) {
            navigate('to', '/pages-sub/pay/deposit-refund/deposit-refund')
          }
        },
      })
      return false
    }
    if (![5, 6, 7, 9].includes(tpe)) return true
    uni.showModal({
      title: t('account.noQualification'),
      confirmText: t('account.goGetQualification'),
      cancelText: t('common.cancel'),
      success: (r) => {
        if (r.confirm) {
          navigate('to', '/pages-sub/account/qualification/qualification')
        }
      },
    })
    return false
  }

  async function checkNeedVerified(profile: Record<string, unknown>): Promise<boolean> {
    if (profile.izNeedAuth === true || profile.izNeedAuth === 1) {
      return !(profile.izAuth === true || profile.izAuth === 1)
    }
    if (Number(profile.authState) === 1 && !profile.izAuth) return true
    return false
  }

  async function ensureCanScan(opts: { skipVerified?: boolean } = {}): Promise<ScanGateResult> {
    user.hydrateFromStorage()
    if (!user.isLoggedIn || !user.userInfo?.pin) {
      navigate('to', '/pages/auth/quick-login')
      return { ok: false, reason: 'login' }
    }

    const profile = await refreshProfile()
    if (!profile.pin && !user.userInfo.pin) {
      navigate('to', '/pages/auth/quick-login')
      return { ok: false, reason: 'login' }
    }

    if (readLimitRiding()) {
      const info = readCreditLimitInfo()
      uni.showModal({
        title: t('pay.creditScore'),
        content: (info && creditTipTitle(info)) || t('pay.creditForbid', { score: info?.score ?? '' }),
        confirmText: t('pay.creditView'),
        showCancel: true,
        success: (r) => {
          if (r.confirm) navigate('to', '/pages-sub/account/credit/credit')
        },
      })
      return { ok: false, reason: 'blocked' }
    }

    const ridingState = Number(profile.ridingState)
    if ([4, 5, 6].includes(ridingState)) {
      navigate('reLaunch', '/pages/riding/riding')
      return { ok: false, reason: 'riding' }
    }

    if (checkUnpaid(profile)) {
      return { ok: false, reason: 'unpaid' }
    }

    // Legacy scanCodeToUseBike: require usable location before opening scanner / enter-id
    const loc = await ensureLocationAuthorized(t('ride.needLocationAuth'))
    if (!loc.ok) {
      return { ok: false, reason: 'cancel' }
    }

    if (!opts.skipVerified) {
      const needVerified = await checkNeedVerified(profile)
      if (needVerified) {
        const gate = await ensureVerifiedOrNavigate({ directReturn: true })
        return {
          ok: false,
          reason: !gate.ok && gate.reason === 'reviewing' ? 'reviewing' : 'verified',
        }
      }
    }

    return { ok: true }
  }

  async function scanThenPrecycling(onUnpaid?: () => void): Promise<ScanGateResult> {
    const gate = await ensureCanScan({ skipVerified: true })
    if (!gate.ok) {
      if (gate.reason === 'unpaid') onUnpaid?.()
      return gate
    }

    await showBeforeUseCarPopup()
    return openPlatformScanner('use')
  }

  /** Scan only to fill a carId field (repair / violation). No unlock gates. */
  async function scanForFillCarId(): Promise<ScanGateResult> {
    return openPlatformScanner('fill')
  }

  async function enterIdThenPrecycling(onUnpaid?: () => void): Promise<ScanGateResult> {
    const gate = await ensureCanScan({ skipVerified: true })
    if (!gate.ok) {
      if (gate.reason === 'unpaid') onUnpaid?.()
      return gate
    }
    await showBeforeUseCarPopup()
    navigate('to', '/pages-sub/ride/enter-id/enter-id')
    return { ok: true }
  }

  /** Full gate including real-name + riding qualification + face — for precycling unlock */
  async function ensureCanUnlock(onUnpaid?: () => void): Promise<ScanGateResult> {
    const gate = await ensureCanScan({ skipVerified: false })
    if (!gate.ok) {
      if (gate.reason === 'unpaid') onUnpaid?.()
      return gate
    }
    const profile = (user.userInfo || {}) as Record<string, unknown>
    if (!checkQualification(profile)) {
      return { ok: false, reason: 'qualification' }
    }
    if (await face.checkNeedFace(profile)) {
      if (temp.ride.carId) {
        temp.setPreCyclingCar({ carId: temp.ride.carId, imei: temp.ride.imei })
      }
      face.goFaceAuth(profile, 'scan_use_bike')
      return { ok: false, reason: 'face' }
    }
    return { ok: true }
  }

  return {
    extractCarId,
    ensureCanScan,
    ensureCanUnlock,
    scanThenPrecycling,
    scanForFillCarId,
    enterIdThenPrecycling,
    checkQualification,
  }
}
