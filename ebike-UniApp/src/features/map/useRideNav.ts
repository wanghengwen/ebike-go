import { ref } from 'vue'
import {
  getFenceByServiceId,
  getNearFenceByPoiAndService,
  navigateToCarOrP,
  turnOffNavigation,
  turnOnNavigation,
} from '@/api/map'
import { closestPoint } from '@/features/map/polylinePoint'
import { useMapLocation } from '@/features/map/useMapLocation'
import { useTempDataStore, type AppNavigation } from '@/stores/tempData'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import { t } from '@/locales'

function hasAccessToken(): boolean {
  const login = storage.get<{ accessToken?: string }>('loginInfo', {}) || {}
  return Boolean(login.accessToken)
}

export type NavPoint = { latitude: number; longitude: number }

export type NavPolyline = {
  points: NavPoint[]
  color: string
  width: number
  arrowLine: boolean
  dottedLine: boolean
}

export type ParkMarker = {
  id: number
  latitude: number
  longitude: number
  width: number
  height: number
  title?: string
  iconPath?: string
  type?: number
  /** Parking full — hide on riding map */
  fullCar?: boolean
  callout?: {
    content?: string
    display?: string
    color?: string
    fontSize?: number
    borderRadius?: number
    padding?: number
    bgColor?: string
    textAlign?: string
  }
}

export type NavPolygon = {
  points: NavPoint[]
  strokeWidth: number
  strokeColor: string
  fillColor: string
  dashArray?: number[]
  /** Parking full — hide on riding map */
  fullCar?: boolean
}

const FENCE_COLORS: Record<number, [string, string]> = {
  0: ['#1890ff', '#1890ff33'],
  1: ['#52c41a', '#52c41a33'],
  2: ['#ff4d4f', '#ff4d4f33'],
}

type PolyPt = { lat: number; lng: number }

function toPolyPts(
  list: Array<{ lat?: number; lng?: number; latitude?: number; longitude?: number }> = [],
): PolyPt[] {
  return list
    .map((i) => ({
      lat: Number(i.lat ?? i.latitude),
      lng: Number(i.lng ?? i.longitude),
    }))
    .filter((i) => Number.isFinite(i.lat) && Number.isFinite(i.lng))
}

function toLatLngList(raw: unknown): NavPoint[] {
  if (!Array.isArray(raw)) return []
  return raw
    .map((p) => {
      if (Array.isArray(p) && p.length >= 2) {
        return { latitude: Number(p[1]), longitude: Number(p[0]) }
      }
      const o = p as Record<string, unknown>
      return {
        latitude: Number(o.lat ?? o.latitude),
        longitude: Number(o.lng ?? o.longitude),
      }
    })
    .filter((p) => Number.isFinite(p.latitude) && Number.isFinite(p.longitude))
}

/**
 * Riding / park-search navigation helpers.
 * Legacy riding goNearPark: stay on page, draw walk line to nearest P.
 */
export function useRideNav() {
  const temp = useTempDataStore()
  const { locate } = useMapLocation()
  const polyline = ref<NavPolyline[]>([])
  const parkMarkers = ref<ParkMarker[]>([])
  const fencePolygons = ref<NavPolygon[]>([])
  const routeMeta = ref<{ distance?: number; durationMin?: number; instruction?: string }>({})
  const instrumentActive = ref(false)
  /** Last navigateToCarOrP payload for park-search → riding handoff */
  const lastNavPayload = ref<AppNavigation | null>(null)

  function serviceId() {
    return String(
      storage.get<string>('serviceId', '') ||
        (storage.get<Record<string, unknown>>('userInfo', {}) as { serviceId?: string })?.serviceId ||
        '',
    )
  }

  function clearRoute() {
    polyline.value = []
    routeMeta.value = {}
    lastNavPayload.value = null
    parkMarkers.value = parkMarkers.value.filter((m) => m.id !== -55)
  }

  function clearInstrument() {
    instrumentActive.value = false
    clearRoute()
  }

  /** Draw instrument nav: remaining blue + passed grey + compass marker. */
  async function drawInstrumentLine(rawPoly: PolyPt[]) {
    if (!rawPoly.length) return
    const loc = temp.location || (await locate())
    const user = loc
      ? { lat: loc.latitude, lng: loc.longitude }
      : { lat: rawPoly[0].lat, lng: rawPoly[0].lng }
    const cut = closestPoint(rawPoly, user)
    const passed = rawPoly.slice(0, Math.max(cut, 1))
    const cursor = passed[passed.length - 1] || rawPoly[0]
    polyline.value = [
      {
        points: rawPoly.map((i) => ({ latitude: i.lat, longitude: i.lng })),
        color: '#006efe',
        width: 6,
        arrowLine: true,
        dottedLine: false,
      },
      {
        points: passed.map((i) => ({ latitude: i.lat, longitude: i.lng })),
        color: '#97999b',
        width: 6,
        arrowLine: true,
        dottedLine: false,
      },
    ]
    const compass = getIconCfg('navigate_compass') || getMapCfg('centerMarker')
    const others = parkMarkers.value.filter((m) => m.id !== -55)
    others.push({
      id: -55,
      latitude: cursor.lat,
      longitude: cursor.lng,
      width: 28,
      height: 28,
      ...(compass ? { iconPath: compass } : {}),
    })
    parkMarkers.value = others
    instrumentActive.value = true
  }

  /**
   * Draw route via navigateToCarOrP.
   * Walking (type=1) on riding page uses legacy origin=P / destination=user swap.
   * Riding nav (type=2) uses origin=user / destination=P.
   */
  async function drawRoute(
    park: NavPoint,
    navType: 1 | 2 = 1,
  ): Promise<{ success: boolean; msg?: string }> {
    const loc = temp.location || (await locate())
    if (!loc) {
      return { success: false, msg: t('ride.needLocation') }
    }
    const user = `${loc.longitude},${loc.latitude}`
    const dest = `${park.longitude},${park.latitude}`
    const body =
      navType === 1
        ? { type: 1, origin: dest, destination: user }
        : { type: 2, origin: user, destination: dest }

    try {
      const res = await navigateToCarOrP(body)
      if (!res.success || !res.data) {
        clearRoute()
        return { success: false, msg: res.msg || t('ride.navFail') }
      }
      const data = res.data as {
        polyline?: Array<{ lat?: number; lng?: number; latitude?: number; longitude?: number }>
        distance?: number
        duration?: number
      }
      const pts = toPolyPts(data.polyline || [])
      polyline.value = [
        {
          points: pts.map((i) => ({ latitude: i.lat, longitude: i.lng })),
          color: navType === 2 ? '#1890ff' : '#006efe',
          width: 6,
          arrowLine: true,
          dottedLine: false,
        },
      ]
      routeMeta.value = {
        distance: data.distance != null ? Number(data.distance) : undefined,
        durationMin: data.duration != null ? Math.round(Number(data.duration) / 60) : undefined,
      }
      lastNavPayload.value = {
        ...body,
        ...data,
        origin: body.origin,
        destination: body.destination,
        type: navType,
        pathType: navType,
      }
      // Bubble on destination park marker when present
      if (routeMeta.value.distance != null && routeMeta.value.durationMin != null) {
        const content =
          navType === 2
            ? t('ride.rideCallout', {
                m: Math.round(routeMeta.value.distance),
                min: routeMeta.value.durationMin,
              })
            : t('ride.walkCallout', {
                m: Math.round(routeMeta.value.distance),
                min: routeMeta.value.durationMin,
              })
        parkMarkers.value = parkMarkers.value.map((m) => {
          const near =
            Math.abs(m.latitude - park.latitude) < 0.00001 &&
            Math.abs(m.longitude - park.longitude) < 0.00001
          if (near || m.id === 9001) {
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
      }
      return { success: true }
    } catch (e) {
      logger.warn('drawRoute fail', e)
      clearRoute()
      return { success: false, msg: t('ride.navFail') }
    }
  }

  /** Closest parking from getNearFence + walk line (legacy goNearPark). */
  async function goNearParkWalk(): Promise<{ success: boolean; msg?: string }> {
    uni.showToast({ title: t('common.loading'), icon: 'loading', mask: true })
    try {
      const loc = temp.location || (await locate())
      if (!loc) {
        return { success: false, msg: t('ride.needLocation') }
      }
      const res = await getNearFenceByPoiAndService({
        locationDTO: { lat: loc.latitude, lng: loc.longitude },
        serviceId: serviceId(),
      })
      if (!res.success || !res.data) {
        return { success: false, msg: res.msg || t('ride.parkEmpty') }
      }
      const data = res.data as {
        parkings?: Array<{
          centerLat?: number
          centerLng?: number
          name?: string
          id?: string | number
          pointList?: unknown
        }>
        noParkings?: Array<Record<string, unknown>>
        serviceAreas?: Array<Record<string, unknown>>
        pointList?: unknown
      }
      applyFencePayload(data)
      const first = (data.parkings || [])[0]
      const lat = Number(first?.centerLat)
      const lng = Number(first?.centerLng)
      if (!Number.isFinite(lat) || !Number.isFinite(lng) || !lat || !lng) {
        return { success: false, msg: t('ride.parkEmpty') }
      }
      const iconPath = getMapCfg('station') || getMapCfg('parking')
      parkMarkers.value = [
        {
          id: 9001,
          latitude: lat,
          longitude: lng,
          width: 28,
          height: 32,
          title: String(first?.name || ''),
          type: 0,
          ...(iconPath ? { iconPath } : {}),
        },
      ]
      return drawRoute({ latitude: lat, longitude: lng }, 1)
    } catch (e) {
      logger.warn('goNearParkWalk fail', e)
      return { success: false, msg: t('ride.parkEmpty') }
    } finally {
      uni.hideToast()
    }
  }

  function applyFencePayload(data: {
    parkings?: Array<Record<string, unknown>>
    noParkings?: Array<Record<string, unknown>>
    serviceAreas?: Array<Record<string, unknown>>
    pointList?: unknown
  }) {
    const out: NavPolygon[] = []
    const pushPoly = (raw: unknown, type: number, fullCar = false) => {
      const pts = toLatLngList(raw)
      if (pts.length < 3) return
      const colors = FENCE_COLORS[type] || FENCE_COLORS[0]
      out.push({
        points: pts,
        strokeWidth: type === 0 ? 3 : 1,
        strokeColor: colors[0],
        fillColor: colors[1],
        dashArray: type === 0 ? [5, 5] : [0, 0],
        ...(fullCar ? { fullCar: true } : {}),
      })
    }
    if (Array.isArray(data.serviceAreas) && data.serviceAreas.length) {
      for (const area of data.serviceAreas) pushPoly(area.pointList, 0)
    } else if (data.pointList) {
      pushPoly(data.pointList, 0)
    }
    for (const p of data.parkings || []) pushPoly(p.pointList, 1, Boolean(p.fullCar))
    for (const p of data.noParkings || []) pushPoly(p.pointList, 2)
    fencePolygons.value = out

    const stationIcon = getMapCfg('station') || getMapCfg('parking')
    const forbidIcon = getMapCfg('noParking') || getMapCfg('forbid')
    const markers: ParkMarker[] = []
    ;(data.parkings || []).forEach((p, i) => {
      const lat = Number(p.centerLat ?? p.lat ?? p.latitude)
      const lng = Number(p.centerLng ?? p.lng ?? p.longitude)
      if (!Number.isFinite(lat) || !Number.isFinite(lng)) return
      markers.push({
        id: 9100 + i,
        latitude: lat,
        longitude: lng,
        width: 28,
        height: 32,
        title: String(p.name || ''),
        type: 0,
        ...(p.fullCar ? { fullCar: true } : {}),
        ...(stationIcon ? { iconPath: stationIcon } : {}),
      })
    })
    ;(data.noParkings || []).forEach((p, i) => {
      const lat = Number(p.centerLat ?? p.lat ?? p.latitude)
      const lng = Number(p.centerLng ?? p.lng ?? p.longitude)
      if (!Number.isFinite(lat) || !Number.isFinite(lng)) return
      markers.push({
        id: 9200 + i,
        latitude: lat,
        longitude: lng,
        width: 28,
        height: 32,
        title: String(p.name || ''),
        type: 1,
        ...(forbidIcon ? { iconPath: forbidIcon } : {}),
      })
    })
    if (markers.length) parkMarkers.value = markers
  }

  async function loadRidingFences(): Promise<void> {
    const loc = temp.location || (await locate())
    const sid = serviceId()
    if (!loc) return
    try {
      if (sid && hasAccessToken()) {
        const byId = await getFenceByServiceId({ id: sid })
        if (byId.success && byId.data) {
          applyFencePayload(byId.data as Parameters<typeof applyFencePayload>[0])
          if (fencePolygons.value.length) return
        }
      }
    } catch (e) {
      logger.warn('getFenceByServiceId soft fail', e)
    }
    try {
      const near = await getNearFenceByPoiAndService({
        locationDTO: { lat: loc.latitude, lng: loc.longitude },
        serviceId: sid,
      })
      if (near.success && near.data) {
        applyFencePayload(near.data as Parameters<typeof applyFencePayload>[0])
      }
    } catch (e) {
      logger.warn('getNearFence soft fail', e)
    }
  }

  /** Consume temp.appNavigation → turnOnNavigation (legacy getRideLine). */
  async function applyAppNavigation(carId: string): Promise<{ success: boolean; msg?: string }> {
    const nav = temp.appNavigation
    if (!nav?.origin || !nav?.destination || !carId) {
      return { success: false }
    }
    temp.setAppNavigation(null)
    try {
      const res = await turnOnNavigation({
        carId,
        imei: '',
        pathType: 2,
        mapType: 'AMap',
        origin: nav.origin,
        destination: nav.destination,
      })
      if (!res.success || !res.data) {
        return { success: false, msg: res.msg || t('ride.navFail') }
      }
      const data = res.data as {
        polyline?: Array<{ lat?: number; lng?: number }>
        paths?: Array<{
          instruction?: string
          roadName?: string
          orientation?: string
          stepDistance?: string | number
        }>
        distance?: number
        duration?: number
      }
      const path0 = data.paths?.[0]
      if (path0?.instruction) {
        routeMeta.value = {
          distance: data.distance != null ? Number(data.distance) : undefined,
          durationMin: data.duration != null ? Math.round(Number(data.duration) / 60) : undefined,
          instruction: String(path0.instruction),
        }
      }
      const pts = toPolyPts(data.polyline || (nav.polyline as PolyPt[]) || [])
      if (pts.length) await drawInstrumentLine(pts)
      else if (nav.polyline) await drawInstrumentLine(toPolyPts(nav.polyline))
      return { success: true }
    } catch (e) {
      logger.warn('applyAppNavigation fail', e)
      return { success: false, msg: t('ride.navFail') }
    }
  }

  async function endInstrumentNav(carId: string): Promise<void> {
    temp.setAppNavigation(null)
    try {
      if (carId) await turnOffNavigation({ carId, imei: '' })
    } catch (e) {
      logger.warn('turnOffNavigation soft fail', e)
    }
    clearInstrument()
  }

  function saveNavForHandoff() {
    if (lastNavPayload.value?.origin && lastNavPayload.value?.destination) {
      temp.setAppNavigation(lastNavPayload.value)
    }
  }

  return {
    polyline,
    parkMarkers,
    fencePolygons,
    routeMeta,
    instrumentActive,
    lastNavPayload,
    clearRoute,
    clearInstrument,
    drawRoute,
    goNearParkWalk,
    loadRidingFences,
    applyAppNavigation,
    endInstrumentNav,
    saveNavForHandoff,
    drawInstrumentLine,
  }
}
