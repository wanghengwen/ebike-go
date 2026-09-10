import { getTenantConfig } from '@/shared/config'

/** Round battery % to nearest mapCfg icon bucket (0,10,...,100). */
export function batteryIconBucket(pct: unknown): number {
  const n = Number(pct)
  if (!Number.isFinite(n) || n <= 0) return 0
  if (n >= 100) return 100
  return Math.round(n / 10) * 10
}

/** Tenant mapCfg icon_ebike_battery{N} — falls back to empty (platform default pin). */
export function getBikeBatteryIcon(restBattery?: unknown): string {
  const mapCfg = getTenantConfig().customSetting?.mapCfg || {}
  const bucket = batteryIconBucket(restBattery)
  return (
    mapCfg[`icon_ebike_battery${bucket}`] ||
    mapCfg.iconEbike ||
    mapCfg.icon_ebike ||
    ''
  )
}

export function getIconCfg(key: string): string {
  return getTenantConfig().customSetting?.iconCfg?.[key] || ''
}

export function getMapCfg(key: string): string {
  return getTenantConfig().customSetting?.mapCfg?.[key] || ''
}

export function getCyclingCfg(key: string): string {
  const c = getTenantConfig().customSetting?.cyclingCfg as Record<string, unknown> | undefined
  return String(c?.[key] || '')
}
