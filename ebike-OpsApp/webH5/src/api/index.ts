import { post } from './http'
import { Endpoints } from './endpoints'
import type { ApiResponse, ServiceArea, UserInfo } from './types'
import type {
  AgeStatistic,
  AmountSummary,
  AverageDataResult,
  CardSellTimesResult,
  CarsDataQuery,
  CarsDataResult,
  DailySeriesResult,
  DeviceItem,
  MemberStatistic,
  OperationDataInfo,
  OrderInfoResult,
  OrderMountResult,
  ScreenRevenueConfig,
  SnakeRangeQuery,
  TaskInfoResult,
  TicketOrderResult,
  UserGrowQuery,
  UserGrowResult,
} from './models'

export * from './models'
export * from './order'
export type { ApiResponse, ServiceArea, UserInfo } from './types'

// --- 通用 -----------------------------------------------------------------

export const getUserInfo = () => post<UserInfo>(Endpoints.userInfo)
export const getServiceArea = () => post<ServiceArea[]>(Endpoints.serviceArea)

// --- 运营大屏 -------------------------------------------------------------

export const getCarStates = (serviceIdList: string[]) =>
  post<DeviceItem[]>(Endpoints.carStates, { serviceIdList })

export const getUserSunburst = (serviceIds: string[]) =>
  post<MemberStatistic>(Endpoints.userSunburst, { serviceIds })

export const getUserNum = (serviceIds: string[]) =>
  post<AgeStatistic>(Endpoints.userNum, { serviceIds })

export const getOperationData = (query: SnakeRangeQuery) =>
  post<OperationDataInfo>(Endpoints.operationData, query)

export const getOrderInfoTime = (query: SnakeRangeQuery) =>
  post<OrderInfoResult>(Endpoints.orderInfoTime, query)

export const getOrderInfoArea = (query: SnakeRangeQuery) =>
  post<OrderInfoResult>(Endpoints.orderInfoArea, query)

export const getTaskInfo = (query: SnakeRangeQuery) =>
  post<TaskInfoResult>(Endpoints.taskInfo, query)

export const getOrderMount = (query: SnakeRangeQuery) =>
  post<OrderMountResult>(Endpoints.orderMount, query)

export const getAverageData = (query: SnakeRangeQuery) =>
  post<AverageDataResult>(Endpoints.averageData, query)

export const getTicketOrder = (query: SnakeRangeQuery) =>
  post<TicketOrderResult>(Endpoints.ticketOrder, query)

export const getUserGrow = (query: UserGrowQuery) =>
  post<UserGrowResult>(Endpoints.userGrow, query)

export const getCarsData = (query: CarsDataQuery) =>
  post<CarsDataResult>(Endpoints.carsData, query)

// --- 营收大屏 -------------------------------------------------------------
// 这批接口出参形状高度一致（金额 + 笔数汇总，或按天折线），共用两个泛型签名。

type SummaryCall = (query: SnakeRangeQuery) => Promise<ApiResponse<AmountSummary>>
type SeriesCall = (query: SnakeRangeQuery) => Promise<ApiResponse<DailySeriesResult>>
type CardTimesCall = (query: SnakeRangeQuery) => Promise<ApiResponse<CardSellTimesResult>>

const summary =
  (path: string): SummaryCall =>
  (query) =>
    post<AmountSummary>(path, query)

const series =
  (path: string): SeriesCall =>
  (query) =>
    post<DailySeriesResult>(path, query)

const cardTimes =
  (path: string): CardTimesCall =>
  (query) =>
    post<CardSellTimesResult>(path, query)

export const getMerchantWallet = summary(Endpoints.merchantWallet)
export const getDepositCard = summary(Endpoints.depositCard)
export const getDepositCardTimes = summary(Endpoints.depositCardTimes)
export const getRidingCard = summary(Endpoints.ridingCard)
export const getRidingCardTimes = summary(Endpoints.ridingCardTimes)
export const getFavorableCards = summary(Endpoints.favorableCards)
export const getFavorableCardTimes = summary(Endpoints.favorableCardTimes)
export const getFavorableCardSellTimes = cardTimes(Endpoints.favorableCardSellTimes)
export const getDeposit = summary(Endpoints.deposit)

export const getWeiXin = summary(Endpoints.weixin)
export const getALi = summary(Endpoints.ali)
export const getUnion = summary(Endpoints.union)

export const getPayment = summary(Endpoints.payment)
export const getOrderPayTime = summary(Endpoints.orderPayTime)
export const getRefund = summary(Endpoints.refund)
export const getReportManage = summary(Endpoints.reportManage)
export const getTotal = summary(Endpoints.total)
export const getPenalty = summary(Endpoints.penalty)

export const getGiving = summary(Endpoints.giving)
export const getRechargeGiving = summary(Endpoints.rechargeGiving)
export const getRecharge = summary(Endpoints.recharge)
export const getActivityDepositCard = summary(Endpoints.activityDepositCard)
export const getActivityDepositTimes = summary(Endpoints.activityDepositTimes)
export const getActivityRidingCard = summary(Endpoints.activityRidingCard)

export const getRevenuesCostDay = series(Endpoints.revenuesCostDay)
export const getRevenuesNumDay = series(Endpoints.revenuesNumDay)
export const getRevenuesPayTimeDay = series(Endpoints.revenuesPayTimeDay)
export const getRevenueRefundDay = series(Endpoints.revenueRefundDay)
export const getRevenueReportManageDay = series(Endpoints.revenueReportManageDay)
export const getRevenuePenaltyDay = series(Endpoints.revenuePenaltyDay)
export const getRevenueMerchantDays = series(Endpoints.revenueMerchantDays)
export const getDepositCardSellTimes = cardTimes(Endpoints.depositCardSellTimes)
export const getRidingCardSellTimes = cardTimes(Endpoints.ridingCardSellTimes)
export const getRevenueDepositDays = series(Endpoints.revenueDepositDays)
export const getActivityRechargeDay = series(Endpoints.activityRechargeDay)
export const getActivityRechargeGivingDay = series(Endpoints.activityRechargeGivingDay)
export const getActivityGivingDay = series(Endpoints.activityGivingDay)

/** 遗留实现不带入参，返回租户级配置；不要擅自加 serviceIds，见 `revenue-home/displayFactor.ts`。 */
export const getScreenRevenueConfig = () => post<ScreenRevenueConfig>(Endpoints.screenRevenueConfig)
