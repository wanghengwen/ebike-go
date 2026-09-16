<template>
  <view class="page">
    <map
      id="parkSearchMap"
      class="map"
      :provider="mapProvider()"
      :latitude="latitude"
      :longitude="longitude"
      :scale="16"
      :markers="markers"
      :polyline="polyline"
      :polygons="fencePolygons"
      show-location
      @tap="onMapTap"
      @callouttap="onCalloutTap"
    />
    <view class="card panel">
      <view class="title" :class="{ danger: izOutService }">
        {{ izOutService ? t('ride.parkOutService') : t('ride.parkInService') }}
      </view>
      <view v-if="izOutService" class="out-tip">{{ t('ride.parkOutServiceReselect') }}</view>
      <view v-if="loading">{{ t('common.loading') }}</view>
      <template v-else>
        <view class="count" v-if="!izOutService && parkingNum != null">
          {{ t('ride.parkNear', { n: parkingNum }) }}
        </view>
        <view v-else-if="!izOutService" class="empty">{{ t('ride.parkEmpty') }}</view>
        <view class="dist" v-if="!izOutService && distance != null">
          {{ t('ride.parkDistance', { m: distance }) }}
        </view>
        <view class="dist" v-if="!izOutService && routeHint">{{ routeHint }}</view>
      </template>
      <view
        class="btn-primary"
        @click="drawNearRoute"
        v-if="!izOutService && nearest?.lat && nearest?.lng"
      >
        {{ mode === 'ride' ? t('ride.showRideRoute') : t('ride.showWalkRoute') }}
      </view>
      <view
        class="btn-primary"
        @click="handoffToRiding"
        v-if="mode === 'ride' && lastNavPayload"
      >
        {{ t('ride.startInstrumentNav') }}
      </view>
      <view class="btn-primary" @click="openLocation" v-if="latitude && longitude">
        {{ t('ride.openLocation') }}
      </view>
      <view class="btn-ghost" @click="load">{{ t('common.retry') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onHide, onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getNearParkingNum, getServiceByPoi } from '@/api/map'
import { useMapLocation } from '@/features/map/useMapLocation'
import { useRideNav } from '@/features/map/useRideNav'
import { useTempDataStore } from '@/stores/tempData'
import { getMapCfg } from '@/shared/tenantSkin'
import { openThirdPartyMap } from '@/shared/openMapApp'
import { navigate, setNavTitle } from '@/shared/navigate'
import { mapProvider } from '@/shared/mapProvider'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const temp = useTempDataStore()
const { locate } = useMapLocation()
const {
  polyline,
  routeMeta,
  drawRoute,
  parkMarkers,
  fencePolygons,
  lastNavPayload,
  loadRidingFences,
  saveNavForHandoff,
} = useRideNav()
const loading = ref(false)
const mode = ref<'walk' | 'ride'>('walk')
const latitude = ref(temp.location?.latitude || 30.25)
const longitude = ref(temp.location?.longitude || 120.15)
const parkingNum = ref<number | null>(null)
const distance = ref<number | null>(null)
const nearest = ref<{ lat?: number; lng?: number; name?: string } | null>(null)
const izOutService = ref(false)
/** Destination serviceId from getServiceByPoi — not cached storage */
const destServiceId = ref('')

const routeHint = computed(() => {
  const d = routeMeta.value.distance
  const min = routeMeta.value.durationMin
  if (d == null || min == null) return ''
  return mode.value === 'ride'
    ? t('ride.rideRouteHint', { m: Math.round(d), min })
    : t('ride.walkRouteHint', { m: Math.round(d), min })
})

const markers = computed(() => {
  const centerIcon = getMapCfg('centerMarker')
  const stationIcon = getMapCfg('station') || getMapCfg('parking') || getMapCfg('searchStationIcon')
  const list: Array<Record<string, unknown>> = [
    {
      id: 1,
      latitude: latitude.value,
      longitude: longitude.value,
      width: 21,
      height: 35,
      // 查询中心：旧版 centerMarker，避免微信默认红针
      ...(centerIcon ? { iconPath: centerIcon } : {}),
    },
  ]
  if (nearest.value?.lat && nearest.value?.lng) {
    list.push({
      id: 2,
      latitude: nearest.value.lat,
      longitude: nearest.value.lng,
      width: 28,
      height: 32,
      // 最近停车点：必须用 P 点图
      ...(stationIcon ? { iconPath: stationIcon } : {}),
      callout:
        routeMeta.value.distance != null && routeMeta.value.durationMin != null
          ? {
              content:
                mode.value === 'ride'
                  ? t('ride.rideCallout', {
                      m: Math.round(routeMeta.value.distance),
                      min: routeMeta.value.durationMin,
                    })
                  : t('ride.walkCallout', {
                      m: Math.round(routeMeta.value.distance),
                      min: routeMeta.value.durationMin,
                    }),
              display: 'ALWAYS',
              color: '#333333',
              fontSize: 12,
              borderRadius: 8,
              padding: 6,
              bgColor: '#ffffff',
              textAlign: 'center',
            }
          : nearest.value.name
            ? { content: nearest.value.name, display: 'ALWAYS' }
            : undefined,
    })
  }
  // 围栏站点：跳过与最近停车点几乎重合的，避免叠两个针
  for (const m of parkMarkers.value) {
    if (m.id === 2) continue
    if (nearest.value?.lat && nearest.value?.lng) {
      const near =
        Math.abs(m.latitude - nearest.value.lat) < 0.00001 &&
        Math.abs(m.longitude - nearest.value.lng) < 0.00001
      if (near) continue
    }
    const withIcon =
      m.iconPath ||
      (m.type === 1
        ? getMapCfg('notAllowStation') || getMapCfg('noParking')
        : stationIcon)
    list.push({
      ...m,
      ...(withIcon ? { iconPath: withIcon } : {}),
    })
  }
  return list
})

onShow(() => setNavTitle(t('ride.parkSearch')))

onHide(() => {
  // Legacy back(): persist nav for riding/precycling handoff
  if (mode.value === 'ride') saveNavForHandoff()
})

onLoad(async (q) => {
  if (q?.mode === 'ride') mode.value = 'ride'
  if (q?.latitude) latitude.value = Number(q.latitude)
  if (q?.longitude) longitude.value = Number(q.longitude)
  if (!q?.latitude || !q?.longitude) {
    const loc = temp.location || (await locate())
    if (loc) {
      latitude.value = loc.latitude
      longitude.value = loc.longitude
    }
  }
  await loadRidingFences()
  await load()
})

function parseOutService(data: Record<string, unknown>): boolean {
  const v = data.izOutService ?? data.outOfService ?? data.isOutService
  if (v === true || v === 1 || v === '1' || v === 'true') return true
  if (v === false || v === 0 || v === '0' || v === 'false') return false
  return false
}

async function load() {
  loading.value = true
  try {
    // Legacy parkSearch: resolve destination serviceId via getServiceByPoi(dest)
    try {
      const svc = await getServiceByPoi({
        lat: latitude.value,
        lng: longitude.value,
      })
      if (svc.success && svc.data) {
        destServiceId.value = String((svc.data as { id?: string }).id || '')
      } else {
        destServiceId.value = ''
      }
    } catch (e) {
      logger.warn('park-search getServiceByPoi soft fail', e)
      destServiceId.value = ''
    }

    const res = await getNearParkingNum({
      locationDTO: {
        lng: longitude.value,
        lat: latitude.value,
      },
      serviceId: destServiceId.value,
    })
    if (!res.success || res.data == null) {
      parkingNum.value = null
      distance.value = null
      nearest.value = null
      izOutService.value = false
      return
    }
    const data = res.data as Record<string, unknown> | number
    if (typeof data === 'number') {
      parkingNum.value = data
      izOutService.value = false
      return
    }
    izOutService.value = parseOutService(data)
    parkingNum.value = Number(data.parkingNum ?? data.num ?? data.count ?? 0) || null
    if (!parkingNum.value) parkingNum.value = null
    const dist = data.distance
    distance.value = dist != null ? Number(Number(dist).toFixed(2)) : null
    if (izOutService.value) {
      nearest.value = null
      return
    }
    const near = (data.nearest || data.parking || data) as Record<string, unknown>
    const nLat = Number(near.lat ?? near.latitude ?? near.centerLat)
    const nLng = Number(near.lng ?? near.longitude ?? near.centerLng)
    if (!Number.isNaN(nLat) && !Number.isNaN(nLng) && nLat && nLng) {
      nearest.value = { lat: nLat, lng: nLng, name: String(near.name || near.title || '') }
      void drawRoute(
        { latitude: nLat, longitude: nLng },
        mode.value === 'ride' ? 2 : 1,
      )
    }
  } finally {
    loading.value = false
  }
}

async function drawNearRoute() {
  if (!nearest.value?.lat || !nearest.value?.lng) return
  const res = await drawRoute(
    { latitude: nearest.value.lat, longitude: nearest.value.lng },
    mode.value === 'ride' ? 2 : 1,
  )
  if (!res.success) {
    uni.showToast({ title: res.msg || t('ride.navFail'), icon: 'none' })
  }
}

function handoffToRiding() {
  saveNavForHandoff()
  navigate('back')
}

function onMapTap(e: { detail?: { latitude?: number; longitude?: number } }) {
  if (e.detail?.latitude != null && e.detail?.longitude != null) {
    latitude.value = e.detail.latitude
    longitude.value = e.detail.longitude
    void load()
  }
}

function openLocation() {
  const lat = nearest.value?.lat || latitude.value
  const lng = nearest.value?.lng || longitude.value
  void openThirdPartyMap('parkSearchMap', {
    latitude: lat,
    longitude: lng,
    name: nearest.value?.name || t('ride.parkSearch'),
  })
}

async function onCalloutTap(e: { detail?: { markerId?: number } }) {
  const id = Number(e?.detail?.markerId)
  if (id === 2 && nearest.value?.lat && nearest.value?.lng) {
    await openThirdPartyMap('parkSearchMap', {
      latitude: nearest.value.lat,
      longitude: nearest.value.lng,
      name: nearest.value.name || t('ride.parkSearch'),
    })
    return
  }
  const park = parkMarkers.value.find((m) => m.id === id)
  if (park) {
    await openThirdPartyMap('parkSearchMap', {
      latitude: park.latitude,
      longitude: park.longitude,
      name: park.title || t('ride.parkSearch'),
    })
  }
}

onMounted(() => {
  /* load via onLoad */
})
</script>

<style scoped lang="scss">
.map {
  width: 100%;
  height: 55vh;
}
.panel {
  margin-top: -24rpx;
  position: relative;
  z-index: 2;
}
.title {
  font-weight: 700;
  margin-bottom: 12rpx;
}
.title.danger {
  color: #e34d59;
}
.out-tip {
  color: #e34d59;
  font-size: 26rpx;
  margin-bottom: 16rpx;
}
.count {
  font-size: 32rpx;
  margin-bottom: 8rpx;
}
.dist {
  color: #666;
  margin-bottom: 20rpx;
}
.empty {
  color: #999;
  margin-bottom: 20rpx;
}
.btn-ghost {
  margin-top: 16rpx;
}
</style>
