<template>
  <view class="scan-bar">
    <view class="scan-bar__row">
      <view
        class="scan-bar__enter"
        :class="{ 'is-disabled': disabled }"
        @click="onEnter"
      >
        <text>{{ enterLabel }}</text>
      </view>
      <view
        class="scan-bar__scan"
        :class="{ 'is-disabled': disabled }"
        @click="onScan"
      >
        <text>{{ scanLabel }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
const props = defineProps<{
  scanLabel: string
  enterLabel: string
  disabled?: boolean
}>()
const emit = defineEmits<{
  (e: 'scan'): void
  (e: 'enterId'): void
}>()

function onScan() {
  if (props.disabled) return
  emit('scan')
}
function onEnter() {
  if (props.disabled) return
  emit('enterId')
}
</script>

<style scoped lang="scss">
.scan-bar {
  padding: 24rpx 24rpx calc(24rpx + env(safe-area-inset-bottom));
}
.scan-bar__row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}
.scan-bar__enter {
  width: 36%;
  background: #fff;
  color: #333;
  border: 2rpx solid #d8d8d8;
  text-align: center;
  border-radius: 48rpx;
  padding: 26rpx 12rpx;
  font-size: 30rpx;
  font-weight: 600;
  box-sizing: border-box;
}
.scan-bar__scan {
  flex: 1;
  background: var(--brand-color, #3aa0e8);
  color: #fff;
  text-align: center;
  border-radius: 48rpx;
  padding: 26rpx 12rpx;
  font-size: 32rpx;
  font-weight: 600;
}
.scan-bar__enter.is-disabled,
.scan-bar__scan.is-disabled {
  opacity: 0.45;
  pointer-events: none;
}
</style>
