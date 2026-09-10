#!/usr/bin/env node
/**
 * Import legacy wechat/config/*_{mode}.json into UniApp config/.
 * Keeps platformSecret / platformSign inline (same as legacy publish).
 * Usage: node scripts/import-legacy-config.mjs
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '..')
const legacyDir = path.resolve(root, '..', '..', 'wechat', 'config')
const outDir = path.join(root, 'config')

/** Normalize legacy tenant JSON → UniApp TenantConfig shape (secrets included). */
function normalize(raw) {
  const mp = raw.platform?.['mp-weixin'] || {}
  const pay = mp.pay || {}
  const h5 = raw.platform?.h5 || {}
  return {
    name: raw.name || '',
    alias: raw.alias || '',
    description: raw.description || '',
    introduce: raw.introduce || '',
    platformTenantId: String(raw.platformTenantId ?? ''),
    platformSecret: raw.platformSecret || '',
    platformSign: raw.platformSign || '',
    api: {
      baseUrl: raw.api?.baseUrl || '',
      logApi: raw.api?.logApi || '',
    },
    qrDomain: raw.qrDomain || [],
    appId: mp.appid || raw.appId || '',
    auth: {
      wechatGrantType: 'wechat_miniapp',
      wechatPartnerCode: '',
    },
    pay: {
      channelType: pay.channelType || 'WXLITE',
      payBaseApi: pay.payBaseApi || '',
    },
    networkTimeout: raw.networkTimeout || { request: 60000 },
    customSetting: raw.customSetting || {},
    globalStyle: raw.globalStyle || {
      navigationBarTitleText: raw.name || 'E-Bike',
      navigationBarBackgroundColor: '#FFFFFF',
    },
    platform: {
      'mp-weixin': {
        appid: mp.appid || '',
        setting: mp.setting || {},
        permission: mp.permission || {},
        plugins: mp.plugins || {},
      },
      h5: h5,
    },
    logLevel: raw.logLevel,
    returnReturnCarStatistics: raw.returnReturnCarStatistics,
  }
}

if (!fs.existsSync(legacyDir)) {
  console.error(`[import] legacy config dir not found: ${legacyDir}`)
  process.exit(1)
}
fs.mkdirSync(outDir, { recursive: true })

const files = fs.readdirSync(legacyDir).filter((f) => f.endsWith('.json'))
let n = 0
for (const file of files) {
  const raw = JSON.parse(fs.readFileSync(path.join(legacyDir, file), 'utf8'))
  const normalized = normalize(raw)
  const outName = file // keep {env}_{mode}.json naming
  fs.writeFileSync(path.join(outDir, outName), JSON.stringify(normalized, null, 2) + '\n')
  n += 1
  const hasSecret = Boolean(normalized.platformSecret && normalized.platformSign)
  console.log(
    `[import] ${file} → config/${outName} (tenant=${normalized.alias}, appId=${normalized.appId}, secrets=${hasSecret ? 'yes' : 'no'})`,
  )
}
console.log(`[import] done, ${n} files. Source of truth: config/{env}_{mode}.json`)
