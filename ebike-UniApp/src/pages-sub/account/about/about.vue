<template>
  <!-- Legacy about.vue -->
  <view class="page">
    <scroll-view scroll-y class="scroll-view">
      <view class="container">
        <view class="logo">
          <image v-if="logo" class="img" :src="logo" mode="widthFix" />
        </view>
        <text class="name">{{ brandName }} V{{ version }}</text>
        <view v-if="loading" class="loading">{{ t('common.loading') }}</view>
        <view v-else-if="protocolHtml" class="protocol">
          <rich-text :nodes="protocolHtml" />
        </view>
        <view v-else class="empty">
          <image v-if="emptyIcon" class="empty-img" :src="emptyIcon" mode="aspectFit" />
          <text>{{ t('common.empty') }}</text>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getProtocolsByType } from '@/api/protocol'
import { getTenantConfig } from '@/shared/config'
import { formatRichText } from '@/shared/format'
import { setNavTitle } from '@/shared/navigate'
import { PROTOCOL_CONFIGS } from '@/shared/protocolConfigs'
import { storage } from '@/shared/storage'
import { getIconCfg } from '@/shared/tenantSkin'

const { t } = useI18n()
const loading = ref(false)
const protocolHtml = ref('')
const version = ref('1.0.0')
const logo = computed(() => getTenantConfig().customSetting?.logo || '')
const brandName = computed(() => getTenantConfig().name || t('brand.name'))
const emptyIcon = computed(() => getIconCfg('no_message') || getIconCfg('notLogin'))

onShow(() => setNavTitle(t('account.about')))

try {
  // #ifdef MP-WEIXIN
  const info = uni.getAccountInfoSync?.()
  const v = info?.miniProgram?.version
  if (v) version.value = v
  // #endif
} catch {
  /* keep default */
}

async function loadAbout() {
  loading.value = true
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getProtocolsByType({
      serviceId: sid || 0,
      type: PROTOCOL_CONFIGS.aboutUs.type,
    })
    if (res.success && res.data) {
      const data = res.data as { content?: string }
      protocolHtml.value = formatRichText(data.content || '')
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadAbout()
})
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #fff;
}
.scroll-view {
  height: 100%;
}
.container {
  padding: 80rpx 48rpx 48rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.logo {
  width: 160rpx;
  height: 160rpx;
}
.img {
  width: 100%;
  height: 100%;
}
.name {
  margin-top: 48rpx;
  font-size: 32rpx;
  font-weight: 500;
  color: #333;
}
.loading,
.empty {
  margin-top: 64rpx;
  color: #666;
  font-size: 28rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.empty-img {
  width: 300rpx;
  height: 300rpx;
  margin-bottom: 24rpx;
}
.protocol {
  margin-top: 48rpx;
  width: 100%;
  font-size: 28rpx;
  line-height: 1.6;
  color: #333;
}
</style>
