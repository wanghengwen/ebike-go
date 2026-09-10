import { getTenantConfig } from '@/shared/config'

/** Legacy isTimeExpired — true when missing or past. */
export function isTimeExpired(val: unknown): boolean {
  if (val == null || val === '') return true
  const target = new Date(String(val)).getTime()
  if (!Number.isFinite(target)) {
    const n = Number(val)
    if (!Number.isFinite(n)) return true
    return n <= Date.now()
  }
  return target <= Date.now()
}

export function getRedEnvelopeCfg(key: string): string {
  return getTenantConfig().customSetting?.redEnvelopeCfg?.[key] || ''
}

/** Bike qualifies as red-envelope car when activity not expired. */
export function isRedEnvelopeBike(item?: Record<string, unknown> | null): boolean {
  if (!item) return false
  if (item.izRedPaket === true || item.izRedPaket === 1 || item.isRedEnvelope === true) return true
  const activityId = item.activityId ?? item.redEnvelopeActivityId
  if (activityId == null || activityId === '') return false
  return !isTimeExpired(item.expirationTime)
}

export function redEnvelopeActivityId(item?: Record<string, unknown> | null): string {
  if (!item || !isRedEnvelopeBike(item)) return ''
  const id = item.activityId ?? item.redEnvelopeActivityId
  return id != null && id !== '' ? String(id) : ''
}

/** Reward yuan from order deduct fields (fen). */
export function redEnvelopeRewardYuan(detail?: Record<string, unknown> | null): string {
  if (!detail) return '0.00'
  const fen =
    Number(detail.redPaketCarDeduct || 0) + Number(detail.redPaketParkingDeduct || 0)
  return (fen / 100).toFixed(2)
}

export function mapLegacyRedEnvelopePath(url: string): string {
  if (/redEnvelopeTips/i.test(url)) {
    const m = /[?&]id=([^&]+)/i.exec(url)
    const id = m?.[1] ? decodeURIComponent(m[1]) : ''
    return id
      ? `/pages-sub/ride/red-envelope-tips/red-envelope-tips?id=${encodeURIComponent(id)}`
      : '/pages-sub/ride/red-envelope-tips/red-envelope-tips'
  }
  return url
}
