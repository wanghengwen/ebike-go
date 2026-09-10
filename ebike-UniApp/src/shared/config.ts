import demoTenant from '@/tenants/demo.release.json'
import type { TenantConfig } from '@/tenants/types'
import { logger } from '@/shared/logger'

function readMergedTenant(): TenantConfig | null {
  const modules = import.meta.glob<{ default: TenantConfig }>('../tenants/merged.runtime.json', {
    eager: true,
  })
  const hit = Object.values(modules)[0]
  const cfg = hit?.default
  if (!cfg?.api?.baseUrl) return null
  return {
    ...(demoTenant as TenantConfig),
    ...cfg,
    api: { ...(demoTenant as TenantConfig).api, ...cfg.api },
  }
}

/**
 * Must resolve at module load — do not wait for App.onLaunch.
 * Hot-reload / "appLaunch with non-empty page stack" can skip onLaunch while
 * pages already fire requests; empty demo keys then produce HTTP 400 on sign.
 */
let runtimeTenant: TenantConfig = readMergedTenant() || (demoTenant as TenantConfig)

export function getTenantConfig(): TenantConfig {
  return runtimeTenant
}

export function setTenantConfig(cfg: TenantConfig) {
  runtimeTenant = cfg
}

export function getBrandColor(): string {
  return runtimeTenant.customSetting?.buttonGreenColor || '#3AA0E8'
}

export function getButtonWhiteColor(): string {
  return runtimeTenant.customSetting?.buttonWhiteColor || '#FFFFFF'
}

export function getButtonDisabledColor(): string {
  return runtimeTenant.customSetting?.buttonDisabledColor || '#CCCCCC'
}

export function getApiBaseUrl(): string {
  return runtimeTenant.api?.baseUrl || ''
}

/** Re-apply merged runtime tenant (safe to call from App.onLaunch). */
export function loadTenantRuntime() {
  const cfg = readMergedTenant()
  if (cfg) {
    setTenantConfig(cfg)
    logger.info('tenant runtime loaded', cfg.alias, cfg.platformTenantId)
    if (!cfg.platformTenantId || !cfg.platformSign || !cfg.platformSecret) {
      logger.warn('tenant platform keys incomplete — API sign will fail')
    }
    return
  }
  logger.info('tenant runtime fallback to demo.release')
}
