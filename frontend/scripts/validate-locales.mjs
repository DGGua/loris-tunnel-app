import { readFileSync } from 'node:fs'
const locales = ['zh-CN', 'zh-TW', 'zh-HK', 'en', 'ru']
let fail = 0
const keySets = new Map()
for (const l of locales) {
  const p = new URL(`../src/locales/${l}.json`, import.meta.url)
  try {
    const data = JSON.parse(readFileSync(p, 'utf8'))
    const flat = (obj, prefix = '') => Object.entries(obj).flatMap(([k, v]) => typeof v === 'object' && v !== null ? flat(v, prefix + k + '.') : [prefix + k])
    keySets.set(l, flat(data).sort())
    console.log(`${l}: OK (${keySets.get(l).length} keys)`)
  } catch (e) {
    fail++
    console.log(`${l}: INVALID - ${e.message}`)
  }
}
const base = keySets.get('zh-CN')
for (const l of locales.slice(1)) {
  const keys = keySets.get(l)
  const missing = base.filter(k => !keys.includes(k))
  const extra = keys.filter(k => !base.includes(k))
  if (missing.length || extra.length) {
    fail++
    console.log(`${l} parity mismatch; missing: ${missing.join(', ')}; extra: ${extra.join(', ')}`)
  }
}
process.exit(fail ? 1 : 0)
