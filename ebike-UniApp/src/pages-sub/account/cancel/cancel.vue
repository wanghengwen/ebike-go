<template>
  <!-- Legacy cancelAccount.vue -->
  <view class="page">
    <view class="cancel_confirm">
      <view class="confirm_title">
        {{ isConfirm ? t('account.cancelTitleConfirm') : t('account.cancelTitle') }}
      </view>

      <view v-if="!isConfirm">
        <view class="confirm_des">{{ t('account.cancelCheckIntro') }}</view>
        <view class="confirm_explain">
          <view v-for="(item, i) in checklist" :key="i" class="confirm_box">
            <view class="confirm_icon">{{ i + 1 }}</view>
            <view class="confirm_content">{{ item }}</view>
          </view>
        </view>
      </view>

      <view v-else>
        <view class="confirm_num">{{ maskPhone }}</view>
        <view class="confirm_des1">
          {{ isSatisfy ? t('account.cancelReadyDetail') : t('account.cancelBlockedDetail') }}
        </view>
        <view v-if="!isSatisfy">
          <view v-if="flags.izRecharge" class="confirm_order">
            <view class="confirm_name">{{ t('account.cancelFlagBalanceName') }}</view>
            <view class="confirm_describe">{{ t('account.cancelFlagBalanceDesc') }}</view>
          </view>
          <view v-if="flags.izToPay" class="confirm_order">
            <view class="confirm_name">{{ t('account.cancelFlagOrderName') }}</view>
            <view class="confirm_describe">{{ t('account.cancelFlagOrderDesc') }}</view>
          </view>
          <view v-if="flags.izDeposit" class="confirm_order">
            <view class="confirm_name">{{ t('account.cancelFlagDepositName') }}</view>
            <view class="confirm_describe">{{ t('account.cancelFlagDepositDesc') }}</view>
          </view>
          <view v-if="flags.izHaveUnauditedUserTicket" class="confirm_order">
            <view class="confirm_name">{{ t('account.cancelFlagTicketName') }}</view>
            <view class="confirm_describe">{{ t('account.cancelFlagTicketDesc') }}</view>
          </view>
        </view>
        <view v-else class="confirm_order">
          <view class="confirm_name">{{ t('account.cancelIncludeTitle') }}</view>
          <view v-for="item in clearItems" :key="item" class="confirm_describes">{{ item }}</view>
        </view>
      </view>
    </view>

    <view v-if="!isConfirm" class="cancel_bottom">
      <view class="cancel_bottom_des" @click="checkAgree = !checkAgree">
        <image
          v-if="checkAgree && checkIcon"
          class="agree_icon"
          :src="checkIcon"
          mode="aspectFit"
        />
        <image
          v-else-if="!checkAgree && uncheckIcon"
          class="agree_icon"
          :src="uncheckIcon"
          mode="aspectFit"
        />
        <view v-else class="agree_fallback" :class="{ on: checkAgree }" />
        <view class="cancel_bottom_des_title">{{ t('account.cancelAgree') }}</view>
      </view>
      <button
        class="cancel_bottom_btn"
        :style="cancelButtonStyle"
        :disabled="!checkAgree"
        @click="onCancel"
      >
        {{ t('account.cancelAction') }}
      </button>
    </view>
    <view v-else class="cancel_bottom">
      <button
        v-if="!isSatisfy"
        class="cancel_bottom_btn"
        :style="confirmBtnStyle"
        @click="navigateBack"
      >
        {{ t('account.cancelGotIt') }}
      </button>
      <button
        v-else
        class="cancel_bottom_btn"
        :style="confirmBtnStyle"
        @click="onConfirmCancel"
      >
        {{ t('account.cancelConfirmAction') }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { queryCancelAccount } from '@/api/account'
import {
  getBrandColor,
  getButtonDisabledColor,
  getButtonWhiteColor,
} from '@/shared/config'
import { navigate, setNavTitle } from '@/shared/navigate'
import { phoneDesensitize } from '@/shared/phone'
import { getIconCfg } from '@/shared/tenantSkin'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const user = useUserStore()

const checkAgree = ref(false)
const isConfirm = ref(false)
const isSatisfy = ref(false)
const flags = ref({
  izRecharge: false,
  izToPay: false,
  izDeposit: false,
  izHaveUnauditedUserTicket: false,
})

const checkIcon = computed(() => getIconCfg('checkbox') || getIconCfg('checked_square_round'))
const uncheckIcon = computed(() => getIconCfg('invoiceUncheck'))
const maskPhone = computed(() => phoneDesensitize(user.userInfo.phone))

const checklist = computed(() => [
  t('account.cancelCheck1'),
  t('account.cancelCheck2'),
  t('account.cancelCheck3'),
  t('account.cancelCheck4'),
  t('account.cancelCheck5'),
])

const clearItems = computed(() => [
  t('account.cancelClear1'),
  t('account.cancelClear2'),
  t('account.cancelClear3'),
  t('account.cancelClear4'),
  t('account.cancelClear5'),
  t('account.cancelClear6'),
])

const cancelButtonStyle = computed(() => {
  const bg = checkAgree.value ? getBrandColor() : getButtonDisabledColor()
  const color = checkAgree.value ? '#1E4A38' : getButtonWhiteColor()
  return `background-color:${bg};border-color:${bg};color:${color};font-weight:bold;`
})

const confirmBtnStyle = computed(() => {
  const bg = getBrandColor()
  return `background-color:${bg};border-color:${bg};color:#1E4A38;font-weight:bold;`
})

onShow(() => {
  user.hydrateFromStorage()
  setNavTitle(t('account.cancelAccount'))
})

function navigateBack() {
  navigate('back')
}

async function onCancel() {
  if (!checkAgree.value) return
  const res = await queryCancelAccount()
  if (!res.success || !res.data) return
  const data = res.data as {
    success?: boolean
    izDeposit?: boolean
    izToPay?: boolean
    izRecharge?: boolean
    izHaveUnauditedUserTicket?: boolean
  }
  isConfirm.value = true
  isSatisfy.value = Boolean(data.success)
  flags.value = {
    izRecharge: Boolean(data.izRecharge),
    izToPay: Boolean(data.izToPay),
    izDeposit: Boolean(data.izDeposit),
    izHaveUnauditedUserTicket: Boolean(data.izHaveUnauditedUserTicket),
  }
}

function onConfirmCancel() {
  uni.showModal({
    title: t('account.cancelModalTitle'),
    content: t('account.cancelModalContent'),
    success: (res) => {
      if (res.confirm) {
        navigate('to', '/pages-sub/account/cancel/cancel-verify')
      }
    },
  })
}
</script>

<style scoped lang="scss">
.page {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  right: 0;
  background: #fff;
}
.cancel_confirm {
  margin-top: 80rpx;
  padding: 0 32rpx;
  box-sizing: border-box;
  padding-bottom: 280rpx;
}
.confirm_title {
  font-size: 40rpx;
  font-weight: bold;
  color: #262626;
  line-height: 56rpx;
}
.confirm_des {
  font-size: 28rpx;
  font-weight: 400;
  color: #121212;
  margin-top: 32rpx;
  line-height: 54rpx;
  margin-bottom: 32rpx;
}
.confirm_explain {
  padding-right: 32rpx;
}
.confirm_box {
  display: flex;
  margin-bottom: 32rpx;
}
.confirm_icon {
  width: 32rpx;
  height: 32rpx;
  background: #ff8401;
  border-radius: 50%;
  color: #fff;
  font-size: 22rpx;
  text-align: center;
  line-height: 32rpx;
  margin-right: 16rpx;
  flex-shrink: 0;
  margin-top: 6rpx;
}
.confirm_content {
  font-size: 28rpx;
  color: #333;
  line-height: 44rpx;
}
.confirm_num {
  margin-top: 48rpx;
  font-size: 40rpx;
  font-weight: 600;
  color: #121212;
}
.confirm_des1 {
  margin-top: 24rpx;
  font-size: 28rpx;
  color: #121212;
  line-height: 44rpx;
}
.confirm_order {
  margin-top: 40rpx;
}
.confirm_name {
  font-size: 30rpx;
  font-weight: 600;
  color: #262626;
  margin-bottom: 12rpx;
}
.confirm_describe,
.confirm_describes {
  font-size: 26rpx;
  color: #666;
  line-height: 40rpx;
  margin-bottom: 12rpx;
}
.cancel_bottom {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 24rpx 32rpx 80rpx;
  background: #fff;
}
.cancel_bottom_des {
  display: flex;
  align-items: flex-start;
  margin-bottom: 24rpx;
}
.agree_icon {
  width: 36rpx;
  height: 36rpx;
  margin-right: 12rpx;
  flex-shrink: 0;
  margin-top: 4rpx;
}
.agree_fallback {
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  border: 2rpx solid #ccc;
  box-sizing: border-box;
  margin-right: 12rpx;
  margin-top: 4rpx;
  flex-shrink: 0;
  &.on {
    background: #3aa0e8;
    border-color: #3aa0e8;
  }
}
.cancel_bottom_des_title {
  font-size: 24rpx;
  color: #666;
  line-height: 36rpx;
}
.cancel_bottom_btn {
  width: 100%;
  height: 96rpx;
  line-height: 96rpx;
  border-radius: 32rpx;
  font-size: 32rpx;
  padding: 0;
}
button::after {
  display: none;
}
</style>
