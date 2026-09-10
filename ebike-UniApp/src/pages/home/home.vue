<template>
  <view class="home">
    <view class="nav-bar" :style="{ paddingTop: statusBarPx + 'px' }">
      <view class="me" @click="goProfile">
        <image v-if="avatarUrl" class="me__img" :src="avatarUrl" mode="aspectFill" />
        <view v-else class="me__fallback">
          <text class="me__text">{{ avatarText }}</text>
        </view>
      </view>
      <text class="brand">{{ brandName }}</text>
    </view>

    <view class="map-wrap" @click="goMap">
      <map
        class="map"
        id="homeMap"
        show-location
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
    >
      <swiper class="scroll-tips__swiper" vertical autoplay circular :interval="3000" :duration="500">
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

    <!-- 底部：动态快捷入口 + 立即用车（贴底，避免中间大块留白） -->
    <view class="bottom-panel" :style="{ paddingBottom: safeBottomPx + 'px' }">
      <view v-if="shortcutList.length" class="shortcut-wrapper">
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

      <view
        class="dock__btn"
        :class="{ 'is-disabled': isLimitRiding }"
        :style="isLimitRiding ? disabledBtnStyle : brandBtnStyle"
        @click="onScan"
      >
        <image v-if="scanIcon" class="dock__scan-icon" :src="scanIcon" mode="widthFix" />
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
import BizPopup from '@/widgets/BizPopup.vue'
import OpsPopup from '@/widgets/OpsPopup.vue'
import GuideSheet from '@/widgets/GuideSheet.vue'
import PrivacyAuthorizePopup from '@/widgets/PrivacyAuthorizePopup.vue'
import { useHomeMap } from '@/features/map/useHomeMap'
import { useScanGate, consumeAutoScanUseBike } from '@/features/bike/useScanGate'
import { useCreditLimit } from '@/features/credit/useCreditLimit'
import { useGuidePopup } from '@/features/guide/useGuidePopup'
import { useUserStore } from '@/stores/user'
import { useTempDataStore } from '@/stores/tempData'
import { getPersonInfo } from '@/api/user'
import { getHomeNav } from '@/api/map'
import { getBrandColor, getButtonDisabledColor, getButtonWhiteColor, getTenantConfig } from '@/shared/config'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { openScrollerNotice, openUserClickAction } from '@/shared/openNotice'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

type HomeShortcut = {
  name?: string
  icon?: string
  jumpPage?: {
    linkUrl?: string
    appId?: string
    param?: unknown
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
const shortcutList = ref<HomeShortcut[]>([])

const brandName = computed(() => getTenantConfig().name || t('brand.name'))
const brandBtnStyle = computed(() => {
  const c = getBrandColor()
  return { backgroundColor: c, borderColor: c, color: getButtonWhiteColor() }
})
const disabledBtnStyle = computed(() => {
  const c = getButtonDisabledColor()
  return { backgroundColor: c, borderColor: c, color: getButtonWhiteColor() }
})

const avatarUrl = computed(() => {
  const u = String(user.userInfo?.avatar || user.userInfo?.avatarUrl || '')
  return u || getIconCfg('default_avatar') || ''
})
const avatarText = computed(() => {
  const name = String(user.userInfo?.nickName || user.userInfo?.phone || t('account.profile'))
  return name.slice(0, 1)
})

const notifyIcon = computed(() => getMapCfg('notify'))
const notifyIconWhite = computed(() => getMapCfg('notifyWhite'))
const arrowIcon = computed(() => getMapCfg('iconRight'))
const arrowIconWhite = computed(() => getMapCfg('iconRightWhite'))
const scanIcon = computed(() => getMapCfg('iconScan'))

onShow(() => {
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
      .catch((e) => logger.warn('home personInfo soft fail', e))
  }
  if (consumeAutoScanUseBike()) {
    setTimeout(() => {
      void scanThenPrecycling(() => {
        showUnpaid.value = true
      })
    }, 400)
  }
})

onMounted(() => {
  try {
    const info = uni.getSystemInfoSync()
    statusBarPx.value = info.statusBarHeight || 20
    safeBottomPx.value = Number(
      (info.safeAreaInsets as { bottom?: number } | undefined)?.bottom || 0,
    )
  } catch {
    statusBarPx.value = 20
  }
  void refreshMap()
  void loadHomeNav()
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
  background: #fff;
  overflow: hidden;
  box-sizing: border-box;
}
.nav-bar {
  width: 100%;
  min-height: 60px;
  background: #fff;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  box-sizing: border-box;
  padding-bottom: 12rpx;
}
.me {
  position: absolute;
  left: 20rpx;
  bottom: 10rpx;
  z-index: 11;
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  overflow: hidden;
  box-shadow: 0 8rpx 16rpx rgba(0, 0, 0, 0.19);
  background: var(--brand-color, #3aa0e8);
}
.me__img,
.me__fallback {
  width: 100%;
  height: 100%;
}
.me__fallback {
  display: flex;
  align-items: center;
  justify-content: center;
}
.me__text {
  color: #fff;
  font-weight: 700;
  font-size: 28rpx;
}
.brand {
  font-size: 34rpx;
  font-weight: 600;
  color: #333;
}
.map-wrap {
  flex: 1;
  min-height: 0;
  margin: 20rpx;
  border: 5rpx solid #fff;
  border-radius: 30rpx;
  overflow: hidden;
  background: #f4f5f7;
  box-sizing: border-box;
}
.map {
  width: 100%;
  height: 100%;
}
.scroll-tips {
  width: 100%;
  flex-shrink: 0;
  z-index: 11;
  padding: 0 30rpx;
  box-sizing: border-box;
  height: 80rpx;
  background: linear-gradient(90deg, #c3deff 0%, #e3f1fa 100%);
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
.bottom-panel {
  flex-shrink: 0;
  width: 100%;
  background: #fff;
  border-top-left-radius: 32rpx;
  border-top-right-radius: 32rpx;
  box-shadow: 0 -8rpx 24rpx rgba(0, 0, 0, 0.04);
  padding-top: 8rpx;
  box-sizing: border-box;
  z-index: 13;
}
.shortcut-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 0 56rpx;
  min-height: 146rpx;
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
  width: 88rpx;
  height: 88rpx;
  margin-top: 16rpx;
  margin-bottom: 12rpx;
}
.shortcut-text {
  height: 34rpx;
  line-height: 34rpx;
  font-size: 24rpx;
  font-weight: 400;
  color: #666;
  margin-bottom: 8rpx;
}
.dock__btn {
  width: calc(100% - 48rpx);
  height: 96rpx;
  margin: 16rpx 24rpx 24rpx;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 34rpx;
  border-radius: 48rpx;
  box-sizing: border-box;
}
.dock__scan-icon {
  width: 40rpx;
  height: 40rpx;
  margin-right: 16rpx;
}
.dock__btn.is-disabled {
  pointer-events: none;
}
</style>
