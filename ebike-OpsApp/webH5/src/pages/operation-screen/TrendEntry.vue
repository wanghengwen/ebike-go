<script setup lang="ts">
defineProps<{ label: string }>()
</script>

<template>
  <van-cell :title="label" is-link class="trend-entry" />
</template>

<style scoped>
.trend-entry :deep(.van-cell__title) {
  font-size: 14px;
  color: #282828;
}
</style>
