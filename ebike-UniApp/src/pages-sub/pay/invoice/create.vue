<template>
  <view class="page">
    <view class="card filter-card">
      <view class="filter-top">
        <view class="tabs">
          <text class="tab" :class="{ on: orderType === 0 }" @click="setOrderType(0)">
            {{ t('pay.invoiceCanBill') }}
          </text>
          <text class="tab" :class="{ on: orderType === 1 }" @click="setOrderType(1)">
            {{ t('pay.invoiceOtherOrders') }}
          </text>
        </view>
        <text class="filter-btn" @click="showFilter = !showFilter">{{ t('pay.invoiceFilter') }}</text>
      </view>
      <view v-if="showFilter" class="filter-body">
        <view class="label">{{ t('pay.invoiceDateRange') }}</view>
        <view class="range">
          <picker mode="date" start="2010-01-01" end="2030-12-31" @change="onStartDate">
            <view class="picker">{{ startDate || t('pay.invoiceStartDate') }}</view>
          </picker>
          <text>-</text>
          <picker mode="date" start="2010-01-01" end="2030-12-31" @change="onEndDate">
            <view class="picker">{{ endDate || t('pay.invoiceEndDate') }}</view>
          </picker>
        </view>
        <view class="label">{{ t('pay.invoiceAmountRange') }}</view>
        <view class="range">
          <input class="input mini" type="digit" v-model="minYuan" :placeholder="t('pay.invoiceMinAmount')" />
          <text>-</text>
          <input class="input mini" type="digit" v-model="maxYuan" :placeholder="t('pay.invoiceMaxAmount')" />
        </view>
        <view class="filter-actions">
          <view class="btn-ghost" @click="resetFilter">{{ t('pay.invoiceReset') }}</view>
          <view class="btn-primary sm" @click="applyFilter">{{ t('common.confirm') }}</view>
        </view>
      </view>
    </view>

    <view class="card">
      <view class="label row-between">
        <text>{{ t('pay.selectOrders') }}</text>
        <text class="link" v-if="orders.length" @click="toggleAll">{{ t('pay.invoiceSelectAll') }}</text>
      </view>
      <view v-if="loadingOrders">{{ t('common.loading') }}</view>
      <view v-else-if="!orders.length" class="empty">{{ t('common.empty') }}</view>
      <view
        v-for="item in orders"
        :key="String(item.id)"
        class="order"
        :class="{ on: selectedIds.includes(String(item.id)), disabled: !canSelect(item) }"
        @click="toggleOrder(item)"
      >
        <view class="order__main">
          <view class="order__top">
            <text class="bike">{{ bikeLabel(item.bikeType) }}</text>
            <text class="car-id">{{ item.carId || item.id }}</text>
            <text class="muted" v-if="!item.izPaid">{{ t('account.orderUnpaid') }}</text>
          </view>
          <view class="sub muted" v-if="complainText(item)">{{ complainText(item) }}</view>
          <view class="route" v-if="item.startAddress || item.endAddress">
            <view class="route__row" v-if="item.startAddress">
              <text class="dot start" />
              <text>{{ item.startAddress }}</text>
            </view>
            <view class="route__row" v-if="item.endAddress">
              <text class="dot end" />
              <text>{{ item.endAddress }}</text>
            </view>
          </view>
          <view class="time" v-if="item.startTime || item.endTime">
            <text>{{ item.startTime || '' }}</text>
            <text v-if="item.startTime || item.endTime"> - </text>
            <text>{{ item.endTime || '' }}</text>
          </view>
        </view>
        <view class="right">
          <view class="sub">¥{{ fenToYuan(item.payCost as number) }}</view>
        </view>
      </view>
      <view class="sum" v-if="selectedIds.length">
        {{ t('pay.invoiceSelectedSum', { n: selectedIds.length, amount: selectedSum }) }}
      </view>
    </view>

    <view class="card" v-if="orderType === 0">
      <view class="type-row">
        <view class="type" :class="{ on: form.type === 0 }" @click="form.type = 0">{{ t('pay.invoicePersonal') }}</view>
        <view class="type" :class="{ on: form.type === 1 }" @click="form.type = 1">{{ t('pay.invoiceCompany') }}</view>
      </view>
      <input class="input" v-model="form.title" :placeholder="t('pay.invoiceTitle')" />
      <input v-if="form.type === 1" class="input" v-model="form.companyEin" :placeholder="t('pay.invoiceTaxNo')" />
      <input v-if="form.type === 1" class="input" v-model="form.companyAddress" :placeholder="t('pay.invoiceCompanyAddress')" />
      <input v-if="form.type === 1" class="input" v-model="form.companyPhone" :placeholder="t('pay.invoiceCompanyPhone')" />
      <input v-if="form.type === 1" class="input" v-model="form.bank" :placeholder="t('pay.invoiceBank')" />
      <input v-if="form.type === 1" class="input" v-model="form.bankAccount" :placeholder="t('pay.invoiceBankAccount')" />
      <input class="input" v-model="form.email" :placeholder="t('pay.invoiceEmail')" />
      <view class="btn-primary" @click="onSubmit">{{ t('common.confirm') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { createInvoice, getInvoicedOrders } from '@/api/invoice'
import { fenToYuan } from '@/features/pay/usePay'
import { useUserStore } from '@/stores/user'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const user = useUserStore()
const loadingOrders = ref(false)
const orders = ref<Array<Record<string, unknown>>>([])
const selectedIds = ref<string[]>([])
const orderType = ref(0) // 0 can invoice, 1 other
const showFilter = ref(false)
const startDate = ref('')
const endDate = ref('')
const minYuan = ref('')
const maxYuan = ref('')

const form = reactive({
  type: 0,
  title: '',
  companyEin: '',
  companyAddress: '',
  companyPhone: '',
  bank: '',
  bankAccount: '',
  email: '',
})

const selectedSum = computed(() => {
  const fen = orders.value
    .filter((o) => selectedIds.value.includes(String(o.id)))
    .reduce((s, o) => s + Number(o.payCost || 0), 0)
  return fenToYuan(fen)
})

onShow(() => setNavTitle(t('pay.invoiceCreate')))

onLoad((q) => {
  try {
    if (q?.orderIds) {
      const ids = JSON.parse(decodeURIComponent(String(q.orderIds)))
      if (Array.isArray(ids)) selectedIds.value = ids.map(String)
    }
  } catch {
    /* ignore */
  }
})

onMounted(() => {
  void loadOrders()
})

function dayStart(d: string) {
  return new Date(`${d}T00:00:00`).getTime()
}
function dayEnd(d: string) {
  return new Date(`${d}T23:59:59`).getTime()
}

async function loadOrders() {
  loadingOrders.value = true
  selectedIds.value = []
  const now = Date.now()
  const yearAgo = now - 365 * 24 * 60 * 60 * 1000
  const start = startDate.value ? dayStart(startDate.value) : yearAgo
  const end = endDate.value ? dayEnd(endDate.value) : now
  const minCost = minYuan.value ? Math.round(Number(minYuan.value) * 100) : null
  const maxCost = maxYuan.value ? Math.round(Number(maxYuan.value) * 100) : null
  const res = await getInvoicedOrders({
    userPin: user.userInfo.pin,
    izCanInvoiced: orderType.value === 0,
    startTime: [start, end],
    minCost,
    maxCost,
  })
  loadingOrders.value = false
  const data = res.data as { records?: Array<Record<string, unknown>>; list?: Array<Record<string, unknown>> } | Array<Record<string, unknown>>
  orders.value = Array.isArray(data) ? data : data?.list || data?.records || []
}

function setOrderType(v: number) {
  if (orderType.value === v) return
  orderType.value = v
  void loadOrders()
}

function onStartDate(e: { detail: { value: string } }) {
  startDate.value = e.detail.value
}
function onEndDate(e: { detail: { value: string } }) {
  endDate.value = e.detail.value
}

function resetFilter() {
  startDate.value = ''
  endDate.value = ''
  minYuan.value = ''
  maxYuan.value = ''
  void loadOrders()
}

function applyFilter() {
  showFilter.value = false
  void loadOrders()
}

function canSelect(item: Record<string, unknown>) {
  if (orderType.value !== 0) return false
  // Legacy blocks unpaid / complaint-pending
  if (!item.izPaid) return false
  if (item.izComplained === 0) return false
  return true
}

function complainText(item: Record<string, unknown>) {
  if (item.izComplained === 0) return t('pay.invoiceObjectionPending')
  if (item.izComplained === 1) return t('pay.invoiceObjectionDone')
  return ''
}

function bikeLabel(bikeType: unknown) {
  if (bikeType == null || bikeType === '') return ''
  return bikeType ? t('account.bikeNormal') : t('account.bikeElectric')
}

function toggleOrder(item: Record<string, unknown>) {
  if (!canSelect(item)) return
  const id = String(item.id)
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

function toggleAll() {
  const ids = orders.value.filter(canSelect).map((o) => String(o.id))
  if (selectedIds.value.length === ids.length) selectedIds.value = []
  else selectedIds.value = ids
}

async function onSubmit() {
  if (!selectedIds.value.length) {
    uni.showToast({ title: t('pay.selectOrders'), icon: 'none' })
    return
  }
  if (!form.title.trim() || !form.email.trim()) {
    uni.showToast({ title: t('pay.invoiceTitle'), icon: 'none' })
    return
  }
  if (form.type === 1 && !form.companyEin.trim()) {
    uni.showToast({ title: t('pay.invoiceTaxNo'), icon: 'none' })
    return
  }
  const res = await createInvoice({
    ...form,
    content: '运输服务',
    orderIds: selectedIds.value,
    serviceId: storage.get('serviceId', ''),
    userPin: user.userInfo.pin,
  })
  if (res.success) {
    uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
    setTimeout(() => navigate('back'), 500)
  }
}
</script>

<style scoped lang="scss">
.filter-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.tabs {
  display: flex;
  gap: 24rpx;
}
.tab {
  color: #888;
  &.on {
    color: #3aa0e8;
    font-weight: 600;
  }
}
.filter-btn {
  color: #3aa0e8;
  font-size: 26rpx;
}
.filter-body {
  margin-top: 20rpx;
  padding-top: 16rpx;
  border-top: 1px solid #f0f0f0;
}
.range {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 16rpx;
}
.picker,
.input.mini {
  flex: 1;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 20rpx;
  text-align: center;
}
.filter-actions {
  display: flex;
  gap: 16rpx;
}
.btn-primary.sm {
  flex: 1;
  text-align: center;
  padding: 20rpx;
}
.filter-actions .btn-ghost {
  flex: 1;
  text-align: center;
  padding: 20rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
}
.label {
  color: #666;
  margin-bottom: 12rpx;
}
.row-between {
  display: flex;
  justify-content: space-between;
}
.link {
  color: #3aa0e8;
  font-size: 26rpx;
}
.order {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  padding: 20rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.order.on {
  color: #3aa0e8;
}
.order.disabled {
  opacity: 0.45;
}
.order__main {
  flex: 1;
  min-width: 0;
}
.order__top {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12rpx;
}
.bike {
  font-weight: 600;
}
.car-id {
  color: #666;
  font-size: 26rpx;
}
.route {
  margin-top: 12rpx;
}
.route__row {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #666;
}
.dot {
  width: 12rpx;
  height: 12rpx;
  border-radius: 50%;
  margin-top: 10rpx;
  flex-shrink: 0;
}
.dot.start {
  background: #3aa0e8;
}
.dot.end {
  background: #fd2d30;
}
.time {
  margin-top: 10rpx;
  font-size: 22rpx;
  color: #999;
}
.right {
  text-align: right;
  flex-shrink: 0;
}
.sub {
  font-weight: 600;
}
.muted {
  color: #999;
  font-size: 22rpx;
  font-weight: 400;
  margin-top: 6rpx;
}
.sum {
  margin-top: 16rpx;
  color: #666;
  font-size: 26rpx;
}
.type-row {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
}
.type {
  flex: 1;
  text-align: center;
  padding: 16rpx;
  background: #f5f6f8;
  border-radius: 8rpx;
}
.type.on {
  background: #3aa0e8;
  color: #fff;
}
.input {
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.empty {
  color: #999;
}
</style>
