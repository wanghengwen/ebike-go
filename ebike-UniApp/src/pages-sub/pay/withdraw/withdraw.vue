<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('account.withdraw') }}</view>
      <view class="balance-label">{{ t('account.withdrawable') }}</view>
      <view class="balance">¥{{ withdrawBalance }}</view>
      <view class="threshold" v-if="thresholdText">{{ thresholdText }}</view>
      <view class="hint">{{ t('account.withdrawHint') }}</view>
      <view
        class="btn-primary"
        :class="{ disabled: !canWithdraw }"
        @click="onSubmit"
      >
        {{ t('account.withdraw') }}
      </view>
    </view>
    <view class="card">
      <view class="title">{{ t('account.withdrawRecords') }}</view>
      <view v-if="loading">{{ t('common.loading') }}</view>
      <view v-else-if="!records.length">{{ t('common.empty') }}</view>
      <view v-for="(item, i) in records" :key="i" class="item">
        <view>{{ fenToYuan(item.amount ?? item.withdrawAmount) }}</view>
        <view class="sub">{{ item.createTime || item.createdAt || item.status || '-' }}</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { createWithdraw, getWithdrawRecord, getWithdrawConfig } from '@/api/withdraw'
import { getUserAccount } from '@/api/user'
import { fenToYuan } from '@/features/pay/usePay'
import { useUserStore } from '@/stores/user'
import { useTempDataStore } from '@/stores/tempData'
import { getTenantConfig } from '@/shared/config'
import { setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const user = useUserStore()
const temp = useTempDataStore()
const withdrawBalance = ref('0.00')
const loading = ref(false)
const records = ref<Array<Record<string, unknown>>>([])
const izWithdraw = ref(false)
const minAmount = ref(1)

const canWithdraw = computed(() => izWithdraw.value && Number(withdrawBalance.value) >= minAmount.value)

const thresholdText = computed(() => {
  if (!izWithdraw.value) return t('account.withdrawDisabled')
  return t('account.withdrawMin', { min: minAmount.value })
})

onShow(() => setNavTitle(t('account.withdraw')))

async function loadRecords() {
  loading.value = true
  const res = await getWithdrawRecord({ page: 1, size: 20 })
  loading.value = false
  if (res.success) {
    const data = res.data as { list?: Array<Record<string, unknown>>; records?: Array<Record<string, unknown>> }
    records.value = data?.list || data?.records || (Array.isArray(res.data) ? (res.data as Array<Record<string, unknown>>) : [])
  }
}

onMounted(async () => {
  const serviceId = storage.get('serviceId', '')
  const [cfg, account] = await Promise.all([
    getWithdrawConfig(serviceId ? { serviceId } : {}),
    getUserAccount(serviceId ? { serviceId } : {}),
  ])
  if (cfg.success && cfg.data) {
    const data = cfg.data as { izWithdraw?: boolean; withdrawMinAmount?: number; minWithdraw?: number }
    izWithdraw.value = Boolean(data.izWithdraw)
    if (data.withdrawMinAmount != null) minAmount.value = Number(data.withdrawMinAmount)
    if (data.minWithdraw != null) minAmount.value = Number(data.minWithdraw)
  }
  if (account.success && account.data) {
    const data = account.data as { userWallet?: { recharge?: number } }
    const recharge = Number(data.userWallet?.recharge || 0)
    withdrawBalance.value = fenToYuan(recharge)
  }
  await loadRecords()
})

async function onSubmit() {
  if (!izWithdraw.value) {
    uni.showToast({ title: t('account.withdrawDisabled'), icon: 'none' })
    return
  }
  // Legacy: unpaid order blocks withdraw
  const profile = user.userInfo as { payState?: number } | null
  if (Number(profile?.payState) === 7 || Boolean(temp.unpaidOrderId)) {
    uni.showToast({ title: t('account.withdrawBlockedUnpaid'), icon: 'none' })
    return
  }
  if (!canWithdraw.value) {
    uni.showToast({ title: thresholdText.value, icon: 'none' })
    return
  }
  const tenant = getTenantConfig()
  const channel = tenant.pay?.channelType || 'BAOFU_WXLITE'
  const res = await createWithdraw({
    openid: user.loginInfo.openid,
    channel,
  })
  if (res.success) {
    uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
    await loadRecords()
  }
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.balance-label {
  color: #888;
}
.balance {
  font-size: 56rpx;
  font-weight: 700;
  margin: 12rpx 0;
}
.threshold,
.hint {
  color: #888;
  font-size: 24rpx;
  margin-bottom: 12rpx;
}
.disabled {
  opacity: 0.5;
}
.item {
  padding: 24rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.sub {
  color: #888;
  margin-top: 8rpx;
}
</style>
