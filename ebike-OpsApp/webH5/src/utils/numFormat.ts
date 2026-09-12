/**
 * 取代 Vue 2 的全局过滤器 `NumFormat` / `NumFormatInt`（Vue 3 已移除 filters）。
 * 行为与遗留实现逐字符对齐，包括无数据时的 `-- --` 占位。
 */
const PLACEHOLDER = '-- --'

function groupThousands(intPart: string): string {
  return intPart.replace(/(\d)(?=(?:\d{3})+$)/g, '$1,')
}

/** 千分位 + 强制两位小数。 */
export function numFormat(value: unknown): string {
  if (value === null || value === '' || value === undefined) return PLACEHOLDER
  const num = Number(value)
  if (!Number.isFinite(num)) return PLACEHOLDER

  const [intPart, decimalPart] = num.toFixed(2).split('.')
  return `${groupThousands(intPart!)}.${decimalPart}`
}

/** 千分位，保留原有小数位（无小数则不补 `.00`）。 */
export function numFormatInt(value: unknown): string {
  if (value === null || value === '' || value === undefined) return PLACEHOLDER

  const [intPart, decimalPart] = String(value).split('.')
  const grouped = groupThousands(intPart ?? '')
  if (decimalPart === undefined) return grouped
  return decimalPart.length === 1 ? `${grouped}.${decimalPart}0` : `${grouped}.${decimalPart}`
}

/** ECharts 轴标签用的万 / 千缩写。 */
export function abbreviateAxisValue(value: number): string {
  if (value >= 10000) return `${value / 10000}万`
  if (value > 1000) return `${value / 1000}千`
  return `${value}`
}
