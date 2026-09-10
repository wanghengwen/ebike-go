<template>
  <view class="page">
    <view v-if="maskedName || maskedPhone" class="profile-head">
      <view class="name">{{ maskedName || '-' }}</view>
      <view class="phone">{{ maskedPhone || '-' }}</view>
    </view>

    <view class="cell-group">
      <view class="cell-row" @click="go('/pages/auth/verified')">
        <text>{{ t('auth.verifiedTitle') }}</text>
        <image v-if="arrow" class="cell-row__arrow" :src="arrow" mode="aspectFit" />
      </view>
      <view class="cell-line" />
      <view class="cell-row" @click="go('/pages-sub/account/change-phone/change-phone')">
        <text>{{ t('account.changePhone') }}</text>
        <image v-if="arrow" class="cell-row__arrow" :src="arrow" mode="aspectFit" />
      </view>
      <view class="cell-line" />
      <view class="cell-row" @click="go('/pages-sub/account/career/career')">
        <text>{{ t('account.careerAuth') }}</text>
        <image v-if="arrow" class="cell-row__arrow" :src="arrow" mode="aspectFit" />
      </view>
      <template v-if="!hideCancelAccount">
        <view class="cell-line" />
        <view class="cell-row" @click="go('/pages-sub/account/cancel/cancel')">
          <text class="danger">{{ t('account.cancelAccount') }}</text>
          <image v-if="arrow" class="cell-row__arrow" :src="arrow" mode="aspectFit" />
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
import { useUserStore } from '@/stores/user'
import { getMapCfg } from '@/shared/tenantSkin'

const { t } = useI18n()
const user = useUserStore()
const arrow = computed(() => getMapCfg('iconRightRound'))
const hideCancelAccount = computed(
  () => Boolean(getTenantConfig().customSetting?.hideCancelAccount),
)

const maskedName = computed(() => {
  const info = user.userInfo as Record<string, unknown>
  const raw = String(info.authName || info.realName || info.name || info.userName || '')
  if (!raw) return ''
  if (raw.length <= 1) return '*'
  return `${raw[0]}${'*'.repeat(Math.max(raw.length - 1, 1))}`
})

const maskedPhone = computed(() => {
  const p = String(user.userInfo.phone || '').replace(/^\+86-?/, '').replace(/\D/g, '')
  if (!p) return ''
  if (p.length < 7) return p
  return `${p.slice(0, 3)}****${p.slice(-4)}`
})

onShow(() => setNavTitle(t('account.security')))

onMounted(async () => {
  user.hydrateFromStorage()
  const res = await getPersonInfo()
  if (res.success && res.data) user.setUserInfo(res.data as never)
})

function go(url: string) {
  navigate('to', url)
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: #fff;
}
.profile-head {
  padding: 48rpx 48rpx 16rpx;
}
.name {
  font-size: 36rpx;
  font-weight: 700;
  color: #333;
}
.phone {
  margin-top: 8rpx;
  color: #888;
  font-size: 28rpx;
}
.danger {
  color: #e34d59;
}
</style>
