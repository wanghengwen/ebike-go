<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('support.faq') }}</view>
      <view v-if="loading">{{ t('common.loading') }}</view>
      <template v-else-if="list.length">
        <view
          v-for="(item, i) in list"
          :key="String(item.id || i)"
          class="item"
          @click="goDetail(item)"
        >
          <text class="label">{{ item.title || item.question || item.name || '-' }}</text>
          <image v-if="arrowIcon" class="arrow" :src="arrowIcon" mode="widthFix" />
        </view>
      </template>
      <template v-else>
        <view
          v-for="(item, i) in fallbackDocs"
          :key="i"
          class="item"
          @click="openFallback(item.key)"
        >
          <text class="label">{{ t(item.labelKey) }}</text>
          <image v-if="arrowIcon" class="arrow" :src="arrowIcon" mode="widthFix" />
        </view>
      </template>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getFaqByServiceId, getAllService } from '@/api/service'
import { getServiceByPoi } from '@/api/map'
import { useMapLocation } from '@/features/map/useMapLocation'
import { openProtocol } from '@/shared/protocol'
import { useTempDataStore } from '@/stores/tempData'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { getMapCfg } from '@/shared/tenantSkin'

const { t } = useI18n()
const temp = useTempDataStore()
const { locate } = useMapLocation()
const loading = ref(false)
const list = ref<Array<Record<string, unknown>>>([])
const arrowIcon = computed(() => getMapCfg('iconRight'))

const fallbackDocs = [
  { key: 'registeredDesc', labelKey: 'support.faqRegister' },
  { key: 'depositAndBalance', labelKey: 'support.faqDeposit' },
  { key: 'useEbike', labelKey: 'support.faqUseBike' },
  { key: 'vehicleProblem', labelKey: 'support.faqBikeIssue' },
  { key: 'ridingInstructions', labelKey: 'support.faqRidingTips' },
] as const

onShow(() => setNavTitle(t('support.faq')))

async function resolveServiceId(): Promise<string | undefined> {
  const cached = storage.get<string>('serviceId', '')
  if (cached) return cached
  const loc = temp.location || (await locate())
  if (loc) {
    const service = await getServiceByPoi({ lat: loc.latitude, lng: loc.longitude })
    const id = (service.data as { id?: string } | undefined)?.id
    if (id) return id
  }
  const all = await getAllService()
  const rows = (Array.isArray(all.data) ? all.data : []) as Array<Record<string, unknown>>
  return rows[0]?.id ? String(rows[0].id) : undefined
}

onMounted(async () => {
  loading.value = true
  try {
    const serviceId = await resolveServiceId()
    const res = await getFaqByServiceId(serviceId ? { serviceId } : {})
    if (res.success) {
      const data = res.data as { records?: Array<Record<string, unknown>> } | Array<Record<string, unknown>>
      list.value = Array.isArray(data) ? data : data?.records || []
    }
  } finally {
    loading.value = false
  }
})

function goDetail(item: Record<string, unknown>) {
  const id = item.id || item.faqId
  if (!id) return
  navigate('to', `/pages-sub/support/faq/faq-detail?id=${encodeURIComponent(String(id))}&type=2`)
}

function openFallback(key: string) {
  openProtocol(key)
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 28rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.label {
  flex: 1;
  padding-right: 16rpx;
}
.arrow {
  width: 34rpx;
  flex-shrink: 0;
}
</style>
