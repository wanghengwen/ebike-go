import { toHexString, toPad2HexStr, tokenToDataArray } from './convert'
import type { BleCmdCode } from './commands'

const DEFAULT_TOKEN = [0x0a, 0x0a, 0x05, 0x05]

export type BleFrameResult = {
  cmd: number
  code: number
  data: string
}

export type TokenProvider = (imei: string) => Promise<string | number | null | undefined>

/**
 * Build BLE command frame: cmd + len + token[+payload] + crc
 */
export async function buildCmd(
  imei: string,
  cmd: BleCmdCode | number,
  data: number[] = [],
  getToken?: TokenProvider,
): Promise<string> {
  let dataArray = [...DEFAULT_TOKEN]
  if (getToken) {
    try {
      const token = await getToken(imei)
      if (token != null && token !== '') {
        dataArray = tokenToDataArray(token)
      }
    } catch {
      dataArray = [...DEFAULT_TOKEN]
    }
  }
  for (let i = 0; i < data.length; i++) dataArray.push(data[i])
  const datalen = dataArray.length
  let datasum = 0
  dataArray.forEach((el) => {
    datasum += el
  })
  const crc = toPad2HexStr(cmd + datasum + datalen)
  return toPad2HexStr(cmd) + toPad2HexStr(datalen) + toHexString(dataArray) + crc
}

export function resolveResult(result: string): BleFrameResult {
  const cmd = parseInt(result.substring(0, 2), 16)
  const length = parseInt(result.substring(2, 4), 16)
  const retCode = parseInt(result.substring(4, 4 + length * 2), 16)
  const crc = parseInt(result.substring(4 + length * 2, 6 + length * 2), 16)
  if (crc === cmd + length + retCode) {
    return { cmd, code: retCode, data: result }
  }
  return { cmd, code: 0, data: result }
}

export function hexFrameToArrayBuffer(hex: string): ArrayBuffer {
  const bytes = new Uint8Array(hex.length / 2)
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(hex.substr(i * 2, 2), 16)
  }
  return bytes.buffer
}
