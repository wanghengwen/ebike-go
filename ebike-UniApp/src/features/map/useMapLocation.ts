import { getNearBike, getServiceByPoi } from '@/api/map'
import { isNative, nativeHost } from '@/shared/nativeHost'
import { useTempDataStore } from '@/stores/tempData'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

export function useMapLocation() {
  const temp = useTempDataStore()

  async function locate(): Promise<{ latitude: number; longitude: number } | null> {
    if (isNative()) {
      try {
        const loc = await nativeHost()?.currentLocation()
        const lat = Number(loc?.latitude)
        const lng = Number(loc?.longitude)
        if (Number.isFinite(lat) && Number.isFinite(lng)) {
          temp.setLocation(lat, lng)
          return { latitude: lat, longitude: lng }
        }
      } catch (e) {
        logger.warn('native locate fail', e)
      }
    }
    return new Promise((resolve) => {
      uni.getLocation({
        type: 'gcj02',
        success: (res) => {
          temp.setLocation(res.latitude, res.longitude)
          resolve({ latitude: res.latitude, longitude: res.longitude })
        },
        fail: (err) => {
          logger.warn('locate fail', err)
          resolve(null)
        },
      })
    })
  }

  async function refreshNearBikes() {
    const loc = temp.location || (await locate())
    if (!loc) return []
    const service = await getServiceByPoi({
      lat: loc.latitude,
      lng: loc.longitude,
    })
    const serviceId = (service.data as { id?: string } | undefined)?.id
    if (serviceId) storage.set('serviceId', String(serviceId))
    const bikes = await getNearBike({
      lat: loc.latitude,
      lng: loc.longitude,
      serviceId,
    })
    const list = (Array.isArray(bikes.data) ? bikes.data : []) as Array<Record<string, unknown>>
    temp.setNearBikes(list)
    return list
  }

  return { locate, refreshNearBikes }
}
