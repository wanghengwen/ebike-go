<script setup lang="ts">
import { onLaunch, onShow, onHide, onPageNotFound } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { initLocale, t } from '@/locales'
import { loadTenantRuntime, getBrandColor } from '@/shared/config'
import { logger } from '@/shared/logger'
import { mapLegacyPagePath } from '@/shared/mapLegacyPath'

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
  // Keep shared/mapLegacyPath.js in the main-package require graph for WeChat.
  mapLegacyPagePath('/')
  // #ifdef H5
  if (typeof document !== 'undefined') {
    document.documentElement.style.setProperty('--brand-color', getBrandColor())
  }
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
