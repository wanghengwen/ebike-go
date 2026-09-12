import { computed, ref } from 'vue'
import {
  getALi,
  getActivityDepositCard,
  getActivityRidingCard,
  getDeposit,
  getDepositCard,
  getFavorableCards,
  getGiving,
  getMerchantWallet,
  getOrderPayTime,
  getPayment,
  getPenalty,
  getRecharge,
  getRechargeGiving,
  getRefund,
  getReportManage,
  getRidingCard,
  getTotal,
  getUnion,
  getWeiXin,
  type AmountSummary,
  type SnakeRangeQuery,
} from '@/api'
import type { TimeRange } from '@/composables/useDateRange'
import { fetchDisplayFactor } from './displayFactor'

export interface StatDetailRow {
  label: string
  value: number
  integer?: boolean
}

export interface StatItem {
  label: string
  value: number
  /** 公式连接符，渲染在本项之前；空串表示该项不参与合计，只是附带展示。 */
  symbol: '' | '+' | '-'
  detail: StatDetailRow[]
}

export interface StatGroup {
  key: string
  title: string
  permissionCode: string
  total: number
  /** 合计值后的小字用途说明。 */
  tip: string
  formula: string
  items: StatItem[]
}

const SOURCES = {
  wallet: getMerchantWallet,
  ridingCard: getRidingCard,
  depositCard: getDepositCard,
  favorableCard: getFavorableCards,
  deposit: getDeposit,
  weixin: getWeiXin,
  ali: getALi,
  union: getUnion,
  payment: getPayment,
  orderPayTime: getOrderPayTime,
  refund: getRefund,
  reportManage: getReportManage,
  total: getTotal,
  penalty: getPenalty,
  giving: getGiving,
  rechargeGiving: getRechargeGiving,
  platformRecharge: getRecharge,
  givenDepositCard: getActivityDepositCard,
  givenRidingCard: getActivityRidingCard,
} as const

type SourceKey = keyof typeof SOURCES
type Sources = Record<SourceKey, AmountSummary | null>

const EMPTY_SOURCES = Object.fromEntries(
  (Object.keys(SOURCES) as SourceKey[]).map((key) => [key, null]),
) as Sources

/** 两位小数定档，与遗留 `parseFloat(x).toFixed(2) * 1` 一致。 */
function round(value: number): number {
  return Number(value.toFixed(2))
}

/**
 * 「数据」页签：营收 / 商户 / 结算 / 赠送四组汇总。
 *
 * 展示放大系数逐组适用范围不同，不能统一处理：
 * - 营收统计四项 + 诚信金额：放大后展示；
 * - 结算统计：用原值，这些数字要和代理商对账，放大了就对不上账；
 * - 商户统计、赠送统计：也用原值。
 */
export function useRevenueData() {
  const sources = ref<Sources>({ ...EMPTY_SOURCES })
  const factor = ref(1)

  /** 字段缺失时后续运算得到 NaN，由 `numFormat` 渲染成 `-- --`，不要在这里补 0。 */
  const field = (key: SourceKey, name: string): number | undefined => sources.value[key]?.[name]

  /** 乘展示放大系数并按两位小数定档，与遗留 `(x * factor).toFixed(2) * 1` 一致。 */
  const scaled = (key: SourceKey, name: string): number =>
    round(Number(field(key, name)) * factor.value)
  const raw = (key: SourceKey, name: string): number => Number(field(key, name))

  const revenue = computed<StatGroup>(() => {
    const walletAll = scaled('wallet', 'wallet')
    const walletRefund = scaled('wallet', 'wallet_refund')
    const walletNet = walletAll - walletRefund

    const ridingAll = scaled('ridingCard', 'riding_card')
    const ridingRefund = scaled('ridingCard', 'riding_card_refund')
    // 骑行卡与诚信金额沿用遗留的 ×1000 再除，用来规避浮点误差；
    // 会员卡、优惠卡在遗留里是直接相减，两种写法的口径不完全一致，保留现状。
    const ridingNet = (ridingAll * 1000 - ridingRefund * 1000) / 1000

    const memberAll = scaled('depositCard', 'deposit_card')
    const memberRefund = scaled('depositCard', 'deposit_card_refund')
    const memberNet = memberAll - memberRefund

    const favorableAll = scaled('favorableCard', 'favorable_card')
    const favorableRefund = scaled('favorableCard', 'favorable_card_refund')
    const favorableNet = favorableAll - favorableRefund

    const depositAll = scaled('deposit', 'deposit')
    const depositRefund = scaled('deposit', 'deposit_refund')
    const depositNet = (depositAll * 1000 - depositRefund * 1000) / 1000

    return {
      key: 'revenue',
      title: '营收统计',
      permissionCode: '022101',
      tip: '(用于统计进账收入)',
      formula: '营收统计=钱包充值金额+会员卡金额+骑行卡金额+优惠卡金额',
      total: round(walletNet + ridingNet + memberNet + favorableNet),
      items: [
        {
          label: '钱包充值金额',
          value: walletNet,
          symbol: '',
          detail: [
            { label: '充值总金额', value: walletAll },
            { label: '退款金额', value: walletRefund },
            { label: '净充值金额', value: walletNet },
          ],
        },
        {
          label: '会员卡金额',
          value: memberNet,
          symbol: '+',
          detail: [
            { label: '售卖总金额', value: memberAll },
            { label: '退款金额', value: memberRefund },
            { label: '净售卖金额', value: memberNet },
          ],
        },
        {
          label: '骑行卡金额',
          value: ridingNet,
          symbol: '+',
          detail: [
            { label: '售卖总金额', value: ridingAll },
            { label: '退款金额', value: ridingRefund },
            { label: '净售卖金额', value: ridingNet },
          ],
        },
        {
          label: '优惠卡金额',
          value: favorableNet,
          symbol: '+',
          detail: [
            { label: '售卖总金额', value: favorableAll },
            { label: '退款金额', value: favorableRefund },
            { label: '净售卖金额', value: favorableNet },
          ],
        },
        {
          label: '诚信金额(不用于统计)',
          value: depositNet,
          symbol: '',
          detail: [
            { label: '诚信金总额', value: depositAll },
            { label: '退款金额', value: depositRefund },
            { label: '净诚信金额', value: depositNet },
          ],
        },
      ],
    }
  })

  const merchant = computed<StatGroup>(() => {
    const net = (key: SourceKey): number => raw(key, 'sum') - raw(key, 'refund_sum')
    const detail = (key: SourceKey): StatDetailRow[] => [
      { label: '收入金额', value: raw(key, 'sum') },
      { label: '收入笔数', value: raw(key, 'count_sum'), integer: true },
      { label: '退款金额', value: raw(key, 'refund_sum') },
      { label: '退款笔数', value: raw(key, 'refund_count_sum'), integer: true },
      { label: '净额', value: net(key) },
    ]

    return {
      key: 'merchant',
      title: '商户统计',
      permissionCode: '022102',
      tip: '(用于统计商户收入)',
      formula: '商户统计=微信商户+支付宝商户+银联商户',
      total: round(net('weixin') + net('ali') + net('union')),
      items: [
        { label: '微信商户(净额)', value: net('weixin'), symbol: '', detail: detail('weixin') },
        { label: '支付宝商户(净额)', value: net('ali'), symbol: '+', detail: detail('ali') },
        { label: '银联商户(净额)', value: net('union'), symbol: '+', detail: detail('union') },
      ],
    }
  })

  const settlement = computed<StatGroup>(() => {
    const orderRecharge = raw('payment', 'recharge_cost')
    const refundRecharge = raw('refund', 'recharge_cost')
    const reportRecharge = raw('reportManage', 'recharge_cost')

    const memberNet = raw('depositCard', 'deposit_card') - raw('depositCard', 'deposit_card_refund')
    const ridingNet =
      (raw('ridingCard', 'riding_card') * 1000 - raw('ridingCard', 'riding_card_refund') * 1000) /
      1000
    const favorableNet =
      raw('favorableCard', 'favorable_card') - raw('favorableCard', 'favorable_card_refund')

    return {
      key: 'settlement',
      title: '结算统计',
      permissionCode: '022103',
      tip: '(用于和代理商结算对账)',
      formula:
        '结算统计(充值金额)=实收订单(含调度费)(充值金额)-用户工单退款(充值金额)+举报管理费(充值金额)+会员卡金额+骑行卡金额+优惠卡金额',
      total: round(
        orderRecharge - refundRecharge + reportRecharge + memberNet + ridingNet + favorableNet,
      ),
      items: [
        {
          label: '实收订单(充值金额)',
          value: orderRecharge,
          symbol: '',
          detail: [
            { label: '实收订单总金额', value: raw('payment', 'cost') },
            { label: '充值金额', value: orderRecharge },
            { label: '赠送金额', value: raw('payment', 'present_cost') },
            { label: '实收订单量', value: raw('payment', 'num'), integer: true },
            { label: '历史欠款补缴', value: raw('orderPayTime', 'history_paid_cost') },
            { label: '历史欠款补缴(充值)', value: raw('orderPayTime', 'history_recharge_cost') },
            { label: '历史欠款补缴(赠送)', value: raw('orderPayTime', 'history_present_cost') },
            { label: '新增实收订单', value: raw('orderPayTime', 'new_paid_cost') },
            { label: '新增实收订单(充值)', value: raw('orderPayTime', 'new_recharge_cost') },
            { label: '新增实收订单(赠送)', value: raw('orderPayTime', 'new_present_cost') },
            { label: '调度费', value: raw('penalty', 'penalty') },
          ],
        },
        {
          label: '用户工单退款(充值金额)',
          value: refundRecharge,
          symbol: '-',
          detail: [
            { label: '退款总金额', value: raw('refund', 'cost') },
            { label: '退款充值金额', value: refundRecharge },
            { label: '退款赠送金额', value: raw('refund', 'present_cost') },
          ],
        },
        {
          label: '举报管理费(充值金额)',
          value: reportRecharge,
          symbol: '+',
          detail: [
            { label: '管理费总金额', value: raw('reportManage', 'cost') },
            { label: '收入充值金额', value: reportRecharge },
            { label: '收入赠送金额', value: raw('reportManage', 'present_cost') },
          ],
        },
        {
          label: '会员卡金额',
          value: memberNet,
          symbol: '+',
          detail: [
            { label: '售卖总金额', value: raw('depositCard', 'deposit_card') },
            { label: '退款金额', value: raw('depositCard', 'deposit_card_refund') },
            { label: '净售卖金额', value: memberNet },
          ],
        },
        {
          label: '骑行卡金额',
          value: ridingNet,
          symbol: '+',
          detail: [
            { label: '售卖总金额', value: raw('ridingCard', 'riding_card') },
            { label: '退款金额', value: raw('ridingCard', 'riding_card_refund') },
            { label: '净售卖金额', value: ridingNet },
          ],
        },
        {
          label: '优惠卡金额',
          value: favorableNet,
          symbol: '+',
          detail: [
            { label: '售卖总金额', value: raw('favorableCard', 'favorable_card') },
            { label: '退款金额', value: raw('favorableCard', 'favorable_card_refund') },
            { label: '净售卖金额', value: favorableNet },
          ],
        },
        {
          label: '应收订单(不用于结算)',
          value: raw('total', 'cost'),
          symbol: '',
          detail: [
            { label: '应收订单金额', value: raw('total', 'cost') },
            { label: '应收订单量', value: raw('total', 'num'), integer: true },
          ],
        },
        {
          label: '调度费(已包含在实收订单里)',
          value: raw('penalty', 'penalty'),
          symbol: '',
          detail: [],
        },
      ],
    }
  })

  const givingGroup = computed<StatGroup>(() => {
    const walletGiven =
      raw('giving', 'giving') +
      raw('rechargeGiving', 'recharge_giving') +
      raw('platformRecharge', 'platform_wallet')
    const memberGiven = raw('givenDepositCard', 'deposit_card')
    const ridingGiven = raw('givenRidingCard', 'riding_card')

    return {
      key: 'giving',
      title: '赠送统计',
      permissionCode: '022104',
      tip: '(用于统计赠送的金额，不对账)',
      formula: '赠送统计=钱包余额赠送+会员卡赠送+骑行卡赠送+优惠卡赠送',
      // 合计不含优惠卡赠送，且不做两位小数定档，与遗留一致。
      total: memberGiven + walletGiven + ridingGiven,
      items: [
        {
          label: '钱包余额赠送',
          value: walletGiven,
          symbol: '',
          detail: [
            { label: '活动赠送余额', value: raw('giving', 'giving') },
            { label: '充值赠送余额', value: raw('rechargeGiving', 'recharge_giving') },
            { label: '平台充值余额', value: raw('platformRecharge', 'platform_wallet') },
          ],
        },
        { label: '会员卡赠送', value: memberGiven, symbol: '+', detail: [] },
        { label: '骑行卡赠送', value: ridingGiven, symbol: '+', detail: [] },
        // 优惠卡赠送没有对应接口，遗留页面直接写死 0。
        { label: '优惠卡赠送', value: 0, symbol: '+', detail: [] },
      ],
    }
  })

  const groups = computed<StatGroup[]>(() => [
    revenue.value,
    merchant.value,
    settlement.value,
    givingGroup.value,
  ])

  async function loadAll(range: TimeRange, serviceIds: string[]): Promise<void> {
    const query: SnakeRangeQuery = { ...range, service_ids: serviceIds }
    factor.value = await fetchDisplayFactor()

    await Promise.all(
      (Object.keys(SOURCES) as SourceKey[]).map(async (key) => {
        const { success, data } = await SOURCES[key](query)
        sources.value[key] = success ? (data ?? null) : null
      }),
    )
  }

  return { groups, loadAll }
}
