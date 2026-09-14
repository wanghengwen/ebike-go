<template>
  <view class="page">
    <scroll-view scroll-y class="scroll" :style="{ height: scrollHeight }">
      <view v-if="loading" class="state">{{ t('common.loading') }}</view>
      <template v-else-if="order">
        <!-- 支付明细 -->
        <view class="card">
          <view class="card-head">
            <view class="card-head__left">
              <image v-if="iconPay" class="card-head__icon" :src="iconPay" mode="aspectFit" />
              <text class="card-head__title">{{ t('pay.payDetail') }}</text>
            </view>
            <view class="rules" @click="goBilling">
              <text>{{ t('account.billingRules') }}</text>
              <image v-if="arrowRound" class="rules__arrow" :src="arrowRound" mode="aspectFit" />
            </view>
          </view>

          <view class="detail">
            <view class="cost-row" @click="toggleOrigin">
              <text class="cost-row__label">{{ t('pay.originPrice') }}</text>
              <view class="cost-row__right">
                <text class="cost-row__price">{{ money(order.originCost) }}{{ t('pay.yuan') }}</text>
                <image
                  v-if="!hidePriceBreakdown && arrowToggle"
                  class="cost-row__arrow"
                  :src="arrowToggle"
                  mode="aspectFit"
                />
              </view>
            </view>
            <template v-if="openOrigin && !hidePriceBreakdown">
              <view class="cost-child">
                <text>{{ t('pay.startPrice') }}</text>
                <text>{{ money(order.startPrice) }}{{ t('pay.yuan') }}</text>
              </view>
              <view class="cost-child">
                <text>{{ t('pay.durationCost') }}</text>
                <text>{{ money(order.timeCost) }}{{ t('pay.yuan') }}</text>
              </view>
              <view class="cost-child">
                <text>{{ t('pay.mileCost') }}</text>
                <text>{{ money(order.mileCost) }}{{ t('pay.yuan') }}</text>
              </view>
            </template>

            <template v-if="hasPenalty">
              <view class="cost-row" @click="openOther = !openOther">
                <text class="cost-row__label">{{ t('pay.otherCost') }}</text>
                <view class="cost-row__right">
                  <text class="cost-row__price">{{ money(order.penalty) }}{{ t('pay.yuan') }}</text>
                  <image v-if="arrowOther" class="cost-row__arrow" :src="arrowOther" mode="aspectFit" />
                </view>
              </view>
              <template v-if="openOther">
                <view v-if="Number(order.dispatchCost)" class="cost-child">
                  <text>{{ t('pay.dispatchCost') }}</text>
                  <text>{{ money(order.dispatchCost) }}{{ t('pay.yuan') }}</text>
                </view>
                <view v-if="Number(order.helmetPenalty)" class="cost-child">
                  <text>{{ t('pay.helmetPenalty') }}</text>
                  <text>{{ money(order.helmetPenalty) }}{{ t('pay.yuan') }}</text>
                </view>
              </template>
            </template>

            <template v-if="hasDeduction">
              <view class="cost-row" @click="openDiscount = !openDiscount">
                <text class="cost-row__label">{{ t('pay.enjoyDiscount') }}</text>
                <view class="cost-row__right">
                  <text class="cost-row__price warn">-{{ money(order.deduction) }}{{ t('pay.yuan') }}</text>
                  <image
                    v-if="arrowDiscount"
                    class="cost-row__arrow"
                    :src="arrowDiscount"
                    mode="aspectFit"
                  />
                </view>
              </view>
              <template v-if="openDiscount">
                <view v-if="showActFree" class="cost-child">
                  <text>
                    {{ t('pay.actFree') }}
                    <text class="hint">{{ t('pay.actFreeHint') }}</text>
                  </text>
                  <text class="warn">-{{ money(order.actityFreeCount || 0) }}{{ t('pay.yuan') }}</text>
                </view>
                <view v-if="showRidingCard" class="cost-child">
                  <text>
                    {{ t('pay.ridingCardAct') }}
                    <text class="hint">{{ t('pay.actNoPenaltyHint') }}</text>
                  </text>
                  <text class="warn">-{{ money(order.ridingCardCount || 0) }}{{ t('pay.yuan') }}</text>
                </view>
                <view v-if="showDiscountAct" class="cost-child">
                  <text>
                    {{ t('pay.discountAct', { n: discountFold }) }}
                    <text class="hint">{{ t('pay.discountNoPenaltyHint') }}</text>
                  </text>
                  <text class="warn">-{{ money(order.discountCount || 0) }}{{ t('pay.yuan') }}</text>
                </view>
                <view v-if="showRandom" class="cost-child">
                  <text>{{ t('pay.randomDeduct') }}</text>
                  <text class="warn">-{{ money(order.randomCount) }}{{ t('pay.yuan') }}</text>
                </view>
                <view v-if="showRedPacket" class="cost-child">
                  <text>{{ t('pay.redEnvelopeDeduct') }}</text>
                  <text class="warn">-{{ money(order.deduction) }}{{ t('pay.yuan') }}</text>
                </view>
              </template>
            </template>
          </view>

          <view v-if="order.izImpunity" class="impunity">{{ t('pay.impunityTip') }}</view>
          <view class="divider" />
          <view class="actual">
            <text>{{ t('pay.actualPay') }}</text>
            <text class="actual__money">{{ money(order.payCost) }}</text>
            <text>{{ t('pay.yuan') }}</text>
          </view>
          <view class="actual-sub">
            {{
              t('pay.includeRechargePresent', {
                recharge: money(order.rechargeCost),
                present: money(order.presentCost),
              })
            }}
          </view>
        </view>

        <!-- 行程信息 -->
        <view class="card card--trip">
          <view class="card-head__left">
            <image v-if="iconTrack" class="card-head__icon" :src="iconTrack" mode="aspectFit" />
            <text class="card-head__title">{{ t('pay.tripInfo') }}</text>
          </view>

          <view class="trip-row">
            <text>{{ t('pay.vehicleNo') }}</text>
            <text class="trip-row__val">{{ order.carId || '-' }}</text>
          </view>
          <view class="trip-row">
            <text>{{ t('pay.startTime') }}</text>
            <text class="trip-row__val">{{ order.startTime || '-' }}</text>
          </view>
          <view class="trip-row">
            <text>{{ t('pay.endTime') }}</text>
            <text class="trip-row__val">{{ order.endTime || '-' }}</text>
          </view>
          <view class="trip-row">
            <text class="trip-row__label">{{ t('pay.rideDuration') }}</text>
            <text class="trip-row__val">{{ formatRidingTimeUnit(order.ridingTime) }}</text>
          </view>
          <view class="trip-row">
            <text class="trip-row__label">{{ t('pay.startPoint') }}</text>
            <text class="trip-row__val">{{ startAddress }}</text>
          </view>
          <view class="trip-row">
            <text class="trip-row__label">{{ t('pay.endPoint') }}</text>
            <text class="trip-row__val">{{ endAddress }}</text>
          </view>
          <view class="trip-row">
            <text class="trip-row__label">{{ t('pay.rideMile') }}</text>
            <text class="trip-row__val">{{ formatMile(order.mile) }}{{ t('pay.kmUnit') }}</text>
          </view>
        </view>
      </template>
      <view v-else class="state">{{ t('common.empty') }}</view>
    </scroll-view>

    <view v-if="order" class="tools">
      <view class="tool" @click="goTripMap">
        <image v-if="iconTrip" class="tool__icon" :src="iconTrip" mode="aspectFit" />
        <text class="tool__text">{{ t('ride.tripMap') }}</text>
      </view>
      <view v-if="canObjection" class="tool" @click="goObjection">
        <image v-if="iconObjection" class="tool__icon" :src="iconObjection" mode="aspectFit" />
        <text class="tool__text">{{ t('support.objection') }}</text>
      </view>
      <view v-if="showRepair" class="tool" @click="goRepair">
        <image v-if="iconRepair" class="tool__icon" :src="iconRepair" mode="aspectFit" />
        <text class="tool__text">{{ t('account.repair') }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getOrderDetail } from '@/api/order'
import { getConfigBaseItem } from '@/api/user'
import { formatMile, formatMoney, formatRidingTimeUnit, resolveAddress } from '@/shared/format'
import { getTenantConfig } from '@/shared/config'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'

const { t } = useI18n()

const loading = ref(false)
const orderId = ref('')
const order = ref<Record<string, unknown> | null>(null)
const startAddress = ref('--')
const endAddress = ref('--')
const showRepair = ref(false)
const openOrigin = ref(false)
const openOther = ref(false)
const openDiscount = ref(false)
const toolHeightPx = ref(90)

const iconPay = computed(() => getIconCfg('accountDetail'))
const iconTrack = computed(() => getIconCfg('track'))
const iconTrip = computed(() => getIconCfg('trackGrey'))
const iconObjection = computed(() => getIconCfg('qusetion'))
const iconRepair = computed(() => getIconCfg('carRepair'))
const arrowRound = computed(() => getMapCfg('iconRightRound'))
const arrowUp = computed(() => getIconCfg('topArrow'))
const arrowDown = computed(() => getIconCfg('bottomArrow'))
const arrowToggle = computed(() => (openOrigin.value ? arrowUp.value : arrowDown.value))
const arrowOther = computed(() => (openOther.value ? arrowUp.value : arrowDown.value))
const arrowDiscount = computed(() => (openDiscount.value ? arrowUp.value : arrowDown.value))

const hidePriceBreakdown = computed(() =>
  Boolean(getTenantConfig().customSetting?.tempHideOrderPrice),
)

const hasPenalty = computed(() => Boolean(Number(order.value?.penalty)))
const hasDeduction = computed(() => Boolean(Number(order.value?.deduction)))
const canObjection = computed(() => Number(order.value?.izComplained) === -1)

const showActFree = computed(() => Boolean(order.value?.izActityFree || order.value?.izFreeN))
const showRidingCard = computed(() => Boolean(order.value?.izRidingCard))
const showDiscountAct = computed(() => {
  const d = Number(order.value?.discount)
  return Boolean(order.value?.izDiscount) && Number.isFinite(d) && d < 1
})
const showRandom = computed(() => Number(order.value?.izRandomDeduct) === 1)
const showRedPacket = computed(
  () => Boolean(order.value?.izRedPaket) && Boolean(Number(order.value?.deduction)),
)
const discountFold = computed(() => {
  const d = Number(order.value?.discount)
  if (!Number.isFinite(d)) return ''
  const fold = d * 10
  return Number.isInteger(fold) ? String(fold) : fold.toFixed(1)
})

const scrollHeight = computed(() => `calc(100% - ${toolHeightPx.value}px)`)

onShow(() => {
  setNavTitle(t('pay.costDetail'))
  if (orderId.value) void loadDetail()
})

onLoad(async (q) => {
  orderId.value = decodeURIComponent(String(q?.orderId || ''))
  measureTools()
  await loadRepairFlag()
  if (orderId.value) await loadDetail()
})

function money(val: unknown) {
  return formatMoney(val)
}

function measureTools() {
  uni
    .createSelectorQuery()
    .select('.tools')
    .boundingClientRect((res) => {
      const box = Array.isArray(res) ? res[0] : res
      if (box && typeof box === 'object' && 'height' in box && box.height) {
        toolHeightPx.value = Number(box.height)
      }
    })
    .exec()
}

async function loadRepairFlag() {
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getConfigBaseItem(sid ? { serviceId: sid } : {})
    if (res.success && res.data) {
      const data = res.data as { izCanAfterRidingRepair?: boolean }
      showRepair.value = Boolean(data.izCanAfterRidingRepair)
    }
  } catch {
    showRepair.value = false
  }
}

async function loadDetail() {
  if (!orderId.value) return
  loading.value = true
  try {
    const res = await getOrderDetail({ orderId: orderId.value, izNewApp: true })
    if (!res.success || !res.data) {
      order.value = null
      uni.showToast({ title: t('account.ordersLoadFail'), icon: 'none' })
      return
    }
    const data = res.data as Record<string, unknown>
    order.value = data
    const paid = Number(data.izPaid)
    if (paid === 3) {
      navigate(
        'redirect',
        `/pages/pay/pay?orderId=${encodeURIComponent(orderId.value)}&isNotPollingWxScoreOrder=true`,
      )
      return
    }
    startAddress.value = await resolveAddress(data.startLat, data.startLng)
    endAddress.value = await resolveAddress(data.endLat, data.endLng)
  } finally {
    loading.value = false
    setTimeout(measureTools, 50)
  }
}

function toggleOrigin() {
  if (hidePriceBreakdown.value) return
  openOrigin.value = !openOrigin.value
}

function goBilling() {
  const sid = String(order.value?.serviceId || storage.get('serviceId', '') || '')
  navigate(
    'to',
    sid
      ? `/pages-sub/account/billing-rules/billing-rules?serviceId=${encodeURIComponent(sid)}`
      : '/pages-sub/account/billing-rules/billing-rules',
  )
}

function goTripMap() {
  // Legacy: deviceTrajectory || [] 原样传入
  const params = order.value?.deviceTrajectory || []
  navigate(
    'to',
    `/pages-sub/ride/trip-map/trip-map?params=${encodeURIComponent(JSON.stringify(params))}`,
  )
}

function goObjection() {
  navigate('to', `/pages-sub/support/objection/objection?orderId=${encodeURIComponent(orderId.value)}`)
}

function goRepair() {
  if (Number(order.value?.izRepair) === 1) {
    uni.showToast({ title: t('pay.alreadyRepaired'), icon: 'none' })
    return
  }
  const carId = String(order.value?.carId || '')
  const q = carId ? `?carId=${encodeURIComponent(carId)}` : ''
  navigate('to', `/pages-sub/support/repair/repair${q}`)
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #f8f8f8;
  position: relative;
  box-sizing: border-box;
}
.scroll {
  padding-top: 32rpx;
  box-sizing: border-box;
}
.state {
  text-align: center;
  color: #999;
  padding: 80rpx 0;
}
.card {
  margin: 0 32rpx 32rpx;
  padding: 46rpx 32rpx 48rpx;
  background: #fff;
  border-radius: 32rpx;
  box-sizing: border-box;
}
.card--trip {
  margin-bottom: 32rpx;
}
.card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-head__left {
  display: flex;
  align-items: center;
}
.card-head__icon {
  width: 48rpx;
  height: 48rpx;
  margin-right: 8rpx;
}
.card-head__title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.rules {
  display: flex;
  align-items: center;
  font-size: 24rpx;
  color: #999;
}
.rules__arrow {
  width: 24rpx;
  height: 24rpx;
}
.detail {
  margin-top: 36rpx;
}
.cost-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 32rpx;
  font-size: 28rpx;
  color: #333;
}
.cost-row:first-child {
  margin-top: 0;
}
.cost-row__label {
  font-weight: 700;
}
.cost-row__right {
  display: flex;
  align-items: center;
}
.cost-row__price {
  font-weight: 600;
  margin-right: 8rpx;
}
.cost-row__arrow {
  width: 24rpx;
  height: 24rpx;
}
.cost-child {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 16rpx 0;
  font-size: 28rpx;
  color: #666;
}
.hint {
  color: #999;
  font-size: 24rpx;
}
.warn {
  color: #ff461d !important;
}
.impunity {
  margin-top: 32rpx;
  font-size: 24rpx;
  color: #999;
}
.divider {
  height: 1px;
  margin-top: 48rpx;
  margin-bottom: 22rpx;
  background: #f6f6f6;
}
.actual {
  text-align: right;
  font-size: 28rpx;
  color: #333;
}
.actual__money {
  margin: 0 8rpx 0 16rpx;
  font-size: 40rpx;
  font-weight: 700;
}
.actual-sub {
  text-align: right;
  font-size: 28rpx;
  color: #333;
  margin-top: 8rpx;
}
.trip-row {
  margin-top: 32rpx;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  font-size: 28rpx;
  color: #333;
}
.trip-row__label {
  width: 116rpx;
  flex-shrink: 0;
}
.trip-row__val {
  flex: 1;
  font-weight: 600;
  text-align: right;
  word-break: break-all;
}
.tools {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  justify-content: space-around;
  align-items: center;
  min-height: 112rpx;
  padding-bottom: calc(24rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-sizing: border-box;
}
.tool {
  display: flex;
  flex-direction: row;
  align-items: center;
  padding: 24rpx 0;
}
.tool__icon {
  width: 40rpx;
  height: 40rpx;
  margin-right: 8rpx;
}
.tool__text {
  font-size: 28rpx;
  color: #666;
}
</style>
