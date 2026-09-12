import { PermissionCode } from '@/stores/permission'

/** 搜索框输入被判成哪一类。搜索框只有一个，不让用户先选类型，与遗留 App 一致。 */
export type SearchKind = 'phone' | 'imei' | 'carId' | 'name'

/** 大陆号段。逐字符抄自遗留 `RegexUtil.CHINA_PATTERN`，不要顺手"优化"，号段是业务口径。 */
const CHINA_PHONE =
  /^((13[0-9])|(14[0,1,4-9])|(15[0-3,5-9])|(16[2,5,6,7])|(17[0-8])|(18[0-9])|(19[0-3,5-9]))\d{8}$/
/** 香港 8 位。遗留 `RegexUtil.HK_PATTERN`。 */
const HK_PHONE = /^(5|6|8|9)\d{7}$/

const DIGITS = /^[0-9]+$/

export function isPhone(value: string): boolean {
  return CHINA_PHONE.test(value) || HK_PHONE.test(value)
}

export function isImei(value: string): boolean {
  return DIGITS.test(value) && value.length === 15
}

/**
 * 车号：9 位纯数字，或 `08` 开头的 7/8 位。
 *
 * 遗留实现里第二条是死的——它写的是 `Regex("^08").matches(v0)`，而 Kotlin 的
 * `matches` 要求整串匹配（部分匹配得用 `containsMatchIn`），所以只有恰好等于 "08"
 * 时才为真，那时长度又是 2，后面的长度判断必然不成立。结果 `08` 开头的短车号会
 * 一路掉到「姓名查人」，最后报"用户不存在"。这里按它显然的本意实现。
 */
export function isCarId(value: string): boolean {
  if (!DIGITS.test(value)) return false
  if (value.length === 9) return true
  return value.startsWith('08') && (value.length === 7 || value.length === 8)
}

/**
 * 判定优先级与遗留 App 的 `when` 一致：手机号 → IMEI → 车号 → 兜底当姓名。
 * 顺序不能改：15 位纯数字既不是手机号也不是车号，只可能是 IMEI。
 */
export function classify(raw: string): SearchKind {
  const value = raw.replace(/\s/g, '')
  if (isPhone(value)) return 'phone'
  if (isImei(value)) return 'imei'
  if (isCarId(value)) return 'carId'
  return 'name'
}

/** 人员类（手机号 / 姓名）走 121202，车辆类（车号 / IMEI）走 121201。 */
export function requiredCode(kind: SearchKind): string {
  return kind === 'phone' || kind === 'name'
    ? PermissionCode.OrderQueryPersonal
    : PermissionCode.OrderQueryVehicle
}

/**
 * 搜索框 placeholder 按可查范围变化。
 *
 * 遗留 App 把这四句写进了 `obsOrderQueryScopeTip`，但布局 XML 绑的是写死的全量文案，
 * 所以旧版实际永远显示第四句——只有请求会被权限拦掉，提示不会变。这里按本意接上。
 */
export function scopeHint(canQueryPerson: boolean, canQueryVehicle: boolean): string {
  if (canQueryPerson && canQueryVehicle) return '请输入手机号、姓名、车辆号、设备号'
  if (canQueryPerson) return '请输入手机号、姓名'
  if (canQueryVehicle) return '请输入车辆号、设备号'
  return ''
}
