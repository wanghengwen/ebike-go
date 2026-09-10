<template>
  <view v-if="visible" class="mask" @click.self="onClose">
    <!-- Full pile dialog -->
    <view v-if="isDialog" class="dialog">
      <view class="dialog__title">{{ titleText }}</view>
      <view class="dialog__body">{{ reasonText }}</view>
      <view class="btn-primary" @click="onClose">{{ t('common.confirm') }}</view>
    </view>

    <!-- Bottom sheet -->
    <view v-else class="sheet">
      <view v-if="returnType === 10001" class="out-banner">
        <text>{{ t('returnSheet.outServiceBanner') }}</text>
        <text v-if="canTempUnlock" class="link" @click="$emit('recoverPower')">
          {{ t('ride.tempUnlock') }}
        </text>
      </view>
      <view class="reason">{{ reasonText }}</view>
      <view class="tips">{{ tipsText }}</view>
      <image v-if="coverImage" class="cover" :src="coverImage" mode="widthFix" />
      <view class="btns">
        <view class="btn-ghost" @click="$emit('refresh')">{{ t('returnSheet.refresh') }}</view>
        <view
          v-if="canReturn"
          class="btn-primary"
          @click="$emit('payDispatch')"
        >
          {{ t('returnSheet.payDispatch') }}
        </view>
      </view>
      <view class="btn-ghost near" @click="$emit('nearPark')">{{ t('returnSheet.nearPark') }}</view>
      <view
        v-if="showApply"
        class="apply"
        @click="$emit('apply')"
      >
        {{ t('returnSheet.locationWrong') }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getReturnCoverImage } from '@/features/bike/ridingFenceTips'
import { getReturnMeta } from '@/features/bike/returnTypes'

const props = withDefaults(
  defineProps<{
    visible: boolean
    returnType?: number | string
    penalty?: number
    canReturn?: boolean
    showApply?: boolean
    canTempUnlock?: boolean
    autoLockMinutes?: number
  }>(),
  {
    returnType: 2101,
    penalty: 0,
    canReturn: false,
    showApply: false,
    canTempUnlock: false,
    autoLockMinutes: 3,
  },
)

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'payDispatch'): void
  (e: 'refresh'): void
  (e: 'nearPark'): void
  (e: 'apply'): void
  (e: 'recoverPower'): void
}>()

const { t } = useI18n()

const meta = computed(() => getReturnMeta(props.returnType))
const isDialog = computed(() => meta.value.popupType === 'dialog')
const coverImage = computed(() => getReturnCoverImage(props.returnType))
const titleText = computed(() =>
  meta.value.titleKey ? t(meta.value.titleKey) : t('returnSheet.tipTitle'),
)
const reasonText = computed(() => t(meta.value.reasonKey))
const tipsText = computed(() => {
  const type = Number(props.returnType)
  const yuan = (Number(props.penalty || 0) / 100).toFixed(2)
  if (type === 10001) {
    return props.canReturn
      ? t('returnSheet.outServiceAutoPay', { min: props.autoLockMinutes, amount: yuan })
      : t('returnSheet.outServiceAuto', { min: props.autoLockMinutes })
  }
  if (props.canReturn) {
    const prefix = meta.value.outOfServiceKey ? `${t(meta.value.outOfServiceKey)}, ` : ''
    return `${prefix}${t('returnSheet.penaltyHint', { amount: yuan })}`
  }
  return t('returnSheet.rideToPark')
})

function onClose() {
  emit('close')
}
</script>

<style scoped lang="scss">
.mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 1100;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}
.dialog {
  align-self: center;
  width: 620rpx;
  background: #fff;
  border-radius: 20rpx;
  padding: 32rpx;
  margin-bottom: 20vh;
}
.dialog__title {
  font-weight: 700;
  font-size: 32rpx;
  margin-bottom: 16rpx;
  text-align: center;
}
.dialog__body {
  color: #666;
  line-height: 1.5;
  margin-bottom: 28rpx;
  text-align: center;
}
.sheet {
  width: 100%;
  background: #fff;
  border-radius: 32rpx 32rpx 0 0;
  padding: 32rpx 32rpx calc(32rpx + env(safe-area-inset-bottom));
  box-sizing: border-box;
}
.out-banner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff4e8;
  color: #ff8401;
  padding: 16rpx 20rpx;
  border-radius: 12rpx;
  margin-bottom: 20rpx;
  font-size: 24rpx;
}
.link {
  color: #3aa0e8;
}
.reason {
  font-size: 34rpx;
  font-weight: 700;
  text-align: center;
}
.tips {
  margin: 16rpx 0 28rpx;
  color: #666;
  text-align: center;
  line-height: 1.5;
}
.cover {
  display: block;
  width: 70%;
  margin: 0 auto 28rpx;
}
.btns {
  display: flex;
  gap: 20rpx;
  margin-bottom: 16rpx;
}
.btns > view {
  flex: 1;
}
.near {
  width: 100%;
  text-align: center;
}
.apply {
  margin-top: 24rpx;
  text-align: center;
  color: #3aa0e8;
  font-size: 26rpx;
}
</style>
