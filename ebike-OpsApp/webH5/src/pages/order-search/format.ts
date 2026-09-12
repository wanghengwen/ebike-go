import { OrderPayState } from '@/api'

const DASH = '--'

/** 分 → 元。接口金额一律是分，界面一律是元，转换只在这里发生。 */
export function yuan(fen: number | null | undefined): string {
  if (fen === null || fen === undefined) return DASH
  return (fen / 100).toFixed(2)
}

/**
 * 骑行时长 → 「x天x小时x分x秒」。
 *
 * 接口给的是毫秒，但序列化成了字符串。遗留实现按位数猜单位：超过 10 位当毫秒先除 1000，
 * 否则当秒。这个约定保留——历史数据里确实两种都存在。
 */
export function duration(raw: string | number | null | undefined): string {
  if (raw === null || raw === undefined || raw === '') return DASH
  const value = Number(raw)
  if (!Number.isFinite(value) || value < 0) return DASH

  const totalSeconds = Math.floor(String(Math.trunc(value)).length > 10 ? value / 1000 : value)
  const days = Math.floor(totalSeconds / 86400)
  const hours = Math.floor((totalSeconds % 86400) / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  const parts: string[] = []
  if (days > 0) parts.push(`${days}天`)
  if (hours > 0) parts.push(`${hours}小时`)
  if (minutes > 0) parts.push(`${minutes}分`)
  // 不满一分钟时也要出个数，否则界面上会是空白。
  if (seconds > 0 || parts.length === 0) parts.push(`${seconds}秒`)
  return parts.join('')
}

/** 里程。接口是米，不足 1 公里按米显示，与遗留一致。 */
export function distance(meters: number | null | undefined): string {
  if (meters === null || meters === undefined) return DASH
  if (meters < 1000) return `${meters}米`
  return `${(meters / 1000).toFixed(2)}公里`
}

const PAY_STATE_LABELS: Record<number, string> = {
  [OrderPayState.Riding]: '骑行中',
  [OrderPayState.Frozen]: '支付中',
  [OrderPayState.ToPay]: '待支付',
  [OrderPayState.Paid]: '已支付',
}

export function payStateLabel(izPaid: number | null | undefined): string {
  if (izPaid === null || izPaid === undefined) return DASH
  return PAY_STATE_LABELS[izPaid] ?? DASH
}

/** 只有已支付算终态，其余都当未结清处理（决定卡片上的强调色）。 */
export function isSettled(izPaid: number | null | undefined): boolean {
  return izPaid === OrderPayState.Paid
}

const RIDING_STATE_LABELS: Record<number, string> = {
  1: '已注册',
  2: '已实名',
  3: '可用车',
  4: '预约中',
  5: '临停中',
  6: '骑行中',
  7: '待支付',
}

/** 用户的 `ridingState`。注意车辆也有个同名字段但取值不同，别混用。 */
export function ridingStateLabel(state: number | null | undefined): string {
  if (state === null || state === undefined) return DASH
  return RIDING_STATE_LABELS[state] ?? DASH
}

/** 用户详情的 `createdAt` 等是带 `T` 的 ISO 串，订单时间则是空格分隔，这里只处理前者。 */
export function isoDateTime(value: string | null | undefined): string {
  const text = (value ?? '').trim()
  if (!text) return DASH
  return text.replace('T', ' ')
}

/** 空字符串和 `-` 都当无值——接口这两种都会出现。 */
export function orDash(value: string | null | undefined): string {
  const text = (value ?? '').trim()
  return text && text !== '-' ? text : DASH
}
