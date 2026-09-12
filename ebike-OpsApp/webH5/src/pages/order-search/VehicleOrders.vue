<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import OrderScopeView from './OrderScopeView.vue'

const route = useRoute()

const carId = computed(() => String(route.query.carId ?? '') || undefined)
const imei = computed(() => String(route.query.imei ?? '') || undefined)

/**
 * 车侧只出订单，不重建车辆信息卡——设防 / 电门 / 后轮锁 / 电量那一屏
 * OpsApp 已经有原生实现（`VehicleDetailSection` + `TencentMapView`），
 * 在 H5 再做一遍等于两套并行维护。
 */
const scope = computed(() => ({ carId: carId.value, imei: imei.value }))
const label = computed(() => carId.value ?? imei.value ?? '')
</script>

<template>
  <div class="page">
    <p class="page__title">{{ carId ? '车辆号' : '设备号' }} {{ label }}</p>
    <OrderScopeView :scope="scope" />
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f6f8;
  padding-bottom: 20px;
}

.page__title {
  padding: 14px 14px 0;
  font-size: 17px;
  font-weight: 600;
  color: #282828;
}
</style>
