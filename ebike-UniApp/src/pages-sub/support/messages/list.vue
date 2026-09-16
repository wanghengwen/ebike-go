<template>
  <view class="page">
    <scroll-view
      v-if="list.length"
      class="list"
      scroll-y
      @scrolltolower="loadMore"
    >
      <view
        v-for="(item, i) in list"
        :key="String(item.id || item.msgId || i)"
        class="card item"
        @click="goDetail(item, i)"
      >
        <view class="top">
          <view class="icon-wrap">
            <image v-if="typeIcon(item)" class="icon" :src="typeIcon(item)" mode="aspectFit" />
            <view v-if="!item.izRead" class="dot" />
          </view>
          <text class="title">{{ item.topic || item.title || '-' }}</text>
          <text class="clock">{{ item.createdAt || item.createTime || '' }}</text>
        </view>
        <text class="content">{{ item.content || '' }}</text>
        <view class="bottom">
          <text>{{ t('account.msgViewDetail') }}</text>
          <image v-if="arrowIcon" class="arrow" :src="arrowIcon" mode="widthFix" />
        </view>
      </view>
      <view class="footer" v-if="loading">{{ t('common.loading') }}</view>
      <view class="footer" v-else-if="!hasMore">{{ t('common.noMore') }}</view>
    </scroll-view>

    <view v-else-if="loading" class="empty">{{ t('common.loading') }}</view>
    <view v-else class="empty">
      <image v-if="emptyIcon" class="empty-img" :src="emptyIcon" mode="widthFix" />
      <text>{{ t('account.messagesEmpty') }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getMsgList } from '@/api/message'
import { getLoginPath, navigate, setNavTitle } from '@/shared/navigate'
import { ensureLoggedIn } from '@/shared/ensureLoggedIn'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'

const { t } = useI18n()
const loading = ref(false)
const list = ref<Array<Record<string, unknown>>>([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const emptyIcon = computed(() => getIconCfg('no_message') || getIconCfg('notLogin'))
const arrowIcon = computed(() => getMapCfg('iconRight'))
const hasMore = computed(() => !total.value || list.value.length < total.value)

onShow(() => setNavTitle(t('account.messages')))

onLoad(async () => {
  if (!(await ensureLoggedIn())) {
    uni.showModal({
      title: t('auth.loginTitle'),
      content: t('account.needLogin'),
      showCancel: false,
      confirmText: t('auth.loginNow'),
      success: (res) => {
        if (res.confirm) navigate('redirect', getLoginPath())
      },
    })
    return
  }
  void reload()
})

function typeIcon(item: Record<string, unknown>) {
  const msgType = Number(item.msgType)
  if (msgType === 1) return getIconCfg('systemInform')
  return getIconCfg('activityInforms') || getIconCfg('systemInform')
}

async function fetchPage(reset: boolean) {
  if (loading.value) return
  if (!reset && total.value && list.value.length >= total.value) {
    uni.showToast({ title: t('common.noMore'), icon: 'none' })
    return
  }
  loading.value = true
  try {
    const res = await getMsgList({
      msgTypes: [1],
      pageNum: pageNum.value,
      pageSize,
    })
    if (!res.success) {
      uni.showToast({ title: t('account.messagesLoadFail'), icon: 'none' })
      return
    }
    const data = res.data as {
      count?: number
      total?: number
      list?: Array<Record<string, unknown>>
      records?: Array<Record<string, unknown>>
    }
    total.value = Number(data.count ?? data.total ?? 0)
    const rows = data.list || data.records || []
    list.value = reset ? rows : list.value.concat(rows)
  } catch {
    uni.showToast({ title: t('account.messagesLoadFail'), icon: 'none' })
  } finally {
    loading.value = false
  }
}

async function reload() {
  pageNum.value = 1
  list.value = []
  total.value = 0
  await fetchPage(true)
}

function loadMore() {
  if (loading.value || !hasMore.value) return
  pageNum.value += 1
  void fetchPage(false)
}

function goDetail(item: Record<string, unknown>, index: number) {
  const id = item.id || item.msgId
  if (!id) return
  list.value[index] = { ...item, izRead: 1 }
  navigate('to', `/pages-sub/support/messages/detail?id=${encodeURIComponent(String(id))}`)
}
</script>

<style scoped lang="scss">
.list {
  height: 100vh;
  box-sizing: border-box;
}
.item {
  margin-bottom: 8rpx;
}
.top {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
}
.icon-wrap {
  position: relative;
  width: 48rpx;
  height: 48rpx;
  flex-shrink: 0;
}
.icon {
  width: 48rpx;
  height: 48rpx;
}
.dot {
  position: absolute;
  top: -2rpx;
  right: -2rpx;
  width: 16rpx;
  height: 16rpx;
  background: #ff5936;
  border-radius: 50%;
}
.title {
  flex: 1;
  font-size: 32rpx;
  font-weight: 500;
  color: #333;
  min-width: 0;
}
.clock {
  font-size: 24rpx;
  color: #999;
  flex-shrink: 0;
}
.content {
  margin: 20rpx 0;
  font-size: 28rpx;
  color: #666;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}
.bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 20rpx;
  border-top: 1px solid #f0f0f0;
  font-size: 28rpx;
  color: #333;
}
.arrow {
  width: 32rpx;
}
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 70vh;
  color: #999;
  gap: 24rpx;
}
.empty-img {
  width: 280rpx;
}
.footer {
  text-align: center;
  color: #999;
  font-size: 24rpx;
  padding: 24rpx 0 48rpx;
}
</style>
