<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('account.repairList') }}</view>
      <view v-if="loading">{{ t('common.loading') }}</view>
      <view v-else-if="!list.length" class="empty">
        <image v-if="emptyIcon" class="empty-img" :src="emptyIcon" mode="widthFix" />
        <text>{{ t('common.empty') }}</text>
      </view>
      <view v-for="(item, i) in list" :key="i" class="item" @click="goProgress(item)">
        <image v-if="bikeIcon(item.bikeType)" class="bike" :src="bikeIcon(item.bikeType)" mode="aspectFit" />
        <view class="body">
          <view class="main">NO.{{ item.carId || '-' }}</view>
          <view class="sub">
            <image v-if="locIcon" class="loc" :src="locIcon" mode="aspectFit" />
            {{ stateText(item) }} · {{ item.address || item.createTime || item.reportTime || '' }}
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { repairList } from '@/api/repair'
import { repairBikeIcon } from '@/features/support/repairIcons'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const loading = ref(false)
const list = ref<Array<Record<string, unknown>>>([])
const emptyIcon = computed(() => getIconCfg('notLogin'))
const locIcon = computed(() => getIconCfg('locationOutline'))

onShow(() => setNavTitle(t('account.repairList')))

onMounted(async () => {
  loading.value = true
  const res = await repairList({ pageNum: 1, pageSize: 20 })
  loading.value = false
  const data = res.data as {
    records?: Array<Record<string, unknown>>
    list?: Array<Record<string, unknown>>
  } | Array<Record<string, unknown>>
  list.value = Array.isArray(data) ? data : data?.list || data?.records || []
})

function bikeIcon(bikeType: unknown) {
  return repairBikeIcon(bikeType)
}

function stateText(item: Record<string, unknown>) {
  return Number(item.state) === 2 ? t('account.repairProcessed') : t('account.repairProcessing')
}

function goProgress(item: Record<string, unknown>) {
  navigate(
    'to',
    `/pages-sub/support/repair/repair-progress?json=${encodeURIComponent(JSON.stringify(item))}`,
  )
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.empty {
  text-align: center;
  color: #999;
  padding: 40rpx 0;
}
.empty-img {
  width: 240rpx;
  margin: 0 auto 16rpx;
  display: block;
}
.item {
  display: flex;
  gap: 16rpx;
  padding: 20rpx 0;
  border-bottom: 1px solid #f0f0f0;
  align-items: center;
}
.bike {
  width: 64rpx;
  height: 64rpx;
  flex-shrink: 0;
}
.body {
  flex: 1;
  min-width: 0;
}
.sub {
  color: #888;
  margin-top: 8rpx;
  font-size: 24rpx;
  display: flex;
  align-items: center;
  gap: 6rpx;
}
.loc {
  width: 24rpx;
  height: 24rpx;
}
</style>
