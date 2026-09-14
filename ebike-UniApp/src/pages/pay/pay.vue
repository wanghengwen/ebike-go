<template>
  <view class="page">
    <view
      class="hint-bar"
      v-if="[1, 2].includes(izPaid) && frozenTime && cutRemainMs > 0"
    >
      <image v-if="iconClock" class="hint-bar__clock" :src="iconClock" mode="aspectFit" />
      <text class="hint-bar__text">{{ t('pay.frozenPayWindow') }}</text>
      <text class="hint-bar__time">{{ frozenCountdownText }}</text>
    </view>

    <scroll-view class="scroll" scroll-y :style="{ paddingBottom: dockPad + 'px' }">
      <view v-if="loading" class="loading-tip">{{ t('common.loading') }}</view>

      <template v-else>
        <view class="pay-card">
          <view v-if="paid" class="hero">
            <text class="hero__label">{{ t('pay.paidLabel') }}</text>
            <text class="hero__value">{{ paidAmountText }}</text>
            <text class="hero__unit">{{ t('pay.yuan') }}</text>
          </view>
          <view v-else class="hero">
            <text class="hero__label">{{ t('pay.unpaidLabel') }}</text>
            <text class="hero__value">{{ formatMoney(waitPayMoney) }}</text>
            <text class="hero__unit">{{ t('pay.yuan') }}</text>
          </view>

          <view class="hint tip" v-if="wxScore && !paid">{{ t('pay.wxScorePaying') }}</view>
          <view class="hint tip" v-else-if="wxScore && paid">{{ t('pay.wxScorePaid') }}</view>
          <view class="frozen" v-if="frozenTip && !([1, 2].includes(izPaid) && frozenTime && cutRemainMs > 0)">
            {{ frozenTip }}
          </view>

          <view v-if="!paid && showWalletBuckets" class="balance-box">
            <text class="balance">
              {{ t('pay.rechargeBalance', { m: formatMoney(rechargeBalance) }) }}
            </text>
            <text class="balance">
              {{ t('pay.presentBalance', { m: formatMoney(presentBalance) }) }}
            </text>
          </view>

          <view v-if="!paid" class="mile-time">
            <text>
              {{ t('pay.rideDuration') }}
              <text class="mile-time__val">{{ formatRidingTimeUnit(detail.ridingTime) }}</text>
            </text>
            <text v-if="!tempHideOrderPrice">
              ，{{ t('pay.rideMile') }}
              <text class="mile-time__val">
                {{ formatMile(detail.mile) }}{{ t('pay.kmUnit') }}
              </text>
            </text>
          </view>

          <!-- 待支付：费用明细 -->
          <view v-if="!paid && Number(payCost) > 0" class="cost">
            <view class="cost-head">
              <text class="cost-head__title">{{ t('pay.costDetail') }}</text>
              <image
                v-if="iconQuestion"
                class="cost-head__icon"
                :src="iconQuestion"
                mode="aspectFit"
                @click="goBillingRules"
              />
            </view>
            <view class="cost-body">
              <view class="cost-row" @click="isOpen = !isOpen">
                <text class="cost-row__label">{{ t('pay.rideCost') }}</text>
                <view class="cost-row__right">
                  <text class="cost-row__price">
                    {{ formatMoney(detail.originCost) }}{{ t('pay.yuan') }}
                  </text>
                  <image
                    v-if="!tempHideOrderPrice && arrowToggle(isOpen)"
                    class="cost-row__arrow"
                    :src="arrowToggle(isOpen)"
                    mode="aspectFit"
                  />
                </view>
              </view>
              <template v-if="isOpen && !tempHideOrderPrice">
                <view v-if="detail.startPrice != null" class="cost-child">
                  <text>{{ t('pay.startPrice') }}</text>
                  <text>{{ formatMoney(detail.startPrice) }}{{ t('pay.yuan') }}</text>
                </view>
                <view v-if="detail.timeCost != null || detail.durationCost != null" class="cost-child">
                  <text>{{ t('pay.durationCost') }}</text>
                  <text>
                    {{ formatMoney(detail.timeCost ?? detail.durationCost) }}{{ t('pay.yuan') }}
                  </text>
                </view>
                <view v-if="detail.mileCost != null" class="cost-child">
                  <text>{{ t('pay.mileCost') }}</text>
                  <text>{{ formatMoney(detail.mileCost) }}{{ t('pay.yuan') }}</text>
                </view>
              </template>

              <view
                v-if="detail.penalty || detail.helmetPenalty || detail.dispatchCost"
                class="cost-row"
                @click="isOpen1 = !isOpen1"
              >
                <text class="cost-row__label">{{ t('pay.otherCost') }}</text>
                <view class="cost-row__right">
                  <text class="cost-row__price">
                    {{
                      formatMoney(
                        Number(detail.penalty || 0) +
                          Number(detail.helmetPenalty || 0) +
                          Number(detail.dispatchCost || 0),
                      )
                    }}{{ t('pay.yuan') }}
                  </text>
                  <image
                    v-if="arrowToggle(isOpen1)"
                    class="cost-row__arrow"
                    :src="arrowToggle(isOpen1)"
                    mode="aspectFit"
                  />
                </view>
              </view>
              <template v-if="isOpen1">
                <view v-if="detail.dispatchCost != null" class="cost-child">
                  <text>{{ t('pay.dispatchCost') }}</text>
                  <text>{{ formatMoney(detail.dispatchCost) }}{{ t('pay.yuan') }}</text>
                </view>
                <view v-if="detail.helmetPenalty != null" class="cost-child">
                  <text>{{ t('pay.helmetPenalty') }}</text>
                  <text>{{ formatMoney(detail.helmetPenalty) }}{{ t('pay.yuan') }}</text>
                </view>
              </template>

              <view v-if="showDiscountBlock" class="cost-row" @click="isOpen2 = !isOpen2">
                <text class="cost-row__label">{{ t('pay.enjoyDiscount') }}</text>
                <view class="cost-row__right">
                  <text class="cost-row__price warn">
                    -{{ formatMoney(discountTotal) }}{{ t('pay.yuan') }}
                  </text>
                  <image
                    v-if="arrowToggle(isOpen2)"
                    class="cost-row__arrow"
                    :src="arrowToggle(isOpen2)"
                    mode="aspectFit"
                  />
                </view>
              </view>
              <template v-if="isOpen2 && showDiscountBlock">
                <view v-if="detail.izActityFree || detail.izFreeN" class="cost-child">
                  <text>
                    {{ t('pay.actFree') }}
                    <text class="hint">{{ t('pay.actFreeHint') }}</text>
                  </text>
                  <text class="warn">{{ t('pay.actFreeTag') }}</text>
                </view>
                <view v-if="Number(detail.ridingCardCount) > 0" class="cost-child">
                  <text>
                    {{ t('pay.ridingCardAct') }}
                    <text class="hint">{{ t('pay.actNoPenaltyHint') }}</text>
                  </text>
                  <text class="warn">-{{ formatMoney(detail.ridingCardCount) }}{{ t('pay.yuan') }}</text>
                </view>
                <view v-if="Number(detail.discountCount) > 0" class="cost-child">
                  <text>
                    {{ t('pay.discountAct', { n: discountFold }) }}
                    <text class="hint">{{ t('pay.discountNoPenaltyHint') }}</text>
                  </text>
                  <text class="warn">-{{ formatMoney(detail.discountCount) }}{{ t('pay.yuan') }}</text>
                </view>
                <view v-if="Number(detail.randomCount) > 0" class="cost-child">
                  <text>{{ t('pay.randomDeduct') }}</text>
                  <text class="warn">-{{ formatMoney(detail.randomCount) }}{{ t('pay.yuan') }}</text>
                </view>
                <view
                  v-if="
                    Number(detail.redPaketCarDeduct) > 0 || Number(detail.redPaketParkingDeduct) > 0
                  "
                  class="cost-child"
                >
                  <text>{{ t('pay.redEnvelopeDeduct') }}</text>
                  <text class="warn">
                    -{{
                      formatMoney(
                        Number(detail.redPaketCarDeduct || 0) +
                          Number(detail.redPaketParkingDeduct || 0),
                      )
                    }}{{ t('pay.yuan') }}
                  </text>
                </view>
                <view
                  v-if="
                    Number(detail.deduction) > 0 &&
                    !Number(detail.ridingCardCount) &&
                    !Number(detail.discountCount) &&
                    !Number(detail.randomCount)
                  "
                  class="cost-child"
                >
                  <text>{{ t('pay.otherDeduct') }}</text>
                  <text class="warn">-{{ formatMoney(detail.deduction) }}{{ t('pay.yuan') }}</text>
                </view>
              </template>
              <view class="cost-line" />
            </view>
            <view class="subtotal">
              <text>
                {{ t('pay.subtotal') }}
                <text class="subtotal__price">{{ formatMoney(payCost) }}</text>
                {{ t('pay.yuan') }}
              </text>
            </view>
          </view>

          <!-- 已支付：费用明细入口 + 骑行数据 -->
          <view v-if="paid" class="cost cost--paid">
            <view class="cost-paid-head" @click="goCostDetail">
              <text class="cost-paid-head__title">{{ t('pay.costDetail') }}</text>
              <image
                v-if="iconRight"
                class="cost-paid-head__arrow"
                :src="iconRight"
                mode="aspectFit"
              />
            </view>
            <view class="cost-paid-row">
              <text>{{ t('pay.rideDuration') }}</text>
              <text class="cost-paid-row__val">{{ formatRidingTimeUnit(detail.ridingTime) }}</text>
            </view>
            <view v-if="!tempHideOrderPrice" class="cost-paid-row">
              <text>{{ t('pay.rideMile') }}</text>
              <text class="cost-paid-row__val">
                {{ formatMile(detail.mile) }}{{ t('pay.kmUnit') }}
              </text>
            </view>
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

        <view class="card wallet-pwd" v-if="!paid && !wxScore && waitPayMoney > 0">
          <view class="wallet-pwd__title">
            {{ t('pay.waitPay') }} {{ formatMoney(waitPayMoney) }}{{ t('pay.yuan') }}
          </view>
          <input
            class="input"
            password
            v-model="walletPwd"
            :placeholder="t('pay.walletPwdPlaceholder')"
            maxlength="6"
          />
        </view>
      </template>
    </scroll-view>

    <view class="dock" id="pay-dock">
      <view class="extra-view">
        <view v-if="showObjectionEntry" class="extra-view__item" @click="goObjection">
          <image
            v-if="iconObjection"
            class="extra-view__icon"
            :src="iconObjection"
            mode="aspectFit"
          />
          <text>{{ t('support.objection') }}</text>
        </view>
        <view v-if="showRepairEntry" class="extra-view__item" @click="goRepair">
          <image v-if="iconRepair" class="extra-view__icon" :src="iconRepair" mode="aspectFit" />
          <text>{{ t('support.bikeRepair') }}</text>
        </view>
      </view>
      <button
        v-if="paid"
        class="dock-btn"
        :style="homeBtnStyle"
        @click="goHome"
      >
        {{ t('pay.backHome') }}
      </button>
      <button
        v-else-if="!wxScore"
        class="dock-btn"
        :style="payBtnStyle"
        @click="onPay"
      >
        {{ payButtonText }}
      </button>
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
import { usePay } from '@/features/pay/usePay'
import { checkPayCertification } from '@/features/pay/checkPayCertification'
import { getRedEnvelopeCfg, redEnvelopeRewardYuan } from '@/features/bike/redEnvelope'
import { formatMoney, formatMile, formatRidingTimeUnit } from '@/shared/format'
import { getBrandColor, getButtonWhiteColor, getTenantConfig } from '@/shared/config'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
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
const frozenReady = ref(false)
const dockPad = ref(200)
let frozenPollTimer: ReturnType<typeof setInterval> | null = null
let cutTimer: ReturnType<typeof setInterval> | null = null
let wxScorePollTimer: ReturnType<typeof setInterval> | null = null
let routeOrderId = ''

const payCost = computed(() => Number(detail.value.payCost ?? detail.value.cost ?? 0))
const waitPayMoney = computed(() => calcWaitPayMoney(detail.value))
const paid = computed(() => Number(detail.value.izPaid) === 4)
const orderId = computed(() => String(detail.value.id || detail.value.orderId || routeOrderId || ''))
const izPaid = computed(() => Number(detail.value.izPaid))
const tempHideOrderPrice = computed(() =>
  Boolean(getTenantConfig().customSetting?.tempHideOrderPrice),
)
const paidAmountText = computed(() =>
  formatMoney(detail.value.hasPaid ?? detail.value.payCost ?? 0),
)
const iconClock = computed(() => getIconCfg('clock'))
const iconQuestion = computed(() => getIconCfg('qusetionIcon'))
const iconObjection = computed(() => getIconCfg('qusetion'))
const iconRepair = computed(() => getIconCfg('carRepair'))
const iconRight = computed(() => getMapCfg('iconRight'))
const arrowDown = computed(() => getIconCfg('bottomArrow'))
const arrowUp = computed(() => getIconCfg('topArrow'))
const homeBtnStyle = computed(() => ({
  backgroundColor: getBrandColor(),
  borderColor: getBrandColor(),
  color: '#1E4A38',
}))
const payBtnStyle = computed(() => ({
  backgroundColor: getBrandColor(),
  borderColor: getBrandColor(),
  color: getButtonWhiteColor(),
}))
const payButtonText = computed(() => {
  if (waitPayMoney.value > 0) {
    return `${t('pay.goPay')}${formatMoney(waitPayMoney.value)}${t('pay.yuan')}`
  }
  return t('pay.goPay')
})
const showObjectionEntry = computed(() => {
  if (!frozenReady.value) return false
  const complained = detail.value.izComplained
  if (complained !== -1 && complained !== '-1') return false
  if (
    !paid.value &&
    [1, 2].includes(izPaid.value) &&
    frozenTime.value &&
    cutRemainMs.value > 0
  ) {
    return false
  }
  return true
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
const isOpen = ref(true)
const isOpen1 = ref(true)
const isOpen2 = ref(true)
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
const discountFold = computed(() => {
  const d = Number(detail.value.discount)
  if (!Number.isFinite(d) || d >= 1) return 10
  return Math.round(d * 10)
})

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
const frozenCountdownText = computed(() => {
  if (cutRemainMs.value <= 0) return '0:00'
  const sec = Math.ceil(cutRemainMs.value / 1000)
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m}:${String(s).padStart(2, '0')}`
})

function arrowToggle(open: boolean) {
  return open ? arrowUp.value : arrowDown.value
}

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
  if (paidState === 3 || paidState === 4) return
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
  frozenReady.value = false
  const res = await loadOrder(routeOrderId)
  loading.value = false
  if (res.success && res.data) {
    detail.value = res.data as Record<string, unknown>
    if (Boolean((res.data as { izWxScoreOrder?: boolean }).izWxScoreOrder)) {
      wxScore.value = true
    }
    if (wxScore.value && !paid.value) startWxScorePoll()
    await maybeHandleFrozen()
    void maybeShortTripRepair()
    void maybeEndTripPopup()
  }
  frozenReady.value = true
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

function measureDock() {
  uni
    .createSelectorQuery()
    .select('#pay-dock')
    .boundingClientRect((rect) => {
      if (rect && !Array.isArray(rect) && rect.height) {
        dockPad.value = rect.height + 16
      }
    })
    .exec()
}

onMounted(() => {
  user.hydrateFromStorage()
  void loadRepairFlag()
  void loadWalletBuckets()
  void refresh()
  setTimeout(measureDock, 100)
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
      setTimeout(measureDock, 100)
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
  const params = detail.value?.deviceTrajectory || []
  navigate(
    'to',
    `/pages-sub/ride/trip-map/trip-map?params=${encodeURIComponent(JSON.stringify(params))}`,
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

function goBillingRules() {
  navigate('to', '/pages-sub/account/billing-rules/billing-rules')
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f8f8f8;
}

.scroll {
  flex: 1;
  height: 0;
  background: #f8f8f8;
}

.loading-tip {
  padding: 48rpx 32rpx;
  text-align: center;
  color: #999;
}

.hint-bar {
  display: flex;
  align-items: center;
  margin: 32rpx 32rpx 0;
  padding: 0 32rpx;
  height: 80rpx;
  background: rgba(255, 146, 43, 0.1);
  border: 2rpx solid rgba(255, 146, 43, 0.5);
  border-radius: 32rpx;
  flex-shrink: 0;
}

.hint-bar__clock {
  width: 32rpx;
  height: 32rpx;
}

.hint-bar__text {
  flex: 1;
  margin-left: 16rpx;
  font-size: 28rpx;
  font-weight: 500;
  color: #ff922b;
}

.hint-bar__time {
  font-size: 28rpx;
  font-weight: 500;
  color: #ff922b;
}

.pay-card {
  margin: 32rpx 32rpx 0;
  padding: 36rpx 32rpx 52rpx;
  background: #fff;
  border-radius: 32rpx;
}

.hero {
  margin-bottom: 16rpx;
  font-size: 48rpx;
  font-weight: 800;
  color: #1a1a1a;
}

.hero__value {
  margin: 0 8rpx;
}

.hero__unit {
  font-size: 28rpx;
  font-weight: 800;
}

.balance-box {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx 24rpx;
}

.balance {
  font-size: 28rpx;
  color: #666;
}

.hint,
.frozen {
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #999;
}

.frozen {
  color: #ff8401;
  line-height: 1.4;
}

.mile-time {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  margin: 24rpx 0;
  padding: 24rpx 0;
  border-top: 2rpx solid #f6f6f6;
  border-bottom: 2rpx solid #f6f6f6;
  font-size: 28rpx;
  color: #666;
}

.mile-time__val {
  color: #333;
  font-weight: 700;
}

.cost {
  margin-top: 8rpx;
}

.cost-head {
  display: flex;
  align-items: center;
}

.cost-head__title {
  margin-right: 8rpx;
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}

.cost-head__icon {
  width: 32rpx;
  height: 32rpx;
}

.cost-body {
  margin-top: 36rpx;
}

.cost-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 32rpx;
}

.cost-row__label {
  font-size: 28rpx;
  font-weight: 700;
  color: #333;
}

.cost-row__right {
  display: flex;
  align-items: center;
}

.cost-row__price {
  margin-right: 8rpx;
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
}

.cost-row__arrow {
  width: 24rpx;
  height: 24rpx;
}

.cost-child {
  display: flex;
  justify-content: space-between;
  margin: 16rpx 0;
  font-size: 28rpx;
  color: #666;
}

.cost-line {
  height: 2rpx;
  margin: 24rpx 0 22rpx;
  background: #f6f6f6;
}

.subtotal {
  text-align: right;
  font-size: 28rpx;
  color: #333;
}

.subtotal__price {
  margin: 0 8rpx;
  font-size: 40rpx;
  font-weight: 700;
}

.cost--paid {
  margin-top: 24rpx;
}

.cost-paid-head {
  display: flex;
  align-items: center;
}

.cost-paid-head__title {
  font-size: 24rpx;
  color: #999;
}

.cost-paid-head__arrow {
  width: 24rpx;
  height: 24rpx;
  margin-left: 8rpx;
}

.cost-paid-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20rpx;
  font-size: 28rpx;
  color: #333;
}

.cost-paid-row__val {
  font-weight: 600;
}

.warn {
  color: #ff461d;
}

.hint {
  color: #999;
  font-size: 24rpx;
}

.card {
  background: #fff;
  border-radius: 32rpx;
  padding: 32rpx;
  margin: 24rpx 32rpx;
}

.wallet-pwd__title {
  font-weight: 700;
  margin-bottom: 16rpx;
}

.input {
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 24rpx;
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

.dock {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 10;
  padding: 10rpx 48rpx calc(32rpx + env(safe-area-inset-bottom));
  background: #f8f8f8;
  box-sizing: border-box;
}

.extra-view {
  width: 654rpx;
  max-width: 100%;
  margin: 10rpx auto 22rpx;
  display: flex;
  justify-content: space-around;
}

.extra-view__item {
  display: flex;
  align-items: center;
  height: 40rpx;
  font-size: 28rpx;
  line-height: 40rpx;
  color: #666;
}

.extra-view__icon {
  width: 40rpx;
  height: 40rpx;
  margin-right: 8rpx;
}

.dock-btn {
  width: 654rpx;
  max-width: 100%;
  height: 96rpx;
  margin: 0 auto;
  border-radius: 32rpx;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 32rpx;
  font-weight: 600;
  border-width: 0;
  line-height: 96rpx;
}

.dock-btn::after {
  border: none;
}
</style>
