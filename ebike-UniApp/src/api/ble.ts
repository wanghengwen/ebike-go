import { request } from '@/shared/request'

export function getBlueToothToken(data: Record<string, unknown>) {
  return request({
    url: '/client/paas/device/getBlueToothToken',
    method: 'POST',
    data,
  })
}

export function bleRidePermission(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/blue/ridePermission',
    method: 'POST',
    data,
  })
}

export function bleRideReport(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/blue/ride',
    method: 'POST',
    data,
  })
}

export function bleReturnReport(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/blue/return',
    method: 'POST',
    data,
  })
}

export function bleTempParkReport(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/blue/tempParking',
    method: 'POST',
    data,
  })
}

export function bleEndTempParkReport(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/blue/endParking',
    method: 'POST',
    data,
  })
}

/** BLE unlock: accessories to check before open */
export function openConfig(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/blue/getReturnConfig',
    method: 'POST',
    data,
  })
}

/** Alias of openConfig path */
export function getReturnConfig(data: Record<string, unknown>) {
  return openConfig(data)
}

/** BLE return: accessories to check before lock */
export function getbackCarConfigByCarId(data: Record<string, unknown>) {
  return request({
    url: '/client/system/getbackCarConfigByCarId',
    method: 'POST',
    data,
  })
}

/** Legacy alias of getbackCarConfigByCarId */
export function returnConfig(data: Record<string, unknown>) {
  return getbackCarConfigByCarId(data)
}
