import { request } from '@/shared/request'

export function userAddVoucher(data: Record<string, unknown>) {
  return request({ url: '/client/ebike-marketing/voucher/user_add_voucher', method: 'POST', data })
}
