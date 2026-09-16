<template>
  <!-- Legacy setting.vue -->
  <view class="page">
    <view class="quesition-items">
      <template v-if="user.isLoggedIn">
        <view class="question-item" @click="goSecurity">
          <text class="question-item-label">{{ t('account.security') }}</text>
          <image v-if="arrow" class="image" :src="arrow" mode="aspectFit" />
        </view>
        <view class="line" />
      </template>
      <view class="question-item" @click="goProtocol">
        <text class="question-item-label">{{ t('account.termsPrivacy') }}</text>
        <image v-if="arrow" class="image" :src="arrow" mode="aspectFit" />
      </view>
      <view class="line" />
      <view class="question-item" @click="goBilling">
        <text class="question-item-label">{{ t('account.billingRules') }}</text>
        <image v-if="arrow" class="image" :src="arrow" mode="aspectFit" />
      </view>
      <view class="line" />
      <view class="question-item" @click="goAbout">
        <text class="question-item-label">{{ t('account.about') }}</text>
        <image v-if="arrow" class="image" :src="arrow" mode="aspectFit" />
      </view>
      <template v-if="canSwitchLanguage">
        <view class="line" />
        <view class="question-item" @click="onToggleLanguage">
          <text class="question-item-label">{{ t('common.language') }}</text>
          <text class="question-item-label">{{ languageLabel }}</text>
        </view>
      </template>
    </view>

    <view v-if="user.isLoggedIn" class="logout-btn">
      <button class="button" :style="logoutButtonStyle" @click="onLogout">
        {{ t('auth.logout') }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import { useAuth } from '@/features/auth/useAuth'
import { getBrandColor } from '@/shared/config'
import { navigate, setNavTitle } from '@/shared/navigate'
import { getMapCfg } from '@/shared/tenantSkin'
import { SUPPORTED_LOCALES, setLocale, getAcceptLanguage, type AppLocale } from '@/locales'
import { isNative, nativeHost } from '@/shared/nativeHost'

const { t } = useI18n()
const user = useUserStore()
const { logout } = useAuth()
const arrow = computed(() => getMapCfg('iconRightRound'))

const logoutButtonStyle = computed(() => {
  const bg = getBrandColor()
  return `background-color:${bg};border-color:${bg};color:#1E4A38;font-weight:bold;`
})

const canSwitchLanguage = SUPPORTED_LOCALES.length > 1
const languageLabel = computed(() =>
  getAcceptLanguage().toLowerCase().startsWith('en') ? t('common.enUS') : t('common.zhCN'),
)

onShow(() => {
  user.hydrateFromStorage()
  setNavTitle(t('common.settings'))
})

async function onToggleLanguage() {
  const cur = getAcceptLanguage()
  const next = (cur.toLowerCase().startsWith('en') ? 'zh-CN' : 'en-US') as AppLocale
  if (!(SUPPORTED_LOCALES as readonly string[]).includes(next)) return
  setLocale(next)
  if (isNative()) {
    await nativeHost()?.setLanguage(next)
  }
  setNavTitle(t('common.settings'))
}

function goSecurity() {
  navigate('to', '/pages-sub/account/security/security')
}

function goProtocol() {
  navigate('to', '/pages-sub/account/protocol/protocol')
}

function goBilling() {
  navigate('to', '/pages-sub/account/billing-rules/billing-rules')
}

function goAbout() {
  navigate('to', '/pages-sub/account/about/about')
}

function onLogout() {
  logout()
  navigate('reLaunch', '/pages/home/home')
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #fff;
  position: relative;
}
.quesition-items {
  padding: 32rpx 48rpx 0;
}
.question-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.question-item-label {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.image {
  width: 32rpx;
  height: 32rpx;
}
.line {
  height: 2rpx;
  margin: 48rpx 0 46rpx;
  background: #f6f6f6;
}
.logout-btn {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 100rpx;
  display: flex;
  justify-content: center;
}
.button {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 686rpx;
  height: 96rpx;
  border-radius: 32rpx;
  font-size: 32rpx;
  font-weight: 600;
  padding: 0;
  margin: 0;
}
button::after {
  display: none;
}
</style>
