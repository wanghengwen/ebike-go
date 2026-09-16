<template>
  <!-- Legacy cancelPhoneNum.vue -->
  <view class="page">
    <view class="phone-title">{{ t('account.cancelVerifyTitle') }}</view>
    <view class="phone-content">
      <view class="phone-num dark">+86 {{ maskPhone }}</view>
      <view class="phone-num row">
        <input
          class="phone-input"
          type="number"
          maxlength="6"
          v-model="smsCode"
          :placeholder="t('auth.codePlaceholder')"
        />
        <text
          class="phnone-code"
          :style="{ color: countdown > 0 ? brandColor : '#999999' }"
          @click="getCode"
        >
          {{ countdown > 0 ? t('account.resendIn', { n: countdown }) : t('auth.sendCode') }}
        </text>
      </view>
    </view>
    <view class="phone-bottom">
      <view
        class="phone-btn"
        :style="{ background: verified ? brandColor : '#CCCCCC' }"
        @click="onSubmit"
      >
        {{ t('account.cancelNext') }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { submitCancel } from '@/api/account'
import { sendSmsCode } from '@/api/user'
import { useAuth } from '@/features/auth/useAuth'
import { getBrandColor } from '@/shared/config'
import { getLoginPath, navigate, setNavTitle } from '@/shared/navigate'
import { phoneDesensitize } from '@/shared/phone'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const user = useUserStore()
const { logout } = useAuth()

const smsCode = ref('')
const countdown = ref(0)
const verified = ref(false)
const brandColor = computed(() => getBrandColor())
const maskPhone = computed(() => {
  const raw = String(user.userInfo.phone || '')
  const digits = raw.replace(/^\+?86-?/, '')
  return phoneDesensitize(digits)
})

watch(smsCode, (v) => {
  verified.value = String(v).length >= 6
})

onShow(() => {
  user.hydrateFromStorage()
  setNavTitle(t('account.cancelAccount'))
})

function normalizePhone(phone: string) {
  const raw = String(phone || '').trim()
  if (!raw) return raw
  if (raw.includes('+86')) return raw
  return `+86-${raw.replace(/^86-?/, '')}`
}

async function getCode() {
  if (countdown.value > 0) return
  uni.showLoading({ title: t('common.loading'), mask: true })
  const phone = normalizePhone(String(user.userInfo.phone || ''))
  const res = await sendSmsCode({ phone, scene: 3 })
  uni.hideLoading()
  if (!res.success) {
    uni.showToast({ title: t('common.submitFail'), icon: 'none' })
    return
  }
  uni.showToast({ title: t('auth.codeSent'), icon: 'none' })
  countdown.value = 60
  const id = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) clearInterval(id)
  }, 1000)
}

async function onSubmit() {
  if (!verified.value) return
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const res = await submitCancel({ code: smsCode.value })
    uni.hideLoading()
    if (res.success) {
      uni.showModal({
        title: t('account.cancelSuccess'),
        content: '',
        showCancel: false,
        success: () => {
          logout()
          navigate('reLaunch', getLoginPath())
        },
      })
      return
    }
    uni.showToast({ title: t('common.submitFail'), icon: 'none' })
  } catch {
    uni.hideLoading()
    uni.showToast({ title: t('common.submitFail'), icon: 'none' })
  }
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #fff;
  padding: 80rpx 48rpx 0;
  box-sizing: border-box;
}
.phone-title {
  font-size: 40rpx;
  font-weight: 600;
  color: #121212;
}
.phone-content {
  margin-top: 64rpx;
}
.phone-num {
  font-size: 28rpx;
  color: #999;
  margin-bottom: 32rpx;
  &.dark {
    color: #121212;
    font-size: 32rpx;
  }
  &.row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-bottom: 2rpx solid #f6f6f6;
    padding-bottom: 24rpx;
  }
}
.phone-input {
  flex: 1;
  font-size: 28rpx;
}
.phnone-code {
  margin-left: 24rpx;
  font-size: 28rpx;
  flex-shrink: 0;
}
.phone-bottom {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 100rpx;
  display: flex;
  justify-content: center;
}
.phone-btn {
  width: 686rpx;
  height: 96rpx;
  line-height: 96rpx;
  text-align: center;
  border-radius: 32rpx;
  color: #1e4a38;
  font-size: 32rpx;
  font-weight: 600;
}
</style>
