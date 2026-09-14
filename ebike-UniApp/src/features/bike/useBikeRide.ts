import {
  getCarInfo,
  networkRide,
  returnByNet,
  returnPermission,
  tempPark,
  endTempPark,
  partMatchByTempParking,
} from '@/api/riding'
import {
  bleRidePermission,
  bleRideReport,
  bleReturnReport,
  bleTempParkReport,
  bleEndTempParkReport,
  returnConfig,
} from '@/api/ble'
import { submitScanData } from '@/api/preCycling'
import { useTempDataStore } from '@/stores/tempData'
import { useUserStore } from '@/stores/user'
import { useBle } from '@/features/ble/useBle'
import { useMapLocation } from '@/features/map/useMapLocation'
import { BLE_VOICE } from '@/features/ble/commands'
import { decideReturnKind } from '@/features/bike/returnTypes'
import { t } from '@/locales'
import { navigate } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import { getTenantConfig } from '@/shared/config'

function userPin(): string {
  const user = useUserStore()
  return String(user.userInfo?.pin || storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
}

export type ReturnPermissionData = {
  izCanReturn?: boolean
  returnType?: number | string
  penalty?: number
  analyze?: unknown
}

export type CommitReturnResult = {
  success: boolean
  code?: string | number
  msg?: string
  data?: ReturnPermissionData | null
  /** Legacy: returnByNet failed with code 0 → re-run UI permission matrix */
  needPermissionRetry?: boolean
  /** Already navigated to pay / frozen pay */
  navigated?: boolean
  izWxScoreOrder?: boolean
}

export function useBikeRide() {
  const temp = useTempDataStore()
  const ble = useBle()
  const map = useMapLocation()

  async function withLocation(payload: Record<string, unknown> = {}) {
    const loc =
      payload.userLat != null && payload.userLng != null
        ? { latitude: Number(payload.userLat), longitude: Number(payload.userLng) }
        : temp.location || (await map.locate())
    if (!loc?.latitude || !loc?.longitude) {
      return {
        userLat: undefined as number | undefined,
        userLng: undefined as number | undefined,
        userPin: payload.userPin || userPin(),
      }
    }
    return {
      userLat: loc.latitude,
      userLng: loc.longitude,
      userPin: payload.userPin || userPin(),
    }
  }

  async function prepareUnlock(carId: string) {
    uni.showLoading({ title: t('common.loading'), mask: true })
    try {
      void submitScanData({ carId })
      const info = await getCarInfo({ carId })
      if (!info.success) return null
      temp.setRide({
        carId,
        imei: (info.data as { imei?: string })?.imei,
        status: 'precycling',
      })
      return info.data
    } finally {
      uni.hideLoading()
    }
  }

  async function unlock(payload: Record<string, unknown>) {
    const skipNavigate = Boolean(payload.skipNavigate)
    const skipLoading = Boolean(payload.skipLoading)
    if (!skipLoading) uni.showLoading({ title: t('ride.unlock'), mask: true })
    try {
      const locFields = await withLocation(payload)
      if (locFields.userLat == null || locFields.userLng == null) {
        uni.showModal({ title: t('ride.unlock'), content: t('ride.needLocation'), showCancel: false })
        return { success: false as const, code: 'LOC', msg: t('ride.needLocation') }
      }
      // Align with current old package: pay-score createOrder is commented out — do not attach channelEnum=14
      const body: Record<string, unknown> = {
        ...payload,
        ...locFields,
        carId: payload.carId || temp.ride.carId,
      }
      // imei not required by legacy networkRide body
      delete body.imei
      delete body.preferBle
      delete body.result
      delete body.skipNavigate
      delete body.skipLoading
      delete body.onPhase
      delete body.isDisconnect
      delete body.deferFailUi
      const res = await networkRide(body)
      if (res.success) {
        storage.set('showHelmetModal', true)
        const unlockedCarId = String(payload.carId || temp.ride.carId || '')
        const unlockedImei = String(payload.imei || temp.ride.imei || '')
        temp.setRide({
          status: 'riding',
          orderId: (res.data as { orderId?: string })?.orderId,
          startTime: Date.now(),
          ...(unlockedCarId ? { carId: unlockedCarId } : {}),
          ...(unlockedImei ? { imei: unlockedImei } : {}),
        })
        if (unlockedCarId) storage.set('currentRidingCarId', unlockedCarId)
        if (!skipNavigate) {
          setTimeout(() => navigate('reLaunch', '/pages/riding/riding'), 800)
        }
      } else if (/15030/.test(String(res.code || ''))) {
        // Legacy: insufficient balance → recharge
        uni.showModal({
          title: t('ride.unlockFail'),
          content: res.msg || t('ride.unlockFail'),
          showCancel: false,
          success: (r) => {
            if (r.confirm) navigate('to', '/pages-sub/pay/recharge/recharge')
          },
        })
      } else if (String(res.code || '').startsWith('17012')) {
        if (!skipNavigate && !payload.deferFailUi) {
          uni.showModal({
            title: t('ride.unlock'),
            content: res.msg || t('ride.unlockFail'),
            showCancel: false,
          })
        }
      }
      return res
    } finally {
      if (!skipLoading) {
        try {
          uni.hideLoading()
        } catch {
          /* already hidden */
        }
      }
    }
  }

  async function unlockByBle(payload: Record<string, unknown> = {}) {
    const skipNavigate = Boolean(payload.skipNavigate)
    const skipLoading = Boolean(payload.skipLoading)
    const imei = String(payload.imei || temp.ride.imei || '')
    const carId = String(payload.carId || temp.ride.carId || '')
    if (!imei) return { success: false as const, code: 'NO_IMEI', msg: t('ride.unlockFail') }

    if (!skipLoading) uni.showLoading({ title: t('ride.bleConnecting'), mask: true })
    try {
      const locFields = await withLocation(payload)
      if (locFields.userLat == null || locFields.userLng == null) {
        uni.showModal({ title: t('ride.unlock'), content: t('ride.needLocation'), showCancel: false })
        return { success: false as const, code: 'LOC', msg: t('ride.needLocation') }
      }
      const perm = await bleRidePermission({
        carId,
        izSw: 1,
        result: payload.result || [],
        userLat: locFields.userLat,
        userLng: locFields.userLng,
        userPin: locFields.userPin,
      })
      if (!perm.success) return perm

      const bleRes = await ble.unlock(imei, { mute: true })
      if (!bleRes.success) return { success: false as const, code: 'BLE', msg: t('ride.unlockFail') }

      const reported = await bleRideReport({
        carId,
        userLat: locFields.userLat,
        userLng: locFields.userLng,
        userPin: locFields.userPin,
      })
      if (reported.success) {
        await ble.playVoice(imei, BLE_VOICE.START, 100)
        storage.set('showHelmetModal', true)
        temp.setRide({
          status: 'riding',
          orderId: (reported.data as { orderId?: string })?.orderId,
          startTime: Date.now(),
          ...(carId ? { carId } : {}),
          ...(imei ? { imei } : {}),
        })
        if (carId) storage.set('currentRidingCarId', carId)
        if (!skipNavigate) {
          setTimeout(() => navigate('reLaunch', '/pages/riding/riding'), 800)
        }
        return { success: true as const, data: reported.data }
      }
      await ble.lock(imei, { mute: true })
      return { success: false as const, code: reported.code, msg: reported.msg || t('ride.unlockFail') }
    } finally {
      if (!skipLoading) {
        try {
          uni.hideLoading()
        } catch {
          /* already hidden */
        }
      }
    }
  }

  /**
   * Unlock: network first; auto BLE on 17012 (or preferBle / onlyBluetooth / isDisconnect).
   * onPhase('ble') fires before BLE so UI can show「蓝牙开锁中」without flashing fail.
   */
  async function unlockWithBleFallback(payload: Record<string, unknown> = {}) {
    const onPhase = payload.onPhase as ((phase: 'network' | 'ble') => void) | undefined
    const forceBle =
      Boolean(payload.preferBle) ||
      Boolean(getTenantConfig().onlyBluetooth) ||
      Boolean(payload.isDisconnect)

    const blePayload = {
      ...payload,
      skipLoading: true,
      deferFailUi: true,
    }

    if (forceBle) {
      onPhase?.('ble')
      return unlockByBle(blePayload)
    }

    onPhase?.('network')
    const net = await unlock({
      ...payload,
      skipLoading: true,
      deferFailUi: true,
    })
    if (net.success) return net
    if (/15030/.test(String(net.code || ''))) return net

    if (String(net.code || '').startsWith('17012')) {
      onPhase?.('ble')
      return unlockByBle(blePayload)
    }
    return net
  }

  /**
   * Only run returnPermission — UI (civilization / penalty sheets) decides next step.
   */
  async function checkReturnPermission(payload: Record<string, unknown> = {}) {
    const locFields = await withLocation(payload)
    if (locFields.userLat == null || locFields.userLng == null) {
      uni.showModal({ title: t('ride.returnBike'), content: t('ride.needLocation'), showCancel: false })
      return { success: false as const, data: null as ReturnPermissionData | null }
    }
    const carId = String(payload.carId || temp.ride.carId || storage.get('currentRidingCarId', '') || '').trim()
    if (!carId || carId === 'null' || carId === 'undefined') {
      return {
        success: false as const,
        data: null as ReturnPermissionData | null,
        code: 'NO_CAR',
        msg: t('ride.carIdMissing'),
      }
    }
    const body = {
      orderId: payload.orderId || temp.ride.orderId,
      carId,
      userPin: locFields.userPin,
      userLat: locFields.userLat,
      userLng: locFields.userLng,
      izFrontSuppotFullPile: true,
      ...(payload.izSw != null ? { izSw: payload.izSw } : {}),
    }
    const perm = await returnPermission(body)
    if (perm.success && perm.data) {
      const data = perm.data as ReturnPermissionData
      temp.setLastReturnPermission({ ...data, ...body })
      return { success: true as const, data, locFields, body }
    }
    return { success: false as const, data: null, code: perm.code, msg: perm.msg, locFields, body }
  }

  /**
   * Actually return: returnByNet with returnType 0|1, BLE fallback on 17012 / preferBle.
   * Failure matrix (legacy returnCarByNet):
   * - success → pay (or wx-score pay)
   * - 17012 → BLE lock + report
   * - 15002 → pay?isTempFrozon=true
   * - 15009 → treat as finished → pay
   * - code 0 / falsy → needPermissionRetry with data for UI sheets
   * - other → toast msg
   */
  async function commitReturn(payload: Record<string, unknown> = {}): Promise<CommitReturnResult> {
    uni.showLoading({ title: t('ride.returnBike'), mask: true })
    try {
      const locFields = await withLocation(payload)
      const forcePenalty = Boolean(payload.forcePenalty)
      const preferBle = Boolean(payload.preferBle)
      const orderId = payload.orderId || temp.ride.orderId
      const carId = String(
        payload.carId || temp.ride.carId || storage.get('currentRidingCarId', '') || '',
      ).trim()
      const imei = String(payload.imei || temp.ride.imei || '')
      const oid = String(orderId || '')
      if (!carId || carId === 'null' || carId === 'undefined') {
        uni.showToast({ title: t('ride.carIdMissing'), icon: 'none' })
        return { success: false, msg: t('ride.carIdMissing') }
      }

      const goPay = (extra = '') => {
        temp.setRide({ status: 'ended' })
        const qs = `orderId=${encodeURIComponent(oid)}${extra}`
        navigate('reLaunch', `/pages/pay/pay?${qs}`)
      }

      const bleLockAndReport = async (): Promise<CommitReturnResult> => {
        if (!imei) {
          uni.showToast({ title: t('ride.returnFail'), icon: 'none' })
          return { success: false, msg: t('ride.returnFail') }
        }

        // Legacy returnBikeByBLE: returnConfig.izBeacon → checkBeacon → accessoriesList
        const accessoriesList: Array<Record<string, unknown>> = Array.isArray(payload.result)
          ? [...(payload.result as Array<Record<string, unknown>>)]
          : []
        try {
          const cfgRes = await returnConfig({ id: carId })
          const cfg = (cfgRes.data || {}) as { izBeacon?: boolean | number }
          const needBeacon = Boolean(Number(cfg.izBeacon) || cfg.izBeacon === true)
          if (needBeacon) {
            const beaconRes = await ble.checkBeacon(imei)
            const valueDecode = (beaconRes.data?.valueDecode || {}) as Record<string, unknown>
            if (beaconRes.success && valueDecode.event) {
              accessoriesList.push({
                name: 'beacon',
                izExist: true,
                canUse: true,
                result: true,
                state: valueDecode,
              })
            } else {
              logger.warn('BLE beacon check fail/empty', beaconRes)
            }
          }
        } catch (e) {
          logger.warn('returnConfig/beacon soft fail', e)
        }

        const blePerm = await returnPermission({
          orderId,
          carId,
          userPin: locFields.userPin,
          userLat: locFields.userLat,
          userLng: locFields.userLng,
          izFrontSuppotFullPile: true,
          izSw: 1,
          result: accessoriesList,
        })
        if (!blePerm.success) {
          return {
            success: false,
            code: blePerm.code,
            msg: blePerm.msg,
            data: (blePerm.data as ReturnPermissionData) || null,
          }
        }
        const bleRes = await ble.lock(imei, { mute: true })
        if (!bleRes.success) {
          uni.showToast({ title: t('ride.returnFail'), icon: 'none' })
          return { success: false, msg: t('ride.returnFail') }
        }
        const reported = await bleReturnReport({
          orderId,
          carId,
          izSw: 1,
          returnType: 3,
          userLng: locFields.userLng,
          userLat: locFields.userLat,
          userPin: locFields.userPin,
          result: accessoriesList,
          izFrontSuppotFullPile: true,
        })
        if (reported.success) {
          await ble.playVoice(imei, BLE_VOICE.LOCK, 100)
          uni.showToast({ title: t('ride.returnSuccess'), icon: 'none' })
          setTimeout(() => goPay(), 800)
          return { success: true, navigated: true }
        }
        uni.showToast({ title: reported.msg || t('ride.returnFail'), icon: 'none' })
        return { success: false, code: reported.code, msg: reported.msg }
      }

      // Weak-net / forced BLE: skip network return, lock via BLE
      if (preferBle) {
        return bleLockAndReport()
      }

      const netBody = {
        orderId,
        carId,
        userPin: locFields.userPin,
        userLat: locFields.userLat,
        userLng: locFields.userLng,
        returnType: forcePenalty ? 1 : 0,
        izFrontSuppotFullPile: true,
      }
      const res = await returnByNet(netBody)
      if (res.success) {
        const data = (res.data || {}) as { izWxScoreOrder?: boolean }
        uni.showToast({ title: t('ride.returnSuccess'), icon: 'none' })
        const wx = Boolean(data.izWxScoreOrder)
        setTimeout(() => goPay(wx ? '&wxScore=1' : ''), 800)
        return { success: true, navigated: true, izWxScoreOrder: wx }
      }

      const codeStr = String(res.code ?? '')
      const codeNum = Number(res.code)

      // Network device control failed → BLE
      if (imei && codeStr.startsWith('17012')) {
        return bleLockAndReport()
      }

      // Temp frozen order — pay countdown
      if (codeNum === 15002) {
        navigate(
          'reLaunch',
          `/pages/pay/pay?orderId=${encodeURIComponent(oid)}&isTempFrozon=true`,
        )
        return { success: false, code: 15002, navigated: true, msg: res.msg }
      }

      // Order already finished
      if (codeNum === 15009) {
        uni.showToast({ title: t('ride.returnSuccess'), icon: 'none' })
        setTimeout(() => goPay(), 800)
        return { success: true, navigated: true, code: 15009 }
      }

      // Fence reject with code 0 → re-open permission UI matrix
      if (!codeNum) {
        return {
          success: false,
          code: res.code,
          msg: res.msg,
          data: (res.data as ReturnPermissionData) || null,
          needPermissionRetry: true,
        }
      }

      uni.showToast({ title: res.msg || t('ride.returnFail'), icon: 'none' })
      return { success: false, code: res.code, msg: res.msg, data: (res.data as ReturnPermissionData) || null }
    } finally {
      uni.hideLoading()
    }
  }

  /** Convenience: permission + auto commit for simple cases (used rarely). Prefer UI matrix. */
  async function returnBike(payload: Record<string, unknown>) {
    if (payload.forcePenalty || payload.skipPermission) {
      return commitReturn(payload)
    }
    const checked = await checkReturnPermission(payload)
    if (!checked.success || !checked.data) return checked
    const pdata = checked.data
    if (!pdata.izCanReturn) return { success: false, data: pdata }
    if (decideReturnKind(pdata.returnType) === 'unnormal' && !payload.forcePenalty) {
      return { success: false, data: pdata, needPenaltyConfirm: true }
    }
    return commitReturn({
      ...payload,
      returnType: pdata.returnType,
      forcePenalty: false,
    })
  }

  async function pauseRide(payload: Record<string, unknown>) {
    const carId = String(payload.carId || temp.ride.carId || '').trim()
    if (!carId || carId === 'null' || carId === 'undefined') {
      return { success: false as const, code: 'NO_CAR', msg: t('ride.carIdMissing') }
    }
    const pin = payload.userPin || userPin()
    const orderId = payload.orderId || temp.ride.orderId
    const serviceId =
      payload.serviceId ||
      storage.get<string>('serviceId', '') ||
      (storage.get<Record<string, unknown>>('userInfo', {}) as { serviceId?: string })?.serviceId ||
      ''

    // Legacy matchPartByTempParking: success + partName=helmet + !izMatch → show helmet 102
    try {
      const match = await partMatchByTempParking({
        carId,
        orderId,
        serviceId,
        userPin: pin,
      })
      if (match.success) {
        const data = (match.data || {}) as { partName?: string; izMatch?: boolean }
        if (data.partName === 'helmet' && !data.izMatch) {
          return {
            success: false as const,
            code: 102,
            needHelmetReturn: true as const,
            msg: match.msg || t('ride.helmetNotBack'),
          }
        }
      } else if (match.msg) {
        uni.showToast({ title: String(match.msg), icon: 'none' })
      }
    } catch (e) {
      logger.warn('partMatchByTempParking soft fail', e)
    }

    const res = await tempPark({ carId, userPin: pin })
    if (res.success) {
      temp.setRide({ status: 'tempPark' })
      return res
    }
    if (!String(res.code || '').startsWith('17012')) return res

    const imei = String(payload.imei || temp.ride.imei || '')
    if (!imei) return res
    const locFields = await withLocation(payload)
    const bleRes = await ble.tempLock(imei)
    if (!bleRes.success) return { success: false }
    const reported = await bleTempParkReport({
      carId,
      userPin: pin,
      userLat: locFields.userLat,
      userLng: locFields.userLng,
    })
    if (reported.success) {
      temp.setRide({ status: 'tempPark' })
      return { success: true }
    }
    return res
  }

  async function resumeRide(payload: Record<string, unknown>) {
    const carId = String(payload.carId || temp.ride.carId || '').trim()
    if (!carId || carId === 'null' || carId === 'undefined') {
      return { success: false as const, code: 'NO_CAR', msg: t('ride.carIdMissing') }
    }
    const pin = payload.userPin || userPin()
    const res = await endTempPark({ carId, userPin: pin })
    if (res.success) {
      temp.setRide({ status: 'riding' })
      return res
    }
    if (!String(res.code || '').startsWith('17012')) return res

    const imei = String(payload.imei || temp.ride.imei || '')
    if (!imei) return res
    const locFields = await withLocation(payload)
    const bleRes = await ble.unlock(imei, { mute: true })
    if (!bleRes.success) return { success: false }
    const reported = await bleEndTempParkReport({
      carId,
      userPin: pin,
      userLat: locFields.userLat,
      userLng: locFields.userLng,
    })
    if (reported.success) {
      await ble.playVoice(imei, BLE_VOICE.START, 100)
      temp.setRide({ status: 'riding' })
      return { success: true }
    }
    return res
  }

  return {
    prepareUnlock,
    unlock,
    unlockWithBleFallback,
    checkReturnPermission,
    commitReturn,
    returnBike,
    pauseRide,
    resumeRide,
  }
}
