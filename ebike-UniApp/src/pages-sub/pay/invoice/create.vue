<template>
  <!-- Legacy issueInvoice：按订单开票 -->
  <view class="page">
    <view v-if="showOrderType || showFilter" class="mask" @click="closePanels" />

    <view class="condition">
      <view class="top_row">
        <view class="top_left" @click="toggleOrderType">
          <text>{{ orderType ? t('pay.invoiceOtherOrders') : t('pay.invoiceCanBill') }}</text>
          <image
            v-if="!showOrderType && arrowBottom"
            class="img_arrow"
            :src="arrowBottom"
            mode="aspectFit"
          />
          <image
            v-else-if="showOrderType && arrowTop"
            class="img_arrow"
            :src="arrowTop"
            mode="aspectFit"
          />
        </view>
        <image
          v-if="filterIcon"
          class="img_filter"
          :src="filterIcon"
          mode="aspectFit"
          @click="toggleFilter"
        />
        <text v-else class="filter_text" @click="toggleFilter">{{ t('pay.invoiceFilter') }}</text>
      </view>

      <view v-if="showOrderType" class="order_type">
        <view
          class="type_item"
          :style="{ color: orderType === 0 ? brandColor : '' }"
          @click="selectOrderType(0)"
        >
          {{ t('pay.invoiceCanBill') }}
        </view>
        <view class="line margin32" />
        <view
          class="type_item"
          :style="{ color: orderType === 1 ? brandColor : '' }"
          @click="selectOrderType(1)"
        >
          {{ t('pay.invoiceOtherOrders') }}
        </view>
      </view>

      <view v-if="showFilter" class="filter">
        <view class="filter_title">{{ t('pay.invoiceDateRange') }}</view>
        <view class="input_container">
          <picker mode="date" start="2010-01-01" end="2030-01-01" @change="onStartDate">
            <view class="input">{{ startDate || t('pay.invoiceStartDate') }}</view>
          </picker>
          <text>-</text>
          <picker mode="date" start="2010-01-01" end="2030-01-01" @change="onEndDate">
            <view class="input">{{ endDate || t('pay.invoiceEndDate') }}</view>
          </picker>
        </view>
        <view class="filter_title">{{ t('pay.invoiceAmountRange') }}</view>
        <view class="input_container">
          <input class="input" type="digit" v-model="minYuan" :placeholder="t('pay.invoiceMinAmount')" />
          <text>-</text>
          <input class="input" type="digit" v-model="maxYuan" :placeholder="t('pay.invoiceMaxAmount')" />
        </view>
        <view class="line" />
        <view class="button_container">
          <button class="button reset" @click="resetFilter">{{ t('pay.invoiceReset') }}</button>
          <button
            class="button confirm"
            :style="{ background: brandColor, color: '#fff' }"
            @click="confirmFilter"
          >
            {{ t('common.confirm') }}
          </button>
        </view>
      </view>
    </view>

    <scroll-view v-if="orders.length" class="scroll" scroll-y>
      <view v-for="(item, index) in orders" :key="String(item.id || index)" class="container">
        <view class="check" @click="toggleOrder(item)">
          <image
            v-if="isSelected(item) && checkIcon"
            class="img_check"
            :src="checkIcon"
            mode="aspectFit"
          />
          <image
            v-else-if="!isSelected(item) && uncheckIcon"
            class="img_check"
            :src="uncheckIcon"
            mode="aspectFit"
          />
          <view v-else class="img_check_fallback" :class="{ on: isSelected(item) }" />
        </view>
        <view class="order">
          <view v-if="!item.izComplained" class="objection">
            {{ t('pay.invoiceObjectionPending') }}
          </view>
          <view v-else-if="item.izComplained === 1" class="objection">
            {{ t('pay.invoiceObjectionDone') }}
          </view>
          <view class="top">
            <view class="top_info">
              <image
                v-if="bikeIcon(item.bikeType)"
                class="img"
                :src="bikeIcon(item.bikeType)"
                mode="aspectFit"
              />
              <text class="top_title">{{ bikeLabel(item.bikeType) }}</text>
              <text class="top_carid">{{ item.carId || '' }}</text>
            </view>
            <text v-if="!item.izPaid" class="unpaid">{{ t('account.orderUnpaid') }}</text>
            <text v-else>{{ t('account.orderPaid') }}</text>
          </view>
          <view class="content">
            <view class="content_item" style="margin-bottom: 24rpx">
              <text class="dot" />
              <text class="text">{{ item.startAddress || '--' }}</text>
            </view>
            <view class="content_item">
              <view class="dot end" />
              <text class="text">{{ item.endAddress || '--' }}</text>
            </view>
          </view>
          <view class="bottom">
            <view class="time">
              <text>{{ item.startTime || '' }}</text>
              <text>-</text>
              <text>{{ item.endTime || '' }}</text>
            </view>
            <view>
              <text class="yen">￥</text>
              <text class="price">{{ fenToYuan(item.payCost as number) }}</text>
            </view>
          </view>
        </view>
      </view>
    </scroll-view>

    <view v-else class="empty">
      <image v-if="emptyIcon" class="notice_img" :src="emptyIcon" mode="aspectFit" />
      <text class="notice_text">{{ t('pay.invoiceNoOrders') }}</text>
    </view>

    <view class="next_container">
      <view class="checkbox" @click="toggleAll">
        <image
          v-if="isChooseAll && checkIcon"
          class="img_check"
          :src="checkIcon"
          mode="aspectFit"
        />
        <image
          v-else-if="!isChooseAll && uncheckIcon"
          class="img_check"
          :src="uncheckIcon"
          mode="aspectFit"
        />
        <view v-else class="img_check_fallback" :class="{ on: isChooseAll }" />
        <text class="all">{{ t('pay.invoiceSelectAll') }}</text>
      </view>
      <text class="total">
        <text :style="{ color: brandColor }">{{ selectedIds.length }}</text>
        {{ t('pay.invoiceTripUnit') }}，{{ t('pay.invoiceTotalPrefix') }}
        <text :style="{ color: brandColor }">{{ totalMoney }}</text>
        {{ t('pay.invoiceYuan') }}
      </text>
      <button
        class="next"
        :style="nextBtnStyle"
        :disabled="!canNext"
        @click="goNext"
      >
        {{ t('account.cancelNext') }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onHide, onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getInvoicedOrders } from '@/api/invoice'
import { fenToYuan } from '@/features/pay/usePay'
import { resolveAddress } from '@/shared/format'
import { getBrandColor, getButtonDisabledColor, getButtonWhiteColor } from '@/shared/config'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const user = useUserStore()

const orders = ref<Array<Record<string, unknown>>>([])
const selectedIds = ref<Array<string | number>>([])
const orderType = ref(0)
const showOrderType = ref(false)
const showFilter = ref(false)
const startDate = ref<string | null>(null)
const endDate = ref<string | null>(null)
const minYuan = ref<string | null>(null)
const maxYuan = ref<string | null>(null)
const isChooseAll = ref(false)
const totalMoney = ref(0)

const brandColor = computed(() => getBrandColor())
const arrowBottom = computed(() => getIconCfg('invoiceArrowBottom'))
const arrowTop = computed(() => getIconCfg('invoiceArrowTop'))
const filterIcon = computed(() => getIconCfg('invoiceFilter'))
const checkIcon = computed(() => getIconCfg('checkbox') || getIconCfg('checked_square_round'))
const uncheckIcon = computed(() => getIconCfg('invoiceUncheck'))
const emptyIcon = computed(() => getIconCfg('notLogin'))
const canNext = computed(() => selectedIds.value.length > 0 && orderType.value === 0)
const nextBtnStyle = computed(() => {
  const enabled = brandColor.value
  const disabled = getButtonDisabledColor()
  const white = getButtonWhiteColor()
  const bg = canNext.value ? enabled : disabled
  return `background:${bg};color:${white};border:none;`
})

onShow(() => setNavTitle(t('pay.invoiceByOrder')))

onLoad(() => {
  void loadOrders()
})

onHide(() => {
  orderType.value = 0
  selectedIds.value = []
  resetFilterFields()
  recalcMoney()
  isChooseAll.value = false
})

function bikeIcon(bikeType: unknown) {
  // Legacy issueInvoice: bikeType truthy → electricBicycle
  return bikeType ? getIconCfg('electricBicycle') : getIconCfg('bicycle')
}

function bikeLabel(bikeType: unknown) {
  // Legacy text: bikeType ? 单车 : 电单车
  return bikeType ? t('account.bikeNormal') : t('account.bikeElectric')
}

function isSelected(item: Record<string, unknown>) {
  return selectedIds.value.includes(item.id as string | number)
}

function transTimeStamp(value: string | null | undefined) {
  if (!value) return null
  const ms = +new Date(value)
  return ms || null
}

/** 未选筛选日期时，默认查近 3 个月可开票订单 */
function defaultMonthRange(): [number, number] {
  const end = Date.now()
  const start = end - 90 * 24 * 60 * 60 * 1000
  return [start, end]
}

function resolveStartTime(): [number | null, number | null] {
  const hasStart = Boolean(startDate.value)
  const hasEnd = Boolean(endDate.value)
  if (!hasStart && !hasEnd) return defaultMonthRange()
  return [transTimeStamp(startDate.value), transTimeStamp(endDate.value)]
}

function resetFilterFields() {
  startDate.value = null
  endDate.value = null
  minYuan.value = null
  maxYuan.value = null
}

async function loadOrders() {
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const pin =
      user.userInfo.pin ||
      (user.userInfo as Record<string, unknown>).userPin ||
      ''
    const params = {
      userPin: pin,
      izCanInvoiced: !Boolean(orderType.value),
      startTime: resolveStartTime(),
      minCost: minYuan.value ? Math.round(Number(minYuan.value) * 100) : null,
      maxCost: maxYuan.value ? Math.round(Number(maxYuan.value) * 100) : null,
    }
    const res = await getInvoicedOrders(params)
    if (!res.success) {
      orders.value = []
      uni.showToast({ title: t('pay.invoiceOrdersFail'), icon: 'none' })
      return
    }
    const data = res.data
    const list = (Array.isArray(data) ? data : []) as Array<Record<string, unknown>>
    orders.value = list
    selectedIds.value = []
    isChooseAll.value = false
    recalcMoney()
    // Legacy: reverse geocode start/end
    void Promise.all(
      list.map(async (item) => {
        item.startAddress = await resolveAddress(item.startLat, item.startLng)
        item.endAddress = await resolveAddress(item.endLat, item.endLng)
      }),
    ).then(() => {
      orders.value = [...orders.value]
    })
  } finally {
    uni.hideLoading()
  }
}

function closePanels() {
  showOrderType.value = false
  showFilter.value = false
}

function toggleOrderType() {
  if (showFilter.value) return
  showOrderType.value = !showOrderType.value
}

function toggleFilter() {
  if (showOrderType.value) return
  showFilter.value = !showFilter.value
}

function selectOrderType(type: number) {
  orderType.value = type
  showOrderType.value = false
  void loadOrders()
}

function onStartDate(e: { detail: { value: string } }) {
  const v = e.detail.value
  if (endDate.value) {
    const end = transTimeStamp(endDate.value)
    const now = transTimeStamp(v)
    if (end != null && now != null && end < now) {
      uni.showToast({ title: t('pay.invoiceDateInvalid'), icon: 'none' })
      return
    }
  }
  startDate.value = v
}

function onEndDate(e: { detail: { value: string } }) {
  const v = e.detail.value
  if (startDate.value) {
    const start = transTimeStamp(startDate.value)
    const now = transTimeStamp(v)
    if (start != null && now != null && start > now) {
      uni.showToast({ title: t('pay.invoiceDateInvalid'), icon: 'none' })
      return
    }
  }
  endDate.value = v
}

function resetFilter() {
  // Legacy reset: only clear fields, no reload
  resetFilterFields()
}

function confirmFilter() {
  showFilter.value = false
  void loadOrders()
}

function toggleOrder(item: Record<string, unknown>) {
  const id = item.id as string | number
  const idx = selectedIds.value.indexOf(id)
  if (idx > -1) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
  isChooseAll.value = selectedIds.value.length === orders.value.length && orders.value.length > 0
  recalcMoney()
}

function toggleAll() {
  isChooseAll.value = !isChooseAll.value
  if (isChooseAll.value) {
    orders.value.forEach((item) => {
      const id = item.id as string | number
      if (!selectedIds.value.includes(id)) selectedIds.value.push(id)
    })
  } else {
    selectedIds.value = []
  }
  recalcMoney()
}

function recalcMoney() {
  let fen = 0
  orders.value.forEach((item) => {
    if (selectedIds.value.includes(item.id as string | number)) {
      fen += Number(item.payCost || 0)
    }
  })
  totalMoney.value = fen / 100
}

function goNext() {
  if (!canNext.value) return
  const params = {
    orderIds: selectedIds.value,
    money: totalMoney.value,
  }
  navigate(
    'to',
    `/pages-sub/pay/invoice/apply?params=${encodeURIComponent(JSON.stringify(params))}`,
  )
}
</script>

<style scoped lang="scss">
.page {
  position: relative;
  width: 100vw;
  height: 100vh;
  background: #f8f8f8;
}
.mask {
  position: absolute;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  z-index: 100;
}
.condition {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 200;
}
.top_row {
  display: flex;
  align-items: center;
  padding: 24rpx 32rpx;
  background: #fff;
}
.top_left {
  flex: 1;
  font-size: 28rpx;
  font-weight: 500;
  color: #000;
  display: flex;
  align-items: center;
  gap: 8rpx;
}
.img_arrow {
  width: 24rpx;
  height: 24rpx;
}
.img_filter {
  width: 48rpx;
  height: 48rpx;
}
.filter_text {
  font-size: 26rpx;
  color: #666;
}
.order_type {
  padding: 32rpx 32rpx 48rpx;
  font-size: 28rpx;
  font-weight: 500;
  border-radius: 0 0 32rpx 32rpx;
  background: #fff;
  color: #666;
}
.type_item {
  font-size: 28rpx;
}
.line {
  height: 2rpx;
  margin: 48rpx 0 46rpx;
  background: #f6f6f6;
}
.margin32 {
  margin: 32rpx 0 30rpx;
}
.filter {
  padding: 32rpx;
  border-radius: 0 0 32rpx 32rpx;
  background: #fff;
}
.filter_title {
  font-size: 28rpx;
  font-weight: 500;
  color: #000;
}
.input_container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 32rpx;
  margin-bottom: 48rpx;
}
.input {
  width: 310rpx;
  height: 80rpx;
  line-height: 80rpx;
  background: #f6f6f6;
  border-radius: 32rpx;
  text-align: center;
  color: #999;
  font-size: 28rpx;
}
.button_container {
  display: flex;
  gap: 18rpx;
}
.button {
  flex: 1;
  height: 96rpx;
  line-height: 96rpx;
  border-radius: 32rpx;
  font-size: 32rpx;
  font-weight: 500;
  &.reset {
    border: 2rpx solid #ccc;
    background: #fff;
    color: #333;
  }
}
.scroll {
  box-sizing: border-box;
  height: calc(100vh - 220rpx);
  padding: 128rpx 0 220rpx;
}
.container {
  position: relative;
  display: flex;
  align-items: center;
  margin: 0 24rpx 32rpx;
  padding: 32rpx 32rpx 32rpx 16rpx;
  border-radius: 32rpx;
  background: #fff;
}
.check {
  margin-right: 16rpx;
}
.img_check {
  width: 32rpx;
  height: 32rpx;
}
.img_check_fallback {
  width: 32rpx;
  height: 32rpx;
  border-radius: 50%;
  border: 2rpx solid #ccc;
  box-sizing: border-box;
  &.on {
    background: #3aa0e8;
    border-color: #3aa0e8;
  }
}
.order {
  flex: 1;
  min-width: 0;
}
.objection {
  position: absolute;
  top: 0;
  right: 0;
  padding: 6rpx 32rpx;
  background: rgba(255, 171, 44, 0.2);
  border-radius: 0 32rpx 0 32rpx;
  font-size: 24rpx;
  font-weight: 500;
  color: #ffab2c;
}
.top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 28rpx;
  color: #999;
}
.top_info {
  display: flex;
  align-items: flex-end;
}
.img {
  width: 48rpx;
  height: 48rpx;
}
.top_title {
  margin: 0 8rpx;
  font-size: 32rpx;
  font-weight: 500;
  color: #000;
}
.top_carid {
  line-height: 25rpx;
  font-size: 20rpx;
  color: #999;
}
.unpaid {
  color: #ff3434;
}
.content {
  margin: 24rpx 0;
}
.content_item {
  display: flex;
  align-items: center;
}
.dot {
  margin-right: 8rpx;
  width: 16rpx;
  height: 16rpx;
  background: #10d60e;
  border-radius: 50%;
  flex-shrink: 0;
  &.end {
    background: #fd2d30;
  }
}
.text {
  font-size: 24rpx;
  color: #333;
}
.bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16rpx;
}
.time {
  font-size: 24rpx;
  color: #999;
}
.yen {
  font-size: 24rpx;
  font-weight: 500;
  color: #333;
}
.price {
  font-size: 40rpx;
  font-weight: bold;
  color: #333;
}
.empty {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  width: 100%;
  background: #fff;
  padding-top: 128rpx;
  box-sizing: border-box;
}
.notice_img {
  width: 300rpx;
  height: 300rpx;
}
.notice_text {
  margin-top: 48rpx;
  font-size: 28rpx;
  color: #666;
}
.next_container {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 32rpx 48rpx 100rpx;
  background: #fff;
  font-size: 28rpx;
  z-index: 50;
}
.checkbox {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 130rpx;
}
.all {
  margin-left: 8rpx;
  margin-right: 10rpx;
  color: #999;
  white-space: nowrap;
}
.total {
  flex: 1;
  margin-right: 24rpx;
  text-align: right;
  color: #666;
}
.next {
  width: 240rpx;
  height: 96rpx;
  line-height: 96rpx;
  border-radius: 32rpx;
  font-size: 32rpx;
  padding: 0;
}
button::after {
  display: none;
}
</style>
