<template>
  <view class="card" @click="$emit('click')">
    <view class="card_content">
      <image v-if="cardCover" class="riding_card_img" :src="cardCover" mode="aspectFit" />
      <view class="card_des">
        <view class="card_title">{{ cardName }}</view>
        <view v-if="tags.length" class="card_tips">
          <block v-for="(item, index) in tags" :key="index">
            <text class="tip_item">{{ item }}</text>
            <view v-if="index + 1 !== tags.length" class="line" />
          </block>
        </view>
        <view v-if="cardTime" class="carTime">{{ cardTime }}{{ t('account.cardValidDays') }}</view>
        <view class="rest_times">
          <text class="new_price">{{ formatMoney(cardCurrentCost) }}</text>
          <text>{{ t('ride.yuan') }}</text>
          <text v-if="Number(cardOriginCost) > 0" class="old_price">
            {{ formatMoney(cardOriginCost) }}{{ t('ride.yuan') }}
          </text>
        </view>
      </view>
    </view>
    <view v-if="cardPromotionTag" class="using">{{ cardPromotionTag }}</view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatMoney } from '@/shared/format'

const props = defineProps<{
  cardName?: string
  cardTag?: string[] | string
  cardPromotionTag?: string
  cardOriginCost?: number | string
  cardCurrentCost?: number | string
  cardTime?: string | number
  cardCover?: string
}>()

defineEmits<{ (e: 'click'): void }>()

const { t } = useI18n()

const tags = computed(() => {
  if (Array.isArray(props.cardTag)) return props.cardTag.filter(Boolean)
  const raw = String(props.cardTag || '')
  return raw
    .split('|')
    .map((s) => s.trim())
    .filter(Boolean)
})
</script>

<style scoped lang="scss">
.card {
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 16rpx 32rpx 16rpx 16rpx;
  background: #ffffff;
  border-radius: 32rpx;
}
.card_content {
  display: flex;
}
.riding_card_img {
  width: 184rpx;
  height: 214rpx;
  flex-shrink: 0;
}
.card_des {
  flex: 1;
  margin-top: 16rpx;
  margin-left: 32rpx;
  min-width: 0;
}
.card_title {
  font-size: 40rpx;
  font-weight: 600;
  color: #333333;
}
.carTime {
  font-size: 28rpx;
  color: #666;
}
.card_tips {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  margin-top: 8rpx;
  margin-bottom: 16rpx;
  font-size: 28rpx;
  color: #666666;
}
.line {
  width: 2rpx;
  height: 26rpx;
  background: #666666;
  margin: 0 12rpx;
}
.rest_times {
  margin-top: 16rpx;
  font-size: 24rpx;
  font-weight: 500;
  color: #333333;
}
.new_price {
  font-size: 48rpx;
  font-weight: bold;
  color: #333333;
}
.old_price {
  margin-left: 20rpx;
  font-size: 24rpx;
  font-weight: 500;
  color: #999;
  text-decoration: line-through;
}
.using {
  position: absolute;
  top: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 144rpx;
  height: 48rpx;
  border-radius: 0 32rpx 0 32rpx;
  background: #ff8a5e;
  font-size: 24rpx;
  font-weight: 600;
  color: #ffffff;
}
</style>
