import { request } from '@/shared/request'
import { uploadFile as uploadLocalFile } from '@/shared/upload'

/** Multipart upload — prefer file path (legacy `/client/file/upload`). */
export function uploadFile(filePathOrData: string | Record<string, unknown>) {
  if (typeof filePathOrData === 'string') {
    return uploadLocalFile(filePathOrData).then((url) => ({
      success: Boolean(url),
      data: url,
    }))
  }
  return request({ url: '/client/file/upload', method: 'POST', data: filePathOrData })
}

export function getAddressByLatAndLng(data: Record<string, unknown>) {
  return request({ url: '/client/management/gaode/getAddress', method: 'POST', data })
}

export function fetchRemoteTenantConfig(data: Record<string, unknown> = {}) {
  return request({ url: '/client/tenant/config', method: 'POST', data, auth: false })
}
