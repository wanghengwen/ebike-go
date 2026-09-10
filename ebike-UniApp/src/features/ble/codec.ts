export { BLE_CMD, BLE_VOICE } from './commands'
export { buildCmd, resolveResult, hexFrameToArrayBuffer } from './buildCmd'
export {
  parseNotify,
  resolveDeviceStatusInfo,
  resolveGpsInfo,
  resolveLastBeaconInfo,
  resolveRFIDInfo,
  resolveStatusResult,
  crcCheckValid,
} from './resolve'
export { toHexString, hexToBytes, bytesToHex, tokenToDataArray } from './convert'

import { buildCmd, hexFrameToArrayBuffer } from './buildCmd'
import { BLE_CMD } from './commands'
import type { TokenProvider } from './buildCmd'

export type BleCommand = {
  cmd: number
  hex: string
  buffer: ArrayBuffer
}

export async function buildUnlockCommand(imei: string, getToken?: TokenProvider): Promise<BleCommand> {
  const hex = await buildCmd(imei, BLE_CMD.UNLOCK, [], getToken)
  return { cmd: BLE_CMD.UNLOCK, hex, buffer: hexFrameToArrayBuffer(hex) }
}

export async function buildLockCommand(imei: string, getToken?: TokenProvider): Promise<BleCommand> {
  const hex = await buildCmd(imei, BLE_CMD.LOCK, [], getToken)
  return { cmd: BLE_CMD.LOCK, hex, buffer: hexFrameToArrayBuffer(hex) }
}

export async function buildPlayVoiceCommand(
  imei: string,
  voiceIdx: number,
  volume = 100,
  getToken?: TokenProvider,
): Promise<BleCommand> {
  const hex = await buildCmd(imei, BLE_CMD.PLAY_VOICE, [voiceIdx, volume], getToken)
  return { cmd: BLE_CMD.PLAY_VOICE, hex, buffer: hexFrameToArrayBuffer(hex) }
}

export async function buildGetDeviceInfoCommand(imei: string, getToken?: TokenProvider): Promise<BleCommand> {
  const hex = await buildCmd(imei, BLE_CMD.GET_DEVICE_INFO, [], getToken)
  return { cmd: BLE_CMD.GET_DEVICE_INFO, hex, buffer: hexFrameToArrayBuffer(hex) }
}
