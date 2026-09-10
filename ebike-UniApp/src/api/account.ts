import { request } from '@/shared/request'

export function queryCancelAccount(data: Record<string, unknown> = {}) {
  return request({ url: '/client/user/user/cancelAccount', method: 'POST', data })
}

export function submitCancel(data: Record<string, unknown> = {}) {
  return request({ url: '/client/user/user/submitCancel', method: 'POST', data })
}

export function careerAuth(data: Record<string, unknown>) {
  return request({ url: '/client/user/career/upload', method: 'POST', data })
}

export function cancelCareerAuth(data: Record<string, unknown> = {}) {
  return request({ url: '/client/user/career/cancel', method: 'POST', data })
}

export function userEnableConfig(data: Record<string, unknown> = {}) {
  return request({ url: '/client/user/config/enable', method: 'POST', data })
}
