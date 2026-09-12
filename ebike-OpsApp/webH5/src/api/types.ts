/** 网关统一响应包（Java `com.xyy.dto.Result`）。 */
export interface ApiResponse<T = unknown> {
  /** 成功是 `"0"`，字符串不是数字。业务错误码见 [ApiCode]。 */
  code: string
  success: boolean
  /** 网关发的是 `msg`，不是 `message`。 */
  msg?: string
  data: T
}

/** `/oauth/token` 返回体。 */
export interface TokenPayload {
  accessToken: string
  refreshToken: string
  expiresIn?: number
}

export interface UserInfo {
  userId?: string
  userName?: string
  phone?: string
  /** 权限码列表，控制大屏各卡片可见性。 */
  codes: string[]
}

export interface ServiceArea {
  id: string
  name: string
}

export interface AgentItem {
  id: string
  name: string
}

/** 网关业务错误码。 */
export const ApiCode = {
  /** token 无效或过期 */
  TokenExpired: '00005',
  /** token 无效 */
  TokenInvalid: '00013',
  /** 刷新 token 时账号被挤下线 */
  Kicked: '00015',
  /** Authorization 头缺失或格式不符 */
  MissingAuth: '00006',
  /** 通用失败 */
  Failed: '00001',
} as const

/** 命中后不把响应体交给业务层，由拦截器接管。 */
export const SILENT_CODES: readonly string[] = [
  ApiCode.TokenExpired,
  ApiCode.MissingAuth,
  ApiCode.TokenInvalid,
  ApiCode.Kicked,
]
