<template>
  <view class="page">
    <scroll-view scroll-y class="list" :style="{ bottom: '160rpx' }">
      <view v-if="loading" class="hint">{{ t('common.loading') }}</view>
      <template v-else-if="usedList.length">
        <view v-for="(item, index) in usedList" :key="index" class="card">
          <view class="card_content" @click="unFold(item, index)">
            <image
              class="riding_card_img"
              :src="String(item.backOfCardUrl || coverFallback)"
              mode="widthFix"
            />
            <view class="card_des">
              <view class="card_title">{{ item.name || item.cardName || '-' }}</view>
              <view v-if="tagList(item).length" class="card_tips">
                <block v-for="(tag, ti) in tagList(item)" :key="ti">
                  <text>{{ tag }}</text>
                  <view v-if="ti + 1 !== tagList(item).length" class="line" />
                </block>
              </view>
              <view class="rest_times">
                {{ t('account.cardRemainToday', { n: item.remainTimes ?? 0 }) }}
              </view>
              <view class="rest_times rest_days">
                {{ t('account.cardRemainTime') }}{{ remainText(item) }}
                <image
                  v-if="expanded !== index && bottomArrow"
                  class="arrow_img"
                  :src="bottomArrow"
                  mode="aspectFit"
                />
              </view>
            </view>
          </view>
          <view v-if="expanded === index" class="card_remind">
            <rich-text v-if="currentHtml" :nodes="currentHtml" />
            <view class="pack_up">
              <text>{{ t('account.cardDispatchHint') }}</text>
              <view class="pack_up_warpper" @click.stop="expanded = -1">
                <text class="pack_up_text">{{ t('account.collapse') }}</text>
                <image v-if="topArrow" class="arrow_img" :src="topArrow" mode="widthFix" />
              </view>
            </view>
          </view>
        </view>
        <view class="expired-link" @click="goExpired">
          <text>{{ t('account.viewExpiredCards') }}</text>
          <image v-if="arrowIcon" class="img_right" :src="arrowIcon" mode="aspectFit" />
        </view>
      </template>
      <view v-else class="noCard">
        <image v-if="noCardIcon" class="img_big" :src="noCardIcon" mode="aspectFit" />
        <text class="noCard_text">{{ t('account.cardsActiveEmpty') }}</text>
        <view class="buy-btn" :style="btnStyle" @click="goShop">{{ t('account.goBuyCard') }}</view>
        <view class="expired-link" @click="goExpired">
          <text>{{ t('account.viewExpiredCards') }}</text>
          <image v-if="arrowIcon" class="img_right" :src="arrowIcon" mode="aspectFit" />
        </view>
      </view>
    </scroll-view>

    <view class="button_warpper">
      <view class="shop-btn" :style="btnStyle" @click="goShop">{{ t('account.cardShop') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getUserRidingCard } from '@/api/card'
import { useUserStore } from '@/stores/user'
import { getBrandColor } from '@/shared/config'
import { formatRichText } from '@/shared/format'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const user = useUserStore()
const loading = ref(false)
const usedList = ref<Array<Record<string, unknown>>>([])
const expanded = ref(-1)
const currentHtml = ref('')

const noCardIcon = computed(() => getIconCfg('noCard'))
const coverFallback = computed(() => getIconCfg('walletCyclingCard'))
const arrowIcon = computed(() => getMapCfg('iconRight'))
const bottomArrow = computed(() => getIconCfg('bottomArrow') || getMapCfg('iconRight'))
const topArrow = computed(() => getIconCfg('topArrow') || getMapCfg('iconRight'))
const btnStyle = computed(() => {
  const bg = getBrandColor()
  return {
    backgroundColor: bg,
    borderColor: bg,
    color: '#1E4A38',
    fontWeight: 'bold',
  }
})

function requireLogin(): boolean {
  user.hydrateFromStorage()
  if (user.isLoggedIn) return true
  uni.showModal({
    title: t('auth.loginTitle'),
    showCancel: false,
    confirmText: t('auth.loginNow'),
    success: (r) => {
      if (r.confirm) navigate('redirect', '/pages/auth/quick-login')
    },
  })
  return false
}

onShow(() => {
  setNavTitle(t('account.myCards'))
  if (user.isLoggedIn) void load()
})

onLoad(() => {
  if (!requireLogin()) return
  void load()
})

function tagList(item: Record<string, unknown>): string[] {
  return String(item.description_tag || item.descriptionTag || '')
    .split('|')
    .map((s) => s.trim())
    .filter(Boolean)
}

function remainText(item: Record<string, unknown>): string {
  const raw = String(item.cardExpiredDate || item.expireTime || '')
  if (!raw) return '-'
  const end = new Date(raw.replace(/-/g, '/')).getTime()
  if (!Number.isFinite(end)) return raw
  const diff = end - Date.now()
  if (diff <= 0) return '0'
  const days = Math.floor(diff / (24 * 60 * 60 * 1000))
  if (days >= 1) return `${days}${t('ride.days')}`
  const h = Math.floor(diff / (60 * 60 * 1000))
  const m = Math.floor((diff % (60 * 60 * 1000)) / (60 * 1000))
  const s = Math.floor((diff % (60 * 1000)) / 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(h)}${t('ride.hours')}${pad(m)}${t('ride.minutes')}${pad(s)}${t('ride.seconds')}`
}

function decodeDetail(raw: string): string {
  if (!raw) return ''
  try {
    if (typeof atob === 'function' && /^[A-Za-z0-9+/=]+$/.test(raw) && raw.length % 4 === 0) {
      const text = decodeURIComponent(
        Array.prototype.map
          .call(atob(raw), (c: string) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join(''),
      )
      return formatRichText(text)
    }
  } catch {
    /* fallthrough */
  }
  return formatRichText(raw)
}

function unFold(item: Record<string, unknown>, index: number) {
  if (expanded.value === index) {
    expanded.value = -1
    return
  }
  currentHtml.value = decodeDetail(String(item.detail_info || item.detailInfo || ''))
  expanded.value = index
}

async function load() {
  loading.value = true
  expanded.value = -1
  try {
    const res = await getUserRidingCard()
    const data = res.data as { used?: unknown } | unknown[] | undefined
    if (Array.isArray(data)) {
      usedList.value = data as Array<Record<string, unknown>>
    } else if (data && typeof data === 'object') {
      usedList.value = Array.isArray(data.used) ? (data.used as Array<Record<string, unknown>>) : []
    } else {
      usedList.value = []
    }
  } finally {
    loading.value = false
  }
}

function goShop() {
  navigate('to', '/pages-sub/account/card-shop/shop')
}
function goExpired() {
  navigate('to', '/pages-sub/account/cards/cards-expired')
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #f8f8f8;
  position: relative;
}
.list {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  box-sizing: border-box;
}
.hint {
  text-align: center;
  color: #999;
  padding: 80rpx 0;
}
.card {
  position: relative;
  display: flex;
  flex-direction: column;
  margin: 32rpx;
  padding: 16rpx 32rpx 16rpx 16rpx;
  background: #ffffff;
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
.card_des {
  flex: 1;
  margin-left: 32rpx;
  margin-top: 8rpx;
  min-width: 0;
}
.card_title {
  font-size: 36rpx;
  font-weight: 600;
  color: #333;
}
.card_tips {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  margin-top: 8rpx;
  font-size: 26rpx;
  color: #666;
}
.line {
  width: 2rpx;
  height: 24rpx;
  background: #666;
  margin: 0 12rpx;
}
.rest_times {
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #666;
}
.rest_days {
  display: flex;
  align-items: center;
}
.arrow_img {
  width: 24rpx;
  height: 24rpx;
  margin-left: 8rpx;
}
.card_remind {
  margin-top: 16rpx;
  padding-top: 16rpx;
  border-top: 2rpx solid #f6f6f6;
  color: #666;
  font-size: 24rpx;
  line-height: 1.5;
}
.pack_up {
  margin-top: 16rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #999;
  font-size: 22rpx;
}
.pack_up_warpper {
  display: flex;
  align-items: center;
  color: #666;
}
.pack_up_text {
  margin-right: 4rpx;
}
.expired-link {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24rpx 0 48rpx;
  font-size: 24rpx;
  color: #999;
}
.img_right {
  width: 24rpx;
  height: 24rpx;
  margin-left: 8rpx;
}
.noCard {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120rpx;
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
.buy-btn {
  margin-top: 48rpx;
  width: 360rpx;
  height: 96rpx;
  border-radius: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
}
.button_warpper {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 24rpx 48rpx calc(24rpx + env(safe-area-inset-bottom));
  background: #fff;
  z-index: 10;
}
.shop-btn {
  height: 96rpx;
  border-radius: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
}
</style>
