<template>
  <!-- Legacy protocol.vue：条款与隐私政策入口列表 -->
  <view class="page">
    <view class="container">
      <view class="row" @click="openDetail(userProtocol)">
        <text class="row_text">{{ userAgreementLabel }}</text>
        <image v-if="arrow" class="image" :src="arrow" mode="aspectFit" />
      </view>
      <view class="line" />
      <view class="row" @click="openDetail(reChargeProtocol)">
        <text class="row_text">{{ t('account.rechargeProtocolTitle') }}</text>
        <image v-if="arrow" class="image" :src="arrow" mode="aspectFit" />
      </view>
      <view class="line" />
      <view class="row" @click="openDetail(privacyProtocol)">
        <text class="row_text">{{ t('account.privacyPolicyTitle') }}</text>
        <image v-if="arrow" class="image" :src="arrow" mode="aspectFit" />
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getTenantConfig } from '@/shared/config'
import { navigate, setNavTitle } from '@/shared/navigate'
import { PROTOCOL_CONFIGS, type ProtocolMeta } from '@/shared/protocolConfigs'
import { getMapCfg } from '@/shared/tenantSkin'

const { t } = useI18n()
const arrow = computed(() => getMapCfg('iconRightRound'))
const userProtocol = PROTOCOL_CONFIGS.userProtocol
const reChargeProtocol = PROTOCOL_CONFIGS.reChargeProtocol
const privacyProtocol = PROTOCOL_CONFIGS.privacyProtocol

const userAgreementLabel = computed(() => {
  const textCfg = (getTenantConfig().customSetting as { textCfg?: Record<string, string> } | undefined)
    ?.textCfg
  return textCfg?.MotorcycleUserAgreement || t('account.motorcycleUserAgreement')
})

onShow(() => setNavTitle(t('account.termsPrivacy')))

function openDetail(protocol: ProtocolMeta) {
  navigate(
    'to',
    `/pages-sub/account/protocol/detail?protocol=${encodeURIComponent(JSON.stringify(protocol))}`,
  )
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #fff;
}
.container {
  padding: 32rpx 48rpx 0;
}
.row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.image {
  width: 32rpx;
  height: 32rpx;
}
.row_text {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.line {
  height: 2rpx;
  margin: 48rpx 0 46rpx;
  background: #f6f6f6;
}
</style>
