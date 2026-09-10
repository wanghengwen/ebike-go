import { ref } from 'vue'
import { getGuidePage, getSpecialTips } from '@/api/guide'
import { storage } from '@/shared/storage'
import { useUserStore } from '@/stores/user'
import { logger } from '@/shared/logger'

export const PopUpTime = {
  beforeGuide: 0,
  afterGuide: 1,
  beforeUseCar: 2,
  endTrip: 3,
} as const

type PopupItem = Record<string, unknown> & {
  id?: string | number
  frequency?: number
  popUpTime?: number
}

type GuideConfig = Record<string, unknown> & {
  frequency?: number
  guidePages?: Array<Record<string, unknown>>
  allowSuperEsc?: boolean
  pageNumEsc?: number
}

type HistoryItem = { id: string | number; invalidDate: number | null }

const popupList = ref<PopupItem[]>([])
const guideConfig = ref<GuideConfig>({})
const currPopup = ref<PopupItem>({})
const showPopup = ref(false)
const showGuide = ref(false)
let beforeUseResolve: (() => void) | null = null
let fetching: Promise<void> | null = null
/** 本次冷启动成功跑过一次首页弹窗链路后置 true（防止每次 onShow 重复弹） */
let homeGuideAttemptedThisLaunch = false

function todayDay() {
  return new Date().getDate()
}

/** 对齐旧版 guideAndPopupMixins.getPopupInfoByStorageForJudge */
function judgePopup(list: PopupItem[], popUpTime: number): { state: boolean; popupItem: PopupItem } {
  const popupItem = (list || []).find((el) => Number(el.popUpTime) === popUpTime)
  if (!popupItem) return { state: false, popupItem: {} }
  const history = storage.get<HistoryItem[]>('popupHistoryList', null)
  if (!history) return { state: true, popupItem }
  const { id, frequency } = popupItem
  if (Number(frequency) === 0) {
    const hit = history.find((el) => el.id === id && el.invalidDate === todayDay())
    return hit ? { state: false, popupItem: {} } : { state: true, popupItem }
  }
  // frequency==1：每次调用都可弹；由 homeGuideAttemptedThisLaunch 限制为冷启动一次
  return { state: true, popupItem }
}

/** 对齐旧版 getGuideInfoByStorageForJudge */
function judgeGuide(cfg?: GuideConfig | null): boolean {
  const pages = cfg?.guidePages
  if (!pages || !pages.length) return false
  const history = storage.get<{ invalidDate?: number }>('guideHistory', null)
  if (!history) return true
  if (Number(cfg?.frequency) === 0) return history.invalidDate !== todayDay()
  // frequency!=0：旧版每次可弹；由 session 门控限制为冷启动一次
  return true
}

function markPopupShown(current: PopupItem, list: PopupItem[]) {
  if (current?.id == null) return
  const day = todayDay()
  const history = storage.get<HistoryItem[]>('popupHistoryList', null)
  if (history && Array.isArray(history)) {
    const next = history.map((el) =>
      el.id === current.id ? { ...el, invalidDate: day } : el,
    )
    if (!next.some((el) => el.id === current.id)) {
      next.push({ id: current.id, invalidDate: day })
    }
    storage.set('popupHistoryList', next)
    return
  }
  storage.set(
    'popupHistoryList',
    (list || []).map((el) => ({
      id: el.id as string | number,
      invalidDate: el.id === current.id ? day : null,
    })),
  )
}

function markGuideShown() {
  storage.set('guideHistory', { invalidDate: todayDay(), isAleardOpen: true })
}

function openPopup(item: PopupItem) {
  if (!item || !Object.keys(item).length) return
  currPopup.value = item
  showPopup.value = true
  markPopupShown(item, popupList.value)
}

function openGuide() {
  if (!judgeGuide(guideConfig.value)) return false
  showGuide.value = true
  markGuideShown()
  return true
}

function hasGuidePin(): { serviceId: string; userPin: string } | null {
  const user = useUserStore()
  user.hydrateFromStorage()
  const serviceId = storage.get<string>('serviceId', '') || ''
  const userPin = String(
    user.userInfo?.pin || storage.get<Record<string, unknown>>('userInfo', {})?.pin || '',
  )
  if (!serviceId || !userPin) return null
  return { serviceId, userPin }
}

async function fetchConfigs() {
  if (fetching) return fetching
  fetching = (async () => {
    const keys = hasGuidePin()
    if (!keys) return
    try {
      const [popupRes, guideRes] = await Promise.all([
        getSpecialTips({ serviceId: keys.serviceId, userPin: keys.userPin }),
        getGuidePage({ serviceId: keys.serviceId, userPin: keys.userPin }),
      ])
      const tips = (popupRes.success ? popupRes.data : null) as PopupItem[] | null
      const guide = (guideRes.success ? guideRes.data : null) as GuideConfig | null
      popupList.value = Array.isArray(tips) ? tips : []
      guideConfig.value = guide && typeof guide === 'object' ? guide : {}
      storage.set('popupList', popupList.value)
      storage.set('guideList', guideConfig.value)
    } catch (e) {
      logger.warn('guide/popup fetch soft fail', e)
    } finally {
      fetching = null
    }
  })()
  return fetching
}

/**
 * 首页弹窗：冷启动最多成功跑一次。
 * serviceId/pin 未就绪时不标记，等下次 onShow 再试（避免彻底弹不出来）。
 */
export async function initHomeOrMapGuide() {
  if (homeGuideAttemptedThisLaunch) return
  const user = useUserStore()
  user.hydrateFromStorage()
  if (!user.isLoggedIn) return
  if (!hasGuidePin()) return

  await fetchConfigs()
  // 接口无数据也算完成本次尝试，避免空配置反复请求；有数据再弹
  homeGuideAttemptedThisLaunch = true
  if (showPopup.value || showGuide.value) return

  const { state, popupItem } = judgePopup(popupList.value, PopUpTime.beforeGuide)
  if (state && popupItem && Object.keys(popupItem).length) {
    openPopup(popupItem)
    return
  }
  openGuide()
}

/** Pay success / paid view */
export async function showEndTripPopup() {
  await fetchConfigs()
  const { state, popupItem } = judgePopup(popupList.value, PopUpTime.endTrip)
  if (state && popupItem && Object.keys(popupItem).length) openPopup(popupItem)
}

/** Before scan / use bike — legacy live OFF (showUsePopup call commented in old wechat) */
export function showBeforeUseCarPopup(): Promise<void> {
  return Promise.resolve()
}

export function onOpsPopupClose() {
  const popTime = Number(currPopup.value.popUpTime)
  showPopup.value = false
  if (popTime === PopUpTime.beforeGuide) {
    if (!openGuide()) {
      const after = judgePopup(popupList.value, PopUpTime.afterGuide)
      if (after.state && after.popupItem && Object.keys(after.popupItem).length) {
        openPopup(after.popupItem)
      }
    }
  } else if (popTime === PopUpTime.beforeUseCar) {
    beforeUseResolve?.()
    beforeUseResolve = null
  }
}

export function onGuideSheetClose() {
  showGuide.value = false
  const { state, popupItem } = judgePopup(popupList.value, PopUpTime.afterGuide)
  if (state && popupItem && Object.keys(popupItem).length) openPopup(popupItem)
}

/** Bind shared reactive state in page templates. */
export function useGuidePopup() {
  return {
    popupList,
    guideConfig,
    currPopup,
    showPopup,
    showGuide,
    initHomeOrMap: initHomeOrMapGuide,
    showEndTripPopup,
    showBeforeUseCarPopup,
    onPopupClose: onOpsPopupClose,
    onGuideClose: onGuideSheetClose,
    fetchConfigs,
  }
}
