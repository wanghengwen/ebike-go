import { defineStore } from 'pinia'

export type RideSession = {
  orderId?: string
  imei?: string
  carId?: string
  startTime?: number
  status?: 'idle' | 'precycling' | 'riding' | 'tempPark' | 'ended'
}

/** Destination nav payload from park-search → riding turnOnNavigation */
export type AppNavigation = {
  origin: string
  destination: string
  type?: number
  pathType?: number
  polyline?: Array<{ lat?: number; lng?: number; latitude?: number; longitude?: number }>
  distance?: number
  duration?: number
  [key: string]: unknown
}

export const useTempDataStore = defineStore('tempData', {
  state: () => ({
    location: null as { latitude: number; longitude: number } | null,
    nearBikes: [] as Array<Record<string, unknown>>,
    ride: { status: 'idle' } as RideSession,
    unpaidOrderId: '' as string,
    /** Face verify result '1' | '0' | '' — anti back-stack retake */
    faceResult: '' as string,
    /** Saved before face gate so result page can resume precycling */
    preCyclingCar: null as { carId?: string; imei?: string } | null,
    /** Cached returnPermission payload for sheets */
    lastReturnPermission: null as Record<string, unknown> | null,
    /** park-search → riding instrument navigation handoff */
    appNavigation: null as AppNavigation | null,
  }),
  actions: {
    setLocation(lat: number, lng: number) {
      this.location = { latitude: lat, longitude: lng }
    },
    setNearBikes(list: Array<Record<string, unknown>>) {
      this.nearBikes = list
    },
    setRide(patch: Partial<RideSession>) {
      const next = { ...this.ride }
      for (const [k, v] of Object.entries(patch) as Array<[keyof RideSession, RideSession[keyof RideSession]]>) {
        // Keep last known carId/imei/orderId when poll omits them
        if ((k === 'carId' || k === 'imei' || k === 'orderId') && (v == null || v === '')) continue
        ;(next as Record<string, unknown>)[k] = v
      }
      this.ride = next
    },
    resetRide() {
      this.ride = { status: 'idle' }
    },
    setFaceResult(v: string) {
      this.faceResult = v
    },
    setPreCyclingCar(car: { carId?: string; imei?: string } | null) {
      this.preCyclingCar = car
    },
    setLastReturnPermission(data: Record<string, unknown> | null) {
      this.lastReturnPermission = data
    },
    setAppNavigation(data: AppNavigation | null) {
      this.appNavigation = data
    },
  },
})
