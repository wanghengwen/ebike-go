<template>
  <view class="page">
    <view class="tip">{{ t('common.loading') }}</view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { navigate } from '@/shared/navigate'

const { t } = useI18n()

onLoad((q) => {
  const orderId = decodeURIComponent(String(q?.orderId || ''))
  if (!orderId) return
  navigate(
    'redirect',
    `/pages-sub/pay/cost-detail/cost-detail?orderId=${encodeURIComponent(orderId)}`,
  )
})
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8f8f8;
}
.tip {
  color: #999;
}
</style>
