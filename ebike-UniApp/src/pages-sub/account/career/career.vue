<template>
  <view class="page">
    <view class="card" v-if="statusVisible">
      <view class="title">{{ t('account.careerAuth') }}</view>
      <view class="status">{{ statusText }}</view>
      <view class="hint" v-if="careerState === 0">{{ t('account.careerReviewTip') }}</view>
      <view class="btn-primary" v-if="careerState === 0" @click="goCustomerService">
        {{ t('support.customerService') }}
      </view>
      <view class="btn-ghost" v-if="careerState === 0" @click="onCancel">{{ t('account.careerCancel') }}</view>
      <view class="btn-primary" v-if="careerState === 2 || careerState === 3" @click="showForm = true">
        {{ t('common.retry') }}
      </view>
    </view>

    <view class="card" v-if="showForm || careerState === 4 || careerState == null">
      <view class="title">{{ t('account.careerAuth') }}</view>
      <view class="hint">{{ t('account.careerHint') }}</view>

      <view class="label">{{ t('account.careerType') }}</view>
      <picker v-if="tags.length" :range="tagNames" @change="onTagPick">
        <view class="picker">{{ selectedTagName || t('account.careerTypePlaceholder') }}</view>
      </picker>
      <input
        v-else
        class="input"
        v-model="company"
        :placeholder="t('account.careerNamePlaceholder')"
      />

      <view class="label">{{ t('account.careerCertNo') }}</view>
      <input class="input" v-model="certNo" :placeholder="t('account.careerCertNoPlaceholder')" />

      <view class="label">{{ t('account.careerFront') }}</view>
      <view class="upload" @click="pickImage('front')">
        <image v-if="frontCard" class="preview" :src="frontCard" mode="aspectFill" />
        <text v-else>{{ t('account.careerUpload') }}</text>
      </view>

      <view class="label">{{ t('account.careerBack') }}</view>
      <view class="upload" @click="pickImage('back')">
        <image v-if="backCard" class="preview" :src="backCard" mode="aspectFill" />
        <text v-else>{{ t('account.careerUploadOptional') }}</text>
      </view>

      <view class="btn-primary" :class="{ disabled: !canSubmit }" @click="onSubmit">{{ t('common.confirm') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { cancelCareerAuth, careerAuth } from '@/api/account'
import { getQualificationList, getPersonInfo } from '@/api/user'
import { uploadFile } from '@/api/common'
import { useUserStore } from '@/stores/user'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const user = useUserStore()

const company = ref('')
const certNo = ref('')
const frontCard = ref('')
const backCard = ref('')
const tags = ref<Array<Record<string, unknown>>>([])
const selectedTag = ref<Record<string, unknown> | null>(null)
const careerState = ref<number | null>(null)
const showForm = ref(false)
const submitting = ref(false)

const tagNames = computed(() => tags.value.map((i) => String(i.tagName || i.name || i.title || '')))
const selectedTagName = computed(() =>
  selectedTag.value
    ? String(selectedTag.value.tagName || selectedTag.value.name || selectedTag.value.title || '')
    : '',
)

const statusVisible = computed(() => careerState.value != null && careerState.value !== 4)
const statusText = computed(() => {
  switch (careerState.value) {
    case 0:
      return t('account.careerPending')
    case 1:
      return t('account.careerPassed')
    case 2:
      return t('account.careerRejected')
    case 3:
      return t('account.careerCancelled')
    default:
      return ''
  }
})

const canSubmit = computed(
  () =>
    Boolean((company.value.trim() || selectedTag.value) && certNo.value.trim() && frontCard.value) &&
    !submitting.value,
)

onShow(() => setNavTitle(t('account.careerAuth')))

onLoad((q) => {
  if (q?.tagName) {
    selectedTag.value = {
      id: q.tagId,
      tagName: decodeURIComponent(String(q.tagName)),
    }
    company.value = decodeURIComponent(String(q.tagName))
  }
})

async function refreshStatus() {
  user.hydrateFromStorage()
  const info = user.userInfo as Record<string, unknown>
  if (info.careerState != null) careerState.value = Number(info.careerState)
  const profile = await getPersonInfo()
  if (profile.success && profile.data) {
    user.setUserInfo(profile.data as never)
    const data = profile.data as { careerState?: number }
    if (data.careerState != null) careerState.value = Number(data.careerState)
  }
  if (careerState.value === 0) showForm.value = false
  if (careerState.value === 4 || careerState.value == null) showForm.value = true
}

async function loadTags() {
  const sid = storage.get<string>('serviceId', '') || ''
  const res = await getQualificationList({ serviceId: sid })
  if (!res.success || !res.data) return
  const data = res.data as {
    careerTags?: Array<Record<string, unknown>>
    career?: string
    izCareer?: boolean
  }
  tags.value = Array.isArray(data.careerTags) ? data.careerTags : []
  if (!company.value && data.career) company.value = String(data.career)
}

onMounted(async () => {
  await Promise.all([refreshStatus(), loadTags()])
})

function onTagPick(e: { detail: { value: string } }) {
  const idx = Number(e.detail.value)
  selectedTag.value = tags.value[idx] || null
  if (selectedTag.value) {
    company.value = String(selectedTag.value.tagName || selectedTag.value.name || '')
  }
}

function pickImage(side: 'front' | 'back') {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: async (res) => {
      const path = res.tempFilePaths?.[0]
      if (!path) return
      uni.showLoading({ title: t('common.loading'), mask: true })
      try {
        const up = await uploadFile(path)
        const url = typeof up === 'string' ? up : (up?.data as string | undefined)
        if (!url) {
          uni.showToast({ title: t('account.careerUploadFail'), icon: 'none' })
          return
        }
        if (side === 'front') frontCard.value = url
        else backCard.value = url
      } finally {
        uni.hideLoading()
      }
    },
  })
}

async function onSubmit() {
  if (!canSubmit.value) return
  if (!frontCard.value) {
    uni.showToast({ title: t('account.careerNeedFront'), icon: 'none' })
    return
  }
  submitting.value = true
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const res = await careerAuth({
      certNo: certNo.value.trim(),
      company: company.value.trim() || selectedTagName.value,
      frontCard: frontCard.value,
      backCard: backCard.value || '',
      tagId: selectedTag.value?.id,
    })
    if (res.success) {
      uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
      careerState.value = 0
      showForm.value = false
      await refreshStatus()
    } else {
      uni.showToast({ title: res.msg || t('common.networkError'), icon: 'none' })
    }
  } finally {
    submitting.value = false
    uni.hideLoading()
  }
}

async function onCancel() {
  uni.showModal({
    title: t('account.careerCancel'),
    content: t('common.confirm'),
    success: async (r) => {
      if (!r.confirm) return
      const res = await cancelCareerAuth()
      if (res.success) {
        uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
        careerState.value = 3
        showForm.value = true
      }
    },
  })
}

function goCustomerService() {
  navigate('to', '/pages-sub/support/customer-service/customer-service')
}
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 12rpx;
}
.hint,
.status {
  color: #888;
  margin-bottom: 24rpx;
}
.label {
  font-size: 24rpx;
  color: #666;
  margin-bottom: 8rpx;
}
.input,
.picker {
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 24rpx;
}
.picker {
  color: #333;
}
.upload {
  height: 220rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
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
.disabled {
  opacity: 0.5;
}
.btn-ghost {
  margin-top: 16rpx;
}
</style>
