import { computed, ref } from 'vue'
import dayjs, { type Dayjs } from 'dayjs'

/** 累计口径的起点，与遗留写死的 2019-01-01 一致。 */
const EPOCH = '2019-01-01'

export const DatePreset = {
  Today: 0,
  Last7Days: 1,
  Last30Days: 2,
  AllTime: 4,
  Custom: 10,
} as const

export type DatePresetValue = (typeof DatePreset)[keyof typeof DatePreset]

export interface TimeRange {
  start_time: number
  end_time: number
}

const dayStart = (d: Dayjs) => d.startOf('day').valueOf()
const dayEnd = (d: Dayjs) => d.endOf('day').valueOf()

function presetRange(preset: DatePresetValue): { current: TimeRange; previous: TimeRange | null } {
  const today = dayjs()
  switch (preset) {
    case DatePreset.Today:
      return {
        current: { start_time: dayStart(today), end_time: dayEnd(today) },
        previous: {
          start_time: dayStart(today.subtract(1, 'day')),
          end_time: dayEnd(today.subtract(1, 'day')),
        },
      }
    case DatePreset.Last7Days:
      return {
        current: { start_time: dayStart(today.subtract(6, 'day')), end_time: dayEnd(today) },
        previous: {
          start_time: dayStart(today.subtract(13, 'day')),
          end_time: dayEnd(today.subtract(7, 'day')),
        },
      }
    case DatePreset.Last30Days:
      return {
        current: { start_time: dayStart(today.subtract(29, 'day')), end_time: dayEnd(today) },
        previous: {
          start_time: dayStart(today.subtract(59, 'day')),
          end_time: dayEnd(today.subtract(30, 'day')),
        },
      }
    default:
      // 累计没有同期可比，页面据此隐藏环比行。
      return {
        current: { start_time: dayStart(dayjs(EPOCH)), end_time: dayEnd(today) },
        previous: null,
      }
  }
}

/**
 * 大屏顶部的「今天 / 近 7 天 / 近 30 天 / 累计 + 自定义区间」筛选。
 * 同期区间总是取与当前区间等长、且紧邻其前的一段。
 */
export function useDateRange(initial: DatePresetValue = DatePreset.Today) {
  const preset = ref<DatePresetValue>(initial)
  const customStart = ref('')
  const customEnd = ref('')

  const customRange = computed<{ current: TimeRange; previous: TimeRange } | null>(() => {
    if (!customStart.value || !customEnd.value) return null
    const start = dayjs(customStart.value)
    const end = dayjs(customEnd.value)
    const span = end.diff(start, 'day')
    return {
      current: { start_time: dayStart(start), end_time: dayEnd(end) },
      previous: {
        start_time: dayStart(start.subtract(span + 1, 'day')),
        end_time: dayEnd(start.subtract(1, 'day')),
      },
    }
  })

  const range = computed(() =>
    preset.value === DatePreset.Custom && customRange.value
      ? customRange.value
      : presetRange(preset.value),
  )

  const current = computed(() => range.value.current)
  const previous = computed(() => range.value.previous)
  const hasComparison = computed(() => previous.value !== null)

  function selectPreset(next: DatePresetValue): boolean {
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
    if (start && end) preset.value = DatePreset.Custom
    return true
  }

  return {
    preset,
    customStart,
    customEnd,
    current,
    previous,
    hasComparison,
    selectPreset,
    setCustom,
  }
}

export function formatDate(value: Date | string | number): string {
  return dayjs(value).format('YYYY/MM/DD')
}
