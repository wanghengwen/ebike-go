<template>
  <view class="page">
    <view class="card self">
      <view class="section-title">{{ t('support.selfService') }}</view>
      <view class="services">
        <view class="service" @click="goNews">
          <image v-if="newsIcon" class="service__icon" :src="newsIcon" mode="aspectFit" />
          <text class="service__label">{{ t('support.latestNews') }}</text>
        </view>
        <view v-if="showRepair" class="service" @click="goRepair">
          <image v-if="repairIcon" class="service__icon" :src="repairIcon" mode="aspectFit" />
          <text class="service__label">{{ t('account.repair') }}</text>
        </view>
        <view v-if="showApplyStation" class="service" @click="goApplyStation">
          <image v-if="stationIcon" class="service__icon" :src="stationIcon" mode="aspectFit" />
          <text class="service__label">{{ t('support.applyStation') }}</text>
        </view>
      </view>
    </view>

    <view class="card guide">
      <view class="section-title">{{ t('support.usageGuide') }}</view>
      <view v-if="faqLoading" class="loading">{{ t('common.loading') }}</view>
      <template v-else-if="questionList.length">
        <view
          v-for="(item, i) in questionList"
          :key="String(item.id || i)"
          class="q-item"
          @click="goFaqDetail(item)"
        >
          <text class="q-label">{{ item.title || item.question || item.name || '-' }}</text>
          <image v-if="arrowIcon" class="arrow" :src="arrowIcon" mode="widthFix" />
        </view>
      </template>
      <template v-else>
        <view
          v-for="(item, i) in fallbackDocs"
          :key="i"
          class="q-item"
          @click="openFallback(item.key)"
        >
          <text class="q-label">{{ t(item.labelKey) }}</text>
          <image v-if="arrowIcon" class="arrow" :src="arrowIcon" mode="widthFix" />
        </view>
      </template>
    </view>

    <view v-if="hasServiceConfig" class="button-wrap">
      <view class="choose-button" :style="contactBtnStyle" @click="goCustomerService">
        {{ t('support.customerService') }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getFaqByServiceId, getAllService, getApplyStationConfig, getCustomerService } from '@/api/service'
import { getConfigBaseItem } from '@/api/user'
import { getServiceByPoi } from '@/api/map'
import { useMapLocation } from '@/features/map/useMapLocation'
import { openProtocol } from '@/shared/protocol'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
import { getBrandColor, getButtonWhiteColor } from '@/shared/config'
import { useTempDataStore } from '@/stores/tempData'

const { t } = useI18n()
const temp = useTempDataStore()
const { locate } = useMapLocation()

const carId = ref('')
const showRepair = ref(true)
const showApplyStation = ref(true)
const faqLoading = ref(true)
const questionList = ref<Array<Record<string, unknown>>>([])
const hasServiceConfig = ref(false)

const repairIcon = computed(() => getIconCfg('newReportIcon') || getIconCfg('report'))
const newsIcon = computed(
  () => getIconCfg('activityInforms') || getIconCfg('newManualIcon') || getIconCfg('systemInform'),
)
const stationIcon = computed(
  () => getIconCfg('newPepairIcon') || getIconCfg('newReportIcon'),
)
const arrowIcon = computed(() => getMapCfg('iconRight') || getMapCfg('iconRightRound'))
const contactBtnStyle = computed(() => {
  const bg = getBrandColor()
  return {
    backgroundColor: bg,
    borderColor: bg,
    color: getButtonWhiteColor() === '#FFFFFF' ? '#1E4A38' : getButtonWhiteColor(),
    fontWeight: 'bold',
  }
})

const fallbackDocs = [
  { key: 'registeredDesc', labelKey: 'support.faqRegister' },
  { key: 'depositAndBalance', labelKey: 'support.faqDeposit' },
  { key: 'useEbike', labelKey: 'support.faqUseBike' },
  { key: 'vehicleProblem', labelKey: 'support.faqBikeIssue' },
  { key: 'ridingInstructions', labelKey: 'support.faqRidingTips' },
] as const

onShow(() => setNavTitle(t('account.help')))

onLoad(async (q) => {
  carId.value = String(q?.carId || '')
  const token = storage.get<Record<string, unknown>>('loginInfo', {})?.accessToken
  if (!token) {
    uni.showModal({
      title: t('auth.loginTitle'),
      content: t('account.needLogin'),
      showCancel: false,
      confirmText: t('auth.loginNow'),
      success: (res) => {
        if (res.confirm) navigate('redirect', '/pages/auth/quick-login')
      },
    })
    return
  }
  await Promise.all([loadRepairFlag(), loadApplyStation(), loadFaq(), loadCsConfig()])
})

async function loadCsConfig() {
  try {
    const serviceId = await resolveServiceId()
    if (!serviceId) return
    const res = await getCustomerService({ serviceId })
    if (res.success && res.data) {
      const data = res.data as { izOnlineEntrance?: boolean; izArtificialEntrance?: boolean }
      hasServiceConfig.value = Boolean(data.izOnlineEntrance || data.izArtificialEntrance)
    }
  } catch {
    hasServiceConfig.value = false
  }
}

async function resolveServiceId(): Promise<string | undefined> {
  const cached = storage.get<string>('serviceId', '')
  if (cached) return cached
  const loc = temp.location || (await locate())
  if (loc) {
    const service = await getServiceByPoi({ lat: loc.latitude, lng: loc.longitude })
    const id = (service.data as { id?: string } | undefined)?.id
    if (id) {
      storage.set('serviceId', id)
      return id
    }
  }
  const all = await getAllService()
  const rows = (Array.isArray(all.data) ? all.data : []) as Array<Record<string, unknown>>
  return rows[0]?.id ? String(rows[0].id) : undefined
}

async function loadRepairFlag() {
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getConfigBaseItem(sid ? { serviceId: sid } : {})
    if (res.success && res.data) {
      const data = res.data as { izCanInitiativeRepair?: boolean | number }
      // Legacy may hide when false; default show for screenshot parity when unset
      if (data.izCanInitiativeRepair != null) {
        showRepair.value = Boolean(data.izCanInitiativeRepair)
      }
    }
  } catch {
    showRepair.value = true
  }
}

async function loadApplyStation() {
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getApplyStationConfig(sid ? { serviceId: sid } : {})
    if (res.success && res.data != null) {
      const data = res.data as boolean | { enable?: boolean; izOpen?: boolean }
      if (typeof data === 'boolean') showApplyStation.value = data
      else showApplyStation.value = Boolean(data.enable ?? data.izOpen ?? true)
    }
  } catch {
    showApplyStation.value = true
  }
}

async function loadFaq() {
  faqLoading.value = true
  try {
    const serviceId = await resolveServiceId()
    const res = await getFaqByServiceId(serviceId ? { serviceId } : {})
    if (res.success) {
      const data = res.data as { records?: Array<Record<string, unknown>> } | Array<Record<string, unknown>>
      questionList.value = Array.isArray(data) ? data : data?.records || []
    }
  } finally {
    faqLoading.value = false
  }
}

function goNews() {
  navigate('to', '/pages-sub/support/messages/list')
}

function goRepair() {
  const q = carId.value ? `?carId=${encodeURIComponent(carId.value)}` : ''
  navigate('to', `/pages-sub/support/repair/repair${q}`)
}

function goApplyStation() {
  navigate('to', '/pages-sub/support/apply-station/apply-station')
}

function goFaqDetail(item: Record<string, unknown>) {
  const id = item.id || item.faqId
  if (!id) return
  navigate('to', `/pages-sub/support/faq/faq-detail?id=${encodeURIComponent(String(id))}&type=2`)
}

function openFallback(key: string) {
  openProtocol(key)
}

function goCustomerService() {
  navigate('to', '/pages-sub/support/customer-service/customer-service')
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: #f7f8fa;
  padding: 24rpx 0 180rpx;
  box-sizing: border-box;
}
.card {
  background: #fff;
  border-radius: 24rpx;
  margin: 0 24rpx 24rpx;
  padding: 32rpx 32rpx 24rpx;
}
.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #000;
  margin-bottom: 32rpx;
}
.services {
  display: flex;
  flex-direction: row;
  justify-content: flex-start;
  gap: 56rpx;
  padding-bottom: 16rpx;
}
.service {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 160rpx;
}
.service__icon {
  width: 112rpx;
  height: 112rpx;
  border-radius: 24rpx;
}
.service__label {
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #333;
  text-align: center;
}
.loading {
  color: #999;
  font-size: 26rpx;
  padding: 16rpx 0;
}
.q-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  padding: 32rpx 0;
  border-bottom: 2rpx solid #f6f6f6;
}
.q-item:last-child {
  border-bottom: none;
}
.q-label {
  font-size: 30rpx;
  color: #333;
  font-weight: 500;
}
.arrow {
  width: 28rpx;
  height: 28rpx;
  flex-shrink: 0;
}
.button-wrap {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 24rpx 48rpx calc(24rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -4rpx 16rpx rgba(0, 0, 0, 0.04);
}
.choose-button {
  height: 96rpx;
  border-radius: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
}
</style>
