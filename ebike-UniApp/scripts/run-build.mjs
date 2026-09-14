#!/usr/bin/env node
/**
 * Build wrapper: merge tenant config then run uni build.
 * Passes --env / --mode to merge (legacy: npm run build:mp-weixin --env=x --mode=release)
 *
 * Usage:
 *   node scripts/run-build.mjs mp-weixin --env=xiaolongyu --mode=release
 *   npm run build:mp-weixin -- --env=xiaolongyu --mode=release
 *   npm run build:mp-weixin --env=xiaolongyu --mode=release
 */
import { spawnSync } from 'node:child_process'
import { existsSync, mkdirSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '..')
const platform = process.argv[2] || 'mp-weixin'
const passthrough = process.argv.slice(3)

function pick(name) {
  const hit = passthrough.find((a) => a.startsWith(`--${name}=`))
  if (hit) return hit.slice(name.length + 3)
  if (process.env[`npm_config_${name}`]) return process.env[`npm_config_${name}`]
  return ''
}

const env = pick('env') || pick('alias') || 'demo'
const mode = pick('mode') || 'release'

const mergeArgs = [`--env=${env}`, `--mode=${mode}`]
console.log(`[build] merge tenant env=${env} mode=${mode}`)
const merge = spawnSync(process.execPath, [path.join(root, 'scripts', 'merge-tenant.mjs'), ...mergeArgs], {
  cwd: root,
  stdio: 'inherit',
})
if (merge.status !== 0) process.exit(merge.status || 1)

// Use the local @dcloudio/vite-plugin-uni CLI. `npx uni` resolves to the
// unrelated npm package `uni@0.0.6` when node_modules is missing / PATH is empty.
const uniCli = path.join(root, 'node_modules', '@dcloudio', 'vite-plugin-uni', 'bin', 'uni.js')
if (!existsSync(uniCli)) {
  console.error('[build] missing @dcloudio/vite-plugin-uni. Run npm install in ebike-UniApp first.')
  process.exit(1)
}

const extraArgs = passthrough.filter(
  (a) => !a.startsWith('--env=') && !a.startsWith('--mode=') && !a.startsWith('--alias='),
)
const uniArgs =
  platform === 'h5' ? [uniCli, 'build', ...extraArgs] : [uniCli, 'build', '-p', platform, ...extraArgs]

console.log(`[build] uni ${uniArgs.slice(1).join(' ')}`)
const build = spawnSync(process.execPath, uniArgs, {
  cwd: root,
  stdio: 'inherit',
})
if (build.status !== 0) process.exit(build.status || 1)

if (platform === 'mp-weixin') {
  console.log('[build] relocate subpackage-only JS out of main package')
  const relocate = spawnSync(
    process.execPath,
    [path.join(root, 'scripts', 'relocate-subpackage-js.mjs')],
    { cwd: root, stdio: 'inherit' },
  )
  if (relocate.status !== 0) process.exit(relocate.status || 1)

  // DevTools may still resolve locales/en-US.js from prior graphs / HMR; ship a zh stub.
  const localesDir = path.join(root, 'dist', 'build', 'mp-weixin', 'locales')
  if (existsSync(localesDir)) {
    mkdirSync(localesDir, { recursive: true })
    writeFileSync(
      path.join(localesDir, 'en-US.js'),
      '"use strict";const e=require("./zh-CN.js");exports.enUS=e.zhCN;\n',
      'utf8',
    )
    console.log('[build] wrote locales/en-US.js stub → zh-CN')
  }
}

process.exit(0)
