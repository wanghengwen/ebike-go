import { createPay, getLastDetail, deductWallet, closeOrder } from '@/api/pay'
import { getOrderDetail } from '@/api/order'
import { getOpenIdByJsCode } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { storage } from '@/shared/storage'
import { getTenantConfig } from '@/shared/config'
import { t } from '@/locales'
import { uniLoginCode } from '@/features/auth/useAuth'
import { invokePayment, normalizeWxPayParams } from '@/vendor/pay/invokePayment'
import type { ApiResult } from '@/shared/request'
import { logger } from '@/shared/logger'
import { isNative, nativeHost } from '@/shared/nativeHost'

export type PayChannelType = 'BAOFU_WXLITE' | 'WXLITE' | 'UMS_WXLITE' | 'UNION_WXLITE' | string

export type PayMoneyOptions = {
  sale_type: string
  sale_info?: Record<string, unknown>
  channel_type?: PayChannelType
  /** Extra fields merged into createPay body (orderId, amount, …) */
  order?: Record<string, unknown>
  showLoading?: boolean
  /** When false, only createPay (no requestPayment) */
  invokeWx?: boolean
}

function userPin(): string {
  const user = useUserStore()
  return String(user.userInfo?.pin || storage.get<Record<string, unknown>>('userInfo', {})?.pin || '')
}

function serviceId(): string {
  return String(storage.get<string>('serviceId', '') || '')
}

/** Align with legacy: customDefaultConfig.platform['mp-weixin'].pay.channelType */
function channelTypeFromTenant(): PayChannelType {
  const cfg = getTenantConfig()
  const fromMp = cfg.platform?.['mp-weixin']?.pay?.channelType
  if (fromMp) return fromMp
  return cfg.pay?.channelType || 'BAOFU_WXLITE'
}

/** fen → yuan display string */
export function fenToYuan(fen: number | string | null | undefined): string {
  const n = Number(fen ?? 0)
  if (Number.isNaN(n)) return '0.00'
  return (n / 100).toFixed(2)
}

/** Normalize createPay response for uni.requestPayment (BAOFU/WXLITE: package | packages) */
export function normalizePayParams(data: Record<string, unknown>): Record<string, unknown> {
  const normalized = normalizeWxPayParams(data)
  if (normalized) return normalized as unknown as Record<string, unknown>
  return {
    timeStamp: String(data.timeStamp ?? ''),
    nonceStr: data.nonceStr,
    package: data.package || data.packages,
    signType: data.signType || 'MD5',
    paySign: data.paySign,
    orderInfo: data.tradeNo || data.orderInfo,
  }
}

export function usePay() {
  const user = useUserStore()

  async function ensureOpenId(): Promise<string | null> {
    const cached =
      user.loginInfo?.openid || storage.get<{ openid?: string }>('loginInfo', {})?.openid
    if (cached) return String(cached)

    const jsCode = await uniLoginCode()
    if (!jsCode) return null

    const res = await getOpenIdByJsCode({ code: jsCode, type: 'WXLITE' })
    if (!res.success || res.data == null) return null

    const openid =
      typeof res.data === 'string'
        ? res.data
        : String(
            (res.data as { openid?: string; open_id?: string }).openid ||
              (res.data as { open_id?: string }).open_id ||
              '',
          )
    if (!openid) return null

    user.setLoginInfo({ ...user.loginInfo, openid })
    return openid
  }

  function buildCreatePayBody(
    openid: string,
    channelType: PayChannelType,
    opts: PayMoneyOptions,
  ): Record<string, unknown> {
    const saleInfo = { ...(opts.sale_info || {}) }
    if (saleInfo.total_fee != null) {
      saleInfo.total_fee = Math.floor(Number(saleInfo.total_fee))
    }

    const channel_info: Record<string, unknown> = { open_id: openid }
    if (channelType === 'UNION_WXLITE') {
      channel_info.service_id = serviceId()
    }

    return {
      channel_type: channelType,
      channel_info,
      sale_type: opts.sale_type,
      sale_info: saleInfo,
      pin: userPin(),
      service_id: serviceId(),
      ...(opts.order || {}),
    }
  }

  async function loadLastOrder() {
    return getLastDetail({ userPin: userPin(), izNewApp: true })
  }

  /** Prefer detail by orderId (legacy pay.vue); fall back to last order. */
  async function loadOrder(orderId?: string) {
    const oid = String(orderId || '').trim()
    if (oid) {
      const res = await getOrderDetail({
        orderId: oid,
        id: oid,
        userPin: userPin(),
        izNewApp: true,
      })
      if (res.success && res.data) return res
    }
    return loadLastOrder()
  }

  /**
   * Create channel pay (+ optional WeChat requestPayment).
   * Mini program: BAOFU_WXLITE | WXLITE | UMS_WXLITE | UNION_WXLITE.
   * App: native WeChat/Alipay cashier not wired yet — block with clear message.
   */
  async function payMoney(opts: PayMoneyOptions): Promise<ApiResult & { paid?: boolean }> {
    if (isNative()) {
      const host = nativeHost()
      const raw = host
        ? await host.pay({
            sale_type: opts.sale_type,
            sale_info: opts.sale_info,
            order: opts.order,
            channel_type: opts.channel_type,
          })
        : {}
      const msg = String(raw.msg || t('pay.nativeUnsupported'))
      uni.showToast({ title: msg, icon: 'none' })
      return {
        success: false,
        code: String(raw.code || 'UNSUPPORTED'),
        msg,
        paid: false,
      }
    }

    const showLoading = opts.showLoading !== false
    if (showLoading) {
      uni.showLoading({ title: t('pay.payNow'), mask: true })
    }

    try {
      // #ifdef APP-PLUS
      return { success: false, code: 'APP_CHANNEL_UNSUPPORTED', msg: t('pay.appChannelUnsupported'), paid: false }
      // #endif

      // #ifndef APP-PLUS
      if (!opts.sale_type) {
        return { success: false, msg: '售卖类型错误', paid: false }
      }
      const saleInfo = opts.sale_info || {}
      if (opts.sale_type !== 'DEPOSIT' && saleInfo.total_fee == null) {
        return { success: false, msg: '请输入金额', paid: false }
      }

      const openid = await ensureOpenId()
      if (!openid) {
        return { success: false, msg: '获取openid失败', paid: false }
      }

      const channelType = opts.channel_type || channelTypeFromTenant()
      const body = buildCreatePayBody(openid, channelType, opts)
      const created = await createPay(body)
      if (!created.success || !created.data) {
        return { ...created, paid: false }
      }

      if (opts.invokeWx === false) {
        return { ...created, paid: false }
      }

      // Hide loading before requestPayment — mask blocks WeChat cashier sheet.
      if (showLoading) uni.hideLoading()

      const payParams = normalizePayParams(created.data as Record<string, unknown>)
      if (!normalizeWxPayParams(payParams)) {
        logger.warn('wxpay params incomplete after createPay', created.data)
        return { success: false, msg: t('pay.payFail'), data: created.data, paid: false }
      }

      const payRes = await invokePayment('wxpay', payParams, { pin: userPin() })
      return { success: payRes.success, data: payRes.data ?? created.data, paid: payRes.success }
      // #endif
    } finally {
      if (showLoading) {
        try {
          uni.hideLoading()
        } catch {
          /* already hidden */
        }
      }
    }
  }

  /** Wallet balance enough: deduct without channel pay */
  async function payByWallet(payload: Record<string, unknown> = {}) {
    uni.showLoading({ title: t('pay.payNow'), mask: true })
    try {
      return deductWallet({ userPin: userPin(), ...payload })
    } finally {
      uni.hideLoading()
    }
  }

  /** Zero-cost pending order */
  async function closeZeroOrder(orderId: string | number, payload: Record<string, unknown> = {}) {
    uni.showLoading({ title: t('pay.payNow'), mask: true })
    try {
      return closeOrder({ orderId, userPin: userPin(), ...payload })
    } finally {
      uni.hideLoading()
    }
  }

  /**
   * createPay with tenant channel_type, then invokePayment.
   * Convenience wrapper used by recharge / card purchase.
   */
  async function createChannelPay(opts: {
    saleType: string
    totalFee: number
    saleInfo?: Record<string, unknown>
    extra?: Record<string, unknown>
    channelType?: PayChannelType
    invokeWx?: boolean
  }) {
    return payMoney({
      sale_type: opts.saleType,
      channel_type: opts.channelType,
      invokeWx: opts.invokeWx,
      sale_info: {
        total_fee: Math.floor(Number(opts.totalFee) || 0),
        ...(opts.saleInfo || {}),
      },
      order: {
        ...(opts.extra || {}),
      },
    })
  }

  /**
   * waitPayMoney: present cannot cover penalty; apply modify* adjustments.
   * Returns fen still owed after wallet buckets (recharge/present).
   */
  function calcWaitPayMoney(detail: Record<string, unknown>): number {
    if (detail.payCost == null && detail.cost == null) return 0
    let payCost = Number(detail.payCost ?? detail.cost ?? 0)
    const penalty = Number(detail.penalty || 0)
    const recharge = Number(detail.recharge || 0)
    const present = Number(detail.present || 0)
    const modifyPayCost = Number(detail.modifyPayCost || 0)
    const modifyDispatchCost = Number(detail.modifyDispatchCost || 0)
    const modifyHelmetPenalty = Number(detail.modifyHelmetPenalty || 0)

    const finalPenalty = penalty - modifyDispatchCost - modifyHelmetPenalty
    const penaltyCost = finalPenalty > recharge ? finalPenalty - recharge : 0
    const rechargeRemain = recharge - finalPenalty
    void modifyPayCost
    payCost = payCost - finalPenalty
    const ridingCost = payCost - present - (finalPenalty > recharge ? 0 : rechargeRemain)
    const wait = penaltyCost + (ridingCost > 0 ? ridingCost : 0)
    return wait > 0 ? Math.floor(wait) : 0
  }

  /**
   * Settle last ride order:
   * payCost===0 → closeOrder
   * waitPayMoney===0 && payCost>0 → deductWallet
   * waitPayMoney>0 → need recharge first (no custom wallet-pwd channel)
   */
  async function settleLastOrder(detail: Record<string, unknown>) {
    const fresh = await loadLastOrder()
    const d = (fresh.success && fresh.data ? fresh.data : detail) as Record<string, unknown>
    const cost = Number(d.payCost ?? d.cost ?? 0)
    const orderId = d.id || d.orderId

    if (Number(d.izPaid) === 4) {
      return { success: true, msg: 'already paid' }
    }

    if (cost === 0) {
      return closeZeroOrder(orderId as string | number)
    }

    const waitPayMoney = calcWaitPayMoney(d)
    if (waitPayMoney === 0) {
      return payByWallet({})
    }

    return {
      success: false,
      code: 'NEED_RECHARGE',
      msg: t('pay.needRechargeBalance', { m: fenToYuan(waitPayMoney) }),
    }
  }

  /** Prefer payMoney / payByWallet / closeZeroOrder / settleLastOrder */
  async function payOrder(payload: Record<string, unknown>) {
    if (payload.useWallet) {
      return payByWallet(payload)
    }
    if (payload.zeroCost || payload.closeOrder) {
      return closeZeroOrder(payload.orderId as string | number, payload)
    }

    const sale_type = String(payload.sale_type || payload.saleType || 'WALLET')
    const sale_info = (payload.sale_info as Record<string, unknown>) || {
      total_fee: payload.total_fee ?? payload.amount,
      ...(payload.orderId ? { orderId: payload.orderId } : {}),
    }
    return payMoney({
      sale_type,
      sale_info,
      channel_type: payload.channel_type as PayChannelType | undefined,
      order: payload.order as Record<string, unknown> | undefined,
    })
  }

  return {
    loadLastOrder,
    loadOrder,
    ensureOpenId,
    payMoney,
    payByWallet,
    closeZeroOrder,
    createChannelPay,
    settleLastOrder,
    calcWaitPayMoney,
    payOrder,
    channelTypeFromTenant,
    fenToYuan,
  }
}

/** Standalone helper for pages that don't need the full composable */
export async function createChannelPay(opts: {
  saleType: string
  totalFee: number
  saleInfo?: Record<string, unknown>
  extra?: Record<string, unknown>
  channelType?: PayChannelType
  invokeWx?: boolean
}) {
  return usePay().createChannelPay(opts)
}
