/**
 * 大屏接口的出入参类型。字段名沿用网关的下划线 / 驼峰混用现状（同一批接口两种风格都有），
 * 不在前端做重命名，避免和 `pc` 后台、Kotlin 侧的口径再分叉。
 */

// --- 通用查询参数 ---------------------------------------------------------

/** `/ebike_visual/*` 一族使用下划线风格。 */
export interface SnakeRangeQuery {
  start_time: number
  end_time: number
  service_ids: string[]
  type?: number
}

/** `/business/user/*` 一族使用 begin / end。 */
export interface UserGrowQuery {
  begin: number
  end: number
  serviceIds: string[]
}

/** `/business/operatingBigScreen/*` 一族使用驼峰。 */
export interface CarsDataQuery {
  startTime: number
  endTime: number
  serviceIds: string[]
}

// --- 运营大屏 -------------------------------------------------------------

/** `/business/paas/device/list` 单车条目，只声明大屏用到的字段。 */
export interface DeviceItem {
  ridingState?: number
  /** 服务端返回数组（一辆车可同时处于多个运维状态）。 */
  operationState: number[]
  alarmState?: number[]
  restBattery?: number
  lockTime?: number | string
  unlockTime?: number | string
}

export interface MemberStatistic {
  info: {
    authentication: number
    noAuthentication: number
    nonMember: number
    historicalMember: number
    member: number
    validMember: number
    invalidMember: number
    freeUser: number
    deposit: number
    career: number
    memberCard: number
  }
}

export interface AgeStatistic {
  user_treemap: {
    have_riding_qualification: { total: number }
    no_riding_qualification: { total: number }
  }
}

export interface OperationDataInfo {
  average_order_cost_vehicle: number
  average_duration_order: number
  average_order_vehicle: number
  average_itinerary_order: number
  order_sum: number
  order_amount: number
}

export interface OrderPie {
  general_order: number
  long_order_time: number
  short_order_time: number
  normal_order: number
  is_parking_zone: number
  is_out_of_service_zone: number
  is_in_no_parking_zone: number
  total: number
}

export interface OrderLinePoint {
  date: string
  general_order: number
  long_order_time: number
  short_order_time: number
  normal_order: number
  is_parking_zone: number
  is_out_of_service_zone: number
  is_in_no_parking_zone: number
}

export interface OrderInfoResult {
  order_pie: OrderPie
  order_amount_pie: OrderPie
  order_line: OrderLinePoint[]
}

export interface TaskInfoResult {
  operation_pie: { move: number; exchange_battery: number; fix: number; total: number }
  operation_line: Array<{ date: string; move: number; exchange_battery: number; fix: number }>
}

export interface OrderMountResult {
  discount_order_sum: number
  free_order_sum: number
  no_discount_order_sum: number
  riding_card_order_sum: number
  date_list: Array<{
    date: string
    no_discount_order: number
    discount_order: number
    free_order: number
    riding_card_order: number
  }>
}

export interface AverageDataResult {
  data_list: Array<{ date: string; order_num: number; order_amount: number }>
}

export interface TicketOrderResult {
  refund_sum: number
  data_list: Array<{ date: string; refund: number }>
}

export interface UserGrowResult {
  total_user: number
  new_user: number
  active_user: number
  new_data: Array<{ date: string; total: number; new: number; active: number }>
}

/** 注意：服务端把折线包了一层数组，`data[0]` 才是点位列表。 */
export type CarsDataResult = Array<Array<{ time: string; count: number }>>

// --- 营收大屏 -------------------------------------------------------------

/** 金额 + 笔数的通用汇总，营收大屏大部分卡片都是这个形状。 */
export interface AmountSummary {
  amount?: number
  num?: number
  total?: number
  [key: string]: number | undefined
}

/** 按天的折线点位，键名随接口而异。 */
export interface DailyPoint {
  date: string
  [key: string]: string | number
}

/** 售卖次数类接口的卡片条目，`name` 由后台按「金额 + 天数」拼好下发，前端不再拼。 */
export interface CardTimesEntry {
  name: string
  times?: number
  money?: number
  days?: number
}

export interface CardTimesPoint {
  date: string
  cardTime: CardTimesEntry[]
}

/** `*_card/sell_times` 三个接口：`*_sum` 是卡片维度合计，`*_times` 是按天 × 卡片的点位。 */
export interface CardSellTimesResult {
  deposit_card_sum?: CardTimesEntry[]
  deposit_card_times?: CardTimesPoint[]
  riding_card_sum?: CardTimesEntry[]
  riding_card_times?: CardTimesPoint[]
  favorable_card_sum?: CardTimesEntry[]
  favorable_card_times?: CardTimesPoint[]
}

export interface DailySeriesResult {
  data_list?: DailyPoint[]
  date_list?: DailyPoint[]
  [key: string]: unknown
}

export interface ScreenRevenueConfig {
  /** 展示放大系数，大屏按它把金额放大后展示；对账口径的数字不适用。 */
  displayCoefficient?: number
  [key: string]: unknown
}
