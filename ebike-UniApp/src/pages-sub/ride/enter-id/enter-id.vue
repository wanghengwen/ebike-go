<template>
  <view class="pages">
    <view class="spacer" />
    <view class="enter-tips">{{ t('ride.enterIdHint') }}</view>
    <view class="enter-bg" v-if="scanBg">
      <image class="enter-bg__img" :src="scanBg" mode="aspectFit" />
    </view>
    <view
      class="enter-input__wrapper"
      :class="{ current: isInputFocus }"
      @click="focusInput"
    >
      <input
        class="enter-input"
        type="number"
        v-model="carId"
        :focus="isInputFocus"
        :maxlength="12"
        :placeholder="t('ride.enterIdPlaceholder')"
        placeholder-class="enter-ph"
        @focus="isInputFocus = true"
        @blur="isInputFocus = false"
        @confirm="go"
      />
    </view>
    <view class="btn_text">
      <button class="btn" :style="btnStyle" :disabled="!canSubmit" @click="go">
        {{ t('common.confirm') }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onHide, onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { useScanGate } from '@/features/bike/useScanGate'
import { getBrandColor, getTenantConfig } from '@/shared/config'
import { getMapCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'

/** Align legacy eBikeIdLength */
const MIN_LEN = 7

const { t } = useI18n()
const { ensureCanScan } = useScanGate()

const carId = ref('')
const isInputFocus = ref(true)
const isRedEnvelope = ref(false)
const scanBg = computed(() => getMapCfg('scanBg'))

const canSubmit = computed(() => carId.value.trim().replace(/\D/g, '').length >= MIN_LEN)

const btnStyle = computed(() => {
  const cs = getTenantConfig().customSetting || {}
  const enabled = cs.buttonGreenColor || getBrandColor()
  const disabled = cs.buttonDisabledColor || '#cccccc'
  return {
    background: canSubmit.value ? enabled : disabled,
    color: '#fff',
  }
})

onLoad((q) => {
  isRedEnvelope.value = String(q?.isRedEnvelope || '') === 'true'
})

onShow(() => {
  setNavTitle(t('ride.enterId'))
  isInputFocus.value = true
})

onHide(() => reset())
onUnload(() => reset())

function reset() {
  carId.value = ''
  isInputFocus.value = false
}

function focusInput() {
  isInputFocus.value = true
}

async function go() {
  const id = carId.value.trim().replace(/\D/g, '')
  if (id.length < MIN_LEN) {
    uni.showToast({ title: t('ride.enterIdTooShort', { n: MIN_LEN }), icon: 'none' })
    return
  }
  const gate = await ensureCanScan({ skipVerified: true })
  if (!gate.ok) {
    if (gate.reason === 'unpaid') navigate('to', '/pages/pay/pay')
    return
  }
  const qs = [`carId=${encodeURIComponent(id)}`]
  if (isRedEnvelope.value) qs.push('isRedEnvelope=true')
  navigate('to', `/pages-sub/ride/precycling/precycling?${qs.join('&')}`)
}
</script>

<style scoped lang="scss">
.pages {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-height: 100vh;
  background: #fff;
  box-sizing: border-box;
}

.spacer {
  height: 40px;
}

.enter-tips {
  text-align: center;
  font-size: 32rpx;
  font-weight: 550;
  padding: 30rpx 0;
  color: #333;
}

.enter-bg {
  width: 454rpx;
  height: 190rpx;
  margin: 64rpx auto 80rpx;
}

.enter-bg__img {
  width: 100%;
  height: 100%;
}

.enter-input__wrapper {
  width: 654rpx;
  height: 100rpx;
  margin: 0 auto;
  box-sizing: border-box;
  background: #ffffff;
  border: 2rpx solid #cccccc;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.enter-input__wrapper.current {
  box-shadow: 0 8px 16px 0 rgba(99, 209, 68, 0.3);
  border: 2px solid #63d144;
}

.enter-input {
  width: 100%;
  height: 100%;
  text-align: center;
  font-size: 56rpx;
  font-weight: 500;
  color: #333333;
  padding: 0 24rpx;
  box-sizing: border-box;
}

.enter-ph {
  font-size: 32rpx;
  color: #cccccc;
  font-weight: 400;
}

.btn_text {
  margin: 64rpx 0;
}

.btn {
  width: 654rpx;
  height: 96rpx;
  border-radius: 32rpx;
  line-height: 96rpx;
  font-size: 32rpx;
  border: none;
}

.btn::after {
  border: none;
}

.btn[disabled] {
  color: #fff !important;
}
</style>
