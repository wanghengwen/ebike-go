/** Closest polyline vertex index to a point — legacy utils/polyline-point.js */
export function closestPoint(
  polyline: Array<{ lat?: number; lng?: number; latitude?: number; longitude?: number }>,
  point: { lat: number; lng: number },
): number {
  if (!polyline.length) return 0
  const dists = polyline.map((i) => {
    const lng = Number(i.lng ?? i.longitude)
    const lat = Number(i.lat ?? i.latitude)
    return Math.abs(lng - point.lng) + Math.abs(lat - point.lat)
  })
  const sorted = [...dists].sort((a, b) => Math.abs(a) - Math.abs(b) || b - a)
  const idx = dists.indexOf(sorted[0])
  return idx >= 0 ? idx : 0
}
