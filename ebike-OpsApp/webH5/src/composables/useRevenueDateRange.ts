import { computed, ref } from 'vue'
import dayjs, { type Dayjs } from 'dayjs'
import type { TimeRange } from './useDateRange'

/** 累计口径的起点，与遗留写死的 2019-01-01 一致。 */
const EPOCH = '2019-01-01'

/**
 * 营收大屏的日期口径，与运营大屏（今天 / 近 7 天 / 近 30 天 / 累计）不同：
 * 「数据」页签是今日 / 昨日 / 七天 / 本月 / 上个月 / 全部，「图表」页签只取前面的一个子集。
 * 营收侧没有同期环比，所以这里不产出 previous 区间。
 */
export const RevenuePreset = {
  Today: 0,
  Yesterday: 1,
  Last7Days: 2,
  ThisMonth: 3,
  LastMonth: 4,
  AllTime: 5,
  Custom: 10,
} as const

export type RevenuePresetValue = (typeof RevenuePreset)[keyof typeof RevenuePreset]

export const DATA_PRESETS: Array<[label: string, value: RevenuePresetValue]> = [
  ['今日', RevenuePreset.Today],
  ['昨日', RevenuePreset.Yesterday],
  ['七天', RevenuePreset.Last7Days],
  ['本月', RevenuePreset.ThisMonth],
  ['上个月', RevenuePreset.LastMonth],
  ['全部', RevenuePreset.AllTime],
]

export const CHART_PRESETS: Array<[label: string, value: RevenuePresetValue]> = [
  ['今日', RevenuePreset.Today],
  ['七天', RevenuePreset.Last7Days],
  ['本月', RevenuePreset.ThisMonth],
  ['上个月', RevenuePreset.LastMonth],
]

const dayStart = (d: Dayjs) => d.startOf('day').valueOf()
const dayEnd = (d: Dayjs) => d.endOf('day').valueOf()

function presetRange(preset: RevenuePresetValue): TimeRange {
  const today = dayjs()
  switch (preset) {
    case RevenuePreset.Yesterday: {
      const yesterday = today.subtract(1, 'day')
      return { start_time: dayStart(yesterday), end_time: dayEnd(yesterday) }
    }
    case RevenuePreset.Last7Days:
      return { start_time: dayStart(today.subtract(6, 'day')), end_time: dayEnd(today) }
    case RevenuePreset.ThisMonth:
      return { start_time: dayStart(today.startOf('month')), end_time: dayEnd(today) }
    case RevenuePreset.LastMonth: {
      const lastMonth = today.subtract(1, 'month')
      return {
        start_time: dayStart(lastMonth.startOf('month')),
        end_time: dayEnd(lastMonth.endOf('month')),
      }
    }
    case RevenuePreset.AllTime:
      return { start_time: dayStart(dayjs(EPOCH)), end_time: dayEnd(today) }
    default:
      return { start_time: dayStart(today), end_time: dayEnd(today) }
  }
}

export function useRevenueDateRange(initial: RevenuePresetValue) {
  const preset = ref<RevenuePresetValue>(initial)
  const customStart = ref('')
  const customEnd = ref('')

  const current = computed<TimeRange>(() => {
    if (preset.value === RevenuePreset.Custom && customStart.value && customEnd.value) {
      return {
        start_time: dayStart(dayjs(customStart.value)),
        end_time: dayEnd(dayjs(customEnd.value)),
      }
    }
    return presetRange(preset.value)
  })

  /** 返回 false 表示没变化，调用方据此跳过重新取数。 */
  function selectPreset(next: RevenuePresetValue): boolean {
    if (next === preset.value) return false
    customStart.value = ''
    customEnd.value = ''
    preset.value = next
    return true
  }

  /** 返回 false 表示区间非法（开始晚于结束），调用方负责提示。 */
  function setCustom(start: string, end: string): boolean {
    if (start && end && dayjs(start).isAfter(dayjs(end))) return false
    customStart.value = start
    customEnd.value = end
    if (start && end) preset.value = RevenuePreset.Custom
    return true
  }

  return { preset, customStart, customEnd, current, selectPreset, setCustom }
}
