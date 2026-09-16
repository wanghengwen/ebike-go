<script setup lang="ts">
import { onLaunch, onShow, onHide, onPageNotFound } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { initLocale, t } from '@/locales'
import { loadTenantRuntime, getBrandColor } from '@/shared/config'
import { logger } from '@/shared/logger'
import { mapLegacyPagePath } from '@/shared/mapLegacyPath'
import { syncFromNativeHost } from '@/shared/nativeHostSync'
import { isNative } from '@/shared/nativeHost'

/** Rider WebView：藏掉 Uni 默认标题栏整块区域，只留原生容器顶栏。 */
function applyNativeHostChrome() {
  // #ifdef H5
  if (typeof document === 'undefined' || !isNative()) return
  document.documentElement.classList.add('rider-native-host')
  const css = `
/* 整栏隐藏（含安全区占位），避免与原生顶栏双标题 */
.rider-native-host uni-page-head,
.rider-native-host .uni-page-head {
  display: none !important;
  height: 0 !important;
  min-height: 0 !important;
  padding: 0 !important;
  margin: 0 !important;
  overflow: hidden !important;
  pointer-events: none !important;
  opacity: 0 !important;
}
/* 个人中心等 custom 页的页内返回箭头 */
.rider-native-host .nav-back {
  display: none !important;
  width: 0 !important;
  height: 0 !important;
}
/*
 * 个人中心顶距（仅原生）：与 profile.vue .is-native-host 一致。
 * 禁止 height:auto —— 会与 .my-property 负 margin 叠盖快捷入口文字。
 * 单位用 px（注入样式不经 uni rpx 换算）；按 375 宽 1rpx≈0.5px。
 */
.rider-native-host .pages .info {
  padding-top: 24px !important;
}
.rider-native-host .pages .top_container {
  /* 238px + 24px，与 profile.vue 524rpx 对齐 */
  height: 262px !important;
  min-height: 262px !important;
}
/* 原先为默认顶栏让出的高度，藏栏后收回 */
.rider-native-host uni-page-head[uni-page-head-type='default'] ~ uni-page-wrapper,
.rider-native-host uni-page-wrapper {
  height: 100% !important;
}
.rider-native-host uni-page,
.rider-native-host .uni-page,
.rider-native-host .uni-page--open,
.rider-native-host .uni-page--close,
.rider-native-host .uni-page--show,
.rider-native-host .uni-page--hide,
.rider-native-host .uni-page-head,
.rider-native-host .uni-main {
  animation: none !important;
  transition: none !important;
  -webkit-transition: none !important;
  -webkit-animation: none !important;
}
`
  let style = document.getElementById('rider-native-host-chrome') as HTMLStyleElement | null
  if (!style) {
    style = document.createElement('style')
    style.id = 'rider-native-host-chrome'
    document.head.appendChild(style)
  }
  style.textContent = css
  // #endif
}

function setupUpdateManager() {
  // #ifdef MP-WEIXIN
  try {
    const updateManager = uni.getUpdateManager()
    updateManager.onUpdateReady(() => {
      uni.showModal({
        title: t('common.updateTitle'),
        content: t('common.updateContent'),
        showCancel: false,
        success(res) {
          if (res.confirm) updateManager.applyUpdate()
        },
      })
    })
    updateManager.onUpdateFailed(() => {
      logger.warn('mini program update failed')
    })
  } catch (e) {
    logger.warn('getUpdateManager unavailable', e)
  }
  // #endif
}

onLaunch(() => {
  loadTenantRuntime()
  initLocale()
  const user = useUserStore()
  user.hydrateFromStorage()
  void syncFromNativeHost()
  // Keep shared/mapLegacyPath.js in the main-package require graph for WeChat.
  mapLegacyPagePath('/')
  // #ifdef H5
  if (typeof document !== 'undefined') {
    document.documentElement.style.setProperty('--brand-color', getBrandColor())
  }
  applyNativeHostChrome()
  // 桥注入可能略晚于 onLaunch，短延迟再补一次 class。
  setTimeout(applyNativeHostChrome, 0)
  setTimeout(applyNativeHostChrome, 300)
  // #endif
  logger.info('app launch')
  setTimeout(() => setupUpdateManager(), 5000)
})

onShow(() => {
  logger.info('app show')
})

onHide(() => {
  logger.info('app hide')
})

/** Legacy: offline QR / unknown path → launch with q for car deep-link */
onPageNotFound((res) => {
  const q = (res as { query?: { q?: string } })?.query?.q
  const url = q
    ? `/pages/launch/launch?q=${encodeURIComponent(String(q))}`
    : '/pages/launch/launch'
  uni.redirectTo({ url })
})
</script>
