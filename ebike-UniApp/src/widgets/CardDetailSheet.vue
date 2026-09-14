<template>
  <view v-if="visible" class="mask" @click.self="emit('close')">
    <view class="dialog">
      <view class="dialog_top">
        <FoldCard
          :card-name="String(card.name || card.ridingCardName || '')"
          :card-tag="tagList"
          :card-origin-cost="originCost"
          :card-current-cost="curCost"
          :card-time="card.cardTime || card.card_time"
          :card-cover="cover"
          :card-promotion-tag="String(card.promotionTag || '')"
        />
      </view>
      <view class="dialog_middle">
        <view class="container_top">
          <text class="title">{{ t('account.cardUsage') }}</text>
          <text class="all" @click="emit('rule')">{{ t('account.cardRule') }}</text>
          <image v-if="arrowIcon" class="arrow_img" :src="arrowIcon" mode="aspectFit" />
        </view>
        <rich-text v-if="detailHtml" class="extra" :nodes="detailHtml" />
        <view class="line" />
        <view class="pay_channel">
          <view class="channel_title">{{ t('pay.payChannel') }}</view>
          <view class="channel_item">
            <image v-if="wechatPayIcon" class="wechatpay" :src="wechatPayIcon" mode="aspectFit" />
            <text class="item_title">{{ t('pay.wechatPay') }}</text>
            <image v-if="checkIcon" class="item_checkbox" :src="checkIcon" mode="aspectFit" />
          </view>
        </view>
      </view>
      <view class="dialog_bottom">
        <text class="cancle_button" @click="emit('close')">{{ t('common.cancel') }}</text>
        <view class="pay_button" :style="payBtnStyle" @click="emit('pay')">
          <text>{{ t('account.payCardNow', { m: formatMoney(curCost) }) }}</text>
          <text v-if="saveFen > 0" class="discount">
            {{ t('pay.savedAmount', { m: formatMoney(saveFen) }) }}
          </text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import FoldCard from '@/widgets/FoldCard.vue'
import { formatMoney, formatRichText } from '@/shared/format'
import { getBrandColor } from '@/shared/config'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'

const props = defineProps<{
  visible: boolean
  card: Record<string, unknown>
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'pay'): void
  (e: 'rule'): void
}>()

const { t } = useI18n()

const arrowIcon = computed(() => getMapCfg('iconRight'))
const wechatPayIcon = computed(() => getIconCfg('wechatPay'))
const checkIcon = computed(() => getIconCfg('checkbox') || getIconCfg('checked_square_round'))
const defaultCover = computed(() => getIconCfg('walletDiscountCard') || getIconCfg('walletCyclingCard'))

const cover = computed(
  () =>
    String(props.card.backOfCardUrl || props.card.cardCover || '') || defaultCover.value || '',
)

const tagList = computed(() => {
  const raw = String(props.card.descriptionTag || props.card.description_tag || '')
  return raw
    .split('|')
    .map((s) => s.trim())
    .filter(Boolean)
})

const curCost = computed(() =>
  Number(props.card.curCost ?? props.card.current_cost ?? props.card.price ?? 0),
)
const originCost = computed(() =>
  Number(props.card.originCost ?? props.card.origin_cost ?? props.card.originalPrice ?? 0),
)
const saveFen = computed(() =>
  originCost.value > curCost.value ? originCost.value - curCost.value : 0,
)

const detailHtml = computed(() => {
  const raw = String(props.card.detailInfo || props.card.detail_info || '')
  return raw ? formatRichText(raw) : ''
})

const payBtnStyle = computed(() => {
  const bg = getBrandColor()
  return {
    backgroundColor: bg,
    borderColor: bg,
    color: '#1E4A38',
    fontWeight: 'bold',
  }
})
</script>

<style scoped lang="scss">
.mask {
  position: fixed;
  inset: 0;
  z-index: 1200;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: flex-end;
}
.dialog {
  width: 750rpx;
  background: #ffffff;
  padding-top: 32rpx;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
  border-radius: 32rpx 32rpx 0 0;
  max-height: 90vh;
  overflow-y: auto;
  box-sizing: border-box;
}
.dialog_top {
  margin: 0 32rpx 32rpx;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.04);
  border-radius: 32rpx;
}
.dialog_middle {
  margin: 0 32rpx;
}
.container_top {
  display: flex;
  align-items: center;
  margin-bottom: 16rpx;
}
.title {
  flex: 1;
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
}
.all {
  font-size: 24rpx;
  color: #666;
}
.arrow_img {
  width: 24rpx;
  height: 24rpx;
  margin-left: 4rpx;
}
.extra {
  color: #666;
  font-size: 24rpx;
  line-height: 1.5;
}
.line {
  height: 2rpx;
  background: #f6f6f6;
  margin: 24rpx 0;
}
.pay_channel {
  margin-bottom: 24rpx;
}
.channel_title {
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 16rpx;
}
.channel_item {
  display: flex;
  align-items: center;
}
.wechatpay {
  width: 48rpx;
  height: 48rpx;
  margin-right: 16rpx;
}
.item_title {
  flex: 1;
  font-size: 28rpx;
  color: #333;
}
.item_checkbox {
  width: 40rpx;
  height: 40rpx;
}
.dialog_bottom {
  display: flex;
  align-items: center;
  margin: 24rpx 32rpx 0;
  gap: 20rpx;
}
.cancle_button {
  width: 160rpx;
  height: 96rpx;
  line-height: 96rpx;
  text-align: center;
  font-size: 30rpx;
  color: #666;
  flex-shrink: 0;
}
.pay_button {
  flex: 1;
  min-height: 96rpx;
  border-radius: 48rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  padding: 12rpx 16rpx;
  box-sizing: border-box;
}
.discount {
  font-size: 22rpx;
  font-weight: 400;
  margin-top: 4rpx;
  opacity: 0.9;
}
</style>
