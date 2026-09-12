import {
  getActivityGivingDay,
  getActivityRechargeDay,
  getActivityRechargeGivingDay,
  getDepositCardSellTimes,
  getFavorableCardSellTimes,
  getRevenueDepositDays,
  getRevenueMerchantDays,
  getRevenuePenaltyDay,
  getRevenueRefundDay,
  getRevenueReportManageDay,
  getRevenuesCostDay,
  getRevenuesNumDay,
  getRevenuesPayTimeDay,
  getRidingCardSellTimes,
  type ApiResponse,
  type CardSellTimesResult,
  type DailySeriesResult,
  type SnakeRangeQuery,
} from '@/api'
import { isDealPenaltyNum, revenueDataPenaltyNum } from '@/utils/commonNum'

const COLOR = {
  blue: '#3868FF',
  orange: '#FF9C80',
  purple: '#993A9E',
  green: '#1DB996',
  pink: '#FF5F93',
} as const

/**
 * 售卖次数图与「用户充值总金额」用的配色：相对默认表把紫、绿对调并补了粉色，
 * 顺序必须和汇总行的 `color` 一致。
 */
export const CARD_PALETTE = [
  '#3868FF',
  '#FF9C80',
  '#993A9E',
  '#1DB996',
  '#FF5F93',
  '#749f83',
  '#ca8622',
  '#bda29a',
  '#91c7ae',
  '#546570',
  '#c4ccd3',
]

export interface SummaryField {
  /** 汇总值在响应顶层的字段名。 */
  key: string
  label: string
  color: string
  integer?: boolean
  /** 主值下方的小字，`[字段名, 后缀]`，用于实收订单的结算充值 / 结算赠送拆分。 */
  notes?: Array<[key: string, suffix: string]>
}

interface BaseChartConfig {
  key: string
  label: string
  permissionCode: string
  unit: string
  seriesType: 'bar' | 'line'
  /** 存在负值时 Y 轴刻度要保留符号。 */
  signedAxis?: boolean
}

export interface DailyChartConfig extends BaseChartConfig {
  source: 'daily'
  fetch: (query: SnakeRangeQuery) => Promise<ApiResponse<DailySeriesResult>>
  /** 系列字段 → 图例名，顺序即图例与颜色顺序。 */
  fields: Array<[key: string, name: string]>
  summary: SummaryField[]
  /** 是否乘展示放大系数。逐图口径不同（工单退款、举报管理费、充值总额都用原值），不能统一。 */
  scaled: boolean
  /** 计数类指标取 0 位小数。 */
  decimals: 0 | 2
  /** 逐日值的再加工；返回 null 表示当前环境不加工。 */
  dailyAdjust?: () => ((value: number) => number) | null
  /** 覆盖默认配色，用于汇总行颜色与默认表不一致的图。 */
  palette?: string[]
}

export interface CardChartConfig extends BaseChartConfig {
  source: 'cardTimes'
  fetch: (query: SnakeRangeQuery) => Promise<ApiResponse<CardSellTimesResult>>
  sumKey: 'deposit_card_sum' | 'riding_card_sum' | 'favorable_card_sum'
  timesKey: 'deposit_card_times' | 'riding_card_times' | 'favorable_card_times'
  summaryLabel: (name: string) => string
}

export type RevenueChartConfig = DailyChartConfig | CardChartConfig

/** 前三个常驻显示，其余折叠在「展开更多」里。 */
export const PRIMARY_CHART_COUNT = 3

export const REVENUE_CHARTS: RevenueChartConfig[] = [
  {
    key: 'receivableOrder',
    label: '应收订单',
    permissionCode: '02210501',
    unit: '单位:元',
    seriesType: 'line',
    source: 'daily',
    fetch: getRevenuesCostDay,
    fields: [
      ['total', '应收订单'],
      ['paid', '实收订单'],
    ],
    summary: [
      { key: 'total_sum', label: '应收订单金额(元)', color: COLOR.blue },
      { key: 'paid_sum', label: '实收订单金额(元)', color: COLOR.orange },
    ],
    scaled: true,
    decimals: 2,
  },
  {
    key: 'actualOrder',
    label: '实收订单',
    permissionCode: '02210502',
    unit: '单位:元',
    seriesType: 'line',
    source: 'daily',
    fetch: getRevenuesPayTimeDay,
    fields: [
      ['payment_cost', '实收订单'],
      ['history_paid_cost', '历史欠款补缴'],
      ['new_paid_cost', '新增实收订单'],
    ],
    summary: [
      {
        key: 'payment_cost_sum',
        label: '实收订单金额(元)',
        color: COLOR.blue,
        notes: [
          ['recharge_cost_sum', '(结算充值)'],
          ['present_cost_sum', '(结算赠送)'],
        ],
      },
      {
        key: 'history_paid_cost_sum',
        label: '历史欠款补缴(元)',
        color: COLOR.orange,
        notes: [
          ['history_recharge_cost_sum', '(结算充值)'],
          ['history_present_cost_sum', '(结算赠送)'],
        ],
      },
      {
        key: 'new_paid_cost_sum',
        label: '新增实收订单(元)',
        color: COLOR.green,
        notes: [
          ['new_recharge_cost_sum', '(结算充值)'],
          ['new_present_cost_sum', '(结算赠送)'],
        ],
      },
    ],
    scaled: true,
    decimals: 2,
  },
  {
    key: 'orderNum',
    label: '订单量',
    permissionCode: '02210503',
    unit: '单位:单',
    seriesType: 'line',
    source: 'daily',
    fetch: getRevenuesNumDay,
    fields: [
      ['num', '应收订单量'],
      ['paid_num', '实收订单量'],
    ],
    // 实收订单量的汇总值在遗留实现里也是蓝色：颜色由 `datas.length == 2` 决定，
    // 而 `length` 被写死成 1，判断永远不成立。保留原显示。
    summary: [
      { key: 'num_sum', label: '应收订单量(单)', color: COLOR.blue, integer: true },
      { key: 'paid_num_sum', label: '实收订单量(单)', color: COLOR.blue, integer: true },
    ],
    scaled: true,
    decimals: 0,
  },
  {
    key: 'ticketRefund',
    label: '用户工单退款',
    permissionCode: '02210504',
    unit: '单位:元',
    seriesType: 'bar',
    source: 'daily',
    fetch: getRevenueRefundDay,
    fields: [
      ['cost', '用户工单退款'],
      ['recharge_cost', '退款充值金额'],
      ['present_cost', '退款赠送金额'],
    ],
    summary: [
      { key: 'cost_sum', label: '用户工单退款(元)', color: COLOR.blue },
      { key: 'recharge_sum', label: '退款充值金额(元)', color: COLOR.orange },
      { key: 'present_sum', label: '退款赠送金额(元)', color: COLOR.green },
    ],
    scaled: false,
    decimals: 2,
  },
  {
    key: 'reportManage',
    label: '举报管理费',
    permissionCode: '02210505',
    unit: '单位:元',
    seriesType: 'bar',
    source: 'daily',
    fetch: getRevenueReportManageDay,
    fields: [
      ['cost', '举报管理费'],
      ['recharge_cost', '收入充值金额'],
      ['present_cost', '收入赠送金额'],
    ],
    summary: [
      { key: 'cost_sum', label: '举报管理费(元)', color: COLOR.blue },
      { key: 'recharge_sum', label: '收入充值金额(元)', color: COLOR.orange },
      { key: 'present_sum', label: '收入赠送金额(元)', color: COLOR.green },
    ],
    scaled: false,
    decimals: 2,
  },
  {
    key: 'dispatchCost',
    label: '调度费',
    permissionCode: '02210506',
    unit: '单位:元',
    seriesType: 'bar',
    source: 'daily',
    fetch: getRevenuePenaltyDay,
    fields: [['penalty', '调度费统计']],
    summary: [{ key: 'penalty_sum', label: '调度费统计(元)', color: COLOR.blue }],
    scaled: true,
    decimals: 2,
    // 命中 `isDealPenaltyNum` 的链接要把每天的调度费向下取整到 0 / 5 结尾，
    // 合计随之改成「取整后日值累加」，不能再用接口的 penalty_sum，否则和柱子对不上。
    dailyAdjust: () => (isDealPenaltyNum() ? revenueDataPenaltyNum : null),
  },
  {
    key: 'totalRecharge',
    label: '用户充值总金额',
    permissionCode: '02210507',
    unit: '单位:元',
    seriesType: 'bar',
    source: 'daily',
    fetch: getRevenueMerchantDays,
    // 逐日的总额字段就叫 `total_sum`（和顶层汇总同名），不是 `total`。
    fields: [
      ['total_sum', '总充值金额'],
      ['wallet', '钱包充值金额'],
      ['deposit_card', '会员卡售卖金额'],
      ['riding_card', '骑行卡售卖金额'],
      ['favorable_card', '优惠卡售卖金额'],
    ],
    summary: [
      { key: 'total_sum', label: '总充值金额(元)', color: COLOR.blue },
      { key: 'wallet_sum', label: '钱包充值金额(元)', color: COLOR.orange },
      { key: 'deposit_card_sum', label: '会员卡售卖金额(元)', color: COLOR.purple },
      { key: 'riding_card_sum', label: '骑行卡售卖金额(元)', color: COLOR.green },
      { key: 'favorable_card_sum', label: '优惠卡售卖金额(元)', color: COLOR.pink },
    ],
    palette: CARD_PALETTE,
    scaled: false,
    decimals: 2,
  },
  {
    key: 'memberCardTimes',
    label: '会员卡售卖次数',
    permissionCode: '02210508',
    unit: '单位:次',
    seriesType: 'bar',
    signedAxis: true,
    source: 'cardTimes',
    fetch: getDepositCardSellTimes,
    sumKey: 'deposit_card_sum',
    timesKey: 'deposit_card_times',
    summaryLabel: (name) => `${name}次数统计(次)`,
  },
  {
    key: 'ridingCardTimes',
    label: '骑行卡售卖次数',
    permissionCode: '02210509',
    unit: '单位:次',
    seriesType: 'bar',
    signedAxis: true,
    source: 'cardTimes',
    fetch: getRidingCardSellTimes,
    sumKey: 'riding_card_sum',
    timesKey: 'riding_card_times',
    summaryLabel: (name) => `${name}统计(次)`,
  },
  {
    key: 'favorableCardTimes',
    label: '优惠卡售卖次数',
    permissionCode: '02210510',
    unit: '单位:次',
    seriesType: 'bar',
    signedAxis: true,
    source: 'cardTimes',
    fetch: getFavorableCardSellTimes,
    sumKey: 'favorable_card_sum',
    timesKey: 'favorable_card_times',
    summaryLabel: (name) => `${name}次数统计(次)`,
  },
  {
    key: 'deposit',
    label: '诚信金额',
    permissionCode: '02210511',
    unit: '单位:元',
    seriesType: 'line',
    signedAxis: true,
    source: 'daily',
    fetch: getRevenueDepositDays,
    fields: [['deposit', '诚信金额统计']],
    // 「数据」页签的诚信金额会放大，这张图不放大；两处口径本来就不一致，保留。
    summary: [{ key: 'deposit_sum', label: '诚信金额统计(元)', color: COLOR.blue }],
    scaled: false,
    decimals: 2,
  },
  {
    key: 'platformRecharge',
    label: '平台充值余额',
    permissionCode: '02210512',
    unit: '单位:元',
    seriesType: 'bar',
    source: 'daily',
    fetch: getActivityRechargeDay,
    fields: [['platform_wallet', '平台充值余额统计']],
    summary: [{ key: 'platform_wallet_sum', label: '平台充值余额统计(元)', color: COLOR.blue }],
    scaled: false,
    decimals: 2,
  },
  {
    key: 'rechargeGiving',
    label: '充值赠送余额',
    permissionCode: '02210513',
    unit: '单位:元',
    seriesType: 'bar',
    source: 'daily',
    fetch: getActivityRechargeGivingDay,
    fields: [['recharge_giving', '充值赠送余额统计']],
    summary: [{ key: 'recharge_giving_sum', label: '充值赠送余额统计(元)', color: COLOR.blue }],
    scaled: true,
    decimals: 2,
  },
  {
    key: 'activityGiving',
    label: '活动赠送余额',
    permissionCode: '02210514',
    unit: '单位:元',
    seriesType: 'bar',
    source: 'daily',
    fetch: getActivityGivingDay,
    fields: [['giving', '活动赠送余额统计']],
    summary: [{ key: 'giving_sum', label: '活动赠送余额统计(元)', color: COLOR.blue }],
    scaled: true,
    decimals: 2,
  },
]
