import { ref } from 'vue'
import { listOrders, OrderPayState, type OrderItem, type OrderListQuery } from '@/api'

/** 查询维度。三者互斥，都不传就是「当前服务区的近期订单」。 */
export interface OrderListScope {
  userPin?: string
  carId?: string
  imei?: string
  /**
   * 只在没有 userPin / carId / imei 时才带。
   * 接口的 `serviceId` 是单个 `int64`，收不了多选；而按人/按车查时它本来就是多余的。
   */
  serviceId?: number
}

export interface UseOrderListOptions {
  scope: () => OrderListScope
  /** 只取最近 N 天。搜索首页是 3，人/车的历史列表不限。 */
  limitDays?: number
  /** 客户端累计条数上限。搜索首页是 200。 */
  limitMaxCount?: number
  pageSize?: number
}

const DEFAULT_PAGE_SIZE = 10

function dedupeById(rows: OrderItem[]): OrderItem[] {
  const seen = new Set<string>()
  return rows.filter((row) => {
    const key = row.id ?? ''
    if (!key || seen.has(key)) return !key
    seen.add(key)
    return true
  })
}

/**
 * 订单列表分页。
 *
 * 接口虽然返回 `count`，但遗留客户端从不读它——无筛选时下游会给一个写死的巨大默认值
 * （约 86000），当总数用会算出一堆不存在的页。所以这里也按「拉到空页就停」判断结束，
 * 另加一个客户端累计上限兜底。
 */
export function useOrderList(options: UseOrderListOptions) {
  const { limitDays, limitMaxCount, pageSize = DEFAULT_PAGE_SIZE } = options

  const items = ref<OrderItem[]>([])
  const loading = ref(false)
  const finished = ref(false)
  const refreshing = ref(false)
  const error = ref('')
  let pageNum = 1

  function recentRange(): [number, number] | undefined {
    if (!limitDays) return undefined
    const end = Date.now()
    const from = new Date(end)
    from.setDate(from.getDate() - limitDays)
    return [from.getTime(), end]
  }

  async function fetchPage(extra?: Partial<OrderListQuery>): Promise<OrderItem[]> {
    const range = recentRange()
    const { success, data } = await listOrders({
      pageNum,
      pageSize,
      ...options.scope(),
      ...(range ? { startTime: range } : {}),
      ...extra,
    })
    return success ? (data?.list ?? []) : []
  }

  async function loadMore(): Promise<void> {
    if (finished.value) return
    loading.value = true
    error.value = ''
    try {
      if (limitMaxCount && items.value.length >= limitMaxCount) {
        finished.value = true
        return
      }

      let batch = await fetchPage()

      // 骑行中的单还没结束时间，会被时间范围过滤掉，所以第一页额外单独捞一刀。
      // 遗留实现直接把两批拼起来，重复的单会出现两次；这里按 id 去重。
      if (pageNum === 1 && limitDays) {
        const riding = await fetchPage({ izPaid: OrderPayState.Riding })
        batch = dedupeById([...riding, ...batch])
      }

      if (batch.length === 0) {
        finished.value = true
        return
      }

      items.value = dedupeById([...items.value, ...batch])
      pageNum += 1

      if (batch.length < pageSize) finished.value = true
      if (limitMaxCount && items.value.length >= limitMaxCount) {
        items.value = items.value.slice(0, limitMaxCount)
        finished.value = true
      }
    } catch {
      // van-list 的 error 态要求 loading 归位，否则它不会再触发 onLoad。
      error.value = '加载失败，请下拉重试'
      finished.value = true
    } finally {
      loading.value = false
      refreshing.value = false
    }
  }

  async function refresh(): Promise<void> {
    pageNum = 1
    items.value = []
    finished.value = false
    refreshing.value = true
    await loadMore()
  }

  return { items, loading, finished, refreshing, error, loadMore, refresh }
}
