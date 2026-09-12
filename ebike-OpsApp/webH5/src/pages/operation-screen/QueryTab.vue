<script setup lang="ts">
import { ref } from 'vue'
import { showToast } from 'vant'
import MetricCard from './MetricCard.vue'
import OrderBreakdownPanel from './OrderBreakdownPanel.vue'
import TrendEntry from './TrendEntry.vue'
import { numFormat, numFormatInt } from '@/utils/numFormat'
import { theme } from '@/utils/theme'
import { DatePreset, formatDate, type DatePresetValue } from '@/composables/useDateRange'
import type { KeyMetrics, OrderBreakdown } from './useQueryData'
import type { PieSlice } from '@/components/charts/options'

const props = defineProps<{
  preset: DatePresetValue
  customStart: string
  customEnd: string
  metrics: KeyMetrics
  comparison: KeyMetrics | null
  orderCount: OrderBreakdown
  orderAmount: OrderBreakdown
  taskSlices: PieSlice[]
  taskTotal: number
}>()

export type TrendKey =
  | 'orderAmount'
  | 'ticketOrder'
  | 'averageAmount'
  | 'averageNum'
  | 'userGrow'
  | 'taskManage'
  | 'carsData'

const emit = defineEmits<{
  'select-preset': [preset: DatePresetValue]
  'select-custom': [start: string, end: string]
  'open-trend': [key: TrendKey]
}>()

const PRESETS: Array<[label: string, value: DatePresetValue]> = [
  ['今天', DatePreset.Today],
  ['近7天', DatePreset.Last7Days],
  ['近30天', DatePreset.Last30Days],
  ['累计', DatePreset.AllTime],
]

const METRICS: Array<[label: string, unit: string, key: keyof KeyMetrics, integer?: boolean]> = [
  ['订单收益', '元', 'orderAmount'],
  ['车均收益', '元', 'carCost'],
  ['订单量', '单', 'orderSum', true],
  ['车均单量', '单', 'orderNum'],
  ['均单时长', '分', 'orderTime'],
  ['均单里程', '公里', 'orderItinerary'],
]

const TRENDS: Array<[label: string, key: TrendKey]> = [
  ['订单收益', 'orderAmount'],
  ['订单量', 'ticketOrder'],
  ['车均收益', 'averageAmount'],
  ['车均单量', 'averageNum'],
  ['用户增长', 'userGrow'],
  ['任务管理', 'taskManage'],
  ['车辆统计', 'carsData'],
]

// 遗留把 `activeNames: [1]`（数字）配 `name="1"`（字符串），两者比不上，实际是全部折叠。
// 这里保持折叠的观感，折叠头本身已经展示了总订单量与总收益。
const activeCollapse = ref<string[]>([])
const showPicker = ref(false)
const pickingEnd = ref(false)
const today = new Date()
const pickedDate = ref<string[]>([
  String(today.getFullYear()),
  String(today.getMonth() + 1).padStart(2, '0'),
  String(today.getDate()).padStart(2, '0'),
])

function openPicker(isEnd: boolean): void {
  pickingEnd.value = isEnd
  showPicker.value = true
}

function confirmDate({ selectedValues }: { selectedValues: string[] }): void {
  const value = formatDate(selectedValues.join('/'))
  const start = pickingEnd.value ? props.customStart : value
  const end = pickingEnd.value ? value : props.customEnd
  showPicker.value = false
  if (start && end && new Date(start) > new Date(end)) {
    showToast('开始时间应小于或等于结束时间')
    return
  }
  emit('select-custom', start, end)
}
</script>

<template>
  <div>
    <div class="date-filter">
      <span class="date-filter__label">自定义时间:</span>
      <van-field
        :model-value="customStart"
        readonly
        center
        placeholder="开始时间"
        class="date-filter__field"
        @click="openPicker(false)"
      />
      <span class="date-filter__dash">—</span>
      <van-field
        :model-value="customEnd"
        readonly
        center
        placeholder="结束时间"
        class="date-filter__field"
        @click="openPicker(true)"
      />
    </div>

    <div class="presets">
      <div
        v-for="[label, value] in PRESETS"
        :key="value"
        class="presets__item"
        :style="preset === value ? { color: theme.themeColor } : undefined"
        @click="emit('select-preset', value)"
      >
        <div>{{ label }}</div>
        <div
          class="presets__underline"
          :style="preset === value ? { backgroundColor: theme.themeColor } : undefined"
        />
      </div>
    </div>

    <div class="metrics">
      <MetricCard
        v-for="[label, unit, key, integer] in METRICS"
        :key="key"
        :label="label"
        :unit="unit"
        :value="metrics[key]"
        :delta="comparison ? comparison[key] : null"
        :integer="integer"
      />
    </div>

    <div class="grey-block grey-block--gap" />
    <div class="section-title">数据统计</div>

    <van-collapse v-model="activeCollapse">
      <van-collapse-item name="1">
        <template #title>
          <div class="collapse-head">
            <div>
              总订单量:<span class="collapse-head__value" :style="{ color: theme.themeColor }">
                {{ numFormatInt(orderCount.total) }}
              </span>
            </div>
            <div>
              总收益:<span class="collapse-head__value" :style="{ color: theme.themeColor }">
                {{ numFormat(orderAmount.total) }}
              </span>
            </div>
          </div>
        </template>
        <OrderBreakdownPanel :count="orderCount" :amount="orderAmount" />
      </van-collapse-item>

      <van-collapse-item name="2">
        <template #title>
          <div class="collapse-head">
            <div>总运维量:</div>
            <div class="collapse-head__value" :style="{ color: theme.themeColor }">
              {{ numFormatInt(taskTotal) }}
            </div>
          </div>
        </template>
        <div v-for="slice in taskSlices" :key="slice.name" class="task-row">
          <span>{{ slice.name }}:</span>
          <span class="task-row__value">{{ numFormatInt(slice.value) }}</span>
        </div>
      </van-collapse-item>
    </van-collapse>

    <div class="grey-block grey-block--spaced" />
    <div class="section-title">趋势折线图</div>

    <TrendEntry
      v-for="[label, key] in TRENDS"
      :key="key"
      :label="label"
      @click="emit('open-trend', key)"
    />

    <van-popup v-model:show="showPicker" position="bottom">
      <van-date-picker
        v-model="pickedDate"
        @cancel="showPicker = false"
        @confirm="confirmDate"
      />
    </van-popup>
  </div>
</template>

<style scoped>
.date-filter {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  margin: 20px 10px 0;
}

.date-filter__label {
  font-size: 12px;
  font-weight: 600;
  color: #282828;
}

.date-filter__field {
  width: 120px;
  height: 36px;
  border: 1px solid #f2f2f2;
  background-color: #ffffff;
  padding: 8px;
}

.date-filter__dash {
  color: #a0a0a0;
}

.presets {
  display: flex;
  margin: 20px 10px 0 0;
}

.presets__item {
  font-size: 15px;
  font-weight: 600;
  line-height: 21px;
  color: #282828;
  margin-left: 16px;
}

.presets__underline {
  width: 12px;
  height: 4px;
  margin: 0 auto;
  border-radius: 2px;
}

.metrics {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  text-align: center;
  margin: 20px 16px 16px;
}

.grey-block--gap {
  margin-top: 8px;
}

.grey-block--spaced {
  margin-top: 34px;
}

/* 折叠面板展开区的底色，Vant 默认是白色 */
:deep(.van-collapse-item__content) {
  background: #fafafa;
}

.collapse-head {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  font-size: 14px;
  line-height: 20px;
  color: #282828;
}

.collapse-head__value {
  font-size: 16px;
  font-weight: bold;
  margin-left: 6px;
}

/* 分隔线在遗留里是独立的兄弟 div，通栏，不跟着行内容缩进 */
.task-row {
  display: flex;
  justify-content: space-between;
  margin-top: 13px;
  padding: 0 20px 12px 16px;
  font-size: 13px;
  border-bottom: 1px solid #dcdcdc;
}

.task-row:last-child {
  border-bottom: none;
}

.task-row__value {
  font-size: 15px;
  font-weight: bold;
}
</style>
