<template>
  <view class="launch page">
    <image v-if="launchBg" class="bg" :src="launchBg" mode="aspectFill" />
    <view class="brand">{{ brandName }}</view>
    <view class="company">{{ t('brand.company') }}</view>
  </view>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { useScanGate } from '@/features/bike/useScanGate'
import { useUserStore } from '@/stores/user'
import { getTenantConfig } from '@/shared/config'
import { isNative } from '@/shared/nativeHost'
import { navigate } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const user = useUserStore()
const { extractCarId } = useScanGate()

const brandName = ref(getTenantConfig().name || t('brand.name'))
const launchBg = ref(getTenantConfig().customSetting?.launchBg || '')
const queryQ = ref('')
let jumped = false
let timer: ReturnType<typeof setTimeout> | null = null

onLoad((q) => {
  if (q?.q) queryQ.value = decodeURIComponent(String(q.q))
})

function readEnterQ(): string {
  if (queryQ.value) return queryQ.value
  try {
    const enter = uni.getEnterOptionsSync?.() as { query?: { q?: string } } | undefined
    if (enter?.query?.q) return decodeURIComponent(String(enter.query.q))
  } catch (e) {
    logger.warn('getEnterOptionsSync soft fail', e)
  }
  try {
    const launch = uni.getLaunchOptionsSync?.() as { query?: { q?: string } } | undefined
    if (launch?.query?.q) return decodeURIComponent(String(launch.query.q))
  } catch {
    /* ignore */
  }
  return ''
}

function goHome() {
  if (jumped) return
  jumped = true
  navigate('reLaunch', '/pages/home/home')
}

function goPrecycling(carId: string) {
  if (jumped) return
  jumped = true
  storage.set('scanCarId', carId)
  navigate(
    'reLaunch',
    `/pages-sub/ride/precycling/precycling?carId=${encodeURIComponent(carId)}`,
  )
}

onMounted(() => {
  user.hydrateFromStorage()
  brandName.value = getTenantConfig().name || t('brand.name')
  launchBg.value = getTenantConfig().customSetting?.launchBg || ''

  // Rider WebView：入口 hash 已由原生指定，勿再 reLaunch 首页，否则会污染返回栈。
  if (isNative()) {
    jumped = true
    return
  }

  const rawQ = readEnterQ()
  if (rawQ) {
    const carId = extractCarId(rawQ)
    if (carId && /^\w+$/.test(carId)) {
      logger.info('launch deep-link carId', carId)
      goPrecycling(carId)
      return
    }
  }
  storage.remove('scanCarId')
  timer = setTimeout(goHome, 600)
})

onUnmounted(() => {
  if (timer) clearTimeout(timer)
})
</script>

<style scoped lang="scss">
.launch {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(180deg, #e8f4fc 0%, #ffffff 70%);
  position: relative;
}
.bg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}
.brand {
  position: relative;
  font-size: 48rpx;
  font-weight: 700;
  color: #1a1a1a;
}
.company {
  position: relative;
  margin-top: 12rpx;
  color: #888;
  font-size: 24rpx;
  letter-spacing: 2rpx;
}
</style>
