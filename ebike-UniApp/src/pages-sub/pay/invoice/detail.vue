<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('pay.invoiceDetail') }}</view>
      <view class="row"><text>{{ t('pay.invoiceTitle') }}</text><text>{{ detail.title || '-' }}</text></view>
      <view class="row"><text>{{ t('pay.invoiceEmail') }}</text><text>{{ detail.email || '-' }}</text></view>
      <view class="row"><text>{{ t('pay.invoiceAmount') }}</text><text>¥{{ amountYuan }}</text></view>
      <view class="row"><text>{{ t('pay.invoiceStatus') }}</text><text>{{ stateText }}</text></view>
      <view class="row" v-if="detail.content"><text>{{ t('pay.invoiceContent') }}</text><text>{{ detail.content }}</text></view>
      <view class="fail" v-if="Number(detail.state) === 2">{{ detail.failedReason || t('pay.invoiceFailed') }}</view>
      <view class="btn-primary" v-if="Number(detail.state) === 2" @click="reopen">{{ t('pay.invoiceRetry') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { fenToYuan } from '@/features/pay/usePay'
import { navigate, setNavTitle } from '@/shared/navigate'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const detail = ref<Record<string, unknown>>({})

const amountYuan = computed(() => fenToYuan(Number(detail.value.amount || detail.value.money || 0)))
const stateText = computed(() => {
  const s = Number(detail.value.state)
  if (s === 1) return t('pay.invoiceDone')
  if (s === 2) return t('pay.invoiceFailed')
  return t('pay.invoicePending')
})

onShow(() => setNavTitle(t('pay.invoiceDetail')))

onLoad((q) => {
  try {
    const raw = q?.params ? decodeURIComponent(String(q.params)) : ''
    if (raw) detail.value = JSON.parse(raw)
  } catch (e) {
    logger.warn('invoice detail parse fail', e)
  }
})

function reopen() {
  const orderIds = detail.value.orderIds
  const money = Number(detail.value.amount || 0) / 100
  const params = {
    orderIds: orderIds || [],
    money,
  }
  navigate(
    'to',
    `/pages-sub/pay/invoice/apply?params=${encodeURIComponent(JSON.stringify(params))}`,
  )
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 20rpx;
}
.row {
  display: flex;
  justify-content: space-between;
  gap: 24rpx;
  padding: 16rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.fail {
  margin-top: 20rpx;
  color: #e34d59;
}
.btn-primary {
  margin-top: 32rpx;
}
</style>
