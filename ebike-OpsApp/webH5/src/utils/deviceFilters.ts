import type { DeviceItem } from '@/api'

const HOUR_MS = 3600 * 1000

/**
 * 闲置时长分桶：最近一次锁车距今落在 [beginHours, endHours) 的车辆。
 * `lockTime <= unlockTime` 说明车还在骑行中，不算闲置。
 */
export function filterByIdleHours(
  devices: DeviceItem[],
  beginHours: number,
  endHours: number,
): DeviceItem[] {
  const now = Date.now()
  return devices.filter((device) => {
    const lockTime = Number(device.lockTime ?? 0)
    const unlockTime = Number(device.unlockTime ?? 0)
    if (lockTime === 0 || lockTime <= unlockTime) return false
    const idle = now - lockTime
    return idle >= beginHours * HOUR_MS && idle < endHours * HOUR_MS
  })
}

/**
 * 电量分桶。遗留口径是左开右闭 `(begin, end]`，而 0% 这一档是精确等于 0，
 * 所以 begin === end === 0 时走等值判断。
 */
export function filterByBattery(
  devices: DeviceItem[],
  beginPercent: number,
  endPercent: number,
): DeviceItem[] {
  if (beginPercent === endPercent) {
    return devices.filter((device) => device.restBattery === beginPercent)
  }
  return devices.filter((device) => {
    const battery = device.restBattery ?? -1
    return battery > beginPercent && battery <= endPercent
  })
}
