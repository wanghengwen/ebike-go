<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getUserByPhone, listUsersByName } from '@/api'
import { usePermissionStore, PermissionCode } from '@/stores/permission'
import { useServiceAreaStore } from '@/stores/serviceArea'
import { RouteName } from '@/router/names'
import { classify, requiredCode, scopeHint } from './searchKind'
import { useOrderList } from './useOrderList'
import OrderCard from './OrderCard.vue'

/** 搜索首页只看近 3 天、最多 200 条，与遗留 App 的 `limitDays` / `limitMaxCount` 一致。 */
const RECENT_DAYS = 3
const MAX_ROWS = 200

const router = useRouter()
const permission = usePermissionStore()
const serviceArea = useServiceAreaStore()

const keyword = ref('')
const searching = ref(false)

const canQueryPerson = computed(() => permission.has(PermissionCode.OrderQueryPersonal))
const canQueryVehicle = computed(() => permission.has(PermissionCode.OrderQueryVehicle))
const placeholder = computed(() => scopeHint(canQueryPerson.value, canQueryVehicle.value))

/**
 * `bList` 的 `serviceId` 只收单个，装不下这里的多选。
 * 首页默认列表取第一个已选服务区，并把区名显示出来——静默只查一个区太容易被误读。
 */
const areaId = computed(() => serviceArea.selectedIds[0])
const areaName = computed(
  () => serviceArea.list.find((item) => item.id === areaId.value)?.name ?? '',
)

const { items, loading, finished, refreshing, error, loadMore, refresh } = useOrderList({
  scope: () => ({ serviceId: areaId.value ? Number(areaId.value) : undefined }),
  limitDays: RECENT_DAYS,
  limitMaxCount: MAX_ROWS,
})

onMounted(async () => {
  if (serviceArea.isEmpty) {
    try {
      await serviceArea.load()
    } catch {
      showToast('服务区加载失败')
    }
  }
  await refresh()
})

async function search(): Promise<void> {
  const value = keyword.value.trim()
  if (!value) {
    showToast('请输入查询内容')
    return
  }

  const kind = classify(value)
  if (!permission.has(requiredCode(kind))) {
    showToast('无权限')
    return
  }

  searching.value = true
  try {
    if (kind === 'carId') {
      await router.push({ name: RouteName.OrderVehicle, query: { carId: value } })
      return
    }
    if (kind === 'imei') {
      await router.push({ name: RouteName.OrderVehicle, query: { imei: value } })
      return
    }
    if (kind === 'phone') {
      const { success, data } = await getUserByPhone(value, serviceArea.selectedIds)
      if (!success || !data?.pin) {
        showToast('用户不存在')
        return
      }
      await router.push({ name: RouteName.OrderUser, params: { pin: data.pin } })
      return
    }

    // 姓名是精确匹配，且可能命中同名。先探一页确认存在，再交给选人页。
    const areaIds = serviceArea.selectedIds.map(Number).filter(Number.isFinite)
    const { success, data } = await listUsersByName(value, areaIds, 1, 1)
    if (!success || !data?.list?.length) {
      showToast('用户不存在')
      return
    }
    await router.push({ name: RouteName.OrderUserPicker, query: { name: value } })
  } catch {
    showToast('查询失败，请重试')
  } finally {
    searching.value = false
  }
}

/** 清空输入回到默认列表，与遗留一致；查询是跳走，不在这张列表里过滤。 */
function onClear(): void {
  keyword.value = ''
}

/** 点列表行看这台车的订单。遗留 App 进的是车辆详情，车辆详情这边由原生页面承担。 */
function openVehicle(carId: string | null): void {
  if (!carId) return
  void router.push({ name: RouteName.OrderVehicle, query: { carId } })
}
</script>

<template>
  <div class="search">
    <van-search
      v-model="keyword"
      shape="round"
      :placeholder="placeholder"
      :disabled="!placeholder"
      @search="search"
      @clear="onClear"
    >
      <template #action>
        <van-button size="small" type="primary" :loading="searching" @click="search">
          查询
        </van-button>
      </template>
    </van-search>

    <p v-if="!placeholder" class="search__denied">当前账号没有订单查询权限</p>

    <template v-else>
      <p class="search__hint">
        近 {{ RECENT_DAYS }} 天订单<span v-if="areaName"> · {{ areaName }}</span>
      </p>

      <van-pull-refresh v-model="refreshing" @refresh="refresh">
        <van-list
          v-model:loading="loading"
          :finished="finished"
          :error-text="error"
          finished-text="没有更多了"
          @load="loadMore"
        >
          <OrderCard
            v-for="order in items"
            :key="order.id ?? ''"
            :order="order"
            clickable
            @click="openVehicle(order.carId)"
          />
        </van-list>
        <NoData v-if="finished && items.length === 0" />
      </van-pull-refresh>
    </template>
  </div>
</template>

<style scoped>
.search {
  min-height: 100vh;
  background: #f5f6f8;
}

.search__hint {
  padding: 8px 14px 0;
  font-size: 12px;
  color: #999999;
}

.search__denied {
  padding: 40px 16px;
  text-align: center;
  font-size: 14px;
  color: #999999;
}
</style>
