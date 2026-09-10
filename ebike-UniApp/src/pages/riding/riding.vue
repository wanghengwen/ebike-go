<template>
  <view class="riding page">
    <map
      v-if="hasBikeLoc || visibleParkMarkers.length || visibleFencePolygons.length"
      id="rideMap"
      class="ride-map"
      show-location
      :latitude="mapLat"
      :longitude="mapLng"
      :scale="16"
      :markers="bikeMarkers"
      :polyline="ridePolyline"
      :polygons="visibleFencePolygons"
      @markertap="onMapMarkerTap"
      @callouttap="onCalloutTap"
    />

    <view v-if="navInstruction" class="route-hint">{{ navInstruction }}</view>
    <view v-else-if="routeHint" class="route-hint">{{ routeHint }}</view>
    <view v-if="instrumentActive" class="exit-nav" @click="onExitNav">{{ t('ride.exitNav') }}</view>
    <view
      v-if="nearServiceTip"
      class="edge-tip"
    >
      {{ nearServiceTip }}
    </view>
    <view
      v-if="fenceBanner"
      class="fence-banner"
      :style="{ background: fenceBanner.background, color: fenceBanner.color }"
    >
      <image v-if="fenceBanner.icon" class="fence-banner__icon" :src="fenceBanner.icon" mode="aspectFit" />
      <text class="fence-banner__text">{{ fenceBanner.text }}</text>
    </view>
    <view v-if="isRedEnvelopeCar" class="red-banner">{{ t('ride.redEnvelopeFree') }}</view>

    <view class="card">
      <view class="status">{{ statusText }}</view>
      <view class="car" v-if="ride.carId">NO.{{ ride.carId }}</view>
      <view class="meta">
        <view class="meta__item">
          <text class="meta__val">{{ costText }}</text>
          <text class="meta__unit">{{ t('ride.yuan') }}</text>
          <view class="meta__label">
            {{ t('ride.cost') }}
            <text class="q" @click="goBillingRules">?</text>
          </view>
        </view>
        <view class="meta__item">
          <text class="meta__val">{{ durationText }}</text>
          <view class="meta__label">{{ t('ride.duration') }}</view>
        </view>
        <view class="meta__item">
          <text class="meta__val">{{ distanceText }}</text>
          <text class="meta__unit">km</text>
          <view class="meta__label">{{ t('ride.distance') }}</view>
        </view>
      </view>
      <view class="extra" v-if="battery != null || restMileage != null">
        <view class="bat" v-if="battery != null">
          <image v-if="batteryIcon" class="bat-icon" :src="batteryIcon" mode="aspectFit" />
          <text>{{ t('ride.battery') }} {{ battery }}%</text>
        </view>
        <text v-if="restMileage != null">{{ t('ride.restMileage') }} {{ restMileage }}km</text>
      </view>
      <view class="tip" v-if="cardTip">{{ cardTip }}</view>
    </view>

    <view class="actions">
      <view
        class="btn-ghost"
        v-if="showHelmet"
        @click="onHelmet"
      >
        {{ t('ride.unlockHelmet') }}
      </view>
      <view class="btn-ghost" v-if="showTempUnlock" @click="onTempUnlock">{{ t('ride.tempUnlock') }}</view>
      <view class="btn-ghost" @click="onRing">{{ t('ride.findBike') }}</view>
      <view class="btn-ghost" @click="onTemp">
        {{ isTempPark ? t('ride.endTempPark') : t('ride.tempPark') }}
      </view>
      <view class="btn-primary" @click="onReturn">{{ t('ride.returnBike') }}</view>
      <view class="btn-ghost" @click="goApplyReturn">{{ t('ride.applyReturn') }}</view>
      <view class="btn-ghost" @click="goParkSearch">{{ t('ride.parkSearch') }}</view>
    </view>

    <CivilizationSheet
      :visible="showCivilization"
      :show-apply="showApplyLink"
      @close="showCivilization = false"
      @confirm="onCivilizationConfirm"
      @apply="onApplyFromSheet"
    />
    <ReturnCarSheet
      :visible="showReturnSheet"
      :return-type="sheetReturnType"
      :penalty="sheetPenalty"
      :can-return="sheetCanReturn"
      :show-apply="showApplyLink && sheetCanReturn"
      :can-temp-unlock="showTempUnlock"
      :auto-lock-minutes="autoLockMinutes"
      @close="showReturnSheet = false"
      @pay-dispatch="onPayDispatch"
      @refresh="onRefreshLocation"
      @near-park="onNearPark"
      @apply="onApplyFromSheet"
      @recover-power="onTempUnlock"
    />
    <HelmetSheet
      :visible="showHelmetSheet"
      :popup-type="helmetPopupType"
      :car-id="String(ride.carId || temp.ride.carId || '')"
      @close="showHelmetSheet = false"
      @cancel="onHelmetCancel"
    />
    <CarTipSheet
      :type="carTipType"
      :car-id="String(ride.carId || temp.ride.carId || '')"
      :is-use-ble="preferBleReturn"
      @change-bike="onLockChangeBike"
      @retry="onLockRetry"
    />
    <BizPopup
      :visible="showTempParkOk"
      :title="t('ride.tempParkSuccess')"
      :confirm-text="t('common.confirm')"
      :show-cancel="false"
      @confirm="showTempParkOk = false"
      @close="showTempParkOk = false"
    >
      <text>{{ t('ride.tempParkBilling') }}</text>
      <text v-if="showTempParkAutoReturn" class="temp-auto">{{ t('ride.tempParkAutoReturn') }}</text>
    </BizPopup>
    <BizPopup
      :visible="showOverload"
      :title="overloadTitle"
      :confirm-text="t('common.confirm')"
      :show-cancel="false"
      @confirm="showOverload = false"
      @close="showOverload = false"
    >
      <text>{{ t('ride.overloadHint') }}</text>
    </BizPopup>
  </view>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { onHide, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import {
  getBackCarConfig,
  getCarInfo,
  getRideInfo,
  getTempUnlockConfig,
  tempUnlock,
  unlockHelmet,
  unFrozenOrder,
} from '@/api/riding'
import { getUseCarConfig } from '@/api/user'
import { playBikeVoice } from '@/api/map'
import { useBikeRide } from '@/features/bike/useBikeRide'
import type { CommitReturnResult, ReturnPermissionData } from '@/features/bike/useBikeRide'
import { useFaceCheck } from '@/features/auth/useFaceCheck'
import {
  applyTypeFrom,
  decideReturnFlow,
  guidePageType,
} from '@/features/bike/returnTypes'
import { getRidingFenceTip } from '@/features/bike/ridingFenceTips'
import { isTimeExpired } from '@/features/bike/redEnvelope'
import { useMapLocation } from '@/features/map/useMapLocation'
import { useRideNav } from '@/features/map/useRideNav'
import { useTempDataStore } from '@/stores/tempData'
import { useUserStore } from '@/stores/user'
import CivilizationSheet from '@/widgets/CivilizationSheet.vue'
import ReturnCarSheet from '@/widgets/ReturnCarSheet.vue'
import HelmetSheet from '@/widgets/HelmetSheet.vue'
import CarTipSheet from '@/widgets/CarTipSheet.vue'
import BizPopup from '@/widgets/BizPopup.vue'
import { getTenantConfig } from '@/shared/config'
import { getBikeBatteryIcon } from '@/shared/tenantSkin'
import { openThirdPartyMap } from '@/shared/openMapApp'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import { t as i18nT } from '@/locales'

const { t } = useI18n()
const temp = useTempDataStore()
const user = useUserStore()
const { checkReturnPermission, commitReturn, pauseRide, resumeRide } = useBikeRide()
const { onEnterRiding } = useFaceCheck()
const { locate } = useMapLocation()
const {
  polyline: ridePolyline,
  parkMarkers,
  fencePolygons,
  routeMeta,
  instrumentActive,
  goNearParkWalk,
  drawRoute,
  loadRidingFences,
  applyAppNavigation,
  endInstrumentNav,
} = useRideNav()

const ride = ref<Record<string, unknown>>({})
const showHelmetSheet = ref(false)
const helmetPopupType = ref(0)
let helmet101Shown = false
const tempUnlockEnabled = ref(false)
const backCarConfig = ref<Record<string, unknown>>({})
const showCivilization = ref(false)
const showReturnSheet = ref(false)
const sheetReturnType = ref<number | string>(2101)
const sheetPenalty = ref(0)
const sheetCanReturn = ref(false)
const outServiceAutoLockMin = ref(3)
/** Legacy isUseBLE: weak-net car or tenant onlyBluetooth */
const preferBleReturn = ref(false)
const showTempParkOk = ref(false)
const showOverload = ref(false)
const showTempParkAutoReturn = ref(false)
const carTipType = ref(-1)
const overloadState = ref(0)
const lastReturnPayload = ref<Record<string, unknown>>({})
let autoLockConsumed = false
let recoverPowerPrompted = false
let pollTimer: ReturnType<typeof setInterval> | null = null
let tickTimer: ReturnType<typeof setInterval> | null = null
const localRideSec = ref(0)

/** dispatchFee in backCarConfig is boolean (show apply link), not amount */
const showApplyLink = computed(() => Boolean(backCarConfig.value.dispatchFee))
const autoLockMinutes = computed(() =>
  Number(outServiceAutoLockMin.value || backCarConfig.value.autoLockTime || 3),
)

const isTempPark = computed(() => {
  if (temp.ride.status === 'tempPark') return true
  // Legacy rideInfo uses ridingState === 3 for temp park
  return Number(ride.value.ridingState) === 3
})

const statusText = computed(() => (isTempPark.value ? t('ride.tempPark') : t('ride.riding')))

const costText = computed(() => {
  const fee = Number(ride.value.costFee ?? 0)
  if (Number.isNaN(fee)) return '0.00'
  return (fee / 100).toFixed(2)
})

const durationText = computed(() => {
  const total = Math.max(0, localRideSec.value)
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  return `${m}:${String(s).padStart(2, '0')}`
})

const distanceText = computed(() => {
  const d = Number(ride.value.rideDistance ?? 0)
  if (Number.isNaN(d)) return '0.00'
  const km = d > 1000 ? d / 1000 : d
  return km.toFixed(2)
})

const battery = computed(() => {
  const v = ride.value.restBattery ?? ride.value.soc
  return v == null ? null : Number(v)
})
const restMileage = computed(() => {
  const v = ride.value.restMileage
  return v == null ? null : Number(v)
})
const showHelmet = computed(() => {
  const state = ride.value.helmetState
  return state !== undefined && Number(state) !== 2
})

const bikeLat = computed(() => {
  const v = Number(ride.value.lat ?? ride.value.latitude ?? ride.value.carLat)
  return Number.isFinite(v) ? v : 0
})
const bikeLng = computed(() => {
  const v = Number(ride.value.lng ?? ride.value.longitude ?? ride.value.carLng)
  return Number.isFinite(v) ? v : 0
})
const hasBikeLoc = computed(() => bikeLat.value !== 0 && bikeLng.value !== 0)
/** Legacy polygonsWithFilter / markersWithFilter — hide fullCar parks */
const visibleFencePolygons = computed(() => fencePolygons.value.filter((p) => !p.fullCar))
const visibleParkMarkers = computed(() => parkMarkers.value.filter((m) => !m.fullCar))
const mapLat = computed(() => {
  const p = visibleParkMarkers.value[0]
  return p?.latitude || bikeLat.value
})
const mapLng = computed(() => {
  const p = visibleParkMarkers.value[0]
  return p?.longitude || bikeLng.value
})
const batteryIcon = computed(() =>
  getBikeBatteryIcon(ride.value.restBattery ?? ride.value.soc ?? battery.value),
)
const bikeMarkers = computed(() => {
  const list: Array<Record<string, unknown>> = []
  if (hasBikeLoc.value) {
    const iconPath = getBikeBatteryIcon(ride.value.restBattery ?? ride.value.soc ?? battery.value)
    list.push({
      id: 1,
      latitude: bikeLat.value,
      longitude: bikeLng.value,
      width: 28,
      height: 36,
      title: String(ride.value.carId || ''),
      ...(iconPath ? { iconPath } : {}),
    })
  }
  for (const m of visibleParkMarkers.value) {
    list.push({ ...m })
  }
  return list
})
const routeHint = computed(() => {
  const d = routeMeta.value.distance
  const min = routeMeta.value.durationMin
  if (d == null || min == null) return ''
  return t('ride.walkRouteHint', { m: Math.round(d), min })
})
const navInstruction = computed(() => routeMeta.value.instruction || '')

const fenceBanner = computed(() =>
  getRidingFenceTip(ride.value.type, ride.value.dispatchCost, ride.value.izCanReturn),
)
/** Legacy: !isTimeExpired(rideInfoData.expirationTime) */
const isRedEnvelopeCar = computed(() => !isTimeExpired(ride.value.expirationTime))
const overloadTitle = computed(() =>
  overloadState.value === 2 ? t('ride.overloadPowerOff') : t('ride.overloadWarn'),
)
const nearServiceTip = computed(() =>
  ride.value.izNearService ? '您在运营区边缘，骑出运营区将会自动断电' : '',
)
/** Card footer tip (apply / civilization), separate from fence banner. */
const cardTip = computed(() => {
  if (showApplyLink.value) return t('ride.returnDispatchTip')
  if (backCarConfig.value.izCivilizationRemind) return t('ride.returnRemindTip')
  return ''
})

const showTempUnlock = computed(() => {
  if (!tempUnlockEnabled.value) return false
  const co = ride.value.tempUnLockCO as Record<string, unknown> | undefined
  if (co && co.izTemp) return false
  return Boolean(ride.value.carId || temp.ride.carId)
})

function userPin() {
  return String(user.userInfo?.pin || storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
}

function serviceId() {
  return String(storage.get<string>('serviceId', '') || '')
}

function parseTempUnlockEnabled(data: unknown): boolean {
  if (data == null) return false
  if (typeof data === 'boolean') return data
  if (typeof data === 'number') return data > 0
  if (typeof data === 'object') {
    const o = data as Record<string, unknown>
    const flag = o.enabled ?? o.izTempUnlock ?? o.izOpen ?? o.open ?? o.izEnable
    if (flag != null) return Boolean(Number(flag) || flag === true)
    if (o.tempTime != null) return Number(o.tempTime) > 0
  }
  return false
}

async function loadTempUnlockConfig() {
  try {
    const res = await getTempUnlockConfig({
      serviceId: serviceId(),
      userPin: userPin(),
    })
    if (res.success) tempUnlockEnabled.value = parseTempUnlockEnabled(res.data)
  } catch (e) {
    logger.warn('getTempUnlockConfig soft fail', e)
  }
}

async function loadBackCarConfig() {
  try {
    const res = await getBackCarConfig({ serviceId: serviceId() })
    if (res.success && res.data) {
      backCarConfig.value = res.data as Record<string, unknown>
    }
  } catch (e) {
    logger.warn('getBackCarConfig soft fail', e)
  }
}

async function loadOutServiceAutoLock() {
  try {
    const res = await getUseCarConfig(serviceId() ? { serviceId: serviceId() } : {})
    const data = (res.data || {}) as {
      outServiceAreaAutoLock?: number
      parkingTime?: number
      izParkingTriggerReturnBike?: boolean
    }
    if (res.success && data.outServiceAreaAutoLock != null) {
      outServiceAutoLockMin.value = Number(data.outServiceAreaAutoLock) || 3
    }
    if (res.success) {
      showTempParkAutoReturn.value = Boolean(data.izParkingTriggerReturnBike)
    }
  } catch (e) {
    logger.warn('getUseCarConfig outService soft fail', e)
  }
}

async function fetchRideInfo() {
  const loc = temp.location || (await locate())
  try {
    const res = await getRideInfo({
      userLat: loc?.latitude,
      userLng: loc?.longitude,
      userPin: userPin(),
      version: '1.0.0',
    })
    if (!res.success || !res.data) {
      const code = Number(res.code)
      if (code === 15009) {
        stopPoll()
        navigate('reLaunch', '/pages/pay/pay')
        return
      }
      if (code === 15042) return
      // Legacy: other errors tip and go home
      stopPoll()
      uni.showModal({
        title: t('ride.riding'),
        content: res.msg || t('common.networkError'),
        showCancel: false,
        success: () => navigate('reLaunch', '/pages/home/home'),
      })
      return
    }
    const data = res.data as Record<string, unknown>
    ride.value = data
    const ov = Number(data.overloadState || 0)
    overloadState.value = ov
    if (ov === 2 || ov === 3) showOverload.value = true
    else showOverload.value = false

    const orderId = data.orderId != null ? String(data.orderId) : undefined
    const imei = data.imei != null ? String(data.imei) : undefined
    const carId = data.carId != null ? String(data.carId) : undefined
    const ridingState = Number(data.ridingState)
    temp.setRide({
      orderId,
      imei,
      carId,
      status: ridingState === 3 ? 'tempPark' : 'riding',
      startTime: temp.ride.startTime || Date.now(),
    })

    // Legacy always treats rideTime as ms
    const apiMs = Number(data.rideTime)
    if (Number.isFinite(apiMs) && apiMs > 0) {
      localRideSec.value = Math.floor(apiMs / 1000)
    }

    // Ride payload may also signal temp unlock capability
    const co = data.tempUnLockCO as Record<string, unknown> | undefined
    if (co && (co.tempTime != null || co.izTemp != null)) {
      if (Number(co.tempTime) > 0 || co.izTemp != null) tempUnlockEnabled.value = true
    }
    // Legacy: platform-initiated tempUnlockState==1 → auto show recoverPowerModal
    if (co && Number(co.tempUnlockState) === 1) {
      if (!recoverPowerPrompted) {
        recoverPowerPrompted = true
        void onTempUnlock()
      }
    } else {
      recoverPowerPrompted = false
    }

    // Legacy autoLock from customizedReturn / apply-return success
    if (!autoLockConsumed && storage.get('autoLock', false)) {
      autoLockConsumed = true
      storage.remove('autoLock')
      stopPoll()
      await onReturn()
      startPoll()
      return
    }

    syncOutOfServiceSheet(data)
    syncHelmetPopup(data)

    // park-search handoff → instrument nav (once per payload)
    if (temp.appNavigation && !instrumentActive.value) {
      const cid = String(data.carId || temp.ride.carId || '')
      void applyAppNavigation(cid)
    }
  } catch (e) {
    logger.warn('getRideInfo fail', e)
  }
}

/** Legacy isShowHelmetPopup: 1→101 once, 2→wear tip */
function syncHelmetPopup(data: Record<string, unknown>) {
  const hp = Number(data.helmetPopup)
  if (hp === 1) {
    if (storage.get('showHelmetModal', false)) {
      storage.remove('showHelmetModal')
      if (data.helmetState && !helmet101Shown) {
        helmet101Shown = true
        helmetPopupType.value = 101
        showHelmetSheet.value = true
      }
    }
    return
  }
  if (hp === 2) {
    helmetPopupType.value = 2
    showHelmetSheet.value = true
  }
}

function startPoll() {
  stopPoll()
  void fetchRideInfo()
  pollTimer = setInterval(() => {
    void fetchRideInfo()
  }, 8000)
  tickTimer = setInterval(() => {
    if (!isTempPark.value) localRideSec.value += 1
  }, 1000)
}

function stopPoll() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  if (tickTimer) {
    clearInterval(tickTimer)
    tickTimer = null
  }
}

onShow(() => {
  setNavTitle(t('ride.riding'))
  user.hydrateFromStorage()
  onEnterRiding()
  if (!temp.ride.startTime) temp.setRide({ startTime: Date.now(), status: 'riding' })
  void unFrozenOrder({
    userPin: userPin(),
    userLat: temp.location?.latitude,
    userLng: temp.location?.longitude,
  }).catch((e) => logger.warn('unFrozenOrder soft fail', e))
  void loadTempUnlockConfig()
  void loadBackCarConfig()
  void loadOutServiceAutoLock()
  void loadRidingFences()
  startPoll()
})

onHide(() => stopPoll())
onUnmounted(() => stopPoll())

async function onTemp() {
  const payload = {
    orderId: temp.ride.orderId || ride.value.orderId,
    imei: temp.ride.imei || ride.value.imei,
    carId: temp.ride.carId || ride.value.carId,
  }
  if (isTempPark.value) {
    const res = await resumeRide(payload)
    if (!res?.success) {
      const code = String((res as { code?: string | number })?.code || '')
      if (!/^17012/.test(code)) {
        uni.showToast({
          title: String((res as { msg?: string })?.msg || t('ride.resumeTempParkFail')),
          icon: 'none',
        })
      }
    }
    await fetchRideInfo()
    return
  }
  const res = await pauseRide(payload)
  if ((res as { needHelmetReturn?: boolean })?.needHelmetReturn || Number(res?.code) === 102) {
    helmetPopupType.value = 102
    showHelmetSheet.value = true
    return
  }
  if (res?.success) {
    showTempParkOk.value = true
    await fetchRideInfo()
    return
  }
  // pauseRide already tried BLE on 17012; any remaining failure matches legacy modal
  uni.showModal({
    title: t('ride.tempParkFailTitle'),
    content: String(res?.msg || t('ride.tempParkFailContent')),
    showCancel: false,
  })
  await fetchRideInfo()
}

function openReturnSheet(returnType: number | string, penalty: number, canReturn: boolean) {
  sheetReturnType.value = returnType
  sheetPenalty.value = penalty
  sheetCanReturn.value = canReturn
  showReturnSheet.value = true
}

/** Legacy getRideInfoSuccessFun: type 4101/4102 + tempUnLockCO → 10001; leave area → close */
function syncOutOfServiceSheet(data: Record<string, unknown>) {
  const fenceType = Number(data.type ?? data.ridingType ?? data.izRidingType ?? 0)
  const out = fenceType === 4101 || fenceType === 4102
  const co = data.tempUnLockCO as Record<string, unknown> | undefined

  if (!out || !co) {
    if (showReturnSheet.value && Number(sheetReturnType.value) === 10001) {
      showReturnSheet.value = false
    }
    return
  }
  if (co.izTemp) return // already powered temporarily — legacy hides / shows countdown instead
  if (showCivilization.value) return
  if (showReturnSheet.value && Number(sheetReturnType.value) === 10001) return

  const penalty = Number(data.dispatchCost ?? data.penalty ?? 0)
  const canReturn = Boolean(data.izCanReturn)
  openReturnSheet(10001, penalty, canReturn)
}

async function detectPreferBle(carId: unknown) {
  const id = String(carId || '')
  const forced = Boolean(getTenantConfig().onlyBluetooth)
  if (!id) {
    preferBleReturn.value = forced
    return preferBleReturn.value
  }
  try {
    const res = await getCarInfo({ carId: id })
    const data = (res.data || {}) as { isDisconnect?: boolean; imei?: string }
    if (data.imei) temp.setRide({ imei: String(data.imei) })
    preferBleReturn.value = forced || Boolean(data.isDisconnect)
  } catch (e) {
    logger.warn('getCarInfo soft fail', e)
    preferBleReturn.value = forced
  }
  return preferBleReturn.value
}

function applyPermissionFlow(perm: ReturnPermissionData, payload: Record<string, unknown>) {
  const flow = decideReturnFlow(perm, Boolean(backCarConfig.value.izCivilizationRemind))
  const rt = perm.returnType ?? ''
  const penalty = Number(perm.penalty ?? 0)
  sheetReturnType.value = rt || sheetReturnType.value

  if (flow === 'civilization') {
    showCivilization.value = true
    return
  }
  if (flow === 'penalty_sheet') {
    openReturnSheet(rt, penalty, true)
    return
  }
  if (flow === 'guide') {
    const pageType = guidePageType(rt)
    if (pageType != null) {
      navigate(
        'to',
        `/pages-sub/ride/return-guide/return-guide?pageType=${pageType}&returnType=${encodeURIComponent(String(rt))}&penalty=${penalty}&orderId=${encodeURIComponent(String(payload.orderId || ''))}&carId=${encodeURIComponent(String(payload.carId || ''))}`,
      )
      return
    }
  }
  if (flow === 'block_sheet') {
    openReturnSheet(rt, penalty, false)
    return
  }
  if (flow === 'silent') return
  void doCommit({ ...payload, forcePenalty: false })
}

async function handleCommitResult(res: CommitReturnResult, payload: Record<string, unknown>) {
  if (res.navigated || res.success) {
    if (res.success || res.navigated) carTipType.value = 4
    return
  }
  if (res.needPermissionRetry && res.data) {
    carTipType.value = -1
    temp.setLastReturnPermission(res.data as Record<string, unknown>)
    applyPermissionFlow(res.data, payload)
    return
  }
  carTipType.value = 5
  // toast already shown inside commitReturn for most failures
}

async function doCommit(payload: Record<string, unknown>) {
  lastReturnPayload.value = payload
  carTipType.value = 3
  const res = await commitReturn({
    ...payload,
    preferBle: preferBleReturn.value,
  })
  await handleCommitResult(res, payload)
}

function onLockChangeBike() {
  carTipType.value = -1
  navigate('reLaunch', '/pages/map/map')
}

function onLockRetry() {
  void doCommit(lastReturnPayload.value)
}

async function onReturn() {
  if (!backCarConfig.value || !Object.keys(backCarConfig.value).length) {
    await loadBackCarConfig()
  }
  const payload = {
    orderId: temp.ride.orderId || ride.value.orderId,
    imei: temp.ride.imei || ride.value.imei,
    carId: temp.ride.carId || ride.value.carId,
  }
  await detectPreferBle(payload.carId)
  const checked = await checkReturnPermission({
    ...payload,
    ...(preferBleReturn.value ? { izSw: 1 } : {}),
  })
  if (!checked.success || !checked.data) {
    // Legacy: permission 17012 → BLE return path
    if (String(checked.code || '').startsWith('17012')) {
      preferBleReturn.value = true
      const again = await checkReturnPermission({ ...payload, izSw: 1 })
      if (again.success && again.data) {
        applyPermissionFlow(again.data, payload)
        return
      }
    }
    if (checked.msg) uni.showToast({ title: String(checked.msg), icon: 'none' })
    return
  }
  applyPermissionFlow(checked.data, payload)
}

async function onCivilizationConfirm() {
  showCivilization.value = false
  await doCommit({
    orderId: temp.ride.orderId || ride.value.orderId,
    imei: temp.ride.imei || ride.value.imei,
    carId: temp.ride.carId || ride.value.carId,
    forcePenalty: false,
  })
}

async function onPayDispatch() {
  showReturnSheet.value = false
  await doCommit({
    orderId: temp.ride.orderId || ride.value.orderId,
    imei: temp.ride.imei || ride.value.imei,
    carId: temp.ride.carId || ride.value.carId,
    forcePenalty: true,
  })
}

async function onRefreshLocation() {
  // Legacy: close sheet + reload ride info; do NOT auto re-run return decision
  showReturnSheet.value = false
  uni.showToast({ title: t('returnSheet.refreshing'), icon: 'none' })
  try {
    await locate()
  } catch (e) {
    logger.warn('locate soft fail', e)
  }
  await fetchRideInfo()
}

function onApplyFromSheet() {
  showCivilization.value = false
  showReturnSheet.value = false
  const rt = sheetReturnType.value || temp.lastReturnPermission?.returnType || ''
  const at = applyTypeFrom(rt)
  const orderId = String(temp.ride.orderId || ride.value.orderId || '')
  navigate(
    'to',
    `/pages-sub/ride/apply-return/apply-return?orderId=${encodeURIComponent(orderId)}&applyType=${at}&returnType=${encodeURIComponent(String(rt))}`,
  )
}

async function onTempUnlock() {
  // Legacy recoverPowerModal: confirm before tempUnlock
  const minutes = autoLockMinutes.value
  const ok = await new Promise<boolean>((resolve) => {
    uni.showModal({
      title: t('ride.tempUnlock'),
      content: t('returnSheet.recoverPowerConfirm', { min: minutes }),
      success: (r) => resolve(Boolean(r.confirm)),
      fail: () => resolve(false),
    })
  })
  if (!ok) return
  showReturnSheet.value = false
  const carId = String(temp.ride.carId || ride.value.carId || '')
  if (!carId) return
  uni.showLoading({ title: i18nT('common.loading'), mask: true })
  try {
    const res = await tempUnlock({ carId })
    if (res.success) {
      uni.showToast({ title: t('ride.tempUnlockSuccess'), icon: 'success' })
      await fetchRideInfo()
    } else {
      uni.showToast({ title: res.msg || t('ride.tempUnlockFail'), icon: 'none' })
    }
  } finally {
    uni.hideLoading()
  }
}

async function onHelmet() {
  const carId = String(temp.ride.carId || ride.value.carId || '')
  if (!carId) return
  uni.showLoading({ title: i18nT('common.loading'), mask: true })
  try {
    const res = await unlockHelmet({ carId })
    if (res.success) {
      uni.showToast({ title: t('ride.unlockHelmet'), icon: 'success' })
    }
  } finally {
    uni.hideLoading()
  }
}

async function onRing() {
  const imei = String(temp.ride.imei || ride.value.imei || '')
  if (!imei) return
  uni.showLoading({ title: t('ride.findBike'), mask: true })
  try {
    const res = await playBikeVoice({ imei })
    uni.showToast({
      title: res.success ? t('ride.ringSuccess') : t('ride.ringFail'),
      icon: 'none',
    })
  } finally {
    uni.hideLoading()
  }
}

function goApplyReturn() {
  const orderId = String(temp.ride.orderId || ride.value.orderId || '')
  const rt = sheetReturnType.value || temp.lastReturnPermission?.returnType || ''
  const qs = [
    orderId ? `orderId=${encodeURIComponent(orderId)}` : '',
    rt ? `applyType=${applyTypeFrom(rt)}` : '',
  ]
    .filter(Boolean)
    .join('&')
  navigate('to', `/pages-sub/ride/apply-return/apply-return${qs ? `?${qs}` : ''}`)
}

function goBillingRules() {
  navigate('to', '/pages-sub/account/billing-rules/billing-rules')
}

/** Legacy returnCarModal goNearPark: close sheet, draw walk line on riding map. */
async function onNearPark() {
  showReturnSheet.value = false
  const res = await goNearParkWalk()
  if (!res.success) {
    uni.showToast({ title: res.msg || t('ride.parkEmpty'), icon: 'none' })
  }
}

async function onMapMarkerTap(e: { detail?: { markerId?: number } }) {
  const id = Number(e?.detail?.markerId)
  const park = parkMarkers.value.find((m) => m.id === id)
  if (!park) return
  const res = await drawRoute(
    { latitude: park.latitude, longitude: park.longitude },
    1,
  )
  if (!res.success) {
    uni.showToast({ title: res.msg || t('ride.navFail'), icon: 'none' })
  }
}

async function onCalloutTap(e: { detail?: { markerId?: number } }) {
  const id = Number(e?.detail?.markerId)
  const park = parkMarkers.value.find((m) => m.id === id)
  const bike =
    id === 1 && hasBikeLoc.value
      ? { latitude: bikeLat.value, longitude: bikeLng.value, title: String(ride.value.carId || '') }
      : null
  const target = park || bike
  if (!target) return
  await openThirdPartyMap('rideMap', {
    latitude: target.latitude,
    longitude: target.longitude,
    name: ('title' in target ? target.title : undefined) || t('ride.parkSearch'),
  })
}

function goParkSearch() {
  uni.chooseLocation({
    success: (res) => {
      const qs = `latitude=${res.latitude}&longitude=${res.longitude}&mode=ride`
      navigate('to', `/pages-sub/ride/park-search/park-search?${qs}`)
    },
    fail: () => {
      navigate('to', '/pages-sub/ride/park-search/park-search?mode=ride')
    },
  })
}

async function onExitNav() {
  const carId = String(ride.value.carId || temp.ride.carId || '')
  await endInstrumentNav(carId)
}

function onHelmetCancel() {
  showHelmetSheet.value = false
}
</script>

<style scoped lang="scss">
.ride-map {
  width: 100%;
  height: 360rpx;
  margin-bottom: 16rpx;
}
.edge-tip {
  margin: 0 24rpx 12rpx;
  padding: 16rpx 20rpx;
  background: #fff7e8;
  color: #b36b00;
  font-size: 24rpx;
  border-radius: 12rpx;
  line-height: 1.4;
}
.route-hint {
  margin: 0 24rpx 12rpx;
  padding: 12rpx 20rpx;
  background: #e8f3ff;
  color: #006efe;
  font-size: 24rpx;
  border-radius: 12rpx;
}
.exit-nav {
  margin: 0 24rpx 12rpx;
  padding: 12rpx 20rpx;
  text-align: center;
  color: #e65c00;
  font-size: 26rpx;
  background: #fff7e8;
  border-radius: 12rpx;
}
.fence-banner {
  margin: 0 24rpx 12rpx;
  padding: 16rpx 20rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.fence-banner__icon {
  width: 36rpx;
  height: 36rpx;
  flex-shrink: 0;
}
.fence-banner__text {
  font-size: 26rpx;
  line-height: 1.4;
  flex: 1;
}
.red-banner {
  margin: 0 24rpx 16rpx;
  padding: 16rpx 24rpx;
  background: #fff5f3;
  color: #ff5936;
  border-radius: 12rpx;
  font-size: 26rpx;
  font-weight: 600;
  text-align: center;
}
.temp-auto {
  display: block;
  margin-top: 12rpx;
  color: #ffab2c;
}
.status {
  font-size: 40rpx;
  font-weight: 700;
  margin-bottom: 8rpx;
}
.car {
  color: #888;
  margin-bottom: 24rpx;
  font-size: 26rpx;
}
.meta {
  display: flex;
  gap: 12rpx;
}
.meta__item {
  flex: 1;
  text-align: center;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 20rpx 8rpx;
}
.meta__val {
  font-size: 36rpx;
  font-weight: 700;
}
.meta__unit {
  font-size: 22rpx;
  margin-left: 4rpx;
  color: #666;
}
.meta__label {
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #888;
}
.q {
  margin-left: 6rpx;
  color: #2f80ed;
}
.extra {
  margin-top: 20rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #666;
  font-size: 24rpx;
}
.bat {
  display: flex;
  align-items: center;
  gap: 8rpx;
}
.bat-icon {
  width: 36rpx;
  height: 36rpx;
}
.tip {
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #ff8401;
  line-height: 1.4;
}
.actions {
  padding: 24rpx 32rpx;
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
</style>
