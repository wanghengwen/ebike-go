import { getCyclingCfg, getIconCfg, getMapCfg } from '@/shared/tenantSkin'

export type RidingFenceTip = {
  text: string
  tip: string
  showTip: boolean
  background: string
  color: string
  icon: string
}

/** Legacy pages/riding/components/ridingType.js — fence banner while riding. */
export function getRidingFenceTip(
  type: unknown,
  dispatchCost: unknown,
  izCanReturn: unknown,
): RidingFenceTip | null {
  const n = Number(type)
  if (!Number.isFinite(n) || !n) return null
  const penalty = Number(dispatchCost || 0)
  const can = Boolean(izCanReturn)
  const fee = (penalty / 100).toFixed(2)
  const warnIcon = getCyclingCfg('tipsWarning') || getMapCfg('notify')
  const notifyIcon = getMapCfg('notify') || warnIcon

  if (n === 1101) {
    return {
      text: '请在区域内骑行,站点P还车',
      tip: '',
      showTip: false,
      background: '#fff',
      color: '#333333',
      icon: notifyIcon,
    }
  }
  if (n === 2101 || n === 2211) {
    const withFee = can && penalty !== 0
    return {
      text: withFee
        ? `不在站点P内,还车收取${fee}元调度费`
        : '站点外，请将车辆停在非机动车道上的站点P',
      tip: withFee ? `在站点P外,还车收取${fee}元调度费` : '在站点外',
      showTip: true,
      background: '#FF5936',
      color: '#fff',
      icon: warnIcon,
    }
  }
  if (n === 3101) {
    return {
      text: '您在禁停区，请骑至站点P内还车',
      tip: '禁停区,停车请注意',
      showTip: true,
      background: '#FF5936',
      color: '#fff',
      icon: warnIcon,
    }
  }
  if (n === 4101 || n === 4102) {
    return {
      text: '您已骑出服务区,车辆已断电',
      tip: '服务区外,骑行请注意',
      showTip: true,
      background: '#FF5936',
      color: '#fff',
      icon: warnIcon,
    }
  }
  return null
}

/** iconCfg cover image for return sheets — legacy returnType.config.js */
export function getReturnCoverImage(returnType: unknown): string {
  const n = Number(returnType)
  if ([2251, 2252].includes(n)) return getIconCfg('helmetUnlockGif')
  if (n === 3101) return getIconCfg('inStop')
  if ([4101, 10001].includes(n)) return getIconCfg('outOfService')
  if ([6101].includes(n)) return getIconCfg('outOfPark')
  if ([2101, 1102, 2201, 2211, 2231, 2241, 2242, 2243].includes(n)) return getIconCfg('outOfPark')
  return getIconCfg('outOfPark')
}
