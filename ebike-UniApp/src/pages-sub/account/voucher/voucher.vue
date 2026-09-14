<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('account.voucher') }}</view>
      <input
        class="form-input form-input--border"
        v-model="code"
        :placeholder="t('account.voucherPlaceholder')"
      />
      <view class="btn-primary" :class="{ disabled: !canSubmit }" @click="onSubmit">
        {{ t('account.voucherRedeem') }}
      </view>
    </view>

    <view class="card rules">
      <view class="title">{{ t('account.voucherRulesTitle') }}</view>
      <view class="rule" v-for="n in 7" :key="n">{{ t(`account.voucherRule${n}`) }}</view>
    </view>

    <BizPopup
      :visible="showSuccess"
      :title="t('account.voucherSuccess')"
      :confirm-text="t('common.gotIt')"
      :cancel-text="t('account.voucherGoWallet')"
      @confirm="onSuccessConfirm"
      @cancel="goWallet"
      @close="showSuccess = false"
    >
      <view class="success-body">
        <text>{{ rewardText }}</text>
        <text class="success-hint">{{ t('account.voucherSuccessHint') }}</text>
      </view>
    </BizPopup>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { userAddVoucher } from '@/api/voucher'
import BizPopup from '@/widgets/BizPopup.vue'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const user = useUserStore()
const code = ref('')
const showSuccess = ref(false)
const rewardText = ref('')
const submitting = ref(false)

const canSubmit = computed(() => Boolean(code.value.trim()) && !submitting.value)

onShow(() => setNavTitle(t('account.voucher')))

function buildRewardText(type: number, info: unknown) {
  const raw = String(info ?? '')
  switch (type) {
    case 0: {
      const times = raw.split(',')[0] || '0'
      return t('account.voucherRewardFree', { n: times })
    }
    case 1:
      return t('account.voucherRewardCard')
    case 2: {
      const yuan = (Number(info) / 100).toFixed(2)
      return t('account.voucherRewardBalance', { m: yuan })
    }
    case 3: {
      const off = Number(info) * 10
      return t('account.voucherRewardDiscount', { n: off })
    }
    case 4:
      return t('account.voucherRewardMember', { d: raw })
    default:
      return t('common.submitSuccess')
  }
}

async function onSubmit() {
  if (!canSubmit.value) return
  user.hydrateFromStorage()
  if (!user.isLoggedIn) {
    uni.showModal({
      title: t('account.needLogin'),
      showCancel: false,
      confirmText: t('auth.loginNow'),
      success: (r) => {
        if (r.confirm) navigate('redirect', '/pages/auth/quick-login')
      },
    })
    return
  }
  submitting.value = true
  try {
    const res = await userAddVoucher({
      voucherCode: code.value.trim(),
      code: code.value.trim(),
      serviceId: storage.get('serviceId', '') || user.userInfo?.serviceId,
    })
    if (res.success) {
      const data = (res.data || {}) as { rewardType?: number; rewardInfo?: unknown }
      rewardText.value = buildRewardText(Number(data.rewardType), data.rewardInfo)
      showSuccess.value = true
      code.value = ''
      return
    }
    uni.showToast({ title: res.msg || t('account.voucherFail'), icon: 'none' })
  } finally {
    submitting.value = false
  }
}

function onSuccessConfirm() {
  showSuccess.value = false
}

function goWallet() {
  showSuccess.value = false
  navigate('to', '/pages-sub/pay/wallet/wallet')
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 600;
  font-size: 28rpx;
  margin-bottom: 24rpx;
  color: #333;
}
.rules .rule {
  margin-bottom: 24rpx;
  font-size: 28rpx;
  color: #666;
  line-height: 1.5;
}
.rules .rule:last-child {
  margin-bottom: 0;
}
.success-body {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  font-size: 26rpx;
}
.success-hint {
  color: #888;
  font-size: 24rpx;
}
</style>
