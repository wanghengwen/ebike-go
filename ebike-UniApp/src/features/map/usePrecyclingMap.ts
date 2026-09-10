import { computed, ref } from 'vue'
import {
  getFenceByServiceId,
  getHomeActivityEntrance,
  getHomeActivityEntranceById,
  getNearBike,
  getNearFence,
} from '@/api/map'
import {
  buildBikeMarkers,
  buildCarMarker,
  buildFencePolygons,
  fenceToPolygon,
  toPoints,
  type LatLng,
  type MapMarker,
  type MapPolygon,
} from '@/features/map/mapFenceUtils'
import { mapLegacyRedEnvelopePath } from '@/features/bike/redEnvelope'
import { useMapLocation } from '@/features/map/useMapLocation'
import { useTempDataStore } from '@/stores/tempData'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import { navigate } from '@/shared/navigate'

function hasAccessToken(): boolean {
  const login = storage.get<{ accessToken?: string }>('loginInfo', {}) || {}
  return Boolean(login.accessToken)
}

export type ActivityEntrance = {
  id?: string | number
  linkUrl?: string
  linkTitle?: string
  title?: string
  name?: string
  appId?: string
  param?: string
  iconUrl?: string
  imageUrl?: string
  picUrl?: string
  position?: number
}

/** Map layer for precycling — car pin + fences + optional near bikes / activity. */
export function usePrecyclingMap() {
  const temp = useTempDataStore()
  const { locate } = useMapLocation()

  const latitude = ref(30.25)
  const longitude = ref(120.15)
  const carMarker = ref<MapMarker | null>(null)
  const fenceMarkers = ref<MapMarker[]>([])
  const fencePolygons = ref<MapPolygon[]>([])
  const nearBikeMarkers = ref<MapMarker[]>([])
  const activityEntrances = ref<ActivityEntrance[]>([])
  const mapReady = ref(false)

  const markers = computed(() => {
    const list: MapMarker[] = [...fenceMarkers.value, ...nearBikeMarkers.value]
    if (carMarker.value) list.push(carMarker.value)
    return list
  })
  const polygons = computed(() => fencePolygons.value)

  const floatingActs = computed(() =>
    activityEntrances.value.filter((a) => {
      const hasImg = Boolean(a.picUrl || a.imageUrl || a.iconUrl)
      if (!hasImg) return false
      const pos = Number(a.position)
      // Legacy floating slots: 0/2 top, 1/3 bottom
      return !Number.isFinite(pos) || [0, 1, 2, 3].includes(pos)
    }),
  )

  function centerOn(lat: number, lng: number) {
    if (Number.isNaN(lat) || Number.isNaN(lng)) return
    latitude.value = lat
    longitude.value = lng
  }

  function setCarFromInfo(info: Record<string, unknown>, isRedEnvelope = false) {
    const marker = buildCarMarker(info, { isRedEnvelope })
    carMarker.value = marker
    if (marker) {
      centerOn(marker.latitude, marker.longitude)
      mapReady.value = true
    }
  }

  async function loadFences(
    sid: string,
    loc: LatLng,
    servicePointList?: unknown,
    opts: { isRedEnvelope?: boolean } = {},
  ) {
    const polygonsAcc: MapPolygon[] = []
    if (servicePointList) {
      const poly = fenceToPolygon(toPoints(servicePointList), 0)
      if (poly) polygonsAcc.push(poly)
    }
    const fenceOpts = {
      redEnvelopeMode: Boolean(opts.isRedEnvelope),
      onlyShowRedEnvelope: false,
    }

    try {
      if (sid && hasAccessToken()) {
        const byId = await getFenceByServiceId({ id: sid })
        if (byId.success && byId.data) {
          const data = byId.data as {
            serviceAreas?: Array<Record<string, unknown>>
            parkings?: Array<Record<string, unknown>>
            noParkings?: Array<Record<string, unknown>>
          }
          const built = buildFencePolygons(data, fenceOpts)
          if (built.polygons.length) {
            fencePolygons.value = built.polygons
            fenceMarkers.value = built.markers
            return
          }
        }
      }
    } catch (e) {
      logger.warn('precycling getFenceByServiceId soft fail', e)
    }

    try {
      const near = await getNearFence({
        locationDTO: { lat: loc.latitude, lng: loc.longitude },
        serviceId: sid,
      })
      if (near.success && near.data) {
        const data = near.data as {
          parkings?: Array<Record<string, unknown>>
          noParkings?: Array<Record<string, unknown>>
        }
        const built = buildFencePolygons(
          {
            pointList: servicePointList,
            parkings: data.parkings,
            noParkings: data.noParkings,
          },
          fenceOpts,
        )
        fencePolygons.value = built.polygons.length ? built.polygons : polygonsAcc
        fenceMarkers.value = built.markers
        return
      }
    } catch (e) {
      logger.warn('precycling getNearFence soft fail', e)
    }

    fencePolygons.value = polygonsAcc
    fenceMarkers.value = []
  }

  async function loadNearBikes(
    sid: string,
    loc: LatLng,
    opts: { excludeCarId?: string; asMarkers?: boolean } = {},
  ): Promise<{ list: Array<Record<string, unknown>>; count: number; nearestM: number }> {
    if (!sid) {
      nearBikeMarkers.value = []
      return { list: [], count: 0, nearestM: 0 }
    }
    try {
      const res = await getNearBike({
        lat: loc.latitude,
        lng: loc.longitude,
        serviceId: sid,
      })
      const list = (Array.isArray(res.data) ? res.data : []) as Array<Record<string, unknown>>
      temp.setNearBikes(list)
      const others = opts.excludeCarId
        ? list.filter((b) => String(b.carId || b.id || '') !== opts.excludeCarId)
        : list
      if (opts.asMarkers) {
        nearBikeMarkers.value = buildBikeMarkers(others, {
          excludeCarId: opts.excludeCarId,
        }).markers
      } else {
        nearBikeMarkers.value = []
      }
      const dists = others
        .map((b) => Number(b.distance ?? b.dis ?? NaN))
        .filter((n) => Number.isFinite(n))
      return {
        list: others,
        count: others.length,
        nearestM: dists.length ? Math.round(Math.min(...dists)) : 0,
      }
    } catch (e) {
      logger.warn('precycling near bikes soft fail', e)
      nearBikeMarkers.value = []
      return { list: [], count: 0, nearestM: 0 }
    }
  }

  async function loadActivityEntrances(sid: string) {
    if (!sid) {
      activityEntrances.value = []
      return
    }
    try {
      const pin = String(storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
      const res = await getHomeActivityEntrance({ serviceId: sid, userPin: pin })
      if (res.success && res.data) {
        activityEntrances.value = Array.isArray(res.data)
          ? (res.data as ActivityEntrance[])
          : []
      } else {
        activityEntrances.value = []
      }
    } catch (e) {
      logger.warn('precycling activity soft fail', e)
      activityEntrances.value = []
    }
  }

  async function openActivityEntrance(item: ActivityEntrance) {
    if (!item) return
    if (item.id != null) {
      try {
        const pin = String(storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
        const sid = String(storage.get<string>('serviceId', '') || '')
        const detail = await getHomeActivityEntranceById({
          serviceId: sid,
          userPin: pin,
          id: item.id,
        })
        const data = detail.data as { izJumpReminder?: boolean; jumpReminder?: boolean } | undefined
        if (detail.success && data && (data.izJumpReminder || data.jumpReminder)) {
          return
        }
      } catch (e) {
        logger.warn('precycling activity detail soft fail', e)
      }
    }
    const url = String(item.linkUrl || '')
    if (!url) return
    const mapped = mapLegacyRedEnvelopePath(url)
    if (mapped.startsWith('/pages') || mapped.startsWith('/pages-sub')) {
      navigate('to', mapped)
      return
    }
    if (/^https?:\/\//i.test(mapped)) {
      navigate('to', `/pages/webview/webview?url=${encodeURIComponent(mapped)}`)
    }
  }

  /** After prepareUnlock: center car, load fences + optional near markers + activity. */
  async function hydrateFromCar(
    info: Record<string, unknown>,
    opts: { isRedEnvelope?: boolean; showNearMarkers?: boolean; carId?: string } = {},
  ) {
    setCarFromInfo(info, Boolean(opts.isRedEnvelope))
    const sid = String(info.serviceId || storage.get<string>('serviceId', '') || '')
    if (sid) storage.set('serviceId', sid)

    const lat = Number(info.lat ?? info.latitude)
    const lng = Number(info.lng ?? info.longitude)
    let carLoc: LatLng
    if (!Number.isNaN(lat) && !Number.isNaN(lng)) {
      carLoc = { latitude: lat, longitude: lng }
    } else if (temp.location) {
      carLoc = temp.location
    } else {
      carLoc = { latitude: latitude.value, longitude: longitude.value }
    }

    // Center map on car; near bikes use USER location (legacy userPoiFun)
    centerOn(carLoc.latitude, carLoc.longitude)
    const userLoc = temp.location || (await locate()) || carLoc

    const near = await loadNearBikes(sid, userLoc, {
      excludeCarId: opts.carId,
      asMarkers: Boolean(opts.showNearMarkers),
    })
    await Promise.all([
      loadFences(sid, carLoc, undefined, { isRedEnvelope: Boolean(opts.isRedEnvelope) }),
      loadActivityEntrances(sid),
    ])
    mapReady.value = true
    return { serviceId: sid, near }
  }

  return {
    latitude,
    longitude,
    markers,
    polygons,
    mapReady,
    carMarker,
    activityEntrances,
    floatingActs,
    setCarFromInfo,
    hydrateFromCar,
    loadNearBikes,
    loadFences,
    loadActivityEntrances,
    openActivityEntrance,
  }
}
