<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import BaseChart from '@/components/charts/BaseChart.vue'
import NoData from '@/components/NoData.vue'
import { buildLineOption } from '@/components/charts/options'
import { numFormat } from '@/utils/numFormat'
import { useTrendStore } from '@/stores/trend'

const router = useRouter()
const trend = useTrendStore()

const option = computed(() => {
  const payload = trend.payload
  if (!payload || payload.xData.length === 0) return null
  return buildLineOption({
    xData: payload.xData,
    series: payload.series,
    yAxisName: payload.unit,
    palette: payload.palette,
  })
})
</script>

<template>
  <div class="trend">
    <van-nav-bar :title="trend.payload?.title ?? ''" left-arrow @click-left="router.back()" />

    <template v-if="option">
      <div class="trend__total">
        <span>合计</span>
        <span class="trend__total-value">{{ numFormat(trend.payload?.total) }}</span>
      </div>
      <BaseChart :option="option" height="360px" />
    </template>

    <NoData v-else />
  </div>
</template>

<style scoped>
.trend {
  min-height: 100vh;
  background: #ffffff;
}

.trend__total {
  display: flex;
  align-items: baseline;
  gap: 8px;
  padding: 16px;
  font-size: 14px;
  color: #a0a0a0;
}

.trend__total-value {
  font-size: 20px;
  font-weight: bold;
  color: var(--ops-theme-color);
}
</style>
