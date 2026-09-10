<template>
  <view v-if="visible" class="mask" @click.self="onClose">
    <view class="sheet">
      <view class="reason">{{ t('returnSheet.civilizationTitle') }}</view>
      <view class="tips">{{ t('returnSheet.civilizationTips') }}</view>
      <image v-if="parkTipsImg" class="park-img" :src="parkTipsImg" mode="widthFix" />
      <view class="btns">
        <view class="btn-ghost" @click="onClose">{{ t('common.cancel') }}</view>
        <view class="btn-primary" @click="$emit('confirm')">{{ t('returnSheet.returnNow') }}</view>
      </view>
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
import { getTenantConfig } from '@/shared/config'

withDefaults(
  defineProps<{
    visible: boolean
    showApply?: boolean
  }>(),
  { showApply: false },
)

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm'): void
  (e: 'apply'): void
}>()

const { t } = useI18n()
const parkTipsImg = computed(
  () => getTenantConfig().customSetting?.iconCfg?.parkTips || '',
)

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
}
.sheet {
  width: 100%;
  background: #fff;
  border-radius: 32rpx 32rpx 0 0;
  padding: 40rpx 32rpx calc(40rpx + env(safe-area-inset-bottom));
  box-sizing: border-box;
}
.reason {
  font-size: 34rpx;
  font-weight: 700;
  text-align: center;
}
.tips {
  margin: 20rpx 0 24rpx;
  color: #666;
  text-align: center;
  line-height: 1.5;
}
.park-img {
  width: 100%;
  margin-bottom: 24rpx;
}
.btns {
  display: flex;
  gap: 20rpx;
}
.btns > view {
  flex: 1;
}
.apply {
  margin-top: 28rpx;
  text-align: center;
  color: #3aa0e8;
  font-size: 26rpx;
}
</style>
