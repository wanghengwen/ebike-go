import { ref } from 'vue'
import {
  getOperationData,
  getOrderInfoArea,
  getOrderInfoTime,
  getTaskInfo,
  type OrderLinePoint,
  type SnakeRangeQuery,
} from '@/api'
import type { TimeRange } from '@/composables/useDateRange'
import type { PieSlice } from '@/components/charts/options'

export interface KeyMetrics {
  orderAmount: number
  carCost: number
  orderSum: number
  orderNum: number
  orderTime: number
  orderItinerary: number
}

const EMPTY_METRICS: KeyMetrics = {
  orderAmount: 0,
  carCost: 0,
  orderSum: 0,
  orderNum: 0,
  orderTime: 0,
  orderItinerary: 0,
}

export interface OrderBreakdown {
  general_order: number
  long_order: number
  short_order: number
  normal_order: number
  parking_zone: number
  out_of_service: number
  in_no_parking: number
  total: number
}

const EMPTY_BREAKDOWN: OrderBreakdown = {
  general_order: 0,
  long_order: 0,
  short_order: 0,
  normal_order: 0,
  parking_zone: 0,
  out_of_service: 0,
  in_no_parking: 0,
  total: 0,
}

/**
 * 环比百分比。分母为 0 时结果是 Infinity / NaN，交给 `numFormat` 渲染成 `-- --`，
 * 与遗留行为一致（不要在这里补 0，会把「从无到有」显示成 0% 增长）。
 */
function percentDelta(current: number, previous: number): number {
  return (current / previous) * 100 - 100
}

function deltas(current: KeyMetrics, previous: KeyMetrics): KeyMetrics {
  return {
    orderAmount: percentDelta(current.orderAmount, previous.orderAmount),
    carCost: percentDelta(current.carCost, previous.carCost),
    orderSum: percentDelta(current.orderSum, previous.orderSum),
    orderNum: percentDelta(current.orderNum, previous.orderNum),
    orderTime: percentDelta(current.orderTime, previous.orderTime),
    orderItinerary: percentDelta(current.orderItinerary, previous.orderItinerary),
  }
}

export function useQueryData() {
  const metrics = ref<KeyMetrics>({ ...EMPTY_METRICS })
  const comparison = ref<KeyMetrics | null>(null)

  const orderCount = ref<OrderBreakdown>({ ...EMPTY_BREAKDOWN })
  const orderAmount = ref<OrderBreakdown>({ ...EMPTY_BREAKDOWN })
  const orderLineByTime = ref<OrderLinePoint[]>([])
  const orderLineByZone = ref<OrderLinePoint[]>([])

  const taskSlices = ref<PieSlice[]>([])
  const taskTotal = ref(0)
  const taskLine = ref<Array<{ date: string; move: number; exchange_battery: number; fix: number }>>(
    [],
  )

  function query(range: TimeRange, serviceIds: string[]): SnakeRangeQuery {
    return { ...range, service_ids: serviceIds, type: 1 }
  }

  async function fetchMetrics(range: TimeRange, serviceIds: string[]): Promise<KeyMetrics> {
    const { success, data } = await getOperationData(query(range, serviceIds))
    if (!success || !data) return { ...EMPTY_METRICS }
    return {
      orderAmount: data.order_amount,
      carCost: data.average_order_cost_vehicle,
      orderSum: data.order_sum,
      orderNum: data.average_order_vehicle,
      orderTime: data.average_duration_order,
      orderItinerary: Number(Number(data.average_itinerary_order).toFixed(2)),
    }
  }

  async function loadMetrics(
    current: TimeRange,
    previous: TimeRange | null,
    serviceIds: string[],
  ): Promise<void> {
    metrics.value = await fetchMetrics(current, serviceIds)
    if (!previous) {
      comparison.value = null
      return
    }
    comparison.value = deltas(metrics.value, await fetchMetrics(previous, serviceIds))
  }

  async function loadOrderInfo(range: TimeRange, serviceIds: string[]): Promise<void> {
    const params = query(range, serviceIds)
    const [byTime, byZone] = await Promise.all([
      getOrderInfoTime(params),
      getOrderInfoArea(params),
    ])

    if (byTime.success && byTime.data) {
      const { order_pie, order_amount_pie, order_line } = byTime.data
      orderCount.value = {
        ...orderCount.value,
        general_order: order_pie.general_order,
        long_order: order_pie.long_order_time,
        short_order: order_pie.short_order_time,
        total: order_pie.total,
      }
      orderAmount.value = {
        ...orderAmount.value,
        general_order: order_amount_pie.general_order,
        long_order: order_amount_pie.long_order_time,
        short_order: order_amount_pie.short_order_time,
        total: order_amount_pie.total,
      }
      orderLineByTime.value = order_line ?? []
    }

    if (byZone.success && byZone.data) {
      const { order_pie, order_amount_pie, order_line } = byZone.data
      orderCount.value = {
        ...orderCount.value,
        normal_order: order_pie.normal_order,
        parking_zone: order_pie.is_parking_zone,
        out_of_service: order_pie.is_out_of_service_zone,
        in_no_parking: order_pie.is_in_no_parking_zone,
      }
      orderAmount.value = {
        ...orderAmount.value,
        normal_order: order_amount_pie.normal_order,
        parking_zone: order_amount_pie.is_parking_zone,
        out_of_service: order_amount_pie.is_out_of_service_zone,
        in_no_parking: order_amount_pie.is_in_no_parking_zone,
      }
      orderLineByZone.value = order_line ?? []
    }
  }

  async function loadTaskInfo(range: TimeRange, serviceIds: string[]): Promise<void> {
    const { success, data } = await getTaskInfo(query(range, serviceIds))
    if (!success || !data) return
    taskSlices.value = [
      { name: '挪车量', value: data.operation_pie.move },
      { name: '换电量', value: data.operation_pie.exchange_battery },
      { name: '维修量', value: data.operation_pie.fix },
    ]
    taskTotal.value = data.operation_pie.total
    taskLine.value = data.operation_line ?? []
  }

  async function loadAll(
    current: TimeRange,
    previous: TimeRange | null,
    serviceIds: string[],
  ): Promise<void> {
    await Promise.all([
      loadMetrics(current, previous, serviceIds),
      loadOrderInfo(current, serviceIds),
      loadTaskInfo(current, serviceIds),
    ])
  }

  return {
    metrics,
    comparison,
    orderCount,
    orderAmount,
    orderLineByTime,
    orderLineByZone,
    taskSlices,
    taskTotal,
    taskLine,
    loadAll,
  }
}
