import { request } from '@/shared/request'

export function loginByPassword(data: Record<string, unknown>) {
  return request({
    url: '/oauth/token',
    method: 'POST',
    data,
    auth: false,
  })
}

export function getPersonInfo(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/user/user/personInfo',
    method: 'POST',
    data,
  })
}

export function sendSmsCode(data: Record<string, unknown>) {
  return request({
    url: '/client/code/send',
    method: 'POST',
    data,
  })
}

export function submitAuth(data: Record<string, unknown>) {
  return request({
    url: '/client/user/auth',
    method: 'POST',
    data,
  })
}

export function getUserAccount(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-account/user_account',
    method: 'POST',
    data,
  })
}

export function getOpenIdByJsCode(data: Record<string, unknown>) {
  return request({
    url: '/client/ebike-pay/pay/getOpenIdByJsCode',
    method: 'POST',
    data,
  })
}

export function refundDeposit(data: Record<string, unknown>) {
  return request({
    url: '/client/ebike-pay/pay/refund',
    method: 'POST',
    data,
  })
}

export function getQualificationList(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/user/user/qualificationList',
    method: 'POST',
    data,
  })
}

export function getUseCarConfig(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/system/getUseCarConfig',
    method: 'POST',
    data,
  })
}

/** Legacy getConfigBaseApi — system base config (repair flags, invoice, etc.). */
export function getConfigBaseItem(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/systemConfig/getConfigBaseItem',
    method: 'POST',
    data,
  })
}

export function userLogin(data: Record<string, unknown>) {
  return loginByPassword(data)
}

export function updateUser(data: Record<string, unknown>) {
  return request({
    url: '/client/user/user/update',
    method: 'POST',
    data,
  })
}

/** Scan / unlock face gate */
export function rentCheck(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/user/rentCheck',
    method: 'POST',
    data,
  })
}

export function authIzNeed(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/user/auth/izNeed',
    method: 'POST',
    data,
  })
}

export function authState(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/user/auth/state',
    method: 'POST',
    data,
  })
}

export function blacklistInfo(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/user/blacklist/info',
    method: 'POST',
    data,
  })
}

export function withPhoneAvoidAudit(data: Record<string, unknown>) {
  return request({
    url: '/client/user/changeBind/withPhoneAviodAudit',
    method: 'POST',
    data,
  })
}

export function changeBindAdd(data: Record<string, unknown>) {
  return request({
    url: '/client/user/changBind/add',
    method: 'POST',
    data,
  })
}

export function changeBindCancel(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/user/changeBind/cancel',
    method: 'POST',
    data,
  })
}

export function changeBindWithFace(data: Record<string, unknown>) {
  return request({
    url: '/client/user/changeBind/withFace',
    method: 'POST',
    data,
  })
}

export function getCreditScoreConfig(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-fence/creditScore/getConfig',
    method: 'POST',
    data,
  })
}

export function getCreditScoreList(data: Record<string, unknown>) {
  return request({
    url: '/client/ebike-user/creditScore/page',
    method: 'POST',
    data,
  })
}

/** Legacy aliases */
export function checkUserizNeedAuth(data: Record<string, unknown> = {}) {
  return authIzNeed(data)
}

export function checkVerify(data: Record<string, unknown> = {}) {
  return authState(data)
}

export function getBlacklistDetail(data: Record<string, unknown> = {}) {
  return blacklistInfo(data)
}

export function useBikeBeforeFaceCheck(data: Record<string, unknown> = {}) {
  return rentCheck(data)
}

export function updateAvatar(data: Record<string, unknown>) {
  return updateUser(data)
}

export function withPhoneAviodAudit(data: Record<string, unknown>) {
  return withPhoneAvoidAudit(data)
}

export function changBindAdd(data: Record<string, unknown>) {
  return changeBindAdd(data)
}
