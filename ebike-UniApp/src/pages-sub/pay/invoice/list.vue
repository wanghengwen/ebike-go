<template>
  <!-- Legacy selectInvoiceType：出行服务发票 / 开票历史 -->
  <view class="page">
    <view class="items">
      <view class="item" @click="goCreate">
        <text class="label">{{ t('pay.invoiceTravelService') }}</text>
        <image v-if="arrowIcon" class="arrow" :src="arrowIcon" mode="aspectFit" />
        <text v-else class="arrow-text">›</text>
      </view>
      <view class="line" />
      <view class="item" @click="goHistory">
        <text class="label">{{ t('pay.invoiceHistory') }}</text>
        <image v-if="arrowIcon" class="arrow" :src="arrowIcon" mode="aspectFit" />
        <text v-else class="arrow-text">›</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getMapCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const arrowIcon = computed(() => getMapCfg('iconRight') || getMapCfg('iconRightRound'))

onShow(() => {
  setNavTitle(t('pay.invoiceCreate'))
})

function goCreate() {
  navigate('to', '/pages-sub/pay/invoice/create')
}

function goHistory() {
  navigate('to', '/pages-sub/pay/invoice/history')
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #fff;
}
.items {
  padding: 32rpx 48rpx 0;
}
.item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.label {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.arrow {
  width: 28rpx;
  height: 28rpx;
}
.arrow-text {
  font-size: 36rpx;
  color: #999;
  line-height: 1;
}
.line {
  height: 2rpx;
  margin: 48rpx 0 46rpx;
  background: #f6f6f6;
}
</style>
