<template>
  <view class="riding-container">
    <view class="riding-map">
      <map
        id="rideMap"
        class="ride-map"
        show-location
        :provider="mapProvider()"
        :latitude="mapLat"
        :longitude="mapLng"
        :scale="17"
        :markers="bikeMarkers"
        :polyline="ridePolyline"
        :polygons="visibleFencePolygons"
        @markertap="onMapMarkerTap"
        @callouttap="onCalloutTap"
      />
    </view>

    <view class="modal-wrapper">
      <!-- 与旧版 xMapSide 一致：普通 view 挂在底部面板上，负 top 浮到地图区，弹窗可正常盖住 -->
      <view class="map-side" style="top: -320rpx">
        <view class="map-side-left">
          <view
            v-if="!instrumentActive"
            class="map-side__search"
            @tap="goSearchDest"
          >
            <image
              v-if="searchAddrIcon"
              class="map-side__search-icon"
              mode="aspectFit"
              :src="searchAddrIcon"
            />
            <text class="map-side__search-text">{{ t('ride.searchDest') }}</text>
          </view>
          <view
            v-else
            class="map-side__search map-side__search--exit"
            @tap="onExitNav"
          >
            <text class="map-side__search-text">{{ t('ride.exitNav') }}</text>
          </view>

          <view
            v-if="!instrumentActive"
            class="map-side__search map-side__park"
            @tap="goParkSearch"
          >
            <image
              v-if="parkNearIcon"
              class="map-side__search-icon"
              mode="aspectFit"
              :src="parkNearIcon"
            />
            <text class="map-side__search-text">{{ t('ride.parkSearch') }}</text>
          </view>

          <view v-if="freeReturnTip" class="map-side__tip">
            <image
              v-if="notifyIcon"
              class="map-side__tip-icon"
              mode="widthFix"
              :src="notifyIcon"
            />
            <text class="map-side__tip-text">{{ freeReturnTip }}</text>
          </view>
          <view
            v-else-if="navInstruction || routeHint"
            class="map-side__tip map-side__tip--nav"
          >
            <text class="map-side__tip-text">{{ navInstruction || routeHint }}</text>
          </view>
          <view
            v-else-if="fenceBanner"
            class="map-side__tip"
            :style="{ backgroundColor: fenceBanner.background || '#fff', color: fenceBanner.color || '#333' }"
          >
            <image
              v-if="fenceBanner.icon"
              class="map-side__tip-icon"
              mode="widthFix"
              :src="fenceBanner.icon"
            />
            <text class="map-side__tip-text">{{ fenceBanner.text }}</text>
          </view>
        </view>

        <view class="map-side-right">
          <view class="map-side__btn" @tap="goHelp">
            <image
              v-if="contactIcon"
              class="map-side__btn-img"
              mode="aspectFit"
              :src="contactIcon"
            />
          </view>
          <view class="map-side__btn" @tap="onLocate">
            <image
              v-if="locateIcon"
              class="map-side__btn-img"
              mode="aspectFit"
              :src="locateIcon"
            />
          </view>
        </view>
      </view>

      <view v-if="nearServiceTip" class="map-edge-tip">
        <view class="map-edge-tip__inner">{{ nearServiceTip }}</view>
      </view>

      <view class="modal-content" v-if="carTipType === -1">
        <view class="model-top">
          <view class="model-top__left">
            <image
              v-if="panelBatteryIcon"
              class="model-top__bat"
              mode="widthFix"
              :src="panelBatteryIcon"
            />
            <text class="km-text">{{ t('ride.restMileage') }}{{ restMileage ?? 0 }}km</text>
          </view>
          <view
            v-if="showHelmet"
            class="model-top__right"
            :style="helmetChipStyle"
            @click="onHelmet"
          >
            <image
              v-if="helmetIcon"
              class="model-top__helmet"
              mode="widthFix"
              :src="helmetIcon"
            />
            {{ t('ride.useHelmet') }}
          </view>
        </view>

        <view class="modal-ridingInfo">
          <view class="modal-ridingInfo___item modal-ridingInfo___item--cost">
            <text class="ridingInfo___item_price">{{ costText }}</text>
            <text class="ridingInfo___item_unit">{{ t('ride.yuan') }}</text>
            <view v-if="isRedEnvelopeCar" class="ridingInfo___red_envelope">
              <text class="ridingInfo___red_envelope_text">{{ t('ride.redEnvelopeFree') }}</text>
            </view>
            <view v-else class="ridingInfo___item_tips">
              <text>{{ t('ride.cost') }}</text>
              <view class="tips_question" @tap.stop="goBillingRules">
                <text class="tips_question__dot">?</text>
              </view>
            </view>
          </view>
          <view class="modal-ridingInfo___item">
            <text class="ridingInfo___item_price">{{ durationText }}</text>
            <view class="ridingInfo___item_tips">{{ t('ride.duration') }}</view>
          </view>
          <view class="modal-ridingInfo___item">
            <text class="ridingInfo___item_price">{{ distanceText }}</text>
            <text class="ridingInfo___item_unit">km</text>
            <view class="ridingInfo___item_tips">{{ t('ride.distance') }}</view>
          </view>
        </view>

        <view class="modal-return">
          <view class="left-tools">
            <view class="tool-btn" @click="onRing">
              <image
                v-if="carBellIcon"
                class="tool-btn__icon"
                mode="widthFix"
                :src="carBellIcon"
              />
              <text class="tool-btn__text">{{ t('ride.findBikeShort') }}</text>
            </view>
          </view>
          <view class="modal-return__btns">
            <view
              class="return-temping"
              :style="tempParkButtonStyle"
              @click="onTemp"
            >
              {{ isTempPark ? t('ride.endTempPark') : t('ride.tempPark') }}
            </view>
            <view
              class="return-car"
              :style="returnCarButtonStyle"
              @click="onReturn"
            >
              {{ t('ride.returnBike') }}
            </view>
          </view>
        </view>

        <view
          v-if="showApplyLink"
          class="modal-ebikeInfo"
          @click="goApplyReturn"
        >
          <text>NO.{{ resolveCarId() || ride.carId || temp.ride.carId }}</text>
          <text>{{ t('ride.cannotReturn') }}</text>
        </view>
      </view>
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
      :car-id="resolveCarId()"
      @close="showHelmetSheet = false"
      @cancel="onHelmetCancel"
    />
    <CarTipSheet
      :type="carTipType"
      :car-id="resolveCarId()"
      :is-use-ble="preferBleReturn"
      @change-bike="onLockChangeBike"
      @retry="onLockRetry"
    />
    <TempParkModal
      :visible="showTempParkOk"
      :temp-park-time="tempParkTime"
      :show-auto-return="showTempParkAutoReturn"
      @close="showTempParkOk = false"
    />
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
  tempUnlock,
  unlockHelmet,
  unFrozenOrder,
} from '@/api/riding'
import { getUseCarConfig, getConfigBaseItem } from '@/api/user'
import { getBillingConfig, playBikeVoice } from '@/api/map'
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
import TempParkModal from './components/TempParkModal.vue'
import { getBrandColor, getButtonWhiteColor, getTenantConfig } from '@/shared/config'
import {
  batteryIconBucket,
  getBikeBatteryIcon,
  getMapCfg,
} from '@/shared/tenantSkin'
import { openThirdPartyMap } from '@/shared/openMapApp'
import { navigate, setNavTitle } from '@/shared/navigate'
import { mapProvider } from '@/shared/mapProvider'
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
const tempParkTime = ref(0)
const showTempParkAutoReturn = ref(false)
const carTipType = ref(-1)
const overloadState = ref(0)
const lastReturnPayload = ref<Record<string, unknown>>({})
const billingFreeTimeMs = ref(0)
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
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
})

const distanceText = computed(() => {
  // Legacy formatMile: API rideDistance is meters → display km
  const meters = Number(ride.value.rideDistance ?? 0)
  if (!Number.isFinite(meters) || meters <= 0) return '0.00'
  return (meters / 1000).toFixed(2)
})

/** Prefer live ride info, then session, then unlock-time cache — never "null"/empty. */
function resolveCarId(): string {
  const candidates = [
    ride.value.carId,
    temp.ride.carId,
    storage.get<string>('currentRidingCarId', ''),
    temp.preCyclingCar?.carId,
  ]
  for (const c of candidates) {
    if (c == null || c === '') continue
    const s = String(c).trim()
    if (!s || s === 'null' || s === 'undefined') continue
    return s
  }
  return ''
}

function resolveImei(): string {
  const candidates = [ride.value.imei, temp.ride.imei, temp.preCyclingCar?.imei]
  for (const c of candidates) {
    if (c == null || c === '') continue
    const s = String(c).trim()
    if (!s || s === 'null' || s === 'undefined') continue
    return s
  }
  return ''
}

function resolveOrderId(): string {
  const candidates = [ride.value.orderId, temp.ride.orderId]
  for (const c of candidates) {
    if (c == null || c === '') continue
    const s = String(c).trim()
    if (!s || s === 'null' || s === 'undefined') continue
    return s
  }
  return ''
}

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
const mapLat = computed(
  () => bikeLat.value || visibleParkMarkers.value[0]?.latitude || Number(temp.location?.latitude) || 30.57,
)
const mapLng = computed(
  () => bikeLng.value || visibleParkMarkers.value[0]?.longitude || Number(temp.location?.longitude) || 104.07,
)
const panelBatteryIcon = computed(() => {
  const bucket = batteryIconBucket(ride.value.restBattery ?? ride.value.soc ?? battery.value)
  return getMapCfg(`battery${bucket}`) || getMapCfg('battery0')
})
const carBellIcon = computed(
  () =>
    String(getTenantConfig().customSetting?.carBell || '') ||
    getMapCfg('carBell') ||
    'https://ebike.luopingtech.com/miniapp/project_bwcx/carBell.png',
)
const parkNearIcon = computed(
  () => getMapCfg('searchStationIcon') || getMapCfg('station') || getMapCfg('iconSearch'),
)
const helmetIcon = computed(() => getMapCfg('helmetIcon'))
const searchAddrIcon = computed(() => getMapCfg('searchAddr') || getMapCfg('iconSearch'))
const notifyIcon = computed(() => getMapCfg('notify'))
const contactIcon = computed(() => getMapCfg('sideGetContact') || getMapCfg('customerService'))
const locateIcon = computed(() => getMapCfg('sideGetLocation'))
const returnCarButtonStyle = computed(() => {
  const bg = getBrandColor()
  return {
    backgroundColor: bg,
    borderColor: bg,
    color: '#1E4A38',
  }
})
const tempParkButtonStyle = computed(() => {
  if (!isTempPark.value) return {}
  const bg = getBrandColor()
  return {
    backgroundColor: bg,
    borderColor: bg,
    color: getButtonWhiteColor(),
  }
})
const helmetChipStyle = computed(() => {
  const bg = getBrandColor()
  return {
    backgroundColor: bg,
    color: getButtonWhiteColor(),
  }
})
/** Legacy returnFreeTime: billing freeTime(ms)/1000 - rideTime(s) */
const freeReturnSec = computed(() => {
  if (!ride.value.izFreeTime) return 0
  const freeSec = Math.floor(Number(billingFreeTimeMs.value || 0) / 1000)
  if (freeSec <= 0) return 0
  return Math.max(0, freeSec - Math.max(0, localRideSec.value))
})
const freeReturnTip = computed(() => {
  if (freeReturnSec.value <= 0) return ''
  return t('ride.freeReturnTip', { s: freeReturnSec.value })
})
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

const showTempUnlock = computed(() => {
  if (!tempUnlockEnabled.value) return false
  const co = ride.value.tempUnLockCO as Record<string, unknown> | undefined
  if (co && co.izTemp) return false
  return Boolean(resolveCarId())
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
  // Legacy riding uses getConfigBaseItem.izTempUnlock — NOT getTempUnlockConfig
  // (that API requires carId and toasts 「carId不能为null」 when missing).
  try {
    const sid = serviceId()
    const res = await getConfigBaseItem(sid ? { serviceId: sid } : {})
    if (res.success) {
      tempUnlockEnabled.value = parseTempUnlockEnabled(res.data)
    }
  } catch (e) {
    logger.warn('getConfigBaseItem tempUnlock soft fail', e)
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

async function loadBillingFreeTime() {
  try {
    const res = await getBillingConfig({
      serviceId: serviceId(),
      userPin: userPin(),
    })
    const data = (res.data || {}) as { freeTime?: number }
    if (res.success && data.freeTime != null) {
      billingFreeTimeMs.value = Number(data.freeTime) || 0
    }
  } catch (e) {
    logger.warn('getBillingConfig soft fail', e)
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
      if (data.parkingTime != null) tempParkTime.value = Number(data.parkingTime) || 0
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

    const orderId = data.orderId != null && data.orderId !== '' ? String(data.orderId) : undefined
    const imei = data.imei != null && data.imei !== '' ? String(data.imei) : undefined
    const carId =
      data.carId != null && data.carId !== '' ? String(data.carId) : undefined
    const ridingState = Number(data.ridingState)
    // Do not wipe known carId/imei/orderId when a poll omits them (legacy keeps ebikeCarId).
    const ridePatch: Record<string, unknown> = {
      status: ridingState === 3 ? 'tempPark' : 'riding',
      startTime: temp.ride.startTime || Date.now(),
    }
    if (orderId) ridePatch.orderId = orderId
    if (imei) ridePatch.imei = imei
    if (carId) {
      ridePatch.carId = carId
      storage.set('currentRidingCarId', carId)
    }
    temp.setRide(ridePatch as Parameters<typeof temp.setRide>[0])

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
    // Legacy: platform-initiated tempUnlockState==1 → confirm recover power (needs carId)
    if (co && Number(co.tempUnlockState) === 1) {
      if (!recoverPowerPrompted && resolveCarId()) {
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
      const cid = resolveCarId() || String(data.carId || temp.ride.carId || '')
      if (cid) void applyAppNavigation(cid)
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

async function runUnFrozenOrder() {
  try {
    // Legacy updatePoi first — backend requires userLat/userLng
    const loc = temp.location || (await locate())
    const lat = Number(loc?.latitude)
    const lng = Number(loc?.longitude)
    if (!Number.isFinite(lat) || !Number.isFinite(lng) || !lat || !lng) {
      logger.warn('unFrozenOrder skipped: no location')
      return
    }
    await unFrozenOrder({
      userPin: userPin(),
      userLat: lat,
      userLng: lng,
      version: '1.0.0',
    })
  } catch (e) {
    logger.warn('unFrozenOrder soft fail', e)
  }
}

onShow(() => {
  setNavTitle(t('ride.riding'))
  user.hydrateFromStorage()
  onEnterRiding()
  if (!temp.ride.startTime) temp.setRide({ startTime: Date.now(), status: 'riding' })
  void runUnFrozenOrder()
  void loadTempUnlockConfig()
  void loadBackCarConfig()
  void loadBillingFreeTime()
  void loadOutServiceAutoLock()
  void loadRidingFences()
  startPoll()
})

onHide(() => stopPoll())
onUnmounted(() => stopPoll())

async function ensureRideCarId(): Promise<string> {
  let id = resolveCarId()
  if (id) return id
  await fetchRideInfo()
  id = resolveCarId()
  return id
}

async function onTemp() {
  // Legacy tempPark: refresh getRideInfo first when carId missing
  const carId = await ensureRideCarId()
  if (!carId) {
    uni.showToast({ title: t('ride.carIdMissing'), icon: 'none' })
    return
  }
  const payload = {
    orderId: resolveOrderId(),
    imei: resolveImei(),
    carId,
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
  const carId = await ensureRideCarId()
  if (!carId) {
    uni.showToast({ title: t('ride.carIdMissing'), icon: 'none' })
    return
  }
  const payload = {
    orderId: resolveOrderId(),
    imei: resolveImei(),
    carId,
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
  const carId = resolveCarId()
  if (!carId) {
    uni.showToast({ title: t('ride.carIdMissing'), icon: 'none' })
    return
  }
  await doCommit({
    orderId: resolveOrderId(),
    imei: resolveImei(),
    carId,
    forcePenalty: false,
  })
}

async function onPayDispatch() {
  showReturnSheet.value = false
  const carId = resolveCarId()
  if (!carId) {
    uni.showToast({ title: t('ride.carIdMissing'), icon: 'none' })
    return
  }
  await doCommit({
    orderId: resolveOrderId(),
    imei: resolveImei(),
    carId,
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
  const orderId = resolveOrderId()
  navigate(
    'to',
    `/pages-sub/ride/apply-return/apply-return?orderId=${encodeURIComponent(orderId)}&applyType=${at}&returnType=${encodeURIComponent(String(rt))}`,
  )
}

async function onTempUnlock() {
  // Legacy recoverPowerModal: confirm before tempUnlock
  const carId = resolveCarId()
  if (!carId) {
    // Wait for next getRideInfo poll to retry — do not toast carId不能为null
    recoverPowerPrompted = false
    return
  }
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
  const carId = await ensureRideCarId()
  if (!carId) {
    uni.showToast({ title: t('ride.carIdMissing'), icon: 'none' })
    return
  }
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
  const imei = resolveImei()
  const carId = resolveCarId()
  if (!imei && !carId) {
    uni.showToast({ title: t('ride.carIdMissing'), icon: 'none' })
    return
  }
  uni.showLoading({ title: t('ride.findBike'), mask: true })
  try {
    const res = await playBikeVoice({
      ...(imei ? { imei } : {}),
      ...(carId ? { carId } : {}),
    })
    uni.showToast({
      title: res.success ? t('ride.ringSuccess') : res.msg || t('ride.ringFail'),
      icon: 'none',
    })
  } finally {
    uni.hideLoading()
  }
}

function goApplyReturn() {
  const orderId = resolveOrderId()
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
  const sid = serviceId() || storage.get<string>('serviceId', '') || ''
  navigate(
    'to',
    sid
      ? `/pages-sub/account/billing-rules/billing-rules?serviceId=${encodeURIComponent(sid)}`
      : '/pages-sub/account/billing-rules/billing-rules',
  )
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
  // 附近停车点：直接进停车点页，不弹微信选点
  const lat = bikeLat.value || Number(temp.location?.latitude) || 0
  const lng = bikeLng.value || Number(temp.location?.longitude) || 0
  const qs = [
    'mode=ride',
    lat ? `latitude=${lat}` : '',
    lng ? `longitude=${lng}` : '',
  ]
    .filter(Boolean)
    .join('&')
  navigate('to', `/pages-sub/ride/park-search/park-search?${qs}`)
}

/** 地图侧「搜索目的地」：先微信选点再进停车点页 */
function goSearchDest() {
  uni.chooseLocation({
    success: (res) => {
      const qs = `latitude=${res.latitude}&longitude=${res.longitude}&mode=ride`
      navigate('to', `/pages-sub/ride/park-search/park-search?${qs}`)
    },
    fail: () => {
      /* 用户取消选点，不跳转 */
    },
  })
}

function goHelp() {
  const carId = resolveCarId()
  const q = carId ? `?carId=${encodeURIComponent(carId)}` : ''
  navigate('to', `/pages-sub/support/help/help${q}`)
}

function onLocate() {
  try {
    const ctx = uni.createMapContext('rideMap')
    ctx.moveToLocation({})
  } catch (e) {
    logger.warn('moveToLocation fail', e)
  }
}

async function onExitNav() {
  const carId = resolveCarId()
  await endInstrumentNav(carId)
}

function onHelmetCancel() {
  showHelmetSheet.value = false
}
</script>

<style scoped lang="scss">
.riding-container {
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #fff;
  overflow: hidden;
}
.riding-map {
  flex: 1;
  position: relative;
  overflow: hidden;
  min-height: 0;
}
.ride-map {
  width: 100%;
  height: 100%;
}

.map-side {
  position: absolute;
  left: 0;
  width: 100%;
  z-index: 11;
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  pointer-events: none;
}
.map-side-left {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  max-width: 500rpx;
}
.map-side-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  padding-right: 24rpx;
}
.map-side__search {
  pointer-events: auto;
  display: flex;
  flex-direction: row;
  align-items: center;
  background: #fff;
  height: 70rpx;
  padding: 0 30rpx 0 20rpx;
  margin-bottom: 24rpx;
  border-top-right-radius: 30rpx;
  border-bottom-right-radius: 30rpx;
  box-shadow: 0 8rpx 16rpx rgba(0, 0, 0, 0.16);
}
.map-side__search--exit {
  background: #00a2ff;
}
.map-side__search--exit .map-side__search-text {
  color: #fff;
}
.map-side__search-icon {
  width: 40rpx;
  height: 40rpx;
  margin-right: 16rpx;
  flex-shrink: 0;
}
.map-side__search-text {
  font-size: 24rpx;
  color: #999;
  line-height: 32rpx;
}
.map-side__tip {
  pointer-events: auto;
  display: flex;
  flex-direction: row;
  align-items: center;
  max-width: 480rpx;
  background: #fff;
  height: 80rpx;
  padding: 0 20rpx;
  margin-bottom: 24rpx;
  border-top-right-radius: 30rpx;
  border-bottom-right-radius: 30rpx;
  box-shadow: 0 8rpx 16rpx rgba(0, 0, 0, 0.16);
}
.map-side__tip--nav {
  background: #e8f3ff;
}
.map-side__tip-icon {
  width: 48rpx;
  height: 48rpx;
  margin-right: 16rpx;
  flex-shrink: 0;
}
.map-side__tip-text {
  font-size: 22rpx;
  color: #333;
  line-height: 1.35;
  white-space: normal;
}
.map-side__btn {
  pointer-events: auto;
  width: 112rpx;
  height: 116rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4rpx;
}
.map-side__btn-img {
  width: 112rpx;
  height: 116rpx;
}
.map-edge-tip {
  position: absolute;
  left: 0;
  top: -280rpx;
  z-index: 11;
  pointer-events: none;
}
.map-edge-tip__inner {
  padding: 15rpx;
  font-size: 24rpx;
  color: #fff;
  background: #ff5936;
  border-top-right-radius: 30rpx;
  border-bottom-right-radius: 30rpx;
}

.modal-wrapper {
  z-index: 20;
  position: relative;
  flex-shrink: 0;
  box-sizing: border-box;
  padding-bottom: env(safe-area-inset-bottom);
  background: transparent;
}
.modal-content {
  padding-bottom: 24rpx;
  background: #fff;
  border-top-left-radius: 20rpx;
  border-top-right-radius: 20rpx;
  box-shadow: 0 -8rpx 24rpx rgba(0, 0, 0, 0.06);
  position: relative;
  z-index: 12;
}
.model-top {
  height: 114rpx;
  margin: 0 44rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 24rpx;
  border-bottom: 2rpx #f6f6f6 solid;
}
.model-top__left {
  display: flex;
  align-items: center;
}
.model-top__bat {
  width: 76rpx;
  height: 44rpx;
  margin-right: 8rpx;
  display: block;
}
.km-text {
  font-size: 28rpx;
  color: #333;
}
.model-top__right {
  display: flex;
  align-items: center;
  padding: 0 20rpx;
  border-radius: 64rpx;
  height: 64rpx;
  font-size: 28rpx;
}
.model-top__helmet {
  width: 44rpx;
  height: 44rpx;
  margin-right: 10rpx;
}

.modal-ridingInfo {
  display: flex;
  justify-content: space-between;
  margin: 46rpx 44rpx 48rpx 44rpx;
}
.modal-ridingInfo___item {
  position: relative;
  text-align: center;
  flex: 1;
}
.modal-ridingInfo___item--cost {
  z-index: 2;
}
.ridingInfo___item_price {
  font-size: 50rpx;
  font-weight: 550;
  color: #333;
}
.ridingInfo___item_unit {
  padding-left: 8rpx;
  font-size: 24rpx;
  color: #333;
}
.ridingInfo___item_tips {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #666;
  position: relative;
}
.tips_question {
  position: absolute;
  right: -48rpx;
  top: -12rpx;
  z-index: 3;
  width: 56rpx;
  height: 56rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.tips_question__dot {
  width: 30rpx;
  height: 30rpx;
  line-height: 28rpx;
  text-align: center;
  border-radius: 50%;
  color: #999;
  border: 2rpx #999 solid;
  font-size: 20rpx;
  background: #fff;
}
.ridingInfo___red_envelope {
  width: 216rpx;
  height: 44rpx;
  margin: 8rpx auto 0;
  background-color: #ffc7bb;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 22rpx;
}
.ridingInfo___red_envelope_text {
  font-size: 20rpx;
  color: #ff5936;
}

.modal-return {
  display: flex;
  align-items: flex-start;
  margin: 0 48rpx 16rpx 48rpx;
}
.left-tools {
  display: flex;
  flex-direction: column;
  justify-content: center;
  margin-right: 16rpx;
  flex-shrink: 0;
}
.tool-btn {
  background: #f0f1f3;
  border-radius: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 96rpx;
  height: 96rpx;
  flex-direction: column;
  font-size: 24rpx;
  color: #333;
  box-sizing: border-box;
  padding: 4rpx;
}
.tool-btn__icon {
  width: 40rpx;
  height: 40rpx;
  display: block;
}
.tool-btn__text {
  margin-top: 2rpx;
  font-size: 22rpx;
  line-height: 1.1;
  text-align: center;
}
.tool-btn__text--sm {
  font-size: 18rpx;
  line-height: 1.15;
}
.modal-return__btns {
  flex: 1;
  display: flex;
  align-items: center;
  height: 96rpx;
  min-width: 0;
}
.return-temping,
.return-car {
  font-size: 32rpx;
  border-radius: 30rpx;
  height: 96rpx;
  line-height: 96rpx;
  padding: 0 28rpx;
  text-align: center;
  box-sizing: border-box;
  font-weight: 600;
}
.return-temping {
  border: 2rpx #ccc solid;
  background: #fff;
  color: #333;
  margin-right: 20rpx;
  white-space: nowrap;
}
.return-car {
  flex: 1;
  min-width: 0;
}
.modal-ebikeInfo {
  text-align: center;
  font-size: 24rpx;
  color: #666;
  padding-bottom: 8rpx;
}
</style>
