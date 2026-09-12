import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { LineSeries } from '@/components/charts/options'

export interface TrendPayload {
  title: string
  unit: string
  total: number
  xData: string[]
  series: LineSeries[]
  palette?: string[]
}

/**
 * 趋势折线详情页的数据载体。
 *
 * 遗留实现把整个数据对象 `JSON.stringify` 后塞进路由 path
 * （`/orderAmount/${encodeURIComponent(json)}`），30 天的多系列数据能把 URL 顶到几十 KB；
 * 而且那些路由压根没在 router 里注册过，点了只会静默失败。改成走 store 传递。
 */
export const useTrendStore = defineStore('trend', () => {
  const payload = ref<TrendPayload | null>(null)

  function set(next: TrendPayload): void {
    payload.value = next
  }

  function clear(): void {
    payload.value = null
  }

  return { payload, set, clear }
})
