<template>
  <view class="page">
    <view v-if="loading" class="card">{{ t('common.loading') }}</view>

    <template v-else-if="authorized">
      <view class="card auth">
        <view class="auth__ok">{{ t('account.payScoreAuthorized') }}</view>
        <view class="auth__benefit">{{ benefitText }}</view>
        <view class="auth__badge">
          <image v-if="wxScoreIcon" class="auth__icon" :src="wxScoreIcon" mode="aspectFit" />
          <text>{{ t('account.payScore') }}</text>
          <text class="auth__sep">|</text>
          <text>{{ t('account.payScoreCreditTip') }}</text>
        </view>
        <image v-if="wxScoreBg" class="auth__bg" :src="wxScoreBg" mode="widthFix" />
        <view class="auth__unbind">{{ t('account.payScoreUnbindHint') }}</view>
        <view class="btn-primary" @click="goHome">{{ t('common.back') }}</view>
      </view>
    </template>

    <template v-else>
      <view class="card">
        <view class="title">{{ t('account.payScore') }}</view>
        <view class="status">{{ statusText }}</view>
        <view class="btn-primary" @click="onOpen">{{ t('common.enable') }}</view>
        <view class="btn-ghost" @click="onClose">{{ t('common.disable') }}</view>
      </view>
      <view class="card">
        <view v-if="!records.length">{{ t('common.empty') }}</view>
        <view v-for="(item, i) in records" :key="i" class="item">
          <view>{{ item.status || item.permissionState || item.authorization_state || '-' }}</view>
          <view class="sub">{{ item.createTime || item.updateTime || item.authorization_success_time || '-' }}</view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { closeScorePermission, getWechatPayScoreRecord, scorePermission } from '@/api/wechatScore'
import { uniLoginCode } from '@/features/auth/useAuth'
import { navigate, setNavTitle } from '@/shared/navigate'
import { getIconCfg } from '@/shared/tenantSkin'

const { t } = useI18n()
const loading = ref(false)
const authorized = ref(false)
const statusText = ref('--')
const records = ref<Array<Record<string, unknown>>>([])
/** 0 = deposit-free + ride-first; other = recharge-free + ride-first */
const scoreMode = ref(0)

const wxScoreIcon = computed(() => getIconCfg('wxScore'))
const wxScoreBg = computed(() => getIconCfg('wxScoreBg'))
const benefitText = computed(() =>
  scoreMode.value === 0 ? t('account.payScoreBenefitDeposit') : t('account.payScoreBenefitRecharge'),
)

onShow(() => setNavTitle(t('account.payScore')))

onLoad((q) => {
  if (q?.wxCredictScoreConfig != null) scoreMode.value = Number(q.wxCredictScoreConfig) || 0
  if (q?.mode != null) scoreMode.value = Number(q.mode) || 0
})

async function withLoginCode<T>(fn: (code: string) => Promise<T>): Promise<T | null> {
  const code = await uniLoginCode()
  if (!code) {
    uni.showToast({ title: t('auth.wxAuthFail'), icon: 'none' })
    return null
  }
  return fn(code)
}

function applyAuthState(data: Record<string, unknown>) {
  const state = String(data.authorization_state || data.status || data.permissionState || '')
  authorized.value = state === 'AVAILABLE' || state === 'AUTHORIZED' || state === '1'
  statusText.value = state || '--'
}

async function load() {
  loading.value = true
  const res = await withLoginCode((code) => getWechatPayScoreRecord({ code }))
  loading.value = false
  if (!res?.success || !res.data) return
  const data = res.data as Record<string, unknown> | Array<Record<string, unknown>>
  if (Array.isArray(data)) {
    records.value = data
    const first = data[0] || {}
    applyAuthState(first)
    return
  }
  applyAuthState(data)
  const list = data.records || data.list
  records.value = Array.isArray(list) ? (list as Array<Record<string, unknown>>) : [data]
}

onMounted(() => {
  void load()
})

function goHome() {
  navigate('reLaunch', '/pages/home/home')
}

async function onOpen() {
  const res = await withLoginCode((code) => scorePermission({ code }))
  if (!res) return
  if (res.success) {
    const token = res.data
    // #ifdef MP-WEIXIN
    if (token && typeof token === 'string') {
      uni.navigateToMiniProgram({
        appId: 'wxd8f3793ea3b935b8',
        path: 'pages/use/enable',
        extraData: { apply_permissions_token: token },
      })
    }
    // #endif
    uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
    await load()
  }
}

async function onClose() {
  uni.showModal({
    title: t('account.payScore'),
    content: t('common.confirm'),
    success: async (r) => {
      if (!r.confirm) return
      const res = await withLoginCode((code) => closeScorePermission({ code }))
      if (res?.success) {
        uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
        await load()
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
.status {
  font-size: 40rpx;
  font-weight: 700;
  color: #3aa0e8;
  margin-bottom: 28rpx;
}
.btn-ghost {
  margin-top: 20rpx;
}
.item {
  padding: 24rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.sub {
  color: #888;
  margin-top: 8rpx;
}
.auth__ok {
  font-size: 28rpx;
  color: #999;
}
.auth__benefit {
  margin-top: 16rpx;
  font-size: 40rpx;
  font-weight: 700;
  color: #111;
  line-height: 1.35;
}
.auth__badge {
  margin-top: 16rpx;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8rpx;
  font-size: 24rpx;
  color: #6f6f6f;
}
.auth__icon {
  width: 30rpx;
  height: 30rpx;
}
.auth__sep {
  color: #d4d4d4;
  margin: 0 4rpx;
}
.auth__bg {
  width: 100%;
  margin-top: 24rpx;
}
.auth__unbind {
  margin: 32rpx 0;
  font-size: 26rpx;
  color: #999;
  line-height: 1.5;
}
</style>
