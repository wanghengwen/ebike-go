<template>
  <view class="page">
    <view class="card score-card">
      <view class="title">{{ t('pay.creditScore') }}</view>
      <view class="score">{{ score }}</view>
      <view class="level" v-if="levelText">{{ levelText }}</view>
      <view class="forbid" v-if="izLowScoreOn">
        {{ t('pay.creditForbid', { score: noRidingScore }) }}
      </view>
    </view>

    <view class="card">
      <view class="section-title">{{ t('pay.creditRecords') }}</view>
      <view v-if="loading">{{ t('common.loading') }}</view>
      <view v-else-if="!list.length">{{ t('common.empty') }}</view>
      <view v-for="(item, i) in list" :key="i" class="item">
        <view class="item-left">
          <view>{{ item.reason || item.title || '-' }}</view>
          <view class="sub">{{ item.createdAt || item.createTime || '' }}</view>
        </view>
        <view class="delta" :class="{ minus: Number(item.type) === 2 }">
          {{ Number(item.type) === 2 ? '-' : '+' }}{{ item.score ?? '' }}
        </view>
      </view>
      <view class="more" v-if="hasMore" @click="loadMore">{{ t('common.more') }}</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getCreditScoreDetail } from '@/api/wechatScore'
import { getCreditScoreConfig, getCreditScoreList } from '@/api/user'
import { refreshCreditLimit } from '@/features/credit/useCreditLimit'
import { useUserStore } from '@/stores/user'
import { setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const { t } = useI18n()
const user = useUserStore()
const score = ref('--')
const maxScore = ref(0)
const status = ref(0)
const loading = ref(false)
const list = ref<Array<Record<string, unknown>>>([])
const pageNum = ref(1)
const totalCount = ref(0)
const noRidingScore = ref(0)
const izLowScoreOn = ref(false)

const levelText = computed(() => {
  if (status.value === 2) return t('pay.creditBan')
  const max = maxScore.value || 100
  const cur = Number(score.value)
  if (!Number.isFinite(cur)) return ''
  const percent = cur / max
  if (percent < 0.6) return t('pay.creditPoor')
  if (percent < 0.9) return t('pay.creditGood')
  return t('pay.creditExcellent')
})

const hasMore = computed(() => list.value.length < totalCount.value)

function pin() {
  return String(user.userInfo?.pin || storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
}

onShow(() => setNavTitle(t('pay.creditScore')))

async function loadScore() {
  const sid = storage.get<string>('serviceId', '') || ''
  await refreshCreditLimit({ pin: pin(), serviceId: sid, showPopup: false })
  const credit = await getCreditScoreDetail({
    service_id: sid,
    pin: pin(),
  })
  if (credit.success && credit.data) {
    const data = credit.data as { score?: number | string; maxScore?: number; status?: number }
    score.value = String(data.score ?? '--')
    maxScore.value = Number(data.maxScore || 0)
    status.value = Number(data.status || 0)
  }
  const cfg = await getCreditScoreConfig()
  if (cfg.success && cfg.data) {
    const data = cfg.data as { noRiddingScore?: number; izLowScoreOn?: boolean | number }
    noRidingScore.value = Number(data.noRiddingScore || 0)
    izLowScoreOn.value = Boolean(data.izLowScoreOn)
  }
}

async function loadList(reset = false) {
  if (reset) {
    pageNum.value = 1
    list.value = []
  }
  loading.value = true
  try {
    const res = await getCreditScoreList({
      pin: pin(),
      pageSize: 10,
      pageNum: pageNum.value,
      orders: [{ column: 'id', asc: false }],
    })
    if (res.success && res.data) {
      const data = res.data as {
        count?: number
        total?: number
        list?: Array<Record<string, unknown>>
        records?: Array<Record<string, unknown>>
      }
      totalCount.value = Number(data.count ?? data.total ?? 0)
      const rows = data.list || data.records || []
      list.value = reset ? rows : list.value.concat(rows)
    }
  } finally {
    loading.value = false
  }
}

function loadMore() {
  if (!hasMore.value) return
  pageNum.value += 1
  void loadList(false)
}

onMounted(async () => {
  user.hydrateFromStorage()
  await Promise.all([loadScore(), loadList(true)])
})
</script>

<style scoped lang="scss">
.score-card {
  text-align: center;
}
.title {
  color: #888;
}
.score {
  font-size: 72rpx;
  font-weight: 700;
  margin: 16rpx 0;
  color: #3aa0e8;
}
.level {
  color: #666;
  margin-bottom: 8rpx;
}
.forbid {
  color: #e34d59;
  font-size: 24rpx;
}
.section-title {
  font-weight: 700;
  margin-bottom: 16rpx;
}
.item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx 0;
  border-bottom: 1px solid #f0f0f0;
}
.item-left {
  flex: 1;
  padding-right: 16rpx;
}
.sub {
  color: #999;
  font-size: 22rpx;
  margin-top: 8rpx;
}
.delta {
  color: #63d144;
  font-weight: 600;
}
.delta.minus {
  color: #333;
}
.more {
  text-align: center;
  color: #3aa0e8;
  padding: 24rpx 0 8rpx;
}
</style>
