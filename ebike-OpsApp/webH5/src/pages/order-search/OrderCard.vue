<script setup lang="ts">
import type { OrderItem } from '@/api'
import { distance, duration, isSettled, orDash, payStateLabel, yuan } from './format'

defineProps<{ order: OrderItem; clickable?: boolean }>()
</script>

<template>
  <div class="order" :class="{ 'order--clickable': clickable }">
    <div class="order__head">
      <span class="order__car">{{ orDash(order.carId) }}</span>
      <span class="order__state" :class="{ 'order__state--open': !isSettled(order.izPaid) }">
        {{ payStateLabel(order.izPaid) }}
      </span>
    </div>

    <div class="order__grid">
      <p><span>金额</span><b>{{ yuan(order.payCost) }} 元</b></p>
      <p><span>用时</span><b>{{ duration(order.ridingTime) }}</b></p>
      <p><span>里程</span><b>{{ distance(order.mile) }}</b></p>
      <p><span>用户</span><b>{{ orDash(order.phone) }}</b></p>
    </div>

    <div class="order__time">
      <p>起 {{ orDash(order.startTime) }}</p>
      <p>止 {{ orDash(order.endTime) }}</p>
    </div>

    <p class="order__id">订单号 {{ orDash(order.id) }}</p>
  </div>
</template>

<style scoped>
.order {
  background: #ffffff;
  border-radius: 8px;
  box-shadow: 0 2px 4px 0 rgba(0, 0, 0, 0.04);
  padding: 12px 14px;
  margin: 10px 12px 0;
}

.order--clickable {
  cursor: pointer;
}

.order__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.order__car {
  font-size: 17px;
  font-weight: 600;
  color: #282828;
}

.order__state {
  font-size: 12px;
  color: #999999;
}

/* 未结清的单要一眼看出来，运维找的多半就是这些。 */
.order__state--open {
  color: #ee0a24;
  font-weight: 600;
}

.order__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px 12px;
  margin-top: 10px;
}

.order__grid p {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.order__grid span {
  color: #999999;
}

.order__grid b {
  color: #282828;
  font-weight: 500;
}

.order__time {
  margin-top: 10px;
  padding-top: 8px;
  border-top: 1px solid #f2f2f2;
  font-size: 12px;
  color: #646464;
  line-height: 18px;
}

.order__id {
  margin-top: 6px;
  font-size: 11px;
  color: #bbbbbb;
  word-break: break-all;
}
</style>
