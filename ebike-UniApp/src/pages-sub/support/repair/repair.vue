<template>
  <view class="page">
    <view class="car-num">
      <input
        class="car-num__input"
        v-model="carId"
        :placeholder="t('account.carIdPlaceholder')"
        @blur="loadConfig"
      />
    </view>

    <view class="record" @tap="goList">
      <text>{{ t('account.viewRepairRecord') }}</text>
      <image v-if="arrowIcon" class="record__arrow" :src="arrowIcon" mode="aspectFit" />
    </view>

    <view class="title-box">
      <view class="title-box__main">{{ t('account.repairPartsPick') }}</view>
      <view class="title-box__sub">{{ t('account.repairPartsHint') }}</view>
    </view>

    <view class="bike-box">
      <view class="bike-box__inner">
        <image v-if="wholeImg" class="bike-box__img" :src="wholeImg" mode="aspectFit" />
        <view class="parts">
          <view
            v-for="(item, index) in baseParts"
            :key="String(item.id)"
            class="part"
            :class="[
              index % 2 === 0 ? 'part--left' : 'part--right',
              {
                on: selectedBaseIds.includes(String(item.id)),
                disabled: item.disabled,
              },
            ]"
            @tap="toggleBase(item)"
          >
            <view
              class="part__check"
              :class="{ on: selectedBaseIds.includes(String(item.id)) }"
              :style="checkStyle(selectedBaseIds.includes(String(item.id)))"
            />
            <text class="part__text">{{ item.content || item.title || item.name }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="otherParts.length" class="section">
      <view class="section__title">{{ t('account.repairOther') }}</view>
      <view class="tags">
        <view
          v-for="item in otherParts"
          :key="String(item.id)"
          class="tag"
          :class="{ on: selectedOtherIds.includes(String(item.id)) }"
          :style="tagStyle(selectedOtherIds.includes(String(item.id)))"
          @tap="toggleOther(item)"
        >
          {{ item.content || item.title || item.name }}
        </view>
      </view>
    </view>

    <view class="section">
      <view class="section__title">{{ t('account.problemDesc') }}</view>
      <view class="desc">
        <textarea
          class="desc__area"
          :value="content"
          :placeholder="t('account.repairPlaceholder')"
          :maxlength="50"
          @input="onInput"
        />
        <text class="desc__count">{{ content.length }}/50</text>
      </view>
    </view>

    <view class="submit-wrap">
      <view
        class="submit"
        :class="{ disabled: submitDisabled }"
        :style="submitStyle"
        @tap="onSubmit"
      >
        {{ t('common.submit') }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getRepairConfigList, submitRepair } from '@/api/repair'
import { getCarInfo } from '@/api/riding'
import { useMapLocation } from '@/features/map/useMapLocation'
import { REPAIR_SORT_TYPES, repairWholeIcon } from '@/features/support/repairIcons'
import { useTempDataStore } from '@/stores/tempData'
import {
  getBrandColor,
  getButtonDisabledColor,
  getButtonWhiteColor,
} from '@/shared/config'
import { getIconCfg } from '@/shared/tenantSkin'
import { consumePendingScanCarId } from '@/features/bike/useScanGate'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

type RepairItem = Record<string, unknown> & { disabled?: boolean }

const { t } = useI18n()
const temp = useTempDataStore()
const { locate } = useMapLocation()

const carId = ref('')
const content = ref('')
const loading = ref(false)
const configList = ref<RepairItem[]>([])
const selectedBaseIds = ref<string[]>([])
const selectedOtherIds = ref<string[]>([])
const carModel = ref('1')

const brand = computed(() => getBrandColor())
const white = computed(() => getButtonWhiteColor())
const disabledColor = computed(() => getButtonDisabledColor())
const wholeImg = computed(() => repairWholeIcon(carModel.value))
const arrowIcon = computed(() => getIconCfg('rightArrowDark') || getIconCfg('newUserInfoRight'))

const baseParts = computed(() => {
  const list = configList.value
    .filter(
      (item) =>
        String(item.tenantId ?? '') === '0' && REPAIR_SORT_TYPES.includes(Number(item.type)),
    )
    .map((item) => ({
      ...item,
      disabled: item.izClientUserRepair === false || item.izClientUserRepair === 0,
    }))
  return list.sort(
    (a, b) => REPAIR_SORT_TYPES.indexOf(Number(a.type)) - REPAIR_SORT_TYPES.indexOf(Number(b.type)),
  )
})

const otherParts = computed(() =>
  configList.value.filter(
    (item) =>
      String(item.tenantId ?? '') !== '0' || !REPAIR_SORT_TYPES.includes(Number(item.type)),
  ),
)

const submitDisabled = computed(
  () => !carId.value.trim() || loading.value,
)

const submitStyle = computed(() => ({
  background: submitDisabled.value ? disabledColor.value : brand.value,
  color: white.value,
}))

function checkStyle(on: boolean) {
  if (!on) return {}
  return {
    background: brand.value,
    borderColor: brand.value,
  }
}

function tagStyle(on: boolean) {
  if (!on) return {}
  return {
    background: brand.value,
    color: white.value,
  }
}

onShow(() => {
  setNavTitle(t('account.repair'))
  const filled = consumePendingScanCarId()
  if (filled) {
    carId.value = filled
    void loadConfig()
  }
})

onLoad((q) => {
  if (q?.carId) carId.value = decodeURIComponent(String(q.carId))
})

async function loadConfig() {
  selectedBaseIds.value = []
  selectedOtherIds.value = []
  const res = await getRepairConfigList({
    ...(carId.value.trim() ? { carId: carId.value.trim() } : {}),
  })
  if (res.success && Array.isArray(res.data)) {
    configList.value = res.data as RepairItem[]
    const first = configList.value[0]
    if (first?.carModel != null) carModel.value = String(first.carModel)
  }
  if (!carId.value.trim()) return
  try {
    const info = await getCarInfo({ carId: carId.value.trim() })
    if (info.success && info.data) {
      const d = info.data as { carModel?: string | number; model?: string | number }
      if (d.carModel != null || d.model != null) carModel.value = String(d.carModel ?? d.model)
    }
  } catch {
    /* soft */
  }
}

onMounted(() => {
  void loadConfig()
})

function onInput(e: { detail?: { value?: string } }) {
  content.value = String(e.detail?.value || '').substring(0, 50)
}

function toggleBase(item: RepairItem) {
  if (item.disabled) return
  const id = String(item.id)
  const idx = selectedBaseIds.value.indexOf(id)
  if (idx >= 0) selectedBaseIds.value.splice(idx, 1)
  else selectedBaseIds.value.push(id)
}

function toggleOther(item: RepairItem) {
  const id = String(item.id)
  const idx = selectedOtherIds.value.indexOf(id)
  if (idx >= 0) selectedOtherIds.value.splice(idx, 1)
  else selectedOtherIds.value.push(id)
}

function goList() {
  navigate('to', '/pages-sub/support/repair/repair-list')
}

async function onSubmit() {
  if (submitDisabled.value) return
  if (!carId.value.trim()) {
    uni.showToast({ title: t('account.carIdPlaceholder'), icon: 'none' })
    return
  }
  const repairPart = [...selectedBaseIds.value, ...selectedOtherIds.value]
  if (!repairPart.length) {
    uni.showToast({ title: t('account.faultType'), icon: 'none' })
    return
  }
  loading.value = true
  try {
    const loc = temp.location || (await locate())
    let imei = ''
    let serviceId = String(storage.get('serviceId', '') || '')
    try {
      const info = await getCarInfo({ carId: carId.value.trim() })
      if (info.success && info.data) {
        const d = info.data as { imei?: string; serviceId?: string }
        imei = String(d.imei || '')
        if (d.serviceId) serviceId = String(d.serviceId)
      }
    } catch {
      /* soft */
    }
    const res = await submitRepair({
      carId: carId.value.trim(),
      imei,
      repairPart,
      reportDesc: content.value,
      reportPhoto: ['nophoto'],
      latitude: loc?.latitude,
      longitude: loc?.longitude,
      serviceId,
    })
    if (res.success) {
      uni.showToast({ title: t('account.repairSuccess'), icon: 'success' })
      setTimeout(() => navigate('back'), 1200)
    } else {
      uni.showToast({ title: res.msg || t('common.submitFail'), icon: 'none' })
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: #fff;
  padding: 32rpx 0 48rpx;
  box-sizing: border-box;
}
.car-num {
  margin: 0 32rpx;
  height: 108rpx;
  background: #f6f6f6;
  border-radius: 32rpx;
  padding: 0 32rpx;
  display: flex;
  align-items: center;
  box-sizing: border-box;
}
.car-num__input {
  width: 100%;
  height: 100rpx;
  font-size: 32rpx;
  color: #333;
}
.record {
  padding: 32rpx 32rpx 0;
  font-size: 32rpx;
  color: #333;
  display: flex;
  align-items: center;
}
.record__arrow {
  width: 26rpx;
  height: 26rpx;
  margin-left: 8rpx;
}
.title-box {
  margin: 32rpx 32rpx 0;
}
.title-box__main {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.title-box__sub {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #999;
}
.bike-box {
  margin: 32rpx 32rpx 0;
  padding-bottom: 32rpx;
  border-radius: 32rpx;
}
.bike-box__inner {
  position: relative;
  min-height: 516rpx;
  margin-top: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.bike-box__img {
  position: absolute;
  width: 430rpx;
  height: 516rpx;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  z-index: 0;
}
.parts {
  position: relative;
  z-index: 1;
  width: 100%;
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  box-sizing: border-box;
}
.part {
  display: flex;
  align-items: center;
  padding: 10rpx 0;
  box-sizing: border-box;
}
.part--left {
  flex-direction: row-reverse;
  margin-right: 40%;
}
.part--right {
  margin-right: 0;
}
.part.disabled {
  opacity: 0.4;
}
.part__check {
  width: 36rpx;
  height: 36rpx;
  border-radius: 8rpx;
  border: 2rpx solid #ccc;
  box-sizing: border-box;
  flex-shrink: 0;
  position: relative;
}
.part__check.on::after {
  content: '';
  position: absolute;
  left: 10rpx;
  top: 4rpx;
  width: 10rpx;
  height: 18rpx;
  border: solid #fff;
  border-width: 0 4rpx 4rpx 0;
  transform: rotate(45deg);
}
.part__text {
  font-size: 28rpx;
  color: #333;
  width: 100rpx;
}
.part--left .part__text {
  margin-right: 10rpx;
  margin-left: 0;
  text-align: right;
}
.part--right .part__text {
  margin-left: 10rpx;
  text-align: left;
}
.section {
  margin: 48rpx 32rpx 0;
}
.section__title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.tags {
  display: flex;
  flex-wrap: wrap;
  margin-top: 6rpx;
}
.tag {
  margin-top: 24rpx;
  margin-right: 24rpx;
  padding: 16rpx 32rpx;
  background: #f5f5f5;
  border-radius: 22rpx;
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
}
.tag.on {
  box-shadow: 0 4rpx 8rpx rgba(99, 209, 68, 0.3);
}
.desc {
  position: relative;
  height: 154rpx;
  margin-top: 32rpx;
  background: #f6f6f6;
  border-radius: 32rpx;
  box-sizing: border-box;
}
.desc__area {
  width: 100%;
  height: 100%;
  padding: 16rpx 32rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}
.desc__count {
  position: absolute;
  bottom: 16rpx;
  right: 32rpx;
  font-size: 28rpx;
  color: #999;
}
.submit-wrap {
  margin: 48rpx 32rpx 100rpx;
}
.submit {
  width: 100%;
  height: 96rpx;
  border-radius: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 600;
}
.submit.disabled {
  pointer-events: none;
}
</style>
