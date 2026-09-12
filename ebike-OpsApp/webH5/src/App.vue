<script setup lang="ts"></script>

<template>
  <!-- Vue 3 里 keep-alive 必须包在 router-view 的插槽内容外层 -->
  <router-view v-slot="{ Component, route }">
    <keep-alive>
      <component :is="Component" v-if="route.meta.keepAlive" :key="String(route.name)" />
    </keep-alive>
    <component :is="Component" v-if="!route.meta.keepAlive" :key="String(route.name)" />
  </router-view>
</template>

<style>
html,
body,
p,
div {
  margin: 0;
  padding: 0;
}

body {
  background: #ffffff;
  color: #282828;
  font-family:
    -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Helvetica Neue', Helvetica, Arial,
    sans-serif;
}
</style>
