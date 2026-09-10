<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('account.idOccupiedTitle') }}</view>
      <view class="hint">{{ t('account.idOccupiedHint', { phone: maskedPhone }) }}</view>
      <view class="btn-primary" @click="goVerifyPhone">{{ t('account.knowPhone') }}</view>
      <view class="btn-ghost" @click="goIdAppeal">{{ t('account.unknownPhone') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { changeBindWithFace } from '@/api/user'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const handlePhone = ref('')
const authName = ref('')
const authNo = ref('')
const izOnCertification = ref('')

const maskedPhone = computed(() => {
  const raw = handlePhone.value
  if (raw.includes('*')) return raw.replace('+86-', '')
  const p = raw.replace(/\D/g, '')
  if (p.length < 7) return raw || '****'
  return `${p.slice(0, 3)}****${p.slice(-4)}`
})

const faceOn = computed(() => {
  const v = izOnCertification.value
  return v === 'true' || v === '1' || v === 'True'
})

onShow(() => setNavTitle(t('account.idOccupiedTitle')))

onLoad((q) => {
  handlePhone.value = decodeURIComponent(String(q?.handlePhone || ''))
  authName.value = decodeURIComponent(String(q?.authName || ''))
  authNo.value = decodeURIComponent(String(q?.authNo || ''))
  izOnCertification.value = String(q?.izOnCertification || '')
})

function qs() {
  const parts = [
    `handlePhone=${encodeURIComponent(handlePhone.value)}`,
    `authName=${encodeURIComponent(authName.value)}`,
    `authNo=${encodeURIComponent(authNo.value)}`,
  ]
  if (izOnCertification.value) {
    parts.push(`izOnCertification=${encodeURIComponent(izOnCertification.value)}`)
  }
  return parts.join('&')
}

function goVerifyPhone() {
  navigate('to', `/pages-sub/account/change-phone/change-phone?mode=occupied&${qs()}`)
}

/** Legacy: face on → changeBindWithFace(applyType=2); face off → upload ID form. */
function goIdAppeal() {
  if (faceOn.value) {
    uni.showModal({
      title: t('account.confirmIrreversible'),
      cancelText: t('common.cancel'),
      confirmText: t('common.confirm'),
      success: (r) => {
        if (r.confirm) void submitFaceAppeal()
      },
    })
    return
  }
  navigate('to', `/pages-sub/account/bind-card/bind-card-form?${qs()}`)
}

async function submitFaceAppeal() {
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const res = await changeBindWithFace({
      applyType: 2,
      authNo: authNo.value,
      authName: authName.value,
    })
    if (res.success) {
      navigate('reLaunch', '/pages/home/home')
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
  font-weight: 700;
  font-size: 36rpx;
  margin-bottom: 16rpx;
}
.hint {
  color: #666;
  line-height: 1.5;
  margin-bottom: 40rpx;
}
.btn-ghost {
  margin-top: 20rpx;
  text-align: center;
  color: #3aa0e8;
  padding: 24rpx;
}
</style>
