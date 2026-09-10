<template>
  <view class="page">
    <view class="tabs">
      <view class="tab" :class="{ on: tab === 'used' }" @click="switchTab('used')">
        {{ t('account.cardsActive') }}
      </view>
      <view class="tab" :class="{ on: tab === 'expired' }" @click="switchTab('expired')">
        {{ t('account.cardsExpired') }}
      </view>
    </view>

    <view class="card">
      <view v-if="loading">{{ t('common.loading') }}</view>
      <template v-else-if="!list.length">
        <view class="empty">
          {{ tab === 'expired' ? t('account.cardsExpiredEmpty') : t('account.cardsActiveEmpty') }}
        </view>
        <view v-if="tab === 'used'" class="btn-primary" @click="goShop">{{ t('account.goBuyCard') }}</view>
      </template>
      <view
        v-for="(item, i) in list"
        :key="i"
        class="item"
        :class="{ expired: tab === 'expired' }"
        @click="toggle(i)"
      >
        <image
          v-if="item.backOfCardUrl"
          class="item__bg"
          :src="String(item.backOfCardUrl)"
          mode="aspectFill"
        />
        <view class="item__body">
          <view class="item__name">{{ item.name || item.cardName || '-' }}</view>
          <view class="item__meta" v-if="item.remainTimes != null && tab === 'used'">
            {{ t('account.cardRemain', { n: item.remainTimes }) }}
          </view>
          <view class="item__meta">
            {{
              tab === 'expired'
                ? t('account.cardExpiredAt', { d: item.cardExpiredDate || item.expireTime || '-' })
                : t('account.cardExpireAt', { d: item.cardExpiredDate || item.expireTime || '-' })
            }}
          </view>
          <view class="item__tags" v-if="tagList(item).length">
            <text v-for="(tag, ti) in tagList(item)" :key="ti" class="tag">{{ tag }}</text>
          </view>
          <view class="item__detail" v-if="expanded === i && detailText(item)">
            {{ detailText(item) }}
          </view>
        </view>
        <image v-if="tab === 'expired' && stampIcon" class="item__stamp" :src="stampIcon" mode="aspectFit" />
      </view>

      <view class="btn-ghost" v-if="tab === 'used'" @click="switchTab('expired')">
        {{ t('account.viewExpiredCards') }}
      </view>
      <view class="btn-ghost" v-else @click="goShop">{{ t('account.cardShop') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getUserRidingCard } from '@/api/card'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const loading = ref(false)
const tab = ref<'used' | 'expired'>('used')
const usedList = ref<Array<Record<string, unknown>>>([])
const expiredList = ref<Array<Record<string, unknown>>>([])
const expanded = ref(-1)
const stampIcon = computed(() => getIconCfg('expired'))

const list = computed(() => (tab.value === 'expired' ? expiredList.value : usedList.value))

onShow(() => setNavTitle(t('account.myCards')))

function normalizeList(raw: unknown): Array<Record<string, unknown>> {
  if (Array.isArray(raw)) return raw as Array<Record<string, unknown>>
  return []
}

function tagList(item: Record<string, unknown>): string[] {
  const raw = String(item.description_tag || item.descriptionTag || '')
  return raw
    .split('|')
    .map((s) => s.trim())
    .filter(Boolean)
}

function detailText(item: Record<string, unknown>): string {
  const raw = String(item.detail_info || item.detailInfo || '')
  if (!raw) return ''
  try {
    // Legacy stores base64 text
    if (typeof atob === 'function' && /^[A-Za-z0-9+/=]+$/.test(raw) && raw.length % 4 === 0) {
      return decodeURIComponent(
        Array.prototype.map
          .call(atob(raw), (c: string) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join(''),
      )
    }
  } catch {
    /* fallthrough */
  }
  return raw
}

async function load() {
  loading.value = true
  try {
    const res = await getUserRidingCard()
    const data = res.data as
      | {
          used?: unknown
          expired?: unknown
          records?: Array<Record<string, unknown>>
        }
      | Array<Record<string, unknown>>
      | undefined
    if (Array.isArray(data)) {
      usedList.value = data
      expiredList.value = []
    } else if (data && typeof data === 'object') {
      usedList.value = normalizeList(data.used)
      expiredList.value = normalizeList(data.expired)
      if (!usedList.value.length && !expiredList.value.length && Array.isArray(data.records)) {
        usedList.value = data.records
      }
    } else {
      usedList.value = []
      expiredList.value = []
    }
  } finally {
    loading.value = false
  }
}

function switchTab(next: 'used' | 'expired') {
  tab.value = next
  expanded.value = -1
}

function toggle(i: number) {
  expanded.value = expanded.value === i ? -1 : i
}

onMounted(() => {
  void load()
})

function goShop() {
  navigate('to', '/pages-sub/account/card-shop/shop')
}
</script>

<style scoped lang="scss">
.tabs {
  display: flex;
  gap: 16rpx;
  padding: 24rpx 32rpx 0;
}
.tab {
  flex: 1;
  text-align: center;
  padding: 16rpx 0;
  border-radius: 12rpx;
  background: #f5f6f8;
  color: #666;
}
.tab.on {
  background: #e8f4fc;
  color: #3aa0e8;
  font-weight: 600;
}
.empty {
  text-align: center;
  color: #999;
  padding: 48rpx 0 24rpx;
}
.item {
  position: relative;
  padding: 24rpx;
  margin-bottom: 20rpx;
  border-radius: 16rpx;
  background: #f8fafc;
  overflow: hidden;
  min-height: 160rpx;
}
.item.expired {
  opacity: 0.85;
}
.item__bg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0.35;
}
.item__body {
  position: relative;
  z-index: 1;
}
.item__name {
  font-size: 30rpx;
  font-weight: 700;
}
.item__meta {
  margin-top: 8rpx;
  color: #666;
  font-size: 24rpx;
}
.item__tags {
  margin-top: 12rpx;
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
}
.tag {
  background: rgba(58, 160, 232, 0.12);
  color: #3aa0e8;
  font-size: 22rpx;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}
.item__detail {
  margin-top: 16rpx;
  color: #555;
  font-size: 24rpx;
  line-height: 1.5;
  white-space: pre-wrap;
}
.item__stamp {
  position: absolute;
  right: 16rpx;
  top: 16rpx;
  width: 120rpx;
  height: 120rpx;
  z-index: 2;
}
.btn-ghost,
.btn-primary {
  margin-top: 28rpx;
}
</style>
