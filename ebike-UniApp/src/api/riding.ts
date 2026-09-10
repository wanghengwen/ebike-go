import { request } from '@/shared/request'

export function getRideInfo(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/rent/getRideInfo',
    method: 'POST',
    data,
  })
}

export function networkRide(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/network/ride',
    method: 'POST',
    data,
  })
}

export function returnPermission(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/returnPermission',
    method: 'POST',
    data,
  })
}

export function returnByNet(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/network/return',
    method: 'POST',
    data,
  })
}

export function tempPark(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/tempParking',
    method: 'POST',
    data,
  })
}

export function endTempPark(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/endParking',
    method: 'POST',
    data,
  })
}

export function getCarInfo(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/getCarInfo',
    method: 'POST',
    data,
  })
}

export function unlockHelmet(data: Record<string, unknown>) {
  return request({
    url: '/client/helmet/unlock',
    method: 'POST',
    data,
  })
}

export function unFrozenOrder(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/rent/unFrozenOrder',
    method: 'POST',
    data,
  })
}

export function tempUnlock(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/tempUnlock',
    method: 'POST',
    data,
  })
}

export function getTempUnlockConfig(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/rent/getTempUnlockConfig',
    method: 'POST',
    data,
  })
}

/** 还车配置 */
export function getBackCarConfig(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/system/getbackCarConfig',
    method: 'POST',
    data,
  })
}

export function getReturnCarConfig(data: Record<string, unknown> = {}) {
  return getBackCarConfig(data)
}

/** 能否申请摄像头免罚 */
export function izCanCameraAudit(data: Record<string, unknown>) {
  return request({
    url: '/client/returnBikeAudit/izCanCameraAudit',
    method: 'POST',
    data,
  })
}

export function isCanCameraAudit(data: Record<string, unknown>) {
  return izCanCameraAudit(data)
}

export function partMatchByTempParking(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/part/partMatchByTempParking',
    method: 'POST',
    data,
  })
}
