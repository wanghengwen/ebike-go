<template>
  <view class="page">
    <map
      id="applyStationMap"
      class="map"
      :provider="mapProvider()"
      :latitude="latitude"
      :longitude="longitude"
      :scale="16"
      :markers="mapMarkers"
      :polygons="polygons"
      show-location
      @tap="onMapTap"
    />
    <view class="card panel">
      <view class="title-row">
        <text class="title">{{ t('support.stationApplyLoc') }}</text>
        <text v-if="izOutService" class="out-tip">{{ t('support.stationOutService') }}</text>
      </view>
      <view class="hint" v-if="configHint">{{ configHint }}</view>
      <view class="addr-row" @click="pickLocation">
        <text class="addr">{{ address || t('support.stationAddress') }}</text>
        <text class="pick">{{ t('support.stationPickLocation') }} ›</text>
      </view>
      <view class="desc-label">{{ t('support.stationDescLabel') }}</view>
      <textarea
        class="area"
        v-model="content"
        maxlength="20"
        :placeholder="t('support.stationDescOptional')"
      />
      <view class="len">{{ content.length }}/20</view>
      <view class="photos">
        <view v-for="(url, i) in imgs" :key="i" class="photo">
          <image :src="url" mode="aspectFill" class="img" />
          <view class="del" @click="imgs.splice(i, 1)">×</view>
        </view>
        <view v-if="imgs.length < 3" class="add" @click="pickImage">+</view>
      </view>
      <view class="btn-primary" :class="{ disabled: izOutService }" @click="onSubmit">
        {{ t('common.confirm') }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getApplyStationConfig, siteApplication } from '@/api/service'
import { getFenceByServiceId, getNearParkingNum } from '@/api/map'
import { uploadFile } from '@/api/common'
import { useTempDataStore } from '@/stores/tempData'
import { useUserStore } from '@/stores/user'
import { useMapLocation } from '@/features/map/useMapLocation'
import { buildFencePolygons, type MapMarker, type MapPolygon } from '@/features/map/mapFenceUtils'
import { resolveAddress } from '@/shared/format'
import { getMapCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { mapProvider } from '@/shared/mapProvider'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const temp = useTempDataStore()
const user = useUserStore()
const { locate } = useMapLocation()
const address = ref('')
const content = ref('')
const configHint = ref('')
const imgs = ref<string[]>([])
const latitude = ref(temp.location?.latitude || 30.25)
const longitude = ref(temp.location?.longitude || 120.15)
const izOutService = ref(false)
const fenceMarkers = ref<MapMarker[]>([])
const polygons = ref<MapPolygon[]>([])

const pinMarker = computed(() => {
  const iconPath = getMapCfg('centerMarker') || getMapCfg('station')
  return {
    id: 1,
    latitude: latitude.value,
    longitude: longitude.value,
    width: 28,
    height: 36,
    ...(iconPath ? { iconPath } : {}),
  }
})

const mapMarkers = computed(() => [pinMarker.value, ...fenceMarkers.value])

onShow(() => setNavTitle(t('support.applyStation')))

function parseOutService(data: Record<string, unknown>): boolean {
  const v = data.izOutService ?? data.outOfService ?? data.isOutService
  return v === true || v === 1 || v === '1' || v === 'true'
}

async function checkServiceArea() {
  const res = await getNearParkingNum({
    locationDTO: { lng: longitude.value, lat: latitude.value },
    serviceId: storage.get('serviceId', ''),
  })
  if (res.success && res.data && typeof res.data === 'object') {
    izOutService.value = parseOutService(res.data as Record<string, unknown>)
  } else {
    izOutService.value = false
  }
}

async function loadServiceFences(sid: string) {
  if (!sid) {
    fenceMarkers.value = []
    polygons.value = []
    return
  }
  try {
    const byId = await getFenceByServiceId({ id: sid })
    if (byId.success && byId.data) {
      const data = byId.data as {
        serviceAreas?: Array<Record<string, unknown>>
        parkings?: Array<Record<string, unknown>>
        noParkings?: Array<Record<string, unknown>>
      }
      const built = buildFencePolygons(data)
      polygons.value = built.polygons
      // Avoid id clash with pin marker id=1
      fenceMarkers.value = built.markers.map((m, i) => ({
        ...m,
        id: m.id === 1 ? 800000 + i : m.id,
      }))
    }
  } catch (e) {
    logger.warn('apply-station fences soft fail', e)
  }
}

async function applyPoi(lat: number, lng: number, addr?: string) {
  latitude.value = lat
  longitude.value = lng
  if (addr) address.value = addr
  else {
    const resolved = await resolveAddress(lat, lng)
    address.value = resolved === '--' ? '' : resolved
  }
  await checkServiceArea()
}

function pickLocation() {
  uni.chooseLocation({
    latitude: latitude.value,
    longitude: longitude.value,
    success: (res) => {
      void applyPoi(res.latitude, res.longitude, res.address || res.name || '')
    },
  })
}

function onMapTap(e: { detail?: { latitude?: number; longitude?: number } }) {
  if (e.detail?.latitude == null || e.detail?.longitude == null) return
  void applyPoi(e.detail.latitude, e.detail.longitude)
}

onMounted(async () => {
  const sid = storage.get<string>('serviceId', '') || ''
  const cfg = await getApplyStationConfig(sid ? { serviceId: sid } : {})
  if (cfg.success && cfg.data) {
    const d = cfg.data as { tip?: string; content?: string; remark?: string }
    configHint.value = String(d.tip || d.content || d.remark || '')
  }
  void loadServiceFences(sid)
  const loc = temp.location || (await locate())
  if (loc) await applyPoi(loc.latitude, loc.longitude)
})

function pickImage() {
  uni.chooseImage({
    count: 3 - imgs.value.length,
    sizeType: ['compressed'],
    sourceType: ['camera', 'album'],
    success: async (res) => {
      for (const path of res.tempFilePaths || []) {
        const up = await uploadFile(path)
        const url = typeof up === 'string' ? up : (up?.data as string | undefined)
        if (url) imgs.value.push(url)
      }
    },
  })
}

async function onSubmit() {
  if (izOutService.value) {
    uni.showToast({ title: t('support.stationOutService'), icon: 'none' })
    return
  }
  // Legacy: description optional — allow empty
  const photoList = imgs.value.length ? imgs.value : ['nophoto']
  const res = await siteApplication({
    address: address.value.trim(),
    location: address.value.trim(),
    content: content.value.trim(),
    applicantRemark: content.value.trim(),
    lat: latitude.value,
    lng: longitude.value,
    imgs: photoList,
    photoUrl: photoList.join(','),
    serviceId: storage.get<string>('serviceId', '') || undefined,
    userPin: user.userInfo?.pin || undefined,
  })
  if (res.success) {
    uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
    address.value = ''
    content.value = ''
    imgs.value = []
    setTimeout(() => navigate('back'), 500)
  }
}
</script>

<style scoped lang="scss">
.map {
  width: 100%;
  height: 42vh;
}
.panel {
  margin-top: -24rpx;
  position: relative;
  z-index: 2;
}
.title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 12rpx;
}
.title {
  font-weight: 700;
}
.out-tip {
  color: #e34d59;
  font-size: 24rpx;
}
.hint {
  color: #888;
  font-size: 24rpx;
  margin-bottom: 20rpx;
  line-height: 1.5;
}
.addr-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.addr {
  flex: 1;
  color: #333;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.pick {
  color: var(--brand-color, #3aa0e8);
  font-size: 24rpx;
  white-space: nowrap;
}
.desc-label {
  font-weight: 600;
  margin-bottom: 12rpx;
}
.area {
  width: 100%;
  min-height: 160rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 20rpx;
  box-sizing: border-box;
}
.len {
  text-align: right;
  color: #999;
  font-size: 22rpx;
  margin: 8rpx 0 20rpx;
}
.photos {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-bottom: 28rpx;
}
.photo {
  position: relative;
  width: 160rpx;
  height: 160rpx;
}
.img {
  width: 100%;
  height: 100%;
  border-radius: 12rpx;
}
.del {
  position: absolute;
  top: -8rpx;
  right: -8rpx;
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  background: #333;
  color: #fff;
  text-align: center;
  line-height: 36rpx;
  font-size: 24rpx;
}
.add {
  width: 160rpx;
  height: 160rpx;
  border-radius: 12rpx;
  background: #f5f6f8;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48rpx;
  color: #999;
}
.btn-primary.disabled {
  opacity: 0.45;
}
</style>
