<template>
  <view class="map-page">
    <!-- 返回/定位必须用 cover-view，否则会被原生 map 盖住点不到 -->
    <map
      class="map"
      id="fullMap"
      show-location
      :provider="mapProvider()"
      :latitude="latitude"
      :longitude="longitude"
      :scale="16"
      :markers="markers"
      :polygons="polygons"
      :polyline="polyline"
      @markertap="onMarkertap"
      @callouttap="onCalloutTap"
      @tap="onMapTap"
    >
      <cover-view
        v-if="showPageBack"
        class="back"
        :style="{ top: statusBarPx + 12 + 'px' }"
        @tap="goBack"
      >
        <cover-image v-if="backIcon" class="back__img" :src="backIcon" />
      </cover-view>

      <cover-view class="side" :style="{ bottom: sideBottom }">
        <cover-view class="side__btn" @tap.stop="onLocate">
          <cover-image v-if="locateIcon" class="side__img" :src="locateIcon" />
        </cover-view>
        <cover-view
          v-if="redEnvelopeCount > 0"
          class="side__btn side__btn--text"
          @tap.stop="onToggleRedEnvelope"
        >
          <cover-view class="side__text">{{
            redEnvelopeMapMode ? t('ride.redEnvelopeExit') : t('ride.redEnvelopeEnter')
          }}</cover-view>
        </cover-view>
      </cover-view>
    </map>

    <view class="bike-panel card" v-if="selectedBike">
      <view class="bike-panel__row">
        <text class="bike-panel__no">NO.{{ selectedBike.carId || selectedBike.title }}</text>
        <text class="bike-panel__battery" v-if="batteryText">{{ batteryText }}</text>
      </view>
      <view class="bike-panel__mileage" v-if="mileageText">{{ mileageText }}</view>
      <view class="bike-panel__bill" v-if="billingText" @click="goBillingRules">
        <text>{{ billingText }}</text>
        <text class="bike-panel__bill-link">{{ t('ride.billingRules') }} ›</text>
      </view>
      <view class="bike-panel__red" v-if="selectedIsRedEnvelope" @click="goRedTips">
        <text class="bike-panel__red-text">{{ t('ride.redEnvelopeFree') }}</text>
        <text class="bike-panel__red-link">{{ t('ride.redEnvelopeTips') }} ›</text>
      </view>
      <view class="bike-panel__actions">
        <view class="btn-ghost mini" @click="onRing">{{ t('ride.findBike') }}</view>
        <view class="btn-ghost mini" @click="onNav">{{ t('ride.navigate') }}</view>
        <view class="btn-primary mini" @click="onUseBike">{{ t('ride.useBike') }}</view>
      </view>
    </view>

    <view class="dock" :style="{ height: dockH }">
      <view
        class="dock__btn"
        :class="{ 'is-disabled': isLimitRiding }"
        :style="isLimitRiding ? disabledBtnStyle : brandBtnStyle"
        @click="onScan"
      >
        <image v-if="scanIcon" class="dock__scan-icon" :src="scanIcon" mode="widthFix" />
        <text>{{ t('ride.useBike') }}</text>
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

    <PrivacyAuthorizePopup />
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import BizPopup from '@/widgets/BizPopup.vue'
import PrivacyAuthorizePopup from '@/widgets/PrivacyAuthorizePopup.vue'
import { useHomeMap } from '@/features/map/useHomeMap'
import { useScanGate, consumeAutoScanUseBike } from '@/features/bike/useScanGate'
import { useCreditLimit } from '@/features/credit/useCreditLimit'
import { showBeforeUseCarPopup } from '@/features/guide/useGuidePopup'
import { useTempDataStore } from '@/stores/tempData'
import { useUserStore } from '@/stores/user'
import { getPersonInfo } from '@/api/user'
import { navigate } from '@/shared/navigate'
import { isNative } from '@/shared/nativeHost'
import { mapProvider } from '@/shared/mapProvider'
import { openThirdPartyMap } from '@/shared/openMapApp'
import { logger } from '@/shared/logger'
import { isRedEnvelopeBike, redEnvelopeActivityId } from '@/features/bike/redEnvelope'
import { getBillingConfig } from '@/api/map'
import { storage } from '@/shared/storage'
import {
  getBrandColor,
  getButtonDisabledColor,
  getButtonWhiteColor,
  getTenantConfig,
} from '@/shared/config'
import { getMapCfg } from '@/shared/tenantSkin'

const { t } = useI18n()
const temp = useTempDataStore()
const user = useUserStore()
const {
  latitude,
  longitude,
  markers,
  polygons,
  polyline,
  selectedBike,
  redEnvelopeCount,
  redEnvelopeMapMode,
  refreshMap,
  relocate,
  onMarkerTap,
  ringSelectedBike,
  navigateToMarker,
  clearRoute,
  findMarkerById,
  setRedEnvelopeMapMode,
} = useHomeMap()
const { scanThenPrecycling, ensureCanScan } = useScanGate()
const { isLimitRiding, refresh: refreshCredit } = useCreditLimit()

const showUnpaid = ref(false)
const statusBarPx = ref(20)
const safeBottom = ref(0)
const billingText = ref('')
const billingServiceId = ref('')

const brandBtnStyle = computed(() => {
  const c = getBrandColor()
  return { backgroundColor: c, borderColor: c, color: getButtonWhiteColor() }
})
const disabledBtnStyle = computed(() => {
  const c = getButtonDisabledColor()
  return { backgroundColor: c, borderColor: c, color: getButtonWhiteColor() }
})
const dockH = computed(() => (safeBottom.value > 0 ? '212rpx' : '160rpx'))
const sideBottom = computed(() => (safeBottom.value > 0 ? '240rpx' : '188rpx'))
const scanIcon = computed(() => getMapCfg('iconScan'))
const backIcon = computed(
  () =>
    String(getTenantConfig().customSetting?.backIcon || '') ||
    getMapCfg('iconBack') ||
    String(getTenantConfig().customSetting?.navBackIcon || ''),
)
/** 原生宿主用容器顶栏后退，地图页不再叠 cover-view 返回。 */
const showPageBack = computed(() => !isNative())
const locateIcon = computed(() => getMapCfg('sideGetLocation'))

const batteryText = computed(() => {
  const src = selectedBike.value?.sourceData
  if (!src) return ''
  const bat = src.restBattery ?? src.soc ?? src.battery
  if (bat == null) return ''
  return `${t('ride.battery')}: ${bat}%`
})

const mileageText = computed(() => {
  const src = selectedBike.value?.sourceData
  const km = src?.restMileage
  if (km == null) return ''
  return `${t('ride.restMileage')}: ${km} km`
})

const selectedIsRedEnvelope = computed(() =>
  isRedEnvelopeBike(selectedBike.value?.sourceData as Record<string, unknown> | undefined),
)

async function loadSelectedBilling() {
  billingText.value = ''
  billingServiceId.value = ''
  const src = selectedBike.value?.sourceData as Record<string, unknown> | undefined
  if (!src) return
  const sid = String(src.serviceId || storage.get<string>('serviceId', '') || '')
  if (!sid) return
  billingServiceId.value = sid
  try {
    const res = await getBillingConfig({
      type: 0,
      serviceId: sid,
      userPin: storage.get<Record<string, unknown>>('userInfo', {})?.pin,
    })
    if (!res.success || !res.data) return
    const b = res.data as Record<string, unknown>
    const starting = b.startingPrice != null ? Number(b.startingPrice) / 100 : null
    const mins =
      b.startingTime != null ? Math.round(Number(b.startingTime) / 60000) : null
    if (starting != null && !Number.isNaN(starting)) {
      billingText.value =
        mins != null && !Number.isNaN(mins)
          ? `${starting.toFixed(2)}${t('ride.yuan')}/${mins}${t('ride.minutes')}`
          : `${starting.toFixed(2)}${t('ride.yuan')}`
    }
  } catch (e) {
    logger.warn('map billing soft fail', e)
  }
}

onShow(() => {
  user.hydrateFromStorage()
  void refreshMap({ keepSelection: true })
  if (temp.unpaidOrderId) showUnpaid.value = true
  if (user.isLoggedIn) {
    void refreshCredit(false)
    void getPersonInfo()
      .then((res) => {
        if (!res.success || !res.data) return
        const profile = res.data as { ridingState?: number; payState?: number }
        user.setUserInfo(res.data as never)
        if ([4, 5, 6].includes(Number(profile.ridingState))) {
          navigate('reLaunch', '/pages/riding/riding')
          return
        }
        if (Number(profile.payState) === 7) showUnpaid.value = true
      })
      .catch((e) => logger.warn('map personInfo soft fail', e))
  }
  if (consumeAutoScanUseBike()) {
    setTimeout(() => {
      void scanThenPrecycling(() => {
        showUnpaid.value = true
      })
    }, 400)
  }
})

onMounted(() => {
  try {
    const info = uni.getSystemInfoSync()
    statusBarPx.value = info.statusBarHeight || 20
    safeBottom.value = Number((info.safeAreaInsets as { bottom?: number } | undefined)?.bottom || 0)
  } catch {
    statusBarPx.value = 20
  }
  void refreshMap()
})

function goBack() {
  navigate('back')
}
function goPay() {
  showUnpaid.value = false
  const oid = temp.unpaidOrderId
  const qs = ['isEBikeLock=true']
  if (oid) qs.push(`orderId=${encodeURIComponent(oid)}`)
  navigate('reLaunch', `/pages/pay/pay?${qs.join('&')}`)
}
async function onLocate() {
  await relocate()
}
function onToggleRedEnvelope() {
  const next = !redEnvelopeMapMode.value
  setRedEnvelopeMapMode(next)
  void refreshMap({ keepSelection: true, redEnvelopeOnly: next })
}
function goRedTips() {
  const src = selectedBike.value?.sourceData as Record<string, unknown> | undefined
  const id = redEnvelopeActivityId(src)
  navigate(
    'to',
    id
      ? `/pages-sub/ride/red-envelope-tips/red-envelope-tips?id=${encodeURIComponent(id)}`
      : '/pages-sub/ride/red-envelope-tips/red-envelope-tips',
  )
}
function onMapTap() {
  clearRoute()
  selectedBike.value = null
  billingText.value = ''
}
async function onMarkertap(e: { detail?: { markerId?: number }; markerId?: number }) {
  const id = e?.detail?.markerId ?? e?.markerId
  if (id == null) return
  await onMarkerTap(id)
  await loadSelectedBilling()
}
function goBillingRules() {
  const sid = billingServiceId.value || storage.get<string>('serviceId', '') || ''
  navigate(
    'to',
    sid
      ? `/pages-sub/account/billing-rules/billing-rules?serviceId=${encodeURIComponent(sid)}`
      : '/pages-sub/account/billing-rules/billing-rules',
  )
}
async function onRing() {
  await ringSelectedBike()
}
async function onNav() {
  await navigateToMarker()
}
async function onCalloutTap(e: { detail?: { markerId?: number }; markerId?: number }) {
  const id = e?.detail?.markerId ?? e?.markerId
  if (id == null) return
  const marker = findMarkerById(id) || selectedBike.value
  if (!marker) return
  await openThirdPartyMap('fullMap', {
    latitude: marker.latitude,
    longitude: marker.longitude,
    name: String(marker.title || marker.carId || t('ride.navigate')),
  })
}
async function onUseBike() {
  const carId = selectedBike.value?.carId
  if (!carId) return
  const gate = await ensureCanScan({ skipVerified: true })
  if (!gate.ok) {
    if (gate.reason === 'unpaid') showUnpaid.value = true
    return
  }
  await showBeforeUseCarPopup()
  navigate('to', `/pages-sub/ride/precycling/precycling?carId=${encodeURIComponent(carId)}`)
}
async function onScan() {
  await scanThenPrecycling(() => {
    showUnpaid.value = true
  })
}
</script>

<style scoped lang="scss">
.map-page {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: #f5f6f8;
}
.map {
  width: 100%;
  height: 100%;
}
.back {
  position: absolute;
  left: 20rpx;
  width: 96rpx;
  height: 96rpx;
}
.back__img {
  width: 96rpx;
  height: 96rpx;
}
.side {
  position: absolute;
  right: 24rpx;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}
.side__btn {
  width: 112rpx;
  height: 116rpx;
  margin-bottom: 16rpx;
  background-color: transparent;
}
.side__img {
  width: 112rpx;
  height: 116rpx;
}
.side__btn--text {
  height: auto;
  min-height: 64rpx;
  padding: 12rpx 16rpx;
  box-sizing: border-box;
  background-color: #ffffff;
  border-radius: 24rpx;
}
.side__text {
  font-size: 22rpx;
  color: #ff5936;
  text-align: center;
  line-height: 1.3;
}
.bike-panel {
  position: absolute;
  left: 24rpx;
  right: 24rpx;
  bottom: 300rpx;
  z-index: 18;
  margin: 0;
}
.bike-panel__row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8rpx;
}
.bike-panel__no {
  font-weight: 700;
  font-size: 32rpx;
}
.bike-panel__battery {
  color: #63d144;
  font-size: 24rpx;
}
.bike-panel__mileage {
  color: #666;
  font-size: 24rpx;
  margin-bottom: 12rpx;
}
.bike-panel__bill {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #333;
  font-size: 24rpx;
  margin-bottom: 12rpx;
}
.bike-panel__bill-link {
  color: #3aa0e8;
  font-size: 22rpx;
}
.bike-panel__red {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff5f3;
  border-radius: 8rpx;
  padding: 12rpx 16rpx;
  margin-bottom: 16rpx;
}
.bike-panel__red-text {
  color: #ff5936;
  font-size: 24rpx;
  font-weight: 600;
}
.bike-panel__red-link {
  color: #ff5936;
  font-size: 22rpx;
}
.bike-panel__actions {
  display: flex;
  gap: 16rpx;
}
.mini {
  flex: 1;
  padding: 18rpx 12rpx !important;
  font-size: 24rpx;
}
.dock {
  position: absolute;
  left: 0;
  bottom: 0;
  width: 100%;
  z-index: 20;
  background-color: #ffffff;
  border-top-left-radius: 32rpx;
  border-top-right-radius: 32rpx;
  box-sizing: border-box;
}
.dock__btn {
  height: 96rpx;
  margin: 32rpx 24rpx 0;
  border-radius: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 34rpx;
  box-sizing: border-box;
}
.dock__scan-icon {
  width: 40rpx;
  height: 40rpx;
  margin-right: 16rpx;
}
.dock__btn.is-disabled {
  pointer-events: none;
}
</style>
