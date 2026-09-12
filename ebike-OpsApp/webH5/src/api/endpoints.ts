/**
 * 网关路径表。遗留 `common/api.js` 里另有约 40 个标注「未调用」的 `/mieba/v2/big_screen/*`
 * 路径，本次迁移一并删除，只保留两块大屏真正请求的接口。
 */
export const Endpoints = {
  refreshToken: '/oauth/token',
  userInfo: '/business/ebike-management/user/getUserByToken',
  serviceArea: '/business/fence/serviceArea/getByToken',

  // --- 运营大屏 ---
  carStates: '/business/paas/device/list',
  userSunburst: '/business/user/user/memberStatistic',
  userNum: '/business/user/getAgeStatistic',
  operationData: '/ebike_visual/revenue/business/order/data_info',
  orderInfoTime: '/ebike_visual/revenue/business/order/time',
  orderInfoArea: '/ebike_visual/revenue/business/order/distance',
  taskInfo: '/ebike_operation/platform/business/operation_info',
  orderMount: '/ebike_visual/revenue/business/order/discount',
  ticketOrder: '/ebike_visual/revenue/business/order/refund_nums_day',
  averageData: '/ebike_visual/revenue/business/order/data_days',
  userGrow: '/business/user/getAllUserStatistic',
  carsData: '/business/operatingBigScreen/queryData/getCarNum',

  // --- 营收大屏：营收统计 ---
  merchantWallet: '/ebike_visual/merchant/business/wallet',
  depositCard: '/ebike_visual/merchant/business/deposit_card',
  depositCardTimes: '/ebike_visual/merchant/business/deposit_card/buy_times',
  ridingCard: '/ebike_visual/merchant/business/riding_card',
  ridingCardTimes: '/ebike_visual/merchant/business/riding_card/buy_times',
  favorableCards: '/ebike_visual/merchant/business/favorable_card',
  favorableCardTimes: '/ebike_visual/merchant/business/favorable_card/buy_times',
  favorableCardSellTimes: '/ebike_visual/merchant/business/favorable_card/sell_times',
  deposit: '/ebike_visual/merchant/business/deposit',

  // --- 营收大屏：商户统计 ---
  weixin: '/ebike_visual/merchant/business/weixin',
  ali: '/ebike_visual/merchant/business/ali',
  union: '/ebike_visual/merchant/business/union',

  // --- 营收大屏：结算统计 ---
  payment: '/ebike_visual/revenue/business/order/payment',
  orderPayTime: '/ebike_visual/revenue/business/order/pay_time',
  refund: '/ebike_visual/revenue/business/order/refund',
  reportManage: '/ebike_visual/revenue/business/order/report_manage',
  total: '/ebike_visual/revenue/business/order/total',
  penalty: '/ebike_visual/revenue/business/order/penalty',

  // --- 营收大屏：赠送统计 ---
  giving: '/ebike_visual/activity/business/wallet/giving',
  rechargeGiving: '/ebike_visual/activity/business/wallet/recharge_giving',
  recharge: '/ebike_visual/activity/business/wallet/recharge_platform',
  activityDepositCard: '/ebike_visual/activity/business/deposit_card/money',
  activityDepositTimes: '/ebike_visual/activity/business/deposit_card/times',
  activityRidingCard: '/ebike_visual/activity/business/riding_card/money',

  // --- 营收大屏：趋势折线 ---
  revenuesCostDay: '/ebike_visual/revenue/business/order/cost_day',
  revenuesNumDay: '/ebike_visual/revenue/business/order/num_day',
  revenuesPayTimeDay: '/ebike_visual/revenue/business/order/pay_time_day',
  revenueRefundDay: '/ebike_visual/revenue/business/order/refund_day',
  revenueReportManageDay: '/ebike_visual/revenue/business/order/report_manage_day',
  revenuePenaltyDay: '/ebike_visual/revenue/business/order/penalty_day',
  revenueMerchantDays: '/ebike_visual/merchant/business/days',
  depositCardSellTimes: '/ebike_visual/merchant/business/deposit_card/sell_times',
  ridingCardSellTimes: '/ebike_visual/merchant/business/riding_card/sell_times',
  revenueDepositDays: '/ebike_visual/merchant/business/deposit/days',
  activityRechargeDay: '/ebike_visual/activity/business/wallet/recharge_platform_day',
  activityRechargeGivingDay: '/ebike_visual/activity/business/wallet/recharge_giving_day',
  activityGivingDay: '/ebike_visual/activity/business/wallet/giving_day',

  screenRevenueConfig: '/business/fence/bigScreen/getConfigByServiceId',
} as const
