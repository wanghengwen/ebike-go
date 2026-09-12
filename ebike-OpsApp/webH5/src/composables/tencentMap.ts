/**
 * 腾讯地图 GL JS 的按需加载。
 *
 * key 走构建期 `VITE_TENCENT_MAP_JS_KEY`，和 Android 侧 `ops.map.tencentKey` 不是同一个：
 * 腾讯位置服务把 Android SDK 和 JSAPI 算两类应用，key 不通用。
 *
 * SDK 只接受 JSONP 式的 `callback` 参数，必须往 window 上挂全局回调，没法用 ESM import。
 * 全应用共用一个 Promise，重复调用不会重复插 script。
 */

/** 只声明用得到的部分；腾讯没有官方 d.ts，全量描一遍不划算。 */
export interface TMapLatLng {
  getLat(): number
  getLng(): number
}

export interface TMapNamespace {
  Map: new (el: HTMLElement, options: Record<string, unknown>) => TMapInstance
  LatLng: new (lat: number, lng: number) => TMapLatLng
  LatLngBounds: new (sw: TMapLatLng, ne: TMapLatLng) => unknown
  MultiMarker: new (options: Record<string, unknown>) => TMapOverlay
  MultiPolyline: new (options: Record<string, unknown>) => TMapOverlay
}

export interface TMapInstance {
  destroy(): void
  setCenter(latLng: TMapLatLng): void
  fitBounds(bounds: unknown, options?: Record<string, unknown>): void
}

export interface TMapOverlay {
  setMap(map: TMapInstance | null): void
  setGeometries(geometries: unknown[]): void
  destroy(): void
}

const CALLBACK = '__tmapReady__'
const SDK_VERSION = '1.exp'

declare global {
  interface Window {
    TMap?: TMapNamespace
    [CALLBACK]?: () => void
  }
}

let loading: Promise<TMapNamespace> | null = null

/** 没配 key 时页面应降级成「不显示地图」，而不是弹一个加载失败。 */
export function hasMapKey(): boolean {
  return Boolean(import.meta.env.VITE_TENCENT_MAP_JS_KEY)
}

export function loadTencentMap(): Promise<TMapNamespace> {
  if (window.TMap) return Promise.resolve(window.TMap)

  loading ??= new Promise<TMapNamespace>((resolve, reject) => {
    const key = import.meta.env.VITE_TENCENT_MAP_JS_KEY
    if (!key) {
      reject(new Error('未配置 VITE_TENCENT_MAP_JS_KEY'))
      return
    }

    const script = document.createElement('script')
    script.async = true
    script.src = `https://map.qq.com/api/gljs?v=${SDK_VERSION}&key=${encodeURIComponent(key)}&callback=${CALLBACK}`

    const cleanup = () => {
      delete window[CALLBACK]
      script.onerror = null
    }

    window[CALLBACK] = () => {
      cleanup()
      if (window.TMap) resolve(window.TMap)
      else reject(new Error('腾讯地图加载完成但 window.TMap 缺失'))
    }
    // key 无效时腾讯返回 200 + 一段报错脚本，callback 永远不触发，这里只兜网络级失败。
    script.onerror = () => {
      cleanup()
      // 失败后允许重试，否则一次网络抖动会让整个会话都用不了地图。
      loading = null
      reject(new Error('腾讯地图脚本加载失败'))
    }

    document.head.appendChild(script)
  })

  return loading
}
