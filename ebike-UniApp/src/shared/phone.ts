/**
 * Legacy Vue filter `phoneDesensitize` (wechat/src/directives/common.js).
 * Phone is often stored as `+86-13800138000` — strip country prefix before masking.
 */
export function phoneDesensitize(value?: string | null): string {
  if (!value) return '--'
  let v = String(value)
  if (v.split('-').length > 1) {
    v = v.split('-')[1]
  }
  return v ? `${v.slice(0, 3)}****${v.slice(7)}` : '--'
}

/** Strip `+86-` / country segment for display or API that expects local number. */
export function stripPhoneCountry(value?: string | null): string {
  if (!value) return ''
  const parts = String(value).split('-')
  return parts.length > 1 ? parts[1] : String(value).replace(/^\+?86/, '')
}
