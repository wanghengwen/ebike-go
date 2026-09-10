<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('account.cardShop') }}</view>
      <view v-if="loading">{{ t('common.loading') }}</view>
      <view v-else-if="!list.length">{{ t('common.empty') }}</view>
      <view
        v-for="(item, i) in list"
        :key="i"
        class="item"
        :class="{ on: selected === i }"
        @click="onSelect(i)"
      >
        <view>
          <view class="name">{{ item.name || item.cardName || '-' }}</view>
          <view class="sub">{{ tagText(item) }}</view>
        </view>
        <view class="price">¥{{ displayPrice(item) }}</view>
      </view>
      <view class="btn-primary" v-if="list.length" @click="openDetail">{{ t('pay.payNow') }}</view>
      <view class="link" @click="openRule">{{ t('account.cardRule') }}</view>
    </view>

    <view v-if="showDetail" class="mask" @click.self="showDetail = false">
      <view class="sheet">
        <view class="sheet__head">
          <view>
            <view class="sheet__name">{{ current.name || current.cardName || '-' }}</view>
            <view class="sheet__price">
              ¥{{ displayPrice(current) }}
              <text
                v-if="originPrice(current) > Number(current.current_cost ?? current.price ?? current.amount ?? 0)"
                class="sheet__origin"
              >
                ¥{{ fenToYuan(originPrice(current)) }}
              </text>
            </view>
          </view>
          <text class="sheet__close" @click="showDetail = false">×</text>
        </view>
        <view class="sheet__section">
          <view class="sheet__row">
            <text class="sheet__label">{{ t('account.cardUsage') }}</text>
            <text class="sheet__link" @click="openRule">{{ t('account.cardRule') }} ›</text>
          </view>
          <rich-text v-if="detailHtml" class="sheet__html" :nodes="detailHtml" />
          <view v-else class="sheet__empty">{{ t('common.empty') }}</view>
        </view>
        <view class="sheet__section">
          <view class="sheet__label">{{ t('pay.walletPwd') }}</view>
          <input
            class="sheet__input"
            password
            maxlength="6"
            v-model="walletPwd"
            :placeholder="t('pay.walletPwd6')"
          />
        </view>
        <view class="sheet__actions">
          <view class="btn-ghost" @click="showDetail = false">{{ t('common.cancel') }}</view>
          <view class="btn-primary" @click="onBuy">
            {{ t('pay.payNow') }} ¥{{ displayPrice(current) }}
            <text v-if="saveFen > 0" class="sheet__save">
              {{ t('pay.savedAmount', { m: fenToYuan(saveFen) }) }}
            </text>
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
import {
  getRidingCardList,
  ridingConfigGetRule,
  getServiceRidingCard,
} from '@/api/card'
import { createChannelPay, fenToYuan } from '@/features/pay/usePay'
import { checkPayCertification } from '@/features/pay/checkPayCertification'
import { useUserStore } from '@/stores/user'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const user = useUserStore()
const loading = ref(false)
const list = ref<Array<Record<string, unknown>>>([])
const selected = ref(0)
const ruleUrl = ref('')
const maxNum = ref(0)
const showDetail = ref(false)
const walletPwd = ref('')
const paying = ref(false)

const current = computed(() => list.value[selected.value] || {})
const detailHtml = computed(() =>
  String(current.value.detailInfo || current.value.detail_info || ''),
)
const saveFen = computed(() => {
  const cur = Number(current.value.current_cost ?? current.value.price ?? current.value.amount ?? 0)
  const origin = originPrice(current.value)
  return origin > cur ? origin - cur : 0
})

function requireLogin(): boolean {
  user.hydrateFromStorage()
  if (user.isLoggedIn) return true
  uni.showToast({ title: t('account.needLogin'), icon: 'none' })
  setTimeout(() => navigate('to', '/pages/auth/quick-login'), 400)
  return false
}

onShow(() => {
  setNavTitle(t('account.cardShop'))
  if (!requireLogin()) return
})

function displayPrice(item: Record<string, unknown>) {
  const p = Number(item.current_cost ?? item.curCost ?? item.price ?? item.amount ?? 0)
  return fenToYuan(p)
}

function originPrice(item: Record<string, unknown>) {
  return Number(item.origin_cost ?? item.originCost ?? item.originalPrice ?? 0)
}

function tagText(item: Record<string, unknown>) {
  const raw = String(item.description_tag || item.descriptionTag || item.desc || '')
  return raw
    .split('|')
    .map((s) => s.trim())
    .filter(Boolean)
    .join(' · ')
}

function onSelect(i: number) {
  selected.value = i
}

function openDetail() {
  if (!list.value[selected.value]) return
  walletPwd.value = ''
  showDetail.value = true
}

async function checkCanBuy(): Promise<boolean> {
  if (!maxNum.value) return true
  try {
    const res = await getServiceRidingCard()
    const data = res.data as { used?: unknown[] } | undefined
    const used = Array.isArray(data?.used) ? data!.used! : []
    if (used.length >= maxNum.value) {
      uni.showToast({ title: t('account.cardBuyLimit'), icon: 'none' })
      return false
    }
  } catch (e) {
    logger.warn('checkCanBuy soft fail', e)
  }
  return true
}

async function onBuy() {
  const item = list.value[selected.value]
  if (!item) return
  if (!walletPwd.value) {
    uni.showToast({ title: t('pay.needWalletPwd'), icon: 'none' })
    return
  }
  if (walletPwd.value.length !== 6) {
    uni.showToast({ title: t('pay.walletPwd6'), icon: 'none' })
    return
  }
  if (!(await checkCanBuy())) return
  const ok = await checkPayCertification()
  if (!ok) return
  if (paying.value) return
  paying.value = true
  try {
    const id = item.card_id || item.cardId || item.id
    const cost = Number(item.current_cost ?? item.curCost ?? item.price ?? item.amount ?? 0)
    const res = await createChannelPay({
      saleType: 'RIDING_CARD',
      totalFee: cost,
      channelType: 'YUDAOXING_APP',
      invokeWx: false,
      walletPwd: walletPwd.value,
      saleInfo: {
        total_fee: cost,
        riding_card_id: id,
      },
    })
    if (res.paid || res.success) {
      showDetail.value = false
      uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
      setTimeout(() => navigate('to', '/pages-sub/account/cards/cards'), 500)
    } else if (String(res.code) === '24005') {
      uni.showToast({ title: t('account.cardBuyLimit'), icon: 'none' })
    } else {
      uni.showToast({ title: res.msg || t('pay.payFail'), icon: 'none' })
    }
  } finally {
    paying.value = false
  }
}

function openRule() {
  if (!ruleUrl.value) {
    uni.showToast({ title: t('common.empty'), icon: 'none' })
    return
  }
  navigate('to', `/pages/webview/webview?url=${encodeURIComponent(ruleUrl.value)}`)
}

onMounted(async () => {
  if (!requireLogin()) return
  loading.value = true
  const sid = storage.get<string>('serviceId', '') || ''
  try {
    const [cards, rule] = await Promise.all([
      getRidingCardList(sid ? { serviceId: sid } : {}),
      ridingConfigGetRule(sid ? { serviceId: sid } : {}),
    ])
    const data = cards.data as { records?: Array<Record<string, unknown>> } | Array<Record<string, unknown>>
    const raw = Array.isArray(data) ? data : data?.records || []
    // Legacy cardCenter: only sell cards with state === 1
    list.value = raw.filter((item) => Number(item.state) === 1)
    selected.value = 0
    if (rule.success && rule.data) {
      const r = rule.data as { url?: string; maxNum?: number }
      ruleUrl.value = String(r.url || '')
      maxNum.value = Number(r.maxNum || 0)
    }
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.item {
  padding: 24rpx 16rpx;
  border: 1px solid #f0f0f0;
  border-radius: 12rpx;
  margin-bottom: 16rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.item.on {
  border-color: #3aa0e8;
}
.name {
  font-weight: 600;
}
.sub {
  color: #888;
  font-size: 24rpx;
  margin-top: 8rpx;
  max-width: 420rpx;
}
.price {
  color: #3aa0e8;
  font-weight: 600;
}
.link {
  margin-top: 28rpx;
  color: #3aa0e8;
  text-align: center;
}
.mask {
  position: fixed;
  inset: 0;
  z-index: 1200;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: flex-end;
}
.sheet {
  width: 100%;
  background: #fff;
  border-radius: 32rpx 32rpx 0 0;
  padding: 32rpx 32rpx calc(32rpx + env(safe-area-inset-bottom));
  max-height: 85vh;
  overflow-y: auto;
}
.sheet__head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24rpx;
}
.sheet__name {
  font-size: 34rpx;
  font-weight: 700;
}
.sheet__price {
  margin-top: 8rpx;
  color: #ff5936;
  font-size: 36rpx;
  font-weight: 700;
}
.sheet__origin {
  margin-left: 12rpx;
  color: #999;
  font-size: 24rpx;
  font-weight: 400;
  text-decoration: line-through;
}
.sheet__close {
  font-size: 44rpx;
  color: #999;
  line-height: 1;
  padding: 0 8rpx;
}
.sheet__section {
  margin-bottom: 28rpx;
}
.sheet__row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12rpx;
}
.sheet__label {
  font-weight: 600;
  color: #333;
}
.sheet__link {
  color: #3aa0e8;
  font-size: 24rpx;
}
.sheet__html {
  color: #666;
  font-size: 24rpx;
  line-height: 1.5;
}
.sheet__empty {
  color: #999;
  font-size: 24rpx;
}
.sheet__input {
  margin-top: 12rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 24rpx;
}
.sheet__actions {
  display: flex;
  gap: 16rpx;
}
.sheet__actions > view {
  flex: 1;
}
.sheet__save {
  display: block;
  font-size: 20rpx;
  font-weight: 400;
  opacity: 0.9;
}
</style>
