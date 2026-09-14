<template>
  <view class="page">
    <scroll-view scroll-y class="list">
      <view v-if="loading" class="hint">{{ t('common.loading') }}</view>
      <template v-else-if="list.length">
        <view v-for="(item, index) in list" :key="index" class="card">
          <view class="card_content">
            <image
              class="riding_card_img gray"
              :src="String(item.backOfCardUrl || expiredCover || coverFallback)"
              mode="widthFix"
            />
            <view class="card_des">
              <view class="card_title">{{ item.name || item.cardName || '-' }}</view>
              <view class="rest_times">
                {{ t('account.cardExpiredAt', { d: item.cardExpiredDate || item.expireTime || '-' }) }}
              </view>
            </view>
            <view v-if="stampIcon" class="card_expired_icon">
              <image class="stamp" :src="stampIcon" mode="widthFix" />
            </view>
          </view>
        </view>
      </template>
      <view v-else class="noCard">
        <image v-if="noCardIcon" class="img_big" :src="noCardIcon" mode="aspectFit" />
        <text class="noCard_text">{{ t('account.cardsExpiredEmpty') }}</text>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getUserRidingCard } from '@/api/card'
import { getIconCfg } from '@/shared/tenantSkin'
import { setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const loading = ref(false)
const list = ref<Array<Record<string, unknown>>>([])

const noCardIcon = computed(() => getIconCfg('noCard'))
const coverFallback = computed(() => getIconCfg('walletCyclingCard'))
const expiredCover = computed(() => getIconCfg('walletCyclingCardExpired'))
const stampIcon = computed(() => getIconCfg('expired'))

onShow(() => {
  setNavTitle(t('account.cardsExpiredTitle'))
  void load()
})

async function load() {
  loading.value = true
  try {
    const res = await getUserRidingCard()
    const data = res.data as { expired?: unknown } | undefined
    list.value =
      data && typeof data === 'object' && Array.isArray(data.expired)
        ? (data.expired as Array<Record<string, unknown>>)
        : []
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #f8f8f8;
}
.list {
  height: 100%;
  box-sizing: border-box;
}
.hint {
  text-align: center;
  color: #999;
  padding: 80rpx 0;
}
.card {
  margin: 32rpx;
  padding: 16rpx 32rpx 16rpx 16rpx;
  background: #fff;
  border-radius: 32rpx;
}
.card_content {
  display: flex;
  position: relative;
}
.riding_card_img {
  width: 184rpx;
  flex-shrink: 0;
}
.riding_card_img.gray {
  filter: grayscale(1);
  opacity: 0.85;
}
.card_des {
  flex: 1;
  margin-left: 32rpx;
  margin-top: 8rpx;
}
.card_title {
  font-size: 36rpx;
  font-weight: 600;
  color: #333;
}
.rest_times {
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #666;
}
.card_expired_icon {
  position: absolute;
  right: 0;
  top: 0;
}
.stamp {
  width: 200rpx;
}
.noCard {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 160rpx;
}
.img_big {
  width: 300rpx;
  height: 300rpx;
}
.noCard_text {
  margin-top: 24rpx;
  color: #999;
  font-size: 28rpx;
}
</style>
