<template>
  <view class="page">
    <view class="invoice-entry" @click="goInvoice">
      <image v-if="invoiceIcon" class="invoice-entry__icon" :src="invoiceIcon" mode="aspectFit" />
      <text class="invoice-entry__text">{{ t('account.issueInvoice') }}</text>
    </view>

    <scroll-view
      v-if="orders.length"
      class="list"
      scroll-y
      @scrolltolower="loadMore"
    >
      <view
        v-for="(item, index) in orders"
        :key="String(item.id || item.orderId || index)"
        class="card order"
        @click="goDetail(item)"
      >
        <view v-if="item.izComplained === 0" class="badge badge--warn">
          {{ t('pay.invoiceObjectionPending') }}
        </view>
        <view v-else-if="item.izComplained === 1" class="badge badge--ok">
          {{ t('pay.invoiceObjectionDone') }}
        </view>

        <view class="top">
          <view class="top__info">
            <image
              v-if="bikeIcon(item.bikeType)"
              class="bike"
              :src="bikeIcon(item.bikeType)"
              mode="aspectFit"
            />
            <text class="bike-name">{{ bikeLabel(item.bikeType) }}</text>
            <text class="car-id">{{ item.carId || '-' }}</text>
          </view>
          <text :class="['pay-state', { unpaid: isUnpaid(item) }]">{{ payStateText(item) }}</text>
        </view>

        <view class="route">
          <view class="route__row">
            <text class="dot dot--start" />
            <text class="addr">{{ item.startAddress || '--' }}</text>
          </view>
          <view class="route__row">
            <text class="dot dot--end" />
            <text class="addr">{{ item.endAddress || '--' }}</text>
          </view>
        </view>

        <view class="meta">
          <text>{{ t('account.orderMile', { n: formatMile(item.mile) }) }}</text>
          <text>{{ t('account.orderDuration', { time: formatRidingTime(item.ridingTime) }) }}</text>
        </view>

        <view class="bottom">
          <view class="time">
            <text>{{ item.startTime || '' }}</text>
            <text v-if="item.startTime || item.endTime">-</text>
            <text>{{ item.endTime || '' }}</text>
          </view>
          <view class="amount">
            <text class="yen">¥</text>
            <text class="price">{{ formatMoney(item.payCost ?? item.cost, 2) }}</text>
          </view>
        </view>
      </view>
      <view class="footer" v-if="loading">{{ t('common.loading') }}</view>
      <view class="footer" v-else-if="!hasMore && orders.length">{{ t('common.noMore') }}</view>
    </scroll-view>

    <view v-else-if="loading" class="empty">{{ t('common.loading') }}</view>
    <view v-else-if="netError" class="empty">
      <text>{{ t('common.networkError') }}</text>
      <view class="btn-primary retry" @click="retry">{{ t('common.retry') }}</view>
    </view>
    <view v-else class="empty">
      <image v-if="emptyIcon" class="empty-img" :src="emptyIcon" mode="widthFix" />
      <text>{{ t('account.ordersEmpty') }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getOrderList } from '@/api/order'
import { repairBikeIcon } from '@/features/support/repairIcons'
import { formatMile, formatMoney, formatRidingTime, resolveAddress } from '@/shared/format'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { getIconCfg } from '@/shared/tenantSkin'
import { useUserStore } from '@/stores/user'

type OrderRow = Record<string, unknown> & {
  startAddress?: string
  endAddress?: string
}

const { t } = useI18n()
const user = useUserStore()
const loading = ref(false)
const netError = ref(false)
const orders = ref<OrderRow[]>([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const emptyIcon = computed(() => getIconCfg('notLogin'))
const invoiceIcon = computed(() => getIconCfg('walletInvoice') || getIconCfg('invoice'))
const hasMore = computed(() => !total.value || orders.value.length < total.value)

onShow(() => {
  setNavTitle(t('account.orders'))
  void reload()
})

function pin() {
  return String(user.userInfo?.pin || storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
}

function bikeIcon(bikeType: unknown) {
  return repairBikeIcon(bikeType)
}

function bikeLabel(bikeType: unknown) {
  return bikeType ? t('account.bikeNormal') : t('account.bikeElectric')
}

function isUnpaid(item: OrderRow) {
  return Number(item.izPaid) === 3 || Number(item.payState) === 7
}

function payStateText(item: OrderRow) {
  const paid = Number(item.izPaid)
  if (paid === 3 || Number(item.payState) === 7) return t('account.orderUnpaid')
  if (paid === 4) return t('account.orderPaid')
  if (paid === 1) return t('account.orderRiding')
  if (paid === 2) return t('account.orderPaying')
  return ''
}

async function enrichAddresses(rows: OrderRow[]) {
  await Promise.all(
    rows.map(async (item) => {
      if (!item.startAddress) {
        item.startAddress = await resolveAddress(item.startLat, item.startLng)
      }
      if (!item.endAddress) {
        item.endAddress = await resolveAddress(item.endLat, item.endLng)
      }
    }),
  )
}

async function fetchPage(reset: boolean) {
  if (loading.value) return
  if (!reset && total.value && orders.value.length >= total.value) {
    uni.showToast({ title: t('common.noMore'), icon: 'none' })
    return
  }
  loading.value = true
  netError.value = false
  try {
    const res = await getOrderList({
      userPin: pin(),
      pageNum: pageNum.value,
      pageSize,
    })
    if (!res.success) {
      uni.showToast({ title: t('account.ordersLoadFail'), icon: 'none' })
      return
    }
    const data = res.data as {
      count?: number
      total?: number
      list?: OrderRow[]
      records?: OrderRow[]
    }
    total.value = Number(data.count ?? data.total ?? 0)
    const rows = (data.list || data.records || []).map((r) => ({ ...r }))
    await enrichAddresses(rows)
    orders.value = reset ? rows : orders.value.concat(rows)
  } catch {
    netError.value = true
    uni.showToast({ title: t('account.ordersLoadFail'), icon: 'none' })
  } finally {
    loading.value = false
  }
}

async function reload() {
  user.hydrateFromStorage()
  pageNum.value = 1
  orders.value = []
  total.value = 0
  await fetchPage(true)
}

function loadMore() {
  if (loading.value || !hasMore.value) return
  pageNum.value += 1
  void fetchPage(false)
}

function retry() {
  void reload()
}

function goDetail(item: OrderRow) {
  const id = item.id || item.orderId
  if (!id) return
  if (isUnpaid(item)) {
    navigate(
      'to',
      `/pages/pay/pay?orderId=${encodeURIComponent(String(id))}&isNotPollingWxScoreOrder=true`,
    )
    return
  }
  navigate('to', `/pages-sub/pay/cost-detail/cost-detail?orderId=${encodeURIComponent(String(id))}`)
}

function goInvoice() {
  navigate('to', '/pages-sub/pay/invoice/list')
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-top: 16rpx;
  box-sizing: border-box;
}
.invoice-entry {
  margin: 0 24rpx 16rpx;
  background: #fff;
  border-radius: 20rpx;
  height: 96rpx;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  padding: 0 32rpx;
  box-sizing: border-box;
}
.invoice-entry__icon {
  width: 48rpx;
  height: 48rpx;
  margin-right: 16rpx;
}
.invoice-entry__text {
  font-size: 30rpx;
  font-weight: 500;
  color: #333;
}
.list {
  height: calc(100vh - 128rpx);
  box-sizing: border-box;
  padding: 0 24rpx 40rpx;
}
.order {
  position: relative;
  margin: 0 0 16rpx;
  border-radius: 20rpx;
  overflow: hidden;
}
.badge {
  position: absolute;
  top: 0;
  right: 0;
  z-index: 2;
  font-size: 20rpx;
  padding: 6rpx 16rpx;
  border-radius: 0 20rpx 0 16rpx;
}
.badge--warn {
  background: #fff3e6;
  color: #ff922b;
}
.badge--ok {
  background: #e8f8ef;
  color: #2bb673;
}
.top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}
.top__info {
  display: flex;
  align-items: center;
  min-width: 0;
}
.bike {
  width: 48rpx;
  height: 48rpx;
  margin-right: 12rpx;
  flex-shrink: 0;
}
.bike-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  margin-right: 12rpx;
}
.car-id {
  font-size: 24rpx;
  color: #999;
}
.pay-state {
  font-size: 26rpx;
  color: #999;
  flex-shrink: 0;
}
.pay-state.unpaid {
  color: #e34d59;
}
.route__row {
  display: flex;
  align-items: flex-start;
  margin-bottom: 12rpx;
}
.dot {
  width: 14rpx;
  height: 14rpx;
  border-radius: 50%;
  margin-top: 10rpx;
  margin-right: 16rpx;
  flex-shrink: 0;
}
.dot--start {
  background: #63d144;
}
.dot--end {
  background: #ff5936;
}
.addr {
  flex: 1;
  font-size: 26rpx;
  color: #333;
  line-height: 36rpx;
}
.meta {
  display: flex;
  justify-content: space-between;
  margin: 8rpx 0 20rpx;
  font-size: 24rpx;
  color: #666;
}
.bottom {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}
.time {
  font-size: 22rpx;
  color: #999;
  flex: 1;
  margin-right: 16rpx;
}
.amount {
  display: flex;
  align-items: baseline;
  flex-shrink: 0;
}
.yen {
  font-size: 24rpx;
  font-weight: 600;
  color: #333;
  margin-right: 4rpx;
}
.price {
  font-size: 36rpx;
  font-weight: 700;
  color: #333;
}
.footer {
  text-align: center;
  color: #999;
  font-size: 24rpx;
  padding: 24rpx 0;
}
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 160rpx;
  color: #999;
}
.empty-img {
  width: 240rpx;
  margin-bottom: 24rpx;
}
.retry {
  margin-top: 32rpx;
  width: 280rpx;
}
</style>
