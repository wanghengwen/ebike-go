<template>
  <!-- Legacy objection.vue -->
  <view class="page">
    <scroll-view class="scroll" scroll-y>
      <view class="explain">
        <text>{{ t('support.objectionHint') }}</text>
      </view>

      <view class="ques_container">
        <view class="ques_title">{{ t('support.objectionType') }}</view>
        <view class="type">
          <text
            v-for="(item, index) in reasons"
            :key="item.value"
            class="item"
            :class="{ border: currIndex !== index }"
            :style="tagStyle(index)"
            @click="chooseReason(index)"
          >
            {{ t(item.titleKey) }}
          </text>
        </view>
      </view>

      <view class="des_container">
        <view class="ques_title">{{ t('support.objectionDescTitle') }}</view>
        <view class="sec_content">
          <textarea
            v-model="quesContent"
            class="sec_content_txt"
            maxlength="50"
            :placeholder="t('support.objectionDesc')"
          />
          <text class="corner_marker">{{ quesContent.length }}/50</text>
        </view>
      </view>

      <view class="photo_container">
        <view class="ques_title">{{ t('support.objectionPhoto') }}</view>
        <view class="photo-wrap">
          <view v-for="(item, index) in imgs" :key="index" class="img-wrap">
            <image :src="item" class="img-item" mode="aspectFill" @click="preview(index)" />
            <view class="delete" @click="delImage(index)">
              <image v-if="deleteIcon" class="delete-icon" :src="deleteIcon" mode="widthFix" />
              <text v-else class="delete-fallback">×</text>
            </view>
          </view>
          <image
            v-if="imgs.length < 3 && addIcon"
            class="add"
            :src="addIcon"
            mode="aspectFit"
            @click="shooting"
          />
          <view v-else-if="imgs.length < 3" class="add add--fallback" @click="shooting">+</view>
        </view>
      </view>
    </scroll-view>

    <view class="btn-box">
      <button
        class="btn-submit"
        :style="submitButtonStyle"
        :disabled="isDisabled"
        @click="uploadAndSubmit"
      >
        {{ t('common.submit') }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { createUserTicket } from '@/api/order'
import { getConfigBaseItem } from '@/api/user'
import {
  getBrandColor,
  getButtonDisabledColor,
  getButtonWhiteColor,
  getTenantConfig,
} from '@/shared/config'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { getIconCfg } from '@/shared/tenantSkin'
import { uploadFile } from '@/shared/upload'
import { logger } from '@/shared/logger'

const { t } = useI18n()

const reasons = [
  { titleKey: 'support.objUnlockFail', value: 3 },
  { titleKey: 'support.objBikeFault', value: 4 },
  { titleKey: 'support.objNoParking', value: 1 },
  { titleKey: 'support.objOutService', value: 0 },
  { titleKey: 'support.objParkPoint', value: 2 },
]

const quesContent = ref('')
const currIndex = ref<number | null>(null)
const imgs = ref<string[]>([])
const isDisabled = ref(true)
const orderId = ref('')
const chooseImgWay = ref<Array<'camera' | 'album'>>(['camera'])
const submitting = ref(false)

const deleteIcon = computed(() => getIconCfg('iconDelete'))
const addIcon = computed(() => getIconCfg('uploadImg') || getIconCfg('add_photo') || '')
const brand = computed(() => getBrandColor())
const white = computed(() => getButtonWhiteColor())

const submitButtonStyle = computed(() => {
  const disabled = isDisabled.value
  const bg = disabled ? getButtonDisabledColor() : brand.value
  const color = disabled ? white.value : '#1E4A38'
  return `background-color:${bg};border-color:${bg};color:${color};font-weight:bold;`
})

onShow(() => setNavTitle(t('support.objection')))

onLoad(async (q) => {
  orderId.value = decodeURIComponent(String(q?.orderId || ''))
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await getConfigBaseItem(sid ? { serviceId: sid } : {})
    if (res.success && res.data) {
      const ways = (res.data as { userTicketPhotoWays?: number[] }).userTicketPhotoWays
      const mapped = (ways || [])
        .map((i) => (['camera', 'album'] as const)[i])
        .filter(Boolean) as Array<'camera' | 'album'>
      if (mapped.length) chooseImgWay.value = mapped
    }
  } catch (e) {
    logger.warn('objection getConfigBaseItem soft fail', e)
  }
})

function tagStyle(index: number) {
  if (currIndex.value === index) {
    return {
      backgroundColor: brand.value,
      color: white.value,
    }
  }
  return {}
}

function chooseReason(index: number) {
  if (currIndex.value === index) {
    currIndex.value = null
    isDisabled.value = true
  } else {
    currIndex.value = index
    isDisabled.value = false
  }
}

function preview(i: number) {
  uni.previewImage({ urls: [imgs.value[i]] })
}

function delImage(i: number) {
  imgs.value.splice(i, 1)
}

function shooting() {
  if (imgs.value.length >= 3) {
    uni.showToast({ title: t('support.objectionPhotoMax'), icon: 'none' })
    return
  }
  const maxSize = Number((getTenantConfig() as { fileSize?: number }).fileSize) || 10485760
  uni.chooseImage({
    count: 3 - imgs.value.length,
    sizeType: ['compressed'],
    sourceType: chooseImgWay.value,
    success: async (res) => {
      const tempFiles = (res.tempFiles || []) as Array<{ path?: string; size?: number }>
      let oversized = false
      for (const item of tempFiles) {
        if (Number(item.size || 0) > maxSize) {
          oversized = true
          continue
        }
        const path = item.path || ''
        if (!path) continue
        const url = await uploadFile(path)
        if (url) imgs.value.push(url)
      }
      if (oversized) {
        uni.showToast({ title: t('support.objectionUploadFail'), icon: 'none' })
      }
    },
  })
}

async function uploadAndSubmit() {
  if (submitting.value) return
  if (isDisabled.value) {
    uni.showToast({ title: t('support.objectionDupSubmit'), icon: 'none' })
    return
  }
  if (currIndex.value == null) {
    uni.showToast({ title: t('support.objectionNeedReason'), icon: 'none' })
    return
  }
  if (imgs.value.length <= 0) {
    uni.showToast({ title: t('support.objectionNeedPhoto'), icon: 'none' })
    return
  }
  if (!orderId.value) {
    uni.showToast({ title: t('account.orders'), icon: 'none' })
    return
  }
  submitting.value = true
  isDisabled.value = true
  const reason = reasons[currIndex.value]
  try {
    const res = await createUserTicket({
      photoUrl: imgs.value,
      orderId: orderId.value,
      initiator: 0,
      userReason: `${t(reason.titleKey)}、${quesContent.value}`,
    })
    if (!res.success) {
      uni.showToast({ title: t('common.submitFail'), icon: 'none' })
      isDisabled.value = false
      return
    }
    setTimeout(() => {
      uni.showToast({ title: t('support.objectionSuccess'), icon: 'none' })
    }, 500)
    navigate('back')
  } catch {
    uni.showToast({ title: t('common.submitFail'), icon: 'none' })
    isDisabled.value = false
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  padding-top: 52rpx;
  background: #fff;
  box-sizing: border-box;
}
.scroll {
  height: calc(100% - 220rpx);
  box-sizing: border-box;
}
.explain {
  margin-left: 48rpx;
  margin-right: 48rpx;
  padding: 16rpx 32rpx;
  font-size: 28rpx;
  font-weight: 400;
  color: #666;
  background-color: rgba(99, 209, 68, 0.1);
  border-radius: 32rpx;
  line-height: 35rpx;
}
.ques_title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.ques_container {
  margin: 32rpx 48rpx 0;
}
.type {
  display: flex;
  flex-wrap: wrap;
}
.item {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 25rpx;
  margin-right: 24rpx;
  padding: 20rpx;
  font-size: 28rpx;
  font-weight: 500;
  color: #666;
  border-radius: 30rpx;
}
.border {
  border: 2rpx solid #e5e5e5;
}
.des_container {
  margin: 48rpx 48rpx 0;
}
.sec_content {
  position: relative;
  width: 654rpx;
  height: 154rpx;
  margin-top: 32rpx;
  background: #f6f6f6;
  border-radius: 32rpx;
}
.sec_content_txt {
  width: 100%;
  height: 154rpx;
  padding: 16rpx 32rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}
.corner_marker {
  position: absolute;
  bottom: 13rpx;
  right: 32rpx;
  color: #999;
  font-size: 24rpx;
}
.photo_container {
  margin: 48rpx 48rpx 0;
  padding-bottom: 40rpx;
}
.photo-wrap {
  position: relative;
  margin-top: 32rpx;
  display: flex;
  flex-wrap: wrap;
}
.img-wrap {
  position: relative;
  margin-bottom: 20rpx;
  margin-right: 32rpx;
}
.img-item {
  height: 120rpx;
  width: 120rpx;
  border-radius: 29rpx;
}
.delete {
  position: absolute;
  top: -6px;
  right: -3px;
  width: 40rpx;
  height: 40rpx;
  display: flex;
  justify-content: center;
  align-items: center;
}
.delete-icon {
  height: 32rpx;
  width: 32rpx;
}
.delete-fallback {
  width: 32rpx;
  height: 32rpx;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 24rpx;
  line-height: 32rpx;
  text-align: center;
}
.add {
  display: inline-block;
  height: 120rpx;
  width: 120rpx;
  background: #f6f6f6;
  border-radius: 29rpx;
}
.add--fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2rpx solid #e5e5e5;
  font-size: 64rpx;
  color: #696969;
  box-sizing: border-box;
}
.btn-box {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 100rpx;
  display: flex;
  justify-content: center;
}
.btn-submit {
  width: 654rpx;
  height: 96rpx;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 32rpx;
  font-size: 32rpx;
  font-weight: 600;
  padding: 0;
  margin: 0;
}
button::after {
  display: none;
}
</style>
