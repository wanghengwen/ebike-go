/** 单独一个模块，避免页面 import 路由名时把整个 router（含页面 chunk）拉回来形成环。 */
export const RouteName = {
  OperationScreen: 'operation-screen',
  RevenueHome: 'revenue-home',
  TrendDetail: 'trend-detail',
} as const

export type RouteNameValue = (typeof RouteName)[keyof typeof RouteName]
