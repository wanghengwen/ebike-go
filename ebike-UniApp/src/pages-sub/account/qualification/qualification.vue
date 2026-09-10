<template>
  <view class="page">
    <view v-if="loading" class="card">{{ t('common.loading') }}</view>
    <template v-else>
      <view class="card" v-if="izWxScorePayDeposited">
        <view class="title">{{ t('account.payScore') }}</view>
        <view class="row">
          <view class="desc">
            <view class="main">{{ t('account.qualPayScoreTitle') }}</view>
            <view class="sub">{{ t('account.qualPayScoreSub') }}</view>
          </view>
          <view class="btn-mini" @click="goPayScore">{{ t('common.enable') }}</view>
        </view>
      </view>

      <view class="card" v-if="izCareer">
        <view class="title">{{ t('account.careerAuth') }}</view>
        <view class="row">
          <view class="desc">
            <view class="main">{{ t('account.qualCareerTitle') }}</view>
            <view class="sub">{{ t('account.qualCareerSub') }}</view>
          </view>
          <view class="btn-mini orange" @click="goCareer">{{ t('account.careerAuth') }}</view>
        </view>
      </view>

      <view class="card" v-if="izDepositCard && depositCards.length">
        <view class="title">{{ t('account.qualDepositCard') }}</view>
        <view
          v-for="(item, i) in depositCards"
          :key="i"
          class="deposit-item"
          @click="buyDepositCard(item)"
        >
          <view class="main">{{ item.name || '-' }}</view>
          <view class="price-row">
            <text class="price">¥{{ fenToYuan(item.discountMoney) }}</text>
            <text class="old" v-if="item.curMoney">¥{{ fenToYuan(item.curMoney) }}</text>
          </view>
          <view class="sub">
            {{ t('account.qualDepositCardDays', { days: item.cardDurationDays || '-' }) }}
          </view>
        </view>
      </view>

      <view class="card" v-if="izDeposit && deposit > 0">
        <view class="title">{{ t('account.qualPayDeposit') }}</view>
        <view class="deposit-item" @click="payDeposit">
          <view class="main">{{ t('account.qualDepositYear') }}</view>
          <view class="price-row">
            <text class="price">¥{{ fenToYuan(deposit) }}</text>
          </view>
          <view class="sub">{{ t('account.qualDepositRefundable') }}</view>
        </view>
      </view>

      <view class="card" v-if="!hasAnyOption">
        <view>{{ t('common.empty') }}</view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getQualificationList } from '@/api/user'
import { getWechatPayScoreRecord } from '@/api/wechatScore'
import { uniLoginCode } from '@/features/auth/useAuth'
import { usePay, fenToYuan } from '@/features/pay/usePay'
import { useUserStore } from '@/stores/user'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const { payMoney } = usePay()
const user = useUserStore()

const loading = ref(false)
const izDepositCard = ref(false)
const izCareer = ref(false)
const izDeposit = ref(false)
const izWxScorePayDeposited = ref(false)
const deposit = ref(0)
const depositCards = ref<Array<Record<string, unknown>>>([])
const careerTags = ref<Array<Record<string, unknown>>>([])

const hasAnyOption = computed(
  () =>
    izWxScorePayDeposited.value ||
    izCareer.value ||
    (izDepositCard.value && depositCards.value.length > 0) ||
    (izDeposit.value && deposit.value > 0),
)

onShow(() => {
  setNavTitle(t('account.qualification'))
  void maybeRedirectAuthorizedPayScore()
})

async function maybeRedirectAuthorizedPayScore() {
  try {
    const code = await uniLoginCode()
    if (!code) return
    const res = await getWechatPayScoreRecord({ code })
    if (!res.success || !res.data) return
    const data = res.data as Record<string, unknown> | Array<Record<string, unknown>>
    const row = Array.isArray(data) ? data[0] || {} : data
    const state = String(row.authorization_state || row.status || row.permissionState || '')
    if (state === 'AVAILABLE' || state === 'AUTHORIZED' || state === '1') {
      navigate('redirect', '/pages-sub/account/pay-score/pay-score')
    }
  } catch (e) {
    logger.warn('pay score auth check soft fail', e)
  }
}

async function load() {
  loading.value = true
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getQualificationList({ serviceId: sid })
    if (!res.success || !res.data) {
      depositCards.value = []
      return
    }
    const data = res.data as {
      depositCardCoList?: Array<Record<string, unknown>>
      izDepositCard?: boolean
      izCareer?: boolean
      izDeposit?: boolean
      deposit?: number
      izWxScorePayDeposited?: boolean
      careerTags?: Array<Record<string, unknown>>
    }
    depositCards.value = data.depositCardCoList || []
    izDepositCard.value = Boolean(data.izDepositCard)
    izCareer.value = Boolean(data.izCareer)
    izDeposit.value = Boolean(data.izDeposit)
    deposit.value = Number(data.deposit || 0)
    izWxScorePayDeposited.value = Boolean(data.izWxScorePayDeposited)
    careerTags.value = data.careerTags || []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  user.hydrateFromStorage()
  void load()
})

function goPayScore() {
  navigate('to', '/pages-sub/account/pay-score/pay-score')
}

function goCareer() {
  const info = user.userInfo as Record<string, unknown>
  if (Number(info.careerState) === 0) {
    navigate('to', '/pages-sub/account/career/career')
    return
  }
  if (!careerTags.value.length) {
    navigate('to', '/pages-sub/account/career/career')
    return
  }
  const first = careerTags.value[0]
  const tagName = encodeURIComponent(String(first.tagName || first.name || ''))
  navigate(
    'to',
    `/pages-sub/account/career/career?tagId=${encodeURIComponent(String(first.id || ''))}&tagName=${tagName}`,
  )
}

async function payDeposit() {
  const res = await payMoney({ sale_type: 'DEPOSIT' })
  if (res.success && res.paid) {
    uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
    setTimeout(() => navigate('back'), 500)
  } else if (!res.success) {
    uni.showToast({ title: res.msg || t('pay.payFail'), icon: 'none' })
  }
}

async function buyDepositCard(item: Record<string, unknown>) {
  const res = await payMoney({
    sale_type: 'DEPOSIT_CARD',
    sale_info: {
      total_fee: Number(item.discountMoney),
      deposit_card_id: item.id,
    },
  })
  if (res.success && res.paid) {
    uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
    setTimeout(() => navigate('back'), 500)
  } else if (!res.success) {
    uni.showToast({ title: res.msg || t('pay.payFail'), icon: 'none' })
  }
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}
.desc {
  flex: 1;
}
.main {
  font-weight: 600;
}
.sub {
  margin-top: 6rpx;
  color: #888;
  font-size: 24rpx;
}
.btn-mini {
  padding: 12rpx 24rpx;
  border-radius: 999rpx;
  border: 1px solid #07c160;
  color: #07c160;
  font-size: 24rpx;
  white-space: nowrap;
}
.btn-mini.orange {
  border-color: #ff922b;
  color: #ff922b;
}
.deposit-item {
  padding: 20rpx 0;
  border-top: 1px solid #f0f0f0;
}
.price-row {
  margin-top: 8rpx;
  display: flex;
  align-items: baseline;
  gap: 12rpx;
}
.price {
  font-size: 40rpx;
  font-weight: 700;
  color: #3aa0e8;
}
.old {
  color: #999;
  text-decoration: line-through;
  font-size: 24rpx;
}
</style>
