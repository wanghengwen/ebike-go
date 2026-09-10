import { request } from '@/shared/request'

export function getProtocolsByType(data: Record<string, unknown>) {
  return request({ url: '/client/fence/config/protocol/default', method: 'POST', data })
}
