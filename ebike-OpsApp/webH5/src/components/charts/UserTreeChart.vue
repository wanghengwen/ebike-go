<script setup lang="ts">
import { computed } from 'vue'
import BaseChart from './BaseChart.vue'
import type { ChartOption } from '@/composables/useEcharts'

export interface UserTreeNode {
  name: string
  value: number
  itemStyle?: { color: string }
  children?: UserTreeNode[]
}

const props = defineProps<{
  nodes: UserTreeNode[]
  authentication: number
  noAuthentication: number
}>()

const totalUsers = computed(() => {
  const total = props.authentication + props.noAuthentication
  return total ? total : '--'
})

function branchLabel(param: { data: UserTreeNode }): string {
  return `${param.data.name}\n(${param.data.value})`
}

/** 叶子节点挤在最右侧一列，名称超过 4 字要折行才不互相叠字。 */
function leafLabel(param: { data: UserTreeNode }): string {
  const { name, value } = param.data
  const wrapped = name.length >= 4 ? `${name.slice(0, 3)}\n${name.slice(3)}` : name
  return `${wrapped}\n(${value})`
}

const option = computed<ChartOption>(() => ({
  tooltip: {
    trigger: 'item',
    triggerOn: 'mousemove',
    enterable: false,
    alwaysShowContent: false,
    hideDelay: 100,
    formatter: '{b}:<br />{c}',
  },
  animation: false,
  series: [
    {
      type: 'tree',
      symbol: 'circle',
      symbolSize: 16,
      top: 60,
      data: props.nodes,
      orient: 'horizontal',
      expandAndCollapse: true,
      initialTreeDepth: -1,
      lineStyle: { curveness: 0 },
      itemStyle: { borderColor: '#FFFFFF' },
      label: {
        position: 'top',
        verticalAlign: 'bottom',
        align: 'center',
        fontSize: 12,
        fontWeight: 600,
        formatter: branchLabel,
      },
      leaves: {
        label: {
          position: 'bottom',
          verticalAlign: 'top',
          align: 'center',
          distance: 1,
          fontSize: 12,
          fontWeight: 600,
          formatter: leafLabel,
        },
      },
    },
  ],
}))
</script>

<template>
  <div>
    <div class="chart-title">
      <span>总用户:</span>
      <span class="chart-title__value">{{ totalUsers }}</span>
      <span>人</span>
    </div>
    <div class="chart-body">
      <BaseChart :option="option" height="400px" width="100%" />
    </div>
  </div>
</template>

<style scoped>
.chart-title {
  height: 25px;
  font-size: 14px;
  font-weight: 550;
  color: #333333;
  line-height: 17px;
  margin: 16px 0 2px 16px;
}

.chart-title__value {
  font-size: 20px;
  margin: 0 2px 0 5px;
}

.chart-body {
  display: flex;
  justify-content: center;
  margin: -40px 10px 20px 10px;
}
</style>
