<template>
  <view class="page share">
    <view
      class="bg"
      :style="bottomBg ? { backgroundImage: `url(${bottomBg})` } : undefined"
    >
      <view
        class="hero"
        :style="topBg ? { backgroundImage: `url(${topBg})` } : undefined"
      >
        <view class="hero__title">{{ t('account.inviteShareTitle') }}</view>
        <view class="hero__reward" v-if="rewardText">
          <text>{{ t('account.inviteEachGet') }}</text>
          <text class="hero__num">{{ rewardText }}</text>
        </view>
      </view>

      <view class="form card">
        <input
          class="input"
          type="number"
          maxlength="11"
          v-model="phone"
          :placeholder="t('auth.phonePlaceholder')"
        />
        <view class="code-row">
          <input
            class="input code"
            type="number"
            maxlength="6"
            v-model="code"
            :placeholder="t('auth.codePlaceholder')"
          />
          <view class="send" @click="onSend">{{ codeBtn }}</view>
        </view>
        <view class="btn-primary" @click="onSubmit">{{ t('account.inviteClaim') }}</view>
        <view class="rules" @click="goRules">{{ t('account.inviteViewRules') }}</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { acceptInvite, getInviteDetail } from '@/api/invite'
import { useAuth } from '@/features/auth/useAuth'
import { useUserStore } from '@/stores/user'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const user = useUserStore()
const { loginWithSms, requestCode } = useAuth()

const phone = ref('')
const code = ref('')
const serviceId = ref('')
const inviteId = ref('')
const detail = ref<Record<string, unknown>>({})
const countdown = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const topBg = computed(() => getIconCfg('invite_share_top'))
const bottomBg = computed(() => getIconCfg('invite_share_bottom'))
const codeBtn = computed(() =>
  countdown.value > 0 ? t('account.resendIn', { n: countdown.value }) : t('auth.sendCode'),
)

const rewardText = computed(() => {
  const info = String(detail.value.rewardInfo || '')
  const first = info.split(',')[0]
  if (!first) return ''
  if (Number(detail.value.rewardType) === 1) {
    const days = detail.value.validDay != null ? Number(detail.value.validDay) : 0
    return `${first}${t('account.inviteCardUnit')}${days}${t('account.inviteDayUnit')}${t('account.inviteRideCard')}`
  }
  const yuan = (Number(first) / 100).toFixed(2)
  return `${yuan}${t('ride.yuan')}${t('account.inviteBalance')}`
})

onShow(() => setNavTitle(t('account.inviteShareNav')))

onLoad(async (q) => {
  serviceId.value = decodeURIComponent(String(q?.serviceId || storage.get('serviceId', '') || ''))
  inviteId.value = decodeURIComponent(String(q?.inviteId || ''))
  try {
    const res = await getInviteDetail(serviceId.value ? { serviceId: serviceId.value } : {})
    if (res.success && res.data) detail.value = res.data as Record<string, unknown>
  } catch (e) {
    logger.warn('invite share detail soft fail', e)
  }
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function startCountdown() {
  countdown.value = 60
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0 && timer) {
      clearInterval(timer)
      timer = null
    }
  }, 1000)
}

async function onSend() {
  if (countdown.value > 0) return
  if (!/^1\d{10}$/.test(phone.value.trim())) {
    uni.showToast({ title: t('auth.phonePlaceholder'), icon: 'none' })
    return
  }
  const res = await requestCode(phone.value.trim())
  if (res?.success !== false) {
    uni.showToast({ title: t('auth.codeSent'), icon: 'none' })
    startCountdown()
  }
}

async function onSubmit() {
  if (!/^1\d{10}$/.test(phone.value.trim())) {
    uni.showToast({ title: t('auth.phonePlaceholder'), icon: 'none' })
    return
  }
  if (code.value.trim().length < 4) {
    uni.showToast({ title: t('auth.codePlaceholder'), icon: 'none' })
    return
  }
  const loginRes = await loginWithSms(phone.value.trim(), code.value.trim(), { skipNavigate: true })
  if (!loginRes?.success) return

  const pin = String(user.userInfo?.pin || '')
  try {
    const res = await acceptInvite({
      inviteePin: pin,
      inviterPin: inviteId.value,
      serviceId: serviceId.value,
      phone: phone.value.trim(),
    })
    if (res.success) {
      uni.showToast({ title: t('account.inviteClaimOk'), icon: 'success' })
      setTimeout(() => navigate('reLaunch', '/pages/home/home'), 1200)
    } else {
      uni.showToast({ title: res.msg || t('common.networkError'), icon: 'none' })
      setTimeout(() => navigate('reLaunch', '/pages/home/home'), 1500)
    }
  } catch (e) {
    logger.warn('acceptInvite soft fail', e)
    navigate('reLaunch', '/pages/home/home')
  }
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
.share {
  min-height: 100vh;
  padding: 0;
}
.bg {
  min-height: 100vh;
  background: linear-gradient(180deg, #fff5f0, #f5f6f8);
  background-size: cover;
  background-repeat: no-repeat;
  padding-bottom: 48rpx;
}
.hero {
  min-height: 360rpx;
  background-size: cover;
  background-position: center;
  padding: 64rpx 40rpx 40rpx;
  text-align: center;
}
.hero__title {
  font-size: 40rpx;
  font-weight: 700;
  color: #333;
}
.hero__reward {
  margin-top: 24rpx;
  font-size: 28rpx;
  color: #666;
}
.hero__num {
  color: #ff5936;
  font-size: 40rpx;
  font-weight: 700;
  margin: 0 8rpx;
}
.form {
  margin: 24rpx 32rpx;
}
.input {
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 24rpx;
}
.code-row {
  display: flex;
  gap: 16rpx;
  align-items: center;
  margin-bottom: 24rpx;
}
.code-row .code {
  flex: 1;
  margin-bottom: 0;
}
.send {
  color: var(--brand-color, #3aa0e8);
  font-size: 26rpx;
  white-space: nowrap;
  padding: 0 8rpx;
}
.rules {
  margin-top: 28rpx;
  text-align: center;
  color: #3aa0e8;
  font-size: 26rpx;
}
</style>
