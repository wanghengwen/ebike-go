import { request } from '@/shared/request'

export function createPay(data: Record<string, unknown>) {
  return request({
    url: '/client/ebike-pay/pay/create',
    method: 'POST',
    data,
  })
}

export function getLastDetail(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/order/detailLast',
    method: 'POST',
    data,
  })
}

export function getWalletInfo(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-account/wallet/get_wallet_info',
    method: 'POST',
    data,
  })
}

export function getRechargeList(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-marketing/recharge_config/list',
    method: 'POST',
    data,
  })
}

export function getRechargeScope(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-marketing/recharge_config/recharge_scope',
    method: 'POST',
    data,
  })
}

export function getRechargeConfig(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-marketing/recharge_config/get_recharge_config',
    method: 'POST',
    data,
  })
}

export function queryFrozen(data: Record<string, unknown>) {
  return request({
    url: '/client/order/queryFrozen',
    method: 'POST',
    data,
  })
}

/** Legacy alias */
export function checkFrozen(data: Record<string, unknown>) {
  return queryFrozen(data)
}

export function setFrozenOrder(data: Record<string, unknown>) {
  return request({
    url: '/client/order/frozen',
    method: 'POST',
    data,
  })
}

/** Legacy typo alias */
export function setFrozonOrder(data: Record<string, unknown>) {
  return setFrozenOrder(data)
}

export function getConfigPay(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/systemConfig/getConfigPay',
    method: 'POST',
    data,
  })
}

export function deductWallet(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/return/deductWallet',
    method: 'POST',
    data,
  })
}

export function closeOrder(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/closeOrder',
    method: 'POST',
    data,
  })
}

/** Report cancelled / failed channel payment */
export function cancelPay(data: Record<string, unknown>) {
  return request({
    url: '/client/ebike-pay/pay/cancel',
    method: 'POST',
    data,
  })
}

/** Legacy alias */
export function payCancel(data: Record<string, unknown>) {
  return cancelPay(data)
}

/** Wallet buy / recharge ledger */
export function walletBuyRecord(data: Record<string, unknown>) {
  return request({
    url: '/ebike_visual/merchant/client/wallet/user_buy_record',
    method: 'POST',
    data,
  })
}

export function getUserBuyRecord(data: Record<string, unknown>) {
  return walletBuyRecord(data)
}

export function ridingCardRecord(data: Record<string, unknown>) {
  return request({
    url: '/ebike_visual/merchant/client/riding_card/user_record',
    method: 'POST',
    data,
  })
}

export function depositRecord(data: Record<string, unknown>) {
  return request({
    url: '/ebike_visual/merchant/client/deposit/user_record',
    method: 'POST',
    data,
  })
}

/** Wallet consumption ledger */
export function walletConsumptionRecord(data: Record<string, unknown>) {
  return request({
    url: '/ebike_visual/merchant/client/wallet/user_consumption_record',
    method: 'POST',
    data,
  })
}

export function getConsumptionRecord(data: Record<string, unknown>) {
  return walletConsumptionRecord(data)
}
