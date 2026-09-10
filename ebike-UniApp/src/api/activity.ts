import { request } from '@/shared/request'

export function getActivityList(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-marketing/activity_center/default_activity_center_list',
    method: 'POST',
    data,
  })
}

export function regularGetRegisterReward(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-marketing/user_reward/regular_get_register_reward',
    method: 'POST',
    data,
  })
}

export function regularGetVerifyReward(data: Record<string, unknown> = {}) {
  return request({
    url: '/client/ebike-marketing/user_reward/regular_get_verify_reward',
    method: 'POST',
    data,
  })
}
