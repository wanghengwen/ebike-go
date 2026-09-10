<template>
  <view class="page">
    <view class="card" v-if="step === 0">
      <view class="title">{{ t('account.cancelAccount') }}</view>
      <view class="hint">{{ t('account.cancelHint') }}</view>
      <view class="check-row" @click="agreed = !agreed">
        <view class="check" :class="{ on: agreed }" />
        <text>{{ t('account.cancelAgree') }}</text>
      </view>
      <view class="btn-primary danger" :class="{ disabled: !agreed }" @click="onPrecheck">
        {{ t('account.cancelConfirm') }}
      </view>
    </view>

    <view class="card" v-else-if="step === 1">
      <view class="title">{{ t('account.cancelAccount') }}</view>
      <view class="phone">{{ maskPhone }}</view>
      <view class="hint" :class="{ warn: !canCancel }">
        {{ canCancel ? t('account.cancelReady') : t('account.cancelBlocked') }}
      </view>
      <view v-if="flags.izRecharge" class="flag">{{ t('account.cancelFlagBalance') }}</view>
      <view v-if="flags.izToPay" class="flag">{{ t('account.cancelFlagOrder') }}</view>
      <view v-if="flags.izDeposit" class="flag">{{ t('account.cancelFlagDeposit') }}</view>
      <view v-if="flags.izHaveUnauditedUserTicket" class="flag">{{ t('account.cancelFlagTicket') }}</view>
      <view v-if="canCancel" class="btn-primary danger" @click="step = 2">{{ t('account.cancelNext') }}</view>
      <view v-else class="btn-ghost" @click="navigateBack">{{ t('common.confirm') }}</view>
    </view>

    <view class="card" v-else>
      <view class="title">{{ t('account.cancelVerify') }}</view>
      <view class="phone">{{ maskPhone }}</view>
      <view class="row">
        <input class="input flex" type="number" maxlength="6" v-model="smsCode" :placeholder="t('auth.codePlaceholder')" />
        <view class="code-btn" @click="onSendCode">{{ countdown > 0 ? `${countdown}s` : t('auth.sendCode') }}</view>
      </view>
      <view class="btn-primary danger" @click="onSubmitCancel">{{ t('account.cancelConfirm') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { queryCancelAccount, submitCancel } from '@/api/account'
import { sendSmsCode } from '@/api/user'
import { useAuth } from '@/features/auth/useAuth'
import { useUserStore } from '@/stores/user'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const { logout } = useAuth()
const user = useUserStore()

const step = ref(0)
const agreed = ref(false)
const canCancel = ref(false)
const smsCode = ref('')
const countdown = ref(0)
const flags = ref({
  izRecharge: false,
  izToPay: false,
  izDeposit: false,
  izHaveUnauditedUserTicket: false,
})

const maskPhone = computed(() => {
  const p = String(user.userInfo.phone || '').replace(/^\+86-?/, '')
  if (p.length < 7) return p || '-'
  return `${p.slice(0, 3)}****${p.slice(-4)}`
})

onShow(() => setNavTitle(t('account.cancelAccount')))

function navigateBack() {
  navigate('back')
}

async function onPrecheck() {
  if (!agreed.value) return
  const res = await queryCancelAccount()
  if (!res.success || !res.data) return
  const data = res.data as {
    success?: boolean
    izDeposit?: boolean
    izToPay?: boolean
    izRecharge?: boolean
    izHaveUnauditedUserTicket?: boolean
  }
  canCancel.value = Boolean(data.success)
  flags.value = {
    izRecharge: Boolean(data.izRecharge),
    izToPay: Boolean(data.izToPay),
    izDeposit: Boolean(data.izDeposit),
    izHaveUnauditedUserTicket: Boolean(data.izHaveUnauditedUserTicket),
  }
  step.value = 1
}

function normalizePhone(phone: string) {
  const raw = String(phone || '').trim()
  if (!raw) return raw
  if (raw.startsWith('+')) return raw
  return `+86-${raw.replace(/^86-?/, '')}`
}

async function onSendCode() {
  if (countdown.value > 0) return
  const phone = normalizePhone(String(user.userInfo.phone || ''))
  if (!phone) return
  const res = await sendSmsCode({ phone, scene: 3 })
  if (!res.success) return
  uni.showToast({ title: t('auth.codeSent'), icon: 'none' })
  countdown.value = 60
  const timer = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) clearInterval(timer)
  }, 1000)
}

function onSubmitCancel() {
  if (smsCode.value.length < 6) {
    uni.showToast({ title: t('auth.codePlaceholder'), icon: 'none' })
    return
  }
  uni.showModal({
    title: t('account.cancelAccount'),
    content: t('account.cancelConfirm'),
    success: async (r) => {
      if (!r.confirm) return
      const phone = normalizePhone(String(user.userInfo.phone || ''))
      const res = await submitCancel({ code: smsCode.value, phone })
      if (res.success) {
        uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
        setTimeout(() => logout(), 500)
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
.hint {
  color: #666;
  line-height: 1.5;
  margin-bottom: 24rpx;
}
.warn {
  color: #ff8401;
}
.phone {
  font-size: 32rpx;
  font-weight: 700;
  margin-bottom: 16rpx;
}
.flag {
  color: #666;
  padding: 12rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.check-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 28rpx;
  color: #666;
  font-size: 24rpx;
}
.check {
  width: 32rpx;
  height: 32rpx;
  border: 2rpx solid #ccc;
  border-radius: 6rpx;
}
.check.on {
  background: #3aa0e8;
  border-color: #3aa0e8;
}
.row {
  display: flex;
  gap: 16rpx;
  align-items: center;
  margin-bottom: 28rpx;
}
.input {
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 24rpx;
}
.flex {
  flex: 1;
}
.code-btn {
  color: #3aa0e8;
  white-space: nowrap;
}
.danger {
  background: #e34d59;
}
.disabled {
  opacity: 0.5;
}
.btn-ghost {
  margin-top: 20rpx;
}
</style>
