<template>
  <view v-if="visible && config && Object.keys(config).length" class="ops-mask">
    <view class="ops-wrap">
      <view
        v-if="!config.closePosition"
        class="ops-close ops-close--top"
        @click="emit('close')"
      >
        <image v-if="closeIcon" class="ops-close__img" :src="closeIcon" mode="aspectFit" />
        <text v-else class="ops-close__x">×</text>
      </view>

      <!-- Image popup: popUpType truthy -->
      <view v-if="isImage" class="ops-img" :style="imgStyle" @click.stop>
        <text class="ops-img__title" :style="{ color: String(config.titleColor || '#333') }">
          {{ config.title || '' }}
        </text>
        <text
          v-if="config.izSubtitle"
          class="ops-img__sub"
          :style="{ color: String(config.subtitleColor || '#666') }"
        >
          {{ config.subtitle || '' }}
        </text>
        <text class="ops-img__body" :style="{ color: String(config.bodyColor || '#666') }">
          {{ config.body || '' }}
        </text>
        <view
          v-if="config.izButton"
          class="ops-btn"
          :style="{
            color: String(config.buttonTextColor || '#fff'),
            background: String(config.buttonColor || 'var(--brand-color, #3aa0e8)'),
          }"
          @click="onAction"
        >
          {{ config.buttonText || t('common.confirm') }}
        </view>
      </view>

      <!-- Words popup -->
      <view v-else class="ops-words" :style="wordsBgStyle" @click.stop>
        <view class="ops-words__title">
          <text class="ops-words__main" :style="{ color: String(config.titleColor || '#333') }">
            {{ config.title || '' }}
          </text>
          <text
            v-if="config.izSubtitle"
            class="ops-words__sub"
            :style="{ color: String(config.subtitleColor || '#666') }"
          >
            {{ config.subtitle || '' }}
          </text>
        </view>
        <scroll-view scroll-y class="ops-words__body">
          <text :style="{ color: String(config.bodyColor || '#666') }">{{ config.body || '' }}</text>
        </scroll-view>
        <view
          v-if="config.izCheckRead"
          class="ops-check"
          @click="checked = !checked"
        >
          <text class="ops-check__box">{{ checked ? '☑' : '☐' }}</text>
          <text>{{ config.checkReadContent || '' }}</text>
        </view>
        <view
          v-if="config.izButton"
          class="ops-btn ops-btn--words"
          :class="{ disabled: config.izCheckRead && !checked }"
          :style="{
            color: String(config.buttonTextColor || '#fff'),
            background: String(config.buttonColor || 'var(--brand-color, #3aa0e8)'),
          }"
          @click="onAction"
        >
          {{ config.buttonText || t('common.confirm') }}
        </view>
      </view>

      <view
        v-if="config.closePosition === 1"
        class="ops-close ops-close--bottom"
        @click="emit('close')"
      >
        <image v-if="closeIcon" class="ops-close__img" :src="closeIcon" mode="aspectFit" />
        <text v-else class="ops-close__x">×</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getIconCfg } from '@/shared/tenantSkin'
import { openJumpAction } from '@/shared/openNotice'

const props = defineProps<{
  visible: boolean
  config: Record<string, unknown>
}>()

const emit = defineEmits<{ (e: 'close'): void }>()
const { t } = useI18n()
const checked = ref(true)
const closeIcon = computed(() => getIconCfg('iconClose'))

watch(
  () => props.visible,
  (v) => {
    if (v) checked.value = true
  },
)

const bgList = computed(() => {
  const bg = props.config?.bgUrl
  return Array.isArray(bg) ? (bg as string[]) : bg ? [String(bg)] : []
})

const isImage = computed(() => Boolean(props.config?.popUpType))

const imgStyle = computed(() => ({
  backgroundImage: bgList.value[0] ? `url(${bgList.value[0]})` : undefined,
}))

const wordsBgStyle = computed(() => ({
  backgroundImage: bgList.value[0] ? `url(${bgList.value[0]})` : undefined,
}))

function onAction() {
  if (props.config.izCheckRead && !checked.value) return
  const clickEvent = Number(props.config.clickEvent ?? 0)
  const jump = (props.config.jumpPage || {}) as Record<string, unknown>
  if (clickEvent === 0) {
    emit('close')
    return
  }
  emit('close')
  openJumpAction(clickEvent, jump)
}
</script>

<style scoped lang="scss">
.ops-mask {
  position: fixed;
  inset: 0;
  z-index: 1300;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}
.ops-wrap {
  position: relative;
  width: 600rpx;
}
.ops-close {
  width: 60rpx;
  height: 60rpx;
  border-radius: 50%;
  border: 4rpx solid #eee;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.25);
}
.ops-close--top {
  position: absolute;
  top: -100rpx;
  right: -20rpx;
}
.ops-close--bottom {
  margin: 40rpx auto 0;
}
.ops-close__img {
  width: 40rpx;
  height: 40rpx;
}
.ops-close__x {
  color: #fff;
  font-size: 40rpx;
  line-height: 1;
}
.ops-img {
  width: 414rpx;
  min-height: 520rpx;
  margin: 0 auto;
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48rpx 24rpx 120rpx;
  position: relative;
}
.ops-img__title {
  font-size: 44rpx;
  font-weight: 700;
}
.ops-img__sub,
.ops-img__body {
  margin-top: 12rpx;
  text-align: center;
  white-space: pre-wrap;
  font-size: 26rpx;
}
.ops-words {
  min-height: 520rpx;
  background: #fff;
  border-radius: 24rpx;
  background-size: 100% 100%;
  background-repeat: no-repeat;
  padding: 40rpx 32rpx 48rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.ops-words__main {
  font-size: 40rpx;
  font-weight: 700;
  display: block;
  text-align: center;
}
.ops-words__sub {
  display: block;
  text-align: center;
  margin-top: 8rpx;
  font-size: 28rpx;
}
.ops-words__body {
  margin-top: 24rpx;
  max-height: 360rpx;
  width: 100%;
  font-size: 28rpx;
  line-height: 1.5;
  text-align: center;
}
.ops-check {
  margin-top: 24rpx;
  display: flex;
  align-items: center;
  gap: 12rpx;
  font-size: 24rpx;
  color: #666;
  align-self: flex-start;
}
.ops-check__box {
  font-size: 32rpx;
}
.ops-btn {
  margin-top: 28rpx;
  min-width: 200rpx;
  height: 70rpx;
  padding: 0 32rpx;
  border-radius: 54rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
}
.ops-btn--words.disabled {
  opacity: 0.45;
}
.ops-img .ops-btn {
  position: absolute;
  bottom: 40rpx;
  left: 50%;
  transform: translateX(-50%);
}
</style>
