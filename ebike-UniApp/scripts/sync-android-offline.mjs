#!/usr/bin/env node
/**
 * Sync uni-app App resources into DCloud HBuilder-Integrate-AS for Android Studio offline debug.
 *
 * Prereq:
 *   1. Download Android offline SDK (HBuilderX 5.24.2026081301) matching CLI 3.0.0-5020420260813003
 *      https://nativesupport.dcloud.net.cn/AppDocs/download/android.html
 *      Baidu: https://pan.baidu.com/s/1AFjLggD7g6ue0iKgZ8yVyA?pwd=jrrb
 *   2. Extract so that this folder exists:
 *      android-offline/HBuilder-Integrate-AS/simpleDemo/
 *   3. npm run build:app-android  (or pass --build)
 *
 * Usage:
 *   node scripts/sync-android-offline.mjs
 *   node scripts/sync-android-offline.mjs --build --env=demo --mode=release
 */
import { spawnSync } from 'node:child_process'
import {
  cpSync,
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '..')
const integrateRoot = path.join(root, 'android-offline', 'HBuilder-Integrate-AS')
const simpleDemo = path.join(integrateRoot, 'simpleDemo')
const assetsApps = path.join(simpleDemo, 'src', 'main', 'assets', 'apps')
const controlXml = path.join(simpleDemo, 'src', 'main', 'assets', 'data', 'dcloud_control.xml')
const manifestSrc = path.join(root, 'src', 'manifest.json')

function argFlag(name) {
  return process.argv.includes(`--${name}`)
}
function argValue(name, fallback = '') {
  const hit = process.argv.find((a) => a.startsWith(`--${name}=`))
  return hit ? hit.slice(name.length + 3) : fallback
}

function readAppId() {
  const manifest = JSON.parse(readFileSync(manifestSrc, 'utf8'))
  return String(manifest.appid || '__UNI__LUOPINGTECH')
}

function findWwwDir() {
  const candidates = [
    path.join(root, 'dist', 'build', 'app-android'),
    path.join(root, 'dist', 'build', 'app'),
  ]

  function walk(d, depth = 0) {
    if (depth > 5 || !existsSync(d)) return null
    const hasManifest = existsSync(path.join(d, 'manifest.json'))
    const hasService =
      existsSync(path.join(d, 'app-service.js')) || existsSync(path.join(d, 'app-config.js'))
    if (hasManifest && hasService) return d
    let entries = []
    try {
      entries = readdirSync(d, { withFileTypes: true })
    } catch {
      return null
    }
    for (const ent of entries) {
      if (!ent.isDirectory()) continue
      if (ent.name === 'node_modules' || ent.name.startsWith('.')) continue
      const hit = walk(path.join(d, ent.name), depth + 1)
      if (hit) return hit
    }
    return null
  }

  for (const dir of candidates) {
    if (existsSync(path.join(dir, 'manifest.json'))) {
      if (
        existsSync(path.join(dir, 'app-service.js')) ||
        existsSync(path.join(dir, 'app-config.js'))
      ) {
        return dir
      }
    }
    if (existsSync(path.join(dir, 'www', 'manifest.json'))) return path.join(dir, 'www')
    const hit = walk(dir)
    if (hit) return hit
  }
  return null
}

function patchControlXml(appid) {
  if (!existsSync(controlXml)) {
    console.warn(`[android-offline] missing ${controlXml}`)
    return
  }
  let xml = readFileSync(controlXml, 'utf8')
  if (/appid\s*=\s*"[^"]*"/.test(xml)) {
    xml = xml.replace(/appid\s*=\s*"[^"]*"/, `appid="${appid}"`)
  } else if (/<apps>[\s\S]*?<\/apps>/.test(xml)) {
    xml = xml.replace(
      /<apps>[\s\S]*?<\/apps>/,
      `<apps>\n    <app appid="${appid}" appver="1.0.0"/>\n  </apps>`,
    )
  }
  writeFileSync(controlXml, xml)
  console.log(`[android-offline] patched dcloud_control.xml appid=${appid}`)
}

function syncWww(appid, wwwDir) {
  mkdirSync(assetsApps, { recursive: true })
  // Remove other demo apps to avoid confusion
  for (const ent of readdirSync(assetsApps, { withFileTypes: true })) {
    if (ent.isDirectory() && ent.name !== appid) {
      rmSync(path.join(assetsApps, ent.name), { recursive: true, force: true })
    }
  }
  const dest = path.join(assetsApps, appid, 'www')
  rmSync(dest, { recursive: true, force: true })
  mkdirSync(path.dirname(dest), { recursive: true })
  cpSync(wwwDir, dest, { recursive: true })
  console.log(`[android-offline] synced ${wwwDir} → ${dest}`)
}

if (!existsSync(simpleDemo)) {
  console.error(`[android-offline] missing Integrate project at:
  ${simpleDemo}

Download Android offline SDK (HBuilderX 5.24.2026081301):
  https://nativesupport.dcloud.net.cn/AppDocs/download/android.html
  Baidu pan: https://pan.baidu.com/s/1AFjLggD7g6ue0iKgZ8yVyA  pwd=jrrb
  和彩云: https://yun.139.com/shareweb/#/w/i/2w2KLgoTHEPz0

Extract HBuilder-Integrate-AS into:
  ${path.join(root, 'android-offline')}
`)
  process.exit(1)
}

if (argFlag('build')) {
  const env = argValue('env', 'demo')
  const mode = argValue('mode', 'release')
  console.log(`[android-offline] building app-android env=${env} mode=${mode}`)
  const build = spawnSync(
    process.execPath,
    [path.join(root, 'scripts', 'run-build.mjs'), 'app-android', `--env=${env}`, `--mode=${mode}`],
    { cwd: root, stdio: 'inherit' },
  )
  if (build.status !== 0) process.exit(build.status || 1)
}

const appid = readAppId()
const wwwDir = findWwwDir()
if (!wwwDir) {
  console.error(
    '[android-offline] no app build output. Run: npm run build:app-android  (or pass --build)',
  )
  process.exit(1)
}

syncWww(appid, wwwDir)
patchControlXml(appid)

console.log(`
[android-offline] done.

Next:
  1. Open Android Studio: "C:\\Program Files\\Android\\Android Studio\\bin\\studio64.exe"
  2. Open project: ${integrateRoot}
  3. Sync Gradle, select simpleDemo, Run on emulator/device

Notes:
  - Set applicationId / packagename to com.luopingtech.ebike.demo (or tenant package)
  - Apply dcloud_appkey from https://dev.dcloud.net.cn/ (离线打包 AppKey)
  - Maps/Payment need AAR + keys per Feature-Android.xls in the SDK zip
`)
