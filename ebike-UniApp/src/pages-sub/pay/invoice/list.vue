<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('pay.invoice') }}</view>
      <view class="btn-primary" @click="goCreate">{{ t('pay.invoiceCreate') }}</view>
    </view>

    <scroll-view
      v-if="list.length"
      class="list"
      scroll-y
      @scrolltolower="loadMore"
    >
      <view v-for="(item, i) in list" :key="i" class="item card" @click="goDetail(item)">
        <view>{{ item.title || item.invoiceNo || item.id || '-' }}</view>
        <view class="sub">{{ stateLabel(item) }} · {{ item.createdAt || item.createTime || '' }}</view>
        <view class="sub" v-if="item.amount != null">
          ¥{{ fenToYuan(item.amount as number) }}
        </view>
        <view class="sub" v-if="tripCount(item) != null">
          {{ t('pay.invoiceTripCount', { n: tripCount(item) }) }}
        </view>
      </view>
      <view class="footer" v-if="loading">{{ t('common.loading') }}</view>
      <view class="footer" v-else-if="!hasMore">{{ t('common.noMore') }}</view>
    </scroll-view>

    <view v-else-if="loading" class="card">{{ t('common.loading') }}</view>
    <view v-else class="card">{{ t('common.empty') }}</view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getInvoiceList } from '@/api/invoice'
import { fenToYuan } from '@/features/pay/usePay'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const loading = ref(false)
const list = ref<Array<Record<string, unknown>>>([])
const pageNum = ref(1)
const pageSize = 20
const total = ref(0)
const hasMore = computed(() => !total.value || list.value.length < total.value)

onShow(() => {
  setNavTitle(t('pay.invoice'))
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
      total.value = Number((data as { count?: number; total?: number }).count ?? (data as { total?: number }).total ?? 0)
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

function goCreate() {
  navigate('to', '/pages-sub/pay/invoice/create')
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
  display: flex;
  flex-direction: column;
  height: 100vh;
  box-sizing: border-box;
}
.title {
  font-weight: 700;
  margin-bottom: 24rpx;
}
.list {
  flex: 1;
  height: 0;
  padding: 0 0 24rpx;
  box-sizing: border-box;
}
.item {
  margin: 0 24rpx 16rpx;
}
.sub {
  color: #888;
  margin-top: 8rpx;
  font-size: 24rpx;
}
.footer {
  text-align: center;
  color: #999;
  padding: 24rpx;
  font-size: 24rpx;
}
</style>
