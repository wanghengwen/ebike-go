<template>
  <view class="page">
    <scroll-view scroll-y class="scroll" :class="{ 'scroll--pad': !showOther }">
      <view v-if="rechargeBeforeUse" class="hint">
        <image v-if="warnIcon" class="hint__icon" :src="warnIcon" mode="aspectFit" />
        <text class="hint__text">{{ t('pay.rechargeBeforeUse', { m: fenToYuan(rechargeCost) }) }}</text>
      </view>

      <view class="card amount-card">
        <view class="card__title">{{ t('pay.rechargeAmount') }}</view>
        <view v-if="loading" class="loading">{{ t('common.loading') }}</view>
        <view v-else class="grid">
          <view
            v-for="(item, idx) in list"
            :key="'p' + idx"
            class="cell"
            :class="{ active: !otherMode && selected === idx }"
            :style="cellStyle(!otherMode && selected === idx)"
            @tap="selectPackage(idx)"
          >
            <text class="cell__amt">{{ displayAmount(item) }}{{ t('pay.yuan') }}</text>
            <text v-if="item.activityType && item.details" class="cell__gift">
              {{ t('pay.giftPrefix') }}{{ item.details }}
            </text>
            <text v-else-if="item.present" class="cell__gift">
              {{ t('pay.giftPrefix') }}{{ fenToYuan(item.present as number) }}
            </text>
            <view v-if="item.activityType" class="cell__tag">{{ t('pay.activity') }}</view>
          </view>
          <view
            class="cell"
            :class="{ active: otherMode }"
            :style="cellStyle(otherMode)"
            @tap="openOther"
          >
            <text class="cell__amt cell__amt--other">
              {{ otherMode && otherFen > 0 ? `${fenToYuan(otherFen)}${t('pay.yuan')}` : t('pay.otherAmount') }}
            </text>
          </view>
        </view>
      </view>

      <view class="card channel-card">
        <view class="card__title">{{ t('pay.payChannel') }}</view>
        <view class="channel-row">
          <image v-if="wechatPayIcon" class="channel-row__icon" :src="wechatPayIcon" mode="aspectFit" />
          <text class="channel-row__name">{{ t('pay.wechatPay') }}</text>
          <image
            v-if="checkedIcon"
            class="channel-row__check"
            :src="checkedIcon"
            mode="aspectFit"
          />
        </view>
      </view>

      <view class="card" v-if="notice.izOpen && notice.content">
        <view class="card__title">{{ notice.title || t('pay.rechargeNotice') }}</view>
        <rich-text class="notice" :nodes="String(notice.content)" />
      </view>
    </scroll-view>

    <view v-if="!showOther" class="footer">
      <view class="agree">
        <text>{{ t('pay.rechargeAgree') }}</text>
        <text class="link" @tap="openProtocol('reChargeProtocol')">{{ t('pay.rechargeProtocol') }}</text>
      </view>
      <view class="pay-btn" :class="{ disabled: paying }" :style="payBtnStyle" @tap="onPay">
        {{ t('pay.goRecharge') }}
      </view>
    </view>

    <view v-if="showOther" class="mask" @tap="closeOther">
      <view class="pad" @tap.stop>
        <view class="pad__top">
          <text class="pad__title">{{ t('pay.otherAmount') }}</text>
          <text class="pad__cancel" @tap="closeOther">{{ t('common.cancel') }}</text>
        </view>
        <view class="pad__input">¥ {{ otherInput || '0' }}</view>
        <view class="pad__hint">
          {{ t('pay.otherAmountRange', { min: fenToYuan(minFen), max: fenToYuan(maxFen) }) }}
        </view>
        <view class="pad__keys">
          <view class="pad__left">
            <view class="pad__row" v-for="(row, ri) in keyRows" :key="ri">
              <view
                v-for="(k, ki) in row"
                :key="ki"
                class="pad__key"
                :class="{ zero: k === '0' }"
                @tap="onKey(k)"
              >
                {{ k }}
              </view>
            </view>
          </view>
          <view class="pad__right">
            <view class="pad__key del" @tap="onDel">⌫</view>
            <view class="pad__key ok" :style="{ background: brand }" @tap="confirmOther">
              {{ t('pay.recharge') }}
            </view>
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
import { getRechargeList, getRechargeConfig, getRechargeScope } from '@/api/pay'
import { getUseCarConfig } from '@/api/user'
import { createChannelPay, fenToYuan } from '@/features/pay/usePay'
import { checkPayCertification } from '@/features/pay/checkPayCertification'
import { openProtocol } from '@/shared/protocol'
import { getBrandColor, getTenantConfig } from '@/shared/config'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import { formatRichText } from '@/shared/format'

const { t } = useI18n()
const list = ref<Array<Record<string, unknown>>>([])
const selected = ref(0)
const loading = ref(false)
const paying = ref(false)
const otherMode = ref(false)
const otherFen = ref(0)
const showOther = ref(false)
const otherInput = ref('')
const minFen = ref(100)
const maxFen = ref(2000000)
const rechargeBeforeUse = ref(false)
const rechargeCost = ref(0)
const notice = ref<Record<string, unknown>>({})

const keyRows = [
  ['1', '2', '3'],
  ['4', '5', '6'],
  ['7', '8', '9'],
  ['0', '.'],
]

const brand = computed(() => getBrandColor())
const white = computed(
  () => String(getTenantConfig().customSetting?.buttonWhiteColor || '#FFFFFF'),
)
const warnIcon = computed(() => getIconCfg('tipsWarning'))
const wechatPayIcon = computed(() => getIconCfg('wechatPay'))
const checkedIcon = computed(() => getIconCfg('checked_square_round'))
const serviceId = computed(() => storage.get<string>('serviceId', '') || '')

const payBtnStyle = computed(() => ({
  background: brand.value,
  color: white.value,
}))

onShow(() => setNavTitle(t('pay.recharge')))

function cellStyle(active: boolean) {
  if (!active) return {}
  return {
    backgroundColor: brand.value,
    color: white.value,
  }
}

function displayAmount(item: Record<string, unknown>) {
  return fenToYuan(item.amount as number)
}

function selectPackage(idx: number) {
  selected.value = idx
  otherMode.value = false
  otherFen.value = 0
}

function openOther() {
  showOther.value = true
  otherInput.value = otherFen.value > 0 ? fenToYuan(otherFen.value) : ''
}

function closeOther() {
  showOther.value = false
}

function onKey(k: string) {
  if (k === '.' && otherInput.value.includes('.')) return
  if (k === '.' && !otherInput.value) {
    otherInput.value = '0.'
    return
  }
  const next = otherInput.value + k
  const parts = next.split('.')
  if (parts[1] && parts[1].length > 2) return
  otherInput.value = next
}

function onDel() {
  otherInput.value = otherInput.value.slice(0, -1)
}

function confirmOther() {
  const yuan = Number(otherInput.value)
  if (!Number.isFinite(yuan) || yuan <= 0) {
    uni.showToast({ title: t('pay.selectAmount'), icon: 'none' })
    return
  }
  const fen = Math.round(yuan * 100)
  if (fen < minFen.value || fen > maxFen.value) {
    uni.showToast({
      title: t('pay.otherAmountRange', {
        min: fenToYuan(minFen.value),
        max: fenToYuan(maxFen.value),
      }),
      icon: 'none',
    })
    return
  }
  otherFen.value = fen
  otherMode.value = true
  showOther.value = false
}

async function onPay() {
  let amount = 0
  let activeId: string | number = 0
  if (otherMode.value) {
    amount = otherFen.value
    activeId = 12
  } else {
    const item = list.value[selected.value]
    if (!item) {
      uni.showToast({ title: t('pay.selectAmount'), icon: 'none' })
      return
    }
    amount = Number(item.amount || 0)
    activeId = (item.id || item.active_id || 0) as string | number
  }
  if (!amount) {
    uni.showToast({ title: t('pay.selectAmount'), icon: 'none' })
    return
  }
  if (paying.value) return
  const ok = await checkPayCertification()
  if (!ok) return
  paying.value = true
  try {
    const res = await createChannelPay({
      saleType: 'WALLET',
      totalFee: amount,
      saleInfo: { total_fee: amount, active_id: activeId },
    })
    if (res.paid || res.success) {
      uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
      setTimeout(() => navigate('back'), 500)
    } else {
      uni.showToast({ title: res.msg || t('pay.payFail'), icon: 'none' })
    }
  } catch (e) {
    logger.warn('recharge pay fail', e)
    uni.showToast({ title: t('pay.payFail'), icon: 'none' })
  } finally {
    paying.value = false
  }
}

onMounted(async () => {
  loading.value = true
  const { ensureServiceId } = await import('@/shared/ensureServiceId')
  const sid = await ensureServiceId()
  try {
    const [packages, cfg, scope, useCar] = await Promise.all([
      getRechargeList(sid ? { serviceId: sid } : {}),
      getRechargeConfig(sid ? { serviceId: sid } : {}),
      getRechargeScope(sid ? { serviceId: sid } : {}),
      getUseCarConfig(sid ? { serviceId: sid } : {}),
    ])
    if (packages.success && Array.isArray(packages.data)) {
      list.value = packages.data as Array<Record<string, unknown>>
      if (list.value.length) selected.value = 0
    }
    if (cfg.success && cfg.data) {
      const data = { ...(cfg.data as Record<string, unknown>) }
      if (data.content) data.content = formatRichText(data.content)
      notice.value = data
    }
    if (scope.success && scope.data) {
      const data = scope.data as { scope?: number[] }
      if (Array.isArray(data.scope) && data.scope.length >= 2) {
        // Legacy scope is yuan
        minFen.value = Math.round(Number(data.scope[0]) * 100) || 100
        maxFen.value = Math.round(Number(data.scope[1]) * 100) || 2000000
      }
    }
    if (useCar.success && useCar.data) {
      const data = useCar.data as { rechargeBeforeUse?: boolean; rechargeCost?: number }
      rechargeBeforeUse.value = Boolean(data.rechargeBeforeUse)
      rechargeCost.value = Number(data.rechargeCost || 0)
    }
  } catch (e) {
    logger.warn('recharge init soft fail', e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: #f7f8fa;
  position: relative;
}
.scroll {
  height: 100vh;
  box-sizing: border-box;
}
.scroll--pad {
  padding-bottom: 240rpx;
}
.hint {
  margin: 24rpx 32rpx 0;
  padding: 20rpx 24rpx;
  background: #fff7e8;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
}
.hint__icon {
  width: 32rpx;
  height: 32rpx;
  margin-right: 12rpx;
  flex-shrink: 0;
}
.hint__text {
  font-size: 26rpx;
  color: #e65c00;
  line-height: 1.4;
}
.card {
  margin: 32rpx 30rpx 0;
  background: #fff;
  border-radius: 32rpx;
  padding: 32rpx;
  box-sizing: border-box;
}
.card__title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 32rpx;
}
.loading {
  color: #999;
  font-size: 26rpx;
}
.grid {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
}
.cell {
  position: relative;
  width: calc((100% - 20rpx) / 2);
  height: 135rpx;
  border-radius: 22rpx;
  background: #f5f5f5;
  margin-bottom: 20rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  color: #333;
}
.cell__amt {
  font-size: 32rpx;
  font-weight: 600;
}
.cell__amt--other {
  font-size: 28rpx;
}
.cell__gift {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: inherit;
}
.cell__tag {
  position: absolute;
  top: 0;
  right: 0;
  width: 60rpx;
  height: 32rpx;
  background: #ff3200;
  border-radius: 0 12rpx 0 12rpx;
  font-size: 16rpx;
  font-weight: 500;
  color: #fff;
  text-align: center;
  line-height: 32rpx;
}
.channel-row {
  display: flex;
  align-items: center;
}
.channel-row__icon {
  width: 48rpx;
  height: 48rpx;
}
.channel-row__name {
  flex: 1;
  margin-left: 16rpx;
  font-size: 28rpx;
  color: #333;
}
.channel-row__check {
  width: 32rpx;
  height: 32rpx;
}
.notice {
  color: #666;
  font-size: 24rpx;
  line-height: 1.5;
}
.footer {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: #fff;
  padding: 16rpx 32rpx calc(24rpx + env(safe-area-inset-bottom));
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.04);
  z-index: 20;
}
.agree {
  text-align: center;
  color: #999;
  font-size: 22rpx;
  margin-bottom: 16rpx;
}
.link {
  color: var(--brand-color, #3aa0e8);
}
.pay-btn {
  height: 96rpx;
  border-radius: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  font-weight: 500;
}
.pay-btn.disabled {
  opacity: 0.6;
}
.mask {
  position: fixed;
  inset: 0;
  z-index: 1200;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: flex-end;
}
.pad {
  width: 100%;
  background: #fff;
  border-radius: 24rpx 24rpx 0 0;
  padding: 24rpx 24rpx calc(24rpx + env(safe-area-inset-bottom));
  box-sizing: border-box;
}
.pad__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}
.pad__title {
  font-weight: 700;
  font-size: 32rpx;
}
.pad__cancel {
  color: #888;
  padding: 8rpx;
}
.pad__input {
  font-size: 48rpx;
  font-weight: 700;
  padding: 24rpx 8rpx 16rpx;
  border-bottom: 1px solid #f0f0f0;
}
.pad__hint {
  margin-top: 12rpx;
  font-size: 22rpx;
  color: #999;
}
.pad__keys {
  display: flex;
  margin-top: 16rpx;
  gap: 12rpx;
}
.pad__left {
  flex: 3;
}
.pad__row {
  display: flex;
  gap: 12rpx;
  margin-bottom: 12rpx;
}
.pad__key {
  flex: 1;
  background: #f5f6f8;
  border-radius: 12rpx;
  text-align: center;
  padding: 28rpx 0;
  font-size: 36rpx;
  font-weight: 600;
}
.pad__key.zero {
  flex: 2;
}
.pad__right {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}
.pad__key.del {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}
.pad__key.ok {
  flex: 2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
}
</style>
