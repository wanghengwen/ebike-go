<script setup lang="ts">
import { computed } from 'vue'
import BaseChart from './BaseChart.vue'
import { buildDonutOption, type PieSlice } from './options'

const props = withDefaults(
  defineProps<{
    title: string
    total: number
    unit?: string
    slices: PieSlice[]
    palette: string[]
    height?: string
    width?: string
  }>(),
  { unit: '', height: '450px', width: '450px' },
)

const option = computed(() =>
  buildDonutOption({
    slices: props.slices,
    total: props.total,
    palette: props.palette,
    legendTop: '82%',
    radius: ['50%', '70%'],
  }),
)
</script>

<template>
  <div>
    <div class="chart-title">
      <span>{{ title }}:</span>
      <span class="chart-title__value">{{ total || '0' }}</span>
      <span>{{ unit }}</span>
    </div>
    <div class="chart-body">
      <BaseChart :option="option" :height="height" :width="width" />
    </div>
  </div>
</template>

<style scoped>
.chart-title {
  height: 25px;
  font-size: 14px;
  font-weight: 550;
  color: #333333;
  line-height: 17px;
  margin: 24px 0 2px 16px;
}

.chart-title__value {
  font-size: 20px;
  margin: 0 2px 0 5px;
}

.chart-body {
  display: flex;
  align-items: center;
  justify-content: center;
  margin: -75px 10px 20px 10px;
}
</style>
