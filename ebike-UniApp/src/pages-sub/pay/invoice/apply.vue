<template>
  <!-- Legacy invoiceDetail：填写抬头并提交 -->
  <view class="page">
    <scroll-view scroll-y class="scroll">
      <view class="card">
        <view class="title">{{ t('pay.invoiceSelfService') }}</view>
        <view class="content">
          <view class="row">
            <text class="label">{{ t('pay.invoiceHeaderType') }}</text>
            <view class="options" @click="info.type = 1">
              <image
                v-if="info.type === 1 && checkIcon"
                class="img_check"
                :src="checkIcon"
                mode="aspectFit"
              />
              <image
                v-else-if="info.type !== 1 && uncheckIcon"
                class="img_check"
                :src="uncheckIcon"
                mode="aspectFit"
              />
              <text>{{ t('pay.invoiceCompany') }}</text>
            </view>
            <view class="options" @click="info.type = 0">
              <image
                v-if="info.type === 0 && checkIcon"
                class="img_check"
                :src="checkIcon"
                mode="aspectFit"
              />
              <image
                v-else-if="info.type !== 0 && uncheckIcon"
                class="img_check"
                :src="uncheckIcon"
                mode="aspectFit"
              />
              <text>{{ t('pay.invoicePersonal') }}</text>
            </view>
          </view>
          <view class="row">
            <text class="label">{{ t('pay.invoiceCompanyName') }}</text>
            <input
              v-model="info.title"
              class="input"
              type="text"
              :placeholder="t('pay.invoiceTitlePlaceholder')"
            />
          </view>
          <view v-if="info.type === 1" class="row">
            <text class="label">{{ t('pay.invoiceTaxNoLabel') }}</text>
            <input
              v-model="info.companyEin"
              class="input"
              type="text"
              :placeholder="t('pay.invoiceTaxNoPlaceholder')"
            />
          </view>
          <view v-if="info.type === 1" class="row" @click="showMore = !showMore">
            <text class="label">{{ t('pay.invoiceMore') }}</text>
            <text class="text">
              {{ t('pay.invoiceMoreFilled', { n: filledCount }) }}
            </text>
          </view>
          <template v-if="showMore && info.type === 1">
            <view class="row">
              <text class="label">{{ t('pay.invoiceBank') }}</text>
              <input v-model="info.bank" class="input" type="text" :placeholder="t('pay.invoiceBankPlaceholder')" />
            </view>
            <view class="row">
              <text class="label">{{ t('pay.invoiceBankAccount') }}</text>
              <input
                v-model="info.bankAccount"
                class="input"
                type="text"
                :placeholder="t('pay.invoiceBankAccountPlaceholder')"
              />
            </view>
            <view class="row">
              <text class="label">{{ t('pay.invoiceCompanyAddressShort') }}</text>
              <input
                v-model="info.companyAddress"
                class="input"
                type="text"
                :placeholder="t('pay.invoiceCompanyAddressPlaceholder')"
              />
            </view>
            <view class="row">
              <text class="label">{{ t('pay.invoicePhoneShort') }}</text>
              <input
                v-model="info.companyPhone"
                class="input"
                type="text"
                :placeholder="t('pay.invoiceCompanyPhonePlaceholder')"
              />
            </view>
          </template>
          <view class="row row--last">
            <text class="label">{{ t('pay.invoiceAmountTotal') }}</text>
            <text class="text">
              <text class="money">{{ money }}</text>{{ t('pay.invoiceYuan') }} {{ t('pay.invoiceOneSheet') }}
            </text>
          </view>
        </view>
      </view>

      <view class="card">
        <view class="title">{{ t('pay.invoiceReceiveWay') }}</view>
        <view class="content">
          <view class="row row--last">
            <text class="label">{{ t('pay.invoiceEmail') }}</text>
            <input
              v-model="info.email"
              class="input"
              type="text"
              :placeholder="t('pay.invoiceEmailPlaceholder')"
            />
          </view>
        </view>
      </view>

      <view class="tips">
        <view class="tip">{{ t('pay.invoiceTip1') }}</view>
        <view class="tip">{{ t('pay.invoiceTip2') }}</view>
      </view>
    </scroll-view>

    <view class="button_container">
      <button
        class="button"
        :style="{ background: brandColor, color: '#fff' }"
        :loading="subLoading"
        @click="submit"
      >
        {{ t('common.submit') }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { createInvoice } from '@/api/invoice'
import { getBrandColor } from '@/shared/config'
import { getIconCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { useUserStore } from '@/stores/user'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const user = useUserStore()
const brandColor = computed(() => getBrandColor())
const checkIcon = computed(() => getIconCfg('checkbox') || getIconCfg('checked_square_round'))
const uncheckIcon = computed(() => getIconCfg('invoiceUncheck'))

const showMore = ref(false)
const money = ref(0)
const orderIds = ref<Array<string | number>>([])
const subLoading = ref(false)

const info = reactive({
  type: 1,
  title: '',
  companyEin: '',
  email: '',
  phone: '',
  bank: '',
  companyPhone: '',
  companyAddress: '',
  bankAccount: '',
  content: '运输服务',
})

const filledCount = computed(() => {
  let n = 0
  if (info.bank) n += 1
  if (info.bankAccount) n += 1
  if (info.companyAddress) n += 1
  if (info.companyPhone) n += 1
  return n
})

onShow(() => setNavTitle(t('pay.invoiceApplyTitle')))

onLoad((q) => {
  try {
    const raw = q?.params ? decodeURIComponent(String(q.params)) : ''
    if (!raw) return
    const params = JSON.parse(raw) as { money?: number; orderIds?: Array<string | number> }
    money.value = Number(params.money || 0)
    orderIds.value = Array.isArray(params.orderIds) ? params.orderIds : []
  } catch (e) {
    logger.warn('invoice apply parse fail', e)
  }
})

async function submit() {
  if (subLoading.value) return
  if (!info.title) {
    uni.showToast({ title: t('pay.invoiceNeedTitle'), icon: 'none' })
    return
  }
  if (!info.email) {
    uni.showToast({ title: t('pay.invoiceNeedEmail'), icon: 'none' })
    return
  }
  if (info.type && !info.companyEin) {
    uni.showToast({ title: t('pay.invoiceNeedTaxNo'), icon: 'none' })
    return
  }
  subLoading.value = true
  try {
    const res = await createInvoice({
      ...info,
      orderIds: orderIds.value,
      serviceId: storage.get('serviceId', ''),
      userPin: user.userInfo.pin,
    })
    if (res.success) {
      uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
      setTimeout(() => {
        // Legacy navigate('back', 2)
        uni.navigateBack({ delta: 2 })
      }, 1000)
    } else {
      uni.showToast({ title: t('common.submitFail'), icon: 'none' })
    }
  } finally {
    subLoading.value = false
  }
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #f8f8f8;
  position: relative;
}
.scroll {
  height: calc(100% - 200rpx);
  box-sizing: border-box;
}
.img_check {
  width: 32rpx;
  height: 32rpx;
  margin-right: 16rpx;
}
.card {
  margin: 32rpx;
  padding: 32rpx 32rpx 48rpx;
  background: #fff;
  border-radius: 32rpx;
}
.title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 40rpx 0;
  border-bottom: 2rpx solid #f6f6f6;
}
.row--last {
  border: none;
  padding-bottom: 0;
}
.label {
  font-size: 28rpx;
  font-weight: 500;
  color: #666;
  flex-shrink: 0;
  margin-right: 24rpx;
}
.options {
  display: flex;
  align-items: center;
  font-size: 28rpx;
  color: #333;
  margin-left: 16rpx;
}
.input {
  flex: 1;
  font-size: 28rpx;
  color: #999;
  text-align: right;
}
.text {
  font-size: 28rpx;
  color: #333;
}
.money {
  color: #ff922b;
}
.tips {
  margin: 32rpx 32rpx 0;
}
.tip {
  font-size: 24rpx;
  color: #999;
}
.button_container {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 32rpx 32rpx 100rpx;
  background: #f8f8f8;
}
.button {
  width: 686rpx;
  height: 96rpx;
  line-height: 96rpx;
  border-radius: 32rpx;
  font-size: 32rpx;
}
button::after {
  display: none;
}
</style>
