import { request } from '@/shared/request'

export function getMsgList(data: Record<string, unknown>) {
  return request({ url: '/client/messageCenter/page/list', method: 'POST', data })
}

export function getMsgDetail(data: Record<string, unknown>) {
  return request({ url: '/client/messageCenter/getById', method: 'POST', data })
}

export function getUnreadCount(data: Record<string, unknown> = {}) {
  return request({ url: '/client/messageCenter/judgeIsHaveNoReadMsg', method: 'POST', data })
}
