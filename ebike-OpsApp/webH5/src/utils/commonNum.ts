/** 营收大屏中调度费向下取整到 0 或 5 结尾：163 → 160，168 → 165。 */
export function revenueDataPenaltyNum(num: number): number {
  if (num <= 0) return num
  const tens = Math.trunc(num / 10)
  const ones = num % 10
  return ones <= 4 ? tens * 10 : tens * 10 + 5
}

/** 运营大屏是否需要对调度费做上述取整（仅未指定 tenantId 的旧链接需要）。 */
export function isDealPenaltyNum(): boolean {
  return /tenantId=()($|&|\?){1}/i.test(window.location.href)
}
