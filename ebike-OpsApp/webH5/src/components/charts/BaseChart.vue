<script setup lang="ts">
import { ref, toRef } from 'vue'
import { closeToast } from 'vant'
import { useEcharts, type ChartOption } from '@/composables/useEcharts'

const props = withDefaults(
  defineProps<{
    option: ChartOption | null
    height?: string
    width?: string
    /** 首帧渲染完成后关闭全局 loading，沿用遗留大屏的 `chart.on('finished')` 行为。 */
    closeToastOnFinished?: boolean
  }>(),
  { height: '350px', width: '100%', closeToastOnFinished: true },
)

const container = ref<HTMLElement | null>(null)

useEcharts(container, toRef(props, 'option'), {
  onFinished: () => {
    if (props.closeToastOnFinished) closeToast()
  },
})
</script>

<template>
  <div ref="container" :style="{ height, width }" />
</template>
