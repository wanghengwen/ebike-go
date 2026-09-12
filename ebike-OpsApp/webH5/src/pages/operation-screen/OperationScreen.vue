<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { closeToast, showLoadingToast, showToast } from 'vant'
import MultiSelectPanel from '@/components/MultiSelectPanel.vue'
import RealtimeTab from './RealtimeTab.vue'
import QueryTab from './QueryTab.vue'
import { useRealtimeData } from './useRealtimeData'
import { useQueryData } from './useQueryData'
import { useTrends } from './useTrends'
import { useServiceAreaStore } from '@/stores/serviceArea'
import { useDateRange, type DatePresetValue } from '@/composables/useDateRange'
import { theme } from '@/utils/theme'

defineOptions({ name: 'OperationScreen' })

const serviceArea = useServiceAreaStore()
const realtime = useRealtimeData()
const query = useQueryData()
const dateRange = useDateRange()

const trends = useTrends(
  () => dateRange.current.value,
  () => serviceArea.selectedIds,
)

const activeTab = ref(0)
const refreshing = ref(false)
const showServicePanel = ref(false)

async function loadAll(): Promise<void> {
  if (serviceArea.isEmpty) return
  showLoadingToast({ duration: 0, forbidClick: true, message: '加载中...' })
  try {
    await Promise.all([
      realtime.loadAll(serviceArea.selectedIds),
      query.loadAll(dateRange.current.value, dateRange.previous.value, serviceArea.selectedIds),
    ])
  } catch {
    showToast({ message: '获取数据失败', duration: 1000 })
  } finally {
    closeToast()
  }
}

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

function onSelectPreset(preset: DatePresetValue): void {
  if (dateRange.selectPreset(preset)) void loadAll()
}

function onSelectCustom(start: string, end: string): void {
  if (dateRange.setCustom(start, end) && start && end) void loadAll()
}

const TREND_ACTIONS = {
  orderAmount: () => trends.openOrderAmount(),
  ticketOrder: () => trends.openTicketOrder(),
  averageAmount: () => trends.openAverage('amount'),
  averageNum: () => trends.openAverage('num'),
  userGrow: () => trends.openUserGrow(),
  carsData: () => trends.openCarsData(),
  taskManage: () => trends.openTaskManage(query.taskLine.value, query.taskTotal.value),
} as const

function onOpenTrend(key: keyof typeof TREND_ACTIONS): void {
  void TREND_ACTIONS[key]()
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
        <span class="service-bar__action" :style="{ color: theme.themeColor }" @click="openServicePanel">
          选择<van-icon name="arrow-down" />
        </span>
      </div>

      <van-tabs v-model:active="activeTab" type="card" :color="theme.themeColor" class="main-tabs">
        <van-tab title="实时数据" />
        <van-tab title="查询数据" />
      </van-tabs>

      <div class="grey-line" />

      <RealtimeTab
        v-if="activeTab === 0"
        :vehicle-stats="realtime.vehicleStats.value"
        :idle-stats="realtime.idleStats.value"
        :battery-stats="realtime.batteryStats.value"
        :user-tree="realtime.userTree.value"
        :authentication="realtime.userInfo.value?.authentication ?? 0"
        :no-authentication="realtime.userInfo.value?.noAuthentication ?? 0"
        :riding-qualification="realtime.ridingQualification.value"
      />

      <QueryTab
        v-else
        :preset="dateRange.preset.value"
        :custom-start="dateRange.customStart.value"
        :custom-end="dateRange.customEnd.value"
        :metrics="query.metrics.value"
        :comparison="query.comparison.value"
        :order-count="query.orderCount.value"
        :order-amount="query.orderAmount.value"
        :task-slices="query.taskSlices.value"
        :task-total="query.taskTotal.value"
        @select-preset="onSelectPreset"
        @select-custom="onSelectCustom"
        @open-trend="onOpenTrend"
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
  margin: 24px auto 22px;
  border-radius: 6px;
}

.service-popup {
  height: 100%;
  width: 100%;
}
</style>
