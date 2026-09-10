export type MapPoint = { latitude: number; longitude: number }

export type MapAdapter = {
  platform: 'mp-weixin' | 'app' | 'h5'
}

export function createMapAdapter(): MapAdapter {
  let platform: MapAdapter['platform'] = 'h5'
  // #ifdef MP-WEIXIN
  platform = 'mp-weixin'
  // #endif
  // #ifdef APP-PLUS
  platform = 'app'
  // #endif
  return { platform }
}
