<template>
  <!-- #ifdef MP-WEIXIN -->
  <view v-if="visible" class="mask">
    <view class="panel">
      <view class="title">{{ t('auth.privacyTitle') }}</view>
      <view class="content">
        <text class="gray">{{ t('auth.privacyIntro', { name: brandName }) }}</text>
        <text class="gray">{{ t('auth.privacyBody') }}</text>
        <view class="privacy-line">
          <text>{{ t('auth.privacyReadPrefix') }}</text>
          <text class="link" @click="openPrivacy">《{{ t('auth.privacyAgreement') }}》</text>
          <text>{{ t('auth.privacyReadSuffix') }}</text>
        </view>
        <text>{{ t('auth.privacyProtect') }}</text>
      </view>
      <view class="actions">
        <button class="disagree" @click="onDisagree">{{ t('auth.privacyDisagree') }}</button>
        <button
          id="agree-privacy-btn"
          class="agree"
          open-type="agreePrivacyAuthorization"
          @agreeprivacyauthorization="onAgree"
        >
          {{ t('auth.privacyAgree') }}
        </button>
      </view>
    </view>
  </view>
  <!-- #endif -->
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { openProtocol } from '@/shared/protocol'
import { getTenantConfig } from '@/shared/config'

const { t } = useI18n()
const visible = ref(false)
const brandName = computed(() => getTenantConfig().name || t('brand.name'))

type WxPrivacy = {
  getPrivacySetting?: (opts: {
    success?: (res: { needAuthorization?: boolean; privacyContractName?: string }) => void
    fail?: () => void
  }) => void
  requirePrivacyAuthorize?: (opts: {
    success?: () => void
    fail?: () => void
  }) => void
  openPrivacyContract?: (opts?: {
    success?: () => void
    fail?: () => void
  }) => void
  onNeedPrivacyAuthorization?: (cb: (resolve: (payload: { event: string; buttonId?: string }) => void) => void) => void
  exitMiniProgram?: (opts?: Record<string, unknown>) => void
}

function getWx(): WxPrivacy | null {
  // #ifdef MP-WEIXIN
  try {
    const g = globalThis as { wx?: WxPrivacy }
    return g.wx || null
  } catch {
    return null
  }
  // #endif
  // #ifndef MP-WEIXIN
  return null
  // #endif
}

function checkNeedAuthorize() {
  const w = getWx()
  if (!w) return
  if (typeof w.getPrivacySetting === 'function') {
    w.getPrivacySetting({
      success: (res) => {
        if (res?.needAuthorization) visible.value = true
      },
      fail: () => {
        // Older bases: try requirePrivacyAuthorize to surface need
        if (typeof w.requirePrivacyAuthorize === 'function') {
          w.requirePrivacyAuthorize({
            fail: () => {
              visible.value = true
            },
          })
        }
      },
    })
    return
  }
  if (typeof w.onNeedPrivacyAuthorization === 'function') {
    w.onNeedPrivacyAuthorization(() => {
      visible.value = true
    })
  }
}

function openPrivacy() {
  const w = getWx()
  if (w && typeof w.openPrivacyContract === 'function') {
    w.openPrivacyContract({})
    return
  }
  openProtocol('privacyProtocol')
}

function onDisagree() {
  const w = getWx()
  if (w && typeof w.exitMiniProgram === 'function') {
    w.exitMiniProgram({})
    return
  }
  visible.value = false
}

function onAgree() {
  visible.value = false
}

onMounted(() => {
  // #ifdef MP-WEIXIN
  checkNeedAuthorize()
  // #endif
})
</script>

<style scoped lang="scss">
.mask {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48rpx;
  box-sizing: border-box;
}
.panel {
  width: 100%;
  max-width: 640rpx;
  background: #fff;
  border-radius: 24rpx;
  padding: 36rpx 32rpx 28rpx;
}
.title {
  text-align: center;
  font-size: 34rpx;
  font-weight: 700;
  margin-bottom: 24rpx;
}
.content {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  font-size: 26rpx;
  line-height: 1.55;
  color: #333;
  max-height: 55vh;
  overflow-y: auto;
}
.gray {
  color: #666;
}
.privacy-line {
  line-height: 1.55;
}
.link {
  color: var(--brand-color, #3aa0e8);
}
.actions {
  margin-top: 28rpx;
  display: flex;
  gap: 20rpx;
}
.disagree,
.agree {
  flex: 1;
  height: 80rpx;
  line-height: 80rpx;
  border-radius: 16rpx;
  font-size: 28rpx;
  padding: 0;
  margin: 0;
  &::after {
    border: none;
  }
}
.disagree {
  background: #f5f6f8;
  color: #666;
}
.agree {
  background: var(--brand-color, #3aa0e8);
  color: #fff;
}
</style>
