<template>
  <view class="page">
    <scroll-view scroll-y class="scroll">
      <image v-if="headerImg" class="header" :src="headerImg" mode="widthFix" />
      <view class="card introduce">
        <view class="step">
          <view class="step__title">
            <image v-if="step1Img" class="step__num" :src="step1Img" mode="aspectFit" />
            <text>{{ t('ride.redEnvelopeRuleTitle1') }}</text>
          </view>
          <view class="step__imgs" v-if="intro1Img || intro2Img">
            <image v-if="intro1Img" class="step__illust half" :src="intro1Img" mode="aspectFit" />
            <image v-if="intro2Img" class="step__illust half" :src="intro2Img" mode="aspectFit" />
          </view>
        </view>

        <view class="step">
          <view class="step__title">
            <image v-if="step2Img" class="step__num" :src="step2Img" mode="aspectFit" />
            <text>{{ t('ride.redEnvelopeRuleTitle2') }}</text>
          </view>
          <image v-if="intro3Img" class="step__illust full" :src="intro3Img" mode="widthFix" />
        </view>

        <view class="step">
          <view class="step__title">
            <image v-if="step3Img" class="step__num" :src="step3Img" mode="aspectFit" />
            <text>{{ t('ride.redEnvelopeRuleTitle3') }}</text>
          </view>
          <view class="rules">
            <text class="rule">1) {{ t('ride.redEnvelopeRule1') }}</text>
            <text class="rule">2) {{ t('ride.redEnvelopeRule2') }}</text>
            <text class="rule">3) {{ t('ride.redEnvelopeRule3') }}</text>
            <view class="rule rule--multi">
              <text>4) {{ t('ride.redEnvelopeRule4Prefix') }}</text>
              <text v-if="ridingTime > 0">{{ t('ride.redEnvelopeRule4Time') }}</text>
              <text v-if="ridingTime > 0" class="special">{{ ridingTime }}{{ t('ride.minutes') }}</text>
              <text v-if="ridingTime > 0 && ridingDistance > 0">、</text>
              <text v-if="ridingDistance > 0">{{ t('ride.redEnvelopeRule4Dist') }}</text>
              <text v-if="ridingDistance > 0" class="special">{{ ridingDistance }}{{ t('ride.meters') }}</text>
              <text>{{ t('ride.redEnvelopeRule4Suffix') }}</text>
              <text class="special">{{ t('ride.redEnvelopeRule4Free') }}</text>
              <text>{{ t('ride.redEnvelopeRule4Billing') }}</text>
            </view>
            <text class="rule">5) {{ t('ride.redEnvelopeRule5') }}</text>
            <text class="rule">6) {{ t('ride.redEnvelopeRule6') }}</text>
            <text class="rule">7) {{ t('ride.redEnvelopeRule7') }}</text>
            <text class="rule">8) {{ t('ride.redEnvelopeRule8') }}</text>
          </view>
        </view>

        <view class="btn-primary" @click="goBack">{{ t('common.confirm') }}</view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getRedEnvelopeRule } from '@/api/redEnvelope'
import { getRedEnvelopeCfg } from '@/features/bike/redEnvelope'
import { navigate, setNavTitle } from '@/shared/navigate'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const activityId = ref('')
const ridingTime = ref(5)
const ridingDistance = ref(500)

onShow(() => setNavTitle(t('ride.redEnvelopeTips')))

onLoad(async (q) => {
  activityId.value = decodeURIComponent(String(q?.id || ''))
  if (!activityId.value) return
  try {
    const res = await getRedEnvelopeRule({ id: activityId.value })
    if (res.success && res.data) {
      const data = res.data as { ridingTime?: number; ridingDistance?: number }
      ridingTime.value = Number(data.ridingTime) || 0
      ridingDistance.value = Number(data.ridingDistance) || 0
    } else if (res.msg) {
      uni.showToast({ title: res.msg, icon: 'none' })
    }
  } catch (e) {
    logger.warn('getRedEnvelopeRule fail', e)
  }
})

const headerImg = computed(() => getRedEnvelopeCfg('redEnvelopeTipsHeader'))
const step1Img = computed(() => getRedEnvelopeCfg('redEnvelopeTipsStep1'))
const step2Img = computed(() => getRedEnvelopeCfg('redEnvelopeTipsStep2'))
const step3Img = computed(() => getRedEnvelopeCfg('redEnvelopeTipsStep3'))
const intro1Img = computed(() => getRedEnvelopeCfg('redEnvelopeTipsIntroduce1'))
const intro2Img = computed(() => getRedEnvelopeCfg('redEnvelopeTipsIntroduce2'))
const intro3Img = computed(() => getRedEnvelopeCfg('redEnvelopeTipsIntroduce3'))

function goBack() {
  navigate('back')
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: #f5f6f8;
}
.scroll {
  height: 100vh;
}
.header {
  width: 100%;
  display: block;
}
.introduce {
  margin-top: -48rpx;
  position: relative;
  z-index: 1;
  border-radius: 40rpx 40rpx 0 0;
  padding: 24rpx 32rpx 48rpx;
}
.step {
  margin-top: 36rpx;
}
.step__title {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  line-height: 1.4;
}
.step__num {
  width: 56rpx;
  height: 56rpx;
  flex-shrink: 0;
}
.step__imgs {
  margin-top: 24rpx;
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
}
.step__illust.half {
  width: 48%;
  height: 240rpx;
}
.step__illust.full {
  width: 100%;
  margin-top: 24rpx;
  display: block;
}
.rules {
  margin-top: 24rpx;
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}
.rule {
  font-size: 26rpx;
  color: #666;
  line-height: 1.5;
}
.rule--multi {
  display: block;
}
.special {
  color: #ff8b00;
}
.btn-primary {
  margin-top: 36rpx;
}
</style>
