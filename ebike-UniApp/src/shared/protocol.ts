import { getTenantConfig } from '@/shared/config'
import { navigate } from '@/shared/navigate'

export type ProtocolKey =
  | 'userProtocol'
  | 'privacyProtocol'
  | 'reChargeProtocol'
  | 'aboutUs'
  | 'aboutProtocol'
  | 'useEbike'
  | 'registeredDesc'
  | 'depositAndBalance'
  | 'vehicleProblem'
  | 'ridingInstructions'

export function getProtocolUrl(key: string): string {
  const docs = getTenantConfig().customSetting?.documentCfg || {}
  return docs[key] || ''
}

export function openProtocol(key: ProtocolKey | string) {
  const url = getProtocolUrl(key)
  if (!url) return
  navigate('to', `/pages/webview/webview?url=${encodeURIComponent(url)}`)
}
