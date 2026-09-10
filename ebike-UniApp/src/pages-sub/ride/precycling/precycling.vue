<template>
  <view class="page precycling">
    <map
      id="precyclingMap"
      class="map"
      show-location
      :latitude="latitude"
      :longitude="longitude"
      :scale="17"
      :markers="markers"
      :polygons="polygons"
    />

    <view
      v-for="(act, idx) in floatingActs"
      :key="act.id || idx"
      class="act-float"
      :class="actFloatClass(act)"
      @click="onActivity(act)"
    >
      <image
        v-if="act.picUrl || act.imageUrl || act.iconUrl"
        class="act-float__img"
        :src="String(act.picUrl || act.imageUrl || act.iconUrl)"
        mode="aspectFit"
      />
      <text v-else class="act-float__text">{{ act.linkTitle || act.title || act.name || '' }}</text>
    </view>

    <view class="dock">
      <view v-if="areaTip" class="area-tip">
        <image v-if="areaTipIcon" class="area-tip__icon" :src="areaTipIcon" mode="aspectFit" />
        <text class="area-tip__text">{{ areaTip }}</text>
      </view>

      <view class="panel card">
      <template v-if="unavailable">
        <view class="err-title">{{ unavailableText }}</view>
        <view class="err-near" v-if="nearHint">{{ nearHint }}</view>
        <view class="err-near" v-else>{{ t('ride.bikeNearEmpty') }}</view>
        <view class="btn-primary" v-if="nearCount > 0" @click="onScanOther">{{ t('ride.scanOther') }}</view>
        <view class="btn-ghost" @click="goNearMap">{{ t('ride.viewNearBikes') }}</view>
        <view class="btn-ghost" @click="onChangeBike">{{ t('ride.changeBike') }}</view>
      </template>

      <template v-else>
        <view class="panel-tips" v-if="dispatchTips.length">
          <image
            v-if="panelTipIcon"
            class="panel-tips__icon"
            :src="panelTipIcon"
            mode="aspectFit"
          />
          <view class="panel-tips__list">
            <view class="panel-tips__item" v-for="(tip, i) in dispatchTips" :key="i">{{ tip }}</view>
          </view>
        </view>

        <view class="km-flex" v-if="loaded">
          <view class="km-item">
            <view class="km-val">
              <text class="km-num">{{ restMileage != null ? restMileage : '--' }}</text>
              <text class="km-unit">km</text>
            </view>
            <view class="km-tit">{{ t('ride.rideableDistance') }}</view>
            <view class="km-no">NO.{{ carId || '--' }}</view>
          </view>
          <view class="km-item" v-if="isRedEnvelope">
            <view class="km-val km-val--end">
              <text class="km-red">{{ t('ride.redEnvelopeEnter') }}</text>
            </view>
            <view class="km-tit">{{ t('ride.redEnvelopeRule4Free') }}</view>
            <view class="km-link" @click="goRedTips">
              {{ t('ride.redEnvelopeTips') }}
              <image v-if="arrowSmall" class="km-arrow" :src="arrowSmall" mode="aspectFit" />
            </view>
          </view>
          <view class="km-item" v-else>
            <view class="km-val km-val--end">
              <text class="km-num">{{ startPriceText }}</text>
              <text class="km-unit">{{ t('ride.yuan') }}</text>
            </view>
            <view class="km-tit">{{ startPriceTitle }}</view>
            <view class="km-link" @click="goBilling">
              {{ t('ride.billingRules') }}
              <image v-if="arrowSmall" class="km-arrow" :src="arrowSmall" mode="aspectFit" />
            </view>
          </view>
        </view>
        <view class="hint" v-else-if="loading">{{ t('common.loading') }}</view>
        <view class="hint" v-else>{{ t('common.empty') }}</view>

        <view
          class="btn-primary unlock-btn"
          :class="{ disabled: unlockDisabled }"
          @click="onUnlock"
        >
          {{ t('ride.unlock') }}
        </view>
      </template>
      </view>
    </view>

    <BizPopup
      :visible="showUnpaid"
      :title="t('pay.unpaidTip')"
      :confirm-text="t('pay.goPay')"
      :cancel-text="t('common.cancel')"
      @cancel="showUnpaid = false"
      @close="showUnpaid = false"
      @confirm="goPay"
    >
      <text>{{ t('pay.unpaidTip') }}</text>
    </BizPopup>

    <BizPopup
      :visible="showBillingPopup"
      :title="t('ride.priceAdjustNotice')"
      :confirm-text="t('common.confirm')"
      :show-cancel="false"
      @confirm="showBillingPopup = false"
      @close="showBillingPopup = false"
    >
      <text class="billing-popup-body">{{ billingPopupContent }}</text>
    </BizPopup>

    <HelmetSheet
      :visible="showHelmetSheet"
      :popup-type="helmetPopupType"
      :car-id="carId"
      @close="onHelmetClose"
      @cancel="onHelmetCancel"
      @unlock="onHelmetUnlockConfirm"
    />

    <CarTipSheet
      :type="carTipType"
      :car-id="carId"
      :is-use-ble="pendingPreferBle"
      @change-bike="onChangeBike"
      @retry="onCarTipRetry"
    />
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import BizPopup from '@/widgets/BizPopup.vue'
import HelmetSheet from '@/widgets/HelmetSheet.vue'
import CarTipSheet from '@/widgets/CarTipSheet.vue'
import { useBikeRide } from '@/features/bike/useBikeRide'
import { useScanGate, markAutoScanUseBike } from '@/features/bike/useScanGate'
import { bikeErrMessageKey, isBikeUnavailable } from '@/features/bike/bikeErrType'
import { isRedEnvelopeBike, redEnvelopeActivityId } from '@/features/bike/redEnvelope'
import { usePrecyclingMap, type ActivityEntrance } from '@/features/map/usePrecyclingMap'
import { useTempDataStore } from '@/stores/tempData'
import { getBillingConfig } from '@/api/map'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import { getTenantConfig } from '@/shared/config'
import { getMapCfg } from '@/shared/tenantSkin'

const { t } = useI18n()
const temp = useTempDataStore()
const { prepareUnlock, unlockWithBleFallback } = useBikeRide()
const { ensureCanUnlock, scanThenPrecycling } = useScanGate()
const {
  latitude,
  longitude,
  markers,
  polygons,
  floatingActs,
  hydrateFromCar,
  openActivityEntrance,
} = usePrecyclingMap()

const carId = ref('')
const loading = ref(false)
const loaded = ref(false)
const showUnpaid = ref(false)
const showBillingPopup = ref(false)
const billingPopupContent = ref('')
const carInfo = ref<Record<string, unknown>>({})
const billing = ref<Record<string, unknown>>({})
const showHelmetSheet = ref(false)
const helmetPopupType = ref(0)
const pendingPreferBle = ref(false)
const carTipType = ref(-1)
const nearCount = ref(0)
const nearDistance = ref(0)

const imei = computed(() => String(carInfo.value.imei || temp.ride.imei || ''))
const isRedEnvelope = computed(() => isRedEnvelopeBike(carInfo.value))
const unavailable = computed(() => isBikeUnavailable(carInfo.value))
const unavailableText = computed(() => {
  const key = bikeErrMessageKey(carInfo.value.errType)
  return key ? t(key) : t('ride.bikeErrDefault')
})
const nearHint = computed(() => {
  if (!nearCount.value) return ''
  return t('ride.bikeNearHint', { n: nearCount.value, m: nearDistance.value })
})
const battery = computed(() => {
  const v = carInfo.value.restBattery ?? carInfo.value.soc ?? carInfo.value.battery
  return v == null ? null : Number(v)
})
const restMileage = computed(() => {
  const v = carInfo.value.restMileage
  return v == null ? null : Number(v)
})
const startPriceText = computed(() => {
  const v = billing.value.startingPrice
  if (v == null) return '--'
  return fenYuan(v)
})
const startPriceTitle = computed(() => {
  const mins = msToMin(billing.value.startingTime)
  if (mins > 0) return `${t('ride.startingPriceShort')}(${mins}${t('ride.minutes')}${t('ride.within')})`
  return t('ride.startingPriceShort')
})
const arrowSmall = computed(() => getMapCfg('iconRightSmall') || getMapCfg('iconRight'))

/** Legacy isOutofServAera — disable unlock while outside service area. */
const unlockDisabled = computed(() => Boolean(carInfo.value.isOutofServAera))

const areaTip = computed(() => t('ride.areaRideTip'))
const areaTipIcon = computed(() => getMapCfg('notify') || getMapCfg('parkIconPanel') || '')
const panelTipIcon = computed(() => getMapCfg('parkIconPanel') || getMapCfg('notify') || '')

function fenYuan(v: unknown) {
  const n = Number(v || 0)
  return Number.isFinite(n) ? (n / 100).toFixed(2) : '0.00'
}

function msToMin(v: unknown) {
  const n = Number(v || 0)
  if (!Number.isFinite(n) || n <= 0) return 0
  return Math.round(n / 60000)
}

/** Legacy billConfigPopup body from getBillingConfig when izPopup. */
function buildBillingPopupContent(b: Record<string, unknown>): string {
  const custom = String(b.content || b.popupContent || b.notice || '').trim()
  if (custom) return custom
  if (Number(b.type) === 2) {
    const lines = (Array.isArray(b.ladderItem) ? b.ladderItem : []).map((raw) => {
      const item = raw as Record<string, unknown>
      const from = Number(item.fromIndex || 0) / 1000 / 60
      const end = Number(item.endIndex || 0) / 1000 / 60
      const price = Number(item.price || 0) / 100
      return `${from}-${end}${t('ride.minutes')} ${price.toFixed(2)}${t('ride.yuan')}`
    })
    const perMin = fenYuan(b.timeOutCostPerMin)
    const unit = msToMin(b.timeUnit) || 1
    lines.push(`${t('ride.billing')}: ${perMin}${t('ride.yuan')}/${unit}${t('ride.minutes')}`)
    return lines.filter(Boolean).join('\n') || t('ride.priceAdjustNotice')
  }
  const free = msToMin(b.freeTime)
  const start = fenYuan(b.startingPrice)
  const startMin = msToMin(b.startingTime)
  const perMin = fenYuan(b.timeOutCostPerMin)
  const unit = msToMin(b.timeUnit) || 1
  return [
    free > 0 ? `${t('ride.freeTime')}${free}${t('ride.minutes')}` : '',
    `${t('ride.startingPrice')}${start}${t('ride.yuan')}（${startMin}${t('ride.minutes')}）`,
    `${t('ride.overtimeFee')}${perMin}${t('ride.yuan')}/${unit}${t('ride.minutes')}`,
  ]
    .filter(Boolean)
    .join('\n')
}

/** Legacy ridePanelTips from billing / ruleConfig (tenant showRidePanelTips). */
const dispatchTips = computed(() => {
  const cfg = getTenantConfig().customSetting
  if (!cfg?.showRidePanelTips) return [] as string[]
  const b = billing.value
  if (!b || !Object.keys(b).length) return [] as string[]
  const tips: string[] = []
  if (b.allowOutofParking && Number(b.dispatchCost)) {
    tips.push(t('billing.parkingFee', { fee: fenYuan(b.dispatchCost) }))
  }
  if (b.allowInNostop && Number(b.penaltyInNostop)) {
    tips.push(t('billing.noParkingFee', { fee: fenYuan(b.penaltyInNostop) }))
  }
  if (b.allowOutofService && Number(b.penaltyOutofService)) {
    tips.push(t('billing.outServiceFee', { fee: fenYuan(b.penaltyOutofService) }))
  }
  return tips
})

function actFloatClass(act: ActivityEntrance) {
  const pos = Number(act.position)
  if (pos === 2 || pos === 3) return 'act-float--left'
  return 'act-float--right'
}

function onActivity(act: ActivityEntrance) {
  void openActivityEntrance(act)
}

/** Legacy checkIsShowHelmetPopup: helmetPopup === 1 */
function needHelmetGate() {
  return Number(carInfo.value.helmetPopup) === 1
}

onShow(() => setNavTitle(t('ride.precycling')))

onLoad(async (q) => {
  carId.value = decodeURIComponent((q?.carId as string) || '')
  if (!carId.value) return
  loading.value = true
  try {
    const data = await prepareUnlock(carId.value)
    carInfo.value = (data || {}) as Record<string, unknown>
    loaded.value = Boolean(data)
    temp.setPreCyclingCar({
      carId: carId.value,
      imei: String((data as { imei?: string } | null)?.imei || ''),
    })

    if (data) {
      const near = await hydrateFromCar(carInfo.value, {
        isRedEnvelope: isRedEnvelope.value,
        showNearMarkers: unavailable.value,
        carId: carId.value,
      })
      nearCount.value = near.near.count
      nearDistance.value = near.near.nearestM
    }

    if (unavailable.value) return

    const sid = String(
      carInfo.value.serviceId || storage.get<string>('serviceId', '') || '',
    )
    if (sid) {
      try {
        const res = await getBillingConfig({
          type: 0,
          serviceId: sid,
          scene: 1,
          userPin: storage.get<Record<string, unknown>>('userInfo', {})?.pin,
        })
        if (res.success && res.data) {
          billing.value = res.data as Record<string, unknown>
          // Legacy: scene=1 + izPopup → price adjust notice
          if (billing.value.izPopup) {
            billingPopupContent.value = buildBillingPopupContent(billing.value)
            showBillingPopup.value = true
          }
        }
      } catch (e) {
        logger.warn('billing soft fail', e)
      }
    }
  } finally {
    loading.value = false
  }
})

function goPay() {
  showUnpaid.value = false
  const oid = temp.unpaidOrderId
  navigate('to', oid ? `/pages/pay/pay?orderId=${encodeURIComponent(oid)}` : '/pages/pay/pay')
}

function goRedTips() {
  const id = redEnvelopeActivityId(carInfo.value)
  navigate(
    'to',
    id
      ? `/pages-sub/ride/red-envelope-tips/red-envelope-tips?id=${encodeURIComponent(id)}`
      : '/pages-sub/ride/red-envelope-tips/red-envelope-tips',
  )
}

function onChangeBike() {
  carTipType.value = -1
  // Legacy auto_scan_use_bike: return to map then open scanner
  markAutoScanUseBike()
  navigate('reLaunch', '/pages/map/map')
}

function goNearMap() {
  navigate('reLaunch', '/pages/map/map')
}

async function onScanOther() {
  await scanThenPrecycling(() => {
    showUnpaid.value = true
  })
}

function goBilling() {
  const sid = String(carInfo.value.serviceId || storage.get('serviceId', '') || '')
  navigate(
    'to',
    sid
      ? `/pages-sub/account/billing-rules/billing-rules?serviceId=${encodeURIComponent(sid)}`
      : '/pages-sub/account/billing-rules/billing-rules',
  )
}

async function beforeUnlock() {
  if (unlockDisabled.value) {
    uni.showToast({ title: t('ride.bikeErr41'), icon: 'none' })
    return false
  }
  const gate = await ensureCanUnlock(() => {
    showUnpaid.value = true
  })
  return gate.ok
}

async function onUnlock() {
  if (!(await beforeUnlock())) return
  pendingPreferBle.value = false
  if (needHelmetGate()) {
    helmetPopupType.value = 1
    showHelmetSheet.value = true
    return
  }
  await doUnlock(false)
}

/** Legacy goUnlockHelmet: show unlocking (100), then openBike — backend unlocks helmet. */
async function onHelmetUnlockConfirm() {
  helmetPopupType.value = 100
  showHelmetSheet.value = true
  const ok = await doUnlock(pendingPreferBle.value)
  if (!ok) {
    showHelmetSheet.value = false
    helmetPopupType.value = 0
  }
}

async function doUnlock(preferBle: boolean) {
  showHelmetSheet.value = false
  carTipType.value = 0
  pendingPreferBle.value = preferBle
  const res = await unlockWithBleFallback({
    carId: carId.value,
    imei: imei.value,
    skipNavigate: true,
    skipLoading: true,
    isDisconnect: Boolean(carInfo.value.isDisconnect),
    ...(preferBle ? { preferBle: true } : {}),
    onPhase: (phase: 'network' | 'ble') => {
      pendingPreferBle.value = phase === 'ble'
    },
  })
  if (res?.success) {
    carTipType.value = 1
    setTimeout(() => navigate('reLaunch', '/pages/riding/riding'), 800)
    return true
  }
  // 15030 already navigates to recharge in useBikeRide.unlock
  if (/15030/.test(String((res as { code?: string | number })?.code || ''))) {
    carTipType.value = -1
    return false
  }
  carTipType.value = 2
  return false
}

function onCarTipRetry() {
  // Retry still goes network-first; 17012 will auto-fallback to BLE again.
  void doUnlock(false)
}

function onHelmetClose() {
  if (helmetPopupType.value === 100) return
  showHelmetSheet.value = false
}

function onHelmetCancel() {
  showHelmetSheet.value = false
  helmetPopupType.value = 0
}
</script>

<style scoped lang="scss">
.precycling {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  padding: 0;
  background: #f5f6f8;
}
.map {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
}
.act-float {
  position: absolute;
  z-index: 7;
  width: 96rpx;
  height: 96rpx;
  bottom: calc(280rpx + env(safe-area-inset-bottom));
}
.act-float--right {
  right: 24rpx;
}
.act-float--left {
  left: 24rpx;
}
.act-float__img {
  width: 96rpx;
  height: 96rpx;
}
.act-float__text {
  font-size: 20rpx;
  color: #333;
  background: #fff;
  border-radius: 12rpx;
  padding: 8rpx;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.08);
}
.dock {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 6;
  padding: 0 24rpx calc(24rpx + env(safe-area-inset-bottom));
  pointer-events: none;
}
/* WXSS does not support universal selector `*` */
.dock .area-tip,
.dock .panel {
  pointer-events: auto;
}
.area-tip {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 16rpx 20rpx;
  margin-bottom: 16rpx;
  background: #fff;
  border-radius: 12rpx;
  box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.06);
}
.area-tip__icon {
  width: 36rpx;
  height: 36rpx;
  flex-shrink: 0;
}
.area-tip__text {
  flex: 1;
  font-size: 24rpx;
  color: #333;
  line-height: 1.4;
}
.panel {
  max-height: 58vh;
  overflow-y: auto;
  margin: 0;
  padding: 32rpx 32rpx 28rpx;
}
.panel-tips {
  display: flex;
  gap: 12rpx;
  align-items: flex-start;
  margin: 0 0 16rpx;
  padding: 16rpx;
  background: #f7fafc;
  border-radius: 12rpx;
}
.panel-tips__icon {
  width: 36rpx;
  height: 36rpx;
  flex-shrink: 0;
  margin-top: 2rpx;
}
.panel-tips__list {
  flex: 1;
}
.panel-tips__item {
  font-size: 22rpx;
  color: #666;
  line-height: 1.5;
}
.panel-tips__item + .panel-tips__item {
  margin-top: 6rpx;
}
.km-flex {
  display: flex;
  margin: 8rpx 0 36rpx;
}
.km-item {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.km-val {
  display: flex;
  align-items: baseline;
}
.km-val--end {
  justify-content: flex-end;
}
.km-num {
  font-size: 56rpx;
  font-weight: 700;
  color: #333;
  line-height: 1.1;
}
.km-unit {
  margin-left: 6rpx;
  font-size: 24rpx;
  color: #666;
}
.km-red {
  font-size: 40rpx;
  font-weight: 700;
  color: #ff5936;
}
.km-tit {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #666;
}
.km-val--end + .km-tit,
.km-item:last-child .km-tit {
  text-align: right;
}
.km-no {
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #999;
}
.km-link {
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #999;
  display: flex;
  align-items: center;
  justify-content: flex-end;
}
.km-arrow {
  width: 24rpx;
  height: 24rpx;
  margin-left: 4rpx;
}
.err-title {
  margin: 28rpx 0 12rpx;
  color: #ff5936;
  font-size: 30rpx;
  font-weight: 600;
  line-height: 1.4;
}
.err-near {
  color: #666;
  font-size: 26rpx;
  margin-bottom: 28rpx;
  line-height: 1.4;
}
.hint {
  margin: 24rpx 0 32rpx;
  color: #888;
}
.unlock-btn {
  margin-top: 8rpx;
}
.btn-ghost {
  margin-top: 20rpx;
}
.btn-primary.disabled,
.btn-ghost.disabled {
  opacity: 0.45;
  pointer-events: none;
}
.billing-popup-body {
  white-space: pre-wrap;
  line-height: 1.5;
}
</style>
