import { getServiceByPoi } from '@/api/map'
import { getAllService } from '@/api/service'
import { isNative, nativeHost } from '@/shared/nativeHost'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

type LatLng = { latitude: number; longitude: number }

/**
 * 与小程序首页 `useHomeMap.refreshMap` 一致：用定位调 getByLocation，缓存 `serviceId`。
 * Rider WebView 下优先原生 profile / currentLocation，避免 uni.getLocation 不可用。
 */
export async function ensureServiceId(): Promise<string> {
  const cached = String(storage.get<string>('serviceId', '') || '').trim()
  if (cached) return cached

  if (isNative()) {
    try {
      const profile = await nativeHost()?.getProfile()
      const fromProfile = String(profile?.serviceAreaId || '').trim()
      if (fromProfile) {
        storage.set('serviceId', fromProfile)
        return fromProfile
      }
    } catch (e) {
      logger.warn('ensureServiceId profile soft fail', e)
    }
  }

  const loc = await resolveLocation()
  if (loc) {
    try {
      const service = await getServiceByPoi({
        lat: loc.latitude,
        lng: loc.longitude,
      })
      const id = String((service.data as { id?: string } | undefined)?.id || '').trim()
      if (id) {
        storage.set('serviceId', id)
        return id
      }
    } catch (e) {
      logger.warn('ensureServiceId getServiceByPoi soft fail', e)
    }
  }

  try {
    const all = await getAllService()
    const rows = (Array.isArray(all.data) ? all.data : []) as Array<Record<string, unknown>>
    const id = String(rows[0]?.id || '').trim()
    if (id) {
      storage.set('serviceId', id)
      return id
    }
  } catch (e) {
    logger.warn('ensureServiceId getAllService soft fail', e)
  }

  return ''
}

async function resolveLocation(): Promise<LatLng | null> {
  if (isNative()) {
    try {
      const loc = await nativeHost()?.currentLocation()
      const lat = Number(loc?.latitude)
      const lng = Number(loc?.longitude)
      if (Number.isFinite(lat) && Number.isFinite(lng)) {
        return { latitude: lat, longitude: lng }
      }
    } catch (e) {
      logger.warn('ensureServiceId native location soft fail', e)
    }
  }

  return new Promise((resolve) => {
    uni.getLocation({
      type: 'gcj02',
      success: (res) => resolve({ latitude: res.latitude, longitude: res.longitude }),
      fail: (err) => {
        logger.warn('ensureServiceId uni.getLocation fail', err)
        resolve(null)
      },
    })
  })
}
