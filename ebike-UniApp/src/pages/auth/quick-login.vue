<template>
  <view class="login">
    <view class="logo-wrap">
      <image v-if="logo" class="logo-image" :src="logo" mode="aspectFit" />
      <text class="logo-title">{{ welcomeText }}</text>
    </view>

    <view class="remind">
      <view class="protocol-wrap">
        <view class="click-area" @click="agreed = !agreed">
          <image
            v-if="!agreed"
            class="checkbox"
            :src="uncheckIcon"
            mode="aspectFit"
          />
          <image
            v-else
            class="checkbox checked"
            :src="checkIcon"
            mode="aspectFit"
          />
        </view>
        <text class="hint">{{ t('auth.agreePrefix') }}</text>
        <text class="hint link" @click="openProtocol('userProtocol')">
          {{ userAgreementLabel }}
        </text>
        <text class="hint">{{ t('auth.and') }}</text>
        <text class="hint link" @click="openProtocol('privacyProtocol')">
          {{ privacyAgreementLabel }}
        </text>
      </view>
      <text class="hint age">
        <text class="red">{{ t('auth.ageForbidPrefix') }}</text>{{ t('auth.ageForbidSuffix') }}
      </text>
    </view>

    <view class="operate-wrap">
      <!-- #ifdef MP-WEIXIN -->
      <view v-if="!agreed" class="wechat-login-hide" @click="onNeedAgree" />
      <button
        class="wechat-login"
        :style="brandBtnStyle"
        open-type="getPhoneNumber"
        :loading="loading"
        :disabled="loading"
        @getphonenumber="onWxPhone"
      >
        {{ loading ? t('common.loading') : t('auth.quickLogin') }}
      </button>
      <!-- #endif -->
      <!-- #ifndef MP-WEIXIN -->
      <view class="wechat-login" :style="brandBtnStyle" @click="onNonWxLogin">
        {{ t('auth.quickLogin') }}
      </view>
      <!-- #endif -->
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { useAuth } from '@/features/auth/useAuth'
import { getBrandColor, getButtonWhiteColor, getTenantConfig } from '@/shared/config'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { openProtocol } from '@/shared/protocol'

const { t } = useI18n()
const { loginWithWechatPhone } = useAuth()
const loading = ref(false)
const agreed = ref(false)

const tenant = computed(() => getTenantConfig())
const logo = computed(() => String(tenant.value.customSetting?.logo || ''))
const brandName = computed(() => tenant.value.name || t('brand.name'))
const welcomeText = computed(() => t('auth.welcomeUse', { name: brandName.value }))
const uncheckIcon = computed(() => getIconCfg('invoiceUncheck'))
const checkIcon = computed(() => getIconCfg('checked_square_round'))
const userAgreementLabel = computed(
  () =>
    String((tenant.value.customSetting as { textCfg?: Record<string, string> })?.textCfg?.UserAgreement || '') ||
    `《${t('auth.userAgreement')}》`,
)
const privacyAgreementLabel = computed(
  () =>
    String((tenant.value.customSetting as { textCfg?: Record<string, string> })?.textCfg?.PrivacyAgreement || '') ||
    `《${t('auth.privacyAgreement')}》`,
)
const brandBtnStyle = computed(() => {
  const c = getBrandColor()
  return {
    backgroundColor: c,
    borderColor: c,
    color: getButtonWhiteColor(),
  }
})

onShow(() => {
  setNavTitle(brandName.value)
  // 登录页若为页面栈底层会显示「返回首页」，与旧版一致隐藏，避免跳过登录进主页
  try {
    // #ifdef MP-WEIXIN
    uni.hideHomeButton?.({})
    // #endif
  } catch {
    /* ignore */
  }
})

function onNeedAgree() {
  uni.showToast({ title: t('auth.agreeRequired'), icon: 'none' })
}

function onNonWxLogin() {
  if (!agreed.value) {
    onNeedAgree()
    return
  }
  navigate('to', '/pages/auth/phone-login')
}

async function onWxPhone(e: {
  detail?: { errMsg?: string; encryptedData?: string; iv?: string; code?: string }
}) {
  if (!agreed.value) {
    onNeedAgree()
    return
  }
  const detail = e.detail || {}
  if (!detail.encryptedData && !detail.code) {
    uni.showToast({ title: t('auth.wxAuthFail'), icon: 'none' })
    return
  }
  loading.value = true
  try {
    await loginWithWechatPhone(detail)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.login {
  width: 100vw;
  min-height: 100vh;
  background: #fff;
  padding-top: 40rpx;
  box-sizing: border-box;
}
.logo-wrap {
  margin-top: 180rpx;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}
.logo-image {
  height: 200rpx;
  width: 200rpx;
  border-radius: 50%;
}
.logo-title {
  margin-top: 40rpx;
  font-size: 40rpx;
  color: #2c2e39;
  font-weight: 600;
}
.remind {
  margin-top: 170rpx;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}
.protocol-wrap {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  justify-content: center;
  padding: 0 48rpx;
}
.click-area {
  height: 80rpx;
  width: 80rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.checkbox {
  width: 34rpx;
  height: 34rpx;
}
.hint {
  color: #999;
  font-size: 24rpx;
}
.hint.link {
  color: #333;
}
.age {
  margin-top: 8rpx;
}
.red {
  color: #e34d59;
}
.operate-wrap {
  margin-top: 50rpx;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  position: relative;
  padding: 0 56rpx 80rpx;
  box-sizing: border-box;
}
.wechat-login-hide {
  width: 636rpx;
  height: 120rpx;
  position: absolute;
  background: transparent;
  top: 0;
  z-index: 99;
}
.wechat-login {
  width: 636rpx;
  height: 120rpx;
  border-radius: 60rpx;
  font-size: 32rpx;
  font-weight: 600;
  display: flex;
  justify-content: center;
  align-items: center;
  border: none;
  line-height: 120rpx;
  padding: 0;
}
.wechat-login::after {
  border: none;
}
</style>
