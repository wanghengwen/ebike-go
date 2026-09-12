import * as echarts from 'echarts/core'
import { BarChart, LineChart, PieChart, TreeChart } from 'echarts/charts'
import {
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent,
} from 'echarts/components'
import { LabelLayout, UniversalTransition } from 'echarts/features'
import { CanvasRenderer } from 'echarts/renderers'

/**
 * 按需注册。两块大屏只用到条形 / 折线 / 环形 / 树四种图，
 * 全量引入 echarts 是 370KB gzip，按需后不到三分之一。
 * 新增图表类型时记得在这里补 `use()`，否则运行期会静默不渲染。
 */
echarts.use([
  BarChart,
  LineChart,
  PieChart,
  TreeChart,
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent,
  LabelLayout,
  UniversalTransition,
  CanvasRenderer,
])

export { echarts }
export type { ECharts, EChartsCoreOption } from 'echarts/core'
