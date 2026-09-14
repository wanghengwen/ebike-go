<template>
  <view class="page">
    <scroll-view
      v-if="list.length"
      class="list"
      scroll-y
      @scrolltolower="loadMore"
    >
      <view v-for="(item, i) in list" :key="i" class="item" @click="goDetail(item)">
        <view class="item-header">
          <view class="row">
            <text class="datetime">{{ item.createdAt || item.createTime || '--' }}</text>
            <view class="money-wrap">
              <text class="money">{{ fenToYuan(item.amount as number) }}</text>
              <text class="money-unit">元</text>
            </view>
          </view>
          <view class="row sub-row">
            <view class="left">
              <text class="content">{{ item.content || '-' }}</text>
              <text class="type">电子发票</text>
            </view>
            <text class="title">{{ item.title || '-' }}</text>
          </view>
        </view>
        <view class="item-footer">
          <text class="status" :class="{ fail: Number(item.state) === 2 }">{{ stateLabel(item) }}</text>
          <text class="trips" v-if="tripCount(item) != null">
            {{ t('pay.invoiceTripCount', { n: tripCount(item) }) }}
          </text>
        </view>
      </view>
      <view class="footer" v-if="loading">{{ t('common.loading') }}</view>
      <view class="footer" v-else-if="!hasMore">{{ t('common.noMore') }}</view>
    </scroll-view>

    <view v-else-if="loading" class="empty">{{ t('common.loading') }}</view>
    <view v-else class="empty">
      <image v-if="emptyIcon" class="empty-img" :src="emptyIcon" mode="aspectFit" />
      <text class="empty-text">{{ t('pay.invoiceHistoryEmpty') }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getInvoiceList } from '@/api/invoice'
import { fenToYuan } from '@/features/pay/usePay'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const loading = ref(false)
const list = ref<Array<Record<string, unknown>>>([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const hasMore = computed(() => !total.value || list.value.length < total.value)
const emptyIcon = computed(() => getIconCfg('notLogin'))

onShow(() => {
  setNavTitle(t('pay.invoiceHistory'))
  void reload()
})

async function fetchPage(reset: boolean) {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getInvoiceList({ pageNum: pageNum.value, pageSize })
    if (!res.success) return
    const data = res.data as {
      records?: Array<Record<string, unknown>>
      list?: Array<Record<string, unknown>>
      count?: number
      total?: number
    } | Array<Record<string, unknown>>
    const rows = Array.isArray(data) ? data : data?.list || data?.records || []
    if (!Array.isArray(data)) {
      total.value = Number(
        (data as { count?: number; total?: number }).count ??
          (data as { total?: number }).total ??
          0,
      )
    } else if (rows.length < pageSize) {
      total.value = (reset ? 0 : list.value.length) + rows.length
    }
    list.value = reset ? rows : list.value.concat(rows)
    if (!total.value && rows.length < pageSize) {
      total.value = list.value.length
    }
  } finally {
    loading.value = false
  }
}

async function reload() {
  pageNum.value = 1
  list.value = []
  total.value = 0
  await fetchPage(true)
}

function loadMore() {
  if (loading.value || !hasMore.value) {
    if (!hasMore.value && list.value.length) {
      uni.showToast({ title: t('common.noMore'), icon: 'none' })
    }
    return
  }
  pageNum.value += 1
  void fetchPage(false)
}

function stateLabel(item: Record<string, unknown>) {
  const s = Number(item.state)
  if (s === 1) return t('pay.invoiceDone')
  if (s === 2) return t('pay.invoiceFailed')
  return t('pay.invoicePending')
}

function tripCount(item: Record<string, unknown>): number | null {
  const ids = item.orderIds ?? item.order_ids
  if (Array.isArray(ids) && ids.length) return ids.length
  const n = Number(item.orderCount ?? item.tripCount ?? item.orderNum ?? 0)
  return Number.isFinite(n) && n > 0 ? n : null
}

function goDetail(item: Record<string, unknown>) {
  navigate(
    'to',
    `/pages-sub/pay/invoice/detail?params=${encodeURIComponent(JSON.stringify(item))}`,
  )
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #fff;
}
.list {
  height: 100%;
  padding-bottom: 20rpx;
  box-sizing: border-box;
}
.item {
  padding: 32rpx 48rpx 24rpx;
  border-bottom: 2rpx solid #f6f6f6;
}
.row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.datetime {
  font-size: 28rpx;
  color: #333;
}
.money-wrap {
  display: flex;
  align-items: baseline;
}
.money {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.money-unit {
  margin-left: 4rpx;
  font-size: 24rpx;
  color: #333;
}
.sub-row {
  margin-top: 16rpx;
}
.left {
  display: flex;
  align-items: center;
  min-width: 0;
}
.content {
  max-width: 140rpx;
  font-size: 24rpx;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.type {
  margin-left: 15rpx;
  font-size: 24rpx;
  color: #999;
}
.title {
  margin-left: 20rpx;
  font-size: 24rpx;
  color: #999;
  flex: 1;
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.item-footer {
  margin-top: 20rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.status {
  font-size: 24rpx;
  color: #999;
}
.status.fail {
  color: #ff4a4a;
}
.trips {
  margin-right: 70rpx;
  font-size: 24rpx;
  color: #999;
}
.footer {
  text-align: center;
  color: #999;
  padding: 24rpx;
  font-size: 24rpx;
}
.empty {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.empty-img {
  width: 240rpx;
  height: 240rpx;
}
.empty-text {
  margin-top: 24rpx;
  font-size: 28rpx;
  color: #999;
}
</style>
