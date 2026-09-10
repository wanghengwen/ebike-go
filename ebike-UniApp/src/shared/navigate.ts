export function navigate(
  type: 'to' | 'redirect' | 'reLaunch' | 'back' | 'switchTab',
  url?: string,
  delta = 1,
) {
  if (type === 'back') {
    uni.navigateBack({
      delta,
      fail: () => {
        uni.reLaunch({ url: '/pages/home/home' })
      },
    })
    return
  }
  if (!url) return
  const map = {
    to: uni.navigateTo,
    redirect: uni.redirectTo,
    reLaunch: uni.reLaunch,
    switchTab: uni.switchTab,
  } as const
  map[type]({ url })
}

export function setNavTitle(title: string) {
  uni.setNavigationBarTitle({ title })
}
