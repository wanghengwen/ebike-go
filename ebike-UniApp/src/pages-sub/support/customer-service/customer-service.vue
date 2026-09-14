<template>
  <view class="page">
    <image v-if="bgImg" class="img" :src="bgImg" mode="aspectFit" />

    <view class="des">
      <text>{{ tips }}</text>
    </view>

    <view class="time">
      <text v-if="izOnlineEntrance" class="leave-msg">{{ t('support.offHoursLeaveMsg') }}</text>
      <text class="work-hours">{{ t('support.workTime') }}：{{ workTime || '--' }}</text>
    </view>

    <view class="button-wrapper">
      <button
        v-if="artificialVisible"
        class="button"
        :style="btnStyle"
        @click="openTelSheet"
      >
        <image v-if="phoneIcon" class="but-img" :src="phoneIcon" mode="aspectFit" />
        {{ t('support.humanService') }}
      </button>
      <!-- #ifdef MP-WEIXIN -->
      <button
        v-if="izOnlineEntrance"
        class="button"
        :class="{ 'button--gap': artificialVisible }"
        :style="btnStyle"
        open-type="contact"
      >
        <image v-if="msgIcon" class="but-img" :src="msgIcon" mode="aspectFit" />
        {{ t('support.onlineService') }}
      </button>
      <!-- #endif -->
    </view>

    <view v-if="telSheetOpen" class="mask" @click="telSheetOpen = false">
      <view class="sheet" @click.stop>
        <view class="popup-content">
          <view
            v-for="(tel, i) in phones"
            :key="i"
            class="service-row"
            @click="onCall(tel)"
          >
            {{ t('support.contactTel', { tel }) }}
          </view>
        </view>
        <view class="cancel-btn" @click="telSheetOpen = false">{{ t('common.cancel') }}</view>
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
import { getBrandColor, getButtonWhiteColor } from '@/shared/config'
import { getIconCfg } from '@/shared/tenantSkin'
import { setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const temp = useTempDataStore()
const { locate } = useMapLocation()

const phones = ref<string[]>([])
const tips = ref('')
const startTime = ref('')
const endTime = ref('')
const workTime = ref('')
const izOnlineEntrance = ref(false)
const izArtificialEntrance = ref(false)
const izWorkTime = ref(true)
const telSheetOpen = ref(false)
let lastCallAt = 0

const bgImg = computed(() => getIconCfg('customerServiceBg'))
const phoneIcon = computed(() => getIconCfg('customerServicePhone'))
const msgIcon = computed(() => getIconCfg('customerServiceMessage'))

/** Legacy: brand green + #1E4A38 text (bwcx / dark-on-green) */
const btnStyle = computed(() => {
  const bg = getBrandColor()
  const white = getButtonWhiteColor()
  const color = white === '#333333' || white === '#1E4A38' ? white : '#1E4A38'
  return {
    background: bg,
    color,
    fontWeight: 'bold',
  }
})

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

/** Legacy artificialEntranceVisible */
const artificialVisible = computed(() => {
  if (!izArtificialEntrance.value) return false
  if (!startTime.value || !endTime.value) return true
  return withinWorkHours.value
})

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
  tips.value = t('support.customerTips')
  const serviceId = await resolveServiceId()
  if (!serviceId) return
  const res = await getCustomerService({ serviceId })
  if (!res.success || !res.data) {
    uni.showToast({ title: t('support.csConfigFail'), icon: 'none' })
    return
  }
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
  izArtificialEntrance.value = Boolean(data.izArtificialEntrance)
  izWorkTime.value = data.izWorkTime !== false
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
    uni.showToast({ title: t('support.callFail'), icon: 'none' })
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
.page {
  width: 100vw;
  min-height: 100vh;
  background: #fff;
  box-sizing: border-box;
}
.img {
  display: block;
  margin-left: 80rpx;
  margin-top: 96rpx;
  width: 588rpx;
  height: 400rpx;
}
.des {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  margin-top: 96rpx;
  margin-left: 30rpx;
  margin-right: 28rpx;
  font-size: 28rpx;
  font-weight: 600;
  color: #333333;
  text {
    word-break: break-word;
    text-align: center;
  }
}
.time {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  margin-top: 32rpx;
  font-size: 24rpx;
  color: #666666;
}
.leave-msg {
  margin-bottom: 16rpx;
  text-align: center;
  padding: 0 40rpx;
}
.work-hours {
  color: #999999;
}
.button-wrapper {
  position: fixed;
  left: 0;
  right: 0;
  bottom: calc(116rpx + env(safe-area-inset-bottom));
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 0 32rpx;
}
.button {
  display: flex;
  flex: 1;
  justify-content: center;
  align-items: center;
  height: 96rpx;
  border-radius: 32rpx;
  font-size: 32rpx;
  font-weight: 600;
  padding: 0;
  margin: 0;
  line-height: 96rpx;
  border: none;
  &::after {
    border: 0;
  }
}
.button--gap {
  margin-left: 20rpx;
}
.but-img {
  width: 40rpx;
  height: 40rpx;
  margin-right: 16rpx;
  flex-shrink: 0;
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
}
.popup-content {
  border-bottom: 10rpx solid #f7f8fa;
}
.service-row {
  height: 90rpx;
  line-height: 90rpx;
  font-size: 28rpx;
  text-align: center;
  color: #666666;
  & + .service-row {
    border-top: 2rpx solid #f6f6f6;
  }
}
.cancel-btn {
  line-height: 80rpx;
  color: #333;
  font-size: 28rpx;
  text-align: center;
  padding-bottom: calc(50rpx + env(safe-area-inset-bottom));
}
</style>
