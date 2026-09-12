import { useRouter } from 'vue-router'
import { closeToast, showLoadingToast, showToast } from 'vant'
import {
  getAverageData,
  getCarsData,
  getOrderMount,
  getTicketOrder,
  getUserGrow,
} from '@/api'
import type { TimeRange } from '@/composables/useDateRange'
import { RouteName } from '@/router/names'
import { useTrendStore } from '@/stores/trend'
import type { LineSeries } from '@/components/charts/options'

type Point = Record<string, string | number>

/** 把 `[{date, a, b}, ...]` 转成 ECharts 需要的 `{xData, series}`。 */
function toSeries(
  points: Point[],
  fields: Record<string, string>,
  dateKey = 'date',
): { xData: string[]; series: LineSeries[] } {
  return {
    xData: points.map((point) => String(point[dateKey] ?? '')),
    series: Object.entries(fields).map(([key, name]) => ({
      name,
      data: points.map((point) => Number(point[key] ?? 0)),
    })),
  }
}

/**
 * 「趋势折线图」分组下的七个入口：点一下拉数据、写入 trend store、跳详情页。
 * 任务管理不用再拉接口，数据在查询页签里已经有了。
 */
export function useTrends(range: () => TimeRange, serviceIds: () => string[]) {
  const router = useRouter()
  const trend = useTrendStore()

  function open(payload: Parameters<typeof trend.set>[0]): void {
    trend.set(payload)
    void router.push({ name: RouteName.TrendDetail })
  }

  async function withLoading<T>(task: () => Promise<T>): Promise<T | null> {
    showLoadingToast({ duration: 0, forbidClick: true, message: '加载中...' })
    try {
      return await task()
    } catch {
      showToast('获取数据失败')
      return null
    } finally {
      closeToast()
    }
  }

  const params = () => ({ ...range(), service_ids: serviceIds() })

  async function openOrderAmount(): Promise<void> {
    await withLoading(async () => {
      const { success, data } = await getOrderMount(params())
      if (!success || !data) return
      const { xData, series } = toSeries(data.date_list as unknown as Point[], {
        no_discount_order: '无优惠订单',
        discount_order: '折扣支付订单',
        free_order: '新用户免单',
        riding_card_order: '骑行卡订单',
      })
      open({
        title: '订单收益',
        unit: '单位:元',
        total:
          data.discount_order_sum +
          data.free_order_sum +
          data.no_discount_order_sum +
          data.riding_card_order_sum,
        xData,
        series,
        palette: ['#DD1CEF', '#A0EC30', '#5E5BDF', '#5BC4DF'],
      })
    })
  }

  async function openTicketOrder(): Promise<void> {
    await withLoading(async () => {
      const { success, data } = await getTicketOrder(params())
      if (!success || !data) return
      const { xData, series } = toSeries(data.data_list as unknown as Point[], {
        refund: '异议工单',
      })
      open({
        title: '订单量',
        unit: '单位:单',
        total: data.refund_sum,
        xData,
        series,
        palette: ['#5BC4DF'],
      })
    })
  }

  async function openAverage(kind: 'amount' | 'num'): Promise<void> {
    await withLoading(async () => {
      const { success, data } = await getAverageData(params())
      if (!success || !data) return
      const field = kind === 'amount' ? 'order_amount' : 'order_num'
      const label = kind === 'amount' ? '车均收益' : '车均单量'
      const { xData, series } = toSeries(data.data_list as unknown as Point[], { [field]: label })
      open({
        title: label,
        unit: kind === 'amount' ? '单位:元' : '单位:单',
        total: series[0]?.data.reduce<number>((sum, n) => sum + (n ?? 0), 0) ?? 0,
        xData,
        series,
        palette: ['#5BC4DF'],
      })
    })
  }

  async function openUserGrow(): Promise<void> {
    await withLoading(async () => {
      const { start_time, end_time } = range()
      const { success, data } = await getUserGrow({
        begin: start_time,
        end: end_time,
        serviceIds: serviceIds(),
      })
      if (!success || !data) return
      const { xData, series } = toSeries(data.new_data as unknown as Point[], {
        total: '总用户数',
        new: '新增用户',
        active: '活跃用户',
      })
      open({
        title: '用户增长',
        unit: '单位:人',
        total: data.total_user,
        xData,
        series,
        palette: ['#5BC4DF', '#DD1CEF', '#5E5BDF'],
      })
    })
  }

  async function openCarsData(): Promise<void> {
    await withLoading(async () => {
      const { start_time, end_time } = range()
      const { success, data } = await getCarsData({
        startTime: start_time,
        endTime: end_time,
        serviceIds: serviceIds(),
      })
      if (!success || !data?.[0]) return
      const { xData, series } = toSeries(
        data[0] as unknown as Point[],
        { count: '运营车辆' },
        'time',
      )
      open({
        title: '车辆统计',
        unit: '单位:辆',
        total: series[0]?.data.at(-1) ?? 0,
        xData,
        series,
        palette: ['#5BC4DF'],
      })
    })
  }

  function openTaskManage(
    points: Array<{ date: string; move: number; exchange_battery: number; fix: number }>,
    total: number,
  ): void {
    const { xData, series } = toSeries(points as unknown as Point[], {
      move: '挪车量',
      exchange_battery: '换电量',
      fix: '维修量',
    })
    open({
      title: '任务管理',
      unit: '单位:次',
      total,
      xData,
      series,
      palette: ['#DD1CEF', '#5BC4DF', '#5E5BDF'],
    })
  }

  return {
    openOrderAmount,
    openTicketOrder,
    openAverage,
    openUserGrow,
    openCarsData,
    openTaskManage,
  }
}
