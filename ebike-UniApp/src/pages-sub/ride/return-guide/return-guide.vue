<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ title }}</view>
      <view class="sub" :class="{ ok: canReturn }">{{ subText }}</view>
      <view class="body">{{ body }}</view>

      <view class="tips_bg" v-if="hasGuideArt">
        <image v-if="warnIcon" class="warn" :src="warnIcon" mode="widthFix" />
        <!-- page0 helmet lock -->
        <image v-if="pageType === 0 && artMain" class="tip_img" :src="artMain" mode="widthFix" />
        <!-- direction -->
        <image v-if="pageType === 4 && artMain" class="tip_img" :src="artMain" mode="widthFix" />
        <image
          v-if="pageType === 4 && artOverlay"
          class="tip_ebike"
          :style="{ transform: `rotate(${bikeAngle}deg)` }"
          :src="artOverlay"
          mode="widthFix"
        />
        <!-- RFID -->
        <image v-if="pageType === 5 && artMain" class="tip_img" :src="artMain" mode="widthFix" />
        <image v-if="pageType === 5 && artOverlay" class="tip_ebike" :src="artOverlay" mode="widthFix" />
        <!-- helmet unlock gif -->
        <image v-if="pageType === 6 && artMain" class="tip_img" :src="artMain" mode="widthFix" />
        <!-- kickstand -->
        <image v-if="pageType === 8 && artMain" class="tip_img" :src="artMain" mode="widthFix" />
        <!-- camera -->
        <image v-if="pageType === 11 && artMain" class="tip_img" :src="artMain" mode="widthFix" />
      </view>

      <view class="steps" v-if="steps.length">
        <view v-for="(s, i) in steps" :key="i" class="step">{{ i + 1 }}. {{ s }}</view>
      </view>
      <view
        class="btn-primary"
        :class="{ disabled: !canReturn && pageType !== 0 }"
        @click="onPrimary"
      >
        {{ primaryText }}
      </view>
      <view class="btn-ghost" @click="goPark">{{ t('ride.parkSearch') }}</view>
      <view class="help" v-if="orderId">
        <text class="help-q">{{ helpQuestion }}</text>
        <text v-if="showPhotoApply" class="link" @click="goApply">{{ t('returnGuide.photoApply') }}</text>
        <text v-else class="link" @click="goService">{{ t('returnGuide.contactService') }}</text>
      </view>
      <view class="btn-ghost" @click="goBack">{{ t('common.cancel') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { onHide, onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { partMatch } from '@/api/applyReturn'
import { getReturnCarConfig, isCanCameraAudit, unlockHelmet } from '@/api/riding'
import { guideApplyTypeFrom } from '@/features/bike/returnTypes'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const pageType = ref(0)
const orderId = ref('')
const carId = ref('')
const returnType = ref('')
const canReturn = ref(false)
const matchHint = ref('')
const bikeAngle = ref(0)
const freeCfg = ref({
  direction: false,
  rfidBeacon: false,
  helmet: false,
  kickstand: false,
  camera: false,
})
const cameraAuditOk = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null
let cameraTimer: ReturnType<typeof setInterval> | null = null

const GUIDE: Record<
  number,
  { titleKey: string; bodyKey: string; stepKeys?: string[]; part?: string; primaryKey: string; helpKey: string }
> = {
  0: {
    titleKey: 'returnGuide.helmetTitle',
    bodyKey: 'returnGuide.helmetBody',
    stepKeys: ['returnGuide.stepHelmet', 'returnGuide.stepRetry'],
    part: 'helmet',
    primaryKey: 'returnGuide.returnNow',
    helpKey: 'returnGuide.helpHelmet',
  },
  4: {
    titleKey: 'returnGuide.directionTitle',
    bodyKey: 'returnGuide.directionBody',
    stepKeys: ['returnGuide.stepPark', 'returnGuide.stepAlign', 'returnGuide.stepRetry'],
    part: 'direction',
    primaryKey: 'returnGuide.returnNow',
    helpKey: 'returnGuide.helpDirection',
  },
  5: {
    titleKey: 'returnGuide.rfidTitle',
    bodyKey: 'returnGuide.rfidBody',
    stepKeys: ['returnGuide.stepPark', 'returnGuide.stepRfid', 'returnGuide.stepRetry'],
    part: 'rfid',
    primaryKey: 'returnGuide.returnNow',
    helpKey: 'returnGuide.helpRfid',
  },
  6: {
    titleKey: 'returnGuide.helmetTitle',
    bodyKey: 'returnGuide.helmetBody',
    stepKeys: ['returnGuide.stepHelmet', 'returnGuide.stepRetry'],
    part: 'helmet',
    primaryKey: 'returnGuide.returnNow',
    helpKey: 'returnGuide.helpHelmet',
  },
  8: {
    titleKey: 'returnGuide.kickTitle',
    bodyKey: 'returnGuide.kickBody',
    stepKeys: ['returnGuide.stepKick', 'returnGuide.stepRetry'],
    part: 'kickstand',
    primaryKey: 'returnGuide.returnNow',
    helpKey: 'returnGuide.helpKick',
  },
  11: {
    titleKey: 'returnGuide.cameraTitle',
    bodyKey: 'returnGuide.cameraBody',
    stepKeys: ['returnGuide.stepPark', 'returnGuide.stepCamera', 'returnGuide.stepRetry'],
    part: 'camera',
    primaryKey: 'returnGuide.returnNow',
    helpKey: 'returnGuide.helpCamera',
  },
}

const conf = computed(() => GUIDE[pageType.value] || GUIDE[4])
const title = computed(() => t(conf.value.titleKey))
const body = computed(() => t(conf.value.bodyKey))
const steps = computed(() => (conf.value.stepKeys || []).map((k) => t(k)))
const primaryText = computed(() =>
  pageType.value === 6 && !canReturn.value ? t('ride.unlockHelmet') : t(conf.value.primaryKey),
)
const subText = computed(() => {
  if (matchHint.value) return matchHint.value
  return canReturn.value ? t('returnGuide.matched') : t('returnGuide.notMatched')
})
const helpQuestion = computed(() => t(conf.value.helpKey))

const warnIcon = computed(() => getIconCfg('tipsWarning'))
const artMain = computed(() => {
  const pt = pageType.value
  if (pt === 0) return getIconCfg('helmetLock')
  if (pt === 4) return getIconCfg('ebikePlaceholder')
  if (pt === 5) return canReturn.value ? getIconCfg('ebikeReturnSuccess') : getIconCfg('ebikeReturnFail')
  if (pt === 6) return getIconCfg('helmetUnlockGif')
  if (pt === 8) return canReturn.value ? getIconCfg('kickstand_green') : getIconCfg('kickstand_gray')
  if (pt === 11) return getIconCfg('cameraReturnCar')
  return ''
})
const artOverlay = computed(() => {
  if (pageType.value === 4) return getIconCfg('ebikeLookdown')
  if (pageType.value === 5) return getIconCfg('ebikeLookdownInductor')
  return ''
})
const hasGuideArt = computed(() => Boolean(artMain.value || artOverlay.value || warnIcon.value))

/** Legacy isShowTakePhoto: free-penalty switch per accessory; camera uses audit quota */
const showPhotoApply = computed(() => {
  switch (pageType.value) {
    case 4:
      return Boolean(freeCfg.value.direction)
    case 5:
      return Boolean(freeCfg.value.rfidBeacon)
    case 6:
      return Boolean(freeCfg.value.helmet)
    case 8:
      return Boolean(freeCfg.value.kickstand)
    case 11:
      return Boolean(cameraAuditOk.value)
    default:
      return false
  }
})

onShow(() => {
  setNavTitle(t('returnGuide.title'))
  startPoll()
})
onHide(() => stopAll())
onUnmounted(() => stopAll())

onLoad((q) => {
  pageType.value = Number(q?.pageType || 0)
  orderId.value = decodeURIComponent(String(q?.orderId || ''))
  carId.value = decodeURIComponent(String(q?.carId || ''))
  returnType.value = decodeURIComponent(String(q?.returnType || ''))
  void loadFreeConfig()
})

function stopAll() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  if (cameraTimer) {
    clearInterval(cameraTimer)
    cameraTimer = null
  }
}

function startPoll() {
  stopAll()
  if (!conf.value.part) return
  void pollMatch()
  pollTimer = setInterval(() => {
    void pollMatch()
  }, 5000)
  if (pageType.value === 11 && freeCfg.value.camera) {
    void pollCameraAudit()
    cameraTimer = setInterval(() => {
      void pollCameraAudit()
    }, 5000)
  }
}

async function loadFreeConfig() {
  const serviceId = String(storage.get('serviceId', '') || '')
  try {
    const res = await getReturnCarConfig(serviceId ? { serviceId } : {})
    const data = (res.data || {}) as Record<string, unknown>
    if (res.success) {
      freeCfg.value = {
        direction: Boolean(data.direction),
        rfidBeacon: Boolean(data.rfidBeacon),
        helmet: Boolean(data.helmet),
        kickstand: Boolean(data.kickstand),
        camera: Boolean(data.camera),
      }
      if (pageType.value === 11 && freeCfg.value.camera) {
        void pollCameraAudit()
        if (!cameraTimer) {
          cameraTimer = setInterval(() => {
            void pollCameraAudit()
          }, 5000)
        }
      }
    }
  } catch (e) {
    logger.warn('getReturnCarConfig soft fail', e)
  }
}

async function pollCameraAudit() {
  if (!carId.value || !orderId.value) return
  try {
    const serviceId = String(storage.get('serviceId', '') || '')
    const res = await isCanCameraAudit({
      serviceId,
      carId: carId.value,
      orderId: orderId.value,
    })
    if (res.success) cameraAuditOk.value = Boolean(res.data)
  } catch (e) {
    logger.warn('isCanCameraAudit soft fail', e)
  }
}

async function pollMatch() {
  const part = conf.value.part
  if (!part || !carId.value) return
  try {
    const res = await partMatch({
      carId: carId.value,
      orderId: orderId.value,
      partNames: [part],
    })
    const data = (res.data || {}) as {
      izMatch?: boolean
      ext?: { izDirection?: boolean; angle?: number }
    }
    if (!res.success) {
      if (res.msg) uni.showToast({ title: String(res.msg), icon: 'none' })
      return
    }
    canReturn.value = Boolean(data.izMatch)
    if (pageType.value === 4) {
      const angle = Number(data.ext?.angle ?? 0)
      const izDirection = Boolean(data.ext?.izDirection)
      if (angle < 0 && !izDirection) {
        uni.showToast({ title: t('returnGuide.directionUnavailable'), icon: 'none' })
      }
      bikeAngle.value = canReturn.value ? 0 : angle || 0
    }
    if (pageType.value === 8 && canReturn.value) {
      matchHint.value = t('returnGuide.kickOk')
    } else if (pageType.value === 11 && canReturn.value) {
      matchHint.value = t('returnGuide.cameraOk')
    } else {
      matchHint.value = ''
    }
  } catch (e) {
    logger.warn('partMatch soft fail', e)
  }
}

async function onPrimary() {
  if (pageType.value === 6 && !canReturn.value) {
    await onHelmet()
    return
  }
  if (!canReturn.value) {
    uni.showToast({ title: t('returnGuide.notMatched'), icon: 'none' })
    return
  }
  storage.set('autoLock', true)
  navigate('reLaunch', '/pages/riding/riding')
}

async function onHelmet() {
  if (!carId.value) return
  const res = await unlockHelmet({ carId: carId.value })
  if (res.success) {
    uni.showToast({ title: t('ride.unlockHelmet'), icon: 'success' })
    void pollMatch()
  }
}

function goPark() {
  navigate('to', '/pages-sub/ride/park-search/park-search')
}

function goApply() {
  const applyType = guideApplyTypeFrom(pageType.value)
  navigate(
    'to',
    `/pages-sub/ride/apply-return/apply-return?orderId=${encodeURIComponent(orderId.value)}&applyType=${applyType}`,
  )
}

function goService() {
  navigate('to', '/pages-sub/support/customer-service/customer-service')
}

function goBack() {
  navigate('back')
}
</script>

<style scoped lang="scss">
.title {
  font-size: 36rpx;
  font-weight: 700;
  margin-bottom: 12rpx;
}
.sub {
  color: #e65c00;
  margin-bottom: 12rpx;
  &.ok {
    color: #2a9d4a;
  }
}
.body {
  color: #666;
  line-height: 1.5;
  margin-bottom: 20rpx;
}
.tips_bg {
  position: relative;
  width: 100%;
  min-height: 280rpx;
  margin: 12rpx 0 24rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.warn {
  width: 40rpx;
  margin-bottom: 12rpx;
}
.tip_img {
  width: 70%;
  max-width: 480rpx;
}
.tip_ebike {
  position: absolute;
  width: 36%;
  max-width: 220rpx;
  top: 42%;
  left: 50%;
  margin-left: -18%;
  transition: transform 1s ease;
  transform-origin: center center;
}
.step {
  color: #888;
  font-size: 26rpx;
  margin-bottom: 8rpx;
}
.btn-primary,
.btn-ghost {
  margin-top: 16rpx;
}
.btn-primary.disabled {
  opacity: 0.45;
}
.help {
  margin-top: 28rpx;
  text-align: center;
  font-size: 26rpx;
}
.help-q {
  color: #888;
  margin-right: 12rpx;
}
.link {
  color: #3aa0e8;
}
</style>
