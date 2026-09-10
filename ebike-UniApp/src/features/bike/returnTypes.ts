/** Fence returnType codes from returnPermission (not the 0/1 used by returnByNet). */

export type ReturnKind = 'normal' | 'unnormal'

export type ReturnTypeMeta = {
  reasonKey: string
  tipsKey?: string
  outOfServiceKey?: string
  popupType?: 'bottom' | 'dialog'
  titleKey?: string
}

export const RETURN_TYPE_META: Record<number, ReturnTypeMeta> = {
  1101: { reasonKey: 'returnSheet.normal' },
  1102: { reasonKey: 'returnSheet.outOfPark' },
  2101: { reasonKey: 'returnSheet.outOfPark' },
  2201: { reasonKey: 'returnSheet.outOfPark' },
  2211: { reasonKey: 'returnSheet.outOfPark' },
  2231: { reasonKey: 'returnSheet.outOfPark' },
  2241: { reasonKey: 'returnSheet.cameraDirection' },
  2242: { reasonKey: 'returnSheet.cameraPoint' },
  2243: { reasonKey: 'returnSheet.cameraMiss' },
  2251: { reasonKey: 'returnSheet.helmetMiss' },
  2252: { reasonKey: 'returnSheet.helmetMiss' },
  3101: { reasonKey: 'returnSheet.noParking' },
  3102: { reasonKey: 'returnSheet.helmetMiss' },
  4101: {
    reasonKey: 'returnSheet.outService',
    tipsKey: 'returnSheet.outServiceTips',
    outOfServiceKey: 'returnSheet.outServiceTips',
  },
  4102: { reasonKey: 'returnSheet.helmetMiss' },
  5101: { reasonKey: 'returnSheet.forbidZone' },
  6101: {
    titleKey: 'returnSheet.fullPileTitle',
    reasonKey: 'returnSheet.fullPile',
    popupType: 'dialog',
  },
  10001: {
    reasonKey: 'returnSheet.outService',
    tipsKey: 'returnSheet.outServiceTips',
    outOfServiceKey: 'returnSheet.outServiceTips',
  },
}

export function decideReturnKind(type: unknown): ReturnKind {
  return String(type ?? '').startsWith('11') ? 'normal' : 'unnormal'
}

/** UI matrix after returnPermission: civilization / penalty / guide / block / direct. */
export type ReturnFlowKind =
  | 'civilization'
  | 'penalty_sheet'
  | 'guide'
  | 'block_sheet'
  | 'normal'
  | 'silent'

export function decideReturnFlow(
  perm: { izCanReturn?: boolean; returnType?: unknown },
  izCivilizationRemind: boolean,
): ReturnFlowKind {
  const can = Boolean(perm.izCanReturn)
  const kind = decideReturnKind(perm.returnType)
  if (can && kind === 'normal') {
    return izCivilizationRemind ? 'civilization' : 'normal'
  }
  if (can && kind === 'unnormal') return 'penalty_sheet'
  // !izCanReturn — legacy goReturnBikeGuide: only known codes; unknown = silent
  if (guidePageType(perm.returnType) != null) return 'guide'
  if (showSheetWhenCannotReturn(perm.returnType)) return 'block_sheet'
  return 'silent'
}

/** Legacy applyReturnEbike applyType: 1 out-of-spot / 2 no-parking / 3 other */
export function applyTypeFrom(type: unknown): 1 | 2 | 3 {
  const s = String(type ?? '')
  if (s.startsWith('2')) return 1
  if (s.startsWith('3')) return 2
  return 3
}

/** When izCanReturn=false, which guide pageType to open (customizedReturn). */
export function guidePageType(type: unknown): number | null {
  const n = Number(type)
  switch (n) {
    case 2201:
      return 4 // directional
    case 2211:
      return 5 // RFID
    case 2251:
    case 2252:
    case 4102:
    case 3102:
      return 6 // helmet
    case 2231:
      return 8 // kickstand
    case 2243:
      return 11 // camera
    default:
      return null
  }
}

/** When izCanReturn=false but still show penalty sheet instead of guide.
 * Legacy goReturnBikeGuide: 2101/3101/4101/6101 only (10001 is ride-info auto popup, not permission).
 */
export function showSheetWhenCannotReturn(type: unknown): boolean {
  const n = Number(type)
  return [2101, 3101, 4101, 6101].includes(n)
}

/** Guide page photo-apply applyType (legacy takePhotoReturnCarType). */
export function guideApplyTypeFrom(pageType: number): number {
  // pageType 11 → applyType 7 (camera); others map 1:1
  if (pageType === 11) return 7
  return pageType
}

export function getReturnMeta(type: unknown): ReturnTypeMeta {
  const n = Number(type)
  return (
    RETURN_TYPE_META[n] || {
      reasonKey: 'returnSheet.outOfPark',
    }
  )
}
