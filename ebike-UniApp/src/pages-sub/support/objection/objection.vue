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
            <image :src="item" class="img-item" mode="aspectFill" @tap="preview(index)" />
            <view class="delete" @tap.stop="delImage(index)">
              <image v-if="deleteIcon" class="delete-icon" :src="deleteIcon" mode="widthFix" />
              <text v-else class="delete-fallback">×</text>
            </view>
          </view>
          <view
            v-if="imgs.length < 3"
            class="add"
            :class="{ 'add--fallback': !addIcon }"
            @tap.stop="shooting"
            @click.stop="shooting"
          >
            <image v-if="addIcon" class="add-icon" :src="addIcon" mode="aspectFit" />
            <text v-else class="add-plus">+</text>
          </view>
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
import { uploadFile, pickAndUploadPhoto } from '@/shared/upload'
import { logger } from '@/shared/logger'
import { isNative, nativeHost } from '@/shared/nativeHost'

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
const picking = ref(false)

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
  if (picking.value) return
  if (imgs.value.length >= 3) {
    uni.showToast({ title: t('support.objectionPhotoMax'), icon: 'none' })
    return
  }

  // 有原生桥就走桥：不要依赖 isNative() 的 class 检测，也不要先弹 sheet（WebView 里常看不见）。
  const host = nativeHost()
  if (host || isNative()) {
    const ways = chooseImgWay.value.length
      ? chooseImgWay.value
      : (['camera'] as Array<'camera' | 'album'>)
    const source: 'camera' | 'album' = ways.includes('camera') ? 'camera' : 'album'
    void runNativePick(source)
    return
  }

  const maxSize = Number((getTenantConfig() as { fileSize?: number }).fileSize) || 10485760
  // WebView 里 chooseImage 可能既不回 success 也不回 fail，看起来就是「点了没反应」。
  let settled = false
  const watchdog = setTimeout(() => {
    if (settled) return
    settled = true
    uni.showModal({
      title: '无法选择照片',
      content: `当前环境不支持选图，且未检测到原生拍照能力。\nplatform=${nativeEnvHint()}`,
      showCancel: false,
    })
  }, 1500)

  uni.chooseImage({
    count: 3 - imgs.value.length,
    sizeType: ['compressed'],
    sourceType: chooseImgWay.value,
    success: async (res) => {
      settled = true
      clearTimeout(watchdog)
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
    fail: (err) => {
      settled = true
      clearTimeout(watchdog)
      logger.warn('chooseImage fail', err)
      uni.showModal({
        title: '无法选择照片',
        content: String((err as { errMsg?: string })?.errMsg || t('support.objectionUploadFail')),
        showCancel: false,
      })
    },
  })
}

/** 排查用：告诉我们原生桥到底有没有注入。 */
function nativeEnvHint(): string {
  const w = window as unknown as {
    __riderNative?: unknown
    __riderNativePlatform?: string
    webkit?: { messageHandlers?: { riderNative?: unknown } }
  }
  const flags = [
    `bridge=${w.__riderNative ? 'y' : 'n'}`,
    `wk=${w.webkit?.messageHandlers?.riderNative ? 'y' : 'n'}`,
    `plat=${w.__riderNativePlatform || '-'}`,
  ]
  return flags.join(' ')
}

async function runNativePick(source: 'camera' | 'album') {
  if (picking.value) return
  picking.value = true
  try {
    const url = await pickAndUploadPhoto(source)
    if (url) imgs.value.push(url)
  } catch (e) {
    logger.error('runNativePick fail', e)
    uni.showModal({
      title: '提示',
      content: t('support.objectionUploadFail'),
      showCancel: false,
    })
  } finally {
    picking.value = false
  }
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
  display: flex;
  align-items: center;
  justify-content: center;
  height: 120rpx;
  width: 120rpx;
  background: #f6f6f6;
  border-radius: 29rpx;
  box-sizing: border-box;
}
.add-icon {
  height: 120rpx;
  width: 120rpx;
  border-radius: 29rpx;
  /* Android WebView 上子 image 常抢走点击，导致外层 @tap 完全无反应 */
  pointer-events: none;
}
.add--fallback {
  border: 2rpx solid #e5e5e5;
}
.add-plus {
  font-size: 64rpx;
  color: #696969;
  line-height: 1;
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
