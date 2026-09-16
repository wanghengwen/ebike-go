<template>
  <view class="page">
    <view class="page-bg" :style="headerBgStyle" />

    <scroll-view scroll-y class="scroll">
      <view class="wallet-card">
        <image v-if="walletBg" class="wallet-card__bg" :src="walletBg" mode="scaleToFill" />
        <view class="wallet-card__content">
          <text class="wallet-card__label">{{ t('account.walletAvailable') }}</text>
          <view class="wallet-card__main">
            <text class="wallet-card__balance">{{ balance }}</text>
            <view class="wallet-card__detail" @tap="go('/pages-sub/account/wallet-records/wallet-records')">
              <image v-if="detailIcon" class="wallet-card__detail-icon" :src="detailIcon" mode="aspectFit" />
            </view>
          </view>
          <view class="wallet-card__line" />
          <view class="wallet-card__parts">
            <view class="wallet-card__part">
              <text class="wallet-card__part-label">{{ t('account.walletRecharge') }}</text>
              <text class="wallet-card__part-val">{{ rechargeYuan }}</text>
            </view>
            <view class="wallet-card__vline" />
            <view class="wallet-card__part">
              <text class="wallet-card__part-label">{{ t('account.walletPresent') }}</text>
              <text class="wallet-card__part-val">{{ presentYuan }}</text>
            </view>
          </view>
        </view>
      </view>

      <view class="actions">
        <view class="action action--primary" @tap="goRecharge">
          <image v-if="rechargeIcon" class="action__icon" :src="rechargeIcon" mode="aspectFit" />
          <text class="action__text action__text--white">{{ t('pay.recharge') }}</text>
        </view>
        <view v-if="izWithdraw && !hasWithdrawRecord" class="action action--light" @tap="onWithdrawEntry">
          <image v-if="withdrawIcon" class="action__icon" :src="withdrawIcon" mode="aspectFit" />
          <text class="action__text action__text--blue">{{ t('account.withdraw') }}</text>
        </view>
        <view v-if="izWithdraw && hasWithdrawRecord" class="action action--light" @tap="goWithdrawProgress">
          <image
            v-if="withdrawProcessIcon"
            class="action__icon"
            :src="withdrawProcessIcon"
            mode="aspectFit"
          />
          <text class="action__text action__text--blue">{{ t('account.withdrawProgress') }}</text>
        </view>
        <view v-if="depositFen > 0" class="action action--light" @tap="goDepositRefund">
          <text class="action__text action__text--blue">{{ t('account.depositRefund') }}</text>
        </view>
      </view>

      <view v-if="izWxScorePayNoPassword" class="more">
        <view class="more__title">{{ t('account.moreServices') }}</view>
        <view class="more__item" @tap="go('/pages-sub/account/pay-score/pay-score')">
          <image v-if="payScoreIcon" class="more__icon" :src="payScoreIcon" mode="aspectFit" />
          <text class="more__name">{{ t('account.noPasswordPay') }}</text>
          <text v-if="wxPayStatusKnown" class="more__des">
            {{ wxPayStatus ? t('account.payScoreOn') : t('account.payScoreOff') }}
          </text>
        </view>
      </view>
    </scroll-view>

    <!-- 提现说明弹层：子节点必须 @tap.stop，微信端 .self 不可靠 -->
    <view v-if="tipsVisible" class="mask" @tap="closeTips">
      <view class="tips-alert" @tap.stop>
        <template v-if="!hasUnpaid">
          <text class="tips-alert__title">{{ t('account.withdrawTipTitle') }}</text>
          <view class="tips-alert__body">
            <text class="tips-alert__line">{{ t('account.withdrawTip1') }}</text>
            <text class="tips-alert__line">{{ t('account.withdrawTip2') }}</text>
            <text class="tips-alert__line">{{ t('account.withdrawTip3') }}</text>
            <text class="tips-alert__line">{{ t('account.withdrawTip4') }}</text>
          </view>
          <view class="tips-alert__agree" @tap.stop="toggleAgree">
            <view class="tips-alert__hit">
              <image
                v-if="!agreed && uncheckIcon"
                class="tips-alert__check"
                :src="uncheckIcon"
                mode="aspectFit"
              />
              <image
                v-else-if="agreed && checkedIcon"
                class="tips-alert__check"
                :src="checkedIcon"
                mode="aspectFit"
              />
              <view v-else class="tips-alert__box" :class="{ on: agreed }" />
            </view>
            <text class="tips-alert__agree-text">{{ t('account.withdrawAgreeRead') }}</text>
          </view>
          <view class="tips-alert__btns">
            <view class="tips-alert__cancel" @tap.stop="closeTips">{{ t('common.cancel') }}</view>
            <view class="tips-alert__confirm" @tap.stop="confirmWithdraw">
              {{ t('account.withdrawContinue') }}
            </view>
          </view>
        </template>
        <template v-else>
          <text class="tips-alert__title">{{ t('common.tip') }}</text>
          <view class="tips-alert__body">
            <text class="tips-alert__line">{{ t('account.withdrawBlockedUnpaid') }}</text>
          </view>
          <view class="tips-alert__btns">
            <view class="tips-alert__confirm tips-alert__confirm--single" @tap.stop="closeTips">
              {{ t('common.gotIt') }}
            </view>
          </view>
        </template>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getWalletInfo } from '@/api/pay'
import { getPersonInfo } from '@/api/user'
import { getWithdrawConfig, getWithdrawRecord } from '@/api/withdraw'
import { getPermissionConfig, getWechatPayScoreRecord } from '@/api/wechatScore'
import { fenToYuan } from '@/features/pay/usePay'
import { useUserStore } from '@/stores/user'
import { useTempDataStore } from '@/stores/tempData'
import { getBrandColor } from '@/shared/config'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { ensureServiceId } from '@/shared/ensureServiceId'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const user = useUserStore()
const temp = useTempDataStore()
const balance = ref('0.00')
const rechargeYuan = ref('0.00')
const presentYuan = ref('0.00')
const depositFen = ref(0)
const izWithdraw = ref(false)
const hasWithdrawRecord = ref(false)
const hasUnpaid = ref(false)
const tipsVisible = ref(false)
const agreed = ref(false)
const izWxScorePayNoPassword = ref(false)
const wxPayStatus = ref(false)
const wxPayStatusKnown = ref(false)

const detailIcon = computed(() => getIconCfg('walletDetail'))
const rechargeIcon = computed(() => getIconCfg('iconRecharge'))
const withdrawIcon = computed(() => getIconCfg('iconWithdraw'))
const withdrawProcessIcon = computed(() => getIconCfg('iconWithdrawProcess') || withdrawIcon.value)
const payScoreIcon = computed(() => getIconCfg('walletNoPasswordPay'))
const headerBg = computed(() => getIconCfg('walletHeaderBg'))
const walletBg = computed(() => getIconCfg('walletBg'))
const uncheckIcon = computed(() => getIconCfg('invoiceUncheck'))
const checkedIcon = computed(() => getIconCfg('checked_square_round'))
const brand = computed(() => getBrandColor())

const headerBgStyle = computed(() =>
  headerBg.value
    ? {
        backgroundImage: `url(${headerBg.value})`,
        backgroundRepeat: 'no-repeat',
        backgroundSize: '100%',
      }
    : { background: `linear-gradient(180deg, ${brand.value} 0%, #f7f8fa 100%)` },
)

onShow(() => setNavTitle(t('pay.wallet')))

async function loadPayScoreConfig(sid: string) {
  try {
    const res = await getPermissionConfig(sid ? { serviceId: sid } : {})
    if (res.success && res.data) {
      izWxScorePayNoPassword.value = Boolean(
        (res.data as { izWxScorePayNoPassword?: boolean }).izWxScorePayNoPassword,
      )
    }
  } catch (e) {
    logger.warn('getPermissionConfig soft fail', e)
  }
  if (!izWxScorePayNoPassword.value) return
  try {
    const login = await new Promise<UniApp.LoginRes>((resolve, reject) => {
      uni.login({ success: resolve, fail: reject })
    })
    const code = String(login.code || '')
    if (!code) return
    const rec = await getWechatPayScoreRecord({ code })
    if (rec.success && rec.data) {
      const state = String(
        (rec.data as { authorization_state?: string; authorizationState?: string })
          .authorization_state ||
          (rec.data as { authorizationState?: string }).authorizationState ||
          '',
      )
      wxPayStatus.value = state === 'AVAILABLE'
      wxPayStatusKnown.value = true
    }
  } catch (e) {
    logger.warn('pay score status soft fail', e)
  }
}

onMounted(async () => {
  user.hydrateFromStorage()
  const sid = await ensureServiceId()
  const [wallet, profile, cfg, record] = await Promise.all([
    getWalletInfo(),
    getPersonInfo(),
    getWithdrawConfig(sid ? { serviceId: sid } : {}),
    getWithdrawRecord({}).catch(() => ({ success: false, data: null })),
  ])
  if (wallet.success && wallet.data) {
    const data = wallet.data as {
      balance?: string | number
      recharge?: number
      present?: number
      deposit?: number
      depositedMount?: number
      depositBalance?: number
    }
    const recharge = Number(data.recharge || 0)
    const present = Number(data.present || 0)
    rechargeYuan.value = fenToYuan(recharge)
    presentYuan.value = fenToYuan(present)
    if (data.balance != null) {
      const b = Number(data.balance)
      balance.value = Number.isInteger(b) || Math.abs(b) >= 1 ? fenToYuan(b) : String(data.balance)
    } else {
      balance.value = fenToYuan(recharge + present)
    }
    depositFen.value = Math.floor(
      Number(data.depositedMount ?? data.deposit ?? data.depositBalance ?? 0),
    )
  }
  if (profile.success && profile.data) {
    user.setUserInfo(profile.data as never)
    const data = profile.data as { depositedMount?: number; payState?: number }
    if (data.depositedMount != null) {
      depositFen.value = Math.floor(Number(data.depositedMount))
    }
    hasUnpaid.value = Number(data.payState) === 7 || Boolean(temp.unpaidOrderId)
  }
  if (cfg.success && cfg.data) {
    izWithdraw.value = Boolean((cfg.data as { izWithdraw?: boolean }).izWithdraw)
  }
  if (record.success && record.data) {
    const list = (record.data as { list?: Array<{ state?: number }> }).list || []
    hasWithdrawRecord.value = list.length > 0 && Number(list[0]?.state) !== 4
  }
  await loadPayScoreConfig(String(sid || ''))
})

function go(url: string) {
  navigate('to', url)
}

function goRecharge() {
  navigate('to', '/pages-sub/pay/recharge/recharge')
}

function onWithdrawEntry() {
  if (!izWithdraw.value) return
  agreed.value = false
  tipsVisible.value = true
}

function goWithdrawProgress() {
  navigate('to', '/pages-sub/pay/withdraw/withdraw?isProcessing=true')
}

function closeTips() {
  tipsVisible.value = false
  agreed.value = false
}

function toggleAgree() {
  agreed.value = !agreed.value
}

function confirmWithdraw() {
  if (hasUnpaid.value) {
    closeTips()
    return
  }
  if (!agreed.value) {
    uni.showToast({ title: t('account.withdrawNeedAgree'), icon: 'none' })
    return
  }
  closeTips()
  navigate('to', '/pages-sub/pay/withdraw/withdraw')
}

function goDepositRefund() {
  const pin = String(user.userInfo?.pin || '')
  navigate(
    'to',
    `/pages-sub/pay/deposit-refund/deposit-refund?depositedMount=${depositFen.value}&userPin=${encodeURIComponent(pin)}`,
  )
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #f7f8fa;
  position: relative;
  box-sizing: border-box;
}
.page-bg {
  position: absolute;
  left: 0;
  top: 0;
  width: 100vw;
  height: 600rpx;
  z-index: 0;
}
.scroll {
  position: relative;
  z-index: 1;
  height: 100%;
  box-sizing: border-box;
  padding-top: 24rpx;
}
.wallet-card {
  position: relative;
  margin: 0 32rpx;
  width: 686rpx;
  height: 324rpx;
  box-sizing: border-box;
}
.wallet-card__bg {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  border-radius: 24rpx;
}
.wallet-card__content {
  position: absolute;
  top: 44rpx;
  left: 50rpx;
  right: 50rpx;
  z-index: 1;
}
.wallet-card__label {
  font-size: 24rpx;
  color: #fff;
  line-height: 24rpx;
  height: 24rpx;
}
.wallet-card__main {
  margin-top: 36rpx;
  display: flex;
  align-items: center;
}
.wallet-card__balance {
  font-size: 64rpx;
  font-weight: 700;
  color: #fff;
  line-height: 36rpx;
  height: 36rpx;
}
.wallet-card__detail {
  margin-left: 16rpx;
  width: 36rpx;
  height: 36rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.wallet-card__detail-icon {
  width: 24rpx;
  height: 24rpx;
}
.wallet-card__line {
  margin-top: 40rpx;
  margin-left: 14rpx;
  width: 558rpx;
  height: 2rpx;
  background: rgba(255, 255, 255, 0.3);
}
.wallet-card__parts {
  margin-top: 20rpx;
  display: flex;
  align-items: center;
  width: 100%;
}
.wallet-card__part {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.wallet-card__part:last-child {
  margin-left: 48rpx;
}
.wallet-card__part-label {
  font-size: 20rpx;
  font-weight: 400;
  color: #fff;
  line-height: 24rpx;
  height: 24rpx;
}
.wallet-card__part-val {
  margin-top: 12rpx;
  font-size: 44rpx;
  font-weight: 700;
  color: #fff;
  line-height: 36rpx;
  height: 36rpx;
}
.wallet-card__vline {
  margin-top: 4rpx;
  width: 2rpx;
  height: 72rpx;
  background: rgba(255, 255, 255, 0.3);
  flex-shrink: 0;
}
.actions {
  margin-top: 46rpx;
  padding: 0 32rpx;
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}
.action {
  width: 686rpx;
  height: 96rpx;
  border-radius: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}
.action--primary {
  background: #4ca1fe;
}
.action--light {
  background: #dbecff;
}
.action__icon {
  width: 40rpx;
  height: 40rpx;
  margin-right: 8rpx;
}
.action__text {
  font-weight: 500;
}
.action__text--white {
  color: #fff;
  font-size: 32rpx;
}
.action__text--blue {
  color: #1e78ff;
  font-size: 28rpx;
}
.more {
  margin: 32rpx;
  padding: 32rpx 32rpx 48rpx;
  background: #fff;
  border-radius: 32rpx;
}
.more__title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.more__item {
  margin-top: 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 154rpx;
}
.more__icon {
  width: 80rpx;
  height: 80rpx;
}
.more__name {
  margin-top: 16rpx;
  font-size: 28rpx;
  color: #666;
}
.more__des {
  font-size: 24rpx;
  color: #999;
}
.mask {
  position: fixed;
  left: 0;
  top: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.tips-alert {
  width: 590rpx;
  background: #fff;
  border-radius: 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-bottom: 32rpx;
  box-sizing: border-box;
}
.tips-alert__title {
  margin-top: 34rpx;
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  line-height: 44rpx;
}
.tips-alert__body {
  padding: 32rpx 48rpx 0;
  width: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}
.tips-alert__line {
  font-size: 28rpx;
  color: #666;
  line-height: 40rpx;
  margin-bottom: 40rpx;
}
.tips-alert__line:last-child {
  margin-bottom: 0;
}
.tips-alert__agree {
  margin-top: 18rpx;
  width: 100%;
  display: flex;
  align-items: center;
  padding: 16rpx 32rpx;
  box-sizing: border-box;
}
.tips-alert__hit {
  width: 48rpx;
  height: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.tips-alert__check {
  width: 32rpx;
  height: 32rpx;
}
.tips-alert__box {
  width: 32rpx;
  height: 32rpx;
  border: 2rpx solid #ccc;
  border-radius: 6rpx;
  box-sizing: border-box;
}
.tips-alert__box.on {
  background: #4ca1fe;
  border-color: #4ca1fe;
}
.tips-alert__agree-text {
  margin-left: 8rpx;
  font-size: 24rpx;
  color: #999;
}
.tips-alert__btns {
  margin-top: 32rpx;
  display: flex;
  flex-direction: row;
  align-items: center;
}
.tips-alert__cancel {
  width: 238rpx;
  height: 80rpx;
  border-radius: 40rpx;
  border: 2rpx solid #ccc;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 500;
  color: #333;
  box-sizing: border-box;
}
.tips-alert__confirm {
  margin-left: 16rpx;
  width: 240rpx;
  height: 80rpx;
  background: #4ca1fe;
  border-radius: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 500;
  color: #fff;
}
.tips-alert__confirm--single {
  margin-left: 0;
  width: 494rpx;
}
</style>
