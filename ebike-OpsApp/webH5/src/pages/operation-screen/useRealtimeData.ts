import { ref } from 'vue'
import { getCarStates, getUserNum, getUserSunburst, type MemberStatistic } from '@/api'
import { bucketDevices, excludeOffShelves } from '@/utils/carType'
import { filterByBattery, filterByIdleHours } from '@/utils/deviceFilters'
import type { UserTreeNode } from '@/components/charts/UserTreeChart.vue'
import type { PieSlice } from '@/components/charts/options'

export interface BarBlock {
  title: string
  unit: string
  total: number
  categories: string[]
  values: number[]
  gradient: [string, string]
}

const IDLE_BUCKETS: Array<[label: string, from: number, to: number]> = [
  ['48h以上', 48, Number.MAX_VALUE],
  ['24-48h', 24, 48],
  ['12-24h', 12, 24],
  ['6-12h', 6, 12],
  ['3-6h', 3, 6],
  ['1-3h', 1, 3],
]

const BATTERY_BUCKETS: Array<[label: string, from: number, to: number]> = [
  ['0%', 0, 0],
  ['0-20%', 0, 20],
  ['20-35%', 20, 35],
  ['35%以上', 35, 101],
]

function buildUserTree(info: MemberStatistic['info']): UserTreeNode[] {
  return [
    {
      name: '总用户数',
      value: info.authentication + info.noAuthentication,
      itemStyle: { color: '#6DD400' },
      children: [
        {
          name: '实名用户',
          value: info.authentication,
          itemStyle: { color: '#E43D2E' },
          children: [
            {
              name: '非会员用户',
              value: info.nonMember,
              itemStyle: { color: '#32C5FF' },
              children: [
                {
                  name: '历史非会员',
                  value: info.nonMember - info.historicalMember,
                  itemStyle: { color: '#C20879' },
                },
                { name: '历史会员', value: info.historicalMember, itemStyle: { color: '#6CF1CB' } },
              ],
            },
            {
              name: '会员用户',
              value: info.member,
              itemStyle: { color: '#08C28D' },
              children: [
                {
                  name: '有效会员',
                  value: info.validMember,
                  itemStyle: { color: '#FF4545' },
                  children: [
                    { name: '一键免押会员', value: info.freeUser, itemStyle: { color: '#4304DE' } },
                    { name: '诚信金额', value: info.deposit, itemStyle: { color: '#B698FF' } },
                    { name: '学生认证', value: info.career, itemStyle: { color: '#009531' } },
                    { name: '会员卡', value: info.memberCard, itemStyle: { color: '#BF5D03' } },
                  ],
                },
                {
                  name: '待失效会员',
                  value: info.invalidMember,
                  itemStyle: { color: '#BE37FF' },
                },
              ],
            },
          ],
        },
        {
          name: '非实名用户',
          value: info.noAuthentication,
          itemStyle: { color: '#ECAF06' },
        },
      ],
    },
  ]
}

type BarSpec = Pick<BarBlock, 'title' | 'categories' | 'gradient'>

const VEHICLE_SPEC: BarSpec = {
  title: '运营车辆',
  categories: ['运维中', '临停中', '骑行中', '预约中', '可使用'],
  gradient: ['#60AA00', '#A9F138'],
}
const IDLE_SPEC: BarSpec = {
  title: '闲置统计',
  categories: IDLE_BUCKETS.map(([label]) => label),
  gradient: ['#44D7B6', '#37F2C8'],
}
const BATTERY_SPEC: BarSpec = {
  title: '电量统计',
  categories: BATTERY_BUCKETS.map(([label]) => label),
  gradient: ['#0A97FF', '#48A3FF'],
}

/**
 * 数据到达前三块条形图就要占位渲染（标题显示 `--`、柱子为空），
 * 数据到达后才挂载会让首屏出现明显的高度跳动。
 */
function block(spec: BarSpec, total: number, values: number[]): BarBlock {
  return { ...spec, unit: '辆', total, values }
}

/** 运营大屏「实时数据」页签的三块车辆统计 + 用户树 + 实名认证环形图。 */
export function useRealtimeData() {
  const zeros = (spec: BarSpec) => block(spec, 0, spec.categories.map(() => 0))

  const vehicleStats = ref<BarBlock>(zeros(VEHICLE_SPEC))
  const idleStats = ref<BarBlock>(zeros(IDLE_SPEC))
  const batteryStats = ref<BarBlock>(zeros(BATTERY_SPEC))

  const userTree = ref<UserTreeNode[]>([])
  const userInfo = ref<MemberStatistic['info'] | null>(null)

  const ridingQualification = ref<{ slices: PieSlice[]; total: number }>({ slices: [], total: 0 })

  async function loadVehicles(serviceIds: string[]): Promise<void> {
    const { success, data } = await getCarStates(serviceIds)
    if (!success || !data) return

    const buckets = bucketDevices(data)
    const active = excludeOffShelves(buckets.all ?? [])

    vehicleStats.value = block(VEHICLE_SPEC, active.length, [
      excludeOffShelves(buckets.operating ?? []).length,
      (buckets.normalTempParking ?? []).length,
      (buckets.normalRiding ?? []).length,
      (buckets.normalSubscribing ?? []).length,
      (buckets.normalAvailable ?? []).length,
    ])

    const idleValues = IDLE_BUCKETS.map(([, from, to]) => filterByIdleHours(active, from, to).length)
    idleStats.value = block(
      IDLE_SPEC,
      idleValues.reduce((sum, n) => sum + n, 0),
      idleValues,
    )

    batteryStats.value = block(
      BATTERY_SPEC,
      active.length,
      BATTERY_BUCKETS.map(([, from, to]) => filterByBattery(active, from, to).length),
    )
  }

  async function loadUserTree(serviceIds: string[]): Promise<void> {
    const { success, data } = await getUserSunburst(serviceIds)
    if (!success || !data?.info) return
    userInfo.value = data.info
    userTree.value = buildUserTree(data.info)
  }

  async function loadRidingQualification(serviceIds: string[]): Promise<void> {
    const { success, data } = await getUserNum(serviceIds)
    if (!success || !data) return
    const have = data.user_treemap.have_riding_qualification.total
    const none = data.user_treemap.no_riding_qualification.total
    ridingQualification.value = {
      total: have + none,
      slices: [
        { name: '有骑行资格用户', value: have },
        { name: '无骑行资格用户', value: none },
      ],
    }
  }

  async function loadAll(serviceIds: string[]): Promise<void> {
    await Promise.all([
      loadVehicles(serviceIds),
      loadUserTree(serviceIds),
      loadRidingQualification(serviceIds),
    ])
  }

  return {
    vehicleStats,
    idleStats,
    batteryStats,
    userTree,
    userInfo,
    ridingQualification,
    loadAll,
  }
}
