import { logger } from '@/shared/logger'
import {
  BLE_CONNECT_TIMEOUT,
  BLE_RX_UUID,
  BLE_SEARCH_TIMEOUT,
  BLE_SEND_TIMEOUT,
  BLE_SERVICE_FILTER,
  BLE_SERVICE_UUID,
  BLE_TX_UUID,
} from './config'
import type { BleCommand } from './codec'
import { buildCmd, hexFrameToArrayBuffer, parseNotify, resolveResult } from './codec'
import { BLE_CMD, BLE_VOICE } from './commands'
import { getBlueToothToken } from '@/api/ble'
import { tokenToDataArray } from './convert'

/** Mute payload for unlock/lock when voice is deferred (legacy openBike isMute). */
const MUTE_PAYLOAD = [0, 0]

export type BleDevice = {
  deviceId: string
  name?: string
  advertisData?: ArrayBuffer
  advertisDataFormat?: string
  [key: string]: unknown
}

type HistoryItem = { imei: string; device: BleDevice }

type BleLink = {
  deviceId: string
  serviceId: string
  writeId: string
  notifyId: string
}

function ab2hex(buffer: ArrayBuffer): string {
  return Array.prototype.map
    .call(new Uint8Array(buffer), (bit: number) => ('00' + bit.toString(16)).slice(-2))
    .join('')
}

/** Match scanned advertisData hex against IMEI mid segment (legacy rule). */
export function matchImeiAdvert(imei: string, advertisDataFormat?: string): boolean {
  return Boolean(imei && advertisDataFormat && imei.indexOf(advertisDataFormat) === 3 && advertisDataFormat.length === 12)
}

async function fetchTokenBytes(imei: string): Promise<number[]> {
  try {
    const res = await getBlueToothToken({ imei })
    if (res.success && res.data && (res.data as { token?: string | number }).token != null) {
      return tokenToDataArray((res.data as { token: string | number }).token)
    }
  } catch (err) {
    logger.warn('ble token fail', err)
  }
  return [0x0a, 0x0a, 0x05, 0x05]
}

class BleSession {
  connected = false
  devices: BleDevice[] = []
  history: HistoryItem[] = []
  private searchTimer: ReturnType<typeof setTimeout> | null = null
  private currentLink: BleLink | null = null
  private notifyHandler: ((hex: string, decoded: Record<string, unknown>) => void) | null = null

  async openAdapter(): Promise<boolean> {
    if (this.connected) return true
    await this.closeAdapter()
    return new Promise((resolve) => {
      uni.openBluetoothAdapter({
        success: () => resolve(true),
        fail: (err) => {
          logger.warn('openBluetoothAdapter fail', err)
          resolve(false)
        },
      })
    })
  }

  closeAdapter(): Promise<boolean> {
    return new Promise((resolve) => {
      uni.closeBluetoothAdapter({
        complete: () => {
          this.connected = false
          this.currentLink = null
          resolve(true)
        },
      })
    })
  }

  stopSearch() {
    if (this.searchTimer) {
      clearTimeout(this.searchTimer)
      this.searchTimer = null
    }
    uni.stopBluetoothDevicesDiscovery({})
  }

  startDiscovery(): Promise<boolean> {
    return new Promise((resolve) => {
      uni.startBluetoothDevicesDiscovery({
        services: BLE_SERVICE_FILTER,
        allowDuplicatesKey: false,
        interval: 0,
        success: () => resolve(true),
        fail: (err) => {
          logger.warn('startDiscovery fail', err)
          resolve(false)
        },
      })
    })
  }

  /**
   * Search nearby devices until IMEI matches or timeout.
   */
  findDeviceByImei(imei: string, timeout = BLE_SEARCH_TIMEOUT): Promise<BleDevice | null> {
    this.devices = []
    return new Promise(async (resolve) => {
      let settled = false
      const finish = (device: BleDevice | null) => {
        if (settled) return
        settled = true
        this.stopSearch()
        // @ts-expect-error uni off API variance
        uni.offBluetoothDeviceFound?.()
        resolve(device)
      }

      const fromHistory = this.history.find((h) => h.imei === imei)
      if (fromHistory) {
        const ok = await this.connectDevice(fromHistory.device.deviceId)
        if (ok) {
          finish(fromHistory.device)
          return
        }
      }

      const started = await this.startDiscovery()
      if (!started) {
        finish(null)
        return
      }

      uni.onBluetoothDeviceFound((res) => {
        const list = (res.devices || []) as BleDevice[]
        list.forEach((item) => {
          if (item.advertisData) item.advertisDataFormat = ab2hex(item.advertisData)
        })
        const hit = list.find((d) => matchImeiAdvert(imei, d.advertisDataFormat))
        if (hit) {
          this.devices = [hit]
          this.history = this.history.filter((h) => h.device.deviceId !== hit.deviceId)
          this.history.push({ imei, device: hit })
          finish(hit)
          return
        }
        this.devices = this.devices.concat(list)
      })

      this.searchTimer = setTimeout(() => finish(null), timeout)
    })
  }

  connectDevice(deviceId: string, timeout = BLE_CONNECT_TIMEOUT): Promise<boolean> {
    return new Promise((resolve) => {
      uni.createBLEConnection({
        deviceId,
        timeout,
        success: () => {
          this.connected = true
          resolve(true)
        },
        fail: (err) => {
          if (err && (err as { errCode?: number }).errCode === -1) {
            this.connected = true
            resolve(true)
            return
          }
          logger.warn('createBLEConnection fail', err)
          resolve(false)
        },
      })
    })
  }

  async prepareLink(deviceId: string): Promise<BleLink | null> {
    const services = await new Promise<Array<{ uuid: string; isPrimary?: boolean }>>((resolve) => {
      uni.getBLEDeviceServices({
        deviceId,
        success: (res) => resolve((res.services || []) as Array<{ uuid: string; isPrimary?: boolean }>),
        fail: () => resolve([]),
      })
    })
    const service =
      services.find((s) => s.uuid.toUpperCase().includes(BLE_SERVICE_UUID.slice(-12).toUpperCase())) ||
      services.find((s) => s.isPrimary) ||
      services[0]
    if (!service) return null

    const chars = await new Promise<
      Array<{
        uuid: string
        properties: {
          write?: boolean
          writeNoResponse?: boolean
          notify?: boolean
          indicate?: boolean
        }
      }>
    >((resolve) => {
      uni.getBLEDeviceCharacteristics({
        deviceId,
        serviceId: service.uuid,
        success: (res) => resolve((res.characteristics || []) as never),
        fail: () => resolve([]),
      })
    })

    const write =
      chars.find((c) => c.uuid.toUpperCase() === BLE_RX_UUID.toUpperCase()) ||
      chars.find((c) => c.properties.write || c.properties.writeNoResponse)
    const notify =
      chars.find((c) => c.uuid.toUpperCase() === BLE_TX_UUID.toUpperCase()) ||
      chars.find((c) => c.properties.notify || c.properties.indicate)
    if (!write || !notify) return null

    await new Promise<void>((resolve) => {
      uni.notifyBLECharacteristicValueChange({
        deviceId,
        serviceId: service.uuid,
        characteristicId: notify.uuid,
        state: true,
        complete: () => resolve(),
      })
    })

    uni.onBLECharacteristicValueChange((res) => {
      const hex = ab2hex(res.value)
      const decoded = parseNotify(hex)
      this.notifyHandler?.(hex, decoded)
    })

    this.currentLink = {
      deviceId,
      serviceId: service.uuid,
      writeId: write.uuid,
      notifyId: notify.uuid,
    }
    return this.currentLink
  }

  onNotify(handler: ((hex: string, decoded: Record<string, unknown>) => void) | null) {
    this.notifyHandler = handler
  }

  async connectByImei(imei: string): Promise<BleLink | null> {
    const ok = await this.openAdapter()
    if (!ok) return null
    const device = await this.findDeviceByImei(imei)
    if (!device) return null
    const connected = await this.connectDevice(device.deviceId)
    if (!connected) return null
    return this.prepareLink(device.deviceId)
  }

  write(command: BleCommand, timeout = BLE_SEND_TIMEOUT): Promise<{ success: boolean; hex?: string; code?: number }> {
    const link = this.currentLink
    if (!link) return Promise.resolve({ success: false })
    return new Promise((resolve) => {
      let settled = false
      const timer = setTimeout(() => {
        if (!settled) {
          settled = true
          this.onNotify(null)
          resolve({ success: false })
        }
      }, timeout)

      this.onNotify((hex) => {
        if (settled) return
        settled = true
        clearTimeout(timer)
        this.onNotify(null)
        const { code } = resolveResult(hex)
        // Legacy: only code === 0 counts as success
        resolve({ success: code === 0, hex, code })
      })

      uni.writeBLECharacteristicValue({
        deviceId: link.deviceId,
        serviceId: link.serviceId,
        characteristicId: link.writeId,
        // @ts-expect-error uni typing
        value: command.buffer,
        fail: (err) => {
          logger.error('ble write fail', err)
          if (!settled) {
            settled = true
            clearTimeout(timer)
            this.onNotify(null)
            resolve({ success: false })
          }
        },
      })
    })
  }

  async sendRaw(imei: string, cmd: number, payload: number[] = []) {
    const tokenBytes = await fetchTokenBytes(imei)
    // buildCmd already fetches token via provider; pass empty and inject via custom:
    const hex = await buildCmd(imei, cmd, payload, async () => {
      // reconstruct numeric token from bytes if needed — use API again through provider path
      const res = await getBlueToothToken({ imei })
      return (res.data as { token?: string | number } | undefined)?.token
    })
    void tokenBytes
    const command: BleCommand = { cmd, hex, buffer: hexFrameToArrayBuffer(hex) }
    return this.write(command)
  }

  /**
   * Unlock. Default mute=[0,0] like legacy (voice played after ride report succeeds).
   */
  async unlock(imei: string, opts: { mute?: boolean } = {}) {
    const mute = opts.mute !== false
    let link = this.currentLink
    if (!link) link = await this.connectByImei(imei)
    if (!link) return { success: false as const }
    return this.sendRaw(imei, BLE_CMD.UNLOCK, mute ? MUTE_PAYLOAD : [])
  }

  async lock(imei: string, opts: { mute?: boolean } = {}) {
    const mute = opts.mute !== false
    let link = this.currentLink
    if (!link) link = await this.connectByImei(imei)
    if (!link) return { success: false as const }
    return this.sendRaw(imei, BLE_CMD.LOCK, mute ? MUTE_PAYLOAD : [])
  }

  /** Temporary park lock with TEMP_LOCK voice index. */
  async tempLock(imei: string) {
    let link = this.currentLink
    if (!link) link = await this.connectByImei(imei)
    if (!link) return { success: false as const }
    return this.sendRaw(imei, BLE_CMD.LOCK, [BLE_VOICE.TEMP_LOCK, 100])
  }

  async playVoice(imei: string, voiceIdx: number, volume = 100) {
    let link = this.currentLink
    if (!link) link = await this.connectByImei(imei)
    if (!link) return { success: false as const }
    return this.sendRaw(imei, BLE_CMD.PLAY_VOICE, [voiceIdx, volume])
  }

  async queryDeviceInfo(imei: string) {
    let link = this.currentLink
    if (!link) link = await this.connectByImei(imei)
    if (!link) return { success: false as const, data: {} }
    const res = await this.sendRaw(imei, BLE_CMD.GET_DEVICE_INFO)
    return {
      success: res.success,
      data: res.hex ? parseNotify(res.hex) : {},
      code: res.code,
    }
  }

  /** Query last beacon / road-nail info (0x42). Legacy bleCommand.checkBeacon. */
  async checkBeacon(imei: string) {
    let link = this.currentLink
    if (!link) link = await this.connectByImei(imei)
    if (!link) return { success: false as const, data: { valueDecode: {} as Record<string, unknown> } }
    const res = await this.sendRaw(imei, BLE_CMD.GET_LAST_BEACON_INFO)
    const valueDecode = (res.hex ? parseNotify(res.hex) : {}) as Record<string, unknown>
    return {
      success: Boolean(res.success),
      data: { valueDecode, hex: res.hex },
      code: res.code,
    }
  }
}

export const bleSession = new BleSession()
