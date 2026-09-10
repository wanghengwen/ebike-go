<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('support.customerService') }}</view>
      <view v-if="loading">{{ t('common.loading') }}</view>
      <template v-else>
        <view class="tips">{{ tips }}</view>
        <view class="off-hours" v-if="showLeaveMsgTip">{{ t('support.offHoursLeaveMsg') }}</view>
        <view class="time" v-if="workTime">{{ t('support.workTime') }}：{{ workTime }}</view>

        <view class="actions">
          <view
            v-if="artificialVisible"
            class="btn-primary action"
            @click="openTelSheet"
          >
            {{ t('support.humanService') }}
          </view>
          <!-- #ifdef MP-WEIXIN -->
          <button
            v-if="izOnlineEntrance"
            class="btn-primary action contact-btn"
            open-type="contact"
          >
            {{ t('support.onlineService') }}
          </button>
          <!-- #endif -->
        </view>

        <view v-if="!phones.length && !izOnlineEntrance" class="empty">{{ t('common.empty') }}</view>
        <view v-for="(tel, i) in phones" :key="i" class="phone-row" @click="onCall(tel)">
          <text class="phone">{{ tel }}</text>
          <text class="call">{{ t('support.call') }}</text>
        </view>
      </template>
    </view>

    <view v-if="telSheetOpen" class="mask" @click="telSheetOpen = false">
      <view class="sheet" @click.stop>
        <view
          v-for="(tel, i) in phones"
          :key="i"
          class="sheet-row"
          @click="onCall(tel)"
        >
          {{ t('support.contactTel', { tel }) }}
        </view>
        <view class="sheet-cancel" @click="telSheetOpen = false">{{ t('common.cancel') }}</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getCustomerService, getAllService } from '@/api/service'
import { getServiceByPoi } from '@/api/map'
import { useMapLocation } from '@/features/map/useMapLocation'
import { useTempDataStore } from '@/stores/tempData'
import { setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const temp = useTempDataStore()
const { locate } = useMapLocation()
const loading = ref(false)
const phones = ref<string[]>([])
const workTime = ref('')
const tips = ref('')
const startTime = ref('')
const endTime = ref('')
const izOnlineEntrance = ref(false)
const izArtificialEntrance = ref(true)
const izWorkTime = ref(true)
const telSheetOpen = ref(false)
let lastCallAt = 0

onShow(() => setNavTitle(t('support.customerService')))

function formatTime(val?: string) {
  if (!val) return ''
  const parts = val.split(/[:\-]/)
  if (parts.length >= 2) return `${parts[0]}:${parts[1]}`
  return val
}

function minutesOfDay(val?: string): number | null {
  if (!val) return null
  const parts = val.split(/[:\-]/).map((x) => Number(x))
  if (parts.length < 2 || Number.isNaN(parts[0]) || Number.isNaN(parts[1])) return null
  return parts[0] * 60 + parts[1]
}

const withinWorkHours = computed(() => {
  if (!startTime.value || !endTime.value) return true
  const start = minutesOfDay(startTime.value)
  const end = minutesOfDay(endTime.value)
  if (start == null || end == null) return true
  if (start === 0 && end >= 23 * 60 + 59) return true
  const now = new Date()
  const cur = now.getHours() * 60 + now.getMinutes()
  return cur >= start && cur <= end
})

const artificialVisible = computed(() => {
  if (!izArtificialEntrance.value) return false
  if (!startTime.value || !endTime.value) return true
  return withinWorkHours.value
})

const showLeaveMsgTip = computed(() => izOnlineEntrance.value && !withinWorkHours.value)

async function resolveServiceId(): Promise<string | undefined> {
  const cached = storage.get<string>('serviceId', '')
  if (cached) return cached
  const loc = temp.location || (await locate())
  if (loc) {
    const service = await getServiceByPoi({ lat: loc.latitude, lng: loc.longitude })
    const id = (service.data as { id?: string } | undefined)?.id
    if (id) return String(id)
  }
  const all = await getAllService()
  const list = (Array.isArray(all.data) ? all.data : []) as Array<Record<string, unknown>>
  return list[0]?.id ? String(list[0].id) : undefined
}

onMounted(async () => {
  loading.value = true
  const serviceId = await resolveServiceId()
  const res = await getCustomerService(serviceId ? { serviceId } : {})
  loading.value = false
  if (res.success && res.data) {
    const data = res.data as {
      tel?: string
      phone?: string
      mobile?: string
      startTime?: string
      endTime?: string
      tips?: string
      izOnlineEntrance?: boolean
      izArtificialEntrance?: boolean
      izWorkTime?: boolean
    }
    const telStr = String(data.tel || data.phone || data.mobile || '')
    phones.value = telStr
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    startTime.value = data.startTime || ''
    endTime.value = data.endTime || ''
    const start = formatTime(data.startTime)
    const end = formatTime(data.endTime)
    workTime.value = start && end ? `${start}-${end}` : ''
    tips.value = data.tips || t('support.customerTips')
    izOnlineEntrance.value = Boolean(data.izOnlineEntrance)
    izArtificialEntrance.value =
      data.izArtificialEntrance == null ? phones.value.length > 0 : Boolean(data.izArtificialEntrance)
    izWorkTime.value = data.izWorkTime !== false
  }
})

function canCall(): boolean {
  if (!startTime.value || !endTime.value) return true
  if (!izWorkTime.value) return true
  if (!withinWorkHours.value) {
    uni.showToast({ title: t('support.offHoursRest'), icon: 'none' })
    return false
  }
  if (Date.now() - lastCallAt < 1000) return false
  return true
}

function openTelSheet() {
  if (!canCall()) return
  if (!phones.value.length) {
    uni.showToast({ title: t('common.empty'), icon: 'none' })
    return
  }
  telSheetOpen.value = true
}

function onCall(tel: string) {
  if (!tel) return
  if (!canCall()) return
  lastCallAt = Date.now()
  telSheetOpen.value = false
  uni.makePhoneCall({ phoneNumber: tel })
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.tips {
  color: #333;
  line-height: 1.5;
  margin-bottom: 16rpx;
}
.off-hours {
  color: #666;
  font-size: 24rpx;
  margin-bottom: 12rpx;
  line-height: 1.5;
}
.time {
  color: #999;
  font-size: 24rpx;
  margin-bottom: 24rpx;
}
.actions {
  display: flex;
  gap: 20rpx;
  margin-bottom: 24rpx;
}
.action {
  flex: 1;
  text-align: center;
}
.contact-btn {
  margin: 0;
  line-height: normal;
  font-size: inherit;
  &::after {
    border: none;
  }
}
.phone-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 28rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.phone {
  font-size: 36rpx;
  font-weight: 700;
  color: #3aa0e8;
}
.call {
  color: #3aa0e8;
}
.empty {
  color: #999;
}
.mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
}
.sheet {
  width: 100%;
  background: #fff;
  border-radius: 24rpx 24rpx 0 0;
  padding-bottom: calc(24rpx + env(safe-area-inset-bottom));
}
.sheet-row {
  text-align: center;
  padding: 32rpx;
  color: #666;
  border-bottom: 1px solid #f6f6f6;
}
.sheet-cancel {
  text-align: center;
  padding: 32rpx;
  color: #333;
  border-top: 12rpx solid #f7f8fa;
}
</style>
