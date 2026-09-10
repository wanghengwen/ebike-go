<template>
  <view class="page">
    <view class="card" v-if="mode === 'occupied'">
      <template v-if="needFace">
        <view class="title">{{ t('account.verifyOldPhoneLastFour') }}</view>
        <view class="hint">{{ t('account.verifyOldPhoneLastFourSub') }}</view>
        <view class="hint sub">{{ t('account.verifyOldPhoneHint', { phone: maskedPhone }) }}</view>
        <input
          class="form-input"
          type="number"
          maxlength="4"
          v-model="digitInput"
          :placeholder="t('account.lastFourPlaceholder')"
        />
      </template>
      <template v-else>
        <view class="title">{{ t('account.verifyOldPhoneFull') }}</view>
        <view class="hint">{{ t('account.verifyOldPhoneFullSub') }}</view>
        <view class="mask-row">
          <text class="mask-part">{{ phoneParts.front }}</text>
          <input
            class="form-input mid"
            type="number"
            maxlength="6"
            v-model="digitInput"
            :placeholder="t('account.middleSixPlaceholder')"
          />
          <text class="mask-part">{{ phoneParts.end }}</text>
        </view>
      </template>
      <view class="btn-primary" @click="onOccupiedSubmit">{{ t('common.confirm') }}</view>
    </view>
    <view class="card" v-else>
      <view class="title">{{ t('account.changePhone') }}</view>
      <view class="hint">{{ t('account.changePhoneHint') }}</view>
      <view class="current" v-if="currentPhone">{{ t('account.currentPhone') }}：{{ currentPhone }}</view>
      <input class="form-input" type="number" maxlength="11" v-model="phone" :placeholder="t('auth.phonePlaceholder')" />
      <view class="row">
        <input class="form-input flex" type="number" maxlength="6" v-model="code" :placeholder="t('auth.codePlaceholder')" />
        <view class="code-btn" @click="onSend">{{ t('auth.sendCode') }}</view>
      </view>
      <view class="btn-primary" @click="onSubmit">{{ t('common.confirm') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { sendSmsCode, withPhoneAvoidAudit, changeBindWithFace, getPersonInfo } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const user = useUserStore()
const phone = ref('')
const code = ref('')
const mode = ref<'change' | 'occupied'>('change')
const handlePhone = ref('')
const authName = ref('')
const authNo = ref('')
const digitInput = ref('')
/** Face-verify path: last 4 + changeBindWithFace. Otherwise middle 6 + withPhoneAvoidAudit. */
const needFace = ref(false)

const currentPhone = computed(() => String(user.userInfo?.phone || ''))
const maskedPhone = computed(() => {
  const raw = handlePhone.value
  if (raw.includes('*')) return raw.replace('+86-', '')
  const p = raw.replace(/\D/g, '')
  if (p.length < 7) return raw || '****'
  return `${p.slice(0, 3)}****${p.slice(-4)}`
})

/** Legacy returnFrontNum: split masked phone on '*'. */
const phoneParts = computed(() => {
  const phoneStr = handlePhone.value.replace('+86-', '')
  if (!phoneStr.includes('*')) {
    const digits = phoneStr.replace(/\D/g, '')
    return { front: digits.slice(0, 3) || '', end: digits.slice(-4) || '' }
  }
  const parts = phoneStr.split('*')
  return {
    front: parts[0] || '',
    end: parts[parts.length - 1] || '',
  }
})

onShow(() => {
  setNavTitle(mode.value === 'occupied' ? t('account.verifyOldPhone') : t('account.changePhone'))
  user.hydrateFromStorage()
})

function truthyFlag(v: unknown) {
  return v === true || v === 'true' || v === '1' || v === 1
}

onLoad((q) => {
  if (String(q?.mode || '') === 'occupied') {
    mode.value = 'occupied'
    handlePhone.value = decodeURIComponent(String(q?.handlePhone || ''))
    authName.value = decodeURIComponent(String(q?.authName || ''))
    authNo.value = decodeURIComponent(String(q?.authNo || ''))
    needFace.value = truthyFlag(q?.izOnCertification)
  }
})

function normalizePhone(p: string) {
  const raw = String(p || '').trim()
  if (!raw) return raw
  if (raw.startsWith('+')) return raw
  return `+86-${raw.replace(/^86-?/, '')}`
}

async function onSend() {
  if (!phone.value || phone.value.length < 11) {
    uni.showToast({ title: t('auth.phonePlaceholder'), icon: 'none' })
    return
  }
  const res = await sendSmsCode({ phone: normalizePhone(phone.value), scene: 1 })
  if (res.success) uni.showToast({ title: t('auth.codeSent'), icon: 'none' })
}

/** Active change of own phone number (product extension; not legacy occupied bind). */
async function onSubmit() {
  if (!phone.value || !code.value) return
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const info = user.userInfo as Record<string, unknown>
    const res = await withPhoneAvoidAudit({
      phone: normalizePhone(phone.value),
      messageCode: code.value,
      authName: info.authName || info.realName || '',
      authNo: info.authNo || info.idCard || '',
      applyType: 1,
    })
    if (res.success) {
      const profile = await getPersonInfo()
      if (profile.success && profile.data) user.setUserInfo(profile.data as never)
      uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
      setTimeout(() => navigate('back'), 500)
    } else {
      uni.showToast({ title: res.msg || t('common.networkError'), icon: 'none' })
    }
  } finally {
    uni.hideLoading()
  }
}

/**
 * Legacy bindCardPhoneVertify:
 * - Face on: last 4 → changeBindWithFace
 * - Face off: middle 6 to complete masked phone → withPhoneAvoidAudit
 */
async function onOccupiedSubmit() {
  const expectedLen = needFace.value ? 4 : 6
  if (digitInput.value.length !== expectedLen) {
    uni.showToast({
      title: needFace.value ? t('account.lastFourPlaceholder') : t('account.middleSixPlaceholder'),
      icon: 'none',
    })
    return
  }

  let fullPhone = ''
  if (needFace.value) {
    fullPhone = String(phoneParts.value.front) + String(digitInput.value) + String(phoneParts.value.end)
    const digits = handlePhone.value.replace(/\D/g, '')
    if (!phoneParts.value.front && digits.length >= 4) {
      fullPhone = digits.slice(0, -4) + digitInput.value
    }
  } else {
    fullPhone = String(phoneParts.value.front) + String(digitInput.value) + String(phoneParts.value.end)
  }

  const payload = {
    authName: authName.value,
    authNo: authNo.value,
    fourNumber: digitInput.value,
    phone: normalizePhone(fullPhone.replace(/\D/g, '') || fullPhone),
    applyType: 1,
  }

  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const res = needFace.value
      ? await changeBindWithFace(payload)
      : await withPhoneAvoidAudit(payload)
    if (res.success) {
      uni.showToast({ title: res.msg || t('common.submitSuccess'), icon: 'success' })
      setTimeout(() => navigate('reLaunch', '/pages/home/home'), 500)
    } else {
      uni.showToast({ title: res.msg || t('common.networkError'), icon: 'none' })
    }
  } finally {
    uni.hideLoading()
  }
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 600;
  font-size: 32rpx;
  margin-bottom: 12rpx;
  color: #333;
}
.hint {
  color: #888;
  font-size: 24rpx;
  margin-bottom: 20rpx;
  line-height: 1.5;
}
.hint.sub {
  margin-top: -8rpx;
}
.current {
  margin-bottom: 16rpx;
  color: #666;
}
.mask-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 20rpx;
}
.mask-part {
  font-size: 36rpx;
  font-weight: 600;
  letter-spacing: 4rpx;
}
.form-input.mid {
  flex: 1;
  margin-bottom: 0;
  text-align: center;
  letter-spacing: 8rpx;
  padding: 0 16rpx;
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
  padding: 20rpx 12rpx;
}
</style>
