import { request } from '@/shared/request'

export function getInviteDetail(data: Record<string, unknown> = {}) {
  return request({ url: '/client/ebike-marketing/invite/detail', method: 'POST', data })
}

export function getInviteRecord(data: Record<string, unknown> = {}) {
  return request({ url: '/client/ebike-marketing/invite/record', method: 'POST', data })
}

export function createInvite(data: Record<string, unknown> = {}) {
  return request({ url: '/client/ebike-marketing/invite/create', method: 'POST', data })
}

export function acceptInvite(data: Record<string, unknown>) {
  return request({ url: '/client/ebike-marketing/invite/accept', method: 'POST', data })
}

export function getInviteRule(data: Record<string, unknown> = {}) {
  return request({ url: '/client/ebike-marketing/invite/rule', method: 'POST', data })
}
