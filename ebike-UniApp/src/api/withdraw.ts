import { request } from '@/shared/request'

export function getWithdrawRecord(data: Record<string, unknown>) {
  return request({ url: '/client/ebike-pay/pay/withdraw/page', method: 'POST', data })
}

export function createWithdraw(data: Record<string, unknown>) {
  return request({ url: '/client/ebike-pay/pay/withdraw', method: 'POST', data })
}

export function getWithdrawConfig(data: Record<string, unknown> = {}) {
  return request({ url: '/client/systemConfig/getConfigBaseItem', method: 'POST', data })
}
