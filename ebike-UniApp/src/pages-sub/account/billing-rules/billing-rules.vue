<template>
  <view class="page">
    <view v-if="loading" class="card">{{ t('common.loading') }}</view>
    <template v-else-if="rule">
      <view class="card">
        <view class="sec-title">{{ t('account.billingRules') }}</view>
        <template v-if="Number(rule.type) !== 2">
          <view class="row">
            <text>{{ t('billing.startPrice') }}</text>
            <text class="des">{{ fenYuan(rule.startingPrice) }}{{ t('ride.yuan') }}</text>
          </view>
          <view class="row sub">
            <text>{{ t('billing.includeTime') }}</text>
            <text class="des">{{ msMin(rule.startingTime) }}{{ t('ride.minutes') }}</text>
          </view>
          <view class="row sub">
            <text>{{ t('billing.includeMile') }}</text>
            <text class="des">{{ meterKm(rule.startingDistance) }}km</text>
          </view>
          <view class="row">
            <text>{{ t('billing.overTime', { n: msMin(rule.startingTime) }) }}</text>
            <text class="des">{{ timeBillText }}</text>
          </view>
          <view class="row" v-if="!hidePrice">
            <text>{{ t('billing.overMile', { n: meterKm(rule.startingDistance) }) }}</text>
            <text class="des">{{ distanceBillText }}</text>
          </view>
        </template>
        <template v-else>
          <view class="row head"><text>{{ t('billing.startPrice') }}</text></view>
          <view class="row" v-for="(item, i) in ladder" :key="i">
            <text>
              {{ item.fromIndex }}-{{ item.endIndex }}{{ t('ride.minutes') }}
              <text class="tips">({{ t('billing.excludeEnd', { n: item.endIndex }) }})</text>
            </text>
            <text class="des">{{ item.price }}{{ t('ride.yuan') }}</text>
          </view>
          <view class="tips-line" v-if="rule.izAccumulate">{{ t('billing.accumulate') }}</view>
          <view class="tips-line">{{ t('billing.overByTime') }}</view>
          <view class="row">
            <text>{{ t('billing.timeFee') }}</text>
            <text class="des">
              {{ fenYuan(rule.timeOutCostPerMin) }}{{ t('ride.yuan') }}/{{ msMin(rule.timeUnit) }}{{ t('ride.minutes') }}
            </text>
          </view>
        </template>
      </view>

      <view class="card">
        <view class="sec-title">{{ t('billing.discountTitle') }}</view>
        <view class="row">
          <text>{{ t('billing.freeTime') }}</text>
          <text class="des">{{ freeTimeText }}</text>
        </view>
        <view class="row">
          <text>{{ t('billing.freeMile') }}</text>
          <text class="des">{{ freeMileText }}</text>
        </view>
        <view class="row">
          <text>{{ t('billing.discount') }}</text>
          <text class="des">{{ discountText }}</text>
        </view>
      </view>

      <view class="card">
        <view class="sec-title">{{ t('billing.rideRules') }}</view>
        <view class="rule-block">
          <view class="rule-name">{{ t('billing.serviceArea') }}</view>
          <view class="rule-body" v-if="rule.allowOutofService">
            {{ t('billing.outServiceFee', { fee: fenYuan(rule.penaltyOutofService) }) }}
          </view>
          <view class="rule-body" v-else>{{ t('billing.outServiceNoReturn') }}</view>
          <view class="rule-hint">{{ t('billing.outServiceHint') }}</view>
        </view>
        <view class="rule-block">
          <view class="rule-name">{{ t('billing.noParking') }}</view>
          <view class="rule-body" v-if="rule.allowInNostop">
            {{ t('billing.noParkingFee', { fee: fenYuan(rule.penaltyInNostop) }) }}
          </view>
          <view class="rule-body" v-else>{{ t('billing.noParkingNoReturn') }}</view>
          <view class="rule-hint">{{ t('billing.noParkingHint') }}</view>
        </view>
        <view class="rule-block">
          <view class="rule-name">{{ t('billing.parking') }}</view>
          <view class="rule-body" v-if="rule.allowOutofParking">
            {{ t('billing.parkingFee', { fee: fenYuan(rule.dispatchCost) }) }}
          </view>
          <view class="rule-body" v-else>{{ t('billing.parkingNoReturn') }}</view>
          <view class="rule-hint">{{ t('billing.parkingHint') }}</view>
        </view>
      </view>
    </template>
    <view v-else class="card">{{ t('common.empty') }}</view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getBillingConfig } from '@/api/map'
import { getTenantConfig } from '@/shared/config'
import { useUserStore } from '@/stores/user'
import { setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

type LadderItem = { fromIndex: number; endIndex: number; price: number }

const { t } = useI18n()
const user = useUserStore()
const loading = ref(true)
const rule = ref<Record<string, unknown> | null>(null)
const hidePrice = computed(() => Boolean(getTenantConfig().customSetting?.tempHideOrderPrice))

const ladder = computed<LadderItem[]>(() => {
  const list = (rule.value?.ladderItem || []) as Array<Record<string, unknown>>
  return list.map((item) => ({
    fromIndex: Number(item.fromIndex || 0) / 60000,
    endIndex: Number(item.endIndex || 0) / 60000,
    price: Number(item.price || 0) / 100,
  }))
})

const timeBillText = computed(() => {
  if (!rule.value) return ''
  const fee = fenYuan(rule.value.timeOutCostPerMin)
  const unit = msMin(rule.value.timeUnit)
  return `${fee}${t('ride.yuan')}/${unit}${t('ride.minutes')}`
})
const distanceBillText = computed(() => {
  if (!rule.value) return ''
  return `${fenYuan(rule.value.overDistanceCostPerMter)}${t('ride.yuan')}/km`
})
const freeTimeText = computed(() => {
  const v = Number(rule.value?.freeTime || 0)
  return v ? `${msMin(v)}${t('ride.minutes')}/${t('billing.perRide')}` : t('billing.none')
})
const freeMileText = computed(() => {
  const v = Number(rule.value?.freeDistance || 0)
  return v ? `${v}m/${t('billing.perRide')}` : t('billing.none')
})
const discountText = computed(() => {
  const d = Number(rule.value?.discount)
  const x = Number.isFinite(d) ? d * 10 : 10
  if (x === 10) return t('billing.noDiscount')
  if (x === 0) return t('billing.freeRide')
  return t('billing.discountOff', { n: x })
})

function fenYuan(v: unknown) {
  return (Number(v || 0) / 100).toFixed(2).replace(/\.?0+$/, '') || '0'
}
function msMin(v: unknown) {
  return Math.round(Number(v || 0) / 60000)
}
function meterKm(v: unknown) {
  return (Number(v || 0) / 1000).toFixed(2).replace(/\.?0+$/, '') || '0'
}

onShow(() => setNavTitle(t('account.billingRules')))

onLoad(async (q) => {
  loading.value = true
  user.hydrateFromStorage()
  const serviceId = String(q?.serviceId || storage.get('serviceId', '') || '')
  const pin = String(user.userInfo?.pin || '')
  const res = await getBillingConfig({
    ...(serviceId ? { serviceId } : {}),
    ...(pin ? { userPin: pin } : {}),
  })
  loading.value = false
  if (res.success && res.data && typeof res.data === 'object' && !Array.isArray(res.data)) {
    rule.value = res.data as Record<string, unknown>
  }
})
</script>

<style scoped lang="scss">
.sec-title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.row {
  display: flex;
  justify-content: space-between;
  padding: 18rpx 0;
  border-bottom: 1px solid #f3f3f3;
  &.sub {
    color: #888;
    font-size: 26rpx;
  }
  &.head {
    border-bottom: none;
    font-weight: 600;
  }
}
.des {
  color: #1a1a1a;
}
.tips {
  color: #aaa;
  font-size: 22rpx;
  margin-left: 8rpx;
}
.tips-line {
  color: #888;
  font-size: 24rpx;
  margin: 8rpx 0;
}
.rule-block {
  margin-top: 28rpx;
}
.rule-name {
  font-weight: 600;
  margin-bottom: 8rpx;
}
.rule-body {
  color: #333;
  line-height: 1.5;
}
.rule-hint {
  color: #888;
  font-size: 24rpx;
  margin-top: 6rpx;
}
</style>
