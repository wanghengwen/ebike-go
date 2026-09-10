<template>
  <view class="page invite">
    <view
      class="hero"
      :style="heroBg ? { backgroundImage: `url(${heroBg})` } : undefined"
    >
      <view class="hero__title">{{ t('account.invite') }}</view>
      <view class="hero__reward" v-if="rewardParts.main">
        <text class="hero__label">{{ t('account.inviteEachGet') }}</text>
        <text class="hero__num">{{ rewardParts.main }}</text>
        <text class="hero__unit">{{ rewardParts.unit }}</text>
        <text v-if="rewardParts.extra" class="hero__extra">{{ rewardParts.extra }}</text>
      </view>
      <view class="hero__hint">{{ detail.desc || detail.title || t('account.inviteHint') }}</view>
      <view class="hero__pool" v-if="poolVisible">
        <text class="hero__pool-label">{{ t('account.inviteRewardPool') }}</text>
        <text class="hero__pool-val">
          {{ t('account.inviteRewardPoolRemain', { remain: poolRemain, total: poolTotal }) }}
        </text>
        <text class="hero__pool-count" v-if="detail.inviteNum != null || detail.rewardNum != null">
          {{ t('account.inviteRewardPoolCount', { invite: detail.inviteNum ?? 0, reward: detail.rewardNum ?? 0 }) }}
        </text>
      </view>
    </view>

    <view class="card actions">
      <view class="btn-primary" @click="onInvite">{{ t('account.inviteShare') }}</view>
      <view class="btn-ghost" @click="goRecords">{{ t('account.inviteRecords') }}</view>
      <view class="btn-ghost" @click="goRules">{{ t('account.inviteViewRules') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShareAppMessage, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { createInvite, getInviteDetail } from '@/api/invite'
import { useUserStore } from '@/stores/user'
import { getTenantConfig } from '@/shared/config'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const user = useUserStore()
const detail = ref<Record<string, unknown>>({})

const serviceId = computed(
  () => String(storage.get('serviceId', '') || detail.value.serviceId || ''),
)
const heroBg = computed(
  () => getIconCfg('invite_share_top') || getIconCfg('polite_invite_share'),
)

const rewardParts = computed(() => {
  const info = String(detail.value.rewardInfo || '')
  const first = info.split(',')[0]
  if (!first) return { main: '', unit: '', extra: '' }
  if (Number(detail.value.rewardType) === 1) {
    return {
      main: first,
      unit: t('account.inviteCardUnit') + t('account.inviteRideCard'),
      extra: detail.value.validDay != null ? `${detail.value.validDay}${t('account.inviteDayUnit')}` : '',
    }
  }
  return {
    main: (Number(first) / 100).toFixed(2),
    unit: t('ride.yuan') + t('account.inviteBalance'),
    extra: '',
  }
})

const poolVisible = computed(
  () => detail.value.rewardPool != null && Number(detail.value.rewardPool) >= 0,
)
const poolTotal = computed(() => {
  const pool = Number(detail.value.rewardPool || 0)
  return Number(detail.value.rewardType) === 1 ? String(pool) : (pool / 100).toFixed(2)
})
const poolRemain = computed(() => {
  const pool = Number(detail.value.rewardPool || 0)
  const used = Number(detail.value.rewardNum || 0)
  const remain = Math.max(0, pool - used)
  return Number(detail.value.rewardType) === 1 ? String(remain) : (remain / 100).toFixed(2)
})

onShow(() => setNavTitle(t('account.invite')))

onMounted(async () => {
  user.hydrateFromStorage()
  try {
    const sid = serviceId.value
    const res = await getInviteDetail(sid ? { serviceId: sid } : {})
    if (res.success && res.data) detail.value = res.data as Record<string, unknown>
  } catch (e) {
    logger.warn('invite detail soft fail', e)
  }
})

onShareAppMessage(() => {
  const pin = String(user.userInfo?.pin || '')
  const sid = serviceId.value
  const brand = getTenantConfig().name || 'E-Bike'
  return {
    title: `${brand}${t('account.invite')}`,
    path: `/pages-sub/account/invite/share?serviceId=${encodeURIComponent(sid)}&inviteId=${encodeURIComponent(pin)}`,
    imageUrl: getIconCfg('polite_invite_share') || undefined,
  }
})

async function onInvite() {
  const sid = serviceId.value
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const res = await createInvite(sid ? { serviceId: sid } : {})
    if (!res.success) {
      uni.showToast({ title: res.msg || t('common.networkError'), icon: 'none' })
      return
    }
    // #ifdef MP-WEIXIN
    uni.showShareMenu({ withShareTicket: true })
    uni.showToast({ title: t('account.inviteShareTip'), icon: 'none' })
    // #endif
    // #ifndef MP-WEIXIN
    uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
    // #endif
  } finally {
    uni.hideLoading()
  }
}

function goRecords() {
  navigate('to', '/pages-sub/account/invite/records')
}

function goRules() {
  const sid = serviceId.value
  navigate(
    'to',
    sid
      ? `/pages-sub/account/invite/rules?serviceId=${encodeURIComponent(sid)}`
      : '/pages-sub/account/invite/rules',
  )
}
</script>

<style scoped lang="scss">
.invite {
  padding-bottom: 48rpx;
}
.hero {
  margin: 24rpx 32rpx 0;
  min-height: 280rpx;
  border-radius: 24rpx;
  background: linear-gradient(135deg, #fff5f0, #ffe8e0);
  background-size: cover;
  background-position: center;
  padding: 48rpx 32rpx;
  text-align: center;
}
.hero__title {
  font-size: 40rpx;
  font-weight: 700;
  color: #333;
}
.hero__reward {
  margin-top: 24rpx;
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: center;
  gap: 8rpx;
}
.hero__label {
  color: #666;
  font-size: 28rpx;
}
.hero__num {
  color: #ff5936;
  font-size: 56rpx;
  font-weight: 700;
  line-height: 1;
}
.hero__unit,
.hero__extra {
  color: #ff5936;
  font-size: 28rpx;
  font-weight: 600;
}
.hero__hint {
  margin-top: 20rpx;
  color: #666;
  font-size: 26rpx;
  line-height: 1.5;
}
.hero__pool {
  margin-top: 28rpx;
  padding: 20rpx;
  background: rgba(255, 255, 255, 0.7);
  border-radius: 16rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}
.hero__pool-label {
  font-size: 24rpx;
  color: #888;
}
.hero__pool-val {
  font-size: 30rpx;
  font-weight: 700;
  color: #ff5936;
}
.hero__pool-count {
  font-size: 24rpx;
  color: #666;
}
.actions {
  margin-top: 24rpx;
}
.btn-ghost {
  margin-top: 20rpx;
}
</style>
