#!/usr/bin/env node
/**
 * After mp-weixin build: move root modules that are only required by one
 * subpackage into that subpackage, rewriting require() paths.
 *
 * Covers:
 * - api/*.js (legacy)
 * - features/ tree (WeChat optimization.subPackages remaps these into
 *   pages-sub/<pkg>/features/... — file must physically exist there)
 *
 * Multi-subpackage (no main): copy into each consumer, delete root copy.
 * Never leaves a broken require.
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '..')
const distRoot = path.join(root, 'dist', 'build', 'mp-weixin')

if (!fs.existsSync(distRoot)) {
  console.warn('[relocate-js] dist missing, skip')
  process.exit(0)
}

const appJson = JSON.parse(fs.readFileSync(path.join(distRoot, 'app.json'), 'utf8'))
const subRoots = (appJson.subPackages || []).map((s) => String(s.root || '').replace(/\\/g, '/'))

function walk(dir, out = []) {
  if (!fs.existsSync(dir)) return out
  for (const name of fs.readdirSync(dir)) {
    const full = path.join(dir, name)
    const st = fs.statSync(full)
    if (st.isDirectory()) {
      if (name === 'node_modules') continue
      walk(full, out)
    } else if (name.endsWith('.js')) out.push(full)
  }
  return out
}

function toPosix(p) {
  return p.split(path.sep).join('/')
}

function relFromDist(abs) {
  return toPosix(path.relative(distRoot, abs))
}

function ownerOfReferrer(relPath) {
  const p = toPosix(relPath)
  for (const r of subRoots) {
    if (p === r || p.startsWith(r + '/')) return r
  }
  return 'main'
}

function relativeRequire(fromFile, toFile) {
  let rel = toPosix(path.relative(path.dirname(fromFile), toFile))
  if (!rel.startsWith('.')) rel = './' + rel
  return rel
}

/** Resolve a require() argument against the file that contains it. */
function resolveRequire(fromFile, reqPath) {
  if (!reqPath.startsWith('.')) return null
  return path.normalize(path.join(path.dirname(fromFile), reqPath))
}

function rewriteFileRequires(fileAbs, oldFileAbs) {
  let text = fs.readFileSync(fileAbs, 'utf8')
  const re = /require\((['"])([^'"\n]+)\1\)/g
  text = text.replace(re, (m, q, reqPath) => {
    const resolved = resolveRequire(oldFileAbs, reqPath)
    if (!resolved || (!fs.existsSync(resolved) && !fs.existsSync(resolved + '.js'))) {
      const tryNew = resolveRequire(fileAbs, reqPath)
      if (tryNew && fs.existsSync(tryNew)) return m
      if (tryNew && fs.existsSync(tryNew + '.js')) return m
    }
    const target = fs.existsSync(resolved)
      ? resolved
      : fs.existsSync(resolved + '.js')
        ? resolved + '.js'
        : null
    if (!target) return m
    const next = relativeRequire(fileAbs, target)
    return `require(${q}${next}${q})`
  })
  fs.writeFileSync(fileAbs, text)
}

function findReferrers(moduleAbs) {
  const moduleRel = relFromDist(moduleAbs)
  const hits = []
  for (const file of walk(distRoot)) {
    if (path.resolve(file) === path.resolve(moduleAbs)) continue
    const text = fs.readFileSync(file, 'utf8')
    const re = /require\((['"])([^'"\n]+)\1\)/g
    let m
    while ((m = re.exec(text))) {
      const reqPath = m[2]
      let resolved = null
      if (reqPath.startsWith('.')) {
        resolved = resolveRequire(file, reqPath)
        if (resolved && !resolved.endsWith('.js')) resolved += '.js'
      } else if (reqPath.replace(/\\/g, '/').endsWith(moduleRel)) {
        hits.push(file)
        break
      }
      if (resolved && path.resolve(resolved) === path.resolve(moduleAbs)) {
        hits.push(file)
        break
      }
    }
  }
  return hits
}

function rewriteReferrer(ref, oldAbs, destAbs, moduleRel) {
  let text = fs.readFileSync(ref, 'utf8')
  const re = /require\((['"])([^'"\n]+)\1\)/g
  text = text.replace(re, (m, q, reqPath) => {
    if (reqPath.startsWith('.')) {
      const oldResolved = resolveRequire(ref, reqPath)
      const oldJs = oldResolved && !oldResolved.endsWith('.js') ? oldResolved + '.js' : oldResolved
      if (oldJs && path.normalize(oldJs) === path.normalize(oldAbs)) {
        return `require(${q}${relativeRequire(ref, destAbs)}${q})`
      }
    }
    if (reqPath.replace(/\\/g, '/').endsWith(moduleRel)) {
      return `require(${q}${relativeRequire(ref, destAbs)}${q})`
    }
    return m
  })
  fs.writeFileSync(ref, text)
}

/**
 * Dest path under a subpackage, preserving path under the top-level bucket
 * (api/foo.js → sub/api/foo.js, features/map/x.js → sub/features/map/x.js).
 */
function destUnderSub(subRoot, moduleAbs) {
  const rel = relFromDist(moduleAbs)
  return path.join(distRoot, subRoot, rel)
}

function pageOwnersOf(referrers) {
  // Ownership from every referrer (including features/api still in main).
  // Keeps shared infra in main when any main-package module still requires it.
  return new Set(referrers.map((r) => ownerOfReferrer(relFromDist(r))))
}

function collectCandidates() {
  const list = []
  const apiDir = path.join(distRoot, 'api')
  if (fs.existsSync(apiDir)) {
    for (const n of fs.readdirSync(apiDir)) {
      if (n.endsWith('.js')) list.push(path.join(apiDir, n))
    }
  }
  const featuresDir = path.join(distRoot, 'features')
  if (fs.existsSync(featuresDir)) {
    list.push(...walk(featuresDir))
  }
  return list
}

let moved = 0
let skipped = 0

for (const abs of collectCandidates()) {
  if (!fs.existsSync(abs)) {
    skipped++
    continue
  }
  const moduleRel = relFromDist(abs)
  const referrers = findReferrers(abs)
  if (!referrers.length) {
    skipped++
    continue
  }

  const owners = pageOwnersOf(referrers)
  // Still referenced from main package → keep in main
  if (owners.has('main')) {
    skipped++
    continue
  }
  if (owners.size === 0) {
    skipped++
    continue
  }

  const targets = [...owners]
  if (targets.length === 1) {
    const subRoot = targets[0]
    const destAbs = destUnderSub(subRoot, abs)
    fs.mkdirSync(path.dirname(destAbs), { recursive: true })
    const oldAbs = abs
    fs.renameSync(abs, destAbs)
    rewriteFileRequires(destAbs, oldAbs)
    for (const ref of referrers) rewriteReferrer(ref, oldAbs, destAbs, moduleRel)
    console.log(`[relocate-js] ${moduleRel} → ${relFromDist(destAbs)}`)
    moved++
    continue
  }

  for (const subRoot of targets) {
    const destAbs = destUnderSub(subRoot, abs)
    fs.mkdirSync(path.dirname(destAbs), { recursive: true })
    fs.copyFileSync(abs, destAbs)
    rewriteFileRequires(destAbs, abs)
    const refs = referrers.filter((r) => ownerOfReferrer(relFromDist(r)) === subRoot)
    for (const ref of refs) rewriteReferrer(ref, abs, destAbs, moduleRel)
    console.log(`[relocate-js] copy ${moduleRel} → ${relFromDist(destAbs)}`)
  }
  fs.unlinkSync(abs)
  console.log(`[relocate-js] removed main ${moduleRel}`)
  moved++
}

console.log(`[relocate-js] done moved=${moved} skipped=${skipped}`)
