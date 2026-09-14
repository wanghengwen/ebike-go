<template>
  <!-- Legacy accountSecurity -->
  <view class="page">
    <view class="container">
      <view class="row">
        <text class="row_text">{{ t('auth.verifiedTitle') }}</text>
        <text class="row_content">{{ formatName }}</text>
      </view>
      <view class="line" />
      <view class="row">
        <text class="row_text">{{ t('account.mobileNumber') }}</text>
        <text class="row_content">{{ maskedPhone }}</text>
      </view>
      <template v-if="!hideCancelAccount">
        <view class="line" />
        <view class="row" @click="goCancel">
          <text class="row_text">{{ t('account.cancelAccount') }}</text>
          <text class="row_content">{{ t('account.cancelAccountHint') }}</text>
          <image v-if="arrow" class="image" :src="arrow" mode="aspectFit" />
        </view>
      </template>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getPersonInfo } from '@/api/user'
import { getTenantConfig } from '@/shared/config'
import { navigate, setNavTitle } from '@/shared/navigate'
import { phoneDesensitize } from '@/shared/phone'
import { getMapCfg } from '@/shared/tenantSkin'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const user = useUserStore()
const arrow = computed(() => getMapCfg('iconRightRound'))
const hideCancelAccount = computed(
  () => Boolean(getTenantConfig().customSetting?.hideCancelAccount),
)

/** Legacy: new Array(len).join('*') + last char */
const formatName = computed(() => {
  const info = user.userInfo as Record<string, unknown>
  const value = String(info.authName || '')
  if (!value) return '--'
  return `${'*'.repeat(Math.max(value.length - 1, 0))}${value.slice(-1)}`
})

const maskedPhone = computed(() => {
  const masked = phoneDesensitize(user.userInfo.phone)
  return masked === '--' ? '--' : masked
})

onShow(() => setNavTitle(t('account.security')))

onMounted(async () => {
  user.hydrateFromStorage()
  const res = await getPersonInfo()
  if (res.success && res.data) user.setUserInfo(res.data as never)
})

function goCancel() {
  navigate('to', '/pages-sub/account/cancel/cancel')
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #fff;
}
.container {
  padding: 32rpx 48rpx 0;
}
.row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.image {
  width: 32rpx;
  height: 32rpx;
  margin-left: 8rpx;
  flex-shrink: 0;
}
.row_text {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  flex-shrink: 0;
}
.row_content {
  flex: 1;
  text-align: right;
  font-size: 28rpx;
  color: #666;
  margin-left: 24rpx;
}
.line {
  height: 2rpx;
  margin: 48rpx 0 46rpx;
  background: #f6f6f6;
}
</style>
