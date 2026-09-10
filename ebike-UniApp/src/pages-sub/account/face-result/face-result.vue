<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ ok ? t('face.success') : t('face.fail') }}</view>
      <view v-if="ok" class="msg">
        <view class="name">{{ maskName(authName) }}</view>
        <view class="id">{{ maskId(authNo) }}</view>
      </view>
      <view v-else class="msg">{{ msg || t('face.verifyFail') }}</view>
      <view class="btn-primary" v-if="ok" @click="onUseBike">{{ t('face.useBike') }}</view>
      <view class="btn-primary" v-else @click="onRetry">{{ t('face.retry') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { useFaceCheck } from '@/features/auth/useFaceCheck'
import { useTempDataStore } from '@/stores/tempData'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const temp = useTempDataStore()
const { markFaceSuccess, clearFaceCache } = useFaceCheck()

const state = ref('0')
const authName = ref('')
const authNo = ref('')
const msg = ref('')
const callback = ref('')
const ok = computed(() => state.value === '1')

onShow(() => setNavTitle(ok.value ? t('face.success') : t('face.fail')))

onLoad((q) => {
  state.value = String(q?.state ?? '0')
  authName.value = decodeURIComponent(String(q?.authName || ''))
  authNo.value = decodeURIComponent(String(q?.authNo || ''))
  msg.value = decodeURIComponent(String(q?.msg || ''))
  callback.value = String(q?.callback || '')
  if (state.value === '1') markFaceSuccess()
  else temp.setFaceResult('0')
})

function maskName(v: string) {
  if (!v) return ''
  if (v.length <= 1) return '*'
  return `${v[0]}${'*'.repeat(Math.max(v.length - 1, 1))}`
}

function maskId(v: string) {
  if (!v || v.length < 8) return v
  return `${v.slice(0, 4)}${'*'.repeat(v.length - 8)}${v.slice(-4)}`
}

function onUseBike() {
  if (callback.value === 'scan_use_bike') {
    const carId = temp.preCyclingCar?.carId
    temp.setPreCyclingCar(null)
    // Legacy keeps 3min cache and returns to precycling (even if carId empty → still navigate)
    if (carId) {
      navigate(
        'reLaunch',
        `/pages-sub/ride/precycling/precycling?carId=${encodeURIComponent(carId)}`,
      )
      return
    }
    navigate('reLaunch', '/pages-sub/ride/precycling/precycling')
    return
  }
  clearFaceCache()
  navigate('reLaunch', '/pages/home/home')
}

function onRetry() {
  navigate('back')
}
</script>

<style scoped lang="scss">
.title {
  font-size: 40rpx;
  font-weight: 700;
  margin-bottom: 24rpx;
  text-align: center;
}
.msg {
  text-align: center;
  color: #666;
  margin-bottom: 48rpx;
  line-height: 1.6;
}
.name {
  font-size: 36rpx;
  color: #1a1a1a;
  font-weight: 600;
}
.id {
  margin-top: 8rpx;
}
</style>
