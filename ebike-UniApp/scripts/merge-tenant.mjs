#!/usr/bin/env node
/**
 * Multi-tenant config merge — parity with legacy wechat/build_script/getConfig.js
 *
 * Usage (same as legacy):
 *   npm run tenant:merge -- --env=xiaolongyu --mode=release
 *   npm_config_env=xiaolongyu npm_config_mode=release npm run tenant:merge
 *
 * Reads:  config/{env}_{mode}.json  (primary — includes platformSecret/platformSign like legacy)
 * Optional overlay: tenants/local.secrets.json (only non-empty fields; for local override)
 * Writes:
 *   - src/tenants/merged.runtime.json   (runtime getTenantConfig)
 *   - patches src/manifest.json         (mp-weixin.appid / name / description)
 *   - patches src/pages.json            (globalStyle from tenant)
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '..')

function arg(name, fallback = '') {
  const hit = process.argv.find((a) => a.startsWith(`--${name}=`))
  if (hit) return hit.slice(name.length + 3)
  // npm --env=x sets npm_config_env on some shells
  const npmKey = `npm_config_${name}`
  if (process.env[npmKey]) return process.env[npmKey]
  return fallback
}

function deepMerge(a, b) {
  if (Array.isArray(b)) return b.slice()
  if (b && typeof b === 'object' && !Array.isArray(b)) {
    const out = { ...(a && typeof a === 'object' && !Array.isArray(a) ? a : {}) }
    for (const [k, v] of Object.entries(b)) {
      out[k] = deepMerge(out[k], v)
    }
    return out
  }
  return b === undefined ? a : b
}

const env = arg('env', arg('alias', 'demo'))
const mode = arg('mode', 'release')
const configPath = path.join(root, 'config', `${env}_${mode}.json`)
const defaultsPath = path.join(root, 'config', '_defaults.json')
const secretsPath = path.join(root, 'tenants', 'local.secrets.json')
const outRuntime = path.join(root, 'src', 'tenants', 'merged.runtime.json')
const manifestPath = path.join(root, 'src', 'manifest.json')
const pagesPath = path.join(root, 'src', 'pages.json')

if (!fs.existsSync(configPath)) {
  console.error(`[tenant:merge] missing ${configPath}`)
  console.error(`  Create it or run: npm run tenant:import`)
  console.error(`  Available: list config/*.json`)
  process.exit(1)
}

const defaults = fs.existsSync(defaultsPath)
  ? JSON.parse(fs.readFileSync(defaultsPath, 'utf8'))
  : {}
const tenant = JSON.parse(fs.readFileSync(configPath, 'utf8'))

/** Optional local overlay — only non-empty values, never wipe config secrets with blanks. */
function pickNonEmptyOverlay(raw) {
  if (!raw || typeof raw !== 'object') return {}
  const out = {}
  for (const [k, v] of Object.entries(raw)) {
    if (v == null || v === '') continue
    if (typeof v === 'object' && !Array.isArray(v)) {
      const nested = pickNonEmptyOverlay(v)
      if (Object.keys(nested).length) out[k] = nested
    } else {
      out[k] = v
    }
  }
  return out
}

let overlay = {}
if (fs.existsSync(secretsPath)) {
  const all = JSON.parse(fs.readFileSync(secretsPath, 'utf8'))
  const slice = all[env] || all[tenant.alias] || (all.platformSecret || all.platformSign || all.appId ? all : {})
  overlay = pickNonEmptyOverlay(slice)
  if (Object.keys(overlay).length) {
    console.log('[tenant:merge] applying optional local.secrets overlay (non-empty fields only)')
  }
}

let merged = deepMerge(defaults, tenant)
merged = deepMerge(merged, overlay)

// Normalize pay / appId from nested platform if still present.
// Legacy reads platform['mp-weixin'].pay.channelType — prefer that over _defaults WXLITE.
const mp = merged.platform?.['mp-weixin'] || {}
if (!merged.pay) merged.pay = {}
if (mp.pay?.channelType) {
  merged.pay.channelType = mp.pay.channelType
}
if (mp.pay?.payBaseApi != null && mp.pay.payBaseApi !== '') {
  merged.pay.payBaseApi = mp.pay.payBaseApi
}
if (!merged.appId && mp.appid) merged.appId = mp.appid
if (!merged.globalStyle) {
  merged.globalStyle = {
    navigationBarTitleText: merged.name || 'E-Bike',
    navigationBarBackgroundColor: '#FFFFFF',
  }
}

fs.mkdirSync(path.dirname(outRuntime), { recursive: true })
fs.writeFileSync(outRuntime, JSON.stringify(merged, null, 2) + '\n')
console.log(`[tenant:merge] wrote ${outRuntime}`)
console.log(`[tenant:merge] env=${env} mode=${mode} name=${merged.name} tenantId=${merged.platformTenantId}`)

// --- patch manifest.json ---
if (fs.existsSync(manifestPath)) {
  const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'))
  if (merged.name) manifest.name = merged.name
  if (merged.description) manifest.description = merged.description
  if (!manifest['mp-weixin']) manifest['mp-weixin'] = {}
  const appid = merged.appId || mp.appid
  if (appid) manifest['mp-weixin'].appid = appid
  // Always keep subpackage optimization for mp-weixin main-package size / unused-js checks
  manifest['mp-weixin'].optimization = {
    ...(manifest['mp-weixin'].optimization || {}),
    subPackages: true,
  }
  // Sync location permission desc from tenant (WeChat limit: ≤30 chars)
  const tenantPerm = mp.permission || merged.platform?.['mp-weixin']?.permission
  if (tenantPerm && typeof tenantPerm === 'object') {
    const nextPerm = { ...(manifest['mp-weixin'].permission || {}), ...tenantPerm }
    const loc = nextPerm['scope.userLocation']
    if (loc && typeof loc.desc === 'string' && [...loc.desc].length > 30) {
      loc.desc = [...loc.desc].slice(0, 30).join('')
      console.warn('[tenant:merge] truncated scope.userLocation.desc to 30 chars')
    }
    manifest['mp-weixin'].permission = nextPerm
  } else if (!manifest['mp-weixin'].permission?.['scope.userLocation']?.desc) {
    manifest['mp-weixin'].permission = {
      ...(manifest['mp-weixin'].permission || {}),
      'scope.userLocation': { desc: '需要定位权限以确定您的位置' },
    }
  }
  if (merged.networkTimeout) {
    manifest.networkTimeout = {
      ...(manifest.networkTimeout || {}),
      request: merged.networkTimeout.request || 60000,
      connectSocket: merged.networkTimeout.connectSocket || 60000,
      uploadFile: merged.networkTimeout.uploadFile || 60000,
      downloadFile: merged.networkTimeout.downloadFile || 60000,
    }
  }
  fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2) + '\n')
  console.log(`[tenant:merge] patched manifest.json appid=${appid || '(unchanged)'}`)
}

// --- patch pages.json globalStyle ---
if (fs.existsSync(pagesPath)) {
  const pages = JSON.parse(fs.readFileSync(pagesPath, 'utf8'))
  const gs = merged.globalStyle || {}
  pages.globalStyle = {
    ...(pages.globalStyle || {}),
    navigationBarTitleText: gs.navigationBarTitleText || merged.name || pages.globalStyle?.navigationBarTitleText,
    navigationBarBackgroundColor: gs.navigationBarBackgroundColor || '#FFFFFF',
    backgroundColor: gs.backgroundColor || pages.globalStyle?.backgroundColor || '#F5F6F8',
  }
  fs.writeFileSync(pagesPath, JSON.stringify(pages, null, 2) + '\n')
  console.log(`[tenant:merge] patched pages.json title=${pages.globalStyle.navigationBarTitleText}`)
}
