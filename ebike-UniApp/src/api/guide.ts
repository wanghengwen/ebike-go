import { request } from '@/shared/request'

export function getGuidePage(data: Record<string, unknown>) {
  return request({ url: '/client/helpConfig/getGuidePageConfigByServiceId', method: 'POST', data })
}

export function getSpecialTips(data: Record<string, unknown>) {
  return request({ url: '/client/helpConfig/getSpecialTipsByServiceId', method: 'POST', data })
}
