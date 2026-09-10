import { getBikeBatteryIcon, getMapCfg } from '@/shared/tenantSkin'
import { getRedEnvelopeCfg, isRedEnvelopeBike, isTimeExpired } from '@/features/bike/redEnvelope'

export type LatLng = { latitude: number; longitude: number }

export type MapPolygon = {
  points: LatLng[]
  strokeWidth: number
  strokeColor: string
  fillColor: string
  dashArray?: number[]
  /** Parking full — hide on riding map (legacy polygonsWithFilter) */
  fullCar?: boolean
}

export type MapMarker = {
  id: number
  latitude: number
  longitude: number
  width?: number
  height?: number
  title?: string
  iconPath?: string
  /** 0 parking / 1 no-parking / 2 bike */
  type?: number
  carId?: string
  imei?: string
  /** Parking full — hide on riding map (legacy markersWithFilter) */
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
  sourceData?: Record<string, unknown>
}

export type MapPolyline = {
  points: LatLng[]
  color?: string
  width?: number
  arrowLine?: boolean
  dottedLine?: boolean
}

const FENCE_COLORS: Array<[string, string]> = [
  ['#3AA0E8', '#3AA0E833'], // service
  ['#63D144', '#63D14433'], // parking
  ['#FF5936', '#FF593633'], // no-parking
]

export function toPoints(pointList: unknown): LatLng[] {
  if (!Array.isArray(pointList)) return []
  const points: LatLng[] = []
  for (const item of pointList) {
    if (Array.isArray(item) && item.length >= 2) {
      points.push({ longitude: Number(item[0]), latitude: Number(item[1]) })
    } else if (item && typeof item === 'object') {
      const o = item as { lat?: number; lng?: number; latitude?: number; longitude?: number }
      const latitude = Number(o.latitude ?? o.lat)
      const longitude = Number(o.longitude ?? o.lng)
      if (!Number.isNaN(latitude) && !Number.isNaN(longitude)) {
        points.push({ latitude, longitude })
      }
    }
  }
  return points
}

export function fenceToPolygon(
  points: LatLng[],
  type: number,
  opts: { fullCar?: boolean } = {},
): MapPolygon | null {
  if (!points.length) return null
  const colors = FENCE_COLORS[type] || FENCE_COLORS[0]
  return {
    points,
    strokeWidth: type === 0 ? 3 : 1,
    strokeColor: colors[0],
    fillColor: colors[1],
    dashArray: type === 0 ? [5, 5] : [0, 0],
    ...(opts.fullCar ? { fullCar: true } : {}),
  }
}

export function markerIdFrom(raw: unknown, fallback: number): number {
  const s = String(raw ?? '')
  if (!s) return fallback
  const n = Number(s.length > 9 ? s.slice(-9) : s)
  return Number.isFinite(n) ? n : fallback
}

export function buildFenceMarkers(
  parkings: Array<Record<string, unknown>> = [],
  noParkings: Array<Record<string, unknown>> = [],
  opts: { redEnvelopeMode?: boolean; onlyShowRedEnvelope?: boolean } = {},
): MapMarker[] {
  const out: MapMarker[] = []
  const redParkIcon =
    getRedEnvelopeCfg('redEnvelopeStation') ||
    getRedEnvelopeCfg('redEnvelopeParking') ||
    getRedEnvelopeCfg('redPaketPark')
  const pushFence = (item: Record<string, unknown>, type: 0 | 1, index: number) => {
    const lat = Number(item.centerLat ?? item.lat ?? item.latitude)
    const lng = Number(item.centerLng ?? item.lng ?? item.longitude)
    if (Number.isNaN(lat) || Number.isNaN(lng)) return
    // Legacy: red-envelope map may hide expired red stations
    if (
      type === 0 &&
      opts.redEnvelopeMode &&
      opts.onlyShowRedEnvelope &&
      isTimeExpired(item.expirationTime)
    ) {
      return
    }
    const isRedStation =
      type === 0 && opts.redEnvelopeMode && !isTimeExpired(item.expirationTime) && Boolean(redParkIcon)
    const iconPath =
      type === 0
        ? isRedStation
          ? redParkIcon
          : getMapCfg('station') || getMapCfg('parking')
        : getMapCfg('noParking') || getMapCfg('forbid')
    out.push({
      id: markerIdFrom(item.id ?? `f${type}${index}`, 900000 + type * 10000 + index),
      latitude: lat,
      longitude: lng,
      width: isRedStation ? 26 : 28,
      height: isRedStation ? 40 : 32,
      title: String(item.name || item.id || ''),
      type,
      ...(type === 0 && item.fullCar ? { fullCar: true } : {}),
      ...(iconPath ? { iconPath } : {}),
      sourceData: item,
    })
  }
  parkings.forEach((p, i) => pushFence(p, 0, i))
  noParkings.forEach((p, i) => pushFence(p, 1, i))
  return out
}

export function buildFencePolygons(
  payload: {
    serviceAreas?: Array<Record<string, unknown>>
    pointList?: unknown
    parkings?: Array<Record<string, unknown>>
    noParkings?: Array<Record<string, unknown>>
  },
  opts: { redEnvelopeMode?: boolean; onlyShowRedEnvelope?: boolean } = {},
): { polygons: MapPolygon[]; markers: MapMarker[] } {
  const out: MapPolygon[] = []
  const serviceAreas = payload.serviceAreas
  if (Array.isArray(serviceAreas) && serviceAreas.length) {
    for (const area of serviceAreas) {
      const poly = fenceToPolygon(toPoints(area.pointList), 0)
      if (poly) out.push(poly)
    }
  } else if (payload.pointList) {
    const poly = fenceToPolygon(toPoints(payload.pointList), 0)
    if (poly) out.push(poly)
  }
  for (const p of payload.parkings || []) {
    if (
      opts.redEnvelopeMode &&
      opts.onlyShowRedEnvelope &&
      isTimeExpired(p.expirationTime)
    ) {
      continue
    }
    const poly = fenceToPolygon(toPoints(p.pointList), 1, { fullCar: Boolean(p.fullCar) })
    if (poly) out.push(poly)
  }
  for (const p of payload.noParkings || []) {
    const poly = fenceToPolygon(toPoints(p.pointList), 2)
    if (poly) out.push(poly)
  }
  return {
    polygons: out,
    markers: buildFenceMarkers(payload.parkings || [], payload.noParkings || [], opts),
  }
}

export function buildBikeMarkers(
  list: Array<Record<string, unknown>>,
  opts: { redEnvelopeOnly?: boolean; excludeCarId?: string } = {},
): { markers: MapMarker[]; redCount: number } {
  let redCount = 0
  const markers = list
    .map((item, index) => {
      const lat = Number(item.lat ?? item.latitude)
      const lng = Number(item.lng ?? item.longitude)
      if (Number.isNaN(lat) || Number.isNaN(lng)) return null
      const carId = String(item.carId ?? item.id ?? '')
      if (opts.excludeCarId && carId && carId === opts.excludeCarId) return null
      const isRed = isRedEnvelopeBike(item)
      if (isRed) redCount += 1
      if (opts.redEnvelopeOnly && !isRed) return null
      const redIcon = getRedEnvelopeCfg('iconEbikeRedEnvelope')
      const iconPath = isRed && redIcon ? redIcon : getBikeBatteryIcon(item.restBattery ?? item.soc ?? item.battery)
      return {
        id: markerIdFrom(carId || index + 1, index + 1),
        latitude: lat,
        longitude: lng,
        width: 28,
        height: 36,
        title: carId,
        type: 2,
        carId,
        imei: item.imei != null ? String(item.imei) : undefined,
        ...(iconPath ? { iconPath } : {}),
        sourceData: item,
      } as MapMarker
    })
    .filter(Boolean) as MapMarker[]
  return { markers, redCount }
}

/** Single bike pin for precycling / detail pages. */
export function buildCarMarker(
  info: Record<string, unknown>,
  opts: { isRedEnvelope?: boolean; fallbackId?: number } = {},
): MapMarker | null {
  const lat = Number(info.lat ?? info.latitude)
  const lng = Number(info.lng ?? info.longitude)
  if (Number.isNaN(lat) || Number.isNaN(lng)) return null
  const carId = String(info.carId ?? info.id ?? '')
  const isRed = opts.isRedEnvelope ?? isRedEnvelopeBike(info)
  const redIcon = getRedEnvelopeCfg('iconEbikeRedEnvelope')
  const iconPath = isRed && redIcon ? redIcon : getBikeBatteryIcon(info.restBattery ?? info.soc ?? info.battery)
  return {
    id: markerIdFrom(carId || opts.fallbackId || 1, opts.fallbackId || 1),
    latitude: lat,
    longitude: lng,
    width: isRed ? 28 : 28,
    height: isRed ? 39 : 36,
    title: carId,
    type: 2,
    carId,
    imei: info.imei != null ? String(info.imei) : undefined,
    ...(iconPath ? { iconPath } : {}),
    sourceData: info,
  }
}
