import { getAddressByLatAndLng } from '@/api/common'

/** fen → yuan (same as formatMoney filter; prefers 2 decimals when digits given). */
export function formatMoney(val: unknown, digits?: number): string {
  if (val === null || val === undefined || val === '') return '0'
  const n = Number(val)
  if (Number.isNaN(n)) return '0'
  if (digits != null) return (n / 100).toFixed(digits)
  return String(n / 100)
}

/** meters → km (legacy formatMile). */
export function formatMile(val: unknown, digits?: number): string {
  if (val === null || val === undefined || val === '') return '0'
  const n = Number(val)
  if (Number.isNaN(n)) return '0'
  if (digits != null) return (n / 1000).toFixed(digits)
  const km = n / 1000
  return Number.isInteger(km) ? String(km) : km.toFixed(2)
}

/** ms → HH:mm:ss (legacy formatSecond). */
export function formatRidingTime(val: unknown): string {
  if (val === null || val === undefined || val === '') return '--'
  const ms = Number(val)
  if (!Number.isFinite(ms)) return '--'
  const hours = Math.floor(ms / (1000 * 60 * 60))
  const minutes = Math.floor((ms % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.round((ms % (1000 * 60)) / 1000)
  const pad = (n: number) => (n < 10 ? `0${n}` : String(n))
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`
}

/** ms → 中文单位时长（legacy formatSecondUnit），如「6秒」「1小时2分钟3秒」。 */
export function formatRidingTimeUnit(val: unknown): string {
  if (val === null || val === undefined || val === '') return '1秒'
  const ms = Number(val)
  if (!Number.isFinite(ms)) return '1秒'
  const hours = Math.floor(ms / (1000 * 60 * 60))
  const minutes = Math.floor((ms % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.round((ms % (1000 * 60)) / 1000)
  const hoursText = hours ? `${hours}小时` : ''
  const minutesText = minutes ? `${minutes}分钟` : ''
  const secondsText = seconds ? `${seconds}秒` : '0秒'
  return hoursText + minutesText + secondsText
}

/**
 * Prepare HTML for uni-app rich-text (legacy formatRichText).
 * Fixes img/table sizing and migrated font size/color attributes.
 */
export function formatRichText(richText: unknown): string {
  if (!richText) return ''
  let html = String(richText)
  html = html
    .replace(/<img([\s\w"-=\/\.:;]+)/gi, '<img style="max-width: 100% !important;height: auto;"$1')
    .replace(/<table([\s\w"-=\/\.:;]+)/gi, '<table style="border-collapse: collapse;"$1')
    .replaceAll('<th>', '<th style="border: 1px solid #cccccc;">')
    .replaceAll('<td>', '<td style="border: 1px solid #cccccc;">')
    .replaceAll('<spanyes";', '<span style="')
    .replace(/<\/spanyes([^>]*)>/g, '</span>')
    .replaceAll('=""', '')
    .replace(/color="(.*?)"/gi, 'style="color:$1"')
    .replace(/size="(.*?)"/gi, 'class="quesDetail_font_size_$1"')
  return html
}

/** Reverse geocode; returns '--' on failure (legacy getAddressByLatLng). */
export async function resolveAddress(lat: unknown, lng: unknown): Promise<string> {
  const latitude = Number(lat)
  const longitude = Number(lng)
  if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) return '--'
  try {
    const res = await getAddressByLatAndLng({ latitude, longitude })
    if (res.success && res.data) {
      const data = res.data as { address?: string; formattedAddress?: string; name?: string } | string
      if (typeof data === 'string') return data || '--'
      return String(data.address || data.formattedAddress || data.name || '--')
    }
  } catch {
    /* soft fail */
  }
  return '--'
}
