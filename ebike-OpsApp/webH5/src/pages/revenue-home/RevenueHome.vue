<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { closeToast, showLoadingToast, showToast } from 'vant'
import MultiSelectPanel from '@/components/MultiSelectPanel.vue'
import DataTab from './DataTab.vue'
import ChartTab from './ChartTab.vue'
import { useRevenueData } from './useRevenueData'
import { useRevenueCharts } from './useRevenueCharts'
import { useServiceAreaStore } from '@/stores/serviceArea'
import { theme } from '@/utils/theme'
import {
  RevenuePreset,
  useRevenueDateRange,
  type RevenuePresetValue,
} from '@/composables/useRevenueDateRange'

defineOptions({ name: 'RevenueHome' })

const serviceArea = useServiceAreaStore()
const data = useRevenueData()
const charts = useRevenueCharts()

const dataRange = useRevenueDateRange(RevenuePreset.Today)
// 图表页签默认「七天」，与遗留一致。
const chartRange = useRevenueDateRange(RevenuePreset.Last7Days)

const activeTab = ref(0)
const refreshing = ref(false)
const showServicePanel = ref(false)

async function withLoading(task: () => Promise<unknown>): Promise<void> {
  if (serviceArea.isEmpty) return
  showLoadingToast({ duration: 0, forbidClick: true, message: '加载中...' })
  try {
    await task()
  } catch {
    showToast({ message: '获取数据失败', duration: 1000 })
  } finally {
    closeToast()
  }
}

const loadData = () =>
  withLoading(() => data.loadAll(dataRange.current.value, serviceArea.selectedIds))

const loadCharts = () =>
  withLoading(() => charts.loadAll(chartRange.current.value, serviceArea.selectedIds))

const loadAll = () =>
  withLoading(() =>
    Promise.all([
      data.loadAll(dataRange.current.value, serviceArea.selectedIds),
      charts.loadAll(chartRange.current.value, serviceArea.selectedIds),
    ]),
  )

async function bootstrap(): Promise<void> {
  try {
    await serviceArea.load()
  } catch {
    showToast('获取服务区数据失败')
    return
  }
  await loadAll()
}

onMounted(bootstrap)

async function onRefresh(): Promise<void> {
  refreshing.value = true
  await bootstrap()
  refreshing.value = false
}

/** 打开前先刷一次列表，避免后台新增 / 删除服务区后面板里还是旧的。 */
async function openServicePanel(): Promise<void> {
  try {
    await serviceArea.load()
  } catch {
    showToast('获取服务区失败，请检查网络设置')
    return
  }
  showServicePanel.value = true
}

function onServiceConfirm(ids: string[]): void {
  serviceArea.select(ids)
  showServicePanel.value = false
  void loadAll()
}

function onDataPreset(preset: RevenuePresetValue): void {
  if (dataRange.selectPreset(preset)) void loadData()
}

function onDataCustom(start: string, end: string): void {
  if (dataRange.setCustom(start, end) && start && end) void loadData()
}

function onChartPreset(preset: RevenuePresetValue): void {
  if (chartRange.selectPreset(preset)) void loadCharts()
}

function onChartCustom(start: string, end: string): void {
  if (chartRange.setCustom(start, end) && start && end) void loadCharts()
}
</script>

<template>
  <div class="page">
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <div class="service-bar">
        <span>已选择服务区:</span>
        <span class="service-bar__count" :style="{ color: theme.themeColor }">
          {{ serviceArea.selectedCount }}
        </span>
        <span
          class="service-bar__action"
          :style="{ color: theme.themeColor }"
          @click="openServicePanel"
        >
          选择<van-icon name="arrow-down" />
        </span>
      </div>

      <van-tabs v-model:active="activeTab" type="card" :color="theme.themeColor" class="main-tabs">
        <van-tab title="数据" />
        <van-tab title="图表" />
      </van-tabs>

      <!-- 营收大屏这条分隔线比运营大屏厚一倍 -->
      <div class="grey-line grey-line--thick" />

      <DataTab
        v-if="activeTab === 0"
        :groups="data.groups.value"
        :preset="dataRange.preset.value"
        :custom-start="dataRange.customStart.value"
        :custom-end="dataRange.customEnd.value"
        @select-preset="onDataPreset"
        @select-custom="onDataCustom"
      />

      <ChartTab
        v-else
        :charts="charts.charts.value"
        :preset="chartRange.preset.value"
        :custom-start="chartRange.customStart.value"
        :custom-end="chartRange.customEnd.value"
        @select-preset="onChartPreset"
        @select-custom="onChartCustom"
        @refresh="loadCharts"
      />
    </van-pull-refresh>

    <van-popup v-model:show="showServicePanel" position="right" class="service-popup">
      <MultiSelectPanel
        title="已选择服务区"
        empty-hint="请至少选择一个服务区"
        :items="serviceArea.list"
        :model-value="serviceArea.selectedIds"
        @confirm="onServiceConfirm"
        @cancel="showServicePanel = false"
      />
    </van-popup>
  </div>
</template>

<style scoped>
.service-bar {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  margin: 0 10px;
  padding-top: 10px;
  border-top: 1px solid #f2f2f2;
}

.service-bar__count {
  margin-right: 2px;
}

.service-bar__action {
  cursor: pointer;
}

.main-tabs {
  width: 270px;
  height: 37px;
  font-size: 14px;
  margin: 24px auto 0;
  border-radius: 6px;
}

.grey-line--thick {
  height: 4px;
}

.service-popup {
  height: 100%;
  width: 100%;
}
</style>
