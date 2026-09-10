<template>
  <view class="page">
    <view class="cell-group">
      <view class="cell-row" @click="goSecurity">
        <text>{{ t('account.security') }}</text>
        <image v-if="arrow" class="cell-row__arrow" :src="arrow" mode="aspectFit" />
      </view>
      <view class="cell-line" />
      <view class="cell-row" @click="openProtocol('userProtocol')">
        <text>{{ t('account.termsPrivacy') }}</text>
        <image v-if="arrow" class="cell-row__arrow" :src="arrow" mode="aspectFit" />
      </view>
      <view class="cell-line" />
      <view class="cell-row" @click="goBilling">
        <text>{{ t('account.billingRules') }}</text>
        <image v-if="arrow" class="cell-row__arrow" :src="arrow" mode="aspectFit" />
      </view>
      <view class="cell-line" />
      <view class="cell-row" @click="goAbout">
        <text>{{ t('account.about') }}</text>
        <image v-if="arrow" class="cell-row__arrow" :src="arrow" mode="aspectFit" />
      </view>
      <!-- #ifndef MP-WEIXIN -->
      <view class="cell-line" />
      <view class="cell-row" @click="showLang = !showLang">
        <text>{{ t('common.language') }}</text>
        <image v-if="arrow" class="cell-row__arrow" :src="arrow" mode="aspectFit" />
      </view>
      <view v-if="showLang" class="lang-box">
        <view
          v-for="opt in localeStore.options"
          :key="opt.code"
          class="lang"
          :class="{ active: localeStore.locale === opt.code }"
          @click="onChange(opt.code)"
        >
          {{ t(opt.labelKey) }}
        </view>
      </view>
      <!-- #endif -->
    </view>

    <view v-if="user.isLoggedIn" class="logout-wrap">
      <view class="btn-primary logout" @click="onLogout">{{ t('auth.logout') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { useLocaleStore } from '@/stores/locale'
import { useUserStore } from '@/stores/user'
import { useAuth } from '@/features/auth/useAuth'
import type { AppLocale } from '@/locales'
import { navigate, setNavTitle } from '@/shared/navigate'
import { openProtocol } from '@/shared/protocol'
import { getMapCfg } from '@/shared/tenantSkin'

const { t } = useI18n()
const localeStore = useLocaleStore()
const user = useUserStore()
const { logout } = useAuth()
const showLang = ref(false)
const arrow = computed(() => getMapCfg('iconRightRound'))

onShow(() => {
  user.hydrateFromStorage()
  setNavTitle(t('common.settings'))
})

function onChange(code: AppLocale) {
  localeStore.change(code)
  setNavTitle(t('common.settings'))
  uni.showToast({
    title: t(code === 'zh-CN' ? 'common.zhCN' : 'common.enUS'),
    icon: 'none',
  })
}

function goSecurity() {
  if (!user.isLoggedIn) {
    uni.showToast({ title: t('account.needLogin'), icon: 'none' })
    setTimeout(() => navigate('to', '/pages/auth/quick-login'), 400)
    return
  }
  navigate('to', '/pages-sub/account/security/security')
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
  min-height: 100vh;
  background: #fff;
  position: relative;
}
.cell-group {
  padding-top: 32rpx;
}
.lang-box {
  padding: 0 0 16rpx;
}
.lang {
  padding: 20rpx 0;
  font-size: 28rpx;
  color: #666;
}
.lang.active {
  color: var(--brand-color, #3aa0e8);
  font-weight: 600;
}
.logout-wrap {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 100rpx;
  display: flex;
  justify-content: center;
  padding: 0 32rpx;
  box-sizing: border-box;
}
.logout {
  width: 686rpx;
}
</style>
