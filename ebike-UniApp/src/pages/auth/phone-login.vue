<template>
  <view class="login page">
    <view class="hero">
      <view class="title">{{ t('auth.phoneLogin') }}</view>
    </view>
    <view class="form">
      <input class="form-input" type="number" maxlength="11" v-model="phone" :placeholder="t('auth.phonePlaceholder')" />
      <view class="row">
        <input class="form-input flex" type="number" maxlength="6" v-model="code" :placeholder="t('auth.codePlaceholder')" />
        <view class="code-btn" @click="onSend">{{ t('auth.sendCode') }}</view>
      </view>
      <view class="agree" @click="agreed = !agreed">
        <view class="check" :class="{ on: agreed }" />
        <text>{{ t('auth.agreePrefix') }}</text>
        <text class="link" @click.stop="openProtocol('userProtocol')">{{ t('auth.userAgreement') }}</text>
        <text>&amp;</text>
        <text class="link" @click.stop="openProtocol('privacyProtocol')">{{ t('auth.privacyAgreement') }}</text>
      </view>
      <view class="btn-primary" @click="onLogin">{{ t('auth.loginTitle') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { useAuth } from '@/features/auth/useAuth'
import { setNavTitle } from '@/shared/navigate'
import { openProtocol } from '@/shared/protocol'

const { t } = useI18n()
const { loginWithSms, requestCode } = useAuth()
const phone = ref('')
const code = ref('')
const agreed = ref(false)

onShow(() => setNavTitle(t('auth.phoneLogin')))

function ensureAgreed() {
  if (agreed.value) return true
  uni.showToast({ title: t('auth.agreeRequired'), icon: 'none' })
  return false
}

async function onSend() {
  if (!ensureAgreed()) return
  if (!phone.value) return
  const res = await requestCode(phone.value)
  if (res.success) uni.showToast({ title: t('auth.codeSent'), icon: 'none' })
}

async function onLogin() {
  if (!ensureAgreed()) return
  await loginWithSms(phone.value, code.value)
}
</script>

<style scoped lang="scss">
.page {
  background: #fff;
}
.hero {
  padding: 48rpx 48rpx 16rpx;
}
.title {
  font-size: 48rpx;
  font-weight: 600;
  color: #000;
}
.form {
  padding: 24rpx 48rpx 48rpx;
}
.row {
  display: flex;
  gap: 16rpx;
  align-items: center;
  margin-bottom: 20rpx;
}
.flex {
  flex: 1;
  margin-bottom: 0;
}
.code-btn {
  white-space: nowrap;
  color: var(--brand-color, #3aa0e8);
  padding: 0 8rpx;
  font-size: 28rpx;
  font-weight: 500;
}
.agree {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  font-size: 22rpx;
  color: #888;
  margin: 12rpx 0 40rpx;
  gap: 6rpx;
}
.check {
  width: 28rpx;
  height: 28rpx;
  border: 2rpx solid #ccc;
  border-radius: 6rpx;
  margin-right: 4rpx;
}
.check.on {
  background: var(--brand-color, #3aa0e8);
  border-color: var(--brand-color, #3aa0e8);
}
.link {
  color: var(--brand-color, #3aa0e8);
  margin: 0 6rpx;
}
</style>
