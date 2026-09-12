import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUserInfo } from '@/api'

/** 入口权限码，与遗留 App 及 `shared/.../domain/permission/OpsPermissions.kt` 一致。 */
export const PermissionCode = {
  OperationScreen: '0221',
  RevenueScreen: 'LargeRevenueScreen',
} as const

export const usePermissionStore = defineStore('permission', () => {
  const codes = ref<string[]>([])
  const loaded = ref(false)

  function has(code: string): boolean {
    return codes.value.includes(code)
  }

  /** 逗号 = 全部满足，`||` = 满足其一。语义沿用遗留的 `v-hasCode` 指令。 */
  function hasExpression(expression: string): boolean {
    if (expression.includes(',')) {
      return expression.split(',').every((code) => has(code.trim()))
    }
    if (expression.includes('||')) {
      return expression.split('||').some((code) => has(code.trim()))
    }
    return has(expression.trim())
  }

  async function load(): Promise<void> {
    if (loaded.value) return
    try {
      const { success, data } = await getUserInfo()
      if (success) {
        codes.value = data?.codes ?? []
        loaded.value = true
      }
    } catch {
      // 保留未加载态，下次导航可重试。
    }
  }

  return { codes, loaded, has, hasExpression, load }
})
