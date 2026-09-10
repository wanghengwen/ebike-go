import { getTenantConfig } from '@/shared/config'

/** Decode URL-safe / standard Base64 path segment to plain text. */
function decodeBase64Segment(input: string): string {
  try {
    let s = String(input || '').replace(/[^A-Za-z0-9+/=_-]/g, '')
    s = s.replace(/-/g, '+').replace(/_/g, '/')
    while (s.length % 4) s += '='
    // Prefer atob in H5; fall back to uni / Buffer-free manual decode via decodeURIComponent
    if (typeof atob === 'function') {
      const bin = atob(s)
      try {
        return decodeURIComponent(
          Array.from(bin)
            .map((c) => '%' + c.charCodeAt(0).toString(16).padStart(2, '0'))
            .join(''),
        )
      } catch {
        return bin
      }
    }
    // mini program: use uni.base64ToArrayBuffer if available
    if (typeof uni !== 'undefined' && typeof uni.base64ToArrayBuffer === 'function') {
      const buf = uni.base64ToArrayBuffer(s)
      const bytes = new Uint8Array(buf)
      let out = ''
      for (let i = 0; i < bytes.length; i++) out += String.fromCharCode(bytes[i])
      return out
    }
    return ''
  } catch {
    return ''
  }
}

/**
 * Legacy doQrCode — returns [placeholder, domain, placeholder, carId].
 * Tries: query param digits → Base64 last path segment → domain + 9-digit id.
 */
export function parseQrCode(qrcode: string): { domain: string; carId: string } {
  const _qrcode = String(qrcode || '').trim()
  if (!_qrcode) return { domain: '', carId: '' }

  let decoded = _qrcode
  try {
    decoded = decodeURIComponent(_qrcode)
  } catch {
    decoded = _qrcode
  }

  const primary = /^https?:\/\/(.*?)(\/?)(\?|#|\/)(.*?)=(\d+)?/i.exec(decoded) || []
  let domain = primary[1] || ''
  let carId = primary[5] || ''

  if (!carId) {
    try {
      const seg = decoded.split('/').pop() || ''
      const ebikeId = decodeBase64Segment(seg)
      if (ebikeId && /^\d+$/.test(ebikeId)) carId = ebikeId
    } catch {
      /* ignore */
    }
  }

  if (!carId) {
    const secondary = /^https?:\/\/(.*?)(\?|#|\/)(.*)/i.exec(decoded) || []
    domain = secondary[1] || domain
    const other = secondary[3] || ''
    const nine = /\d{9}/.exec(other)
    if (nine?.[0]) carId = nine[0]
  }

  if (!carId) {
    // Plain car id / non-URL
    if (!/^https?:\/\//i.test(decoded)) {
      const m = /[?&](?:carId|id)=(\w+)/i.exec(decoded)
      if (m?.[1]) carId = m[1]
      else if (/^\d{5,}$/.test(decoded)) carId = decoded
    }
  }

  if (!carId) domain = ''
  return { domain, carId: String(carId || '') }
}

export function extractCarId(raw: string): string {
  const { carId } = parseQrCode(raw)
  if (carId) return carId
  // Last resort for searchParams-style URLs already handled; keep empty if unrecognized URL
  const text = String(raw || '').trim()
  if (!text || /^https?:\/\//i.test(text)) return ''
  return text
}

/** Tenant qrDomain allow-list; empty config = allow all. Non-URL always allowed. */
export function checkQrDomain(raw: string): boolean {
  try {
    const u = new URL(raw)
    const domains = getTenantConfig().qrDomain
    if (!domains || (Array.isArray(domains) && !domains.length)) return true
    const host = u.host
    const list = Array.isArray(domains) ? domains : [domains]
    return list.some((d) => {
      const m = /(https?:\/\/)?(.*)/i.exec(String(d)) || []
      const body = m[2] || String(d)
      return String(d).includes(host) || host.includes(body) || raw.includes(body)
    })
  } catch {
    return true
  }
}
