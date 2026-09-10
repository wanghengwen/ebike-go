<template>
  <view class="page">
    <view class="title">{{ t('face.scanTitle') }}</view>
    <view class="camera-wrap">
      <camera
        v-if="cameraOn"
        class="camera"
        :device-position="position"
        @error="onCameraError"
      />
    </view>
    <view class="flip" @click="flip">{{ t('face.flipCamera') }}</view>
    <view class="btn-primary" :class="{ disabled: busy }" @click="takePhoto">
      {{ t('face.takePhoto') }}
    </view>
    <view class="tip">{{ t('face.scanTip') }}</view>
  </view>
</template>

<script setup lang="ts">
import { onUnmounted, ref } from 'vue'
import { onHide, onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { rentCheck, submitAuth } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { useTempDataStore } from '@/stores/tempData'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const user = useUserStore()
const temp = useTempDataStore()

const authName = ref('')
const authNo = ref('')
const izOnCertification = ref('')
const callback = ref('')
const cameraOn = ref(false)
const position = ref<'front' | 'back'>('front')
const busy = ref(false)
let ctx: UniApp.CameraContext | null = null

onShow(() => {
  setNavTitle(t('face.scanTitle'))
  // Legacy: success then back → force home
  const ts = Number(storage.get('faceResultSuccessTime', 0) || 0)
  if (ts && temp.faceResult === '1') {
    temp.setFaceResult('')
    navigate('reLaunch', '/pages/home/home')
    return
  }
  ctx = uni.createCameraContext()
  cameraOn.value = true
})

onHide(() => {
  cameraOn.value = false
})
onUnmounted(() => {
  cameraOn.value = false
})

onLoad((q) => {
  authName.value = decodeURIComponent(String(q?.authName || ''))
  authNo.value = decodeURIComponent(String(q?.authNo || ''))
  izOnCertification.value = String(q?.izOnCertification || '')
  callback.value = String(q?.callback || '')
})

function flip() {
  position.value = position.value === 'front' ? 'back' : 'front'
}

function onCameraError(e: unknown) {
  logger.warn('camera error', e)
  uni.showModal({
    title: t('face.cameraFail'),
    content: t('face.cameraPermission'),
    confirmText: t('face.openSetting'),
    success: (r) => {
      if (r.confirm) uni.openSetting({})
    },
  })
}

function goResult(state: 0 | 1, msg = '') {
  const qs = [
    `state=${state}`,
    `authName=${encodeURIComponent(authName.value)}`,
    `authNo=${encodeURIComponent(authNo.value)}`,
    `callback=${encodeURIComponent(callback.value)}`,
    `msg=${encodeURIComponent(msg || '')}`,
  ].join('&')
  navigate('to', `/pages-sub/account/face-result/face-result?${qs}`)
}

function readBase64(filePath: string): Promise<string> {
  return new Promise((resolve, reject) => {
    uni.getFileSystemManager().readFile({
      filePath,
      encoding: 'base64',
      success: (res) => resolve(String(res.data || '')),
      fail: reject,
    })
  })
}

async function checkFace(image: string) {
  busy.value = true
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const pin = String(user.userInfo?.pin || storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
    if (callback.value === 'scan_use_bike') {
      const res = await rentCheck({ pin, image })
      goResult(res.success ? 1 : 0, res.msg || '')
      return
    }
    const res = await submitAuth({
      authName: authName.value.trim(),
      authNo: authNo.value.trim(),
      type: 1,
      image,
    })
    const data = res.data as { resultCode?: number; handlePhone?: string } | undefined
    if (res.success && data?.resultCode === 1) {
      goResult(1)
      return
    }
    if (res.success && data?.resultCode === -1) {
      const hp = encodeURIComponent(String(data.handlePhone || ''))
      const cert = encodeURIComponent(izOnCertification.value || 'true')
      navigate(
        'redirect',
        `/pages-sub/account/bind-card/bind-card-select?handlePhone=${hp}&authName=${encodeURIComponent(authName.value)}&authNo=${encodeURIComponent(authNo.value)}&izOnCertification=${cert}`,
      )
      return
    }
    goResult(0, res.msg || t('face.verifyFail'))
  } finally {
    busy.value = false
    uni.hideLoading()
  }
}

function takePhoto() {
  if (busy.value) return
  if (!ctx) ctx = uni.createCameraContext()
  uni.authorize({
    scope: 'scope.camera',
    success: () => {
      ctx?.takePhoto({
        quality: 'normal',
        success: async (res) => {
          try {
            const b64 = await readBase64(res.tempImagePath)
            await checkFace(b64)
          } catch (e) {
            logger.warn('read face photo fail', e)
            uni.showModal({ title: t('face.storagePermission'), showCancel: false })
          }
        },
        fail: () => {
          uni.showToast({ title: t('face.takeFail'), icon: 'none' })
        },
      })
    },
    fail: () => {
      uni.showModal({
        title: t('face.cameraPermission'),
        confirmText: t('face.openSetting'),
        success: (r) => {
          if (r.confirm) uni.openSetting({})
        },
      })
    },
  })
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  padding: 32rpx;
  box-sizing: border-box;
  background: #111;
}
.title {
  color: #fff;
  font-size: 36rpx;
  font-weight: 700;
  margin-bottom: 24rpx;
  text-align: center;
}
.camera-wrap {
  width: 100%;
  height: 720rpx;
  border-radius: 24rpx;
  overflow: hidden;
  background: #222;
}
.camera {
  width: 100%;
  height: 100%;
}
.flip {
  margin: 24rpx 0;
  text-align: center;
  color: #3aa0e8;
}
.tip {
  margin-top: 24rpx;
  text-align: center;
  color: #aaa;
  font-size: 24rpx;
}
.disabled {
  opacity: 0.5;
}
</style>
