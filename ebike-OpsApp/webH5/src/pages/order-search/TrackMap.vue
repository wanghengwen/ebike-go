<script setup lang="ts">
import { onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import type { TrackPoint } from '@/api'
import {
  hasMapKey,
  loadTencentMap,
  type TMapInstance,
  type TMapLatLng,
  type TMapNamespace,
  type TMapOverlay,
} from '@/composables/tencentMap'

interface Coord {
  lat: number | null | undefined
  lng: number | null | undefined
}

const props = defineProps<{
  points?: TrackPoint[] | null
  start?: Coord | null
  end?: Coord | null
}>()

const container = ref<HTMLDivElement | null>(null)
const failed = ref(false)

// 地图实例和覆盖物不该被 Vue 深度代理——SDK 内部持有大量自引用对象，
// 变成 reactive 后既慢又会触发它自己的相等性判断。
const sdk = shallowRef<TMapNamespace | null>(null)
const map = shallowRef<TMapInstance | null>(null)
const polyline = shallowRef<TMapOverlay | null>(null)
const markers = shallowRef<TMapOverlay | null>(null)

function valid(coord: Coord | null | undefined): coord is { lat: number; lng: number } {
  return (
    !!coord &&
    typeof coord.lat === 'number' &&
    typeof coord.lng === 'number' &&
    // 后端拿不到定位时回 0，落在几内亚湾，画出来是条横跨半个地球的线。
    (coord.lat !== 0 || coord.lng !== 0)
  )
}

function path(): Coord[] {
  return (props.points ?? []).filter(valid)
}

function endpoints(): Coord[] {
  return [props.start, props.end].filter(valid)
}

function draw(): void {
  const TMap = sdk.value
  const instance = map.value
  if (!TMap || !instance) return

  const line = path().map((p) => new TMap.LatLng(p.lat as number, p.lng as number))
  const pins = endpoints().map((p) => new TMap.LatLng(p.lat as number, p.lng as number))

  polyline.value?.setGeometries(line.length > 1 ? [{ id: 'track', paths: line }] : [])
  markers.value?.setGeometries(pins.map((position, index) => ({ id: `p${index}`, position })))

  const all: TMapLatLng[] = line.length > 1 ? line : pins
  if (all.length === 0) return
  if (all.length === 1) {
    instance.setCenter(all[0]!)
    return
  }

  const lats = all.map((p) => p.getLat())
  const lngs = all.map((p) => p.getLng())
  instance.fitBounds(
    new TMap.LatLngBounds(
      new TMap.LatLng(Math.min(...lats), Math.min(...lngs)),
      new TMap.LatLng(Math.max(...lats), Math.max(...lngs)),
    ),
    { padding: 40 },
  )
}

async function init(): Promise<void> {
  if (!container.value || map.value) return
  try {
    const TMap = await loadTencentMap()
    sdk.value = TMap
    // 等待期间组件可能已经卸载。
    if (!container.value) return

    const instance = new TMap.Map(container.value, { zoom: 14 })
    map.value = instance
    polyline.value = new TMap.MultiPolyline({ map: instance, geometries: [] })
    markers.value = new TMap.MultiMarker({ map: instance, geometries: [] })
    draw()
  } catch {
    failed.value = true
  }
}

watch(
  () => [props.points, props.start, props.end],
  () => (map.value ? draw() : void init()),
  { immediate: true, flush: 'post' },
)

onBeforeUnmount(() => {
  polyline.value?.destroy()
  markers.value?.destroy()
  map.value?.destroy()
  map.value = null
})
</script>

<template>
  <div class="track">
    <div v-if="!hasMapKey()" class="track__off">未配置地图 Key</div>
    <div v-else-if="failed" class="track__off">地图加载失败</div>
    <div v-else ref="container" class="track__canvas" />
  </div>
</template>

<style scoped>
.track {
  height: 180px;
  margin: 10px 12px 0;
  border-radius: 8px;
  overflow: hidden;
  background: #e9ecf0;
}

.track__canvas {
  width: 100%;
  height: 100%;
}

.track__off {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: #999999;
}
</style>
