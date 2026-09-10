import { request } from '@/shared/request'

export function getUserRidingCard(data: Record<string, unknown> = {}) {
  return request({ url: '/client/ebike-account/riding_card/get_riding_card', method: 'POST', data })
}

export function getRidingCardList(data: Record<string, unknown> = {}) {
  return request({ url: '/client/ebike-marketing/card/riding_config_list', method: 'POST', data })
}

export function ridingConfigGetRule(data: Record<string, unknown> = {}) {
  return request({ url: '/client/ebike-marketing/card/riding_config_get_rule', method: 'POST', data })
}

export function getServiceRidingCard(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-account/riding_card/get_service_riding_card',
    method: 'POST',
    data,
  })
}
