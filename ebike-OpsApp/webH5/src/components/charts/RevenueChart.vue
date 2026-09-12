<script setup lang="ts">
import { computed } from 'vue'
import BaseChart from './BaseChart.vue'
import NoData from '@/components/NoData.vue'
import type { ChartOption } from '@/composables/useEcharts'
import type { LineSeries } from './options'
import { numFormat, numFormatInt } from '@/utils/numFormat'

/**
 * 营收大屏的通用图表。
 *
 * 遗留工程为这块屏写了 14 个组件（`paymentChart` / `memberCardChart` / `depositChart` …），
 * 它们之间只有三处差异：顶部汇总行、Y 轴单位、以及是否需要处理负值刻度。
 * 这里合并成一个组件 + 一张配置表（`revenueCharts.ts`）。
 */

export interface SummaryItem {
  label: string
  value: number
  color: string
  /** 次数类指标不补两位小数。 */
  integer?: boolean
  /** 主值下方的小字拆分行（实收订单的结算充值 / 结算赠送）。 */
  notes?: string[]
}

const props = withDefaults(
  defineProps<{
    summary: SummaryItem[]
    xData: string[]
    series: LineSeries[]
    /** 系列类型，柱状与折线在这块屏上混用。 */
    seriesType?: 'bar' | 'line'
    unit: string
    legend?: string[]
    /** 存在负值时刻度需要保留符号（退款、坏账类图表）。 */
    signedAxis?: boolean
    palette?: string[]
  }>(),
  { seriesType: 'bar', signedAxis: false },
)

/**
 * 金额 / 数量类图表的配色。售卖次数类图表和用户充值总金额用的是另一套
 * （紫绿对调、多一个粉色），由调用方通过 `palette` 传入。
 */
const PALETTE = [
  '#3868FF',
  '#FF9C80',
  '#1DB996',
  '#993A9E',
  '#91c7ae',
  '#749f83',
  '#ca8622',
  '#bda29a',
  '#6e7074',
  '#546570',
  '#c4ccd3',
]

const hasData = computed(() => props.xData.length > 0 && props.series.length > 0)

function axisLabel(value: number): string {
  const abs = Math.abs(value)
  const sign = value < 0 ? '-' : ''
  if (abs >= 10000) return `${sign}${abs / 10000}万`
  if (props.signedAxis ? abs >= 1000 : abs > 1000) return `${sign}${abs / 1000}千`
  return props.signedAxis ? `${sign}${abs}` : `${value}`
}

const option = computed<ChartOption | null>(() => {
  if (!hasData.value) return null
  return {
    color: props.palette ?? PALETTE,
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: {
      show: true,
      orient: 'horizontal',
      top: 0,
      itemWidth: 12,
      itemHeight: 12,
      ...(props.legend ? { data: props.legend } : {}),
    },
    // 遗留 14 个组件的 grid / legend 尺寸各不相同（20%~27% top、10%~15% left），
    // 是复制粘贴漂移出来的，这里统一取出现次数最多的一组。
    grid: { top: '20%', left: '15%' },
    xAxis: [
      {
        type: 'category',
        axisTick: { show: false },
        data: props.xData,
        axisLine: { lineStyle: { color: '#C8C8C8' } },
        axisLabel: { color: '#666666' },
      },
    ],
    yAxis: [
      {
        type: 'value',
        name: props.unit,
        nameTextStyle: { color: '#666666' },
        axisLabel: { color: '#666666', formatter: axisLabel },
        axisLine: { lineStyle: { color: '#C8C8C8' } },
      },
    ],
    series: props.series.map((s) => ({ ...s, type: props.seriesType })),
  }
})
</script>

<template>
  <div>
    <template v-if="hasData">
      <div class="summary">
        <div v-for="item in summary" :key="item.label" class="summary__item">
          <p class="summary__value" :style="{ color: item.color }">
            {{ item.integer ? numFormatInt(item.value) : numFormat(item.value) }}
          </p>
          <p class="summary__label">{{ item.label }}</p>
          <p v-for="note in item.notes" :key="note" class="summary__note">{{ note }}</p>
        </div>
      </div>
      <div class="chart-body">
        <BaseChart :option="option" height="350px" width="350px" />
      </div>
    </template>

    <NoData v-else>
      <slot name="empty" />
    </NoData>
  </div>
</template>

<style scoped>
.summary {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  margin: 35px 16px;
}

.summary__item {
  width: 45%;
  margin-bottom: 10px;
  text-align: center;
}

.summary__value {
  font-size: 16px;
  font-weight: 600;
}

.summary__label {
  margin-top: 5px;
  font-size: 12px;
  font-weight: 400;
  color: #646464;
}

.summary__note {
  margin-top: 2px;
  font-size: 11px;
  font-weight: 400;
  color: #999999;
}

.chart-body {
  display: flex;
  justify-content: center;
  margin: 20px 10px;
}
</style>
