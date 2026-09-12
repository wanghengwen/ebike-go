import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { getServiceArea, type ServiceArea } from '@/api'
import { storage } from '@/utils/storage'

const SELECTION_KEY = 'serviceId'

export const useServiceAreaStore = defineStore('serviceArea', () => {
  const list = ref<ServiceArea[]>([])
  const selectedIds = ref<string[]>(storage.get<string[]>(SELECTION_KEY) ?? [])

  const allIds = computed(() => list.value.map((item) => item.id))
  const selectedCount = computed(() => selectedIds.value.length)
  const isEmpty = computed(() => list.value.length === 0)

  function select(ids: string[]): void {
    selectedIds.value = ids
    storage.set(SELECTION_KEY, ids)
  }

  /**
   * 拉取服务区并校正本地选择：与服务端列表求交集，交集为空则回落到全选。
   * 服务区被后台删除后本地缓存会残留失效 id，这一步是为了避免带着脏 id 去查数据。
   */
  async function load(): Promise<void> {
    const { success, data } = await getServiceArea()
    if (!success) throw new Error('failed to load service areas')

    list.value = data ?? []
    const serverIds = allIds.value
    const cached = selectedIds.value
    const intersection = cached.filter((id) => serverIds.includes(id))
    select(intersection.length > 0 ? intersection : serverIds)
  }

  return { list, selectedIds, allIds, selectedCount, isEmpty, select, load }
})
