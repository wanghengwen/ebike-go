<template>
  <view class="home">
    <view class="nav-bar" :style="{ marginTop: statusBarPx + 'px' }">
      <view
        class="me"
        :class="{ 'default-me': !userAvatar }"
        @click="goProfile"
      >
        <image
          v-if="userAvatar"
          class="me__img"
          :src="userAvatar"
          mode="aspectFill"
        />
        <image
          v-else-if="defaultAvatar"
          class="me__img me__img--default"
          :src="defaultAvatar"
          mode="aspectFit"
        />
      </view>
      <text class="brand">{{ brandName }}</text>
    </view>

    <scroll-view
      class="home-scroll"
      scroll-y
      :style="{ marginBottom: scanBarHeight }"
    >
      <view class="map-container" :style="{ height: mapHeight }">
        <map
          class="map"
          id="homeMap"
          show-location
          :provider="mapProvider()"
          :latitude="latitude"
          :longitude="longitude"
          :scale="17"
          :markers="markers"
          :polygons="polygons"
          :polyline="polyline"
          :enable-scroll="false"
          :enable-zoom="false"
          :enable-rotate="false"
          :enable-overlooking="false"
          @tap="goMap"
          @markertap="goMap"
        />
      </view>

      <view
        v-if="creditTips.length || scrollerMsg.length"
        class="scroll-tips"
        :class="{ notice: isLimitRiding }"
        :style="scrollTipsStyle"
      >
        <swiper
          class="scroll-tips__swiper"
          vertical
          autoplay
          circular
          :interval="3000"
          :duration="500"
        >
          <swiper-item v-for="(item, idx) in creditTips" :key="'c' + idx">
            <view class="scroll-tips__row" @click="goCredit">
              <image
                v-if="notifyIcon"
                class="scroll-tips__icon"
                :src="isLimitRiding ? notifyIconWhite || notifyIcon : notifyIcon"
                mode="widthFix"
              />
              <text class="scroll-tips__text">{{ item.title }}</text>
              <image
                v-if="arrowIcon"
                class="scroll-tips__arrow"
                :src="isLimitRiding ? arrowIconWhite || arrowIcon : arrowIcon"
                mode="widthFix"
              />
            </view>
          </swiper-item>
          <swiper-item v-for="(item, idx) in scrollerMsg" :key="'s' + idx">
            <view class="scroll-tips__row" @click="onNotice(item)">
              <image v-if="notifyIcon" class="scroll-tips__icon" :src="notifyIcon" mode="widthFix" />
              <text class="scroll-tips__text">{{ item.content || item.title || '' }}</text>
              <image v-if="arrowIcon" class="scroll-tips__arrow" :src="arrowIcon" mode="widthFix" />
            </view>
          </swiper-item>
        </swiper>
      </view>

      <view v-if="shortcutList.length" class="shortcut-section">
        <view class="shortcut-wrapper">
          <view
            v-for="(item, index) in shortcutList"
            :key="index"
            class="shortcut"
            @click="onShortcut(item)"
          >
            <image class="shortcut-image" :src="item.icon" mode="aspectFit" />
            <text class="shortcut-text">{{ item.name }}</text>
          </view>
        </view>
      </view>
    </scroll-view>

    <view class="scanBtn-wrapper" :style="{ height: scanBarHeight }">
      <view
        class="scanBtn"
        :class="{ 'is-disabled': isLimitRiding }"
        :style="scanBtnStyle"
        @click="onScan"
      >
        <image v-if="scanIcon" class="scanBtn__icon" :src="scanIcon" mode="widthFix" />
        <text>{{ t('ride.useBike') }}</text>
      </view>
    </view>

    <BizPopup
      :visible="showUnpaid"
      :title="t('pay.unpaidTip')"
      :confirm-text="t('pay.goPay')"
      :cancel-text="t('common.cancel')"
      @cancel="showUnpaid = false"
      @close="showUnpaid = false"
      @confirm="goPay"
    >
      <text>{{ t('pay.unpaidTip') }}</text>
    </BizPopup>

    <OpsPopup :visible="showPopup" :config="currPopup" @close="onPopupClose" />
    <GuideSheet :visible="showGuide" :config="guideConfig" @close="onGuideClose" />
    <PrivacyAuthorizePopup />
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getHomeNav } from '@/api/map'
import { getPersonInfo } from '@/api/user'
import { useHomeMap } from '@/features/map/useHomeMap'
import { useScanGate } from '@/features/bike/useScanGate'
import { useCreditLimit } from '@/features/credit/useCreditLimit'
import { useGuidePopup } from '@/features/guide/useGuidePopup'
import { openUserClickAction, openScrollerNotice } from '@/shared/openNotice'
import { useUserStore } from '@/stores/user'
import { useTempDataStore } from '@/stores/tempData'
import {
  getBrandColor,
  getButtonDisabledColor,
  getTenantConfig,
} from '@/shared/config'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { mapProvider } from '@/shared/mapProvider'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'
import BizPopup from '@/widgets/BizPopup.vue'
import OpsPopup from '@/widgets/OpsPopup.vue'
import GuideSheet from '@/widgets/GuideSheet.vue'
import PrivacyAuthorizePopup from '@/widgets/PrivacyAuthorizePopup.vue'

type HomeShortcut = {
  name?: string
  icon?: string
  jumpPage?: {
    linkUrl?: string
    appId?: string
    param?: string
    linkTitle?: string
    chainType?: number
  }
}

const { t } = useI18n()
const user = useUserStore()
const temp = useTempDataStore()
const {
  latitude,
  longitude,
  markers,
  polygons,
  polyline,
  scrollerMsg,
  refreshMap,
} = useHomeMap()
const { scanThenPrecycling } = useScanGate()
const { isLimitRiding, tipList: creditTips, refresh: refreshCredit } = useCreditLimit()
const {
  showPopup,
  showGuide,
  currPopup,
  guideConfig,
  initHomeOrMap,
  onPopupClose,
  onGuideClose,
} = useGuidePopup()

const showUnpaid = ref(false)
const statusBarPx = ref(20)
const safeBottomPx = ref(0)
/** Legacy mapHeight: windowHeight - 450；加大地图后改为减 380，快捷入口随之下移 */
const mapHeight = ref('300px')
const shortcutList = ref<HomeShortcut[]>([])

const brandName = computed(() => getTenantConfig().name || t('brand.name'))

/** Legacy scanButtonStyle: brand bg + #1E4A38 text */
const scanBtnStyle = computed(() => {
  const activeBg = getBrandColor()
  const disabledBg = getButtonDisabledColor()
  const isDisabled = isLimitRiding.value
  return {
    backgroundColor: isDisabled ? disabledBg : activeBg,
    borderColor: isDisabled ? disabledBg : activeBg,
    color: isDisabled ? '#ffffff' : '#1E4A38',
  }
})

const scrollTipsStyle = computed(() => {
  if (isLimitRiding.value) return {}
  const brand = getBrandColor() || '#AEC8A3'
  return {
    background: `linear-gradient(90deg, ${brand} 0%, #F0F5EE 100%)`,
  }
})

/** Legacy scanBtn-wrapper: 212rpx with safe area, else 160rpx */
const scanBarHeight = computed(() => (safeBottomPx.value > 0 ? '212rpx' : '160rpx'))

const userAvatar = computed(() =>
  String(user.userInfo?.avatar || user.userInfo?.avatarUrl || ''),
)
const defaultAvatar = computed(() => getIconCfg('default_avatar') || '')

const notifyIcon = computed(() => getMapCfg('notify'))
const notifyIconWhite = computed(() => getMapCfg('notifyWhite'))
const arrowIcon = computed(() => getMapCfg('iconRight'))
const arrowIconWhite = computed(() => getMapCfg('iconRightWhite'))
const scanIcon = computed(() => getMapCfg('iconScan'))

function calcMapHeight() {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarPx.value = sys.statusBarHeight || 20
    safeBottomPx.value = sys.safeAreaInsets?.bottom || 0
    const h = (sys.windowHeight || 0) - 380
    mapHeight.value = `${h > 300 ? h : 300}px`
  } catch {
    statusBarPx.value = 20
    safeBottomPx.value = 0
    mapHeight.value = '300px'
  }
}

onMounted(() => {
  calcMapHeight()
})

onShow(() => {
  calcMapHeight()
  setNavTitle(brandName.value)
  user.hydrateFromStorage()
  void refreshMap({ keepSelection: true }).then(() => {
    if (user.isLoggedIn) void initHomeOrMap()
  })
  void loadHomeNav()
  if (temp.unpaidOrderId) showUnpaid.value = true
  if (user.isLoggedIn) {
    void refreshCredit(true)
    void getPersonInfo()
      .then((res) => {
        if (!res.success || !res.data) return
        const profile = res.data as { ridingState?: number; payState?: number }
        user.setUserInfo(res.data as never)
        const ridingState = Number(profile.ridingState)
        if ([4, 5, 6].includes(ridingState)) {
          navigate('reLaunch', '/pages/riding/riding')
          return
        }
        if (Number(profile.payState) === 7) showUnpaid.value = true
      })
      .catch((e) => logger.warn('home getPersonInfo soft fail', e))
  }
})

async function loadHomeNav() {
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getHomeNav({
      carType: 0,
      ...(sid ? { serviceId: sid } : {}),
    })
    if (res.success && Array.isArray(res.data)) {
      shortcutList.value = res.data as HomeShortcut[]
    }
  } catch (e) {
    logger.warn('getHomeNav soft fail', e)
  }
}

function onShortcut(item: HomeShortcut) {
  const jump = item.jumpPage || {}
  openUserClickAction(Number(jump.chainType ?? 0), {
    linkUrl: jump.linkUrl,
    appId: jump.appId,
    param: jump.param,
    linkTitle: jump.linkTitle || item.name,
  })
}

function goProfile() {
  navigate('to', '/pages/account/profile')
}
function goCredit() {
  navigate('to', '/pages-sub/account/credit/credit')
}
function goMap() {
  navigate('to', '/pages/map/map')
}
function onNotice(item: (typeof scrollerMsg.value)[number]) {
  openScrollerNotice(item as never)
}
function goPay() {
  showUnpaid.value = false
  const oid = temp.unpaidOrderId
  const qs = ['isEBikeLock=true']
  if (oid) qs.push(`orderId=${encodeURIComponent(oid)}`)
  navigate('reLaunch', `/pages/pay/pay?${qs.join('&')}`)
}
async function onScan() {
  if (isLimitRiding.value) return
  await scanThenPrecycling(() => {
    showUnpaid.value = true
  })
}
</script>

<style scoped lang="scss">
.home {
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  overflow: hidden;
  box-sizing: border-box;
}
.nav-bar {
  width: 100%;
  height: 60px;
  background: #ffffff;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  box-sizing: border-box;
}
.me {
  position: absolute;
  left: 20rpx;
  top: 50%;
  transform: translateY(-50%);
  z-index: 11;
  width: 52px;
  height: 52px;
  border-radius: 50%;
  overflow: hidden;
  box-shadow: 0 8rpx 16rpx rgba(0, 0, 0, 0.19);
  background: #ffffff;
}
.me.default-me {
  display: flex;
  align-items: center;
  justify-content: center;
}
.me__img {
  width: 100%;
  height: 100%;
}
.me__img--default {
  width: 55%;
  height: 55%;
}
.brand {
  font-size: 18px;
  font-weight: 600;
  color: #333333;
}
.home-scroll {
  flex: 1;
  min-height: 0;
  width: 100%;
  background: #f4f5f7;
  box-sizing: border-box;
}
.map-container {
  background: #f4f5f7;
  border: 5rpx solid #ffffff;
  margin: 20rpx;
  border-radius: 30rpx;
  overflow: hidden;
  box-sizing: border-box;
}
.map {
  width: 100%;
  height: 100%;
}
.scroll-tips {
  width: 100%;
  z-index: 11;
  padding: 0 30rpx;
  box-sizing: border-box;
  height: 80rpx;
  background: linear-gradient(90deg, #aec8a3 0%, #f0f5ee 100%);
}
.scroll-tips.notice {
  background: #ff5936;
}
.scroll-tips__swiper {
  width: 100%;
  height: 80rpx;
}
.scroll-tips__row {
  height: 80rpx;
  display: flex;
  align-items: center;
}
.scroll-tips__icon {
  width: 48rpx;
  height: 48rpx;
  margin-right: 16rpx;
  flex-shrink: 0;
}
.scroll-tips__text {
  flex: 1;
  font-size: 28rpx;
  font-weight: 400;
  color: #333;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.scroll-tips.notice .scroll-tips__text {
  color: #fff;
}
.scroll-tips__arrow {
  width: 30rpx;
  height: 30rpx;
  flex-shrink: 0;
  margin-left: 8rpx;
}
.shortcut-section {
  width: 100%;
  background: #ffffff;
  padding-top: 24rpx;
  padding-bottom: 16rpx;
  box-sizing: border-box;
}
.shortcut-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 0 56rpx;
  height: 204rpx;
  box-sizing: border-box;
}
.shortcut {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  flex: 1;
}
.shortcut-image {
  width: 106rpx;
  height: 106rpx;
  margin-top: 32rpx;
  margin-bottom: 16rpx;
}
.shortcut-text {
  height: 34rpx;
  line-height: 34rpx;
  font-size: 24rpx;
  font-weight: 400;
  color: #666666;
  margin-bottom: 16rpx;
}
.scanBtn-wrapper {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  width: 100%;
  background: #ffffff;
  border-top-left-radius: 32rpx;
  border-top-right-radius: 32rpx;
  display: flex;
  flex-direction: column;
  z-index: 13;
  box-sizing: border-box;
}
.scanBtn {
  height: 96rpx;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 34rpx;
  margin-top: 32rpx;
  margin-left: 24rpx;
  margin-right: 24rpx;
  border-radius: 48rpx;
  box-sizing: border-box;
}
.scanBtn__icon {
  width: 40rpx;
  height: 40rpx;
  margin-right: 16rpx;
}
.scanBtn.is-disabled {
  pointer-events: none;
}
</style>
