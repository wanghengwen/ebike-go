<template>
  <view class="page">
    <view class="title">{{ t('account.authReviewing') }}</view>
    <image v-if="reviewIcon" class="img" :src="reviewIcon" mode="aspectFit" />
    <view class="tip">{{ t('account.authReviewTip') }}</view>
    <view class="btn-primary" @click="goService">{{ t('support.customerService') }}</view>
    <view class="link" @click="onCancel">{{ t('account.cancelAuth') }}</view>
    <view class="btn-ghost" @click="goHome">{{ t('common.back') }}</view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { changeBindCancel } from '@/api/user'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const reviewIcon = computed(() => getIconCfg('underReview'))

onShow(() => setNavTitle(t('account.authReviewing')))

function goService() {
  navigate('to', '/pages-sub/support/customer-service/customer-service')
}

function goHome() {
  navigate('reLaunch', '/pages/home/home')
}

async function onCancel() {
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const res = await changeBindCancel()
    if (res.success) {
      navigate('reLaunch', '/pages/home/home')
    } else {
      uni.showToast({ title: res.msg || t('common.networkError'), icon: 'none' })
    }
  } finally {
    uni.hideLoading()
  }
}
</script>

<style scoped lang="scss">
.page {
  padding: 64rpx 48rpx;
  text-align: center;
}
.title {
  font-size: 40rpx;
  font-weight: 700;
  margin-bottom: 48rpx;
}
.img {
  width: 280rpx;
  height: 280rpx;
  margin: 0 auto 32rpx;
  display: block;
}
.tip {
  color: #666;
  font-size: 28rpx;
  margin-bottom: 48rpx;
  line-height: 1.5;
}
.link {
  margin: 32rpx 0;
  color: #999;
  font-size: 28rpx;
}
.btn-ghost {
  margin-top: 16rpx;
}
</style>
