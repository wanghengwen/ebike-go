<template>
  <view class="pages">
    <image
      v-if="navBackIcon"
      class="nav-back"
      :src="navBackIcon"
      :style="{ top: statusBarPx + 12 + 'px' }"
      @click="navBack"
    />

    <scroll-view scroll-y class="page">
      <view class="top" :style="headerBgStyle">
        <view class="info">
          <view v-if="user.isLoggedIn" class="name-box" @click="goVerified">
            <view class="avatar-wrap">
              <image v-if="avatarUrl" class="avatar" :src="avatarUrl" mode="aspectFit" />
              <image v-else class="avatar" :src="defaultAvatar" mode="aspectFit" />
            </view>
            <view class="meta">
              <view class="name-row">
                <text class="name">{{ displayName }}</text>
                <view class="verify" :style="verifyBgStyle">{{ realNameText }}</view>
              </view>
              <view class="phone">{{ maskedUserPhone }}</view>
            </view>
          </view>
          <view v-else class="name-box" @click="goLogin">
            <view class="avatar-wrap">
              <image class="avatar avatar--full" :src="defaultAvatar" mode="aspectFit" />
            </view>
            <view class="meta">
              <view class="login-row">
                <text class="login-text">{{ t('auth.loginNow') }}</text>
                <image v-if="arrowRound" class="login-arrow" :src="arrowRound" mode="aspectFit" />
              </view>
            </view>
          </view>
        </view>

        <view class="quick">
          <view class="quick-item" @click="goGuarded('/pages-sub/account/orders/orders')">
            <image v-if="iconTrips" class="quick-icon" :src="iconTrips" mode="aspectFit" />
            <text class="quick-label">{{ t('account.quickTrips') }}</text>
          </view>
          <view class="quick-item" @click="goGuarded('/pages-sub/support/messages/list')">
            <image v-if="iconMsg" class="quick-icon" :src="iconMsg" mode="aspectFit" />
            <text class="quick-label">{{ t('account.quickMessages') }}</text>
            <view v-if="unread > 0" class="badge">
              <text class="badge-text">{{ unread > 99 ? '99+' : unread }}</text>
            </view>
          </view>
          <view class="quick-item" @click="goGuarded('/pages-sub/support/help/help')">
            <image v-if="iconCs" class="quick-icon" :src="iconCs" mode="aspectFit" />
            <text class="quick-label">{{ t('account.quickService') }}</text>
          </view>
          <view class="quick-item" @click="go('/pages/account/settings')">
            <image v-if="iconSet" class="quick-icon" :src="iconSet" mode="aspectFit" />
            <text class="quick-label">{{ t('account.quickSettings') }}</text>
          </view>
        </view>
      </view>

      <view class="assets">
        <text class="assets-title">{{ t('account.myAssets') }}</text>
        <view class="assets-body" @click="goCards">
          <view class="assets-item">
            <view class="assets-num">
              <text class="assets-value">{{ ridingCardCount }}</text>
              <text class="assets-unit">{{ t('account.ridingCardUnit') }}</text>
            </view>
            <view class="assets-name">
              <text>{{ t('account.ridingCardLabel') }}</text>
              <image v-if="iconRight" class="assets-arrow" :src="iconRight" mode="aspectFit" />
            </view>
          </view>
        </view>

        <view class="wallet" :style="walletBgStyle" @click="goWallet">
          <view class="wallet__info">
            <text class="wallet__title">{{ t('pay.wallet') }}</text>
            <view class="wallet__money">
              <text class="wallet__yen">¥</text>
              <text class="wallet__value">{{ balanceText }}</text>
            </view>
          </view>
          <view class="wallet__btn" @click.stop="goRecharge">{{ t('pay.goRecharge') }}</view>
        </view>
      </view>

      <view class="other">
        <view class="other-title-wrap">
          <text class="other-title">{{ t('account.otherFunctions') }}</text>
        </view>
        <view class="other-list">
          <view class="other-item other-item--qual" @click="goGuarded('/pages-sub/account/qualification/qualification')">
            <image
              v-if="iconQual"
              class="other-icon"
              :src="iconQual"
              mode="aspectFit"
            />
            <view class="other-qual-text">
              <text class="other-text">{{ t('account.qualification') }}</text>
              <text class="other-sub" :class="qualGot ? 'is-ok' : 'is-todo'">
                {{ qualGot ? t('account.qualGot') : t('account.qualTodo') }}
              </text>
            </view>
          </view>
          <view class="other-item" @click="goGuarded('/pages-sub/account/voucher/voucher')">
            <image v-if="iconVoucher" class="other-icon" :src="iconVoucher" mode="aspectFit" />
            <text class="other-text">{{ t('account.voucher') }}</text>
          </view>
          <view
            v-if="izOpenInvoice"
            class="other-item"
            @click="goGuarded('/pages-sub/pay/invoice/list')"
          >
            <image v-if="iconInvoice" class="other-icon" :src="iconInvoice" mode="aspectFit" />
            <text class="other-text">{{ t('pay.invoice') }}</text>
          </view>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getPersonInfo, getConfigBaseItem } from '@/api/user'
import { getUnreadCount } from '@/api/message'
import { getUserRidingCard } from '@/api/card'
import { useUserStore } from '@/stores/user'
import { checkVerifyAndGo } from '@/features/auth/checkVerifyAndGo'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
import { getTenantConfig } from '@/shared/config'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const user = useUserStore()
const unread = ref(0)
const ridingCardCount = ref(0)
const izOpenInvoice = ref(false)
const statusBarPx = ref(20)

const defaultAvatar = computed(() => getIconCfg('default_avatar') || '')
const avatarUrl = computed(() => {
  const info = user.userInfo as Record<string, unknown>
  return String(info.avatar || info.icon || info.headImg || '')
})
const headerBg = computed(() => getIconCfg('walletHeaderBg'))
const headerBgStyle = computed(() =>
  headerBg.value
    ? {
        backgroundImage: `url(${headerBg.value})`,
        backgroundRepeat: 'no-repeat',
        backgroundSize: '100%',
      }
    : { background: 'linear-gradient(180deg, #d7ecff 0%, #f7f8fa 100%)' },
)
const walletBg = computed(() => getIconCfg('newUserInfoWalletBg'))
const walletBgStyle = computed(() =>
  walletBg.value
    ? {
        backgroundImage: `url(${walletBg.value})`,
        backgroundRepeat: 'no-repeat',
        backgroundSize: '100% 100%',
      }
    : { background: 'linear-gradient(90deg, #9fd0ff 0%, #5babf0 100%)' },
)
const verifyBg = computed(() => getIconCfg('vertifyBg'))
const verifyBgStyle = computed(() =>
  verifyBg.value
    ? {
        backgroundImage: `url(${verifyBg.value})`,
        backgroundRepeat: 'no-repeat',
        backgroundSize: 'cover',
      }
    : { background: 'rgba(202, 139, 0, 0.12)' },
)
const navBackIcon = computed(
  () => String(getTenantConfig().customSetting?.navBackIcon || '') || getMapCfg('iconBack') || '',
)
const arrowRound = computed(() => getMapCfg('iconRightRound'))
const iconTrips = computed(() => getIconCfg('newUserInfoRoute'))
const iconMsg = computed(() => getIconCfg('newUserInfoMessage'))
const iconCs = computed(() => getIconCfg('newUserInfoService'))
const iconSet = computed(() => getIconCfg('newUserInfoSetting'))
const iconRight = computed(() => getIconCfg('newUserInfoRight'))
const iconVoucher = computed(() => getIconCfg('newUserInfoActivityCenter'))
const iconInvoice = computed(() => getIconCfg('newUserInfoInvoice'))
const iconQual = computed(() => getIconCfg('newUserInfoQualification'))

const displayName = computed(
  () => user.userInfo.nickName || user.userInfo.phone || t('account.profile'),
)

const maskedUserPhone = computed(() => {
  const p = String(user.userInfo.phone || '').replace(/\D/g, '')
  if (p.length < 7) return user.userInfo.phone || '-'
  return `${p.slice(0, 3)}****${p.slice(-4)}`
})

const realNameText = computed(() => {
  const info = user.userInfo as Record<string, unknown>
  if (info.izAuth === true || info.realNameStatus === 1 || info.authState === 3) {
    return t('account.realNameDone')
  }
  if (Number(info.authState) === 1) return t('account.authReviewing')
  return t('account.realNameTodo')
})

const balanceText = computed(() => {
  const fen = Number((user.userInfo as Record<string, unknown>).balance ?? 0)
  if (!Number.isFinite(fen)) return '0.00'
  return (fen / 100).toFixed(2)
})

/** Legacy depositType: most non-empty types mean 已获取; 0 / null / need-get → 未获取 */
const qualGot = computed(() => {
  const info = user.userInfo as Record<string, unknown>
  const dt = info.depositType
  if (dt == null || dt === '' || Number(dt) === 0) return false
  // type 10 historically "去获取" in some tenants — treat as not obtained
  if (Number(dt) === 10) return false
  return true
})

onMounted(() => {
  try {
    statusBarPx.value = uni.getSystemInfoSync().statusBarHeight || 20
  } catch {
    statusBarPx.value = 20
  }
})

onShow(async () => {
  setNavTitle(t('account.profile'))
  user.hydrateFromStorage()
  if (!user.isLoggedIn) {
    unread.value = 0
    ridingCardCount.value = 0
    izOpenInvoice.value = false
    return
  }
  const sid = storage.get<string>('serviceId', '') || ''
  const [profile, msg, cards, baseCfg] = await Promise.all([
    getPersonInfo(),
    getUnreadCount({ msgTypes: [1] }),
    getUserRidingCard(),
    getConfigBaseItem(sid ? { serviceId: sid } : {}).catch((e) => {
      logger.warn('getConfigBaseItem soft fail', e)
      return { success: false, data: null }
    }),
  ])
  if (profile.success && profile.data) user.setUserInfo(profile.data as never)
  if (msg.success && msg.data != null) {
    const data = msg.data as number | boolean | { count?: number; unread?: number }
    if (typeof data === 'number') unread.value = data
    else if (typeof data === 'boolean') unread.value = data ? 1 : 0
    else unread.value = Number(data.count ?? data.unread ?? 0)
  }
  if (cards.success && cards.data) {
    const data = cards.data as {
      used?: unknown
      records?: unknown
    } | unknown[]
    if (Array.isArray(data)) {
      ridingCardCount.value = data.length
    } else if (data && typeof data === 'object') {
      const used = Array.isArray(data.used) ? data.used : []
      const records = Array.isArray(data.records) ? data.records : []
      ridingCardCount.value = used.length || records.length
    } else {
      ridingCardCount.value = 0
    }
  }
  if (baseCfg && 'success' in baseCfg && baseCfg.success && baseCfg.data) {
    izOpenInvoice.value = Boolean(
      (baseCfg.data as { izOpenInvoice?: boolean }).izOpenInvoice,
    )
  } else {
    izOpenInvoice.value = false
  }
})

function navBack() {
  navigate('back')
}

function goLogin() {
  navigate('to', '/pages/auth/quick-login')
}

function requireLogin(): boolean {
  if (user.isLoggedIn) return true
  uni.showToast({ title: t('account.needLogin'), icon: 'none' })
  setTimeout(() => goLogin(), 400)
  return false
}

function go(url: string) {
  navigate('to', url)
}

function goGuarded(url: string) {
  if (!requireLogin()) return
  navigate('to', url)
}

function goCards() {
  if (!requireLogin()) return
  if (ridingCardCount.value <= 0) {
    navigate('to', '/pages-sub/account/card-shop/shop')
    return
  }
  navigate('to', '/pages-sub/account/cards/cards')
}

function goWallet() {
  if (!requireLogin()) return
  navigate('to', '/pages-sub/pay/wallet/wallet')
}

function goRecharge() {
  if (!requireLogin()) return
  navigate('to', '/pages-sub/pay/recharge/recharge')
}

function goVerified() {
  if (!requireLogin()) return
  void checkVerifyAndGo()
}
</script>

<style scoped lang="scss">
.pages {
  width: 100vw;
  height: 100vh;
  position: relative;
  background: #f7f8fa;
}
.nav-back {
  position: absolute;
  width: 48rpx;
  height: 48rpx;
  left: 20rpx;
  z-index: 999;
}
.page {
  width: 100%;
  height: 100%;
  background: #f7f8fa;
}
.top {
  position: relative;
  width: 100%;
  height: 600rpx;
  display: flex;
  flex-direction: column;
}
.info {
  display: flex;
  padding-top: 172rpx;
  position: relative;
}
.name-box {
  margin-left: 24rpx;
  height: 156rpx;
  display: flex;
  flex-direction: row;
  align-items: center;
}
.avatar-wrap {
  border-radius: 50%;
  overflow: hidden;
  width: 156rpx;
  height: 156rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #e8e8e8;
}
.avatar {
  width: 100rpx;
  height: 100rpx;
}
.avatar--full {
  width: 156rpx;
  height: 156rpx;
}
.meta {
  display: flex;
  flex-direction: column;
  margin-left: 24rpx;
}
.login-row {
  display: flex;
  align-items: center;
}
.login-text {
  font-size: 32rpx;
  line-height: 44rpx;
  font-weight: 600;
  color: #333333;
}
.login-arrow {
  width: 38rpx;
  height: 32rpx;
  margin-left: 2rpx;
}
.name-row {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
}
.name {
  color: #2c2e39;
  font-size: 32rpx;
  font-weight: 600;
  line-height: 50rpx;
  margin-right: 20rpx;
}
.verify {
  width: 108rpx;
  height: 52rpx;
  margin-left: 16rpx;
  font-size: 20rpx;
  font-weight: 500;
  color: #ca8b00;
  text-align: center;
  line-height: 50rpx;
}
.phone {
  color: #777777;
  font-size: 24rpx;
  line-height: 24rpx;
  margin-top: 12rpx;
}
.quick {
  width: 100%;
  margin-top: 12rpx;
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  padding: 0 16rpx;
  box-sizing: border-box;
}
.quick-item {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin: 0 16rpx;
  flex: 1;
}
.quick-icon {
  width: 144rpx;
  height: 144rpx;
}
.quick-label {
  color: #333333;
  font-size: 24rpx;
  line-height: 34rpx;
}
.badge {
  position: absolute;
  min-width: 40rpx;
  height: 24rpx;
  background: linear-gradient(90deg, #ff4542 0%, #fe735b 100%);
  box-shadow: 0 4rpx 4rpx 0 rgba(255, 76, 70, 0.22);
  border-radius: 11rpx 11rpx 11rpx 0;
  left: 80rpx;
  top: -8rpx;
  padding: 0 10rpx 0 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}
.badge-text {
  color: #fff;
  font-size: 16rpx;
  font-weight: 700;
  line-height: 24rpx;
}
.assets {
  position: relative;
  background-color: #fff;
  border-radius: 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin: -54rpx 32rpx 0;
  padding: 32rpx 0 40rpx;
}
.assets-title {
  color: #333333;
  font-size: 28rpx;
  font-weight: 500;
  line-height: 40rpx;
  align-self: flex-start;
  margin-left: 48rpx;
}
.assets-body {
  width: 100%;
  display: flex;
  flex-direction: row;
  margin-top: 40rpx;
}
.assets-item {
  display: flex;
  flex-direction: column;
  margin-left: 48rpx;
}
.assets-num {
  display: flex;
  flex-direction: row;
  align-items: flex-end;
}
.assets-value {
  color: #333333;
  font-size: 38rpx;
  font-weight: 700;
  line-height: 36rpx;
}
.assets-unit {
  color: #333333;
  font-size: 20rpx;
  line-height: 36rpx;
  margin: 4rpx 0 0 4rpx;
}
.assets-name {
  display: flex;
  flex-direction: row;
  align-items: center;
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #333333;
  line-height: 24rpx;
}
.assets-arrow {
  margin-left: 2rpx;
  width: 24rpx;
  height: 24rpx;
}
.wallet {
  width: 622rpx;
  height: 168rpx;
  margin-top: 40rpx;
  border-radius: 30rpx;
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  box-sizing: border-box;
  padding: 0 32rpx 0 192rpx;
}
.wallet__info {
  display: flex;
  flex-direction: column;
}
.wallet__title {
  font-size: 24rpx;
  color: #333;
  line-height: 24rpx;
}
.wallet__money {
  display: flex;
  flex-direction: row;
  align-items: flex-end;
  margin-top: 16rpx;
}
.wallet__yen {
  font-size: 28rpx;
  font-weight: 700;
  color: #333;
  line-height: 36rpx;
}
.wallet__value {
  margin-left: 8rpx;
  font-size: 48rpx;
  font-weight: 700;
  color: #333;
  line-height: 48rpx;
}
.wallet__btn {
  min-width: 144rpx;
  height: 56rpx;
  padding: 0 24rpx;
  border-radius: 28rpx;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.72) 0%, #fff 100%);
  color: #17a3ff;
  font-size: 28rpx;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}
.other {
  position: relative;
  background-color: #fff;
  border-radius: 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin: 32rpx 32rpx 40rpx;
  padding: 40rpx 0;
}
.other-title-wrap {
  align-self: flex-start;
  margin-left: 48rpx;
}
.other-title {
  color: #333333;
  font-size: 28rpx;
  font-weight: 500;
  line-height: 40rpx;
}
.other-list {
  width: 622rpx;
  margin-top: 40rpx;
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: 40rpx 48rpx;
}
.other-item {
  display: flex;
  flex-direction: row;
  align-items: center;
}
.other-item--qual {
  align-items: flex-start;
}
.other-icon {
  width: 40rpx;
  height: 40rpx;
  flex-shrink: 0;
}
.other-text {
  color: #333333;
  font-size: 24rpx;
  font-weight: 500;
  line-height: 34rpx;
  margin-left: 12rpx;
}
.other-qual-text {
  display: flex;
  flex-direction: column;
  margin-left: 12rpx;
}
.other-qual-text .other-text {
  margin-left: 0;
}
.other-sub {
  font-size: 20rpx;
  line-height: 28rpx;
  margin-top: 2rpx;
}
.other-sub.is-ok {
  color: #17a3ff;
}
.other-sub.is-todo {
  color: #ff922b;
}
</style>
