<template>
  <view class="page">
    <view class="card">
      <view class="label">{{ t('account.walletBalance') }}</view>
      <view class="balance">¥{{ balance }}</view>
    </view>
    <scroll-view scroll-x class="tabs-scroll">
      <view class="tabs">
        <view
          v-for="(label, i) in tabLabels"
          :key="i"
          class="tab"
          :class="{ on: tab === i }"
          @click="switchTab(i)"
        >
          {{ label }}
        </view>
      </view>
    </scroll-view>
    <scroll-view scroll-y class="list-scroll" @scrolltolower="onScrollToLower">
      <view class="card list-card">
        <view v-if="loading && !records.length">{{ t('common.loading') }}</view>
        <view v-else-if="!records.length && !loading">{{ emptyTip }}</view>
        <view v-for="(item, i) in records" :key="i" class="item">
          <view class="main">
            <view class="title-block">
              <text class="title">{{ recordTitle(item) }}</text>
              <text v-if="remindText(item)" class="remind">{{ remindText(item) }}</text>
            </view>
            <text class="price" :class="{ out: isOut(item) }">{{ formatAmount(item) }}</text>
          </view>
          <view class="sub">
            <text>{{ recordTime(item) }}</text>
            <text v-if="item.channel">{{ String(item.channel) }}</text>
          </view>
        </view>
        <view v-if="records.length" class="bot-line">————— {{ t('common.noMore') }} —————</view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import {
  getWalletInfo,
  getConsumptionRecord,
  getUserBuyRecord,
  depositRecord,
  ridingCardRecord,
} from '@/api/pay'
import { fenToYuan } from '@/features/pay/usePay'
import { useUserStore } from '@/stores/user'
import { setNavTitle } from '@/shared/navigate'
import { logger } from '@/shared/logger'

/** Legacy chargeList: earliest year is 2017 */
const EARLIEST_YEAR = 2017
const PAGE_CHUNK = 10

const { t } = useI18n()
const user = useUserStore()
const balance = ref('0.00')
const loading = ref(false)
const tab = ref(0)
const records = ref<Array<Record<string, unknown>>>([])
const yearStart = ref(0)
const yearEnd = ref(0)
/** Total rows fetched for current tab (all years / full list) */
const total = ref(0)
const pageNum = ref(0)
const splitData = ref<Array<Array<Record<string, unknown>>>>([])

const tabLabels = computed(() => [
  t('account.consumeRecords'),
  t('account.buyRecords'),
  t('account.depositRecords'),
  t('account.ridingCardRecords'),
])

const emptyTip = computed(() => {
  if (tab.value === 0) return t('account.walletEmptyConsume')
  if (tab.value === 1) return t('account.walletEmptyBuy')
  if (tab.value === 2) return t('account.walletEmptyDeposit')
  return t('account.walletEmptyCard')
})

const usesYearRange = computed(() => tab.value === 0 || tab.value === 1)

onShow(() => setNavTitle(t('account.walletRecords')))

function resetYearRange() {
  const y = new Date().getFullYear()
  yearStart.value = new Date(y, 0, 1).getTime()
  yearEnd.value = new Date(y, 11, 31, 23, 59, 59).getTime()
}

function isOut(item: Record<string, unknown>) {
  const n = Number(item.amount ?? item.changeAmount ?? item.money ?? 0)
  return n < 0 || item.direction === 'out' || item.type === 'consume'
}

function formatAmount(item: Record<string, unknown>) {
  const raw = item.amount ?? item.changeAmount ?? item.money ?? item.total_fee
  if (raw == null) return '-'
  const n = Number(raw)
  if (Math.abs(n) >= 100 && Number.isInteger(n)) return `¥${fenToYuan(n)}`
  return `¥${Number(n).toFixed(2)}`
}

function fenPart(raw: unknown): string | null {
  if (raw == null || raw === '') return null
  const n = Number(raw)
  if (!Number.isFinite(n) || n === 0) return null
  return Math.abs(n) >= 100 && Number.isInteger(n) ? fenToYuan(n) : Number(n).toFixed(2)
}

function recordTitle(item: Record<string, unknown>) {
  const name = item.name != null && item.name !== '' ? String(item.name) : ''
  const type = item.type != null && item.type !== '' ? String(item.type) : ''
  if (name && type) return `${name}: (${type})`
  if (type) return type
  if (name) return name
  return String(item.remark || item.typeName || item.bizType || tabLabels.value[tab.value] || '-')
}

function remindText(item: Record<string, unknown>) {
  const recharge = fenPart(item.recharge_amount ?? item.rechargeAmount)
  const present = fenPart(item.present_amount ?? item.presentAmount)
  if (!recharge && !present) return ''
  const parts: string[] = []
  if (recharge) parts.push(t('account.walletInclRecharge', { m: recharge }))
  if (recharge && present) parts.push('+')
  if (present) parts.push(t('account.walletInclPresent', { m: present }))
  return parts.join('')
}

function recordTime(item: Record<string, unknown>) {
  return String(item.paid_at || item.paidAt || item.createTime || item.createdAt || item.time || '')
}

function normalizeRecords(data: unknown): Array<Record<string, unknown>> {
  if (Array.isArray(data)) return data as Array<Record<string, unknown>>
  if (data && typeof data === 'object') {
    const obj = data as { records?: unknown; list?: unknown; rows?: unknown }
    if (Array.isArray(obj.records)) return obj.records as Array<Record<string, unknown>>
    if (Array.isArray(obj.list)) return obj.list as Array<Record<string, unknown>>
    if (Array.isArray(obj.rows)) return obj.rows as Array<Record<string, unknown>>
  }
  return []
}

function chunkAndShow(rows: Array<Record<string, unknown>>, append: boolean) {
  if (append) {
    // Previous-year batch: concat onto visible list; rebuild chunks from full set
    const merged = records.value.concat(rows)
    total.value += rows.length
    splitData.value = []
    for (let i = 0; i < merged.length; i += PAGE_CHUNK) {
      splitData.value.push(merged.slice(i, i + PAGE_CHUNK))
    }
    pageNum.value = 0
    records.value = splitData.value[0] ? [...splitData.value[0]] : []
    // Reveal all already-fetched pages so user keeps prior years visible
    for (let p = 1; p < splitData.value.length; p++) {
      records.value = records.value.concat(splitData.value[p])
      pageNum.value = p
    }
    return
  }
  total.value = rows.length
  pageNum.value = 0
  splitData.value = []
  for (let i = 0; i < rows.length; i += PAGE_CHUNK) {
    splitData.value.push(rows.slice(i, i + PAGE_CHUNK))
  }
  records.value = splitData.value[0] ? [...splitData.value[0]] : []
}

async function query(appendYear = false) {
  loading.value = true
  const pin = user.userInfo.pin
  const payload: Record<string, unknown> = { pin }
  if (usesYearRange.value) {
    payload.start_time = yearStart.value
    payload.end_time = yearEnd.value
  }
  try {
    let res
    if (tab.value === 0) res = await getConsumptionRecord(payload)
    else if (tab.value === 1) res = await getUserBuyRecord(payload)
    else if (tab.value === 2) res = await depositRecord(payload)
    else res = await ridingCardRecord(payload)
    const rows = res.success ? normalizeRecords(res.data) : []
    chunkAndShow(rows, appendYear)
  } catch (e) {
    logger.warn('wallet records soft fail', e)
    if (!appendYear) {
      records.value = []
      total.value = 0
      splitData.value = []
    }
  } finally {
    loading.value = false
  }
}

function onScrollToLower() {
  if (loading.value) return
  // Next client chunk
  if (records.value.length < total.value && splitData.value.length) {
    const next = pageNum.value + 1
    if (splitData.value[next]) {
      pageNum.value = next
      records.value = records.value.concat(splitData.value[next])
      return
    }
  }
  if (records.value.length >= total.value && usesYearRange.value) {
    const y = new Date(yearStart.value).getFullYear()
    if (y <= EARLIEST_YEAR) {
      uni.showToast({ title: t('common.noMore'), icon: 'none' })
      return
    }
    const prev = y - 1
    yearStart.value = new Date(prev, 0, 1).getTime()
    yearEnd.value = new Date(prev, 11, 31, 23, 59, 59).getTime()
    void query(true)
    return
  }
  if (records.value.length >= total.value) {
    uni.showToast({ title: t('common.noMore'), icon: 'none' })
  }
}

function switchTab(i: number) {
  tab.value = i
  resetYearRange()
  total.value = 0
  records.value = []
  splitData.value = []
  pageNum.value = 0
  void query(false)
}

onMounted(async () => {
  resetYearRange()
  const wallet = await getWalletInfo()
  if (wallet.success && wallet.data) {
    const data = wallet.data as { balance?: string | number; recharge?: number; present?: number }
    if (data.balance != null) {
      const b = Number(data.balance)
      balance.value = b > 1000 || Number.isInteger(b) ? fenToYuan(b) : String(data.balance)
    } else {
      balance.value = fenToYuan(Number(data.recharge || 0) + Number(data.present || 0))
    }
  }
  await query(false)
})
</script>

<style scoped lang="scss">
.page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  box-sizing: border-box;
}
.label {
  color: #888;
}
.balance {
  font-size: 56rpx;
  font-weight: 700;
  margin: 16rpx 0 8rpx;
}
.tabs-scroll {
  width: 100%;
  white-space: nowrap;
  flex-shrink: 0;
}
.tabs {
  display: inline-flex;
  gap: 8rpx;
  padding: 0 24rpx 0;
  min-width: 100%;
}
.tab {
  display: inline-block;
  padding: 20rpx 24rpx;
  color: #888;
  font-size: 26rpx;
}
.tab.on {
  color: #3aa0e8;
  font-weight: 600;
  border-bottom: 4rpx solid #3aa0e8;
}
.list-scroll {
  flex: 1;
  height: 0;
}
.list-card {
  margin-bottom: 24rpx;
}
.item {
  padding: 24rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16rpx;
}
.title-block {
  flex: 1;
  min-width: 0;
}
.title {
  font-size: 30rpx;
  color: #111;
  font-weight: 500;
}
.remind {
  display: block;
  margin-top: 8rpx;
  font-size: 20rpx;
  color: #999;
}
.price {
  font-size: 30rpx;
  font-weight: 700;
  color: #333;
  flex-shrink: 0;
}
.out {
  color: #e34d59;
}
.sub {
  color: #888;
  margin-top: 16rpx;
  font-size: 24rpx;
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
}
.bot-line {
  text-align: center;
  color: #bbb;
  font-size: 24rpx;
  padding: 24rpx 0 8rpx;
}
</style>
