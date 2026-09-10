<template>
  <view class="page">
    <scroll-view scroll-y class="scroll">
      <view class="hero">
        <view class="logo" v-if="logo">
          <image class="logo__img" :src="logo" mode="aspectFit" />
        </view>
        <text class="name">{{ brandName }} V{{ version }}</text>
      </view>

      <view class="card body" v-if="loading">{{ t('common.loading') }}</view>
      <view class="card body" v-else-if="protocolHtml">
        <rich-text :nodes="protocolHtml" />
      </view>
      <view class="card body muted" v-else-if="aboutUrl">
        <text class="link" @click="openAboutUrl">{{ t('account.aboutProtocol') }}</text>
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
import { navigate, setNavTitle } from '@/shared/navigate'
import { getProtocolUrl, openProtocol } from '@/shared/protocol'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const loading = ref(false)
const protocolHtml = ref('')
const version = '1.0.0'

const logo = computed(() => getTenantConfig().customSetting?.logo || '')
const brandName = computed(() => getTenantConfig().name || t('brand.name'))
const aboutUrl = computed(() => getProtocolUrl('aboutUs') || getProtocolUrl('aboutProtocol') || '')

onShow(() => setNavTitle(t('account.about')))

async function loadProtocolBody() {
  loading.value = true
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getProtocolsByType({
      serviceId: sid || 0,
      type: 5,
    })
    if (res.success && res.data) {
      const data = res.data as { content?: string }
      protocolHtml.value = formatRichText(data.content || '')
    }
  } finally {
    loading.value = false
  }
}

function openAboutUrl() {
  if (getProtocolUrl('aboutUs')) openProtocol('aboutUs')
  else if (getProtocolUrl('aboutProtocol')) openProtocol('aboutProtocol')
  else if (aboutUrl.value) {
    navigate('to', `/pages/webview/webview?url=${encodeURIComponent(aboutUrl.value)}`)
  }
}

onMounted(() => {
  void loadProtocolBody()
})
</script>

<style scoped lang="scss">
.page {
  height: 100vh;
}
.scroll {
  height: 100%;
}
.hero {
  padding: 80rpx 48rpx 24rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.logo {
  width: 160rpx;
  height: 160rpx;
  border-radius: 32rpx;
  overflow: hidden;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.08);
  background: #fff;
}
.logo__img {
  width: 100%;
  height: 100%;
}
.name {
  margin-top: 48rpx;
  font-size: 32rpx;
  font-weight: 500;
  color: #333;
}
.body {
  margin: 24rpx 32rpx 48rpx;
}
.muted {
  color: #888;
  text-align: center;
}
.link {
  color: #3aa0e8;
}
</style>
