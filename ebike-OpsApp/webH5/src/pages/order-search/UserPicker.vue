<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listUsersByName, type UserPageItem } from '@/api'
import { useServiceAreaStore } from '@/stores/serviceArea'
import { RouteName } from '@/router/names'
import { isoDateTime, orDash, ridingStateLabel, yuan } from './format'

const PAGE_SIZE = 15

const route = useRoute()
const router = useRouter()
const serviceArea = useServiceAreaStore()

const name = String(route.query.name ?? '')
const users = ref<UserPageItem[]>([])
const loading = ref(false)
const finished = ref(false)
const error = ref('')
let pageNum = 1

async function loadMore(): Promise<void> {
  if (finished.value) return
  loading.value = true
  error.value = ''
  try {
    const areaIds = serviceArea.selectedIds.map(Number).filter(Number.isFinite)
    // 遗留实现这里把 pageNum 写死成 1（自增了但没用上），翻页会一直重复第一页。
    const { success, data } = await listUsersByName(name, areaIds, pageNum, PAGE_SIZE)
    const batch = success ? (data?.list ?? []) : []
    users.value = [...users.value, ...batch]
    pageNum += 1
    // 用户分页的 count 是数字（订单那边是字符串），可以当总数用。
    if (batch.length < PAGE_SIZE || users.value.length >= (data?.count ?? 0)) {
      finished.value = true
    }
  } catch {
    error.value = '加载失败，请重试'
    finished.value = true
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!name) {
    finished.value = true
    return
  }
  void loadMore()
})

function open(user: UserPageItem): void {
  void router.replace({ name: RouteName.OrderUser, params: { pin: user.pin } })
}
</script>

<template>
  <div class="picker">
    <p class="picker__hint">「{{ name }}」共 {{ users.length }} 人，请选择</p>

    <van-list
      v-model:loading="loading"
      :finished="finished"
      :error-text="error"
      finished-text="没有更多了"
      @load="loadMore"
    >
      <div v-for="user in users" :key="user.pin" class="user" @click="open(user)">
        <div class="user__head">
          <span class="user__name">{{ orDash(user.authName) }}</span>
          <span class="user__state">{{ ridingStateLabel(user.ridingState) }}</span>
        </div>
        <p class="user__row"><span>手机号</span><b>{{ orDash(user.phone) }}</b></p>
        <p class="user__row"><span>实名</span><b>{{ user.izAuth ? '已实名' : '未实名' }}</b></p>
        <p class="user__row"><span>钱包余额</span><b>{{ yuan(user.balance) }} 元</b></p>
        <p class="user__row"><span>注册时间</span><b>{{ isoDateTime(user.createdAt) }}</b></p>
      </div>
    </van-list>

    <NoData v-if="finished && users.length === 0" />
  </div>
</template>

<style scoped>
.picker {
  min-height: 100vh;
  background: #f5f6f8;
}

.picker__hint {
  padding: 12px 14px 0;
  font-size: 12px;
  color: #999999;
}

.user {
  background: #ffffff;
  border-radius: 8px;
  box-shadow: 0 2px 4px 0 rgba(0, 0, 0, 0.04);
  padding: 12px 14px;
  margin: 10px 12px 0;
}

.user__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 8px;
}

.user__name {
  font-size: 17px;
  font-weight: 600;
  color: #282828;
}

.user__state {
  font-size: 12px;
  color: #999999;
}

.user__row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  line-height: 22px;
}

.user__row span {
  color: #999999;
}

.user__row b {
  color: #282828;
  font-weight: 500;
}
</style>
