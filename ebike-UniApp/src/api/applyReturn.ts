import { request } from '@/shared/request'

export function createReturnEbike(data: Record<string, unknown>) {
  return request({ url: '/client/returnBikeAudit/createReturnBikeAudit', method: 'POST', data })
}

export function queryReturnBikeAudit(data: Record<string, unknown> = {}) {
  return request({ url: '/client/returnBikeAudit/izCapable', method: 'POST', data })
}

export function partMatch(data: Record<string, unknown>) {
  return request({ url: '/client/rent/part/partMatch', method: 'POST', data })
}
