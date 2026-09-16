<template>
  <!-- Legacy accountRules.vue -->
  <view class="page">
    <scroll-view v-if="rule" class="scroll_view" scroll-y>
      <view class="container rule-content">
        <view class="top-left">
          <image v-if="iconAccount" class="img" :src="iconAccount" mode="aspectFit" />
          <text class="text">{{ t('account.billingRules') }}</text>
        </view>
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
            <text class="des">{{ meterKm(rule.startingDistance) }}{{ t('billing.km') }}</text>
          </view>
          <view class="row">
            <text>{{ t('billing.overTime', { n: msMin(rule.startingTime) }) }}</text>
            <text class="des">{{ timeBillText }}</text>
          </view>
          <view v-if="!hidePrice" class="row">
            <text>{{ t('billing.overMile', { n: meterKm(rule.startingDistance) }) }}</text>
            <text class="des">{{ distanceBillText }}</text>
          </view>
        </template>
        <template v-else>
          <view class="tit"><text>{{ t('billing.startPrice') }}</text><text /></view>
          <view v-for="(item, index) in ladder" :key="index" class="tit">
            <view class="flex">
              {{ item.fromIndex }}-{{ item.endIndex }}{{ t('ride.minutes') }}
              <text class="tips tips2">({{ t('billing.excludeEnd', { n: item.endIndex }) }})</text>
            </view>
            <text>{{ item.price }}{{ t('ride.yuan') }}</text>
          </view>
          <view v-if="rule.izAccumulate" class="tips"><text>{{ t('billing.accumulate') }}</text></view>
          <view class="tips"><text>{{ t('billing.overByTime') }}</text></view>
          <view class="tit">
            <text>{{ t('billing.timeFee') }}</text>
            <text>
              {{ fenYuan(rule.timeOutCostPerMin) }}{{ t('ride.yuan') }}/{{ msMin(rule.timeUnit)
              }}{{ t('ride.minutes') }}
            </text>
          </view>
          <view class="tips">
            <text>{{ t('billing.timeUnitTip', { n: msMin(rule.timeUnit) }) }}</text>
          </view>
        </template>
      </view>

      <view class="container">
        <view class="top-left">
          <image v-if="iconDiscount" class="img" :src="iconDiscount" mode="aspectFit" />
          <text class="text">{{ t('billing.discountTitle') }}</text>
        </view>
        <view class="row">
          <text>{{ t('billing.freeRideHint') }}</text>
          <text class="remark">{{ t('billing.freeRideOver') }}</text>
        </view>
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

      <view class="container">
        <view class="top-left">
          <image v-if="iconOrder" class="img" :src="iconOrder" mode="aspectFit" />
          <text class="text">{{ t('billing.rideRules') }}</text>
        </view>

        <view class="cycle">
          <image v-if="iconService" class="cyc-img" :src="iconService" mode="aspectFit" />
          <view class="cyc-des">
            <text class="title">{{ t('billing.serviceArea') }}</text>
            <text v-if="rule.allowOutofService" class="sub-title">
              {{ t('billing.outServiceFee', { fee: fenYuan(rule.penaltyOutofService) }) }}
            </text>
            <text v-else class="content">{{ t('billing.outServiceNoReturn') }}</text>
            <text class="content">{{ t('billing.outServiceHint') }}</text>
          </view>
        </view>

        <view class="cycle" style="margin-top: 48rpx">
          <image v-if="iconNoPark" class="cyc-img" :src="iconNoPark" mode="aspectFit" />
          <view class="cyc-des">
            <text class="title">{{ t('billing.noParking') }}</text>
            <text v-if="rule.allowInNostop" class="sub-title">
              {{ t('billing.noParkingFee', { fee: fenYuan(rule.penaltyInNostop) }) }}
            </text>
            <text v-else class="content">{{ t('billing.noParkingNoReturn') }}</text>
            <text class="content">{{ t('billing.noParkingHint') }}</text>
          </view>
        </view>

        <view class="cycle" style="margin-top: 48rpx">
          <image v-if="iconPark" class="cyc-img" :src="iconPark" mode="aspectFit" />
          <view class="cyc-des">
            <text class="title">{{ t('billing.parking') }}</text>
            <text v-if="rule.allowOutofParking" class="sub-title">
              {{ t('billing.parkingFee', { fee: fenYuan(rule.dispatchCost) }) }}
            </text>
            <text v-else class="content">{{ t('billing.parkingNoReturn') }}</text>
            <text class="content">{{ t('billing.parkingHint') }}</text>
          </view>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getBillingConfig } from '@/api/map'
import { getTenantConfig } from '@/shared/config'
import { getLoginPath, navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { getIconCfg } from '@/shared/tenantSkin'
import { useUserStore } from '@/stores/user'

type LadderItem = { fromIndex: number; endIndex: number; price: number }

const { t } = useI18n()
const user = useUserStore()
const rule = ref<Record<string, unknown> | null>(null)
const hidePrice = computed(() => Boolean(getTenantConfig().customSetting?.tempHideOrderPrice))

const iconAccount = computed(() => getIconCfg('accountDetail'))
const iconDiscount = computed(() => getIconCfg('moneyDiscount'))
const iconOrder = computed(() => getIconCfg('order'))
const iconService = computed(() => getIconCfg('priceRuleService'))
const iconNoPark = computed(() => getIconCfg('priceRuleNoParking'))
const iconPark = computed(() => getIconCfg('priceRuleParking'))

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
  const unit = msMin(rule.value.timeUnit)
  const unitText = unit === '--' ? '--' : String(parseInt(String(unit), 10))
  return `${fenYuan(rule.value.timeOutCostPerMin)}${t('ride.yuan')}/${unitText}${t('ride.minutes')}`
})

const distanceBillText = computed(() => {
  if (!rule.value) return ''
  return `${fenYuan(rule.value.overDistanceCostPerMter)}${t('ride.yuan')}/${t('billing.km')}`
})

const freeTimeText = computed(() => {
  const freeTime = rule.value?.freeTime
  if (freeTime || freeTime === 0) {
    return `${Math.floor(Number(freeTime) / 60000)}${t('ride.minutes')}/${t('billing.perRide')}`
  }
  return `--${t('ride.minutes')}/${t('billing.perRide')}`
})

const freeMileText = computed(() => {
  const freeDistance = rule.value?.freeDistance
  if (freeDistance === undefined || freeDistance === null) {
    return `--${t('billing.meterPerRide')}`
  }
  return `${freeDistance}${t('billing.meterPerRide')}`
})

const discountText = computed(() => {
  const discount = rule.value?.discount
  if (discount === undefined || discount === null) return ''
  const x = Number(discount) * 10
  if (x === 10) return t('billing.noDiscount')
  if (x === 0) return t('billing.freeRide')
  return t('billing.discountOff', { n: x })
})

function fenYuan(v: unknown) {
  const n = Number(v || 0) / 100
  if (!Number.isFinite(n)) return '--'
  return String(parseFloat(n.toFixed(2)))
}

function msMin(v: unknown) {
  if (v === undefined || v === null || v === '') return '--'
  return Math.round(Number(v) / 60000)
}

function meterKm(v: unknown) {
  const n = Number(v || 0) / 1000
  if (!Number.isFinite(n)) return '--'
  return String(parseFloat(n.toFixed(2)))
}

onShow(() => setNavTitle(t('account.billingRules')))

onLoad(async (q) => {
  user.hydrateFromStorage()
  if (!user.isLoggedIn) {
    uni.showModal({
      title: t('account.needLoginTitle'),
      showCancel: false,
      confirmText: t('account.goLogin'),
      success: (res) => {
        if (res.confirm) navigate('redirect', getLoginPath())
      },
    })
    return
  }
  const serviceId = String(q?.serviceId || storage.get('serviceId', '') || '')
  const pin = String(user.userInfo?.pin || '')
  uni.showToast({ title: t('common.loading'), icon: 'loading', mask: true })
  try {
    const res = await getBillingConfig({
      ...(serviceId ? { serviceId } : {}),
      ...(pin ? { userPin: pin } : {}),
    })
    uni.hideToast()
    if (res.success && res.data && typeof res.data === 'object' && !Array.isArray(res.data)) {
      rule.value = res.data as Record<string, unknown>
    } else {
      uni.showToast({ title: t('billing.loadFail'), icon: 'none', mask: true })
    }
  } catch {
    uni.hideToast()
    uni.showToast({ title: t('billing.loadFail'), icon: 'none', mask: true })
  }
})
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
}
.scroll_view {
  height: 100%;
  background-color: #f8f8f8;
}
.container {
  margin: 34rpx 32rpx;
  padding: 46rpx 32rpx 48rpx;
  background: #fff;
  border-radius: 32rpx;
}
.top-left {
  display: flex;
  justify-content: flex-start;
  align-items: center;
}
.img {
  width: 48rpx;
  height: 48rpx;
  margin-right: 8rpx;
}
.text {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.row {
  margin-top: 32rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 28rpx;
  font-weight: 400;
  color: #333;
  &.sub {
    margin-top: 16rpx;
    color: #afafaf;
    font-size: 24rpx;
  }
}
.des {
  font-weight: 600;
}
.remark {
  font-size: 24rpx;
  font-weight: 400;
  color: #666;
}
.cycle {
  margin-top: 40rpx;
  display: flex;
}
.cyc-img {
  width: 160rpx;
  height: 138rpx;
  margin-right: 24rpx;
  border-radius: 16rpx;
  flex-shrink: 0;
}
.cyc-des {
  display: flex;
  flex-direction: column;
  font-size: 28rpx;
}
.title {
  font-weight: 600;
  color: #282828;
}
.sub-title {
  margin: 12rpx 0;
  color: #666;
}
.content {
  font-size: 24rpx;
  color: #999;
}
.rule-content {
  .tit,
  .tips {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .tips2 {
    margin-top: 0;
  }
  .flex {
    display: flex;
    align-items: center;
  }
  .tips {
    color: #666;
    font-size: 24rpx;
    margin-top: 16rpx;
  }
  .tit {
    color: #333;
    font-size: 28rpx;
    margin-top: 16px;
    &:first-child {
      margin-top: 0;
    }
    .tips {
      margin-top: 0;
    }
  }
}
</style>
