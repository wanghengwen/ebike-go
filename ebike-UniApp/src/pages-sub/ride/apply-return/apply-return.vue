<template>
  <view class="page">
    <view class="card" v-if="capable">
      <view class="title">{{ t('ride.applyReturn') }}</view>
      <view class="hint">{{ t('ride.applyReturnHint') }}</view>
      <view class="photos">
        <view v-for="(url, i) in imgs" :key="i" class="photo">
          <image :src="url" mode="aspectFill" class="img" />
          <view class="del" @click="imgs.splice(i, 1)">×</view>
        </view>
        <view v-if="imgs.length < 2" class="add" @click="pickImage">+</view>
      </view>
      <textarea class="area" v-model="reason" :placeholder="t('ride.applyReturnDesc')" maxlength="50" />
      <view class="btn-primary" :class="{ disabled: !canSubmit }" @click="onSubmit">{{ t('common.confirm') }}</view>
    </view>
    <view class="card" v-else>
      <view class="title">{{ t('ride.applyReturnPending') }}</view>
      <view class="hint">{{ t('ride.applyReturnBlocked') }}</view>
      <view class="btn-primary" @click="goHelp">{{ t('support.customerService') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { createReturnEbike, queryReturnBikeAudit } from '@/api/applyReturn'
import { uploadFile } from '@/api/common'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const reason = ref('')
const orderId = ref('')
const applyType = ref<number | string>('')
const imgs = ref<string[]>([])
const capable = ref(true)

const canSubmit = computed(() => imgs.value.length > 0 && reason.value.trim().length > 0)

onShow(() => setNavTitle(t('ride.applyReturn')))

onLoad(async (q) => {
  orderId.value = decodeURIComponent((q?.orderId as string) || '')
  applyType.value = q?.applyType != null ? Number(q.applyType) : ''
  const res = await queryReturnBikeAudit(
    applyType.value !== '' ? { applyType: applyType.value } : {},
  )
  if (res.success) {
    capable.value = Boolean(res.data)
  }
})

function pickImage() {
  uni.chooseImage({
    count: 2 - imgs.value.length,
    sizeType: ['compressed'],
    sourceType: ['camera'],
    success: async (res) => {
      for (const path of res.tempFilePaths || []) {
        const up = await uploadFile(path)
        const url = typeof up === 'string' ? up : (up?.data as string | undefined)
        if (url) imgs.value.push(url)
      }
    },
  })
}

function goHelp() {
  navigate('to', '/pages-sub/support/customer-service/customer-service')
}

async function onSubmit() {
  if (!canSubmit.value) return
  const res = await createReturnEbike({
    photoUrl: imgs.value.join(','),
    orderId: orderId.value || undefined,
    applyType: applyType.value !== '' ? applyType.value : undefined,
    userReason: reason.value.trim(),
  })
  if (res.success) {
    storage.set('autoLock', true)
    uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
    setTimeout(() => navigate('reLaunch', '/pages/riding/riding'), 500)
  }
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.hint {
  color: #666;
  margin-bottom: 20rpx;
  line-height: 1.5;
}
.area {
  width: 100%;
  min-height: 160rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-bottom: 28rpx;
  box-sizing: border-box;
}
.photos {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
}
.photo {
  position: relative;
  width: 160rpx;
  height: 160rpx;
}
.img {
  width: 100%;
  height: 100%;
  border-radius: 8rpx;
}
.del {
  position: absolute;
  top: -8rpx;
  right: -8rpx;
  width: 36rpx;
  height: 36rpx;
  background: #e34d59;
  color: #fff;
  border-radius: 50%;
  text-align: center;
  line-height: 36rpx;
}
.add {
  width: 160rpx;
  height: 160rpx;
  background: #f5f6f8;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48rpx;
  color: #999;
}
.disabled {
  opacity: 0.5;
}
</style>
