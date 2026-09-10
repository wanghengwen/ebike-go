import { request } from '@/shared/request'

export function getWechatPayScoreRecord(data: Record<string, unknown> = {}) {
  return request({ url: '/client/pay/score/getPermissionRecord', method: 'POST', data })
}

export function createWechatPayScoreOrder(data: Record<string, unknown>) {
  return request({ url: '/client/pay/score/createOrder', method: 'POST', data })
}

export function getPermissionConfig(data: Record<string, unknown> = {}) {
  return request({ url: '/client/ridingPermission/get', method: 'POST', data })
}

export function getCreditScoreDetail(data: Record<string, unknown> = {}) {
  return request({ url: '/client/ebike-user/creditScore/noRidding/info', method: 'POST', data })
}

export function closeScorePermission(data: Record<string, unknown> = {}) {
  return request({ url: '/client/pay/score/terminatePermission', method: 'POST', data })
}

export function scorePermission(data: Record<string, unknown> = {}) {
  return request({ url: '/client/pay/score/permission', method: 'POST', data })
}

/** Re-export for callers that imported list/config from this module */
export { getCreditScoreList, getCreditScoreConfig } from './user'
