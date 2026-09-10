import { getAddressByLatAndLng } from '@/api/common'
import { logger } from '@/shared/logger'
import { t } from '@/locales'

export type MapAppTarget = {
  latitude: number
  longitude: number
  destination?: string
  name?: string
}

/**
 * Open WeChat / system third-party map picker (legacy x-map openMapApp).
 * Falls back to uni.openLocation when MapContext.openMapApp is unavailable.
 */
export async function openThirdPartyMap(
  mapId: string,
  target: MapAppTarget,
): Promise<boolean> {
  const latitude = Number(target.latitude)
  const longitude = Number(target.longitude)
  if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) return false

  let destination = target.destination || target.name || ''
  if (!destination || destination === '--') {
    try {
      const res = await getAddressByLatAndLng({ latitude, longitude })
      const data = res.data as { address?: string; name?: string; formattedAddress?: string } | string
      if (typeof data === 'string') destination = data
      else destination = String(data?.formattedAddress || data?.address || data?.name || '')
      if (destination === '--') destination = ''
    } catch (e) {
      logger.warn('getAddressByLatAndLng soft fail', e)
    }
  }
  if (!destination || destination === '--') {
    destination = t('ride.parkSearch')
  }

  try {
    const ctx = uni.createMapContext(mapId)
    const openMapApp = (ctx as { openMapApp?: (opt: Record<string, unknown>) => void }).openMapApp
    if (typeof openMapApp === 'function') {
      await new Promise<void>((resolve, reject) => {
        openMapApp.call(ctx, {
          longitude,
          latitude,
          destination,
          success: () => resolve(),
          fail: (err: unknown) => reject(err),
        })
      })
      return true
    }
  } catch (e) {
    logger.warn('openMapApp fail, fallback openLocation', e)
  }

  return new Promise((resolve) => {
    uni.openLocation({
      latitude,
      longitude,
      name: destination,
      scale: 16,
      success: () => resolve(true),
      fail: () => resolve(false),
    })
  })
}
