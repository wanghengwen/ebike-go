import { getPersonInfo } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { useTempDataStore } from '@/stores/tempData'
import { navigate } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

export function useFaceCheck() {
  const user = useUserStore()
  const temp = useTempDataStore()

  /** Legacy checkIsNeedFace — live WeChat has unlock-path face gate commented out */
  async function checkNeedFace(_profile?: Record<string, unknown>): Promise<boolean> {
    // legacy live OFF: unlock does not require face; identity pages still use goFaceAuth directly
    return false
  }

  function goFaceAuth(
    profile: Record<string, unknown> = {},
    callback = 'scan_use_bike',
  ) {
    const authName = encodeURIComponent(String(profile.authName || profile.realName || ''))
    const authNo = encodeURIComponent(String(profile.authNo || profile.idCard || ''))
    navigate(
      'to',
      `/pages-sub/account/identity-auth/identity-auth?authName=${authName}&authNo=${authNo}&izOnCertification=true&callback=${callback}`,
    )
  }

  function markFaceSuccess() {
    storage.set('faceResultSuccessTime', Date.now())
    temp.setFaceResult('1')
  }

  function clearFaceCache() {
    storage.remove('faceResultSuccessTime')
    temp.setFaceResult('')
  }

  /** Call when entering riding page — legacy clears face cache after unlock */
  function onEnterRiding() {
    clearFaceCache()
  }

  async function refreshAndCheck(): Promise<boolean> {
    try {
      const res = await getPersonInfo()
      if (res.success && res.data) {
        user.setUserInfo(res.data as never)
        return checkNeedFace(res.data as Record<string, unknown>)
      }
    } catch (e) {
      logger.warn('face profile soft fail', e)
    }
    return checkNeedFace()
  }

  return {
    checkNeedFace,
    goFaceAuth,
    markFaceSuccess,
    clearFaceCache,
    onEnterRiding,
    refreshAndCheck,
  }
}
