/** BLE hex / float helpers (migrated from legacy convert utils). */

export function toPad2HexStr(num: number): string {
  return ('0' + (num & 0xff).toString(16)).slice(-2)
}

export function toHexString(byteArray: number[]): string {
  return Array.from(byteArray, (byte) => toPad2HexStr(byte)).join('')
}

export function hexToBytes(hex: string): number[] {
  const bytes: number[] = []
  for (let c = 0; c < hex.length; c += 2) {
    bytes.push(parseInt(hex.substr(c, 2), 16))
  }
  return bytes
}

export function bytesToHex(bytes: number[]): string {
  const hex: string[] = []
  for (let i = 0; i < bytes.length; i++) {
    const current = bytes[i] < 0 ? bytes[i] + 256 : bytes[i]
    hex.push((current >>> 4).toString(16))
    hex.push((current & 0xf).toString(16))
  }
  return hex.join('')
}

export function tokenToDataArray(token: string | number): number[] {
  const str = parseInt(String(token), 10).toString(16)
  const zeroes = '00000000'
  const hexStr = zeroes.substr(0, zeroes.length - str.length) + str
  const array: number[] = []
  for (let i = 0; i < hexStr.length; i += 2) {
    array.push(parseInt(hexStr.substring(i, i + 2), 16))
  }
  return array
}

function swap32(val: number): number {
  return (
    ((val & 0xff) << 24) |
    ((val & 0xff00) << 8) |
    ((val >> 8) & 0xff00) |
    ((val >> 24) & 0xff)
  )
}

export function hexToFloat32(str: string, bigEndian = true): number {
  let int = parseInt(str, 16)
  if (!bigEndian) int = swap32(int)
  if (int === 0) return 0
  const sign = int >>> 31 ? -1 : 1
  let exp = ((int >>> 23) & 0xff) - 127
  const mantissa = ((int & 0x7fffff) + 0x800000).toString(2)
  let float32 = 0
  for (let i = 0; i < mantissa.length; i += 1) {
    float32 += parseInt(mantissa[i], 10) ? Math.pow(2, exp) : 0
    exp--
  }
  return float32 * sign
}

export function hexToString(hex: string): string {
  let out = ''
  for (let i = 0; i < hex.length; i += 2) {
    out += String.fromCharCode(parseInt(hex.substr(i, 2), 16))
  }
  return out
}

export function numberToMacAddress(number: string): string {
  const pairs: string[] = []
  for (let i = 0; i < number.length; i += 2) {
    pairs.push(number.substr(i, 2))
  }
  return pairs.reverse().join(':')
}
