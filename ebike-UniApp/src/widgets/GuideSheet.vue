<template>
  <view v-if="visible && pages.length" class="guide-mask">
    <view class="guide-panel">
      <swiper
        class="guide-swiper"
        :style="{ height: swiperH }"
        :current="current"
        :indicator-dots="pages.length > 1"
        indicator-color="rgba(0,0,0,0.35)"
        indicator-active-color="#ffffff"
        @change="onChange"
      >
        <swiper-item v-for="(item, idx) in pages" :key="idx">
          <view class="guide-item" @tap="onItemClick(item)">
            <image
              v-if="item.picUrl"
              class="guide-img"
              :src="String(item.picUrl)"
              mode="aspectFit"
              :style="{ height: imgH }"
            />
          </view>
        </swiper-item>
      </swiper>
      <view class="guide-foot" v-if="showClose">
        <view class="guide-cta" @tap="emit('close')">
          {{ ctaText }}
        </view>
      </view>
      <view class="guide-x" v-else-if="allowEarlyClose" @tap="emit('close')">
        <image v-if="closeIcon" class="guide-x__img" :src="closeIcon" mode="aspectFit" />
        <text v-else>×</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getIconCfg } from '@/shared/tenantSkin'
import { openScrollerNotice } from '@/shared/openNotice'

const props = defineProps<{
  visible: boolean
  config: Record<string, unknown>
}>()

const emit = defineEmits<{ (e: 'close'): void }>()
const { t } = useI18n()
const current = ref(0)
const screenHeightRpx = ref(1334)
const closeIcon = computed(() => getIconCfg('iconClose'))

/** Legacy: image height = windowHeight(rpx) * 0.5, width 564rpx */
const imgH = computed(() => `${Math.round(screenHeightRpx.value * 0.5)}rpx`)
const swiperH = computed(() => imgH.value)

const pages = computed(
  () =>
    (Array.isArray(props.config?.guidePages) ? props.config.guidePages : []) as Array<
      Record<string, unknown>
    >,
)

watch(
  () => props.visible,
  (v) => {
    if (v) current.value = 0
  },
)

onMounted(() => {
  try {
    const info = uni.getSystemInfoSync()
    const w = Number(info.windowWidth || 375)
    const h = Number(info.windowHeight || 667)
    screenHeightRpx.value = Math.round((h * 750) / w)
  } catch {
    screenHeightRpx.value = 1334
  }
})

const atLast = computed(() => current.value >= Math.max(pages.value.length - 1, 0))
const allowEarlyClose = computed(() => {
  const allow = Boolean(props.config?.allowSuperEsc)
  const need = Number(props.config?.pageNumEsc || 0)
  return allow && current.value + 1 >= need
})
const showClose = computed(() => pages.value.length <= 1 || atLast.value || allowEarlyClose.value)

const ctaText = computed(() => {
  if (pages.value.length === 1 || (Boolean(props.config?.allowSuperEsc) && atLast.value)) {
    return t('ride.startUse')
  }
  return t('common.gotIt')
})

function onChange(e: { detail?: { current?: number } }) {
  current.value = Number(e?.detail?.current || 0)
}

function onItemClick(item: Record<string, unknown>) {
  openScrollerNotice({
    skipUrl: String(item.linkUrl || ''),
    appId: String(item.appId || ''),
    params: item.param,
    title: String(item.linkTitle || ''),
    type: Number(item.chainType != null ? Number(item.chainType) + 1 : 1),
  })
}
</script>

<style scoped lang="scss">
.guide-mask {
  position: fixed;
  inset: 0;
  z-index: 1290;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
}
.guide-panel {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.guide-swiper {
  width: 100%;
}
.guide-item {
  height: 100%;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.guide-img {
  width: 564rpx;
}
.guide-foot {
  display: flex;
  justify-content: center;
  margin-top: 40rpx;
  height: 80rpx;
  align-items: center;
}
.guide-cta {
  font-size: 32rpx;
  font-weight: 500;
  color: #fff;
  border: 4rpx solid #eee;
  border-radius: 40rpx;
  padding: 12rpx 32rpx;
  line-height: 1.2;
}
.guide-x {
  margin-top: 40rpx;
  width: 60rpx;
  height: 60rpx;
  border-radius: 50%;
  border: 4rpx solid #eee;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  box-sizing: border-box;
  padding: 10rpx;
}
.guide-x__img {
  width: 40rpx;
  height: 40rpx;
}
</style>
