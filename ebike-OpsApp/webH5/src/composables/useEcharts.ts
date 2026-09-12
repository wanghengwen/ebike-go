import { onBeforeUnmount, onMounted, shallowRef, watch, type Ref } from 'vue'
import { echarts, type ECharts, type EChartsCoreOption } from './echartsCore'

export type ChartOption = EChartsCoreOption

export interface UseEchartsOptions {
  /** 图表首帧渲染完成，页面用它关掉全局 loading。 */
  onFinished?: () => void
}

/**
 * 托管一个 ECharts 实例的完整生命周期。
 *
 * 遗留组件在 `mounted` 和 `updated` 里都调 `echarts.init(this.$refs.myEchart)`，
 * 每次更新都会新建实例而不销毁旧的，滚动几次大屏就能堆出十几个 canvas。
 * 这里改成：只 init 一次，后续走 `setOption`，并在卸载时 dispose。
 */
export function useEcharts(
  container: Ref<HTMLElement | null | undefined>,
  option: Ref<ChartOption | null | undefined>,
  { onFinished }: UseEchartsOptions = {},
) {
  const chart = shallowRef<ECharts | null>(null)
  let observer: ResizeObserver | null = null

  function render(): void {
    if (!chart.value || !option.value) return
    // notMerge = true：系列数量随筛选条件变化，合并会残留上一次的 series。
    chart.value.setOption(option.value, true)
  }

  onMounted(() => {
    if (!container.value) return
    chart.value = echarts.init(container.value)
    if (onFinished) chart.value.on('finished', onFinished)
    render()

    observer = new ResizeObserver(() => chart.value?.resize())
    observer.observe(container.value)
  })

  watch(option, render, { deep: true })

  onBeforeUnmount(() => {
    observer?.disconnect()
    observer = null
    chart.value?.dispose()
    chart.value = null
  })

  return { chart }
}

export { echarts }
