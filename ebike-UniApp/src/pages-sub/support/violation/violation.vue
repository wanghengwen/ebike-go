<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('support.violation') }}</view>
      <view class="car-row">
        <input class="input" v-model="carId" :placeholder="t('account.carIdPlaceholder')" />
        <view class="scan-btn" @click="onScanFill">
          <image v-if="scanIcon" class="scan-icon" :src="scanIcon" mode="aspectFit" />
          <text v-else>{{ t('ride.scanFill') }}</text>
        </view>
      </view>
      <view class="label">{{ t('support.violationType') }}</view>
      <view class="tags">
        <view
          v-for="item in types"
          :key="item.value"
          class="tag"
          :class="{ on: selected.includes(item.value) }"
          @click="toggle(item.value)"
        >
          {{ item.label }}
        </view>
      </view>
      <textarea
        class="area"
        v-model="content"
        :placeholder="t('support.violationDesc')"
        maxlength="100"
      />
      <view class="photos">
        <view v-for="(url, i) in imgs" :key="i" class="photo">
          <image :src="url" mode="aspectFill" class="img" />
          <view class="del" @click="imgs.splice(i, 1)">×</view>
        </view>
        <view v-if="imgs.length < 3" class="add" @click="pickImage">+</view>
      </view>
      <view class="btn-primary" @click="onSubmit">{{ t('common.confirm') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { submitSneak } from '@/api/repair'
import { getLastDetail } from '@/api/pay'
import { uploadFile } from '@/api/common'
import { useUserStore } from '@/stores/user'
import { useScanGate, consumePendingScanCarId } from '@/features/bike/useScanGate'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const user = useUserStore()
const { scanForFillCarId } = useScanGate()
const content = ref('')
const carId = ref('')
const orderId = ref('')
const imgs = ref<string[]>([])
const selected = ref<string[]>([])
const scanIcon = computed(() => getMapCfg('scanCode1') || getIconCfg('scan'))

/** Legacy reportTypeMap values (4=恶意涂画, 12=其他) */
const types = computed(() => [
  { value: '0', label: t('support.vPrivateLock') },
  { value: '1', label: t('support.vOverload') },
  { value: '2', label: t('support.vIllegalPark') },
  { value: '3', label: t('support.vUnderage') },
  { value: '4', label: t('support.vGraffiti') },
  { value: '5', label: t('support.vDamage') },
  { value: '6', label: t('support.vFallen') },
  { value: '7', label: t('support.vWrongWay') },
  { value: '8', label: t('support.vHelmetLost') },
  { value: '9', label: t('support.vMotorLane') },
  { value: '13', label: t('support.vNonMotorLane') },
  { value: '10', label: t('support.vOutFence') },
  { value: '11', label: t('support.vAnglePark') },
  { value: '12', label: t('support.vOther') },
])

onShow(() => {
  setNavTitle(t('support.violation'))
  const filled = consumePendingScanCarId()
  if (filled) carId.value = filled
})

onLoad((q) => {
  orderId.value = decodeURIComponent((q?.orderId as string) || '')
  if (q?.carId) carId.value = decodeURIComponent(String(q.carId))
})

async function onScanFill() {
  await scanForFillCarId()
}

function toggle(v: string) {
  const i = selected.value.indexOf(v)
  if (i >= 0) selected.value.splice(i, 1)
  else selected.value.push(v)
}

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
  if (!selected.value.length && !content.value.trim()) {
    uni.showToast({ title: t('support.violationDesc'), icon: 'none' })
    return
  }
  if (!carId.value.trim()) {
    uni.showToast({ title: t('account.carIdPlaceholder'), icon: 'none' })
    return
  }
  const pin = String(user.userInfo?.pin || storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
  let reportedPin = ''
  let itinId = orderId.value
  // Legacy: getLastDetail by carId to find reported user on that bike
  try {
    const last = await getLastDetail({ carId: carId.value.trim(), izNewApp: true })
    if (last.success && last.data) {
      const d = last.data as Record<string, unknown>
      reportedPin = String(d.userPin || d.pin || '')
      itinId = String(d.id || d.orderId || itinId || '')
    }
  } catch {
    /* soft */
  }
  const res = await submitSneak({
    reported_user_pin: reportedPin,
    description: content.value.trim(),
    car_id: carId.value.trim(),
    imgs: imgs.value.length ? imgs.value : ['nophoto'],
    report_man_pin: pin,
    report_role: 2,
    itin_id: itinId || undefined,
    report_type: { type: selected.value },
  })
  if (res.success) {
    uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
    setTimeout(() => navigate('back'), 500)
  }
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.car-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 20rpx;
}
.input {
  flex: 1;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 0;
}
.scan-btn {
  flex-shrink: 0;
  min-width: 72rpx;
  height: 72rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #3aa0e8;
  font-size: 24rpx;
}
.scan-icon {
  width: 56rpx;
  height: 56rpx;
}
.label {
  margin-bottom: 12rpx;
  color: #666;
  font-size: 26rpx;
}
.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-bottom: 20rpx;
}
.tag {
  padding: 12rpx 24rpx;
  border-radius: 8rpx;
  background: #f5f6f8;
  font-size: 26rpx;
  color: #333;
}
.tag.on {
  background: var(--brand-color, #3aa0e8);
  color: #fff;
}
.area {
  width: 100%;
  min-height: 200rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-bottom: 20rpx;
  box-sizing: border-box;
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
</style>
