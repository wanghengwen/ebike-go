import { request } from '@/shared/request'

export function getOrderList(data: Record<string, unknown>) {
  return request({
    url: '/client/order/list',
    method: 'POST',
    data,
  })
}

export function getOrderDetail(data: Record<string, unknown>) {
  return request({
    url: '/client/order/detail',
    method: 'POST',
    data,
  })
}

export function createUserTicket(data: Record<string, unknown>) {
  return request({
    url: '/client/userTicket/createUserTicket',
    method: 'POST',
    data,
  })
}
