import { hexToFloat32, hexToString, numberToMacAddress } from './convert'

const BLE_PROTOCOL_MIN = 6

export function crcCheckValid(bleRsp: string): boolean {
  if (typeof bleRsp !== 'string') return false
  const len = bleRsp.length
  if (len < BLE_PROTOCOL_MIN || len % 2 !== 0) return false
  const crc = parseInt(bleRsp.substr(-2), 16)
  let sum = 0
  for (let i = 0; i < len - 1; i += 2) {
    sum += parseInt(bleRsp.substr(i, 2), 16)
  }
  return (sum & 0xff) === ((crc + crc) & 0xff)
}

export function resolveStatusResult(statusRsp: string) {
  if (typeof statusRsp !== 'string' || !crcCheckValid(statusRsp)) return {}
  const mode = parseInt(statusRsp.substr(4, 2), 16)
  const gsm = parseInt(statusRsp.substr(6, 2), 16)
  const sw = parseInt(statusRsp.substr(8, 2), 16)
  const voltageMv = parseInt(statusRsp.substr(10, 8), 16)
  return {
    mode,
    gsm,
    acc: (sw & 0b10) > 1 ? 1 : 0,
    defend: sw & 0b01,
    backWheel: (sw & 0b0100) > 2 ? 1 : 0,
    backSeat: (sw & 0b1000) > 3 ? 1 : 0,
    voltageMv,
  }
}

export function resolveGpsInfo(data: string) {
  const rsp: Record<string, number> = {}
  if (!data || !crcCheckValid(data)) return rsp
  rsp.timestamp = parseInt(data.substr(4, 8), 16)
  rsp.longitude = hexToFloat32(data.substr(12, 8), false)
  rsp.latitude = hexToFloat32(data.substr(20, 8), false)
  rsp.speed = parseInt(data.substr(28, 2), 16)
  rsp.course = parseInt(data.substr(30, 4), 16)
  return rsp
}

export function resolveDeviceStatusInfo(data: string) {
  const rsp: Record<string, number> = {}
  if (!data || !crcCheckValid(data)) return rsp
  const sw = parseInt(data.substr(6, 8), 16)
  rsp.gsm = parseInt(data.substr(14, 2), 16)
  rsp.voltage = parseInt(data.substr(16, 4), 16)
  rsp.GPSMajorVsn = parseInt(data.substr(20, 2), 16)
  rsp.GPSManorVsn = parseInt(data.substr(22, 2), 16)
  rsp.GPSMicroVsn = parseInt(data.substr(24, 2), 16)
  rsp.BLEMajorVsn = parseInt(data.substr(26, 2), 16)
  rsp.BLEManorVsn = parseInt(data.substr(28, 2), 16)
  rsp.BLEMicroVsn = parseInt(data.substr(30, 2), 16)
  rsp.timestamp = parseInt(data.substr(32, 8), 16)
  rsp.longitude = parseInt(data.substr(40, 8), 16) / 1000000
  rsp.latitude = parseInt(data.substr(48, 8), 16) / 1000000
  rsp.speed = parseInt(data.substr(56, 2), 16)
  rsp.course = parseInt(data.substr(58, 4), 16)
  rsp.hdop = parseInt(data.substr(62, 4), 16)
  rsp.satellite = parseInt(data.substr(66, 2), 16)
  rsp.totalMiles = parseInt(data.substr(68, 8), 16)
  rsp.isDefendOn = (sw & 0b01) > 0 ? 1 : 0
  rsp.isAccOn = (sw & 0b10) > 0 ? 1 : 0
  rsp.isWheelLocked = (sw & 0b100) > 1 ? 1 : 0
  rsp.isSeatLocked = (sw & 0b1000) > 1 ? 1 : 0
  rsp.isPowerExist = (sw & 0b10000) > 1 ? 1 : 0
  rsp.isMoving = (sw & 0b1000000) > 1 ? 1 : 0
  rsp.isHelmetExist = (sw & 0b10000000000000000) > 1 ? 1 : 0
  return rsp
}

export function resolveLastBeaconInfo(data: string) {
  const rsp: Record<string, unknown> = {}
  if (!data || !crcCheckValid(data)) return rsp
  rsp.event = parseInt(data.substring(4, 6), 16)
  rsp.tBeaconAddr = numberToMacAddress(data.substring(6, 18))
  rsp.tBeaconId = hexToString(data.substring(18, 42))
  rsp.tBeaconSOC = parseInt(data.substring(42, 44), 16)
  rsp.tBeaconVsn = hexToString(data.substring(44, 48))
  rsp.lat = parseInt(data.substring(48, 56), 16) / 1000000
  rsp.lon = parseInt(data.substring(56, 64), 16) / 1000000
  rsp.timestamp = parseInt(data.substring(64, 72), 16)
  return rsp
}

export function resolveRFIDInfo(data: string) {
  const rsp: Record<string, unknown> = {}
  if (!data || !crcCheckValid(data)) return rsp
  rsp.result = parseInt(data.substring(4, 6), 16)
  rsp.version = parseInt(data.substring(6, 54), 16)
  rsp.cardID = hexToString(data.substring(54, 86))
  return rsp
}

export function parseNotify(hexOrBuffer: string | ArrayBuffer): Record<string, unknown> {
  let hex = ''
  if (typeof hexOrBuffer === 'string') {
    hex = hexOrBuffer
  } else {
    const view = new Uint8Array(hexOrBuffer)
    hex = Array.from(view, (b) => ('0' + b.toString(16)).slice(-2)).join('')
  }
  if (!hex) return {}
  const cmd = parseInt(hex.substring(0, 2), 16)
  if (cmd === 0x41) return { cmd, ...resolveDeviceStatusInfo(hex) }
  if (cmd === 0x32) return { cmd, ...resolveGpsInfo(hex) }
  if (cmd === 0x42) return { cmd, ...resolveLastBeaconInfo(hex) }
  if (cmd === 0x54) return { cmd, ...resolveRFIDInfo(hex) }
  return { cmd, raw: hex }
}
