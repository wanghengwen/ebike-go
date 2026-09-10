<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('account.depositRefund') }}</view>
      <view class="amount">¥{{ fenToYuan(deposit) }}</view>
      <view class="hint">{{ t('account.depositRefundHint', { amount: fenToYuan(deposit) }) }}</view>
      <view class="btn-primary" :class="{ disabled: !deposit }" @click="onSubmit">{{ t('common.confirm') }}</view>
      <view class="btn-ghost" @click="goBack">{{ t('common.cancel') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { refundDeposit, getPersonInfo } from '@/api/user'
import { getWalletInfo } from '@/api/pay'
import { fenToYuan } from '@/features/pay/usePay'
import { useUserStore } from '@/stores/user'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const user = useUserStore()
const deposit = ref(0)

onShow(() => setNavTitle(t('account.depositRefund')))

function userPin() {
  return String(user.userInfo?.pin || storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
}

async function resolveDeposit(queryAmount?: string) {
  if (queryAmount != null && queryAmount !== '' && !Number.isNaN(Number(queryAmount))) {
    deposit.value = Math.floor(Number(queryAmount))
    return
  }
  user.hydrateFromStorage()
  const info = user.userInfo as Record<string, unknown>
  if (info.depositedMount != null) {
    deposit.value = Math.floor(Number(info.depositedMount))
    if (deposit.value > 0) return
  }
  const profile = await getPersonInfo()
  if (profile.success && profile.data) {
    user.setUserInfo(profile.data as never)
    const data = profile.data as { depositedMount?: number }
    if (data.depositedMount != null) {
      deposit.value = Math.floor(Number(data.depositedMount))
      if (deposit.value > 0) return
    }
  }
  const wallet = await getWalletInfo()
  if (wallet.success && wallet.data) {
    const data = wallet.data as { deposit?: number; depositedMount?: number; depositBalance?: number }
    deposit.value = Math.floor(
      Number(data.depositedMount ?? data.deposit ?? data.depositBalance ?? 0),
    )
  }
}

onLoad((q) => {
  void resolveDeposit(
    (q?.depositedMount as string) || (q?.deposit as string) || (q?.amount as string),
  )
})

function goBack() {
  navigate('back')
}

function onSubmit() {
  if (!deposit.value) {
    uni.showToast({ title: t('account.depositRefundEmpty'), icon: 'none' })
    return
  }
  uni.showModal({
    title: t('account.depositRefund'),
    content: t('account.depositRefundConfirm', { amount: fenToYuan(deposit.value) }),
    success: async (r) => {
      if (!r.confirm) return
      uni.showLoading({ title: t('common.loading'), mask: true })
      try {
        const res = await refundDeposit({
          sale_type: 'DEPOSIT',
          refund_fee: Number(deposit.value),
          pin: userPin(),
        })
        if (res.success) {
          uni.showToast({ title: t('account.depositRefundSuccess'), icon: 'success' })
          setTimeout(() => navigate('back'), 800)
        } else {
          uni.showToast({ title: res.msg || t('account.depositRefundFail'), icon: 'none' })
        }
      } finally {
        uni.hideLoading()
      }
    },
  })
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.amount {
  font-size: 56rpx;
  font-weight: 700;
  color: #3aa0e8;
  margin-bottom: 16rpx;
}
.hint {
  color: #666;
  line-height: 1.5;
  margin-bottom: 32rpx;
}
.btn-ghost {
  margin-top: 20rpx;
}
.disabled {
  opacity: 0.5;
}
</style>
