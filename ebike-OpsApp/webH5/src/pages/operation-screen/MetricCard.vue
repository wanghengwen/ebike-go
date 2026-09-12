<script setup lang="ts">
import { computed } from 'vue'
import { numFormat, numFormatInt } from '@/utils/numFormat'

const props = defineProps<{
  label: string
  unit: string
  value: number
  /** 环比百分比；null 表示当前区间（累计）没有同期可比。 */
  delta: number | null
  /** 整数指标（订单量）不补两位小数。 */
  integer?: boolean
}>()

const display = computed(() =>
  props.integer ? numFormatInt(props.value) : numFormat(props.value),
)

const deltaText = computed(() => {
  if (props.delta === null) return ''
  const formatted = numFormat(props.delta)
  return props.delta > 0 ? `+${formatted}` : formatted
})
</script>

<template>
  <div class="metric">
    <div class="metric__head">
      <p class="metric__label">{{ label }}</p>
      <p class="metric__unit">({{ unit }})</p>
    </div>
    <div class="metric__value" :class="{ 'metric__value--tall': delta === null }">
      {{ display }}
    </div>
    <div v-if="delta !== null" class="metric__delta">
      同期:<span :class="delta > 0 ? 'value-up' : 'value-down'">{{ deltaText }}%</span>
    </div>
  </div>
</template>

<style scoped>
.metric {
  margin-bottom: 16px;
  width: 164px;
  height: 80px;
  background: #ffffff;
  box-shadow: 0 0 6px 0 rgba(233, 219, 215, 0.5);
  border: 1px solid #e6e6e6;
  border-radius: 4px;
  opacity: 0.92;
  padding: 8px 0;
  display: flex;
  flex-direction: column;
  justify-content: space-around;
  align-items: center;
}

.metric__head {
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
}

/* 遗留写了 12px / 10px，但 `.show_three p` 的优先级更高，实际两行都渲染成 13px */
.metric__label,
.metric__unit {
  height: 17px;
  font-size: 13px;
  color: #a0a0a0;
  line-height: 17px;
}

.metric__value {
  height: 24px;
  font-size: 18px;
  font-weight: bold;
  color: #282828;
  line-height: 24px;
}

.metric__value--tall {
  margin-top: -15px;
}

.metric__delta {
  height: 16px;
  font-size: 13px;
  color: #a0a0a0;
  line-height: 16px;
}
</style>
