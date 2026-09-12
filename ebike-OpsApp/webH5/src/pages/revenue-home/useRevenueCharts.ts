import { ref } from 'vue'
import type { CardSellTimesResult, DailySeriesResult, SnakeRangeQuery } from '@/api'
import type { LineSeries } from '@/components/charts/options'
import type { TimeRange } from '@/composables/useDateRange'
import { numFormat } from '@/utils/numFormat'
import { fetchDisplayFactor } from './displayFactor'
import {
  CARD_PALETTE,
  REVENUE_CHARTS,
  type CardChartConfig,
  type DailyChartConfig,
  type RevenueChartConfig,
} from './revenueCharts'

export interface RevenueSummaryItem {
  label: string
  value: number
  color: string
  integer?: boolean
  notes?: string[]
}

export interface RevenueChartData {
  summary: RevenueSummaryItem[]
  xData: string[]
  series: LineSeries[]
  legend?: string[]
  palette?: string[]
}

function buildDaily(
  config: DailyChartConfig,
  data: DailySeriesResult,
  factor: number,
): RevenueChartData {
  const points = data.data_list ?? []
  const adjust = config.dailyAdjust?.() ?? null

  const scale = (value: unknown): number => {
    const scaled = Number(value) * (config.scaled ? factor : 1)
    return Number(scaled.toFixed(config.decimals))
  }
  const daily = (value: unknown): number => {
    const scaled = scale(value)
    return adjust ? adjust(scaled) : scaled
  }

  const series = config.fields.map(([key, name]) => ({
    name,
    data: points.map((point) => daily(point[key])),
  }))

  const summary = config.summary.map((field, index) => ({
    label: field.label,
    color: field.color,
    integer: field.integer,
    value:
      adjust && index === 0
        ? (series[0]?.data ?? []).reduce<number>((sum, value) => sum + (value ?? 0), 0)
        : scale(data[field.key]),
    notes: field.notes?.map(([key, suffix]) => `${numFormat(scale(data[key]))}${suffix}`),
  }))

  return {
    summary,
    xData: points.map((point) => String(point.date)),
    series,
    ...(config.palette ? { palette: config.palette } : {}),
  }
}

function buildCardTimes(config: CardChartConfig, data: CardSellTimesResult): RevenueChartData {
  const totals = data[config.sumKey] ?? []
  const points = data[config.timesKey] ?? []

  const series = totals.map((card) => ({
    name: card.name,
    // 某一天没有这张卡的售卖记录时留空，不能补 0，否则柱状图会多出一根零柱。
    data: points.map((point) => point.cardTime.find((it) => it.name === card.name)?.times ?? null),
  }))

  return {
    summary: totals.map((card, index) => ({
      label: config.summaryLabel(card.name),
      value: Number(card.times),
      color: CARD_PALETTE[index % CARD_PALETTE.length]!,
      integer: true,
    })),
    xData: points.map((point) => point.date),
    series,
    legend: series.map((item) => item.name),
    palette: CARD_PALETTE,
  }
}

async function load(
  config: RevenueChartConfig,
  query: SnakeRangeQuery,
  factor: number,
): Promise<RevenueChartData | null> {
  if (config.source === 'cardTimes') {
    const { success, data } = await config.fetch(query)
    return success && data ? buildCardTimes(config, data) : null
  }
  const { success, data } = await config.fetch(query)
  return success && data ? buildDaily(config, data, factor) : null
}

/**
 * 「图表」页签的取数。遗留实现每次切时间都把 14 张图全部重新请求，这里保持一致：
 * 切图表按钮时不再发请求，否则来回点会反复打接口。
 */
export function useRevenueCharts() {
  const charts = ref<Record<string, RevenueChartData | null>>({})

  async function loadAll(range: TimeRange, serviceIds: string[]): Promise<void> {
    const query: SnakeRangeQuery = { ...range, service_ids: serviceIds }
    const factor = await fetchDisplayFactor()

    await Promise.all(
      REVENUE_CHARTS.map(async (config) => {
        charts.value[config.key] = await load(config, query, factor)
      }),
    )
  }

  return { charts, loadAll }
}
