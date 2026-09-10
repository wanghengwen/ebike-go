<template>
  <view v-if="visible" class="mask" @click.self="$emit('close')">
    <view class="panel">
      <view class="panel__title">{{ title }}</view>
      <view class="panel__body"><slot /></view>
      <view class="panel__actions">
        <view v-if="showCancel" class="btn-ghost" @click="$emit('cancel')">{{ cancelText }}</view>
        <view class="btn-primary" @click="$emit('confirm')">{{ confirmText }}</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    visible: boolean
    title: string
    confirmText: string
    cancelText?: string
    showCancel?: boolean
  }>(),
  { showCancel: true, cancelText: '' },
)
defineEmits<{
  (e: 'close'): void
  (e: 'cancel'): void
  (e: 'confirm'): void
}>()
</script>

<style scoped lang="scss">
.mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.panel {
  width: 620rpx;
  background: #fff;
  border-radius: 20rpx;
  padding: 32rpx;
}
.panel__title {
  font-size: 32rpx;
  font-weight: 600;
  margin-bottom: 16rpx;
}
.panel__body {
  color: #666;
  margin-bottom: 28rpx;
  line-height: 1.5;
}
.panel__actions {
  display: flex;
  gap: 20rpx;
}
.panel__actions .btn-ghost,
.panel__actions .btn-primary {
  flex: 1;
}
</style>
