import { ref } from 'vue'
import { getCreditScoreDetail } from '@/api/wechatScore'
import { getTenantConfig } from '@/shared/config'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import { t } from '@/locales'

export type CreditLimitInfo = {
  score?: number | string
  days?: number | string
  endTime?: string
  status?: number
  title?: string
}

const LIMIT_KEY = 'isLimitRiding'
const INFO_KEY = 'creditLimitInfo'

/** Whether tenant enables credit-score riding limit (legacy configData.creditScore). */
export function isCreditScoreEnabled(): boolean {
  const cfg = getTenantConfig() as Record<string, unknown>
  if (cfg.creditScore === false || cfg.creditScore === 0) return false
  const custom = cfg.customSetting as Record<string, unknown> | undefined
  if (custom?.creditScore === false || custom?.creditScore === 0) return false
  // Default on — most tenants use credit; API no-op when feature off
  return true
}

export function readLimitRiding(): boolean {
  const v = storage.get<boolean | string | number>(LIMIT_KEY, false)
  return v === true || v === 'true' || v === 1
}

export function readCreditLimitInfo(): CreditLimitInfo | null {
  return storage.get<CreditLimitInfo | null>(INFO_KEY, null)
}

export function clearLimitRiding() {
  storage.remove(LIMIT_KEY)
  storage.remove(INFO_KEY)
}

export function creditTipTitle(data: CreditLimitInfo): string {
  const score = data.score ?? ''
  if (Number(data.status) === 1) {
    return t('pay.creditWarnTip', { score })
  }
  if (Number(data.status) === 2) {
    return t('pay.creditLimitTip', {
      score,
      days: data.days ?? '',
      endTime: data.endTime ?? '',
    })
  }
  return ''
}

/**
 * Fetch credit limit and sync storage — legacy x-common getUserCreditScore.
 * status 1 = warning tip, 2 = forbid riding + popup.
 */
export async function refreshCreditLimit(opts: {
  pin?: string
  serviceId?: string
  showPopup?: boolean
} = {}): Promise<CreditLimitInfo | null> {
  if (!isCreditScoreEnabled()) {
    clearLimitRiding()
    return null
  }
  const pin =
    opts.pin ||
    String((storage.get<Record<string, unknown>>('userInfo', {}) as { pin?: string })?.pin || '')
  if (!pin) {
    clearLimitRiding()
    return null
  }
  const login = storage.get<{ accessToken?: string; nativeHost?: boolean }>('loginInfo', {}) || {}
  if (!login.accessToken && !login.nativeHost) {
    clearLimitRiding()
    return null
  }

  try {
    const res = await getCreditScoreDetail({
      service_id: opts.serviceId || storage.get<string>('serviceId', '') || '',
      pin,
    })
    if (!res.success || !res.data) {
      clearLimitRiding()
      return null
    }
    const data = res.data as CreditLimitInfo
    const status = Number(data.status)
    if (status === 2) {
      storage.set(LIMIT_KEY, true)
      storage.set(INFO_KEY, data)
      if (opts.showPopup !== false) {
        uni.showModal({
          title: t('pay.creditScore'),
          content: creditTipTitle(data) || t('pay.creditForbid', { score: data.score ?? '' }),
          confirmText: t('pay.creditView'),
          cancelText: t('common.cancel'),
          success: (r) => {
            if (r.confirm) {
              uni.navigateTo({ url: '/pages-sub/account/credit/credit' })
            }
          },
        })
      }
    } else {
      storage.remove(LIMIT_KEY)
      if (status === 1) storage.set(INFO_KEY, data)
      else storage.remove(INFO_KEY)
    }
    return data
  } catch (e) {
    logger.warn('refreshCreditLimit soft fail', e)
    return null
  }
}

/** Reactive helper for home / map scan bar. */
export function useCreditLimit() {
  const isLimitRiding = ref(readLimitRiding())
  const tipList = ref<Array<{ title: string; status?: number }>>([])
  const info = ref<CreditLimitInfo | null>(readCreditLimitInfo())

  function applyLocal() {
    isLimitRiding.value = readLimitRiding()
    info.value = readCreditLimitInfo()
    if (info.value) {
      const title = creditTipTitle(info.value)
      tipList.value = title ? [{ title, status: Number(info.value.status) }] : []
    } else {
      tipList.value = []
    }
  }

  async function refresh(showPopup = false) {
    const data = await refreshCreditLimit({ showPopup })
    applyLocal()
    return data
  }

  applyLocal()

  return {
    isLimitRiding,
    tipList,
    info,
    refresh,
    applyLocal,
  }
}
