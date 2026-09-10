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

const uniArgs =
  platform === 'h5'
    ? ['uni', 'build', ...passthrough.filter((a) => !a.startsWith('--env=') && !a.startsWith('--mode=') && !a.startsWith('--alias='))]
    : [
        'uni',
        'build',
        '-p',
        platform,
        ...passthrough.filter((a) => !a.startsWith('--env=') && !a.startsWith('--mode=') && !a.startsWith('--alias=')),
      ]

console.log(`[build] ${uniArgs.join(' ')}`)
const build = spawnSync(process.platform === 'win32' ? 'npx.cmd' : 'npx', uniArgs, {
  cwd: root,
  stdio: 'inherit',
  shell: process.platform === 'win32',
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
