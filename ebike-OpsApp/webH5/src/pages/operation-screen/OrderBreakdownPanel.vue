<script setup lang="ts">
import { numFormat, numFormatInt } from '@/utils/numFormat'
import type { OrderBreakdown } from './useQueryData'

defineProps<{
  count: OrderBreakdown
  amount: OrderBreakdown
}>()

const BY_TIME: Array<[label: string, key: keyof OrderBreakdown]> = [
  ['普通订单', 'general_order'],
  ['超长订单', 'long_order'],
  ['短时订单', 'short_order'],
]

const BY_ZONE: Array<[label: string, key: keyof OrderBreakdown]> = [
  ['正常订单', 'normal_order'],
  ['站点外订单', 'parking_zone'],
  ['服务区外订单', 'out_of_service'],
  ['禁停区订单', 'in_no_parking'],
]
</script>

<template>
  <div>
    <div class="group-title">按时间统计</div>
    <div v-for="[label, key] in BY_TIME" :key="key" class="row">
      <span class="row__label">
        {{ label }}:<span class="row__value">{{ numFormatInt(count[key]) }}</span>
      </span>
      <span class="row__label row__label--amount">
        收益:<span class="row__value">{{ numFormat(amount[key]) }}</span>
      </span>
    </div>

    <div class="group-title group-title--spaced">按区域统计</div>
    <div v-for="[label, key] in BY_ZONE" :key="key" class="row">
      <span class="row__label">
        {{ label }}:<span class="row__value">{{ numFormatInt(count[key]) }}</span>
      </span>
      <span class="row__label row__label--amount">
        收益:<span class="row__value">{{ numFormat(amount[key]) }}</span>
      </span>
    </div>
  </div>
</template>

<style scoped>
.group-title {
  font-size: 13px;
  font-weight: 550;
  color: #282828;
  line-height: 18px;
  margin-left: 16px;
}

.group-title--spaced {
  margin-top: 14px;
}

/* 分隔线在遗留里是独立的兄弟 div，通栏，不跟着行内容缩进 */
.row {
  margin-top: 13px;
  padding: 0 0 12px 16px;
  border-bottom: 1px solid #dcdcdc;
}

.row:last-child {
  border-bottom: none;
}

.row__label {
  display: inline-block;
  width: 170px;
  font-size: 13px;
  color: #282828;
  line-height: 18px;
}

.row__label--amount {
  width: auto;
}

.row__value {
  display: inline-block;
  font-size: 15px;
  font-weight: bold;
  color: #282828;
  margin-left: 6px;
}
</style>
