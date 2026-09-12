<script setup lang="ts">
import { computed, ref } from 'vue'
import DateFilter from './DateFilter.vue'
import StatSection from './StatSection.vue'
import { usePermissionStore } from '@/stores/permission'
import { DATA_PRESETS, type RevenuePresetValue } from '@/composables/useRevenueDateRange'
import type { StatGroup } from './useRevenueData'

const props = defineProps<{
  groups: StatGroup[]
  preset: RevenuePresetValue
  customStart: string
  customEnd: string
}>()

const emit = defineEmits<{
  'select-preset': [preset: RevenuePresetValue]
  'select-custom': [start: string, end: string]
}>()

const permission = usePermissionStore()

const visibleGroups = computed(() =>
  props.groups.filter((group) => permission.has(group.permissionCode)),
)

/** 折叠面板是单开的，默认展开第一组。 */
const active = ref('revenue')
</script>

<template>
  <div>
    <DateFilter
      spread
      :presets="DATA_PRESETS"
      :preset="preset"
      :custom-start="customStart"
      :custom-end="customEnd"
      @select-preset="emit('select-preset', $event)"
      @select-custom="(start, end) => emit('select-custom', start, end)"
    />

    <van-collapse v-model="active" accordion class="groups">
      <van-collapse-item
        v-for="group in visibleGroups"
        :key="group.key"
        :title="group.title"
        :name="group.key"
      >
        <StatSection :group="group" />
      </van-collapse-item>
    </van-collapse>
  </div>
</template>

<style scoped>
.groups {
  margin-top: 12px;
  font-size: 15px;
  font-weight: 600;
  color: #282828;
  line-height: 21px;
}

.groups :deep(.van-collapse-item__content) {
  background: #f8f8f8;
  padding: 12px 0 16px;
}
</style>
