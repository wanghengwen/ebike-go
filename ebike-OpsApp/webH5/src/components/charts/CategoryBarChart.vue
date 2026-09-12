<script setup lang="ts">
import { computed } from 'vue'
import BaseChart from './BaseChart.vue'
import { buildHorizontalBarOption } from './options'

const props = defineProps<{
  title: string
  total: number
  unit: string
  categories: string[]
  values: number[]
  gradient: [string, string]
}>()

const option = computed(() =>
  buildHorizontalBarOption({
    categories: props.categories,
    values: props.values,
    gradient: props.gradient,
  }),
)
</script>

<template>
  <div>
    <div class="chart-title">
      <span>{{ title }}:</span>
      <span class="chart-title__value">{{ total || '--' }}</span>
      <span>{{ unit }}</span>
    </div>
    <div class="chart-body">
      <BaseChart :option="option" height="350px" width="350px" />
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
  margin: 16px 0 2px 16px;
}

.chart-title__value {
  font-size: 20px;
  margin: 0 2px 0 5px;
}

.chart-body {
  display: flex;
  justify-content: center;
  margin: -50px 10px 20px 10px;
}
</style>
