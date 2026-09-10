<template>
  <view v-if="visible && popupType" class="mask" @click.self="onKnow">
    <view class="sheet">
      <view class="title">{{ titleText }}</view>
      <view v-if="subText" class="sub">{{ subText }}</view>
      <image v-if="imgSrc" class="img" :src="imgSrc" mode="widthFix" />
      <view class="btns">
        <view v-if="popupType === 1" class="btn-ghost" @click="onCancel">{{ t('helmet.cancelRide') }}</view>
        <view v-if="popupType === 1" class="btn-primary" @click="onUnlockClick">{{ t('helmet.unlock') }}</view>
        <view
          v-if="[2, 101, 102].includes(popupType)"
          class="btn-primary full"
          @click="onKnow"
        >
          {{ knowText }}
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getCyclingCfg, getIconCfg } from '@/shared/tenantSkin'

const props = withDefaults(
  defineProps<{
    visible?: boolean
    popupType?: number
    carId?: string
  }>(),
  {
    visible: false,
    popupType: 0,
    carId: '',
  },
)

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'cancel'): void
  /** Precycling type1: parent shows 100 then opens bike (backend unlocks helmet). */
  (e: 'unlock'): void
}>()

const { t } = useI18n()
const seconds = ref(0)
let timer: ReturnType<typeof setInterval> | null = null
let lastCloseAt = 0

const titleText = computed(() => {
  switch (props.popupType) {
    case 1:
      return t('helmet.needUnlock')
    case 2:
      return t('helmet.needWear')
    case 100:
      return t('helmet.unlocking')
    case 101:
      return t('helmet.unlocked')
    case 102:
      return t('helmet.tempReturn')
    default:
      return ''
  }
})
const subText = computed(() => (props.popupType === 101 ? t('helmet.returnFirst') : ''))
const knowText = computed(() =>
  seconds.value > 0 ? `${t('helmet.gotIt')}（${seconds.value}s）` : t('helmet.gotIt'),
)
const imgSrc = computed(() => {
  if (props.popupType === 100) return getCyclingCfg('helmetUnlocking')
  if (props.popupType === 102) return getIconCfg('helmetUnlockGif')
  if ([1, 2, 101].includes(props.popupType)) return getCyclingCfg('wearHelmetTip')
  return ''
})

watch(
  () => [props.visible, props.popupType] as const,
  ([vis, type]) => {
    stopTimer()
    if (!vis) return
    if (type === 2 && lastCloseAt && Date.now() - lastCloseAt < 10000) {
      emit('close')
      return
    }
    if (type === 101) {
      seconds.value = 3
      timer = setInterval(() => {
        if (seconds.value <= 1) {
          onKnow()
          return
        }
        seconds.value -= 1
      }, 1000)
    }
  },
)

onUnmounted(() => stopTimer())

function stopTimer() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
  seconds.value = 0
}

function onKnow() {
  if (props.popupType === 100) return
  lastCloseAt = Date.now()
  stopTimer()
  emit('close')
}

function onCancel() {
  stopTimer()
  emit('cancel')
  emit('close')
}

function onUnlockClick() {
  emit('unlock')
}
</script>

<style scoped lang="scss">
.mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 1200;
  display: flex;
  align-items: center;
  justify-content: center;
}
.sheet {
  width: 620rpx;
  background: #fff;
  border-radius: 24rpx;
  padding: 36rpx 32rpx 28rpx;
  box-sizing: border-box;
}
.title {
  font-size: 30rpx;
  font-weight: 600;
  text-align: center;
  line-height: 1.5;
}
.sub {
  margin-top: 12rpx;
  color: #888;
  font-size: 24rpx;
  text-align: center;
}
.img {
  display: block;
  width: 70%;
  margin: 28rpx auto;
}
.btns {
  display: flex;
  gap: 20rpx;
  margin-top: 12rpx;
}
.btns > view {
  flex: 1;
  text-align: center;
}
.full {
  flex: 1;
}
</style>
