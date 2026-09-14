<template>
  <!-- Legacy protocolsCustom + x-protocols -->
  <scroll-view scroll-y class="scroll-view">
    <view v-if="nodes" class="content">
      <rich-text :nodes="nodes" />
    </view>
    <view v-else-if="empty" class="empty-content">
      <image v-if="emptyIcon" class="image" :src="emptyIcon" mode="aspectFit" />
      <text>{{ t('common.empty') }}</text>
    </view>
  </scroll-view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getProtocolsByType } from '@/api/protocol'
import { formatRichText } from '@/shared/format'
import { setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { getIconCfg } from '@/shared/tenantSkin'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const nodes = ref('')
const empty = ref(false)
const emptyIcon = computed(() => getIconCfg('no_message') || getIconCfg('notLogin'))

onLoad(async (q) => {
  let title = t('account.protocolDetail')
  let type = 1
  try {
    if (q?.protocol) {
      const protocol = JSON.parse(decodeURIComponent(String(q.protocol))) as {
        type?: number
        title?: string
      }
      type = Number(protocol.type || 1)
      if (protocol.title) title = protocol.title
    }
  } catch (e) {
    logger.warn('protocol detail parse fail', e)
  }
  setNavTitle(title)
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getProtocolsByType({
      serviceId: sid || 0,
      type,
    })
    if (res.success && res.data) {
      const data = res.data as { content?: string }
      nodes.value = formatRichText(data.content || '')
      empty.value = !nodes.value
    } else {
      empty.value = true
      uni.showToast({ title: t('common.loadFail'), icon: 'none' })
    }
  } catch (e) {
    logger.warn('getProtocolsByType fail', e)
    empty.value = true
    uni.showToast({ title: t('common.loadFail'), icon: 'none' })
  } finally {
    uni.hideLoading()
  }
})
</script>

<style scoped lang="scss">
.scroll-view {
  height: 100vh;
  background: #fff;
}
.content {
  background: #fff;
  padding: 40rpx;
  font-size: 28rpx;
  line-height: 1.6;
  color: #333;
}
.empty-content {
  background: #fff;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #666;
  flex-direction: column;
}
.image {
  width: 300rpx;
  height: 300rpx;
  margin-bottom: 24rpx;
}
</style>
