import { request } from '@/shared/request'

export function getServiceByPoi(data: Record<string, unknown>) {
  return request({
    url: '/client/fence/serviceArea/getByLocation',
    method: 'POST',
    data,
  })
}

export function getNearBike(data: Record<string, unknown>) {
  return request({
    url: '/client/paas/device/eBikeLocation',
    method: 'POST',
    data,
  })
}

/** Alias of getNearFenceByPoiAndService */
export function getNearFence(data: Record<string, unknown>) {
  return request({
    url: '/client/fence/serviceArea/getNearFence',
    method: 'POST',
    data,
  })
}

export function getNearFenceByPoiAndService(data: Record<string, unknown>) {
  return getNearFence(data)
}

export function getFenceByServiceId(data: Record<string, unknown>) {
  return request({
    url: '/client/fence/serviceArea/getFenceByServiceId',
    method: 'POST',
    data,
  })
}

export function getHomeScrollerMsg(data: Record<string, unknown>) {
  return request({
    url: '/client/helpConfig/getHomeScrollerMsgByServiceId/v2',
    method: 'POST',
    data,
  })
}

export function getHomeScrollerMsgById(data: Record<string, unknown>) {
  return request({
    url: '/client/helpConfig/getHomeScrollerMsgById',
    method: 'POST',
    data,
  })
}

export function playBikeVoice(data: Record<string, unknown>) {
  return request({
    url: '/client/paas/device/carSearchVoice',
    method: 'POST',
    data,
  })
}

export function getBillingConfig(data: Record<string, unknown>) {
  return request({
    url: '/client/order/config/get',
    method: 'POST',
    data,
  })
}

export function getNearParkingNum(data: Record<string, unknown>) {
  return request({
    url: '/client/fence/parking/nearParkingNum',
    method: 'POST',
    data,
  })
}

export function navigateToCarOrP(data: Record<string, unknown>) {
  return request({
    url: '/client/management/gaode/v3/navigate',
    method: 'POST',
    data,
  })
}

export function getHomeActivityEntrance(data: Record<string, unknown>) {
  return request({
    url: '/client/helpConfig/getHomeActivityEntranceByServiceId',
    method: 'POST',
    data,
  })
}

export function getHomeActivityEntranceById(data: Record<string, unknown>) {
  return request({
    url: '/client/helpConfig/getHomeActivityById',
    method: 'POST',
    data,
  })
}

export function getHomeNav(data: Record<string, unknown>) {
  return request({
    url: '/client/helpConfig/getHomeNavByServiceId',
    method: 'POST',
    data,
  })
}

export function getAppList(data: Record<string, unknown>) {
  return request({
    url: '/client/fence/resource/management/appList',
    method: 'POST',
    data,
  })
}

export function cancelWeChatPayScore(data: Record<string, unknown>) {
  return request({
    url: '/client/pay/score/cancelOrder',
    method: 'POST',
    data,
  })
}

/** Legacy typo alias */
export function cancelWeChatPaySocre(data: Record<string, unknown>) {
  return cancelWeChatPayScore(data)
}

export function turnOnNavigation(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/navigation/begin',
    method: 'POST',
    data,
  })
}

export function turnOffNavigation(data: Record<string, unknown>) {
  return request({
    url: '/client/rent/navigation/end',
    method: 'POST',
    data,
  })
}

export function resourceViewExposure(data: Record<string, unknown>) {
  return request({
    url: '/client/fence/resourceBit/addExposure',
    method: 'POST',
    data,
  })
}

export function resourceClickExposure(data: Record<string, unknown>) {
  return request({
    url: '/client/fence/resourceBit/addClick',
    method: 'POST',
    data,
  })
}

export function getMainPushRidingCards(data: Record<string, unknown>) {
  return request({
    url: '/client/helpConfig/getIzMainPush',
    method: 'POST',
    data,
  })
}
