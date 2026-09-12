import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import { useConfigStore } from '@/stores/config'
import { PermissionCode, usePermissionStore } from '@/stores/permission'
import { applyTheme } from '@/utils/theme'
import { setLocale, type Locale, SUPPORTED_LOCALES } from '@/i18n'
import { RouteName } from './names'

export { RouteName } from './names'

declare module 'vue-router' {
  interface RouteMeta {
    /** 权限码表达式，语义同遗留 `v-hasCode`：`,` 全部满足，`||` 满足其一。 */
    permission?: string
    keepAlive?: boolean
  }
}

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: { name: RouteName.OperationScreen } },
  {
    path: '/newOperationScreen',
    name: RouteName.OperationScreen,
    component: () => import('@/pages/operation-screen/OperationScreen.vue'),
    meta: { keepAlive: true },
  },
  {
    path: '/newRevenueHome',
    // 租户配置和老 iOS 里写的是全小写的 newrevenueHome，一直靠 vue-router 默认
    // 大小写不敏感匹配兜着。显式列出来，免得哪天有人打开 sensitive 就 404。
    alias: '/newrevenueHome',
    name: RouteName.RevenueHome,
    component: () => import('@/pages/revenue-home/RevenueHome.vue'),
    meta: { keepAlive: true },
  },
  {
    path: '/trend',
    name: RouteName.TrendDetail,
    component: () => import('@/pages/trend/TrendDetail.vue'),
  },

  // --- 订单查询 ---
  // 入口权限按 `1212 || 0204` 放行：同一个功能 App 侧发 1212、PC 后台发 0204，
  // 两边发码习惯不统一。与 `H5ScreenKind.Order.permissionCodes` 保持一致。
  {
    path: '/order/search',
    name: RouteName.OrderSearch,
    component: () => import('@/pages/order-search/OrderSearch.vue'),
    meta: { permission: `${PermissionCode.OrderQuery}||${PermissionCode.PcOrderMenu}` },
  },
  {
    path: '/order/users',
    name: RouteName.OrderUserPicker,
    component: () => import('@/pages/order-search/UserPicker.vue'),
    meta: { permission: PermissionCode.OrderQueryPersonal },
  },
  {
    path: '/order/user/:pin',
    name: RouteName.OrderUser,
    component: () => import('@/pages/order-search/UserOrders.vue'),
    meta: { permission: PermissionCode.OrderQueryPersonal },
  },
  {
    path: '/order/vehicle',
    name: RouteName.OrderVehicle,
    component: () => import('@/pages/order-search/VehicleOrders.vue'),
    meta: { permission: PermissionCode.OrderQueryVehicle },
  },
  // 遗留 App 里散落着 `/operationScreen`、`/opHome` 等旧路径，统一兜到运营大屏。
  { path: '/:pathMatch(.*)*', redirect: { name: RouteName.OperationScreen } },
]

export const router = createRouter({
  history: createWebHashHistory(import.meta.env.VITE_BASE_PATH || '/mop-saas/'),
  routes,
  scrollBehavior(_to, _from, savedPosition) {
    // vue-router 4 起坐标字段由 { x, y } 改为 { left, top }
    return savedPosition ?? { left: 0, top: 0 }
  },
})

function firstQueryValue(value: unknown): string | undefined {
  if (Array.isArray(value)) return typeof value[0] === 'string' ? value[0] : undefined
  return typeof value === 'string' ? value : undefined
}

router.beforeEach(async (to) => {
  const config = useConfigStore()
  config.hydrateFromQuery(to.query)

  applyTheme(firstQueryValue(to.query.themeColor), firstQueryValue(to.query.lightTxtColor))

  const lang = firstQueryValue(to.query.lang)
  if (lang && SUPPORTED_LOCALES.includes(lang as Locale)) setLocale(lang as Locale)

  const permission = usePermissionStore()
  try {
    await permission.load()
  } catch {
    // 大屏是逐卡片判权的，拉不到码顶多少显示几张卡，不要把用户卡在白屏。
  }

  // 管理页不一样：整页都受权限约束，拉不到码就没有放行的依据。
  const required = to.meta.permission
  if (required && (!permission.loaded || !permission.hasExpression(required))) {
    return { name: RouteName.OperationScreen }
  }
  return true
})

export default router
