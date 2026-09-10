import { request } from '@/shared/request'

export function submitScanData(data: Record<string, unknown>) {
  return request({ url: '/client/rent/scan', method: 'POST', data })
}

export function getPreCyclingByEBikeId(data: Record<string, unknown>) {
  return request({ url: '/client/rent/getCarInfo', method: 'POST', data })
}

export function openBikeByNetwork(data: Record<string, unknown>) {
  return request({ url: '/client/rent/network/ride', method: 'POST', data })
}
