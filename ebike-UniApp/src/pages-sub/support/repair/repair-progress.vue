<template>
  <view class="page">
    <view class="card status">
      <image v-if="statusImg" class="status-img" :src="statusImg" mode="widthFix" />
      <view class="status-text">{{ statusText }}</view>
    </view>
    <view class="card">
      <view class="row">
        <image v-if="bikeImg" class="bike" :src="bikeImg" mode="aspectFit" />
        <text class="label">{{ bikeTypeLabel }}</text>
        <text class="val">NO.{{ details.carId || '-' }}</text>
      </view>
      <view class="row" v-if="details.address">
        <image v-if="locIcon" class="loc" :src="locIcon" mode="aspectFit" />
        <text class="label">{{ t('account.repairAddress') }}</text>
        <text class="val">{{ details.address }}</text>
      </view>
      <view class="row" v-if="timeText">
        <text class="label">{{ t('account.repairTime') }}</text>
        <text class="val">{{ timeText }}</text>
      </view>
      <view class="block" v-if="partNames.length">
        <view class="label">{{ t('account.repairParts') }}</view>
        <view class="tags">
          <text v-for="(n, i) in partNames" :key="i" class="tag">{{ n }}</text>
        </view>
      </view>
      <view class="block" v-if="otherNames.length">
        <view class="label">{{ t('account.repairOther') }}</view>
        <view class="tags">
          <text v-for="(n, i) in otherNames" :key="i" class="tag">{{ n }}</text>
        </view>
      </view>
      <view class="block" v-if="details.reportDesc">
        <view class="label">{{ t('account.repairPlaceholder') }}</view>
        <view class="desc">{{ details.reportDesc }}</view>
      </view>
      <view class="block" v-if="photos.length">
        <view class="label">{{ t('account.repairPhotos') }}</view>
        <view class="imgs">
          <image
            v-for="(src, i) in photos"
            :key="i"
            class="img"
            :src="src"
            mode="aspectFill"
            @click="preview(i)"
          />
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getRepairConfigList } from '@/api/repair'
import { repairBikeIcon, repairStatusIcon } from '@/features/support/repairIcons'
import { getIconCfg } from '@/shared/tenantSkin'
import { setNavTitle } from '@/shared/navigate'
import { logger } from '@/shared/logger'

const SORT_ARR = [6, 13, 0, 7, 3, 8, 1, 10, 2, 14, 4, 15, 11, 12, 5, 9]

const { t } = useI18n()
const details = ref<Record<string, unknown>>({})
const partNames = ref<string[]>([])
const otherNames = ref<string[]>([])

const statusText = computed(() => {
  const s = Number(details.value.state)
  if (s === 2) return t('account.repairProcessed')
  return t('account.repairProcessing')
})
const statusImg = computed(() => repairStatusIcon(details.value.state))
const bikeImg = computed(() => repairBikeIcon(details.value.bikeType))
const locIcon = computed(() => getIconCfg('locationOutline'))
const bikeTypeLabel = computed(() =>
  details.value.bikeType ? t('account.bikeNormal') : t('account.bikeElectric'),
)
const timeText = computed(() => {
  const raw = details.value.reportTime || details.value.createTime
  if (!raw) return ''
  const d = new Date(Number(raw) || String(raw))
  if (Number.isNaN(d.getTime())) return String(raw)
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}.${p(d.getMonth() + 1)}.${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
})
const photos = computed(() => {
  const p = details.value.reportPhoto
  if (Array.isArray(p)) return p.map(String)
  if (typeof p === 'string' && p) return p.split(',').filter(Boolean)
  return []
})

onShow(() => setNavTitle(t('account.repairProgress')))

onLoad(async (q) => {
  try {
    const raw = q?.json ? decodeURIComponent(String(q.json)) : ''
    if (raw) details.value = JSON.parse(raw)
  } catch (e) {
    logger.warn('repair progress parse fail', e)
  }
  await resolveParts()
})

async function resolveParts() {
  const carId = String(details.value.carId || '')
  const parts = (details.value.repairPart || details.value.faultType || []) as unknown
  const ids = Array.isArray(parts) ? parts.map(String) : String(parts || '').split(',').filter(Boolean)
  if (!ids.length) return
  try {
    const res = await getRepairConfigList(carId ? { carId } : {})
    const list = (Array.isArray(res.data) ? res.data : (res.data as { list?: unknown[] })?.list || []) as Array<
      Record<string, unknown>
    >
    const primary: string[] = []
    const other: string[] = []
    for (const id of ids) {
      const hit = list.find((c) => String(c.id) === String(id) || String(c.configId) === String(id))
      const name = String(hit?.content || hit?.name || id)
      const type = Number(hit?.type)
      const tenantId = String(hit?.tenantId ?? '')
      if (tenantId === '0' && SORT_ARR.includes(type)) primary.push(name)
      else other.push(name)
    }
    partNames.value = primary
    otherNames.value = other
  } catch (e) {
    logger.warn('repair config soft fail', e)
    partNames.value = ids
  }
}

function preview(i: number) {
  uni.previewImage({ urls: photos.value, current: photos.value[i] })
}
</script>

<style scoped lang="scss">
.status {
  text-align: center;
}
.status-img {
  width: 160rpx;
  margin: 0 auto 12rpx;
  display: block;
}
.status-text {
  font-size: 32rpx;
  font-weight: 600;
  color: #3aa0e8;
}
.row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16rpx;
  padding: 16rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.bike {
  width: 48rpx;
  height: 48rpx;
  flex-shrink: 0;
}
.loc {
  width: 28rpx;
  height: 28rpx;
  flex-shrink: 0;
}
.label {
  color: #888;
  flex-shrink: 0;
}
.val {
  text-align: right;
  flex: 1;
}
.block {
  margin-top: 24rpx;
}
.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 12rpx;
}
.tag {
  background: #f5f6f8;
  padding: 8rpx 16rpx;
  border-radius: 8rpx;
  font-size: 24rpx;
}
.desc {
  margin-top: 12rpx;
  color: #555;
  line-height: 1.5;
}
.imgs {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-top: 12rpx;
}
.img {
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
}
</style>
