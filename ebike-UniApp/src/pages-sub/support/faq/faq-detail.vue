<template>
  <view class="page">
    <view class="card">
      <view v-if="loading">{{ t('common.loading') }}</view>
      <template v-else>
        <view class="title">{{ titleText }}</view>
        <rich-text v-if="html" class="body" :nodes="html" />
        <view v-else class="body plain">{{ plainText || t('common.empty') }}</view>
      </template>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getFaqById } from '@/api/service'
import { formatRichText } from '@/shared/format'
import { setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const loading = ref(false)
const detail = ref<Record<string, unknown>>({})

const titleText = computed(() =>
  String(
    detail.value.detailTitle ||
      detail.value.title ||
      detail.value.question ||
      t('support.faq'),
  ),
)

const html = computed(() => {
  const raw = detail.value.detail || detail.value.content || detail.value.answer || detail.value.desc
  if (!raw) return ''
  const s = String(raw)
  if (!/[<>]/.test(s)) return ''
  return formatRichText(s)
})

const plainText = computed(() => {
  if (html.value) return ''
  return String(detail.value.content || detail.value.answer || detail.value.desc || '')
})

onShow(() => setNavTitle(t('support.faq')))

onLoad(async (q) => {
  const id = q?.id
  if (!id) return
  loading.value = true
  try {
    const serviceId = storage.get<string>('serviceId', '') || ''
    const res = await getFaqById({
      id,
      ...(serviceId ? { serviceId } : {}),
    })
    if (res.success && res.data) {
      detail.value = res.data as Record<string, unknown>
      const nav = String(
        (res.data as { detailTitle?: string }).detailTitle ||
          (res.data as { title?: string }).title ||
          '',
      )
      if (nav) setNavTitle(nav)
    } else {
      uni.showToast({ title: t('support.faqLoadFail'), icon: 'none' })
    }
  } catch {
    uni.showToast({ title: t('support.faqLoadFail'), icon: 'none' })
  } finally {
    loading.value = false
  }
})
</script>

<style>
.quesDetail_font_size_1 {
  font-size: 12px;
}
.quesDetail_font_size_2 {
  font-size: 14px;
}
.quesDetail_font_size_3 {
  font-size: 16px;
}
.quesDetail_font_size_4 {
  font-size: 18px;
}
.quesDetail_font_size_5 {
  font-size: 24px;
}
.quesDetail_font_size_6 {
  font-size: 32px;
}
.quesDetail_font_size_7 {
  font-size: 48px;
}
</style>

<style scoped lang="scss">
.title {
  font-weight: 700;
  margin-bottom: 20rpx;
  font-size: 32rpx;
}
.body {
  color: #555;
  line-height: 1.6;
  font-size: 28rpx;
}
.plain {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
