<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('support.objection') }}</view>
      <view class="meta" v-if="orderId">{{ t('account.orders') }}：{{ orderId }}</view>
      <view class="label">{{ t('support.objectionType') }}</view>
      <view class="tags">
        <view
          v-for="(item, index) in reasons"
          :key="item.value"
          class="tag"
          :class="{ on: currIndex === index }"
          @click="currIndex = index"
        >
          {{ t(item.titleKey) }}
        </view>
      </view>
      <textarea class="area" v-model="content" :placeholder="t('support.objectionDesc')" maxlength="50" />
      <view class="btn-primary" @click="onSubmit">{{ t('common.confirm') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { createUserTicket } from '@/api/order'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const content = ref('')
const orderId = ref('')
const currIndex = ref<number | null>(null)

const reasons = [
  { titleKey: 'support.objUnlockFail', value: 3 },
  { titleKey: 'support.objBikeFault', value: 4 },
  { titleKey: 'support.objNoParking', value: 1 },
  { titleKey: 'support.objOutService', value: 0 },
  { titleKey: 'support.objParkPoint', value: 2 },
]

onShow(() => setNavTitle(t('support.objection')))

onLoad((q) => {
  orderId.value = decodeURIComponent((q?.orderId as string) || '')
})

async function onSubmit() {
  if (!orderId.value) {
    uni.showToast({ title: t('account.orders'), icon: 'none' })
    return
  }
  if (currIndex.value == null) {
    uni.showToast({ title: t('support.objectionType'), icon: 'none' })
    return
  }
  const reason = reasons[currIndex.value]
  const res = await createUserTicket({
    photoUrl: ['nophoto'],
    orderId: orderId.value,
    initiator: 0,
    userReason: `${t(reason.titleKey)}、${content.value.trim()}`,
  })
  if (res.success) {
    uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
    setTimeout(() => navigate('back'), 500)
  }
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.meta {
  color: #888;
  margin-bottom: 16rpx;
}
.label {
  color: #666;
  margin-bottom: 12rpx;
}
.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 20rpx;
}
.tag {
  padding: 12rpx 20rpx;
  background: #f5f6f8;
  border-radius: 8rpx;
  font-size: 24rpx;
}
.tag.on {
  background: #3aa0e8;
  color: #fff;
}
.area {
  width: 100%;
  min-height: 200rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-bottom: 28rpx;
  box-sizing: border-box;
}
</style>
