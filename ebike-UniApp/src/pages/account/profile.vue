<template>
  <view class="pages">
    <image
      v-if="navBackIcon"
      class="nav-back"
      :src="navBackIcon"
      :style="{ top: `${navBackPos}px` }"
      @click="navBack"
    />

    <view class="page" :style="pageBgStyle">
      <scroll-view scroll-y class="scroll" style="height: 100%">
        <view class="top_container">
          <view class="info">
            <view v-if="user.isLoggedIn" class="name_container">
              <view
                class="userinfo-avatar"
                :class="{ 'default-avatar-wrap': !avatarUrl }"
                @click.stop="changeAvatar"
              >
                <image
                  v-if="avatarUrl"
                  class="avatar_img"
                  :src="avatarUrl"
                  mode="aspectFit"
                />
                <image
                  v-else
                  class="avatar_img_default"
                  :src="defaultAvatar"
                  mode="aspectFit"
                />
              </view>
              <view class="userinfo-username">
                <view class="name-view">
                  <!-- #ifdef MP-WEIXIN -->
                  <open-data type="userNickName" />
                  <!-- #endif -->
                  <!-- #ifndef MP-WEIXIN -->
                  <text>{{ displayName }}</text>
                  <!-- #endif -->
                  <view
                    v-if="needAuthEntry"
                    class="verify"
                    :style="verifyBgStyle"
                    @click.stop="goVerified"
                  >
                    {{ realNameText }}
                  </view>
                </view>
                <view class="phone">{{ maskedUserPhone }}</view>
              </view>
            </view>
            <view v-else class="name_container" @click="goLogin">
              <view class="userinfo-avatar default-avatar-wrap">
                <image class="avatar_img_default" :src="defaultAvatar" mode="aspectFit" />
              </view>
              <view class="userinfo-username">
                <view class="unlogin-view">
                  <text class="login_text">{{ t('auth.loginNow') }}</text>
                  <image
                    v-if="arrowRound"
                    class="right_arrow"
                    :src="arrowRound"
                    mode="aspectFit"
                  />
                </view>
              </view>
            </view>
          </view>

          <view class="top-list_container">
            <view class="size-box">
              <view class="route" @click="goGuarded('/pages-sub/account/orders/orders')">
                <image v-if="iconTrips" class="type-imgae" :src="iconTrips" mode="aspectFit" />
                <text class="type-name">{{ t('account.quickTrips') }}</text>
              </view>
              <view class="route" @click="goGuarded('/pages-sub/support/messages/list')">
                <image v-if="iconMsg" class="type-imgae" :src="iconMsg" mode="aspectFit" />
                <text class="type-name">{{ t('account.quickMessages') }}</text>
                <view v-if="unread > 0" class="bg_wrapper">
                  <text class="text_tips">{{ unread > 99 ? '99+' : unread }}</text>
                </view>
              </view>
              <view class="route" @click="goGuarded('/pages-sub/support/help/help')">
                <image v-if="iconCs" class="type-imgae" :src="iconCs" mode="aspectFit" />
                <text class="type-name">{{ t('account.quickService') }}</text>
              </view>
              <view class="route" @click="go('/pages/account/settings')">
                <image v-if="iconSet" class="type-imgae" :src="iconSet" mode="aspectFit" />
                <text class="type-name">{{ t('account.quickSettings') }}</text>
              </view>
            </view>
          </view>
        </view>

        <view class="my-property">
          <text class="my-property-title">{{ t('account.myAssets') }}</text>
          <view class="card_container">
            <view class="card_item" @click="goCards">
              <view class="card_number">
                <text class="card_number_value">{{ ridingCardCount }}</text>
                <text class="card_number_unit">{{ t('account.ridingCardUnit') }}</text>
              </view>
              <view class="card_name">
                <text class="card_name_text">{{ t('account.ridingCardLabel') }}</text>
                <image v-if="iconRight" class="icon_right" :src="iconRight" mode="aspectFit" />
              </view>
            </view>
          </view>

          <view class="my_wallet" :style="walletBgStyle" @click="goWallet">
            <view class="wallet_box">
              <text class="wallet_title">{{ t('pay.wallet') }}</text>
              <view class="wallet_number_content">
                <text class="money_number_unit">¥</text>
                <text class="money_number_value">{{ balanceText }}</text>
              </view>
            </view>
          </view>
        </view>

        <view class="other_function">
          <view class="text-wrapper_13">
            <text class="text_18">{{ t('account.otherFunctions') }}</text>
          </view>
          <view class="other_function_container">
            <view class="other_item" @click="onQualClick">
              <image v-if="iconQual" class="other_item_image" :src="iconQual" mode="aspectFit" />
              <view class="other_item_sub">
                <text class="main_title">{{ t('account.qualification') }}</text>
                <text class="sub_title" :class="qualGot ? 'sub_title_0' : 'sub_title_1'">
                  {{ qualGot ? t('account.qualGot') : t('account.qualTodo') }}
                </text>
              </view>
            </view>
            <view class="other_item" @click="goGuarded('/pages-sub/account/voucher/voucher')">
              <image
                v-if="iconVoucher"
                class="other_item_image"
                :src="iconVoucher"
                mode="aspectFit"
              />
              <text class="other_item_text">{{ t('account.voucher') }}</text>
            </view>
            <view
              class="other_item"
              :style="{ visibility: izOpenInvoice ? 'visible' : 'hidden' }"
              @click="onInvoiceClick"
            >
              <image
                v-if="iconInvoice"
                class="other_item_image"
                :src="iconInvoice"
                mode="aspectFit"
              />
              <text class="other_item_text">{{ t('pay.invoice') }}</text>
            </view>
          </view>
        </view>
      </scroll-view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getPersonInfo, getConfigBaseItem, updateAvatar } from '@/api/user'
import { userEnableConfig } from '@/api/account'
import { getUnreadCount } from '@/api/message'
import { getUserRidingCard } from '@/api/card'
import { uploadFile } from '@/shared/upload'
import { useUserStore } from '@/stores/user'
import { checkVerifyAndGo } from '@/features/auth/checkVerifyAndGo'
import { ensureLocationAuthorized } from '@/features/map/ensureLocationAuth'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
import { getTenantConfig } from '@/shared/config'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import { phoneDesensitize } from '@/shared/phone'

const { t } = useI18n()
const user = useUserStore()
const unread = ref(0)
const ridingCardCount = ref(0)
const izOpenInvoice = ref(false)
const careerEnable = ref(false)
const navBackPos = ref(40)

const defaultAvatar = computed(() => getIconCfg('default_avatar') || '')
const avatarUrl = computed(() => {
  const info = user.userInfo as Record<string, unknown>
  return String(info.avatar || info.icon || info.headImg || '')
})
const headerBg = computed(() => getIconCfg('walletHeaderBg'))
const pageBgStyle = computed(() => {
  if (headerBg.value) {
    return {
      backgroundImage: `url(${headerBg.value})`,
      backgroundSize: '100% 750rpx',
      backgroundRepeat: 'no-repeat',
      backgroundColor: '#F6F7F0',
    }
  }
  return { backgroundColor: '#F6F7F0' }
})
const walletBg = computed(() => getIconCfg('newUserInfoWalletBg'))
const walletBgStyle = computed(() =>
  walletBg.value
    ? {
        backgroundImage: `url(${walletBg.value})`,
        backgroundRepeat: 'no-repeat',
        backgroundSize: '100% 100%',
        backgroundPosition: 'center',
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

const maskedUserPhone = computed(() => phoneDesensitize(user.userInfo.phone))

/** Legacy isneedAuth：配置开启实名时才展示徽章 */
const needAuthEntry = computed(() => {
  const cfg = getTenantConfig().customSetting as { isneedAuth?: boolean; needAuth?: boolean } | undefined
  if (cfg?.isneedAuth === false || cfg?.needAuth === false) return false
  return true
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

/**
 * Legacy userInfo.checkUserState(izRidingType):
 * 仅 7（及未返回）显示「未获取」，其余均为「已获取」。
 */
const qualGot = computed(() => {
  const info = user.userInfo as Record<string, unknown>
  const state = info.izRidingType
  if (state == null || state === '') return false
  return Number(state) !== 7
})

onMounted(() => {
  try {
    const menu = uni.getMenuButtonBoundingClientRect?.()
    if (menu?.top != null && menu?.height != null) {
      navBackPos.value = menu.top + menu.height / 2
    } else {
      const bar = uni.getSystemInfoSync().statusBarHeight || 20
      navBackPos.value = bar + 22
    }
  } catch {
    navBackPos.value = 40
  }
})

onShow(async () => {
  setNavTitle(t('account.profile'))
  user.hydrateFromStorage()
  if (!user.isLoggedIn) {
    unread.value = 0
    ridingCardCount.value = 0
    izOpenInvoice.value = false
    careerEnable.value = false
    return
  }
  const sid = storage.get<string>('serviceId', '') || ''
  const [profile, msg, cards, baseCfg, enableCfg] = await Promise.all([
    getPersonInfo(),
    getUnreadCount({ msgTypes: [1] }),
    getUserRidingCard(),
    getConfigBaseItem(sid ? { serviceId: sid } : {}).catch((e) => {
      logger.warn('getConfigBaseItem soft fail', e)
      return { success: false, data: null }
    }),
    userEnableConfig(sid ? { serviceId: sid } : {}).catch((e) => {
      logger.warn('userEnableConfig soft fail', e)
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
  if (enableCfg && 'success' in enableCfg && enableCfg.success && enableCfg.data) {
    careerEnable.value = Boolean(
      (enableCfg.data as { careerEnable?: boolean }).careerEnable,
    )
  } else {
    careerEnable.value = false
  }
})

function navBack() {
  navigate('back')
}

function goLogin() {
  navigate('to', '/pages/auth/quick-login')
}

function changeAvatar() {
  if (!user.isLoggedIn) {
    goLogin()
    return
  }
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: async (res) => {
      const path = res.tempFilePaths?.[0]
      const file = (res.tempFiles as { size?: number; path?: string }[] | undefined)?.[0]
      if (!path || !file) return
      if (Number(file.size || 0) >= 10485760) {
        uni.showToast({ title: t('account.avatarTooLarge'), icon: 'none' })
        return
      }
      const url = await uploadFile(path)
      if (!url) {
        uni.showToast({ title: t('account.avatarUploadFail'), icon: 'none' })
        return
      }
      const upd = await updateAvatar({ avatar: url })
      if (!upd.success) {
        uni.showToast({ title: upd.msg || t('account.avatarUpdateFail'), icon: 'none' })
        return
      }
      const profile = await getPersonInfo()
      if (profile.success && profile.data) user.setUserInfo(profile.data as never)
    },
  })
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

async function onQualClick() {
  if (!requireLogin()) return
  const loc = await ensureLocationAuthorized()
  if (!loc.ok) return
  if (careerEnable.value) {
    navigate('to', '/pages-sub/account/career/career')
  }
}

function onInvoiceClick() {
  if (!izOpenInvoice.value) return
  goGuarded('/pages-sub/pay/invoice/list')
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

function goVerified() {
  if (!requireLogin()) return
  const info = user.userInfo as Record<string, unknown>
  if (info.izAuth === true || info.realNameStatus === 1 || info.authState === 3) return
  void checkVerifyAndGo()
}
</script>

<style scoped lang="scss">
.pages {
  width: 100vw;
  height: 100vh;
  position: relative;
}
.nav-back {
  position: absolute;
  width: 48rpx;
  height: 48rpx;
  left: 20rpx;
  transform: translateY(-50%);
  z-index: 999;
}
.page {
  width: 100vw;
  height: 100vh;
  background: #f6f7f0;
}
.scroll {
  height: 100%;
}
.top_container {
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
.name_container {
  position: relative;
  margin-left: 24rpx;
  height: 156rpx;
  display: flex;
  flex-direction: row;
  justify-content: flex-start;
  align-items: center;
}
.userinfo-avatar {
  border-radius: 50%;
  overflow: hidden;
  width: 156rpx;
  height: 156rpx;
  flex-shrink: 0;
  background-color: #ffffff;
  &.default-avatar-wrap {
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
.avatar_img {
  width: 156rpx;
  height: 156rpx;
}
.avatar_img_default {
  width: 55%;
  height: 55%;
}
.userinfo-username {
  display: flex;
  flex-direction: column;
  margin-left: 24rpx;
}
.unlogin-view {
  display: flex;
  align-items: center;
}
.login_text {
  font-size: 32rpx;
  line-height: 44rpx;
  font-weight: 600;
  color: #333333;
}
.right_arrow {
  width: 38rpx;
  height: 32rpx;
  margin-left: 2rpx;
}
.name-view {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  color: rgba(44, 46, 57, 1);
  font-size: 32rpx;
  font-weight: 600;
  text-align: left;
  white-space: nowrap;
  line-height: 50rpx;
  margin-right: 20rpx;
}
.verify {
  width: 108rpx;
  height: 52rpx;
  margin-left: 16rpx;
  background-repeat: no-repeat;
  background-size: cover;
  font-size: 20rpx;
  font-weight: 500;
  color: #ca8b00;
  text-align: center;
  line-height: 50rpx;
}
.phone {
  color: rgba(119, 119, 119, 1);
  font-size: 24rpx;
  font-weight: normal;
  text-align: left;
  white-space: nowrap;
  line-height: 24rpx;
  margin-top: 12rpx;
}
.top-list_container {
  position: relative;
  width: 100%;
  margin-top: 56rpx;
}
.size-box {
  width: 648rpx;
  margin-left: 52rpx;
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-self: center;
}
.route {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.type-imgae {
  width: 108rpx;
  height: 108rpx;
}
.type-name {
  color: rgba(51, 51, 51, 1);
  font-size: 24rpx;
  font-weight: normal;
  white-space: nowrap;
  line-height: 34rpx;
}
.bg_wrapper {
  position: absolute;
  width: 40rpx;
  height: 24rpx;
  background: linear-gradient(90deg, #ff4542 0%, #fe735b 100%);
  box-shadow: 0rpx 4rpx 4rpx 0rpx rgba(255, 76, 70, 0.22);
  border-radius: 11rpx 11rpx 11rpx 0rpx;
  display: flex;
  flex-direction: column;
  left: 64rpx;
  top: -8rpx;
  padding: 0 10rpx 0 12rpx;
  box-sizing: content-box;
  align-items: center;
  justify-content: center;
}
.text_tips {
  color: rgba(255, 255, 255, 1);
  font-size: 16rpx;
  font-weight: 700;
  text-align: center;
  white-space: nowrap;
  line-height: 24rpx;
}
.my-property {
  position: relative;
  background-color: white;
  border-radius: 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-left: 32rpx;
  margin-top: -54rpx;
  margin-right: 32rpx;
  padding-top: 32rpx;
  padding-bottom: 48rpx;
}
.my-property-title {
  color: rgba(51, 51, 51, 1);
  font-size: 28rpx;
  font-weight: 500;
  text-align: left;
  white-space: nowrap;
  line-height: 40rpx;
  align-self: flex-start;
  margin-left: 48rpx;
}
.card_container {
  width: 100%;
  display: flex;
  flex-direction: row;
  justify-content: flex-start;
  margin-top: 40rpx;
}
.card_item {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin-left: 48rpx;
}
.card_number {
  display: flex;
  flex-direction: row;
  justify-content: flex-start;
}
.card_number_value {
  color: rgba(51, 51, 51, 1);
  font-size: 38rpx;
  font-weight: 700;
  text-align: left;
  white-space: nowrap;
  line-height: 36rpx;
}
.card_number_unit {
  color: rgba(51, 51, 51, 1);
  font-size: 20rpx;
  font-weight: normal;
  text-align: left;
  white-space: nowrap;
  line-height: 36rpx;
  margin: 4rpx 0 0 4rpx;
}
.card_name {
  flex-direction: row;
  display: flex;
  justify-content: flex-start;
  margin-top: 16rpx;
}
.card_name_text {
  height: 24rpx;
  font-size: 24rpx;
  font-weight: 400;
  color: #333333;
  line-height: 24rpx;
}
.icon_right {
  margin-left: 2rpx;
  width: 24rpx;
  height: 24rpx;
}
.my_wallet {
  border-radius: 30rpx;
  width: 622rpx;
  height: 168rpx;
  margin-top: 40rpx;
  flex-direction: row;
  display: flex;
  justify-content: flex-start;
  background-repeat: no-repeat;
  background-size: 100% 100%;
  background-position: center;
}
.wallet_box {
  display: flex;
  flex-direction: column;
  margin-left: 192rpx;
  margin-top: 44rpx;
}
.wallet_title {
  color: rgba(51, 51, 51, 1);
  font-size: 24rpx;
  font-weight: normal;
  text-align: left;
  white-space: nowrap;
  line-height: 24rpx;
}
.wallet_number_content {
  margin-top: 20rpx;
  flex-direction: row;
  display: flex;
  justify-content: flex-start;
}
.money_number_unit {
  color: rgba(51, 51, 51, 1);
  font-size: 48rpx;
  font-weight: 700;
  text-align: left;
  white-space: nowrap;
  line-height: 36rpx;
}
.money_number_value {
  margin-left: 12rpx;
  color: rgba(51, 51, 51, 1);
  font-size: 48rpx;
  font-weight: 700;
  text-align: left;
  white-space: nowrap;
  line-height: 36rpx;
}
.other_function {
  position: relative;
  background-color: white;
  border-radius: 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-left: 32rpx;
  margin-top: 32rpx;
  margin-right: 32rpx;
  margin-bottom: 40rpx;
  padding-top: 40rpx;
  padding-bottom: 40rpx;
}
.text-wrapper_13 {
  align-self: flex-start;
  margin-left: 48rpx;
  display: flex;
  flex-direction: row;
}
.text_18 {
  color: rgba(51, 51, 51, 1);
  font-size: 28rpx;
  font-weight: 500;
  text-align: left;
  white-space: nowrap;
  line-height: 40rpx;
}
.other_function_container {
  width: 622rpx;
  margin-top: 40rpx;
  flex-direction: row;
  display: flex;
  justify-content: space-between;
}
.other_item {
  flex-direction: row;
  display: flex;
  justify-content: flex-start;
  align-items: flex-start;
}
.other_item_image {
  width: 40rpx;
  height: 40rpx;
  flex-shrink: 0;
}
.other_item_sub {
  margin-top: 4rpx;
  margin-left: 12rpx;
  display: flex;
  flex-direction: column;
}
.main_title {
  color: rgba(51, 51, 51, 1);
  font-size: 24rpx;
  font-weight: 500;
  text-align: left;
  white-space: nowrap;
  line-height: 34rpx;
}
.sub_title {
  font-size: 20rpx;
  font-weight: normal;
  white-space: nowrap;
  line-height: 28rpx;
  align-self: center;
  margin-top: 2rpx;
}
.sub_title_0 {
  color: #17a3ff;
}
.sub_title_1 {
  color: #ff922b;
}
.other_item_text {
  color: rgba(51, 51, 51, 1);
  font-size: 24rpx;
  font-weight: 500;
  text-align: left;
  white-space: nowrap;
  line-height: 34rpx;
  margin-top: 4rpx;
  margin-left: 12rpx;
}
</style>
