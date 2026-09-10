import { computed, ref } from 'vue'
import {
  getFenceByServiceId,
  getHomeActivityEntrance,
  getHomeActivityEntranceById,
  getHomeScrollerMsg,
  getNearBike,
  getNearFence,
  getServiceByPoi,
  navigateToCarOrP,
  playBikeVoice,
} from '@/api/map'
import { useTempDataStore } from '@/stores/tempData'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import { navigate } from '@/shared/navigate'
import { useMapLocation } from '@/features/map/useMapLocation'
import {
  buildBikeMarkers as buildBikeMarkersFromList,
  buildFencePolygons,
  fenceToPolygon,
  toPoints,
  type LatLng,
  type MapMarker,
  type MapPolygon,
  type MapPolyline,
} from '@/features/map/mapFenceUtils'
import { mapLegacyRedEnvelopePath } from '@/features/bike/redEnvelope'
import { t } from '@/locales'

function hasAccessToken(): boolean {
  const login = storage.get<{ accessToken?: string }>('loginInfo', {}) || {}
  return Boolean(login.accessToken)
}

export type { LatLng, MapMarker, MapPolygon, MapPolyline }

export function useHomeMap() {
  const temp = useTempDataStore()
  const { locate } = useMapLocation()

  const latitude = ref(30.25)
  const longitude = ref(120.15)
  const serviceId = ref(String(storage.get<string>('serviceId', '') || ''))
  const loading = ref(false)
  const bikeMarkers = ref<MapMarker[]>([])
  const fenceMarkers = ref<MapMarker[]>([])
  const fencePolygons = ref<MapPolygon[]>([])
  const polyline = ref<MapPolyline[]>([])
  const scrollerMsg = ref<
    Array<{
      content?: string
      title?: string
      id?: string
      skipUrl?: string
      linkUrl?: string
      appid?: string
      appId?: string
      params?: unknown
      type?: number | string
    }>
  >([])
  const activityEntrances = ref<
    Array<{
      id?: string | number
      linkUrl?: string
      linkTitle?: string
      title?: string
      name?: string
      appId?: string
      param?: string
      iconUrl?: string
      imageUrl?: string
    }>
  >([])
  const selectedBike = ref<MapMarker | null>(null)
  /** Legacy isRedEnvelopeMap — only show active red-envelope bikes */
  const redEnvelopeMapMode = ref(false)
  const redEnvelopeCount = ref(0)

  const markers = computed(() => [...fenceMarkers.value, ...bikeMarkers.value])
  const polygons = computed(() => fencePolygons.value)
  const bikeCount = computed(() => bikeMarkers.value.filter((m) => m.type === 2).length)

  function applyLocation(loc: LatLng) {
    latitude.value = loc.latitude
    longitude.value = loc.longitude
    temp.setLocation(loc.latitude, loc.longitude)
  }

  async function ensureLocation(): Promise<LatLng | null> {
    if (temp.location) {
      applyLocation(temp.location)
      return temp.location
    }
    const loc = await locate()
    if (loc) applyLocation(loc)
    return loc
  }

  function buildBikeMarkers(list: Array<Record<string, unknown>>): MapMarker[] {
    const built = buildBikeMarkersFromList(list, { redEnvelopeOnly: redEnvelopeMapMode.value })
    redEnvelopeCount.value = built.redCount
    return built.markers
  }

  async function loadNearBikes(loc: LatLng, sid: string) {
    if (!sid) {
      bikeMarkers.value = []
      temp.setNearBikes([])
      return []
    }
    try {
      const res = await getNearBike({
        lat: loc.latitude,
        lng: loc.longitude,
        serviceId: sid,
      })
      const list = (Array.isArray(res.data) ? res.data : []) as Array<Record<string, unknown>>
      temp.setNearBikes(list)
      bikeMarkers.value = buildBikeMarkers(list)
      return list
    } catch (e) {
      logger.warn('loadNearBikes fail', e)
      return []
    }
  }

  async function loadFences(loc: LatLng, sid: string, servicePointList?: unknown) {
    const polygonsAcc: MapPolygon[] = []
    if (servicePointList) {
      const poly = fenceToPolygon(toPoints(servicePointList), 0)
      if (poly) polygonsAcc.push(poly)
    }
    const fenceOpts = {
      redEnvelopeMode: redEnvelopeMapMode.value,
      onlyShowRedEnvelope: redEnvelopeMapMode.value,
    }

    try {
      // Production security exclude list does NOT include getFenceByServiceId —
      // guests get 00006; use getNearFence (auth-free) instead.
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
      logger.warn('getFenceByServiceId soft fail', e)
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
      logger.warn('getNearFence soft fail', e)
    }

    fencePolygons.value = polygonsAcc
    fenceMarkers.value = []
  }

  async function loadScroller(sid: string) {
    if (!sid) {
      scrollerMsg.value = []
      return
    }
    try {
      const res = await getHomeScrollerMsg({ serviceId: sid })
      if (res.success && res.data) {
        const data = res.data as { content?: string; title?: string; id?: string } | Array<{ content?: string }>
        scrollerMsg.value = Array.isArray(data) ? data : [data]
      } else {
        scrollerMsg.value = []
      }
    } catch {
      scrollerMsg.value = []
    }
  }

  async function loadActivityEntrances(sid: string) {
    if (!sid) {
      activityEntrances.value = []
      return
    }
    try {
      const pin = String(
        storage.get<Record<string, unknown>>('userInfo', {})?.pin || '',
      )
      const res = await getHomeActivityEntrance({ serviceId: sid, userPin: pin })
      if (res.success && res.data) {
        activityEntrances.value = Array.isArray(res.data)
          ? (res.data as typeof activityEntrances.value)
          : []
      } else {
        activityEntrances.value = []
      }
    } catch (e) {
      logger.warn('getHomeActivityEntrance soft fail', e)
      activityEntrances.value = []
    }
  }

  async function openActivityEntrance(item: (typeof activityEntrances.value)[number]) {
    if (!item) return
    if (item.id != null) {
      try {
        const pin = String(storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
        const detail = await getHomeActivityEntranceById({
          serviceId: serviceId.value,
          userPin: pin,
          id: item.id,
        })
        const data = detail.data as { izJumpReminder?: boolean; jumpReminder?: boolean } | undefined
        if (detail.success && data && (data.izJumpReminder || data.jumpReminder)) {
          return
        }
      } catch (e) {
        logger.warn('getHomeActivityEntranceById soft fail', e)
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

  async function refreshMap(opts: { keepSelection?: boolean; redEnvelopeOnly?: boolean } = {}) {
    if (opts.redEnvelopeOnly != null) redEnvelopeMapMode.value = Boolean(opts.redEnvelopeOnly)
    loading.value = true
    try {
      const loc = await ensureLocation()
      if (!loc) return null

      let sid = serviceId.value
      let servicePointList: unknown
      try {
        const service = await getServiceByPoi({
          lat: loc.latitude,
          lng: loc.longitude,
        })
        if (service.success && service.data) {
          const data = service.data as { id?: string; pointList?: unknown }
          sid = String(data.id || '')
          servicePointList = data.pointList
          serviceId.value = sid
          if (sid) storage.set('serviceId', sid)
        }
      } catch (e) {
        logger.warn('getServiceByPoi soft fail', e)
      }

      await Promise.all([
        loadNearBikes(loc, sid),
        loadFences(loc, sid, servicePointList),
        loadScroller(sid),
        loadActivityEntrances(sid),
      ])

      if (!opts.keepSelection) selectedBike.value = null
      return { loc, serviceId: sid }
    } finally {
      loading.value = false
    }
  }

  function setRedEnvelopeMapMode(on: boolean) {
    redEnvelopeMapMode.value = on
  }

  async function relocate() {
    const loc = await locate()
    if (loc) applyLocation(loc)
    await refreshMap()
    return loc
  }

  function findMarkerById(markerId: number | string): MapMarker | undefined {
    const id = Number(markerId)
    return bikeMarkers.value.find((m) => m.id === id || String(m.carId) === String(markerId))
  }

  async function onMarkerTap(markerId: number | string) {
    const marker = findMarkerById(markerId)
    if (!marker || marker.type !== 2) {
      selectedBike.value = null
      return null
    }
    selectedBike.value = marker
    return marker
  }

  async function ringSelectedBike(bike?: MapMarker | null) {
    const target = bike || selectedBike.value
    const imeiRaw = target?.imei ?? target?.sourceData?.imei
    const imei = imeiRaw != null ? String(imeiRaw) : undefined
    const carId = target?.carId
    if (!imei && !carId) return { success: false as const }
    uni.showLoading({ title: t('ride.findBike'), mask: true })
    try {
      const res = await playBikeVoice({ carId, imei })
      uni.showToast({
        title: res.success ? t('ride.ringSuccess') : t('ride.ringFail'),
        icon: 'none',
      })
      return res
    } finally {
      uni.hideLoading()
    }
  }

  async function navigateToMarker(marker?: MapMarker | null) {
    const target = marker || selectedBike.value
    const loc = temp.location || (await ensureLocation())
    if (!target || !loc) return
    try {
      const res = await navigateToCarOrP({
        type: 1,
        origin: `${loc.longitude},${loc.latitude}`,
        destination: `${target.longitude},${target.latitude}`,
      })
      if (res.success && res.data) {
        const data = res.data as {
          polyline?: Array<{ lat?: number; lng?: number; latitude?: number; longitude?: number }>
          distance?: number
          duration?: number
        }
        polyline.value = [
          {
            points: (data.polyline || []).map((i) => ({
              latitude: Number(i.latitude ?? i.lat),
              longitude: Number(i.longitude ?? i.lng),
            })),
            color: '#1890ff',
            width: 6,
            arrowLine: true,
            dottedLine: false,
          },
        ]
        const meters = data.distance != null ? Math.round(Number(data.distance)) : null
        const mins =
          data.duration != null ? Math.round(Number(data.duration) / 60) : null
        if (meters != null && mins != null) {
          const content = t('ride.walkCallout', { m: meters, min: mins })
          const applyCallout = (list: MapMarker[]) =>
            list.map((m) => {
              if (m.id === target.id) {
                return {
                  ...m,
                  callout: {
                    content,
                    display: 'ALWAYS',
                    color: '#333333',
                    fontSize: 12,
                    borderRadius: 8,
                    padding: 6,
                    bgColor: '#ffffff',
                    textAlign: 'center',
                  },
                }
              }
              const { callout: _c, ...rest } = m
              return rest
            })
          bikeMarkers.value = applyCallout(bikeMarkers.value)
          fenceMarkers.value = applyCallout(fenceMarkers.value)
        }
      }
    } catch (e) {
      logger.warn('navigateToMarker fail', e)
    }
  }

  function clearRoute() {
    polyline.value = []
  }

  return {
    latitude,
    longitude,
    serviceId,
    loading,
    markers,
    polygons,
    polyline,
    bikeCount,
    scrollerMsg,
    activityEntrances,
    selectedBike,
    redEnvelopeMapMode,
    redEnvelopeCount,
    ensureLocation,
    refreshMap,
    relocate,
    onMarkerTap,
    ringSelectedBike,
    navigateToMarker,
    clearRoute,
    findMarkerById,
    openActivityEntrance,
    setRedEnvelopeMapMode,
  }
}
