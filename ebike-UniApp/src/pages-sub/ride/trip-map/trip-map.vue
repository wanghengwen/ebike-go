<template>
  <!-- Legacy tripMap.vue：全屏地图 + 轨迹折线 + 起终点 marker -->
  <view class="page">
    <map
      class="map"
      :latitude="latitude"
      :longitude="longitude"
      :scale="18"
      :polyline="polyline"
      :markers="markers"
      :show-location="false"
      enable-scroll
      enable-zoom
    />
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getMapCfg, getIconCfg } from '@/shared/tenantSkin'
import { setNavTitle } from '@/shared/navigate'

const { t } = useI18n()

const latitude = ref(39.909)
const longitude = ref(116.39742)
const points = ref<Array<{ latitude: number; longitude: number }>>([])

const startIcon = computed(
  () => getMapCfg('trajectory_start') || getIconCfg('trajectory_start') || '',
)
const endIcon = computed(
  () => getMapCfg('trajectory_end') || getIconCfg('trajectory_end') || '',
)

const polyline = computed(() => {
  if (!points.value.length) return []
  return [
    {
      points: points.value,
      color: '#1890ff',
      width: 6,
      arrowLine: true,
      dottedLine: false,
    },
  ]
})

const markers = computed(() => {
  if (!points.value.length) return []
  const start = points.value[0]
  const end = points.value[points.value.length - 1]
  return [
    {
      id: 1,
      latitude: end.latitude,
      longitude: end.longitude,
      width: 28,
      height: 30,
      iconPath: endIcon.value || undefined,
    },
    {
      id: 2,
      latitude: start.latitude,
      longitude: start.longitude,
      width: 28,
      height: 30,
      iconPath: startIcon.value || undefined,
    },
  ]
})

onShow(() => setNavTitle(t('ride.tripMap')))

function parsePoints(raw: unknown): Array<{ latitude: number; longitude: number }> {
  if (!Array.isArray(raw)) return []
  return raw
    .map((p) => {
      const item = p as Record<string, unknown>
      const lat = Number(item.lat ?? item.latitude)
      const lng = Number(item.lng ?? item.longitude)
      if (!Number.isFinite(lat) || !Number.isFinite(lng)) return null
      return { latitude: lat, longitude: lng }
    })
    .filter(Boolean) as Array<{ latitude: number; longitude: number }>
}

onLoad((q) => {
  // Legacy: tripMap?params=JSON.stringify(deviceTrajectory)
  if (!q?.params) {
    uni.showToast({ title: t('ride.noTripInfo'), icon: 'none' })
    return
  }
  try {
    const raw = JSON.parse(decodeURIComponent(String(q.params)))
    const parsed = parsePoints(raw)
    if (!parsed.length) {
      uni.showToast({ title: t('ride.noTripInfo'), icon: 'none' })
      return
    }
    points.value = parsed
    latitude.value = parsed[0].latitude
    longitude.value = parsed[0].longitude
  } catch {
    uni.showToast({ title: t('ride.noTripInfo'), icon: 'none' })
  }
})
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
}
.map {
  width: 100%;
  height: 100%;
}
</style>
