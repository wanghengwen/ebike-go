#!/usr/bin/env node
/**
 * Extract a downloaded Android offline SDK zip into android-offline/.
 *
 * Usage:
 *   node scripts/extract-android-sdk.mjs path/to/SDK.zip
 *   node scripts/extract-android-sdk.mjs   # auto-detect under android-offline/sdk-download or Downloads
 */
import { spawnSync } from 'node:child_process'
import { existsSync, mkdirSync, readdirSync, statSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import os from 'node:os'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '..')
const offline = path.join(root, 'android-offline')
const downloadDir = path.join(offline, 'sdk-download')

function findZip(explicit) {
  if (explicit && existsSync(explicit)) return explicit
  const dirs = [
    downloadDir,
    path.join(os.homedir(), 'Downloads'),
    path.join(os.homedir(), 'Desktop'),
  ]
  const hits = []
  for (const dir of dirs) {
    if (!existsSync(dir)) continue
    for (const name of readdirSync(dir)) {
      if (!/\.zip$/i.test(name)) continue
      if (!/android|sdk|hbuilder|uni/i.test(name)) continue
      const full = path.join(dir, name)
      hits.push({ full, mtime: statSync(full).mtimeMs })
    }
  }
  hits.sort((a, b) => b.mtime - a.mtime)
  return hits[0]?.full || ''
}

const zip = findZip(process.argv[2])
if (!zip) {
  console.error(`[extract-android-sdk] zip not found.

1. Download HBuilderX 5.24.2026081301 Android offline SDK:
   https://pan.baidu.com/s/1AFjLggD7g6ue0iKgZ8yVyA?pwd=jrrb
2. Save zip to:
   ${downloadDir}
   or pass path: node scripts/extract-android-sdk.mjs D:\\path\\SDK.zip
`)
  process.exit(1)
}

mkdirSync(offline, { recursive: true })
console.log(`[extract-android-sdk] extracting ${zip} → ${offline}`)
const ps = spawnSync(
  'powershell.exe',
  [
    '-NoProfile',
    '-Command',
    `Expand-Archive -LiteralPath '${zip.replace(/'/g, "''")}' -DestinationPath '${offline.replace(/'/g, "''")}' -Force`,
  ],
  { stdio: 'inherit' },
)
if (ps.status !== 0) process.exit(ps.status || 1)

// Some zips nest one extra folder — lift HBuilder-Integrate-AS if found deeper
function findIntegrate(dir, depth = 0) {
  if (depth > 4 || !existsSync(dir)) return null
  const target = path.join(dir, 'HBuilder-Integrate-AS')
  if (existsSync(path.join(target, 'simpleDemo'))) return target
  for (const name of readdirSync(dir)) {
    const full = path.join(dir, name)
    try {
      if (statSync(full).isDirectory()) {
        const hit = findIntegrate(full, depth + 1)
        if (hit) return hit
      }
    } catch {
      /* ignore */
    }
  }
  return null
}

const found = findIntegrate(offline)
if (found && found !== path.join(offline, 'HBuilder-Integrate-AS')) {
  console.log(`[extract-android-sdk] found Integrate at ${found}`)
  console.log('[extract-android-sdk] please ensure android-offline/HBuilder-Integrate-AS points to that folder')
}

if (existsSync(path.join(offline, 'HBuilder-Integrate-AS', 'simpleDemo'))) {
  console.log('[extract-android-sdk] OK — next: npm run android:sync')
} else {
  console.warn(
    '[extract-android-sdk] HBuilder-Integrate-AS/simpleDemo not found yet; check zip layout under android-offline/',
  )
}
