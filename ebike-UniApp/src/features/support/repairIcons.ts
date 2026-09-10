import { getIconCfg } from '@/shared/tenantSkin'

/** Legacy sortArr1 for fixed body-part fault types. */
export const REPAIR_SORT_TYPES = [6, 13, 0, 7, 3, 8, 1, 10, 2, 14, 4, 15, 11, 12, 5, 9]

/** Map repair fault title → iconCfg key (legacy repair.vue part chips). */
const PART_ICON_RULES: Array<{ match: RegExp; key: string }> = [
  { match: /车把|把手|刹/, key: 'handlebar' },
  { match: /车头|龙头/, key: 'carTap' },
  { match: /车灯|大灯|头灯/, key: 'headlights' },
  { match: /二维码|二维/, key: 'qrCode' },
  { match: /车筐|筐/, key: 'basked' },
  { match: /线路|线束|线/, key: 'line' },
  { match: /车座|坐垫|座/, key: 'carSeat' },
  { match: /电池|电瓶/, key: 'battery' },
  { match: /车轮|轮胎|轮/, key: 'wheels' },
  { match: /脚蹬|脚踏/, key: 'pedal' },
  { match: /车撑|脚撑/, key: 'carHang' },
  { match: /挡板|泥板/, key: 'baffle' },
]

export function repairPartIcon(label: unknown): string {
  const text = String(label || '')
  for (const rule of PART_ICON_RULES) {
    if (rule.match.test(text)) return getIconCfg(rule.key)
  }
  return ''
}

export function repairBikeIcon(bikeType: unknown): string {
  return bikeType ? getIconCfg('bicycle') : getIconCfg('electricBicycle')
}

export function repairStatusIcon(state: unknown): string {
  return Number(state) === 2 ? getIconCfg('repairProcessed') : getIconCfg('repairProcessing')
}

export function repairWholeIcon(carModel: unknown): string {
  const m = String(carModel ?? '0')
  return getIconCfg(`Aicarwhole${m}`) || getIconCfg('Aicarwhole0')
}
