import { getScreenRevenueConfig } from '@/api'

/**
 * 展示放大系数（网关字段 `displayCoefficient`）。
 *
 * 它的作用是把大屏上的金额按倍数放大，让数字好看，**不是**真实的成本折算。
 * 所以「结算统计」那一组必须用未放大的原值——那些数字要和代理商对账。
 *
 * 接口虽然叫 `getConfigByServiceId`，但遗留实现是不带任何入参调的，返回的是租户级系数。
 * 这里保持一致——带上 serviceIds 可能让后端改走按服务区过滤的分支，整屏金额都会跟着变。
 * 取不到时按 1（不放大）处理，配置接口挂了不能连带整屏数据一起不显示。
 */
export async function fetchDisplayFactor(): Promise<number> {
  try {
    const { success, data } = await getScreenRevenueConfig()
    return success ? data?.displayCoefficient || 1 : 1
  } catch {
    return 1
  }
}
