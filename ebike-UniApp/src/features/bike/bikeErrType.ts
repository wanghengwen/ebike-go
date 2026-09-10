/** Legacy precycling ebikeErrType map. */
export const BIKE_ERR_TYPE: Record<number, string> = {
  0: 'ride.bikeErr0',
  1: 'ride.bikeErr1',
  2: 'ride.bikeErr2',
  11: 'ride.bikeErr11',
  12: 'ride.bikeErr12',
  14: 'ride.bikeErr14',
  15: 'ride.bikeErr15',
  16: 'ride.bikeErr16',
  41: 'ride.bikeErr41',
  42: 'ride.bikeErr42',
  51: 'ride.bikeErr51',
}

export function bikeErrMessageKey(errType: unknown): string | null {
  if (errType == null || errType === '') return null
  const n = Number(errType)
  if (!Number.isFinite(n)) return null
  return BIKE_ERR_TYPE[n] || 'ride.bikeErrDefault'
}

/** Legacy: ridingType === 1 means usable; otherwise show errType copy. */
export function isBikeUnavailable(car?: Record<string, unknown> | null): boolean {
  if (!car) return false
  const ridingType = car.ridingType
  if (ridingType != null && ridingType !== '') {
    return Number(ridingType) !== 1
  }
  return bikeErrMessageKey(car.errType) != null
}
