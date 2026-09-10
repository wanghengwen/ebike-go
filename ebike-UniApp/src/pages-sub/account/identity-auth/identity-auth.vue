<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('face.authTitle') }}</view>
      <view class="hint">{{ t('face.authHint') }}</view>
      <view class="btn-primary" @click="onStart">{{ t('face.start') }}</view>
      <view class="exit" @click="onExit">{{ t('face.exit') }}</view>
      <view class="remind">{{ t('face.remind') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const authName = ref('')
const authNo = ref('')
const izOnCertification = ref('')
const callback = ref('')

onShow(() => setNavTitle(t('face.authTitle')))

onLoad((q) => {
  authName.value = decodeURIComponent(String(q?.authName || ''))
  authNo.value = decodeURIComponent(String(q?.authNo || ''))
  izOnCertification.value = String(q?.izOnCertification || '')
  callback.value = String(q?.callback || '')
})

function onStart() {
  const qs = [
    `authName=${encodeURIComponent(authName.value)}`,
    `authNo=${encodeURIComponent(authNo.value)}`,
    `izOnCertification=${encodeURIComponent(izOnCertification.value)}`,
    `callback=${encodeURIComponent(callback.value)}`,
  ].join('&')
  navigate('to', `/pages-sub/account/face-scan/face-scan?${qs}`)
}

function onExit() {
  navigate('back')
}
</script>

<style scoped lang="scss">
.title {
  font-size: 40rpx;
  font-weight: 700;
  margin-bottom: 16rpx;
}
.hint {
  color: #666;
  line-height: 1.6;
  margin-bottom: 48rpx;
}
.exit {
  margin-top: 28rpx;
  text-align: center;
  color: #888;
  font-size: 28rpx;
}
.remind {
  margin-top: 48rpx;
  text-align: center;
  color: #aaa;
  font-size: 22rpx;
}
</style>
