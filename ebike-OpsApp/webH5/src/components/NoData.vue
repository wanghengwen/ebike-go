<script setup lang="ts">
import { onMounted } from 'vue'
import { closeToast } from 'vant'

// 遗留实现在 `updated()` 里调 `Toast.clear()`，等于每次父组件重渲染都关一次全局提示。
// 空态只会出现一次，挂载时关掉就够了。
onMounted(closeToast)
</script>

<template>
  <div class="no-data">
    <van-empty :description="$t('common.noData')" />
    <slot />
  </div>
</template>

<style scoped>
.no-data {
  position: relative;
  text-align: center;
}
</style>
