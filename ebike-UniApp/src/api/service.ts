import { request } from '@/shared/request'

export function getFaqByServiceId(data: Record<string, unknown>) {
  return request({ url: '/client/helpConfig/getFaqByServiceId', method: 'POST', data })
}

export function getFaqById(data: Record<string, unknown>) {
  return request({ url: '/client/helpConfig/getFaqById', method: 'POST', data })
}

export function getCustomerService(data: Record<string, unknown> = {}) {
  return request({ url: '/client/helpConfig/getCustomerServiceByServiceId', method: 'POST', data })
}

export function getAllService(data: Record<string, unknown> = {}) {
  return request({ url: '/client/fence/serviceArea/getAll', method: 'POST', data })
}

export function getApplyStationConfig(data: Record<string, unknown> = {}) {
  return request({ url: '/client/applicationSite/getConfig', method: 'POST', data })
}

export function siteApplication(data: Record<string, unknown>) {
  return request({ url: '/client/SiteApplication/SiteApplication', method: 'POST', data })
}
