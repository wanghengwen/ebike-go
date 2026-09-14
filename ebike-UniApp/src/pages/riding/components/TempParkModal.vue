<template>
  <view v-if="visible" class="mask" @click.self="emit('close')">
    <view class="riding-parking-popup">
      <image
        mode="widthFix"
        class="riding-parking-popup_img"
        :src="imgSrc"
      />
      <view class="riding-parking-popup_content">
        <text class="riding-parking-popup_title">{{ t('ride.tempParkSuccess') }}</text>
        <text class="riding-parking-popup_msg">
          {{ t('ride.tempParkBilling') }}
          <text v-if="showAutoReturn" class="riding-parking-popup_warn">
            {{ t('ride.tempParkAutoReturnMin', { m: tempParkTime }) }}
          </text>
        </text>
        <text class="riding-parking-popup_btn" :style="btnStyle" @click="emit('close')">
          {{ t('ride.tempParkGotIt') }}
        </text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getBrandColor } from '@/shared/config'

const props = withDefaults(
  defineProps<{
    visible: boolean
    tempParkTime?: number
    showAutoReturn?: boolean
    /** Optional tenant override; defaults to legacy illustration */
    imageUrl?: string
  }>(),
  {
    tempParkTime: 0,
    showAutoReturn: false,
    imageUrl: '',
  },
)

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()

const DEFAULT_IMG =
  'https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/tempParking.png'

const imgSrc = computed(() => props.imageUrl || DEFAULT_IMG)

const btnStyle = computed(() => ({
  background: getBrandColor(),
  color: '#1E4A38',
  fontWeight: 'bold' as const,
}))
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
.riding-parking-popup {
  box-sizing: border-box;
  width: 80vw;
  margin: 0 auto;
}
.riding-parking-popup_img {
  width: 100%;
  display: block;
  vertical-align: top;
}
.riding-parking-popup_content {
  padding: 64rpx 48rpx 32rpx;
  box-sizing: border-box;
  background-color: #fff;
  border-radius: 0 0 32rpx 32rpx;
  display: flex;
  flex-direction: column;
  text-align: center;
  margin-top: -40rpx;
  position: relative;
  z-index: 1;
}
.riding-parking-popup_title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333333;
  margin-bottom: 32rpx;
}
.riding-parking-popup_msg {
  font-size: 28rpx;
  font-weight: 400;
  color: #666666;
  line-height: 40rpx;
}
.riding-parking-popup_warn {
  color: #ffab2c;
}
.riding-parking-popup_btn {
  padding: 18rpx 48rpx;
  text-align: center;
  box-sizing: border-box;
  border-radius: 32rpx;
  margin-top: 40rpx;
  font-size: 28rpx;
}
</style>
