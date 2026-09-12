/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>
  export default component
}

declare module 'amfe-flexible'

interface ImportMetaEnv {
  readonly VITE_BASE_PATH?: string
  readonly VITE_TENCENT_MAP_JS_KEY?: string
  readonly VITE_DEV_API_HOST?: string
  readonly VITE_DEV_TENANT_ID?: string
  readonly VITE_DEV_SIGN?: string
  readonly VITE_DEV_TENANT_SECRET?: string
  readonly VITE_DEV_ACCESS_TOKEN?: string
  readonly VITE_DEV_REFRESH_TOKEN?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
