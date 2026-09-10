import { request } from '@/shared/request'

/** Legacy `/client/ebike-marketing/redPaketCar/getRule` */
export function getRedEnvelopeRule(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-marketing/redPaketCar/getRule',
    method: 'POST',
    data,
  })
}
