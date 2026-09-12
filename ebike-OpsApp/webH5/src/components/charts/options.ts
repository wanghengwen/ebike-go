import { echarts, type ChartOption } from '@/composables/useEcharts'
import { abbreviateAxisValue, numFormatInt } from '@/utils/numFormat'

/** 遗留大屏的默认调色板，顺序不能动，改了会影响已有截图与运营习惯。 */
export const DEFAULT_PALETTE = [
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

export const PIE_PALETTE = [
  '#2D66FF',
  '#00F2E2',
  '#AEF93D',
  '#DF21FF',
  '#4A5BED',
  '#E7AB02',
  '#DF392A',
]

const LEGEND_RICH = {
  a: { width: 100, height: 17, fontSize: 14, fontWeight: 400, color: '#333333' },
  b: { width: 70, height: 17, fontSize: 14, fontWeight: 600, color: '#333333' },
  c: { width: 58, height: 17, fontSize: 14, fontWeight: 600, color: '#333333' },
} as const

export interface LineSeries {
  name: string
  data: Array<number | null>
}

/** 多系列折线图，营收 / 运营大屏的趋势图都走这个。 */
export function buildLineOption(params: {
  xData: string[]
  series: LineSeries[]
  yAxisName?: string
  palette?: string[]
}): ChartOption {
  return {
    color: params.palette ?? DEFAULT_PALETTE,
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: { show: true, orient: 'horizontal', top: 'top', itemWidth: 18, itemHeight: 8, height: 30 },
    grid: { left: '15%' },
    xAxis: [{ type: 'category', axisTick: { show: false }, data: params.xData }],
    yAxis: [
      {
        type: 'value',
        name: params.yAxisName,
        axisLabel: { formatter: (value: number) => abbreviateAxisValue(value) },
      },
    ],
    series: params.series.map((s) => ({ ...s, type: 'line' })),
  }
}

/** 横向条形图（车辆状态 / 闲置 / 电量统计）。 */
export function buildHorizontalBarOption(params: {
  categories: string[]
  values: number[]
  gradient: [string, string]
}): ChartOption {
  return {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: '17%' },
    yAxis: [
      {
        type: 'category',
        axisTick: { show: false },
        data: params.categories,
        axisLabel: { color: '#333333', fontSize: 12 },
        axisLine: { lineStyle: { color: '#C8C8C8' } },
      },
    ],
    xAxis: [
      {
        type: 'value',
        axisLabel: { inside: false, color: '#333333' },
        axisTick: { show: false },
        axisLine: { show: false },
      },
    ],
    series: [
      {
        type: 'bar',
        barWidth: 20,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
            { offset: 0, color: params.gradient[0] },
            { offset: 1, color: params.gradient[1] },
          ]),
          opacity: 0.9,
          shadowColor: 'rgba(0,0,0,0.1)',
          shadowBlur: 3,
          shadowOffsetY: 3,
        },
        data: params.values,
        label: { show: true, position: 'right', color: '#333333' },
      },
    ],
  }
}

export interface PieSlice {
  name: string
  value: number
}

/**
 * 环形图 + 富文本图例（名称 / 数值 / 占比三列对齐）。
 * 遗留代码在每个组件里各抄了一份 `numFormatInt` 和 rich 配置，这里收敛成一处。
 */
export function buildDonutOption(params: {
  slices: PieSlice[]
  total: number
  palette: string[]
  legendTop?: string
  legendLeft?: string
  radius?: [string, string]
  center?: [string, string]
}): ChartOption {
  const percent = (value: number) =>
    params.total === 0 || value === 0
      ? '(0.00%)'
      : `(${((value / params.total) * 100).toFixed(2)}%)`

  return {
    color: params.palette,
    tooltip: { trigger: 'item', formatter: '{b}:<br/> {c}&emsp;({d}%)' },
    legend: {
      orient: 'vertical',
      icon: 'circle',
      left: params.legendLeft ?? '14%',
      top: params.legendTop ?? '82%',
      data: params.slices.map((s) => s.name),
      formatter: (name: string) => {
        const slice = params.slices.find((s) => s.name === name)
        const value = slice?.value ?? 0
        return [`{a| ${name}}`, `{b| ${numFormatInt(value)}}`, `{c|${percent(value)}}`].join('  ')
      },
      textStyle: { rich: LEGEND_RICH },
    },
    series: [
      {
        type: 'pie',
        radius: params.radius ?? ['50%', '70%'],
        center: params.center ?? ['50%', '50%'],
        avoidLabelOverlap: false,
        label: { show: false, position: 'center' },
        emphasis: { label: { show: true, fontSize: 16, fontWeight: 'bold' } },
        labelLine: { show: false },
        data: params.slices,
      },
    ],
  }
}
