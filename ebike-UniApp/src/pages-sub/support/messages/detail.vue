<template>
  <view class="page">
    <view class="card">
      <view v-if="loading">{{ t('common.loading') }}</view>
      <template v-else>
        <view class="title">
          {{ detail.topic || detail.title || t('account.messages') }}
          <text v-if="typeLabel" class="type">-{{ typeLabel }}</text>
        </view>
        <view class="time">{{ detail.createdAt || detail.createTime || '' }}</view>
        <text class="body">{{ detail.content || t('common.empty') }}</text>
      </template>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getMsgDetail } from '@/api/message'
import { setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const loading = ref(false)
const detail = ref<Record<string, unknown>>({})

const typeLabel = computed(() => {
  const msgType = Number(detail.value.msgType)
  if (msgType === 1) return t('account.msgTypeSystem')
  if (msgType === 2) return t('account.msgTypeSms')
  if (msgType === 3) return t('account.msgTypePush')
  return ''
})

onShow(() => setNavTitle(t('account.messages')))

onLoad(async (q) => {
  const id = q?.id
  if (!id) return
  loading.value = true
  try {
    const res = await getMsgDetail({ id })
    if (res.success && res.data) detail.value = res.data as Record<string, unknown>
    else uni.showToast({ title: t('account.messagesLoadFail'), icon: 'none' })
  } catch {
    uni.showToast({ title: t('account.messagesLoadFail'), icon: 'none' })
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.title {
  font-size: 40rpx;
  font-weight: 800;
  text-align: center;
  color: #333;
}
.type {
  font-weight: 600;
}
.time {
  margin: 24rpx 0 32rpx;
  text-align: center;
  font-size: 28rpx;
  color: #999;
}
.body {
  font-size: 28rpx;
  color: #666;
  line-height: 1.6;
  word-break: break-all;
  white-space: pre-wrap;
}
</style>
