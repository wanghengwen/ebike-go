<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getLastOrder, OrderPayState, type OrderItem } from '@/api'
import OrderCard from './OrderCard.vue'
import TrackMap from './TrackMap.vue'
import ModifyCostDialog from './ModifyCostDialog.vue'
import { useOrderList, type OrderListScope } from './useOrderList'

const props = defineProps<{ scope: OrderListScope }>()

const lastOrder = ref<OrderItem | null>(null)
const lastError = ref('')
const modifying = ref(false)

const { items, loading, finished, refreshing, error, loadMore, refresh } = useOrderList({
  // 人/车的历史不设近 3 天和 200 条的限制，那两条只约束搜索首页。
  scope: () => props.scope,
})

/**
 * 末单单独一刀。它和列表用的是同一个响应类型，但只有 `detailLast` 会把
 * `hasPaid` / `dispatchCost` / `helmetPenalty` 填上——改价要用这几个值，
 * 所以改价必须挂在末单上，不能拿列表行去改。
 */
async function loadLast(): Promise<void> {
  lastError.value = ''
  try {
    const { success, data, msg } = await getLastOrder(props.scope)
    if (!success) {
      // 车不存在时后端给 10019，这是车侧唯一的存在性校验。
      lastError.value = msg || '未找到订单'
      return
    }
    lastOrder.value = data?.id ? data : null
  } catch {
    lastError.value = '末单加载失败'
  }
}

onMounted(async () => {
  await loadLast()
  await refresh()
})

async function onModified(): Promise<void> {
  await loadLast()
  await refresh()
}
</script>

<template>
  <div class="scope">
    <p v-if="lastError" class="scope__error">{{ lastError }}</p>

    <template v-if="lastOrder">
      <TrackMap
        :points="lastOrder.deviceTrajectory"
        :start="{ lat: lastOrder.startLat, lng: lastOrder.startLng }"
        :end="{ lat: lastOrder.endLat, lng: lastOrder.endLng }"
      />
      <p class="scope__label">最近一单</p>
      <OrderCard :order="lastOrder" />
      <div v-if="lastOrder.izPaid === OrderPayState.ToPay" class="scope__actions">
        <van-button size="small" type="primary" @click="modifying = true">修改订单金额</van-button>
      </div>

      <ModifyCostDialog v-model:show="modifying" :order="lastOrder" @done="onModified" />
    </template>

    <p class="scope__label">历史订单</p>
    <van-pull-refresh v-model="refreshing" @refresh="refresh">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        :error-text="error"
        finished-text="没有更多了"
        @load="loadMore"
      >
        <OrderCard v-for="order in items" :key="order.id ?? ''" :order="order" />
      </van-list>
      <NoData v-if="finished && items.length === 0" />
    </van-pull-refresh>
  </div>
</template>

<style scoped>
.scope__label {
  padding: 14px 14px 0;
  font-size: 12px;
  color: #999999;
}

.scope__error {
  padding: 24px 16px;
  text-align: center;
  font-size: 14px;
  color: #ee0a24;
}

.scope__actions {
  display: flex;
  justify-content: flex-end;
  padding: 10px 12px 0;
}
</style>
