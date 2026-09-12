/**
 * 订单查询域的类型与调用。
 *
 * 单独一个模块而不并进 `models.ts`：那边是两块大屏的口径，这边的字段约定完全不同——
 * 后端把 Java `Long` 全序列化成字符串，所以 `id` / `ridingTime` 是 string；
 * 订单分页的 `count` 是 string，用户分页的 `count` 却是 number。放一起太容易拿错。
 *
 * 单位（有后端代码为据，不要凭字段名猜）：金额一律**分**，`ridingTime` **毫秒**，
 * `mile` **米**；响应里的 `startTime` / `endTime` 是 `"yyyy-MM-dd HH:mm:ss"` 字符串，
 * 而请求里的时间范围是**毫秒时间戳数组**。
 */
import { post } from './http'
import { Endpoints } from './endpoints'

/** 订单分页外壳。`count` 是字符串，要 `parseInt`。 */
export interface OrderPage<T> {
  list: T[] | null
  count: string
  pageNum: number
  pageSize: number
}

/** 用户分页外壳。这里的 `count` 是数字——和订单那边不一样。 */
export interface UserPage<T> {
  list: T[] | null
  count: number
  pageNum: number
  pageSize: number
}

export interface TrackPoint {
  lng: number | null
  lat: number | null
  speed: number | null
  course: number | null
  /** 毫秒时间戳，但被序列化成了字符串。 */
  timestamp: string | null
}

/** `izPaid` 取值。后端 `OrderIzPayEnum`。 */
export const OrderPayState = {
  /** 正常骑行中，未冻结 */
  Riding: 1,
  /** 准备支付中，冻结中 */
  Frozen: 2,
  /** 行程结束，待支付 */
  ToPay: 3,
  /** 完成订单，已支付 */
  Paid: 4,
} as const

/**
 * `bList` 与 `detailLast` 的元素类型（后端 `bLastOrderDetailCO`）。
 *
 * 两个接口共用这个类型，但填充程度不同：`bList` 不填 `userName` / `carState` /
 * `hasPaid`，也不填 item 上的 `penalty` / `dispatchCost` / `deduction` / `helmetPenalty`
 * ——那几个只有 `detailLast` 会补。key 都在，值是 null。改价要用到它们，所以
 * 改价必须基于 `detailLast` 的结果，不能拿列表行凑。
 */
export interface OrderItem {
  /** Long 字符串。 */
  id: string | null
  /** 分。 */
  originCost: number | null
  /** 分。 */
  payCost: number | null
  /** 米。 */
  mile: number | null
  /** 毫秒，字符串。 */
  ridingTime: string | null
  izPaid: number | null
  carId: string | null
  startLat: number | null
  startLng: number | null
  endLat: number | null
  endLng: number | null
  payTime: string | null
  startTime: string | null
  endTime: string | null
  phone: string | null
  userPin: string | null
  deviceTrajectory: TrackPoint[] | null
  payType: number | null
  /** 分。 */
  totalRefundCost: number | null

  // --- 以下仅 detailLast 会填 ---
  userName: string | null
  carState: number | null
  /** 已实付，分。改价后的合计不能低于它。 */
  hasPaid: number | null
  /** 还车围栏判定，`1101` 是站点内定位匹配。后端 `FenceRelation`。 */
  penaltyType: number | null
  /** 分。 */
  penalty: number | null
  /** 分。 */
  dispatchCost: number | null
  /** 分。 */
  deduction: number | null
  /** 分。 */
  helmetPenalty: number | null
}

export interface OrderListQuery {
  pageNum: number
  pageSize: number
  userPin?: string
  imei?: string
  carId?: string
  izPaid?: number
  /** 必须是数字，后端是 `*int64`。 */
  serviceId?: number
  /** `[起, 止]` 毫秒时间戳。 */
  startTime?: [number, number]
}

/** 用户骑行中的概要，`ridingState` 为 1 或 2 时才有内容，否则是 `{}`。 */
export interface RidingInfo {
  cardNo?: string
  /** 米。 */
  distance?: number
  startTime?: string
  /** 分。 */
  fee?: number
  imei?: string
  /** 秒——注意和订单的 `ridingTime`（毫秒）不是一个单位。 */
  duration?: number
}

/** 下游 `UserDetailCo` + BFF 追加字段。只列页面用得到的。 */
export interface UserDetail {
  pin: string
  authName: string
  authNo: string
  phone: string
  /** 带 `T` 的 ISO 串，和订单时间的空格分隔格式不一样。 */
  createdAt: string | null
  izAuth: boolean | null
  izDeposited: boolean | null
  izDepositCard: boolean | null
  /** 1 已注册 / 2 已实名 / 3 可用车 / 4 预约 / 5 临停 / 6 骑行中 / 7 待支付 */
  ridingState: number | null
  serviceId: number | null

  // --- BFF 追加 ---
  /** 诚信金，分。 */
  depositAmount: number | null
  depositedCardDays: number | null
  creditScore: number | null
  /** 3 黑名单未过期 / 2 已认证 / 1 其它。 */
  userState: number | null
  serviceName: string | null
  ridingInfo: RidingInfo | null
}

export interface UserPageItem {
  pin: string
  authName: string
  phone: string
  izAuth: boolean | null
  ridingState: number | null
  createdAt: string | null
  /** 钱包余额，分。BFF 逐行补的。 */
  balance: number | null
  lastOrderNo: number | null
}

export interface UpdateCostPayload {
  orderId: string
  /** 分。 */
  modifyPayCost: number
  /** 分。 */
  modifyDispatchCost: number
  /** 分。可选。 */
  modifyHelmetPenalty?: number
}

// --- 调用 -------------------------------------------------------------------

export const listOrders = (query: OrderListQuery) =>
  post<OrderPage<OrderItem>>(Endpoints.orderList, query)

/**
 * 末单。传 `userPin` 查人的、传 `carId` / `imei` 查车的。
 * 车不存在时返回业务码 `10019`，这也是车侧唯一的存在性校验——
 * 我们不调 `device/detail`，车辆信息卡由原生页面负责。
 */
export const getLastOrder = (query: { userPin?: string; carId?: string; imei?: string }) =>
  post<OrderItem>(Endpoints.orderDetailLast, query)

/** 只为拿 `deviceTrajectory` 画轨迹；列表行自带轨迹时不用调。 */
export const getOrderDetail = (orderId: string) =>
  post<OrderItem>(Endpoints.orderDetail, { orderId })

export const getUserByPin = (pin: string) => post<UserDetail>(Endpoints.userDetail, { pin })

/**
 * 按手机号查用户。
 *
 * 后端这个接口按 `phone = ?` 精确匹配且**不会**自动补 `+86-`（订单列表和用户分页会补），
 * 所以前缀得自己拼。遗留 App 无论号码归属地一律拼 `+86-`，这里照做以免查不到已有数据。
 */
export const getUserByPhone = (phone: string, serviceIds: string[]) =>
  post<UserDetail>(Endpoints.userDetailByPhone, {
    type: '1',
    phone: `+86-${phone}`,
    serviceIds,
  })

/** 姓名是**精确等于**，不是模糊匹配。`serviceId` 至少要一个，后端强制。 */
export const listUsersByName = (authName: string, serviceId: number[], pageNum: number, pageSize: number) =>
  post<UserPage<UserPageItem>>(Endpoints.userPage, { authName, serviceId, pageNum, pageSize })

/** 返回的 `data` 是新建工单的 id（数字）。已支付完成的订单不能改价，后端会拒。 */
export const createUpdateCostTicket = (payload: UpdateCostPayload) =>
  post<number>(Endpoints.updateCostTicket, payload)
