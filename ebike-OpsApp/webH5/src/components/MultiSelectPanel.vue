<script setup lang="ts">
import { ref, watch } from 'vue'
import { showToast } from 'vant'
import { theme } from '@/utils/theme'

/**
 * 服务区 / 加盟商多选面板。遗留的 `xcService.vue` 与 `xcAgent.vue` 是两份除文案外
 * 完全相同的复制体，这里合并成一个组件。
 *
 * 选中态直接用数组维护，不再靠 `this.$refs.checkboxes[i].toggle()`——
 * Vue 3 里 `v-for` 上的 ref 不再自动收集成数组，那种写法也不好测。
 */
const props = defineProps<{
  title: string
  emptyHint: string
  items: Array<{ id: string; name: string }>
  modelValue: string[]
}>()

const emit = defineEmits<{
  'update:modelValue': [ids: string[]]
  confirm: [ids: string[]]
  cancel: []
}>()

const draft = ref<string[]>([...props.modelValue])

/** 面板打开时的选中数；标题底色靠它区分「未改动」与「已改动」。 */
const baseline = ref(props.modelValue.length)

watch(
  () => props.modelValue,
  (value) => {
    draft.value = [...value]
    baseline.value = value.length
  },
)

const isPristine = () => draft.value.length === baseline.value
const isAllSelected = () => props.items.length > 0 && draft.value.length === props.items.length

function toggle(id: string): void {
  draft.value = draft.value.includes(id)
    ? draft.value.filter((item) => item !== id)
    : [...draft.value, id]
}

function toggleAll(): void {
  draft.value = isAllSelected() ? [] : props.items.map((item) => item.id)
}

function cancel(): void {
  draft.value = [...props.modelValue]
  emit('cancel')
}

function confirm(): void {
  if (draft.value.length === 0) {
    showToast(props.emptyHint)
    return
  }
  emit('update:modelValue', [...draft.value])
  emit('confirm', [...draft.value])
}
</script>

<template>
  <div class="panel">
    <div class="panel__title" :class="{ 'panel__title--pristine': isPristine() }">
      {{ title }}({{ draft.length }})
    </div>
    <div
      class="panel__select-all"
      :style="{ backgroundColor: theme.themeColor, color: theme.lightTxtColor }"
      @click="toggleAll"
    >
      {{ isAllSelected() ? '取消全选' : '全选' }}
    </div>

    <div class="panel__list">
      <van-cell-group>
        <van-cell
          v-for="item in items"
          :key="item.id"
          clickable
          :title="item.name"
          class="panel__cell"
          :class="draft.includes(item.id) ? 'panel__cell--active' : ''"
          @click="toggle(item.id)"
        >
          <template #icon>
            <van-checkbox
              :model-value="draft.includes(item.id)"
              :checked-color="theme.themeColor"
              shape="square"
              class="panel__checkbox"
              @click.stop="toggle(item.id)"
            />
          </template>
        </van-cell>
      </van-cell-group>
    </div>

    <div class="panel__actions">
      <div
        class="panel__button panel__button--ghost"
        :style="{ color: theme.themeColor, borderColor: theme.themeColor }"
        @click="cancel"
      >
        取消
      </div>
      <div
        class="panel__button"
        :style="{ backgroundColor: theme.themeColor, color: theme.lightTxtColor }"
        @click="confirm"
      >
        确定
      </div>
    </div>
  </div>
</template>

<style scoped>
.panel {
  position: relative;
  height: 100%;
  width: 100%;
}

.panel__title {
  height: 40px;
  width: 100%;
  line-height: 40px;
  font-size: 16px;
  text-align: center;
  background: #ffffff;
  border-bottom: 1px solid #f2f2f2;
}

.panel__title--pristine {
  background: #f8f8f8;
}

.panel__select-all {
  position: absolute;
  top: 7px;
  right: 5px;
  height: 25px;
  line-height: 25px;
  padding: 0 10px;
  font-size: 14px;
  text-align: center;
}

.panel__list {
  position: absolute;
  top: 40px;
  bottom: 80px;
  left: 0;
  right: 0;
  overflow-y: scroll;
  border-bottom: 1px solid #f2f2f2;
}

.panel__cell {
  width: 344px;
  height: 44px;
  position: relative;
  left: 50%;
  transform: translateX(-50%);
  margin-top: 8px;
  background-color: #ffffff;
}

.panel__cell--active {
  background-color: #f8f8f8;
}

.panel__checkbox {
  margin-right: 14px;
}

.panel__actions {
  position: fixed;
  bottom: 15px;
  width: 100%;
  display: flex;
  justify-content: center;
}

.panel__button {
  height: 42px;
  width: 145px;
  line-height: 42px;
  text-align: center;
  font-size: 14px;
  margin-left: 35px;
}

.panel__button--ghost {
  background-color: #ffffff;
  border: 1px solid currentColor;
  margin-left: 0;
}
</style>
