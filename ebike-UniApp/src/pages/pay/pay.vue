<template>
  <view class="pay page">
    <view class="card">
      <view class="title">{{ t('pay.title') }}</view>
      <view class="amount">¥{{ fenToYuan(payCost) }}</view>
      <view class="balance-box" v-if="showWalletBuckets && !paid">
        <text class="balance">{{ t('pay.rechargeBalance', { m: fenToYuan(rechargeBalance) }) }}</text>
        <text class="balance">{{ t('pay.presentBalance', { m: fenToYuan(presentBalance) }) }}</text>
      </view>
      <view class="hint" v-if="loading">{{ t('common.loading') }}</view>
      <view class="frozen" v-if="frozenTip">{{ frozenTip }}</view>
      <view class="hint tip" v-if="wxScore && !paid">{{ t('pay.wxScorePaying') }}</view>
      <view class="hint tip" v-else-if="wxScore && paid">{{ t('pay.wxScorePaid') }}</view>
      <view class="meta" v-if="detail.ridingTime != null">
        {{ t('ride.duration') }}：{{ durationText }}
      </view>
      <view class="meta" v-if="detail.mile != null">
        {{ t('ride.distance') }}：{{ distanceText }}
      </view>
    </view>

    <view class="card reward" v-if="showRedReward">
      <image v-if="rewardBg" class="reward__bg" :src="rewardBg" mode="aspectFill" />
      <view class="reward__body">
        <text class="reward__done">{{ t('ride.redEnvelopeRewardDone') }}</text>
        <text class="reward__amount">
          {{ t('ride.redEnvelopeRewardAmount', { money: redRewardYuan }) }}
        </text>
        <view class="reward__hint">
          <text>{{ t('ride.redEnvelopeRewardHint') }}</text>
          <text class="reward__link" @click="goRedTips">{{ t('ride.redEnvelopeTips') }} ›</text>
        </view>
      </view>
    </view>

    <view class="card" v-if="!loading && Number(payCost) > 0">
      <view class="section-title">{{ t('pay.costDetail') }}</view>
      <view class="row" v-if="detail.originCost != null">
        <text>{{ t('pay.rideCost') }}</text>
        <text>¥{{ fenToYuan(detail.originCost) }}</text>
      </view>
      <view class="row" v-if="detail.startPrice != null">
        <text>{{ t('pay.startPrice') }}</text>
        <text>¥{{ fenToYuan(detail.startPrice) }}</text>
      </view>
      <view class="row" v-if="detail.timeCost != null || detail.durationCost != null">
        <text>{{ t('pay.durationCost') }}</text>
        <text>¥{{ fenToYuan(detail.timeCost ?? detail.durationCost) }}</text>
      </view>
      <view class="row" v-if="detail.mileCost != null">
        <text>{{ t('pay.mileCost') }}</text>
        <text>¥{{ fenToYuan(detail.mileCost) }}</text>
      </view>
      <view class="row" v-if="detail.dispatchCost != null">
        <text>{{ t('pay.dispatchCost') }}</text>
        <text>¥{{ fenToYuan(detail.dispatchCost) }}</text>
      </view>
      <view class="row" v-if="detail.penalty || detail.helmetPenalty">
        <text>{{ t('pay.otherCost') }}</text>
        <text>¥{{ fenToYuan(Number(detail.penalty || 0) + Number(detail.helmetPenalty || 0)) }}</text>
      </view>
      <view class="row" v-if="showDiscountBlock" @click="discountOpen = !discountOpen">
        <text>{{ t('pay.discount') }} {{ discountOpen ? '▴' : '▾' }}</text>
        <text class="warn">-¥{{ fenToYuan(discountTotal) }}</text>
      </view>
      <template v-if="discountOpen && showDiscountBlock">
        <view class="row sub" v-if="detail.izActityFree || detail.izFreeN">
          <text>{{ t('pay.actFree') }}</text>
          <text class="warn">{{ t('pay.actFreeTag') }}</text>
        </view>
        <view class="row sub" v-if="Number(detail.ridingCardCount) > 0">
          <text>{{ t('pay.ridingCardDeduct') }}</text>
          <text class="warn">-¥{{ fenToYuan(detail.ridingCardCount) }}</text>
        </view>
        <view class="row sub" v-if="Number(detail.discountCount) > 0">
          <text>{{ t('pay.discountDeduct') }}</text>
          <text class="warn">-¥{{ fenToYuan(detail.discountCount) }}</text>
        </view>
        <view class="row sub" v-if="Number(detail.randomCount) > 0">
          <text>{{ t('pay.randomDeduct') }}</text>
          <text class="warn">-¥{{ fenToYuan(detail.randomCount) }}</text>
        </view>
        <view
          class="row sub"
          v-if="
            Number(detail.redPaketCarDeduct) > 0 || Number(detail.redPaketParkingDeduct) > 0
          "
        >
          <text>{{ t('pay.redEnvelopeDeduct') }}</text>
          <text class="warn">
            -¥{{
              fenToYuan(
                Number(detail.redPaketCarDeduct || 0) + Number(detail.redPaketParkingDeduct || 0),
              )
            }}
          </text>
        </view>
        <view
          class="row sub"
          v-if="
            Number(detail.deduction) > 0 &&
            !Number(detail.ridingCardCount) &&
            !Number(detail.discountCount) &&
            !Number(detail.randomCount)
          "
        >
          <text>{{ t('pay.otherDeduct') }}</text>
          <text class="warn">-¥{{ fenToYuan(detail.deduction) }}</text>
        </view>
      </template>
      <view class="row link" @click="goCostDetail">{{ t('pay.costDetail') }} ›</view>
    </view>

    <view class="card" v-if="!paid && !wxScore && waitPayMoney > 0">
      <view class="section-title">{{ t('pay.waitPay') }} ¥{{ fenToYuan(waitPayMoney) }}</view>
      <input
        class="input"
        password
        v-model="walletPwd"
        :placeholder="t('pay.walletPwdPlaceholder')"
        maxlength="6"
      />
    </view>

    <view class="actions">
      <view class="btn-primary" v-if="!paid && !wxScore" @click="onPay">{{ t('pay.payNow') }}</view>
      <view
        class="btn-ghost"
        v-if="showObjectionEntry"
        @click="goObjection"
      >
        {{ t('support.objection') }}
      </view>
      <view class="btn-ghost" v-if="showRepairEntry" @click="goRepair">{{ t('ride.goRepair') }}</view>
      <view class="btn-ghost" v-if="paid && orderId" @click="goTripMap">{{ t('ride.tripTrack') }}</view>
      <view class="btn-ghost" @click="goHome">{{ t('ride.homeTitle') }}</view>
    </view>

    <BizPopup
      :visible="showShortRepair"
      :title="t('ride.shortTripRepairTitle')"
      :confirm-text="t('ride.goRepair')"
      :cancel-text="t('ride.noFault')"
      @confirm="onShortRepairConfirm"
      @cancel="showShortRepair = false"
      @close="showShortRepair = false"
    >
      <text>{{ t('ride.shortTripRepairHint') }}</text>
    </BizPopup>

    <OpsPopup :visible="showPopup" :config="currPopup" @close="onPopupClose" />
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { onHide, onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getWalletInfo, queryFrozen, setFrozenOrder } from '@/api/pay'
import { usePay, fenToYuan } from '@/features/pay/usePay'
import { checkPayCertification } from '@/features/pay/checkPayCertification'
import { getRedEnvelopeCfg, redEnvelopeRewardYuan } from '@/features/bike/redEnvelope'
import { formatMile, formatRidingTime } from '@/shared/format'
import { getConfigBaseItem } from '@/api/user'
import BizPopup from '@/widgets/BizPopup.vue'
import OpsPopup from '@/widgets/OpsPopup.vue'
import { useGuidePopup } from '@/features/guide/useGuidePopup'
import { navigate, setNavTitle } from '@/shared/navigate'
import { useTempDataStore } from '@/stores/tempData'
import { useUserStore } from '@/stores/user'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const { loadLastOrder, loadOrder, settleLastOrder, calcWaitPayMoney } = usePay()
const temp = useTempDataStore()
const user = useUserStore()
const { showPopup, currPopup, showEndTripPopup, onPopupClose } = useGuidePopup()
const loading = ref(false)
const paying = ref(false)
const detail = ref<Record<string, unknown>>({})
const walletPwd = ref('')
const frozenTime = ref(0)
const cutRemainMs = ref(0)
const isTempFrozen = ref(false)
const isEBikeLock = ref(false)
const wxScore = ref(false)
let frozenPollTimer: ReturnType<typeof setInterval> | null = null
let cutTimer: ReturnType<typeof setInterval> | null = null
let wxScorePollTimer: ReturnType<typeof setInterval> | null = null
let routeOrderId = ''

const payCost = computed(() => Number(detail.value.payCost ?? detail.value.cost ?? 0))
const waitPayMoney = computed(() => calcWaitPayMoney(detail.value))
const paid = computed(() => Number(detail.value.izPaid) === 4)
const orderId = computed(() => String(detail.value.id || detail.value.orderId || routeOrderId || ''))
const izPaid = computed(() => Number(detail.value.izPaid))
const durationText = computed(() => formatRidingTime(detail.value.ridingTime))
const distanceText = computed(() => `${formatMile(detail.value.mile, 2)} km`)
const showObjectionEntry = computed(() => {
  if (paid.value || wxScore.value) return false
  if ([1, 2].includes(izPaid.value)) return false
  const complained = detail.value.izComplained
  if (complained === 0 || complained === 1 || complained === '0' || complained === '1') return false
  return complained === -1 || complained === '-1' || complained == null
})
const isRedEnvelopeCar = computed(
  () => detail.value.izRedPaket === true || detail.value.izRedPaket === 1,
)
const redRewardYuan = computed(() => redEnvelopeRewardYuan(detail.value))
const showRedReward = computed(
  () => isRedEnvelopeCar.value && paid.value && Number(redRewardYuan.value) > 0,
)
const rewardBg = computed(() => getRedEnvelopeCfg('redEnvelopeRewardBg'))
const showRepairEntry = ref(false)
const showShortRepair = ref(false)
const discountOpen = ref(true)
const walletRecharge = ref<number | null>(null)
const walletPresent = ref<number | null>(null)
let shortRepairAsked = false

const rechargeBalance = computed(() => {
  const fromDetail = detail.value.recharge
  if (fromDetail != null && fromDetail !== '') return Number(fromDetail)
  return Number(walletRecharge.value ?? 0)
})
const presentBalance = computed(() => {
  const fromDetail = detail.value.present
  if (fromDetail != null && fromDetail !== '') return Number(fromDetail)
  return Number(walletPresent.value ?? 0)
})
const showWalletBuckets = computed(
  () =>
    detail.value.recharge != null ||
    detail.value.present != null ||
    walletRecharge.value != null ||
    walletPresent.value != null,
)

const discountTotal = computed(() => {
  const d = detail.value
  const parts = [
    Number(d.ridingCardCount || 0),
    Number(d.discountCount || 0),
    Number(d.randomCount || 0),
    Number(d.redPaketCarDeduct || 0),
    Number(d.redPaketParkingDeduct || 0),
  ]
  const sum = parts.reduce((a, b) => a + (Number.isFinite(b) ? b : 0), 0)
  if (sum > 0) return sum
  return Number(d.deduction || 0)
})
const showDiscountBlock = computed(
  () =>
    discountTotal.value > 0 ||
    Boolean(detail.value.izActityFree || detail.value.izFreeN || detail.value.deduction),
)

const frozenTip = computed(() => {
  if (![1, 2].includes(izPaid.value)) return ''
  if (!frozenTime.value) return t('pay.frozenTip')
  if (cutRemainMs.value <= 0) return t('pay.frozenExpired')
  const sec = Math.ceil(cutRemainMs.value / 1000)
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return t('pay.frozenCountdown', {
    time: `${m}:${String(s).padStart(2, '0')}`,
  })
})

onShow(() => setNavTitle(t('pay.title')))

onLoad((q) => {
  routeOrderId = decodeURIComponent((q?.orderId as string) || '')
  isTempFrozen.value = String(q?.isTempFrozon || q?.isTempFrozen || '') === 'true' || q?.isTempFrozon === '1'
  isEBikeLock.value =
    String(q?.isEBikeLock || '') === 'true' || q?.isEBikeLock === '1' || q?.isEBikeLock === true
  wxScore.value = String(q?.wxScore || '') === '1' || q?.wxScore === 'true'
})

function userPin() {
  return String(user.userInfo?.pin || storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
}

function stopFrozenTimers() {
  if (frozenPollTimer) {
    clearInterval(frozenPollTimer)
    frozenPollTimer = null
  }
  if (cutTimer) {
    clearInterval(cutTimer)
    cutTimer = null
  }
  if (wxScorePollTimer) {
    clearInterval(wxScorePollTimer)
    wxScorePollTimer = null
  }
}

function startWxScorePoll() {
  if (wxScorePollTimer) clearInterval(wxScorePollTimer)
  wxScorePollTimer = setInterval(async () => {
    if (!wxScore.value || paid.value) {
      if (wxScorePollTimer) {
        clearInterval(wxScorePollTimer)
        wxScorePollTimer = null
      }
      return
    }
    try {
      const res = await loadLastOrder()
      if (res.success && res.data) {
        detail.value = res.data as Record<string, unknown>
        if (Number(detail.value.izPaid) === 4) {
          temp.resetRide()
          if (wxScorePollTimer) {
            clearInterval(wxScorePollTimer)
            wxScorePollTimer = null
          }
        }
      }
    } catch (e) {
      logger.warn('wxScore poll soft fail', e)
    }
  }, 3000)
}

function startCountdown(ts: number) {
  if (cutTimer) clearInterval(cutTimer)
  const end = new Date(ts).valueOf()
  const longTime = 1000 * 60 * 3
  const tick = () => {
    const diff = Date.now() - end
    cutRemainMs.value = Math.max(longTime - diff, 0)
  }
  tick()
  cutTimer = setInterval(tick, 1000)
  if ([1, 2].includes(izPaid.value)) {
    uni.showModal({
      title: t('pay.title'),
      content: t('pay.frozenPayWindow'),
      showCancel: false,
    })
  }
}

async function checkFrozenOnce(oid: string): Promise<boolean> {
  try {
    const res = await queryFrozen({ orderId: oid })
    if (!res.success || !res.data) return false
    const data = res.data as {
      payCost?: number
      frozenTime?: number
      izPaid?: number
      izOverDistance?: boolean
      izOverTime?: boolean
      izFrozen?: boolean
    }
    if (data.payCost != null) detail.value = { ...detail.value, payCost: data.payCost }
    if (data.izPaid != null) detail.value = { ...detail.value, izPaid: data.izPaid }
    frozenTime.value = Number(data.frozenTime || 0)
    if (data.izPaid === 3 || data.izPaid === 4) {
      stopFrozenTimers()
      return true
    }
    if (data.izOverDistance || data.izOverTime || data.izFrozen === false) {
      stopFrozenTimers()
      uni.showModal({
        title: t('pay.tripUpdated'),
        content: t('pay.tripUpdatedHint'),
        showCancel: false,
        success: (r) => {
          if (r.confirm) navigate('reLaunch', '/pages/riding/riding')
        },
      })
      return true
    }
    if (frozenTime.value) startCountdown(frozenTime.value)
    return true
  } catch (e) {
    logger.warn('queryFrozen soft fail', e)
    return false
  }
}

function startFrozenPoll(oid: string) {
  stopFrozenTimers()
  void checkFrozenOnce(oid)
  frozenPollTimer = setInterval(() => {
    if (paid.value || Number(detail.value.izPaid) === 3) {
      stopFrozenTimers()
      return
    }
    void checkFrozenOnce(oid)
  }, 3000)
}

async function maybeHandleFrozen() {
  const oid = orderId.value
  if (!oid) return
  const paidState = Number(detail.value.izPaid)
  // Already locked / settled — skip
  if (paidState === 3 || paidState === 4) return
  // Legacy unpaid-from-ride gate: skip frozen countdown
  if (isEBikeLock.value) return

  if (isTempFrozen.value) {
    try {
      await setFrozenOrder({
        orderId: oid,
        serviceId: detail.value.serviceId || storage.get('serviceId', ''),
        userPin: userPin(),
        type: 1,
        imei: detail.value.imei,
      })
    } catch (e) {
      logger.warn('setFrozenOrder soft fail', e)
    }
  }

  if (paidState === 1 || paidState === 2 || isTempFrozen.value) {
    startFrozenPoll(oid)
  }
}

async function maybeShortTripRepair() {
  // Legacy: only after paid (izPaid===4)
  if (shortRepairAsked || !showRepairEntry.value || !paid.value) return
  if (detail.value.izRepair === 1 || detail.value.izRepair === true) return
  const mins = Number(detail.value.ridingTime || 0) / 1000 / 60
  if (!Number.isFinite(mins) || mins >= 3) return
  shortRepairAsked = true
  showShortRepair.value = true
}

async function maybeEndTripPopup() {
  if (!paid.value) return
  await showEndTripPopup()
}

async function refresh() {
  loading.value = true
  const res = await loadOrder(routeOrderId)
  loading.value = false
  if (res.success && res.data) {
    detail.value = res.data as Record<string, unknown>
    // Backend may mark wx-score order even without query flag
    if (Boolean((res.data as { izWxScoreOrder?: boolean }).izWxScoreOrder)) {
      wxScore.value = true
    }
    if (wxScore.value && !paid.value) startWxScorePoll()
    await maybeHandleFrozen()
    void maybeShortTripRepair()
    void maybeEndTripPopup()
  }
}

async function loadRepairFlag() {
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getConfigBaseItem(sid ? { serviceId: sid } : {})
    if (res.success && res.data) {
      const data = res.data as { izCanAfterRidingRepair?: boolean }
      showRepairEntry.value = Boolean(data.izCanAfterRidingRepair)
    }
  } catch (e) {
    logger.warn('pay repair flag soft fail', e)
  }
}

async function loadWalletBuckets() {
  try {
    const res = await getWalletInfo()
    if (!res.success || !res.data) return
    const data = res.data as { recharge?: number; present?: number }
    if (data.recharge != null) walletRecharge.value = Number(data.recharge)
    if (data.present != null) walletPresent.value = Number(data.present)
  } catch (e) {
    logger.warn('pay wallet buckets soft fail', e)
  }
}

onMounted(() => {
  user.hydrateFromStorage()
  void loadRepairFlag()
  void loadWalletBuckets()
  void refresh()
})

onHide(() => stopFrozenTimers())
onUnmounted(() => stopFrozenTimers())

async function onPay() {
  if (paying.value || paid.value) return
  const ok = await checkPayCertification()
  if (!ok) return
  paying.value = true
  try {
    const res = await settleLastOrder(detail.value, { walletPwd: walletPwd.value })
    if (res.success) {
      temp.resetRide()
      stopFrozenTimers()
      uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
      await refresh()
      void showEndTripPopup()
      return
    }
    if (res.code === 'NEED_WALLET_PWD') {
      uni.showToast({ title: res.msg || t('pay.needWalletPwd'), icon: 'none' })
      return
    }
    if (res.code && Number(res.code) === 15043) {
      uni.showModal({ title: t('pay.title'), content: res.msg || t('common.loading'), showCancel: false })
      return
    }
    if (res.code && Number(res.code) === 15038) {
      uni.showModal({
        title: t('pay.title'),
        content: t('ride.returnBike'),
        showCancel: false,
        success: (r) => {
          if (r.confirm) navigate('reLaunch', '/pages/riding/riding')
        },
      })
    }
  } finally {
    paying.value = false
  }
}

function goHome() {
  navigate('reLaunch', '/pages/home/home')
}

function goRedTips() {
  navigate('to', '/pages-sub/ride/red-envelope-tips/red-envelope-tips')
}

function goRepair() {
  if (detail.value.izRepair === 1 || detail.value.izRepair === true) {
    uni.showToast({ title: t('pay.alreadyRepaired'), icon: 'none' })
    return
  }
  showShortRepair.value = false
  const cid = String(detail.value.carId || '')
  const oid = orderId.value
  const qs = [
    cid ? `carId=${encodeURIComponent(cid)}` : '',
    oid ? `orderId=${encodeURIComponent(oid)}` : '',
  ]
    .filter(Boolean)
    .join('&')
  navigate('to', `/pages-sub/support/repair/repair${qs ? `?${qs}` : ''}`)
}

function onShortRepairConfirm() {
  goRepair()
}

function goTripMap() {
  if (!orderId.value) return
  navigate(
    'to',
    `/pages-sub/ride/trip-map/trip-map?orderId=${encodeURIComponent(orderId.value)}`,
  )
}

function goObjection() {
  if (!orderId.value) return
  navigate('to', `/pages-sub/support/objection/objection?orderId=${encodeURIComponent(orderId.value)}`)
}

function goCostDetail() {
  if (!orderId.value) return
  navigate('to', `/pages-sub/pay/cost-detail/cost-detail?orderId=${encodeURIComponent(orderId.value)}`)
}
</script>

<style scoped lang="scss">
.title {
  font-size: 30rpx;
  color: #666;
}
.amount {
  margin-top: 16rpx;
  font-size: 64rpx;
  font-weight: 700;
}
.balance-box {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx 32rpx;
  margin-top: 16rpx;
}
.balance {
  font-size: 24rpx;
  color: #888;
}
.reward {
  position: relative;
  overflow: hidden;
  min-height: 280rpx;
  padding: 40rpx 24rpx;
  text-align: center;
}
.reward__bg {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  opacity: 0.85;
}
.reward__body {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
}
.reward__done {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.reward__amount {
  font-size: 36rpx;
  font-weight: 700;
  color: #ff461e;
}
.reward__hint {
  font-size: 22rpx;
  color: #999;
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8rpx;
}
.reward__link {
  color: #ff5936;
}
.hint,
.meta {
  margin-top: 12rpx;
  color: #999;
}
.frozen {
  margin-top: 12rpx;
  color: #ff8401;
  font-size: 26rpx;
  line-height: 1.4;
}
.section-title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.row {
  display: flex;
  justify-content: space-between;
  padding: 16rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.row.link {
  color: #3aa0e8;
  border-bottom: none;
  justify-content: flex-end;
}
.row.sub {
  padding-left: 16rpx;
  font-size: 26rpx;
  color: #888;
}
.warn {
  color: #ff8401;
}
.input {
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-top: 8rpx;
}
.actions {
  padding: 24rpx 32rpx;
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
</style>
