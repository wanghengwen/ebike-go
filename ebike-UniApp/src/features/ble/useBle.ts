import { bleSession } from './session'
import { parseNotify } from './codec'
import { ensureBleAdapter } from './adapter'

export { ensureBleAdapter } from './adapter'
export { bleSession } from './session'
export { matchImeiAdvert } from './session'

export function useBle() {
  async function unlock(imei: string, opts?: { mute?: boolean }) {
    const ok = await ensureBleAdapter()
    if (!ok) return { success: false }
    return bleSession.unlock(imei, opts)
  }

  async function lock(imei: string, opts?: { mute?: boolean }) {
    const ok = await ensureBleAdapter()
    if (!ok) return { success: false }
    return bleSession.lock(imei, opts)
  }

  async function tempLock(imei: string) {
    const ok = await ensureBleAdapter()
    if (!ok) return { success: false }
    return bleSession.tempLock(imei)
  }

  async function playVoice(imei: string, voiceIdx: number, volume = 100) {
    const ok = await ensureBleAdapter()
    if (!ok) return { success: false }
    return bleSession.playVoice(imei, voiceIdx, volume)
  }

  async function queryDevice(imei: string) {
    const ok = await ensureBleAdapter()
    if (!ok) return { success: false, data: {} }
    return bleSession.queryDeviceInfo(imei)
  }

  async function checkBeacon(imei: string) {
    const ok = await ensureBleAdapter()
    if (!ok) return { success: false, data: { valueDecode: {} } }
    return bleSession.checkBeacon(imei)
  }

  async function connect(imei: string) {
    const ok = await ensureBleAdapter()
    if (!ok) return null
    return bleSession.connectByImei(imei)
  }

  return {
    unlock,
    lock,
    tempLock,
    playVoice,
    queryDevice,
    checkBeacon,
    connect,
    parseNotify,
    ensureBleAdapter,
    session: bleSession,
  }
}
