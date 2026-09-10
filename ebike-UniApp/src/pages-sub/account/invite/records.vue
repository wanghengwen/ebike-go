<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('account.inviteRecords') }}</view>
      <view v-if="loading">{{ t('common.loading') }}</view>
      <view v-else-if="!list.length">{{ t('common.empty') }}</view>
      <view v-for="(item, i) in list" :key="i" class="item">
        <view>{{ item.phone || item.invitee || item.userId || '-' }}</view>
        <view class="sub">{{ item.createTime || item.status || '' }}</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getInviteRecord } from '@/api/invite'
import { setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const loading = ref(false)
const list = ref<Array<Record<string, unknown>>>([])

onShow(() => setNavTitle(t('account.inviteRecords')))

onMounted(async () => {
  loading.value = true
  const res = await getInviteRecord({ page: 1, size: 30 })
  loading.value = false
  const data = res.data as { records?: Array<Record<string, unknown>> } | Array<Record<string, unknown>>
  list.value = Array.isArray(data) ? data : data?.records || []
})
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.item {
  padding: 20rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.sub {
  margin-top: 8rpx;
  color: #888;
  font-size: 24rpx;
}
</style>
