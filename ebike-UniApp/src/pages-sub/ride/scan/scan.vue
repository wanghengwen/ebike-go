<template>
  <view class="page">
    <!-- #ifdef H5 -->
    <view id="qr-reader" class="reader" />
    <view class="flash" @click="toggleFlash">
      <image v-if="flashIcon" class="flash__icon" :src="flashIcon" mode="aspectFit" />
      <text class="flash__text">{{ flashOn ? t('ride.torchOff') : t('ride.torchOn') }}</text>
    </view>
    <view class="footer">
      <view class="link" @click="goEnterId">{{ t('ride.enterId') }}</view>
    </view>
    <!-- #endif -->

    <!-- #ifndef H5 -->
    <view class="fallback">
      <text class="fallback__tip">{{ t('ride.scanStarting') }}</text>
    </view>
    <!-- #endif -->
  </view>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { applyScanResult, type ScanPurpose } from '@/features/bike/useScanGate'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const purpose = ref<ScanPurpose>('use')
const flashOn = ref(false)
const flashIcon = computed(() => getIconCfg('flash'))
let handled = false

// #ifdef H5
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let html5QrCode: any = null
// #endif

onShow(() => setNavTitle(t('ride.scan')))

onLoad((q) => {
  const p = String(q?.purpose || 'use')
  purpose.value = p === 'fill' ? 'fill' : 'use'
  // #ifdef H5
  setTimeout(() => void startH5Scan(), 200)
  // #endif
  // #ifndef H5
  openNativeScan()
  // #endif
})

onUnload(() => {
  void stopH5()
})

onUnmounted(() => {
  void stopH5()
})

async function onRaw(raw: string) {
  if (handled) return
  handled = true
  await stopH5()
  await applyScanResult(raw, {
    purpose: purpose.value,
    failToEnterId: purpose.value === 'use',
    replaceScanPage: true,
  })
}

function goEnterId() {
  void stopH5()
  if (purpose.value === 'fill') {
    navigate('back')
    return
  }
  navigate('redirect', '/pages-sub/ride/enter-id/enter-id')
}

function openNativeScan() {
  uni.scanCode({
    onlyFromCamera: false,
    success: (res) => {
      void onRaw(res.result || '')
    },
    fail: () => {
      if (purpose.value === 'use') {
        navigate('redirect', '/pages-sub/ride/enter-id/enter-id')
      } else {
        navigate('back')
      }
    },
  })
}

async function startH5Scan() {
  // #ifdef H5
  try {
    uni.showLoading({ title: t('ride.scanStarting'), mask: true })
    const { Html5Qrcode } = await import('html5-qrcode')
    html5QrCode = new Html5Qrcode('qr-reader')
    await html5QrCode.start(
      { facingMode: 'environment' },
      {
        fps: 10,
        qrbox: { width: 280, height: 280 },
        aspectRatio: 1.777778,
      },
      (msg: string) => {
        if (msg) void onRaw(msg)
      },
      () => {
        /* frame miss — ignore */
      },
    )
  } catch (e) {
    logger.warn('h5 scan start fail', e)
    uni.showToast({ title: t('ride.scanFail'), icon: 'none' })
    if (purpose.value === 'use') {
      navigate('redirect', '/pages-sub/ride/enter-id/enter-id')
    } else {
      navigate('back')
    }
  } finally {
    setTimeout(() => uni.hideLoading(), 800)
  }
  // #endif
}

async function stopH5() {
  // #ifdef H5
  try {
    if (html5QrCode) {
      await html5QrCode.stop()
      html5QrCode.clear()
      html5QrCode = null
    }
  } catch (e) {
    logger.warn('h5 scan stop soft fail', e)
    html5QrCode = null
  }
  // #endif
}

function toggleFlash() {
  // #ifdef H5
  try {
    if (!html5QrCode) return
    const torch = html5QrCode.getRunningTrackCameraCapabilities?.()?.torchFeature?.()
    if (!torch) {
      uni.showToast({ title: t('ride.torchUnsupported'), icon: 'none' })
      return
    }
    torch.apply(!torch.value())
    flashOn.value = Boolean(torch.value())
  } catch (e) {
    logger.warn('torch fail', e)
    uni.showToast({ title: t('ride.torchUnsupported'), icon: 'none' })
  }
  // #endif
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: rgba(0, 0, 0, 0.85);
  position: relative;
}

.reader {
  width: 100%;
  min-height: 100vh;
}

.flash {
  position: fixed;
  left: 50%;
  bottom: 22%;
  transform: translateX(-50%);
  z-index: 20;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
}

.flash__icon {
  width: 64rpx;
  height: 64rpx;
}

.flash__text {
  color: #fff;
  font-size: 26rpx;
}

.footer {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 8%;
  z-index: 20;
  display: flex;
  justify-content: center;
}

.link {
  color: #fff;
  font-size: 28rpx;
  padding: 16rpx 32rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.45);
  border-radius: 999rpx;
}

.fallback {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.fallback__tip {
  color: #fff;
  font-size: 28rpx;
}
</style>
