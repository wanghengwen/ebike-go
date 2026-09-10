<template>
  <view class="page">
    <view class="hero">
      <view class="title">{{ t('auth.verifiedTitle') }}</view>
      <view class="hint">{{ t('auth.verifiedHint') }}</view>
    </view>
    <view class="form">
      <input class="form-input" v-model="authName" :placeholder="t('auth.realNamePlaceholder')" />
      <input
        class="form-input"
        type="idcard"
        v-model="authNo"
        :placeholder="t('auth.idCardPlaceholder')"
        maxlength="18"
      />
      <view class="age-tip">{{ t('auth.ageTip', { min: minAge, max: maxAge }) }}</view>
      <view class="btn-primary" :class="{ disabled: submitting }" @click="onSubmit">
        {{ t('common.confirm') }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { submitAuth, getUseCarConfig } from '@/api/user'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const authName = ref('')
const authNo = ref('')
const submitting = ref(false)
const minAge = ref(16)
const maxAge = ref(65)

onShow(() => setNavTitle(t('auth.verifiedTitle')))

onMounted(async () => {
  const serviceId = storage.get('serviceId', '')
  const res = await getUseCarConfig(serviceId ? { serviceId } : {})
  if (res.success && res.data) {
    const data = res.data as { minAge?: number; maxAge?: number }
    if (data.minAge) minAge.value = data.minAge
    if (data.maxAge) maxAge.value = data.maxAge
  }
})

function checkName(name: string) {
  return /^[\u00B7\u3400-\u9FFF\uE000-\uF8FF\uF900-\uFAFF]{2,16}$/u.test(name)
}

function checkIdCard(no: string) {
  return /^\d{17}[\dXx]$/.test(no)
}

async function onSubmit() {
  const name = authName.value.trim()
  const no = authNo.value.trim()
  if (!name || !no) {
    uni.showToast({ title: t('auth.authEmpty'), icon: 'none' })
    return
  }
  if (!checkName(name)) {
    uni.showToast({ title: t('auth.authNameInvalid'), icon: 'none' })
    return
  }
  if (!checkIdCard(no)) {
    uni.showToast({ title: t('auth.authIdInvalid'), icon: 'none' })
    return
  }
  if (submitting.value) return
  submitting.value = true
  try {
    const res = await submitAuth({
      authName: name,
      authNo: no,
      type: 0,
    })
    if (res.success) {
      const data = res.data as { resultCode?: number; handlePhone?: string } | undefined
      if (data && data.resultCode === 1) {
        storage.set('verifiedSuc', 1)
        uni.showToast({ title: t('auth.authSuccess'), icon: 'success' })
        setTimeout(() => navigate('back'), 500)
        return
      }
      if (data && data.resultCode === -1) {
        const hp = encodeURIComponent(String(data.handlePhone || ''))
        const an = encodeURIComponent(name)
        const ao = encodeURIComponent(no)
        navigate(
          'to',
          `/pages-sub/account/bind-card/bind-card-select?handlePhone=${hp}&authName=${an}&authNo=${ao}`,
        )
        return
      }
      uni.showToast({ title: res.msg || t('auth.authFail'), icon: 'none' })
      return
    }
    uni.showToast({ title: res.msg || t('auth.authFail'), icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: #fff;
}
.hero {
  padding: 48rpx 48rpx 24rpx;
}
.title {
  font-size: 48rpx;
  font-weight: 600;
  color: #000;
  margin-bottom: 16rpx;
}
.hint {
  color: #666;
  font-size: 28rpx;
  line-height: 40rpx;
}
.form {
  padding: 24rpx 48rpx 48rpx;
}
.age-tip {
  color: #ff5936;
  font-size: 24rpx;
  margin: 8rpx 0 40rpx;
}
</style>
