<script setup lang="ts">
import { ref } from 'vue'
import { showToast } from 'vant'
import { theme } from '@/utils/theme'
import { formatDate } from '@/composables/useDateRange'
import type { RevenuePresetValue } from '@/composables/useRevenueDateRange'

const props = defineProps<{
  presets: Array<[label: string, value: RevenuePresetValue]>
  preset: RevenuePresetValue
  customStart: string
  customEnd: string
  /** 六个档位时按等分铺满整行（遗留「数据」页签用的是 van-tabs 平铺）。 */
  spread?: boolean
}>()

const emit = defineEmits<{
  'select-preset': [preset: RevenuePresetValue]
  'select-custom': [start: string, end: string]
}>()

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

    <div class="presets" :class="{ 'presets--spread': spread }">
      <div
        v-for="[label, value] in presets"
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

    <van-popup v-model:show="showPicker" position="bottom">
      <van-date-picker v-model="pickedDate" @cancel="showPicker = false" @confirm="confirmDate" />
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
  padding: 8px;
}

.date-filter__dash {
  color: #f2f2f2;
}

.presets {
  display: flex;
  margin: 20px 10px 0 0;
}

.presets--spread {
  margin: 20px 0 5px;
}

.presets--spread .presets__item {
  flex: 1;
  margin-left: 0;
  text-align: center;
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
</style>
