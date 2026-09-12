<script setup lang="ts">
import { ref } from 'vue'
import CategoryBarChart from '@/components/charts/CategoryBarChart.vue'
import DonutChart from '@/components/charts/DonutChart.vue'
import UserTreeChart from '@/components/charts/UserTreeChart.vue'
import type { UserTreeNode } from '@/components/charts/UserTreeChart.vue'
import type { PieSlice } from '@/components/charts/options'
import { theme } from '@/utils/theme'
import type { BarBlock } from './useRealtimeData'

defineProps<{
  vehicleStats: BarBlock
  idleStats: BarBlock
  batteryStats: BarBlock
  userTree: UserTreeNode[]
  authentication: number
  noAuthentication: number
  ridingQualification: { slices: PieSlice[]; total: number }
}>()

const activeTab = ref(0)
</script>

<template>
  <div>
    <van-tabs
      v-model:active="activeTab"
      :color="theme.themeColor"
      :title-active-color="theme.themeColor"
      title-inactive-color="#282828"
      line-width="12px"
      :border="false"
      class="sub-tabs"
    >
      <van-tab title="车辆状态" />
      <van-tab title="用户详情" />
    </van-tabs>

    <div v-if="activeTab === 0">
      <CategoryBarChart v-bind="vehicleStats" />
      <CategoryBarChart v-bind="idleStats" />
      <CategoryBarChart v-bind="batteryStats" />
    </div>

    <div v-else>
      <UserTreeChart
        :nodes="userTree"
        :authentication="authentication"
        :no-authentication="noAuthentication"
      />
      <div class="grey-line" />
      <DonutChart
        title="实名认证用户"
        unit="人"
        :total="ridingQualification.total"
        :slices="ridingQualification.slices"
        :palette="['#6DD400', '#C20879']"
      />
    </div>
  </div>
</template>

<style scoped>
.sub-tabs {
  width: 150px;
  margin: 0 0 5px 7px;
}
</style>
