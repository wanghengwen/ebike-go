<template>
  <view class="page">
    <view class="map-wrap card">
      <view class="title">{{ t('ride.tripMap') }}</view>
      <map
        class="map"
        :latitude="latitude"
        :longitude="longitude"
        :scale="15"
        :polyline="polyline"
        :markers="markers"
        show-location
      />
      <view v-if="loading" class="tip">{{ t('common.loading') }}</view>
      <view v-else-if="!orderId" class="tip">{{ t('common.empty') }}</view>
      <view v-else-if="!points.length" class="tip">{{ t('ride.noTrack') }}</view>
      <view v-else class="tip">{{ orderId }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getOrderDetail } from '@/api/order'
import { useTempDataStore } from '@/stores/tempData'
import { setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const temp = useTempDataStore()
const loading = ref(false)
const orderId = ref('')
const latitude = ref(temp.location?.latitude || 30.25)
const longitude = ref(temp.location?.longitude || 120.15)
const points = ref<Array<{ latitude: number; longitude: number }>>([])

const polyline = computed(() => {
  if (!points.value.length) return []
  return [
    {
      points: points.value,
      color: '#3AA0E8',
      width: 4,
    },
  ]
})

const markers = computed(() => {
  if (!points.value.length) {
    return [
      {
        id: 1,
        latitude: latitude.value,
        longitude: longitude.value,
        width: 24,
        height: 24,
      },
    ]
  }
  return [
    { id: 1, latitude: points.value[0].latitude, longitude: points.value[0].longitude, width: 24, height: 24 },
    {
      id: 2,
      latitude: points.value[points.value.length - 1].latitude,
      longitude: points.value[points.value.length - 1].longitude,
      width: 24,
      height: 24,
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
      if (Number.isNaN(lat) || Number.isNaN(lng)) return null
      return { latitude: lat, longitude: lng }
    })
    .filter(Boolean) as Array<{ latitude: number; longitude: number }>
}

onLoad(async (q) => {
  orderId.value = decodeURIComponent((q?.orderId as string) || '')
  // Legacy: tripMap?params=JSON.stringify(deviceTrajectory)
  if (q?.params) {
    try {
      const raw = JSON.parse(decodeURIComponent(String(q.params)))
      const parsed = parsePoints(raw)
      if (parsed.length) {
        points.value = parsed
        latitude.value = parsed[0].latitude
        longitude.value = parsed[0].longitude
        return
      }
    } catch {
      /* fall through to order detail */
    }
  }
  if (!orderId.value) return
  loading.value = true
  const res = await getOrderDetail({ orderId: orderId.value, izNewApp: true })
  loading.value = false
  if (!res.success || !res.data) return
  const data = res.data as Record<string, unknown>
  // Prefer real track; never invent an offset polyline when empty
  const parsed = parsePoints(
    data.deviceTrajectory || data.track || data.points || data.polyline || data.path,
  )
  if (parsed.length) {
    points.value = parsed
    latitude.value = parsed[0].latitude
    longitude.value = parsed[0].longitude
  } else {
    points.value = []
    const startLat = Number(data.startLat ?? data.startLatitude)
    const startLng = Number(data.startLng ?? data.startLongitude)
    if (!Number.isNaN(startLat) && !Number.isNaN(startLng)) {
      latitude.value = startLat
      longitude.value = startLng
    }
  }
})
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.map {
  width: 100%;
  height: 560rpx;
  border-radius: 12rpx;
}
.tip {
  margin-top: 16rpx;
  color: #888;
}
</style>
