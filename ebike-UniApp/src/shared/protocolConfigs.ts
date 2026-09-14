/** Legacy protocolConfigs.js */
export type ProtocolMeta = {
  type: number
  title: string
  key?: string
}

export const PROTOCOL_CONFIGS: Record<string, ProtocolMeta> = {
  userProtocol: { type: 1, title: '用户协议', key: 'userProtocol' },
  privacyProtocol: { type: 2, title: '隐私协议', key: 'privacyProtocol' },
  critenrion: { type: 3, title: '计费说明', key: 'critenrion' },
  reChargeProtocol: { type: 4, title: '充值协议', key: 'reChargeProtocol' },
  aboutUs: { type: 5, title: '关于我们', key: 'aboutUs' },
}
