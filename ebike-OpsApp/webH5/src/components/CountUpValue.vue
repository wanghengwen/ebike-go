<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { CountUp } from 'countup.js'

/**
 * 取代 `vue-countup-v2`（只支持 Vue 2）。直接包 `countup.js`，行为与遗留一致：
 * 值为 null / undefined 时显示 `-- --`，0 仍然要跑动画。
 */
const props = withDefaults(
  defineProps<{
    value: number | null | undefined
    decimalPlaces?: number
    color?: string
    duration?: number
  }>(),
  { decimalPlaces: 2, color: '#272727', duration: 0.4 },
)

const target = ref<HTMLElement | null>(null)
let counter: CountUp | null = null

function hasValue(): boolean {
  return props.value !== null && props.value !== undefined && Number.isFinite(props.value)
}

function mount(): void {
  if (!target.value || !hasValue()) return
  counter = new CountUp(target.value, props.value as number, {
    useEasing: true,
    useGrouping: true,
    separator: ',',
    decimal: '.',
    decimalPlaces: props.decimalPlaces,
    duration: props.duration,
  })
  if (!counter.error) counter.start()
}

onMounted(mount)

watch(
  () => props.value,
  (next) => {
    if (!hasValue()) {
      counter?.reset()
      counter = null
      return
    }
    if (counter) counter.update(next as number)
    else mount()
  },
)

onBeforeUnmount(() => {
  counter = null
})
</script>

<template>
  <span class="count-up">
    <span v-if="hasValue()" ref="target" :style="{ color }" />
    <span v-else>-- --</span>
  </span>
</template>

<style scoped>
.count-up {
  display: inline-block;
  margin: 0;
}
</style>
