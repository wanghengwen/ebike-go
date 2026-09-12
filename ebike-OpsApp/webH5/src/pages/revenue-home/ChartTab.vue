<script setup lang="ts">
import { computed, ref } from 'vue'
import DateFilter from './DateFilter.vue'
import RevenueChart from '@/components/charts/RevenueChart.vue'
import { usePermissionStore } from '@/stores/permission'
import { theme } from '@/utils/theme'
import { CHART_PRESETS, type RevenuePresetValue } from '@/composables/useRevenueDateRange'
import { PRIMARY_CHART_COUNT, REVENUE_CHARTS, type RevenueChartConfig } from './revenueCharts'
import type { RevenueChartData } from './useRevenueCharts'

const props = defineProps<{
  charts: Record<string, RevenueChartData | null>
  preset: RevenuePresetValue
  customStart: string
  customEnd: string
}>()

const emit = defineEmits<{
  'select-preset': [preset: RevenuePresetValue]
  'select-custom': [start: string, end: string]
  refresh: []
}>()

const permission = usePermissionStore()

const PRIMARY_KEYS = new Set(REVENUE_CHARTS.slice(0, PRIMARY_CHART_COUNT).map((it) => it.key))

const allowed = computed(() =>
  REVENUE_CHARTS.filter((config) => permission.has(config.permissionCode)),
)
const primary = computed(() => allowed.value.filter((config) => PRIMARY_KEYS.has(config.key)))
const extra = computed(() => allowed.value.filter((config) => !PRIMARY_KEYS.has(config.key)))

const activeKey = ref(REVENUE_CHARTS[0]!.key)
const expanded = ref(false)

/** 默认选中「应收订单」；没有该权限时退到第一张有权限的图，避免整页空白。 */
const activeConfig = computed<RevenueChartConfig | undefined>(
  () => allowed.value.find((config) => config.key === activeKey.value) ?? allowed.value[0],
)

const activeData = computed<RevenueChartData | null>(() =>
  activeConfig.value ? props.charts[activeConfig.value.key] ?? null : null,
)

function buttonStyle(key: string): Record<string, string> | undefined {
  if (key !== activeConfig.value?.key) return undefined
  return { color: theme.lightTxtColor, backgroundColor: theme.themeColor }
}
</script>

<template>
  <div>
    <DateFilter
      :presets="CHART_PRESETS"
      :preset="preset"
      :custom-start="customStart"
      :custom-end="customEnd"
      @select-preset="emit('select-preset', $event)"
      @select-custom="(start, end) => emit('select-custom', start, end)"
    />

    <div class="switcher">
      <span
        v-for="config in primary"
        :key="config.key"
        class="switcher__item"
        :style="buttonStyle(config.key)"
        @click="activeKey = config.key"
      >
        {{ config.label }}
      </span>
      <span
        v-if="extra.length > 0"
        class="switcher__more"
        :class="{ 'switcher__more--open': expanded }"
        :style="{ color: theme.themeColor }"
        @click="expanded = !expanded"
      >
        <van-icon name="arrow-down" />
      </span>
    </div>

    <transition name="fade">
      <div v-if="expanded" class="switcher">
        <span
          v-for="config in extra"
          :key="config.key"
          class="switcher__item"
          :style="buttonStyle(config.key)"
          @click="activeKey = config.key"
        >
          {{ config.label }}
        </span>
      </div>
    </transition>

    <RevenueChart
      v-if="activeConfig"
      :summary="activeData?.summary ?? []"
      :x-data="activeData?.xData ?? []"
      :series="activeData?.series ?? []"
      :legend="activeData?.legend"
      :palette="activeData?.palette"
      :series-type="activeConfig.seriesType"
      :unit="activeConfig.unit"
      :signed-axis="activeConfig.signedAxis"
    >
      <template #empty>
        <p class="refresh" :style="{ backgroundColor: theme.themeColor }" @click="emit('refresh')">
          刷新试试
        </p>
      </template>
    </RevenueChart>
  </div>
</template>

<style scoped>
.switcher {
  display: flex;
  flex-wrap: wrap;
}

.switcher__item {
  color: #282828;
  font-size: 12px;
  font-weight: 600;
  line-height: 17px;
  text-align: center;
  background: #efefef;
  border-radius: 4px;
  padding: 8px 9px;
  margin: 8px 0 0 16px;
}

.switcher__more {
  font-size: 14px;
  text-align: center;
  line-height: 33px;
  height: 33px;
  background: #efefef;
  border-radius: 4px;
  padding: 0 9px;
  margin: 8px 0 0 16px;
}

.switcher__more--open {
  transform: rotate(-180deg);
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.4s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.refresh {
  height: 35px;
  width: 100px;
  border-radius: 35px;
  line-height: 35px;
  text-align: center;
  font-size: 14px;
  color: #ffffff;
  margin: 0 auto 20px;
}
</style>
