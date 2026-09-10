import { request } from '@/shared/request'

export function submitSneak(data: Record<string, unknown>) {
  // Legacy wechat imported this from repair.js but export was missing; path follows operation module style
  return request({ url: '/client/operation/sneak/add', method: 'POST', data })
}

export function getRepairConfigList(data: Record<string, unknown>) {
  return request({ url: '/client/management/repairConfig/listByCar', method: 'POST', data })
}

export function submitRepair(data: Record<string, unknown>) {
  return request({ url: '/client/operation/repair/add', method: 'POST', data })
}

export function repairList(data: Record<string, unknown>) {
  return request({ url: '/client/operation/repair/page', method: 'POST', data })
}
