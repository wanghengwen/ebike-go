<template>
  <view v-if="type >= 0" class="cartip">
    <view v-if="type === 0 || type === 3" class="cartip__mask" />
    <view class="cartip__panel">
      <view class="cartip__car" v-if="carId">NO.{{ carId }}</view>
      <view class="cartip__title">
        <image v-if="iconSrc" class="cartip__icon" :src="iconSrc" mode="aspectFit" />
        <text>{{ titleText }}</text>
      </view>
      <view class="cartip__desc">
        <text v-for="(line, i) in descLines" :key="i" class="cartip__line">{{ line }}</text>
        <view v-if="type !== 2 || retryTimes <= 0" class="cartip__placeholder" />
      </view>
      <view v-if="type === 0 || type === 1" class="cartip__illust-wide">
        <image v-if="illustSrc" class="cartip__img-wide" :src="illustSrc" mode="aspectFit" />
      </view>
      <view v-else-if="illustSrc" class="cartip__illust">
        <image class="cartip__img" :src="illustSrc" mode="aspectFit" />
      </view>
      <view class="cartip__btns" v-if="type === 2">
        <view class="cartip__btn-white" @click="emit('changeBike')">{{ t('ride.changeBike') }}</view>
        <view v-if="retryTimes === 0" class="cartip__btn-primary" @click="onRetry">
          {{ t('ride.retryUnlock') }}
        </view>
        <view v-else-if="showRepair" class="cartip__btn-primary" @click="goRepair">
          {{ t('ride.reportFault') }}
        </view>
      </view>
      <view class="cartip__link" v-if="type === 5 && retryTimes === 0" @click="onRetry">
        {{ t('ride.retryLock') }}
      </view>
      <view class="cartip__link" v-else-if="type === 5 && retryTimes > 0" @click="goHelp">
        {{ t('support.customerService') }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getConfigBaseItem } from '@/api/user'
import { getCyclingCfg } from '@/shared/tenantSkin'
import { navigate } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const props = withDefaults(
  defineProps<{
    type?: number
    carId?: string
    isUseBle?: boolean
  }>(),
  {
    type: -1,
    carId: '',
    isUseBle: false,
  },
)

const emit = defineEmits<{
  (e: 'changeBike'): void
  (e: 'retry'): void
}>()

const { t } = useI18n()
const progress = ref(0)
const retryTimes = ref(0)
const showRepair = ref(true)
let timer: ReturnType<typeof setInterval> | null = null

const titles = computed(() => [
  `${props.isUseBle ? t('ride.blePrefix') : ''}${t('ride.unlocking')}${progress.value}%`,
  t('ride.unlockSuccess'),
  t('ride.unlockFail'),
  t('ride.locking'),
  t('ride.lockSuccess'),
  t('ride.lockFail'),
])

const titleText = computed(() => titles.value[props.type] || '')

const descLines = computed(() => {
  if (props.type === 0 || props.type === 1) return [t('ride.unlockSafetyTip')]
  if (props.type === 2 && retryTimes.value === 0) return [t('ride.unlockRetryTip')]
  if (props.type === 2) return [t('ride.unlockFaultTip'), t('ride.unlockRepairTip')]
  if (props.type === 3 || props.type === 4) return [t('ride.lockSafetyTip')]
  if (props.type === 5) return [t('ride.lockRetryTip')]
  return []
})

const iconSrc = computed(() => {
  if (props.type === 0) return getCyclingCfg('cartipUnlockLoading')
  if (props.type === 1) return getCyclingCfg('cartipUnlockSucIcon')
  return ''
})

const illustSrc = computed(() => {
  if (props.type === 0 || props.type === 1) return getCyclingCfg('cartipUnlockingNew')
  if (props.type === 2) return getCyclingCfg('cartipUnlockFail')
  if (props.type === 3) return getCyclingCfg('cartipLocking')
  if (props.type === 4) return getCyclingCfg('cartipLockSuc')
  if (props.type === 5) return getCyclingCfg('cartipLockFail')
  return ''
})

function clearTimer() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

function startProgress() {
  clearTimer()
  progress.value = 0
  timer = setInterval(() => {
    if (progress.value >= 98) {
      clearTimer()
      return
    }
    progress.value += 1
  }, Math.max(20, Math.floor((2 / 98) * 1000)))
}

watch(
  () => props.type,
  (val) => {
    if (val === 0) startProgress()
    else if (val === 1) {
      progress.value = 100
      clearTimer()
    } else {
      progress.value = 0
      clearTimer()
    }
    if (val === -1) retryTimes.value = 0
  },
)

onUnmounted(clearTimer)

void (async () => {
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getConfigBaseItem(sid ? { serviceId: sid } : {})
    if (res.success && res.data) {
      const data = res.data as { izCanAfterRidingRepair?: boolean }
      if (data.izCanAfterRidingRepair != null) showRepair.value = Boolean(data.izCanAfterRidingRepair)
    }
  } catch (e) {
    logger.warn('CarTipSheet config soft fail', e)
  }
})()

function onRetry() {
  retryTimes.value += 1
  emit('retry')
}

function goRepair() {
  const q = props.carId ? `?carId=${encodeURIComponent(props.carId)}` : ''
  navigate('to', `/pages-sub/support/repair/repair${q}`)
}

function goHelp() {
  const q = props.carId ? `?carId=${encodeURIComponent(props.carId)}` : ''
  navigate('to', `/pages-sub/support/help/help${q}`)
}
</script>

<style scoped lang="scss">
.cartip {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1200;
}
.cartip__mask {
  position: fixed;
  left: 0;
  right: 0;
  top: 0;
  bottom: 706rpx;
  background: rgba(0, 0, 0, 0.3);
  z-index: 1199;
}
.cartip__panel {
  position: relative;
  z-index: 1201;
  width: 100%;
  height: 666rpx;
  background: #fff;
  border-radius: 32rpx 32rpx 0 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  box-sizing: border-box;
  padding-bottom: env(safe-area-inset-bottom);
}
.cartip__car {
  position: absolute;
  top: 32rpx;
  left: 32rpx;
  color: #666;
  font-size: 24rpx;
}
.cartip__title {
  margin-top: 98rpx;
  display: flex;
  align-items: center;
  font-size: 36rpx;
  font-weight: 600;
  color: #333;
}
.cartip__icon {
  width: 40rpx;
  height: 40rpx;
  margin-right: 20rpx;
}
.cartip__desc {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin: 16rpx 0 20rpx;
}
.cartip__line {
  display: block;
  color: #666;
  font-size: 32rpx;
  line-height: 1.4;
}
.cartip__placeholder {
  height: 40rpx;
}
.cartip__illust-wide {
  width: 750rpx;
  height: 416rpx;
}
.cartip__img-wide {
  width: 100%;
  height: 100%;
}
.cartip__illust {
  width: 300rpx;
  height: 300rpx;
}
.cartip__img {
  width: 100%;
  height: 100%;
}
.cartip__btns {
  display: flex;
  width: 100%;
  justify-content: space-between;
  padding: 0 32rpx;
  box-sizing: border-box;
  margin-top: 8rpx;
}
.cartip__btn-white,
.cartip__btn-primary {
  font-size: 32rpx;
  border-radius: 30rpx;
  line-height: 80rpx;
  padding: 8rpx 40rpx;
  text-align: center;
  box-sizing: border-box;
}
.cartip__btn-white {
  background: #fff;
  border: 2rpx solid #ccc;
  color: #333;
  min-width: 200rpx;
}
.cartip__btn-primary {
  width: 428rpx;
  background: var(--brand-color, #3aa0e8);
  color: #fff;
}
.cartip__link {
  margin-top: 16rpx;
  color: #0b9ffe;
  font-size: 24rpx;
}
</style>
