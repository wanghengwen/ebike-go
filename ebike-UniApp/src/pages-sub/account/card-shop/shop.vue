<template>
  <view class="page">
    <scroll-view scroll-y class="list" :style="{ bottom: footerH }">
      <view v-if="loading" class="no-data">{{ t('common.loading') }}</view>
      <template v-else-if="list.length">
        <view v-for="(item, i) in list" :key="i" class="cardItem">
          <FoldCard
            :card-cover="String(item.backOfCardUrl || '')"
            :card-name="String(item.ridingCardName || item.name || '')"
            :card-tag="tagArr(item)"
            :card-origin-cost="Number(item.originCost ?? item.origin_cost ?? 0)"
            :card-current-cost="Number(item.curCost ?? item.current_cost ?? 0)"
            :card-promotion-tag="String(item.promotionTag || '')"
            @click="onCardClick(item)"
          />
        </view>
        <view class="expired" @click="goRule">
          <text>{{ t('account.cardRule') }}</text>
          <image v-if="arrowIcon" class="img_right" :src="arrowIcon" mode="aspectFit" />
        </view>
      </template>
      <view v-else class="no-data">{{ t('account.cardShopEmpty') }}</view>
    </scroll-view>

    <view class="button_warpper">
      <view>{{ t('account.buyCardTip') }}</view>
      <view class="go_my_card" @click="goMyCard()">
        <text :style="{ color: brand }">{{ t('account.myCards') }}</text>
        <image v-if="greenArrow" class="img" :src="greenArrow" mode="aspectFit" />
      </view>
    </view>

    <CardDetailSheet
      :visible="showDetail"
      :card="currentCard"
      @close="showDetail = false"
      @rule="goRule"
      @pay="onPay"
    />
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import {
  getRidingCardList,
  ridingConfigGetRule,
  getServiceRidingCard,
} from '@/api/card'
import { createChannelPay } from '@/features/pay/usePay'
import { checkPayCertification } from '@/features/pay/checkPayCertification'
import { useUserStore } from '@/stores/user'
import { getBrandColor } from '@/shared/config'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import FoldCard from '@/widgets/FoldCard.vue'
import CardDetailSheet from '@/widgets/CardDetailSheet.vue'

const { t } = useI18n()
const user = useUserStore()
const loading = ref(false)
const list = ref<Array<Record<string, unknown>>>([])
const maxNum = ref(0)
const showDetail = ref(false)
const currentCard = ref<Record<string, unknown>>({})
const paying = ref(false)
const pendingCardQuery = ref('')
const footerH = '120rpx'

const brand = computed(() => getBrandColor())
const arrowIcon = computed(() => getMapCfg('iconRight'))
const greenArrow = computed(() => getIconCfg('greenRightArrow') || getMapCfg('iconRight'))

function tagArr(item: Record<string, unknown>) {
  const raw = String(item.descriptionTag || item.description_tag || '')
  return raw
    .split('|')
    .map((s) => s.trim())
    .filter(Boolean)
}

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

onShow(() => setNavTitle(t('account.cardShop')))

onLoad(async (q) => {
  if (!requireLogin()) return
  if (q?.card) pendingCardQuery.value = String(q.card)
  await loadList()
  if (pendingCardQuery.value) {
    try {
      const card = JSON.parse(decodeURIComponent(pendingCardQuery.value)) as Record<string, unknown>
      pendingCardQuery.value = ''
      await onCardClick(card)
    } catch (e) {
      logger.warn('parse card query fail', e)
    }
  }
})

async function loadList() {
  loading.value = true
  const sid = storage.get<string>('serviceId', '') || ''
  try {
    const [cards, rule] = await Promise.all([
      getRidingCardList(sid ? { serviceId: sid } : {}),
      ridingConfigGetRule(sid ? { serviceId: sid } : {}),
    ])
    const data = cards.data as { records?: Array<Record<string, unknown>> } | Array<Record<string, unknown>>
    const raw = Array.isArray(data) ? data : data?.records || []
    list.value = raw.filter((item) => Number(item.state) === 1)
    if (rule.success && rule.data) {
      maxNum.value = Number((rule.data as { maxNum?: number }).maxNum || 0)
    }
  } finally {
    loading.value = false
  }
}

async function onCardClick(item: Record<string, unknown>) {
  const ok = await checkPayCertification()
  if (!ok) return
  currentCard.value = {
    ...item,
    name: item.ridingCardName || item.name,
    curCost: item.curCost ?? item.current_cost,
    originCost: item.originCost ?? item.origin_cost,
    descriptionTag: item.descriptionTag || item.description_tag,
    cardId: item.cardId || item.card_id || item.id,
  }
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

async function onPay() {
  const item = currentCard.value
  if (!item || paying.value) return
  if (!(await checkCanBuy())) return
  paying.value = true
  try {
    const id = item.cardId || item.card_id || item.id
    const cost = Number(item.curCost ?? item.current_cost ?? 0)
    const res = await createChannelPay({
      saleType: 'RIDING_CARD',
      totalFee: cost,
      saleInfo: {
        total_fee: cost,
        riding_card_id: id,
      },
    })
    if (res.paid || res.success) {
      showDetail.value = false
      uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
      setTimeout(() => goMyCard({ focus: '0' }), 800)
    } else if (String(res.code) === '24005') {
      uni.showToast({ title: t('account.cardBuyLimit'), icon: 'none' })
    } else {
      uni.showToast({ title: res.msg || t('pay.payFail'), icon: 'none' })
    }
  } finally {
    paying.value = false
  }
}

function goRule() {
  navigate('to', '/pages-sub/account/card-rules/card-rules?type=1')
}

function goMyCard(params?: { focus?: string }) {
  const pages = getCurrentPages() as Array<{ route?: string }>
  const prev = pages[pages.length - 2]
  if (prev?.route?.includes('cards/cards') && !params) {
    navigate('back')
    return
  }
  const q = params?.focus != null ? `?focus=${encodeURIComponent(params.focus)}` : ''
  navigate('to', `/pages-sub/account/cards/cards${q}`)
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #f8f8f8;
  position: relative;
  box-sizing: border-box;
}
.list {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  box-sizing: border-box;
}
.cardItem {
  margin: 32rpx;
}
.expired {
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
.no-data {
  text-align: center;
  color: #999;
  padding: 120rpx 0;
  font-size: 28rpx;
}
.button_warpper {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: 120rpx;
  padding-bottom: env(safe-area-inset-bottom);
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-left: 32rpx;
  padding-right: 32rpx;
  box-sizing: content-box;
  font-size: 28rpx;
  color: #666;
  z-index: 10;
}
.go_my_card {
  display: flex;
  align-items: center;
  font-weight: 600;
}
.img {
  width: 28rpx;
  height: 28rpx;
  margin-left: 8rpx;
}
</style>
