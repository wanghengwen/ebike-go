import { logger } from '@/shared/logger'
import { cancelPay } from '@/api/pay'

export type PayChannel = 'wxpay' | 'alipay' | 'wallet'

export type WxPayParams = {
  timeStamp?: string | number
  nonceStr?: string
  package?: string
  packages?: string
  signType?: string
  paySign?: string
  [key: string]: unknown
}

function unwrapPayPayload(data: Record<string, unknown>): WxPayParams | null {
  let raw: unknown = data?.miniPayRequest ?? data?.pay_info ?? data?.wc_pay_data ?? data
  if (typeof raw === 'string') {
    try {
      raw = JSON.parse(raw)
    } catch {
      return null
    }
  }
  if (!raw || typeof raw !== 'object') return null
  return raw as WxPayParams
}

/** Normalize BAOFU_WXLITE / WXLITE / UMS / UNION createPay response into uni.requestPayment options */
export function normalizeWxPayParams(data: Record<string, unknown>): UniApp.RequestPaymentOptions | null {
  const raw = unwrapPayPayload(data)
  if (!raw) return null

  const timeStamp = raw.timeStamp ?? (raw as { timestamp?: string | number }).timestamp
  const nonceStr = raw.nonceStr ?? (raw as { nonce_str?: string }).nonce_str
  // BAOFU uses `packages`; WXLITE uses `package`
  const pkg =
    raw.package ||
    raw.packages ||
    (raw as { packageValue?: string }).packageValue ||
    ''
  const signType = raw.signType || (raw as { sign_type?: string }).sign_type || 'MD5'
  const paySign = raw.paySign || (raw as { pay_sign?: string }).pay_sign || (raw as { sign?: string }).sign

  if (!timeStamp || !nonceStr || !pkg || !paySign) return null

  return {
    provider: 'wxpay',
    timeStamp: String(timeStamp),
    nonceStr: String(nonceStr),
    package: String(pkg),
    signType: String(signType),
    paySign: String(paySign),
  }
}

export async function invokePayment(
  channel: PayChannel,
  params: Record<string, unknown>,
  opts?: { pin?: string; reportCancel?: boolean },
): Promise<{ success: boolean; data?: unknown }> {
  if (channel === 'wallet') {
    return { success: true }
  }

  // #ifdef MP-WEIXIN
  if (channel === 'wxpay') {
    const payOpts = normalizeWxPayParams(params)
    if (!payOpts) {
      logger.warn('wxpay params incomplete', params)
      return { success: false, data: params }
    }
    return new Promise((resolve) => {
      uni.requestPayment({
        ...payOpts,
        success: (res) => resolve({ success: true, data: res }),
        fail: (err) => {
          logger.warn('wxpay fail', err)
          if (opts?.reportCancel !== false && opts?.pin) {
            cancelPay({ pin: opts.pin }).catch(() => undefined)
          }
          resolve({ success: false, data: err })
        },
      })
    })
  }
  // #endif

  // #ifdef APP-PLUS
  return new Promise((resolve) => {
    uni.requestPayment({
      provider: channel === 'alipay' ? 'alipay' : 'wxpay',
      orderInfo: params.orderInfo as string,
      success: (res) => resolve({ success: true, data: res }),
      fail: (err) => {
        logger.warn('app pay fail', err)
        resolve({ success: false, data: err })
      },
    })
  })
  // #endif

  logger.warn('payment channel unsupported on this platform', channel)
  return { success: false }
}
