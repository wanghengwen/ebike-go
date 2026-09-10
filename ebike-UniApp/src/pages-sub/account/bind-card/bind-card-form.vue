<template>
  <view class="page">
    <view class="card">
      <view class="title">{{ t('account.idAppealTitle') }}</view>
      <view class="hint">{{ t('account.idAppealHint') }}</view>
      <view class="upload" @click="pick('front')">
        <image v-if="frontCard" class="preview" :src="frontCard" mode="aspectFit" />
        <text v-else>{{ t('account.idFront') }}</text>
      </view>
      <view class="upload" @click="pick('back')">
        <image v-if="backCard" class="preview" :src="backCard" mode="aspectFit" />
        <text v-else>{{ t('account.idBack') }}</text>
      </view>
      <view class="btn-primary" @click="onSubmit">{{ t('common.confirm') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { changeBindAdd } from '@/api/user'
import { uploadFile } from '@/shared/upload'
import { navigate, setNavTitle } from '@/shared/navigate'

const { t } = useI18n()
const authName = ref('')
const authNo = ref('')
const frontCard = ref('')
const backCard = ref('')

onShow(() => setNavTitle(t('account.idAppealTitle')))

onLoad((q) => {
  authName.value = decodeURIComponent(String(q?.authName || ''))
  authNo.value = decodeURIComponent(String(q?.authNo || ''))
})

function pick(side: 'front' | 'back') {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['camera', 'album'],
    success: async (res) => {
      const path = res.tempFilePaths?.[0]
      if (!path) return
      const url = await uploadFile(path)
      if (!url) return
      if (side === 'front') frontCard.value = url
      else backCard.value = url
    },
  })
}

async function onSubmit() {
  if (!frontCard.value || !backCard.value) {
    uni.showToast({ title: t('account.idNeedBoth'), icon: 'none' })
    return
  }
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const res = await changeBindAdd({
      authName: authName.value,
      authNo: authNo.value,
      frontCard: frontCard.value,
      backCard: backCard.value,
      applyType: 2,
    })
    if (res.success) {
      navigate(
        'redirect',
        `/pages-sub/account/bind-card/bind-card-state?authName=${encodeURIComponent(authName.value)}&authNo=${encodeURIComponent(authNo.value)}`,
      )
    } else {
      uni.showToast({ title: res.msg || t('common.networkError'), icon: 'none' })
    }
  } finally {
    uni.hideLoading()
  }
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 12rpx;
}
.hint {
  color: #666;
  margin-bottom: 28rpx;
  line-height: 1.5;
}
.upload {
  height: 280rpx;
  border: 1px dashed #ccc;
  border-radius: 16rpx;
  margin-bottom: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #888;
  overflow: hidden;
}
.preview {
  width: 100%;
  height: 100%;
}
</style>
